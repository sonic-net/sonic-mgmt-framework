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
	"os"
	"reflect"
	"strings"
	"testing"
	db "translib/db"
)

func init() {
	fmt.Println("+++++  Init acl_app_test  +++++")
}

func TestMain(m *testing.M) {
	if err := clearAclDataFromDb(); err != nil {
		os.Exit(-1)
	}
	fmt.Println("+++++  Removed All Acl Data from Db  +++++")

	ret := m.Run()

	if err := clearAclDataFromDb(); err != nil {
		os.Exit(-1)
	}

	os.Exit(ret)
}

// This will test GET on /openconfig-acl:acl
func Test_AclApp_TopLevelPath(t *testing.T) {
	url := "/openconfig-acl:acl"

	t.Run("Empty_Response_Top_Level", processGetRequest(url, emptyJson, false))

	t.Run("Bulk_Create_Top_Level", processSetRequest(url, bulkAclCreateJsonRequest, "POST", false))

	t.Run("Get_Full_Acl_Tree_Top_Level", processGetRequest(url, bulkAclCreateJsonResponse, false))

	// Delete all bindings before deleting at top level
	t.Run("Delete_All_Bindings_Top_Level", processDeleteRequest("/openconfig-acl:acl/interfaces"))
	t.Run("Delete_Full_ACl_Tree_Top_Level", processDeleteRequest(url))

	t.Run("Verify_Top_Level_Delete", processGetRequest(url, emptyJson, false))
}

func Test_AclApp_SingleAclOperations(t *testing.T) {
	url := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL3][type=ACL_IPV4]"

	t.Run("Create_One_Acl_With_Multiple_Rules(PATCH)", processSetRequest(url, oneAclCreateWithRulesJsonRequest, "PATCH", false))

	t.Run("Verify_Create_One_Acl_With_Multiple_Rules", processGetRequest(url, oneAclCreateWithRulesJsonResponse, false))

	aclDescrUrl := url + "/config/description"
	t.Run("Delete Acl_Description", processDeleteRequest(aclDescrUrl))
	t.Run("Verify_Acl_Description_Deletion", processGetRequest(aclDescrUrl, emptyAclDescriptionJson, false))

	createAclDescrUrl := url + "/config"
	t.Run("Create_new_Acl_Description", processSetRequest(createAclDescrUrl, aclDescrUpdateJson, "POST", false))
	t.Run("Verify_Description_of_Acl", processGetRequest(aclDescrUrl, aclDescrUpdateJson, false))

	t.Run("Delete_One_Acl_With_All_Its_Rules", processDeleteRequest(url))

	t.Run("Verify_One_Acl_Delete", processGetRequest(url, "", true))
}

func Test_AclApp_SingleRuleOperations(t *testing.T) {
	aclUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]"
	ruleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=8]"

	t.Run("Create_One_Acl_Without_Rule", processSetRequest(aclUrl, oneAclCreateJsonRequest, "POST", false))
	t.Run("Get_One_Acl_Without_Rule", processGetRequest(aclUrl, oneAclCreateJsonResponse, false))

	t.Run("Create_One_Rule", processSetRequest(ruleUrl, requestOneRulePostJson, "POST", false))
	t.Run("Get_One_Rule", processGetRequest(ruleUrl, responseOneRuleJson, false))

	// Change Source/Desination address and protocol
	t.Run("Update_Existing_Rule", processSetRequest(ruleUrl, requestOneRulePatchJson, "PATCH", false))
	t.Run("Verify_One_Rule_Updation", processGetRequest(ruleUrl, responseOneRulePatchJson, false))

	tcpFlagsUrl := ruleUrl + "/transport/config/tcp-flags"
	t.Run("Delete_Tcp_Flags_Field", processDeleteRequest(tcpFlagsUrl))
	t.Run("Verify_Tcp_Flags_Deletion", processGetRequest(tcpFlagsUrl, emptyJson, false))

	dscpUrl := ruleUrl + "/ipv4/config/dscp"
	t.Run("Delete_IPv4_Dscp_Field", processDeleteRequest(dscpUrl))
	t.Run("Verify_IPv4_Dscp_Deletion", processGetRequest(dscpUrl, emptyRuleDscpJson, false))

	protocolUrl := ruleUrl + "/ipv4/config/protocol"
	t.Run("Delete_IPv4_Protocol_Field", processDeleteRequest(protocolUrl))
	t.Run("Verify_IPv4_Protocol_Deletion", processGetRequest(protocolUrl, emptyJson, false))

	transportConfigUrl := ruleUrl + "/transport"
	t.Run("Delete_Transport_Container", processDeleteRequest(transportConfigUrl))
	t.Run("Verify_Transport_Container_Deletion", processGetRequest(transportConfigUrl, emptyJson, false))

	ipv4ConfigUrl := ruleUrl + "/ipv4/config"
	t.Run("Delete_IPv4_Config_Container", processDeleteRequest(ipv4ConfigUrl))
	t.Run("Verify_IPv4_Config_Container_Deletion", processGetRequest(ipv4ConfigUrl, emptyJson, false))

	t.Run("Delete_One_Rule", processDeleteRequest(ruleUrl))
	t.Run("Verify_One_Rule_Delete", processGetRequest(ruleUrl, "", true))

	t.Run("Delete_One_Acl", processDeleteRequest(aclUrl))
	t.Run("Verify_One_Acl_Delete", processGetRequest(aclUrl, "", true))
}

// This will test PUT (Replace) operation by  Replacing multiple Rules with one Rule in an Acl
func Test_AclApp_ReplaceMultipleRulesWithOneRule(t *testing.T) {
	url := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL3][type=ACL_IPV4]"

	t.Run("Create_One_Acl_With_Multiple_Rules(PATCH)", processSetRequest(url, oneAclCreateWithRulesJsonRequest, "PATCH", false))
	t.Run("Verify_Create_One_Acl_With_Multiple_Rules", processGetRequest(url, oneAclCreateWithRulesJsonResponse, false))

	t.Run("Replace_All_Rules_With_One_Rule", processSetRequest(url, replaceMultiRulesWithOneRuleJsonRequest, "PUT", false))
	t.Run("Verify_Acl_With_Replaced_Rules", processGetRequest(url, replaceMultiRulesWithOneRuleJsonResponse, false))

	t.Run("Delete_One_Acl_With_All_Its_Rules", processDeleteRequest(url))
	t.Run("Verify_One_Acl_Delete", processGetRequest(url, "", true))
}

// This will test PATCH operation by  modifying Description of an Acl
func Test_AclApp_AclDescriptionUpdation(t *testing.T) {
	aclUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]"
	descrUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]/config/description"

	t.Run("Create_One_Acl_Without_Rule", processSetRequest(aclUrl, oneAclCreateJsonRequest, "POST", false))

	t.Run("Update_Description_of_Acl", processSetRequest(descrUrl, aclDescrUpdateJson, "PATCH", false))
	t.Run("Verify_Description_of_Acl", processGetRequest(descrUrl, aclDescrUpdateJson, false))

	t.Run("Delete_One_Acl", processDeleteRequest(aclUrl))
	t.Run("Verify_One_Acl_Delete", processGetRequest(aclUrl, "", true))
}

