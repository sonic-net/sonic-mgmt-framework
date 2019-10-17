package transformer

import (
    "errors"
    "translib/ocbinds"
    "reflect"
    log "github.com/golang/glog"
)

func getBgpRoot (inParams XfmrParams) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp, error) {
    pathInfo := NewPathInfo(inParams.uri)
    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name1")
    var err error

    if len(niName) == 0 {
        return nil, errors.New("Network-instance-name missing")
    }

    if bgpId != "BGP" {
        return nil, errors.New("Network-instance-name missing")
    }

    if len(protoName) == 0 {
        return nil, errors.New("Network-instance-name missing")
    }

	deviceObj := (*inParams.ygRoot).(*ocbinds.Device)
    netInstsObj := deviceObj.NetworkInstances

    if netInstsObj.NetworkInstance == nil {
        return nil, errors.New("Network-instance-name missing")
    }

    netInstObj := netInstsObj.NetworkInstance[niName]
    if netInstObj == nil {
        return nil, errors.New("Network-instance-name missing")
    }

    if netInstObj.Protocols == nil || len(netInstObj.Protocols.Protocol) == 0 {
        return nil, errors.New("Network-instance-name missing")
    }

    var protoKey ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Key
    protoKey.Identifier = ocbinds.OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE_BGP
    protoKey.Name1 = protoName
    protoInstObj := netInstObj.Protocols.Protocol[protoKey]
    if protoInstObj == nil {
        return nil, errors.New("Network-instance-name missing")
    }
    return protoInstObj.Bgp, err
}


func init () {
    XlateFuncBind("YangToDb_bgp_gbl_tbl_key_xfmr", YangToDb_bgp_gbl_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_gbl_tbl_key_xfmr", DbToYang_bgp_gbl_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_always_compare_med_enable_xfmr", YangToDb_bgp_always_compare_med_enable_xfmr)
    XlateFuncBind("DbToYang_bgp_always_compare_med_enable_xfmr", DbToYang_bgp_always_compare_med_enable_xfmr)
    XlateFuncBind("YangToDb_bgp_allow_multiple_as_xfmr", YangToDb_bgp_allow_multiple_as_xfmr)
    XlateFuncBind("DbToYang_bgp_allow_multiple_as_xfmr", DbToYang_bgp_allow_multiple_as_xfmr)
    XlateFuncBind("YangToDb_bgp_graceful_restart_status_xfmr", YangToDb_bgp_graceful_restart_status_xfmr)
    XlateFuncBind("DbToYang_bgp_graceful_restart_status_xfmr", DbToYang_bgp_graceful_restart_status_xfmr)
    XlateFuncBind("YangToDb_bgp_ignore_as_path_length_xfmr", YangToDb_bgp_ignore_as_path_length_xfmr)
    XlateFuncBind("DbToYang_bgp_ignore_as_path_length_xfmr", DbToYang_bgp_ignore_as_path_length_xfmr)
    XlateFuncBind("YangToDb_bgp_external_compare_router_id_xfmr", YangToDb_bgp_external_compare_router_id_xfmr)
    XlateFuncBind("DbToYang_bgp_external_compare_router_id_xfmr", DbToYang_bgp_external_compare_router_id_xfmr)
}

var YangToDb_bgp_gbl_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    /* @@TODO Make sure name is vrf-name instead of BGP protocol name in the URI */
    niName := pathInfo.Var("name")

    /* @@TODO Return error for protocols other than BGP here */
    log.Info("URI VRF ", niName)

    return niName, err
}

var DbToYang_bgp_gbl_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_tbl_key: ", entry_key)

    rmap["name"] = entry_key
    return rmap, err
}

var YangToDb_bgp_always_compare_med_enable_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_always_compare_med_enable_xfmr Entry - ", reflect.ValueOf(inParams.param), "Type of : ", reflect.TypeOf(inParams.param));
    enabled, _ := inParams.param.(*bool)
    var enStr string
    if *enabled == true {
        enStr = "true"
    } else {
        enStr = "false"
    }
    res_map["always_compare_med"] = enStr

    return res_map, nil
}

