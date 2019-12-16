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
	"testing"
	db "translib/db"
	"os"
)

const (
	CFG_L2MC_TABLE               = "CFG_L2MC_TABLE"
	CFG_L2MC_MROUTER_TABLE       = "CFG_L2MC_MROUTER_TABLE"
	CFG_L2MC_STATIC_GROUP_TABLE  = "CFG_L2MC_STATIC_GROUP_TABLE"
	CFG_L2MC_STATIC_MEMBER_TABLE = "CFG_L2MC_STATIC_MEMBER_TABLE"
	APP_L2MC_MROUTER_TABLE       = "APP_L2MC_MROUTER_TABLE"
	APP_L2MC_MEMBER_TABLE        = "APP_L2MC_MEMBER_TABLE"
)

func clearDb() {
	fmt.Println("---------  Init IGMP Snooping Go test  --------")

	if err := clearIgmpSnoopingDataFromConfigDb(); err == nil {
		fmt.Println("----- Removed All IGMP Snooping Data from Db  -------")
		createVlanInterface ()
	} else {
		fmt.Printf("Failed to remove All IGMP Snooping from Db: %v", err)
		os.Exit(1) // Cancel any further tests.
	}
}

func createVlanInterface () {
	var createVlan string = "{\"openconfig-interfaces:config\": {\"name\": \"Vlan5\"}}"
	processSetRequest("/openconfig-interfaces:interfaces/interface=Vlan5/config", createVlan, "PATCH", false)
	processGetRequest("/openconfig-interfaces:interfaces/interface=Vlan5/config", createVlan, false)
}

