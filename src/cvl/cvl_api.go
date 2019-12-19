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
	"reflect"
	"encoding/json"
	"github.com/go-redis/redis"
	toposort "github.com/philopon/go-toposort"
	"cvl/internal/yparser"
	. "cvl/internal/util"
	"strings"
	"github.com/antchfx/xmlquery"
	"unsafe"
	"runtime"
	custv "cvl/custom_validation"
)

type CVLValidateType uint
const (
	VALIDATE_NONE CVLValidateType = iota //Data is used as dependent data
	VALIDATE_SYNTAX //Syntax is checked and data is used as dependent data
	VALIDATE_SEMANTICS //Semantics is checked
	VALIDATE_ALL //Syntax and Semantics are checked
)

type CVLOperation uint
const (
	OP_NONE   CVLOperation = 0 //Used to just validate the config without any operation
	OP_CREATE = 1 << 0//For Create operation 
	OP_UPDATE = 1 << 1//For Update operation
	OP_DELETE = 1 << 2//For Delete operation
)

var cvlErrorMap = map[CVLRetCode]string {
	CVL_SUCCESS					: "Config Validation Success",
	CVL_SYNTAX_ERROR				: "Config Validation Syntax Error",
	CVL_SEMANTIC_ERROR				: "Config Validation Semantic Error",
	CVL_SYNTAX_MISSING_FIELD			: "Required Field is Missing", 
	CVL_SYNTAX_INVALID_FIELD			: "Invalid Field Received",
	CVL_SYNTAX_INVALID_INPUT_DATA			: "Invalid Input Data Received", 
	CVL_SYNTAX_MULTIPLE_INSTANCE			: "Multiple Field Instances Received", 
	CVL_SYNTAX_DUPLICATE				: "Duplicate Instances Received", 
	CVL_SYNTAX_ENUM_INVALID			        : "Invalid Enum Value Received",  
	CVL_SYNTAX_ENUM_INVALID_NAME 			: "Invalid Enum Value Received", 
	CVL_SYNTAX_ENUM_WHITESPACE		        : "Enum name with leading/trailing whitespaces Received",
	CVL_SYNTAX_OUT_OF_RANGE                         : "Value out of range/length/pattern (data)",
	CVL_SYNTAX_MINIMUM_INVALID        		: "min-elements constraint not honored",
	CVL_SYNTAX_MAXIMUM_INVALID       		: "max-elements constraint not honored",
	CVL_SEMANTIC_DEPENDENT_DATA_MISSING 		: "Dependent Data is missing",
	CVL_SEMANTIC_MANDATORY_DATA_MISSING  		: "Mandatory Data is missing",
	CVL_SEMANTIC_KEY_ALREADY_EXIST 			: "Key already existing.",
	CVL_SEMANTIC_KEY_NOT_EXIST  			: "Key is missing.",
	CVL_SEMANTIC_KEY_DUPLICATE 			: "Duplicate key received",
	CVL_SEMANTIC_KEY_INVALID  			: "Invalid Key Received",
	CVL_INTERNAL_UNKNOWN			 	: "Internal Unknown Error",
	CVL_ERROR                                       : "Generic Error",
	CVL_NOT_IMPLEMENTED                             : "Error Not Implemented",
	CVL_FAILURE                             	: "Generic Failure",
}

