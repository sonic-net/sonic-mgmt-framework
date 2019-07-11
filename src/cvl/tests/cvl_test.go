package cvl_test

import (
	"cvl"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

type testEditCfgData struct {
	filedescription string
	cfgData         string
	depData         string
	retCode         cvl.CVLRetCode
}

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
func validateEditConfigEachFileBased(t *testing.T) {

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
func validateEditConfigInline(t *testing.T) {

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

func TestValidateConfig(t *testing.T) {
	TestValidateConfig_CfgStrBuffer(t)
	TestValidateConfig_CfgFile(t)
}

/* API to invoke tests for Edit Config. */
func TestValidateEditConfig(t *testing.T) {
	validateEditConfigEachFileBased(t)
	validateEditConfigInline(t)
}
