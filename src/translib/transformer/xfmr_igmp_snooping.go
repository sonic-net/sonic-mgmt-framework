////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom, Inc.                                             //
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
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/kylelemons/godebug/pretty"
	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
	"strconv"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"
	"reflect"
)

//config db tables
var CFG_L2MC_TABLE_TS *db.TableSpec = &db.TableSpec{Name: CFG_L2MC_TABLE}
var CFG_L2MC_MROUTER_TABLE_TS *db.TableSpec = &db.TableSpec{Name: CFG_L2MC_MROUTER_TABLE}
var CFG_L2MC_STATIC_GROUP_TABLE_TS *db.TableSpec = &db.TableSpec{Name: CFG_L2MC_STATIC_GROUP_TABLE}
var CFG_L2MC_STATIC_MEMBER_TABLE_TS *db.TableSpec = &db.TableSpec{Name: CFG_L2MC_STATIC_MEMBER_TABLE}

//app db tables
var APP_L2MC_MROUTER_TABLE_TS *db.TableSpec = &db.TableSpec{Name: APP_L2MC_MROUTER_TABLE}
var APP_L2MC_MEMBER_TABLE_TS *db.TableSpec = &db.TableSpec{Name: APP_L2MC_MEMBER_TABLE}

var L2MC_TABLE_DEFAULT_FIELDS_MAP = map[string]string{
	"enabled":                    "true",
	"version":                    "2",
	"query-interval":             "125",
	"last-member-query-interval": "1000",
	"query-max-response-time":    "10",
}

const (
	CFG_L2MC_TABLE               = "CFG_L2MC_TABLE"
	CFG_L2MC_MROUTER_TABLE       = "CFG_L2MC_MROUTER_TABLE"
	CFG_L2MC_STATIC_GROUP_TABLE  = "CFG_L2MC_STATIC_GROUP_TABLE"
	CFG_L2MC_STATIC_MEMBER_TABLE = "CFG_L2MC_STATIC_MEMBER_TABLE"
	APP_L2MC_MROUTER_TABLE       = "APP_L2MC_MROUTER_TABLE"
	APP_L2MC_MEMBER_TABLE        = "APP_L2MC_MEMBER_TABLE"
)

func init() {
	XlateFuncBind("YangToDb_igmp_snooping_key_xfmr", YangToDb_igmp_snooping_key_xfmr)
	XlateFuncBind("DbToYang_igmp_snooping_key_xfmr", DbToYang_igmp_snooping_key_xfmr)
	XlateFuncBind("YangToDb_igmp_snooping_mrouter_config_key_xfmr", YangToDb_igmp_snooping_mrouter_config_key_xfmr)
	XlateFuncBind("DbToYang_igmp_snooping_mrouter_config_key_xfmr", DbToYang_igmp_snooping_mrouter_config_key_xfmr)
	XlateFuncBind("YangToDb_igmp_snooping_static_group_config_key_xfmr", YangToDb_igmp_snooping_static_group_config_key_xfmr)
	XlateFuncBind("DbToYang_igmp_snooping_static_group_config_key_xfmr", DbToYang_igmp_snooping_static_group_config_key_xfmr)
	XlateFuncBind("YangToDb_igmp_snooping_static_member_state_key_xfmr", YangToDb_igmp_snooping_static_member_state_key_xfmr)
	XlateFuncBind("DbToYang_igmp_snooping_static_member_state_key_xfmr", DbToYang_igmp_snooping_static_member_state_key_xfmr)
	XlateFuncBind("YangToDb_igmp_snooping_subtree_xfmr", YangToDb_igmp_snooping_subtree_xfmr)
	XlateFuncBind("DbToYang_igmp_snooping_subtree_xfmr", DbToYang_igmp_snooping_subtree_xfmr)
}

type reqProcessor struct {
	uri        *string
	uriPath    *gnmipb.Path
	opcode     int
	rootObj    *ocbinds.Device
	targetObj  interface{}
	db         *db.DB
	dbs        [db.MaxDB]*db.DB
	igmpsObj   *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_IgmpSnooping
	intfConfigObj    *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_IgmpSnooping_Interfaces_Interface_Config
	intfStateObj     *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_IgmpSnooping_Interfaces_Interface_State
	targetNode *yang.Entry
}

var YangToDb_igmp_snooping_static_member_state_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	fmt.Println("YangToDb_igmp_snooping_static_member_state_key_xfmr ==> ", inParams)

	path, err := getUriPath(inParams.uri)

	if err != nil {
		return "", tlerr.InvalidArgs("Invalid IGMP snooping key present in the request - %v", err)
	}

	reqP := &reqProcessor{&inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil}

	if err := reqP.setIGMPSnoopingObjFromReq(); err != nil || reqP.igmpsObj == nil {
		return "", err
	}

	var staticMemKey string

	for key, intfObj := range reqP.igmpsObj.Interfaces.Interface {
		
		fmt.Println("YangToDb_igmp_snooping_static_member_state_key_xfmr - IGMP KEY ==> ", key)
		
		for sKey, sObj := range intfObj.State.StaticMulticastGroup {
			if len (sObj.OutgoingInterface) > 0 { 
				staticMemKey = key + ":" + sKey.SourceAddr + ":" + sKey.Group + ":" + sObj.OutgoingInterface[0]
			}
			break // since only one entry should be in the request
		}
		break  // since only one entry should be in the request
	}
	
	fmt.Println("YangToDb_igmp_snooping_static_member_state_key_xfmr - staticMemKey KEY ==> ", staticMemKey)

	if len(reqP.igmpsObj.Interfaces.Interface) == 0 {
		return staticMemKey, tlerr.InvalidArgs("IGMP Snooping key is not present in the request")
	}
	
	return staticMemKey, nil	
}

var DbToYang_igmp_snooping_static_member_state_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	fmt.Println("DbToYang_igmp_snooping_static_member_state_key_xfmr ==> ", inParams)
	rmap := make(map[string]interface{})
	return rmap, nil
}

var YangToDb_igmp_snooping_static_group_config_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	fmt.Println("YangToDb_igmp_snooping_static_group_config_key_xfmr ==> ", inParams)

	path, err := getUriPath(inParams.uri)

	if err != nil {
		return "", tlerr.InvalidArgs("Invalid IGMP snooping key present in the request - %v", err)
	}

	reqP := &reqProcessor{&inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil}

	if err := reqP.setIGMPSnoopingObjFromReq(); err != nil || reqP.igmpsObj == nil {
		return "", err
	}

	var staticMemKey string

	for key, intfObj := range reqP.igmpsObj.Interfaces.Interface {
		
		fmt.Println("YangToDb_igmp_snooping_static_group_config_key_xfmr - IGMP KEY ==> ", key)
		
		if len(intfObj.Config.StaticMulticastGroup) > 0 {
			for sKey, sObj := range intfObj.Config.StaticMulticastGroup {
				if len (sObj.OutgoingInterface) > 0 {
					staticMemKey = key + "|" + sKey + "|" + sObj.OutgoingInterface[0]
				}
				break // since only one entry should be in the request
			}
		}
		
		break // since only one entry should be in the request
	}

	
	fmt.Println("YangToDb_igmp_snooping_static_group_config_key_xfmr - staticMemKey KEY ==> ", staticMemKey)

	if len(reqP.igmpsObj.Interfaces.Interface) == 0 {
		return staticMemKey, tlerr.InvalidArgs("IGMP Snooping key is not present in the request")
	}
	
	return staticMemKey, nil	
}