func TestIGMPSnoopingConfigPostDeleteGetAPIs(t *testing.T) {
	clearDb()
	//POST - config container 
	t.Run("POST - Config container", processSetRequest(igmpsIntfUrl, configNodeReq, "POST", false))
	t.Run("Verify: POST - Config container", processGetRequest(configUrl, configNodeReq, false))
	t.Run("Delete - Config container", processDeleteRequest(configUrl))
	t.Run("Verify: Delete - Config container", processGetRequest(configUrl, "{}", false))
		
	//POST - mrouter 
	t.Run("POST - mrouter", processSetRequest(configUrl, mrouterReq, "POST", false))
	t.Run("Verify: POST - mrouter", processGetRequest(mrouterUrl, mrouterReq, false))
	t.Run("Delete - mrouterr", processDeleteRequest(mrouterUrl))
	t.Run("Verify: Delete - mrouter", processGetRequest(mrouterUrl, "{}", false))
	
	//POST - static-group 
	t.Run("POST - static group", processSetRequest(configUrl, staticGroupReq, "POST", false))
	t.Run("Verify: POST - static group", processGetRequest(staticGrpUrl, staticGroupReq, false))
	t.Run("Delete - static group", processDeleteRequest(staticGrpUrl))
	t.Run("Verify: Delete - static group", processGetRequest(staticGrpUrl, "{}", false))

	//POST - static-group only
	t.Run("POST - static-group only", processSetRequest(configUrl, staticGroupOnlyReq, "POST", false))
	t.Run("Verify: POST - static-group only", processGetRequest(staticGrpUrl, staticGroupOnlyReq, false))
	t.Run("Delete - static-group only", processDeleteRequest(staticGrpUrl))
	t.Run("Verify: Delete - static-group only", processGetRequest(staticGrpUrl, "{}", false))

	//POST - static-group intf only
	t.Run("POST - static group", processSetRequest(configUrl, staticGroupOnlyReq, "POST", false))
	t.Run("Verify: POST - mrouter", processGetRequest(staticGrpUrl, staticGroupOnlyReq, false))
	grpUrl := staticGrpUrl + "[group=1.2.3.4]"	
	t.Run("POST - static-group intf only", processSetRequest(staticGrpUrl, staticGroupIntfReq, "POST", false))
	t.Run("Verify: POST - static-group intf only", processGetRequest(grpUrl+"/outgoing-interface", staticGroupIntfReq, false))
	t.Run("Delete - static-group intf only ", processDeleteRequest(grpUrl+"/outgoing-interface"))
	t.Run("Verify: Delete - static-group intf only ", processGetRequest(grpUrl+"/outgoing-interface", "{}", false))
	t.Run("Delete - static-group", processDeleteRequest(staticGrpUrl))
	
	//DELETE - static-group intf delete only
	t.Run("POST - static group", processSetRequest(configUrl, staticGroupReq, "POST", false))
	t.Run("Verify: POST - static group", processGetRequest(staticGrpUrl, staticGroupReq, false))
	t.Run("Delete - static-group intf delete only ", processDeleteRequest(staticGrpUrl+"/outgoing-interface=Ethernet8"))
	t.Run("Verify: Delete - static-group intf delete only ", processGetRequest(staticGrpUrl+"/outgoing-interface=Ethernet8", "{}", false))
	t.Run("Delete - static-group", processDeleteRequest(staticGrpUrl))
	
	//POST - igmps enable 
	t.Run("POST - igmps enable ", processSetRequest(configUrl, igmpsEnableReq, "POST", false))
	t.Run("Verify: POST - igmps enable ", processGetRequest(igmpsEnableUrl, igmpsEnableReq, false))
	t.Run("Delete - igmps enable ", processDeleteRequest(igmpsEnableUrl))
	t.Run("Verify: Delete - igmps enable ", processGetRequest(igmpsEnableUrl, "{}", false))

	//POST - fast-leave 
	t.Run("POST - static group", processSetRequest(configUrl, fastLeaveReq, "POST", false))
	t.Run("Verify: POST - mrouter", processGetRequest(fastLeaveUrl, fastLeaveReq, false))
	t.Run("Delete - mrouterr", processDeleteRequest(fastLeaveUrl))
	t.Run("Verify: Delete - mrouter", processGetRequest(fastLeaveUrl, "{}", false))

	//POST - querier
	t.Run("POST - querier", processSetRequest(configUrl, querierReq, "POST", false))
	t.Run("Verify: POST - querier", processGetRequest(querierUrl, querierReq, false))
	t.Run("Delete - querier", processDeleteRequest(querierUrl))
	t.Run("Verify: Delete - querier", processGetRequest(querierUrl, "{}", false))

	//POST - last-memeber 
	t.Run("POST - last-memeber ", processSetRequest(configUrl, lastMemReq, "POST", false))
	t.Run("Verify: POST - last-memeber ", processGetRequest(lastMemUrl, lastMemReq, false))
	t.Run("Delete - last-memeber ", processDeleteRequest(lastMemUrl))
	t.Run("Verify: Delete - last-memeber ", processGetRequest(lastMemUrl, "{}", false))

	//POST - version 
	t.Run("POST - version", processSetRequest(configUrl, versionReq, "POST", false))
	t.Run("Verify: POST - version", processGetRequest(versionUrl, versionReq, false))
	t.Run("Delete - version", processDeleteRequest(versionUrl))
	t.Run("Verify: Delete - version", processGetRequest(versionUrl, "{}", false))

	//POST - query-max-response 
	t.Run("POST - query-max-response ", processSetRequest(configUrl, qryMaxTimeReq, "POST", false))
	t.Run("Verify: POST - query-max-response ", processGetRequest(maxRespTimenUrl, qryMaxTimeReq, false))
	t.Run("Delete - query-max-response ", processDeleteRequest(maxRespTimenUrl))
	t.Run("Verify: Delete - query-max-response ", processGetRequest(maxRespTimenUrl, "{}", false))
	
	//POST - query-interval  
	t.Run("POST - query-interval ", processSetRequest(configUrl, qryIntvlReq, "POST", false))
	t.Run("Verify: POST - query-interval ", processGetRequest(qryIntvlUrl, qryIntvlReq, false))
	t.Run("Delete - query-interval ", processDeleteRequest(qryIntvlUrl))
	t.Run("Verify: Delete - query-interval ", processGetRequest(qryIntvlUrl, "{}", false))
}

