package cvl

import (
	"encoding/json"
	"github.com/go-redis/redis"
	log "github.com/golang/glog"
	"path/filepath"
)
/*
#cgo CFLAGS: -I build/pcre-8.43/install/include -I build/libyang/build/include
#cgo LDFLAGS: -L build/pcre-8.43/install/lib -lpcre
#cgo LDFLAGS: -L build/libyang/build -lyang
#include <libyang/libyang.h>
#include <libyang/tree_data.h>
#include <stdlib.h>
#include <stdio.h>

*/
import "C"

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

//Error code
type CVLRetCode int
const (
	CVL_SUCCESS CVLRetCode = iota
	CVL_SYNTAX_ERROR
	CVL_SEMANTIC_ERROR
	CVL_ERROR
	CVL_SYNTAX_MISSING_FIELD
	CVL_SYNTAX_INVALID_FIELD   /* Invalid Field  */
	CVL_SYNTAX_INVALID_INPUT_DATA     /*Invalid Input Data */
	CVL_SYNTAX_MULTIPLE_INSTANCE     /* Multiple Field Instances */
	CVL_SYNTAX_DUPLICATE       /* Duplicate Fields  */
	CVL_SYNTAX_ENUM_INVALID  /* Invalid enum value */
	CVL_SYNTAX_ENUM_INVALID_NAME /* Invalid enum name  */
	CVL_SYNTAX_ENUM_WHITESPACE     /* Enum name with leading/trailing whitespaces */
	CVL_SYNTAX_OUT_OF_RANGE    /* Value out of range/length/pattern (data) */
	CVL_SYNTAX_MINIMUM_INVALID       /* min-elements constraint not honored  */
	CVL_SYNTAX_MAXIMUM_INVALID       /* max-elements constraint not honored */
	CVL_SEMANTIC_DEPENDENT_DATA_MISSING   /* Dependent Data is missing */
	CVL_SEMANTIC_MANDATORY_DATA_MISSING /* Mandatory Data is missing */
	CVL_SEMANTIC_KEY_ALREADY_EXIST /* Key already existing. */
	CVL_SEMANTIC_KEY_NOT_EXIST /* Key is missing. */
	CVL_SEMANTIC_KEY_DUPLICATE  /* Duplicate key. */
        CVL_SEMANTIC_KEY_INVALID
	CVL_NOT_IMPLEMENTED
	CVL_INTERNAL_UNKNOWN
	CVL_FAILURE
)

//Strcture for key and data in API
type CVLEditConfigData struct {
	VType CVLValidateType //Validation type
	VOp CVLOperation      //Operation type
	Key string      //Key format : .PORT|Ethernet4.
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
		log.Fatal(err)
	}

	ctx = C.ly_ctx_new(C.CString(CVL_SCHEMA), 0)

	modelInfo.modelNs =  make(map[string]modelNamespace) //redis table to model name
	modelInfo.tableInfo = make(map[string]modelTableInfo) //model namespace 
	dbNameToDbNum = map[string]uint8{"APPL_DB": APPL_DB, "CONFIG_DB": CONFIG_DB}
	tmpDbCache = make(map[string]interface{})

	C.ly_verb(C.LY_LLERR)

	/* schema */
	for _, modelFilePath := range modelFiles {
		_, modelFile := filepath.Split(modelFilePath)
		storeModelInfo(modelFile)

		TRACE_LOG(4, "Parsing schema file %s ...\n", modelFilePath)
		if (nil == C.lys_parse_path(ctx, C.CString(modelFilePath), C.LYS_IN_YIN)) {
			log.Fatal("Unable to parse schema file %s", modelFile)
			return CVL_ERROR
		}
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
		log.Fatal("Unable to connect with Redis Config DB")
		return CVL_ERROR
	}

	cvlInitialized = true

	return CVL_SUCCESS 
}

func Finish() {
	C.ly_ctx_destroy(ctx, nil)
}

