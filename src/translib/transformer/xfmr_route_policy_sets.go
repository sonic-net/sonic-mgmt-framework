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
    SONIC_MATCH_SET_ACTION_ANY = "ANY"
    SONIC_MATCH_SET_ACTION_ALL = "ALL"
)

/* E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode */
var PREFIX_SET_MODE_MAP = map[string]string{
    strconv.FormatInt(int64(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode_IPV4), 10): SONIC_PREFIX_SET_MODE_IPV4,
    strconv.FormatInt(int64(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode_IPV6), 10): SONIC_PREFIX_SET_MODE_IPV6,
}

/* ocbinds.E_OpenconfigRoutingPolicy_MatchSetOptionsType */
var MATCH_SET_ACTION_MAP = map[string]string{
    strconv.FormatInt(int64(ocbinds.OpenconfigRoutingPolicy_MatchSetOptionsType_ALL), 10): SONIC_MATCH_SET_ACTION_ALL,
    strconv.FormatInt(int64(ocbinds.OpenconfigRoutingPolicy_MatchSetOptionsType_ANY), 10): SONIC_MATCH_SET_ACTION_ANY,
}

func init () {
    XlateFuncBind("YangToDb_prefix_empty_set_name_fld_xfmr", YangToDb_prefix_empty_set_name_fld_xfmr)
    XlateFuncBind("YangToDb_prefix_set_name_fld_xfmr", YangToDb_prefix_set_name_fld_xfmr)
    XlateFuncBind("DbToYang_prefix_set_name_fld_xfmr", DbToYang_prefix_set_name_fld_xfmr)
    XlateFuncBind("YangToDb_prefix_set_mode_fld_xfmr", YangToDb_prefix_set_mode_fld_xfmr)
    XlateFuncBind("DbToYang_prefix_set_mode_fld_xfmr", DbToYang_prefix_set_mode_fld_xfmr)
    XlateFuncBind("YangToDb_prefix_key_xfmr", YangToDb_prefix_key_xfmr)
    XlateFuncBind("DbToYang_prefix_key_xfmr", DbToYang_prefix_key_xfmr)
    XlateFuncBind("YangToDb_prefix_empty_ip_prefix_fld_xfmr", YangToDb_prefix_empty_ip_prefix_fld_xfmr)
    XlateFuncBind("YangToDb_prefix_ip_prefix_fld_xfmr", YangToDb_prefix_ip_prefix_fld_xfmr)
    XlateFuncBind("DbToYang_prefix_ip_prefix_fld_xfmr", DbToYang_prefix_ip_prefix_fld_xfmr)
    XlateFuncBind("YangToDb_prefix_empty_masklength_range_fld_xfmr", YangToDb_prefix_empty_masklength_range_fld_xfmr)
    XlateFuncBind("YangToDb_prefix_masklength_range_fld_xfmr", YangToDb_prefix_masklength_range_fld_xfmr)
    XlateFuncBind("DbToYang_prefix_masklength_range_fld_xfmr", DbToYang_prefix_masklength_range_fld_xfmr)

    XlateFuncBind("YangToDb_community_set_name_fld_xfmr", YangToDb_community_set_name_fld_xfmr)
    XlateFuncBind("DbToYang_community_set_name_fld_xfmr", DbToYang_community_set_name_fld_xfmr)
    XlateFuncBind("YangToDb_community_match_set_options_fld_xfmr", YangToDb_community_match_set_options_fld_xfmr)
    XlateFuncBind("DbToYang_community_match_set_options_fld_xfmr", DbToYang_community_match_set_options_fld_xfmr)
    XlateFuncBind("YangToDb_community_member_fld_xfmr", YangToDb_community_member_fld_xfmr)
    XlateFuncBind("DbToYang_community_member_fld_xfmr", DbToYang_community_member_fld_xfmr)

    XlateFuncBind("YangToDb_ext_community_set_name_fld_xfmr", YangToDb_ext_community_set_name_fld_xfmr)
    XlateFuncBind("DbToYang_ext_community_set_name_fld_xfmr", DbToYang_ext_community_set_name_fld_xfmr)
    XlateFuncBind("YangToDb_ext_community_match_set_options_fld_xfmr", YangToDb_ext_community_match_set_options_fld_xfmr)
    XlateFuncBind("DbToYang_ext_community_match_set_options_fld_xfmr", DbToYang_ext_community_match_set_options_fld_xfmr)
    XlateFuncBind("YangToDb_ext_community_member_fld_xfmr", YangToDb_ext_community_member_fld_xfmr)
    XlateFuncBind("DbToYang_ext_community_member_fld_xfmr", DbToYang_ext_community_member_fld_xfmr)

    XlateFuncBind("YangToDb_as_path_set_name_fld_xfmr", YangToDb_as_path_set_name_fld_xfmr)
    XlateFuncBind("DbToYang_as_path_set_name_fld_xfmr", DbToYang_as_path_set_name_fld_xfmr)

    XlateFuncBind("YangToDb_neighbor_set_name_fld_xfmr", YangToDb_neighbor_set_name_fld_xfmr)
    XlateFuncBind("DbToYang_neighbor_set_name_fld_xfmr", DbToYang_neighbor_set_name_fld_xfmr)

    XlateFuncBind("YangToDb_tag_set_name_fld_xfmr", YangToDb_tag_set_name_fld_xfmr)
    XlateFuncBind("DbToYang_tag_set_name_fld_xfmr", DbToYang_tag_set_name_fld_xfmr)
}

