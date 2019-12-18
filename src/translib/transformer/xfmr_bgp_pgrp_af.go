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
    XlateFuncBind("YangToDb_bgp_pgrp_plist_direction_fld_xfmr", YangToDb_bgp_pgrp_plist_direction_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_plist_direction_fld_xfmr", DbToYang_bgp_pgrp_plist_direction_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_flist_direction_fld_xfmr", YangToDb_bgp_pgrp_flist_direction_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_flist_direction_fld_xfmr", DbToYang_bgp_pgrp_flist_direction_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_pgrp_orf_type_fld_xfmr", YangToDb_bgp_pgrp_orf_type_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_pgrp_orf_type_fld_xfmr", DbToYang_bgp_pgrp_orf_type_fld_xfmr)
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
    community_type, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_CommunityType)
    log.Info("YangToDb_bgp_pgrp_community_type_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " community_type: ", community_type)

    if (community_type == ocbinds.OpenconfigBgpExt_CommunityType_STANDARD) {
        res_map["send_community"] = "standard"
    }  else if (community_type == ocbinds.OpenconfigBgpExt_CommunityType_EXTENDED) {
        res_map["send_community"] = "extended"
    }  else if (community_type == ocbinds.OpenconfigBgpExt_CommunityType_BOTH) {
        res_map["send_community"] = "both"
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
        }
    } else {
        log.Info("send_community not found in DB")
    }
    return result, err
}

var YangToDb_bgp_pgrp_plist_direction_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    direction, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_BgpDirection)
    log.Info("YangToDb_bgp_pgrp_plist_direction_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " direction: ", direction)

    if (direction == ocbinds.OpenconfigBgpExt_BgpDirection_INBOUND) {
        res_map["prefix_list_direction"] = "in"
    }  else if (direction == ocbinds.OpenconfigBgpExt_BgpDirection_OUTBOUND) {
        res_map["prefix_list_direction"] = "out"
    } else {
        err = errors.New("direction Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_pgrp_plist_direction_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_pgrp_plist_direction_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_PEER_GROUP_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_pgrp_plist_direction_fld_xfmr BGP peer group not found : ", inParams.key)
        return result, errors.New("BGP peer group not found : " + inParams.key)
    }
    pGrpKey := pTbl[inParams.key]
    direction, ok := pGrpKey.Field["prefix_list_direction"]

    if ok {
        if (direction == "in") {
            result["direction"] = "INBOUND"
        } else if (direction == "out") {
            result["direction"] = "OUTBOUND"
        }
    } else {
        log.Info("prefix_list_direction field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_pgrp_flist_direction_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    direction, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_BgpDirection)
    log.Info("YangToDb_bgp_pgrp_flist_direction_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " direction: ", direction)

    if (direction == ocbinds.OpenconfigBgpExt_BgpDirection_INBOUND) {
        res_map["filter_list_direction"] = "in"
    }  else if (direction == ocbinds.OpenconfigBgpExt_BgpDirection_OUTBOUND) {
        res_map["filter_list_direction"] = "out"
    } else {
        err = errors.New("direction Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_pgrp_flist_direction_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_pgrp_flist_direction_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_PEER_GROUP_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_pgrp_flist_direction_fld_xfmr BGP peer group not found : ", inParams.key)
        return result, errors.New("BGP peer group not found : " + inParams.key)
    }
    pGrpKey := pTbl[inParams.key]
    direction, ok := pGrpKey.Field["filter_list_direction"]

    if ok {
        if (direction == "in") {
            result["direction"] = "INBOUND"
        }else if (direction == "out") {
            result["direction"] = "OUTBOUND"
        }
    } else {
        log.Info("filter_list_direction field not found in DB")
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
