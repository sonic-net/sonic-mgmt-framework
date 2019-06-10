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


	err := cvl.ValidateConfig(jsonData)

	fmt.Printf("\nValidating data = %v\n\n", jsonData);

	if (err == cvl.CVL_SUCCESS) {
		fmt.Printf("\nConfig Validation succeeded.\n\n");
	} else {
		fmt.Printf("\nConfig Validation failed.\n\n");
	}

	keyData := make([]cvl.KeyData, 4)
	keyData[0].Key = "ACL_TABLE|MyACL55_ACL_IPV4"
	keyData[0].Data = make(map[string]string)
	keyData[0].Data["stage"] =  "INGRESS"
	keyData[0].Data["type"] =  "l3"

	keyData[1].Key = "ACL_RULE|MyACL55_ACL_IPV4|RULE_1"
	keyData[1].Data = make(map[string]string)
	keyData[1].Data["packet_action"] = "forward"
	keyData[1].Data["ip_protocol"] = "ip"
	keyData[1].Data["src_ip"] = "10.1.1.1/32"
	keyData[1].Data["dst_ip"] = "20.2.2.2/32"

	keyData[2].Key = "ACL_TABLE|MyACL11_ACL_IPV4"
	keyData[2].Data = make(map[string]string)
	keyData[2].Data["stage"] =  "INGRESS"

	keyData[3].Key = "VLAN|Vlan901"
	keyData[3].Data = make(map[string]string)
	keyData[3].Data["members"] =  "Ethernet8"
	keyData[3].Data["vlanid"] =  "901"

	fmt.Printf("\n\n\n  cvl.ValidateCreate() = %d\n", cvl.ValidateCreate(keyData))

	keyData1 := make([]cvl.KeyData, 3)
	keyData1[0].Key = "ACL_TABLE|MyACL11_ACL_IPV4"
	keyData1[0].Data = make(map[string]string)
	keyData1[0].Data["stage"] =  "INGRESS"
	keyData1[0].Data["type"] =  "l3"

	keyData1[1].Key = "ACL_RULE|MyACL11_ACL_IPV4|RULE_1"
	keyData1[1].Data = make(map[string]string)
	keyData1[1].Data["packet_action"] = "forward"
	keyData1[1].Data["ip_protocol"] = "ip"
	keyData1[1].Data["src_ip"] = "10.1.1.1/32"
	keyData1[1].Data["dst_ip"] = "20.2.2.2/32"

	keyData1[2].Key = "ACL_TABLE|MyACL33_ACL_IPV4"
	keyData1[2].Data = make(map[string]string)
	keyData1[2].Data["stage"] =  "INGRESS"

	fmt.Printf("\n\n\n  cvl.ValidateUpdate() = %d\n", cvl.ValidateUpdate(keyData1))


	keyData2 := make([]cvl.KeyData, 3)
	keyData2[0].Key = "ACL_TABLE|MyACL11_ACL_IPV4"
	keyData2[0].Data = make(map[string]string)

	keyData2[1].Key = "ACL_RULE|MyACL11_ACL_IPV4|RULE_1"
	keyData2[1].Data = make(map[string]string)

	keyData2[2].Key = "ACL_TABLE|MyACL33_ACL_IPV4"
	keyData2[2].Data = make(map[string]string)

	fmt.Printf("\n\n\n  cvl.ValidateDelete() = %d\n", cvl.ValidateDelete(keyData2))


	cvl.Finish()
	fmt.Printf("\n\n\n Time taken = %v\n", time.Since(start))
}