var YangToDb_prefix_empty_set_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_prefix_empty_set_name_fld_xfmr: ", inParams.key)
    return res_map, nil
}

var YangToDb_prefix_set_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_prefix_cfg_set_name_fld_xfmr: ", inParams.key)
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
    log.Info("DbToYang_prefix_set_mode_fld_xfmr: Input", data, inParams.ygRoot)
    mode, ok := data["PREFIX_SET"][inParams.key].Field["mode"]
    if ok {
        log.Info("DbToYang_prefix_set_mode_fld_xfmr **** ", mode)
        oc_mode := findInMap(PREFIX_SET_MODE_MAP, mode)
        n, err := strconv.ParseInt(oc_mode, 10, 64)
        result["mode"] = ocbinds.E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode(n).ΛMap()["E_OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_PrefixSets_PrefixSet_Config_Mode"][n].Name
        log.Info("DbToYang_prefix_set_mode_fld_xfmr ", result)
        return result, err
    }
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

    if ((inParams.oper == DELETE) && (len(pathInfo.Vars) == 1)) {
        setName = pathInfo.Var("name")
        if len(setName) == 0 {
            err = errors.New("YangToDb_prefix_key_xfmr: Prefix set name is missing");
            log.Error("YangToDb_prefix_key_xfmr: Prefix set name is Missing")
            return setName, err
        }
        // TODO - This Case will not come for CLI, Riht now return dummy key to avoid DB flush
        //   return prefix_del_by_set_name (inParams.d, setName, "PREFIX")
        return "NULL", nil
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

var YangToDb_prefix_empty_ip_prefix_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_prefix_empty_ip_prefix_fld_xfmr: ", inParams.key)
    return res_map, nil
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

var YangToDb_prefix_empty_masklength_range_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_prefix_empty_masklength_range_fld_xfmr: ", inParams.key)
    return res_map, nil
}

var YangToDb_prefix_masklength_range_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_prefix_masklength_range_fld_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
    return res_map, nil
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

/* COMMUNITY SET API's */
var YangToDb_community_set_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_community_set_name_fld_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
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

    res_map["community-set-name"] = setName
    log.Info("config/name  ", res_map)
    return res_map, err
}

func community_set_match_options_get_by_set_name (d *db.DB , setName string, tblName string) (string, error) {
    var err error

    dbspec := &db.TableSpec { Name: tblName }

    log.Info("community_set_match_options_get_by_set_name: key  ", db.Key{Comp: []string{setName}})
    dbEntry, err := d.GetEntry(dbspec, db.Key{Comp: []string{setName}})
    if err != nil {
        log.Error("No Entry found e = ", err)
        return "", err
    }
    match_action, ok := dbEntry.Field["match_action"]
    if ok {
        log.Info("Previous Match options ", match_action)
    } else {
        log.Info("New Table, No previous match option ", match_action)
    }
    return match_action, nil;
}

