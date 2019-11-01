package transformer

import (
    "errors"
    "strings"
    "strconv"
    "translib/ocbinds"
    log "github.com/golang/glog"
)

const (
    SONIC_PREFIX_SET_MODE_IPV4 = "IPv4"
    SONIC_PREFIX_SET_MODE_IPV6 = "IPv6"
)

/* E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode */
var PREFIX_SET_MODE_MAP = map[string]string{
    strconv.FormatInt(int64(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode_IPV4), 10): SONIC_PREFIX_SET_MODE_IPV4,
    strconv.FormatInt(int64(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode_IPV6), 10): SONIC_PREFIX_SET_MODE_IPV6,
}

func init () {
    XlateFuncBind("YangToDb_prefix_set_key_xfmr", YangToDb_prefix_set_key_xfmr)
    XlateFuncBind("DbToYang_prefix_set_key_xfmr", DbToYang_prefix_set_key_xfmr)
    XlateFuncBind("YangToDb_prefix_set_name_xfmr", YangToDb_prefix_set_name_xfmr)
    XlateFuncBind("DbToYang_prefix_set_name_xfmr", DbToYang_prefix_set_name_xfmr)
    XlateFuncBind("YangToDb_prefix_set_mode_field_xfmr", YangToDb_prefix_set_mode_field_xfmr)
    XlateFuncBind("DbToYang_prefix_set_mode_field_xfmr", DbToYang_prefix_set_mode_field_xfmr)
    XlateFuncBind("YangToDb_prefix_key_xfmr", YangToDb_prefix_key_xfmr)
    XlateFuncBind("DbToYang_prefix_key_xfmr", DbToYang_prefix_key_xfmr)
    XlateFuncBind("YangToDb_prefix_ip_prefix_xfmr", YangToDb_prefix_ip_prefix_xfmr)
    XlateFuncBind("DbToYang_prefix_ip_prefix_xfmr", DbToYang_prefix_ip_prefix_xfmr)
    XlateFuncBind("DbToYang_prefix_masklength_range_xfmr", DbToYang_prefix_masklength_range_xfmr)
}

var YangToDb_prefix_set_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
 
    var err error
    var setName string
    
    log.Info("YangToDb_prefix_set_key_xfmr: ", inParams.uri)
    
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <name> */
    setName = pathInfo.Var("name")

    if len(pathInfo.Vars) <  1 {
        err = errors.New("Invalid Key length");
        log.Info("Invalid Key length", len(pathInfo.Vars))
	return setName, err
    }

    if len(setName) == 0 {
        err = errors.New("set name is missing");
        log.Info("Set Name is Missing")
	return setName, err
    }

    log.Info("YangToDb_prefix_set_key_xfmr: setTblKey:", setName)

    return setName, nil
}

var DbToYang_prefix_set_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
 
    var err error
    rmap := make(map[string]interface{})

    key := inParams.key
    log.Info("DbToYang_prefix_set_key: ", key)
    setTblKey := strings.Split(key, "|")
    
    if len(setTblKey) < 1 {
        err = errors.New("Invalid key for prefix set.")
        log.Info("Invalid Keys for prefix sets", key)
        return rmap, err
    }

    setName := setTblKey[0]

    log.Info("DbToYang_prefix_set_key: ", setName)

    rmap["name"] = setName

    return rmap, nil
}

var YangToDb_prefix_set_name_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_prefix_set_name_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
    return res_map, nil
}


var DbToYang_prefix_set_name_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_prefix_set_name_xfmr: ", inParams.key)
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    log.Info("DbToYang_prefix_set_key: ", key)
    setTblKey := strings.Split(key, "|")
    setName := setTblKey[0]

    //setName, _ := getOCAclKeysFromStrDBKey(inParams.key)
    res_map["name"] = setName
    log.Info("prefix-set/config/name  ", res_map)
    return res_map, err
}

var YangToDb_prefix_set_mode_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	if inParams.param == nil {
	    res_map["mode"] = ""
	    return res_map, err
	}

	mode, _ := inParams.param.(ocbinds.E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode)
	log.Info("YangToDb_prefix_set_mode_field_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " Mode: ", mode)
	res_map["mode"] = findInMap(PREFIX_SET_MODE_MAP, strconv.FormatInt(int64(mode), 10))
	return res_map, err
}

var DbToYang_prefix_set_mode_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_prefix_set_mode_field_xfmr", data, inParams.ygRoot)
	oc_mode := findInMap(PREFIX_SET_MODE_MAP, data["PREFIX_SET"][inParams.key].Field["mode"])
	n, err := strconv.ParseInt(oc_mode, 10, 64)
	log.Info("DbToYang_prefix_set_mode_field_xfmr", oc_mode)
	result["mode"] = ocbinds.E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode(n).ΛMap()["E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode"][n].Name
	return result, err
}

var YangToDb_prefix_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var setName string
    var ipPrefix string
    var masklenrange string
    var prefixTblKey string

    log.Info("YangToDb_prefix_key_xfmr: ", inParams.ygRoot, inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    if len(pathInfo.Vars) < 3 {
        err = errors.New("Invalid xpath, key attributes not found")
        log.Info("Prefix set name is Missing")
        return prefixTblKey, err
    }
    setName = pathInfo.Var("name")
    ipPrefix = pathInfo.Var("ip-prefix")
    masklenrange = pathInfo.Var("masklength-range")

    if len(setName) == 0 {
        err = errors.New("Prefix set name is missing");
        log.Info("Prefix set name is Missing")
	return ipPrefix, err
    }

    if len(ipPrefix) == 0 {
        err = errors.New("ipPrefix is missing");
        log.Info("ipPrefix is Missing")
	return ipPrefix, err
    }

    if len(masklenrange) == 0 {
        err = errors.New("masklenrange is missing");
        log.Info("masklength-range is Missing")
	return masklenrange, err
    }

    prefixTblKey = setName + "|" + ipPrefix + "|" + masklenrange

    log.Info("YangToDb_prefix_key_xfmr: prefixTblKey: ", prefixTblKey)

    return prefixTblKey, nil
}

var DbToYang_prefix_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    key := inParams.key
   
    log.Info("DbToYang_prefix_key: ", key)

    prefixTblKey := strings.Split(key, "|")
    ipPrefix     := prefixTblKey[1]
    masklenrange := prefixTblKey[2]

    rmap["ip-prefix"] = ipPrefix
    rmap["masklength-range"] = masklenrange

    log.Info("DbToYang_prefix_key_xfmr:  ipPrefix ", ipPrefix , "masklength-range ", masklenrange)

    return rmap, nil
}

var YangToDb_prefix_ip_prefix_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_prefix_ip_prefix_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_prefix_ip_prefix_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_prefix_ip_prefix_xfmr: ", inParams.key)
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    prefixKey := strings.Split(key, "|")
    ip_prefix := prefixKey[1]

    res_map["ip-prefix"] = ip_prefix
    log.Info("prefix-set/prefix/config/ip-prefix ", res_map)
    return res_map, err
}

var DbToYang_prefix_masklength_range_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_prefix_masklength_range_xfmr: ", inParams.key)
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    prefixKey := strings.Split(key, "|")
    mask := prefixKey[2]

    res_map["masklength-range"] = mask
    log.Info("prefix-set/prefix/config/masklength-range ", res_map)
    return res_map, err
}