func TestIGMPSnoopingConfigPatchDeleteGetAPIs(t *testing.T) {
	clearDb()		
	//PATCH - config container 
	t.Run("PATCH - Config container", processSetRequest(configUrl, configNodeReq, "PATCH", false))
	t.Run("Verify: PATCH - Config container", processGetRequest(configUrl, configNodeReq, false))
	t.Run("Delete - Config container", processDeleteRequest(configUrl))
	t.Run("Verify: Delete - Config container", processGetRequest(configUrl, "{}", false))
		
	//PATCH - mrouter 
	t.Run("PATCH - mrouter", processSetRequest(mrouterUrl, mrouterReq, "PATCH", false))
	t.Run("Verify: PATCH - mrouter", processGetRequest(mrouterUrl, mrouterReq, false))
	t.Run("Delete - mrouterr", processDeleteRequest(mrouterUrl))
	t.Run("Verify: Delete - mrouter", processGetRequest(mrouterUrl, "{}", false))
	
	//PATCH - static-group 
	t.Run("PATCH - static group", processSetRequest(staticGrpUrl, staticGroupReq, "PATCH", false))
	t.Run("Verify: PATCH - static group", processGetRequest(staticGrpUrl, staticGroupReq, false))
	t.Run("Delete - static group", processDeleteRequest(staticGrpUrl))
	t.Run("Verify: Delete - static group", processGetRequest(staticGrpUrl, "{}", false))
	
	//PATCH - static-group only
	t.Run("PATCH - static-group only", processSetRequest(staticGrpUrl, staticGroupOnlyReq, "PATCH", false))
	t.Run("Verify: PATCH - static-group only", processGetRequest(staticGrpUrl, staticGroupOnlyReq, false))
	t.Run("Delete - static-group only", processDeleteRequest(staticGrpUrl))
	t.Run("Verify: Delete - static-group only", processGetRequest(staticGrpUrl, "{}", false))

	//PATCH - static-group intf only
	t.Run("PATCH - PATCH group", processSetRequest(staticGrpUrl, staticGroupOnlyReq, "PATCH", false))
	t.Run("Verify: PATCH - group", processGetRequest(staticGrpUrl, staticGroupOnlyReq, false))
	grpUrl := staticGrpUrl + "[group=1.2.3.4]"	
	t.Run("PATCH - static-group intf only", processSetRequest(grpUrl+"/outgoing-interface", staticGroupIntfReq, "PATCH", false))
	t.Run("Verify: PATCH - static-group intf only", processGetRequest(grpUrl+"/outgoing-interface", staticGroupIntfReq, false))
	t.Run("Delete - static-group intf only ", processDeleteRequest(grpUrl+"/outgoing-interface"))
	t.Run("Verify: Delete - static-group intf only ", processGetRequest(grpUrl+"/outgoing-interface", "{}", false))
	t.Run("Delete - static-group", processDeleteRequest(staticGrpUrl))
	
	//PATCH - igmps enable 
	t.Run("PATCH - igmps enable ", processSetRequest(igmpsEnableUrl, igmpsEnableReq, "PATCH", false))
	t.Run("Verify: PATCH - igmps enable ", processGetRequest(igmpsEnableUrl, igmpsEnableReq, false))
	t.Run("Delete - igmps enable ", processDeleteRequest(igmpsEnableUrl))
	t.Run("Verify: Delete - igmps enable ", processGetRequest(igmpsEnableUrl, "{}", false))

	//PATCH - fast-leave 
	t.Run("PATCH - fast-leave ", processSetRequest(fastLeaveUrl, fastLeaveReq, "PATCH", false))
	t.Run("Verify: PATCH - mrouter", processGetRequest(fastLeaveUrl, fastLeaveReq, false))
	t.Run("Delete - mrouterr", processDeleteRequest(fastLeaveUrl))
	t.Run("Verify: Delete - mrouter", processGetRequest(fastLeaveUrl, "{}", false))

	//PATCH - querier
	t.Run("PATCH - querier", processSetRequest(querierUrl, querierReq, "PATCH", false))
	t.Run("Verify: PATCH - querier", processGetRequest(querierUrl, querierReq, false))
	t.Run("Delete - querier", processDeleteRequest(querierUrl))
	t.Run("Verify: Delete - querier", processGetRequest(querierUrl, "{}", false))

	//PATCH - last-memeber 
	t.Run("PATCH - last-memeber ", processSetRequest(lastMemUrl, lastMemReq, "PATCH", false))
	t.Run("Verify: PATCH - last-memeber ", processGetRequest(lastMemUrl, lastMemReq, false))
	t.Run("Delete - last-memeber ", processDeleteRequest(lastMemUrl))
	t.Run("Verify: Delete - last-memeber ", processGetRequest(lastMemUrl, "{}", false))

	//PATCH - version 
	t.Run("PATCH - version", processSetRequest(versionUrl, versionReq, "PATCH", false))
	t.Run("Verify: PATCH - version", processGetRequest(versionUrl, versionReq, false))
	t.Run("Delete - version", processDeleteRequest(versionUrl))
	t.Run("Verify: Delete - version", processGetRequest(versionUrl, "{}", false))

	//PATCH - query-max-response 
	t.Run("PATCH - query-max-response ", processSetRequest(maxRespTimenUrl, qryMaxTimeReq, "PATCH", false))
	t.Run("Verify: PATCH - query-max-response ", processGetRequest(maxRespTimenUrl, qryMaxTimeReq, false))
	t.Run("Delete - query-max-response ", processDeleteRequest(maxRespTimenUrl))
	t.Run("Verify: Delete - query-max-response ", processGetRequest(maxRespTimenUrl, "{}", false))
	
	//PATCH - query-interval 
	t.Run("PATCH - query-interval ", processSetRequest(qryIntvlUrl, qryIntvlReq, "PATCH", false))
	t.Run("Verify: PATCH - query-interval ", processGetRequest(qryIntvlUrl, qryIntvlReq, false))
	t.Run("Delete - query-interval ", processDeleteRequest(qryIntvlUrl))
	t.Run("Verify: Delete - query-interval", processGetRequest(qryIntvlUrl, "{}", false))
}

