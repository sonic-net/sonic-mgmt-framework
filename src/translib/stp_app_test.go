////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package translib

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	db "translib/db"
)

func init() {
	fmt.Println("+++++  Init stp_app_test  +++++")

	if err := clearStpDataFromConfigDb(); err == nil {
		fmt.Println("+++++  Removed All Stp Data from Db  +++++")
	} else {
		fmt.Printf("Failed to remove All Stp Data from Db: %v", err)
	}
}

// This will Test PVST mode enable/disable
func Test_StpApp_Pvst_Enable_Disable(t *testing.T) {
	topStpUrl := "/openconfig-spanning-tree:stp"
	enableStpUrl := topStpUrl + "/global"

	t.Run("Empty_Response_Top_Level", processGetRequest(topStpUrl, "", true))

	t.Run("Enable_PVST_Mode", processSetRequest(enableStpUrl, enablePVSTModeJsonRequest, "POST", false))
	t.Run("Verify_PVST_Mode_Configured", processGetRequest(enableStpUrl+"/state", pvstmodeVerifyJsonResponse, false))
	t.Run("Disable_PVST_Mode", processDeleteRequest(topStpUrl))

	t.Run("Verify_Empty_Response_Top_Level", processGetRequest(topStpUrl, "", true))
}

// This will Test RAPID PVST mode enable/disable
func Test_StpApp_Rapid_Pvst_Enable_Disable(t *testing.T) {
	topStpUrl := "/openconfig-spanning-tree:stp"
	enableStpUrl := topStpUrl + "/global"

	t.Run("Empty_Response_Top_Level", processGetRequest(topStpUrl, "", true))

	t.Run("Enable_Rapid_PVST_Mode", processSetRequest(enableStpUrl, enableRapidPVSTModeJsonRequest, "POST", false))
	t.Run("Verify_Rapid_PVST_Mode_Configured", processGetRequest(enableStpUrl+"/state", rapidPvstmodeVerifyJsonResponse, false))
	t.Run("Disable_Rapid_PVST_Mode", processDeleteRequest(topStpUrl))

	t.Run("Verify_Empty_Response_Top_Level", processGetRequest(topStpUrl, "", true))
}

func Test_StpApp_TopLevelPathInPvstMode(t *testing.T) {
	topStpUrl := "/openconfig-spanning-tree:stp"
	enableStpUrl := topStpUrl + "/global"
	vlanUrl := "/openconfig-interfaces:interfaces/interface[name=Vlan4090]"
	createswitchportUrl := "/openconfig-interfaces:interfaces/interface[name=Ethernet28]/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/config"
	deleteSwitchportUrl := createswitchportUrl + "/trunk-vlans[trunk-vlans=4090]"

	t.Run("Empty_Response_Top_Level", processGetRequest(topStpUrl, "", true))

	t.Run("Create_Single_Vlan", processSetRequest(vlanUrl, emptyJson, "PATCH", false))
	t.Run("Create_Switchport", processSetRequest(createswitchportUrl, switchportCreateJsonRequest, "PATCH", false))

	t.Run("Enable_PVST_Mode", processSetRequest(enableStpUrl, enablePVSTModeJsonRequest, "POST", false))
	t.Run("Verify_Full_Stp_Top_Level", processGetRequest(topStpUrl, topLevelPvstModeVerifyJsonResponse, false))
	t.Run("Disable_PVST_Mode", processDeleteRequest(topStpUrl))

	t.Run("Delete_Switchport", processDeleteRequest(deleteSwitchportUrl))
	t.Run("Delete_Single_Vlan", processDeleteRequest(vlanUrl))

	t.Run("Verify_Disable_PVST_Mode", processGetRequest(topStpUrl, "", true))
}

func Test_StpApp_TopLevelPathInRapidPvstMode(t *testing.T) {
	topStpUrl := "/openconfig-spanning-tree:stp"
	enableStpUrl := topStpUrl + "/global"
	vlanUrl := "/openconfig-interfaces:interfaces/interface[name=Vlan4090]"
	createswitchportUrl := "/openconfig-interfaces:interfaces/interface[name=Ethernet28]/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/config"
	deleteSwitchportUrl := createswitchportUrl + "/trunk-vlans[trunk-vlans=4090]"

	t.Run("Empty_Response_Top_Level", processGetRequest(topStpUrl, "", true))

	t.Run("Create_Single_Vlan", processSetRequest(vlanUrl, emptyJson, "PATCH", false))
	t.Run("Create_Switchport", processSetRequest(createswitchportUrl, switchportCreateJsonRequest, "PATCH", false))

	t.Run("Enable_Rapid_PVST_Mode", processSetRequest(enableStpUrl, enableRapidPVSTModeJsonRequest, "POST", false))
	t.Run("Verify_Full_Stp_Top_Level", processGetRequest(topStpUrl, topLevelRapidPvstModeVerifyJsonResponse, false))
	t.Run("Disable_Rapid_PVST_Mode", processDeleteRequest(topStpUrl))

	t.Run("Delete_Switchport", processDeleteRequest(deleteSwitchportUrl))
	t.Run("Delete_Single_Vlan", processDeleteRequest(vlanUrl))

	t.Run("Verify_Disable_Rapid_PVST_Mode", processGetRequest(topStpUrl, "", true))
}

