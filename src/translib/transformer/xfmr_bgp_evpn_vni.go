////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package transformer

import (
    "errors"
    "strings"
    "translib/ocbinds"
    "translib/db"
    "strconv"
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
    XlateFuncBind("YangToDb_bgp_advertise_fld_xfmr", YangToDb_bgp_advertise_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_advertise_fld_xfmr", DbToYang_bgp_advertise_fld_xfmr)

    XlateFuncBind("DbToYang_bgp_evpn_vni_state_xfmr", DbToYang_bgp_evpn_vni_state_xfmr)
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


var YangToDb_bgp_advertise_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    afi_safi_list, _ := inParams.param.([]ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE)
    log.Info("YangToDb_bgp_advertise_fld_xfmr: afi_safi_list:", afi_safi_list)

    for _, afi_safi := range afi_safi_list {

        if (afi_safi == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST) {
            res_map["advertise-ipv4-unicast"] = "true"
        }  else if (afi_safi == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST) {
            res_map["advertise-ipv6-unicast"] = "true"
        } else {
            err = errors.New("Unsupported afi_safi");
            return res_map, err
        }
    }
    return res_map, nil
}

var DbToYang_bgp_advertise_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    /*
    entry_key := inParams.key
    afiSafiKey := strings.Split(entry_key, "|")
    afiSafi:= afiSafiKey[2]

    result["advertise-list"] = afiSafi
*/
    return result, err
}

var DbToYang_bgp_evpn_vni_state_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error
    cmn_log := "GET: xfmr for BGP EVPN VNI state"

    vni_obj, vni_key, get_err := validate_vni_get (inParams, cmn_log);
    if get_err != nil {
        return get_err
    }

    err = get_specific_vni_state (vni_obj, inParams.dbs[db.ConfigDB], &vni_key)
    return err;
}

type _xfmr_bgp_vni_state_key struct {
    niName string
    vniNumber string
    afiSafiNameStr string
    afiSafiNameDbStr string
    afiSafiNameEnum ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE
}

func fill_vni_state_info (vni_key *_xfmr_bgp_vni_state_key, vniDataValue interface{}, cfgDb *db.DB,
                          vni_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Global_AfiSafis_AfiSafi_L2VpnEvpn_Vnis_Vni) error {
    var err error

    vniDataJson := vniDataValue.(map[string]interface{})
    
    vniState := vni_obj.State

    vninum := vniDataJson["vni"].(uint32)
    importlist := vniDataJson["importRts"].([]string)
    exportlist := vniDataJson["exportRts"].([]string)

    if value, ok := vniDataJson["kernelFlag"] ; ok {
        switch value {
            case "Yes":
                b := true
                vniState.IsLive = &b
            case "No":
                b := false
                vniState.IsLive = &b
        }
    }

    if value, ok := vniDataJson["advertiseGatewayMacip"] ; ok {
        switch value {
            case "Yes":
                b := true
                vniState.AdvertiseGwMac = &b
            case "No":
                b := false
                vniState.AdvertiseGwMac = &b
        }
    }

    if value, ok := vniDataJson["type"].(string) ; ok {
        vniState.Type = &value
    }

    if value, ok := vniDataJson["rd"].(string) ; ok {
        vniState.RouteDistinguisher = &value
    }

    if value, ok := vniDataJson["mcastGroup"].(string) ; ok {
        vniState.McastGroup = &value
    }

    if value, ok := vniDataJson["originatorIp"].(string) ; ok {
        vniState.Originator = &value
    }

    for _,importrt := range importlist {
        vniState.ImportRts = append(vniState.ImportRts, importrt)
    }

    for _,exportrt := range exportlist {
        vniState.ExportRts = append(vniState.ExportRts, exportrt)
    }

    vniState.VniNumber = &vninum

    return err
}

func get_specific_vni_state (vni_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Global_AfiSafis_AfiSafi_L2VpnEvpn_Vnis_Vni,
                             cfgDb *db.DB, vni_key *_xfmr_bgp_vni_state_key) error {
    var err error
 
    vtysh_cmd := "show ip bgp l2vpn evpn " + vni_key.vniNumber + " json"
    vniMapJson, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bgp l2vpn evpn state info for niName:%s vniNumber:%s. Err: %s\n", vni_key.niName, vni_key.vniNumber, err)
        return cmd_err
    }

    /*
    //This is test data from sample json output for test off-device
    vniMapJson := make(map[string]interface{})
    importlist := []string{"10:10", "20:20", "30:30"} 
    exportlist := []string{"100:100", "200:200"} 
    vniMapJson["output"] = map[string]interface{}{
        "vni":uint32(100), 
        "type":"L2", 
        "importRts":importlist, 
        "exportRts":exportlist, 
        "kernelFlag":"Yes",
        "rd":"2.2.2.2:56",
        "originatorIp":"1.1.1.1",
        "mcastGroup":"0.0.0.0",
        "advertiseGatewayMacip":"No",
    }
    */

    if vniDataJson, ok := vniMapJson["output"].(map[string]interface{}) ; ok {
        err = fill_vni_state_info (vni_key, vniDataJson, cfgDb, vni_obj)
    }    

    return err
}

func validate_vni_get (inParams XfmrParams, dbg_log string) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Global_AfiSafis_AfiSafi_L2VpnEvpn_Vnis_Vni, _xfmr_bgp_vni_state_key, error) {
    var err error
    var ok bool
    oper_err := errors.New("Opertational error")
    var vni_key _xfmr_bgp_vni_state_key
    var bgp_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp

    bgp_obj, vni_key.niName, err = getBgpRoot (inParams)
    if err != nil {
        log.Errorf ("%s failed !! Error:%s", dbg_log , err);
        return nil, vni_key, err
    }

    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)
    vni_key.vniNumber = pathInfo.Var("vni-number")
    vni_key.afiSafiNameStr = pathInfo.Var("afi-safi-name")
    vni_key.afiSafiNameEnum, vni_key.afiSafiNameDbStr, ok = get_afi_safi_name_enum_dbstr_for_ocstr (vni_key.afiSafiNameStr)
    log.Infof("%s : path:%s; template:%s targetUriPath:%s niName:%s vniNumber:%s",
              dbg_log, pathInfo.Path, pathInfo.Template, targetUriPath, vni_key.niName, vni_key.vniNumber)

    afiSafis_obj := bgp_obj.Global.AfiSafis
    if afiSafis_obj == nil {
        log.Errorf("%s failed !! Error: BGP AfiSafis container missing", dbg_log)
        return nil, vni_key, oper_err
    }

    afiSafi_obj, ok := afiSafis_obj.AfiSafi[vni_key.afiSafiNameEnum]
    if !ok {
        log.Errorf("%s failed !! Error: BGP AfiSafi object missing", dbg_log)
        return nil, vni_key, oper_err
    }

    nbrs_obj := afiSafi_obj.L2VpnEvpn.Vnis
    if nbrs_obj == nil {
        log.Errorf("%s failed !! Error: VNIs container missing", dbg_log)
        return nil, vni_key, oper_err
    }

    vninum64, err := strconv.ParseUint(vni_key.vniNumber, 10, 32)
    vninum := uint32(vninum64)

    vni_obj, ok := nbrs_obj.Vni[vninum]
    if !ok {
        log.Errorf("%s failed !! Error: Vni object %u missing", dbg_log, vninum)
        return nil, vni_key, oper_err
    }

    return vni_obj, vni_key, err
}
