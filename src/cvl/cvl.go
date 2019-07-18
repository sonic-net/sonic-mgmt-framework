package cvl
import (
	"fmt"
	"os"
	"strings"
	"regexp"
	"time"
	 log "github.com/golang/glog"
	//"encoding/xml"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/jsonquery"
	"cvl/internal/yparser"
	"cvl/internal/util"
)

//DB number 
const (
	APPL_DB uint8 = 0 + iota
	ASIC_DB
	COUNTERS_DB
	LOGLEVEL_DB
	CONFIG_DB
	PFC_WD_DB
	FLEX_COUNTER_DB = PFC_WD_DB
	STATE_DB
	SNMP_OVERLAY_DB
	INVALID_DB
)

var reLeafRef *regexp.Regexp = nil
var reHashRef *regexp.Regexp = nil
var cvlInitialized bool
var dbNameToDbNum map[string]uint8

//map of lua script loaded
var luaScripts map[string]*redis.Script

//var tmpDbCache map[string]interface{} //map of table storing map of key-value pair
					//m["PORT_TABLE] = {"key" : {"f1": "v1"}}
//Important schema information to be loaded at bootup time
type modelTableInfo struct {
	dbNum uint8
	modelName string
	module *yparser.YParserModule
	keys []string
	redisKeyDelim string
	redisKeyPattern string
	mapLeaf []string //for 'mapping  list'
	leafRef map[string][]string //for storing all leafrefs for a leaf in a table, 
				//multiple leafref possible for union 
	mustExp map[string]string
	tablesForMustExp map[string]bool
}


/* CVL Error Structure. */
type CVLErrorInfo struct {
	TableName string      /* Table having error */
	ErrCode  CVLRetCode   /* Error Code describing type of error. */
	Keys    []string      /* Keys of the Table having error. */
        Value    string        /* Field Value throwing error */
        Field	 string        /* Field Name throwing error . */
	Msg     string        /* Detailed error message. */
	ConstraintErrMsg  string  /* Constraint error message. */
	ErrAppTag string
}

type CVL struct {
	redisClient *redis.Client
	yp *yparser.YParser
	tmpDbCache map[string]interface{} //map of table storing map of key-value pair
	batchLeaf string
}

type modelNamespace struct {
	prefix string
	ns string
}

type modelDataInfo struct {
	modelNs map[string]modelNamespace//model namespace 
	tableInfo map[string]modelTableInfo //redis table to model name and keys
}

//Struct for storing global DB cache to store DB which are needed frequently like PORT
type dbCachedData struct {
	root *yparser.YParserNode //Root of the cached data
	startTime time.Time  //When cache started
	expiry uint16    //How long cache should be maintained in sec
}

//Global data cache for redis table
type cvlGlobalSessionType struct {
	db map[string]dbCachedData
	pubsub *redis.PubSub
	stopChan chan int //stop channel to stop notification listener
	cv *CVL
}
var cvg cvlGlobalSessionType

//Single redis client for validation
var redisClient *redis.Client

//Stores important model info
var modelInfo modelDataInfo

type keyValuePairStruct struct {
	key string
	values []string
}

func TRACE_LOG(level log.Level, fmtStr string, args ...interface{}) {
	util.TRACE_LOG(level, fmtStr, args...)
}

//package init function 
func init() {
	if (os.Getenv("CVL_SCHEMA_PATH") != "") {
		util.CVL_SCHEMA = os.Getenv("CVL_SCHEMA_PATH") + "/"
	}

	if (os.Getenv("CVL_DEBUG") != "") {
		util.SetTrace(true)
	}

	//regular expression for leafref and hashref finding
	reLeafRef = regexp.MustCompile(`.*[/]([a-zA-Z]*:)?(.*)[/]([a-zA-Z]*:)?(.*)`)
	reHashRef = regexp.MustCompile(`\[(.*)\|(.*)\]`)

	Initialize()

	cvg.db = make(map[string]dbCachedData)
	//Global session keeps the global cache
	cvg.cv, _ = ValidatorSessOpen()

	dbCacheSet(false, "PORT", 0)
}

func Debug(on bool) {
	yparser.Debug(on)
}

//Get attribute value of xml node
func getXmlNodeAttr(node *xmlquery.Node, attrName string) string {
	for _, attr := range node.Attr {
		if (attrName == attr.Name.Local) {
			return attr.Value
		}
	}

	return ""
}

//Store useful schema data during initialization
func storeModelInfo(modelFile string, module *yparser.YParserModule) { //such model info can be maintained in C code and fetched from there 
	f, err := os.Open(util.CVL_SCHEMA + modelFile)
	root, err := xmlquery.Parse(f)

	if  err != nil {
		return
	}
	f.Close()

	//model is derived from file name
	tokens := strings.Split(modelFile, ".")
	modelName := tokens[0]

	//Store namespace
	modelNs := modelNamespace{}

	nodes := xmlquery.Find(root, "//module/namespace")
	if (nodes != nil) {
		modelNs.ns = nodes[0].Attr[0].Value
	}

	nodes = xmlquery.Find(root, "//module/prefix")
	if (nodes != nil) {
		modelNs.prefix = nodes[0].Attr[0].Value
	}

	modelInfo.modelNs[modelName] = modelNs

	//Store metadata present in each list
	nodes = xmlquery.Find(root, "//module/container/list")
	if (nodes == nil) {
		return
	}

	for  _, node := range nodes {
		//for each list
		tableName :=  node.Attr[0].Value
		tableInfo := modelTableInfo{modelName: modelName}
		//Store the reference for list node to be used later
		listNode := node
		node = node.FirstChild
		//Default database is CONFIG_DB since CVL works with config db mainly
		tableInfo.module = module
		tableInfo.dbNum = CONFIG_DB
		//default delim '|'
		tableInfo.redisKeyDelim = "|"

		fieldCount := 0

		//Check for meta data in schema
		for node !=  nil {
			switch node.Data {
			case "db-name":
				tableInfo.dbNum = dbNameToDbNum[node.Attr[0].Value]
				fieldCount++
			case "key":
				tableInfo.keys = strings.Split(node.Attr[0].Value," ")
				fieldCount++
			case "key-delim":
				tableInfo.redisKeyDelim = node.Attr[0].Value
				fieldCount++
			case "key-pattern":
				tableInfo.redisKeyPattern = node.Attr[0].Value
				fieldCount++
			case "map-leaf":
				tableInfo.mapLeaf = strings.Split(node.Attr[0].Value," ")
				fieldCount++
			}
			node = node.NextSibling
		}

		//Find and store all leafref under each table
		if (listNode == nil) {
			//Store the tableInfo in global data
			modelInfo.tableInfo[tableName] = tableInfo

			continue
		}

		leafRefNodes := xmlquery.Find(listNode, "//type[@name='leafref']")
		if (leafRefNodes == nil) {
			//Store the tableInfo in global data
			modelInfo.tableInfo[tableName] = tableInfo

			continue
		}

		tableInfo.leafRef = make(map[string][]string)
		for _, leafRefNode := range leafRefNodes {
			if (leafRefNode.Parent == nil || leafRefNode.FirstChild == nil) {
				continue
			}

			//Get the leaf/leaf-list name holding this leafref
			//Note that leaf can have union of leafrefs
			leafName := ""
			for node := leafRefNode.Parent; node != nil; node = node.Parent {
				if  (node.Data == "leaf" || node.Data == "leaf-list") {
					leafName = getXmlNodeAttr(node, "name")
					break
				}
			}

			//Store the leafref path
			if (leafName != "") {
				tableInfo.leafRef[leafName] = append(tableInfo.leafRef[leafName],
				getXmlNodeAttr(leafRefNode.FirstChild, "value"))
			}
		}

		//Find all 'must' expression and store the agains its parent node
		mustExps := xmlquery.Find(listNode, "//must")
		if (mustExps == nil) {
			//Update the tableInfo in global data
			modelInfo.tableInfo[tableName] = tableInfo
			continue
		}

		tableInfo.mustExp = make(map[string]string)
		for _, mustExp := range mustExps {
			if (mustExp.Parent == nil) {
				continue
			}
			parentName := ""
			for node := mustExp.Parent; node != nil; node = node.Parent {
				//assuming must exp is at leaf or list level
				if  (node.Data == "leaf" || node.Data == "leaf-list" ||
				node.Data == "list") {
					parentName = getXmlNodeAttr(node, "name")
					break
				}
			}
			if (parentName != "") {
				tableInfo.mustExp[parentName] = getXmlNodeAttr(mustExp, "condition")
			}
		}

		//Update the tableInfo in global data
		modelInfo.tableInfo[tableName] = tableInfo
	}
}