//Error code
type CVLRetCode int
const (
	CVL_SUCCESS CVLRetCode = iota
	CVL_ERROR
	CVL_NOT_IMPLEMENTED
	CVL_INTERNAL_UNKNOWN
	CVL_FAILURE
	CVL_SYNTAX_ERROR =  CVLRetCode(yparser.YP_SYNTAX_ERROR)
	CVL_SEMANTIC_ERROR = CVLRetCode(yparser.YP_SEMANTIC_ERROR)
	CVL_SYNTAX_MISSING_FIELD = CVLRetCode(yparser.YP_SYNTAX_MISSING_FIELD)
	CVL_SYNTAX_INVALID_FIELD = CVLRetCode(yparser.YP_SYNTAX_INVALID_FIELD)   /* Invalid Field  */
	CVL_SYNTAX_INVALID_INPUT_DATA = CVLRetCode(yparser.YP_SYNTAX_INVALID_INPUT_DATA) /*Invalid Input Data */
	CVL_SYNTAX_MULTIPLE_INSTANCE = CVLRetCode(yparser.YP_SYNTAX_MULTIPLE_INSTANCE)   /* Multiple Field Instances */
	CVL_SYNTAX_DUPLICATE  = CVLRetCode(yparser.YP_SYNTAX_DUPLICATE)      /* Duplicate Fields  */
	CVL_SYNTAX_ENUM_INVALID  = CVLRetCode(yparser.YP_SYNTAX_ENUM_INVALID) /* Invalid enum value */
	CVL_SYNTAX_ENUM_INVALID_NAME = CVLRetCode(yparser.YP_SYNTAX_ENUM_INVALID_NAME) /* Invalid enum name  */
	CVL_SYNTAX_ENUM_WHITESPACE = CVLRetCode(yparser.YP_SYNTAX_ENUM_WHITESPACE)     /* Enum name with leading/trailing whitespaces */
	CVL_SYNTAX_OUT_OF_RANGE = CVLRetCode(yparser.YP_SYNTAX_OUT_OF_RANGE)    /* Value out of range/length/pattern (data) */
	CVL_SYNTAX_MINIMUM_INVALID = CVLRetCode(yparser.YP_SYNTAX_MINIMUM_INVALID)       /* min-elements constraint not honored  */
	CVL_SYNTAX_MAXIMUM_INVALID  = CVLRetCode(yparser.YP_SYNTAX_MAXIMUM_INVALID)      /* max-elements constraint not honored */
	CVL_SEMANTIC_DEPENDENT_DATA_MISSING  = CVLRetCode(yparser.YP_SEMANTIC_DEPENDENT_DATA_MISSING)  /* Dependent Data is missing */
	CVL_SEMANTIC_MANDATORY_DATA_MISSING = CVLRetCode(yparser.YP_SEMANTIC_MANDATORY_DATA_MISSING) /* Mandatory Data is missing */
	CVL_SEMANTIC_KEY_ALREADY_EXIST = CVLRetCode(yparser.YP_SEMANTIC_KEY_ALREADY_EXIST) /* Key already existing. */
	CVL_SEMANTIC_KEY_NOT_EXIST = CVLRetCode(yparser.YP_SEMANTIC_KEY_NOT_EXIST) /* Key is missing. */
	CVL_SEMANTIC_KEY_DUPLICATE  = CVLRetCode(yparser.YP_SEMANTIC_KEY_DUPLICATE) /* Duplicate key. */
        CVL_SEMANTIC_KEY_INVALID = CVLRetCode(yparser.YP_SEMANTIC_KEY_INVALID)
)

//Strcture for key and data in API
type CVLEditConfigData struct {
	VType CVLValidateType //Validation type
	VOp CVLOperation      //Operation type
	Key string      //Key format : "PORT|Ethernet4"
	Data map[string]string //Value :  {"alias": "40GE0/28", "mtu" : 9100,  "admin_status":  down}
}

func Initialize() CVLRetCode {
	if (cvlInitialized == true) {
		//CVL has already been initialized
		return CVL_SUCCESS
	}

	//Initialize redis Client 
	redisClient = redis.NewClient(&redis.Options{
		Addr:     ":6379",
		Password: "", // no password set
		DB:       int(CONFIG_DB),  // use APP DB
	})

	if (redisClient == nil) {
		CVL_LOG(FATAL, "Unable to connect with Redis Config DB")
		return CVL_ERROR
	}

	//Load lua script into redis
	luaScripts = make(map[string]*redis.Script)
	loadLuaScript(luaScripts)

	yparser.Initialize()

	modelInfo.modelNs =  make(map[string]*modelNamespace) //redis table to model name
	modelInfo.tableInfo = make(map[string]*modelTableInfo) //model namespace 
	modelInfo.allKeyDelims = make(map[string]bool) //all key delimiter
	modelInfo.redisTableToYangList = make(map[string][]string) //Redis table to Yang list map
	dbNameToDbNum = map[string]uint8{"APPL_DB": APPL_DB, "CONFIG_DB": CONFIG_DB}

	// Load all YIN schema files
	if retCode := loadSchemaFiles(); retCode != CVL_SUCCESS {
		return retCode
	}

	//Compile leafref path
	compileLeafRefPath()

	//Compile all must exps
	compileMustExps()

	//Compile all when exps
	compileWhenExps()

	//Add all table names to be fetched to validate 'must' expression
	addTableNamesForMustExp()

	//Build reverse leafref info i.e. which table/field uses one table through leafref
	buildRefTableInfo()

	cvlInitialized = true

	return CVL_SUCCESS
}