func Test_AclApp_AclIngressBindingOperations(t *testing.T) {
	aclUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]"
	ruleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=8]"
	bindingUrl := "/openconfig-acl:acl/interfaces/interface[id=Ethernet4]/ingress-acl-sets/ingress-acl-set[set-name=MyACL5][type=ACL_IPV4]"

	t.Run("Create_One_Acl_Without_Rule", processSetRequest(aclUrl, oneAclCreateJsonRequest, "POST", false))

	t.Run("Create_One_Rule", processSetRequest(ruleUrl, requestOneRulePostJson, "POST", false))

	t.Run("Create_Ingress_Acl_set", processSetRequest(bindingUrl, ingressAclSetCreateJsonRequest, "POST", false))
	t.Run("Verify_Ingress_Aclset_Creation", processGetRequest(bindingUrl, ingressAclSetCreateJsonResponse, false))
	t.Run("Get_Port_Binding_From_Ingress_AclEntry_Level", processGetRequest(bindingUrl+"/acl-entries/acl-entry[sequence-id=8]", getBindingAclEntryResponse, false))

	t.Run("Delete_Binding_From_Ingress_Aclset", processDeleteRequest(bindingUrl))
	t.Run("Verify_Binding_From_Ingress_Aclset_Deletion", processGetRequest(bindingUrl, "", true))
	t.Run("Delete_One_Rule", processDeleteRequest(ruleUrl))
	t.Run("Verify_One_Rule_Delete", processGetRequest(ruleUrl, "", true))

	t.Run("Delete_One_Acl", processDeleteRequest(aclUrl))
	t.Run("Verify_One_Acl_Delete", processGetRequest(aclUrl, "", true))
}

func Test_AclApp_AclEgressBindingOperations(t *testing.T) {
	aclUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]"
	ruleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=8]"
	bindingUrl := "/openconfig-acl:acl/interfaces/interface[id=Ethernet4]/egress-acl-sets/egress-acl-set[set-name=MyACL5][type=ACL_IPV4]"

	t.Run("Create_One_Acl_Without_Rule", processSetRequest(aclUrl, oneAclCreateJsonRequest, "POST", false))

	t.Run("Create_One_Rule", processSetRequest(ruleUrl, requestOneRulePostJson, "POST", false))

	t.Run("Create_Egress_Acl_set", processSetRequest(bindingUrl, ingressAclSetCreateJsonRequest, "POST", false))
	t.Run("Verify_Egress_Aclset_Creation", processGetRequest(bindingUrl, egressAclSetCreateJsonResponse, false))
	t.Run("Get_Port_Binding_From_Egress_AclEntry_Level", processGetRequest(bindingUrl+"/acl-entries/acl-entry[sequence-id=8]", getBindingAclEntryResponse, false))

	t.Run("Delete_Binding_From_Egress_Aclset", processDeleteRequest(bindingUrl))
	t.Run("Verify_Binding_From_Egress_Aclset_Deletion", processGetRequest(bindingUrl, "", true))
	t.Run("Delete_One_Rule", processDeleteRequest(ruleUrl))
	t.Run("Verify_One_Rule_Delete", processGetRequest(ruleUrl, "", true))

	t.Run("Delete_One_Acl", processDeleteRequest(aclUrl))
	t.Run("Verify_One_Acl_Delete", processGetRequest(aclUrl, "", true))
}

func Test_AclApp_GetOperationsFromMultipleTreeLevels(t *testing.T) {
	aclUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]"
	ruleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=8]"
	bindingUrl := "/openconfig-acl:acl/interfaces/interface[id=Ethernet4]/egress-acl-sets/egress-acl-set[set-name=MyACL5][type=ACL_IPV4]"

	t.Run("Create_One_Acl_Without_Rule", processSetRequest(aclUrl, oneAclCreateJsonRequest, "POST", false))
	t.Run("Create_One_Rule", processSetRequest(ruleUrl, requestOneRulePostJson, "POST", false))
	t.Run("Create_Egress_Acl_set_Port_Binding", processSetRequest(bindingUrl, ingressAclSetCreateJsonRequest, "POST", false))

	t.Run("Get_Acl_Tree_From_AclSets_level", processGetRequest("/openconfig-acl:acl/acl-sets", getFromAclSetsTreeLevelResponse, false))

	t.Run("Get_All_Ports_Bindings_From_Interfaces_Tree_Level", processGetRequest("/openconfig-acl:acl/interfaces", getAllPortsFromInterfacesTreeLevelResponse, false))

	t.Run("Get_One_Port_Binding_From_Interface_Tree_Level", processGetRequest("/openconfig-acl:acl/interfaces/interface[id=Ethernet4]", getPortBindingFromInterfaceTreeLevelResponse, false))

	t.Run("Delete_Binding_From_Egress_Aclset", processDeleteRequest(bindingUrl))
	t.Run("Verify_Binding_From_Egress_Aclset_Deletion", processGetRequest(bindingUrl, "", true))

	t.Run("Delete_One_Rule", processDeleteRequest(ruleUrl))
	t.Run("Verify_One_Rule_Delete", processGetRequest(ruleUrl, "", true))

	t.Run("Delete_One_Acl", processDeleteRequest(aclUrl))
	t.Run("Verify_One_Acl_Delete", processGetRequest(aclUrl, "", true))
}

func Test_AclApp_AddNewPortBindingToAlreadyBindedAcl(t *testing.T) {
	aclUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]"
	ruleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=8]"
	bindingUrl := "/openconfig-acl:acl/interfaces/interface[id=Ethernet4]/egress-acl-sets/egress-acl-set[set-name=MyACL5][type=ACL_IPV4]"

	t.Run("Create_One_Acl_Without_Rule", processSetRequest(aclUrl, oneAclCreateJsonRequest, "POST", false))
	t.Run("Create_One_Rule", processSetRequest(ruleUrl, requestOneRulePostJson, "POST", false))
	t.Run("Create_Egress_Acl_set_Port_Binding", processSetRequest(bindingUrl, ingressAclSetCreateJsonRequest, "POST", false))

	newBindingUrl := "/openconfig-acl:acl/interfaces/interface[id=Ethernet0]/egress-acl-sets/egress-acl-set[set-name=MyACL5][type=ACL_IPV4]"
	t.Run("Create_New_Egress_Acl_set_Port_Binding", processSetRequest(newBindingUrl, ingressAclSetCreateJsonRequest, "POST", false))

	t.Run("Get_All_Ports_Bindings_From_Interfaces_Tree_Level", processGetRequest("/openconfig-acl:acl/interfaces", getMultiportBindingOnSingleAclResponse, false))

	t.Run("Delete_All_Bindings_Top_Level", processDeleteRequest("/openconfig-acl:acl/interfaces"))
	t.Run("Delete_All_Rules_Not_Acl", processDeleteRequest("/openconfig-acl:acl/acl-sets/acl-set[name=MyACL5][type=ACL_IPV4]/acl-entries"))

	t.Run("Delete_One_Acl", processDeleteRequest(aclUrl))
	t.Run("Verify_One_Acl_Delete", processGetRequest(aclUrl, "", true))

	t.Run("Verify_Top_Level_Delete", processGetRequest("/openconfig-acl:acl", emptyJson, false))
}