//Find the tables names in must expression, these tables data need to be fetched 
//during semantic validation
func addTableNamesForMustExp() {

	for tblName, tblInfo := range  modelInfo.tableInfo {
		if (tblInfo.mustExp == nil) {
			continue
		}

		for _, mustExp := range tblInfo.mustExp {
			tblInfo.tablesForMustExp = make(map[string]bool)
			//store the current table always so that expression check with other row
			tblInfo.tablesForMustExp[tblName] = true

			//check which table name is present in the must expression
			for tblNameSrch, _ := range modelInfo.tableInfo {
				if (tblNameSrch == tblName) {
					continue
				}
				//Table name should appear like "../VLAN_MEMBER/tagging_mode' or '
				// "/prt:PORT/prt:ifname"
				re := regexp.MustCompile(fmt.Sprintf(".*[/]([a-zA-Z]*:)?%s", tblNameSrch))
				matches := re.FindStringSubmatch(mustExp)
				if (len(matches) > 0) {
					//stores the table name 
					tblInfo.tablesForMustExp[tblNameSrch] = true
				}
			}
		}

		//update map
		modelInfo.tableInfo[tblName] = tblInfo
	}
}

//Convert Redis key to Yang keys, if multiple key components are there,
//they are separated based on Yang schema
func getRedisToYangKeys(tableName string, redisKey string)[]keyValuePairStruct{
	keyNames := modelInfo.tableInfo[tableName].keys
	//First split all the keys components
	keyVals := strings.Split(redisKey, modelInfo.tableInfo[tableName].redisKeyDelim) //split by DB separator
	//Store patterns for each key components by splitting using key delim
	keyPatterns := strings.Split(modelInfo.tableInfo[tableName].redisKeyPattern,
	                            modelInfo.tableInfo[tableName].redisKeyDelim) //split by DB separator

	if (len(keyNames) != len(keyVals)) {
		return nil //number key names and values does not match
	}

	mkeys := []keyValuePairStruct{}
	//For each key check the pattern and store key/value pair accordingly
	for  idx, keyName := range keyNames {

		//check if key-pattern contains specific key pattern
		if (keyPatterns[idx+1] == fmt.Sprintf("({%s},)*", keyName)) {   // key pattern is "({key},)*" i.e. repeating keysseperated by ','   
			repeatedKeys := strings.Split(keyVals[idx], ",")
			mkeys = append(mkeys, keyValuePairStruct{keyName, repeatedKeys})

		} else if (keyPatterns[idx+1] == fmt.Sprintf("{%s}", keyName)) { //no specific key pattern - just "{key}"

			//Store key/value mapping     
			mkeys = append(mkeys, keyValuePairStruct{keyName,  []string{keyVals[idx]}})
		}
	}

	return mkeys
}


//Add child node to a parent node
func(c *CVL) addChildNode1(tableName string, parent *yparser.YParserNode, name string) *yparser.YParserNode {

	//return C.lyd_new(parent, modelInfo.tableInfo[tableName].module, C.CString(name))
	return c.yp.AddChildNode(modelInfo.tableInfo[tableName].module, parent, name)
}

/*
func (c *CVL) addChildNode(parent *xmlquery.Node, xmlChildNode *xmlquery.Node) {
	xmlChildNode.Parent = parent
	if (parent.FirstChild == nil) {
		parent.FirstChild = xmlChildNode
	} else  {
		parent.LastChild.NextSibling = xmlChildNode
	}
	parent.LastChild = xmlChildNode
}
*/

//Add all other table data for validating all 'must' exp for tableName
func (c *CVL) addTableDataForMustExp(tableName string) CVLRetCode {
	//Check in cache first and merge
	if (cvg.db[tableName].root != nil) {
		var errObj yparser.YParserError
		//If global cache has the table, add to the session validation
		if errObj = c.yp.CacheSubtree(cvg.db[tableName].root); errObj.ErrCode != yparser.YP_SUCCESS {
			return CVL_SYNTAX_ERROR
		}
		return CVL_SUCCESS
	}


	if (modelInfo.tableInfo[tableName].mustExp == nil) {
		return CVL_SUCCESS
	}

	for mustTblName, _ := range modelInfo.tableInfo[tableName].tablesForMustExp {
		tableKeys, err:= redisClient.Keys(mustTblName +
		modelInfo.tableInfo[mustTblName].redisKeyDelim + "*").Result()

		if (err != nil) {
			continue
		}

		for _, tableKey := range tableKeys {
			tableKey = tableKey[len(mustTblName+ modelInfo.tableInfo[mustTblName].redisKeyDelim):] //remove table prefix
			if (c.tmpDbCache[mustTblName] == nil) {
				c.tmpDbCache[mustTblName] = map[string]interface{}{tableKey: nil}
			} else {
				tblMap := c.tmpDbCache[mustTblName]
				tblMap.(map[string]interface{})[tableKey] =nil
				c.tmpDbCache[mustTblName] = tblMap
			}
		}
	}

	return CVL_SUCCESS
}

