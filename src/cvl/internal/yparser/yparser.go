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

package yparser

/* Yang parser using libyang library */

import (
	"os"
	"strings"
	log "github.com/golang/glog"
	. "cvl/internal/util"
)

/*
#cgo LDFLAGS: -lyang
#include <libyang/libyang.h>
#include <libyang/tree_data.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

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

int lyd_data_validate_all(const char *data, const char *depData, const char *othDepData, int options, struct ly_ctx *ctx)
{
	struct lyd_node *pData;
	struct lyd_node *pDepData;
	struct lyd_node *pOthDepData;

	if ((data == NULL)  || (data[0] == '\0'))
	{
		return -1; 
	}

	pData =  lyd_parse_mem(ctx, data, LYD_XML, LYD_OPT_EDIT | LYD_OPT_NOEXTDEPS);
	if (pData == NULL) 
	{
		return -1;
	}

	if ((depData != NULL) && (depData[0] != '\0'))
	{
		if (NULL != (pDepData = lyd_parse_mem(ctx, depData, LYD_XML, LYD_OPT_EDIT | LYD_OPT_NOEXTDEPS)))
		{
			if (0 != lyd_merge_to_ctx(&pData, pDepData, LYD_OPT_DESTRUCT, ctx))
			{
				return -1;
			}
		}
	}

	if ((othDepData != NULL) && (othDepData[0] != '\0'))
	{
		if (NULL != (pOthDepData = lyd_parse_mem(ctx, othDepData, LYD_XML, LYD_OPT_EDIT | LYD_OPT_NOEXTDEPS)))
		{
			if (0 != lyd_merge_to_ctx(&pData, pOthDepData, LYD_OPT_DESTRUCT, ctx))
			{
				return -1;
			}
		}
	}

	return lyd_validate(&pData, LYD_OPT_CONFIG, ctx);
}

int lyd_multi_new_leaf(struct lyd_node *parent, const struct lys_module *module, const char *leafVal)
{
	char s[4048];
        char *name, *val;
        char *saveptr;

        strcpy(s, leafVal);

	name = strtok_r(s, "#", &saveptr);

	while (name != NULL)
	{
		val = strtok_r(NULL, "#", &saveptr);
		if (val != NULL)
		{
			if (NULL == lyd_new_leaf(parent, module, name, val))
			{
				return -1;
			}
		}

		name = strtok_r(NULL, "#", &saveptr);
	}
}

struct lyd_node *lyd_find_node(struct lyd_node *root, const char *xpath) 
{
	struct ly_set *set = NULL;
	struct lyd_node *node = NULL;

	if (root == NULL)
	{
		return NULL;
	}

	set = lyd_find_path(root, xpath);
	if (set == NULL || set->number == 0) {
		return  NULL;
	}

	node = set->set.d[0];
	ly_set_free(set);

	return node;
}

int lyd_change_leaf_data(struct lyd_node *leaf, const char *val_str)
{
  return lyd_change_leaf((struct lyd_node_leaf_list *)leaf, val_str);
}

*/
import "C"

type YParserCtx C.struct_ly_ctx
type YParserNode C.struct_lyd_node
type YParserModule C.struct_lys_module

var ypCtx *YParserCtx
var ypOpModule *YParserModule
var ypOpRoot *YParserNode  //Operation root
var ypOpNode *YParserNode  //Operation node 


type YParser struct {
	ctx *YParserCtx      //Parser context
	root *YParserNode    //Top evel root for validation
	operation string     //Edit operation
}

/* YParser Error Structure */
type YParserError struct {
	ErrCode  YParserRetCode   /* Error Code describing type of error. */
	Msg     string        /* Detailed error message. */
	ErrTxt  string        /* High level error message. */
	TableName string      /* List/Table having error */
	Keys    []string      /* Keys of the Table having error. */
        Field	string        /* Field Name throwing error . */
        Value	string        /* Field Value throwing error */
	ErrAppTag string      /* Error App Tag. */
}

