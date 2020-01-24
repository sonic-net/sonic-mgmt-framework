package transformer

import (
    "errors"
    "strings"
    "translib/ocbinds"
    "github.com/openconfig/ygot/ygot"
    log "github.com/golang/glog"
)


func init () {
    XlateFuncBind("YangToDb_bgp_pgrp_tbl_key_xfmr", YangToDb_bgp_pgrp_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_tbl_key_xfmr", DbToYang_bgp_pgrp_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_peer_type_fld_xfmr", YangToDb_bgp_pgrp_peer_type_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_peer_type_fld_xfmr", DbToYang_bgp_pgrp_peer_type_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_name_fld_xfmr", YangToDb_bgp_pgrp_name_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_name_fld_xfmr", DbToYang_bgp_pgrp_name_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_peer_group_mbrs_state_xfmr", DbToYang_bgp_peer_group_mbrs_state_xfmr)
}

var YangToDb_bgp_pgrp_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_bgp_pgrp_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_pgrp_name_fld_xfmr : ", data, "inParams : ", inParams)

    entry_key := inParams.key
    peer_group_Key := strings.Split(entry_key, "|")
    peer_group_name:= peer_group_Key[1]
    result["peer-group-name"] = peer_group_name

    return result, err
}

var YangToDb_bgp_pgrp_peer_type_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    if inParams.oper == DELETE {
        res_map["peer_type"] = ""
        return res_map, nil
    }

    peer_type, _ := inParams.param.(ocbinds.E_OpenconfigBgp_PeerType)
    log.Info("YangToDb_bgp_pgrp_peer_type_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " peer-type: ", peer_type)

    if (peer_type == ocbinds.OpenconfigBgp_PeerType_INTERNAL) {
        res_map["peer_type"] = "internal"
    }  else if (peer_type == ocbinds.OpenconfigBgp_PeerType_EXTERNAL) {
        res_map["peer_type"] = "external"
    } else {
        err = errors.New("Peer Type Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_pgrp_peer_type_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_pgrp_peer_type_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_PEER_GROUP"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_pgrp_peer_type_fld_xfmr BGP peer-groups not found : ", inParams.key)
        return result, errors.New("BGP peer-groups not found : " + inParams.key)
    }
    pGrpKey := pTbl[inParams.key]
    peer_type, ok := pGrpKey.Field["peer_type"]

    if ok {
        if (peer_type == "internal") {
            result["peer-type"] = "INTERNAL" 
        } else if (peer_type == "external") {
            result["peer-type"] = "EXTERNAL"
        }
    } else {
        log.Info("peer_type field not found in DB")
    }
    return result, err
}


var YangToDb_bgp_pgrp_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_pgrp_tbl_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, peer group name> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    pGrpName   := pathInfo.Var("peer-group-name")

    if len(pathInfo.Vars) <  3 {
        err = errors.New("Invalid Key length");
        log.Info("Invalid Key length", len(pathInfo.Vars))
        return vrfName, err
    }

    if len(vrfName) == 0 {
        err = errors.New("vrf name is missing");
        log.Info("VRF Name is Missing")
        return vrfName, err
    }
    if strings.Contains(bgpId,"BGP") == false {
        err = errors.New("BGP ID is missing");
        log.Info("BGP ID is missing")
        return bgpId, err
    }
    if len(protoName) == 0 {
        err = errors.New("Protocol Name is missing");
        log.Info("Protocol Name is Missing")
        return protoName, err
    }
    if len(pGrpName) == 0 {
        err = errors.New("Peer Group Name is missing")
        log.Info("Peer Group Name is Missing")
        return pGrpName, err
    }

    log.Info("URI VRF", vrfName)
    log.Info("URI Peer Group", pGrpName)

    var pGrpKey string

    pGrpKey = vrfName + "|" + pGrpName

    log.Info("YangToDb_bgp_pgrp_tbl_key_xfmr: pGrpKey:", pGrpKey)
    return pGrpKey, nil
}

var DbToYang_bgp_pgrp_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_pgrp_tbl_key: ", entry_key)

    pgrpKey := strings.Split(entry_key, "|")
    pgrpName:= pgrpKey[1]

    rmap["peer-group-name"] = pgrpName

    return rmap, nil
}


type _xfmr_bgp_pgrp_state_key struct {
    niName string
    pgrp string
}

