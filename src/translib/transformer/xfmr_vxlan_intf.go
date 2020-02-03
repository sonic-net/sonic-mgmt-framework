//////////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom, Inc.
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

package transformer

import (
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
)

type vxlanReqProcessor struct {
	userReqUri         *string
	uri                *string
	uriPath            *gnmipb.Path
	opcode             int
	rootObj            *ocbinds.Device
	targetObj          interface{}
	db                 *db.DB
	dbs                [db.MaxDB]*db.DB
	vxlanIntfConfigObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface_VxlanIf_Config
	vxlanNetInstObj    *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance
	intfObject         *ocbinds.OpenconfigInterfaces_Interfaces_Interface
	targetNode         *yang.Entry
	reqParams          *XfmrParams
}

var applDbPtr, _ = db.NewDB(getDBOptions(db.ApplDB))
var stateDbPtr, _ = db.NewDB(getDBOptions(db.StateDB))
var configDbPtr, _ = db.NewDB(getDBOptions(db.ConfigDB))

func init() {
	// vxlan: interface config
	XlateFuncBind("YangToDb_intf_vxlan_config_xfmr", YangToDb_intf_vxlan_config_xfmr)
	XlateFuncBind("DbToYang_intf_vxlan_config_xfmr", DbToYang_intf_vxlan_config_xfmr)

	// vxlan: vni-network-instance - config - replaced by subtree - vxlan_vni_instance_subtree_xfmr
	//	XlateFuncBind("YangToDb_nw_inst_vxlan_key_xfmr", YangToDb_nw_inst_vxlan_key_xfmr)
	//	XlateFuncBind("DbToYang_nw_inst_vxlan_key_xfmr", DbToYang_nw_inst_vxlan_key_xfmr)
	//	XlateFuncBind("YangToDb_nw_inst_vxlan_vni_id_xfmr", YangToDb_nw_inst_vxlan_vni_id_xfmr)
	//	XlateFuncBind("DbToYang_nw_inst_vxlan_vni_id_xfmr", DbToYang_nw_inst_vxlan_vni_id_xfmr)
	//	XlateFuncBind("YangToDb_nw_inst_vxlan_source_nve_xfmr", YangToDb_nw_inst_vxlan_source_nve_xfmr)
	//	XlateFuncBind("DbToYang_nw_inst_vxlan_source_nve_xfmr", DbToYang_nw_inst_vxlan_source_nve_xfmr)

	// state - vxlan: peer info
	XlateFuncBind("YangToDb_vxlan_vni_state_peer_info_key_xfmr", YangToDb_vxlan_vni_state_peer_info_key_xfmr)
	XlateFuncBind("DbToYang_vxlan_vni_state_peer_info_key_xfmr", DbToYang_vxlan_vni_state_peer_info_key_xfmr)
	XlateFuncBind("YangToDb_vxlan_state_peer_tunnel_type_xfmr", YangToDb_vxlan_state_peer_tunnel_type_xfmr)
	XlateFuncBind("DbToYang_vxlan_state_peer_tunnel_type_xfmr", DbToYang_vxlan_state_peer_tunnel_type_xfmr)

	//state - vxlan: tunnel info
	XlateFuncBind("YangToDb_vxlan_state_tunnel_info_key_xfmr", YangToDb_vxlan_state_tunnel_info_key_xfmr)
	XlateFuncBind("DbToYang_vxlan_state_tunnel_info_key_xfmr", DbToYang_vxlan_state_tunnel_info_key_xfmr)
	XlateFuncBind("YangToDb_vxlan_state_tunnel_info_tunnel_type_xfmr", YangToDb_vxlan_state_tunnel_info_tunnel_type_xfmr)
	XlateFuncBind("DbToYang_vxlan_state_tunnel_info_tunnel_type_xfmr", DbToYang_vxlan_state_tunnel_info_tunnel_type_xfmr)

	XlateFuncBind("YangToDb_vxlan_vni_instance_subtree_xfmr", YangToDb_vxlan_vni_instance_subtree_xfmr)
	XlateFuncBind("DbToYang_vxlan_vni_instance_subtree_xfmr", DbToYang_vxlan_vni_instance_subtree_xfmr)
	
	XlateFuncBind("YangToDb_vlan_nd_suppress_key_xfmr", YangToDb_vlan_nd_suppress_key_xfmr)
	XlateFuncBind("DbToYang_vlan_nd_suppress_key_xfmr", DbToYang_vlan_nd_suppress_key_xfmr)
	XlateFuncBind("YangToDb_vlan_nd_suppress_fld_xfmr", YangToDb_vlan_nd_suppress_fld_xfmr)
	XlateFuncBind("DbToYang_vlan_nd_suppress_fld_xfmr", DbToYang_vlan_nd_suppress_fld_xfmr)
}

var YangToDb_vxlan_vni_state_peer_info_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {

	if log.V(3) {
		log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==>inParams.uri => ", inParams.uri)
		log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==>inParams.requestUri => ", inParams.requestUri)
	}

	pathInfo := NewPathInfo(inParams.uri)
	vniIdStr := pathInfo.Var("vni-id")
	srcIpStr := pathInfo.Var("source-ip")
	peerIpStr := pathInfo.Var("peer-ip")

//	if log.V(3) {
		log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==>vniIdStr => ", vniIdStr)
		log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==>srcIpStr => ", srcIpStr)
		log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==>peerIpStr => ", peerIpStr)
//	}

	if vniIdStr != "" {

		var VXLAN_TUNNEL_TABLE_TS *db.TableSpec = &db.TableSpec{Name: "VXLAN_TUNNEL"}
		tunnelTblData, err := configDbPtr.GetTable(VXLAN_TUNNEL_TABLE_TS)
		if err != nil {
			retErr := tlerr.NotFound("Resource Not Found")
			log.Error("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> returning err ==> ", err)
			return "", retErr
		}

		log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> tunnelTblData ==> ", tunnelTblData)

		tunnelKeys, err := tunnelTblData.GetKeys()
		if err != nil || len(tunnelKeys) != 1 {
			retErr := tlerr.NotFound("Resource Not Found")
			log.Error("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROr ==> ", err)
			return "", retErr
		}

		if log.V(3) {
			log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> tunnelKeys ==> ", tunnelKeys)
		}

		tunnelEntry, err := tunnelTblData.GetEntry(tunnelKeys[0])

		if log.V(3) {
			log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> tunnelEntry ==> ", tunnelEntry)
		}

		if err != nil || len(tunnelEntry.Field) == 0 {
			retErr := tlerr.NotFound("Resource Not Found")
			log.Error("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROr ==> ", err)
			return "", retErr
		}

		if tunnelEntry.Field["src_ip"] != srcIpStr {
			log.Error("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> srcIpStr mismatch")
			retErr := tlerr.NotFound("Resource Not Found")
			log.Error("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROr ==> ", retErr)
			return "", retErr
		}

		var VXLAN_TUNNEL_MAP_TABLE_TS *db.TableSpec = &db.TableSpec{Name: "VXLAN_TUNNEL_MAP"}

		tunnelMapKeyStr := tunnelKeys[0].Comp[0] + "|map_" + vniIdStr + "_Vlan*"

//		if log.V(3) {
			log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> tunnelMapKeyStr ==> ", tunnelMapKeyStr)
//		}

		tblVxlanMapKeys, err := configDbPtr.GetKeysPattern(VXLAN_TUNNEL_MAP_TABLE_TS, db.Key{[]string{tunnelMapKeyStr}})

		if log.V(3) {
			log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> tblVxlanMapKeys ==> ", tblVxlanMapKeys)
		}

		if err != nil || len(tblVxlanMapKeys) != 1 {
			retErr := tlerr.NotFound("Resource Not Found")
			log.Error("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROr ==> ", err)
			return "", retErr
		}

//		if log.V(3) {
			log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> tblVxlanMapKeys ==> ", tblVxlanMapKeys)
//		}

		vlanIdList := strings.Split(tblVxlanMapKeys[0].Comp[1], "_Vlan")

//		if log.V(3) {
			log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> vlanIdList ==> ", vlanIdList)
//		}

		if len(vlanIdList) != 2 {
			retErr := tlerr.NotFound("Resource Not Found")
			log.Error("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROr ==> ", retErr)
			return "", retErr
		}

		var APP_EVPN_REMOTE_VNI_TABLE_TS *db.TableSpec = &db.TableSpec{Name: "EVPN_REMOTE_VNI_TABLE"}
		remote_ip := peerIpStr
		evpnRemoteKey, err := applDbPtr.GetEntry(APP_EVPN_REMOTE_VNI_TABLE_TS, db.Key{[]string{"Vlan" + vlanIdList[1], remote_ip}})

//		if log.V(3) {
			log.Info("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> evpnRemoteKey ==> ", evpnRemoteKey)
//		}

		if err == nil && len(evpnRemoteKey.Field) > 0 {
			retKey := "Vlan" + vlanIdList[1] + ":" + remote_ip
			log.Error("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> final retKey ==> ", retKey)
			return retKey, nil
		} else {
			retErr := tlerr.NotFound("Resource Not Found")
			log.Error("YangToDb_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROr ==> ", retErr)
			return "", retErr
		}
	}

	return "", nil
}

