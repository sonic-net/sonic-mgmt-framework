////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package cvl
import (
	"fmt"
	"os"
	"strings"
	"regexp"
	"time"
	"github.com/go-redis/redis"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/antchfx/jsonquery"
	"cvl/internal/yparser"
	. "cvl/internal/util"
	"sync"
	"flag"
	"io/ioutil"
	"path/filepath"
	custv "cvl/custom_validation"
	"unsafe"
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

const DEFAULT_CACHE_DURATION uint16 = 300 /* 300 sec */
const MAX_BULK_ENTRIES_IN_PIPELINE int = 50
const MAX_DEVICE_METADATA_FETCH_RETRY = 60
const PLATFORM_SCHEMA_PATH = "platform/"

var reLeafRef *regexp.Regexp = nil
var reHashRef *regexp.Regexp = nil
var reSelKeyVal *regexp.Regexp = nil
var reLeafInXpath *regexp.Regexp = nil

var cvlInitialized bool
var dbNameToDbNum map[string]uint8

//map of lua script loaded
var luaScripts map[string]*redis.Script

type tblFieldPair struct {
	tableName string
	field string
}

type mustInfo struct {
	expr string //must expression
	exprTree *xpath.Expr //compiled expression tree
	errCode string //err-app-tag
	errStr string //error message
}

type leafRefInfo struct {
	path string //leafref path
	exprTree *xpath.Expr //compiled expression tree
	yangListNames []string //all yang list in path
	targetNodeName string //target node name
}

type whenInfo struct {
	expr string //when expression
	exprTree *xpath.Expr //compiled expression tree
	nodeNames []string //list of nodes under when condition
	yangListNames []string //all yang list in expression
}

//Important schema information to be loaded at bootup time
type modelTableInfo struct {
	dbNum uint8
	modelName string
	redisTableName string //To which Redis table it belongs to, used for 1 Redis to N Yang List
	module *yparser.YParserModule
	keys []string
	redisKeyDelim string
	redisKeyPattern string
	redisTableSize int
	mapLeaf []string //for 'mapping  list'
	leafRef map[string][]*leafRefInfo //for storing all leafrefs for a leaf in a table, 
				//multiple leafref possible for union 
	mustExpr map[string][]*mustInfo
	whenExpr map[string][]*whenInfo
	tablesForMustExp map[string]CVLOperation
	refFromTables []tblFieldPair //list of table or table/field referring to this table
	custValidation map[string]string // Map for custom validation node and function name
	dfltLeafVal map[string]string //map of leaf names and default value
}


/* CVL Error Structure. */
type CVLErrorInfo struct {
	TableName string      /* Table having error */
	ErrCode  CVLRetCode   /* CVL Error return Code. */
	CVLErrDetails string  /* CVL Error Message details. */
	Keys    []string      /* Keys of the Table having error. */
        Value    string        /* Field Value throwing error */
	Field	 string        /* Field Name throwing error . */
	Msg     string        /* Detailed error message. */
	ConstraintErrMsg  string  /* Constraint error message. */
	ErrAppTag string
}

// Struct for request data and YANG data
type requestCacheType struct {
	reqData CVLEditConfigData
	yangData *xmlquery.Node
}

// Struct for CVL session 
type CVL struct {
	redisClient *redis.Client
	yp *yparser.YParser
	tmpDbCache map[string]interface{} //map of table storing map of key-value pair
	requestCache map[string]map[string][]*requestCacheType//Cache of validated data,
				//per table, per key. Can be used as dependent data in next request
	maxTableElem map[string]int //max element count per table
	batchLeaf string
	chkLeafRefWithOthCache bool
	yv *YValidator //Custom YANG validator for validating external dependencies
}

// Struct for model namepsace and prefix
type modelNamespace struct {
	prefix string
	ns string
}

// Struct for storing all YANG list schema info
type modelDataInfo struct {
	modelNs map[string]*modelNamespace //model namespace 
	tableInfo map[string]*modelTableInfo //redis table to model name and keys
	redisTableToYangList map[string][]string //Redis table to all YANG lists when it is not 1:1 mapping
	allKeyDelims map[string]bool
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
	mutex *sync.Mutex
}

// Struct for storing key and value pair
type keyValuePairStruct struct {
	key string
	values []string
}

var cvg cvlGlobalSessionType

//Single redis client for validation
var redisClient *redis.Client

//Stores important model info
var modelInfo modelDataInfo

func TRACE_LOG(tracelevel CVLTraceLevel, fmtStr string, args ...interface{}) {
	TRACE_LEVEL_LOG(tracelevel, fmtStr, args...)
}

func CVL_LOG(level CVLLogLevel, fmtStr string, args ...interface{}) {
	CVL_LEVEL_LOG(level, fmtStr, args...)
}

//package init function 
func init() {
	if (os.Getenv("CVL_SCHEMA_PATH") != "") {
		CVL_SCHEMA = os.Getenv("CVL_SCHEMA_PATH") + "/"
	}

	if (os.Getenv("CVL_DEBUG") != "") {
		SetTrace(true)
	}

	xpath.SetKeyGetClbk(func(listName string) []string {
		if modelInfo.tableInfo[listName] != nil {
			return modelInfo.tableInfo[listName].keys
		}

		return nil
	})


	ConfigFileSyncHandler()

	cvlCfgMap := ReadConfFile()

	if (cvlCfgMap != nil) {
		flag.Set("v", cvlCfgMap["VERBOSITY"])
		if (strings.Compare(cvlCfgMap["LOGTOSTDERR"], "true") == 0) {
			flag.Set("logtostderr", "true")
			flag.Set("stderrthreshold", cvlCfgMap["STDERRTHRESHOLD"])
		}

		CVL_LOG(INFO ,"Current Values of CVL Configuration File %v", cvlCfgMap)
	}

	//regular expression for leafref and hashref finding
	reLeafRef = regexp.MustCompile(`.*[/]([-_a-zA-Z]*:)?(.*)[/]([-_a-zA-Z]*:)?(.*)`)
	reHashRef = regexp.MustCompile(`\[(.*)\|(.*)\]`)
	//Regular expression to select key value
	reSelKeyVal = regexp.MustCompile("=[ ]*['\"]?([0-9_a-zA-Z]+)['\"]?|(current[(][)])")
	//Regular expression to find leafref in xpath
	reLeafInXpath = regexp.MustCompile("(.*[:/]{1})([a-zA-Z0-9_-]+)([^a-zA-Z0-9_-]*)")

	if Initialize() != CVL_SUCCESS {
		CVL_LOG(FATAL, "CVL initialization failed")
	}

	cvg.db = make(map[string]dbCachedData)

	//Global session keeps the global cache
	cvg.cv, _ = ValidationSessOpen()
	//Create buffer channel of length 1
	cvg.stopChan = make(chan int, 1)
	//Initialize mutex
	cvg.mutex = &sync.Mutex{}

	_, err := redisClient.ConfigSet("notify-keyspace-events", "AKE").Result()
	if err != nil {
		CVL_LOG(ERROR ,"Could not enable notification error %s", err)
	}

	//Store PORT table details in global cache and update the cache if PORT entry changes
	//dbCacheSet(false, "PORT", 0)

	xpath.SetLogCallback(func(fmt string, args ...interface{}) {
		if (IsTraceLevelSet(TRACE_SEMANTIC) == false) {
			return
		}

		TRACE_LOG(TRACE_SEMANTIC, "XPATH: " + fmt, args...)
	})
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

// Load all YIN schema files, apply deviation files 
func loadSchemaFiles() CVLRetCode {

	platformName := ""
	// Wait to check if CONFIG_DB is populated with DEVICE_METADATA.
	// This is needed to apply deviation file
	retryCnt := 0
	for ; (retryCnt < MAX_DEVICE_METADATA_FETCH_RETRY); retryCnt++ {
		deviceMetaDataKey, err := redisClient.Keys("DEVICE_METADATA|localhost").Result()
		if (err != nil) || (len(deviceMetaDataKey) == 0)  {
			//Retry for 1 min
			time.Sleep(100 * time.Millisecond) //sleep for 1 sec and then retry
			continue
		}

		//Redis is populated with DEVICE_METADATA
		break
	}

	//Now try to fetch the platform details
	if (retryCnt < MAX_DEVICE_METADATA_FETCH_RETRY) {
		deviceMetaData, err := redisClient.HGetAll("DEVICE_METADATA|localhost").Result()
		var exists bool
		platformName, exists = deviceMetaData["platform"]
		if (err != nil) || (exists == false) || (platformName == "") {
			CVL_LOG(WARNING, "Could not fetch 'platform' details from CONFIG_DB")
		}
	}

	//Scan schema directory to get all schema files
	modelFiles, err := filepath.Glob(CVL_SCHEMA + "/*.yin")
	if err != nil {
		CVL_LOG(FATAL ,"Could not read schema files %v", err)
	}

	moduleMap := map[string]*yparser.YParserModule{}
	// Load all common schema files
	for _, modelFilePath := range modelFiles {
		_, modelFile := filepath.Split(modelFilePath)

		TRACE_LOG(TRACE_LIBYANG, "Parsing schema file %s ...",
		modelFilePath)

		// Now parse each schema file 
		var module *yparser.YParserModule
		if module, _ = yparser.ParseSchemaFile(modelFilePath); module == nil {

			CVL_LOG(FATAL,fmt.Sprintf("Unable to parse schema file %s", modelFile))
			return CVL_ERROR
		}

		moduleMap[modelFile] = module
	}

	// Load all platform specific schema files based on platform details 
	// present in DEVICE_METADATA
	for {
		if (platformName == "") {
			CVL_LOG(INFO, "Skipping parsing of any platform specific YIN schema " +
			"files as platform name can't be determined")
			break
		}

		// Read directory under 'platform' directory
		allDirs, errDir := ioutil.ReadDir(CVL_SCHEMA + "/" + PLATFORM_SCHEMA_PATH)
		if (errDir != nil) || (len(allDirs) == 0) {
			CVL_LOG(INFO, "Could not read platform schema location or no platform " +
			"specific schema exists. %v", err)
			break
		}

		// For matched platform directory parse all schema files
		for _, sDir := range allDirs {

			sDirName := sDir.Name()
			//Check which directory matches
			if (strings.Contains(platformName, sDirName) == false) {
				continue
			}

			//Get all platform specific YIN schema file names
			modelFiles, err := filepath.Glob(CVL_SCHEMA + "/" +
			PLATFORM_SCHEMA_PATH + "/" + sDirName + "/*.yin")
			if err != nil {
				CVL_LOG(WARNING,"Could not read platform schema directory %v", err)
				break
			}


			//Now parse platform schema files
			for _, modelFilePath := range modelFiles {
				_, modelFile := filepath.Split(modelFilePath)

				TRACE_LOG(TRACE_YPARSER, "Parsing platform specific schema" + 				"file %s ...\n", modelFilePath)

				var module *yparser.YParserModule
				if module, _ = yparser.ParseSchemaFile(modelFilePath); module == nil {

					CVL_LOG(ERROR, "Unable to parse schema file %s", modelFile)
					return CVL_ERROR
				}

				moduleMap[modelFile] = module
			}

			//platform found
			break
		}

		break
	}

	for modelFile, parsedModule := range moduleMap {
		//store schema related info to use in validation
		storeModelInfo(modelFile, parsedModule)
	}

	return CVL_SUCCESS
}

//Get list of YANG list names used in xpath expression
func getYangListNamesInExpr(expr string) []string {
	tbl := []string{}

	//Check with all table names
	for tblName, _ := range modelInfo.tableInfo {

		//Match 1 - Prefix is used in path
		//Match 2 - Prefix is not used in path, it is in same YANG model
		if (strings.Contains(expr, ":" + tblName + "_LIST") == true) ||
		(strings.Contains(expr, "/" + tblName + "_LIST") == true) {
			tbl = append(tbl, tblName)
		}
	}

	return tbl
}

//Get all YANG lists referred and the target node for leafref
//Ex: leafref { path "../../../ACL_TABLE/ACL_TABLE_LIST[aclname=current()]/aclname";}
//will return [ACL_TABLE] and aclname
func getLeafRefTargetInfo(path string) ([]string, string) {
	target := ""

	//Get list of all YANG list used in the path
	tbl := getYangListNamesInExpr(path)

	//Get the target node name from end of the path
	idx := strings.LastIndex(path, ":") //check with prefix first
	if idx > 0 {
		target = path[idx+1:]
	} else if idx = strings.LastIndex(path, "/"); idx > 0{ //no prefix there
		target = path[idx+1:]
	}

	return tbl, target
}

func storeModelInfo(modelFile string, module *yparser.YParserModule) { //such model info can be maintained in C code and fetched from there 

	//model is derived from file name
	tokens := strings.Split(modelFile, ".")
	modelName := tokens[0]

	//Store namespace and prefix
	ns, prefix := yparser.GetModelNs(module)
	modelInfo.modelNs[modelName] = &modelNamespace{ns:ns, prefix:prefix}

	list := yparser.GetModelListInfo(module)

	if (list == nil) {
		CVL_LOG(ERROR, "Unable to get schema details for %s", modelFile)
		return
	}


	for _, lInfo := range list {
		TRACE_LOG(TRACE_YPARSER,
		"Storing schema details for list %s", lInfo.ListName)

		tInfo := modelTableInfo{modelName: modelName}

		tInfo.dbNum = dbNameToDbNum[lInfo.DbName]
		tInfo.redisTableName = lInfo.RedisTableName
		tInfo.module = module
		tInfo.redisKeyDelim = lInfo.RedisKeyDelim
		tInfo.redisKeyPattern = lInfo.RedisKeyPattern
		tInfo.redisTableSize = lInfo.RedisTableSize
		tInfo.keys = lInfo.Keys
		tInfo.mapLeaf = lInfo.MapLeaf
		tInfo.custValidation = lInfo.CustValidation

		//store default values used in must and when exp
		tInfo.dfltLeafVal = make(map[string]string, len(lInfo.DfltLeafVal))
		for nodeName, val := range lInfo.DfltLeafVal {
			tInfo.dfltLeafVal[nodeName] = val
		}

		//tInfo.leafRef = lInfo.LeafRef
		tInfo.leafRef =  make(map[string][]*leafRefInfo, len(lInfo.LeafRef))
		for nodeName, lpathArr := range lInfo.LeafRef { //for each leaf or leaf-list
			leafRefInfoArr := []*leafRefInfo{}
			for _, lpath := range lpathArr {
				//just store the leafref path
				leafRefData := leafRefInfo{path: lpath}
				leafRefInfoArr = append(leafRefInfoArr, &leafRefData)
			}
			tInfo.leafRef[nodeName] = leafRefInfoArr
		}

		tInfo.mustExpr = make(map[string][]*mustInfo, len(lInfo.XpathExpr))
		for nodeName, xprArr := range lInfo.XpathExpr {
			for _, xpr := range xprArr {
				tInfo.mustExpr[nodeName] = append(tInfo.mustExpr[nodeName],
				&mustInfo{
					expr: xpr.Expr,
					errCode: xpr.ErrCode,
					errStr: xpr.ErrStr,
				})
			}
		}

		tInfo.whenExpr = make(map[string][]*whenInfo, len(lInfo.WhenExpr))
		for nodeName, whenExprArr := range lInfo.WhenExpr {
			for _, whenExpr := range whenExprArr {
				tInfo.whenExpr[nodeName] = append(tInfo.whenExpr[nodeName],
					&whenInfo {
					expr: whenExpr.Expr,
					nodeNames: whenExpr.NodeNames,
				})
			}
		}

		modelInfo.allKeyDelims[tInfo.redisKeyDelim] = true
		yangList := modelInfo.redisTableToYangList[tInfo.redisTableName]
		yangList = append(yangList, lInfo.ListName)
		//Update the map
		modelInfo.redisTableToYangList[tInfo.redisTableName] = yangList

		modelInfo.tableInfo[lInfo.ListName] = &tInfo
	}
}

// Get YANG list to Redis table name
func getYangListToRedisTbl(yangListName string) string {
	if (strings.HasSuffix(yangListName, "_LIST")) {
		yangListName = yangListName[0:len(yangListName) - len("_LIST")]
	}
	tInfo, exists := modelInfo.tableInfo[yangListName]

	if (exists == true) && (tInfo.redisTableName != "") {
		return tInfo.redisTableName
	}

	return yangListName
}

//This functions build info of dependent table/fields 
//which uses a particular table through leafref
func buildRefTableInfo() {

	CVL_LOG(INFO_API, "Building reverse reference info from leafref")

	for tblName, tblInfo := range modelInfo.tableInfo {
		if (len(tblInfo.leafRef) == 0) {
			continue
		}

		//For each leafref update the table used through leafref
		for fieldName, leafRefs  := range tblInfo.leafRef {
			for _, leafRef := range leafRefs {
				/*
				matches := reLeafRef.FindStringSubmatch(leafRef)

				//We have the leafref table name
				if (matches != nil && len(matches) == 5) { //whole + 4 sub matches
					refTable := getYangListToRedisTbl(matches[2])
					refTblInfo :=  modelInfo.tableInfo[refTable]

					refFromTables := &refTblInfo.refFromTables
					*refFromTables = append(*refFromTables, tblFieldPair{tblName, fieldName})
					modelInfo.tableInfo[refTable] = refTblInfo
				} */

				for _, yangListName := range leafRef.yangListNames {
					refTblInfo :=  modelInfo.tableInfo[yangListName]

					refFromTables := &refTblInfo.refFromTables
					 *refFromTables = append(*refFromTables, tblFieldPair{tblName, fieldName})
					 modelInfo.tableInfo[yangListName] = refTblInfo
				}

			}
		}

	}

	//Now sort list 'refFromTables' under each table based on dependency among them 
	for tblName, tblInfo := range modelInfo.tableInfo {
		if (len(tblInfo.refFromTables) == 0) {
			continue
		}

		depTableList := []string{}
		for i:=0; i < len(tblInfo.refFromTables); i++ {
			depTableList = append(depTableList, tblInfo.refFromTables[i].tableName)
		}

		sortedTableList, _ := cvg.cv.SortDepTables(depTableList)
		if (len(sortedTableList) == 0) {
			continue
		}

		newRefFromTables := []tblFieldPair{}

		for i:=0; i < len(sortedTableList); i++ {
			//Find fieldName
			fieldName := ""
			for j :=0; j < len(tblInfo.refFromTables); j++ {
				if (sortedTableList[i] == tblInfo.refFromTables[j].tableName) {
					fieldName =  tblInfo.refFromTables[j].field
					break
				}
			}
			newRefFromTables = append(newRefFromTables, tblFieldPair{sortedTableList[i], fieldName})
		}
		//Update sorted refFromTables
		tblInfo.refFromTables = newRefFromTables
		modelInfo.tableInfo[tblName] = tblInfo 
	}

}

//Find the tables names in must expression, these tables data need to be fetched 
//during semantic validation
func addTableNamesForMustExp() {

	for tblName, tblInfo := range  modelInfo.tableInfo {
		if (len(tblInfo.mustExpr) == 0) {
			continue
		}

		tblInfo.tablesForMustExp = make(map[string]CVLOperation)

		for _, mustExpArr := range tblInfo.mustExpr {
			for _, mustExp := range mustExpArr {
				var op CVLOperation = OP_NONE
				//Check if 'must' expression should be executed for a particular operation
				if (strings.Contains(mustExp.expr,
				":operation != 'CREATE'") == true) {
					op = op | OP_CREATE
				} else if (strings.Contains(mustExp.expr,
				":operation != 'UPDATE'") == true) {
					op = op | OP_UPDATE
				} else if (strings.Contains(mustExp.expr,
				":operation != 'DELETE'") == true) {
					op = op | OP_DELETE
				}

				//store the current table if aggregate function like count() is used
				/*if (strings.Contains(mustExp, "count") == true) {
					tblInfo.tablesForMustExp[tblName] = op
				}*/

				//check which table name is present in the must expression
				for tblNameSrch, _ := range modelInfo.tableInfo {
					if (tblNameSrch == tblName) {
						continue
					}
					//Table name should appear like "../VLAN_MEMBER_LIST/tagging_mode' or '
					// "/prt:PORT/prt:ifname"
					re := regexp.MustCompile(fmt.Sprintf(".*[/]([-_a-zA-Z]*:)?%s_LIST[\\[/]", tblNameSrch))
					matches := re.FindStringSubmatch(mustExp.expr)
					if (len(matches) > 0) {
						//stores the table name 
						tblInfo.tablesForMustExp[tblNameSrch] = op
					}
				}
			}
		}

		//update map
		modelInfo.tableInfo[tblName] = tblInfo
	}
}

//Split key into table prefix and key
func splitRedisKey(key string) (string, string) {

	var foundIdx int = -1
	//Check with all key delim
	for keyDelim, _ := range modelInfo.allKeyDelims {
		foundIdx = strings.Index(key, keyDelim)
		if (foundIdx >= 0) {
			//Matched with key delim
			break
		}
	}

	if (foundIdx < 0) {
		//No matches
		CVL_LOG(ERROR, "Could not find any of key delimeter %v in key '%s'",
		modelInfo.allKeyDelims, key)
		return "", ""
	}

	tblName := key[:foundIdx]

	if _, exists := modelInfo.tableInfo[tblName]; exists == false {
		//Wrong table name
		CVL_LOG(ERROR, "Could not find table '%s' in schema", tblName)
		return "", ""
	}

	prefixLen := foundIdx + 1


	TRACE_LOG(TRACE_SYNTAX, "Split Redis Key %s into (%s, %s)",
	key, tblName, key[prefixLen:])

	return tblName, key[prefixLen:]
}

//Get the YANG list name from Redis key and table name
//This just returns same YANG list name as Redis table name
//when 1:1 mapping is there. For one Redis table to 
//multiple YANG list, it returns appropriate YANG list name
//INTERFACE:Ethernet12 returns ==> INTERFACE
//INTERFACE:Ethernet12:1.1.1.0/32 ==> INTERFACE_IPADDR
func getRedisTblToYangList(tableName, key string) (yangList string) {
	defer func() {
		pYangList := &yangList
		TRACE_LOG(TRACE_SYNTAX, "Got YANG list '%s' " +
		"from Redis Table '%s', Key '%s'", *pYangList, tableName, key)
	}()

	mapArr, exists := modelInfo.redisTableToYangList[tableName]

	if (exists == false) || (len(mapArr) == 1) { //no map or only one
		//1:1 mapping case
		return tableName
	}

	//As of now determine the mapping based on number of keys
	var foundIdx int = -1
	numOfKeys := 1 //Assume only one key initially
	for keyDelim, _ := range modelInfo.allKeyDelims {
		foundIdx = strings.Index(key, keyDelim)
		if (foundIdx >= 0) {
			//Matched with key delim
			keyComps := strings.Split(key, keyDelim)
			numOfKeys = len(keyComps)
			break
		}
	}

	//Check which list has number of keys as 'numOfKeys' 
	for i := 0; i < len(mapArr); i++ {
		tblInfo, exists := modelInfo.tableInfo[mapArr[i]]
		if exists == true {
			if (len(tblInfo.keys) == numOfKeys) {
				//Found the YANG list matching the number of keys
				return mapArr[i]
			}
		}
	}

	//No matches
	return tableName
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

	/* TBD. Workaround for optional keys in INTERFACE Table.
	   Code will be removed once model is finalized. */
	if  ((tableName == "INTERFACE") && (len(keyNames) != len(keyVals))) {
		keyVals = append(keyVals, "0.0.0.0/0")

	} else if (len(keyNames) != len(keyVals)) {
		return nil //number key names and values does not match
	}

	mkeys := []keyValuePairStruct{}
	//For each key check the pattern and store key/value pair accordingly
	for  idx, keyName := range keyNames {

		//check if key-pattern contains specific key pattern
		if (keyPatterns[idx+1] == ("{" + keyName + "}")) { //fmt.Sprintf("{%s}", keyName)) { //no specific key pattern - just "{key}"
			//Store key/value mapping     
			mkeys = append(mkeys, keyValuePairStruct{keyName,  []string{keyVals[idx]}})
		} else if (keyPatterns[idx+1] == ("({" + keyName + "},)*")) { //fmt.Sprintf("({%s},)*", keyName)) {   // key pattern is "({key},)*" i.e. repeating keys seperated by ','   
			repeatedKeys := strings.Split(keyVals[idx], ",")
			mkeys = append(mkeys, keyValuePairStruct{keyName, repeatedKeys})
		}
	}

	TRACE_LOG(TRACE_SYNTAX, "getRedisToYangKeys() returns %v " +
	"from Redis Table '%s', Key '%s'", mkeys, tableName, redisKey)

	return mkeys
}

//Check for path resolution. Find Redis entry to be fetched from the path.
func (c *CVL) checkPathForTableEntry(tableName string, currentValue string, cfgData *CVLEditConfigData, mustExpStk []string, token string) ([]string, string, CVLRetCode) {

	n := len(mustExpStk) - 1
	xpath := ""

	if (token  == ")") {
		for n = len(mustExpStk) - 1; mustExpStk[n] != "(";  n  = len(mustExpStk) - 1 {
			//pop until "("
			xpath = mustExpStk[n] + xpath
			mustExpStk = mustExpStk[:n]
		}
	} else if (token == "]") {
		//pop until "["
		for n = len(mustExpStk) - 1; mustExpStk[n] != "[";  n  = len(mustExpStk) - 1 {
			xpath = mustExpStk[n] + xpath
			mustExpStk = mustExpStk[:n]
		}
	}

	mustExpStk = mustExpStk[:n]
	targetTbl := ""
	//Search the table name in xpath
	for tblNameSrch, _ := range modelInfo.tableInfo {
		if (tblNameSrch == tableName) {
			continue
		}
		//Table name should appear like "../VLAN_MEMBER/tagging_mode' or '
		// "/prt:PORT/prt:ifname"
		//re := regexp.MustCompile(fmt.Sprintf(".*[/]([a-zA-Z]*:)?%s[\\[/]", tblNameSrch))
		tblSrchIdx := strings.Index(xpath, fmt.Sprintf("/%s_LIST", tblNameSrch)) //no preifx
		if (tblSrchIdx < 0) {
			tblSrchIdx = strings.Index(xpath, fmt.Sprintf(":%s_LIST", tblNameSrch)) //with prefix
		}
		if (tblSrchIdx < 0) {
			continue
		}

		tblSrchIdxEnd := strings.Index(xpath[tblSrchIdx+len(tblNameSrch)+1:], "[")
		if (tblSrchIdxEnd < 0) {
			tblSrchIdxEnd = strings.Index(xpath[tblSrchIdx+len(tblNameSrch)+1:], "/")
		}

		if (tblSrchIdxEnd >= 0) { //match found
			targetTbl = tblNameSrch
			break
		}
	}

	//No match with table found, could be just keys like 'aclname='TestACL1'
	//just return the same
	if (targetTbl == "") {
		return mustExpStk, xpath, CVL_SUCCESS
	}

	tableKey := targetTbl
	//Add the keys
	keyNames := modelInfo.tableInfo[tableKey].keys

	//Form the Redis Key to fetch the entry
	for  idx, keyName := range keyNames {
		//Key value is string/numeric literal, extract the same
		keySrchIdx := strings.Index(xpath, keyName)
		if (keySrchIdx < 0 ) {
			continue
		}

		matches := reSelKeyVal.FindStringSubmatch(xpath[keySrchIdx+len(keyName):])
		if (len(matches) > 1) {
			if (matches[1] == "current()") {
				//replace with current field value
				tableKey = tableKey + "*" + modelInfo.tableInfo[tableName].redisKeyDelim + currentValue
			} else {
				//Use literal
				tableKey = tableKey + "*" + modelInfo.tableInfo[tableName].redisKeyDelim + matches[1]
			}

			if  (idx != len(keyNames) - 1) {
				tableKey = tableKey + "|*"
			}
		}
	}

	//Fetch the entries
	redisTableKeys, err:= redisClient.Keys(tableKey).Result()

	if (err !=nil || len (redisTableKeys) > 1) { //more than one entry is returned, can't proceed further 
		//Just add all the entries for caching
		for _, redisTableKey := range redisTableKeys {
			c.addTableEntryToCache(splitRedisKey(redisTableKey))
		}

		return mustExpStk, "", CVL_SUCCESS
	}

	for _, redisTableKey := range redisTableKeys {

		var entry map[string]string

		if tmpEntry, mergeNeeded := c.fetchDataFromRequestCache(splitRedisKey(redisTableKey)); (tmpEntry == nil || mergeNeeded == true)  {
			//If data is not available in validated cache fetch from Redis DB
			entry, err = redisClient.HGetAll(redisTableKey).Result()

			if (mergeNeeded) {
				mergeMap(entry, tmpEntry)
			}
		}

		//Get the entry fields from Redis
		if (entry != nil) {
			//Just add all the entries for caching
			c.addTableEntryToCache(splitRedisKey(redisTableKey))

			leafInPath := ""
			index := strings.LastIndex(xpath, "/")
			if (index >= 0) {

				matches := reLeafInXpath.FindStringSubmatch(xpath[index:])
				if (len(matches) > 2) { //should return atleasts two subgroup and entire match
					leafInPath = matches[2]
				} else {
					//No leaf requested in xpath selection
					return mustExpStk, "", CVL_SUCCESS
				}

				index = strings.Index(xpath, "=")
				tblIndex := strings.Index(xpath, targetTbl)
				if (index >= 0 && tblIndex >=0) {

					if leafVal, existing := entry[leafInPath + "@"]; existing == true {
						//Get the field value referred in the xpath
						if (index  < tblIndex) {
							// case like - [ifname=../../ACL_TABLE[aclname=current()]
							return mustExpStk, xpath[:index+1] + leafVal, CVL_SUCCESS
						} else {
							// case like - [ifname=current()]
							return mustExpStk, leafVal, CVL_SUCCESS
						}
					} else {

						if (index  < tblIndex) {
							return mustExpStk, xpath[:index+1] + entry[leafInPath], CVL_SUCCESS
						} else {
							return mustExpStk, entry[leafInPath], CVL_SUCCESS
						}
					}
				}
			}
		}
	}

	return mustExpStk, "", CVL_FAILURE
}

//Add specific entries by looking at must expression
//Must expression may need single or multiple entries
//It can be within same table or across multiple tables
//Node-set function such count() can be quite expensive and 
//should be avoided through this function
func (c *CVL) addTableEntryForMustExp(cfgData *CVLEditConfigData, tableName string) CVLRetCode {
	if (modelInfo.tableInfo[tableName].mustExpr == nil) {
		return CVL_SUCCESS
	}

	for fieldName, mustExpArr := range modelInfo.tableInfo[tableName].mustExpr {
		mustExp := mustExpArr[0]

		currentValue := "" // Current value for current() function

		//Get the current() field value from the entry being created/updated/deleted
		keyValuePair := getRedisToYangKeys(tableName, cfgData.Key[len(tableName)+1:])

		//Try to get the current() from the 'key' provided
		if (keyValuePair != nil) {
			for _, keyValItem := range keyValuePair {
				if (keyValItem.key == fieldName) {
					currentValue = keyValItem.values[0]
				}
			}
		}

		TRACE_LOG(TRACE_CACHE, "Attempting adding data for must expression %s",
		mustExp.expr)

		//current() value needs to be fetched from other field
		if (currentValue == "") {
			if (cfgData.VOp == OP_CREATE) {
				if (tableName != fieldName) { //must expression is not at list level
					currentValue = cfgData.Data[fieldName]
					if (currentValue == "") {
					currentValue = cfgData.Data[fieldName + "@"]
					}
				}
			} else if (cfgData.VOp == OP_UPDATE || cfgData.VOp == OP_DELETE) {
				//fetch the entry to get current() value
				c.clearTmpDbCache()
				entryKey := cfgData.Key[len(tableName)+1:]
				c.tmpDbCache[tableName] = map[string]interface{}{entryKey: nil}

				if (c.fetchTableDataToTmpCache(tableName,
				map[string]interface{}{entryKey: nil}) > 0) {
					mapTable := c.tmpDbCache[tableName]
					if  fields, existing := mapTable.(map[string]interface{})[entryKey]; existing == true {
						currentValue = fmt.Sprintf("%v", fields.(map[string]interface{})[fieldName])
					}
				}
			}
		}

		mustExpStk := []string{} //Use the string slice as stack
		mustExpStr := "(" + mustExp.expr + ")"
		strLen :=  len(mustExpStr)
		strTmp := ""
		//Parse the xpath expression and fetch Redis entry by looking at xpath,
		// any xpath function call is ignored except current().
		for i := 0; i < strLen; i++ {
			switch mustExpStr[i] {
			case '(':
				if (mustExpStr[i+1] == ')') {
					strTmp = strTmp + "()"
					if index := strings.Index(strTmp, "current()"); index >= 0 {
						strTmp = strTmp[:index] + currentValue
					}
					i = i + 1
					continue
				}
				if (strTmp != "") {
					mustExpStk = append(mustExpStk, strTmp)
				}
				mustExpStk = append(mustExpStk, "(")
				strTmp = ""
			case ')':
				if (strTmp != "") {
					mustExpStk = append(mustExpStk, strTmp)
				}
				strTmp = ""
				//Check Path - pop until ')'
				mustExpStk, evalPath,_ := c.checkPathForTableEntry(tableName, currentValue, cfgData,
				mustExpStk, ")")
				if (evalPath != "") {
					mustExpStk = append(mustExpStk, evalPath)
				}
				mustExpStk = append(mustExpStk, ")")
			case '[':
				if (strTmp != "") {
					mustExpStk = append(mustExpStk, strTmp)
				}
				mustExpStk = append(mustExpStk, "[")
				strTmp = ""
			case ']':
				if (strTmp != "") {
					mustExpStk = append(mustExpStk, strTmp)
				}
				//Check Path - pop until = or '['
				mustExpStk, evalPath,_ := c.checkPathForTableEntry(tableName, currentValue, cfgData,
				mustExpStk, "]")
				if (evalPath != "") {
					mustExpStk = append(mustExpStk, "[" + evalPath + "]")
				}
				strTmp = ""
			default:
				strTmp = fmt.Sprintf("%s%c", strTmp, mustExpStr[i])
			}
		}

		//Get the redis data for accumulated keys and add them to session cache
		depData := c.fetchDataToTmpCache() //fetch data to temp cache for temporary validation

		if (depData != nil) {
			if (Tracing == true) {
				TRACE_LOG(TRACE_CACHE, "Adding entries for 'must' expression : %s", c.yp.NodeDump(depData))
			}
		} else {
			CVL_LOG(WARNING, "Could not fetch entries for 'must' expression : %s", mustExp.expr)
			//Could not fetch any entry from Redis after xpath evaluation
			return CVL_FAILURE
		}

		if errObj := c.yp.CacheSubtree(false, depData); errObj.ErrCode != yparser.YP_SUCCESS {
			CVL_LOG(WARNING, "Unable to add dependent data to session cache " +
			" for 'must' expression : %s", mustExp.expr)
			return CVL_FAILURE
		}

	} //for each must expression

	return CVL_SUCCESS
}

//Add all other table data for validating all 'must' exp for tableName
func (c *CVL) addTableDataForMustExp(op CVLOperation, tableName string) CVLRetCode {
	if (len(modelInfo.tableInfo[tableName].mustExpr) == 0) {
		return CVL_SUCCESS
	}

	for mustTblName, mustOp := range modelInfo.tableInfo[tableName].tablesForMustExp {
		//First check if must expression should be executed for the given operation
		if (mustOp != OP_NONE) && ((mustOp & op) == OP_NONE) {
			//must to be excuted for particular operation, but current operation 
			//is not the same one
			continue
		}

		//Check in global cache first and merge to session cache
		if topNode, _ := dbCacheGet(mustTblName); topNode != nil {
			var errObj yparser.YParserError
			//If global cache has the table, add to the session validation
			TRACE_LOG(TRACE_CACHE, "Adding global cache to session cache for table %s", tableName)
			if errObj = c.yp.CacheSubtree(true, topNode); errObj.ErrCode != yparser.YP_SUCCESS {
				return CVL_SYNTAX_ERROR
			}
		} else { //Put the must table in global table and add to session cache
			cvg.cv.chkLeafRefWithOthCache = true
			dbCacheSet(false, mustTblName, 100*DEFAULT_CACHE_DURATION) //Keep the cache for default duration
			cvg.cv.chkLeafRefWithOthCache = false

			if topNode, ret := dbCacheGet(mustTblName); topNode != nil {
				var errObj yparser.YParserError
				//If global cache has the table, add to the session validation
				TRACE_LOG(TRACE_CACHE, "Global cache created, add the data to session cache for table %s", tableName)
				if errObj = c.yp.CacheSubtree(true, topNode); errObj.ErrCode != yparser.YP_SUCCESS {
					CVL_LOG(WARNING, "Unable to add dependent data from " + 
					"global to session cache for table %s", mustTblName)
					return CVL_SYNTAX_ERROR
				}
			} else if (ret == CVL_SUCCESS) {
				TRACE_LOG(TRACE_CACHE, "Global cache empty, no data in Redis for table %s", tableName)
				return CVL_SUCCESS
			} else {
				CVL_LOG(ERROR ,"Could not create global cache for table %s", mustTblName)
				return CVL_ERROR
			}


			/*
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
			*/
		}
	}

	return CVL_SUCCESS
}

//Add leafref entry for caching
//It has to be recursive in nature, as there can be chained leafref
/*
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

				//Check if leafref dependency can also be met from 'must' table
				if (c.chkLeafRefWithOthCache == true) {
					found := false
					for mustTbl, _ := range modelInfo.tableInfo[tableName].tablesForMustExp {
						if mustTbl == refTableName {
							found = true
							break
						}
					}
					if (found == true) {
						//Leafref data will be available from must table dep data, skip this leafref entry
						continue
					}
				}

				//only key is there, value wil be fetched and stored here, 
				//if value can't fetched this entry will be deleted that time
				//Strip "_LIST" suffix
				refRedisTableName := refTableName[0:len(refTableName) - len("_LIST")]
				if (c.tmpDbCache[refRedisTableName] == nil) {
					c.tmpDbCache[refRedisTableName] = map[string]interface{}{redisKey: nil}
				} else {
					tblMap := c.tmpDbCache[refRedisTableName]
					_, exist := tblMap.(map[string]interface{})[redisKey]
					if (exist == false) {
						tblMap.(map[string]interface{})[redisKey] = nil
						c.tmpDbCache[refRedisTableName] = tblMap
					}
				}
			}
		}
	}
}
*/

//Checks field map values and removes "NULL" entry, create array for leaf-list
func (c *CVL) checkFieldMap(fieldMap *map[string]string) map[string]interface{} {
	fieldMapNew := map[string]interface{}{}

	for field, value := range *fieldMap {
		if (field == "NULL") {
			continue
		} else if (field[len(field)-1:] == "@") {
			//last char @ means it is a leaf-list/array of fields
			field = field[:len(field)-1] //strip @
			//split the values seprated using ','
			strArr := strings.Split(value, ",")
			//fieldMapNew[field] = strings.Split(value, ",")
			arrMap := make([]interface{}, 0)//len(strArr))
			for _, strArrItem := range strArr {
				arrMap = append(arrMap, strArrItem)
			}
			fieldMapNew[field] = arrMap//make([]interface{}, len(strArr))
		} else {
			fieldMapNew[field] = value
		}
	}

	return fieldMapNew
}

//Merge 'src' map to 'dest' map of map[string]string type
func mergeMap(dest map[string]string, src map[string]string) {
	TRACE_LOG(TRACE_SEMANTIC,
	"Merging map %v into %v", src, dest)

	for key, data := range src {
		dest[key] = data
	}
}

func (c *CVL) translateToYang(jsonMap *map[string]interface{}) (*yparser.YParserNode, CVLErrorInfo) {

	var  cvlErrObj CVLErrorInfo
	//Parse the map data to json tree
	data, _ := jsonquery.ParseJsonMap(jsonMap)
	var root *yparser.YParserNode
	root = nil
	var errObj yparser.YParserError

	for jsonNode := data.FirstChild; jsonNode != nil; jsonNode=jsonNode.NextSibling {
		TRACE_LOG(TRACE_LIBYANG, "Translating, Top Node=%v\n", jsonNode.Data)
		//Visit each top level list in a loop for creating table data
		topNode, cvlErrObj  := c.generateTableData(true, jsonNode)

		//Generate YANG data for Yang Validator
		topYangNode, cvlYErrObj := c.generateYangListData(jsonNode, true)

		if  topNode == nil {
			cvlErrObj.ErrCode = CVL_SYNTAX_ERROR
			CVL_LOG(ERROR, "Unable to translate request data to YANG format")
			return nil, cvlErrObj
		}

		if  topYangNode == nil {
			cvlYErrObj.ErrCode = CVL_SYNTAX_ERROR
			CVL_LOG(ERROR, "Unable to translate request data to YANG format")
			return nil, cvlYErrObj
		}

		if (root == nil) {
			root = topNode
		} else {
			if root, errObj = c.yp.MergeSubtree(root, topNode); errObj.ErrCode != yparser.YP_SUCCESS {
				CVL_LOG(ERROR, "Unable to merge translated YANG data(libyang) " +
				"while translating from request data to YANG format")
				return nil, cvlErrObj
			}
		}

		//Create a full document and merge with main YANG data
		doc := &xmlquery.Node{Type: xmlquery.DocumentNode}
		doc.FirstChild = topYangNode
		doc.LastChild = topYangNode
		topYangNode.Parent = doc

		if (IsTraceLevelSet(TRACE_CACHE)) {
			TRACE_LOG(TRACE_CACHE, "Before merge, YANG data tree = %s, source = %s",
			c.yv.root.OutputXML(false),
			doc.OutputXML(false))
		}

		if c.mergeYangData(c.yv.root, doc) != CVL_SUCCESS {
			CVL_LOG(ERROR, "Unable to merge translated YANG data while " +
			"translating from request data to YANG format")
			cvlYErrObj.ErrCode = CVL_SYNTAX_ERROR
			return nil, cvlErrObj
		}
		if (IsTraceLevelSet(TRACE_CACHE)) {
			TRACE_LOG(TRACE_CACHE, "After merge, YANG data tree = %s",
			c.yv.root.OutputXML(false))
		}
	}

	return root, cvlErrObj
}

//Validate config - syntax and semantics
func (c *CVL) validate (data *yparser.YParserNode) CVLRetCode {

	depData := c.fetchDataToTmpCache()
	/*
	if (depData != nil) {
		if (0 != C.lyd_merge_to_ctx(&data, depData, C.LYD_OPT_DESTRUCT, ctx)) {
			TRACE_LOG(1, "Failed to merge status data\n")
		}
	}

	if (0 != C.lyd_data_validate(&data, C.LYD_OPT_CONFIG, ctx)) {
		fmt.Println("Validation failed\n")
		return CVL_SYNTAX_ERROR
	}*/

	TRACE_LOG(TRACE_LIBYANG, "\nValidate1 data=%v\n", c.yp.NodeDump(data))
	errObj := c.yp.ValidateData(data, depData)
	if yparser.YP_SUCCESS != errObj.ErrCode {
		return CVL_FAILURE
	}

	cvlErrObj := c.validateCfgExtDep(c.yv.root)
	if CVL_SUCCESS != cvlErrObj.ErrCode {
		return cvlErrObj.ErrCode
	}

	return CVL_SUCCESS
}

func  createCVLErrObj(errObj yparser.YParserError) CVLErrorInfo {

	cvlErrObj :=  CVLErrorInfo {
		TableName : errObj.TableName,
		ErrCode   : CVLRetCode(errObj.ErrCode),
		CVLErrDetails : cvlErrorMap[CVLRetCode(errObj.ErrCode)],
		Keys      : errObj.Keys,
		Value     : errObj.Value,
		Field     : errObj.Field,
		Msg       : errObj.Msg,
		ConstraintErrMsg : errObj.ErrTxt,
		ErrAppTag  : errObj.ErrAppTag,
	}


	return cvlErrObj

}

//Perform syntax checks
func (c *CVL) validateSyntax(data *yparser.YParserNode) (CVLErrorInfo, CVLRetCode) {
	var cvlErrObj CVLErrorInfo
	TRACE_LOG(TRACE_YPARSER, "Validating syntax ....")

	//Get dependent data from 
	c.fetchDataToTmpCache() //fetch data to temp cache for temporary validation

	if errObj  := c.yp.ValidateSyntax(data); errObj.ErrCode != yparser.YP_SUCCESS {

		retCode := CVLRetCode(errObj.ErrCode)

			cvlErrObj =  CVLErrorInfo {
		             TableName : errObj.TableName,
		             ErrCode   : CVLRetCode(errObj.ErrCode),		
			     CVLErrDetails : cvlErrorMap[retCode], 
			     Keys      : errObj.Keys,
			     Value     : errObj.Value,
			     Field     : errObj.Field,
			     Msg       : errObj.Msg,
			     ConstraintErrMsg : errObj.ErrTxt,
			     ErrAppTag	: errObj.ErrAppTag,
			}

			CVL_LOG(ERROR,"Syntax validation failed. Error - %v", cvlErrObj)

		return  cvlErrObj, retCode
	}

	return cvlErrObj, CVL_SUCCESS
}

//Perform semantic checks 
func (c *CVL) validateSemantics(data *yparser.YParserNode, appDepData *yparser.YParserNode) (CVLErrorInfo, CVLRetCode) {
	var cvlErrObj CVLErrorInfo
	
	if (SkipSemanticValidation() == true) {
		return cvlErrObj, CVL_SUCCESS
	}

	//TODO: Clean up this function
	return cvlErrObj, CVL_SUCCESS

	//Get dependent data from 
	depData := c.fetchDataToTmpCache() //fetch data to temp cache for temporary validation


	if (Tracing == true) {
		TRACE_LOG(TRACE_SEMANTIC, "Validating semantics data=%s\n depData =%s\n, appDepData=%s\n....", c.yp.NodeDump(data), c.yp.NodeDump(depData), c.yp.NodeDump(appDepData))
	}

	if errObj := c.yp.ValidateSemantics(data, depData, appDepData); errObj.ErrCode != yparser.YP_SUCCESS {

		retCode := CVLRetCode(errObj.ErrCode)

		cvlErrObj =  CVLErrorInfo {
			TableName : errObj.TableName,
		        ErrCode   : CVLRetCode(errObj.ErrCode),		
			CVLErrDetails : cvlErrorMap[retCode], 
			Keys      : errObj.Keys,
			Value     : errObj.Value,
			Field     : errObj.Field,
			Msg       : errObj.Msg,
			ConstraintErrMsg : errObj.ErrTxt,
			ErrAppTag	: errObj.ErrAppTag,
		}

		CVL_LOG(ERROR,"Semantic validation failed. Error - %v", cvlErrObj)

		return  cvlErrObj, retCode
	}

	return cvlErrObj ,CVL_SUCCESS
}

//Add config data item to accumulate per table
func (c *CVL) addCfgDataItem(configData *map[string]interface{},
			cfgDataItem CVLEditConfigData) (string, string){
	var cfgData map[string]interface{}
	cfgData = *configData

	tblName, key := splitRedisKey(cfgDataItem.Key)
	if (tblName == "" || key == "") {
		//Bad redis key
		return "", ""
	}

	if _, existing := cfgData[tblName]; existing {
		fieldsMap := cfgData[tblName].(map[string]interface{})
		if (cfgDataItem.VOp == OP_DELETE) {
			return tblName, key
		}
		fieldsMap[key] = c.checkFieldMap(&cfgDataItem.Data)
	} else {
		fieldsMap := make(map[string]interface{})
		if (cfgDataItem.VOp == OP_DELETE) {
			fieldsMap[key] = nil
		} else {
			fieldsMap[key] = c.checkFieldMap(&cfgDataItem.Data)
		}
		cfgData[tblName] = fieldsMap
	}

	return tblName, key
}

//Perform user defined custom validation
func (c *CVL) doCustomValidation(custvCfg []custv.CVLEditConfigData,
	curCustvCfg *custv.CVLEditConfigData, yangListName,
	tbl, key string) CVLErrorInfo {

	cvlErrObj := CVLErrorInfo{ErrCode : CVL_SUCCESS}

	for nodeName, custFunc := range modelInfo.tableInfo[tbl].custValidation {
		//check in requestCache for YANG data first
		var node *xmlquery.Node = nil
		if (c.requestCache[tbl][key][0].yangData != nil) {
			node = c.requestCache[tbl][key][0].yangData
		} else {
			//Find the node from YANG tree
			node = c.moveToYangList(yangListName, key)
		}

		if (node == nil) {
			CVL_LOG(WARNING, "Could not find data for custom validation, " +
			"node name=%s, func name =%s", nodeName, custFunc)
			continue
		}

		//find the node value
		//node value is empty for custom validation function at list level
		nodeVal := ""
		if (strings.HasSuffix(nodeName, "_LIST") == false) {
			for nodeLeaf := node.FirstChild; nodeLeaf != nil;
			nodeLeaf = nodeLeaf.NextSibling {
				if (nodeName != nodeLeaf.Data) {
					continue
				}

				if (len(nodeLeaf.Attr) > 0) &&
				(nodeLeaf.Attr[0].Name.Local == "leaf-list") {
					nodeVal = curCustvCfg.Data[nodeName]
				} else {
					nodeVal = nodeLeaf.FirstChild.Data
				}
			}

		}

		//Call csutom validation functions
		CVL_LOG(INFO_TRACE, "Calling custom validation function %s", custFunc)
		errObj := custv.InvokeCustomValidation(&custv.CustomValidation{},
		custFunc,
		&custv.CustValidationCtxt{
			ReqData: custvCfg,
			CurCfg: curCustvCfg,
			YNodeName: nodeName,
			YNodeVal: nodeVal,
			YCur: node,
			RClient: redisClient})

		cvlErrObj = *(*CVLErrorInfo)(unsafe.Pointer(&errObj))

		if (cvlErrObj.ErrCode != CVL_SUCCESS) {
			CVL_LOG(ERROR, "Custom validation failed, Error = %v", cvlErrObj)
		}
	}

	return cvlErrObj
}

