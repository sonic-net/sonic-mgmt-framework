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
	"fmt"
	"testing"
	"cvl"
)

func TestValidateEditConfig_Delete_Must_Check_Positive(t *testing.T) {
	depDataMap := map[string]interface{} {
		"PORT" : map[string]interface{} {
			"Ethernet3" : map[string]interface{} {
				"alias":"hundredGigE1",
				"lanes": "81,82,83,84",
				"mtu": "9100",
			},
			"Ethernet5" : map[string]interface{} {
				"alias":"hundredGigE1",
				"lanes": "85,86,87,89",
				"mtu": "9100",
			},
		},
		"ACL_TABLE" : map[string]interface{} {
			"TestACL1": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
				"ports@": "Ethernet3,Ethernet5",
			},
			"TestACL2": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
			},
		},
		"ACL_RULE" : map[string]interface{} {
			"TestACL1|Rule1": map[string] interface{} {
				"PACKET_ACTION": "FORWARD",
				"IP_TYPE": "IPV4",
				"SRC_IP": "10.1.1.1/32",
				"L4_SRC_PORT": "1909",
				"IP_PROTOCOL": "103",
				"DST_IP": "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
			"TestACL2|Rule2": map[string] interface{} {
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

	//Prepare data in Redis
	loadConfigDB(rclient, depDataMap)

	cfgDataAclRule :=  []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"ACL_RULE|TestACL2|Rule2",
			map[string]string {
			},
		},
	}

	cvSess, _ := cvl.ValidationSessOpen()

	cvlErrObj, err := cvSess.ValidateEditConfig(cfgDataAclRule)

	 cvl.ValidationSessClose(cvSess)

	if err != cvl.CVL_SUCCESS { //should not succeed
		t.Errorf("Config Validation failed. %v", cvlErrObj)
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_Delete_Must_Check_Negative(t *testing.T) {
	depDataMap := map[string]interface{} {
		"PORT" : map[string]interface{} {
			"Ethernet3" : map[string]interface{} {
				"alias":"hundredGigE1",
				"lanes": "81,82,83,84",
				"mtu": "9100",
			},
			"Ethernet5" : map[string]interface{} {
				"alias":"hundredGigE1",
				"lanes": "85,86,87,89",
				"mtu": "9100",
			},
		},
		"ACL_TABLE" : map[string]interface{} {
			"TestACL1": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
				"ports@": "Ethernet3,Ethernet5",
			},
		},
		"ACL_RULE" : map[string]interface{} {
			"TestACL1|Rule1": map[string] interface{} {
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

	//Prepare data in Redis
	loadConfigDB(rclient, depDataMap)

	cfgDataAclRule :=  []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string {
			},
		},
	}

	cvSess, _ := cvl.ValidationSessOpen()

	cvlErrObj, err := cvSess.ValidateEditConfig(cfgDataAclRule)

	 cvl.ValidationSessClose(cvSess)

	if err == cvl.CVL_SUCCESS { //should not succeed
		t.Errorf("Config Validation failed. %v", cvlErrObj)
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_Create_ErrAppTag_In_Must_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"VLAN|Vlan1001",
			map[string]string{
				"vlanid":   "102",
				"members@": "Ethernet24,Ethernet8",
			},
		},
	}

	cvSess, _ := cvl.ValidationSessOpen()

	cvlErrInfo, retCode := cvSess.ValidateEditConfig(cfgData)

	cvl.ValidationSessClose(cvSess)

	WriteToFile(fmt.Sprintf("\nCVL Error Info is  %v\n", cvlErrInfo))

	if retCode == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v %v", cvlErrInfo, retCode)
	}

}

func TestValidateEditConfig_MustExp_With_Default_Value_Positive(t *testing.T) {
	depDataMap := map[string]interface{} {
		"VLAN" : map[string]interface{} {
			"Vlan2001": map[string] interface{} {
				"vlanid":   "2001",
			},
		},
	}


	//Try to create er interface collding with vlan interface IP prefix
        cfgData := []cvl.CVLEditConfigData{
                cvl.CVLEditConfigData{
                        cvl.VALIDATE_ALL,
                        cvl.OP_CREATE,
                        "CFG_L2MC_TABLE|Vlan2001",
			map[string]string{
				"enabled":   "true",
				"query-max-response-time": "25", //default query-interval = 125
			},
                },
        }

	loadConfigDB(rclient, depDataMap)

        cvSess, _ := cvl.ValidationSessOpen()

	//Try to add second element
	cvlErrInfo, _ := cvSess.ValidateEditConfig(cfgData)


	unloadConfigDB(rclient, depDataMap)

        cvl.ValidationSessClose(cvSess)

        if cvlErrInfo.ErrCode != cvl.CVL_SUCCESS {
                t.Errorf("CFG_L2MC_TABLE creation should succeed %v", cvlErrInfo)
        }

}

