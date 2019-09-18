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
	"encoding/json"
	"github.com/go-redis/redis"
	"path/filepath"
	"cvl/internal/yparser"
	. "cvl/internal/util"
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

	//Scan schema directory to get all schema files
	modelFiles, err := filepath.Glob(CVL_SCHEMA + "/*.yin")
	if err != nil {
		CVL_LOG(FATAL ,"Could not read schema %v", err)
	}

	yparser.Initialize()

	modelInfo.modelNs =  make(map[string]modelNamespace) //redis table to model name
	modelInfo.tableInfo = make(map[string]modelTableInfo) //model namespace 
	modelInfo.allKeyDelims = make(map[string]bool) //all key delimiter
	dbNameToDbNum = map[string]uint8{"APPL_DB": APPL_DB, "CONFIG_DB": CONFIG_DB}

	/* schema */
	for _, modelFilePath := range modelFiles {
		_, modelFile := filepath.Split(modelFilePath)

		TRACE_LOG(INFO_DEBUG, TRACE_LIBYANG, "Parsing schema file %s ...\n", modelFilePath)
		var module *yparser.YParserModule
		if module, _ = yparser.ParseSchemaFile(modelFilePath); module == nil {

			CVL_LOG(FATAL,fmt.Sprintf("Unable to parse schema file %s", modelFile))
			return CVL_ERROR
		}

		storeModelInfo(modelFile, module)
	}

	//Add all table names to be fetched to validate 'must' expression
	addTableNamesForMustExp()

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
	loadLuaScript()

	cvlInitialized = true

	return CVL_SUCCESS
}

func Finish() {
	yparser.Finish()
}