var YangToDb_community_match_set_options_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    if inParams.param == nil {
        res_map["match_action"] = ""
        return res_map, err
    }

    log.Info("YangToDb_community_match_set_options_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri)

    pathInfo := NewPathInfo(inParams.uri)
    if len(pathInfo.Vars) <  1 {
        err = errors.New("Invalid Key length");
        log.Error("Invalid Key length", len(pathInfo.Vars))
        return res_map, err
    }

    setName := pathInfo.Var("community-set-name")
    log.Info("YangToDb_community_match_set_options_fld_xfmr: setName ", setName)
    if len(setName) == 0 {
        err = errors.New("set name is missing");
        log.Error("Set Name is Missing")
        return res_map, err
    }

    prev_match_action, _ := community_set_match_options_get_by_set_name (inParams.d, setName, "COMMUNITY_SET");

    match_opt, _ := inParams.param.(ocbinds.E_OpenconfigRoutingPolicy_MatchSetOptionsType)
    new_match_action := findInMap(MATCH_SET_ACTION_MAP, strconv.FormatInt(int64(match_opt), 10))
    log.Info("YangToDb_community_match_set_options_fld_xfmr: New match Opt: ", new_match_action)
    if len(prev_match_action) > 0 {
        if prev_match_action != new_match_action {
            log.Error("YangToDb_community_match_set_options_fld_xfmr: Match option difference, Error previous", prev_match_action, " new ", new_match_action)
            err = errors.New("Match option difference");
            return nil, err
        } else {
            prev_match_action = new_match_action
        }
    }

    res_map["match_action"] = new_match_action

    return res_map, err
}

var DbToYang_community_match_set_options_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    log.Info("DbToYang_community_match_set_options_fld_xfmr", inParams.ygRoot)
    data := (*inParams.dbDataMap)[inParams.curDb]
    opt, ok := data["COMMUNITY_SET"][inParams.key].Field["match_action"]
    if ok {
        match_opt := findInMap(MATCH_SET_ACTION_MAP, opt)
        n, err := strconv.ParseInt(match_opt, 10, 64)
        result["match-set-options"] = ocbinds.E_OpenconfigRoutingPolicy_MatchSetOptionsType(n).ΛMap()["E_OpenconfigRoutingPolicy_MatchSetOptionsType"][n].Name
        log.Info("DbToYang_community_match_set_options_fld_xfmr ", result["match-set-options"])
        return result, err
    }
    return result, err
}

func community_set_type_get_by_set_name (d *db.DB , setName string, tblName string) (string, error) {
    var err error

    dbspec := &db.TableSpec { Name: tblName }

    log.Info("community_set_type_get_by_set_name: key  ", db.Key{Comp: []string{setName}})
    dbEntry, err := d.GetEntry(dbspec, db.Key{Comp: []string{setName}})
    if err != nil {
        log.Error("No Entry found e = ", err)
        return "", err
    }
    prev_type, ok := dbEntry.Field["set_type"]
    if ok {
        log.Info("Previous type ", prev_type)
    } else {
        log.Info("New Table, No previous type ", prev_type)
    }
    return prev_type, nil;
}

func community_set_is_community_members_exits (d *db.DB , setName string, tblName string, fieldName string) (bool, error) {
    var err error
    var community_list string

    dbspec := &db.TableSpec { Name: tblName }

    log.Info("community_set_is_community_members_exits: key  ", db.Key{Comp: []string{setName}})
    dbEntry, err := d.GetEntry(dbspec, db.Key{Comp: []string{setName}})
    if err != nil {
        log.Error("No Entry found e = ", err)
        return false, err
    }

    community_list, ok := dbEntry.Field[fieldName]
    if ok {
        if len(community_list) > 0 {
            log.Info("community_set_is_community_members_exits: Comminuty members eixts")
            return true, nil
        }
    } else {
        log.Info("community_set_is_community_members_exits: No Comminuty members eixts ")
    }

    return false, nil
}