func TestIGMPSnoopingConfigPutDeleteGetAPIs(t *testing.T) {
	clearDb()		
	//PUT - config container 
	t.Run("PUT - Config container", processSetRequest(configUrl, configNodeReq, "PUT", false))
	t.Run("Verify: PUT - Config container", processGetRequest(configUrl, configNodeReq, false))
	t.Run("Delete - Config container", processDeleteRequest(configUrl))
	t.Run("Verify: Delete - Config container", processGetRequest(configUrl, "{}", false))
		
	//PUT - mrouter 
	t.Run("PUT - mrouter", processSetRequest(mrouterUrl, mrouterReq, "PUT", false))
	t.Run("Verify: PUT - mrouter", processGetRequest(mrouterUrl, mrouterReq, false))
	t.Run("Delete - mrouterr", processDeleteRequest(mrouterUrl))
	t.Run("Verify: Delete - mrouter", processGetRequest(mrouterUrl, "{}", false))
	
	//PUT - static-group 
	t.Run("PUT - static group", processSetRequest(staticGrpUrl, staticGroupReq, "PUT", false))
	t.Run("Verify: PUT - mrouter", processGetRequest(staticGrpUrl, staticGroupReq, false))
	t.Run("Delete - mrouterr", processDeleteRequest(staticGrpUrl))
	t.Run("Verify: Delete - mrouter", processGetRequest(staticGrpUrl, "{}", false))
	
	//PUT - static-group only
	t.Run("PUT - static-group only", processSetRequest(staticGrpUrl, staticGroupOnlyReq, "PUT", false))
	t.Run("Verify: PUT - static-group only", processGetRequest(staticGrpUrl, staticGroupOnlyReq, false))
	t.Run("Delete - static-group only", processDeleteRequest(staticGrpUrl))
	t.Run("Verify: Delete - static-group only", processGetRequest(staticGrpUrl, "{}", false))

	//PUT - static-group intf only
	t.Run("PUT - PUT group", processSetRequest(staticGrpUrl, staticGroupOnlyReq, "PUT", false))
	t.Run("Verify: PUT - mrouter", processGetRequest(staticGrpUrl, staticGroupOnlyReq, false))
	grpUrl := staticGrpUrl + "[group=1.2.3.4]"	
	t.Run("PUT - static-group intf only", processSetRequest(grpUrl+"/outgoing-interface", staticGroupIntfReq, "PUT", false))
	t.Run("Verify: PUT - static-group intf only", processGetRequest(grpUrl+"/outgoing-interface", staticGroupIntfReq, false))
	t.Run("Delete - static-group intf only ", processDeleteRequest(grpUrl+"/outgoing-interface"))
	t.Run("Verify: Delete - static-group intf only ", processGetRequest(grpUrl+"/outgoing-interface", "{}", false))
	t.Run("Delete - static-group", processDeleteRequest(staticGrpUrl))	
	
	//PUT - igmps enable 
	t.Run("PUT - igmps enable ", processSetRequest(igmpsEnableUrl, igmpsEnableReq, "PUT", false))
	t.Run("Verify: PUT - igmps enable ", processGetRequest(igmpsEnableUrl, igmpsEnableReq, false))
	t.Run("Delete - igmps enable ", processDeleteRequest(igmpsEnableUrl))
	t.Run("Verify: Delete - igmps enable ", processGetRequest(igmpsEnableUrl, "{}", false))

	//PUT - fast-leave 
	t.Run("PUT - fast-leave ", processSetRequest(fastLeaveUrl, fastLeaveReq, "PUT", false))
	t.Run("Verify: PUT - mrouter", processGetRequest(fastLeaveUrl, fastLeaveReq, false))
	t.Run("Delete - mrouterr", processDeleteRequest(fastLeaveUrl))
	t.Run("Verify: Delete - mrouter", processGetRequest(fastLeaveUrl, "{}", false))

	//PUT - querier
	t.Run("PUT - querier", processSetRequest(querierUrl, querierReq, "PUT", false))
	t.Run("Verify: PUT - querier", processGetRequest(querierUrl, querierReq, false))
	t.Run("Delete - querier", processDeleteRequest(querierUrl))
	t.Run("Verify: Delete - querier", processGetRequest(querierUrl, "{}", false))

	//PUT - last-memeber 
	t.Run("PUT - last-memeber ", processSetRequest(lastMemUrl, lastMemReq, "PUT", false))
	t.Run("Verify: PUT - last-memeber ", processGetRequest(lastMemUrl, lastMemReq, false))
	t.Run("Delete - last-memeber ", processDeleteRequest(lastMemUrl))
	t.Run("Verify: Delete - last-memeber ", processGetRequest(lastMemUrl, "{}", false))

	//PUT - version 
	t.Run("PUT - version", processSetRequest(versionUrl, versionReq, "PUT", false))
	t.Run("Verify: PUT - version", processGetRequest(versionUrl, versionReq, false))
	t.Run("Delete - version", processDeleteRequest(versionUrl))
	t.Run("Verify: Delete - version", processGetRequest(versionUrl, "{}", false))

	//PUT - query-max-response 
	t.Run("PUT - query-max-response ", processSetRequest(maxRespTimenUrl, qryMaxTimeReq, "PUT", false))
	t.Run("Verify: PUT - query-max-response ", processGetRequest(maxRespTimenUrl, qryMaxTimeReq, false))
	t.Run("Delete - query-max-response ", processDeleteRequest(maxRespTimenUrl))
	t.Run("Verify: Delete - query-max-response ", processGetRequest(maxRespTimenUrl, "{}", false))
	
	//PUT - query-interval 
	t.Run("PUT - query-interval ", processSetRequest(qryIntvlUrl, qryIntvlReq, "PUT", false))
	t.Run("Verify: PUT - query-interval ", processGetRequest(qryIntvlUrl, qryIntvlReq, false))
	t.Run("Delete - query-interval ", processDeleteRequest(qryIntvlUrl))
	t.Run("Verify: Delete - query-interval", processGetRequest(qryIntvlUrl, "{}", false))	
}

