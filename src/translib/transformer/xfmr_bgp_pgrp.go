package transformer

import (
    "errors"
    "strings"
    "translib/ocbinds"
    log "github.com/golang/glog"
)


func init () {
    XlateFuncBind("YangToDb_bgp_pgrp_tbl_key_xfmr", YangToDb_bgp_pgrp_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_tbl_key_xfmr", DbToYang_bgp_pgrp_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_peer_type_fld_xfmr", YangToDb_bgp_pgrp_peer_type_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_peer_type_fld_xfmr", DbToYang_bgp_pgrp_peer_type_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_name_fld_xfmr", YangToDb_bgp_pgrp_name_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_name_fld_xfmr", DbToYang_bgp_pgrp_name_fld_xfmr)
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