var DbToYang_igmp_snooping_static_group_config_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	fmt.Println("DbToYang_igmp_snooping_static_group_config_key_xfmr ==> ", inParams)
	rmap := make(map[string]interface{})
	return rmap, nil
}

var YangToDb_igmp_snooping_mrouter_config_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	fmt.Println("YangToDb_igmp_snooping_mrouter_config_key_xfmr ==> ", inParams)

	path, err := getUriPath(inParams.uri)

	if err != nil {
		return "", tlerr.InvalidArgs("Invalid IGMP snooping key present in the request - %v", err)
	}

	reqP := &reqProcessor{&inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil}

	if err := reqP.setIGMPSnoopingObjFromReq(); err != nil || reqP.igmpsObj == nil {
		return "", err
	}

	var mrouterKey string

	for key, intfObj := range reqP.igmpsObj.Interfaces.Interface {
		
		fmt.Println("YangToDb_igmp_snooping_mrouter_config_key_xfmr - IGMMP interface KEY ==> ", key)
				
		if len (intfObj.Config.MrouterInterface) > 0 {
			mrouterKey = key + "|" + intfObj.Config.MrouterInterface[0]	
		}
		
		break // since only one entry should be in the request
	}

	fmt.Println("YangToDb_igmp_snooping_mrouter_config_key_xfmr - mrouterKey KEY ==> ", mrouterKey)

	if len(reqP.igmpsObj.Interfaces.Interface) == 0 {
		return mrouterKey, tlerr.InvalidArgs("IGMP Snooping key is not present in the request")
	}
	
	return mrouterKey, nil
}

var DbToYang_igmp_snooping_mrouter_config_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	fmt.Println("DbToYang_igmp_snooping_mrouter_config_key_xfmr ==> ", inParams)
	rmap := make(map[string]interface{})
	return rmap, nil
}

var YangToDb_igmp_snooping_mrouter_state_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	fmt.Println("YangToDb_igmp_snooping_mrouter_state_key_xfmr ==> ", inParams)

	path, err := getUriPath(inParams.uri)

	if err != nil {
		return "", tlerr.InvalidArgs("Invalid IGMP snooping key present in the request - %v", err)
	}

	reqP := &reqProcessor{&inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil}

	if err := reqP.setIGMPSnoopingObjFromReq(); err != nil || reqP.igmpsObj == nil {
		return "", err
	}

	var mrouterKey string

	for key, intfObj := range reqP.igmpsObj.Interfaces.Interface {
		fmt.Println("YangToDb_igmp_snooping_mrouter_state_key_xfmr - IGMP intf KEY ==> ", key)
		if len (intfObj.State.MrouterInterface) > 0 {
			mrouterKey = key + ":" + intfObj.State.MrouterInterface[0]	
		}
		break // since only one entry should be in the request
	}

	
	fmt.Println("YangToDb_igmp_snooping_mrouter_state_key_xfmr - mrouterKey KEY ==> ", mrouterKey)

	if len(reqP.igmpsObj.Interfaces.Interface) == 0 {
		return mrouterKey, tlerr.InvalidArgs("IGMP Snooping key is not present in the request")
	}
	
	return mrouterKey, nil
}

var DbToYang_igmp_snooping_mrouter_state_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	fmt.Println("DbToYang_igmp_snooping_mrouter_state_key_xfmr ==> ", inParams)
	rmap := make(map[string]interface{})
	return rmap, nil
}

var YangToDb_igmp_snooping_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	fmt.Println("YangToDb_igmp_snooping_key_xfmr ==> ", inParams)

	path, err := getUriPath(inParams.uri)

	if err != nil {
		return "", tlerr.InvalidArgs("Invalid IGMP snooping key present in the request - %v", err)
	}

	reqP := &reqProcessor{&inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil}

	if err := reqP.setIGMPSnoopingObjFromReq(); err != nil || reqP.igmpsObj == nil {
		return "", err
	}

	var igmpsKey string

	for key, _ := range reqP.igmpsObj.Interfaces.Interface {
		igmpsKey = key
		break
	}

	fmt.Println("YangToDb_igmp_snooping_key_xfmr - IGMP KEY ==> ", igmpsKey)

	if len(reqP.igmpsObj.Interfaces.Interface) == 0 {
		return igmpsKey, tlerr.InvalidArgs("IGMP Snooping key is not present in the request")
	}

	return igmpsKey, nil
}

var DbToYang_igmp_snooping_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	fmt.Println("DbToYang_igmp_snooping_key_xfmr ==> ", inParams)
	rmap := make(map[string]interface{})
	return rmap, nil
}

func getYangNode(path *gnmipb.Path) (*yang.Entry, error) {
	pathStr, err := ygot.PathToSchemaPath(path)

	if err != nil {
		return nil, errors.New("path to schema path conversion failed")
	}

	fmt.Println("tmpStr pathStr ==> ", pathStr)

	pathStr = pathStr[1:len(pathStr)]

	fmt.Println("tmpStr pathStr ==> ", pathStr)

	ygNode := ocbinds.SchemaTree["Device"].Find(pathStr)

	fmt.Println("translate == ygNode => ", ygNode)

	return ygNode, err
}

func getUriPath(uri string) (*gnmipb.Path, error) {
	uriPath := strings.Replace(uri, "openconfig-network-instance-deviation:", "", -1)
	path, err := ygot.StringToPath(uriPath, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		return nil, errors.New("URI to path conversion failed")
	}
	for _, p := range path.Elem {
		pathSlice := strings.Split(p.Name, ":")
		p.Name = pathSlice[len(pathSlice)-1]
	}
	return path, nil
}

func (reqP *reqProcessor) setIGMPSnoopingObjFromReq() error {
	var err error

	igmpsPath := &gnmipb.Path{}

	var pathList []*gnmipb.PathElem = reqP.uriPath.Elem

	for i := 0; i < len(pathList); i++ {
		igmpsPath.Elem = append(igmpsPath.Elem, pathList[i])
		if pathList[i].Name == "igmp-snooping" {
			break
		}
	}

	fmt.Println("igmpsPath => ", igmpsPath)

	targetNodeList, err := ytypes.GetNode(ocbinds.SchemaTree["Device"], reqP.rootObj, igmpsPath)

	if err != nil {
		return tlerr.InvalidArgs("Interface list node not found in the request: %v", err)
	}

	if len(targetNodeList) == 0 {
		return tlerr.InvalidArgs("Interfaces node not found in the request: %s", *reqP.uri)
	}

	reqP.igmpsObj = targetNodeList[0].Data.(*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_IgmpSnooping)

	fmt.Println("igmpSnoopingObj ==> ", reqP.igmpsObj)

	return err
}