func Finish() {
	yparser.Finish()
}

func ValidationSessOpen() (*CVL, CVLRetCode) {
	cvl :=  &CVL{}
	cvl.tmpDbCache = make(map[string]interface{})
	cvl.requestCache = make(map[string]map[string][]*requestCacheType)
	cvl.maxTableElem = make(map[string]int)
	cvl.yp = &yparser.YParser{}
	cvl.yv = &YValidator{}
	cvl.yv.root = &xmlquery.Node{Type: xmlquery.DocumentNode}

	if (cvl == nil || cvl.yp == nil) {
		return nil, CVL_FAILURE
	}

	return cvl, CVL_SUCCESS
}

func ValidationSessClose(c *CVL) CVLRetCode {
	c.yp.DestroyCache()
	c = nil

	return CVL_SUCCESS
}

func (c *CVL) ValidateStartupConfig(jsonData string) CVLRetCode {
	//Check config data syntax
	//Finally validate
	return CVL_NOT_IMPLEMENTED
}

func (c *CVL) ValidateIncrementalConfig(jsonData string) CVLRetCode {
	//Check config data syntax
	//Fetch the depedent data
	//Merge config and dependent data
	//Finally validate
	c.clearTmpDbCache()
	var  v interface{}

	b := []byte(jsonData)
	if err := json.Unmarshal(b, &v); err != nil {
		return CVL_SYNTAX_ERROR
	}

	var dataMap map[string]interface{} = v.(map[string]interface{})

	root, _ := c.translateToYang(&dataMap)
	if root == nil {
		return CVL_SYNTAX_ERROR

	}

	//Add and fetch entries if already exists in Redis
	for tableName, data := range dataMap {
		for key, _ := range data.(map[string]interface{}) {
			c.addTableEntryToCache(tableName, key)
		}
	}

	existingData := c.fetchDataToTmpCache()

	//Merge existing data for update syntax or checking duplicate entries
	var errObj yparser.YParserError
	if (existingData != nil) {
		if root, errObj = c.yp.MergeSubtree(root, existingData);
				errObj.ErrCode != yparser.YP_SUCCESS {
			return CVL_ERROR
		}
	}

	//Clear cache
	c.clearTmpDbCache()

	//Add tables for 'must' expression
	for tableName, _ := range dataMap {
		c.addTableDataForMustExp(OP_NONE, tableName)
	}

	//Perform validation
	if _, cvlRetCode := c.validateSemantics(root, nil); cvlRetCode != CVL_SUCCESS {
		return cvlRetCode
	}

	return CVL_SUCCESS
}

//Validate data for operation
func (c *CVL) ValidateConfig(jsonData string) CVLRetCode {
	c.clearTmpDbCache()
	var  v interface{}

	b := []byte(jsonData)
	if err := json.Unmarshal(b, &v); err == nil {
		var value map[string]interface{} = v.(map[string]interface{})
		root, _ := c.translateToYang(&value)
		//if ret == CVL_SUCCESS && root != nil {
		if root == nil {
			return CVL_FAILURE

		}

		if (c.validate(root) != CVL_SUCCESS) {
			return CVL_FAILURE
		}

	}

	return CVL_SUCCESS
}

