package cvl_test

import (
	"cvl"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type testEditCfgData struct {
	filedescription string
	cfgData         string
	depData         string
	retCode         cvl.CVLRetCode
}

var rclient *redis.Client
var cv *cvl.CVL
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

func compareErrorDetails(cvlErr cvl.CVLErrorInfo, expCode cvl.CVLRetCode, errAppTag string, constraintmsg string) bool {

	if ((cvlErr.ErrCode == expCode) && ((cvlErr.ErrAppTag == errAppTag) || (cvlErr.ConstraintErrMsg == constraintmsg))) {
		return true
	}

	return false
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

	redisAlreadyRunning := false
	pidOfRedis, err := exec.Command("/bin/pidof", "redis-server").Output()
	if err == nil &&  string(pidOfRedis) != "\n" {
		redisAlreadyRunning = true
	}

	if (redisAlreadyRunning == false) {
		//Redis not running, lets start it
		output, err := exec.Command("/bin/sh", "-c", "sudo /etc/init.d/redis-server start").Output()
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println(string(output))
	}

	cv, _ = cvl.ValidationSessOpen()

	/* Prepare the Redis database. */
	prepareDb()
	code := m.Run()
	//os.Exit(m.Run())

	unloadConfigDB(rclient, port_map)
	cvl.ValidationSessClose(cv)
	cvl.Finish()
	rclient.Close()
	rclient.FlushDb()

	if (redisAlreadyRunning == false) {
		//If Redis was not already running, close the instance that we ran
		output, err := exec.Command("/bin/sh", "-c", "sudo /etc/init.d/redis-server stop").Output()
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println(string(output))
	}

	os.Exit(code)

}

//Test Initialize() API
func TestInitialize(t *testing.T) {
	ret := cvl.Initialize()
	if (ret != cvl.CVL_SUCCESS) {
		t.Errorf("CVl initialization failed")
	}

	ret = cvl.Initialize()
	if (ret != cvl.CVL_SUCCESS) {
		t.Errorf("CVl re-initialization should not fail")
	}
}

//Test Initialize() API
func TestFinish(t *testing.T) {
	ret := cvl.Initialize()
	if (ret != cvl.CVL_SUCCESS) {
		t.Errorf("CVl initialization failed")
	}

	cvl.Finish()
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

	cvSess, _ := cvl.ValidationSessOpen()

	for index, tc := range tests {
		t.Logf("Running Testcase %d with Description %s", index+1, tc.filedescription)

		t.Run(tc.filedescription, func(t *testing.T) {

			jsonEditCfg_Create_DependentMap := convertJsonFileToMap(t, tc.depDataFile)
			jsonEditCfg_Create_ConfigMap := convertJsonFileToMap(t, tc.cfgDataFile)

			cfgData := []cvl.CVLEditConfigData{
				cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_TABLE|TestACL1", jsonEditCfg_Create_DependentMap},
			}

			fmt.Printf("\n\n Validating create data = %v\n\n", cfgData)

			cvlErrObj, err := cvSess.ValidateEditConfig(cfgData)

			if err != tc.retCode {
				t.Errorf("Config Validation failed. %v", cvlErrObj)
			}

			cfgData = []cvl.CVLEditConfigData{
				cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_RULE|TestACL1|Rule1", jsonEditCfg_Create_ConfigMap},
			}

			fmt.Printf("\n\n Validating create data = %v\n\n", cfgData)

			cvlErrObj, err = cvSess.ValidateEditConfig(cfgData)

			if err != tc.retCode {
				t.Errorf("Config Validation failed. %v", cvlErrObj)
			}
		})
	}

	cvl.ValidationSessClose(cvSess)
}

/* ValidateEditConfig with user input inline. */
func TestValidateEditConfig_CfgStrBuffer(t *testing.T) {

	type testStruct struct {
		filedescription string
		cfgData         string
		depData         string
		retCode         cvl.CVLRetCode
	}

	cvSess, _ := cvl.ValidationSessOpen()

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
				cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_TABLE|TestACL1", jsonEditCfg_Create_DependentMap},
			}

			fmt.Printf("\n\n Validating create data = %v\n\n", cfgData)

			cvlErrObj, err := cvSess.ValidateEditConfig(cfgData)

			if err != tc.retCode {
				t.Errorf("Config Validation failed. %v", cvlErrObj)
			}

			cfgData = []cvl.CVLEditConfigData{
				cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_RULE|TestACL1|Rule1", jsonEditCfg_Create_ConfigMap},
			}

			fmt.Printf("\n\n Validating create data = %v\n\n", cfgData)

			cvlErrObj, err = cvSess.ValidateEditConfig(cfgData)

			if err != tc.retCode {
				t.Errorf("Config Validation failed. %v", cvlErrObj)
			}
		})
	}

	cvl.ValidationSessClose(cvSess)
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

	cv, _ := cvl.ValidationSessOpen()

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

	cv, _ := cvl.ValidationSessOpen()

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