type YParserRetCode int
const (
	YP_SUCCESS YParserRetCode = 1000 + iota
	YP_SYNTAX_ERROR
	YP_SEMANTIC_ERROR
	YP_SYNTAX_MISSING_FIELD
	YP_SYNTAX_INVALID_FIELD   /* Invalid Field  */
	YP_SYNTAX_INVALID_INPUT_DATA     /*Invalid Input Data */
	YP_SYNTAX_MULTIPLE_INSTANCE     /* Multiple Field Instances */
	YP_SYNTAX_DUPLICATE       /* Duplicate Fields  */
	YP_SYNTAX_ENUM_INVALID  /* Invalid enum value */
	YP_SYNTAX_ENUM_INVALID_NAME /* Invalid enum name  */
	YP_SYNTAX_ENUM_WHITESPACE     /* Enum name with leading/trailing whitespaces */
	YP_SYNTAX_OUT_OF_RANGE    /* Value out of range/length/pattern (data) */
	YP_SYNTAX_MINIMUM_INVALID       /* min-elements constraint not honored  */
	YP_SYNTAX_MAXIMUM_INVALID       /* max-elements constraint not honored */
	YP_SEMANTIC_DEPENDENT_DATA_MISSING   /* Dependent Data is missing */
	YP_SEMANTIC_MANDATORY_DATA_MISSING /* Mandatory Data is missing */
	YP_SEMANTIC_KEY_ALREADY_EXIST /* Key already existing */
	YP_SEMANTIC_KEY_NOT_EXIST /* Key is missing */
	YP_SEMANTIC_KEY_DUPLICATE  /* Duplicate key */
        YP_SEMANTIC_KEY_INVALID    /* Invalid key */
	YP_INTERNAL_UNKNOWN
)

const (
	YP_NOP = 1 + iota
	YP_OP_CREATE
	YP_OP_UPDATE
	YP_OP_DELETE
)

var yparserInitialized bool = false

func TRACE_LOG(level log.Level, tracelevel CVLTraceLevel, fmtStr string, args ...interface{}) {
	TRACE_LEVEL_LOG(level, tracelevel , fmtStr, args...)
}

func CVL_LOG(level CVLLogLevel, fmtStr string, args ...interface{}) {
	CVL_LEVEL_LOG(level, fmtStr, args...)
}

//package init function 
func init() {
	if (os.Getenv("CVL_DEBUG") != "") {
		Debug(true)
	}
}

func Debug(on bool) {
	if  (on == true) {
		C.ly_verb(C.LY_LLDBG)
	} else {
		C.ly_verb(C.LY_LLERR)
	}
}

func Initialize() {
	if (yparserInitialized != true) {
		ypCtx = (*YParserCtx)(C.ly_ctx_new(C.CString(CVL_SCHEMA), 0))
		C.ly_verb(C.LY_LLERR)
	//	yparserInitialized = true
	}
}

func Finish() {
	if (yparserInitialized == true) {
		C.ly_ctx_destroy((*C.struct_ly_ctx)(ypCtx), nil)
	//	yparserInitialized = false
	}
}

//Parse YIN schema file
func ParseSchemaFile(modelFile string) (*YParserModule, YParserError) {
	/* schema */
	TRACE_LOG(INFO_DEBUG, TRACE_YPARSER, "Parsing schema file %s ...\n", modelFile)

	module :=  C.lys_parse_path((*C.struct_ly_ctx)(ypCtx), C.CString(modelFile), C.LYS_IN_YIN)
	if module == nil {
		return nil, getErrorDetails()
	}

	if (strings.Contains(modelFile, "sonic-common.yin") == true) {
		ypOpModule = (*YParserModule)(module)
		ypOpRoot = (*YParserNode)(C.lyd_new(nil, (*C.struct_lys_module)(ypOpModule), C.CString("operation")))
		ypOpNode = (*YParserNode)(C.lyd_new_leaf((*C.struct_lyd_node)(ypOpRoot), (*C.struct_lys_module)(ypOpModule), C.CString("operation"), C.CString("NOP")))
	}

	return (*YParserModule)(module), YParserError {ErrCode : YP_SUCCESS,}
}

