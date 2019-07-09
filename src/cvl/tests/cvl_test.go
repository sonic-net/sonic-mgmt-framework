package cvl_test

import (
	"cvl"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

/* Converts JSON Data in a File to Map. */
func convertFileJsonToMap(t *testing.T, fileName string) map[string]string {

	jsonData := convertFileJsonToString(t, fileName)

	byteData := []byte(jsonData)

	var mapstr map[string]string
	err1 := json.Unmarshal(byteData, &mapstr)

	if err1 != nil {
		fmt.Println("error:", err1)
	}

	return mapstr

}

/* Converts JSON Data in a File to String. */
func convertFileJsonToString(t *testing.T, fileName string) string {
	var jsonData string
	b, e := ioutil.ReadFile(fileName)
	if e != nil {
		fmt.Printf("\nFailed to read data file : %v\n", e)
	} else {
		jsonData = string(b)
	}

	return jsonData
}

func TestValidateConfig(t *testing.T) {
	APIValidateConfigEachFileBased(t)
	APIValidateConfigInline(t)
}

/* ValidateConfig with user input in file. */
func APIValidateConfigEachFileBased(t *testing.T) {
	tests := []struct {
		filedescription string
		fileName        string
		retCode         cvl.CVLRetCode
	}{
		{filedescription: "DEVICE_METADATA", fileName: "./device.json", retCode: cvl.CVL_SUCCESS},
	}
	for _, tc := range tests {
		t.Run(tc.filedescription, func(t *testing.T) {
			jsonString := convertFileJsonToString(t, tc.fileName)
			err := cvl.ValidateConfig(jsonString)
			fmt.Printf("\nValidating data = %v\n\n", jsonString)

			if err != tc.retCode {
				t.Errorf("Config Validation failed.")
			}

		})
	}

}

/* ValidateConfig with user input inline. */
func APIValidateConfigInline(t *testing.T) {
	jsonData := `{
		  	 "INTERFACE": {
			  "Ethernet668|10.0.0.0/31": {},
				  "Ethernet24|10.0.0.2/31": {},
				  "Ethernet112|10.0.0.4/31": {}
		  }
	  }`

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

	tests := []struct {
		filedescription string
		jsonString      string
		retCode         cvl.CVLRetCode
	}{
		{filedescription: "INTERFACE", jsonString: jsonData, retCode: cvl.CVL_SUCCESS},
	}
	for _, tc := range tests {
		t.Run(tc.filedescription, func(t *testing.T) {
			err := cvl.ValidateConfig(tc.jsonString)

			fmt.Printf("\nValidating data = %v\n\n", jsonData)

			if err != tc.retCode {
				t.Errorf("Config Validation failed.")
			}

		})
	}

}

/* ValidateEditConfig with user input in filr . */
func TestValidateEditConfigFileBased(t *testing.T) {

	tests := []struct {
		filedescription     string
		configStringFile    string
		dependentStringFile string
		retCode             cvl.CVLRetCode
	}{
		{filedescription: "ACL_DATA", configStringFile: "./acltable.json", dependentStringFile: "./aclrule.json", retCode: cvl.CVL_SUCCESS},
	}

	for _, tc := range tests {
		t.Run(tc.filedescription, func(t *testing.T) {
			jsonConfigMap := convertFileJsonToMap(t, tc.configStringFile)
			jsonDependentMap := convertFileJsonToMap(t, tc.dependentStringFile)

			cfgData := []cvl.CVLEditConfigData{
				cvl.CVLEditConfigData{
					cvl.VALIDATE_NONE,
					cvl.OP_NONE,
					"ACL_TABLE|TestACL1",
					jsonDependentMap,
				},
				cvl.CVLEditConfigData{
					cvl.VALIDATE_ALL,
					cvl.OP_CREATE,
					"ACL_RULE|TestACL1|Rule1",
					jsonConfigMap,
				},
			}
			fmt.Printf("\n\n Validating create data = %v\n\n", cfgData)

			err := cvl.ValidateEditConfig(cfgData)

			if err == cvl.CVL_SUCCESS {
				fmt.Printf("\nConfig Validation succeeded.\n\n")
			} else {
				t.Errorf("Config Validation failed.")
			}

		})
	}

}
