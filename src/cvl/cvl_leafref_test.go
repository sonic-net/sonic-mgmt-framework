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

// EditConfig(Create) with chained leafref from redis
func TestValidateEditConfig_Create_Chained_Leafref_DepData_Positive(t *testing.T) {
	depDataMap := map[string]interface{} {
		"VLAN" : map[string]interface{} {
			"Vlan100": map[string]interface{} {
				"members@": "Ethernet1",
				"vlanid": "100",
			},
		},
		"PORT" : map[string]interface{} {
			"Ethernet1" : map[string]interface{} {
				"alias":"hundredGigE1",
				"lanes": "81,82,83,84",
				"mtu": "9100",
			},
			"Ethernet2" : map[string]interface{} {
				"alias":"hundredGigE1",
				"lanes": "85,86,87,89",
				"mtu": "9100",
			},
		},
		"ACL_TABLE" : map[string]interface{} {
			"TestACL1": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
				"ports@":"Ethernet2",
			},
		},
	}

	//Prepare data in Redis
	loadConfigDB(rclient, depDataMap)

	cvSess, _ := cvl.ValidationSessOpen()

	cfgDataVlan := []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"VLAN_MEMBER|Vlan100|Ethernet1",
			map[string]string {
				"tagging_mode" : "tagged",
			},
		},
	}

	_, err := cvSess.ValidateEditConfig(cfgDataVlan)

	if err != cvl.CVL_SUCCESS { //should succeed
		t.Errorf("Config Validation failed.")
		return
	}

	cfgDataAclRule :=  []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string {
				"PACKET_ACTION": "FORWARD",
				"IP_TYPE": "IPV4",
				"SRC_IP": "10.1.1.1/32",
				"L4_SRC_PORT": "1909",
				"IP_PROTOCOL": "103",
				"DST_IP": "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}


	_, err = cvSess.ValidateEditConfig(cfgDataAclRule)

	cvl.ValidationSessClose(cvSess)

	if err != cvl.CVL_SUCCESS { //should succeed
		t.Errorf("Config Validation failed.")
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_Create_Leafref_To_NonKey_Positive(t *testing.T) {
	depDataMap := map[string]interface{} {
		"BGP_GLOBALS" : map[string]interface{} {
			"default": map[string] interface{} {
				"router_id": "1.1.1.1",
				"local_asn": "12338",
			},
		},
	}

	//Prepare data in Redis
	loadConfigDB(rclient, depDataMap)

	cfgData :=  []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"DEVICE_METADATA|localhost",
			map[string]string {
				"vrf": "default",
				"bgp_asn": "12338",
			},
		},
	}
	cvSess, _ := cvl.ValidationSessOpen()

	_, err := cvSess.ValidateEditConfig(cfgData)

	cvl.ValidationSessClose(cvSess)

	if err != cvl.CVL_SUCCESS { //should not succeed
		t.Errorf("Leafref to non key : Config Validation failed.")
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_Update_Leafref_To_NonKey_Negative(t *testing.T) {
	depDataMap := map[string]interface{} {
		"BGP_GLOBALS" : map[string]interface{} {
			"default": map[string] interface{} {
				"router_id": "1.1.1.1",
				"local_asn": "12338",
			},
		},
	}

	//Prepare data in Redis
	loadConfigDB(rclient, depDataMap)

	cfgData :=  []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"DEVICE_METADATA|localhost",
			map[string]string {
				"vrf": "default",
				"bgp_asn": "17698",
			},
		},
	}
	cvSess, _ := cvl.ValidationSessOpen()

	_, err := cvSess.ValidateEditConfig(cfgData)

	cvl.ValidationSessClose(cvSess)

	if err == cvl.CVL_SUCCESS { //should not succeed
		t.Errorf("Leafref to non key : Config Validation failed.")
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_Create_Leafref_Multi_Key_Positive(t *testing.T) {
	depDataMap := map[string]interface{} {
		"ACL_TABLE" : map[string]interface{} {
			"TestACL901": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
			},
			"TestACL902": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
			},
		},
		"ACL_RULE" : map[string]interface{} {
			"TestACL901|Rule1": map[string] interface{} {
				"PACKET_ACTION": "FORWARD",
				"IP_TYPE": "IPV4",
				"SRC_IP": "10.1.1.1/32",
				"DST_IP": "20.2.2.2/32",
			},
			"TestACL902|Rule1": map[string] interface{} {
				"PACKET_ACTION": "FORWARD",
				"IP_TYPE": "IPV4",
				"SRC_IP": "10.1.1.2/32",
				"DST_IP": "20.2.2.4/32",
			},
		},
	}

	//Prepare data in Redis
	loadConfigDB(rclient, depDataMap)

	cfgData :=  []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"TAM_INT_IFA_FLOW_TABLE|Flow_1",
			map[string]string {
				"acl-table-name": "TestACL901",
				"acl-rule-name": "Rule1",
			},
		},
	}
	cvSess, _ := cvl.ValidationSessOpen()

	_, err := cvSess.ValidateEditConfig(cfgData)

	cvl.ValidationSessClose(cvSess)

	if err != cvl.CVL_SUCCESS { //should succeed
		t.Errorf("Leafref to unique key in multiple keys: Config Validation failed.")
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_Create_Leafref_Multi_Key_Negative(t *testing.T) {
	depDataMap := map[string]interface{} {
		"ACL_TABLE" : map[string]interface{} {
			"TestACL901": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
			},
			"TestACL902": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
			},
		},
		"ACL_RULE" : map[string]interface{} {
			"TestACL902|Rule1": map[string] interface{} {
				"PACKET_ACTION": "FORWARD",
				"IP_TYPE": "IPV4",
				"SRC_IP": "10.1.1.2/32",
				"DST_IP": "20.2.2.4/32",
			},
		},
	}

	//Prepare data in Redis
	loadConfigDB(rclient, depDataMap)

	cfgData :=  []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"TAM_INT_IFA_FLOW_TABLE|Flow_1",
			map[string]string {
				"acl-table-name": "TestACL901",
				"acl-rule-name": "Rule1", //This is not there in above depDataMap
			},
		},
	}
	cvSess, _ := cvl.ValidationSessOpen()

	_, err := cvSess.ValidateEditConfig(cfgData)

	cvl.ValidationSessClose(cvSess)

	if err == cvl.CVL_SUCCESS {
		//should not succeed
		t.Errorf("Leafref to unique key in multiple keys: Config Validation failed.")
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_Create_Leafref_With_Other_DataType_Positive(t *testing.T) {

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
			"STP_PORT|Test12", //Non-leafref
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