func (reqP *reqProcessor) handleDeleteReq() (*map[string]map[string]db.Value, error) {
	var res_map map[string]map[string]db.Value = make(map[string]map[string]db.Value)

	var igmpsTblMap map[string]db.Value = make(map[string]db.Value)
	var igmpsMrouterTblMap map[string]db.Value = make(map[string]db.Value)
	var igmpsMcastGroupTblMap map[string]db.Value = make(map[string]db.Value)
	var igmpsMcastGroupMemTblMap map[string]db.Value = make(map[string]db.Value)

	igmpsObj := reqP.igmpsObj

	if igmpsObj == nil || igmpsObj.Interfaces == nil || len(igmpsObj.Interfaces.Interface) == 0 {
		res_map[CFG_L2MC_TABLE] = igmpsTblMap
		res_map[CFG_L2MC_MROUTER_TABLE] = igmpsMrouterTblMap
		res_map[CFG_L2MC_STATIC_GROUP_TABLE] = igmpsMcastGroupTblMap
		res_map[CFG_L2MC_STATIC_MEMBER_TABLE] = igmpsMcastGroupMemTblMap
	} else {
		if len(igmpsObj.Interfaces.Interface) == 1 {
			for igmpsKey, igmpsVal := range igmpsObj.Interfaces.Interface {
				if igmpsVal.Config == nil {
					res_map[CFG_L2MC_TABLE] = igmpsTblMap
					res_map[CFG_L2MC_MROUTER_TABLE] = igmpsMrouterTblMap
					res_map[CFG_L2MC_STATIC_GROUP_TABLE] = igmpsMcastGroupTblMap
					res_map[CFG_L2MC_STATIC_MEMBER_TABLE] = igmpsMcastGroupMemTblMap
					break
				}

				dbV := db.Value{Field: make(map[string]string)}

				if reqP.targetNode.Name == "version" {
					dbV.Field["version"] = ""
					igmpsTblMap[igmpsKey] = dbV
					res_map[CFG_L2MC_TABLE] = igmpsTblMap
					fmt.Println("handleDeleteReq version res_map ==> ", res_map)
				} else if reqP.targetNode.Name == "fast-leave" {
					dbV.Field["fast-leave"] = "false"
					igmpsTblMap[igmpsKey] = dbV
					res_map[CFG_L2MC_TABLE] = igmpsTblMap
					fmt.Println("handleDeleteReq fast-leave res_map ==> ", res_map)
				} else if reqP.targetNode.Name == "querier" {
					dbV.Field["querier"] = ""
					igmpsTblMap[igmpsKey] = dbV
					res_map[CFG_L2MC_TABLE] = igmpsTblMap
					fmt.Println("handleDeleteReq querier res_map ==> ", res_map)
				} else if reqP.targetNode.Name == "query-interval" {
					dbV.Field["query-interval"] = ""
					igmpsTblMap[igmpsKey] = dbV
					res_map[CFG_L2MC_TABLE] = igmpsTblMap
					fmt.Println("handleDeleteReq query-interval res_map ==> ", res_map)
				} else if reqP.targetNode.Name == "query-max-response-time" {
					dbV.Field["query-max-response-time"] = ""
					igmpsTblMap[igmpsKey] = dbV
					res_map[CFG_L2MC_TABLE] = igmpsTblMap
					fmt.Println("handleDeleteReq query-max-response-time res_map ==> ", res_map)
				} else if reqP.targetNode.Name == "last-member-query-interval" {
					dbV.Field["last-member-query-interval"] = ""
					igmpsTblMap[igmpsKey] = dbV
					res_map[CFG_L2MC_TABLE] = igmpsTblMap
					fmt.Println("handleDeleteReq last-member-query-interval res_map ==> ", res_map)
				} else if reqP.targetNode.Name == "enabled" {
					dbV.Field["enabled"] = ""
					igmpsTblMap[igmpsKey] = dbV
					res_map[CFG_L2MC_TABLE] = igmpsTblMap
					fmt.Println("handleDeleteReq enabled res_map ==> ", res_map)
				} else if len(igmpsVal.Config.MrouterInterface) == 0 && reqP.isConfigTargetNode ("mrouter-interface") == true {
					res_map[CFG_L2MC_MROUTER_TABLE] = igmpsMrouterTblMap
				} else if len(igmpsVal.Config.MrouterInterface) == 1 {
					for _, mrVal := range igmpsVal.Config.MrouterInterface {
						igmpsMrouterKey := igmpsKey + "|" + mrVal
						igmpsMrouterTblMap[igmpsMrouterKey] = db.Value{Field: make(map[string]string)}
					}
					res_map[CFG_L2MC_MROUTER_TABLE] = igmpsMrouterTblMap
				} else if len(igmpsVal.Config.StaticMulticastGroup) == 0 && reqP.isConfigTargetNode ("static-multicast-group") == true {
					res_map[CFG_L2MC_STATIC_GROUP_TABLE] = igmpsMcastGroupTblMap
					res_map[CFG_L2MC_STATIC_MEMBER_TABLE] = igmpsMcastGroupMemTblMap
				} else if len(igmpsVal.Config.StaticMulticastGroup) == 1 {
					for grpKey, grpObj := range igmpsVal.Config.StaticMulticastGroup {
						if len (grpObj.OutgoingInterface) == 0 {
							var err error
							var staticGrpDbTbl db.Table
							if staticGrpDbTbl, err = reqP.db.GetTable(CFG_L2MC_STATIC_GROUP_TABLE_TS); err != nil {
								fmt.Println("DB error in test GetEntry => ", err)
								return nil, err;
							}
							fmt.Println("handleDeleteReq - printing db staticGrDbTbl data")
							pretty.Print(staticGrpDbTbl)
					
							staticGrpKeys, _ := staticGrpDbTbl.GetKeys()
							fmt.Println("handleDeleteReq - printing db table staticGrpKeys keys")
							pretty.Print(staticGrpKeys)
							// fetch all group entries from the db and delete the entries matches with the given grpKey							
							for k, _ := range staticGrpKeys {
								if staticGrpKeys[k].Comp[1] == grpKey {
									igmpsGrpKey := igmpsKey + "|" + grpKey
									igmpsMcastGroupTblMap[igmpsGrpKey] = db.Value{Field: make(map[string]string)}
									
									staticGrpDbV, err := staticGrpDbTbl.GetEntry(staticGrpKeys[k])
									if err != nil {
										return nil, err
									}
									outIntfs := staticGrpDbV.GetList("out-intf")									
									for _, intf := range outIntfs {
										igmpsGrpMemKey := igmpsKey + "|" + grpKey + "|" + intf
										igmpsMcastGroupMemTblMap[igmpsGrpMemKey] = db.Value{Field: make(map[string]string)}
									}
									break
								}
							}
						} else {
							dbV := db.Value{Field: make(map[string]string)}
							dbV.SetList("out-intf", grpObj.OutgoingInterface)
							igmpsGrpKey := igmpsKey + "|" + grpKey
							igmpsMcastGroupTblMap[igmpsGrpKey] = dbV 
							for _, outIntf := range grpObj.OutgoingInterface {
								igmpsGrpMemKey := igmpsKey + "|" + grpKey + "|" + outIntf
								igmpsMcastGroupMemTblMap[igmpsGrpMemKey] = db.Value{Field: make(map[string]string)}
							}
						}
					}
					if len(igmpsMcastGroupTblMap) > 0 {
						res_map[CFG_L2MC_STATIC_GROUP_TABLE] = igmpsMcastGroupTblMap	
					}
					if len(igmpsMcastGroupMemTblMap) > 0 {
						res_map[CFG_L2MC_STATIC_MEMBER_TABLE] = igmpsMcastGroupMemTblMap	
					}
				}
			}
		}
	}

	fmt.Println(" handleDeleteReq ============> res_map")
	pretty.Print(res_map)

	return &res_map, nil
}

