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
	"github.com/go-redis/redis"
	"strings"
	)

//Custom validation code for sonic-acl.yang//
/////////////////////////////////////////////
//Path : /sonic-acl/ACL_TABLE/ACL_TABLE_LIST
//Purpose: Allow maximum 1024 ACL tables
//vc : Custom Validation Context
//Returns -  CVL Error object
const MAX_ACL_TABLE_INSTANCES = 1024
func (t *CustomValidation) ValidateMaxAclTable(
	vc *CustValidationCtxt) CVLErrorInfo {

	var nokey []string
	ls := redis.NewScript(`return #redis.call('KEYS', "ACL_TABLE|*")`)

	//Get current coutnt from Redis
	redisEntries, err := ls.Run(vc.RClient, nokey).Result()
	if err != nil {
		return CVLErrorInfo{ErrCode: CVL_SEMANTIC_ERROR}
	}

	aclTblCount := int(redisEntries.(int64))
	//Get count from user request
	for idx := 0; idx < len(vc.ReqData); idx++ {
		if (vc.ReqData[idx].VOp == OP_CREATE) &&
		(strings.HasPrefix(vc.ReqData[idx].Key, "ACL_TABLE|")) {
			aclTblCount = aclTblCount + 1
		}
	}

	if (aclTblCount > MAX_ACL_TABLE_INSTANCES) {
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "ACL_TABLE",
			CVLErrDetails : "Max table count exceeded",
		}
	}

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}

//Path : /sonic-acl/ACL_RULE/ACL_RULE_LIST/IP_TYPE
//Purpose: Check correct for IP address provided
//         based on type IP_TYPE
//vc : Custom Validation Context
//Returns -  CVL Error object
func (t *CustomValidation) ValidateAclRuleIPAddress(
	vc *CustValidationCtxt) CVLErrorInfo {

	if (vc.CurCfg.VOp == OP_DELETE) || (vc.CurCfg.VOp == OP_UPDATE) {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	if (vc.YNodeVal == "") {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	if  (vc.YNodeVal == "ANY" || vc.YNodeVal == "IP" ||
	vc.YNodeVal == "IPV4" || vc.YNodeVal == "IPV4ANY") {

		_, srcIpV4exists := vc.CurCfg.Data["SRC_IP"]
		_, dstIpV4exists := vc.CurCfg.Data["DST_IP"]

		if (srcIpV4exists == false) || (dstIpV4exists == false) {
			return CVLErrorInfo{
				ErrCode: CVL_SEMANTIC_ERROR,
				TableName: "ACL_RULE",
				CVLErrDetails : "IP address is missing for " +
				"IP_TYPE=" + vc.YNodeVal,
			}
		}

	} else if  (vc.YNodeVal == "ANY" || vc.YNodeVal == "IP" ||
	vc.YNodeVal == "IPV6" || vc.YNodeVal == "IPV6ANY") {

		_, srcIpV6exists := vc.CurCfg.Data["SRC_IPV6"]
		_, dstIpV6exists := vc.CurCfg.Data["DST_IPV6"]

		if (srcIpV6exists == false) || (dstIpV6exists == false) {
			return CVLErrorInfo{
				ErrCode: CVL_SEMANTIC_ERROR,
				TableName: "ACL_RULE",
				CVLErrDetails : "IP address is missing for " +
				"IP_TYPE=" + vc.YNodeVal,
			}
		}
	}

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}