//Validate invalid json data
func TestValidateConfig_Negative(t *testing.T) {
	cvSess, _ := cvl.ValidationSessOpen()
	jsonData := `{
		"VLANjunk": {
			"Vlan100": {
				"members": [
				"Ethernet4",
				"Ethernet8"
				],
				"vlanid": "100"
			}
		}
	}`

	err := cv.ValidateConfig(jsonData)

	if err == cvl.CVL_SUCCESS { //Should return failure
		t.Errorf("Config Validation failed.")
	}

	cvl.ValidationSessClose(cvSess)
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

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, retCode := cv.ValidateEditConfig(cfgData)

	if retCode != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, mpi_acl_table_map)

}

/* API to test edit config with invalid field value. */
func TestValidateEditConfig_Create_Syntax_CableLength(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"CABLE_LENGTH|AZURE",
			map[string]string{
			  "Ethernet8": "5m",
			  "Ethernet12": "5m",
			  "Ethernet16": "5m",
			},
		},
	 }

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)
	fmt.Println(err)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

/* API to test edit config with invalid field value. */
func TestValidateEditConfig_Create_Syntax_Invalid_FieldValue(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{cvl.VALIDATE_ALL, cvl.OP_CREATE, "ACL_TABLE|TestACL1", map[string]string{
			"stage": "INGRESS",
			"type":  "junk",
		},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)
	fmt.Println(err)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

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

			cvlErrInfo, err := cv.ValidateEditConfig(testDataItem.data)

			if err != testDataItem.retCode {
				t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
			}
		})
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_Invalid_PacketAction_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD222",
				"SRC_IP":            "77.10.1.1.1/32",
				"L4_SRC_PORT":       "aa1909",
				"IP_PROTOCOL":       "10388888",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_Invalid_SrcPrefix_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD777",
				"SRC_IP":            "10.1.1.1/3288888",
				"L4_SRC_PORT":       "aa1909",
				"IP_PROTOCOL":       "10388888",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_InvalidIPAddress_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1a.1.1/32888",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_OutofBound_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "19099090909090",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_InvalidProtocol_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "10388888",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_InvalidRange_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "777779000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_InvalidCharNEw_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1jjjj|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Create_Syntax_SpecialChar_Positive(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
	       cvl.CVLEditConfigData{
                        cvl.VALIDATE_ALL,
                        cvl.OP_CREATE,
                        "ACL_TABLE|TestACL1",
                        map[string]string{
                                "stage": "INGRESS",
                                "type":  "MIRROR",
                        },
                },
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	cfgData = []cvl.CVLEditConfigData{
	       cvl.CVLEditConfigData{
                        cvl.VALIDATE_NONE,
                        cvl.OP_CREATE,
                        "ACL_TABLE|TestACL1",
                        map[string]string{
                                "stage": "INGRESS",
                                "type":  "MIRROR",
                        },
                },
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule@##",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err = cv.ValidateEditConfig(cfgData)

	if err != cvl.CVL_SUCCESS { //Should succeed
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Create_Syntax_InvalidKeyName_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"AC&&***L_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Create_Semantic_AdditionalInvalidNode_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
				"extra":             "shhs",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Create_Semantic_MissingMandatoryNode_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"VLAN|Vlan101",
			map[string]string{
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}
}

func TestValidateEditConfig_Create_Syntax_Invalid_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULERule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Create_Syntax_IncompleteKey_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Create_Syntax_InvalidKey_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Update_Syntax_DependentData_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_NONE,
			"MIRROR_SESSION|everflow",
			map[string]string{
				"src_ip": "10.1.0.32",
				"dst_ip": "2.2.2.2",
			},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"ACL_RULE|MyACL11_ACL_IPV4|RULE_1",
			map[string]string{
				"MIRROR_ACTION": "everflow",
			},
		},
	}

	cvlErrObj, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrObj)
	}

}

