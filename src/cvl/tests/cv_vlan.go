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
		vlanNoStart, _ := strconv.Atoi(os.Args[2])
		vlanNoEnd, _ := strconv.Atoi(os.Args[3])
		for vlanNum:= vlanNoStart ;vlanNum <= vlanNoEnd; vlanNum++ {
			cvSess, _ := cvl.ValidationSessOpen()

			cfgDataVlan := []cvl.CVLEditConfigData {
				cvl.CVLEditConfigData {
					cvl.VALIDATE_ALL,
					cvl.OP_CREATE,
					fmt.Sprintf("VLAN|Vlan%d", vlanNum),
					map[string]string {
						"vlanid":  fmt.Sprintf("%d", vlanNum),
						"members@": "Ethernet0,Ethernet4,Ethernet8,Ethernet12,Ethernet16,Ethernet20,Ethernet24,Ethernet28",
					},
				},
			}

			_, ret := cvSess.ValidateEditConfig(cfgDataVlan)

			if (ret != cvl.CVL_SUCCESS) {
				fmt.Printf("Validation failure\n")
				return
			}

			cfgDataVlan[0].VType = cvl.VALIDATE_NONE

			for i:=0; i<7; i++ {
				cfgDataVlan = append(cfgDataVlan, cvl.CVLEditConfigData {
					cvl.VALIDATE_ALL,
					cvl.OP_CREATE,
					fmt.Sprintf("VLAN_MEMBER|Vlan%d|Ethernet%d", vlanNum, i * 4),
					map[string]string {
						"tagging_mode" : "tagged",
					},
				})

				_, ret1 := cvSess.ValidateEditConfig(cfgDataVlan)
				if (ret1 != cvl.CVL_SUCCESS) {
					fmt.Printf("Validation failure\n")
					return
				}

				cfgDataVlan[1 + i].VType = cvl.VALIDATE_NONE
			}

			//Write to DB
			for _, cfgDataItem := range cfgDataVlan {
				loadConfigDB(rclient, cfgDataItem.Key, cfgDataItem.Data)
			}

			cvl.ValidationSessClose(cvSess)
		}

		return
	} else if ((len(os.Args) > 1) && (os.Args[1] == "del")) {
		vlanNoStart, _ := strconv.Atoi(os.Args[2])
		vlanNoEnd, _ := strconv.Atoi(os.Args[3])
		for vlanNum:= vlanNoStart ;vlanNum <= vlanNoEnd; vlanNum++ {
			cvSess,_ := cvl.ValidationSessOpen()

			//Delete ACL

			cfgDataVlan := []cvl.CVLEditConfigData{}

			//Create 7 ACL rules
			for i:=0; i<7; i++ {
				cfgDataVlan = append(cfgDataVlan, cvl.CVLEditConfigData {
					cvl.VALIDATE_ALL,
					cvl.OP_DELETE,
					fmt.Sprintf("VLAN_MEMBER|Vlan%d|Ethernet%d", vlanNum, i * 4),
					map[string]string {
					},
				})

				_, ret := cvSess.ValidateEditConfig(cfgDataVlan)
				if (ret != cvl.CVL_SUCCESS) {
					fmt.Printf("Validation failure\n")
					//return
				}

				cfgDataVlan[i].VType = cvl.VALIDATE_NONE
			}

			cfgDataVlan = append(cfgDataVlan,	cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_DELETE,
				fmt.Sprintf("VLAN|Vlan%d", vlanNum),
				map[string]string {
				},
			})

			_, ret := cvSess.ValidateEditConfig(cfgDataVlan)
			if (ret != cvl.CVL_SUCCESS) {
				fmt.Printf("Validation failure\n")
				//return
			}

			//Write to DB
			for _, cfgDataItem := range cfgDataVlan {
				unloadConfigDB(rclient, cfgDataItem.Key, cfgDataItem.Data)
			}

			cvl.ValidationSessClose(cvSess)
		}
		return
	}
	cv, ret := cvl.ValidationSessOpen()
	if (ret != cvl.CVL_SUCCESS) {
		fmt.Printf("NewDB: Could not create CVL session")
		return
	}

	{
		count++
		keyData :=  []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_NONE,
				cvl.OP_NONE,
				"PORTCHANNEL|ch1",
				map[string]string {
					"admin_status": "up",
					"mtu": "9100",
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_NONE,
				cvl.OP_NONE,
				"PORTCHANNEL|ch2",
				map[string]string {
					"admin_status": "up",
					"mtu": "9100",
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_NONE,
				cvl.OP_NONE,
				"PORTCHANNEL_MEMBER|ch1|Ethernet4",
				map[string]string {
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_NONE,
				cvl.OP_NONE,
				"PORTCHANNEL_MEMBER|ch1|Ethernet8",
				map[string]string {
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_NONE,
				cvl.OP_NONE,
				"PORTCHANNEL_MEMBER|ch2|Ethernet12",
				map[string]string {
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_NONE,
				cvl.OP_NONE,
				"PORTCHANNEL_MEMBER|ch2|Ethernet16",
				map[string]string {
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_NONE,
				cvl.OP_NONE,
				"PORTCHANNEL_MEMBER|ch2|Ethernet20",
				map[string]string {
				},
			},
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_CREATE,
				"VLAN|Vlan1001",
				map[string]string {
					"vlanid": "1001",
					"members@": "Ethernet24,ch1,Ethernet8",
				},
			},
		}

		fmt.Printf("\nValidating data for must = %v\n\n", keyData);

		_, err := cv.ValidateEditConfig(keyData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}

	}

	{
		keyData :=  []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_DELETE,
				"ACL_TABLE|MyACL1_ACL_IPV4",
				map[string]string {
					"type": "L3",
				},
			},
		}

		_, err := cv.ValidateEditConfig(keyData)

		fmt.Printf("\nValidating field delete...\n\n");

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}

	}

	{
		count++
		jsonData :=`{
			"VLAN": {
				"Vlan100": {
					"members": [
					"Ethernet44",
					"Ethernet64"
					],
					"vlanid": "100"
				}
			}
		}`


		err := cv.ValidateConfig(jsonData)

		fmt.Printf("\nValidating data = %v\n\n", jsonData);

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}

	{
		count++
		jsonData :=`{
			"VLAN": {
				"Vln100": {
					"members": [
					"Ethernet44",
					"Ethernet64"
					],
					"vlanid": "100"
				}
			}
		}`


		err := cv.ValidateConfig(jsonData)

		fmt.Printf("\nValidating data for key syntax = %v\n\n", jsonData);

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}

	{
		count++
		jsonData :=`{
			"VLAN": {
				"Vlan4096": {
					"members": [
					"Ethernet44",
					"Ethernet64"
					],
					"vlanid": "100"
				}
			}
		}`


		err := cv.ValidateConfig(jsonData)

		fmt.Printf("\nValidating data for range check = %v\n\n", jsonData);

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		jsonData :=`{
			"VLAN": {
				"Vlan201": {
					"members": [
					"Ethernet44",
					"Ethernet64"
					],
					"vlanid": "100"
				}
			}
		}`


		err := cv.ValidateConfig(jsonData)

		fmt.Printf("\nValidating data for internal dependency check = %v\n\n", jsonData);

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		/*jsonData :=`{
			"VLAN": {
				"Vlan100": {
					"members": [
					"Ethernet44",
					"Ethernet964"
					],
					"vlanid": "100"
				},
				"Vlan1200": {
					"members": [
					"Ethernet64",
					"Ethernet1008"
					],
					"vlanid": "1200"
				}
			}
		}`*/
		jsonData :=`{
			"VLAN": {
				"Vlan4095": {
					"vlanid": "4995"
				}
			}
		}`

		err := cv.ValidateConfig(jsonData)

		fmt.Printf("\nValidating data for external dependency check = %v\n\n", jsonData);

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