func TestIGMPSnoopingConfigNegativeAPIs(t *testing.T) {
	clearDb()
	var lastMemNegReq string = "{\"last-member-query-interval\":30000}"
	
	var versionNegReq string = "{\"version\":10}"
	
	var qryMaxTimeNegReq string = "{\"query-max-response-time\":11000}"
	
	var qryIntvlNegReq string = "{\"query-interval\":126000}"
	
	//PATCH - last-memeber 
	t.Run("PATCH - last-memeber ", processSetRequest(lastMemUrl, lastMemNegReq, "PATCH", false))
	t.Run("Verify: PATCH - last-memeber ", processGetRequest(lastMemUrl, lastMemNegReq, false))

	//PATCH - version 
	t.Run("PATCH - version", processSetRequest(versionUrl, versionNegReq, "PATCH", false))
	t.Run("Verify: PATCH - version", processGetRequest(versionUrl, versionNegReq, false))

	//PATCH - query-max-response 
	t.Run("PATCH - query-max-response ", processSetRequest(maxRespTimenUrl, qryMaxTimeNegReq, "PATCH", false))
	t.Run("Verify: PATCH - query-max-response ", processGetRequest(maxRespTimenUrl, qryMaxTimeNegReq, false))
	
	//PATCH - query-interval 
	t.Run("PATCH - query-interval ", processSetRequest(qryIntvlUrl, qryIntvlNegReq, "PATCH", false))
	t.Run("Verify: PATCH - query-interval ", processGetRequest(qryIntvlUrl, qryIntvlNegReq, false))
	t.Run("Verify: config container", processGetRequest(configUrl, "{}", false))
}