func TestValidateEditConfig_Create_Syntax_DependentData_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_NONE,
			"PORTCHANNEL|ch1",
			map[string]string{
				"admin_status": "up",
				"mtu":          "9100",
			},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_NONE,
			"PORTCHANNEL|ch2",
			map[string]string{
				"admin_status": "up",
				"mtu":          "9100",
			},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_NONE,
			"PORTCHANNEL_MEMBER|ch1|Ethernet4",
			map[string]string{},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_NONE,
			"PORTCHANNEL_MEMBER|ch1|Ethernet8",
			map[string]string{},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_NONE,
			"PORTCHANNEL_MEMBER|ch2|Ethernet12",
			map[string]string{},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_NONE,
			"PORTCHANNEL_MEMBER|ch2|Ethernet16",
			map[string]string{},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_NONE,
			"PORTCHANNEL_MEMBER|ch2|Ethernet20",
			map[string]string{},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"VLAN|Vlan1001",
			map[string]string{
				"vlanid":   "1002",
				"members@": "Ethernet24,ch1,Ethernet8",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Delete_Syntax_InvalidKey_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Update_Syntax_InvalidKey_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Delete_InvalidKey_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"ACL_RULE|TestACL1:Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrObj, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrObj)
	}

}

func TestValidateEditConfig_Update_Syntax_Invalid_Field_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "CATCH",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Update_Semantic_Invalid_Key_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"ACL_RULE|TestACL1Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103uuuu",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Delete_Semantic_Positive(t *testing.T) {
	depDataMap := map[string]interface{}{
		"MIRROR_SESSION": map[string]interface{}{
			"everflow": map[string]interface{}{
				"src_ip": "10.1.0.32",
				"dst_ip": "2.2.2.2",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"MIRROR_SESSION|everflow",
			map[string]string{},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, depDataMap)

}

func TestValidateEditConfig_Delete_Semantic_MissingKey_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"MIRROR_SESSION|everflow0",
			map[string]string{},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"ACL_RULE|MyACL11_ACL_IPV4|RULE_1",
			map[string]string{},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Update_Semantic_MissingKey_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"ACL_RULE|TestACL177|Rule1",
			map[string]string{
				"MIRROR_ACTION": "everflow",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Create_Duplicate_Key_Negative(t *testing.T) {
	depDataMap := map[string]interface{}{
		"ACL_TABLE": map[string]interface{} {
			"TestACL100": map[string]interface{} {
				"stage": "INGRESS",
				"type": "L3",
			},
		},
	}

	//Load same key in DB
	loadConfigDB(rclient, depDataMap)

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_TABLE|TestACL100",
			map[string]string{
				"stage": "INGRESS",
				"type":  "L3",
			},
		},
	}

	cvlErrInfo, retCode := cv.ValidateEditConfig(cfgData)

	if retCode == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, depDataMap)
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

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"ACL_TABLE|TestACL1",
			map[string]string{
				"stage": "INGRESS",
				"type":  "MIRROR",
			},
		},
	}

	cvlErrInfo, retCode := cv.ValidateEditConfig(cfgData)

	if retCode != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, mpi_acl_table_map)

}