var YangToDb_vxlan_state_tunnel_info_tunnel_type_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	if log.V(3) {
		log.Info("Entering YangToDb_vxlan_state_tunnel_info_tunnel_type_xfmr ===> ")
	}
	return res_map, err
}

var DbToYang_vxlan_state_tunnel_info_tunnel_type_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	rmap["type"] = "dynamic"
	if log.V(3) {
		log.Info("DbToYang_vxlan_state_tunnel_info_tunnel_type_xfmr ==> returning  type field ==> ", rmap)
	}
	return rmap, nil
}

var YangToDb_vxlan_state_peer_tunnel_type_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	if log.V(3) {
		log.Info("Entering YangToDb_vxlan_state_peer_tunnel_type_xfmr ===> ")
	}
	return res_map, err
}

var DbToYang_vxlan_state_peer_tunnel_type_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	//	rmap["tunnel-type"] = ocbinds.OpenconfigVxlan_PeerType_dynamic
	rmap["tunnel-type"] = "dynamic"
	if log.V(3) {
		log.Info("DbToYang_vxlan_state_peer_tunnel_type_xfmr ==> returning  tunnel-type field ==> ", rmap)
	}
	return rmap, nil
}

var DbToYang_vxlan_vni_state_peer_info_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {

	log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr Entering ==> ", inParams)

	rmap := make(map[string]interface{})

	if inParams.key != "" {

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> inParams.key => ", inParams.key)
//		}

		/*
			1) Fetch the VTEP name from the VXLAN_TUNNEL_TABLE (config db)
			2) Use this to fetch the keys with pattern - VXLAN_TUNNEL_MAP_TABLE | <vtep_name>_<vni>_* (config db)
			    get the vlan from this returned key. (only 1 key will be returned)
			3) then fetch the entry from REMOTE_VNI table in APP DB
		*/

		evpnKeyList := strings.Split(inParams.key, ":")

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> evpnKeyList => ", evpnKeyList)
//		}

		var VXLAN_TUNNEL_TABLE_TS *db.TableSpec = &db.TableSpec{Name: "VXLAN_TUNNEL"}
		tunnelTblData, err := configDbPtr.GetTable(VXLAN_TUNNEL_TABLE_TS)
		if err != nil {
			log.Error("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROR => ", err)
			return rmap, tlerr.NotFound("Resource Not Found")
		}

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> tunnelTblData ==> ", tunnelTblData)
//		}

		tunnelKeys, err := tunnelTblData.GetKeys()
		if err != nil || len(tunnelKeys) != 1 {
			log.Error("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROR => ", err)
			return rmap, tlerr.NotFound("Resource Not Found")
		}

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> tunnelKeys ==> ", tunnelKeys)
//		}

		tunnelEntry, err := tunnelTblData.GetEntry(tunnelKeys[0])

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> tunnelEntry ==> ", tunnelEntry)
//		}

		if err != nil || len(tunnelEntry.Field) == 0 {
			log.Error("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROR")
			return rmap, tlerr.NotFound("Resource Not Found")
		}

		var VXLAN_TUNNEL_MAP_TABLE_TS *db.TableSpec = &db.TableSpec{Name: "VXLAN_TUNNEL_MAP"}

		tunnelMapKeyStr := tunnelKeys[0].Comp[0] + "|map_" + "*_" + evpnKeyList[0]

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> tunnelMapKeyStr ==> ", tunnelMapKeyStr)
//		}

		tblVxlanMapKeys, err := configDbPtr.GetKeysPattern(VXLAN_TUNNEL_MAP_TABLE_TS, db.Key{[]string{tunnelMapKeyStr}})

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> tblVxlanMapKeys ==> ", tblVxlanMapKeys)
//		}

		if len(tblVxlanMapKeys) != 1 {
			log.Error("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROR")
			return rmap, tlerr.NotFound("Resource Not Found")
		}

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> tblVxlanMapKeys ==> ", tblVxlanMapKeys)
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> tblVxlanMapKeys[1].Comp[0] ==> ", tblVxlanMapKeys[0].Comp[1])
//		}

		tunnelMapList := strings.Split(tblVxlanMapKeys[0].Comp[1], "_")

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> tunnelMapList ==> ", tunnelMapList)
//		}

		if len(tunnelMapList) != 3 {
			log.Error("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROR")
			return rmap, tlerr.NotFound("Resource Not Found")
		}

		vniIdInt, _ := strconv.ParseInt(tunnelMapList[1], 10, 64)
		rmap["vni-id"] = uint32(vniIdInt)
		rmap["source-ip"] = tunnelEntry.Field["src_ip"]
		rmap["peer-ip"] = evpnKeyList[1]

//		if log.V(3) {
			log.Info("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> final result map ==> ", rmap)
//		}

		return rmap, nil
	} else {
		log.Error("DbToYang_vxlan_vni_state_peer_info_key_xfmr ==> returning ERROR ==> Resource Not Found with empty result => ", rmap)
		return rmap, tlerr.NotFound("Resource Not Found")
	}
}

