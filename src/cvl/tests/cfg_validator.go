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

package main

import (
	"fmt"
	"os"
	"time"
	"io/ioutil"
	"cvl"
)

func main() {
	jsonData :=`{
		"VLAN": {
			"Vlan100": {
				"members": [
				"Ethernet44",
				"Ethernet64"
				],
				"vlanid": "100"
			},
			"Vlan1200": {
				"members": [
				"Ethernet64",
				"Ethernet8"
				],
				"vlanid": "1200"
			},
			"Vlan2500": {
				"members": [
				"Ethernet8",
				"Ethernet64"
				],
				"vlanid": "2500"
			}
		}
	}`
	/*
		"ACL_TABLE": {
			"TestACL1": {
				"stage": "INGRESS",
				"type": "l3"
			}
		},
		"ACL_RULE": {
			"TestACL1|Rule1": {
				"packet_action": "forward",
				"src_ip": "10.1.1.1/32",
				"l4_src_port": "ABC",
				"ip_protocol": "ip",
				"dst_ip": "20.2.2.2/32",
				"l4_dst_port_range": "9000-12000",
				"mirror_action" : "mirror1"
			}
		}*/

/*jsonData :=  `{
		  "DEVICE_METADATA": {
        "localhost": {
        "hwsku": "Force10-S6100",
        "default_bgp_status": "up",
        "docker_routing_config_mode": "unified",
        "hostname": "sonic-s6100-01",
        "platform": "x86_64-dell_s6100_c2538-r0",
        "mac": "4c:76:25:f4:70:82",
        "default_pfcwd_status": "disable",
        "deployment_id": "1",
        "type": "ToRRouter"
    }
  }
	  }`*/
/*jsonData :=  `{
		  "DEVICE_NEIGHBOR": {
        "ARISTA04T1": {
                "mgmt_addr": "10.20.0.163",
                "hwsku": "Arista",
		"lo_addr": "2.2.2.2",
                "local_port": "Ethernet124",
                "type": "LeafRouter",
                "port": "Ethernet68"
        }
	  }
	  }`*/
/*jsonData :=  `{
		  "BGP_NEIGHBOR": {
        "10.0.0.61": {
            "local_addr": "10.0.0.60",
            "asn": 64015,
            "name": "ARISTA15T0"
        }
		  }
	  }`*/

/* jsonData :=  `{
		  "INTERFACE": {
        "Ethernet68|10.0.0.0/31": {},
        "Ethernet24|10.0.0.2/31": {},
        "Ethernet112|10.0.0.4/31": {}
     }
	  }`*/

/*jsonData :=  `{
		  "INTERFACE": {
        "Ethernet68|10.0.0.0/31": {},
        "Ethernet24|10.0.0.2/31": {},
        "Ethernet112|10.0.0.4/31": {}
     }
	  }`*/
/*jsonData :=  `{
		  "PORTCHANNEL_INTERFACE": {
        "PortChannel01|10.0.0.56/31": {},
        "PortChannel01|FC00::71/126": {},
        "PortChannel02|10.0.0.58/31": {},
        "PortChannel02|FC00::75/126": {}
    }

	  }`*/
/*jsonData :=  `{
  "VLAN_INTERFACE": {
        "Vlan1000|192.168.0.1/27": {}
    }
  }`*/
	start := time.Now()

	dataFile := ""
	if (len(os.Args) >= 2) {
		if (os.Args[1] == "debug") {
			cvl.Debug(true)
		} else {
			dataFile =  os.Args[1]
		}
	}
	if (len(os.Args) == 3) {
		dataFile = os.Args[2]
	}

	//cvl.Initialize()

	b, e := ioutil.ReadFile(dataFile)
	if e != nil {
		fmt.Printf("\nFailed to read data file : %v\n", e)
	} else {
		jsonData = string(b)
	}


	cv, ret := cvl.ValidationSessOpen()
	if (ret != cvl.CVL_SUCCESS) {
		fmt.Printf("NewDB: Could not create CVL session")
		return
	}

	err := cv.ValidateConfig(jsonData)

	fmt.Printf("\nValidating data = %v\n\n", jsonData);

	if (err == cvl.CVL_SUCCESS) {
		fmt.Printf("\nConfig Validation succeeded.\n\n");
	} else {
		fmt.Printf("\nConfig Validation failed.\n\n");
	}

	keyData := make([]cvl.CVLEditConfigData, 4)
	keyData[0].VType = cvl.VALIDATE_NONE
	keyData[0].VOp = cvl.OP_NONE
	keyData[0].Key = "ACL_TABLE|MyACL55_ACL_IPV4"
	keyData[0].Data = make(map[string]string)
	keyData[0].Data["stage"] =  "INGRESS"
	keyData[0].Data["type"] =  "l3"

	keyData[1].VType = cvl.VALIDATE_NONE
	keyData[1].VOp = cvl.OP_NONE
	keyData[1].Key = "ACL_RULE|MyACL55_ACL_IPV4|RULE_1"
	keyData[1].Data = make(map[string]string)
	keyData[1].Data["packet_action"] = "forward"
	keyData[1].Data["ip_protocol"] = "ip"
	keyData[1].Data["src_ip"] = "10.1.1.1/32"
	keyData[1].Data["dst_ip"] = "20.2.2.2/32"

	keyData[2].VType = cvl.VALIDATE_NONE
	keyData[2].VOp = cvl.OP_NONE
	keyData[2].Key = "ACL_TABLE|MyACL11_ACL_IPV4"
	keyData[2].Data = make(map[string]string)
	keyData[2].Data["stage"] =  "INGRESS"

	keyData[3].VType = cvl.VALIDATE_ALL
	keyData[3].VOp = cvl.OP_CREATE
	keyData[3].Key = "VLAN|Vlan901"
	keyData[3].Data = make(map[string]string)
	keyData[3].Data["members"] =  "Ethernet8"
	keyData[3].Data["vlanid"] =  "901"

	_, ret = cv.ValidateEditConfig(keyData)
	fmt.Printf("\n\n\n  cvl.ValidateEditConfig() = %d\n", ret)

	keyData1 := make([]cvl.CVLEditConfigData, 3)
	keyData1[0].VType = cvl.VALIDATE_NONE
	keyData1[0].VOp = cvl.OP_NONE
	keyData1[0].Key = "ACL_TABLE|MyACL11_ACL_IPV4"
	keyData1[0].Data = make(map[string]string)
	keyData1[0].Data["stage"] =  "INGRESS"
	keyData1[0].Data["type"] =  "l3"

	keyData1[1].VType = cvl.VALIDATE_NONE
	keyData1[1].VOp = cvl.OP_NONE
	keyData1[1].Key = "ACL_RULE|MyACL11_ACL_IPV4|RULE_1"
	keyData1[1].Data = make(map[string]string)
	keyData1[1].Data["packet_action"] = "forward"
	keyData1[1].Data["ip_protocol"] = "ip"
	keyData1[1].Data["src_ip"] = "10.1.1.1/32"
	keyData1[1].Data["dst_ip"] = "20.2.2.2/32"

	keyData1[2].VType = cvl.VALIDATE_ALL
	keyData1[2].VOp = cvl.OP_UPDATE
	keyData1[2].Key = "ACL_TABLE|MyACL33_ACL_IPV4"
	keyData1[2].Data = make(map[string]string)
	keyData1[2].Data["stage"] =  "INGRESS"

	_, ret = cv.ValidateEditConfig(keyData)
	fmt.Printf("\n\n\n  cvl.ValidateEditConfig() = %d\n", ret)


	keyData2 := make([]cvl.CVLEditConfigData, 3)
	keyData2[0].VType = cvl.VALIDATE_ALL
	keyData2[0].VOp = cvl.OP_DELETE
	keyData2[0].Key = "ACL_TABLE|MyACL11_ACL_IPV4"
	keyData2[0].Data = make(map[string]string)

	keyData2[1].VType = cvl.VALIDATE_ALL
	keyData2[1].VOp = cvl.OP_DELETE
	keyData2[1].Key = "ACL_RULE|MyACL11_ACL_IPV4|RULE_1"
	keyData2[1].Data = make(map[string]string)

	keyData2[2].VType = cvl.VALIDATE_ALL
	keyData2[2].VOp = cvl.OP_DELETE
	keyData2[2].Key = "ACL_TABLE|MyACL33_ACL_IPV4"
	keyData2[2].Data = make(map[string]string)

	_, ret = cv.ValidateEditConfig(keyData)
	fmt.Printf("\n\n\n  cvl.ValidateEditConfig() = %d\n", ret)


	cvl.ValidationSessClose(cv)
	cvl.Finish()
	fmt.Printf("\n\n\n Time taken = %v\n", time.Since(start))

	stopChan := make(chan int, 1)
	for {
		select {
		case <- stopChan:
		}
	}


}