func clearStpDataFromConfigDb() error {
	var err error
	stpGlobalTbl := db.TableSpec{Name: "STP"}
	stpVlanTbl := db.TableSpec{Name: "STP_VLAN"}
	stpVlanIntfTbl := db.TableSpec{Name: "STP_VLAN_INTF"}
	stpIntfTbl := db.TableSpec{Name: "STP_INTF"}

	d := getConfigDb()
	if d == nil {
		err = errors.New("Failed to connect to config Db")
		return err
	}

	if err = d.DeleteTable(&stpVlanIntfTbl); err != nil {
		err = errors.New("Failed to delete STP Vlan Intf Table")
		return err
	}

	if err = d.DeleteTable(&stpIntfTbl); err != nil {
		err = errors.New("Failed to delete STP Intf Table")
		return err
	}

	if err = d.DeleteTable(&stpVlanTbl); err != nil {
		err = errors.New("Failed to delete STP Vlan Table")
		return err
	}

	if err = d.DeleteTable(&stpGlobalTbl); err != nil {
		err = errors.New("Failed to delete STP Global Table")
		return err
	}

	/* Temporary
	if err = d.DeleteTable(&db.TableSpec{Name: "PORTCHANNEL_MEMBER"}); err != nil {
		err = errors.New("Failed to delete PORTCHANNEL_MEMBER Table")
		return err
	}
	if err = d.DeleteTable(&db.TableSpec{Name: "PORTCHANNEL"}); err != nil {
		err = errors.New("Failed to delete PORTCHANNEL Table")
		return err
	}
	if err = d.DeleteTable(&db.TableSpec{Name: "VLAN_MEMBER"}); err != nil {
		err = errors.New("Failed to delete VLAN_MEMBER Table")
		return err
	}
	if err = d.DeleteTable(&db.TableSpec{Name: "VLAN"}); err != nil {
		err = errors.New("Failed to delete VLAN Table")
		return err
	}
	*/

	return err
}

func processGetRequestToFile(url string, expectedRespJson string, errorCase bool) func(*testing.T) {
	return func(t *testing.T) {
		response, err := Get(GetRequest{Path:url})
		if err != nil && !errorCase {
			t.Errorf("Error %v received for Url: %s", err, url)
		}

		respJson := response.Payload
		err = ioutil.WriteFile("/tmp/TmpResp.json", respJson, 0644)
		if err != nil {
			fmt.Println(err)
		}
		if string(respJson) != expectedRespJson {
			t.Errorf("Response for Url: %s received is not expected:\n%s", url, string(respJson))
		}
	}
}

/***************************************************************************/
///////////                  JSON Data for Tests              ///////////////
/***************************************************************************/

var switchportCreateJsonRequest string = "{\"openconfig-vlan:config\": {\"trunk-vlans\": [4090], \"interface-mode\": \"TRUNK\"}}"

var enablePVSTModeJsonRequest string = "{ \"openconfig-spanning-tree:config\": { \"enabled-protocol\": [ \"PVST\" ], \"openconfig-spanning-tree-ext:rootguard-timeout\": 401, \"openconfig-spanning-tree-ext:hello-time\": 7, \"openconfig-spanning-tree-ext:max-age\": 16, \"openconfig-spanning-tree-ext:forwarding-delay\": 22, \"openconfig-spanning-tree-ext:bridge-priority\": 20480 }}"

var pvstmodeVerifyJsonResponse string = "{\"openconfig-spanning-tree:state\":{\"bpdu-filter\":false,\"enabled-protocol\":[\"openconfig-spanning-tree-ext:PVST\"],\"openconfig-spanning-tree-ext:bridge-priority\":20480,\"openconfig-spanning-tree-ext:forwarding-delay\":22,\"openconfig-spanning-tree-ext:hello-time\":7,\"openconfig-spanning-tree-ext:max-age\":16,\"openconfig-spanning-tree-ext:rootguard-timeout\":401}}"