//Validate config data based on edit operation - no marshalling in between
func (c *CVL) ValidateEditConfig(cfgData []CVLEditConfigData) (cvlErr CVLErrorInfo, ret CVLRetCode) {

	defer func() {
		if (cvlErr.ErrCode != CVL_SUCCESS) {
			CVL_LOG(ERROR, "ValidateEditConfig() failed , %+v", cvlErr)
		}
	}()

	var cvlErrObj CVLErrorInfo

	caller := ""
	if (IsTraceSet()) {
		pc := make([]uintptr, 10)
		runtime.Callers(2, pc)
		f := runtime.FuncForPC(pc[0])
		caller = f.Name()
	}

	TRACE_LOG(INFO_TRACE, "ValidateEditConfig() called from %s() : %v", caller, cfgData)

	if (SkipValidation() == true) {
		CVL_LOG(INFO_TRACE, "Skipping CVL validation.")
		return cvlErrObj, CVL_SUCCESS
	}

	//Type cast to custom validation cfg data
	sliceHeader := *(*reflect.SliceHeader)(unsafe.Pointer(&cfgData))
	custvCfg := *(*[]custv.CVLEditConfigData)(unsafe.Pointer(&sliceHeader))

	c.clearTmpDbCache()
	//c.yv.root.FirstChild = nil
	//c.yv.root.LastChild = nil


	//Step 1: Get requested data first
	//add all dependent data to be fetched from Redis
	requestedData := make(map[string]interface{})

	cfgDataLen := len(cfgData)
	for i := 0; i < cfgDataLen; i++ {
		if (VALIDATE_ALL != cfgData[i].VType) {
			continue
		}

		//Add config data item to be validated
		tbl,key := c.addCfgDataItem(&requestedData, cfgData[i])

		//Add to request cache
		reqTbl, exists := c.requestCache[tbl]
		if (exists == false) {
			//Create new table key data
			reqTbl = make(map[string][]*requestCacheType)
		}
		cfgDataItemArr, _ := reqTbl[key]
		cfgDataItemArr = append(cfgDataItemArr, &requestCacheType{cfgData[i], nil})
		reqTbl[key] = cfgDataItemArr
		c.requestCache[tbl] = reqTbl

		//Invalid table name or invalid key separator 
		if key == "" {
			cvlErrObj.ErrCode = CVL_SYNTAX_ERROR
			cvlErrObj.Msg = "Invalid table or key for " + cfgData[i].Key
			cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
			return cvlErrObj, CVL_SYNTAX_ERROR
		}

		switch cfgData[i].VOp {
		case OP_CREATE:
			//Check max-element constraint 
			if ret := c.checkMaxElemConstraint(tbl); ret != CVL_SUCCESS {
				cvlErrObj.ErrCode = CVL_SYNTAX_ERROR
				cvlErrObj.ErrAppTag = "too-many-elements"
				cvlErrObj.Msg = "Max elements limit reached"
				cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
				return cvlErrObj, CVL_SYNTAX_ERROR
			}

			/*if (c.addTableEntryForMustExp(&cfgData[i], tbl) != CVL_SUCCESS) {
				c.addTableDataForMustExp(cfgData[i].VOp, tbl)
			}*/

		case OP_UPDATE:
			//Get the existing data from Redis to cache, so that final validation can be done after merging this dependent data
			/*if (c.addTableEntryForMustExp(&cfgData[i], tbl) != CVL_SUCCESS) {
				c.addTableDataForMustExp(cfgData[i].VOp, tbl)
			}*/
			c.addTableEntryToCache(tbl, key)

		case OP_DELETE:
			if (len(cfgData[i].Data) > 0) {
				//Check constraints for deleting field(s)
				for field, _ := range cfgData[i].Data {
					if (c.checkDeleteConstraint(cfgData, tbl, key, field) != CVL_SUCCESS) {
						cvlErrObj.ErrCode = CVL_SEMANTIC_ERROR
						cvlErrObj.Msg = "Delete constraint failed"
						cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
						return cvlErrObj, CVL_SEMANTIC_ERROR
					}
				}
			} else {
				//Entire entry to be deleted
				if (c.checkDeleteConstraint(cfgData, tbl, key, "") != CVL_SUCCESS) {
					cvlErrObj.ErrCode = CVL_SEMANTIC_ERROR
					cvlErrObj.Msg = "Delete constraint failed"
					cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
					return cvlErrObj, CVL_SEMANTIC_ERROR
				}
				//TBD : Can we do this ?
				//No entry has depedency on this key, 
				//remove from requestCache, we don't need any more as depedent data
				//delete(c.requestCache[tbl], key)
			}

			/*if (c.addTableEntryForMustExp(&cfgData[i], tbl) != CVL_SUCCESS) {
				c.addTableDataForMustExp(cfgData[i].VOp, tbl)
			}*/

			c.addTableEntryToCache(tbl, key)
		}
	}

	//Only for tracing
	if (IsTraceSet()) {
		jsonData := ""

		jsonDataBytes, err := json.Marshal(requestedData)
		if (err == nil) {
			jsonData = string(jsonDataBytes)
		} else {
			cvlErrObj.ErrCode = CVL_SYNTAX_ERROR
			cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
			return cvlErrObj, CVL_SYNTAX_ERROR
		}

		TRACE_LOG(TRACE_LIBYANG, "Requested JSON Data = [%s]\n", jsonData)
	}

	//Step 2 : Perform syntax validation only
	yang, errN := c.translateToYang(&requestedData)
	if (errN.ErrCode == CVL_SUCCESS) {
		if cvlErrObj, cvlRetCode := c.validateSyntax(yang); cvlRetCode != CVL_SUCCESS {
			return cvlErrObj, cvlRetCode
		}
	} else {
		return errN,errN.ErrCode
	}

	//Step 3 : Check keys and update dependent data
	//dependentData := make(map[string]interface{})

	//Step 4 : Perform validation
	if cvlErrObj, cvlRetCode1 := c.validateSemantics(yang, nil);
	cvlRetCode1 != CVL_SUCCESS {
			return cvlErrObj, cvlRetCode1
	}

	for i := 0; i < cfgDataLen; i++ {

		if (cfgData[i].VType != VALIDATE_ALL && cfgData[i].VType != VALIDATE_SEMANTICS) {
			continue
		}

		tbl, key := splitRedisKey(cfgData[i].Key)

		//Step 3.1 : Check keys
		switch cfgData[i].VOp {
		case OP_CREATE:
			//Check key should not already exist
			n, err1 := redisClient.Exists(cfgData[i].Key).Result()
			if (err1 == nil && n > 0) {
				//Check if key deleted and CREATE done in same session, 
				//allow to create the entry
				deletedInSameSession := false
				if  tbl != ""  && key != "" {
					for _, cachedCfgData := range c.requestCache[tbl][key] {
						if cachedCfgData.reqData.VOp == OP_DELETE {
							deletedInSameSession = true
							break
						}
					}
				}

				if deletedInSameSession == false {
					CVL_LOG(ERROR, "\nValidateEditConfig(): Key = %s already exists", cfgData[i].Key)
					cvlErrObj.ErrCode = CVL_SEMANTIC_KEY_ALREADY_EXIST
					cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
					return cvlErrObj, CVL_SEMANTIC_KEY_ALREADY_EXIST 

				} else {
					TRACE_LOG(TRACE_CREATE, "\nKey %s is deleted in same session, skipping key existence check for OP_CREATE operation", cfgData[i].Key)
				}
			}

			c.yp.SetOperation("CREATE")

		case OP_UPDATE:
			n, err1 := redisClient.Exists(cfgData[i].Key).Result()
			if (err1 != nil || n == 0) { //key must exists
				CVL_LOG(ERROR, "\nValidateEditConfig(): Key = %s does not exist", cfgData[i].Key)
				cvlErrObj.ErrCode = CVL_SEMANTIC_KEY_NOT_EXIST
				cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
				return cvlErrObj, CVL_SEMANTIC_KEY_NOT_EXIST
			}

			c.yp.SetOperation("UPDATE")

		case OP_DELETE:
			n, err1 := redisClient.Exists(cfgData[i].Key).Result()
			if (err1 != nil || n == 0) { //key must exists
				CVL_LOG(ERROR, "\nValidateEditConfig(): Key = %s does not exist", cfgData[i].Key)
				cvlErrObj.ErrCode = CVL_SEMANTIC_KEY_NOT_EXIST
				cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
				return cvlErrObj, CVL_SEMANTIC_KEY_NOT_EXIST
			}

			c.yp.SetOperation("DELETE")
			//store deleted keys
		}

		yangListName := getRedisTblToYangList(tbl, key)

		//Run all custom validations
		cvlErrObj= c.doCustomValidation(custvCfg, &custvCfg[i], yangListName,
		tbl, key)
		if cvlErrObj.ErrCode != CVL_SUCCESS {
			return cvlErrObj,cvlErrObj.ErrCode
		}

		if cvlErrObj = c.validateEditCfgExtDep(yangListName, key, &cfgData[i]);
		cvlErrObj.ErrCode != CVL_SUCCESS {
			return cvlErrObj,cvlErrObj.ErrCode
		}
	}

	/* TBD
	var depYang *yparser.YParserNode = nil
	if (len(dependentData) > 0) {
		depYang, errN = c.translateToYang(&dependentData)
	}
	//Step 4 : Perform validation
	if cvlErrObj, cvlRetCode1 := c.validateSemantics(yang, depYang);
	cvlRetCode1 != CVL_SUCCESS {
			return cvlErrObj, cvlRetCode1
	}
	*/

	//Cache validated data
	/*
	if errObj := c.yp.CacheSubtree(false, yang); errObj.ErrCode != yparser.YP_SUCCESS {
		TRACE_LOG(TRACE_CACHE, "Could not cache validated data")
	}
	*/

	c.yp.DestroyCache()
	return cvlErrObj, CVL_SUCCESS
}

