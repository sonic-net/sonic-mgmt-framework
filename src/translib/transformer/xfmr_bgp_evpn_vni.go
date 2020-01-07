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
    "fmt"
    log "github.com/golang/glog"
    "github.com/openconfig/ygot/ygot"
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

    vniNumberKey := strings.Split(entry_key, "|")
    vniNumber, _  := strconv.ParseFloat(vniNumberKey[2], 64)

    rmap["vni-number"] = vniNumber

    log.Info("Rmap", rmap)

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

    log.Info("DbToYang_bgp_vni_number_fld_xfmr: ", inParams.key)

    entry_key := inParams.key
    vniKey := strings.Split(entry_key, "|")
    vniNumber, _ := strconv.ParseFloat(vniKey[2], 64)

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
    routeTarget  := routeTargetKey[2]

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
    routeTarget  := vniRouteTargetKey[3]

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
    var afi_list []string

    log.Info(inParams.key)

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_advertise_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS_AF"]
    log.Info("Table: ", pTbl)
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_advertise_fld_xfmr BGP AF not found : ", inParams.key)
        return result, errors.New("BGP AF not found : " + inParams.key)
    }
    GblAfData := pTbl[inParams.key]

    adv_ipv4_uni, ok := GblAfData.Field["advertise-ipv4-unicast"]
    if ok {
        if adv_ipv4_uni == "true" {
            afi_list = append(afi_list, "IPV4_UNICAST")
        }
    } else {
        log.Info("advertise-ipv4-unicast field not found in DB")
    }

    adv_ipv6_uni, ok := GblAfData.Field["advertise-ipv6-unicast"]
    if ok {
        if adv_ipv6_uni == "true" {
            afi_list = append(afi_list, "IPV6_UNICAST")
        }
    } else {
        log.Info("advertise-ipv6-unicast field not found in DB")
    }

    result["advertise-list"] = afi_list
    
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

    var vniState *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Global_AfiSafis_AfiSafi_L2VpnEvpn_Vnis_Vni_State
    if vniState = vni_obj.State ; vniState == nil {
        var _vniState ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Global_AfiSafis_AfiSafi_L2VpnEvpn_Vnis_Vni_State
        vni_obj.State = &_vniState
        vniState = vni_obj.State
    }

    if value, ok := vniDataJson["vni"].(float64) ; ok {
        vninum := uint32(value)
        vniState.VniNumber = &vninum
    }   

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

    if importlist, ok := vniDataJson["importRts"].([]interface{}) ; ok {
        s := make([]string, len(importlist))
        for i, v := range importlist {
            s[i] = fmt.Sprint(v)
        }
        for _,importrt := range s {
            vniState.ImportRts = append(vniState.ImportRts, importrt)
        }
    }

    if exportlist, ok := vniDataJson["exportRts"].([]interface{}) ; ok {
        s := make([]string, len(exportlist))
        for i, v := range exportlist {
            s[i] = fmt.Sprint(v)
        }
        for _,exportrt := range s {
            vniState.ExportRts = append(vniState.ExportRts, exportrt)
        }
    }

    return err
}

func get_specific_vni_state (vni_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Global_AfiSafis_AfiSafi_L2VpnEvpn_Vnis_Vni,
                             cfgDb *db.DB, vni_key *_xfmr_bgp_vni_state_key) error {
    var err error
    vniMapJson := make(map[string]interface{})
 
    vtysh_cmd := "show bgp l2vpn evpn vni " + vni_key.vniNumber + " json"
    output, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bgp l2vpn evpn state info for niName:%s vniNumber:%s. Err: %s\n", vni_key.niName, vni_key.vniNumber, err)
        return cmd_err
    }

    vniMapJson["output"] = output

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
        log.Infof("%s Vni object %u missing. Create", dbg_log, vninum)
        vni_obj, _ = nbrs_obj.NewVni (vninum)
    }

    return vni_obj, vni_key, err
}


