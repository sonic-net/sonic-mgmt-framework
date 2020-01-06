package transformer

import (
    "errors"
    "strings"
    "translib/ocbinds"
    log "github.com/golang/glog"
)


func init () {
    XlateFuncBind("YangToDb_bgp_af_pgrp_tbl_key_xfmr", YangToDb_bgp_af_pgrp_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_af_pgrp_tbl_key_xfmr", DbToYang_bgp_af_pgrp_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_afi_safi_name_fld_xfmr", YangToDb_bgp_pgrp_afi_safi_name_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_afi_safi_name_fld_xfmr", DbToYang_bgp_pgrp_afi_safi_name_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_af_pgrp_proto_tbl_key_xfmr", YangToDb_bgp_af_pgrp_proto_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_af_pgrp_proto_tbl_key_xfmr", DbToYang_bgp_af_pgrp_proto_tbl_key_xfmr)

    XlateFuncBind("YangToDb_bgp_pgrp_community_type_fld_xfmr", YangToDb_bgp_pgrp_community_type_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_community_type_fld_xfmr", DbToYang_bgp_pgrp_community_type_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_orf_type_fld_xfmr", YangToDb_bgp_pgrp_orf_type_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_orf_type_fld_xfmr", DbToYang_bgp_pgrp_orf_type_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_tx_add_paths_fld_xfmr", YangToDb_bgp_pgrp_tx_add_paths_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_tx_add_paths_fld_xfmr", DbToYang_bgp_pgrp_tx_add_paths_fld_xfmr)
}

var YangToDb_bgp_pgrp_afi_safi_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_bgp_pgrp_afi_safi_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    entry_key := inParams.key
    pgrpAfKey := strings.Split(entry_key, "|")
	pgrpAfName := ""

	switch pgrpAfKey[2] {
	case "ipv4_unicast":
		pgrpAfName = "IPV4_UNICAST"
	case "ipv6_unicast":
		pgrpAfName = "IPV6_UNICAST"
	case "l2vpn_evpn":
		pgrpAfName = "L2VPN_EVPN"
	}

    result["afi-safi-name"] = pgrpAfName

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
        afName = "ipv4_unicast"
    } else if strings.Contains(afName, "IPV6_UNICAST") { 
        afName = "ipv6_unicast"
    } else if strings.Contains(afName, "L2VPN_EVPN") {
        afName = "l2vpn_evpn"
    } else  {
	err = errors.New("Unsupported AFI SAFI")
	log.Info("Unsupported AFI SAFI ", afName);
	return afName, err
    }

    log.Info("URI VRF ", vrfName)
    log.Info("URI Peer Group ", pGrpName)
    log.Info("URI AFI SAFI ", afName)

    var afPgrpKey string

    afPgrpKey = vrfName + "|" + pGrpName + "|" + afName

    log.Info("YangToDb_bgp_af_pgrp_tbl_key_xfmr: afPgrpKey:", afPgrpKey)
    return afPgrpKey, nil
}

var DbToYang_bgp_af_pgrp_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_af_pgrp_tbl_key: ", entry_key)

    afPgrpKey := strings.Split(entry_key, "|")
	afName := ""

	switch afPgrpKey[2] {
	case "ipv4_unicast":
		afName = "IPV4_UNICAST"
	case "ipv6_unicast":
		afName = "IPV6_UNICAST"
	case "l2vpn_evpn":
		afName = "L2VPN_EVPN"
	}

    rmap["afi-safi-name"]   = afName

    return rmap, nil
}

var YangToDb_bgp_af_pgrp_proto_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_af_pgrp_proto_tbl_key_xfmr***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

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
        afName = "ipv4_unicast"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		err = errors.New("IPV4_UNICAST supported only on ipv4-config container")
		log.Info("IPV4_UNICAST supported only on ipv4-config container: ", afName);
		return afName, err
        }
    } else if strings.Contains(afName, "IPV6_UNICAST") { 
        afName = "ipv6_unicast"
        if strings.Contains(inParams.uri, "ipv4-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		err = errors.New("IPV6_UNICAST supported only on ipv6-config container")
		log.Info("IPV6_UNICAST supported only on ipv6-config container: ", afName);
		return afName, err
        }
    } else if strings.Contains(afName, "L2VPN_EVPN") {
        afName = "l2vpn_evpn"
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
    log.Info("URI Peer Group ", pGrpName)
    log.Info("URI AFI SAFI ", afName)

    var afPgrpKey string

    afPgrpKey = vrfName + "|" + pGrpName + "|" + afName

    log.Info("YangToDb_bgp_af_pgrp_tbl_key_xfmr: afPgrpKey:", afPgrpKey)
    return afPgrpKey, nil
}

var DbToYang_bgp_af_pgrp_proto_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_af_pgrp_proto_tbl_key_xfmr: ", entry_key)

    afPgrpKey := strings.Split(entry_key, "|")
	afName := ""

	switch afPgrpKey[2] {
	case "ipv4_unicast":
		afName = "IPV4_UNICAST"
	case "ipv6_unicast":
		afName = "IPV6_UNICAST"
	case "l2vpn_evpn":
		afName = "L2VPN_EVPN"
	}

    rmap["afi-safi-name"]   = afName

    return rmap, nil
}

