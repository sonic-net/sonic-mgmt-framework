package transformer

import (
    "errors"
    "strings"
    "strconv"
    "translib/ocbinds"
    "github.com/openconfig/ygot/ygot"
    "translib/db"
    "encoding/json"
    log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("YangToDb_fdb_mac_table_xfmr", YangToDb_fdb_mac_table_xfmr)
    XlateFuncBind("DbToYang_fdb_mac_table_xfmr", DbToYang_fdb_mac_table_xfmr)
    XlateFuncBind("rpc_clear_fdb", rpc_clear_fdb)
}

const (
    FDB_TABLE                = "FDB_TABLE"
    SONIC_ENTRY_TYPE_STATIC  = "SAI_FDB_ENTRY_TYPE_STATIC"
    SONIC_ENTRY_TYPE_DYNAMIC = "SAI_FDB_ENTRY_TYPE_DYNAMIC"
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


func getFdbMacTableRoot (s *ygot.GoStruct, instance string, build bool) *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable {
    var fdbMacTableObj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable

    deviceObj := (*s).(*ocbinds.Device)
    niObj := deviceObj.NetworkInstances

    if instance == "" {
        instance = "default"
    }
    if niObj != nil {
        if niObj.NetworkInstance != nil && len(niObj.NetworkInstance) > 0 {
            if _, ok := niObj.NetworkInstance[instance]; ok {
                niInst := niObj.NetworkInstance[instance]
                if niInst.Fdb != nil {
                    fdbMacTableObj = niInst.Fdb.MacTable
                }
            }
        }
    }

    if fdbMacTableObj == nil && build == true {
        if niObj.NetworkInstance == nil || len(niObj.NetworkInstance) < 1 {
            ygot.BuildEmptyTree(niObj)
        }
        var niInst *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance
        if _, ok := niObj.NetworkInstance[instance]; !ok {
            niInst, _  = niObj.NewNetworkInstance(instance)
        } else {
            niInst = niObj.NetworkInstance[instance]
        }
        ygot.BuildEmptyTree(niInst)
        if niInst.Fdb.MacTable == nil {
            ygot.BuildEmptyTree(niInst.Fdb)
        }
        fdbMacTableObj = niInst.Fdb.MacTable

    }

    return fdbMacTableObj
}

var YangToDb_fdb_mac_table_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    retMap := make(map[string]map[string]db.Value)

    return retMap, nil
}

func getOidToIntfNameMap (d *db.DB) (map[string]string, error) {
    tblTs := &db.TableSpec{Name:"COUNTERS_PORT_NAME_MAP"}
    oidToIntf :=  make(map[string]string)
    intfOidEntry, err := d.GetMapAll(tblTs)
    if err != nil || !intfOidEntry.IsPopulated() {
        log.Error("Reading Port OID map failed.", err)
        return oidToIntf, err
    }
    for intf, oid := range intfOidEntry.Field {
        oidToIntf[oid] = intf
    }

    return oidToIntf, nil
}