//Validate data for Create operation
func ValidateConfig(jsonData string) CVLRetCode {
	clearTmpDbCache()

	//Convert JSON to YANG XML 
	doc, err := translateToYang(jsonData)
	if (err == CVL_SUCCESS) {
		yangXml := doc.OutputXML(true)

		return validate(yangXml)
	}

	return  err
}

//Validate edit config 
func ValidateEditConfig(cfgData []CVLEditConfigData) CVLRetCode {
	op := OP_NONE

	//find the operation
	for _, cfgDataItem := range cfgData {
		if ((cfgDataItem.VOp != OP_NONE) && (cfgDataItem.VType != VALIDATE_NONE)) {
			op = cfgDataItem.VOp
			break //assume all are same operation for now
		}
	}

	switch (op) {
	case OP_CREATE:
		return ValidateCreate(cfgData)
	case OP_UPDATE:
		return ValidateUpdate(cfgData)
	case OP_DELETE:
		return ValidateDelete(cfgData)
	}

	return CVL_ERROR
}

//Validate data for Create operation
func ValidateCreate(keyData []CVLEditConfigData) CVLRetCode {
	clearTmpDbCache()

	requestedData := keyDataToMap(VALIDATE_ALL, keyData)
	jsonData := ""

	jsonDataBytes, err := json.Marshal(requestedData) //Optimize TBD:
	if (err == nil) {
		jsonData = string(jsonDataBytes)
	} else {
		return CVL_SYNTAX_ERROR
	}

	TRACE_LOG(4, "JSON Data = %s\n", jsonData)
	//Precheck, throw error if key already exists
	//TBD: For leaf creation we can't throw error if key exists
        for tbl, kd := range requestedData {
		for key, _ := range kd.(map[string]interface{}) {
			n, err1 := redisClient.Exists(tbl+modelInfo.tableInfo[tbl].redisKeyDelim+key).Result()
			if (err1 == nil && n > 0) {
				TRACE_LOG(1, "\nValidateCreate(): Table = %s, Key = %s alreday exists", tbl, key)
				return CVL_SEMANTIC_KEY_ALREADY_EXIST
			}
		}

		addTableDataForMustExp(tbl)
	}

	//Perform syntax validation
	doc, errN := translateToYang(jsonData)
	var lydData *C.struct_lyd_node

	if (errN == CVL_SUCCESS) {
		yangXml := doc.OutputXML(true)

		if errN, lydData = validateSyntax(yangXml); errN != CVL_SUCCESS {
			return errN
		}
	}


	//TBD: Optimization - translateToYang() function stores the dependency details which are
	//populated from Redis during validateSemantics() call.
	//But dependent data 'dependentData' provided by caller is not really there in Redis,
	//so need to skip such Redis call later
	dependentData := keyDataToMap(VALIDATE_NONE, keyData)
	yangXml := ""
	if (len(dependentData) > 0) {
		jsonDataBytes, err = json.Marshal(dependentData) //Optimize TBD:
		if (err == nil) {
			jsonData = string(jsonDataBytes)
		} else {
			return CVL_SYNTAX_ERROR
		}

		doc, errN = translateToYang(jsonData)
		if (errN == CVL_SUCCESS) {
			yangXml = doc.OutputXML(true)

		}
	}
	//Perform validation
	if errN := validateSemantics(lydData, yangXml); errN != CVL_SUCCESS {
		return errN
	}

	return CVL_SUCCESS
}