var DbToYang_bgp_always_compare_med_enable_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_always_compare_med_enable_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_always_compare_med_enable_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    always_compare_med_enable, ok := niInst.Field["always_compare_med"]
    if ok {
        if always_compare_med_enable == "true" {
            result["always-compare-med"] = true
        } else {
            result["always-compare-med"] = false
        }
    } else {
        log.Info("always_compare_med field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_allow_multiple_as_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_allow_multiple_as_xfmr Entry - ", reflect.ValueOf(inParams.param), "Type of : ", reflect.TypeOf(inParams.param));
    allow_multiple_as, _ := inParams.param.(*bool)
    var allowMultipleAsStr string
    if *allow_multiple_as == true {
        allowMultipleAsStr = "true"
    } else {
        allowMultipleAsStr = "false"
    }
    res_map["load_balance_mp_relax"] = allowMultipleAsStr

    return res_map, nil
}

var DbToYang_bgp_allow_multiple_as_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_allow_multiple_as_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_allow_multiple_as_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    load_balance_mp_relax_val, ok := niInst.Field["load_balance_mp_relax"]
    if ok {
        if load_balance_mp_relax_val == "true" {
            result["load_balance_mp_relax"] = true
        } else {
            result["load_balance_mp_relax"] = false
        }
    } else {
        log.Info("load_balance_mp_relax field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_graceful_restart_status_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_graceful_restart_status_xfmr Entry - ", reflect.ValueOf(inParams.param), "Type of : ", reflect.TypeOf(inParams.param));
    gr_status, _ := inParams.param.(*bool)
    var gr_statusStr string
    if *gr_status == true {
        gr_statusStr = "true"
    } else {
        gr_statusStr = "false"
    }
    res_map["graceful_restart_enable"] = gr_statusStr

    return res_map, nil
}

var DbToYang_bgp_graceful_restart_status_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_graceful_restart_status_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_graceful_restart_status_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    gr_enable_val, ok := niInst.Field["graceful_restart_enable"]
    if ok {
        if gr_enable_val == "true" {
            result["graceful_restart_enable"] = true
        } else {
            result["graceful_restart_enable"] = false
        }
    } else {
        log.Info("graceful_restart_enable field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_ignore_as_path_length_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_ignore_as_path_length_xfmr Entry - ", reflect.ValueOf(inParams.param), "Type of : ", reflect.TypeOf(inParams.param));
    ignore_as_path_length, _ := inParams.param.(*bool)
    var ignoreAsPathLen string
    if *ignore_as_path_length == true {
        ignoreAsPathLen = "true"
    } else {
        ignoreAsPathLen = "false"
    }
    res_map["ignore_as_path_length"] = ignoreAsPathLen

    return res_map, nil
}

var DbToYang_bgp_ignore_as_path_length_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_ignore_as_path_length_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_ignore_as_path_length_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    ignore_as_path_length_val, ok := niInst.Field["ignore_as_path_length"]
    if ok {
        if ignore_as_path_length_val == "true" {
            result["ignore_as_path_length"] = true
        } else {
            result["ignore_as_path_length"] = false
        }
    } else {
        log.Info("ignore_as_path_length field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_external_compare_router_id_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_external_compare_router_id_xfmr Entry - ", reflect.ValueOf(inParams.param), "Type of : ", reflect.TypeOf(inParams.param));
    external_compare_router_id, _ := inParams.param.(*bool)
    var externalCompareRouterIdStr string
    if *external_compare_router_id == true {
        externalCompareRouterIdStr = "true"
    } else {
        externalCompareRouterIdStr = "false"
    }
    res_map["external_compare_router_id"] = externalCompareRouterIdStr

    return res_map, nil
}

var DbToYang_bgp_external_compare_router_id_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_external_compare_router_id_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_external_compare_router_id_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    external_compare_router_id_val, ok := niInst.Field["external_compare_router_id"]
    if ok {
        if external_compare_router_id_val == "true" {
            result["external_compare_router_id"] = true
        } else {
            result["external_compare_router_id"] = false
        }
    } else {
        log.Info("external_compare_router_id field not found in DB")
    }
    return result, err
}