func (c *CVL) addUpdateDataToCache(tableName string, redisKey string) {
	if (c.tmpDbCache[tableName] == nil) {
		c.tmpDbCache[tableName] = map[string]interface{}{redisKey: nil}
	} else {
		tblMap := c.tmpDbCache[tableName]
		tblMap.(map[string]interface{})[redisKey] =nil
		c.tmpDbCache[tableName] = tblMap
	}
}

//Check delete constraint for leafref if key/field is deleted
func (c *CVL) checkDeleteConstraint(tableName, keyVal, field string) CVLRetCode {
	var leafRefs []tblFieldPair
	if (field != "") {
		//Leaf or field is getting deleted
		leafRefs = c.findUsedAsLeafRef(tableName, field)
	} else {
		//Entire entry is getting deleted
		leafRefs = c.findUsedAsLeafRef(tableName, modelInfo.tableInfo[tableName].keys[0])
	}

	//The entry getting deleted might have been referred from multiple tables
	//Return failure if at-least one table is using this entry
	for _, leafRef := range leafRefs {
		TRACE_LOG(1, "Checking delete constraint for leafRef %s/%s", leafRef.tableName, leafRef.field)
		var nokey []string
		refKeyVal, err := luaScripts["find_key"].Run(redisClient, nokey, leafRef.tableName,
		modelInfo.tableInfo[leafRef.tableName].redisKeyDelim, leafRef.field, keyVal).Result()
		if (err == nil &&  refKeyVal != "") {
			TRACE_LOG(1, "Delete will violate the constraint as entry %s is referred in %s", tableName, refKeyVal)

			return CVL_SEMANTIC_ERROR
		}
	}


	return CVL_SUCCESS
}

//Add the data which are referring this key
func (c *CVL) updateDeleteDataToCache(tableName string, redisKey string) {
	if _, existing := c.tmpDbCache[tableName]; existing == false {
		return
	} else {
		tblMap := c.tmpDbCache[tableName]
		if _, existing := tblMap.(map[string]interface{})[redisKey]; existing == true {
			delete(tblMap.(map[string]interface{}), redisKey)
		}
		c.tmpDbCache[tableName] = tblMap
	}
}

//Find which all tables (and which field) is using given (tableName/field)
// as leafref
//Use LUA script to find if table has any entry for this leafref

type tblFieldPair struct {
	tableName string
	field string
}

func (c *CVL) findUsedAsLeafRef(tableName, field string) []tblFieldPair {

	var tblFieldPairArr []tblFieldPair

	for tblName, tblInfo := range  modelInfo.tableInfo {
		if (tableName == tblName) {
			continue
		}
		if (len(tblInfo.leafRef) == 0) {
			continue
		}

		for fieldName, leafRefs  := range tblInfo.leafRef {
			found := false
			//Find leafref by searching table and field name
			for _, leafRef := range leafRefs {
				if ((strings.Contains(leafRef, tableName) == true) &&
				(strings.Contains(leafRef, field) == true)) {
					tblFieldPairArr = append(tblFieldPairArr,
					tblFieldPair{tblName, fieldName})
					//Found as leafref, no need to search further
					found = true
					break
				}
			}

			if (found == true) {
				break
			}
		}
	}

	return tblFieldPairArr
}

//Add leafref entry for caching
func (c *CVL) addLeafRef(config bool, tableName string, name string, value string) {

	if (config == false) {
		return
	}

	//Check if leafRef entry is there for this field
	if (len(modelInfo.tableInfo[tableName].leafRef[name]) > 0) { //array of leafrefs for a leaf
		for _, leafRef  := range modelInfo.tableInfo[tableName].leafRef[name] {

			//Get reference table name from the path and the leaf name
			matches := reLeafRef.FindStringSubmatch(leafRef)

			//We have the leafref table name and the leaf name as well
			if (matches != nil && len(matches) == 5) { //whole + 4 sub matches
				refTableName := matches[2]
				redisKey := value
				//only key is there, value wil be fetched and stored here, 
				//if value can fetched this entry will be deleted that time
				if (c.tmpDbCache[refTableName] == nil) {
					c.tmpDbCache[refTableName] = map[string]interface{}{redisKey: nil}
				} else {
					tblMap := c.tmpDbCache[refTableName]
					_, exist := tblMap.(map[string]interface{})[redisKey]
					if (exist == false) {
						tblMap.(map[string]interface{})[redisKey] = nil
						c.tmpDbCache[refTableName] = tblMap
					}
				}
			}
		}
	}
}


func (c *CVL) addChildLeaf1(config bool, tableName string, parent *yparser.YParserNode, name string, value string) {
	//Batch leaf creation
	c.batchLeaf = c.batchLeaf + name + "#" + value + "#"
	//Check if this leaf has leafref,
	//If so add the add redis key to its table so that those 
	// details can be fetched for dependency validation

	c.addLeafRef(config, tableName, name, value)
}

/*
func (c *CVL) addChildLeaf(config bool, tableName string, parent *xmlquery.Node, name string, value string) *xmlquery.Node{
	//Create leaf name
	xmlLeafNode := &xmlquery.Node{
		Data: name,
		Type: xmlquery.ElementNode,
	}

	//Create leaf text
	xmlLeafNodeText := &xmlquery.Node{
		Data: value,
		Type: xmlquery.TextNode,
	}

	//Attach leaf node to its parent
	xmlLeafNode.Parent = parent
	if (parent.FirstChild == nil) {
		parent.FirstChild = xmlLeafNode
	} else  {
		parent.LastChild.NextSibling = xmlLeafNode
	}
	parent.LastChild = xmlLeafNode

	//Attach leaf text to leaf node
	xmlLeafNodeText.Parent = xmlLeafNode
	xmlLeafNode.FirstChild = xmlLeafNodeText
	xmlLeafNode.LastChild = xmlLeafNodeText

	//Check if this leaf has leafref,
	//If so add the add redis key to its table so that those 
	// details can be fetched for dependency validation

	c.addLeafRef(config, tableName, name, value)

	return xmlLeafNode
}
*/

func (c *CVL) checkFieldMap(fieldMap *map[string]string) map[string]interface{} {
	fieldMapNew := map[string]interface{}{}

	for field, value := range *fieldMap {
		if (field == "NULL") {
			continue
		} else if (field[len(field)-1:] == "@") {
			//last char @ means it is a leaf-list/array of fields
			field = field[:len(field)-1] //strip @
			//split the values seprated using ','
			fieldMapNew[field] = strings.Split(value, ",")
		} else {
			fieldMapNew[field] = value
		}
	}

	return fieldMapNew
}

