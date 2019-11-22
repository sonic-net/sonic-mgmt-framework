///////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Dell, Inc.                                                 //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//  http://www.apache.org/licenses/LICENSE-2.0                                //
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
	log "github.com/golang/glog"
	"strconv"
	"strings"
	"translib/db"
    "translib/ocbinds"
)

func init() {
	XlateFuncBind("YangToDb_nat_instance_key_xfmr", YangToDb_nat_instance_key_xfmr)
	XlateFuncBind("DbToYang_nat_instance_key_xfmr", DbToYang_nat_instance_key_xfmr)
	XlateFuncBind("YangToDb_nat_global_key_xfmr", YangToDb_nat_global_key_xfmr)
	XlateFuncBind("DbToYang_nat_global_key_xfmr", DbToYang_nat_global_key_xfmr)
	XlateFuncBind("YangToDb_nat_enable_xfmr", YangToDb_nat_enable_xfmr)
	XlateFuncBind("DbToYang_nat_enable_xfmr", DbToYang_nat_enable_xfmr)
        XlateFuncBind("YangToDb_napt_mapping_key_xfmr", YangToDb_napt_mapping_key_xfmr)
        XlateFuncBind("DbToYang_napt_mapping_key_xfmr", DbToYang_napt_mapping_key_xfmr)
        XlateFuncBind("YangToDb_napt_mapping_state_key_xfmr", YangToDb_napt_mapping_state_key_xfmr)
        XlateFuncBind("DbToYang_napt_mapping_state_key_xfmr", DbToYang_napt_mapping_state_key_xfmr)
        XlateFuncBind("YangToDb_nat_mapping_key_xfmr", YangToDb_nat_mapping_key_xfmr)
        XlateFuncBind("DbToYang_nat_mapping_key_xfmr", DbToYang_nat_mapping_key_xfmr)
        XlateFuncBind("YangToDb_nat_pool_key_xfmr", YangToDb_nat_pool_key_xfmr)
        XlateFuncBind("DbToYang_nat_pool_key_xfmr", DbToYang_nat_pool_key_xfmr)
        XlateFuncBind("YangToDb_nat_binding_key_xfmr", YangToDb_nat_binding_key_xfmr)
        XlateFuncBind("DbToYang_nat_binding_key_xfmr", DbToYang_nat_binding_key_xfmr)
        XlateFuncBind("YangToDb_nat_zone_key_xfmr", YangToDb_nat_zone_key_xfmr)
        XlateFuncBind("DbToYang_nat_zone_key_xfmr", DbToYang_nat_zone_key_xfmr)
        XlateFuncBind("YangToDb_nat_twice_mapping_key_xfmr", YangToDb_nat_twice_mapping_key_xfmr)
        XlateFuncBind("DbToYang_nat_twice_mapping_key_xfmr", DbToYang_nat_twice_mapping_key_xfmr)
        XlateFuncBind("YangToDb_napt_twice_mapping_key_xfmr", YangToDb_napt_twice_mapping_key_xfmr)
        XlateFuncBind("DbToYang_napt_twice_mapping_key_xfmr", DbToYang_napt_twice_mapping_key_xfmr)
        XlateFuncBind("YangToDb_nat_type_field_xfmr", YangToDb_nat_type_field_xfmr)
        XlateFuncBind("DbToYang_nat_type_field_xfmr", DbToYang_nat_type_field_xfmr)
        XlateFuncBind("YangToDb_nat_entry_type_field_xfmr", YangToDb_nat_entry_type_field_xfmr)
        XlateFuncBind("DbToYang_nat_entry_type_field_xfmr", DbToYang_nat_entry_type_field_xfmr)
}

const (
    ADMIN_MODE       = "admin_mode"
    NAT_GLOBAL_TN    = "NAT_GLOBAL"
    ENABLED          = "enabled"
    DISABLED         = "disabled"
    ENABLE           = "enable"
    INSTANCE_ID      = "id"
    GLOBAL_KEY       = "Values"
    NAT_TABLE        = "NAT_TABLE"
    STATIC_NAT       = "STATIC_NAT"
    NAT_TYPE         = "nat_type"
    NAT_ENTRY_TYPE   = "entry_type"
    STATIC           = "static"
    DYNAMIC          = "dynamic"
    SNAT             = "snat"
    DNAT             = "dnat"
)

