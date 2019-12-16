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

package custom_validation

import (
	"reflect"
	"github.com/antchfx/xmlquery"
	"github.com/go-redis/redis"
	. "cvl/internal/util"
	"cvl/internal/yparser"
	)

type CustomValidation struct {}

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

//Custom validation context passed to custom validation function 
type CustValidationCtxt struct {
	ReqData []CVLEditConfigData //All request data
	CurCfg *CVLEditConfigData //Current request data for which validation should be done
	YNodeName string //YANG node name
	YNodeVal string  //YANG node value, leaf-list will have "," separated value
	YCur *xmlquery.Node //YANG data tree
	RClient *redis.Client //Redis client
}

//Common function to invoke custom validation
//TBD should we do this using GO plugin feature ?
func InvokeCustomValidation(cv *CustomValidation, name string, args... interface{}) CVLErrorInfo {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	f := reflect.ValueOf(cv).MethodByName(name)
	if (f.IsNil() == false) {
		v := f.Call(inputs)
		TRACE_LEVEL_LOG(TRACE_SEMANTIC,
		"InvokeCustomValidation: %s(), return value = %v", v[0])

		return (v[0].Interface()).(CVLErrorInfo)
	}

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}

