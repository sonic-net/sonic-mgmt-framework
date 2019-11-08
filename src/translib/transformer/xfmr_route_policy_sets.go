package transformer

import (
    "bytes"
    "errors"
    "strings"
    "strconv"
    "translib/ocbinds"
    "translib/db"
    log "github.com/golang/glog"
    "reflect"
    "fmt"
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

    XlateFuncBind("YangToDb_community_set_name_fld_xfmr", YangToDb_community_set_name_fld_xfmr)
    XlateFuncBind("DbToYang_community_set_name_fld_xfmr", DbToYang_community_set_name_fld_xfmr)
    XlateFuncBind("YangToDb_community_match_set_options_fld_xfmr", YangToDb_community_match_set_options_fld_xfmr)
    XlateFuncBind("DbToYang_community_match_set_options_fld_xfmr", DbToYang_community_match_set_options_fld_xfmr)
    XlateFuncBind("YangToDb_community_member_fld_xfmr", YangToDb_community_member_fld_xfmr)
//    XlateFuncBind("DbToYang_community_member_fld_xfmr", DbToYang_community_member_fld_xfmr)
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
/*    if inParams.oper == DELETE {
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
    */
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
 //       delEntry := make(map[string]map[string]db.Value)
//        delEntry[tblName] = key1
//        delEntry[tblName] = key2
  //      inParams.txCache["delete"] = delEntry
  //      inParams.result = make(map[string]interface{})
 //       inParams.result = inParams.txCache["delete"]

        prefixTblKey = setName + "|" + ipPrefix + "|" + masklenrange
 //   }
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

var YangToDb_community_set_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_community_set_name_fld_xfmr: ", inParams.key)
//    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_community_set_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_community_set_name_fld_xfmr: ", inParams.key)
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    log.Info("DbToYang_community_set_name_fld_xfmr: ", key)
    setTblKey := strings.Split(key, "|")
    setName := setTblKey[0]

    res_map["name"] = setName
    log.Info("config/name  ", res_map)
    return res_map, err
}

var YangToDb_community_match_set_options_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	if inParams.param == nil {
	    res_map["mode"] = ""
	    return res_map, err
	}

	match_opt, _ := inParams.param.(ocbinds.E_OpenconfigRoutingPolicy_MatchSetOptionsType)
	log.Info("YangToDb_community_match_set_options_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " match Opt: ", match_opt)
	return res_map, err
}

var DbToYang_community_match_set_options_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_community_match_set_options_fld_xfmr", data, inParams.ygRoot)
	n := strconv.FormatInt(int64(ocbinds.OpenconfigRoutingPolicy_MatchSetOptionsType_ALL), 10)
	log.Info("DbToYang_community_match_set_options_fld_xfmr", n)
//	result["match-set-options"] = strconv.ParseInt(n, 10,64)
	log.Info("DbToYang_community_match_set_options_fld_xfmr", result["match-set-options"])
	return result, err
}

func community_set_action_get_by_set_name (d *db.DB , setName string, tblName string) (string, error) {
    var err error

    dbspec := &db.TableSpec { Name: tblName }
    dbEntry, err := d.GetEntry(dbspec, db.Key{Comp: []string{setName}})
    if err != nil {
        log.Error("No Entry found e = %v key %v", err, setName)
        return "", err
    }
    prev_type, ok := dbEntry.Field["type"]
    if ok {
        log.Info("Previous type ", prev_type)
    } else {
        log.Info("New Table, No previous type ", prev_type)
    }
    return prev_type, nil; 
}

var YangToDb_community_member_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
        var community []string
        var community_list string
        var numCommunity int
      //  var prev_type string
        var new_type string
        log.Info("YangToDb_community_member_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, "inParams : ", inParams)
        if inParams.param == nil {
	    res_map["community_member"] = ""
	    return res_map, errors.New("Invalid Inputs")
	}

        pathInfo := NewPathInfo(inParams.uri)
        if len(pathInfo.Vars) <  1 {
            err = errors.New("Invalid Key length");
            log.Error("Invalid Key length", len(pathInfo.Vars))
            return res_map, err
        }

        setName := pathInfo.Var("community-set-name")
        log.Info("YangToDb_community_member_fld_xfmr: ***** setName ", setName)
     //   setName = "test102"
        if len(setName) == 0 {
            err = errors.New("set name is missing");
            log.Error("Set Name is Missing")
            return res_map, err
        }

        prev_type, _ := community_set_action_get_by_set_name (inParams.d, setName, "COMMUNITY_SET");
 /* 
        key := "COMMUNITY_SET" + "|" + setName
        data := (*inParams.dbDataMap)[inParams.curDb]
        log.Info("YangToDb_community_member_fld_xfmr", data)

        pTbl := data["COMMUNITY_SET"]
        if _, ok := pTbl[key]; !ok {
            log.Info("YangToDb_community_member_fld_xfmr not found : ", key)
        } else {
            setTbl := pTbl[key]
            prev_type, ok = setTbl.Field["type"]
            if ok {
                log.Info("Previous type ", prev_type)
            } else {
                log.Info("New Table, No previous type ", prev_type)
            }
        }
*/
        numCommunity = 0
        members := inParams.param.([]ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union)

        for _, member := range members {

	memberType := reflect.TypeOf(member).Elem()
	log.Info("YangToDb_community_member_fld_xfmr: ", member, " memberType: ", memberType)
        var b bytes.Buffer
        switch memberType {
        case reflect.TypeOf(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_E_OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY{}):
            v := (member).(*ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_E_OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY)
            switch v.E_OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY {
            case ocbinds.OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY_NOPEER:
                community[numCommunity] = "no_peer"
                break
            case ocbinds.OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY_NO_ADVERTISE:
                community[numCommunity] = "no_advertise"
                break
            case ocbinds.OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY_NO_EXPORT:
                community[numCommunity] = "no_export"
                break
            case ocbinds.OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY_NO_EXPORT_SUBCONFED:
                //community = "no_export_subconfed"
                break
            }
            new_type = "STANDARD"
            break
        case reflect.TypeOf(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_Uint32{}):
            v := (member).(*ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_Uint32)
            fmt.Fprintf(&b, "0x%x", v.Uint32)
            community[numCommunity] = b.String()
            new_type = "STANDARD"
            break
        case reflect.TypeOf(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_String{}):
            v := (member).(*ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_String)
            has_regex := strings.HasPrefix(v.String, "regex:")
            if has_regex == true {
                new_type = "EXPANDARD"
            } else {
                new_type = "STANDARD"
            }
          //  community[numCommunity] = strings.TrimPrefix(v.String, "regex:")
          //  community[numCommunity] = v.String
	    log.Info("YangToDb_community_member_fld_xfmr: ", v.String, " new_type: ", new_type)
            community = append(community,v.String)
         //   community[numCommunity] = "5:5"
            break
        }
        if len(prev_type) > 0 {
            if prev_type != new_type {
                log.Error("YangToDb_community_member_fld_xfmr: Type Difference, Error previous", prev_type, " newType: ", new_type)
                err = errors.New("Type difference");
                return nil, err
            } else {
              prev_type = new_type
            }
        }

        log.Info("YangToDb_community_member_fld_xfmr: ", community[numCommunity])

        numCommunity++
    }
    var i int
    for i = 0; i < numCommunity - 1; i++ {
        community_list = community[i] + ","
    }
    community_list = community[i]
    res_map["community_member"] = community_list
    res_map["type"] = new_type

    log.Info("YangToDb_community_member_fld_xfmr: ", res_map["community_member"], " type ", res_map["type"])
    return res_map, err
}