var YangToDb_nat_instance_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var nat_inst_key string
    var err error

    return nat_inst_key, err
}

var DbToYang_nat_instance_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    rmap[INSTANCE_ID] = 0
    return rmap, err
}


var YangToDb_nat_global_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var nat_global_key string
    var err error

    nat_global_key = GLOBAL_KEY

    return nat_global_key, err
}

var DbToYang_nat_global_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    return rmap, err
}

var YangToDb_nat_enable_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    enabled, _ := inParams.param.(*bool)
    var enStr string
    if *enabled == true {
        enStr = ENABLED
    } else {
        enStr = DISABLED
    }
    res_map[ADMIN_MODE] = enStr

    return res_map, nil
}

var DbToYang_nat_enable_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]

    pTbl := data[NAT_GLOBAL_TN]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_intf_enabled_xfmr Values entry not found : ", inParams.key)
        return result, errors.New("Global Values not found : " + inParams.key)
    }

    prtInst := pTbl[inParams.key]
    adminMode, ok := prtInst.Field["admin_mode"]
    if ok {
        if adminMode == ENABLED {
            result[ENABLE] = true
        } else {
            result[ENABLE] = false
        }
    } else {
        result[ENABLE] = false
        log.Info("Admin Mode field not found in DB")
    }
    return result, err
}
var protocol_map  = map[uint8]string{
    1 : "ICMP",
    6 : "TCP",
    17: "UDP",
}

func findProtocolByValue(m map[uint8]string, value string) uint8 {
    for key, val := range m {
        if val == value {
            return key
        }
    }
    return 0
}

func getKeySeparator (dbId db.DBNum) (string) {
    var sep string

    if dbId == db.ConfigDB {
        sep = "|"
    } else if dbId == db.ApplDB {
        sep = ":"
    } else if dbId == db.StateDB {
        sep = "|"
    } else if dbId == db.CountersDB {
        sep = ":"
    } else {
        sep = "|"
    }
    return sep
}

var YangToDb_napt_mapping_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var napt_key string
    var err error

    var key_sep string

    pathInfo := NewPathInfo(inParams.uri)
    extAddress := pathInfo.Var("external-address")
    extPort := pathInfo.Var("external-port")
    protocol,_ := strconv.Atoi(pathInfo.Var("protocol"))


    if _, ok := protocol_map[uint8(protocol)]; !ok {
        log.Info("YangToDb_napt_mapping_key_xfmr - Invalid protocol : ", protocol);
        return napt_key, nil
    }

    key_sep = getKeySeparator(inParams.curDb)
    napt_key = extAddress + key_sep + protocol_map[uint8(protocol)] + key_sep + extPort
    log.Info("YangToDb_napt_mapping_key_xfmr : Key : ", napt_key)
    return napt_key, err
}

var DbToYang_napt_mapping_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    var key_sep string

    napt_key := inParams.key
    key_sep = getKeySeparator(inParams.curDb)

    key := strings.Split(napt_key, key_sep)
    if len(key) < 3 {
        err = errors.New("Invalid key for NAPT ampping entry.")
        log.Info("Invalid Keys, NAPT Mapping entry", napt_key)
        return rmap, err
    }
    oc_protocol := findProtocolByValue(protocol_map, key[1])

    rmap["external-address"] = key[0]
    rmap["external-port"], _ = strconv.Atoi(key[2])
    rmap["protocol"] = oc_protocol
    log.Info(" DbToYang_napt_mapping_key_xfmr : - ", rmap)
    return rmap, err
}

var YangToDb_napt_mapping_state_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var napt_key string
    var err error

    var key_sep string

    pathInfo := NewPathInfo(inParams.uri)
    extAddress := pathInfo.Var("external-address")
    extPort := pathInfo.Var("external-port")
    protocol,_ := strconv.Atoi(pathInfo.Var("protocol"))

    if _, ok := protocol_map[uint8(protocol)]; !ok {
        log.Info("YangToDb_napt_mapping_key_xfmr - Invalid protocol : ", protocol);
        return napt_key, nil
    }

    key_sep = getKeySeparator(inParams.curDb)
    napt_key = protocol_map[uint8(protocol)] + key_sep + extAddress + key_sep + extPort
    log.Info("YangToDb_napt_counters_key_xfmr: Key : ", napt_key)
    return napt_key, err
}

