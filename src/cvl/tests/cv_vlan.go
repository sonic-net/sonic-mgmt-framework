package main


import (
	"fmt"
	"os"
	"time"
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
				"PORTCHANNEL|ch1",
				map[string]string {
					"admin_status": "up",
					"mtu": "9100",
				},
			},
			cvl.KeyData {
				false,
				"PORTCHANNEL|ch2",
				map[string]string {
					"admin_status": "up",
					"mtu": "9100",
				},
			},
			cvl.KeyData {
				false,
				"PORTCHANNEL_MEMBER|ch1|Ethernet4",
				map[string]string {
				},
			},
			cvl.KeyData {
				false,
				"PORTCHANNEL_MEMBER|ch1|Ethernet8",
				map[string]string {
				},
			},
			cvl.KeyData {
				false,
				"PORTCHANNEL_MEMBER|ch2|Ethernet12",
				map[string]string {
				},
			},
			cvl.KeyData {
				false,
				"PORTCHANNEL_MEMBER|ch2|Ethernet16",
				map[string]string {
				},
			},
			cvl.KeyData {
				false,
				"PORTCHANNEL_MEMBER|ch2|Ethernet20",
				map[string]string {
				},
			},
			cvl.KeyData {
				true,
				"VLAN|Vlan1001",
				map[string]string {
					"vlanid": "1001",
					"members@": "Ethernet24,ch1,Ethernet8",
				},
			},
		}

		fmt.Printf("\nValidating data for must = %v\n\n", keyData);

		err := cvl.ValidateCreate(keyData)

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}

		return
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


		err := cvl.ValidateConfig(jsonData)

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


		err := cvl.ValidateConfig(jsonData)

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


		err := cvl.ValidateConfig(jsonData)

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


		err := cvl.ValidateConfig(jsonData)

		fmt.Printf("\nValidating data for internal dependency check = %v\n\n", jsonData);

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
		}`

		err := cvl.ValidateConfig(jsonData)

		fmt.Printf("\nValidating data for external dependency check = %v\n\n", jsonData);

		if (err == cvl.CVL_SUCCESS) {
			fmt.Printf("\nConfig Validation succeeded.\n\n");
		} else {
			fmt.Printf("\nConfig Validation failed.\n\n");
		}
	}

	cvl.Finish()

	fmt.Printf("\n\n\n Time taken for %v requests = %v\n", count, time.Since(start))
}