//populate redis data to cache
func (c *CVL) fetchDataToTmpCache1() *yparser.YParserNode {
	for tableName, dbKeys := range c.tmpDbCache { //for each table

		if (len(dbKeys.(map[string]interface{}))  == 0) {
			continue
		}

		mCmd := map[string]*redis.StringStringMapCmd{}
		pipe := redisClient.Pipeline()

		for dbKey, _ := range dbKeys.(map[string]interface{}) { //for all keys
			redisKey := tableName + modelInfo.tableInfo[tableName].redisKeyDelim + dbKey
			mCmd[dbKey] = pipe.HGetAll(redisKey) //write into pipeline
			if mCmd[dbKey] == nil {
				TRACE_LOG(1, "Failed pipe.HGetAll('%s')", redisKey)
			}
		}

		_, err := pipe.Exec()
		if err != nil {
			TRACE_LOG(1, "Failed to fetch details for table %s", tableName)
		}

		mapTable := c.tmpDbCache[tableName]

		for key, val := range mCmd {
			res, err := val.Result()
			if (err != nil || len(res) == 0) {
				//no data found, don't keep blank entry
				delete(mapTable.(map[string]interface{}), key)
				continue
			}
			//exclude table name and delim
			keyOnly := key
			fieldMap := c.checkFieldMap(&res)
			mapTable.(map[string]interface{})[keyOnly] = fieldMap
		}

		pipe.Close()
	}

	if (util.Tracing == true) {
		jsonDataBytes, _ := json.Marshal(c.tmpDbCache)
		jsonData := string(jsonDataBytes)
		TRACE_LOG(1, "Top Node=%v\n", jsonData)
	}

	data, err := jsonquery.ParseJsonMap(&c.tmpDbCache)

	if (err != nil) {
		return nil
	}

	var root *yparser.YParserNode = nil
	var errObj yparser.YParserError

	for jsonNode := data.FirstChild; jsonNode != nil; jsonNode=jsonNode.NextSibling {
		TRACE_LOG(1, "Top Node=%v\n", jsonNode.Data)
		//Visit each top level list in a loop for creating table data
		topNode, _ := c.generateTableData1(true, jsonNode)
		if (root == nil) {
			root = topNode
		} else {
			if root, errObj = c.yp.MergeSubtree(root, topNode); errObj.ErrCode != yparser.YP_SUCCESS {
				return nil
			}
		}
	}

	if root != nil && util.Tracing == true {
		TRACE_LOG(5, "Dependent Data = %v\n", c.yp.NodeDump(root))
	}

	return root
}


/*
func (c *CVL) fetchDataToTmpCache() string {
	for tableName, dbKeys := range c.tmpDbCache { //for each table

		if (len(dbKeys.(map[string]interface{}))  == 0) {
			continue
		}

		mCmd := map[string]*redis.StringStringMapCmd{}
		pipe := redisClient.Pipeline()

		for dbKey, _ := range dbKeys.(map[string]interface{}) { //for all keys
			redisKey := tableName + modelInfo.tableInfo[tableName].redisKeyDelim + dbKey
			mCmd[dbKey] = pipe.HGetAll(redisKey) //write into pipeline
			if mCmd[dbKey] == nil {
				TRACE_LOG(1, "Failed pipe.HGetAll('%s')", redisKey)
			}
		}

		_, err := pipe.Exec()
		if err != nil {
			TRACE_LOG(1, "Failed to fetch details for table %s", tableName)
		}

		mapTable := c.tmpDbCache[tableName]

		for key, val := range mCmd {
			res, err := val.Result()
			if (err != nil || len(res) == 0) {
				//no data found, don't keep blank entry
				delete(mapTable.(map[string]interface{}), key)
				continue
			}
			//exclude table name and delim
			keyOnly := key
			//TBD: Need to check field name like <NULL>(to be deleted)
			//and 'members@' (strip '@')
			//store all field values
			fieldMap := c.checkFieldMap(&res)
			mapTable.(map[string]interface{})[keyOnly] = fieldMap
		}

		pipe.Close()
	}

	jsonDataBytes, _ := json.Marshal(c.tmpDbCache)
	jsonData := string(jsonDataBytes)

	data, err := jsonquery.Parse(strings.NewReader(jsonData))

	if (err != nil) {
		return ""
	}

	doc := &xmlquery.Node{
		Type: xmlquery.DeclarationNode,
		Data: "xml",
		Attr: []xml.Attr{
			xml.Attr{Name: xml.Name{Local: "version"}, Value: "1.0"},
		},
	}

	var xmlNode *xmlquery.Node
	//Generate xml data from status JSON data
	for jsonNode := data.FirstChild; jsonNode != nil; jsonNode=jsonNode.NextSibling {
		//TRACE_LOG(3, "Top Status Node=%v\n", jsonNode.Data)
		//Visit each top level list in a loop for creating table data
		xmlNode, _= c.generateTableData(false, jsonNode)

		xmlNode.Parent = doc
		if doc.FirstChild == nil {
			doc.FirstChild = xmlNode
			doc.LastChild = xmlNode
		} else {
			//If duplicate top node exists add child subtree in the existing top node
			dupTopNode := xmlquery.Find(doc, fmt.Sprintf("/%s", xmlNode.Data))
			if (dupTopNode != nil) {
				dupTopNode[0].LastChild.NextSibling = xmlNode.FirstChild
				dupTopNode[0].LastChild = xmlNode.LastChild
			} else {
				doc.LastChild.NextSibling = xmlNode
				doc.LastChild = xmlNode
			}
		}
	}

	if (doc.FirstChild == nil) {
		return "";
	}

	yangXml := doc.OutputXML(true)
	TRACE_LOG(5, "Dependent Data = %s\n", yangXml)
	return yangXml
}
*/

func (c *CVL) clearTmpDbCache() {
	for key := range c.tmpDbCache {
		delete(c.tmpDbCache, key)
	}
}