//Validate data for Update operation
func ValidateUpdate(keyData []CVLEditConfigData) CVLRetCode {
	clearTmpDbCache()

	requestedData := keyDataToMap(VALIDATE_ALL, keyData)
	jsonData := ""

	jsonDataBytes, err := json.Marshal(requestedData) //Optimize TBD:
	if (err == nil) {
		jsonData = string(jsonDataBytes)
	} else {
		return CVL_SYNTAX_ERROR
	}


	TRACE_LOG(2, "JSON Data = %s\n", jsonData)
	//Precheck, throw error if key already exists
	//TBD: For leaf update we just need to check leaf syntax/semantic and key existence
        for tbl, kd := range requestedData {
		for key, _ := range kd.(map[string]interface{}) {
			n, err1 := redisClient.Exists(tbl+modelInfo.tableInfo[tbl].redisKeyDelim+key).Result()
			if (err1 != nil || n == 0) { //key must exists
				TRACE_LOG(1, "\nValidateUpdate(): Table = %s, Key = %s does not exist", tbl, key)
				return CVL_SEMANTIC_KEY_NOT_EXIST
			}

			//Get the existing data from Redis to cache, so that final validation can be done after merging this dependent data
			addUpdateDataToCache(tbl, key)
			addTableDataForMustExp(tbl)
		}
	}

	//Perform syntax validation
	doc, errN := translateToYang(jsonData)
	var lydData *C.struct_lyd_node

	if (errN == CVL_SUCCESS) {
		yangXml := doc.OutputXML(true)

		if errN, lydData = validateSyntax(yangXml); errN != CVL_SUCCESS {
			return errN
		}
	}

	dependentData := keyDataToMap(VALIDATE_NONE, keyData)
	yangXml := ""
	if (len(dependentData) > 0) {
		jsonDataBytes, err = json.Marshal(dependentData) //Optimize TBD:
		if (err == nil) {
			jsonData = string(jsonDataBytes)
		} else {
			return CVL_SYNTAX_ERROR
		}

		doc, errN = translateToYang(jsonData)
		if (errN == CVL_SUCCESS) {
			yangXml = doc.OutputXML(true)

		}
	}

	//Perform validation
	if errN := validateSemantics(lydData, yangXml); errN != CVL_SUCCESS {
		return errN
	}

	return CVL_SUCCESS
}

//Validate data for Delete operation
func ValidateDelete(keyData []CVLEditConfigData) CVLRetCode {
	clearTmpDbCache()

	requestedData := keyDataToMap(VALIDATE_ALL, keyData)
	jsonData := ""

	jsonDataBytes, err := json.Marshal(requestedData) //Optimize TBD:
	if (err == nil) {
		jsonData = string(jsonDataBytes)
	} else {
		return CVL_SYNTAX_ERROR
	}

	TRACE_LOG(2, "JSON Data = %s\n", jsonData)
	//Key must exist for deletion
	//TBD: For leaf deletion also key must exists and leaf/field should be present in DB
	//TBD: If leaf is getting deleted and has been referred somewhere else and not found in list of delete keys, should be reported as error
        for tbl, kd := range requestedData {
		for key, _ := range kd.(map[string]interface{}) {
			n, err1 := redisClient.Exists(tbl+modelInfo.tableInfo[tbl].redisKeyDelim+key).Result()
			if (err1 != nil || n == 0) { //key must exists
				TRACE_LOG(1, "\nValidateDelete(): Table = %s, Key = %s does not exist", tbl, key)
				return CVL_SEMANTIC_KEY_NOT_EXIST
			}
		}
	}

	//TBD:
	//For delete we need to delete the data from dependent data cache (simulating delete in Redis)and then check validation

	// Check if current key/leaf (in case of leaf delete - TBD) is used as reference in other tables
	// Find if it has been used in key or as a simple field
	// If it is used in key get all keys where current value is used, all such entries must be deleted as well by caller
	// If it is used in field (e.g. for ABNF hash-ref case), the entry/field(leaf delete - TBD) should be deleted by caller

	return CVL_SUCCESS
}

//Validate key only
func ValidateKeys(key []string) CVLRetCode {
	return CVL_NOT_IMPLEMENTED
}

//Validate key and data
func ValidateKeyData(key string, data string) CVLRetCode {
	return CVL_NOT_IMPLEMENTED
}

//Validate key, field and value
func ValidateFields(key string, field string, value string) CVLRetCode {
	return CVL_NOT_IMPLEMENTED
}
