package cvl

import (
	"encoding/json"
	"github.com/go-redis/redis"
	log "github.com/golang/glog"
	"path/filepath"
)
/*
#cgo LDFLAGS: -lyang
#cgo LDFLAGS: -lpcre
#include <libyang/libyang.h>
#include <libyang/tree_data.h>
#include <stdlib.h>
#include <stdio.h>

*/
import "C"


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

//Validate data for Create operation
func ValidateCreate(keyData []KeyData) CVLRetCode {
	clearTmpDbCache()

	requestedData := keyDataToMap(true, keyData)
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
				return CVL_KEY_ALREADY_EXIST
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
	dependentData := keyDataToMap(false, keyData)
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
func ValidateUpdate(keyData []KeyData) CVLRetCode {
	clearTmpDbCache()

	requestedData := keyDataToMap(true, keyData)
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
				return CVL_KEY_NOT_EXIST
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

	dependentData := keyDataToMap(false, keyData)
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
func ValidateDelete(keyData []KeyData) CVLRetCode {
	clearTmpDbCache()

	requestedData := keyDataToMap(true, keyData)
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
				return CVL_KEY_NOT_EXIST
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