/*
//This function is not used currently
func (c *CVL) addStatusData(tableName string) *xmlquery.Node {
	//Get all keys from DB
	pattern := tableName + modelInfo.tableInfo[tableName].redisKeyDelim + "*"
	dbKeys, err := redisClient.Keys(pattern).Result()

	if (err != nil) {
		return nil
	}

	mCmd := map[string]*redis.StringStringMapCmd{}
	pipe := redisClient.Pipeline()

	for _, dbKey := range dbKeys {
		if ((dbKey == "PORT_TABLE:PortConfigDone") ||
		(dbKey == "PORT_TABLE:PortInitDone")) {
			continue
		}
		mCmd[dbKey] = pipe.HGetAll(dbKey)
		if err != nil {
			return nil
		}
	}
	_, err = pipe.Exec()
	if err != nil {
		return nil
	}

	mapTable := make(map[string]interface{})
	mapData := make(map[string]interface{})

	for key, val := range mCmd {
		res, _ := val.Result()
		//exclude table name and delim
		keyOnly := key[len(tableName + modelInfo.tableInfo[tableName].redisKeyDelim):]
		mapData[keyOnly] = res 
	}
	mapTable[tableName] = mapData
	jsonData, err := json.Marshal(mapTable)

	data, err := jsonquery.Parse(strings.NewReader(string(jsonData)))

	if (err != nil) {
		return nil
	}

	doc := &xmlquery.Node{   //module level top level node will reduce namespace repition
		Data: "top",
		Type: xmlquery.ElementNode,
	}

	var xmlNode *xmlquery.Node

	for jsonNode := data.FirstChild; jsonNode != nil; jsonNode=jsonNode.NextSibling {
		TRACE_LOG(1, "Leafref table top node=%v\n", jsonNode.Data)
		//Visit each top level list in a loop for creating table data
		xmlNode, _= c.generateTableData(false, jsonNode)

		xmlNode.Parent = doc
		if doc.FirstChild == nil {
			doc.FirstChild = xmlNode
			doc.LastChild = xmlNode
		} else {
			//If duplicate top node exists add child subtree in the existing top node
			dupTopNode := xmlquery.Find(doc, fmt.Sprintf("/%s", xmlNode.Data))
			if (dupTopNode != nil) {
				dupTopNode[0].LastChild.NextSibling = xmlNode.FirstChild
				dupTopNode[0].LastChild = xmlNode.LastChild
			} else {
				doc.LastChild.NextSibling = xmlNode
				doc.LastChild = xmlNode
			}
		}
	}

	return doc
}
*/

func (c *CVL) generateTableFieldsData1(config bool, tableName string, jsonNode *jsonquery.Node,
parent *yparser.YParserNode) CVLRetCode {

	//Traverse fields
	for jsonFieldNode := jsonNode.FirstChild; jsonFieldNode!= nil;
	jsonFieldNode = jsonFieldNode.NextSibling {
		//Add fields as leaf to the list
		if (jsonFieldNode.Type == jsonquery.ElementNode &&
		jsonFieldNode.FirstChild != nil &&
		jsonFieldNode.FirstChild.Type == jsonquery.TextNode) {

			if (len(modelInfo.tableInfo[tableName].mapLeaf) == 2) {//mapping should have two leaf always
				//Values should be stored inside another list as map table
				listNode := c.addChildNode1(tableName, parent, tableName) //Add the list to the top node
				c.addChildLeaf1(config, tableName,
				listNode, modelInfo.tableInfo[tableName].mapLeaf[0],
				jsonFieldNode.Data)

				c.addChildLeaf1(config, tableName,
				listNode, modelInfo.tableInfo[tableName].mapLeaf[1],
				jsonFieldNode.FirstChild.Data)

			} else {
				//check if it is hash-ref, then need to add only key from "TABLE|k1"
				hashRefMatch := reHashRef.FindStringSubmatch(jsonFieldNode.FirstChild.Data)

				if (hashRefMatch != nil && len(hashRefMatch) == 3) {
				/*if (strings.HasPrefix(jsonFieldNode.FirstChild.Data, "[")) &&
				(strings.HasSuffix(jsonFieldNode.FirstChild.Data, "]")) &&
				(strings.Index(jsonFieldNode.FirstChild.Data, "|") > 0) {*/

					c.addChildLeaf1(config, tableName,
					parent, jsonFieldNode.Data,
					hashRefMatch[2]) //take hashref key value
				} else {
					c.addChildLeaf1(config, tableName,
					parent, jsonFieldNode.Data,
					jsonFieldNode.FirstChild.Data)
				}
			}

		} else if (jsonFieldNode.Type == jsonquery.ElementNode &&
		jsonFieldNode.FirstChild != nil &&
		jsonFieldNode.FirstChild.Type == jsonquery.ElementNode) {
			//Array data e.g. VLAN members
			for  arrayNode:=jsonFieldNode.FirstChild; arrayNode != nil;

			arrayNode = arrayNode.NextSibling {
				c.addChildLeaf1(config, tableName,
				parent, jsonFieldNode.Data,
				arrayNode.FirstChild.Data)
			}
		}
	}

	return CVL_SUCCESS
}

/*
//Generate Yang xml for all fields in a hash
func (c *CVL) generateTableFieldsData(config bool, tableName string, jsonNode *jsonquery.Node,
parent *xmlquery.Node) CVLRetCode {

	//Traverse fields
	for jsonFieldNode := jsonNode.FirstChild; jsonFieldNode!= nil;
	jsonFieldNode = jsonFieldNode.NextSibling {
		//Add fields as leaf to the list
		if (jsonFieldNode.Type == jsonquery.ElementNode &&
		jsonFieldNode.FirstChild.Type == jsonquery.TextNode) {

			if (len(modelInfo.tableInfo[tableName].mapLeaf) == 2) {//mapping should have two leaf always
				//Values should be stored inside another list as map table
				xmlNode := &xmlquery.Node{
					Data: tableName,
					Type: xmlquery.ElementNode,
				}
				c.addChildNode(parent, xmlNode)

				c.addChildLeaf(config, tableName,
				xmlNode, modelInfo.tableInfo[tableName].mapLeaf[0],
				jsonFieldNode.Data)

				c.addChildLeaf(config, tableName,
				xmlNode, modelInfo.tableInfo[tableName].mapLeaf[1],
				jsonFieldNode.FirstChild.Data)

			} else {
				//check if it is hash-ref, then need to add only key from "TABLE|k1"
				//re := regexp.MustCompile(`\[(.*)\|(.*)\]`)
				hashRefMatch := reHashRef.FindStringSubmatch(jsonFieldNode.FirstChild.Data)

				if (hashRefMatch != nil && len(hashRefMatch) == 3) {
					c.addChildLeaf(config, tableName,
					parent, jsonFieldNode.Data,
					hashRefMatch[2]) //take hashref key value
				} else {
					c.addChildLeaf(config, tableName,
					parent, jsonFieldNode.Data,
					jsonFieldNode.FirstChild.Data)
				}
			}

		} else if (jsonFieldNode.Type == jsonquery.ElementNode &&
		jsonFieldNode.FirstChild.Type == jsonquery.ElementNode) {
			//Array data e.g. VLAN members
			for  arrayNode:=jsonFieldNode.FirstChild; arrayNode != nil;

			arrayNode = arrayNode.NextSibling {
				c.addChildLeaf(config, tableName,
				parent, jsonFieldNode.Data,
				arrayNode.FirstChild.Data)
			}
		}
	}

	return CVL_SUCCESS
}
*/

