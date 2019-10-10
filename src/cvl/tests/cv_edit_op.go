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
	"time"
	"os"
	"cvl"
)

func main() {
	start := time.Now()
	count := 0

	cvl.Initialize()
	cv, _ := cvl.ValidationSessOpen()

	if ((len(os.Args) > 1) && (os.Args[1] == "debug")) {
		cvl.Debug(true)
	}

	{
		count++

		cfgData :=  []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_NONE,
				cvl.OP_NONE,
				"ACL_TABLE|TestACL1",
				map[string]string {
					"stage": "INGRESS",
					"type": "L3",
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_CREATE,
				"ACL_RULE|TestACL1|Rule1",
				map[string]string {
					"PACKET_ACTION": "FORWARD",
					"SRC_IP": "10.1.1.1/32",
					"L4_SRC_PORT": "1909",
					"IP_PROTOCOL": "103",
					"DST_IP": "20.2.2.2/32",
					"L4_DST_PORT_RANGE": "9000-12000",
				},
			},
		}

		fmt.Printf("\n\n%d. Validating create data = %v\n\n", count, cfgData);

		_, err := cv.ValidateEditConfig(cfgData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++

		cfgData :=  []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_UPDATE,
				"ACL_TABLE|MyACL11_ACL_IPV4",
				map[string]string {
					"stage": "INGRESS",
					"type": "MIRROR",
				},
			},
		}

		fmt.Printf("\n\n%d. Validating update data = %v\n\n", count, cfgData);

		_, err := cv.ValidateEditConfig(cfgData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		cfgData :=  []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_CREATE,
				"MIRROR_SESSION|everflow",
				map[string]string {
					"src_ip": "10.1.0.32",
					"dst_ip": "2.2.2.2",
				},
			},
		}

		fmt.Printf("\n\n%d. Validating create data = %v\n\n", count, cfgData);
		_, err := cv.ValidateEditConfig(cfgData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}

		count++
		cfgData = []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_NONE,
				cvl.OP_NONE,
				"MIRROR_SESSION|everflow",
				map[string]string {
					"src_ip": "10.1.0.32",
					"dst_ip": "2.2.2.2",
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_UPDATE,
				"ACL_RULE|MyACL11_ACL_IPV4|RULE_1",
				map[string]string {
					"MIRROR_ACTION": "everflow",
				},
			},
		}

		fmt.Printf("\n\n%d. Validating data for update = %v\n\n", count, cfgData);

		_, err = cv.ValidateEditConfig(cfgData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++

		cfgData :=  []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_DELETE,
				"MIRROR_SESSION|everflow",
				map[string]string {
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_DELETE,
				"ACL_RULE|MyACL11_ACL_IPV4|RULE_1",
				map[string]string {
				},
			},
		}

		fmt.Printf("\n\n%d. Validating data for delete = %v\n\n", count, cfgData);

		_, err := cv.ValidateEditConfig(cfgData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}

	cvl.ValidationSessClose(cv)
	cvl.Finish()

	fmt.Printf("\n\n\n Time taken for %v requests = %v\n", count, time.Since(start))
}