//Add child node to a parent node
func(yp *YParser) AddChildNode(module *YParserModule, parent *YParserNode, name string) *YParserNode {

	ret := (*YParserNode)(C.lyd_new((*C.struct_lyd_node)(parent), (*C.struct_lys_module)(module), C.CString(name)))
	if (ret == nil) {
		TRACE_LOG(INFO_DEBUG, TRACE_YPARSER, "Failed parsing node %s\n", name)
	}

	return ret
}

//Add child node to a parent node
func(yp *YParser) AddMultiLeafNodes(module *YParserModule, parent *YParserNode, multiLeaf string) YParserError {
	if (0 != C.lyd_multi_new_leaf((*C.struct_lyd_node)(parent), (*C.struct_lys_module)(module), C.CString(multiLeaf))) {
		if (Tracing == true) {
			TRACE_LOG(INFO_API, TRACE_ONERROR, "Failed to create Multi Leaf Data = %v", multiLeaf)
		}
		return getErrorDetails()
	}

	return YParserError {ErrCode : YP_SUCCESS,}

}

//Return entire subtree in XML format in string
func (yp *YParser) NodeDump(root *YParserNode) string {
	if (root == nil) {
		return ""
	} else {
		var outBuf *C.char
		C.lyd_print_mem(&outBuf, (*C.struct_lyd_node)(root), C.LYD_XML, C.LYP_WITHSIBLINGS)
		return C.GoString(outBuf)
	}
}

//Merge source with destination
func (yp *YParser) MergeSubtree(root, node *YParserNode) (*YParserNode, YParserError) {
	rootTmp := (*C.struct_lyd_node)(root)

	if (root == nil || node == nil) {
		return root, YParserError {ErrCode: YP_SUCCESS}
	}

	if (Tracing == true) {
		rootdumpStr := yp.NodeDump((*YParserNode)(rootTmp))
		TRACE_LOG(INFO_API, TRACE_YPARSER, "Root subtree = %v\n", rootdumpStr)
	}

	if (0 != C.lyd_merge_to_ctx(&rootTmp, (*C.struct_lyd_node)(node), C.LYD_OPT_DESTRUCT,
	(*C.struct_ly_ctx)(ypCtx))) {
		return (*YParserNode)(rootTmp), getErrorDetails()
	}

	if (Tracing == true) {
		dumpStr := yp.NodeDump((*YParserNode)(rootTmp))
		TRACE_LOG(INFO_API, TRACE_YPARSER, "Merged subtree = %v\n", dumpStr)
	}

	return (*YParserNode)(rootTmp), YParserError {ErrCode : YP_SUCCESS,}
}

//Cache subtree
func (yp *YParser) CacheSubtree(dupSrc bool, node *YParserNode) YParserError {
	rootTmp := (*C.struct_lyd_node)(yp.root)
	var dup *C.struct_lyd_node

	if (node == nil) {
		//nothing to merge
		return YParserError {ErrCode : YP_SUCCESS,}
	}

	if (dupSrc == true) {
		dup = C.lyd_dup_withsiblings((*C.struct_lyd_node)(node), C.LYD_DUP_OPT_RECURSIVE | C.LYD_DUP_OPT_NO_ATTR)
	} else {
		dup = (*C.struct_lyd_node)(node)
	}

	if (yp.root != nil) {
		if (0 != C.lyd_merge_to_ctx(&rootTmp, (*C.struct_lyd_node)(dup), C.LYD_OPT_DESTRUCT,
		(*C.struct_ly_ctx)(ypCtx))) {
			return getErrorDetails()
		}
	} else {
		yp.root = (*YParserNode)(dup)
	}

	if (Tracing == true) {
		dumpStr := yp.NodeDump((*YParserNode)(rootTmp))
		TRACE_LOG(INFO_API, TRACE_YPARSER, "Cached subtree = %v\n", dumpStr)
	}

	return YParserError {ErrCode : YP_SUCCESS,}
}