func Test_AclApp_IPv6AclAndRule(t *testing.T) {
	aclUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL6][type=ACL_IPV6]"
	ruleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL6][type=ACL_IPV6]/acl-entries/acl-entry[sequence-id=6]"
	bindingUrl := "/openconfig-acl:acl/interfaces/interface[id=Ethernet4]/ingress-acl-sets/ingress-acl-set[set-name=MyACL6][type=ACL_IPV6]"

	t.Run("Create_One_IPv6_Acl_Without_Rule", processSetRequest(aclUrl, oneIPv6AclCreateJsonRequest, "POST", false))
	t.Run("Verify_One_IPv6_Acl_Without_Rule_Creation", processGetRequest(aclUrl, oneIPv6AclCreateJsonResponse, false))

	t.Run("Create_One_IPv6_Rule", processSetRequest(ruleUrl, oneIPv6RuleCreateJsonRequest, "POST", false))
	t.Run("Verify_One_IPv6_Rule_Creation", processGetRequest(ruleUrl, oneIPv6RuleCreateJsonResponse, false))

	t.Run("Create_Ingress_Acl_set", processSetRequest(bindingUrl, ingressIPv6AclSetCreateJsonRequest, "POST", false))
	t.Run("Verify_Ingress_Aclset_Creation", processGetRequest(bindingUrl, ingressIPv6AclSetCreateJsonResponse, false))

	t.Run("Get_Acl_Tree_From_AclSet_level", processGetRequest("/openconfig-acl:acl/acl-sets/acl-set", getIPv6AclsFromAclSetListLevelResponse, false))
	t.Run("Get_All_Ports_Bindings_From_Interfaces_Tree_Level", processGetRequest("/openconfig-acl:acl/interfaces", getIPv6AllPortsBindingsResponse, false))

	t.Run("Delete_Binding_From_Ingress_Aclset", processDeleteRequest(bindingUrl))
	t.Run("Verify_Binding_From_Ingress_Aclset_Deletion", processGetRequest(bindingUrl, "", true))

	pktActionUrl := ruleUrl + "/actions/config/forwarding-action"
	t.Run("Delete_Packet_Action_Field", processDeleteRequest(pktActionUrl))
	t.Run("Verify_Packet_Action_Field_Deletion", processGetRequest(pktActionUrl, emptyJson, false))

	ipv6ConfigUrl := ruleUrl + "/ipv6/config"
	t.Run("Delete_IPv6_Config", processDeleteRequest(ipv6ConfigUrl))
	t.Run("Verify_IPv6_Config_Deletion", processGetRequest(ipv6ConfigUrl, emptyJson, false))

	t.Run("Delete_One_Rule", processDeleteRequest(ruleUrl))
	t.Run("Delete_One_Acl", processDeleteRequest(aclUrl))
	t.Run("Verify_One_Acl_Delete", processGetRequest(aclUrl, "", true))
}

func Test_AclApp_L2AclAndRule(t *testing.T) {
	aclUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL2][type=ACL_L2]"
	ruleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL2][type=ACL_L2]/acl-entries/acl-entry[sequence-id=2]"
	bindingUrl := "/openconfig-acl:acl/interfaces/interface[id=Ethernet0]/ingress-acl-sets/ingress-acl-set[set-name=MyACL2][type=ACL_L2]"

	t.Run("Create_One_L2_Acl_Without_Rule", processSetRequest(aclUrl, oneL2AclCreateJsonRequest, "POST", false))
	t.Run("Verify_One_L2_Acl_Without_Rule_Creation", processGetRequest(aclUrl, oneL2AclCreateJsonResponse, false))

	t.Run("Create_One_L2_Rule", processSetRequest(ruleUrl, oneL2RuleCreateJsonRequest, "POST", false))
	t.Run("Verify_One_L2_Rule_Creation", processGetRequest(ruleUrl, oneL2RuleCreateJsonResponse, false))

	t.Run("Create_Ingress_L2_Acl_set", processSetRequest(bindingUrl, ingressL2AclSetCreateJsonRequest, "POST", false))
	t.Run("Verify_Ingress_L2_Aclset_Creation", processGetRequest(bindingUrl, ingressL2AclSetCreateJsonResponse, false))

	t.Run("Get_Acl_Tree_From_AclSet_level", processGetRequest("/openconfig-acl:acl/acl-sets/acl-set", getL2AclsFromAclSetListLevelResponse, false))
	t.Run("Get_All_Ports_Bindings_From_Interfaces_Tree_Level", processGetRequest("/openconfig-acl:acl/interfaces", getL2AllPortsBindingsResponse, false))

	t.Run("Delete_Binding_From_Ingress_Aclset", processDeleteRequest(bindingUrl))
	t.Run("Verify_Binding_From_Ingress_Aclset_Deletion", processGetRequest(bindingUrl, "", true))

	etherTypeUrl := ruleUrl + "/l2/config/ethertype"
	t.Run("Delete_Ethertype_Field", processDeleteRequest(etherTypeUrl))
	t.Run("Verify_L2_Ethertype_Field_Deletion", processGetRequest(ruleUrl+"/l2/config", emptyJson, false))

	t.Run("Delete_Transport_Src_Port_Field", processDeleteRequest(ruleUrl+"/transport/config/source-port"))
	t.Run("Delete_Transport_Dst_Port_Field", processDeleteRequest(ruleUrl+"/transport/config/destination-port"))
	t.Run("Delete_Transport_Tcp_Flags_Field", processDeleteRequest(ruleUrl+"/transport/config/tcp-flags"))
	t.Run("Verify_Transport_Src_Dst_Fields_Deletion", processGetRequest(ruleUrl+"/transport/config", emptyJson, false))

	t.Run("Delete_One_Rule", processDeleteRequest(ruleUrl))
	t.Run("Delete_One_Acl", processDeleteRequest(aclUrl))
	t.Run("Verify_One_Acl_Delete", processGetRequest(aclUrl, "", true))
}