// handle create/replace/update request
func (reqP *reqProcessor) handleCRUReq() (*map[string]map[string]db.Value, error) {

	fmt.Println(" handleCRUReq entering ============> ")

	var res_map map[string]map[string]db.Value = make(map[string]map[string]db.Value)
	var igmpsTblMap map[string]db.Value = make(map[string]db.Value)
	var igmpsMrouterTblMap map[string]db.Value = make(map[string]db.Value)
	var igmpsMcastGroupTblMap map[string]db.Value = make(map[string]db.Value)
	var igmpsMcastGroupMemTblMap map[string]db.Value = make(map[string]db.Value)

	igmpsObj := reqP.igmpsObj

	for igmpsKey, igmpsVal := range igmpsObj.Interfaces.Interface {

		if igmpsVal.Config == nil {
			fmt.Println(" handleCRUReq ============> igmpsVal.Config is NULL")
			continue
		}

		dbV := db.Value{Field: make(map[string]string)}

		if igmpsVal.Config.Version != nil {
			dbV.Field["version"] = strconv.Itoa(int(*igmpsVal.Config.Version))
			fmt.Println(" handleCRUReq ============> setting version => ", strconv.Itoa(int(*igmpsVal.Config.Version)))
		}

		if igmpsVal.Config.FastLeave != nil {
			dbV.Field["fast-leave"] = strconv.FormatBool(*igmpsVal.Config.FastLeave)
			fmt.Println(" handleCRUReq ============> setting fast-leave => ", strconv.FormatBool(*igmpsVal.Config.FastLeave))
		}

		if igmpsVal.Config.QueryInterval != nil {
			dbV.Field["query-interval"] = strconv.Itoa(int(*igmpsVal.Config.QueryInterval))
			fmt.Println(" handleCRUReq ============> setting query-interval => ", strconv.Itoa(int(*igmpsVal.Config.QueryInterval)))
		}

		if igmpsVal.Config.QueryMaxResponseTime != nil {
			dbV.Field["query-max-response-time"] = strconv.Itoa(int(*igmpsVal.Config.QueryMaxResponseTime))
			fmt.Println(" handleCRUReq ============> setting query-max-response-time => ", strconv.Itoa(int(*igmpsVal.Config.QueryMaxResponseTime)))
		}

		if igmpsVal.Config.LastMemberQueryInterval != nil {
			dbV.Field["last-member-query-interval"] = strconv.Itoa(int(*igmpsVal.Config.LastMemberQueryInterval))
			fmt.Println(" handleCRUReq ============> setting last-member-query-interval => ", strconv.Itoa(int(*igmpsVal.Config.LastMemberQueryInterval)))
		}

		if igmpsVal.Config.Querier != nil {
			dbV.Field["querier"] = strconv.FormatBool(*igmpsVal.Config.Querier)
			fmt.Println(" handleCRUReq ============> setting querier => ", strconv.FormatBool(*igmpsVal.Config.Querier))
		}

		if igmpsVal.Config.Enabled != nil {
			dbV.Field["enabled"] = strconv.FormatBool(*igmpsVal.Config.Enabled)
			fmt.Println(" handleCRUReq ============> setting querier => ", strconv.FormatBool(*igmpsVal.Config.Enabled))
		}

		if len(dbV.Field) > 0 {
			igmpsTblMap[igmpsKey] = dbV
			res_map[CFG_L2MC_TABLE] = igmpsTblMap
		}
		
		if len(igmpsVal.Config.MrouterInterface) > 0 {

			fmt.Println(" handleCRUReq ============> setting igmpsVal.Config.MrouterInterface")

			for _, mrVal := range igmpsVal.Config.MrouterInterface {
				igmpsMrouterKey := igmpsKey + "|" + mrVal
				dbV := db.Value{Field: make(map[string]string)}
				dbV.Field["NULL"] = "NULL" // to represent empty value
				igmpsMrouterTblMap[igmpsMrouterKey] = dbV
				fmt.Println(" handleCRUReq ============> setting igmpsMrouterKey => ", igmpsMrouterKey)
			}

			if len(igmpsMrouterTblMap) > 0 {
//				if len(dbV.Field) == 0 {
//					dbV.Field["enabled"] = "true"
//					igmpsTblMap[igmpsKey] = dbV
//					res_map[CFG_L2MC_TABLE] = igmpsTblMap
//				}
				fmt.Println(" handleCRUReq ============> setting CFG_L2MC_MROUTER_TABLE igmpsMrouterTblMap => ", igmpsMrouterTblMap)
				res_map[CFG_L2MC_MROUTER_TABLE] = igmpsMrouterTblMap
			}
		}

		if len(igmpsVal.Config.StaticMulticastGroup) > 0 {

			fmt.Println(" handleCRUReq ============> setting igmpsVal.Config.StaticMulticastGroup")

			for grpKey, grpObj := range igmpsVal.Config.StaticMulticastGroup {
				if len (grpObj.OutgoingInterface) > 0 {
					igmpsGrpKey := igmpsKey + "|" + grpKey
					dbV := db.Value{Field: make(map[string]string)}
					dbV.Field["NULL"] = "NULL" // since deleting the field "out-intf" from the db removes the key also, to avoid that insert the dummy field/value as NULL/NULL
					dbV.SetList("out-intf", grpObj.OutgoingInterface)
					igmpsMcastGroupTblMap[igmpsGrpKey] = dbV
					for _, outIntf := range grpObj.OutgoingInterface {
						igmpsGrpMemKey := igmpsKey + "|" + grpKey + "|" + outIntf
						dbV := db.Value{Field: make(map[string]string)}
						dbV.Field["NULL"] = "NULL" // to represent empty value
						igmpsMcastGroupMemTblMap[igmpsGrpMemKey] = dbV
						fmt.Println(" handleCRUReq ============> setting igmpsVal.Config.StaticMulticastGroup igmpsGrpMemKey => ", igmpsGrpMemKey)
					}
				} else {
					igmpsGrpKey := igmpsKey + "|" + grpKey
					dbV := db.Value{Field: make(map[string]string)}
					dbV.Field["NULL"] = "NULL" // to represent empty value
					igmpsMcastGroupTblMap[igmpsGrpKey] = dbV
					fmt.Println(" handleCRUReq ============> setting igmpsVal.Config.StaticMulticastGroup igmpsGrpKey => ", igmpsGrpKey)
				}
			}

			if len(igmpsMcastGroupTblMap) > 0 {
//				if len(dbV.Field) == 0 {
//					dbV.Field["enabled"] = "true"
//					igmpsTblMap[igmpsKey] = dbV
//					res_map[CFG_L2MC_TABLE] = igmpsTblMap
//				}
				fmt.Println(" handleCRUReq ============> setting CFG_L2MC_STATIC_MEMBER_TABLE igmpsMcastGroupTblMap => ", igmpsMcastGroupTblMap)
				res_map[CFG_L2MC_STATIC_GROUP_TABLE] = igmpsMcastGroupTblMap
			}
			
			if len(igmpsMcastGroupMemTblMap) > 0 {
				fmt.Println(" handleCRUReq ============> setting CFG_L2MC_STATIC_MEMBER_TABLE igmpsMcastGroupMemTblMap => ", igmpsMcastGroupMemTblMap)
				res_map[CFG_L2MC_STATIC_MEMBER_TABLE] = igmpsMcastGroupMemTblMap				
			}
		}
	}

	fmt.Println(" handleCRUReq ============> printing  res_map ")
	pretty.Print(res_map)

	return &res_map, nil
}

