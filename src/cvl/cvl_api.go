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
	OP_NONE   CVLOperation = iota //Used to just validate the config without any operation
	OP_CREATE //For Create operation 
	OP_UPDATE //For Update operation
	OP_DELETE //For Delete operation
)

var cvlErrorMap = map[CVLRetCode]string {
		CVL_SUCCESS					: "Config Validation Success",
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
}

//Error code
type CVLRetCode int
const (
	CVL_SUCCESS CVLRetCode = iota
	CVL_SYNTAX_ERROR =  CVLRetCode(yparser.YP_SYNTAX_ERROR)
	CVL_SEMANTIC_ERROR = CVLRetCode(yparser.YP_SEMANTIC_ERROR)
	CVL_ERROR
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
	CVL_NOT_IMPLEMENTED
	CVL_INTERNAL_UNKNOWN
	CVL_FAILURE
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
	return CVL_SUCCESS
}

func (c *CVL) ValidateIncrementalConfig(jsonData string) CVLRetCode {
	//Check config data syntax
	//Fetch the depedent data
	//Merge config and dependent data
	//Finally validate
	return CVL_SUCCESS
}

//Validate data for operation
func (c *CVL) ValidateConfig(jsonData string) CVLRetCode {
	c.clearTmpDbCache()
	var  v interface{}

	b := []byte(jsonData)
	if err := json.Unmarshal(b, &v); err == nil {
		var value map[string]interface{} = v.(map[string]interface{})
		root, _ := c.translateToYang1(&value)
		//if ret == CVL_SUCCESS && root != nil {
		if root == nil {
			//var outBuf *C.char
			//C.lyd_print_mem(&outBuf, root, C.LYD_XML, 0)
			return CVL_FAILURE

		}

		if (c.validate1(root) != CVL_SUCCESS) {
			return CVL_FAILURE
		}

	}

	/*
	//Convert JSON to YANG XML 
	doc, err := c.translateToYang(jsonDat cvlErrObj, cvlRetCode)
	if (err == CVL_SUCCESS) {
		yangXml := doc.OutputXML(true)

		return c.validate(yangXml)
	}
	*/

	return CVL_SUCCESS
}

func (c *CVL) ValidateEditConfig1(cfgData []CVLEditConfigData) (CVLErrorInfo, CVLRetCode) {
	return c.ValidateEditConfig(cfgData)
}

