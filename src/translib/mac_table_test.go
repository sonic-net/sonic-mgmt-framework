//////////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Dell, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//////////////////////////////////////////////////////////////////////////

package translib

import (
        "fmt"
        "testing"
)

func init() {
        fmt.Println("+++++   Init MAC_Table__test  +++++")
}

func Test_MacTableOperations(t *testing.T) {
	fmt.Println("+++++  Start MacTable Testing  +++++ ")

	allUrl := "/openconfig-network-instance:network-instances/network-instance[name=default]/fdb/mac-table/entries"
	macAddressVlanUrl := allUrl + "/entry[mac-address=00:00:00:00:00:01][vlan=10]"

	// GET all mac-address
	t.Run("Verify_All_Mac_Address", processGetRequest(allUrl, macAllGetJsonResponse, false))

        // GET mac-entry
        t.Run("Verify__Mac_Address", processGetRequest(macAddressVlanUrl, macVlanMacAddressGetJsonResponse, false))
}

/***************************************************************************/
///////////                  JSON Data for Tests              ///////////////
/***************************************************************************/

var macAllGetJsonResponse string = "{\"openconfig-network-instance:entries\":{\"entry\":[{\"interface\":{\"interface-ref\":{\"state\":{\"interface\":\"Ethernet0\"}}},\"mac-address\":\"00:00:00:00:00:01\",\"state\":{\"entry-type\":\"STATIC\"},\"vlan\":10},{\"interface\":{\"interface-ref\":{\"state\":{\"interface\":\"Ethernet4\"}}},\"mac-address\":\"00:00:00:00:00:02\",\"state\":{\"entry-type\":\"STATIC\"},\"vlan\":20}]}}"

var macVlanMacAddressGetJsonResponse string = "{\"openconfig-network-instance:entry\":[{\"interface\":{\"interface-ref\":{\"state\":{\"interface\":\"Ethernet0\"}}},\"mac-address\":\"00:00:00:00:00:01\",\"state\":{\"entry-type\":\"STATIC\"},\"vlan\":10}]}"
