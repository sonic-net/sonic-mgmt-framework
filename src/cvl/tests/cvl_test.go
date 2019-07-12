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

	for index, tc := range tests {
		t.Logf("Running Testcase %d with Description %s", index+1, tc.filedescription)
		t.Run(fmt.Sprintf("%s [%d]", tc.filedescription, index+1), func(t *testing.T) {
			err := cvl.ValidateConfig(tc.jsonString)

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
		{filedescription: "Config File - VLAN,ACL,PORTCHANNEL", fileName: "./config_db1.json", retCode: cvl.CVL_SUCCESS},
		{filedescription: "Config File - BUFFER_PG", fileName: "./config_db2.json", retCode: cvl.CVL_SUCCESS},
	}

	for index, tc := range tests {

		t.Logf("Running Testcase %d with Description %s", index+1, tc.filedescription)
		t.Run(tc.filedescription, func(t *testing.T) {
			jsonString := convertJsonFileToString(t, tc.fileName)
			err := cvl.ValidateConfig(jsonString)

			fmt.Printf("\nValidating data = %v\n\n", jsonString)

			if err != tc.retCode {
				t.Errorf("Config Validation failed.")
			}

		})
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_Valid_FieldValue(t *testing.T) {

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

}

/* API to test edit config with invalid syntax. */
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

		if err != cvl.CVL_SUCCESS {
			t.Errorf("Config Validation failed -- error details.")
		}
	})

}

/* API to test edit config with invalid type. */
func TestValidateEditConfig_Create_Syntax_Invalid_Type(t *testing.T) {

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
		{
			data: []cvl.CVLEditConfigData{
				cvl.CVLEditConfigData{cvl.VALIDATE_NONE, cvl.OP_CREATE, "MIRROR_SESSION|everflow", map[string]string{
					"src_ip": "10.1.0.3288888",
					"dst_ip": "2.2.2.2",
				},
				},
				cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_RULE|MyACL11_ACL_IPV4|RULE_1", map[string]string{
					"MIRROR_ACTION": "everflow",
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
        defer rclient.Close()
        rclient.FlushDb()

	/* Create port table. */
        fileName := "tests/config_db1.json"
        countersPortAliasMapByte, err := ioutil.ReadFile(fileName)
        if err != nil {
                fmt.Printf("read file %v err: %v", fileName, err)
        }

        mpi_alias_map := loadConfig("", countersPortAliasMapByte)
        loadConfigDB(rclient, mpi_alias_map)

	/* Create ACL Table. */
        fileName = "./create_acl_table.json"
        countersACLTableMapByte, err := ioutil.ReadFile(fileName)
        if err != nil {
                fmt.Printf("read file %v err: %v", fileName, err)
        }

        mpi_acl_table_map := loadConfig("", countersACLTableMapByte)
        loadConfigDB(rclient, mpi_acl_table_map)
}

/* Setup before starting of test. */ 
func TestMain(m *testing.M) {
	/* Prepare the Redis database. */
	prepareDb()
	os.Exit(m.Run())
}