var YangToDb_community_member_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    var community_list string
    var new_type string
    var prev_type string

    log.Info("YangToDb_community_member_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, "inParams : ", inParams)
    if inParams.param == nil {
        res_map["community_member@"] = ""
        return res_map, errors.New("Invalid Inputs")
    }

    pathInfo := NewPathInfo(inParams.uri)
    if len(pathInfo.Vars) <  1 {
        err = errors.New("Invalid Key length");
        log.Error("Invalid Key length", len(pathInfo.Vars))
        return res_map, err
    }

    setName := pathInfo.Var("community-set-name")
    log.Info("YangToDb_community_member_fld_xfmr: setName ", setName)
    if len(setName) == 0 {
        err = errors.New("set name is missing");
        log.Error("Set Name is Missing")
        return res_map, err
    }
    is_member_exits, _ := community_set_is_community_members_exits (inParams.d, setName, "COMMUNITY_SET", "community_member@");
    if is_member_exits == true {
        prev_type, _ = community_set_type_get_by_set_name (inParams.d, setName, "COMMUNITY_SET");

        log.Info("YangToDb_community_member_fld_xfmr: prev_type ", prev_type)
    }
    members := inParams.param.([]ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union)

    for _, member := range members {

        memberType := reflect.TypeOf(member).Elem()
        log.Info("YangToDb_community_member_fld_xfmr: member - ", member, " memberType: ", memberType)
        var b bytes.Buffer
        switch memberType {

        case reflect.TypeOf(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_E_OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY{}):
            v := (member).(*ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_E_OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY)
            switch v.E_OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY {
            case ocbinds.OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY_NOPEER:
                community_list += "local-AS" + ","
                break
            case ocbinds.OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY_NO_ADVERTISE:
                community_list += "no-advertise" + ","
                break
            case ocbinds.OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY_NO_EXPORT:
                community_list += "no-export" + ","
                break
            case ocbinds.OpenconfigBgpTypes_BGP_WELL_KNOWN_STD_COMMUNITY_NO_EXPORT_SUBCONFED:
                err = errors.New("Un supported BGP well known type NO_EXPORT_SUBCONFED");
                return res_map, err
            }
            new_type = "STANDARD"
            break
        case reflect.TypeOf(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_Uint32{}):
            v := (member).(*ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_Uint32)
            fmt.Fprintf(&b, "%d", v.Uint32)
            community_list += b.String() + ","
            new_type = "STANDARD"
            break
        case reflect.TypeOf(ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_String{}):
            v := (member).(*ocbinds.OpenconfigRoutingPolicy_RoutingPolicy_DefinedSets_BgpDefinedSets_CommunitySets_CommunitySet_Config_CommunityMember_Union_String)

            has_regex := strings.HasPrefix(v.String, "REGEX:")
            if has_regex == true {
                new_type = "EXPANDED"
            } else {
                new_type = "STANDARD"
            }
            community_list += strings.TrimPrefix(v.String, "REGEX:") + ","
            break
        }

        log.Info("YangToDb_community_member_fld_xfmr: new_type: ", new_type, " prev_type ", prev_type)
        if ((len(prev_type) > 0) && (prev_type != new_type)){
            log.Error("YangToDb_community_member_fld_xfmr: Type Difference Error, previous", prev_type, " newType: ", new_type)
            err = errors.New("Type difference, Quit Operation");
            return res_map, err
        } else {
            prev_type = new_type
        }
    }

    res_map["community_member@"] = strings.TrimSuffix(community_list, ",")
    res_map["set_type"] = new_type

    log.Info("YangToDb_community_member_fld_xfmr: ", res_map["community_member@"], " type ", res_map["set_type"])
    return res_map, err
}

var DbToYang_community_member_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})
    var result_community string
    data := (*inParams.dbDataMap)[inParams.curDb]

    log.Info("DbToYang_community_member_fld_xfmr", data, inParams.ygRoot, inParams.key)

    set_type := data["COMMUNITY_SET"][inParams.key].Field["set_type"]

    log.Info("DbToYang_community_member_fld_xfmr: type ", set_type)
    var Communities []interface{}

    community_list, ok := data["COMMUNITY_SET"][inParams.key].Field["community_member@"]
    if ok {
        log.Info("DbToYang_community_member_fld_xfmr: DB Memebers ", community_list)
        for _, community := range strings.Split(community_list, ",") {
            if set_type == "EXPANDED" {
                result_community = "REGEX:"
            } else  {
                result_community = ""
            }

            if (community == "local-AS") {
                result_community += "NOPEER"
            } else if (community == "no-advertise") {
                result_community += "NO_ADVERTISE"
            } else if (community == "no-export") {
                result_community += "NO_EXPORT"
            } else {
                result_community += community
            }
            log.Info("DbToYang_community_member_fld_xfmr: result_community ", result_community)
            Communities = append(Communities, result_community)
        }
    }
    result["community-member"] = Communities
    log.Info("DbToYang_community_member_fld_xfmr: Comminuty Memebers ", result["community-member"])
    return result, err
}