func TestIGMPSnoopingStateGetAPIs(t *testing.T) {
	clearDb()
	// patch igmp snooping config.
	t.Run("PATCH - Config container", processSetRequest(configUrl, configNodeReq, "PATCH", false))
	t.Run("Verify: PATCH - Config container", processGetRequest(configUrl, configNodeReq, false))

	//GET - state container 
	t.Run("Verify: GET - state container", processGetRequest(stateUrl, configNodeReq, false))

	//GET - state mrouter 
	t.Run("Verify: GET - state mrouter", processGetRequest(mrouterStateUrl, mrouterReq, false))

	//GET - state static-group 
	t.Run("Verify: GET - state static-group ", processGetRequest(staticGrpStateUrl, staticGroupReq, false))

	//GET - state static-group with keys 
	t.Run("Verify: GET - state static-group with keys ", processGetRequest(staticGrpStateUrl+"=1.2.3.4,255.0.0.0", staticGroupReq, false))

	//GET - state static-group intf 
	t.Run("Verify: GET - state static-group intf only", processGetRequest(staticGrpStateUrl+"=1.2.3.4,255.0.0.0/outgoing-interface", staticGroupIntfReq, false))

	//GET - state static-group intf with value
	t.Run("Verify: GET - state static-group intf only", processGetRequest(staticGrpStateUrl+"=1.2.3.4,255.0.0.0/outgoing-interface=Ethernet8", staticGroupIntfReq, false))

	//GET - state igmp enable 
	t.Run("Verify: GET - state state container", processGetRequest(igmpsEnableStateUrl, igmpsEnableReq, false))

	//GET - fast-leave 
	t.Run("Verify: GET - state mrouter", processGetRequest(fastLeaveStateUrl, fastLeaveReq, false))

	//GET - querier
	t.Run("Verify: GET - state querier", processGetRequest(querierStateUrl, querierReq, false))

	//GET - last-memeber 
	t.Run("Verify: GET - state last-memeber ", processGetRequest(lastMemStateUrl, lastMemReq, false))

	//GET - version 
	t.Run("Verify: GET - state version", processGetRequest(versionStateUrl, versionReq, false))

	//GET - query-max-response 
	t.Run("Verify: GET - state query-max-response ", processGetRequest(maxRespTimenStateUrl, qryMaxTimeReq, false))
	
	//GET - query-interval  
	t.Run("Verify: GET - state query-interval ", processGetRequest(qryIntvlStateUrl, qryIntvlReq, false))
}

func clearIgmpSnoopingDataFromConfigDb() error {
	var err error
	
	cgf_l2mc_tbl_ts := db.TableSpec{Name: "CFG_L2MC_TABLE"}
	cgf_l2mc_mrouter_tbl_ts := db.TableSpec{Name: "CFG_L2MC_MROUTER_TABLE"}
	cgf_l2mc_static_grp_tbl_ts := db.TableSpec{Name: "CFG_L2MC_STATIC_GROUP_TABLE"}
	cgf_l2mc_static_mem_tbl_ts := db.TableSpec{Name: "CFG_L2MC_STATIC_MEMBER_TABLE"}

	d := getConfigDb()
	
	if d == nil {
		err = errors.New("Failed to connect to config Db")
		return err
	}

	if err = d.DeleteTable(&cgf_l2mc_tbl_ts); err != nil {
		err = errors.New("Failed to delete CFG_L2MC_TABLE Table")
		return err
	}

	if err = d.DeleteTable(&cgf_l2mc_mrouter_tbl_ts); err != nil {
		err = errors.New("Failed to delete CFG_L2MC_MROUTER_TABLE Table")
		return err
	}

	if err = d.DeleteTable(&cgf_l2mc_static_grp_tbl_ts); err != nil {
		err = errors.New("Failed to delete CFG_L2MC_STATIC_GROUP_TABLE Table")
		return err
	}

	if err = d.DeleteTable(&cgf_l2mc_static_mem_tbl_ts); err != nil {
		err = errors.New("Failed to delete CFG_L2MC_STATIC_MEMBER_TABLE Table")
		return err
	}

	return err
}

