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
	"errors"
	"fmt"
	"testing"
	"time"
	db "translib/db"
)

func init() {
	fmt.Println("+++++  Init PortChannel_Intf_test  +++++")
}

func Test_LagIntfOperations(t *testing.T) {

	fmt.Println("+++++  Start PorChannel Testing  +++++")

	url := "/openconfig-interfaces:interfaces/interface[name=PortChannel100]"
	time.Sleep(2 * time.Second)
	t.Run("Create_PortChannel(PATCH)", processSetRequest(url, emptyJson, "PATCH", false))
	time.Sleep(1 * time.Second)
	fmt.Println("+++++  Done PorChannel Creation  +++++")

	stateUrl := url + "/state"
	t.Run("Verify_PortChannel_Create", processGetRequest(stateUrl, intfStateGetJsonResponse, false))

	// Set MTU
	mtuUrl := url + "/config/mtu"
	t.Run("Configure PortChannel MTU", processSetRequest(mtuUrl, mtuJson, "PATCH", false))
	time.Sleep(1 * time.Second)
	fmt.Println("+++++  Done PorChannel MTU configuration  +++++")

	// Get MTU
	stateMtu := stateUrl + "/mtu"
	t.Run("Verify_PortChannel_MTU_Set", processGetRequest(stateMtu, mtuJson, false))

	adminUrl := url + "/config/enabled"
	t.Run("Configure PortChannel admin-status", processSetRequest(adminUrl, enabledJson, "PATCH", false))
	time.Sleep(1 * time.Second)
	fmt.Println("+++++  Done PorChannel admin-status configuration  +++++")

	// Get admin-status
	stateAdmin := stateUrl + "/admin-status"
	t.Run("Verify_PortChannel_AdminStatus_Set", processGetRequest(stateAdmin, adminJson, false))

	aggregationUrl := url + "/openconfig-if-aggregate:aggregation"

	// Set min-links
	minLinksUrl := aggregationUrl + "/config/min-links"
	t.Run("Configure PortChannel min-links", processSetRequest(minLinksUrl, minLinksJson, "PATCH", false))
	fmt.Println("+++++  Done PorChannel min-links configuration  +++++")

	// Get min-links
	minLinksGetUrl := aggregationUrl + "/state/min-links"
	t.Run("Verify_PortChannel_MinLinks_Set", processGetRequest(minLinksGetUrl, minLinksResp, false))

	// Set fallback
	fallbackUrl := aggregationUrl + "/config/dell-intf-augments:fallback"
	t.Run("Configure PortChannel Fallback", processSetRequest(fallbackUrl, fallbackJson, "PATCH", false))

	// Get fallback mode
	fallbackGetUrl := aggregationUrl + "/state/dell-intf-augments:fallback"
	t.Run("Verify_portchannel_Fallback_Set", processGetRequest(fallbackGetUrl, fallbackResp, false))

	// Set IP address
	ipUrl := url + "/subinterfaces/subinterface[index=0]/openconfig-if-ip:ipv4/addresses/address[ip=11.1.1.1]/config"
	t.Run("Configure PortChannel IPv4 address", processSetRequest(ipUrl, ipJson, "PATCH", false))
	fmt.Println("+++++  Done PorChannel IPv4 configuration  +++++")

	// Set IPv6 address
	ip6Url := url + "/subinterfaces/subinterface[index=0]/openconfig-if-ip:ipv6/addresses/address[ip=a::e]/config"
	t.Run("Configure PortChannel IPv6 address", processSetRequest(ip6Url, ip6Json, "PATCH", false))
	fmt.Println("+++++  Done PorChannel IPv6 configuration  +++++")

	// Add member ports
	memUrl := "/openconfig-interfaces:interfaces/interface[name=Ethernet4]" + "/openconfig-if-ethernet:ethernet/config/openconfig-if-aggregate:aggregate-id"
	t.Run("Configure PortChannel Member Add", processSetRequest(memUrl, memAddJson, "PATCH", false))

	// Get member ports
	memGetUrl := aggregationUrl + "/state/member"
	t.Run("Verify_PortChannel_Member_Add", processGetRequest(memGetUrl, memRespJson, false))

	// Remove member ports
	t.Run("PortChannel_Member_Remove", processDeleteRequest(memUrl))
	time.Sleep(1 * time.Second)

	// Delete PortChannel
	t.Run("Delete_PortChannel", processDeleteRequest(url))

	time.Sleep(2 * time.Second)
	t.Run("Verify_PortChannel_Delete", processGetRequest(stateUrl, "", true))
}

// This will delete configs in PortChannel table, PORTCHANNEL_MEMBER table and PORTCHANNEL_INTERFACE table
func clearLagDataFromDb() error {
	var err error
	lagTable := db.TableSpec{Name: "PORTCHANNEL"}
	memTable := db.TableSpec{Name: "PORTCHANNEL_MEMBER"}
	IpTable := db.TableSpec{Name: "PORTCHANNEL_INTERFACE"}

	d := getConfigDb()
	if d == nil {
		err = errors.New("Failed to connect to config Db")
		return err
	}
	if err = d.DeleteTable(&lagTable); err != nil {
		err = errors.New("Failed to clear PORTCHANNELTable")
		return err
	}
	if err = d.DeleteTable(&memTable); err != nil {
		err = errors.New("Failed to clear PORTCHANNEL_MEMBER Table")
		return err
	}
	if err = d.DeleteTable(&IpTable); err != nil {
		err = errors.New("Failed to clear PORTCHANNEL_INTERFACE Table")
		return err
	}
	return err
}

/***************************************************************************/
///////////                  JSON Data for Tests              ///////////////
/***************************************************************************/

var intfStateGetJsonResponse string = "{\"openconfig-interfaces:state\":{\"admin-status\":\"UP\",\"mtu\":9100,\"name\":\"PortChannel100\",\"oper-status\":\"DOWN\"}}"

var mtuJson string = "{\"openconfig-interfaces:mtu\":9000}"

var enabledJson string = "{\"openconfig-interfaces:enabled\":false}"
var adminJson string = "{\"openconfig-interfaces:admin-status\":\"DOWN\"}"

var minLinksJson string = "{\"openconfig-if-aggregate:min-links\":1}"
var minLinksResp string = "{\"min-links\":1}"

var fallbackJson string = "{\"dell-intf-augments:fallback\":true}"
var fallbackResp string = "{\"fallback\":true}"

var ipJson string = "{\"openconfig-if-ip:config\":{\"ip\":\"11.1.1.1\",\"prefix-length\":24}}"
var ip6Json string = "{\"openconfig-if-ip:config\":{\"ip\":\"a::e\",\"prefix-length\":64}}"

var memAddJson string = "{\"openconfig-if-aggregate:aggregate-id\":\"100\"}"
var memRespJson string = "{\"member\":[\"Ethernet4\"]}"