var YangToDb_vxlan_state_tunnel_info_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
//	if log.V(3) {
		log.Info("YangToDb_vxlan_state_tunnel_info_key_xfmr ==>inParams.uri => ", inParams.uri)
		log.Info("YangToDb_vxlan_state_tunnel_info_key_xfmr ==>inParams.requestUri => ", inParams.requestUri)
//	}

	pathInfo := NewPathInfo(inParams.uri)
	peerIpStr := pathInfo.Var("peer-ip")
//	if log.V(3) {
		log.Info("YangToDb_vxlan_state_tunnel_info_key_xfmr ==> peerIpStr => ", peerIpStr)
//	}

	pathOrigInfo := NewPathInfo(inParams.requestUri)
	peerIpOrigStr := pathOrigInfo.Var("peer-ip")
	
//	if log.V(3) {
		log.Info("YangToDb_vxlan_state_tunnel_info_key_xfmr ==> peerIpOrigStr => ", peerIpOrigStr)
//	}

	if peerIpOrigStr != "" {
		var VXLAN_TUNNEL_TABLE_STATE_TS *db.TableSpec = &db.TableSpec{Name: "VXLAN_TUNNEL_TABLE"}
		evpnPeerkeyStr := "EVPN_" + peerIpOrigStr
		if log.V(3) {
			log.Info("YangToDb_vxlan_state_tunnel_info_key_xfmr ==> evpnPeerkeyStr ==> ", evpnPeerkeyStr)
		}
		_, err := stateDbPtr.GetEntry(VXLAN_TUNNEL_TABLE_STATE_TS, db.Key{[]string{evpnPeerkeyStr}})
		if err != nil {
			log.Info("YangToDb_vxlan_state_tunnel_info_key_xfmr ==> returning error ==> ", err)
			return "", tlerr.NotFound("Resource Not Found")
		}
	}

	if peerIpStr != "" {
		evpnPeerkeyStr := "EVPN_" + peerIpStr
		if log.V(3) {
			log.Info("YangToDb_vxlan_state_tunnel_info_key_xfmr ==> returning KEY => ", evpnPeerkeyStr)
		}
		return evpnPeerkeyStr, nil
	} else {
		return "", nil
	}
}

var DbToYang_vxlan_state_tunnel_info_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
//	if log.V(3) {
		log.Info("DbToYang_vxlan_state_tunnel_info_key_xfmr ==> inParams.key => ", inParams.key)
//	}

	pathOrigInfo := NewPathInfo(inParams.requestUri)
	peerIpOrigStr := pathOrigInfo.Var("peer-ip")
	
//	if log.V(3) {
		log.Info("DbToYang_vxlan_state_tunnel_info_key_xfmr ==> peerIpOrigStr => ", peerIpOrigStr)
//	}

	rmap := make(map[string]interface{})
	if inParams.key != "" {
		//VXLAN_TUNNEL_TABLE|EVPN_6.6.6.2
		keyListTmp := strings.Split(inParams.key, "EVPN_")
		if peerIpOrigStr != "" && peerIpOrigStr != keyListTmp[1] {
			log.Info("DbToYang_vxlan_state_tunnel_info_key_xfmr ==> returning ERROR => peer-ip not exist => ", peerIpOrigStr)
			return rmap, tlerr.NotFound("Resource Not Found")
		}
		if log.V(3) {
			log.Info("DbToYang_vxlan_state_tunnel_info_key_xfmr ==> keyListTmp => ", keyListTmp)
		}
		if len(keyListTmp) == 2 {
			rmap["peer-ip"] = keyListTmp[1]
		}
	}
//	if log.V(3) {
		log.Info("DbToYang_vxlan_state_tunnel_info_key_xfmr ==> returning RESTULT map rmap => ", rmap)
//	}
	return rmap, nil
}

func (reqP *vxlanReqProcessor) setVxlanIntfFromReq() error {
	var err error

	vxlanIntfConfigPath := &gnmipb.Path{}

	var pathList []*gnmipb.PathElem = reqP.uriPath.Elem

	for i := 0; i < len(pathList); i++ {
		vxlanIntfConfigPath.Elem = append(vxlanIntfConfigPath.Elem, pathList[i])
		if pathList[i].Name == "vxlan-if" {
			break
		}
	}

	if log.V(3) {
		log.Info("vxlanIntfConfigPath => ", vxlanIntfConfigPath)
	}

	targetNodeList, err := ytypes.GetNode(ocbinds.SchemaTree["Device"], reqP.rootObj, vxlanIntfConfigPath)

	if err != nil {
		return tlerr.InvalidArgs("Interface list node not found in the request: %v", err)
	}

	if len(targetNodeList) == 0 {
		return tlerr.InvalidArgs("Interfaces node not found in the request: %s", *reqP.uri)
	}

	vxlanItfObj := targetNodeList[0].Data.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_VxlanIf)
	if vxlanItfObj != nil {
		reqP.vxlanIntfConfigObj = vxlanItfObj.Config
	}

	if log.V(3) {
		log.Info("vxlanIntfConfigObj ==> ", reqP.vxlanIntfConfigObj)
	}

	return err
}

func (reqP *vxlanReqProcessor) setVxlanNetInstObjFromReq() error {
	var err error

	if log.V(3) {
		log.Info("setVxlanNetInstObjFromReq entreing => ")
	}

	vxlanNetInstObjPath := &gnmipb.Path{}

	var pathList []*gnmipb.PathElem = reqP.uriPath.Elem

	for i := 0; i < len(pathList); i++ {
		vxlanNetInstObjPath.Elem = append(vxlanNetInstObjPath.Elem, pathList[i])
		if pathList[i].Name == "network-instance" {
			break
		}
	}

	if log.V(3) {
		log.Info("vxlanNetInstObj path => ", vxlanNetInstObjPath)
	}

	targetNodeList, err := ytypes.GetNode(ocbinds.SchemaTree["Device"], reqP.rootObj, vxlanNetInstObjPath)

	if err != nil {
		return tlerr.NotFound("Resource Not Found")
	}

	if len(targetNodeList) == 0 {
		return tlerr.NotFound("Resource Not Found")
	}

	reqP.vxlanNetInstObj = targetNodeList[0].Data.(*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance)

	if log.V(3) {
		log.Info("vxlanNetInstObj ==> ", reqP.vxlanNetInstObj)
	}

	return err
}

func (reqP *vxlanReqProcessor) setIntfObjFromReq() error {
	var err error

	intfTargetPath := &gnmipb.Path{}

	var pathList []*gnmipb.PathElem = reqP.uriPath.Elem

	for i := 0; i < (len(pathList) - 2); i++ {
		intfTargetPath.Elem = append(intfTargetPath.Elem, pathList[i])
		if pathList[i].Name == "interface" {
			break
		}
	}

	if log.V(3) {
		log.Info("intfTargetPath => ", intfTargetPath)
	}

	targetNodeList, err := ytypes.GetNode(ocbinds.SchemaTree["Device"], reqP.rootObj, intfTargetPath)

	if err != nil {
		return tlerr.InvalidArgs("Interface list node not found in the request: %v", err)
	}

	if len(targetNodeList) == 0 {
		return tlerr.InvalidArgs("Interfaces node not found in the request: %s", *reqP.uri)
	}

	reqP.intfObject = targetNodeList[0].Data.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface)

	if log.V(3) {
		log.Info("intfTargetObj ==> ", reqP.intfObject)
	}

	return err
}