/* Fetch the Error Message from CVL Return Code. */
func GetErrorString(retCode CVLRetCode) string{

	return cvlErrorMap[retCode]

}

//Validate key only
func (c *CVL) ValidateKeys(key []string) CVLRetCode {
	return CVL_NOT_IMPLEMENTED
}

//Validate key and data
func (c *CVL) ValidateKeyData(key string, data string) CVLRetCode {
	return CVL_NOT_IMPLEMENTED
}

//Validate key, field and value
func (c *CVL) ValidateFields(key string, field string, value string) CVLRetCode {
	return CVL_NOT_IMPLEMENTED
}

func (c *CVL) addDepEdges(graph *toposort.Graph, tableList []string) {
	//Add all the depedency edges for graph nodes
	for ti :=0; ti < len(tableList); ti++ {

		redisTblTo := getYangListToRedisTbl(tableList[ti])

		for tj :=0; tj < len(tableList); tj++ {

			if (tableList[ti] == tableList[tj]) {
				//same table, continue
				continue
			}

			redisTblFrom := getYangListToRedisTbl(tableList[tj])

			//map for checking duplicate edge
			dupEdgeCheck := map[string]string{}

			for _, leafRefs := range modelInfo.tableInfo[tableList[tj]].leafRef {
				for _, leafRef := range leafRefs {
					if (strings.Contains(leafRef.path, tableList[ti] + "_LIST")) == false {
						continue
					}

					toName, exists := dupEdgeCheck[redisTblFrom]
					if (exists == true) && (toName == redisTblTo) {
						//Don't add duplicate edge
						continue
					}

					//Add and store the edge in map
					graph.AddEdge(redisTblFrom, redisTblTo)
					dupEdgeCheck[redisTblFrom] = redisTblTo

					CVL_LOG(INFO_DEBUG,
					"addDepEdges(): Adding edge %s -> %s", redisTblFrom, redisTblTo)
				}
			}
		}
	}
}