func (c *CVL) generateTableData1(config bool, jsonNode *jsonquery.Node)(*yparser.YParserNode, CVLRetCode) {
	tableName := fmt.Sprintf("%s",jsonNode.Data)
	c.batchLeaf = ""

	var topNode *yparser.YParserNode
	topNode = c.yp.AddChildNode(modelInfo.tableInfo[tableName].module,
	nil, modelInfo.tableInfo[tableName].modelName)


	//Traverse each key instance
	for jsonNode = jsonNode.FirstChild; jsonNode != nil; jsonNode = jsonNode.NextSibling {

		//For each field check if is key 
		//If it is key, create list as child of top container
		// Get all key name/value pairs
		keyValuePair := getRedisToYangKeys(tableName, jsonNode.Data)
		keyCompCount := len(keyValuePair)
		totalKeyComb := 1
		var keyIndices []int

		//Find number of all key combinations
		//Each key can have one or more key values, which results in nk1 * nk2 * nk2 combinations
		idx := 0
		for i,_ := range keyValuePair {
			totalKeyComb = totalKeyComb * len(keyValuePair[i].values)
			keyIndices = append(keyIndices, 0)
		}

		for  ; totalKeyComb > 0 ; totalKeyComb-- {

			//Add table i.e. create list element
			listNode := c.addChildNode1(tableName, topNode, tableName) //Add the list to the top node

			//For each key combination
			//Add keys as leaf to the list
			for idx = 0; idx < keyCompCount; idx++ {
				c.addChildLeaf1(config, tableName,
				listNode, keyValuePair[idx].key,
				keyValuePair[idx].values[keyIndices[idx]])
			}

			//Get all fields under the key field and add them as children of the list
			c.generateTableFieldsData1(config, tableName, jsonNode, listNode)

			//Check which key elements left after current key element
			var next int = keyCompCount - 1
			for  ((next > 0) && ((keyIndices[next] +1) >=  len(keyValuePair[next].values))) {
				next--
			}
			//No more combination possible
			if (next < 0) {
				break
			}

			keyIndices[next]++

			//Reset indices for all other key elements
			for idx = next+1;  idx < keyCompCount; idx++ {
				keyIndices[idx] = 0
			}

			TRACE_LOG(1, "Starting batch leaf creation - %s\n", c.batchLeaf)
			//process batch leaf creation
			if errObj := c.yp.AddMultiLeafNodes(modelInfo.tableInfo[tableName].module, listNode, c.batchLeaf); errObj.ErrCode != yparser.YP_SUCCESS {
				return nil, CVL_SYNTAX_ERROR
			}
			c.batchLeaf = ""
		}
	}

	return topNode, CVL_SUCCESS
}

/*
// Generate YANG xml instance for Redis table
func (c *CVL) generateTableData(config bool, jsonNode *jsonquery.Node)(*xmlquery.Node, CVLRetCode) {

	tableName := fmt.Sprintf("%s",jsonNode.Data)

	//Create top level container based on first list
	xmlTopNode := &xmlquery.Node{   //module level top level node will reduce namespace repition
		Data: modelInfo.tableInfo[tableName].modelName,
		Type: xmlquery.ElementNode,
		Attr: []xml.Attr{
			xml.Attr{Name: xml.Name{Local: "xmlns"},
			Value:  modelInfo.modelNs[modelInfo.tableInfo[tableName].modelName].ns},
		},
	}

	//Traverse each key instance
	for jsonNode = jsonNode.FirstChild; jsonNode != nil; jsonNode = jsonNode.NextSibling {

		//For each field check if is key 
		//If it is key, create list as child of top container
		// Get all key name/value pairs
		keyValuePair := getRedisToYangKeys(tableName, jsonNode.Data)
		keyCompCount := len(keyValuePair)
		totalKeyComb := 1
		var keyIndices []int

		//Find number of all key combinations
		//Each key can have one or more key values, which results in nk1 * nk2 * nk2 combinations
		idx := 0
		for i,_ := range keyValuePair {
			totalKeyComb = totalKeyComb * len(keyValuePair[i].values)
			keyIndices = append(keyIndices, 0)
		}

		for  ; totalKeyComb > 0 ; totalKeyComb-- {

			//Add table i.e. create list element
			xmlNode := &xmlquery.Node{
				Data: tableName,
				Type: xmlquery.ElementNode,
			}

			//For each key combination
			//Add keys as leaf to the list
			for idx = 0; idx < keyCompCount; idx++ {
				c.addChildLeaf(config, tableName,
				xmlNode, keyValuePair[idx].key,
				keyValuePair[idx].values[keyIndices[idx]])
			}


			c.addChildNode(xmlTopNode, xmlNode) //Add the list to the top node

			//Get all fields under the key field and add them as children of the list
			c.generateTableFieldsData(config, tableName, jsonNode, xmlNode)


			//Check which key elements left after current key element
			var next int = keyCompCount - 1
			for  ((next > 0) && ((keyIndices[next] +1) >=  len(keyValuePair[next].values))) {
				next--
			}
			//No more combination possible
			if (next < 0) {
				break
			}

			keyIndices[next]++

			//Reset indices for all other key elements
			for idx = next+1;  idx < keyCompCount; idx++ {
				keyIndices[idx] = 0
			}
		}
	}


	return xmlTopNode, CVL_SUCCESS
}
*/

/*
func jsonMapToYangTree(jsonMap *map[string]interface{}) *yparser.YParserNode {
	jsonDoc, _ := jsonquery.ParseJsonMap(jsonMap)

	if (jsonDoc == nil) {
		return nil
	}
	return nil
}
*/

func (c *CVL) translateToYang1(jsonMap *map[string]interface{}) (*yparser.YParserNode, CVLRetCode) {
	//Parse the map data to json tree
	data, _ := jsonquery.ParseJsonMap(jsonMap)
	var root *yparser.YParserNode
	root = nil
	var errObj yparser.YParserError

	for jsonNode := data.FirstChild; jsonNode != nil; jsonNode=jsonNode.NextSibling {
		TRACE_LOG(1, "Top Node=%v\n", jsonNode.Data)
		//Visit each top level list in a loop for creating table data
		topNode, _ := c.generateTableData1(true, jsonNode)

		if  topNode == nil {
			return nil, CVL_SYNTAX_ERROR
		}

		if (root == nil) {
			root = topNode
		} else {
			if root, errObj = c.yp.MergeSubtree(root, topNode); errObj.ErrCode != yparser.YP_SUCCESS {
				return nil, CVL_SYNTAX_ERROR
			}
		}
	}

	return root, CVL_SUCCESS
}

