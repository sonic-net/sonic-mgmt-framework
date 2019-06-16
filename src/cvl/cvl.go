package cvl

import (
	"fmt"
	"os"
	"strings"
	"regexp"
	 log "github.com/golang/glog"
	"encoding/xml"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/jsonquery"
)

/*
#cgo CFLAGS: -I build/pcre-8.43/install/include -I build/libyang/build/include
#cgo LDFLAGS: -L build/pcre-8.43/install/lib -lpcre
#cgo LDFLAGS: -L build/libyang/build -lyang
#include <libyang/libyang.h>
#include <libyang/tree_data.h>
#include <stdlib.h>
#include <stdio.h>

struct lyd_node* lyd_parse_data_path(struct ly_ctx *ctx,  const char *path, LYD_FORMAT format, int options) {
	return lyd_parse_path(ctx, path, format, options);
}

struct lyd_node *lyd_parse_data_mem(struct ly_ctx *ctx, const char *data, LYD_FORMAT format, int options)
{
	return lyd_parse_mem(ctx, data, format, options);
}

int lyd_data_validate(struct lyd_node **node, int options, struct ly_ctx *ctx)
{
	return lyd_validate(node, options, ctx);
}
*/
import "C"


var ctx *C.struct_ly_ctx
var data *C.struct_lyd_node