func (yp *YParser) DestroyCache() YParserError {

	if (yp.root != nil) {
		C.lyd_free_withsiblings((*C.struct_lyd_node)(yp.root))
		yp.root = nil
	}

	return YParserError {ErrCode : YP_SUCCESS,}
}

//Set operation 
func (yp *YParser) SetOperation(op string) YParserError {
	if (ypOpNode == nil) {
		return YParserError {ErrCode : YP_INTERNAL_UNKNOWN,}
	}

	if (0 != C.lyd_change_leaf_data((*C.struct_lyd_node)(ypOpNode), C.CString(op))) {
		return YParserError {ErrCode : YP_INTERNAL_UNKNOWN,}
	}

	yp.operation = op
	return YParserError {ErrCode : YP_SUCCESS,}
}

//Validate config - syntax and semantics
func (yp *YParser) ValidateData(data, depData *YParserNode) YParserError {

	var dataRoot *YParserNode

	if (depData != nil) {
		if dataRoot, _ = yp.MergeSubtree(data, depData); dataRoot == nil {
			CVL_LOG(ERROR, "Failed to merge dependent data\n")
			return getErrorDetails()
		}
	}

	dataRootTmp := (*C.struct_lyd_node)(dataRoot)

	if (0 != C.lyd_data_validate(&dataRootTmp, C.LYD_OPT_CONFIG, (*C.struct_ly_ctx)(ypCtx))) {
		if (Tracing == true) {
			strData := yp.NodeDump((*YParserNode)(dataRootTmp))
			TRACE_LOG(INFO_API, TRACE_ONERROR, "Failed to validate data = %v", strData)
		}

		CVL_LOG(ERROR, "Validation failed\n")
		return getErrorDetails()
	}

	return YParserError {ErrCode : YP_SUCCESS,}
}

//Perform syntax checks
func (yp *YParser) ValidateSyntax(data *YParserNode) YParserError {
	dataTmp := (*C.struct_lyd_node)(data)

	//Just validate syntax
	if (0 != C.lyd_data_validate(&dataTmp, C.LYD_OPT_EDIT | C.LYD_OPT_NOEXTDEPS,
	(*C.struct_ly_ctx)(ypCtx))) {
		if (Tracing == true) {
			strData := yp.NodeDump((*YParserNode)(dataTmp))
			TRACE_LOG(INFO_API, TRACE_ONERROR, "Failed to validate Syntax, data = %v", strData)
		}
		return  getErrorDetails()
	}
		 //fmt.Printf("Error Code from libyang is %d\n", C.ly_errno) 

	return YParserError {ErrCode : YP_SUCCESS,}
}