func Test_AclApp_NegativeTests(t *testing.T) {
	// Verify GET returns errors for non-existing ACL
	aclUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL2][type=ACL_L2]"
	t.Run("Verify_Non_Existing_Acl_GET_Error", processGetRequest(aclUrl, "", true))

	// Verify GET returns errors for non-existing Rule
	ruleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL2][type=ACL_L2]/acl-entries/acl-entry[sequence-id=2]"
	t.Run("Verify_Non_Existing_Rule_GET_Error", processGetRequest(ruleUrl, "", true))

	// Verify Error on giving Invalid Interface in payload during binding creation
	url := "/openconfig-acl:acl"
	t.Run("Create_Acl_With_Invalid_Interface_Binding", processSetRequest(url, aclCreateWithInvalidInterfaceBinding, "POST", true))

	// Verify error if duplicate Acl is created using POST
	t.Run("Create_One_L2_Acl_Without_Rule", processSetRequest(aclUrl, oneL2AclCreateJsonRequest, "POST", false))
	t.Run("Verify_One_L2_Acl_Without_Rule_Creation", processGetRequest(aclUrl, oneL2AclCreateJsonResponse, false))
	t.Run("Verify_Error_On_Create_Duplicate_L2_Acl", processSetRequest(aclUrl, oneL2AclCreateJsonRequest, "POST", true))

	// Verify error if duplicate Rule is created using POST
	multiRuleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL3][type=ACL_IPV4]"
	t.Run("Create_One_Acl_With_Multiple_Rules(PATCH)", processSetRequest(multiRuleUrl, oneAclCreateWithRulesJsonRequest, "PATCH", false))
	t.Run("Verify_Create_One_Acl_With_Multiple_Rules", processGetRequest(multiRuleUrl, oneAclCreateWithRulesJsonResponse, true))

	duplicateRuleUrl := "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL3][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=1]"
	t.Run("Create_One_Duplicate_Rule", processSetRequest(duplicateRuleUrl, requestOneDuplicateRulePostJson, "POST", true))

	topLevelUrl := "/openconfig-acl:acl"
	t.Run("Delete_Full_ACl_Tree_Top_Level", processDeleteRequest(topLevelUrl))
	t.Run("Verify_Top_Level_Delete", processGetRequest(topLevelUrl, emptyJson, false))
}

func processGetRequest(url string, expectedRespJson string, errorCase bool) func(*testing.T) {
	return func(t *testing.T) {
		response, err := Get(GetRequest{url})
		if err != nil && !errorCase {
			t.Errorf("Error %v received for Url: %s", err, url)
		}

		respJson := response.Payload
		if string(respJson) != expectedRespJson {
			t.Errorf("Response for Url: %s received is not expected:\n%s", url, string(respJson))
		}
	}
}

func processSetRequest(url string, jsonPayload string, oper string, errorCase bool) func(*testing.T) {
	return func(t *testing.T) {
		var err error
		switch oper {
		case "POST":
			_, err = Create(SetRequest{Path: url, Payload: []byte(jsonPayload)})
		case "PATCH":
			_, err = Update(SetRequest{Path: url, Payload: []byte(jsonPayload)})
		case "PUT":
			_, err = Replace(SetRequest{Path: url, Payload: []byte(jsonPayload)})
		default:
			t.Errorf("Operation not supported")
		}
		if err != nil && !errorCase {
			t.Errorf("Error %v received for Url: %s", err, url)
		}
	}
}

func processDeleteRequest(url string) func(*testing.T) {
	return func(t *testing.T) {
		_, err := Delete(SetRequest{Path: url})
		if err != nil {
			t.Errorf("Error %v received for Url: %s", err, url)
		}
	}
}

// THis will delete ACL table and Rules Table from DB
func clearAclDataFromDb() error {
	var err error
	ruleTable := db.TableSpec{Name: "ACL_RULE"}
	aclTable := db.TableSpec{Name: "ACL_TABLE"}

	d := getConfigDb()
	if d == nil {
		err = errors.New("Failed to connect to config Db")
		return err
	}
	if err = d.DeleteTable(&ruleTable); err != nil {
		err = errors.New("Failed to delete Rules Table")
		return err
	}
	if err = d.DeleteTable(&aclTable); err != nil {
		err = errors.New("Failed to delete Acl Table")
		return err
	}
	return err
}

func getConfigDb() *db.DB {
	configDb, _ := db.NewDB(db.Options{
		DBNo:               db.ConfigDB,
		InitIndicator:      "CONFIG_DB_INITIALIZED",
		TableNameSeparator: "|",
		KeySeparator:       "|",
	})

	return configDb
}

func Test_AclApp_Subscribe(t *testing.T) {
	app := new(AclApp)

	t.Run("top", testSubsError(app, "/"))
	t.Run("unknown", testSubsError(app, "/some/unknown/path"))
	t.Run("topacl", testSubsError(app, "/openconfig-acl:acl"))
	t.Run("aclsets", testSubsError(app, "/openconfig-acl:acl/acl-sets"))
	t.Run("aclset*", testSubsError(app, "/openconfig-acl:acl/acl-sets/acl-set"))
	t.Run("aclset", testSubsError(app, "/openconfig-acl:acl/acl-sets/acl-set[name=X][type=ACL_IPV4]"))

	t.Run("acl_config", testSubs(app,
		"/openconfig-acl:acl/acl-sets/acl-set[name=X][type=ACL_IPV4]/config/description",
		"ACL_TABLE", "X_ACL_IPV4", true))

	t.Run("acl_state", testSubs(app,
		"/openconfig-acl:acl/acl-sets/acl-set[name=X][type=ACL_IPV4]/state",
		"ACL_TABLE", "X_ACL_IPV4", true))

	t.Run("entries", testSubs(app,
		"/openconfig-acl:acl/acl-sets/acl-set[name=X][type=ACL_IPV4]/acl-entries",
		"ACL_RULE", "X_ACL_IPV4|*", false))

	t.Run("rule*", testSubs(app,
		"/openconfig-acl:acl/acl-sets/acl-set[name=X][type=ACL_IPV4]/acl-entries/acl-entry",
		"ACL_RULE", "X_ACL_IPV4|*", false))

	t.Run("rule", testSubs(app,
		"/openconfig-acl:acl/acl-sets/acl-set[name=X][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=1]",
		"ACL_RULE", "X_ACL_IPV4|RULE_1", false))

	t.Run("rule_state", testSubs(app,
		"/openconfig-acl:acl/acl-sets/acl-set[name=X][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=100]/state",
		"ACL_RULE", "X_ACL_IPV4|RULE_100", true))

	t.Run("rule_sip", testSubs(app,
		"/openconfig-acl:acl/acl-sets/acl-set[name=X][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=200]/ipv4/config/source-address",
		"ACL_RULE", "X_ACL_IPV4|RULE_200", true))

}

// testSubs creates a test case which invokes translateSubscribe on an
// app interafce and check returned notificationInfo matches given values.
func testSubs(app appInterface, path, oTable, oKey string, oCache bool) func(*testing.T) {
	return func(t *testing.T) {
		_, nt, err := app.translateSubscribe([db.MaxDB]*db.DB{}, path)
		if err != nil {
			t.Fatalf("Unexpected error processing '%s'; err=%v", path, err)
		}
		if nt == nil || nt.needCache != oCache || nt.table.Name != oTable ||
			!reflect.DeepEqual(nt.key.Comp, strings.Split(oKey, "|")) {
			t.Logf("translateSubscribe for path '%s'", path)
			t.Logf("Expected table '%s', key '%v', cache %v", oTable, oKey, oCache)
			if nt == nil {
				t.Fatalf("Found nil")
			} else {
				t.Fatalf("Found table '%s', key '%s', cache %v",
					nt.table.Name, strings.Join(nt.key.Comp, "|"), nt.needCache)
			}
		}
	}
}