func (reqP *reqProcessor) translateToDb() (*map[string]map[string]db.Value, error) {
	//DELETE
	if reqP.opcode == 5 {
		// get the target node
		var err error
		if reqP.targetNode, err = getYangNode(reqP.uriPath); err != nil {
			return nil, tlerr.InvalidArgs("Invalid request: %s", *reqP.uri)
		}

		fmt.Println("translateToDb param reqP.targetNode.Name ==> ", reqP.targetNode.Name)
		
		res_map, err := reqP.handleDeleteReq()
		
		if err != nil {
			return nil, tlerr.InvalidArgs("Invlaid IGMP Snooing delete: %s", *reqP.uri)
		}
		
		return res_map, err
		
	} else if reqP.igmpsObj != nil && reqP.igmpsObj.Interfaces != nil {
		res_map, err := reqP.handleCRUReq()
		if err != nil {
			return nil, tlerr.InvalidArgs("Invlaid IGMP Snooing request: %s", *reqP.uri)
		}
		return res_map, err
	} else {
		return nil, tlerr.InvalidArgs("IGMP Snooing object not found in the request: %s", *reqP.uri)
	}
}

var YangToDb_igmp_snooping_subtree_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {

	fmt.Println("YangToDb_igmp_snooping_subtree_xfmr entering => ", inParams)

	path, err := getUriPath(inParams.uri)

	if err != nil {
		return nil, err
	}

	reqP := &reqProcessor{&inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil}

	fmt.Println("YangToDb_igmp_snooping_subtree_xfmr => translateToDb == reqP.uri => ", *reqP.uri)

	if err := reqP.setIGMPSnoopingObjFromReq(); err != nil {
		return nil, err
	}

	fmt.Println("YangToDb_igmp_snooping_subtree_xfmr ==> printing IGMPSnooping object request ==> ")
	pretty.Print(*reqP.igmpsObj)

	res_map, err := reqP.translateToDb()

	if err == nil {
		return *res_map, nil
	} else {
		return nil, err
	}
}

var DbToYang_igmp_snooping_subtree_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
	fmt.Println("DbToYang_igmp_snooping_subtree_xfmr entering => ", inParams)

	path, err := getUriPath(inParams.uri)

	if err != nil {
		return err
	}

	reqP := &reqProcessor{&inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil}

	fmt.Println("YangToDb_igmp_snooping_subtree_xfmr => translateToDb == reqP.uri => ", *reqP.uri)

	if err := reqP.setIGMPSnoopingObjFromReq(); err != nil {
		return err
	}

	if reqP.igmpsObj != nil {
		fmt.Println("YangToDb_igmp_snooping_subtree_xfmr ==> printing IGMPSnooping object request ==> ")
		pretty.Print(*reqP.igmpsObj)
	}

	// get the target node
	reqP.targetNode, err = getYangNode(reqP.uriPath)
	if err != nil {
		return tlerr.InvalidArgs("Invalid request - error: %v", err)
	}

	return reqP.translateToYgotObj()
}

func (reqP *reqProcessor) unMarshalStaticGrpConfigObj() (error) {
	if len(reqP.intfConfigObj.StaticMulticastGroup) > 0 {
		for grpKey, staticGrpObj := range reqP.intfConfigObj.StaticMulticastGroup {
			fmt.Println("unMarshalStaticGrpConfigObj - grpey => ", grpKey)
			fmt.Println("unMarshalStaticGrpConfig - grpObj => ", staticGrpObj)
			var err error
			var staticGrpDbTbl db.Table
			if staticGrpDbTbl, err = reqP.db.GetTable(CFG_L2MC_STATIC_MEMBER_TABLE_TS); err != nil {
				fmt.Println("DB error in GetEntry => ", err)
			}
			
			fmt.Println("unMarshalStaticGrpConfigObj - printing db staticGrDbTbl data")
			pretty.Print(staticGrpDbTbl)
	
			staticGrpKeys, _ := staticGrpDbTbl.GetKeys()
			fmt.Println("unMarshalStaticGrpConfigObj - printing db table staticGrpKeys keys")
			pretty.Print(staticGrpKeys)
			
			for k, _ := range staticGrpKeys {
				staticGrpDbV, err := staticGrpDbTbl.GetEntry(staticGrpKeys[k])
				if err != nil {
					return err
				}
				fmt.Println("unMarshalStaticGrpConfigObj - printing db table staticGrpKeysEntry")
				pretty.Print(staticGrpDbV)
	
				if *reqP.intfConfigObj.Name != staticGrpKeys[k].Comp[0] || grpKey != staticGrpKeys[k].Comp[1] {
					continue
				}
				
				if reqP.targetNode.Name == "group" {
					staticGrpObj.Group = &grpKey
				} else if reqP.targetNode.Name == "outgoing-interface" {
					if len(staticGrpObj.OutgoingInterface) == 0 {
						staticGrpObj.OutgoingInterface = append(staticGrpObj.OutgoingInterface, staticGrpKeys[k].Comp[2])	
					} else if staticGrpKeys[k].Comp[2] == staticGrpObj.OutgoingInterface[0] {
						staticGrpObj.OutgoingInterface = append(staticGrpObj.OutgoingInterface, staticGrpKeys[k].Comp[2])						
					}					
				} else {
					staticGrpObj.Group = &grpKey
					staticGrpObj.OutgoingInterface = append(staticGrpObj.OutgoingInterface, staticGrpKeys[k].Comp[2])
				}
			}			
		}
	} else if reqP.isConfigTargetNode ("static-multicast-group") {
		var staticGrpDbTbl db.Table
		var err error
		if staticGrpDbTbl, err = reqP.db.GetTable(CFG_L2MC_STATIC_MEMBER_TABLE_TS); err != nil {
			fmt.Println("DB error in GetEntry => ", err)
		}

		fmt.Println("unMarshalStaticGrpConfigObj - printing db staticGrDbTbl data")
		pretty.Print(staticGrpDbTbl)

		staticGrpKeys, _ := staticGrpDbTbl.GetKeys()
		fmt.Println("unMarshalStaticGrpConfigObj - printing db table staticGrpKeys keys")
		pretty.Print(staticGrpKeys)

		for k, _ := range staticGrpKeys {
			staticGrpDbV, err := staticGrpDbTbl.GetEntry(staticGrpKeys[k])
			if err != nil {
				return err
			}
			fmt.Println("unMarshalStaticGrpConfigObj - printing db table staticGrpKeysEntry")
			pretty.Print(staticGrpDbV)

			if *reqP.intfConfigObj.Name != staticGrpKeys[k].Comp[0] {
				continue
			}

			staticGrpObj, _ := reqP.intfConfigObj.StaticMulticastGroup[staticGrpKeys[k].Comp[1]]

			if staticGrpObj == nil {
				staticGrpObj, err = reqP.intfConfigObj.NewStaticMulticastGroup(staticGrpKeys[k].Comp[1])
				if err != nil {
					return err
				}
			}

			staticGrpObj.OutgoingInterface = append(staticGrpObj.OutgoingInterface, staticGrpKeys[k].Comp[2])

			fmt.Println("unMarshalStaticGrpConfigObj - printing staticGrpObj => ", *staticGrpObj)
		}
	}
	
	return nil
}

