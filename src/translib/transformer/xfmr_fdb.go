package transformer

import (
    "errors"
    "strings"
    "strconv"
    "translib/ocbinds"
    log "github.com/golang/glog"
)


func init () {
    XlateFuncBind("YangToDb_fdb_tbl_key_xfmr", YangToDb_fdb_tbl_key_xfmr)
    XlateFuncBind("DbToYang_fdb_tbl_key_xfmr", DbToYang_fdb_tbl_key_xfmr)
    XlateFuncBind("YangToDb_entry_type_field_xfmr", YangToDb_entry_type_field_xfmr)
    XlateFuncBind("DbToYang_entry_type_field_xfmr", DbToYang_entry_type_field_xfmr)
}

const (
	FDB_TABLE                = "FDB_TABLE"
	ENTRY_TYPE_STATIC        = "STATIC"
	ENTRY_TYPE_DYNAMIC       = "DYNAMIC"
	ENTRY_TYPE               = "type"
)

/* E_OpenconfigNetworkInstance_ENTRY_TYPE */
var FDB_ENTRY_TYPE_MAP = map[string]string{
	strconv.FormatInt(int64(ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType_STATIC), 10): ENTRY_TYPE_STATIC,
	strconv.FormatInt(int64(ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType_DYNAMIC), 10): ENTRY_TYPE_DYNAMIC,
}


var YangToDb_fdb_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var entry_key string
    var err error

    log.Info("YangToDb_bgp_gbl_tbl_key_xfmr: **** inParams ****", inParams)
    pathInfo := NewPathInfo(inParams.uri)
    vlanName := pathInfo.Var("vlan")
    macAddress := pathInfo.Var("mac-address")

    log.Info("YangToDb_bgp_gbl_tbl_key_xfmr: **** path_info ****", pathInfo)
    log.Info("YangToDb_bgp_gbl_tbl_key_xfmr: **** VLAN ****", vlanName)
    log.Info("YangToDb_bgp_gbl_tbl_key_xfmr: **** mac-address ****", macAddress)

    vlanString := strings.Contains(vlanName,"Vlan")
    if vlanString == false {
        vlanName = "Vlan"+vlanName
    }
    entry_key = vlanName + ":" + macAddress
    log.Info("YangToDb_bgp_gbl_tbl_key_xfmr: **** Final-Key ****", entry_key)
    return entry_key, err
}

var DbToYang_fdb_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_tbl_key: **** entry_key ****", entry_key)

    keyMap := strings.SplitN(entry_key, ":",2)
    if len(keyMap) < 2 {
        err = errors.New("Invalid key for INTERFACE table entry.")
        log.Info("Invalid Keys for INTERFACE table entry : ", entry_key)
        return rmap, err
    }

    vlanName := keyMap[0]
    macAddress := keyMap[1]
    log.Info("DbToYang_bgp_gbl_tbl_key: **** VLAN: ****", vlanName)
    log.Info("DbToYang_bgp_gbl_tbl_key: **** mac-address ****", macAddress)

    rmap["vlan"] = vlanName
    rmap["mac-address"] = macAddress
    log.Info("DbToYang_bgp_gbl_tbl_key: **** rmap ****", rmap)

    return rmap, err
}

var YangToDb_entry_type_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    log.Info("YangToDb ---- FIELD TRANSFORMER: **** res_map ****", res_map)
    var err error
    if inParams.param == nil {
        res_map[ENTRY_TYPE] = ""
	return res_map, err
    }

    entrytype, _ := inParams.param.(ocbinds.E_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType)
    log.Info("YangToDb ---- FIELD TRANSFORMER: **** ", inParams.ygRoot, " Xpath: ", inParams.uri, " ENTRY-TYPE: ", entrytype)
    res_map[ENTRY_TYPE] = findInMap(FDB_ENTRY_TYPE_MAP, strconv.FormatInt(int64(entrytype), 10))
    log.Info("YangToDb ---- FIELD TRANSFORMER: **** FINAL res_map ****", res_map)
    return res_map, err
}
var DbToYang_entry_type_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})
    log.Info("DbToYang ---- FIELD TRANSFORMER: **** PARAMS: ****", inParams)
    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang ---- FIELD TRANSFORMER: **** result ****   data: ", data)
    log.Info("DbToYang ---- FIELD TRANSFORMER: **** result ****   key: ", inParams.key)
    log.Info("DbToYang ---- FIELD TRANSFORMER: **** result ***    TESTING123123")
    oc_entrytype := findInMap(FDB_ENTRY_TYPE_MAP, data[FDB_TABLE][inParams.key].Field[ENTRY_TYPE])
    log.Info("DbToYang ---- FIELD TRANSFORMER: **** oc_entrytype ****", oc_entrytype)
    n, err := strconv.ParseInt(oc_entrytype, 10, 64)
    log.Info("DbToYang ---- FIELD TRANSFORMER: **** n: ****", n)
    result[ENTRY_TYPE] = ocbinds.E_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType(n).ΛMap()["E_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType"][n].Name
    log.Info("DbToYang ---- FIELD TRANSFORMER: **** FINAL: result ****", result)
    return result, err
}