var DbToYang_napt_mapping_state_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    var key_sep string

    napt_key := inParams.key
    key_sep = getKeySeparator(inParams.curDb)

    key := strings.Split(napt_key, key_sep)
    if len(key) < 3 {
        err = errors.New("Invalid key for NAPT ampping entry.")
        log.Info("Invalid Keys, NAPT Mapping entry", napt_key)
        return rmap, err
    }
    oc_protocol := findProtocolByValue(protocol_map, key[0])

    rmap["external-address"] = key[1]
    rmap["external-port"], _ = strconv.Atoi(key[2])
    rmap["protocol"] = oc_protocol
    return rmap, err
}


var YangToDb_nat_mapping_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var nat_key string
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    extAddress := pathInfo.Var("external-address")

    nat_key = extAddress
    log.Info("YangToDb_nat_mapping_key_xfmr : Key : ", nat_key)
    return nat_key, err
}

var DbToYang_nat_mapping_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    nat_key := inParams.key
    rmap["external-address"] = nat_key
    log.Info("DbToYang_nat_mapping_key_xfmr : - ", rmap)
    return rmap, err
}


var YangToDb_nat_pool_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var key string
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    name := pathInfo.Var("pool-name")

    key = name
    log.Info("YangToDb_nat_pool_key_xfmr: Key : ", key)
    return key, err
}

var DbToYang_nat_pool_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    key := inParams.key
    rmap["pool-name"] = key
    log.Info("YangToDb_nat_pool_key_xfmr : - ", rmap)
    return rmap, err
}

var YangToDb_nat_binding_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var key string
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    name := pathInfo.Var("name")

    key = name
    log.Info("YangToDb_nat_binding_key_xfmr : Key : ", key)
    return key, err
}

var DbToYang_nat_binding_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    key := inParams.key
    rmap["name"] = key
    log.Info("YangToDb_nat_binding_key_xfmr : - ", rmap)
    return rmap, err
}


var YangToDb_nat_zone_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var key string
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    name := pathInfo.Var("zone-id")

    key = name
    log.Info("YangToDb_nat_zone_key_xfmr : Key : ", key)
    return key, err
}

var DbToYang_nat_zone_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    key := inParams.key
    rmap["zoneid"],_ = strconv.Atoi(key)
    log.Info("YangToDb_nat_zone_key_xfmr : - ", rmap)
    return rmap, err
}



var YangToDb_nat_twice_mapping_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var nat_key string
    var err error

    var key_sep string

    pathInfo := NewPathInfo(inParams.uri)
    srcIp := pathInfo.Var("src-ip")
    dstIp := pathInfo.Var("dst-ip")

    key_sep = getKeySeparator(inParams.curDb)
    nat_key = srcIp + key_sep + dstIp
    log.Info("YangToDb_nat_twice_mapping_key_xfmr : Key : ", nat_key)
    return nat_key, err
}

var DbToYang_nat_twice_mapping_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    nat_key := inParams.key
    key_sep := getKeySeparator(inParams.curDb)

    key := strings.Split(nat_key, key_sep)
    if len(key) < 2 {
        err = errors.New("Invalid key for NAT mapping entry.")
        log.Info("Invalid Keys, NAT Mapping entry", nat_key)
        return rmap, err
    }

    rmap["src-ip"] = key[0]
    rmap["dst-ip"] = key[1]
    log.Info("DbToYang_nat_twice_mapping_key_xfmr : - ", rmap)
    return rmap, err
}