func (reqP *vxlanReqProcessor) handleDeleteReq() (*map[string]map[string]db.Value, error) {

	var res_map map[string]map[string]db.Value = make(map[string]map[string]db.Value)

	if log.V(3) {
		log.Info(" handleDeleteReq entering ====== reqP.targetNode.Name ======> ", reqP.targetNode.Name)
		log.Info(" handleDeleteReq entering ====== reqP.intfObject.VxlanIf ======> ", reqP.intfObject.VxlanIf)
	}

	if reqP.targetNode.Name == "state" || (reqP.intfObject.VxlanIf != nil && reqP.intfObject.VxlanIf.State != nil) {
		return &res_map, tlerr.InvalidArgs("Method Not Allowed")
	} else {
		var vxlanIntfConfTbl map[string]db.Value = make(map[string]db.Value)
		var evpnNvoTbl map[string]db.Value = make(map[string]db.Value)

		if log.V(3) {
			log.Info(" handleDeleteReq entering ============> reqP.userReqUri ====> ", reqP.userReqUri)
		}

		pathInfo := NewPathInfo(*reqP.userReqUri)
		vxlanIntfName := pathInfo.Var("name")

		if log.V(3) {
			log.Info(" =====> handleDeleteReq ==> handleDeleteReq - vxlanIntfName => ", vxlanIntfName)
		}

		vxlanIntfdbV := db.Value{Field: make(map[string]string)}
		evpnNvodbV := db.Value{Field: make(map[string]string)}

		if vxlanIntfName != "" {
			var VXLAN_TUNNEL_TABLE_TS *db.TableSpec = &db.TableSpec{Name: "VXLAN_TUNNEL"}
			tblValList, err := reqP.db.GetEntry(VXLAN_TUNNEL_TABLE_TS, db.Key{[]string{vxlanIntfName}})
			if log.V(3) {
				log.Info(" =====> handleDeleteReq ==> handleDeleteReq - tblValList => ", tblValList)
			}
			if err != nil {
				return &res_map, tlerr.NotFound("Resource Not Found")
			}

			if log.V(3) {
				log.Info(" =====> handleDeleteReq ==> handleDeleteReq - src_ip => ", tblValList.Field["src_ip"])
			}

			if tblValList.Field["src_ip"] != "" {

				var VXLAN_TUNNEL_MAP_TS *db.TableSpec = &db.TableSpec{Name: "VXLAN_TUNNEL_MAP"}
				tunnelMapKeyStr := vxlanIntfName + "|*"
				if log.V(3) {
					log.Info("handleDeleteReq ==> tunnelMapKeyStr ==> ", tunnelMapKeyStr)
				}
				tblVxlanMapKeys, _ := reqP.db.GetKeysPattern(VXLAN_TUNNEL_MAP_TS, db.Key{[]string{tunnelMapKeyStr}})
				if log.V(3) {
					log.Info("handleDeleteReq ==> tblVxlanMapKeys ==> ", tblVxlanMapKeys)
				}

				if len(tblVxlanMapKeys) > 0 {
					log.Error("handleDeleteReq ==> returning ERROR")
					return &res_map, tlerr.New("source-vtep-ip cannot be deleted since tunnel map (VLAN-VNI) has reference to the vxlan interface \"%s\" of the source-vtep-ip %s", vxlanIntfName, tblValList.Field["src_ip"])
				}

				vxlanIntfdbV.Field["src_ip"] = tblValList.Field["src_ip"]
				subOpMap := make(map[db.DBNum]map[string]map[string]db.Value)
				subOpMap[db.ConfigDB] = make(map[string]map[string]db.Value)
				subOpMap[db.ConfigDB]["VXLAN_TUNNEL"] = make(map[string]db.Value)
				subOpMap[db.ConfigDB]["VXLAN_TUNNEL"][vxlanIntfName] = db.Value{Field: make(map[string]string)}
				subOpMap[db.ConfigDB]["VXLAN_TUNNEL"][vxlanIntfName].Field["NULL"] = "NULL"
				reqP.reqParams.subOpDataMap[UPDATE] = &subOpMap
				vxlanIntfConfTbl[vxlanIntfName] = vxlanIntfdbV
				res_map["VXLAN_TUNNEL"] = vxlanIntfConfTbl
				evpnNvodbV.Field["source_vtep"] = vxlanIntfName
				evpnNvoTbl["nvo1"] = evpnNvodbV
				res_map["EVPN_NVO"] = evpnNvoTbl
			} else {
				return &res_map, tlerr.NotFound("Resource Not Found")
			}
		} else {
			return &res_map, tlerr.NotFound("Resource Not Found")
		}
	}

	if log.V(3) {
		log.Info(" =====> handleDeleteReq ==> handleDeleteReq - res_map => ", res_map)
	}
	return &res_map, nil
}

// handle create/replace/update request
func (reqP *vxlanReqProcessor) handleCRUReq() (*map[string]map[string]db.Value, error) {

	var res_map map[string]map[string]db.Value = make(map[string]map[string]db.Value)

	if log.V(3) {
		log.Info(" handleCRUReq entering ============> reqP.userReqUri ====> ", reqP.userReqUri)
	}

	pathInfo := NewPathInfo(*reqP.userReqUri)
	vxlanIntfName := pathInfo.Var("name")

	if log.V(3) {
		log.Info(" =====> vxlanReqProcessor ==> handleCRUReq - vxlanIntfName => ", vxlanIntfName)
	}

	if vxlanIntfName != "" {
		var VXLAN_TUNNEL_TABLE_TS *db.TableSpec = &db.TableSpec{Name: "VXLAN_TUNNEL"}
		dbv, err := reqP.db.GetEntry(VXLAN_TUNNEL_TABLE_TS, db.Key{[]string{vxlanIntfName}})
		if log.V(3) {
			log.Info("VXLAN testing YangToDb_intf_tbl_key_xfmr ========  GetEntry ===========> dbv => ", dbv)
			log.Info("VXLAN testing YangToDb_intf_tbl_key_xfmr ========  GetEntry ===========> err => ", err)
		}
		if err != nil {
			return &res_map, tlerr.NotFound("Resource Not Found")
		}

		if dbv.Field["src_ip"] != "" && (reqP.opcode == 3 || reqP.opcode == 4) {
			log.Error("VXLAN testing YangToDb_intf_tbl_key_xfmr ========  GetEntry ===========> source-vtep-ip => ", dbv.Field["src_ip"])
			return &res_map, tlerr.New("source-vtep-ip %s is already exist; PUT and PATCH method not allowed on the \"source-vtep-ip\"", dbv.Field["src_ip"])
		}
	}

	var vxlanTunnelTblMap map[string]db.Value = make(map[string]db.Value)
	var evpnNvoTblMap map[string]db.Value = make(map[string]db.Value)

	if reqP.vxlanIntfConfigObj.SourceVtepIp == nil {
		log.Error(" =====> vxlanReqProcessor ==> handleCRUReq - ERROR ")
		pretty.Print(res_map)
		return &res_map, tlerr.InvalidArgs("Cannot configure the Vxlan interface without source-vtep-ip; Please proivde the source-vtep-ip - /openconfig-interfaces:interfaces/interface/openconfig-vxlan:vxlan-if/config/source-vtep-ip")
	} else {
		dbV1 := db.Value{Field: make(map[string]string)}
		srcIp := *(reqP.vxlanIntfConfigObj.SourceVtepIp)
		dbV1.Field["src_ip"] = srcIp
		vxlanTunnelTblMap[*(reqP.intfObject.Name)] = dbV1
		dbV2 := db.Value{Field: make(map[string]string)}
		dbV2.Field["source_vtep"] = *(reqP.intfObject.Name)
		evpnNvoTblMap["nvo1"] = dbV2
		res_map["VXLAN_TUNNEL"] = vxlanTunnelTblMap
		res_map["EVPN_NVO"] = evpnNvoTblMap
	}
	
	if log.V(3) {
		log.Info(" =====> vxlanReqProcessor ==> handleCRUReq - success - res_map => ", res_map)
	}

	return &res_map, nil
}