func TestValidateEditConfig_MustExp_With_Default_Value_Negative(t *testing.T) {
	depDataMap := map[string]interface{} {
		"VLAN" : map[string]interface{} {
			"Vlan2002": map[string] interface{} {
				"vlanid":   "2002",
			},
		},
	}


	//Try to create er interface collding with vlan interface IP prefix
        cfgData := []cvl.CVLEditConfigData{
                cvl.CVLEditConfigData{
                        cvl.VALIDATE_ALL,
                        cvl.OP_CREATE,
                        "CFG_L2MC_TABLE|Vlan2002",
			map[string]string{
				"enabled":   "true",
				"query-interval": "9", //default query-max-response-time = 10
			},
                },
        }

	loadConfigDB(rclient, depDataMap)

        cvSess, _ := cvl.ValidationSessOpen()

	//Try to add second element
	cvlErrInfo, _ := cvSess.ValidateEditConfig(cfgData)


	unloadConfigDB(rclient, depDataMap)

        cvl.ValidationSessClose(cvSess)

        if cvlErrInfo.ErrCode == cvl.CVL_SUCCESS {
                t.Errorf("CFG_L2MC_TABLE creation should fail %v", cvlErrInfo)
        }

}

func TestValidateEditConfig_MustExp_Chained_Predicate_Positive(t *testing.T) {
	depDataMap := map[string]interface{} {
		"VLAN" : map[string]interface{} {
			"Vlan701": map[string] interface{} {
				"vlanid":   "701",
				"members@": "Ethernet20",
			},
			"Vlan702": map[string] interface{} {
				"vlanid":   "702",
				"members@": "Ethernet20,Ethernet24,Ethernet28",
			},
			"Vlan703": map[string] interface{} {
				"vlanid":   "703",
				"members@": "Ethernet20",
			},
		},
		"VLAN_MEMBER" : map[string]interface{} {
			"Vlan701|Ethernet20": map[string] interface{} {
				"tagging_mode": "tagged",
			},
			"Vlan702|Ethernet20": map[string] interface{} {
				"tagging_mode": "tagged",
			},
			"Vlan702|Ethernet24": map[string] interface{} {
				"tagging_mode": "tagged",
			},
			"Vlan702|Ethernet28": map[string] interface{} {
				"tagging_mode": "tagged",
			},
			"Vlan703|Ethernet20": map[string] interface{} {
				"tagging_mode": "tagged",
			},
		},
		"INTERFACE" : map[string]interface{} {
			"Ethernet20|1.1.1.0/32": map[string] interface{} {
				"NULL": "NULL",
			},
			"Ethernet24|1.1.2.0/32": map[string] interface{} {
				"NULL": "NULL",
			},
			"Ethernet28|1.1.2.0/32": map[string] interface{} {
				"NULL": "NULL",
			},
			"Ethernet20|1.1.3.0/32": map[string] interface{} {
				"NULL": "NULL",
			},
		},
		"VLAN_INTERFACE" : map[string]interface{} {
			"Vlan701|2.2.2.0/32": map[string] interface{} {
				"NULL": "NULL",
			},
			"Vlan701|2.2.3.0/32": map[string] interface{} {
				"NULL": "NULL",
			},
			"Vlan702|2.2.4.0/32": map[string] interface{} {
				"NULL": "NULL",
			},
			"Vlan702|2.2.5.0/32": map[string] interface{} {
				"NULL": "NULL",
			},
			"Vlan703|2.2.6.0/32": map[string] interface{} {
				"NULL": "NULL",
			},
		},
	}


	//Try to create er interface collding with vlan interface IP prefix
        cfgData := []cvl.CVLEditConfigData{
                cvl.CVLEditConfigData{
                        cvl.VALIDATE_ALL,
                        cvl.OP_CREATE,
                        "VLAN_INTERFACE|Vlan702|1.1.2.0/32",
			map[string]string{
				"NULL": "NULL",
			},
                },
        }

	loadConfigDB(rclient, depDataMap)

        cvSess, _ := cvl.ValidationSessOpen()

	//Try to add second element
	cvlErrInfo, _ := cvSess.ValidateEditConfig(cfgData)


	unloadConfigDB(rclient, depDataMap)

        cvl.ValidationSessClose(cvSess)

        if cvlErrInfo.ErrCode == cvl.CVL_SUCCESS {
                t.Errorf("INTERFACE creating failed failed -- error details %v", cvlErrInfo)
        }

}

func TestValidateEditConfig_MustExp_Within_Same_Table_Negative(t *testing.T) {
	//Try to create 
        cfgData := []cvl.CVLEditConfigData{
                cvl.CVLEditConfigData{
                        cvl.VALIDATE_ALL,
                        cvl.OP_CREATE,
                        "TAM_COLLECTOR_TABLE|Col10",
			map[string]string{
				"ipaddress-type": "ipv6", //Invalid ip address type
				"ipaddress":   "10.101.1.2",
			},
                },
        }

        cvSess, _ := cvl.ValidationSessOpen()

	cvlErrInfo, _ := cvSess.ValidateEditConfig(cfgData)

        cvl.ValidationSessClose(cvSess)

        if cvlErrInfo.ErrCode == cvl.CVL_SUCCESS {
                t.Errorf("TAM_COLLECTOR_TABLE creation should fail, %v", cvlErrInfo)
        }

}