func (reqP *reqProcessor) unMarshalStaticGrpStateObj() (error) {
	if len(reqP.intfStateObj.StaticMulticastGroup) > 0 {
		for grpKey, staticGrpObj := range reqP.intfStateObj.StaticMulticastGroup {
			fmt.Println("unMarshalStaticGrpStateObj - grpKey => ", grpKey.Group)
			fmt.Println("unMarshalStaticGrpStateObj - grpKey => ", grpKey.SourceAddr)
			fmt.Println("unMarshalStaticGrpStateObj - grpObj => ", staticGrpObj)
			intfKeys := reflect.ValueOf(reqP.igmpsObj.Interfaces.Interface).MapKeys()
			// temp for now - hardcoded
			portIntfName := "Ethernet10"
			_, err := reqP.dbs[0].GetEntry(APP_L2MC_MEMBER_TABLE_TS, db.Key{[]string{intfKeys[0].Interface().(string), grpKey.Group, grpKey.SourceAddr, portIntfName}})
			if err != nil {
				return err
			}
			if reqP.targetNode.Name == "group" {
				staticGrpObj.Group = &grpKey.Group
			} else if reqP.targetNode.Name == "source-addr" {
				staticGrpObj.SourceAddr = &grpKey.SourceAddr								
			} else if reqP.targetNode.Name == "outgoing-interface" {
				staticGrpObj.OutgoingInterface = append(staticGrpObj.OutgoingInterface, portIntfName)				
			} else {
				staticGrpObj.Group = &grpKey.Group
				staticGrpObj.SourceAddr = &grpKey.SourceAddr
				staticGrpObj.OutgoingInterface = append(staticGrpObj.OutgoingInterface, portIntfName)
			}
		}
	} else if reqP.isStateTargetNode ("static-multicast-group") {
		var staticGrpDbTbl db.Table
		var err error
		fmt.Println("printing db type")
		pretty.Print(reqP.dbs[0])
		fmt.Println("printing ALL db type")
		pretty.Print(reqP.dbs)
		fmt.Println("printing trans db type")
		pretty.Print(reqP.db)
		if staticGrpDbTbl, err = reqP.dbs[0].GetTable(APP_L2MC_MEMBER_TABLE_TS); err != nil {
			fmt.Println("DB error in GetEntry => ", err)
		}

		fmt.Println("unMarshalStaticGrpStateObj - printing db staticGrpDbTbl data")
		pretty.Print(staticGrpDbTbl)

		staticGrpKeys, _ := staticGrpDbTbl.GetKeys()
		fmt.Println("unMarshalStaticGrpStateObj - printing db table staticGrpKeys keys")
		pretty.Print(staticGrpKeys)

		for k, _ := range staticGrpKeys {
			staticGrpDbV, err := staticGrpDbTbl.GetEntry(staticGrpKeys[k])
			if err != nil {
				return err
			}
			fmt.Println("unMarshalStaticGrpStateObj - printing db table staticGrpKeysEntry")
			pretty.Print(staticGrpDbV)

			if *reqP.intfStateObj.Name != staticGrpKeys[k].Comp[0] {
				continue
			}

			staticGrpKey := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_IgmpSnooping_Interfaces_Interface_State_StaticMulticastGroup_Key{staticGrpKeys[k].Comp[1], staticGrpKeys[k].Comp[2]}
			staticGrpObj, _ := reqP.intfStateObj.StaticMulticastGroup[staticGrpKey]

			if staticGrpObj == nil {
				staticGrpObj, err = reqP.intfStateObj.NewStaticMulticastGroup(staticGrpKeys[k].Comp[1], staticGrpKeys[k].Comp[2])
				if err != nil {
					return err
				}
			}

			staticGrpObj.OutgoingInterface = append(staticGrpObj.OutgoingInterface, staticGrpKeys[k].Comp[3])

			fmt.Println("unMarshalStaticGrpStateObj - printing staticGrpObj => ", *staticGrpObj)
		}
	}
	
	return nil
}

func (reqP *reqProcessor) unMarshalMrouterState() (error) {
	if reqP.isStateTargetNode ("mrouter-interface") == false {
		return nil
	}

	var mRouterDbTbl db.Table
	var err error
	if mRouterDbTbl, err = reqP.dbs[0].GetTable(APP_L2MC_MROUTER_TABLE_TS); err != nil {
		fmt.Println("DB error in GetEntry => ", err)
	}

	fmt.Println("unMarshalMrouterConfig - printing db mRouterDbTbl data")
	pretty.Print(mRouterDbTbl)

	mrouterKeys, _ := mRouterDbTbl.GetKeys()
	fmt.Println("unMarshalMrouterConfig - printing db table mRouterDbTbl keys")
	pretty.Print(mrouterKeys)

	for j, _ := range mrouterKeys {
		mrouterDbV, err := mRouterDbTbl.GetEntry(mrouterKeys[j])
		if err != nil {
			return err
		}
		fmt.Println("unMarshalMrouterConfig - printing db table mrouterDbVEntry")
		pretty.Print(mrouterDbV)

		if *reqP.intfStateObj.Name != mrouterKeys[j].Comp[0] {
			continue
		}

		reqP.intfStateObj.MrouterInterface = append(reqP.intfStateObj.MrouterInterface, mrouterKeys[j].Comp[1])
	}
	
	return nil
}