func hdl_get_bgp_l2vpn_evpn_local_rib (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                            rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
    oper_err := errors.New("Operational error")
    var ok bool

    log.Infof("%s ==> Local-RIB invoke with keys {%s} afiSafiType:%d", *dbg_log, print_rib_keys(rib_key), afiSafiType)

    cmd := "show bgp l2vpn evpn route detail json"

    bgpRibOutputJson := make(map[string]interface{})

    output, cmd_err := exec_vtysh_cmd (cmd)
    if (cmd_err != nil) {
        log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
        return oper_err
    }
    bgpRibOutputJson["output"] = output

    log.Info(bgpRibOutputJson)

    var ribAfiSafis_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis
    if ribAfiSafis_obj = bgpRib_obj.AfiSafis ; ribAfiSafis_obj == nil {
        var _ribAfiSafis ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis
        bgpRib_obj.AfiSafis = &_ribAfiSafis
        ribAfiSafis_obj = bgpRib_obj.AfiSafis
    }

    var ribAfiSafi_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi
    if ribAfiSafi_obj, ok = ribAfiSafis_obj.AfiSafi[afiSafiType] ; !ok {
        ribAfiSafi_obj, _ = ribAfiSafis_obj.NewAfiSafi (afiSafiType)
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_L2VPN_EVPN {
        err = hdl_get_bgp_evpn_local_rib (ribAfiSafi_obj, rib_key, bgpRibOutputJson, dbg_log)
    }

    return err
}

func hdl_get_bgp_evpn_local_rib (ribAfiSafi_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi,
                                 rib_key *_xfmr_bgp_rib_key, bgpRibOutputJson map[string]interface{}, dbg_log *string) (error) {
    var err error

    var l2vpnEvpn_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn
    if l2vpnEvpn_obj = ribAfiSafi_obj.L2VpnEvpn ; l2vpnEvpn_obj == nil {
        var _l2vpnEvpn ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn
        ribAfiSafi_obj.L2VpnEvpn = &_l2vpnEvpn
        l2vpnEvpn_obj = ribAfiSafi_obj.L2VpnEvpn
    }

    var evpnLocRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn_LocRib
    if evpnLocRib_obj = l2vpnEvpn_obj.LocRib ; evpnLocRib_obj == nil {
        var _evpnLocRib ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn_LocRib
        l2vpnEvpn_obj.LocRib = &_evpnLocRib
        evpnLocRib_obj = l2vpnEvpn_obj.LocRib
    }

    var evpnLocRibRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn_LocRib_Routes
    if evpnLocRibRoutes_obj = evpnLocRib_obj.Routes ; evpnLocRibRoutes_obj == nil {
        var _evpnLocRibRoutes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn_LocRib_Routes
        evpnLocRib_obj.Routes = &_evpnLocRibRoutes
        evpnLocRibRoutes_obj = evpnLocRib_obj.Routes
    }

    rds, _ := bgpRibOutputJson["output"].(map[string]interface{})
    for rd, _ := range rds {
        rd_data, ok := rds[rd].(map[string]interface{}) ; if !ok {continue}
        for prefix, _ := range rd_data {
            prefix_data, ok := rd_data[prefix].(map[string]interface{}) ; if !ok {continue}
            patharrayarray, ok := prefix_data["paths"].([]interface{}) ; if !ok {
                log.Info("patharray not parsed")
            }
            for i, patharray := range patharrayarray {
                if i > 0 {continue}
                for j, path := range patharray.([]interface{}) {
                    if j > 0 {continue}
                    path_data, ok := path.(map[string]interface{}) ; if !ok {
                      log.Info("path not parsed")
                    }
                    pathId := uint32(j)
                    if ok := fill_evpn_spec_pfx_path_loc_rib_data (evpnLocRibRoutes_obj,rd, prefix, pathId, path_data) ; !ok {continue}
                }
            }
        }
    }

    return err
}

func fill_evpn_spec_pfx_path_loc_rib_data (ipv4LocRibRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn_LocRib_Routes,
                                               rd string, prefix string, pathId uint32, pathData map[string]interface{}) bool {
    
    origin, ok := pathData["origin"].(string)
    if !ok {return false}

    _route_origin := &ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn_LocRib_Routes_Route_State_Origin_Union_String{origin}
    ipv4LocRibRoute_obj, err := ipv4LocRibRoutes_obj.NewRoute (rd, prefix)
    if err != nil {return false}
    ygot.BuildEmptyTree(ipv4LocRibRoute_obj)

    var _state ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn_LocRib_Routes_Route_State
    ipv4LocRibRoute_obj.State = &_state
    ipv4LocRibRouteState := ipv4LocRibRoute_obj.State

    ipv4LocRibRouteState.Prefix = &prefix
    ipv4LocRibRouteState.Origin = _route_origin
    ipv4LocRibRouteState.PathId = &pathId

    if value, ok := pathData["valid"].(bool) ; ok {
        ipv4LocRibRouteState.ValidRoute = &value
    }

    ipv4LocRibRouteAttrSets := ipv4LocRibRoute_obj.AttrSets

    if value, ok := pathData["localPref"] ; ok {
        _localPref := uint32(value.(float64))
        ipv4LocRibRouteAttrSets.LocalPref = &_localPref
    }

    if value, ok := pathData["med"] ; ok {
        _med := uint32(value.(float64))
        ipv4LocRibRouteAttrSets.Med = &_med
    }

    if value, ok := pathData["originatorId"].(string) ; ok {
        ipv4LocRibRouteAttrSets.OriginatorId = &value
    }

    if value, ok := pathData["nexthops"].([]interface{}) ; ok {
        if nhop, ok := value[0].(map[string]interface{}) ; ok {
            if ip, ok := nhop["ip"].(string) ; ok {
                ipv4LocRibRouteAttrSets.NextHop = &ip
            }
        }
    }

    if value, ok := pathData["aspath"].(map[string]interface{}) ; ok {
        if asPathSegments, ok := value["segments"].([]interface {}) ; ok {
            for _, asPathSegmentsData := range asPathSegments {
                var _segment ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn_LocRib_Routes_Route_AttrSets_AsPath_AsSegment
                ygot.BuildEmptyTree (&_segment)
                if ok = parse_aspath_segment_data (asPathSegmentsData.(map[string]interface {}), &_segment.State.Type, &_segment.State.Member) ; ok {
                   ipv4LocRibRouteAttrSets.AsPath.AsSegment = append (ipv4LocRibRouteAttrSets.AsPath.AsSegment, &_segment)
                }
            }
        }
    }

    if value, ok := pathData["extendedCommunity"].(map[string]interface{}) ; ok {
        if _value, ok := value["string"] ; ok {
            _community_slice := strings.Split (_value.(string), " ")
            for _, _data := range _community_slice {
                if _ext_community_union, err := ipv4LocRibRouteAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_L2VpnEvpn_LocRib_Routes_Route_AttrSets_ExtCommunity_Union (_data) ; err == nil {
                    ipv4LocRibRouteAttrSets.ExtCommunity = append (ipv4LocRibRouteAttrSets.ExtCommunity, _ext_community_union)
                }
            }
        }
    }

    return true
}