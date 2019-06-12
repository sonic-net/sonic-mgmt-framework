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
	if ((len(os.Args) > 1) && (os.Args[1] == "debug")) {
		cvl.Debug(true)
	}

	{
		count++

		keyData :=  []cvl.KeyData {
			cvl.KeyData {
				false,
				"ACL_TABLE|TestACL1",
				map[string]string {
					"stage": "INGRESS",
					"type": "L3",
				},
			},
			cvl.KeyData {
				true,
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

		fmt.Printf("\n\n%d. Validating create data = %v\n\n", count, keyData);

		err := cvl.ValidateCreate(keyData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++

		keyData :=  []cvl.KeyData {
			cvl.KeyData {
				true,
				"ACL_TABLE|MyACL11_ACL_IPV4",
				map[string]string {
					"stage": "INGRESS",
					"type": "MIRROR",
				},
			},
		}

		fmt.Printf("\n\n%d. Validating update data = %v\n\n", count, keyData);

		err := cvl.ValidateUpdate(keyData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++
		keyData :=  []cvl.KeyData {
			cvl.KeyData {
				true,
				"MIRROR_SESSION|everflow",
				map[string]string {
					"src_ip": "10.1.0.32",
					"dst_ip": "2.2.2.2",
				},
			},
		}

		fmt.Printf("\n\n%d. Validating create data = %v\n\n", count, keyData);
		err := cvl.ValidateCreate(keyData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}

		count++
		keyData = []cvl.KeyData {
			cvl.KeyData {
				false,
				"MIRROR_SESSION|everflow",
				map[string]string {
					"src_ip": "10.1.0.32",
					"dst_ip": "2.2.2.2",
				},
			},
			cvl.KeyData {
				true,
				"ACL_RULE|MyACL11_ACL_IPV4|RULE_1",
				map[string]string {
					"MIRROR_ACTION": "everflow",
				},
			},
		}

		fmt.Printf("\n\n%d. Validating data for update = %v\n\n", count, keyData);

		err = cvl.ValidateUpdate(keyData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}
	{
		count++

		keyData :=  []cvl.KeyData {
			cvl.KeyData {
				true,
				"MIRROR_SESSION|everflow",
				map[string]string {
				},
			},
			cvl.KeyData {
				true,
				"ACL_RULE|MyACL11_ACL_IPV4|RULE_1",
				map[string]string {
				},
			},
		}

		fmt.Printf("\n\n%d. Validating data for delete = %v\n\n", count, keyData);

		err := cvl.ValidateDelete(keyData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}

	cvl.Finish()

	fmt.Printf("\n\n\n Time taken for %v requests = %v\n", count, time.Since(start))
}