func (reqP *vxlanReqProcessor) translateToDb() (*map[string]map[string]db.Value, error) {
	//DELETE
	if reqP.opcode == 5 {
		// get the target node
		var err error
		if reqP.targetNode, err = getYangNode(reqP.uriPath); err != nil {
			return nil, tlerr.InvalidArgs("Invalid request: %s", *reqP.uri)
		}

		if log.V(3) {
			log.Info("translateToDb param reqP.targetNode.Name ==> ", reqP.targetNode.Name)
		}

		res_map, err := reqP.handleDeleteReq()

		if err != nil {
			return nil, err
		}

		return res_map, err

	} else if reqP.vxlanIntfConfigObj != nil {
		res_map, err := reqP.handleCRUReq()
		if err != nil {
			return nil, err
		}
		return res_map, err
	} else {
		return nil, tlerr.InvalidArgs("Invalid Request")
	}
}

func getIntfUriPath(uri string) (*gnmipb.Path, error) {
	uriPath := strings.Replace(uri, "openconfig-interfaces:", "", -1)
	uriPath = strings.Replace(uri, "openconfig-vxlan:", "", -1)
	path, err := ygot.StringToPath(uriPath, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		return nil, tlerr.NotFound("Resource Not Found")
	}
	for _, p := range path.Elem {
		pathSlice := strings.Split(p.Name, ":")
		p.Name = pathSlice[len(pathSlice)-1]
	}
	return path, nil
}

func getVxlanNiUriPath(uri string) (*gnmipb.Path, error) {
	uriPath := strings.Replace(uri, "openconfig-network-instance:", "", -1)
	uriPath = strings.Replace(uri, "openconfig-vxlan:", "", -1)
	path, err := ygot.StringToPath(uriPath, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		return nil, tlerr.NotFound("Resource Not Found")
	}
	for _, p := range path.Elem {
		pathSlice := strings.Split(p.Name, ":")
		p.Name = pathSlice[len(pathSlice)-1]
	}
	return path, nil
}

var YangToDb_intf_vxlan_config_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
	var err error

	if log.V(3) {
		log.Info("YangToDb_intf_vxlan_config_xfmr entering => inParams.uri => ", inParams.uri)
	}

	path, err := getIntfUriPath(inParams.uri)

	if err != nil {
		return nil, err
	}

	reqP := &vxlanReqProcessor{&inParams.requestUri, &inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil, &inParams}

	if err := reqP.setVxlanIntfFromReq(); err != nil {
		return nil, err
	}

	if log.V(3) {
		log.Info("YangToDb_intf_vxlan_config_xfmr ==> printing vxlanIntfConfigPath object request ==> ", (*reqP.vxlanIntfConfigObj))
	}

	if err := reqP.setIntfObjFromReq(); err != nil {
		return nil, err
	}

	if log.V(3) {
		log.Info("YangToDb_intf_vxlan_config_xfmr ==> printing intf object request ==> ", (*reqP.intfObject))
	}

	res_map, err := reqP.translateToDb()

	if err == nil {
		return *res_map, nil
	} else {
		return nil, err
	}
}

var DbToYang_intf_vxlan_config_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {

	if log.V(3) {
		log.Info("Entering DbToYang_intf_vxlan_config_xfmr ===> inParams.uri => ", inParams.uri)
	}

	path, err := getIntfUriPath(inParams.uri)

	if err != nil {
		return err
	}

	reqP := &vxlanReqProcessor{&inParams.requestUri, &inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil, &inParams}

	if err := reqP.setIntfObjFromReq(); err != nil {
		return err
	}

	if reqP.intfObject != nil && reqP.intfObject.Name != nil {

		if log.V(3) {
			log.Info("DbToYang_intf_vxlan_config_xfmr ==> printing intf object request ==> ", (*reqP.intfObject))
		}

		vxlanIntfName := *reqP.intfObject.Name

		if log.V(3) {
			log.Info("DbToYang_intf_vxlan_config_xfmr ==> vxlanIntfName ==> ", vxlanIntfName)
		}

		if vxlanIntfName != "" {
			var VXLAN_TUNNEL_TABLE_TS *db.TableSpec = &db.TableSpec{Name: "VXLAN_TUNNEL"}
			dbv, err := reqP.db.GetEntry(VXLAN_TUNNEL_TABLE_TS, db.Key{[]string{vxlanIntfName}})
			if log.V(3) {
				log.Info("DbToYang_intf_vxlan_config_xfmr ========  GetEntry ===> dbv => ", dbv)
			}
			if err != nil {
				return tlerr.NotFound("Resource Not Found")
			}

			srcIpStr := dbv.Field["src_ip"]

			if log.V(3) {
				log.Info("DbToYang_intf_vxlan_config_xfmr ========  srcIpStr ===> ", srcIpStr)
			}

			if srcIpStr != "" {
				ygot.BuildEmptyTree(reqP.intfObject)
				if reqP.intfObject.VxlanIf != nil {
					if log.V(3) {
						log.Info("DbToYang_intf_vxlan_config_xfmr ========  reqP.intfObject.VxlanIf.Config => ", reqP.intfObject.VxlanIf.Config)
					}
					ygot.BuildEmptyTree(reqP.intfObject.VxlanIf)
					if log.V(3) {
						log.Info("DbToYang_intf_vxlan_config_xfmr ========  reqP.intfObject.VxlanIf.Config => ", reqP.intfObject.VxlanIf.Config)
					}
					reqP.intfObject.VxlanIf.Config.SourceVtepIp = &srcIpStr
					if log.V(3) {
						log.Info("DbToYang_intf_vxlan_config_xfmr ========  reqP.vxlanIntfConfigObj.SourceVtepIp ===> ", reqP.intfObject.VxlanIf.Config.SourceVtepIp)
					}
				} else {
					log.Error("DbToYang_intf_vxlan_config_xfmr ========  reqP.intfObject.VxlanIf is nil")
				}
			}
		}
	}

	return nil
}