/* EXTENDED COMMUNITY SET API's */
var YangToDb_ext_community_set_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_ext_community_set_name_fld_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_ext_community_set_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_ext_community_set_name_fld_xfmr: ", inParams.key)
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    log.Info("DbToYang_ext_community_set_name_fld_xfmr: ", key)
    setTblKey := strings.Split(key, "|")
    setName := setTblKey[0]

    res_map["ext-community-set-name"] = setName
    log.Info("config/name  ", res_map)
    return res_map, err
}

var YangToDb_ext_community_match_set_options_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    if inParams.param == nil {
        res_map["match_action"] = ""
        return res_map, err
    }

    log.Info("YangToDb_ext_community_match_set_options_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri)

    pathInfo := NewPathInfo(inParams.uri)
    if len(pathInfo.Vars) <  1 {
        err = errors.New("Invalid Key length");
        log.Error("Invalid Key length", len(pathInfo.Vars))
        return res_map, err
    }

    setName := pathInfo.Var("ext-community-set-name")
    log.Info("YangToDb_ext_community_match_set_options_fld_xfmr: setName ", setName)
    if len(setName) == 0 {
        err = errors.New("set name is missing");
        log.Error("Set Name is Missing")
        return res_map, err
    }

    prev_match_action, _ := community_set_match_options_get_by_set_name (inParams.d, setName, "EXTENDED_COMMUNITY_SET");

    match_opt, _ := inParams.param.(ocbinds.E_OpenconfigRoutingPolicy_MatchSetOptionsType)
    new_match_action := findInMap(MATCH_SET_ACTION_MAP, strconv.FormatInt(int64(match_opt), 10))
    log.Info("YangToDb_ext_community_match_set_options_fld_xfmr: New match Opt: ", new_match_action)
    if len(prev_match_action) > 0 {
        if prev_match_action != new_match_action {
            log.Error("YangToDb_ext_community_match_set_options_fld_xfmr: Match option difference, Error previous", prev_match_action, " new ", new_match_action)
            err = errors.New("Match option difference");
            return nil, err
        } else {
            prev_match_action = new_match_action
        }
    }

    res_map["match_action"] = new_match_action

    return res_map, err
}

var DbToYang_ext_community_match_set_options_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    log.Info("DbToYang_ext_community_match_set_options_fld_xfmr", inParams.ygRoot)
    data := (*inParams.dbDataMap)[inParams.curDb]
    opt, ok := data["EXTENDED_COMMUNITY_SET"][inParams.key].Field["match_action"]
    if ok {
        match_opt := findInMap(MATCH_SET_ACTION_MAP, opt)
        n, err := strconv.ParseInt(match_opt, 10, 64)
        result["match-set-options"] = ocbinds.E_OpenconfigRoutingPolicy_MatchSetOptionsType(n).ΛMap()["E_OpenconfigRoutingPolicy_MatchSetOptionsType"][n].Name
        log.Info("DbToYang_ext_community_match_set_options_fld_xfmr ", result["match-set-options"])
        return result, err
    }
    return result, err
}