var YangToDb_bgp_pgrp_community_type_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    community_type, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_BgpExtCommunityType)
    log.Info("YangToDb_bgp_pgrp_community_type_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " community_type: ", community_type)

    if (community_type == ocbinds.OpenconfigBgpExt_BgpExtCommunityType_STANDARD) {
        res_map["send_community"] = "standard"
    }  else if (community_type == ocbinds.OpenconfigBgpExt_BgpExtCommunityType_EXTENDED) {
        res_map["send_community"] = "extended"
    }  else if (community_type == ocbinds.OpenconfigBgpExt_BgpExtCommunityType_BOTH) {
        res_map["send_community"] = "both"
    }  else if (community_type == ocbinds.OpenconfigBgpExt_BgpExtCommunityType_NONE) {
        res_map["send_community"] = "none"
    }  else if (community_type == ocbinds.OpenconfigBgpExt_BgpExtCommunityType_LARGE) {
        res_map["send_community"] = "large"
    }  else if (community_type == ocbinds.OpenconfigBgpExt_BgpExtCommunityType_ALL) {
        res_map["send_community"] = "all"
    } else {
        err = errors.New("send_community  Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_pgrp_community_type_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_pgrp_community_type_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_PEER_GROUP_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_pgrp_community_type_fld_xfmr BGP Peer group not found : ", inParams.key)
        return result, errors.New("BGP peer group not found : " + inParams.key)
    }
    pGrpKey := pTbl[inParams.key]
    community_type, ok := pGrpKey.Field["send_community"]

    if ok {
        if (community_type == "standard") {
            result["send-community"] = "STANDARD"
        } else if (community_type == "extended") {
            result["send-community"] = "EXTENDED"
        } else if (community_type == "both") {
            result["send-community"] = "BOTH"
        } else if (community_type == "none") {
            result["send-community"] = "NONE"
        } else if (community_type == "large") {
            result["send-community"] = "LARGE"
        } else if (community_type == "all") {
            result["send-community"] = "ALL"
        }
    } else {
        log.Info("send_community not found in DB")
    }
    return result, err
}

var YangToDb_bgp_pgrp_orf_type_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    orf_type, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_BgpOrfType)
    log.Info("YangToDb_bgp_pgrp_orf_type_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " orf_type: ", orf_type)

    if (orf_type == ocbinds.OpenconfigBgpExt_BgpOrfType_SEND) {
        res_map["cap_orf"] = "send"
    }  else if (orf_type == ocbinds.OpenconfigBgpExt_BgpOrfType_RECEIVE) {
        res_map["cap_orf"] = "receive"
    }  else if (orf_type == ocbinds.OpenconfigBgpExt_BgpOrfType_BOTH) {
        res_map["cap_orf"] = "both"
    } else {
        err = errors.New("ORF type Missing");
        return res_map, err
    }

    return res_map, nil
}

var DbToYang_bgp_pgrp_orf_type_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_pgrp_orf_type_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_PEER_GROUP_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_pgrp_orf_type_fld_xfmr BGP PEER GROUP AF not found : ", inParams.key)
        return result, errors.New("BGP PEER GROUP AF not found : " + inParams.key)
    }
    pGrpKey := pTbl[inParams.key]
    orf_type, ok := pGrpKey.Field["cap_orf"]

    if ok {
        if (orf_type == "send") {
            result["orf-type"] = "SEND"
        } else if (orf_type == "receive") {
            result["orf-type"] = "RECEIVE"
        } else if (orf_type == "both") {
            result["orf-type"] = "BOTH"
        }
    } else {
        log.Info("cap_orf_direction field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_pgrp_tx_add_paths_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    tx_add_paths_type, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_TxAddPathsType)
    log.Info("YangToDb_pgrp_tx_add_paths_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " add-paths-type: ", tx_add_paths_type)

    if (tx_add_paths_type == ocbinds.OpenconfigBgpExt_TxAddPathsType_TX_ALL_PATHS) {
        res_map["tx_add_paths"] = "tx_all_paths"
    }  else if (tx_add_paths_type == ocbinds.OpenconfigBgpExt_TxAddPathsType_TX_BEST_PATH_PER_AS) {
        res_map["tx_add_paths"] = "tx_best_path_per_as"
    } else {
        err = errors.New("Invalid add Paths type Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_pgrp_tx_add_paths_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_pgrp_tx_add_paths_fld_xfmr: ", data, "inParams : ", inParams)

    pTbl := data["BGP_PEER_GROUP_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_pgrp_tx_add_paths_fld_xfmr BGP peer group not found : ", inParams.key)
        return result, errors.New("BGP neighbor not found : " + inParams.key)
    }
    pNbrKey := pTbl[inParams.key]
    tx_add_paths_type, ok := pNbrKey.Field["tx_add_paths"]

    if ok {
        if (tx_add_paths_type == "tx_all_paths") {
            result["tx-add-paths"] = "TX_ALL_PATHS"
        } else if (tx_add_paths_type == "tx_best_path_per_as") {
            result["tx-add-paths"] = "TX_BEST_PATH_PER_AS"
        }
    } else {
        log.Info("Tx add Paths field not found in DB")
    }
    return result, err
}



