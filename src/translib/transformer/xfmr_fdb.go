package transformer

import (
    "errors"
    "strings"
    "strconv"
    "translib/ocbinds"
    "translib/db"
    "encoding/json"
    log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("YangToDb_fdb_tbl_key_xfmr", YangToDb_fdb_tbl_key_xfmr)
    XlateFuncBind("DbToYang_fdb_tbl_key_xfmr", DbToYang_fdb_tbl_key_xfmr)
    XlateFuncBind("YangToDb_entry_type_field_xfmr", YangToDb_entry_type_field_xfmr)
    XlateFuncBind("DbToYang_entry_type_field_xfmr", DbToYang_entry_type_field_xfmr)
    XlateFuncBind("rpc_clear_fdb", rpc_clear_fdb)
}

const (
	FDB_TABLE                = "FDB_TABLE"
        SONIC_ENTRY_TYPE_STATIC  = "static"
        SONIC_ENTRY_TYPE_DYNAMIC = "dynamic"
	ENTRY_TYPE               = "entry-type"
)

/* E_OpenconfigNetworkInstance_ENTRY_TYPE */
var FDB_ENTRY_TYPE_MAP = map[string]string{
	strconv.FormatInt(int64(ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType_STATIC), 10): SONIC_ENTRY_TYPE_STATIC,
	strconv.FormatInt(int64(ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType_DYNAMIC), 10): SONIC_ENTRY_TYPE_DYNAMIC,
}

var rpc_clear_fdb RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
	var err error
	var  valLst [2]string
	var data  []byte

	valLst[0]= "ALL"
	valLst[1] = "ALL"

	data, err = json.Marshal(valLst)

	if err != nil {
		log.Error("Failed to  marshal input data; err=%v", err)
		return nil, err
	}

	err = dbs[db.ApplDB].Publish("FLUSHFDBREQUEST",data)
	return nil, err
}

var YangToDb_fdb_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var entry_key string
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    vlanName := pathInfo.Var("vlan")
    macAddress := pathInfo.Var("mac-address")

    vlanString := strings.Contains(vlanName,"Vlan")
    if vlanString == false {
        vlanName = "Vlan"+vlanName
    }
    entry_key = vlanName + ":" + macAddress
    return entry_key, err
}

var DbToYang_fdb_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    entry_key := inParams.key

    keyMap := strings.SplitN(entry_key, ":",2)
    if len(keyMap) < 2 {
        err = errors.New("Invalid key for INTERFACE table entry.")
        return rmap, err
    }

    vlanName := keyMap[0]
    macAddress := keyMap[1]
    vlanNumber := strings.SplitN(vlanName, "Vlan",2)
    vlanId, err := strconv.ParseFloat(vlanNumber[1],64)

    rmap["vlan"] = vlanId
    rmap["mac-address"] = macAddress
    return rmap, err
}

var YangToDb_entry_type_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error

    return res_map, err
}

var DbToYang_entry_type_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})
    data := (*inParams.dbDataMap)[inParams.curDb]
    fdbTableMap := data["FDB_TABLE"]
    var entryTypeFinal = ""
    if val, keyExist := fdbTableMap[inParams.key]; keyExist {
        if entryType, ok := val.Field["type"]; ok {
            entryTypeFinal = entryType
        } else {
            return result, err
        }
    } else {
        return result, err
    }
    oc_entrytype := findInMap(FDB_ENTRY_TYPE_MAP, entryTypeFinal)
    n, err := strconv.ParseInt(oc_entrytype, 10, 64)
    result[ENTRY_TYPE] = ocbinds.E_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType(n).ΛMap()["E_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType"][n].Name

    return result, err
}
