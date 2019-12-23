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
	"github.com/go-redis/redis"
	"strconv"
)

func getConfigDbClient() *redis.Client {
	rclient := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		Password:    "", // no password set
		DB:          4,
		DialTimeout: 0,
	})
	_, err := rclient.Ping().Result()
	if err != nil {
		fmt.Printf("failed to connect to redis server %v", err)
	}
	return rclient
}

/* Unloads the Config DB based on JSON File. */
func unloadConfigDB(rclient *redis.Client, key string, data map[string]string) {
	_, err := rclient.Del(key).Result()

	if err != nil {
		fmt.Printf("Failed to delete for key %s, data %v, err %v", key, data, err)
	}

}

/* Loads the Config DB based on JSON File. */
func loadConfigDB(rclient *redis.Client, key string, data map[string]string) {

	dataTmp  := make(map[string]interface{})

	for k, v :=  range data {
		dataTmp[k] = v
	}

	_, err := rclient.HMSet(key, dataTmp).Result()

	if err != nil {
		fmt.Printf("Failed to add for key %s, data %v, err %v", key, data, err)
	}

}

func main() {
	start := time.Now()
	count := 0

	cvl.Initialize()

	if ((len(os.Args) > 1) && (os.Args[1] == "debug")) {
		cvl.Debug(true)
	}

	rclient := getConfigDbClient()

	if ((len(os.Args) > 1) && (os.Args[1] == "add")) {

		//Add  ACL
		aclNoStart, _ := strconv.Atoi(os.Args[2])
		aclNoEnd, _ := strconv.Atoi(os.Args[3])
		for aclNum:= aclNoStart ;aclNum <= aclNoEnd; aclNum++ {
			aclNo := fmt.Sprintf("%d", aclNum)

			cvSess, _ := cvl.ValidationSessOpen()

			cfgDataAclRule := []cvl.CVLEditConfigData {
				cvl.CVLEditConfigData {
					cvl.VALIDATE_ALL,
					cvl.OP_CREATE,
					fmt.Sprintf("ACL_TABLE|TestACL%s", aclNo),
					map[string]string {
						"stage": "INGRESS",
						"type": "L3",
						//"ports@": "Ethernet0",
					},
				},
			}

			_, ret := cvSess.ValidateEditConfig(cfgDataAclRule)

			if (ret != cvl.CVL_SUCCESS) {
				fmt.Printf("Validation failure\n")
				return
			}

			cfgDataAclRule[0].VType = cvl.VALIDATE_NONE

			//Create 7 ACL rules
			for i:=0; i<7; i++ {
				cfgDataAclRule = append(cfgDataAclRule, cvl.CVLEditConfigData {
					cvl.VALIDATE_ALL,
					cvl.OP_CREATE,
					fmt.Sprintf("ACL_RULE|TestACL%s|Rule%d", aclNo, i+1),
					map[string]string {
						"PACKET_ACTION":     "FORWARD",
						"IP_TYPE": "IPV4",
						"SRC_IP":            "10.1.1.1/32",
						"L4_SRC_PORT":       fmt.Sprintf("%d", 201 + i),
						"IP_PROTOCOL":       "103",
						"DST_IP":            "20.2.2.2/32",
						"L4_DST_PORT":       fmt.Sprintf("%d", 701 + i),
					},
				})

				_, ret1 := cvSess.ValidateEditConfig(cfgDataAclRule)
				if (ret1 != cvl.CVL_SUCCESS) {
					fmt.Printf("Validation failure\n")
					return
				}

				cfgDataAclRule[1 + i].VType = cvl.VALIDATE_NONE
			}

			//Write to DB
			for _, cfgDataItem := range cfgDataAclRule {
				loadConfigDB(rclient, cfgDataItem.Key, cfgDataItem.Data)
			}

			cvl.ValidationSessClose(cvSess)
		}

		return
	} else if ((len(os.Args) > 1) && (os.Args[1] == "del")) {
		aclNoStart, _ := strconv.Atoi(os.Args[2])
		aclNoEnd, _ := strconv.Atoi(os.Args[3])
		for aclNum:= aclNoStart ;aclNum <= aclNoEnd; aclNum++ {
			aclNo := fmt.Sprintf("%d", aclNum)
			cvSess,_ := cvl.ValidationSessOpen()

			//Delete ACL

			cfgDataAclRule := []cvl.CVLEditConfigData{}

			//Create 7 ACL rules
			for i:=0; i<7; i++ {
				cfgDataAclRule = append(cfgDataAclRule, cvl.CVLEditConfigData {
					cvl.VALIDATE_ALL,
					cvl.OP_DELETE,
					fmt.Sprintf("ACL_RULE|TestACL%s|Rule%d", aclNo, i+1),
					map[string]string {
					},
				})

				_, ret := cvSess.ValidateEditConfig(cfgDataAclRule)
				if (ret != cvl.CVL_SUCCESS) {
					fmt.Printf("Validation failure\n")
					return
				}

				cfgDataAclRule[i].VType = cvl.VALIDATE_NONE
			}

			cfgDataAclRule = append(cfgDataAclRule,	cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_DELETE,
				fmt.Sprintf("ACL_TABLE|TestACL%s", aclNo),
				map[string]string {
				},
			})

			_, ret := cvSess.ValidateEditConfig(cfgDataAclRule)
			if (ret != cvl.CVL_SUCCESS) {
				fmt.Printf("Validation failure\n")
				return
			}

			//Write to DB
			for _, cfgDataItem := range cfgDataAclRule {
				unloadConfigDB(rclient, cfgDataItem.Key, cfgDataItem.Data)
			}

			cvl.ValidationSessClose(cvSess)
		}

		return
	}

	cv, ret := cvl.ValidationSessOpen()
	if (ret != cvl.CVL_SUCCESS) {
		fmt.Printf("Could not create CVL session")
		return
	}

	{
		count++
		jsonData :=`{
			"ACL_TABLE": {
				"TestACL1": {
					"stage": "INGRESS",
					"type": "L3"
				},
				"TestACL2": {
					"stage": "EGRESS",
					"ports": "Ethernet4"
				}
			}
		}`

		fmt.Printf("\nValidating data = %v\n\n", jsonData);

		err := cv.ValidateConfig(jsonData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		jsonData :=`{
			"ACL_TABLE": {
				"TestACL2": {
					"stage": "EGRESS",
					"ports": "Ethernet804"
				}
			}
		}`


		fmt.Printf("\nValidating data for external dependency check = %v\n\n", jsonData);

		err := cv.ValidateConfig(jsonData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		jsonData :=`{
			"ACL_TABLE": {
				"TestACL1": {
					"type": "L3"
				}
			}
		}`


		fmt.Printf("\nValidating data for mandatory element misssing = %v\n\n", jsonData);

		err := cv.ValidateConfig(jsonData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		jsonData :=`{
			"ACL_TABLE": {
				"TestACL1": {
					"stage": "INGRESS",
					"type": "L3"
				}
			},
			"ACL_RULE": {
				"TestACL1|Rule1": {
					"PACKET_ACTION": "FORWARD",
					"IP_PROTOCOL": "103",
					"SRC_IP": "10.1.1.1/32",
					"DST_IP": "20.2.2.2/32"
				}
			}
		}`



		fmt.Printf("\nValidating data for internal dependency check = %v\n\n", jsonData);

		err := cv.ValidateConfig(jsonData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		jsonData :=`{
			"ACL_TABLE": {
				"TestACL1": {
					"stage": "INGRESS",
					"type": "L3"
				}
			},
			"ACL_RULE": {
				"TestACL1|Rule1": {
					"PACKET_ACTION": "FORWARD",
					"IP_PROTOCOL": "103"
				}
			}
		}`



		fmt.Printf("\nValidating data for mandatory element check = %v\n\n", jsonData);

		err := cv.ValidateConfig(jsonData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		jsonData :=`{
			"ACL_TABLE": {
				"TestACL1": {
					"stage": "INGRESS",
					"type": "L3"
				}
			},
			"ACL_RULE": {
				"TestACL1|Rule1": {
					"PACKET_ACTION": "FORWARD",
					"IP_PROTOCOL": "103",
					"DST_IP": "20.2.2.2/32"
				}
			}
		}`



		fmt.Printf("\nValidating data for mandatory element check = %v\n\n", jsonData);

		err := cv.ValidateConfig(jsonData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		jsonData :=`{
			"ACL_TABLE": {
				"TestACL1": {
					"stage": "INGRESS",
					"type": "L3"
				}
			},
			"ACL_RULE": {
				"TestACL1|Rule1": {
					"PACKET_ACTION": "FORWARD",
					"SRC_IP": "10.1.1.1/32",
					"L4_SRC_PORT": 8080,
					"ETHER_TYPE":"0x0800",
					"IP_PROTOCOL": "1",
					"DST_IP": "20.2.2.2/32",
					"L4_DST_PORT_RANGE": "9000-12000"
				}
			}
		}`



		fmt.Printf("\nValidating data for pattern check = %v\n\n", jsonData);

		err := cv.ValidateConfig(jsonData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		jsonData :=`{
			"ACL_TABLE": {
				"TestACL1": {
					"stage": "INGRESS",
					"type": "L3"
				}
			},
			"ACL_RULE": {
				"TestACL1|Rule1": {
					"PACKET_ACTION": "FORWARD",
					"SRC_IP": "10.1.1.1/32",
					"L4_SRC_PORT": "ABC",
					"IP_PROTOCOL": "103",
					"DST_IP": "20.2.2.2/32",
					"L4_DST_PORT_RANGE": "9000-12000"
				}
			}
		}`



		fmt.Printf("\nValidating data for type check = %v\n\n", jsonData);

		err := cv.ValidateConfig(jsonData)

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
