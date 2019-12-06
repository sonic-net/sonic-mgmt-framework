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

var json_edit_config_create_acl_table_dependent_data = []string{`{
							"stage": "INGRESS",
							"type": "L3"
						}`}

var json_edit_config_create_acl_rule_config_data = []string{
						`{
        						"PACKET_ACTION": "FORWARD",
							"IP_TYPE":	"IPV4",
               						 "SRC_IP": "10.1.1.1/32",
                					"L4_SRC_PORT": "1909",
               						 "IP_PROTOCOL": "103",
              						  "DST_IP": "20.2.2.2/32",
               						 "L4_DST_PORT_RANGE": "9000-12000"


						}`}

var json_validate_config_data = []string{`{
					"INTERFACE": {
					"Ethernet8|10.0.0.0/31": {},
					"Ethernet12|10.0.0.2/31": {},
					"Ethernet16|10.0.0.4/31": {}
					}
				}`,
				`{
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
				}`,
                                 `{
					"CABLE_LENGTH": {
					    "AZURE": {
						"Ethernet8": "5m",
						"Ethernet12": "5m",
						"Ethernet16": "5m",
					    }
					  }
					}`}