/*
//Convert Redis JSON to Yang XML using translation metadata
func (c *CVL) translateToYang(jsonData string) (*xmlquery.Node, CVLRetCode) {
	var v interface{}
	jsonData1 :=`{
		"VLAN": {
			"Vlan100": {
				"members": [
				"Ethernet44",
				"Ethernet964"
				],
				"vlanid": "100"
			},
			"Vlan1200": {
				"members": [
				"Ethernet64",
				"Ethernet1008"
				],
				"vlanid": "1200"
			}
		}
	}`
	b := []byte(jsonData1)
	if err1 := json.Unmarshal(b, &v); err1 == nil {
		var value map[string]interface{} = v.(map[string]interface{})
		root, ret := c.translateToYang1(&value)
		if ret == CVL_SUCCESS && root != nil {
			var outBuf *C.char
			C.lyd_print_mem(&outBuf, root, C.LYD_XML, 0)

			fmt.Printf("\nLYD data = %v\n", C.GoString(outBuf))
		}
	}

	//Parse Entire JSON file
	data, err := jsonquery.Parse(strings.NewReader(jsonData))

	if err != nil {
		return nil, CVL_SYNTAX_ERROR
	}

	doc := &xmlquery.Node{
		Type: xmlquery.DeclarationNode,
		Data: "xml",
		Attr: []xml.Attr{
			xml.Attr{Name: xml.Name{Local: "version"}, Value: "1.0"},
		},
	}

	//var xmlNode, prevXmlNode *xmlquery.Node
	var xmlNode *xmlquery.Node

	for jsonNode := data.FirstChild; jsonNode != nil; jsonNode=jsonNode.NextSibling {
		TRACE_LOG(1, "Top Node=%v\n", jsonNode.Data)
		//Visit each top level list in a loop for creating table data
		xmlNode, _= c.generateTableData(true, jsonNode)

		xmlNode.Parent = doc
		if doc.FirstChild == nil {
			doc.FirstChild = xmlNode
			doc.LastChild = xmlNode
		} else {
			//If duplicate top node exists add child subtree in the existing top node
			dupTopNode := xmlquery.Find(doc, fmt.Sprintf("/%s", xmlNode.Data))
			if (dupTopNode != nil) {
				dupTopNode[0].LastChild.NextSibling = xmlNode.FirstChild
				dupTopNode[0].LastChild = xmlNode.LastChild
			} else {
				doc.LastChild.NextSibling = xmlNode
				doc.LastChild = xmlNode
			}
		}
	}

	return doc, CVL_SUCCESS
}
*/

/*
//Validate config - syntax and semantics
func (c *CVL) validate1 (data *yparser.YParserNode) CVLRetCode {

	depData := c.fetchDataToTmpCache1()
	/*
	if (depData != nil) {
		if (0 != C.lyd_merge_to_ctx(&data, depData, C.LYD_OPT_DESTRUCT, ctx)) {
			TRACE_LOG(1, "Failed to merge status data\n")
		}
	}

	if (0 != C.lyd_data_validate(&data, C.LYD_OPT_CONFIG, ctx)) {
		fmt.Println("Validation failed\n")
		return CVL_SYNTAX_ERROR
	}*

	c.yp.ValidateData(data, depData)

	return CVL_SUCCESS
}
*/

/*
	
func (c *CVL) validate (xmlData string) CVLRetCode {
	TRACE_LOG(1, "Validating \n%v\n....", xmlData)

	data := C.lyd_parse_data_mem(ctx, C.CString(xmlData), C.LYD_XML, C.LYD_OPT_EDIT)
	if ((C.ly_errno != 0) || (data == nil)) {
		//fmt.Println("Parsing data failed %d\n", C.ly_errno)
		return CVL_SYNTAX_ERROR
	}

	depData := c.fetchDataToTmpCache()
	if (depData != "") {
		depDataNode := C.lyd_parse_data_mem(ctx, C.CString(depData), C.LYD_XML, C.LYD_OPT_EDIT)

		if (0 != C.lyd_merge_to_ctx(&data, depDataNode, C.LYD_OPT_DESTRUCT, ctx)) {
			TRACE_LOG(1, "Failed to merge status data\n")
		}
	}


	if (0 != C.lyd_data_validate(&data, C.LYD_OPT_CONFIG, ctx)) {
                processErrorResp(ctx)
		return CVL_SYNTAX_ERROR
	}

	return CVL_SUCCESS
}
*/

//Perform syntax checks
func (c *CVL) validateSyntax1(data *yparser.YParserNode) (CVLErrorInfo, CVLRetCode) {
	var cvlErrObj CVLErrorInfo
	TRACE_LOG(1, "Validating syntax \n....")

	if errObj  := c.yp.ValidateSyntax(data); errObj.ErrCode != yparser.YP_SUCCESS {

			cvlErrObj =  CVLErrorInfo {
		             TableName : errObj.TableName,
			     Keys      : errObj.Keys,
			     Value     : errObj.Value,
			     Field     : errObj.Field,
			     Msg       : errObj.Msg,
			     ConstraintErrMsg : errObj.ErrTxt,
			     ErrAppTag	: errObj.ErrAppTag,
	   		} 


		retCode := CVLRetCode(errObj.ErrCode)

		return  cvlErrObj, retCode 
	}

	return cvlErrObj, CVL_SUCCESS
}

/*
func (c *CVL) validateSyntax(xmlData string) (CVLRetCode, *yparser.YParserNode) {
	TRACE_LOG(1, "Validating syntax \n%v\n....", xmlData)
	//parsing only does syntacial checks
	data := C.lyd_parse_data_mem(ctx, C.CString(xmlData), C.LYD_XML, C.LYD_OPT_EDIT)
	if ((C.ly_errno != 0) || (data == nil)) {
                processErrorResp(ctx)
		//fmt.Println("Parsing data failed\n")
		return CVL_SYNTAX_ERROR, nil
	}

	return CVL_SUCCESS, data
	return CVL_SUCCESS, nil
}
*/

//Perform semantic checks 
func (c *CVL) validateSemantics1(data *yparser.YParserNode, otherDepData *yparser.YParserNode) (CVLErrorInfo, CVLRetCode) {
	var cvlErrObj CVLErrorInfo
	//Get dependent data from 
	depData := c.fetchDataToTmpCache1() //fetch data to temp cache for temporary validation
	TRACE_LOG(1, "Validating semantics data=%s\n depData =%s\n, otherDepData=%s\n....", c.yp.NodeDump(data), c.yp.NodeDump(depData), c.yp.NodeDump(otherDepData))

	if errObj := c.yp.ValidateSemantics(data, depData, otherDepData); errObj.ErrCode != yparser.YP_SUCCESS {

			cvlErrObj =  CVLErrorInfo {
		             TableName : errObj.TableName,
			     Keys      : errObj.Keys,
			     Value     : errObj.Value,
			     Field     : errObj.Field,
			     Msg       : errObj.Msg,
			     ConstraintErrMsg : errObj.ErrTxt,
			     ErrAppTag	: errObj.ErrAppTag,
	   		} 


		retCode := CVLRetCode(errObj.ErrCode)

		return  cvlErrObj, retCode 
	}

/*
	//parse dependent data
	if (depData != nil) {

		//merge input data and dependent data for semantic validation
		if (0 != C.lyd_merge_to_ctx(&data, depData, C.LYD_OPT_DESTRUCT, ctx)) {
			TRACE_LOG(1, "Unable to merge dependent data\n")
			return CVL_SEMANTIC_ERROR
		}
	}

	if (otherDepData != nil) { //if other dep data is provided
		if (0 != C.lyd_merge_to_ctx(&data, otherDepData, C.LYD_OPT_DESTRUCT, ctx)) {
			TRACE_LOG(1, "Unable to merge other dependent data\n")
			return CVL_SEMANTIC_ERROR
		}
	}

	//Check semantic validation
	if (0 != C.lyd_data_validate(&data, C.LYD_OPT_CONFIG, ctx)) {
		fmt.Println("Validation failed\n")
		return CVL_SEMANTIC_ERROR
	}
*/
	return cvlErrObj ,CVL_SUCCESS
}