//Sort list of given tables as per their dependency
func (c *CVL) SortDepTables(inTableList []string) ([]string, CVLRetCode) {

	tableList := []string{}

	//Skip all unknown tables
	for ti := 0; ti < len(inTableList); ti++ {
		_, exists := modelInfo.tableInfo[inTableList[ti]]
		if exists == false {
			continue
		}

		tableList = append(tableList, inTableList[ti])
	}

	//Add all the table names in graph nodes
	graph := toposort.NewGraph(len(tableList))
	for ti := 0; ti < len(tableList); ti++ {
		graph.AddNodes(tableList[ti])
	}

	//Add all dependency egdes
	c.addDepEdges(graph, tableList)

	//Now perform topological sort
	result, ret := graph.Toposort()
	if ret == false {
		return nil, CVL_ERROR
	}

	return result, CVL_SUCCESS
}

//Get the order list(parent then child) of tables in a given YANG module
//within a single model this is obtained using leafref relation
func (c *CVL) GetOrderedTables(yangModule string) ([]string, CVLRetCode) {
	tableList := []string{}

	//Get all the table names under this model
	for tblName, tblNameInfo := range modelInfo.tableInfo {
		if (tblNameInfo.modelName == yangModule) {
			tableList = append(tableList, tblName)
		}
	}

	return c.SortDepTables(tableList)
}

func (c *CVL) addDepTables(tableMap map[string]bool, tableName string) {

	//Mark it is added in list
	tableMap[tableName] = true

	//Now find all tables referred in leafref from this table
	for _, leafRefs := range modelInfo.tableInfo[tableName].leafRef {
		for _, leafRef := range leafRefs {
			for _, refTbl := range leafRef.yangListNames {
				c.addDepTables(tableMap, getYangListToRedisTbl(refTbl)) //call recursively
			}
		}
	}
}

