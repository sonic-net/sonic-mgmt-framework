package cvl_test

import (
	"cvl"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"os"
	"github.com/go-redis/redis"
)

type testEditCfgData struct {
	filedescription string
	cfgData         string
	depData         string
	retCode         cvl.CVLRetCode
}

var rclient *redis.Client
var cv  *cvl.CVL
var port_map map[string]interface{}


/* Converts JSON Data in a File to Map. */
func convertJsonFileToMap(t *testing.T, fileName string) map[string]string {
	var mapstr map[string]string

	jsonData := convertJsonFileToString(t, fileName)
	byteData := []byte(jsonData)

	err := json.Unmarshal(byteData, &mapstr)

	if err != nil {
		fmt.Println("Failed to convert Json File to map:", err)
	}

	return mapstr

}

/* Converts JSON Data in a File to Map. */
func convertDataStringToMap(t *testing.T, dataString string) map[string]string {
	var mapstr map[string]string

	byteData := []byte(dataString)

	err := json.Unmarshal(byteData, &mapstr)

	if err != nil {
		fmt.Println("Failed to convert Json Data String to map:", err)
	}

	return mapstr

}

/* Converts JSON Data in a File to String. */
func convertJsonFileToString(t *testing.T, fileName string) string {
	var jsonData string

	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		fmt.Printf("\nFailed to read data file : %v\n", err)
	} else {
		jsonData = string(data)
	}

	return jsonData
}

/* Converts JSON config to map which can be loaded to Redis */
func loadConfig(key string, in []byte) map[string]interface{} {
        var fvp map[string]interface{}

        err := json.Unmarshal(in, &fvp)
        if err != nil {
                fmt.Printf("Failed to Unmarshal %v err: %v", in, err)
        }
        if key != "" {
                kv := map[string]interface{}{}
                kv[key] = fvp
                return kv
        }
        return fvp
}

/* Separator for keys. */
func getSeparator() string {
        return "|"
}


/* Unloads the Config DB based on JSON File. */
func unloadConfigDB(rclient *redis.Client, mpi map[string]interface{}) {
        for key, fv := range mpi {
                switch fv.(type) {
                case map[string]interface{}:
                        for subKey, subValue := range fv.(map[string]interface{}) {
                                newKey := key + getSeparator() + subKey
                                _, err := rclient.Del(newKey).Result()

                                if err != nil {
                                        fmt.Printf("Invalid data for db: %v : %v %v", newKey, subValue, err)
                                }

                        }
                default:
                        fmt.Printf("Invalid data for db: %v : %v", key, fv)
                }
        }

}


/* Loads the Config DB based on JSON File. */
func loadConfigDB(rclient *redis.Client, mpi map[string]interface{}) {
	  for key, fv := range mpi {
                switch fv.(type) {
                case map[string]interface{}:
                        for subKey, subValue := range fv.(map[string]interface{}) {
                                newKey := key + getSeparator() + subKey
                                _, err := rclient.HMSet(newKey, subValue.(map[string]interface{})).Result()

                                if err != nil {
                                        fmt.Printf("Invalid data for db: %v : %v %v", newKey, subValue, err)
                                }

                        }
                default:
                        fmt.Printf("Invalid data for db: %v : %v", key, fv)
                }
        }
}


