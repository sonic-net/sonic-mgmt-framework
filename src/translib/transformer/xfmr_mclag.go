////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Dell, Inc.                                                 //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//  http://www.apache.org/licenses/LICENSE-2.0                                //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package transformer

import (
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
	"strconv"
	"strings"
	"translib/db"
	"translib/ocbinds"
)

func init() {
	XlateFuncBind("YangToDb_mclag_domainid_fld_xfmr", YangToDb_mclag_domainid_fld_xfmr)
	XlateFuncBind("DbToYang_mclag_domainid_fld_xfmr", DbToYang_mclag_domainid_fld_xfmr)
	XlateFuncBind("YangToDb_mclag_vlan_name_fld_xfmr", YangToDb_mclag_vlan_name_fld_xfmr)
	XlateFuncBind("DbToYang_mclag_vlan_name_fld_xfmr", DbToYang_mclag_vlan_name_fld_xfmr)
	XlateFuncBind("YangToDb_mclag_interface_subtree_xfmr", YangToDb_mclag_interface_subtree_xfmr)
	XlateFuncBind("DbToYang_mclag_interface_subtree_xfmr", DbToYang_mclag_interface_subtree_xfmr)

	XlateFuncBind("DbToYang_mclag_domain_oper_status_fld_xfmr", DbToYang_mclag_domain_oper_status_fld_xfmr)
	XlateFuncBind("DbToYang_mclag_domain_role_fld_xfmr", DbToYang_mclag_domain_role_fld_xfmr)
	XlateFuncBind("DbToYang_mclag_domain_system_mac_fld_xfmr", DbToYang_mclag_domain_system_mac_fld_xfmr)
}

var YangToDb_mclag_domainid_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	log.Info("YangToDb_mclag_domainid_fld_xfmr: ", inParams.key)

	return res_map, err
}

var DbToYang_mclag_domainid_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	log.Info("DbToYang_mclag_domainid_fld_xfmr: ", inParams.key)
	result["domain-id"], _ = strconv.ParseUint(inParams.key, 10, 32)

	return result, err
}

var YangToDb_mclag_vlan_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	log.Info("YangToDb_mclag_vlan_name_fld_xfmr: ", inParams.key)

	return res_map, err
}

var DbToYang_mclag_vlan_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	log.Info("DbToYang_mclag_vlan_name_fld_xfmr: ", inParams.key)
	result["name"] = inParams.key

	return result, err
}

var DbToYang_mclag_domain_oper_status_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	log.Infof("DbToYang_mclag_domain_oper_status_fld_xfmr --> key: %v", inParams.key)

	stDb := inParams.dbs[db.StateDB]
	mclagEntry, _ := stDb.GetEntry(&db.TableSpec{Name: "MCLAG_TABLE"}, db.Key{Comp: []string{inParams.key}})
	operStatus := mclagEntry.Get("oper_status")
	if operStatus == "up" {
		result["oper-status"], _ = ygot.EnumName(ocbinds.OpenconfigMclag_Mclag_MclagDomains_MclagDomain_State_OperStatus_OPER_UP)
	} else {
		result["oper-status"], _ = ygot.EnumName(ocbinds.OpenconfigMclag_Mclag_MclagDomains_MclagDomain_State_OperStatus_OPER_DOWN)
	}

	return result, err
}

var DbToYang_mclag_domain_role_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	log.Infof("DbToYang_mclag_domain_role_fld_xfmr --> key: %v", inParams.key)

	stDb := inParams.dbs[db.StateDB]
	mclagEntry, _ := stDb.GetEntry(&db.TableSpec{Name: "MCLAG_TABLE"}, db.Key{Comp: []string{inParams.key}})
	role := mclagEntry.Get("role")
	if role == "active" {
		result["role"], _ = ygot.EnumName(ocbinds.OpenconfigMclag_Mclag_MclagDomains_MclagDomain_State_Role_ROLE_ACTIVE)
	} else {
		result["role"], _ = ygot.EnumName(ocbinds.OpenconfigMclag_Mclag_MclagDomains_MclagDomain_State_Role_ROLE_STANDBY)
	}

	return result, err
}

var DbToYang_mclag_domain_system_mac_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	log.Infof("DbToYang_mclag_domain_system_mac_fld_xfmr --> key: %v", inParams.key)

	stDb := inParams.dbs[db.StateDB]
	mclagEntry, _ := stDb.GetEntry(&db.TableSpec{Name: "MCLAG_TABLE"}, db.Key{Comp: []string{inParams.key}})
	sysmac := mclagEntry.Get("system_mac")
	result["system-mac"] = &sysmac

	return result, err
}