//Schema path
const (
	CVL_SCHEMA = "schema/"
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

var cvlInitialized bool
var dbNameToDbNum map[string]uint8
var tmpDbCache map[string]interface{} //map of table storing map of key-value pair
					//m["PORT_TABLE] = {"key" : {"f1": "v1"}}
//Important schema information to be loaded at bootup time
type modelTableInfo struct {
	dbNum uint8
	modelName string
	keys []string
	redisKeyDelim string
	redisKeyPattern string
	mapLeaf []string //for 'mapping  list'
	leafRef map[string][]string //for storing all leafrefs for a leaf in a table, 
				//multiple leafref possible for union 
	mustExp map[string]string
	tablesForMustExp map[string]bool
}

type modelNamespace struct {
	prefix string
	ns string
}

type modelDataInfo struct {
	modelNs map[string]modelNamespace//model namespace 
	tableInfo map[string]modelTableInfo //redis table to model name and keys
}

var redisClient *redis.Client

//Stores important model info
var modelInfo modelDataInfo

type keyValuePairStruct struct {
	key string
	values []string
}

//package init function 
func init() {
	Initialize()
}

var tracing bool = false
func TRACE_LOG(level log.Level, fmtStr string, args ...interface{}) {
	if  tracing == true {
		fmt.Printf(fmtStr, args...)
	} else {
		log.V(level).Infof(fmtStr, args...)
	}
}

func Debug(on bool) {
	if  (on == true) {
		tracing = true
		C.ly_verb(C.LY_LLDBG)
	} else {
		tracing = false
		C.ly_verb(C.LY_LLERR)
	}

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
func storeModelInfo(modelFile string) { //such model info can be maintained in C code and fetched from there 
	f, err := os.Open(CVL_SCHEMA + modelFile)
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

		//Now store the tableInfo in global data
		modelInfo.tableInfo[tableName] = tableInfo

		//Find and store all leafref under each table
		if (listNode == nil) {
			continue
		}

		leafRefNodes := xmlquery.Find(listNode, "//type[@name='leafref']")
		if (leafRefNodes == nil) {
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
func addChildNode(parent *xmlquery.Node, xmlChildNode *xmlquery.Node) {
	xmlChildNode.Parent = parent
	if (parent.FirstChild == nil) {
		parent.FirstChild = xmlChildNode
	} else  {
		parent.LastChild.NextSibling = xmlChildNode
	}
	parent.LastChild = xmlChildNode
}

//Add all other table data for validating all 'must' exp for tableName
func addTableDataForMustExp(tableName string) {
	if (modelInfo.tableInfo[tableName].mustExp == nil) {
		return
	}

	for mustTblName, _ := range modelInfo.tableInfo[tableName].tablesForMustExp {
		tableKeys, err:= redisClient.Keys(mustTblName +
		modelInfo.tableInfo[mustTblName].redisKeyDelim + "*").Result()

		if (err != nil) {
			continue
		}

		for _, tableKey := range tableKeys {
			tableKey = tableKey[len(mustTblName+ modelInfo.tableInfo[mustTblName].redisKeyDelim):] //remove table prefix
			if (tmpDbCache[mustTblName] == nil) {
				tmpDbCache[mustTblName] = map[string]interface{}{tableKey: nil}
			} else {
				tblMap := tmpDbCache[mustTblName]
				tblMap.(map[string]interface{})[tableKey] =nil
				tmpDbCache[mustTblName] = tblMap
			}
		}
	}
}

func addUpdateDataToCache(tableName string, redisKey string) {
	if (tmpDbCache[tableName] == nil) {
		tmpDbCache[tableName] = map[string]interface{}{redisKey: nil}
	} else {
		tblMap := tmpDbCache[tableName]
		tblMap.(map[string]interface{})[redisKey] =nil
		tmpDbCache[tableName] = tblMap
	}
}

func addLeafRef(config bool, tableName string, name string, value string) {

	if (config == false) {
		return 
	}

	//Check if leafRef entry is there for this field
	if (len(modelInfo.tableInfo[tableName].leafRef[name]) > 0) { //array of leafrefs for a leaf
		for _, leafRef  := range modelInfo.tableInfo[tableName].leafRef[name] {

			//Get reference table name from the path and the leaf name
			re := regexp.MustCompile(`.*[/]([a-zA-Z]*:)?(.*)[/]([a-zA-Z]*:)?(.*)`)
			matches := re.FindStringSubmatch(leafRef)

			//We have the leafref table name and the leaf name as well
			if (matches != nil && len(matches) == 5) { //whole + 4 sub matches
				refTableName := matches[2]
				redisKey := value
				//only key is there, value wil be fetched and stored here, 
				//if value can fetched this entry will be deleted that time
				if (tmpDbCache[refTableName] == nil) {
					tmpDbCache[refTableName] = map[string]interface{}{redisKey: nil}
				} else {
					tblMap := tmpDbCache[refTableName]
					_, exist := tblMap.(map[string]interface{})[redisKey]
					if (exist == false) {
						 tblMap.(map[string]interface{})[redisKey] = nil
						//append(tblMap, map[string]interface{}{redisKey: nil})
						tmpDbCache[refTableName] = tblMap
					}
				}
			}
		}
	}
}


func addChildLeaf(config bool, tableName string, parent *xmlquery.Node, name string, value string) *xmlquery.Node{
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

	addLeafRef(config, tableName, name, value)

	return xmlLeafNode
}

func checkFieldMap(fieldMap *map[string]string) map[string]interface{} {
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
func fetchDataToTmpCache() string {
	prevDbNum := INVALID_DB//start with high db number

	var redisClient *redis.Client

	for tableName, dbKeys := range tmpDbCache { //for each table

		if (prevDbNum != modelInfo.tableInfo[tableName].dbNum) {
			if (prevDbNum != INVALID_DB) {
				redisClient.Close() //close previous connection and start a new one
			}
			redisClient = redis.NewClient(&redis.Options{
				Addr:     ":6379",
				Password: "", // no password set
				DB: int(modelInfo.tableInfo[tableName].dbNum),  // use APP DB

			})
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

		mapTable := tmpDbCache[tableName]

		for key, val := range mCmd {
			res, err := val.Result()
			if (err != nil || len(res) == 0) {
				//no data found, don't keep blank entry
				delete(mapTable.(map[string]interface{}), key)
				continue
			}
			/*if (len(res) == 0){ //means no data
				delete(mapTable.(map[string]interface{}), key)
			} else {*/
			//exclude table name and delim
			keyOnly := key//key[len(tableName + modelInfo.tableInfo[tableName].redisKeyDelim):]
			//TBD: Need to check field name like <NULL>(to be deleted)
			//and 'members@' (strip '@')
			//store all field values
			fieldMap := checkFieldMap(&res)
			/*
			for field, value := range res {
				if (field == "NULL") {
					delete(res, field) 
				} else if (field[len(field)-1:] == "@") {
					//last char @ means it is a leaf-list/array of fields
					delete(res, field)
					field = field[:len(field)-1] //strip @
					res[field] = value
				}
			} */
			mapTable.(map[string]interface{})[keyOnly] = fieldMap
			/*} else {
				delete(mapTable.(map[string]interface{}), key)
				//tmpDbCache[tableName] = mapTable
			}*/
		}

		pipe.Close()
	}

	if (prevDbNum != INVALID_DB) {
		redisClient.Close()
	}

	jsonDataBytes, _ := json.Marshal(tmpDbCache)
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
		xmlNode, _= generateTableData(false, jsonNode)

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

func clearTmpDbCache() {
	for key := range tmpDbCache {
		delete(tmpDbCache, key)
	}
}

//This function is not used currently
func addStatusData(tableName string) *xmlquery.Node {
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
		xmlNode, _= generateTableData(false, jsonNode)

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

//Generate Yang xml for all fields in a hash
func generateTableFieldsData(config bool, tableName string, jsonNode *jsonquery.Node,
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
				addChildNode(parent, xmlNode)

				addChildLeaf(config, tableName,
				xmlNode, modelInfo.tableInfo[tableName].mapLeaf[0],
				jsonFieldNode.Data)

				addChildLeaf(config, tableName,
				xmlNode, modelInfo.tableInfo[tableName].mapLeaf[1],
				jsonFieldNode.FirstChild.Data)

			} else {
				//check if it is hash-ref, then need to add only key from "TABLE|k1"
				re := regexp.MustCompile(`\[(.*)\|(.*)\]`)
				hasRefMatch := re.FindStringSubmatch(jsonFieldNode.FirstChild.Data)

				if (hasRefMatch != nil && len(hasRefMatch) == 3) {
					addChildLeaf(config, tableName,
					parent, jsonFieldNode.Data,
					hasRefMatch[2]) //take hashref key value
				} else {
					addChildLeaf(config, tableName,
					parent, jsonFieldNode.Data,
					jsonFieldNode.FirstChild.Data)
				}
			}

		} else if (jsonFieldNode.Type == jsonquery.ElementNode &&
		jsonFieldNode.FirstChild.Type == jsonquery.ElementNode) {
			//Array data e.g. VLAN members
			for  arrayNode:=jsonFieldNode.FirstChild; arrayNode != nil;

			arrayNode = arrayNode.NextSibling {
				addChildLeaf(config, tableName,
				parent, jsonFieldNode.Data,
				arrayNode.FirstChild.Data)
			}
		}
	}

	return CVL_SUCCESS
}

// Generate YANG xml instance for Redis table
func generateTableData(config bool, jsonNode *jsonquery.Node)(*xmlquery.Node, CVLRetCode) {

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
			//fmt.Sprintf("xmlns:%s", modelInfo.tableInfo[tableName].modelName)},
			//modelInfo.modelNs[modelInfo.tableInfo[tableName].modelName].prefix)},

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
				addChildLeaf(config, tableName,
				xmlNode, keyValuePair[idx].key,
				keyValuePair[idx].values[keyIndices[idx]])
			}


			addChildNode(xmlTopNode, xmlNode) //Add the list to the top node

			//Get all fields under the key field and add them as children of the list
			generateTableFieldsData(config, tableName, jsonNode, xmlNode)


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

//Convert Redis JSON to Yang XML using translation metadata
func translateToYang(jsonData string) (*xmlquery.Node, CVLRetCode) {
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
		xmlNode, _= generateTableData(true, jsonNode)

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

//Validate config - syntax and semantics
func validate (xmlData string) CVLRetCode {
	TRACE_LOG(1, "Validating \n%v\n....", xmlData)

	data := C.lyd_parse_data_mem(ctx, C.CString(xmlData), C.LYD_XML, C.LYD_OPT_EDIT)
	if ((C.ly_errno != 0) || (data == nil)) {
		fmt.Println("Parsing data failed\n")
		return CVL_SYNTAX_ERROR
	}

	depData := fetchDataToTmpCache()
	if (depData != "") {
		depDataNode := C.lyd_parse_data_mem(ctx, C.CString(depData), C.LYD_XML, C.LYD_OPT_EDIT)

		if (0 != C.lyd_merge_to_ctx(&data, depDataNode, C.LYD_OPT_DESTRUCT, ctx)) {
			TRACE_LOG(1, "Failed to merge status data\n")
		}
	}

	if (0 != C.lyd_data_validate(&data, C.LYD_OPT_CONFIG, ctx)) {
		fmt.Println("Validation failed\n")
		return CVL_SYNTAX_ERROR
	}

	return CVL_SUCCESS
}

//Perform syntax checks
func validateSyntax(xmlData string) (CVLRetCode, *C.struct_lyd_node) {
	TRACE_LOG(1, "Validating syntax \n%v\n....", xmlData)

	//parsing only does syntacial checks
	data := C.lyd_parse_data_mem(ctx, C.CString(xmlData), C.LYD_XML, C.LYD_OPT_EDIT)
	if ((C.ly_errno != 0) || (data == nil)) {
		fmt.Println("Parsing data failed\n")
		return CVL_SYNTAX_ERROR, nil
	}

	return CVL_SUCCESS, data
}

//Perform semantic checks 
func validateSemantics(data *C.struct_lyd_node, otherDepData string) CVLRetCode {

	//Get dependent data from 
	depData := fetchDataToTmpCache() //fetch data to temp cache for temporary validation
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
		fmt.Println("Validation failed\n")
		return CVL_SEMANTIC_ERROR
	}

	return CVL_SUCCESS
}

//Convert key-data array to map data for json conversion
func keyDataToMap(dataForValidation CVLValidateType, keyData []CVLEditConfigData) map[string]interface{} {
	topData := make(map[string]interface{})

	//For each keyData
	for _, keyDataItem := range keyData {
		if (dataForValidation != keyDataItem.VType) {
			continue
		}

		for tblName,_ := range modelInfo.tableInfo {
			//Check if table prefix matches to any schema table
			//i.e. has 'VLAN|' or 'PORT|' etc.
			if (strings.HasPrefix(keyDataItem.Key, tblName + modelInfo.tableInfo[tblName].redisKeyDelim)) {
				prefixLen := len(tblName) + 1
				if _, existing := topData[tblName]; existing {
					fieldsMap := topData[tblName].(map[string]interface{})
					fieldsMap[keyDataItem.Key[prefixLen:]] = checkFieldMap(&keyDataItem.Data)
				} else {
					fieldsMap := make(map[string]interface{})
					fieldsMap[keyDataItem.Key[prefixLen:]] = checkFieldMap(&keyDataItem.Data)
					topData[tblName] = fieldsMap
				}

				break
			}
		}
	}

	return topData
}