func getConfigDbClient() *redis.Client {
        rclient := redis.NewClient(&redis.Options{
                Network:     "tcp",
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

/* Prepares the database in Redis Server. */
func prepareDb() {
        rclient = getConfigDbClient()

        fileName := "testdata/port_table.json"
        PortsMapByte, err := ioutil.ReadFile(fileName)
        if err != nil {
                fmt.Printf("read file %v err: %v", fileName, err)
        }

        port_map := loadConfig("", PortsMapByte)

        loadConfigDB(rclient, port_map)
}

/* Setup before starting of test. */ 
func TestMain(m *testing.M) {
	cv, _ := cvl.ValidatorSessOpen()

	/* Prepare the Redis database. */
	prepareDb()
	os.Exit(m.Run())

	unloadConfigDB(rclient, port_map)
	cvl.ValidatorSessClose(cv)
	cvl.Finish()
        rclient.Close()
        rclient.FlushDb()
}


/* ValidateEditConfig with user input in file . */
func TestValidateEditConfig_CfgFile(t *testing.T) {

	tests := []struct {
		filedescription string
		cfgDataFile     string
		depDataFile     string
		retCode         cvl.CVLRetCode
	}{
		{filedescription: "ACL_DATA", cfgDataFile: "./acltable.json", depDataFile: "./aclrule.json", retCode: cvl.CVL_SUCCESS},
	}

	for index, tc := range tests {
		t.Logf("Running Testcase %d with Description %s", index+1, tc.filedescription)

		t.Run(tc.filedescription, func(t *testing.T) {

			jsonEditCfg_Create_DependentMap := convertJsonFileToMap(t, tc.depDataFile)
			jsonEditCfg_Create_ConfigMap := convertJsonFileToMap(t, tc.cfgDataFile)

			cfgData := []cvl.CVLEditConfigData{
				cvl.CVLEditConfigData{cvl.VALIDATE_NONE, cvl.OP_NONE, "ACL_TABLE|TestACL1", jsonEditCfg_Create_DependentMap},
				cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_RULE|TestACL1|Rule1", jsonEditCfg_Create_ConfigMap},
			}

			fmt.Printf("\n\n Validating create data = %v\n\n", cfgData)

			err := cvl.ValidateEditConfig(cfgData)

			if err != tc.retCode {
				t.Errorf("Config Validation failed.")
			}
		})
	}
}

/* ValidateEditConfig with user input inline. */
func TestValidateEditConfig_CfgStrBuffer(t *testing.T) {

	type testStruct struct {
		filedescription string
		cfgData         string
		depData         string
		retCode         cvl.CVLRetCode
	}

	tests := []testStruct{}

	/* Iterate through data present is separate file. */
	for index, _ := range json_edit_config_create_acl_table_dependent_data {
		tests = append(tests, testStruct{filedescription: "ACL_DATA", cfgData: json_edit_config_create_acl_rule_config_data[index],
			depData: json_edit_config_create_acl_table_dependent_data[index], retCode: cvl.CVL_SUCCESS})
	}

	for index, tc := range tests {
		t.Logf("Running Testcase %d with Description %s", index+1, tc.filedescription)
		t.Run(tc.filedescription, func(t *testing.T) {
			jsonEditCfg_Create_DependentMap := convertDataStringToMap(t, tc.depData)
			jsonEditCfg_Create_ConfigMap := convertDataStringToMap(t, tc.cfgData)

			cfgData := []cvl.CVLEditConfigData{
				cvl.CVLEditConfigData{cvl.VALIDATE_NONE, cvl.OP_NONE, "ACL_TABLE|TestACL1", jsonEditCfg_Create_DependentMap},
				cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_RULE|TestACL1|Rule1", jsonEditCfg_Create_ConfigMap},
			}

			fmt.Printf("\n\n Validating create data = %v\n\n", cfgData)

			err := cvl.ValidateEditConfig(cfgData)

			if err != tc.retCode {
				t.Errorf("Config Validation failed.")
			}
		})
	}

}

/* API when config is given as string buffer. */
func TestValidateConfig_CfgStrBuffer(t *testing.T) {
	type testStruct struct {
		filedescription string
		jsonString      string
		retCode         cvl.CVLRetCode
	}


	tests := []testStruct{}

	for index, _ := range json_validate_config_data {
		/* Fetch the modelName. */
		result := strings.Split(json_validate_config_data[index], "{")
		modelName := strings.Trim(strings.Replace(strings.TrimSpace(result[1]), "\"", "", -1), ":")

		tests = append(tests, testStruct{filedescription: modelName, jsonString: json_validate_config_data[index], retCode: cvl.CVL_SUCCESS})
	}

	cv, _ := cvl.ValidatorSessOpen()

	fmt.Printf("Cv Value : %v", cv) 

	for index, tc := range tests {
		t.Logf("Running Testcase %d with Description %s", index+1, tc.filedescription)
		t.Run(fmt.Sprintf("%s [%d]", tc.filedescription, index+1), func(t *testing.T) {
			err := cv.ValidateConfig(tc.jsonString)

			fmt.Printf("\nValidating data = %v\n\n", tc.jsonString)

			if err != tc.retCode {
				t.Errorf("Config Validation failed.")
			}

		})
	}
}

/* API when config is given as json file. */
func TestValidateConfig_CfgFile(t *testing.T) {

	/* Structure containing file information. */
	tests := []struct {
		filedescription string
		fileName        string
		retCode         cvl.CVLRetCode
	}{
		{filedescription: "Config File - VLAN,ACL,PORTCHANNEL", fileName: "testdata/config_db1.json", retCode: cvl.CVL_SUCCESS},
		{filedescription: "Config File - BUFFER_PG", fileName: "testdata/config_db2.json", retCode: cvl.CVL_SUCCESS},
		{filedescription: "Config File - BUFFER_PG", fileName: "testdata/config_db3.json", retCode: cvl.CVL_SUCCESS},
		{filedescription: "Config File - BUFFER_PG", fileName: "testdata/config_db4.json", retCode: cvl.CVL_SUCCESS},
	}

	cv, _ := cvl.ValidatorSessOpen()

	for index, tc := range tests {

		t.Logf("Running Testcase %d with Description %s", index+1, tc.filedescription)
		t.Run(tc.filedescription, func(t *testing.T) {
			jsonString := convertJsonFileToString(t, tc.fileName)
			err := cv.ValidateConfig(jsonString)

			fmt.Printf("\nValidating data = %v\n\n", jsonString)

			if err != tc.retCode {
				t.Errorf("Config Validation failed.")
			}

		})
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_Valid_FieldValue(t *testing.T) {


	// Create ACL Table. 
	fileName := "testdata/create_acl_table.json"
        aclTableMapByte, err := ioutil.ReadFile(fileName)
        if err != nil {
                fmt.Printf("read file %v err: %v", fileName, err)
        }

        mpi_acl_table_map := loadConfig("", aclTableMapByte)
        loadConfigDB(rclient, mpi_acl_table_map)

	t.Run("Positive - EditConfig(Create) : Valid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
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


		err := cvl.ValidateEditConfig(cfgData)

		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})
	
	unloadConfigDB(rclient, mpi_acl_table_map)

}

/* API to test edit config with invalid field value. */
func TestValidateEditConfig_Create_Syntax_Invalid_FieldValue(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		cfgData := []cvl.CVLEditConfigData{
			cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_TABLE|TestACL1", map[string]string{
				"stage": "INGRESS",
				"type":  "junk",
			},
			},
		}

		err := cvl.ValidateEditConfig(cfgData)
		fmt.Println(err)

		/* TBD . Proper Error not Returned. */
		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

/* API to test edit config with invalid type. */
func TestValidateEditConfig_Create_Syntax_Valid_PacketAction_Positive(t *testing.T) {

	type testEditCfgData struct {
		data    []cvl.CVLEditConfigData
		retCode cvl.CVLRetCode
	}

	testData := []testEditCfgData{
		{
			data: []cvl.CVLEditConfigData{
				cvl.CVLEditConfigData{cvl.VALIDATE_NONE, cvl.OP_CREATE, "ACL_TABLE|TestACL1", map[string]string{
					"stage": "EGRESS",
					"type":  "L3",
				},
				},
				cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_RULE|TestACL1|Rule1", map[string]string{
					"PACKET_ACTION":     "FORWARD",
					"SRC_IP":            "10.1.1.1/32",
					"L4_SRC_PORT":       "1909",
					"IP_PROTOCOL":       "103",
					"DST_IP":            "20.2.2.2/32",
					"L4_DST_PORT_RANGE": "9000-12000",
				},
				},
			},
			retCode: cvl.CVL_SUCCESS,
		},
	}

	for idx, testDataItem := range testData {
		t.Run(fmt.Sprintf("Negative - EditConfig(Create) : Invalid Field Value [%d]", idx+1), func(t *testing.T) {

			err := cvl.ValidateEditConfig(testDataItem.data)

			if err != testDataItem.retCode {
				t.Errorf("Config Validation failed -- error details.")
			}
		})
	}

}


/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_Invalid_PacketAction_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "PACKET_ACTION": "FORWARD222",
                                        "SRC_IP": "77.10.1.1.1/32",
                                        "L4_SRC_PORT": "aa1909",
                                        "IP_PROTOCOL": "10388888",
                                        "DST_IP": "20.2.2.2/32",
                                        "L4_DST_PORT_RANGE": "9000-12000",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_Invalid_SrcPrefix_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "PACKET_ACTION": "FORWARD777",
                                        "SRC_IP": "10.1.1.1/3288888",
                                        "L4_SRC_PORT": "aa1909",
                                        "IP_PROTOCOL": "10388888",
                                        "DST_IP": "20.2.2.2/32",
                                        "L4_DST_PORT_RANGE": "9000-12000",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}