var DbToYang_nw_inst_vxlan_vni_id_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})

	entry_key := inParams.key
	if log.V(3) {
		log.Info("DbToYang_nw_inst_vxlan_vni_id_xfmr ==> entry_key  ===> ", entry_key)
	}

	if entry_key != "" {
		keyList := strings.Split(entry_key, "|")

		if log.V(3) {
			log.Info("DbToYang_nw_inst_vxlan_vni_id_xfmr ==> keyList  ===> ", keyList)
		}

		mapNameList := strings.Split(keyList[1], "_")

		vniId, _ := strconv.ParseInt(mapNameList[1], 10, 64)
		rmap["vni-id"] = uint32(vniId)
	}

	if log.V(3) {
		log.Info("DbToYang_nw_inst_vxlan_vni_id_xfmr ==> rmap  ===> ", rmap)
	}

	return rmap, nil
}

var YangToDb_nw_inst_vxlan_vni_id_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error

	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_vni_id_xfmr ==> printing target object inParams.uri ==> ", (inParams.uri))
	}

	path, err := getVxlanNiUriPath(inParams.uri)

	if err != nil {
		return res_map, err
	}

	reqP := &vxlanReqProcessor{&inParams.requestUri, &inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil, &inParams}

	if err := reqP.setVxlanNetInstObjFromReq(); err != nil {
		return nil, err
	}

	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_vni_id_xfmr ==> printing vxlanNetInstObj object request ==> ", (*reqP.vxlanNetInstObj))
	}

	if reqP.vxlanNetInstObj.VxlanVniInstances == nil {
		return res_map, tlerr.NotFound("Resource Not Found")
	}

	for _, vxlanNiMap := range reqP.vxlanNetInstObj.VxlanVniInstances.VniInstance {
		res_map["vni"] = strconv.Itoa(int(*vxlanNiMap.Config.VniId))
		break
	}

	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_vni_id_xfmr ==> res_map  ===> ", res_map)
	}

	return res_map, err
}

var YangToDb_nw_inst_vxlan_source_nve_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error

	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_source_nve_xfmr ==> printing target object request ==> ", inParams.param)
	}

	path, err := getVxlanNiUriPath(inParams.uri)

	if err != nil {
		return res_map, err
	}

	reqP := &vxlanReqProcessor{&inParams.requestUri, &inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil, &inParams}

	if err = reqP.setVxlanNetInstObjFromReq(); err != nil {
		return res_map, err
	}

	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_source_nve_xfmr ==> printing vxlanNetInstObj object request ==> ", (*reqP.vxlanNetInstObj))
	}

	niName := *(reqP.vxlanNetInstObj.Name)
	res_map["vlan"] = niName

	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_source_nve_xfmr ==> res_map  ===> ", res_map)
	}

	return res_map, err
}

var DbToYang_nw_inst_vxlan_source_nve_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})

	entry_key := inParams.key
	if log.V(3) {
		log.Info("DbToYang_nw_inst_vxlan_source_nve_xfmr ==> entry_key  ===> ", entry_key)
	}

	if entry_key != "" {
		keyList := strings.Split(entry_key, "|")
		if log.V(3) {
			log.Info("DbToYang_nw_inst_vxlan_source_nve_xfmr ==> keyList  ===> ", keyList)
		}
		rmap["source-nve"] = keyList[0]
	}

	if log.V(3) {
		log.Info("DbToYang_nw_inst_vxlan_source_nve_xfmr ==> rmap  ===> ", rmap)
	}

	return rmap, nil
}

var YangToDb_nw_inst_vxlan_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	var err error

	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_key_xfmr ==> printing target object request ==> ", (inParams.param))
	}

	path, err := getVxlanNiUriPath(inParams.uri)

	if err != nil {
		return "", err
	}

	reqP := &vxlanReqProcessor{&inParams.requestUri, &inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil, &inParams}

	if err = reqP.setVxlanNetInstObjFromReq(); err != nil {
		return "", err
	}
	
	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_key_xfmr ==> printing vxlanNetInstObj object request ==> ", (*reqP.vxlanNetInstObj))
	}

	var keyStr string
	var srcVetpName string
	var vniIdStr string

	if reqP.opcode == 5 || reqP.opcode == 1 {
		pathInfo := NewPathInfo(inParams.uri)
		srcVetpName = pathInfo.Var("source-nve")
		vniIdStr = pathInfo.Var("vni-id")
		if reqP.vxlanNetInstObj.VxlanVniInstances == nil || len(reqP.vxlanNetInstObj.VxlanVniInstances.VniInstance) == 0 && srcVetpName == "" {
			log.Error("YangToDb_nw_inst_vxlan_key_xfmr ==> returning EMPTY key, since there is no key in the request")
			return "", nil
		}
	}

	if srcVetpName == "" && vniIdStr == "" {
		if reqP.vxlanNetInstObj.VxlanVniInstances == nil {
			log.Error("YangToDb_nw_inst_vxlan_key_xfmr ==> returning EMPTY key, since there is no key in the request")
			return "", tlerr.NotFound("Resource Not Found")
		}

		for _, vxlanNiObj := range reqP.vxlanNetInstObj.VxlanVniInstances.VniInstance {
			srcVetpName = *vxlanNiObj.SourceNve
			vniIdStr = strconv.Itoa(int(*vxlanNiObj.VniId))
			break
		}
	}

	niName := *(reqP.vxlanNetInstObj.Name)
	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_key_xfmr ==> niName  ===> ", niName)
	}

	keyStr = srcVetpName + "|" + "map_" + vniIdStr + "_" + niName

	if log.V(3) {
		log.Info("YangToDb_nw_inst_vxlan_key_xfmr ==> keyStr  ===> ", keyStr)
	}

	//Vtep1|map_100_Vlan5

	return keyStr, nil
}

var DbToYang_nw_inst_vxlan_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	entry_key := inParams.key
	if log.V(3) {
		log.Info("DbToYang_nw_inst_vxlan_key_xfmr ==> entry_key  ===> ", entry_key)
	}

	if entry_key != "" {
		keyList := strings.Split(entry_key, "|")
		if log.V(3) {
			log.Info("DbToYang_nw_inst_vxlan_key_xfmr ==> keyList  ===> ", keyList)
		}

		rmap["source-nve"] = keyList[0]
		mapNameList := strings.Split(keyList[1], "_")

		vniId, _ := strconv.ParseInt(mapNameList[1], 10, 64)
		rmap["vni-id"] = uint32(vniId)
	}

	if log.V(3) {
		log.Info("DbToYang_nw_inst_vxlan_key_xfmr ==> rmap  ===> ", rmap)
	}

	return rmap, nil
}

