package transformer

import (
    "errors"
    "strings"
    log "github.com/golang/glog"
)


func init () {
    XlateFuncBind("YangToDb_ni_tbl_key_xfmr", YangToDb_ni_tbl_key_xfmr)
    XlateFuncBind("YangToDb_ni_proto_tbl_key_xfmr", YangToDb_ni_proto_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_tbl_key_xfmr", YangToDb_bgp_pgrp_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_tbl_key_xfmr", DbToYang_bgp_pgrp_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_peer_type_xfmr", YangToDb_bgp_pgrp_peer_type_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_peer_type_xfmr", DbToYang_bgp_pgrp_peer_type_xfmr)
}



var YangToDb_bgp_pgrp_peer_type_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    
    return res_map, nil
    
}

var DbToYang_bgp_pgrp_peer_type_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    result := make(map[string]interface{})


    return result, nil
}

var YangToDb_ni_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    log.Info("YangToDb_ni_tbl_key_xfmr: ***")
    var err error
    var vrfName string

    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, peer group name> */

    vrfName    =  pathInfo.Var("name")

    if len(pathInfo.Vars) <  1 {
        err = errors.New("Invalid Key length");
	return vrfName, err
    }

    if len(vrfName) == 0 {
        err = errors.New("vrf name is missing");
	    return vrfName, err
    }

    return vrfName, nil
}
var YangToDb_ni_proto_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_ni_proto_tbl_key_xfmr: ***")

    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, peer group name> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")

    if len(pathInfo.Vars) <  3 {
        err = errors.New("Invalid Key length");
	return vrfName, err
    }

    if len(vrfName) == 0 {
        err = errors.New("vrf name is missing");
        log.Info("vrf name is missing")
	return vrfName, err
    }
    if strings.Contains(bgpId,"BGP") == false {
        err = errors.New("BGP ID is missing");
        log.Info("BGP ID is missing")
	return bgpId, err
    }
    if len(protoName) == 0 {
        err = errors.New("Protocol Name is missing");
        log.Info("Protocol Name is missing")
	return protoName, err
    }

    var protoKey string

    protoKey = bgpId + "|" + protoName 

    log.Info("ProtoKey: ", protoKey)
    return protoKey, nil
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
    vrfName := pgrpKey[0]
    pgrpName:= pgrpKey[1]

    rmap["name"] = vrfName
    rmap["name#2"] = "BGP"
    rmap["peer-group-name"] = pgrpName

    return rmap, nil
}