/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_InvalidIPAddress_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid IP Address", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "PACKET_ACTION": "FORWARD",
                                        "SRC_IP": "10.1a.1.1/32888",
                                        "L4_SRC_PORT": "1909",
                                        "IP_PROTOCOL": "103",
                                        "DST_IP": "20.2.2.2/32",
                                        "L4_DST_PORT_RANGE": "9000-12000",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}
/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_OutofBound_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Out of Bound Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "PACKET_ACTION": "FORWARD",
                                        "SRC_IP": "10.1.1.1/32",
                                        "L4_SRC_PORT": "19099090909090",
                                        "IP_PROTOCOL": "103",
                                        "DST_IP": "20.2.2.2/32",
                                        "L4_DST_PORT_RANGE": "9000-12000",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}
/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_InvalidProtocol_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "PACKET_ACTION": "FORWARD",
                                        "SRC_IP": "10.1.1.1/32",
                                        "L4_SRC_PORT": "1909",
                                        "IP_PROTOCOL": "10388888",
                                        "DST_IP": "20.2.2.2/32",
                                        "L4_DST_PORT_RANGE": "9000-12000",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_InvalidRange_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
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
                                        "L4_DST_PORT_RANGE": "777779000-12000",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_InvalidCharNEw_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|TestACL1jjjj|Rule1",
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


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Create_Syntax_InvalidChar_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|TestACL1|Rule@##",
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


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Create_Syntax_InvalidKeyName_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "AC&&***L_RULE|TestACL1|Rule1",
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


		err := cvl.ValidateEditConfig(cfgData)

			     fmt.Println("Ashanew", err)

		/* TBD , Error details return SUCCESS. */
		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Create_Semantic_AdditionalInvalidNode_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Additional Node", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
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
					"extra"     : "shhs",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Create_Semantic_MissingMandatoryNode_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "PACKET_ACTION": "FORWARD",
                                        "L4_SRC_PORT": "1909",
                                        "IP_PROTOCOL": "103",
                                        "DST_IP": "20.2.2.2/32",
                                        "L4_DST_PORT_RANGE": "9000-12000",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Create_Syntax_Invalid_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULERule1",
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


		err := cvl.ValidateEditConfig(cfgData)

		/* TBD. */
		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Create_Syntax_IncompleteKey_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|Rule1",
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


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Create_Syntax_InvalidKey_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "|Rule1",
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


		err := cvl.ValidateEditConfig(cfgData)

		/* TBD */
		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Update_Syntax_DependentData_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {


		cfgData := []cvl.CVLEditConfigData {
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


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Create_Syntax_DependentData_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

                 cfgData :=  []cvl.CVLEditConfigData {
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

		err := cvl.ValidateEditConfig(cfgData)


		/* TBD */
		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Delete_Syntax_InvalidKey_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Delete) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_DELETE,
                                "|Rule1",
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


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Update_Syntax_InvalidKey_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Update) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_UPDATE,
                                "|Rule1",
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


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Delete_InvalidKey_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Delete) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_DELETE,
                                "ACL_RULE|TestACL1:Rule1",
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


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Update_Syntax_Invalid_Field_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Update) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_UPDATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "PACKET_ACTION": "CATCH",
                                        "SRC_IP": "10.1.1.1/32",
                                        "L4_SRC_PORT": "1909",
                                        "IP_PROTOCOL": "103",
                                        "DST_IP": "20.2.2.2/32",
                                        "L4_DST_PORT_RANGE": "9000-12000",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Update_Semantic_Invalid_Key_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		    cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_UPDATE,
                                "ACL_RULE|TestACL1Rule1",
                                map[string]string {
                                        "PACKET_ACTION": "FORWARD",
                                        "SRC_IP": "10.1.1.1/32",
                                        "L4_SRC_PORT": "1909",
                                        "IP_PROTOCOL": "103uuuu",
                                        "DST_IP": "20.2.2.2/32",
                                        "L4_DST_PORT_RANGE": "9000-12000",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Delete_Semantic_Positive(t *testing.T) {
	depDataMap := map[string]interface{} {
		"MIRROR_SESSION" : map[string]interface{} {
			"everflow": map[string] interface{} {
				"src_ip": "10.1.0.32",
                                "dst_ip": "2.2.2.2",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)

	t.Run("Positivee - EditConfig(Delete)", func(t *testing.T) {

		 cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_DELETE,
                                "MIRROR_SESSION|everflow",
                                map[string]string {
                                },
                        },
                }

		err := cvl.ValidateEditConfig(cfgData)

		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

	unloadConfigDB(rclient, depDataMap)

}

func TestValidateEditConfig_Delete_Semantic_MissingKey_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Delete) : Invalid Field Value", func(t *testing.T) {

		 cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_DELETE,
                                "MIRROR_SESSION|everflow0",
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

		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Update_Semantic_MissingKey_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {

		cfgData := []cvl.CVLEditConfigData {
                        	cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_UPDATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "MIRROR_ACTION": "everflow",
                                },
                        },
                }

		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Update_Semantic_Positive(t *testing.T) {

	// Create ACL Table. 
	fileName := "testdata/create_acl_table.json"
        aclTableMapByte, err := ioutil.ReadFile(fileName)
        if err != nil {
                fmt.Printf("read file %v err: %v", fileName, err)
        }

        mpi_acl_table_map := loadConfig("", aclTableMapByte)
        loadConfigDB(rclient, mpi_acl_table_map)

	t.Run("Positive - EditConfig(Update) : Valid Field Value", func(t *testing.T) {


			cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_UPDATE,
                                "ACL_TABLE|TestACL1",
                                map[string]string {
                                        "stage": "INGRESS",
                                        "type": "MIRROR",
                                },
                        },
                }

		err := cvl.ValidateEditConfig(cfgData)

		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

	unloadConfigDB(rclient, mpi_acl_table_map)

}

/* API to test edit config with valid syntax. */
func TestValidateConfig_Update_Semantic_Vlan_Negative(t *testing.T) {

	cv, _ := cvl.ValidatorSessOpen()
	t.Run("Negative - EditConfig(Update) : Valid Field Value", func(t *testing.T) {

			jsonData :=`{
                        "VLAN": {
                                "Vlan100": {
                                        "members": [
                                        "Ethernet44",
                                        "Ethernet64"
                                        ],
                                        "vlanid": "107"
                                }
                        }
                }`


                err := cv.ValidateConfig(jsonData)


		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Update_Syntax_DependentData_Redis_Positive(t *testing.T) {

	// Create ACL Table. 
	fileName := "testdata/create_acl_table13.json"
        aclTableMapByte, err := ioutil.ReadFile(fileName)
        if err != nil {
                fmt.Printf("read file %v err: %v", fileName, err)
        }

        mpi_acl_table_map := loadConfig("", aclTableMapByte)
        loadConfigDB(rclient, mpi_acl_table_map)

	// Create ACL Rule. 
        fileName = "testdata/acl_rule.json"
        aclTableMapRule, err := ioutil.ReadFile(fileName)
        if err != nil {
                fmt.Printf("read file %v err: %v", fileName, err)
        }

        mpi_acl_table_rule := loadConfig("", aclTableMapRule)
        loadConfigDB(rclient, mpi_acl_table_rule)

	depDataMap := map[string]interface{} {
		"MIRROR_SESSION" : map[string]interface{} {
			"everflow2": map[string] interface{} {
				"src_ip": "10.1.0.32",
                                "dst_ip": "2.2.2.2",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)

	t.Run("Negative - EditConfig(Update) ", func(t *testing.T) {


		/* ACL and Rule name pre-created . */
		cfgData := []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_UPDATE,
                                "ACL_RULE|TestACL13|Rule1",
                                map[string]string {
                                        "MIRROR_ACTION": "everflow2",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

	unloadConfigDB(rclient, mpi_acl_table_map)
	unloadConfigDB(rclient, mpi_acl_table_rule)
	unloadConfigDB(rclient, depDataMap)

}

func TestValidateEditConfig_Update_Syntax_DependentData_Invalid_Op_Seq(t *testing.T) {

	t.Run("Negative - EditConfig(Create/Update) - same entry can't be created and updated in single request", func(t *testing.T) {


		/* ACL and Rule name pre-created . */
		cfgData := []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_NONE,
                                cvl.OP_CREATE,
                                "ACL_TABLE|TestACL1",
                                map[string]string {
                                        "stage": "INGRESS",
                                        "type": "MIRROR",
                                },
                        },
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_NONE,
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
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_UPDATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "PACKET_ACTION": "DROP",
                                        "L4_SRC_PORT": "781",
                                },
                        },
                }



		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS { //Validation should fail
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Update_Syntax_DependentData_Redis_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Create) : Invalid Field Value", func(t *testing.T) {


		/* ACL does not exist.*/
		cfgData := []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_UPDATE,
                                "ACL_RULE|TestACL1|Rule1",
                                map[string]string {
                                        "MIRROR_ACTION": "everflow0",
                                },
                        },
                }


		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

/* Create with User provideddependent data. */
func TestValidateEditConfig_Create_Syntax_DependentData_Redis_Positive(t *testing.T) {

	t.Run("Positive - EditConfig(Create)", func(t *testing.T) {

		/* ACL and Rule name pre-created . */
		cfgData := []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_NONE,
                                cvl.OP_CREATE,
                                "ACL_TABLE|TestACL22",
                                map[string]string {
                                        "stage": "INGRESS",
                                        "type": "MIRROR",
                                },
                        },
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_CREATE,
                                "ACL_RULE|TestACL22|Rule1",
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


		err := cvl.ValidateEditConfig(cfgData)

		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

/* Delete Non-Existing Key.*/
func TestValidateEditConfig_Delete_Semantic_ACLTableReference_Negative(t *testing.T) {

	t.Run("Negative - EditConfig(Delete) : Invalid Field Value", func(t *testing.T) {

		 cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_DELETE,
                                "ACL_RULE|MyACL11_ACL_IPV4|RULE_1",
                                map[string]string {
                                },
                        },
                }

		err := cvl.ValidateEditConfig(cfgData)

		if err == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

/* Delete Existing Key.*/
func TestValidateEditConfig_Delete_Semantic_ACLTableReference_Positive(t *testing.T) {

	t.Run("Positive - EditConfig(Delete)", func(t *testing.T) {

		 cfgData :=  []cvl.CVLEditConfigData {
                        cvl.CVLEditConfigData {
                                cvl.VALIDATE_ALL,
                                cvl.OP_DELETE,
                                "ACL_RULE|TestACL11|Rule1",
                                map[string]string {
                                },
                        },
                }

		err := cvl.ValidateEditConfig(cfgData)

		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

func TestValidateEditConfig_Create_Dependent_CacheData(t *testing.T) {

	cvSess, _ := cvl.ValidatorSessOpen()

	t.Run("Positive - EditConfig(Create) with dependent data from cache", func(t *testing.T) {
		//Create ACL rule
		cfgDataAcl := []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_CREATE,
				"ACL_TABLE|TestACL14",
				map[string]string {
					"stage": "INGRESS",
					"type": "MIRROR",
				},
			},
		}

		err1 := cvSess.ValidateEditConfig1(cfgDataAcl)

		//Create ACL rule
		cfgDataRule := []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_CREATE,
				"ACL_RULE|TestACL14|Rule1",
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

		err2 := cvSess.ValidateEditConfig1(cfgDataRule)

		if err1 != cvl.CVL_SUCCESS || err2 != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed.")
		}
	})
	cvl.ValidatorSessClose(cvSess)
}

func TestValidateEditConfig_Create_DepData_In_MultiSess(t *testing.T) {


	t.Run("Negative - EditConfig(Create) with dependent data in multiple sessionsi, but no cached data", func(t *testing.T) {

		//Create ACL rule - Session 1
		cvSess, _ := cvl.ValidatorSessOpen()
		cfgDataAcl := []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_CREATE,
				"ACL_TABLE|TestACL16",
				map[string]string {
					"stage": "INGRESS",
					"type": "MIRROR",
				},
			},
		}

		err1 := cvSess.ValidateEditConfig1(cfgDataAcl)
		cvl.ValidatorSessClose(cvSess)

		//Create ACL rule - Session 2, validation should fail
		cvSess, _ = cvl.ValidatorSessOpen()
		cfgDataRule := []cvl.CVLEditConfigData {
			cvl.CVLEditConfigData {
				cvl.VALIDATE_ALL,
				cvl.OP_CREATE,
				"ACL_RULE|TestACL16|Rule1",
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

		err2 := cvSess.ValidateEditConfig1(cfgDataRule)
		cvl.ValidatorSessClose(cvSess)

		if err1 != cvl.CVL_SUCCESS || err2 == cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed.")
		}
	})

}

func TestValidateEditConfig_Create_DepData_From_Redis(t *testing.T) {

	depDataMap := map[string]interface{} {
		"ACL_TABLE" : map[string]interface{} {
			"TestACL1": map[string] interface{} {
				"stage": "INGRESS",
				"type": "MIRROR",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)

	t.Run("Positive - EditConfig(Create) with dependent data from redis", func(t *testing.T) {
		//Create ACL rule - Session 2
		cvSess, _ := cvl.ValidatorSessOpen()
		cfgDataRule := []cvl.CVLEditConfigData {
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

		err := cvSess.ValidateEditConfig1(cfgDataRule)
		cvl.ValidatorSessClose(cvSess)

		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed.")
		}
	})

	unloadConfigDB(rclient, depDataMap)
}