func (reqP *reqProcessor) unMarshalMrouterConfig() (error) {
	if reqP.isConfigTargetNode ("mrouter-interface") == false {
		return nil
	}

	var mRouterDbTbl db.Table
	var err error
	if mRouterDbTbl, err = reqP.db.GetTable(CFG_L2MC_MROUTER_TABLE_TS); err != nil {
		fmt.Println("DB error in GetEntry => ", err)
	}

	fmt.Println("unMarshalMrouterConfig - printing db mRouterDbTbl data")
	pretty.Print(mRouterDbTbl)

	mrouterKeys, _ := mRouterDbTbl.GetKeys()
	fmt.Println("unMarshalMrouterConfig - printing db table mRouterDbTbl keys")
	pretty.Print(mrouterKeys)

	for j, _ := range mrouterKeys {
		mrouterDbV, err := mRouterDbTbl.GetEntry(mrouterKeys[j])
		if err != nil {
			return err
		}
		fmt.Println("unMarshalMrouterConfig - printing db table mrouterDbVEntry")
		pretty.Print(mrouterDbV)

		if *reqP.intfConfigObj.Name != mrouterKeys[j].Comp[0] {
			continue
		}

		reqP.intfConfigObj.MrouterInterface = append(reqP.intfConfigObj.MrouterInterface, mrouterKeys[j].Comp[1])
	}
	
	return nil
}

func (reqP *reqProcessor) unMarshalIGMPSnoopingIntfConfigObjInst (dbV *db.Value) {
	isAllFields := reqP.isConfigTargetNode ("")
	if reqP.targetNode.Name == "version" || isAllFields == true {
		if fv, ok := dbV.Field["version"]; ok == true {
			intV, _ := strconv.ParseInt(fv, 10, 64)
			tmp := uint8(intV)
			reqP.intfConfigObj.Version = &tmp
		}
	}
	if reqP.targetNode.Name == "fast-leave" || isAllFields == true {
		if fv, ok := dbV.Field["fast-leave"]; ok == true {
			tmp, _ := strconv.ParseBool(fv)
			reqP.intfConfigObj.FastLeave = &tmp
		}
	}
	if reqP.targetNode.Name == "query-interval" || isAllFields == true {
		if fv, ok := dbV.Field["query-interval"]; ok == true {
			intV, _ := strconv.ParseInt(fv, 10, 64)
			tmp := uint16(intV)
			reqP.intfConfigObj.QueryInterval = &tmp
		}
	} 
	if reqP.targetNode.Name == "last-member-query-interval" || isAllFields == true {
		if fv, ok := dbV.Field["last-member-query-interval"]; ok == true {
			intV, _ := strconv.ParseInt(fv, 10, 64)
			tmp := uint32(intV)
			reqP.intfConfigObj.LastMemberQueryInterval = &tmp
		}
	} 
	if reqP.targetNode.Name == "query-max-response-time" || isAllFields == true {
		if fv, ok := dbV.Field["query-max-response-time"]; ok == true {
			intV, _ := strconv.ParseInt(fv, 10, 64)
			tmp := uint16(intV)
			reqP.intfConfigObj.QueryMaxResponseTime = &tmp
		}
	} 
	if reqP.targetNode.Name == "enabled" || isAllFields == true {
		if fv, ok := dbV.Field["enabled"]; ok == true {
			tmp, _ := strconv.ParseBool(fv)
			reqP.intfConfigObj.Enabled = &tmp
		}
	} 
	if reqP.targetNode.Name == "querier" || isAllFields == true {
		if fv, ok := dbV.Field["querier"]; ok == true {
			tmp, _ := strconv.ParseBool(fv)
			reqP.intfConfigObj.Querier = &tmp
		}
	}
}

func (reqP *reqProcessor) unMarshalIGMPSnoopingIntfStateObjInst (dbV *db.Value) {
	isAllFields := reqP.isStateTargetNode ("")
	if reqP.targetNode.Name == "version" || isAllFields == true {
		if fv, ok := dbV.Field["version"]; ok == true {
			intV, _ := strconv.ParseInt(fv, 10, 64)
			tmp := uint8(intV)
			reqP.intfStateObj.Version = &tmp 
		}
	}
	if reqP.targetNode.Name == "fast-leave" || isAllFields == true {
		if fv, ok := dbV.Field["fast-leave"]; ok == true {
			tmp, _ := strconv.ParseBool(fv)
			reqP.intfStateObj.FastLeave = &tmp
		}
	}
	if reqP.targetNode.Name == "query-interval" || isAllFields == true {
		if fv, ok := dbV.Field["query-interval"]; ok == true {
			intV, _ := strconv.ParseInt(fv, 10, 64)
			tmp := uint16(intV)
			reqP.intfStateObj.QueryInterval = &tmp
		}
	} 
	if reqP.targetNode.Name == "last-member-query-interval" || isAllFields == true {
		if fv, ok := dbV.Field["last-member-query-interval"]; ok == true {
			intV, _ := strconv.ParseInt(fv, 10, 64)
			tmp := uint32(intV)
			reqP.intfStateObj.LastMemberQueryInterval = &tmp
		}
	} 
	if reqP.targetNode.Name == "query-max-response-time" || isAllFields == true {
		if fv, ok := dbV.Field["query-max-response-time"]; ok == true {
			intV, _ := strconv.ParseInt(fv, 10, 64)
			tmp := uint16(intV)
			reqP.intfStateObj.QueryMaxResponseTime = &tmp
		}
	} 
	if reqP.targetNode.Name == "enabled" || isAllFields == true {
		if fv, ok := dbV.Field["enabled"]; ok == true {
			tmp, _ := strconv.ParseBool(fv)
			reqP.intfStateObj.Enabled = &tmp
		}
	} 
	if reqP.targetNode.Name == "querier" || isAllFields == true {
		if fv, ok := dbV.Field["querier"]; ok == true {
			tmp, _ := strconv.ParseBool(fv)
			reqP.intfStateObj.Querier = &tmp
		}
	}
}