var YangToDb_ext_community_member_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    var community_list string
    var new_type string
    var prev_type string

    log.Info("YangToDb_ext_community_member_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, "inParams : ", inParams)
    if inParams.param == nil {
        res_map["community_member@"] = ""
        return res_map, errors.New("Invalid Inputs")
    }

    pathInfo := NewPathInfo(inParams.uri)
    if len(pathInfo.Vars) <  1 {
        err = errors.New("Invalid Key length");
        log.Error("Invalid Key length", len(pathInfo.Vars))
        return res_map, err
    }

    setName := pathInfo.Var("ext-community-set-name")
    log.Info("YangToDb_ext_community_member_fld_xfmr: setName ", setName)
    if len(setName) == 0 {
        err = errors.New("set name is missing");
        log.Error("Set Name is Missing")
        return res_map, err
    }
    is_member_exits, _ := community_set_is_community_members_exits (inParams.d, setName, "EXTENDED_COMMUNITY_SET", "community_member@");
    if is_member_exits == true {
        prev_type, _ = community_set_type_get_by_set_name (inParams.d, setName, "EXTENDED_COMMUNITY_SET");

        log.Info("YangToDb_ext_community_member_fld_xfmr: prev_type ", prev_type)
    }

    members := inParams.param.([]string)

    log.Info("YangToDb_ext_community_member_fld_xfmr: members", members)
    for _, member := range members {

        has_regex := strings.HasPrefix(member, "REGEX:")
        if has_regex == true {
            new_type = "EXPANDED"
        } else {
            new_type = "STANDARD"
        }
        member = strings.TrimPrefix(member, "REGEX:")

        has_rt := strings.HasPrefix(member, "route-target")
        has_ro := strings.HasPrefix(member, "route-origin")
        if ((has_rt == false) && (has_ro == false)){
            err = errors.New("Community member is not of type route-target or route-origin");
            log.Error("Community member is not of type route-target or route-origin")
            return res_map, err
        }
        community_list += member + ","
        log.Info("YangToDb_ext_community_member_fld_xfmr: new_type: ", new_type, " prev_type ", prev_type)
        if ((len(prev_type) > 0) && (prev_type != new_type)){
            log.Error("YangToDb_ext_community_member_fld_xfmr: Type Difference Error, previous", prev_type, " newType: ", new_type)
            err = errors.New("Type difference, Quit Operation");
            return res_map, err
        } else {
            prev_type = new_type
        }
    }
    res_map["community_member@"] = strings.TrimSuffix(community_list, ",")
    res_map["set_type"] = new_type
    log.Info("YangToDb_ext_community_member_fld_xfmr: ", res_map["community_member@"], " type ", res_map["set_type"])
    return res_map, err
}

var DbToYang_ext_community_member_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})
    var result_community string
    data := (*inParams.dbDataMap)[inParams.curDb]

    log.Info("DbToYang_ext_community_member_fld_xfmr", data, inParams.ygRoot, inParams.key)

    set_type := data["EXTENDED_COMMUNITY_SET"][inParams.key].Field["set_type"]

    log.Info("DbToYang_ext_community_member_fld_xfmr: type ", set_type)
    var Communities []interface{}

    community_list, ok := data["EXTENDED_COMMUNITY_SET"][inParams.key].Field["community_member@"]
    if ok {
        log.Info("DbToYang_ext_community_member_fld_xfmr: DB Memebers ", community_list)
        for _, community := range strings.Split(community_list, ",") {
            if set_type == "EXPANDED" {
                result_community = "REGEX:"
            } else  {
                result_community = ""
            }
            result_community += community
            log.Info("DbToYang_ext_community_member_fld_xfmr: result_community ", result_community)
            Communities = append(Communities, result_community)
        }
    }
    result["ext-community-member"] = Communities
    log.Info("DbToYang_ext_community_member_fld_xfmr: Comminuty Memebers ", result["community-member"])
    return result, err
}

/* AS PATH SET API's */
var YangToDb_as_path_set_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_as_path_set_name_fld_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_as_path_set_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    log.Info("DbToYang_as_path_set_name_fld_xfmr: ", key)
    setTblKey := strings.Split(key, "|")
    setName := setTblKey[0]

    res_map["as-path-set-name"] = setName
    log.Info("config/name  ", res_map)
    return res_map, err
}

/* NEIGHBOR SET API's */
var YangToDb_neighbor_set_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_neighbor_set_name_fld_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_neighbor_set_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    log.Info("DbToYang_neighbor_set_name_fld_xfmr: ", key)
    setTblKey := strings.Split(key, "|")
    setName := setTblKey[0]

    res_map["name"] = setName
    log.Info("config/name  ", res_map)
    return res_map, err
}

/* TAG SET API's */
var YangToDb_tag_set_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_tag_set_name_fld_xfmr: ", inParams.key)
    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_tag_set_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    /*name attribute corresponds to key in redis table*/
    key := inParams.key
    log.Info("DbToYang_tag_set_name_fld_xfmr: ", key)
    setTblKey := strings.Split(key, "|")
    setName := setTblKey[0]

    res_map["name"] = setName
    log.Info("config/name  ", res_map)
    return res_map, err
}