var YangToDb_mclag_interface_subtree_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
	var err error
	res_map := make(map[string]map[string]db.Value)
	mclagIntfTblMap := make(map[string]db.Value)
	log.Info("YangToDb_mclag_interface_subtree_xfmr: ", inParams.ygRoot, inParams.uri)

	mclagObj := getMclagRoot(inParams.ygRoot)
	if mclagObj.Interfaces == nil {
		return res_map, err
	}

	for intfId, _ := range mclagObj.Interfaces.Interface {
		intf := mclagObj.Interfaces.Interface[intfId]
		if intf != nil && intf.Config != nil {
			mclagIntfKey := strconv.Itoa(int(*intf.Config.MclagDomainId)) + "|" + *intf.Name
			log.Infof("YangToDb_mclag_interface_subtree_xfmr --> key: %v", mclagIntfKey)

			_, ok := mclagIntfTblMap[mclagIntfKey]
			if !ok {
				mclagIntfTblMap[mclagIntfKey] = db.Value{Field: make(map[string]string)}
			}
			mclagIntfTblMap[mclagIntfKey].Field["if_type"] = "PortChannel"
		}
	}

	res_map["MCLAG_INTERFACE"] = mclagIntfTblMap
	return res_map, err
}

var DbToYang_mclag_interface_subtree_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
	var err error
	data := (*inParams.dbDataMap)[inParams.curDb]
	mclagObj := getMclagRoot(inParams.ygRoot)
	pathInfo := NewPathInfo(inParams.uri)

	log.Info("DbToYang_mclag_interface_subtree_xfmr: ", data, inParams.ygRoot)

	if isSubtreeRequest(pathInfo.Template, "/openconfig-mclag:mclag/interfaces/interface{name}") {
		mclagIntfKeys, _ := inParams.d.GetKeys(&db.TableSpec{Name: "MCLAG_INTERFACE"})
		if len(mclagIntfKeys) > 0 {
			for _, intfKey := range mclagIntfKeys {
				ifname := intfKey.Get(1)
				if ifname == pathInfo.Var("name") && mclagObj.Interfaces != nil {
					for k, _ := range mclagObj.Interfaces.Interface {
						intfData := mclagObj.Interfaces.Interface[k]
						fillMclagIntfDetails(inParams, ifname, intfKey.Get(0), intfData)
					}
				}
			}
		}
	} else {
		var mclagIntfData map[string]map[string]string

		mclagIntfTbl := data["MCLAG_INTERFACE"]
		mclagIntfData = make(map[string]map[string]string)
		for key, _ := range mclagIntfTbl {
			//split key into domain-id and if-name
			tokens := strings.Split(key, "|")
			ifname := tokens[1]
			mclagIntfData[ifname] = make(map[string]string)
			mclagIntfData[ifname]["domainid"] = tokens[0]
			mclagIntfData[ifname]["ifname"] = ifname
		}

		for intfId := range mclagIntfData {
			if mclagObj.Interfaces == nil {
				ygot.BuildEmptyTree(mclagObj)
			}
			intfData, _ := mclagObj.Interfaces.NewInterface(intfId)
			fillMclagIntfDetails(inParams, mclagIntfData[intfId]["ifname"], mclagIntfData[intfId]["domainid"], intfData)
		}
	}

	return err
}

func fillMclagIntfDetails(inParams XfmrParams, ifname string, mclagdomainid string, intfData *ocbinds.OpenconfigMclag_Mclag_Interfaces_Interface) {
	if intfData == nil {
		return
	}

	ygot.BuildEmptyTree(intfData)

	domainid, _ := strconv.ParseUint(mclagdomainid, 10, 32)
	did32 := uint32(domainid)

	intfData.Name = &ifname

	if intfData.Config != nil {
		intfData.Config.MclagDomainId = &did32
		intfData.Config.Name = &ifname
		log.Infof("fillMclagIntfDetails--> filled config container with domain:%v and Interface:%v", did32, ifname)
	}

	// Fetch operational data from StateDb and AppDb
	stDb := inParams.dbs[db.StateDB]
	mclagRemoteIntfEntry, _ := stDb.GetEntry(&db.TableSpec{Name: "MCLAG_REMOTE_INTF_TABLE"}, db.Key{Comp: []string{mclagdomainid + "|" + ifname}})
	operStatus := mclagRemoteIntfEntry.Get("oper_status")

	appDb := inParams.dbs[db.ApplDB]
	lagEntry, _ := appDb.GetEntry(&db.TableSpec{Name: "LAG_TABLE"}, db.Key{Comp: []string{ifname}})
	trafficDisable, _ := strconv.ParseBool(lagEntry.Get("traffic_disable"))

	if intfData.State != nil {
		ygot.BuildEmptyTree(intfData.State)

		intfData.State.MclagDomainId = &did32
		intfData.State.Name = &ifname

		if intfData.State.Local != nil {
			intfData.State.Local.TrafficDisable = &trafficDisable
		}

		if intfData.State.Remote != nil {
			if operStatus == "up" {
				intfData.State.Remote.OperStatus = ocbinds.OpenconfigMclag_Mclag_Interfaces_Interface_State_Remote_OperStatus_OPER_UP
			} else if operStatus == "down" {
				intfData.State.Remote.OperStatus = ocbinds.OpenconfigMclag_Mclag_Interfaces_Interface_State_Remote_OperStatus_OPER_DOWN
			}
		}
	}

}

func getMclagRoot(s *ygot.GoStruct) *ocbinds.OpenconfigMclag_Mclag {
	deviceObj := (*s).(*ocbinds.Device)
	return deviceObj.Mclag
}