var YangToDb_vxlan_vni_instance_subtree_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
	var err error
	var tblName string
	res_map := make(map[string]map[string]db.Value)
	valueMap := make(map[string]db.Value)
	pathInfo := NewPathInfo(inParams.uri)
	if log.V(3) {
		log.Info("YangToDb_vxlan_vni_instance_subtree_xfmr: ", inParams.ygRoot, inParams.uri)
	}

	path, err := getVxlanNiUriPath(inParams.uri)
	if err != nil {
		return res_map, err
	}

	reqP := &vxlanReqProcessor{&inParams.requestUri, &inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil, &inParams}
	if err := reqP.setVxlanNetInstObjFromReq(); err != nil {
		return nil, err
	}

	if reqP.opcode != DELETE && (reqP.vxlanNetInstObj.VxlanVniInstances == nil || len(reqP.vxlanNetInstObj.VxlanVniInstances.VniInstance) == 0) {
		return res_map, tlerr.NotFound("Resource Not Found")
	}

	niName := *(reqP.vxlanNetInstObj.Name)
	if strings.HasPrefix(niName, "Vlan") {
		tblName = "VXLAN_TUNNEL_MAP"
	} else if strings.HasPrefix(niName, "Vrf") {
		tblName = "VRF"
	} else {
		return res_map, tlerr.InvalidArgs("Invalid Network Instance name: %s", niName)
	}

	pathInfoOrig := NewPathInfo(inParams.requestUri) // orignial user given URI
	if log.V(3) {
		log.Info("YangToDb_vxlan_vni_instance_subtree_xfmr: pathInfoOrig => ", pathInfoOrig)
	}
	vniIdKeyStr := pathInfoOrig.Var("vni-id")
	srcNveKeyStr := pathInfoOrig.Var("source-nve")
	if log.V(3) {
		log.Info("YangToDb_vxlan_vni_instance_subtree_xfmr: vniIdKeyStr in URI => ", vniIdKeyStr)
		log.Info("YangToDb_vxlan_vni_instance_subtree_xfmr: srcNveKeyStr in URI => ", srcNveKeyStr)
	}

	if vniIdKeyStr != "" && srcNveKeyStr != "" {
		if tblName == "VXLAN_TUNNEL_MAP" {
			var VXLAN_TUNNEL_MAP_TS *db.TableSpec = &db.TableSpec{Name: tblName}
			tunnelMapKeyStr := "map_" + vniIdKeyStr + "_" + niName
			if log.V(3) {
				log.Info("YangToDb_vxlan_vni_instance_subtree_xfmr: tunnelMapKeyStr => ", tunnelMapKeyStr)
			}
			_, err := reqP.db.GetEntry(VXLAN_TUNNEL_MAP_TS, db.Key{Comp: []string{srcNveKeyStr, tunnelMapKeyStr}})
			if log.V(3) {
				log.Info("YangToDb_vxlan_vni_instance_subtree_xfmr: tblVxlanMapKeys => err => ", err)
			}
			if err != nil {
				log.Error("YangToDb_vxlan_vni_instance_subtree_xfmr ==> returning ERROR, since the key doesn't exist")
				return res_map, tlerr.NotFound("Resource Not Found")
			}
		}
	}

	var vniId uint32
	var vtepName string
	var tblKeyStr string

	if reqP.opcode == DELETE && (pathInfo.Template == "/openconfig-network-instance:network-instances/network-instance{name}/openconfig-vxlan:vxlan-vni-instances/vni-instance" ||
		pathInfo.Template == "/openconfig-network-instance:network-instances/network-instance{name}/openconfig-vxlan:vxlan-vni-instances") {
		dbKeys, err := inParams.d.GetKeys(&db.TableSpec{Name: tblName})
		if err != nil {
			return res_map, err
		}
		if len(dbKeys) > 0 {
			for _, dbkey := range dbKeys {
				if strings.HasPrefix(niName, "Vlan") {
					vtepName = dbkey.Get(0)
					mapNameList := strings.Split(dbkey.Get(1), "_")
					vniNum, _ := strconv.ParseUint(mapNameList[1], 10, 32)
					vniId = uint32(vniNum)
				} else if strings.HasPrefix(niName, "Vrf") {
					vrfEntry, err := inParams.d.GetEntry(&db.TableSpec{Name: tblName}, db.Key{Comp: []string{niName}})
					if err != nil {
						return res_map, err
					}
					if vrfEntry.Has("vni") {
						vniIdStr := vrfEntry.Get("vni")
						vniNum, _ := strconv.ParseUint(vniIdStr, 10, 32)
						vniId = uint32(vniNum)
					}
				}
			}
		}
	} else {
		for vniKey, _ := range reqP.vxlanNetInstObj.VxlanVniInstances.VniInstance {
			vniId = vniKey.VniId
			vtepName = vniKey.SourceNve
			break
		}
	}

	if strings.HasPrefix(niName, "Vlan") {
		tblKeyStr = vtepName + "|" + "map_" + strconv.Itoa(int(vniId)) + "_" + niName
		valueMap[tblKeyStr] = db.Value{Field: make(map[string]string)}
		valueMap[tblKeyStr].Field["vlan"] = niName
		valueMap[tblKeyStr].Field["vni"] = strconv.Itoa(int(vniId))
	} else if strings.HasPrefix(niName, "Vrf") {
		tblKeyStr = niName
		valueMap[tblKeyStr] = db.Value{Field: make(map[string]string)}
		valueMap[tblKeyStr].Field["vni"] = strconv.Itoa(int(vniId))
	}

	res_map[tblName] = valueMap
	return res_map, err
}

