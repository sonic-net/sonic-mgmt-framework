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
	"strings"
	util "cvl/internal/util"
	"fmt"
	)

//Custom validation for Unnumbered interface 
func (t *CustomValidation) ValidateIpv4UnnumIntf(vc *CustValidationCtxt) CVLErrorInfo {

	if (vc.CurCfg.VOp == OP_DELETE) {
		 return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	if (vc.YNodeVal == "") {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS} //Allow empty value
	}

	tableKeys, err:= vc.RClient.Keys("LOOPBACK_INTERFACE|*").Result()

	if (err != nil) {
		util.TRACE_LEVEL_LOG(util.TRACE_SEMANTIC, "LOOPBACK_INTERFACE is empty or invalid argument")
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	count := 0
	for _, dbKey := range tableKeys {
		if (strings.Contains(dbKey, ".")) {
			count++
		}
	}

	if (count > 1) {
		util.TRACE_LEVEL_LOG(util.TRACE_SEMANTIC, "Donor interface has multiple IPv4 address")
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "LOOPBACK_INTERFACE",
			Keys: strings.Split(vc.CurCfg.Key, "|"),
			ConstraintErrMsg: fmt.Sprintf("Multiple IPv4 address configured on Donor interface. Cannot configure IP Unnumbered"),
			CVLErrDetails: "Config Validation Error",
			ErrAppTag:  "donor-multi-ipv4-addr",
		}
	}

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}