//Get the list of dependent tables for a given table in a YANG module
func (c *CVL) GetDepTables(yangModule string, tableName string) ([]string, CVLRetCode) {
	tableList := []string{}
	tblMap := make(map[string]bool)

	if _, exists := modelInfo.tableInfo[tableName]; exists == false {
		CVL_LOG(INFO_DEBUG, "GetDepTables(): Unknown table %s\n", tableName)
		return []string{}, CVL_ERROR
	}

	c.addDepTables(tblMap, tableName)

	for tblName, _ := range tblMap {
		tableList = append(tableList, tblName)
	}

	//Add all the table names in graph nodes
	graph := toposort.NewGraph(len(tableList))
	for ti := 0; ti < len(tableList); ti++ {
		CVL_LOG(INFO_DEBUG, "GetDepTables(): Adding node %s\n", tableList[ti])
		graph.AddNodes(tableList[ti])
	}

	//Add all dependency egdes
	c.addDepEdges(graph, tableList)

	//Now perform topological sort
	result, ret := graph.Toposort()
	if ret == false {
		return nil, CVL_ERROR
	}

	return result, CVL_SUCCESS
}

//Get the dependent (Redis keys) to be deleted or modified
//for a given entry getting deleted
func (c *CVL) GetDepDataForDelete(redisKey string) ([]string, []string) {

	tableName, key := splitRedisKey(redisKey)

	if (tableName == "") || (key == "") {
		CVL_LOG(INFO_DEBUG, "GetDepDataForDelete(): Unknown or invalid table %s\n",
		tableName)
	}

	if _, exists := modelInfo.tableInfo[tableName]; exists == false {
		CVL_LOG(INFO_DEBUG, "GetDepDataForDelete(): Unknown table %s\n", tableName)
		return []string{}, []string{} 
	}

	mCmd := map[string]*redis.StringSliceCmd{}
	mFilterScripts := map[string]string{}
	pipe := redisClient.Pipeline()

	for _, refTbl := range modelInfo.tableInfo[tableName].refFromTables {

		//check if ref field is a key
		numKeys := len(modelInfo.tableInfo[refTbl.tableName].keys)
		idx := 0
		for ; idx < numKeys; idx++ {
			if (modelInfo.tableInfo[refTbl.tableName].keys[idx] == refTbl.field) {
				//field is key comp
				mCmd[refTbl.tableName] = pipe.Keys(fmt.Sprintf("%s|*%s*",
				refTbl.tableName, key)) //write into pipeline}
				break
			}
		}

		if (idx == numKeys) {
			//field is hash-set field, not a key, match with hash-set field
			//prepare the lua filter script
			// ex: (h.members: == 'Ethernet4,' or (string.find(h['members@'], 'Ethernet4') != nil)
			//',' to include leaf-list case
			mFilterScripts[refTbl.tableName] =
			fmt.Sprintf("return (h.%s == '%s') or " +
			"h['ports@'] ~= nil and (string.find(h['%s@'], '%s') ~= nil)",
			refTbl.field, key, refTbl.field, key)
		}
	}

	_, err := pipe.Exec()
	if err != nil {
		CVL_LOG(ERROR, "Failed to fetch dependent key details for table %s", tableName)
	}
	pipe.Close()

	//Add dependent keys which should be modified
	depKeysForMod := []string{}
	for tableName, mFilterScript := range mFilterScripts {
		refKeys, err := luaScripts["filter_keys"].Run(redisClient, []string{},
		tableName + "|*", "", mFilterScript).Result()

		if (err != nil) {
			CVL_LOG(ERROR, "Lua script error (%v)", err)
		}
		if (refKeys == nil) {
		//No reference field found
			continue
		}

		refKeysStr := string(refKeys.(string))

		if (refKeysStr != "") {
			//Add all keys whose fields to be updated
			depKeysForMod = append(depKeysForMod,
			strings.Split(refKeysStr, ",")...)
		}
	}

	depKeys := []string{}
	for tblName, keys := range mCmd {
		res, err := keys.Result()
		if (err != nil) {
			CVL_LOG(ERROR, "Failed to fetch dependent key details for table %s", tblName)
			continue
		}

		//Add keys found
		depKeys = append(depKeys, res...)

		//For each key, find dependent data for delete recursively
		for i :=0; i< len(res); i++ {
			retDepKeys, retDepKeysForMod := c.GetDepDataForDelete(res[i])
			depKeys = append(depKeys, retDepKeys...)
			depKeysForMod = append(depKeysForMod, retDepKeysForMod...)
		}
	}

	return depKeys, depKeysForMod
}