//Perform semantic checks 
func (yp *YParser) ValidateSemantics(data, depData, appDepData *YParserNode) YParserError {

	var dataTmp *C.struct_lyd_node

	if (data != nil) {
		dataTmp = (*C.struct_lyd_node)(data)
	} else if (depData != nil) {
		dataTmp = (*C.struct_lyd_node)(depData)
	} else if (yp.root != nil) {
		dataTmp = (*C.struct_lyd_node)(yp.root)
	} else {
		if (yp.operation == "CREATE") || (yp.operation == "UPDATE") {
			return YParserError {ErrCode : YP_INTERNAL_UNKNOWN,}
		} else {
			return YParserError {ErrCode : YP_SUCCESS,}
		}
	}

	//parse dependent data
	if (data != nil && depData != nil) {

		//merge input data and dependent data for semantic validation
		if (0 != C.lyd_merge_to_ctx(&dataTmp, (*C.struct_lyd_node)(depData),
		C.LYD_OPT_DESTRUCT, (*C.struct_ly_ctx)(ypCtx))) {
			TRACE_LOG(INFO_API, (TRACE_SEMANTIC | TRACE_LIBYANG), "Unable to merge dependent data\n")
			return getErrorDetails()
		}
	}

	//Merge cached data
	if ((data != nil || depData != nil) && yp.root != nil) {
		if (0 != C.lyd_merge_to_ctx(&dataTmp, (*C.struct_lyd_node)(yp.root),
		0, (*C.struct_ly_ctx)(ypCtx))) {
			TRACE_LOG(INFO_API, (TRACE_SEMANTIC | TRACE_LIBYANG), "Unable to merge cached dependent data\n")
			return getErrorDetails()
		}
	}

	//Merge appDepData
	if (appDepData != nil) {
		if (0 != C.lyd_merge_to_ctx(&dataTmp, (*C.struct_lyd_node)(appDepData),
		C.LYD_OPT_DESTRUCT, (*C.struct_ly_ctx)(ypCtx))) {
			TRACE_LOG(INFO_API, (TRACE_SEMANTIC | TRACE_LIBYANG), "Unable to merge other dependent data\n")
			return getErrorDetails()
		}
	}

	//Add operation for constraint check
	if (ypOpRoot != nil) {
		//if (0 != C.lyd_insert_sibling(&dataTmp, (*C.struct_lyd_node)(ypOpRoot))) {
		if (0 != C.lyd_merge_to_ctx(&dataTmp, (*C.struct_lyd_node)(ypOpRoot),
		0, (*C.struct_ly_ctx)(ypCtx))) {
			TRACE_LOG(INFO_API, (TRACE_SEMANTIC | TRACE_LIBYANG), "Unable to insert operation node")
			return getErrorDetails()
		}
	}

	if (Tracing == true) {
		strData := yp.NodeDump((*YParserNode)(dataTmp))
		TRACE_LOG(INFO_API, TRACE_YPARSER, "Semantics data = %v", strData)
	}

	//Check semantic validation
	if (0 != C.lyd_data_validate(&dataTmp, C.LYD_OPT_CONFIG, (*C.struct_ly_ctx)(ypCtx))) {
		if (Tracing == true) {
			strData1 := yp.NodeDump((*YParserNode)(dataTmp))
			TRACE_LOG(INFO_API, TRACE_ONERROR, "Failed to validate Semantics, data = %v", strData1)
		}
		return getErrorDetails()
	}

	return YParserError {ErrCode : YP_SUCCESS,}
}

func (yp *YParser) FreeNode(node *YParserNode) YParserError {

	C.lyd_free_withsiblings((*C.struct_lyd_node)(node))

	return YParserError {ErrCode : YP_SUCCESS,}
}