var topLevelPvstModeVerifyJsonResponse string = "{\"openconfig-spanning-tree:stp\":{\"global\":{\"config\":{\"bpdu-filter\":false,\"enabled-protocol\":[\"openconfig-spanning-tree-ext:PVST\"],\"openconfig-spanning-tree-ext:bridge-priority\":20480,\"openconfig-spanning-tree-ext:forwarding-delay\":22,\"openconfig-spanning-tree-ext:hello-time\":7,\"openconfig-spanning-tree-ext:max-age\":16,\"openconfig-spanning-tree-ext:rootguard-timeout\":401},\"state\":{\"bpdu-filter\":false,\"enabled-protocol\":[\"openconfig-spanning-tree-ext:PVST\"],\"openconfig-spanning-tree-ext:bridge-priority\":20480,\"openconfig-spanning-tree-ext:forwarding-delay\":22,\"openconfig-spanning-tree-ext:hello-time\":7,\"openconfig-spanning-tree-ext:max-age\":16,\"openconfig-spanning-tree-ext:rootguard-timeout\":401}},\"interfaces\":{\"interface\":[{\"config\":{\"bpdu-guard\":false,\"guard\":\"NONE\",\"name\":\"Ethernet28\",\"openconfig-spanning-tree-ext:bpdu-guard-port-shutdown\":false,\"openconfig-spanning-tree-ext:portfast\":true,\"openconfig-spanning-tree-ext:spanning-tree-enable\":true,\"openconfig-spanning-tree-ext:uplink-fast\":false},\"name\":\"Ethernet28\",\"state\":{\"bpdu-guard\":false,\"guard\":\"NONE\",\"name\":\"Ethernet28\",\"openconfig-spanning-tree-ext:bpdu-guard-port-shutdown\":false,\"openconfig-spanning-tree-ext:spanning-tree-enable\":true,\"openconfig-spanning-tree-ext:uplink-fast\":false}}]},\"openconfig-spanning-tree-ext:pvst\":{\"vlan\":[{\"config\":{\"bridge-priority\":20480,\"forwarding-delay\":22,\"hello-time\":7,\"max-age\":16,\"spanning-tree-enable\":true,\"vlan-id\":4090},\"state\":{\"bridge-priority\":20480,\"vlan-id\":4090},\"vlan-id\":4090}]}}}"

var enableRapidPVSTModeJsonRequest string = "{ \"openconfig-spanning-tree:config\": { \"enabled-protocol\": [ \"RAPID_PVST\" ], \"openconfig-spanning-tree-ext:rootguard-timeout\": 305, \"openconfig-spanning-tree-ext:hello-time\": 4, \"openconfig-spanning-tree-ext:max-age\": 10, \"openconfig-spanning-tree-ext:forwarding-delay\": 25, \"openconfig-spanning-tree-ext:bridge-priority\": 20480 }}"

var rapidPvstmodeVerifyJsonResponse string = "{\"openconfig-spanning-tree:state\":{\"bpdu-filter\":false,\"enabled-protocol\":[\"openconfig-spanning-tree-types:RAPID_PVST\"],\"openconfig-spanning-tree-ext:bridge-priority\":20480,\"openconfig-spanning-tree-ext:forwarding-delay\":25,\"openconfig-spanning-tree-ext:hello-time\":4,\"openconfig-spanning-tree-ext:max-age\":10,\"openconfig-spanning-tree-ext:rootguard-timeout\":305}}"

var topLevelRapidPvstModeVerifyJsonResponse string = "{\"openconfig-spanning-tree:stp\":{\"global\":{\"config\":{\"bpdu-filter\":false,\"enabled-protocol\":[\"openconfig-spanning-tree-types:RAPID_PVST\"],\"openconfig-spanning-tree-ext:bridge-priority\":20480,\"openconfig-spanning-tree-ext:forwarding-delay\":25,\"openconfig-spanning-tree-ext:hello-time\":4,\"openconfig-spanning-tree-ext:max-age\":10,\"openconfig-spanning-tree-ext:rootguard-timeout\":305},\"state\":{\"bpdu-filter\":false,\"enabled-protocol\":[\"openconfig-spanning-tree-types:RAPID_PVST\"],\"openconfig-spanning-tree-ext:bridge-priority\":20480,\"openconfig-spanning-tree-ext:forwarding-delay\":25,\"openconfig-spanning-tree-ext:hello-time\":4,\"openconfig-spanning-tree-ext:max-age\":10,\"openconfig-spanning-tree-ext:rootguard-timeout\":305}},\"interfaces\":{\"interface\":[{\"config\":{\"bpdu-guard\":false,\"guard\":\"NONE\",\"name\":\"Ethernet28\",\"openconfig-spanning-tree-ext:bpdu-guard-port-shutdown\":false,\"openconfig-spanning-tree-ext:portfast\":true,\"openconfig-spanning-tree-ext:spanning-tree-enable\":true,\"openconfig-spanning-tree-ext:uplink-fast\":false},\"name\":\"Ethernet28\",\"state\":{\"bpdu-guard\":false,\"guard\":\"NONE\",\"name\":\"Ethernet28\",\"openconfig-spanning-tree-ext:bpdu-guard-port-shutdown\":false,\"openconfig-spanning-tree-ext:spanning-tree-enable\":true,\"openconfig-spanning-tree-ext:uplink-fast\":false}}]},\"rapid-pvst\":{\"vlan\":[{\"config\":{\"bridge-priority\":20480,\"forwarding-delay\":25,\"hello-time\":4,\"max-age\":10,\"openconfig-spanning-tree-ext:spanning-tree-enable\":true,\"vlan-id\":4090},\"state\":{\"bridge-priority\":20480,\"vlan-id\":4090},\"vlan-id\":4090}]}}}"
