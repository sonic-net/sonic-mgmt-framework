package transformer

import (
    "errors"
    "strings"
    "strconv"
    "translib/ocbinds"
    "translib/db"
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
    XlateFuncBind("YangToDb_prefix_set_name_fld_xfmr", YangToDb_prefix_set_name_fld_xfmr)
    XlateFuncBind("DbToYang_prefix_set_name_fld_xfmr", DbToYang_prefix_set_name_fld_xfmr)
    XlateFuncBind("YangToDb_prefix_set_mode_fld_xfmr", YangToDb_prefix_set_mode_fld_xfmr)
    XlateFuncBind("DbToYang_prefix_set_mode_fld_xfmr", DbToYang_prefix_set_mode_fld_xfmr)
    XlateFuncBind("YangToDb_prefix_key_xfmr", YangToDb_prefix_key_xfmr)
    XlateFuncBind("DbToYang_prefix_key_xfmr", DbToYang_prefix_key_xfmr)
    XlateFuncBind("YangToDb_prefix_ip_prefix_fld_xfmr", YangToDb_prefix_ip_prefix_fld_xfmr)
    XlateFuncBind("DbToYang_prefix_ip_prefix_fld_xfmr", DbToYang_prefix_ip_prefix_fld_xfmr)
    XlateFuncBind("DbToYang_prefix_masklength_range_fld_xfmr", DbToYang_prefix_masklength_range_fld_xfmr)
}

var YangToDb_prefix_set_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
 
    var err error
    var setName string
    
    log.Info("YangToDb_prefix_set_key_xfmr: ", inParams.uri)
    
    pathInfo := NewPathInfo(inParams.uri)

    setName = pathInfo.Var("name")

    if len(pathInfo.Vars) <  1 {
        err = errors.New("Invalid Key length");
        log.Error("Invalid Key length", len(pathInfo.Vars))
	return setName, err
    }

    if len(setName) == 0 {
        err = errors.New("set name is missing");
        log.Error("Set Name is Missing")
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
        log.Error("Invalid Keys for prefix sets", key)
        return rmap, err
    }

    setName := setTblKey[0]

    log.Info("DbToYang_prefix_set_key: ", setName)

    rmap["name"] = setName

    return rmap, nil
}

var YangToDb_prefix_set_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_prefix_set_name_fld_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_prefix_set_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_prefix_set_name_fld_xfmr: ", inParams.key)
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    log.Info("DbToYang_prefix_set_name_fld_xfmr: ", key)
    setTblKey := strings.Split(key, "|")
    setName := setTblKey[0]

    res_map["name"] = setName
    log.Info("prefix-set/config/name  ", res_map)
    return res_map, err
}

var YangToDb_prefix_set_mode_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	if inParams.param == nil {
	    res_map["mode"] = ""
	    return res_map, err
	}

	mode, _ := inParams.param.(ocbinds.E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode)
	log.Info("YangToDb_prefix_set_mode_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " Mode: ", mode)
	res_map["mode"] = findInMap(PREFIX_SET_MODE_MAP, strconv.FormatInt(int64(mode), 10))
	return res_map, err
}

var DbToYang_prefix_set_mode_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_prefix_set_mode_fld_xfmr", data, inParams.ygRoot)
	oc_mode := findInMap(PREFIX_SET_MODE_MAP, data["PREFIX_SET"][inParams.key].Field["mode"])
	n, err := strconv.ParseInt(oc_mode, 10, 64)
	log.Info("DbToYang_prefix_set_mode_fld_xfmr", oc_mode)
	result["mode"] = ocbinds.E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode(n).ΛMap()["E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode"][n].Name
	return result, err
}

func prefix_all_keys_get(d *db.DB, dbSpec *db.TableSpec) ([]db.Key, error) {

    var keys []db.Key

    prefixTable, err := d.GetTable(dbSpec)
    if err != nil {
        return keys, err
    }

    keys, err = prefixTable.GetKeys()
    log.Info("prefix_all_keys_get: Found %d PREFIX table keys", len(keys))
    return keys, err
}