/* This function translates LIBYANG error code to valid YPARSER error code. */
func translateLYErrToYParserErr(LYErrcode int) YParserRetCode {
	var ypErrCode YParserRetCode

	switch LYErrcode {
		case C.LYVE_SUCCESS:  /**< no error */
			ypErrCode = YP_SUCCESS
		case C.LYVE_XML_MISS, C.LYVE_INARG, C.LYVE_MISSELEM: /**< missing XML object */
			ypErrCode = YP_SYNTAX_MISSING_FIELD
		case C.LYVE_XML_INVAL, C.LYVE_XML_INCHAR, C.LYVE_INMOD, C.LYVE_INELEM , C.LYVE_INVAL, C.LYVE_MCASEDATA:/**< invalid XML object */
			ypErrCode = YP_SYNTAX_INVALID_FIELD
		case C.LYVE_EOF, C.LYVE_INSTMT,  C.LYVE_INPAR,  C.LYVE_INID,  C.LYVE_MISSSTMT, C.LYVE_MISSARG:   /**< invalid statement (schema) */
			ypErrCode = YP_SYNTAX_INVALID_INPUT_DATA
		case C.LYVE_TOOMANY:  /**< too many instances of some object */
			ypErrCode = YP_SYNTAX_MULTIPLE_INSTANCE
		case C.LYVE_DUPID,  C.LYVE_DUPLEAFLIST, C.LYVE_DUPLIST, C.LYVE_NOUNIQ:/**< duplicated identifier (schema) */
			ypErrCode = YP_SYNTAX_DUPLICATE
		case C.LYVE_ENUM_INVAL:    /**< invalid enum value (schema) */
			ypErrCode = YP_SYNTAX_ENUM_INVALID
		case C.LYVE_ENUM_INNAME:   /**< invalid enum name (schema) */
			ypErrCode = YP_SYNTAX_ENUM_INVALID_NAME
		case C.LYVE_ENUM_WS:  /**< enum name with leading/trailing whitespaces (schema) */
			ypErrCode = YP_SYNTAX_ENUM_WHITESPACE
		case C.LYVE_KEY_NLEAF,  C.LYVE_KEY_CONFIG, C.LYVE_KEY_TYPE : /**< list key is not a leaf (schema) */
			ypErrCode = YP_SEMANTIC_KEY_INVALID
		case C.LYVE_KEY_MISS, C.LYVE_PATH_MISSKEY: /**< list key not found (schema) */
			ypErrCode = YP_SEMANTIC_KEY_NOT_EXIST
		case C.LYVE_KEY_DUP:  /**< duplicated key identifier (schema) */
			ypErrCode = YP_SEMANTIC_KEY_DUPLICATE
		case C.LYVE_NOMIN:/**< min-elements constraint not honored (data) */
			ypErrCode = YP_SYNTAX_MINIMUM_INVALID
		case C.LYVE_NOMAX:/**< max-elements constraint not honored (data) */
			ypErrCode = YP_SYNTAX_MAXIMUM_INVALID
		case C.LYVE_NOMUST, C.LYVE_NOWHEN, C.LYVE_INWHEN, C.LYVE_NOLEAFREF :   /**< unsatisfied must condition (data) */
			ypErrCode = YP_SEMANTIC_DEPENDENT_DATA_MISSING
		case C.LYVE_NOMANDCHOICE:/**< max-elements constraint not honored (data) */
			ypErrCode = YP_SEMANTIC_MANDATORY_DATA_MISSING
		case C.LYVE_PATH_EXISTS:   /**< target node already exists (path) */
			ypErrCode = YP_SEMANTIC_KEY_ALREADY_EXIST
		default:
			ypErrCode = YP_INTERNAL_UNKNOWN

	}
	return ypErrCode
}