//Validate config data based on edit operation - no marshalling in between
func (c *CVL) ValidateEditConfig(cfgData []CVLEditConfigData) (CVLErrorInfo, CVLRetCode) {
	var cvlErrObj CVLErrorInfo
	c.clearTmpDbCache()

	//Step 1: Get requested dat first
	//add all dependent data to be fetched from Redis
	requestedData := make(map[string]interface{})

        for _, cfgDataItem := range cfgData {
		if (VALIDATE_ALL != cfgDataItem.VType) {
			continue
		}

		tbl,key := c.addCfgDataItem(&requestedData, cfgDataItem)

		if key == "" {
			return cvlErrObj, CVL_SEMANTIC_ERROR
		}

		switch cfgDataItem.VOp {
		case OP_CREATE:
			c.addTableDataForMustExp(tbl)

		case OP_UPDATE:
			//Get the existing data from Redis to cache, so that final validation can be done after merging this dependent data
			c.addUpdateDataToCache(tbl, key)
			c.addTableDataForMustExp(tbl)

		case OP_DELETE:
			if (len(cfgDataItem.Data) > 0) {
				//Delete a single field
				if (len(cfgDataItem.Data) != 1)  {
					CVL_LOG(ERROR, "Only single field is allowed for field deletion")
				} else {
					for field, _ := range cfgDataItem.Data {
						if (c.checkDeleteConstraint(cfgData, tbl, key, field) != CVL_SUCCESS) {
							return cvlErrObj, CVL_SEMANTIC_ERROR
						}
						break //only one field there
					}
				}
			} else {
				//Entire entry to be deleted
				if (c.checkDeleteConstraint(cfgData, tbl, key, "") != CVL_SUCCESS) {
					return cvlErrObj, CVL_SEMANTIC_ERROR 
				}
			}

			c.updateDeleteDataToCache(tbl, key)
			c.addTableDataForMustExp(tbl)
		}
	}

	if (IsTraceSet()) {
		jsonData := ""

		jsonDataBytes, err := json.Marshal(requestedData) //Optimize TBD:
		if (err == nil) {
			jsonData = string(jsonDataBytes)
		} else {
			return cvlErrObj, CVL_SYNTAX_ERROR 
		}

		TRACE_LOG(INFO_DATA, TRACE_LIBYANG, "JSON Data = %s\n", jsonData)
	}

	//Step 2 : Perform syntax validation only
	yang, errN := c.translateToYang1(&requestedData)
	if (errN.ErrCode == CVL_SUCCESS) {
		if cvlErrObj, cvlRetCode := c.validateSyntax1(yang); cvlRetCode != CVL_SUCCESS {
			return cvlErrObj, cvlRetCode 
		}
	} else {
		return errN,errN.ErrCode 
	}

	//Step 3 : Check keys and update dependent data
	dependentData := make(map[string]interface{})
	deletedKeys := make(map[string]interface{}) //will store deleted key

        for _, cfgDataItem := range cfgData {

		if (cfgDataItem.VType == VALIDATE_ALL || cfgDataItem.VType == VALIDATE_SEMANTICS) {
			//Step 3.1 : Check keys
			switch cfgDataItem.VOp {
			case OP_CREATE:
				//If key is deleted and CREATE is done in same session, no need to check for key existence in case of create
				if  _, deleted := deletedKeys[cfgDataItem.Key]; deleted == false {

					//Check key should not already exist
					n, err1 := redisClient.Exists(cfgDataItem.Key).Result()
					if (err1 == nil && n > 0) {
						TRACE_LOG(INFO_API, TRACE_CREATE, "\nValidateEditConfig(): Key = %s alreday exists", cfgDataItem.Key)
						return cvlErrObj, CVL_SEMANTIC_KEY_ALREADY_EXIST 
					}
				} else {
					TRACE_LOG(INFO_API, TRACE_CREATE, "\nKey %s is deleted in same session, skipping key existence check for OP_CREATE operation", cfgDataItem.Key)
				}

				c.yp.SetOperation("CREATE")

			case OP_UPDATE:
				n, err1 := redisClient.Exists(cfgDataItem.Key).Result()
				if (err1 != nil || n == 0) { //key must exists
					TRACE_LOG(INFO_API, TRACE_UPDATE, "\nValidateEditConfig(): Key = %s does not exist", cfgDataItem.Key)
					return cvlErrObj, CVL_SEMANTIC_KEY_NOT_EXIST
				}

				c.yp.SetOperation("UPDATE")

			case OP_DELETE:
				n, err1 := redisClient.Exists(cfgDataItem.Key).Result()
				if (err1 != nil || n == 0) { //key must exists
					TRACE_LOG(INFO_API, TRACE_DELETE, "\nValidateDelete(): Key = %s does not exist", cfgDataItem.Key)
					return cvlErrObj, CVL_SEMANTIC_KEY_NOT_EXIST
				}

				c.yp.SetOperation("DELETE")
				//store deleted keys
				deletedKeys[cfgDataItem.Key] = nil;

			}

		} else if (cfgDataItem.VType == VALIDATE_NONE) {
			//Step 3.2 : Get dependent data
			/*
			tbl,key := c.addCfgDataItem(&dependentData, cfgDataItem)

			switch cfgDataItem.VOp {
			case OP_CREATE:
				//NOP
			case OP_UPDATE:
				//NOP
			case OP_DELETE:
				//update cache by removing deleted entry
				c.updateDeleteDataToCache(tbl, key)
				//store deleted keys
				deletedKeys[cfgDataItem.Key] = nil;
			}*/
		}
	}

	var depYang *yparser.YParserNode = nil
	if (len(dependentData) > 0) {
		depYang, errN = c.translateToYang1(&dependentData)
	}
	//Step 4 : Perform validation
	if cvlErrObj, cvlRetCode1 := c.validateSemantics1(yang, depYang); cvlRetCode1 != CVL_SUCCESS {
			return cvlErrObj, cvlRetCode1 
	}

	//Cache validated data
	if errObj := c.yp.CacheSubtree(false, yang); errObj.ErrCode != yparser.YP_SUCCESS {
		TRACE_LOG(INFO_API, TRACE_CACHE, "Could not cache validated data")
	}

	return cvlErrObj, CVL_SUCCESS
}

/* Fetch the Error Message from CVL Return Code. */
func GetErrorString(retCode CVLRetCode) string{

	return cvlErrorMap[retCode] 

}