var DbToYang_vxlan_vni_instance_subtree_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
	var err error
	pathInfo := NewPathInfo(inParams.uri)
	if log.V(3) {
		log.Info("DbToYang_vxlan_vni_instance_subtree_xfmr: ", pathInfo.Template)
	}

	path, err := getVxlanNiUriPath(inParams.uri)
	if err != nil {
		return err
	}
	reqP := &vxlanReqProcessor{&inParams.requestUri, &inParams.uri, path, inParams.oper, (*inParams.ygRoot).(*ocbinds.Device), inParams.param, inParams.d, inParams.dbs, nil, nil, nil, nil, &inParams}
	if err := reqP.setVxlanNetInstObjFromReq(); err != nil {
		return err
	}

	var vniIdStr string
	var vtepName string
	var tblKeyStr string
	var tblName string
	configDb := inParams.dbs[db.ConfigDB]
	niName := pathInfo.Var("name")

	if strings.HasPrefix(niName, "Vlan") {
		tblName = "VXLAN_TUNNEL_MAP"
	} else if strings.HasPrefix(niName, "Vrf") {
		tblName = "VRF"
	} else {
		log.Errorf("Invalid Network Instance name: %s", niName)
		return tlerr.InvalidArgs("Invalid Network Instance name: %s", niName)
	}

	pathInfoOrig := NewPathInfo(inParams.requestUri) // orignial user given URI
	if log.V(3) {
		log.Info("DbToYang_vxlan_vni_instance_subtree_xfmr: pathInfoOrig => ", pathInfoOrig)
	}
	vniIdKeyStr := pathInfoOrig.Var("vni-id")
	srcNveKeyStr := pathInfoOrig.Var("source-nve")
	if log.V(3) {
		log.Info("DbToYang_vxlan_vni_instance_subtree_xfmr: vniIdKeyStr in URI => ", vniIdKeyStr)
		log.Info("DbToYang_vxlan_vni_instance_subtree_xfmr: srcNveKeyStr in URI => ", srcNveKeyStr)
	}

	if vniIdKeyStr != "" && srcNveKeyStr != "" {
		if tblName == "VXLAN_TUNNEL_MAP" {
			var VXLAN_TUNNEL_MAP_TS *db.TableSpec = &db.TableSpec{Name: tblName}
			tunnelMapKeyStr := "map_" + vniIdKeyStr + "_" + niName
			if log.V(3) {
				log.Info("DbToYang_vxlan_vni_instance_subtree_xfmr: tunnelMapKeyStr => ", tunnelMapKeyStr)
			}
			_, err := reqP.db.GetEntry(VXLAN_TUNNEL_MAP_TS, db.Key{Comp: []string{srcNveKeyStr, tunnelMapKeyStr}})
			if log.V(3) {
				log.Info("DbToYang_vxlan_vni_instance_subtree_xfmr: tblVxlanMapKeys => err => ", err)
			}
			if err != nil {
				log.Error("DbToYang_vxlan_vni_instance_subtree_xfmr ==> returning ERROR, since the key doesn't exist")
				return tlerr.NotFound("Resource Not Found")
			}
		}
	}

	if isSubtreeRequest(pathInfo.Template, "/openconfig-network-instance:network-instances/network-instance{name}/openconfig-vxlan:vxlan-vni-instances/vni-instance{vni-id}{source-nve}") {
		vniIdStr = pathInfo.Var("vni-id")
		vtepName = pathInfo.Var("source-nve")

		if strings.HasPrefix(niName, "Vlan") {
			tblKeyStr = vtepName + "|" + "map_" + vniIdStr + "_" + niName
		} else if strings.HasPrefix(niName, "Vrf") {
			tblKeyStr = niName
		}

		dbEntry, err := configDb.GetEntry(&db.TableSpec{Name: tblName}, db.Key{Comp: []string{tblKeyStr}})
		if err != nil {
			return err
		}
		if dbEntry.Get("vni") != vniIdStr {
			log.Errorf("Network instance %s not associated with vni %s", niName, vniIdStr)
			return tlerr.NotFound("Resource Not Found")
		}

		if reqP.vxlanNetInstObj.VxlanVniInstances != nil || len(reqP.vxlanNetInstObj.VxlanVniInstances.VniInstance) > 0 {
			for vniKey := range reqP.vxlanNetInstObj.VxlanVniInstances.VniInstance {
				vniInst := reqP.vxlanNetInstObj.VxlanVniInstances.VniInstance[vniKey]

				vniId := vniKey.VniId
				srcNve := vniKey.SourceNve
				fillVniInstanceDetails(niName, vniId, srcNve, vniInst)
			}
		}
	} else if isSubtreeRequest(pathInfo.Template, "/openconfig-network-instance:network-instances/network-instance{name}/openconfig-vxlan:vxlan-vni-instances") {
		dbKeys, err := configDb.GetKeys(&db.TableSpec{Name: tblName})
		if err != nil {
			return err
		}
		if len(dbKeys) > 0 {
			for _, dbkey := range dbKeys {
				var vniId uint32
				if strings.HasPrefix(niName, "Vlan") {
					vtepName = dbkey.Get(0)
					mapNameList := strings.Split(dbkey.Get(1), "_")
					vniNum, _ := strconv.ParseUint(mapNameList[1], 10, 32)
					vniId = uint32(vniNum)
				} else if strings.HasPrefix(niName, "Vrf") {
					vrfEntry, err := configDb.GetEntry(&db.TableSpec{Name: tblName}, db.Key{Comp: []string{niName}})
					if err != nil {
						return err
					}
					if vrfEntry.Has("vni") {
						vniIdStr = vrfEntry.Get("vni")
						vniNum, _ := strconv.ParseUint(vniIdStr, 10, 32)
						vniId = uint32(vniNum)

						vtepEntries, _ := configDb.GetKeys(&db.TableSpec{Name: "VXLAN_TUNNEL"})
						if len(vtepEntries) > 0 {
							vtepKey := vtepEntries[0]
							vtepName = vtepKey.Get(0)
						}
					} else {
						log.Errorf("Network instance %s not associated with vni %s", niName, vniIdStr)
						return tlerr.NotFound("Resource Not Found")
					}
				}
				vniInst, _ := reqP.vxlanNetInstObj.VxlanVniInstances.NewVniInstance(vniId, vtepName)

				fillVniInstanceDetails(niName, vniId, vtepName, vniInst)
			}
		} else {
			log.Errorf("Network instance %s not found", niName)
		}
	}

	return err
}

func fillVniInstanceDetails(niName string, vniId uint32, vtepName string, vniInstData *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_VxlanVniInstances_VniInstance) {
	if vniInstData == nil {
		return
	}

	ygot.BuildEmptyTree(vniInstData)

	vniInstData.VniId = &vniId
	vniInstData.SourceNve = &vtepName

	if vniInstData.Config != nil {
		vniInstData.Config.VniId = &vniId
		vniInstData.Config.SourceNve = &vtepName
	}

	if vniInstData.State != nil {
		vniInstData.State.VniId = &vniId
		vniInstData.State.SourceNve = &vtepName
	}
}

var YangToDb_vlan_nd_suppress_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	pathInfo := NewPathInfo(inParams.uri)
	vlanIdStr := pathInfo.Var("name")

	if !strings.HasPrefix(vlanIdStr, "Vlan") {
		return "", tlerr.InvalidArgs("Invalid key: %v", vlanIdStr)
	}
	return vlanIdStr, nil
}

var DbToYang_vlan_nd_suppress_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	res_map := make(map[string]interface{})

	log.Info("Vlan Name = ", inParams.key)
	res_map["name"] = inParams.key
	return res_map, nil
}

var YangToDb_vlan_nd_suppress_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)

	pathInfo := NewPathInfo(inParams.uri)
	vlanIdStr := pathInfo.Var("name")

	if !strings.HasPrefix(vlanIdStr, "Vlan") {
		return res_map, tlerr.InvalidArgs("Invalid key: %v", vlanIdStr)
	}
	log.Infof("YangToDb_vlan_nd_suppress_fld_xfmr: Params: %v", inParams.param)

	if inParams.param != nil {
		val, _ := inParams.param.(ocbinds.E_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_NeighbourSuppress_Config_ArpAndNdSuppress)
		if val == ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_NeighbourSuppress_Config_ArpAndNdSuppress_enable {
			res_map["suppress"] = "on"
		}
	}

	return res_map, nil
}

var DbToYang_vlan_nd_suppress_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	res_map := make(map[string]interface{})

	pathInfo := NewPathInfo(inParams.uri)
	vlanIdStr := pathInfo.Var("name")
	data := (*inParams.dbDataMap)[inParams.curDb]

	log.Infof("vlan_nd_suppress_fld_xfmr: key: %v, data: %v", vlanIdStr, data)
	if data != nil && len(data) > 0 {
		val := data["SUPPRESS_VLAN_NEIGH"][vlanIdStr]
		if val.Get("suppress") == "on" {
			res_map["arp-and-nd-suppress"], _ = ygot.EnumName(ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_NeighbourSuppress_Config_ArpAndNdSuppress_enable)
		}
	}

	return res_map, nil
}