//Check if all data is fetched for xpath without predicate
func TestValidateEditConfig_MustExp_Without_Predicate_Positive(t *testing.T) {
	depDataMap := map[string]interface{} {
		"VLAN" : map[string]interface{} {
			"Vlan201": map[string] interface{} {
				"vlanid":   "201",
				"members@": "Ethernet4,Ethernet8,Ethernet12,Ethernet16",
			},
			"Vlan202": map[string] interface{} {
				"vlanid":   "202",
				"members@": "Ethernet4",
			},
		},
	}

	//Try to create 
        cfgData := []cvl.CVLEditConfigData{
                cvl.CVLEditConfigData{
                        cvl.VALIDATE_ALL,
                        cvl.OP_CREATE,
                        "VLAN_INTERFACE|Vlan201",
			map[string]string{
				"NULL": "NULL",
			},
                },
        }

	loadConfigDB(rclient, depDataMap)

        cvSess, _ := cvl.ValidationSessOpen()

	cvlErrInfo, _ := cvSess.ValidateEditConfig(cfgData)
	cvlErrInfo, _ = cvSess.ValidateEditConfig(cfgData) //second time call should succeed also

        cvl.ValidationSessClose(cvSess)

        if cvlErrInfo.ErrCode != cvl.CVL_SUCCESS {
                t.Errorf("No predicate - config validation should succeed, %v", cvlErrInfo)
        }

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_MustExp_Non_Key_As_Predicate_Negative(t *testing.T) {
	depDataMap := map[string]interface{} {
		"VLAN" : map[string]interface{} {
			"Vlan201": map[string] interface{} {
				"vlanid":   "201",
			},
			"Vlan202": map[string] interface{} {
				"vlanid":   "202",
			},
		},
		"VXLAN_TUNNEL" : map[string]interface{} {
			"tun1": map[string] interface{} {
				"src_ip":   "10.10.1.2",
			},
		},
		"VXLAN_TUNNEL_MAP" : map[string]interface{} {
			"tun1|vmap1": map[string] interface{} {
				"vlan": "Vlan201",
				"vni": "300",
			},
		},
	}

	//Try to create 
        cfgData := []cvl.CVLEditConfigData{
                cvl.CVLEditConfigData{
                        cvl.VALIDATE_ALL,
                        cvl.OP_CREATE,
                        "VXLAN_TUNNEL_MAP|tun1|vmap2",
			map[string]string{
				"vlan": "Vlan202",
				"vni": "300", //same VNI is not valid
			},
                },
        }

	loadConfigDB(rclient, depDataMap)

        cvSess, _ := cvl.ValidationSessOpen()

	cvlErrInfo, _ := cvSess.ValidateEditConfig(cfgData)//should fail
	cvlErrInfo, _ = cvSess.ValidateEditConfig(cfgData) //should fail again
	cvlErrInfo, _ = cvSess.ValidateEditConfig(cfgData) //should fail again

        cvl.ValidationSessClose(cvSess)

        if cvlErrInfo.ErrCode == cvl.CVL_SUCCESS {
                t.Errorf("Non key as predicate - config validation should fail, %v", cvlErrInfo)
        }

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_MustExp_Non_Key_As_Predicate_In_External_Table_Positive(t *testing.T) {
	depDataMap := map[string]interface{} {
		"VLAN" : map[string]interface{} {
			"Vlan201": map[string] interface{} {
				"vlanid":   "201",
			},
			"Vlan202": map[string] interface{} {
				"vlanid":   "202",
			},
			"Vlan203": map[string] interface{} {
				"vlanid":   "203",
			},
		},
		"VXLAN_TUNNEL" : map[string]interface{} {
			"tun1": map[string] interface{} {
				"src_ip":   "10.10.1.2",
			},
		},
		"VXLAN_TUNNEL_MAP" : map[string]interface{} {
			"tun1|vmap1": map[string] interface{} {
				"vlan": "Vlan201",
				"vni": "301",
			},
			"tun1|vmap2": map[string] interface{} {
				"vlan": "Vlan202",
				"vni": "302",
			},
			"tun1|vmap3": map[string] interface{} {
				"vlan": "Vlan203",
				"vni": "303",
			},
		},
	}

	//Try to create 
        cfgData := []cvl.CVLEditConfigData{
                cvl.CVLEditConfigData{
                        cvl.VALIDATE_ALL,
                        cvl.OP_CREATE,
                        "VRF|vrf101",
			map[string]string{
				"vni": "302",
			},
                },
        }

	loadConfigDB(rclient, depDataMap)

        cvSess, _ := cvl.ValidationSessOpen()

	cvlErrInfo, _ := cvSess.ValidateEditConfig(cfgData)

        cvl.ValidationSessClose(cvSess)

        if cvlErrInfo.ErrCode != cvl.CVL_SUCCESS {
                t.Errorf("Non key as predicate in external table - config validation should succeed, %v", cvlErrInfo)
        }

	unloadConfigDB(rclient, depDataMap)
}