func ValidationSessOpen() (*CVL, CVLRetCode) {
	cvl :=  &CVL{}
	cvl.tmpDbCache = make(map[string]interface{})
	cvl.requestCache = make(map[string]map[string][]CVLEditConfigData)
	cvl.yp = &yparser.YParser{}

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
func (c *CVL) ValidateEditConfig(cfgData []CVLEditConfigData) (CVLErrorInfo, CVLRetCode) {
	var cvlErrObj CVLErrorInfo

	if (SkipValidation() == true) {

		return cvlErrObj, CVL_SUCCESS
	}

	c.clearTmpDbCache()

	//Step 1: Get requested dat first
	//add all dependent data to be fetched from Redis
	requestedData := make(map[string]interface{})

        for _, cfgDataItem := range cfgData {
		if (VALIDATE_ALL != cfgDataItem.VType) {
			continue
		}

		//Add config data item to be validated
		tbl,key := c.addCfgDataItem(&requestedData, cfgDataItem)

		//Add to request cache
		reqTbl, exists := c.requestCache[tbl]
		if (exists == false) {
			//Create new table key data
			reqTbl = make(map[string][]CVLEditConfigData)
		}
		cfgDataItemArr, _ := reqTbl[key]
		cfgDataItemArr = append(cfgDataItemArr, cfgDataItem)
		reqTbl[key] = cfgDataItemArr
		c.requestCache[tbl] = reqTbl

		//Invalid table name or invalid key separator 
		if key == "" {
			cvlErrObj.ErrCode = CVL_SYNTAX_ERROR
			cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
			return cvlErrObj, CVL_SYNTAX_ERROR
		}

		switch cfgDataItem.VOp {
		case OP_CREATE:
			if (c.addTableEntryForMustExp(&cfgDataItem, tbl) != CVL_SUCCESS) {
				c.addTableDataForMustExp(cfgDataItem.VOp, tbl)
			}

		case OP_UPDATE:
			//Get the existing data from Redis to cache, so that final validation can be done after merging this dependent data
			if (c.addTableEntryForMustExp(&cfgDataItem, tbl) != CVL_SUCCESS) {
				c.addTableDataForMustExp(cfgDataItem.VOp, tbl)
			}
			c.addTableEntryToCache(tbl, key)

		case OP_DELETE:
			if (len(cfgDataItem.Data) > 0) {
				//Delete a single field
				if (len(cfgDataItem.Data) != 1)  {
					CVL_LOG(ERROR, "Only single field is allowed for field deletion")
				} else {
					for field, _ := range cfgDataItem.Data {
						if (c.checkDeleteConstraint(cfgData, tbl, key, field) != CVL_SUCCESS) {
							cvlErrObj.ErrCode = CVL_SEMANTIC_ERROR
							cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
							return cvlErrObj, CVL_SEMANTIC_ERROR
						}
						break //only one field there
					}
				}
			} else {
				//Entire entry to be deleted
				if (c.checkDeleteConstraint(cfgData, tbl, key, "") != CVL_SUCCESS) {
					cvlErrObj.ErrCode = CVL_SEMANTIC_ERROR
					cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
					return cvlErrObj, CVL_SEMANTIC_ERROR 
				}
				//TBD : Can we do this ?
				//No entry has depedency on this key, 
				//remove from requestCache, we don't need any more as depedent data
				//delete(c.requestCache[tbl], key)
			}

			if (c.addTableEntryForMustExp(&cfgDataItem, tbl) != CVL_SUCCESS) {
				c.addTableDataForMustExp(cfgDataItem.VOp, tbl)
			}

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

		TRACE_LOG(INFO_DATA, TRACE_LIBYANG, "Requested JSON Data = [%s]\n", jsonData)
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
	dependentData := make(map[string]interface{})

        for _, cfgDataItem := range cfgData {

		if (cfgDataItem.VType == VALIDATE_ALL || cfgDataItem.VType == VALIDATE_SEMANTICS) {
			//Step 3.1 : Check keys
			switch cfgDataItem.VOp {
			case OP_CREATE:
				//Check key should not already exist
				n, err1 := redisClient.Exists(cfgDataItem.Key).Result()
				if (err1 == nil && n > 0) {
					//Check if key deleted and CREATE done in same session, 
					//allow to create the entry
					tbl, key := splitRedisKey(cfgDataItem.Key)
					deletedInSameSession := false
					if  tbl != ""  && key != "" {
						for _, cachedCfgData := range c.requestCache[tbl][key] {
							if cachedCfgData.VOp == OP_DELETE {
								deletedInSameSession = true
								break
							}
						}
					}

					if deletedInSameSession == false {
						CVL_LOG(ERROR, "\nValidateEditConfig(): Key = %s already exists", cfgDataItem.Key)
						cvlErrObj.ErrCode = CVL_SEMANTIC_KEY_ALREADY_EXIST
						cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
						return cvlErrObj, CVL_SEMANTIC_KEY_ALREADY_EXIST 

					} else {
						TRACE_LOG(INFO_API, TRACE_CREATE, "\nKey %s is deleted in same session, skipping key existence check for OP_CREATE operation", cfgDataItem.Key)
					}
				}

				c.yp.SetOperation("CREATE")

			case OP_UPDATE:
				n, err1 := redisClient.Exists(cfgDataItem.Key).Result()
				if (err1 != nil || n == 0) { //key must exists
					CVL_LOG(ERROR, "\nValidateEditConfig(): Key = %s does not exist", cfgDataItem.Key)
					cvlErrObj.ErrCode = CVL_SEMANTIC_KEY_ALREADY_EXIST
					cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
					return cvlErrObj, CVL_SEMANTIC_KEY_NOT_EXIST
				}

				c.yp.SetOperation("UPDATE")

			case OP_DELETE:
				n, err1 := redisClient.Exists(cfgDataItem.Key).Result()
				if (err1 != nil || n == 0) { //key must exists
					CVL_LOG(ERROR, "\nValidateEditConfig(): Key = %s does not exist", cfgDataItem.Key)
					cvlErrObj.ErrCode = CVL_SEMANTIC_KEY_ALREADY_EXIST
					cvlErrObj.CVLErrDetails = cvlErrorMap[cvlErrObj.ErrCode]
					return cvlErrObj, CVL_SEMANTIC_KEY_NOT_EXIST
				}

				c.yp.SetOperation("DELETE")
				//store deleted keys
			}

		}/* else if (cfgDataItem.VType == VALIDATE_NONE) {
			//Step 3.2 : Get dependent data

			switch cfgDataItem.VOp {
			case OP_CREATE:
				//NOP
			case OP_UPDATE:
				//NOP
			case OP_DELETE:
				tbl,key := c.addCfgDataItem(&dependentData, cfgDataItem)
				//update cache by removing deleted entry
				c.updateDeleteDataToCache(tbl, key)
				//store deleted keys
			}
		}*/
	}

	var depYang *yparser.YParserNode = nil
	if (len(dependentData) > 0) {
		depYang, errN = c.translateToYang(&dependentData)
	}
	//Step 4 : Perform validation
	if cvlErrObj, cvlRetCode1 := c.validateSemantics(yang, depYang); cvlRetCode1 != CVL_SUCCESS {
			return cvlErrObj, cvlRetCode1 
	}

	//Cache validated data
	/*
	if errObj := c.yp.CacheSubtree(false, yang); errObj.ErrCode != yparser.YP_SUCCESS {
		TRACE_LOG(INFO_API, TRACE_CACHE, "Could not cache validated data")
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
