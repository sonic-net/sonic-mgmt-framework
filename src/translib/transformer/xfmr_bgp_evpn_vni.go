package transformer

import (
    "errors"
    "strings"
    //"translib/ocbinds"
    log "github.com/golang/glog"
)


func init () {
    XlateFuncBind("YangToDb_bgp_evpn_vni_key_xfmr", YangToDb_bgp_evpn_vni_key_xfmr)
    XlateFuncBind("DbToYang_bgp_evpn_vni_key_xfmr", DbToYang_bgp_evpn_vni_key_xfmr)
    XlateFuncBind("YangToDb_bgp_vni_number_fld_xfmr", YangToDb_bgp_vni_number_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_vni_number_fld_xfmr", DbToYang_bgp_vni_number_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_evpn_rt_key_xfmr", YangToDb_bgp_evpn_rt_key_xfmr)
    XlateFuncBind("DbToYang_bgp_evpn_rt_key_xfmr", DbToYang_bgp_evpn_rt_key_xfmr)
    XlateFuncBind("YangToDb_bgp_rt_fld_xfmr", YangToDb_bgp_rt_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_rt_fld_xfmr", DbToYang_bgp_rt_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_rt_type_fld_xfmr", YangToDb_bgp_rt_type_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_rt_type_fld_xfmr", DbToYang_bgp_rt_type_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_evpn_vni_rt_key_xfmr", YangToDb_bgp_evpn_vni_rt_key_xfmr)
    XlateFuncBind("DbToYang_bgp_evpn_vni_rt_key_xfmr", DbToYang_bgp_evpn_vni_rt_key_xfmr)
}


var YangToDb_bgp_evpn_vni_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_evpn_vni_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, afi-safi name, vni-number> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    vniNumber   := pathInfo.Var("vni-number")
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
    if len(vniNumber) == 0 {
        err = errors.New("VNI number is missing")
        log.Info("VNI number is Missing")
        return vniNumber, err
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
    log.Info("URI VNI NUMBER ", vniNumber)
    log.Info("URI AFI SAFI ", afName)

    var vniTableKey string

    vniTableKey = vrfName + "|" + afName + "|" + vniNumber

    log.Info("YangToDb_bgp_evpn_vni_key_xfmr: vniTableKey:", vniTableKey)
    return vniTableKey, nil
}

var DbToYang_bgp_evpn_vni_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_evpn_vni_key_xfmr: ", entry_key)

    afPgrpKey := strings.Split(entry_key, "|")
    afName  := afPgrpKey[1]

    rmap["afi-safi-name"]   = afName

    return rmap, nil
}

var YangToDb_bgp_vni_number_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_bgp_vni_number_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    entry_key := inParams.key
    vniKey := strings.Split(entry_key, "|")
    vniNumber:= vniKey[2]

    result["vni-number"] = vniNumber

    return result, err
}






var YangToDb_bgp_evpn_rt_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_evpn_rt_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, afi-safi name, route-target> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    routeTarget   := pathInfo.Var("route-target")
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
    if len(routeTarget) == 0 {
        err = errors.New("routeTarget is missing")
        log.Info("routeTarget is Missing")
        return routeTarget, err
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
    log.Info("URI Route Target ", routeTarget)
    log.Info("URI AFI SAFI ", afName)

    var routeTargetKey string

    routeTargetKey = vrfName + "|" + afName + "|" + routeTarget

    log.Info("YangToDb_bgp_evpn_rt_key_xfmr: routeTargetKey:", routeTargetKey)
    return routeTargetKey, nil
}

var DbToYang_bgp_evpn_rt_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_evpn_rt_key_xfmr: ", entry_key)

    routeTargetKey := strings.Split(entry_key, "|")
    routeTarget  := routeTargetKey[1]

    rmap["route-target"]   = routeTarget

    return rmap, nil
}

var YangToDb_bgp_rt_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_bgp_rt_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    entry_key := inParams.key
    routeTargetKey := strings.Split(entry_key, "|")
    routeTarget:= routeTargetKey[2]

    result["route-target"] = routeTarget

    return result, err
}


var YangToDb_bgp_rt_type_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    /*TODO: Fix this to fill correct value*/
    res_map["route-target-type"] = "import"//ocbinds.E_IETFRoutingTypes_RouteTargetType
    return res_map, nil
}

var DbToYang_bgp_rt_type_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    /*TODO: Fill this with correct value*/
    result["route-target-type"] = "import"

    return result, err
}





var YangToDb_bgp_evpn_vni_rt_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_evpn_vni_rt_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, afi-safo name, vni number, route-target> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    vniNumber   := pathInfo.Var("vni-number")
    afName     := pathInfo.Var("afi-safi-name")
    routeTarget   := pathInfo.Var("route-target")

    if len(pathInfo.Vars) <  5 {
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
    if len(vniNumber) == 0 {
        err = errors.New("vniNumber is missing")
        log.Info("vniNumber is Missing")
        return vniNumber, err
    }

    if len(afName) == 0 {
        err = errors.New("AFI SAFI is missing")
        log.Info("AFI SAFI is Missing")
        return afName, err
    }

    if len(routeTarget) == 0 {
        err = errors.New("Route-target is missing")
        log.Info("route-target is Missing")
        return routeTarget, err
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
    log.Info("URI VNI NUMBER ", vniNumber)
    log.Info("URI AFI SAFI ", afName)
    log.Info("URI Route-target ", routeTarget)

    var vniRouteTargetKey string

    vniRouteTargetKey = vrfName + "|" + afName + "|" + vniNumber + "|" + routeTarget

    log.Info("YangToDb_bgp_evpn_vni_rt_key_xfmr: vniRouteTargetKey:", vniRouteTargetKey)
    return vniRouteTargetKey, nil
}

var DbToYang_bgp_evpn_vni_rt_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_evpn_vni_rt_key_xfmr: ", entry_key)

    vniRouteTargetKey := strings.Split(entry_key, "|")
    routeTarget  := vniRouteTargetKey[1]

    rmap["route-target"]   = routeTarget

    return rmap, nil
}