/*
//Validate config data based on edit operation - xml marshalling in between
func (c *CVL) validateEditConfig(cfgData []CVLEditConfigData) CVLRetCode {
	c.clearTmpDbCache()

	leafref := c.findUsedAsLeafRef("ACL_TABLE", "name")
	TRACE_LOG(1, "%v", leafref[0])
	var nokey []string
	entry, err1 := luaScripts["find_key"].Run(redisClient, nokey, leafref[0].tableName,
	modelInfo.tableInfo[leafref[0].tableName].redisKeyDelim, leafref[0].field, "MyACL1_ACL_IPV4").Result()

	TRACE_LOG(1, "Entry = %s, %v", entry, err1)


	//Step 1: Get requested dat first
	//add all dependent data to be fetched from Redis
	requestedData := make(map[string]interface{})

        for _, cfgDataItem := range cfgData {
		if (VALIDATE_ALL != cfgDataItem.VType) {
			continue
		}

		tbl,key := c.addCfgDataItem(&requestedData, cfgDataItem)

		switch cfgDataItem.VOp {
		case OP_CREATE:
			//c.addTableDataForMustExp(tbl)

		case OP_UPDATE:
			//Get the existing data from Redis to cache, so that final validation can be done after merging this dependent data
			c.addUpdateDataToCache(tbl, key)
			c.addTableDataForMustExp(tbl)

		case OP_DELETE:
			c.updateDeleteDataToCache(tbl, key)
		}
	}

	//Validate syntax
	jsonData := ""

	jsonDataBytes, err := json.Marshal(requestedData) //Optimize TBD:
	if (err == nil) {
		jsonData = string(jsonDataBytes)
	} else {
		return CVL_SYNTAX_ERROR
	}

	TRACE_LOG(4, "JSON Data = %s\n", jsonData)
	//Step 2 : Perform syntax validation only
	doc, errN := c.translateToYang(jsonData)

	yangXml := ""
	if (errN == CVL_SUCCESS) {
		yangXml = doc.OutputXML(true)

		if errN, lydData = c.validateSyntax(yangXml); errN != CVL_SUCCESS {
			return errN
		}
	}

	//Step 3 : Check keys and update dependent data
	dependentData := make(map[string]interface{})
	deletedKeys := make(map[string]interface{}) //will store deleted key

        for _, cfgDataItem := range cfgData {

		if (cfgDataItem.VType == VALIDATE_ALL || cfgDataItem.VType == VALIDATE_SEMANTICS) {
			//Step 3.1 : Check keys
			switch cfgDataItem.VOp {
			case OP_CREATE:
				//If key is deleted and CREATE is done in same session, no need to check for key existence in case of create
				if  _, deleted := deletedKeys[cfgDataItem.Key]; deleted == false {

					//Check key should not already exist
					n, err1 := redisClient.Exists(cfgDataItem.Key).Result()
					if (err1 == nil && n > 0) {
						TRACE_LOG(1, "\nValidateEditConfig(): Key = %s alreday exists", cfgDataItem.Key)
						return CVL_SEMANTIC_KEY_ALREADY_EXIST
					}
				} else {
					TRACE_LOG(1, "\nKey %s is deleted in same session, skipping key existence check for OP_CREATE operation", cfgDataItem.Key)
				}

			case OP_UPDATE:
				n, err1 := redisClient.Exists(cfgDataItem.Key).Result()
				if (err1 != nil || n == 0) { //key must exists
					TRACE_LOG(1, "\nValidateEditConfig(): Key = %s does not exist", cfgDataItem.Key)
					return CVL_SEMANTIC_KEY_NOT_EXIST
				}

			case OP_DELETE:
				n, err1 := redisClient.Exists(cfgDataItem.Key).Result()
				if (err1 != nil || n == 0) { //key must exists
					TRACE_LOG(1, "\nValidateDelete(): Key = %s does not exist", cfgDataItem.Key)
					return CVL_SEMANTIC_KEY_NOT_EXIST
				}

				//store deleted keys
				deletedKeys[cfgDataItem.Key] = nil;

			}

		} else if (cfgDataItem.VType == VALIDATE_NONE) {
			//Step 3.2 : Get dependent data
			tbl,key := c.addCfgDataItem(&dependentData, cfgDataItem)

			switch cfgDataItem.VOp {
			case OP_CREATE:
				//NOP
			case OP_UPDATE:
				//NOP
			case OP_DELETE:
				//update cahce by removing deleted entry
				c.updateDeleteDataToCache(tbl, key)
				//store deleted keys
				deletedKeys[cfgDataItem.Key] = nil;
			}
		}
	}

	yangXmlDep := ""
	if (len(dependentData) > 0) {
		jsonDataBytes, err = json.Marshal(dependentData) //Optimize TBD:
		if (err == nil) {
			jsonData = string(jsonDataBytes)
		} else {
			return CVL_SYNTAX_ERROR
		}

		doc, errN = c.translateToYang(jsonData)
		if (errN == CVL_SUCCESS) {
			yangXmlDep = doc.OutputXML(true)

		}
	}
	//Step 4 : Perform validation
	//if errN := c.validateSemantics(lydData, yangXml); errN != CVL_SUCCESS {
	if errN := c.validateSemantics(yangXml, yangXmlDep); errN != CVL_SUCCESS {
		return errN
	}

	return CVL_SUCCESS
}


func ValidateEditConfig(cfgData []CVLEditConfigData)  (CVLErrorInfo, CVLRetCode){
	 var cvlErrObj CVLErrorInfo
	cv, ret := ValidatorSessOpen()

	if (ret != CVL_SUCCESS) {
		TRACE_LOG(1, "Failed to create validation session")
	}

	cvlErrObj,ret = cv.ValidateEditConfig1(cfgData)

	if (ValidatorSessClose(cv) != CVL_SUCCESS) {
		TRACE_LOG(1, "Failed to close validation session")
	}

	return cvlErrObj, ret
}
*/

/*
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
*/