// testSubsError creates a test case which invokes translateSubscribe on
// an app interafce and expects it to return an error
func testSubsError(app appInterface, path string) func(*testing.T) {
	return func(t *testing.T) {
		_, _, err := app.translateSubscribe([db.MaxDB]*db.DB{}, path)
		if err == nil {
			t.Fatalf("Expected error for path '%s'", path)
		}
	}
}

/***************************************************************************/
///////////                  JSON Data for Tests              ///////////////
/***************************************************************************/
var emptyJson string = "{}"

var bulkAclCreateJsonRequest string = "{\"acl-sets\":{\"acl-set\":[{\"name\":\"MyACL1\",\"type\":\"ACL_IPV4\",\"config\":{\"name\":\"MyACL1\",\"type\":\"ACL_IPV4\",\"description\":\"Description for MyACL1\"},\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":1,\"config\":{\"sequence-id\":1,\"description\":\"Description for MyACL1 Rule Seq 1\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.1/32\",\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":101,\"destination-port\":201}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":2,\"config\":{\"sequence-id\":2,\"description\":\"Description for MyACL1 Rule Seq 2\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.2/32\",\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":102,\"destination-port\":202}},\"actions\":{\"config\":{\"forwarding-action\":\"DROP\"}}},{\"sequence-id\":3,\"config\":{\"sequence-id\":3,\"description\":\"Description for MyACL1 Rule Seq 3\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.3/32\",\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":103,\"destination-port\":203}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":4,\"config\":{\"sequence-id\":4,\"description\":\"Description for MyACL1 Rule Seq 4\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.4/32\",\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":104,\"destination-port\":204}},\"actions\":{\"config\":{\"forwarding-action\":\"DROP\"}}},{\"sequence-id\":5,\"config\":{\"sequence-id\":5,\"description\":\"Description for MyACL1 Rule Seq 5\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.5/32\",\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":105,\"destination-port\":205}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}}]}},{\"name\":\"MyACL2\",\"type\":\"ACL_IPV4\",\"config\":{\"name\":\"MyACL2\",\"type\":\"ACL_IPV4\",\"description\":\"Description for MyACL2\"},\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":1,\"config\":{\"sequence-id\":1,\"description\":\"Description for Rule Seq 1\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.1/32\",\"destination-address\":\"22.1.1.1/32\",\"dscp\":1,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":101,\"destination-port\":201}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":2,\"config\":{\"sequence-id\":2,\"description\":\"Description for Rule Seq 2\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.2/32\",\"destination-address\":\"22.1.1.2/32\",\"dscp\":2,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":102,\"destination-port\":202}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":3,\"config\":{\"sequence-id\":3,\"description\":\"Description for Rule Seq 3\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.3/32\",\"destination-address\":\"22.1.1.3/32\",\"dscp\":3,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":103,\"destination-port\":203}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":4,\"config\":{\"sequence-id\":4,\"description\":\"Description for Rule Seq 4\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.4/32\",\"destination-address\":\"22.1.1.4/32\",\"dscp\":4,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":104,\"destination-port\":204}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":5,\"config\":{\"sequence-id\":5,\"description\":\"Description for Rule Seq 5\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.5/32\",\"destination-address\":\"22.1.1.5/32\",\"dscp\":5,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":105,\"destination-port\":205}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}}]}}]},\"interfaces\":{\"interface\":[{\"id\":\"Ethernet0\",\"config\":{\"id\":\"Ethernet0\"},\"interface-ref\":{\"config\":{\"interface\":\"Ethernet0\"}},\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"set-name\":\"MyACL1\",\"type\":\"ACL_IPV4\",\"config\":{\"set-name\":\"MyACL1\",\"type\":\"ACL_IPV4\"}}]}},{\"id\":\"Ethernet4\",\"config\":{\"id\":\"Ethernet4\"},\"interface-ref\":{\"config\":{\"interface\":\"Ethernet4\"}},\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"set-name\":\"MyACL2\",\"type\":\"ACL_IPV4\",\"config\":{\"set-name\":\"MyACL2\",\"type\":\"ACL_IPV4\"}}]}}]}}"

var bulkAclCreateJsonResponse string = "{\"openconfig-acl:acl\":{\"acl-sets\":{\"acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":1},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.1/32\"},\"state\":{\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.1/32\"}},\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1},\"transport\":{\"config\":{\"destination-port\":201,\"source-port\":101},\"state\":{\"destination-port\":201,\"source-port\":101}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:DROP\"},\"state\":{\"forwarding-action\":\"openconfig-acl:DROP\"}},\"config\":{\"sequence-id\":2},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.2/32\"},\"state\":{\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.2/32\"}},\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2},\"transport\":{\"config\":{\"destination-port\":202,\"source-port\":102},\"state\":{\"destination-port\":202,\"source-port\":102}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":3},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.3/32\"},\"state\":{\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.3/32\"}},\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3},\"transport\":{\"config\":{\"destination-port\":203,\"source-port\":103},\"state\":{\"destination-port\":203,\"source-port\":103}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:DROP\"},\"state\":{\"forwarding-action\":\"openconfig-acl:DROP\"}},\"config\":{\"sequence-id\":4},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.4/32\"},\"state\":{\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.4/32\"}},\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4},\"transport\":{\"config\":{\"destination-port\":204,\"source-port\":104},\"state\":{\"destination-port\":204,\"source-port\":104}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":5},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.5/32\"},\"state\":{\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.5/32\"}},\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5},\"transport\":{\"config\":{\"destination-port\":205,\"source-port\":105},\"state\":{\"destination-port\":205,\"source-port\":105}}}]},\"config\":{\"description\":\"Description for MyACL1\",\"name\":\"MyACL1\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL1\",\"state\":{\"description\":\"Description for MyACL1\",\"name\":\"MyACL1\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"},{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":1},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.1/32\"},\"state\":{\"destination-address\":\"22.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.1/32\"}},\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1},\"transport\":{\"config\":{\"destination-port\":201,\"source-port\":101},\"state\":{\"destination-port\":201,\"source-port\":101}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":2},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.2/32\"},\"state\":{\"destination-address\":\"22.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.2/32\"}},\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2},\"transport\":{\"config\":{\"destination-port\":202,\"source-port\":102},\"state\":{\"destination-port\":202,\"source-port\":102}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":3},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.3/32\"},\"state\":{\"destination-address\":\"22.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.3/32\"}},\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3},\"transport\":{\"config\":{\"destination-port\":203,\"source-port\":103},\"state\":{\"destination-port\":203,\"source-port\":103}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":4},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.4/32\"},\"state\":{\"destination-address\":\"22.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.4/32\"}},\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4},\"transport\":{\"config\":{\"destination-port\":204,\"source-port\":104},\"state\":{\"destination-port\":204,\"source-port\":104}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":5},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.5/32\"},\"state\":{\"destination-address\":\"22.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.5/32\"}},\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5},\"transport\":{\"config\":{\"destination-port\":205,\"source-port\":105},\"state\":{\"destination-port\":205,\"source-port\":105}}}]},\"config\":{\"description\":\"Description for MyACL2\",\"name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL2\",\"state\":{\"description\":\"Description for MyACL2\",\"name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"interfaces\":{\"interface\":[{\"config\":{\"id\":\"Ethernet0\"},\"id\":\"Ethernet0\",\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1}},{\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2}},{\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3}},{\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4}},{\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5}}]},\"config\":{\"set-name\":\"MyACL1\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL1\",\"state\":{\"set-name\":\"MyACL1\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"state\":{\"id\":\"Ethernet0\"}},{\"config\":{\"id\":\"Ethernet4\"},\"id\":\"Ethernet4\",\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1}},{\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2}},{\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3}},{\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4}},{\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5}}]},\"config\":{\"set-name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL2\",\"state\":{\"set-name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"state\":{\"id\":\"Ethernet4\"}}]}}}"