/* API to test edit config with valid syntax. */
func TestValidateConfig_Update_Semantic_Vlan_Negative(t *testing.T) {

	cv, _ := cvl.ValidationSessOpen()

	jsonData := `{
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

	if err == cvl.CVL_SUCCESS { //Expected semantic failure
		t.Errorf("Config Validation failed -- error details.")
	}

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

	depDataMap := map[string]interface{}{
		"MIRROR_SESSION": map[string]interface{}{
			"everflow2": map[string]interface{}{
				"src_ip": "10.1.0.32",
				"dst_ip": "2.2.2.2",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)

	/* ACL and Rule name pre-created . */
	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"ACL_RULE|TestACL13|Rule1",
			map[string]string{
				"MIRROR_ACTION": "everflow2",
			},
		},
	}

	cvlErrInfo, retCode := cv.ValidateEditConfig(cfgData)

	if retCode != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, mpi_acl_table_map)
	unloadConfigDB(rclient, mpi_acl_table_rule)
	unloadConfigDB(rclient, depDataMap)

}

func TestValidateEditConfig_Update_Syntax_DependentData_Invalid_Op_Seq(t *testing.T) {

	/* ACL and Rule name pre-created . */
	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_CREATE,
			"ACL_TABLE|TestACL1",
			map[string]string{
				"stage": "INGRESS",
				"type":  "MIRROR",
			},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION": "DROP",
				"L4_SRC_PORT":   "781",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS { //Validation should fail
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Update_Syntax_DependentData_Redis_Negative(t *testing.T) {

	/* ACL does not exist.*/
	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_UPDATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"MIRROR_ACTION": "everflow0",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

/* Create with User provideddependent data. */
func TestValidateEditConfig_Create_Syntax_DependentData_Redis_Positive(t *testing.T) {

	/* ACL and Rule name pre-created . */
	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_TABLE|TestACL22",
			map[string]string{
				"stage": "INGRESS",
				"type":  "MIRROR",
			},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	cfgData = []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL22|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	if err != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}
}

/* Delete Non-Existing Key.*/
func TestValidateEditConfig_Delete_Semantic_ACLTableReference_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"ACL_RULE|MyACL11_ACL_IPV4|RULE_1",
			map[string]string{},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

/* Delete Existing Key.*/
func TestValidateEditConfig_Delete_Semantic_ACLTableReference_Positive(t *testing.T) {

	depDataMap := map[string]interface{} {
		"ACL_TABLE" : map[string]interface{} {
			"TestACL10": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
			},
		},
		"ACL_RULE": map[string]interface{} {
			"TestACL10|Rule1": map[string] interface{} {
				"PACKET_ACTION": "FORWARD",
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

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"ACL_RULE|TestACL10|Rule1",
			map[string]string{},
		},
	}

	cvlErrInfo, err := cv.ValidateEditConfig(cfgData)

	if err != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_Create_Dependent_CacheData(t *testing.T) {

	cvSess, _ := cvl.ValidationSessOpen()

	//Create ACL rule
	cfgDataAcl := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_TABLE|TestACL14",
			map[string]string{
				"stage": "INGRESS",
				"type":  "MIRROR",
			},
		},
	}

	cvlErrInfo, err1 := cvSess.ValidateEditConfig1(cfgDataAcl)
	fmt.Println(cvlErrInfo)

	//Create ACL rule
	cfgDataRule := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL14|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err2 := cvSess.ValidateEditConfig1(cfgDataRule)
	fmt.Println(cvlErrInfo)

	if err1 != cvl.CVL_SUCCESS || err2 != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}
	cvl.ValidationSessClose(cvSess)
}

func TestValidateEditConfig_Create_DepData_In_MultiSess(t *testing.T) {

	//Create ACL rule - Session 1
	cvSess, _ := cvl.ValidationSessOpen()
	cfgDataAcl := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_TABLE|TestACL16",
			map[string]string{
				"stage": "INGRESS",
				"type":  "MIRROR",
			},
		},
	}

	cvlErrInfo, err1 := cvSess.ValidateEditConfig1(cfgDataAcl)
	fmt.Println(cvlErrInfo)

	cvl.ValidationSessClose(cvSess)

	//Create ACL rule - Session 2, validation should fail
	cvSess, _ = cvl.ValidationSessOpen()
	cfgDataRule := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL16|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlSessErr, err2 := cvSess.ValidateEditConfig1(cfgDataRule)

	fmt.Println("Session Info", cvlSessErr)

	cvl.ValidationSessClose(cvSess)

	if err1 != cvl.CVL_SUCCESS || err2 == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

}

func TestValidateEditConfig_Create_DepData_From_Redis_Negative11(t *testing.T) {

	depDataMap := map[string]interface{}{
		"ACL_TABLE": map[string]interface{}{
			"TestACL1": map[string]interface{}{
				"stage": "INGRESS",
				"type":  "MIRROR",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)

	//Create ACL rule - Session 2
	cvSess, _ := cvl.ValidationSessOpen()
	cfgDataRule := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL188|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cvSess.ValidateEditConfig1(cfgDataRule)

	fmt.Println(cvlErrInfo)

        /* Check for error message. */	
	if (err == cvl.CVL_SEMANTIC_DEPENDENT_DATA_MISSING) {
		fmt.Println(cvl.GetErrorString(err))
	}


	cvl.ValidationSessClose(cvSess)

	if (err != cvl.CVL_SEMANTIC_DEPENDENT_DATA_MISSING) {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, depDataMap)
}

func TestValidateEditConfig_Create_DepData_From_Redis(t *testing.T) {

	depDataMap := map[string]interface{}{
		"ACL_TABLE": map[string]interface{}{
			"TestACL1": map[string]interface{}{
				"stage": "INGRESS",
				"type":  "MIRROR",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)

	//Create ACL rule - Session 2
	cvSess, _ := cvl.ValidationSessOpen()
	cfgDataRule := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	cvlErrInfo, err := cvSess.ValidateEditConfig1(cfgDataRule)

	cvl.ValidationSessClose(cvSess)

	if err != cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}

	unloadConfigDB(rclient, depDataMap)
}


func TestValidateEditConfig_Create_Syntax_InvalidErrAppTag_Negative(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"VLAN|Vlan1001",
			map[string]string{
				"vlanid":   "1002",
				"members@": "Ethernet24,Ethernet8",
			},
		},
	}

	cvlErrInfo, retCode := cv.ValidateEditConfig(cfgData)

	/* Compare expected error details and error tag. */
	if compareErrorDetails(cvlErrInfo, cvl.CVL_SEMANTIC_DEPENDENT_DATA_MISSING ,"vlan-invalid", "") != true {
		t.Errorf("Config Validation failed -- error details %v %v", cvlErrInfo, retCode)
	}

}

/* API to test edit config with valid syntax. */
func TestValidateEditConfig_Create_Syntax_InValid_FieldValue(t *testing.T) {

	cfgData := []cvl.CVLEditConfigData{
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_UPDATE,
			"ACL_TABLE|TestACL1",
			map[string]string{
				"stage": "INGRESS",
				"type":  "MIRROR",
			},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_NONE,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL1|Rule1",
			map[string]string{
				"PACKET_ACTION":     "FORWARD",
				"SRC_IP":            "10.1.1.1/32",
				"L4_SRC_PORT":       "1909",
				"IP_PROTOCOL":       "103",
				"DST_IP":            "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
		cvl.CVLEditConfigData{
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"ACL_RULE|TestACL1",
			map[string]string{},
		},
	}

	cvlErrInfo, retCode := cv.ValidateEditConfig(cfgData)

	if retCode == cvl.CVL_SUCCESS {
		t.Errorf("Config Validation failed -- error details %v", cvlErrInfo)
	}
}

//EditConfig(Create) with dependent data from redis
func TestValidateEditConfig_Create_DepData_From_Redis_Negative(t *testing.T) {

	depDataMap := map[string]interface{} {
		"ACL_TABLE" : map[string]interface{} {
			"TestACL1": map[string] interface{} {
				"stage": "INGRESS",
				"type": "MIRROR",
			},
		},
	}

	loadConfigDB(rclient, depDataMap)

	cfgDataRule := []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_CREATE,
			"ACL_RULE|TestACL2|Rule1",
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

	_, err := cv.ValidateEditConfig1(cfgDataRule)

	if err == cvl.CVL_SUCCESS { //should not succeed
		t.Errorf("Config Validation should fail.")
	}

	unloadConfigDB(rclient, depDataMap)
}

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

	_, err := cv.ValidateEditConfig1(cfgDataVlan)

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
				"SRC_IP": "10.1.1.1/32",
				"L4_SRC_PORT": "1909",
				"IP_PROTOCOL": "103",
				"DST_IP": "20.2.2.2/32",
				"L4_DST_PORT_RANGE": "9000-12000",
			},
		},
	}

	_, err = cv.ValidateEditConfig1(cfgDataAclRule)
	if err != cvl.CVL_SUCCESS { //should succeed
		t.Errorf("Config Validation failed.")
	}

	unloadConfigDB(rclient, depDataMap)
}

//EditConfig(Delete) deleting entry already used by other table as leafref
func TestValidateEditConfig_Delete_Dep_Leafref_Negative(t *testing.T) {
	depDataMap := map[string]interface{} {
		"ACL_TABLE" : map[string]interface{} {
			"TestACL1": map[string] interface{} {
				"stage": "INGRESS",
				"type": "L3",
			},
		},
		"ACL_RULE": map[string]interface{} {
			"TestACL1|Rule1": map[string] interface{} {
				"PACKET_ACTION": "FORWARD",
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

	cfgDataVlan := []cvl.CVLEditConfigData {
		cvl.CVLEditConfigData {
			cvl.VALIDATE_ALL,
			cvl.OP_DELETE,
			"ACL_TABLE|TestACL1",
			map[string]string {
			},
		},
	}

	_, err := cv.ValidateEditConfig1(cfgDataVlan)

	if err != cvl.CVL_SEMANTIC_ERROR { //should be semantic failure
		t.Errorf("Config Validation failed.")
	}

	unloadConfigDB(rclient, depDataMap)
}

//EditConfig(Create) with chained leafref from redis
func TestValidateEditConfig_Create_Chained_Leafref_DepData_Negative(t *testing.T) {
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
				"ports@":"Ethernet2",
			},
		},
	}

	//Prepare data in Redis
	loadConfigDB(rclient, depDataMap)

	cfgDataAclRule :=  []cvl.CVLEditConfigData {
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

	_, err := cv.ValidateEditConfig1(cfgDataAclRule)
	if err != cvl.CVL_SUCCESS { //should succeed
		t.Errorf("Config Validation failed.")
	}

	unloadConfigDB(rclient, depDataMap)
}