func prefix_del_by_set_name (d *db.DB , setName string, tblName string) (string, error) {
    var err error
    var prefixTblKey string
    first := false
    prefixTblKey = "NULL"
    keys,_ := prefix_all_keys_get(d, &db.TableSpec{Name:tblName})
    for _, key := range keys {
        log.Info("prefix_del_by_set_name: Found PREFIX table key set ", key.Get(0), "prefix ", key.Get(1), "mask ", key.Get(2))
        if len(key.Comp) < 3 {
            continue
        }
        if key.Get(0) != setName {
            continue
        }

        if first == false {
           prefixTblKey = key.Get(0) + "|" + key.Get(1) + "|" + key.Get(2)
           first = true
           continue
        }

        log.Info("prefix_del_by_set_name: PREFIX key", key)
        err = d.DeleteEntry(&db.TableSpec{Name:tblName}, key)
        if err != nil {
            log.Error("Prefix Delete Entry fails e = %v key %v", err, key)
            return setName, err
        }
    }
    if first == true {
        return prefixTblKey,nil
    } else {
        err =  errors.New("Unknown prefix set name")
        return prefixTblKey, err
    }
}
var YangToDb_prefix_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var setName string
    var ipPrefix string
    var masklenrange string
    var prefixTblKey string

    log.Info("YangToDb_prefix_key_xfmr: ", inParams.ygRoot, inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    if inParams.oper == DELETE {
        if len(pathInfo.Vars) == 1  {
            setName = pathInfo.Var("name")
            if len(setName) == 0 {
                err = errors.New("YangToDb_prefix_key_xfmr: Prefix set name is missing");
                log.Error("YangToDb_prefix_key_xfmr: Prefix set name is Missing")
                return setName, err
            }
            return prefix_del_by_set_name (inParams.d, setName, "PREFIX")
        }
    } else {
        if len(pathInfo.Vars) < 3 {
            err = errors.New("Invalid xpath, key attributes not found")
            log.Error("YangToDb_prefix_key_xfmr: Prefix keys are Missing, numKeys ", len(pathInfo.Vars))
            return prefixTblKey, err
        }
        setName = pathInfo.Var("name")
        ipPrefix = pathInfo.Var("ip-prefix")
        masklenrange = pathInfo.Var("masklength-range")

        if len(setName) == 0 {
            err = errors.New("YangToDb_prefix_key_xfmr: Prefix set name is missing");
            log.Info("YangToDb_prefix_key_xfmr: Prefix set name is Missing")
            return setName, err
        }

        if len(ipPrefix) == 0 {
            err = errors.New("YangToDb_prefix_key_xfmr: ipPrefix is missing");
            log.Info("YangToDb_prefix_key_xfmr: ipPrefix is Missing")
            return ipPrefix, err
        }

        if len(masklenrange) == 0 {
            err = errors.New("YangToDb_prefix_key_xfmr: masklenrange is missing");
            log.Info("YangToDb_prefix_key_xfmr: masklength-range is Missing")
            return masklenrange, err
        }

        prefixTblKey = setName + "|" + ipPrefix + "|" + masklenrange
    }
    log.Info("YangToDb_prefix_key_xfmr: prefixTblKey: ", prefixTblKey)

    return prefixTblKey, nil
}

var DbToYang_prefix_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    key := inParams.key
   
    log.Info("DbToYang_prefix_key_xfmr: ", key)

    prefixTblKey := strings.Split(key, "|")
    ipPrefix     := prefixTblKey[1]
    masklenrange := prefixTblKey[2]

    rmap["ip-prefix"] = ipPrefix
    rmap["masklength-range"] = masklenrange

    log.Info("DbToYang_prefix_key_xfmr:  ipPrefix ", ipPrefix , "masklength-range ", masklenrange)

    return rmap, nil
}

var YangToDb_prefix_ip_prefix_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_prefix_ip_prefix_fld_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_prefix_ip_prefix_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_prefix_ip_prefix_fld_xfmr: ", inParams.key)
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    prefixKey := strings.Split(key, "|")
    ip_prefix := prefixKey[1]

    res_map["ip-prefix"] = ip_prefix
    log.Info("prefix-set/prefix/config/ip-prefix ", res_map)
    return res_map, err
}

var DbToYang_prefix_masklength_range_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_prefix_masklength_range_fld_xfmr: ", inParams.key)
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    prefixKey := strings.Split(key, "|")
    mask := prefixKey[2]

    res_map["masklength-range"] = mask
    log.Info("prefix-set/prefix/config/masklength-range ", res_map)
    return res_map, err
}