func (reqP *reqProcessor) unMarshalIGMPSnoopingIntfState() (error) {
	var l2McDbTbl db.Table
	var dbErr error
	if l2McDbTbl, dbErr = reqP.db.GetTable(CFG_L2MC_TABLE_TS); dbErr != nil {
		fmt.Println("DB error in GetEntry ====> ", dbErr)
	}

	fmt.Println("translateToYgotObj - printing db data")
	pretty.Print(l2McDbTbl)

	l2McKeys, _ := l2McDbTbl.GetKeys()
	fmt.Println("translateToYgotObj - printing db test ttest table keys")
	pretty.Print(l2McKeys)

	for i, _ := range l2McKeys {
		dbV, err := l2McDbTbl.GetEntry(l2McKeys[i])
		if err != nil {
			return err
		}
		intfName := l2McKeys[i].Comp[0]
		intfObj, err := reqP.igmpsObj.Interfaces.NewInterface(intfName)
		if err != nil {
			return err
		}
		ygot.BuildEmptyTree(intfObj)
		
		intfObj.Config.Name = intfObj.Name
		reqP.intfConfigObj = intfObj.Config
		
		reqP.unMarshalIGMPSnoopingIntfConfigObjInst(&dbV)
		
		if err := reqP.unMarshalMrouterState(); err != nil {
			return err
		}
		if err := reqP.unMarshalStaticGrpStateObj(); err != nil {
			return err
		}
	}
	
	return nil
}

func (reqP *reqProcessor) unMarshalIGMPSnoopingIntf(objType int) (error) {
	var l2McDbTbl db.Table
	var dbErr error
	if l2McDbTbl, dbErr = reqP.db.GetTable(CFG_L2MC_TABLE_TS); dbErr != nil {
		fmt.Println("DB error in GetEntry => ", dbErr)
	}

	fmt.Println("translateToYgotObj - printing db data")
	pretty.Print(l2McDbTbl)

	l2McKeys, _ := l2McDbTbl.GetKeys()
	fmt.Println("translateToYgotObj - printing db table keys")
	pretty.Print(l2McKeys)

	for i, _ := range l2McKeys {
		dbV, err := l2McDbTbl.GetEntry(l2McKeys[i])
		if err != nil {
			return err
		}
		intfName := l2McKeys[i].Comp[0]
		intfObj, err := reqP.igmpsObj.Interfaces.NewInterface(intfName)
		if err != nil {
			return err
		}
		ygot.BuildEmptyTree(intfObj)
		
		if objType == 1 {
			intfObj.Config.Name = intfObj.Name
			reqP.intfConfigObj = intfObj.Config
			
			reqP.unMarshalIGMPSnoopingIntfConfigObjInst(&dbV)
			
			if err := reqP.unMarshalMrouterConfig(); err != nil {
				return err
			}
			if err := reqP.unMarshalStaticGrpConfigObj(); err != nil {
				return err
			}
		} else if objType == 2 {
			//state
			intfObj.State.Name = intfObj.Name
			reqP.intfStateObj = intfObj.State
			
			reqP.unMarshalIGMPSnoopingIntfStateObjInst(&dbV)
			
			if err := reqP.unMarshalMrouterState(); err != nil {
				return err
			}
			
			if err := reqP.unMarshalStaticGrpStateObj(); err != nil {
				return err
			}
		} else if objType == 3 {
			//config
			intfObj.Config.Name = intfObj.Name
			reqP.intfConfigObj = intfObj.Config
			
			reqP.unMarshalIGMPSnoopingIntfConfigObjInst(&dbV)
			
			if err := reqP.unMarshalMrouterConfig(); err != nil {
				return err
			}
			if err := reqP.unMarshalStaticGrpConfigObj(); err != nil {
				return err
			}
			//state
			intfObj.State.Name = intfObj.Name
			reqP.intfStateObj = intfObj.State
			
			reqP.unMarshalIGMPSnoopingIntfStateObjInst(&dbV)
			
			if err := reqP.unMarshalMrouterState(); err != nil {
				return err
			}
			
			if err := reqP.unMarshalStaticGrpStateObj(); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (reqP *reqProcessor) isConfigTargetNode (nodeName string) bool {
	if reqP.targetNode.Name == "igmp-snooping" || reqP.targetNode.Name == "interfaces" || reqP.targetNode.Name == "interface" || reqP.targetNode.Name == "config" || nodeName == reqP.targetNode.Name {
		return true
	}
	return false  
}

func (reqP *reqProcessor) isStateTargetNode (nodeName string) bool {
	if reqP.targetNode.Name == "igmp-snooping" || reqP.targetNode.Name == "interfaces" || reqP.targetNode.Name == "interface" || reqP.targetNode.Name == "state" || nodeName == reqP.targetNode.Name {
		return true
	}
	return false  
}

func (reqP *reqProcessor) translateToYgotObj() error {
	log.Info("translateToYgotObj entering => ")

	var err error
	
	fmt.Println("translateToYgotObj param reqP.targetNode.Name ==> ", reqP.targetNode.Name)

	if reqP.igmpsObj == nil {
		reqP.igmpsObj = &(ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_IgmpSnooping{nil})
	}
	
	if reqP.targetNode.Name == "igmp-snooping" || reqP.targetNode.Name == "interfaces" || len(reqP.igmpsObj.Interfaces.Interface) == 0 {
		ygot.BuildEmptyTree(reqP.igmpsObj)
		reqP.unMarshalIGMPSnoopingIntf(3)
	} else if len(reqP.igmpsObj.Interfaces.Interface) == 1 {
		intfKeys := reflect.ValueOf(reqP.igmpsObj.Interfaces.Interface).MapKeys()
		intfObj := reqP.igmpsObj.Interfaces.Interface[intfKeys[0].Interface().(string)]

		var objType int
		if intfObj.Config != nil {
			objType = 1
		} else if intfObj.State != nil {
			objType = 2
		} else {
			objType = 3
		}
		
		ygot.BuildEmptyTree(intfObj)

		if objType == 1 || objType == 3  {
			intfObj.Config.Name = intfObj.Name
			reqP.intfConfigObj = intfObj.Config
	
			dbV, err := reqP.db.GetEntry(CFG_L2MC_TABLE_TS, db.Key{[]string{intfKeys[0].Interface().(string)}})
			if err != nil {
				return err
			}
			reqP.unMarshalIGMPSnoopingIntfConfigObjInst(&dbV)
	
			if err = reqP.unMarshalMrouterConfig(); err != nil {
				return err
			}
			
			if err = reqP.unMarshalStaticGrpConfigObj(); err != nil {
				return err
			}
		}
		
		if objType == 2 || objType == 3  {
			// state obj
			//state
			intfObj.State.Name = intfObj.Name
			reqP.intfStateObj = intfObj.State
			
			fmt.Println("test ---> 1 => ", reqP.dbs[4])
			
			dbV, err := reqP.dbs[4].GetEntry(CFG_L2MC_TABLE_TS, db.Key{[]string{*intfObj.Name}})
			if err != nil {
				return err
			}
			
			fmt.Println("test ---> 2")
			
			reqP.unMarshalIGMPSnoopingIntfStateObjInst(&dbV)
			
			if err := reqP.unMarshalMrouterState(); err != nil {
				return err
			}
			
			if err := reqP.unMarshalStaticGrpStateObj(); err != nil {
				return err
			}
		}							
	}
	
	fmt.Println("translateToYgotObj printing ygot object after unmarshalled ==> ")
	pretty.Print(reqP.igmpsObj)

	return err
}