func validate_pgrp_state_get (inParams XfmrParams, dbg_log string) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_PeerGroups_PeerGroup, _xfmr_bgp_pgrp_state_key, error) {
    var err error
    oper_err := errors.New("Opertational error")
    var pgrp_key _xfmr_bgp_pgrp_state_key
    var bgp_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp

    bgp_obj, pgrp_key.niName, err = getBgpRoot (inParams)
    if err != nil {
        log.Errorf ("%s failed !! Error:%s", dbg_log , err);
        return nil, pgrp_key, err
    }

    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, _ := getYangPathFromUri(pathInfo.Path)
    pgrp_key.pgrp = pathInfo.Var("peer-group-name")
    log.Info("%s : path:%s; template:%s targetUriPath:%s niName:%s peer group:%s",
              dbg_log, pathInfo.Path, pathInfo.Template, targetUriPath, pgrp_key.niName, pgrp_key.pgrp)

    if pgrp_key.niName == "default" {
       pgrp_key.niName = ""
    }

    pgrps_obj := bgp_obj.PeerGroups
    if pgrps_obj == nil {
        log.Errorf("%s failed !! Error: Peer groups container missing", dbg_log)
        return nil, pgrp_key, oper_err
    }

    pgrp_obj, ok := pgrps_obj.PeerGroup[pgrp_key.pgrp]
    if !ok {
        log.Info("%s Peer group object missing, add new", dbg_log)
        pgrp_obj,_ = pgrps_obj.NewPeerGroup(pgrp_key.pgrp)
    }
    ygot.BuildEmptyTree(pgrp_obj)
    return pgrp_obj, pgrp_key, err
}

func fill_pgrp_state_info (pgrp_key *_xfmr_bgp_pgrp_state_key, frrPgrpDataValue interface{},
                              pgrp_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_PeerGroups_PeerGroup) error {
    var err error
    var pMember ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_PeerGroups_PeerGroup_MembersState
    pgrp_obj.MembersState = &pMember

    frrPgrpDataJson := frrPgrpDataValue.(map[string]interface{})

    if frrPgrpDataJson == nil {
        log.Info("peer group json Data NIL ")
        return err
    }

    if peerGroupMembers,  ok := frrPgrpDataJson["peerGroupMembers"].(map[string]interface{}) ; ok {
        for pgMem,_ := range peerGroupMembers {
            member, ok := pMember.Member[pgMem]
            if !ok {
                member, _ = pMember.NewMember(pgMem)
            }
            temp, ok := peerGroupMembers[pgMem].(map[string]interface{})
            if  ok {
                if value, ok := temp["peerStatus"].(string); ok {
                    member.State = &value
                    log.Info("peer group member status ", member.State)
                }
                if value, ok := temp["isDynamic"].(bool); ok {
                    member.Dynamic = &value
                    log.Info("peer group member Dynamic ", member.Dynamic)
                }
            }
        }
    }

    return err
}

func get_specific_pgrp_state (pgrp_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_PeerGroups_PeerGroup,
                             pgrp_key *_xfmr_bgp_pgrp_state_key) error {
    var err error
    var vtysh_cmd string
    if pgrp_key.niName == "" {
       vtysh_cmd = "show ip bgp peer-group " + pgrp_key.pgrp + " json"
    } else {
       vtysh_cmd = "show ip bgp vrf " + pgrp_key.niName + " peer-group " + pgrp_key.pgrp + " json"
    }
    pgrpMapJson, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bgp peer group member state info for niName:%s peer group :%s. Err: %s\n", pgrp_key.niName, pgrp_key.pgrp, cmd_err)
        return cmd_err
    }

    if frrPgrpDataJson, ok := pgrpMapJson[pgrp_key.pgrp].(map[string]interface{}) ; ok {
        err = fill_pgrp_state_info (pgrp_key, frrPgrpDataJson, pgrp_obj)
    }

    return err
}

var DbToYang_bgp_peer_group_mbrs_state_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error
    cmn_log := "GET: xfmr for BGP Peer Group members state"

    pgrp_obj, pgrp_key, get_err := validate_pgrp_state_get (inParams, cmn_log);
    if get_err != nil {
        log.Info("Peer Group members state get subtree error: ", get_err)
        return get_err
    }

    err = get_specific_pgrp_state (pgrp_obj, &pgrp_key)
    return err;
}