//config-URL
var igmpsIntfUrl string = "/openconfig-network-instance:network-instances/network-instance[name=default]/protocols/protocol[identifier=IGMP_SNOOPING][name=IGMP-SNOOPING]/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface[name=Vlan5]"
var configUrl string = igmpsIntfUrl+"/config"
var mrouterUrl string = configUrl+"/mrouter-interface"
var staticGrpUrl string = configUrl+"/static-multicast-group"
var igmpsEnableUrl string = configUrl+"/enabled"
var fastLeaveUrl string = configUrl+"/fast-leave"
var querierUrl string = configUrl+"/querier"
var lastMemUrl string =  configUrl+"/last-member-query-interval"
var versionUrl string =  configUrl+"/version"
var maxRespTimenUrl string = configUrl+"/query-max-response-time"
var qryIntvlUrl string = configUrl+"/query-interval"

//state-URL
var stateUrl string = igmpsIntfUrl+"/state"
var mrouterStateUrl string = stateUrl+"/mrouter-interface"
var staticGrpStateUrl string = stateUrl+"/static-multicast-group"
var igmpsEnableStateUrl string = stateUrl+"/enabled"
var fastLeaveStateUrl string = stateUrl+"/fast-leave"
var querierStateUrl string = stateUrl+"/querier"
var lastMemStateUrl string =  stateUrl+"/last-member-query-interval"
var versionStateUrl string =  stateUrl+"/version"
var maxRespTimenStateUrl string = stateUrl+"/query-max-response-time"
var qryIntvlStateUrl string = stateUrl+"/query-interval"

//JSON data
var configNodeReq string = "{\"openconfig-network-instance-deviation:config\":{\"enabled\":true,\"fast-leave\":true,\"last-member-query-interval\":1001,\"mrouter-interface\":[\"Ethernet8\",\"Ethernet12\"],\"name\":\"Vlan5\",\"querier\":true,\"query-interval\":126,\"query-max-response-time\":11,\"static-multicast-group\":[{\"group\":\"1.2.3.4\",\"outgoing-interface\":[\"Ethernet12\",\"Ethernet8\"]}],\"version\":3}}" 

var mrouterReq string = "{\"openconfig-network-instance-deviation:mrouter-interface\":[\"Ethernet8\",\"Ethernet12\"]}"

var staticGroupReq string = "{\"openconfig-network-instance-deviation:static-multicast-group\":[{\"group\":\"1.2.3.4\",\"outgoing-interface\":[\"Ethernet12\",\"Ethernet8\"]}]}"

var staticGroupOnlyReq string = "{\"openconfig-network-instance-deviation:static-multicast-group\":[{\"group\":\"1.2.3.4\"}]}"

var staticGroupIntfReq string = "{\"openconfig-network-instance-deviation:outgoing-interface\":[\"Ethernet12\",\"Ethernet8\"]}"

var igmpsEnableReq string = "{\"openconfig-network-instance-deviation:enabled\":true}"

var querierReq string = "{\"openconfig-network-instance-deviation:querier\":true}"

var fastLeaveReq string = "{\"openconfig-network-instance-deviation:fast-leave\":true}"

var lastMemReq string = "{\"openconfig-network-instance-deviation:last-member-query-interval\":1001}"

var versionReq string = "{\"openconfig-network-instance-deviation:version\":3}"

var qryMaxTimeReq string = "{\"openconfig-network-instance-deviation:query-max-response-time\":11}"

var qryIntvlReq string = "{\"openconfig-network-instance-deviation:query-interval\":126}"
