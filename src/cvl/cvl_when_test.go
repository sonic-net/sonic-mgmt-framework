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

package cvl_test

import (
	"testing"
	"cvl"
)

func TestValidateEditConfig_When_Exp_In_Choice_Negative(t *testing.T) {

	depDataMap := map[string]interface{}{
		"ACL_TABLE": map[string]interface{}{
			"TestACL1": map[string]interface{}{
				"stage": "INGRESS",
				"type":  "MIRROR",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)

	cvSess, _ := cvl.ValidationSessOpen()
	cfgDataRule := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"IP_TYPE":	     "IPV6",
				"SRC_IP":            "10.1.1.1/32", //Invalid field
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32", //Invalid field
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}


	cvlErrInfo, err := cvSess.ValidateEditConfig(cfgDataRule)

	cvl.ValidationSessClose(cvSess)

	if err == cvl.CVL_SUCCESS {
		//Should fail
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_When_Exp_In_Leaf_Positive(t *testing.T) {

	depDataMap := map[string]interface{}{
		"STP": map[string]interface{}{
			"GLOBAL": map[string]interface{}{
				"mode": "rpvst",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)
	cvSess, _ := cvl.ValidationSessOpen()

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"STP_PORT|Ethernet4",
			map[string]string{
				"enabled": "true",
				"edge_port": "true",
				"link_type": "shared",
			},
		},
	}


	cvlErrInfo, err := cvSess.ValidateEditConfig(cfgData)

	cvl.ValidationSessClose(cvSess)

	if err != cvl.CVL_SUCCESS {
		//Should succeed
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_When_Exp_In_Leaf_Negative(t *testing.T) {

	depDataMap := map[string]interface{}{
		"STP": map[string]interface{}{
			"GLOBAL": map[string]interface{}{
				"mode": "mstp",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)
	cvSess, _ := cvl.ValidationSessOpen()

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"STP_PORT|Ethernet4",
			map[string]string{
				"enabled": "true",
				"edge_port": "true",
				"link_type": "shared",
			},
		},
	}


	cvlErrInfo, err := cvSess.ValidateEditConfig(cfgData)

	cvl.ValidationSessClose(cvSess)

	if err == cvl.CVL_SUCCESS {
		//Should succeed
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, depDataMap)
}