/* This function performs parsing and processing of LIBYANG error messages. */
func getErrorDetails() YParserError {
	var key []string
	var errtableName string
	var ElemVal string
	var errMessage string
	var ElemName string
	var errText string
	var msg string
	var ypErrCode YParserRetCode =  YP_INTERNAL_UNKNOWN
	var errMsg, errPath, errAppTag string 

	ctx := (*C.struct_ly_ctx)(ypCtx)
	ypErrFirst := C.ly_err_first(ctx);


	if (ypErrFirst == nil) {
               return  YParserError {
                       TableName : errtableName,
                       ErrCode : ypErrCode,
                       Keys    : key,
                       Value : ElemVal,
                       Field : ElemName,
                       Msg        :  errMessage,
                       ErrTxt: errText,
                       ErrAppTag: errAppTag,
               }
       }


	if ((ypErrFirst != nil) && ypErrFirst.prev.no == C.LY_SUCCESS) {
		return YParserError {
			ErrCode : YP_SUCCESS,
		}
	}

	if (ypErrFirst != nil) {
	       errMsg = C.GoString(ypErrFirst.prev.msg)
	       errPath = C.GoString(ypErrFirst.prev.path)
	       errAppTag = C.GoString(ypErrFirst.prev.apptag)
	}


	/* Example error messages. 	
	1. Leafref "/sonic-port:sonic-port/sonic-port:PORT/sonic-port:ifname" of value "Ethernet668" points to a non-existing leaf. 
	(path: /sonic-interface:sonic-interface/INTERFACE[portname='Ethernet668'][ip_prefix='10.0.0.0/31']/portname)
	2. A vlan interface member cannot be part of portchannel which is already a vlan member
	(path: /sonic-vlan:sonic-vlan/VLAN[name='Vlan1001']/members[.='Ethernet8'])
	3. Value "ch1" does not satisfy the constraint "Ethernet([1-3][0-9]{3}|[1-9][0-9]{2}|[1-9][0-9]|[0-9])" (range, length, or pattern). 
	(path: /sonic-vlan:sonic-vlan/VLAN[name='Vlan1001']/members[.='ch1'])*/ 


	/* Fetch the TABLE Name which are in CAPS. */
	resultTable := strings.SplitN(errPath, "[", 2)
	if (len(resultTable) >= 2) {
		resultTab := strings.Split(resultTable[0], "/")
		errtableName = resultTab[len(resultTab) -1]

		/* Fetch the Error Elem Name. */
		resultElem := strings.Split(resultTable[1], "/")
		ElemName = resultElem[len(resultElem) -1]
	}

	/* Fetch the invalid field name. */
	result := strings.Split(errMsg, "\"")
	if (len(result) > 1) {
		for i := range result {
			if (strings.Contains(result[i], "value")) ||
			(strings.Contains(result[i], "Value")) {
				ElemVal = result[i+1]
			}
		}
	} else if (len(result) == 1) {
		/* Custom contraint error message like in must statement. 
		This can be used by App to display to user.
		*/
		errText = errMsg
	}

	// Find key elements 
	resultKey := strings.Split(errPath, "=")
	for i := range resultKey {
		if (strings.Contains(resultKey[i], "]")) {
			newRes := strings.Split(resultKey[i], "]")
			key = append(key, newRes[0])
		}
	}

	/* Form the error message. */
	msg = "["
	for _, elem := range key {
		msg = msg + elem + " "
	}
	msg = msg + "]"

	/* For non-constraint related errors , print below error message. */
	if (len(result) > 1) {
		errMessage = errtableName + " with keys" + msg + " has field " +
		ElemName + " with invalid value " + ElemVal
	}else {
		/* Dependent data validation error. */
		errMessage = "Dependent data validation failed for table " +
		errtableName + " with keys" + msg
	}


	if (C.ly_errno == C.LY_EVALID) {  //Validation failure
		ypErrCode =  translateLYErrToYParserErr(int(ypErrFirst.prev.vecode))
		
	} else {
		switch (C.ly_errno) {
		case C.LY_EMEM:
			errText = "Memory allocation failure"

		case C.LY_ESYS:
			errText = "System call failure"

		case C.LY_EINVAL:
			errText = "Invalid value"

		case C.LY_EINT:
			errText = "Internal error"

		case C.LY_EPLUGIN:
			errText = "Error reported by a plugin"
		}
	}

	errObj := YParserError {
		TableName : errtableName,
		ErrCode : ypErrCode,
		Keys    : key,
		Value : ElemVal,
		Field : ElemName,
		Msg        :  errMessage,
		ErrTxt: errText,
		ErrAppTag: errAppTag,
	}

	TRACE_LOG(INFO_API, TRACE_YPARSER, "YParser error details: %v...", errObj)

	return  errObj
}

func FindNode(root *YParserNode, xpath string) *YParserNode {
	return  (*YParserNode)(C.lyd_find_node((*C.struct_lyd_node)(root), C.CString(xpath)))
}
