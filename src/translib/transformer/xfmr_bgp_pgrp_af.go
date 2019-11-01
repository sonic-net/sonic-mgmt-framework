package transformer

import (
    "errors"
    "strings"
    log "github.com/golang/glog"
)


func init () {
    XlateFuncBind("YangToDb_bgp_af_pgrp_tbl_key_xfmr", YangToDb_bgp_af_pgrp_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_af_pgrp_tbl_key_xfmr", DbToYang_bgp_af_pgrp_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_afi_safi_name_fld_xfmr", YangToDb_bgp_pgrp_afi_safi_name_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_afi_safi_name_fld_xfmr", DbToYang_bgp_pgrp_afi_safi_name_fld_xfmr)
}

var YangToDb_bgp_pgrp_afi_safi_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_bgp_pgrp_afi_safi_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_pgrp_afi_safi_name_fld_xfmr: ", data, "inParams : ", inParams)

    pTbl := data["BGP_PEER_GROUP_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_pgrp_afi_safi_name_fld_xfmr BGP peer-groups AF not found : ", inParams.key)
        return result, errors.New("BGP Peer Group AF Table not found : " + inParams.key)
    }

    pgrpAfKey := pTbl[inParams.key]
    afiSafiName, ok := pgrpAfKey.Field["afi_safi"]

    if ok {
        result["afi-safi-name"] = afiSafiName 
    } else {
        log.Info("afi_safi field not found in DB")
        err = errors.New("afi_safi field not found in DB")
    }

    return result, err
}


var YangToDb_bgp_af_pgrp_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_af_pgrp_tbl_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, peer group name> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    pGrpName   := pathInfo.Var("peer-group-name")
    afName     := pathInfo.Var("afi-safi-name")

    if len(pathInfo.Vars) <  4 {
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

    if len(afName) == 0 {
        err = errors.New("AFI SAFI is missing")
        log.Info("AFI SAFI is Missing")
        return pGrpName, err
    }

    if strings.Contains(afName, "IPV4_UNICAST") {
        afName = "IPV4_UNICAST"
    } else if strings.Contains(afName, "IPV6_UNICAST") { 
        afName = "IPV6_UNICAST"
    } else if strings.Contains(afName, "L2VPN_EVPN") {
        afName = "L2VPN_EVPN"
    } else  {
	err = errors.New("Unsupported AFI SAFI")
	log.Info("Unsupported AFI SAFI ", afName);
	return afName, err
    }

    log.Info("URI VRF ", vrfName)
    log.Info("URI Peer Group ", pGrpName)
    log.Info("URI AFI SAFI ", afName)

    var afPgrpKey string

    afPgrpKey = vrfName + "|" + afName + "|" + pGrpName 

    log.Info("YangToDb_bgp_af_pgrp_tbl_key_xfmr: afPgrpKey:", afPgrpKey)
    return afPgrpKey, nil
}

var DbToYang_bgp_af_pgrp_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_af_pgrp_tbl_key: ", entry_key)

    afPgrpKey := strings.Split(entry_key, "|")
    vrfName := afPgrpKey[0]
    afName  := afPgrpKey[1]
    pgrpName:= afPgrpKey[2]

    rmap["name"] = vrfName
    rmap["name#2"] = "BGP"
    rmap["peer-group-name"] = pgrpName
    rmap["afi-safi-name"]   = afName

    return rmap, nil
}