var oneAclCreateWithRulesJsonRequest string = "{ \"name\": \"MyACL3\", \"type\": \"ACL_IPV4\", \"config\": { \"name\": \"MyACL3\", \"type\": \"ACL_IPV4\", \"description\": \"Description for MyACL3\" }, \"acl-entries\": { \"acl-entry\": [ { \"sequence-id\": 1, \"config\": { \"sequence-id\": 1, \"description\": \"Description for MyACL3 Rule Seq 1\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.1/32\", \"destination-address\": \"21.1.1.1/32\", \"dscp\": 1, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 101, \"destination-port\": 201 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } }, { \"sequence-id\": 2, \"config\": { \"sequence-id\": 2, \"description\": \"Description for MyACL3 Rule Seq 2\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.2/32\", \"destination-address\": \"21.1.1.2/32\", \"dscp\": 2, \"protocol\": \"IP_UDP\" } }, \"transport\": { \"config\": { \"source-port\": 102, \"destination-port\": 202 } }, \"actions\": { \"config\": { \"forwarding-action\": \"DROP\" } } }, { \"sequence-id\": 3, \"config\": { \"sequence-id\": 3, \"description\": \"Description for MyACL3 Rule Seq 3\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.3/32\", \"destination-address\": \"21.1.1.3/32\", \"dscp\": 3, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 103, \"destination-port\": 203 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } }, { \"sequence-id\": 4, \"config\": { \"sequence-id\": 4, \"description\": \"Description for MyACL3 Rule Seq 4\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.4/32\", \"destination-address\": \"21.1.1.4/32\", \"dscp\": 4, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 104, \"destination-port\": 204 } }, \"actions\": { \"config\": { \"forwarding-action\": \"DROP\" } } }, { \"sequence-id\": 5, \"config\": { \"sequence-id\": 5, \"description\": \"Description for MyACL3 Rule Seq 5\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.5/32\", \"destination-address\": \"21.1.1.5/32\", \"dscp\": 5, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 105, \"destination-port\": 205 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } } ] }}"

var oneAclCreateWithRulesJsonResponse string = "{\"openconfig-acl:acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":1},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.1/32\"},\"state\":{\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.1/32\"}},\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1},\"transport\":{\"config\":{\"destination-port\":201,\"source-port\":101},\"state\":{\"destination-port\":201,\"source-port\":101}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:DROP\"},\"state\":{\"forwarding-action\":\"openconfig-acl:DROP\"}},\"config\":{\"sequence-id\":2},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_UDP\",\"source-address\":\"11.1.1.2/32\"},\"state\":{\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_UDP\",\"source-address\":\"11.1.1.2/32\"}},\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2},\"transport\":{\"config\":{\"destination-port\":202,\"source-port\":102},\"state\":{\"destination-port\":202,\"source-port\":102}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":3},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.3/32\"},\"state\":{\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.3/32\"}},\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3},\"transport\":{\"config\":{\"destination-port\":203,\"source-port\":103},\"state\":{\"destination-port\":203,\"source-port\":103}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:DROP\"},\"state\":{\"forwarding-action\":\"openconfig-acl:DROP\"}},\"config\":{\"sequence-id\":4},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.4/32\"},\"state\":{\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.4/32\"}},\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4},\"transport\":{\"config\":{\"destination-port\":204,\"source-port\":104},\"state\":{\"destination-port\":204,\"source-port\":104}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":5},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.5/32\"},\"state\":{\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.5/32\"}},\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5},\"transport\":{\"config\":{\"destination-port\":205,\"source-port\":105},\"state\":{\"destination-port\":205,\"source-port\":105}}}]},\"config\":{\"description\":\"Description for MyACL3\",\"name\":\"MyACL3\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL3\",\"state\":{\"description\":\"Description for MyACL3\",\"name\":\"MyACL3\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]}"

var oneAclCreateJsonRequest string = "{\"config\": {\"name\": \"MyACL5\",\"type\": \"ACL_IPV4\",\"description\": \"Description for MyACL5\"}}"
var oneAclCreateJsonResponse string = "{\"openconfig-acl:acl-set\":[{\"config\":{\"description\":\"Description for MyACL5\",\"name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL5\",\"state\":{\"description\":\"Description for MyACL5\",\"name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]}"

var aclDescrUpdateJson string = "{\"openconfig-acl:description\":\"Verifying ACL Description Update\"}"

var requestOneRulePostJson string = "{\"sequence-id\": 8,\"config\": {\"sequence-id\": 8,\"description\": \"Description for MyACL5 Rule Seq 8\"},\"ipv4\": {\"config\": {\"source-address\": \"4.4.4.4/24\",\"destination-address\": \"5.5.5.5/24\",\"protocol\": \"IP_TCP\"}},\"transport\": {\"config\": {\"source-port\": 101,\"destination-port\": 100,\"tcp-flags\": [\"TCP_FIN\",\"TCP_ACK\"]}},\"actions\": {\"config\": {\"forwarding-action\": \"ACCEPT\"}}}"

var requestOneRulePatchJson string = "{\"sequence-id\": 8,\"config\": {\"sequence-id\": 8,\"description\": \"Description for MyACL5 Rule Seq 8\"},\"ipv4\": {\"config\": {\"source-address\": \"4.8.4.8/24\",\"destination-address\": \"15.5.15.5/24\",\"protocol\": \"IP_L2TP\"}},\"transport\": {\"config\": {\"source-port\": 101,\"destination-port\": 100,\"tcp-flags\": [\"TCP_FIN\",\"TCP_ACK\",\"TCP_RST\",\"TCP_ECE\"]}},\"actions\": {\"config\": {\"forwarding-action\": \"ACCEPT\"}}}"

var responseOneRuleJson string = "{\"openconfig-acl:acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":8},\"ipv4\":{\"config\":{\"destination-address\":\"5.5.5.5/24\",\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"4.4.4.4/24\"},\"state\":{\"destination-address\":\"5.5.5.5/24\",\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"4.4.4.4/24\"}},\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8},\"transport\":{\"config\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]},\"state\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]}}}]}"

