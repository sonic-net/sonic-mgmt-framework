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

	cv, ret := cvl.ValidatorSessOpen()
	if (ret != cvl.CVL_SUCCESS) {
		fmt.Printf("Could not create CVL session")
		return
	}

	if ((len(os.Args) > 1) && (os.Args[1] == "debug")) {
		cvl.Debug(true)
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

	cvl.ValidatorSessClose(cv)
	cvl.Finish()

	fmt.Printf("\n\n\n Time taken for %v requests = %v\n", count, time.Since(start))
}