func getASICStateMaps (d *db.DB) (map[string]string, map[string]string, map[string]map[string]db.Value, error) {
    oidTOVlan := make(map[string]string)
    brPrtOidToIntfOid := make(map[string]string)
    fdbMap := make(map[string]map[string]db.Value)

    tblName := "ASIC_STATE"
    vlanPrefix := "SAI_OBJECT_TYPE_VLAN"
    bridgePortPrefix := "SAI_OBJECT_TYPE_BRIDGE_PORT"
    fdbPrefix := "SAI_OBJECT_TYPE_FDB_ENTRY"

    keys, tblErr := d.GetKeys(&db.TableSpec{Name:tblName, CompCt:2} )
    if tblErr != nil {
        log.Error("Get Keys from ASIC_STATE table failed.", tblErr);
        return oidTOVlan, brPrtOidToIntfOid, fdbMap, tblErr
    }

    for _, key := range keys {

        if key.Comp[0] == vlanPrefix {
            vlanKey := key.Comp[1]
            entry, dbErr := d.GetEntry(&db.TableSpec{Name:tblName}, key)
            if dbErr != nil {
                log.Error("DB GetEntry failed for key : ", key)
                continue
            }
            if entry.Has("SAI_VLAN_ATTR_VLAN_ID") == true {
                oidTOVlan[vlanKey] = entry.Get("SAI_VLAN_ATTR_VLAN_ID")
            }
        } else if key.Comp[0] == bridgePortPrefix {
            brPKey := key.Comp[1]
            entry, dbErr := d.GetEntry(&db.TableSpec{Name:tblName}, key)
            if dbErr != nil {
                log.Error("DB GetEntry failed for key : ", key)
                continue
            }
            if entry.Has("SAI_BRIDGE_PORT_ATTR_PORT_ID") == true {
                brPrtOidToIntfOid[brPKey] = entry.Get("SAI_BRIDGE_PORT_ATTR_PORT_ID")
            }
        } else if key.Comp[0] == fdbPrefix {
            jsonData := make(map[string]interface{})
            err := json.Unmarshal([]byte(key.Get(1)), &jsonData)
            if err != nil {
                log.Info("Failed parsing json")
                continue
            }
            bvid := jsonData["bvid"].(string)
            macstr := jsonData["mac"].(string)

            entry, dbErr := d.GetEntry(&db.TableSpec{Name:tblName}, key)
            if dbErr != nil {
                log.Error("DB GetEntry failed for key : ", key)
                continue
            }
            if _, ok := fdbMap[bvid]; !ok {
                fdbMap[bvid] = make(map[string]db.Value)
            }
            fdbMap[bvid][macstr] = entry
        } else {
            continue
        }
    }
    return oidTOVlan, brPrtOidToIntfOid, fdbMap, nil
}

func fdbMacTableGetAll (inParams XfmrParams) error {

    pathInfo := NewPathInfo(inParams.uri)
    instance := pathInfo.Var("name")
    macTbl := getFdbMacTableRoot(inParams.ygRoot, instance, true)
    oidToVlan, brPrtOidToIntfOid, fdbMap, _ := getASICStateMaps(inParams.dbs[db.AsicDB])
    OidInfMap,_  := getOidToIntfNameMap(inParams.dbs[db.CountersDB])

    ygot.BuildEmptyTree(macTbl.Entries)

    for vlanOid, vlanEntry := range fdbMap {
        if _, ok  := oidToVlan[vlanOid]; !ok {
            continue
        }
        vlan := oidToVlan[vlanOid]
        for mac, _ := range vlanEntry {
            fdbMacTableGetEntry(inParams, vlan, mac, OidInfMap, oidToVlan, brPrtOidToIntfOid, fdbMap, macTbl)
        }
    }
    return nil
}

