package transformer

import (
    "errors"
    "strings"
    log "github.com/golang/glog"
)


func init () {
    XlateFuncBind("YangToDb_bgp_nbr_tbl_key_xfmr", YangToDb_bgp_nbr_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_tbl_key_xfmr", DbToYang_bgp_nbr_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_nbr_pgrp_name_fld_xfmr", YangToDb_bgp_nbr_pgrp_name_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_pgrp_name_fld_xfmr", DbToYang_bgp_nbr_pgrp_name_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_af_nbr_tbl_key_xfmr", YangToDb_bgp_af_nbr_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_af_nbr_tbl_key_xfmr", DbToYang_bgp_af_nbr_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_nbr_afi_safi_name_fld_xfmr", YangToDb_bgp_nbr_afi_safi_name_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_afi_safi_name_fld_xfmr", DbToYang_bgp_nbr_afi_safi_name_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_af_nbr_proto_tbl_key_xfmr", YangToDb_bgp_af_nbr_proto_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_af_nbr_proto_tbl_key_xfmr", DbToYang_bgp_af_nbr_proto_tbl_key_xfmr)
}

var YangToDb_bgp_nbr_pgrp_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_bgp_nbr_pgrp_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_pgrp_name_fld_xfmr : ", data, "inParams : ", inParams)

    entry_key := inParams.key
    peer_group_Key := strings.Split(entry_key, "|")
    peer_group_name:= peer_group_Key[1]
    result["peer-group"] = peer_group_name

    return result, err 
}

var YangToDb_bgp_nbr_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_nbr_tbl_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, peer group name> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    pNbrAddr   := pathInfo.Var("neighbor-address")

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
    if len(pNbrAddr) == 0 {
        err = errors.New("Neighbor address  is missing")
        log.Info("Neighbor address is Missing")
        return pNbrAddr, err
    }

    log.Info("URI VRF", vrfName)
    log.Info("URI Neighbor address", pNbrAddr)

    var pNbrKey string

    pNbrKey = vrfName + "|" + pNbrAddr

    log.Info("YangToDb_bgp_nbr_tbl_key_xfmr: pGrpKey:", pNbrKey)
    return pNbrKey, nil
}

var DbToYang_bgp_nbr_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_nbr_tbl_key: ", entry_key)

    nbrKey := strings.Split(entry_key, "|")
    nbrName:= nbrKey[1]

    rmap["neighbor-address"] = nbrName

    return rmap, nil
}


var YangToDb_bgp_nbr_afi_safi_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_bgp_nbr_afi_safi_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    entry_key := inParams.key
    nbrAfKey := strings.Split(entry_key, "|")
    nbrAfName:= nbrAfKey[2]

    result["afi-safi-name"] = nbrAfName

    return result, err
}


var YangToDb_bgp_af_nbr_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_af_nbr_tbl_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, peer group name> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    pNbr   := pathInfo.Var("neighbor-address")
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
    if len(pNbr) == 0 {
        err = errors.New("Neighbor is missing")
        log.Info("Neighbor is Missing")
        return pNbr, err
    }

    if len(afName) == 0 {
        err = errors.New("AFI SAFI is missing")
        log.Info("AFI SAFI is Missing")
        return afName, err
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
    log.Info("URI Nbr ", pNbr)
    log.Info("URI AFI SAFI ", afName)

    var nbrAfKey string

    nbrAfKey = vrfName + "|" + pNbr + "|" + afName

    log.Info("YangToDb_bgp_af_nbr_tbl_key_xfmr: afPgrpKey:", nbrAfKey)
    return nbrAfKey, nil
}

var DbToYang_bgp_af_nbr_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_af_nbr_tbl_key: ", entry_key)

    nbrAfKey := strings.Split(entry_key, "|")
    afName  := nbrAfKey[2]

    rmap["afi-safi-name"]   = afName

    return rmap, nil
}

var YangToDb_bgp_af_nbr_proto_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_af_nbr_proto_tbl_key_xfmr***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    pNbr   := pathInfo.Var("neighbor-address")
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
    if len(pNbr) == 0 {
        err = errors.New("Neighbor missing")
        log.Info("Neighbo Missing")
        return pNbr, err
    }

    if len(afName) == 0 {
        err = errors.New("AFI SAFI is missing")
        log.Info("AFI SAFI is Missing")
        return afName, err
    }

    if strings.Contains(afName, "IPV4_UNICAST") {
        afName = "IPV4_UNICAST"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		err = errors.New("IPV4_UNICAST supported only on ipv4-config container")
		log.Info("IPV4_UNICAST supported only on ipv4-config container: ", afName);
		return afName, err
        }
    } else if strings.Contains(afName, "IPV6_UNICAST") { 
        afName = "IPV6_UNICAST"
        if strings.Contains(inParams.uri, "ipv4-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		err = errors.New("IPV6_UNICAST supported only on ipv6-config container")
		log.Info("IPV6_UNICAST supported only on ipv6-config container: ", afName);
		return afName, err
        }
    } else if strings.Contains(afName, "L2VPN_EVPN") {
        afName = "L2VPN_EVPN"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "ipv4-unicast") {
		err = errors.New("L2VPN_EVPN supported only on l2vpn-evpn container")
		log.Info("L2VPN_EVPN supported only on l2vpn-evpn container: ", afName);
		return afName, err
        }
    } else  {
	err = errors.New("Unsupported AFI SAFI")
	log.Info("Unsupported AFI SAFI ", afName);
	return afName, err
    }

    log.Info("URI VRF ", vrfName)
    log.Info("URI Nbr ", pNbr)
    log.Info("URI AFI SAFI ", afName)

    var nbrAfKey string

  nbrAfKey = vrfName + "|" + pNbr + "|" + afName

    log.Info("YangToDb_bgp_af_nbr_proto_tbl_key_xfmr: nbrAfKey:", nbrAfKey)
    return nbrAfKey, nil
}

var DbToYang_bgp_af_nbr_proto_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_af_nbr_proto_tbl_key_xfmr: ", entry_key)

    nbrAfKey := strings.Split(entry_key, "|")
    afName  := nbrAfKey[2]

    rmap["afi-safi-name"]   = afName

    return rmap, nil
}