var YangToDb_napt_twice_mapping_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var napt_key string
    var err error

    var key_sep string

    pathInfo := NewPathInfo(inParams.uri)
    protocol, _ := strconv.Atoi(pathInfo.Var("protocol"))
    srcIp    := pathInfo.Var("src-ip")
    srcPort  := pathInfo.Var("src-port")
    dstIp    := pathInfo.Var("dst-ip")
    dstPort  := pathInfo.Var("dst-port")

    if _, ok := protocol_map[uint8(protocol)]; !ok {
        log.Info("YangToDb_napt_twice_mapping_key_xfmr - Invalid protocol : ", protocol);
        return napt_key, nil
    }

    key_sep = getKeySeparator(inParams.curDb)
    napt_key = protocol_map[uint8(protocol)] + key_sep + srcIp + key_sep + srcPort + key_sep + dstIp + key_sep + dstPort
    log.Info("YangToDb_napt_twice_mapping_key_xfmr : Key : ", napt_key)
    return napt_key, err
}

var DbToYang_napt_twice_mapping_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    nat_key := inParams.key
    key_sep := getKeySeparator(inParams.curDb)

    key := strings.Split(nat_key, key_sep)
    if len(key) < 5 {
        err = errors.New("Invalid key for NAPT mapping entry.")
        log.Info("Invalid Keys, NAPT Mapping entry", nat_key)
        return rmap, err
    }
    oc_protocol := findProtocolByValue(protocol_map, key[0])

    rmap["protocol"] = oc_protocol
    rmap["src-ip"] = key[1]
    rmap["src-port"],_ = strconv.Atoi(key[2])
    rmap["dst-ip"] = key[3]
    rmap["dst-port"], _  = strconv.Atoi(key[4])

    log.Info("DbToYang_nat_twice_mapping_key_xfmr : - ", rmap)
    return rmap, err
}

var NAT_TYPE_MAP = map[string]string{
    strconv.FormatInt(int64(ocbinds.OpenconfigNat_NAT_TYPE_SNAT), 10): "snat",
    strconv.FormatInt(int64(ocbinds.OpenconfigNat_NAT_TYPE_DNAT), 10): "dnat",
}

var NAT_ENTRY_TYPE_MAP = map[string]string{
    strconv.FormatInt(int64(ocbinds.OpenconfigNat_NAT_ENTRY_TYPE_STATIC), 10): "static",
    strconv.FormatInt(int64(ocbinds.OpenconfigNat_NAT_ENTRY_TYPE_DYNAMIC), 10): "dynamic",
}


var YangToDb_nat_type_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    result := make(map[string]string)
    var err error

    if inParams.param == nil {
        result[NAT_TYPE] = ""
        return result, err
    }

    t, _ := inParams.param.(ocbinds.E_OpenconfigNat_NAT_TYPE)
    log.Info("YangToDb_nat_type_field_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " type: ", t)
    result[NAT_TYPE] = findInMap(NAT_TYPE_MAP, strconv.FormatInt(int64(t), 10))
    return result, err

}

var DbToYang_nat_type_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_nat_type_field_xfmr", data, inParams.ygRoot)

    t := findInMap(NAT_TYPE_MAP, data[STATIC_NAT][inParams.key].Field[NAT_TYPE])
    n, err := strconv.ParseInt(t, 10, 64)
    result[NAT_TYPE] = ocbinds.E_OpenconfigNat_NAT_TYPE(n).ΛMap()["E_OpenconfigNat_NAT_TYPE"][n].Name
    return result, err
}

var YangToDb_nat_entry_type_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    result := make(map[string]string)
    var err error

    if inParams.param == nil {
        result[NAT_ENTRY_TYPE] = ""
        return result, err
    }

    t, _ := inParams.param.(ocbinds.E_OpenconfigNat_NAT_ENTRY_TYPE)
    log.Info("YangToDb_nat_entry_type_field_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " type: ", t)
    result[NAT_ENTRY_TYPE] = findInMap(NAT_ENTRY_TYPE_MAP, strconv.FormatInt(int64(t), 10))
    return result, err
}

var DbToYang_nat_entry_type_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_nat_entry_type_field_xfmr", data, inParams.ygRoot)

    t := findInMap(NAT_ENTRY_TYPE_MAP, data[NAT_TABLE][inParams.key].Field[NAT_ENTRY_TYPE])
    n, err := strconv.ParseInt(t, 10, 64)
    result[NAT_ENTRY_TYPE] = ocbinds.E_OpenconfigNat_NAT_ENTRY_TYPE(n).ΛMap()["E_OpenconfigNat_NAT_ENTRY_TYPE"][n].Name
    return result, err
}