var responseOneRulePatchJson string = "{\"openconfig-acl:acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":8},\"ipv4\":{\"config\":{\"destination-address\":\"15.5.15.5/24\",\"protocol\":\"openconfig-packet-match-types:IP_L2TP\",\"source-address\":\"4.8.4.8/24\"},\"state\":{\"destination-address\":\"15.5.15.5/24\",\"protocol\":\"openconfig-packet-match-types:IP_L2TP\",\"source-address\":\"4.8.4.8/24\"}},\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8},\"transport\":{\"config\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_RST\",\"openconfig-packet-match-types:TCP_ACK\",\"openconfig-packet-match-types:TCP_ECE\"]},\"state\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_RST\",\"openconfig-packet-match-types:TCP_ACK\",\"openconfig-packet-match-types:TCP_ECE\"]}}}]}"

var emptyAclDescriptionJson string = "{\"openconfig-acl:description\":\"\"}"
var emptyRuleDscpJson string = "{\"openconfig-acl:dscp\":0}"

var ingressAclSetCreateJsonRequest string = "{ \"openconfig-acl:config\": { \"set-name\": \"MyACL5\", \"type\": \"ACL_IPV4\" }}"
var ingressAclSetCreateJsonResponse string = "{\"openconfig-acl:ingress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8}}]},\"config\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL5\",\"state\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]}"

var egressAclSetCreateJsonResponse string = "{\"openconfig-acl:egress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8}}]},\"config\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL5\",\"state\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]}"

var replaceMultiRulesWithOneRuleJsonRequest string = "{\"name\": \"MyACL3\",\"type\": \"ACL_IPV4\",\"config\": {\"name\": \"MyACL3\",\"type\": \"ACL_IPV4\",\"description\": \"Description for MyACL3\"},\"acl-entries\": {\"acl-entry\": [{\"sequence-id\": 8,\"config\": {\"sequence-id\": 8,\"description\": \"Description for MyACL3 Rule Seq 8\"},\"ipv4\": {\"config\": {\"source-address\": \"81.1.1.1/32\",\"destination-address\": \"91.1.1.1/32\",\"protocol\": \"IP_TCP\"}},\"transport\": {\"config\": {\"source-port\": \"801..811\",\"destination-port\": \"901..921\"}},\"actions\": {\"config\": {\"forwarding-action\": \"REJECT\"}}}]}}"

var replaceMultiRulesWithOneRuleJsonResponse string = "{\"openconfig-acl:acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:DROP\"},\"state\":{\"forwarding-action\":\"openconfig-acl:DROP\"}},\"config\":{\"sequence-id\":8},\"ipv4\":{\"config\":{\"destination-address\":\"91.1.1.1/32\",\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"81.1.1.1/32\"},\"state\":{\"destination-address\":\"91.1.1.1/32\",\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"81.1.1.1/32\"}},\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8},\"transport\":{\"config\":{\"destination-port\":\"901-921\",\"source-port\":\"801-811\"},\"state\":{\"destination-port\":\"901-921\",\"source-port\":\"801-811\"}}}]},\"config\":{\"description\":\"Description for MyACL3\",\"name\":\"MyACL3\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL3\",\"state\":{\"description\":\"Description for MyACL3\",\"name\":\"MyACL3\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]}"

var getFromAclSetsTreeLevelResponse string = "{\"openconfig-acl:acl-sets\":{\"acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":8},\"ipv4\":{\"config\":{\"destination-address\":\"5.5.5.5/24\",\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"4.4.4.4/24\"},\"state\":{\"destination-address\":\"5.5.5.5/24\",\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"4.4.4.4/24\"}},\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8},\"transport\":{\"config\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]},\"state\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]}}}]},\"config\":{\"description\":\"Description for MyACL5\",\"name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL5\",\"state\":{\"description\":\"Description for MyACL5\",\"name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]}}"

var getAllPortsFromInterfacesTreeLevelResponse string = "{\"openconfig-acl:interfaces\":{\"interface\":[{\"config\":{\"id\":\"Ethernet4\"},\"egress-acl-sets\":{\"egress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8}}]},\"config\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL5\",\"state\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"id\":\"Ethernet4\",\"state\":{\"id\":\"Ethernet4\"}}]}}"

var getPortBindingFromInterfaceTreeLevelResponse string = "{\"openconfig-acl:interface\":[{\"config\":{\"id\":\"Ethernet4\"},\"egress-acl-sets\":{\"egress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8}}]},\"config\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL5\",\"state\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"id\":\"Ethernet4\",\"state\":{\"id\":\"Ethernet4\"}}]}"

var getBindingAclEntryResponse string = "{\"openconfig-acl:acl-entry\":[{\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8}}]}"

var getMultiportBindingOnSingleAclResponse string = "{\"openconfig-acl:interfaces\":{\"interface\":[{\"config\":{\"id\":\"Ethernet0\"},\"egress-acl-sets\":{\"egress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8}}]},\"config\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL5\",\"state\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"id\":\"Ethernet0\",\"state\":{\"id\":\"Ethernet0\"}},{\"config\":{\"id\":\"Ethernet4\"},\"egress-acl-sets\":{\"egress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8}}]},\"config\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL5\",\"state\":{\"set-name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"id\":\"Ethernet4\",\"state\":{\"id\":\"Ethernet4\"}}]}}"

var oneIPv6AclCreateJsonRequest string = "{\"config\": {\"name\": \"MyACL6\",\"type\": \"ACL_IPV6\",\"description\": \"Description for IPv6 ACL MyACL6\"}}"
var oneIPv6AclCreateJsonResponse string = "{\"openconfig-acl:acl-set\":[{\"config\":{\"description\":\"Description for IPv6 ACL MyACL6\",\"name\":\"MyACL6\",\"type\":\"openconfig-acl:ACL_IPV6\"},\"name\":\"MyACL6\",\"state\":{\"description\":\"Description for IPv6 ACL MyACL6\",\"name\":\"MyACL6\",\"type\":\"openconfig-acl:ACL_IPV6\"},\"type\":\"openconfig-acl:ACL_IPV6\"}]}"

var oneIPv6RuleCreateJsonRequest string = "{\"sequence-id\": 6,\"config\": {\"sequence-id\": 6,\"description\": \"Description for MyACL6 Rule Seq 6\"},\"ipv6\": {\"config\": {\"source-address\": \"11::67/64\",\"destination-address\": \"22::87/64\",\"protocol\": \"IP_TCP\",\"dscp\": 11}},\"transport\": {\"config\": {\"source-port\": 101,\"destination-port\": 100,\"tcp-flags\": [\"TCP_FIN\",\"TCP_ACK\"]}},\"actions\": {\"config\": {\"forwarding-action\": \"ACCEPT\"}}}"
var oneIPv6RuleCreateJsonResponse string = "{\"openconfig-acl:acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":6},\"ipv6\":{\"config\":{\"destination-address\":\"22::87/64\",\"dscp\":11,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11::67/64\"},\"state\":{\"destination-address\":\"22::87/64\",\"dscp\":11,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11::67/64\"}},\"sequence-id\":6,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":6},\"transport\":{\"config\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]},\"state\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]}}}]}"