/*
//func (c *CVL) validateSemantics(data *yparser.YParserNode, otherDepData string) CVLRetCode {
func (c *CVL) validateSemantics(data, otherDepData string) CVLRetCode {

	//Get dependent data from 
	depData := c.fetchDataToTmpCache() //fetch data to temp cache for temporary validation
	//parse dependent data
	if (depData != "") {
		depDataNode := C.lyd_parse_data_mem(ctx, C.CString(depData), C.LYD_XML, C.LYD_OPT_EDIT)

		//merge input data and dependent data for semantic validation
		if (0 != C.lyd_merge_to_ctx(&data, depDataNode, C.LYD_OPT_DESTRUCT, ctx)) {
			TRACE_LOG(1, "Unable to merge dependent data\n")
			return CVL_SEMANTIC_ERROR
		}
	}

	if (otherDepData != "") { //if other dep data is provided
		otherDepDataNode := C.lyd_parse_data_mem(ctx, C.CString(otherDepData),
		C.LYD_XML, C.LYD_OPT_EDIT)

		if (0 != C.lyd_merge_to_ctx(&data, otherDepDataNode, C.LYD_OPT_DESTRUCT, ctx)) {
			TRACE_LOG(1, "Unable to merge other dependent data\n")
			return CVL_SEMANTIC_ERROR
		}
	}

	//Check semantic validation
	if (0 != C.lyd_data_validate(&data, C.LYD_OPT_CONFIG, ctx)) {
	if (0 != C.lyd_data_validate_all(C.CString(data), C.CString(depData), C.CString(otherDepData), C.LYD_OPT_CONFIG, ctx)) {
		fmt.Println("Validation failed\n")
		return CVL_SEMANTIC_ERROR
	}

	return CVL_SUCCESS
}
*/

//Add config data item to accumulate per table
func (c *CVL) addCfgDataItem(configData *map[string]interface{}, cfgDataItem CVLEditConfigData) (string, string){
	var cfgData map[string]interface{}//:= map[string]interface{}
	cfgData = *configData

	for tblName,_ := range modelInfo.tableInfo {
		//Check if table prefix matches to any schema table
		//i.e. has 'VLAN|' or 'PORT|' etc.
		if (strings.HasPrefix(cfgDataItem.Key, tblName + modelInfo.tableInfo[tblName].redisKeyDelim)) {
			prefixLen := len(tblName) + 1
			if (cfgDataItem.VOp == OP_DELETE) {
				//Don't add data it is delete operation
				return tblName, cfgDataItem.Key[prefixLen:]
			}
			if _, existing := cfgData[tblName]; existing {
				fieldsMap := cfgData[tblName].(map[string]interface{})
				fieldsMap[cfgDataItem.Key[prefixLen:]] = c.checkFieldMap(&cfgDataItem.Data)
			} else {
				fieldsMap := make(map[string]interface{})
				fieldsMap[cfgDataItem.Key[prefixLen:]] = c.checkFieldMap(&cfgDataItem.Data)
				cfgData[tblName] = fieldsMap
			}

			return tblName, cfgDataItem.Key[prefixLen:]
		}
	}

	return "",""
}

//Get the table data from redis and cache it in yang node format
//expiry =0 never expire the cache
func dbCacheSet(update bool, tableName string, expiry uint16) CVLRetCode {
	//Get the data from redis and save it
	tableKeys, err:= redisClient.Keys(tableName +
	modelInfo.tableInfo[tableName].redisKeyDelim + "*").Result()

	if (err != nil) {
		return CVL_FAILURE
	}

	tablePrefixLen := len(tableName + modelInfo.tableInfo[tableName].redisKeyDelim)
	for _, tableKey := range tableKeys {
		tableKey = tableKey[tablePrefixLen:] //remove table prefix
		if (cvg.cv.tmpDbCache[tableName] == nil) {
			cvg.cv.tmpDbCache[tableName] = map[string]interface{}{tableKey: nil}
		} else {
			tblMap := cvg.cv.tmpDbCache[tableName]
			tblMap.(map[string]interface{})[tableKey] =nil
			cvg.cv.tmpDbCache[tableName] = tblMap
		}
	}

	cvg.db[tableName] = dbCachedData{startTime:time.Now(), expiry: expiry,
	root: cvg.cv.fetchDataToTmpCache1()}

	//install keyspace notification for the table to update the cache
	if (update == false) {
		installDbChgNotif()
	}

	return CVL_SUCCESS
}

//Receive all updates for all tables on a single channel
func installDbChgNotif() {
	if (len(cvg.db) > 1) { //notif running for at least one table added previously
		cvg.stopChan <- 1 //stop active notification 
	}

	subList := ""
	for tableName, _ := range cvg.db {
		subList = subList + "__keyspace@" +
		fmt.Sprintf("%d", modelInfo.tableInfo[tableName].dbNum) + "__:" + tableName +
		modelInfo.tableInfo[tableName].redisKeyDelim + "*"

	}

	cvg.pubsub = redisClient.PSubscribe(subList)

	go func() {
		notifCh := cvg.pubsub.Channel()
		for {
			select  {
			case <-cvg.stopChan:
				//stop this routine
				return
			case msg:= <-notifCh:
				//Handle update
				dbCacheUpdate(msg.Channel, msg.Payload)

			}
		}
	}()
}

func dbCacheUpdate(tableName, op string) CVLRetCode {
	switch op {
	case "hset", "hmset", "hdel":
		//Delete the existing cache
		dbCacheClear(tableName)
		//Add the cache again in yang tree --> TBD:Optimie
		dbCacheSet(true, tableName, 0)
	case "del":
		//Delete the entry in yang tree --> TBD:Optimize
		dbCacheClear(tableName)
		dbCacheSet(true, tableName, 0)
	}

	return CVL_SUCCESS
}

//Clear cache data for given table
func dbCacheClear(tableName string) CVLRetCode {
	cvg.cv.yp.FreeNode(cvg.db[tableName].root)
	delete(cvg.db, tableName)

	return CVL_SUCCESS
}