func fdbMacTableGetEntry(inParams XfmrParams, vlan string,  macAddress string, oidInfMap map[string]string, oidTOVlan map[string]string, brPrtOidToIntfOid map[string]string, fdbMap map[string]map[string]db.Value, macTbl *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable) error {
    var err error

    vlanOid := findInMap(oidTOVlan, vlan)
    vlanId, _ := strconv.Atoi(vlan)

    mcEntries := macTbl.Entries
    var mcEntry *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry
    var mcEntryKey ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_Key
    mcEntryKey.MacAddress = macAddress 
    mcEntryKey.Vlan = uint16(vlanId)

    if _, ok := fdbMap[vlanOid]; !ok {
        errStr := "vlanOid entry not found in FDB map, vlanOid: " + vlanOid
        log.Error(errStr)
        return errors.New(errStr)
    }
    if _, ok := fdbMap[vlanOid][macAddress]; !ok {
        errStr := "macAddress entry not found FDB map, macAddress: " + macAddress
        log.Error(errStr)
        return errors.New(errStr)
    }
    entry := fdbMap[vlanOid][macAddress]
    if _, ok := mcEntries.Entry[mcEntryKey]; !ok {
        _, err := mcEntries.NewEntry(macAddress, uint16(vlanId))
        if err != nil {
            log.Error("FDB NewEntry create failed." + vlan + " " + macAddress)
            return errors.New("FDB NewEntry create failed, " + vlan + " " + macAddress)
        }
    }
    mcEntry  = mcEntries.Entry[mcEntryKey]
    ygot.BuildEmptyTree(mcEntry)
    mcMac := new(string)
    mcVlan := new(uint16)
    *mcMac = macAddress
    *mcVlan = uint16(vlanId)
    ygot.BuildEmptyTree(mcEntry.Config)
    mcEntry.Config.MacAddress = mcMac
    mcEntry.Config.Vlan = mcVlan
    ygot.BuildEmptyTree(mcEntry.State)
    mcEntry.State.MacAddress = mcMac
    mcEntry.State.Vlan = mcVlan
    if entry.Has("SAI_FDB_ENTRY_ATTR_TYPE") {
        fdbEntryType := entry.Get("SAI_FDB_ENTRY_ATTR_TYPE")
        if fdbEntryType == SONIC_ENTRY_TYPE_STATIC {
            mcEntry.State.EntryType = ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType_STATIC
        } else {
            mcEntry.State.EntryType = ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Fdb_MacTable_Entries_Entry_State_EntryType_DYNAMIC
        }
    }

    if  entry.Has("SAI_FDB_ENTRY_ATTR_BRIDGE_PORT_ID") {
        intfOid := findInMap(brPrtOidToIntfOid, entry.Get("SAI_FDB_ENTRY_ATTR_BRIDGE_PORT_ID"))
        if intfOid != "" {
            intfName := new(string)
            *intfName = findInMap(oidInfMap, intfOid)
            if *intfName != "" {
                ygot.BuildEmptyTree(mcEntry.Interface)
                ygot.BuildEmptyTree(mcEntry.Interface.InterfaceRef)
                ygot.BuildEmptyTree(mcEntry.Interface.InterfaceRef.Config)
                mcEntry.Interface.InterfaceRef.Config.Interface = intfName
                ygot.BuildEmptyTree(mcEntry.Interface.InterfaceRef.State)
                mcEntry.Interface.InterfaceRef.State.Interface = intfName
            }

        }
    }

    return err
}

var DbToYang_fdb_mac_table_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) (error) {
    var err error
    pathInfo := NewPathInfo(inParams.uri)
    instance := pathInfo.Var("name")
    vlan := pathInfo.Var("vlan")
    macAddress := pathInfo.Var("mac-address")

    targetUriPath, err := getYangPathFromUri(inParams.uri)
    log.Info("targetUriPath is ", targetUriPath)

    macTbl := getFdbMacTableRoot(inParams.ygRoot, instance, true)
    if macTbl == nil {
        log.Info("DbToYang_fdb_mac_table_xfmr - getFdbMacTableRoot returned nil, for URI: ", inParams.uri)
        return errors.New("Not able to get FDB MacTable root.");
    }

    ygot.BuildEmptyTree(macTbl)
    if vlan == "" || macAddress == "" {
        err = fdbMacTableGetAll (inParams)
    } else {
        vlanString := strings.Contains(vlan, "Vlan")
        if vlanString == true {
            vlan = strings.Replace(vlan, "", "Vlan", 0)
        }
        oidToVlan, brPrtOidToIntfOid, fdbMap, err := getASICStateMaps(inParams.dbs[db.AsicDB])
        if err != nil {
            log.Error("getASICStateMaps failed.")
            return err
        }
        oidInfMap,_  := getOidToIntfNameMap(inParams.dbs[db.CountersDB])
        err = fdbMacTableGetEntry(inParams, vlan, macAddress, oidInfMap, oidToVlan, brPrtOidToIntfOid, fdbMap, macTbl)
    }

    return err
}