var ingressIPv6AclSetCreateJsonRequest string = "{ \"openconfig-acl:config\": { \"set-name\": \"MyACL6\", \"type\": \"ACL_IPV6\" }}"
var ingressIPv6AclSetCreateJsonResponse string = "{\"openconfig-acl:ingress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":6,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":6}}]},\"config\":{\"set-name\":\"MyACL6\",\"type\":\"openconfig-acl:ACL_IPV6\"},\"set-name\":\"MyACL6\",\"state\":{\"set-name\":\"MyACL6\",\"type\":\"openconfig-acl:ACL_IPV6\"},\"type\":\"openconfig-acl:ACL_IPV6\"}]}"

var getIPv6AclsFromAclSetListLevelResponse string = "{\"openconfig-acl:acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":6},\"ipv6\":{\"config\":{\"destination-address\":\"22::87/64\",\"dscp\":11,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11::67/64\"},\"state\":{\"destination-address\":\"22::87/64\",\"dscp\":11,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11::67/64\"}},\"sequence-id\":6,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":6},\"transport\":{\"config\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]},\"state\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]}}}]},\"config\":{\"description\":\"Description for IPv6 ACL MyACL6\",\"name\":\"MyACL6\",\"type\":\"openconfig-acl:ACL_IPV6\"},\"name\":\"MyACL6\",\"state\":{\"description\":\"Description for IPv6 ACL MyACL6\",\"name\":\"MyACL6\",\"type\":\"openconfig-acl:ACL_IPV6\"},\"type\":\"openconfig-acl:ACL_IPV6\"}]}"

var getIPv6AllPortsBindingsResponse string = "{\"openconfig-acl:interfaces\":{\"interface\":[{\"config\":{\"id\":\"Ethernet4\"},\"id\":\"Ethernet4\",\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":6,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":6}}]},\"config\":{\"set-name\":\"MyACL6\",\"type\":\"openconfig-acl:ACL_IPV6\"},\"set-name\":\"MyACL6\",\"state\":{\"set-name\":\"MyACL6\",\"type\":\"openconfig-acl:ACL_IPV6\"},\"type\":\"openconfig-acl:ACL_IPV6\"}]},\"state\":{\"id\":\"Ethernet4\"}}]}}"

var oneL2AclCreateJsonRequest string = "{\"config\": {\"name\": \"MyACL2\",\"type\": \"ACL_L2\",\"description\": \"Description for L2 ACL MyACL2\"}}"
var oneL2AclCreateJsonResponse string = "{\"openconfig-acl:acl-set\":[{\"config\":{\"description\":\"Description for L2 ACL MyACL2\",\"name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_L2\"},\"name\":\"MyACL2\",\"state\":{\"description\":\"Description for L2 ACL MyACL2\",\"name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_L2\"},\"type\":\"openconfig-acl:ACL_L2\"}]}"

var oneL2RuleCreateJsonRequest string = "{\"sequence-id\": 2,\"config\": {\"sequence-id\": 2,\"description\": \"Description for MyACL2 Rule Seq 2\"},\"l2\": {\"config\": {\"ethertype\": \"ETHERTYPE_VLAN\"}},\"transport\": {\"config\": {\"source-port\": 101,\"destination-port\": 100,\"tcp-flags\": [\"TCP_FIN\",\"TCP_ACK\"]}},\"actions\": {\"config\": {\"forwarding-action\": \"ACCEPT\"}}}"

var oneL2RuleCreateJsonResponse string = "{\"openconfig-acl:acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":2},\"l2\":{\"config\":{\"ethertype\":\"openconfig-packet-match-types:ETHERTYPE_VLAN\"},\"state\":{\"ethertype\":\"openconfig-packet-match-types:ETHERTYPE_VLAN\"}},\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2},\"transport\":{\"config\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]},\"state\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]}}}]}"

var ingressL2AclSetCreateJsonRequest string = "{ \"openconfig-acl:config\": { \"set-name\": \"MyACL2\", \"type\": \"ACL_L2\" }}"
var ingressL2AclSetCreateJsonResponse string = "{\"openconfig-acl:ingress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2}}]},\"config\":{\"set-name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_L2\"},\"set-name\":\"MyACL2\",\"state\":{\"set-name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_L2\"},\"type\":\"openconfig-acl:ACL_L2\"}]}"

var getL2AclsFromAclSetListLevelResponse string = "{\"openconfig-acl:acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"},\"state\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":2},\"l2\":{\"config\":{\"ethertype\":\"openconfig-packet-match-types:ETHERTYPE_VLAN\"},\"state\":{\"ethertype\":\"openconfig-packet-match-types:ETHERTYPE_VLAN\"}},\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2},\"transport\":{\"config\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]},\"state\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]}}}]},\"config\":{\"description\":\"Description for L2 ACL MyACL2\",\"name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_L2\"},\"name\":\"MyACL2\",\"state\":{\"description\":\"Description for L2 ACL MyACL2\",\"name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_L2\"},\"type\":\"openconfig-acl:ACL_L2\"}]}"

var getL2AllPortsBindingsResponse string = "{\"openconfig-acl:interfaces\":{\"interface\":[{\"config\":{\"id\":\"Ethernet0\"},\"id\":\"Ethernet0\",\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2}}]},\"config\":{\"set-name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_L2\"},\"set-name\":\"MyACL2\",\"state\":{\"set-name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_L2\"},\"type\":\"openconfig-acl:ACL_L2\"}]},\"state\":{\"id\":\"Ethernet0\"}}]}}"

var aclCreateWithInvalidInterfaceBinding string = "{ \"acl-sets\": { \"acl-set\": [ { \"name\": \"MyACL1\", \"type\": \"ACL_IPV4\", \"config\": { \"name\": \"MyACL1\", \"type\": \"ACL_IPV4\", \"description\": \"Description for MyACL1\" }, \"acl-entries\": { \"acl-entry\": [ { \"sequence-id\": 1, \"config\": { \"sequence-id\": 1, \"description\": \"Description for MyACL1 Rule Seq 1\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.1/32\", \"destination-address\": \"21.1.1.1/32\", \"dscp\": 1, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 101, \"destination-port\": 201 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } } ] } } ] }, \"interfaces\": { \"interface\": [ { \"id\": \"Ethernet2112\", \"config\": { \"id\": \"Ethernet2112\" }, \"interface-ref\": { \"config\": { \"interface\": \"Ethernet2112\" } }, \"ingress-acl-sets\": { \"ingress-acl-set\": [ { \"set-name\": \"MyACL1\", \"type\": \"ACL_IPV4\", \"config\": { \"set-name\": \"MyACL1\", \"type\": \"ACL_IPV4\" } } ] } } ] }}"

var requestOneDuplicateRulePostJson string = "{\"sequence-id\": 1,\"config\": {\"sequence-id\": 1,\"description\": \"Description for MyACL3 Rule Seq 1\"},\"ipv4\": {\"config\": {\"source-address\": \"4.4.4.4/24\",\"destination-address\": \"5.5.5.5/24\",\"protocol\": \"IP_TCP\"}},\"transport\": {\"config\": {\"source-port\": 101,\"destination-port\": 100,\"tcp-flags\": [\"TCP_FIN\",\"TCP_ACK\"]}},\"actions\": {\"config\": {\"forwarding-action\": \"ACCEPT\"}}}"
