////////////////////////////////////////////////////////////////////////////
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
    "github.com/openconfig/ygot/ygot"
    "translib/ocbinds"
)

func init() {
    XlateFuncBind("YangToDb_nat_instance_key_xfmr", YangToDb_nat_instance_key_xfmr)
    XlateFuncBind("DbToYang_nat_instance_key_xfmr", DbToYang_nat_instance_key_xfmr)
    XlateFuncBind("YangToDb_nat_global_key_xfmr", YangToDb_nat_global_key_xfmr)
    XlateFuncBind("DbToYang_nat_global_key_xfmr", DbToYang_nat_global_key_xfmr)
    XlateFuncBind("YangToDb_nat_enable_xfmr", YangToDb_nat_enable_xfmr)
    XlateFuncBind("DbToYang_nat_enable_xfmr", DbToYang_nat_enable_xfmr)
    XlateFuncBind("YangToDb_napt_mapping_subtree_xfmr", YangToDb_napt_mapping_subtree_xfmr)
    XlateFuncBind("DbToYang_napt_mapping_subtree_xfmr", DbToYang_napt_mapping_subtree_xfmr)
    XlateFuncBind("YangToDb_nat_mapping_subtree_xfmr", YangToDb_nat_mapping_subtree_xfmr)
    XlateFuncBind("DbToYang_nat_mapping_subtree_xfmr", DbToYang_nat_mapping_subtree_xfmr)
    XlateFuncBind("YangToDb_nat_pool_key_xfmr", YangToDb_nat_pool_key_xfmr)
    XlateFuncBind("DbToYang_nat_pool_key_xfmr", DbToYang_nat_pool_key_xfmr)
    XlateFuncBind("YangToDb_nat_ip_field_xfmr", YangToDb_nat_ip_field_xfmr)
    XlateFuncBind("DbToYang_nat_ip_field_xfmr", DbToYang_nat_ip_field_xfmr)
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
    XlateFuncBind("nat_post_xfmr", nat_post_xfmr)
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
    NAPT_TABLE       = "NAPT_TABLE"
    STATIC_NAT       = "STATIC_NAT"
    STATIC_NAPT      = "STATIC_NAPT"
    NAT_TYPE         = "nat_type"
    NAT_ENTRY_TYPE   = "entry_type"
    STATIC           = "static"
    DYNAMIC          = "dynamic"
    SNAT             = "snat"
    DNAT             = "dnat"
    NAT_BINDINGS     = "NAT_BINDINGS"
    NAPT_TWICE_TABLE = "NAPT_TWICE_TABLE"
    NAT_TWICE_TABLE  = "NAT_TWICE_TABLE"
)

var nat_post_xfmr PostXfmrFunc = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    if inParams.oper == DELETE {
        if inParams.skipOrdTblChk != nil {
            *inParams.skipOrdTblChk  = true
        }
    }
    log.Infof("nat_post_xfmr returned : %v, skipOrdTblChk: %v", (*inParams.dbDataMap)[db.ConfigDB], *inParams.skipOrdTblChk)
    return (*inParams.dbDataMap)[db.ConfigDB], nil
}

var YangToDb_nat_instance_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var nat_inst_key string
    var err error
    nat_inst_key = "0"
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

func getNatRoot (s *ygot.GoStruct) *ocbinds.OpenconfigNat_Nat {
    deviceObj := (*s).(*ocbinds.Device)
    return deviceObj.Nat
}

func getNatInstance (s *ygot.GoStruct, build bool) *ocbinds.OpenconfigNat_Nat_Instances_Instance {
    deviceObj := (*s).(*ocbinds.Device)
    var natInst *ocbinds.OpenconfigNat_Nat_Instances_Instance
    natObj := deviceObj.Nat

    if natObj != nil {
        if natObj.Instances != nil {
            if natObj.Instances.Instance != nil && len(natObj.Instances.Instance) > 0 {
                if _, ok := natObj.Instances.Instance[0]; ok {
                    natInst = natObj.Instances.Instance[0]
                }
            }
        }
    }

    if natInst == nil && build == true {
        if natObj.Instances == nil {
            ygot.BuildEmptyTree(natObj)
        }
        if natObj.Instances.Instance == nil || len(natObj.Instances.Instance) < 1 {
            ygot.BuildEmptyTree(natObj.Instances)
            natObj.Instances.NewInstance(0)
            ygot.BuildEmptyTree(natObj.Instances.Instance[0])
        }
        if _, ok := natObj.Instances.Instance[0]; !ok {
            natObj.Instances.NewInstance(0)
            ygot.BuildEmptyTree(natObj.Instances.Instance[0])
        }
        natInst = natObj.Instances.Instance[0]
    }
    return natInst
}

func getNaptTblRoot (s *ygot.GoStruct, build bool) *ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable {
    natInst := getNatInstance(s, build)

    var naptTblObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable

    if natInst != nil {
        if natInst.NaptMappingTable == nil && build == true {
            ygot.BuildEmptyTree(natInst)
        }

        naptTblObj = natInst.NaptMappingTable
    }

    return naptTblObj
}

func getNatTblRoot (s *ygot.GoStruct, build bool) *ocbinds.OpenconfigNat_Nat_Instances_Instance_NatMappingTable {
    natInst := getNatInstance(s, build)

    var natTblObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NatMappingTable

    if natInst != nil {
        if natInst.NatMappingTable == nil && build == true {
            ygot.BuildEmptyTree(natInst)
        }

        natTblObj = natInst.NatMappingTable
    }

    return natTblObj
}

func getAllTableKeys (d *db.DB, dbSpec *db.TableSpec) ([]db.Key, error) {
    var keys []db.Key

    tbl, tblErr := d.GetTable(dbSpec)
    if tblErr != nil {
        return keys, tblErr
    }
    keys, tblErr =  tbl.GetKeys()
    return keys, tblErr
}

var YangToDb_nat_mapping_subtree_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    var err error
    natMap := make(map[string]map[string]db.Value)
    tblName := STATIC_NAT

    natTblObj := getNatTblRoot(inParams.ygRoot, false)
    if natTblObj == nil || natTblObj.NatMappingEntry == nil || len(natTblObj.NatMappingEntry) < 1 {
        errStr := "NAT [container/list] not populated."
        log.Info("YangToDb_nat_mapping_subtree_xfmr: " + errStr)
        return natMap, errors.New(errStr)
    }

    if inParams.oper == DELETE {
        if natTblObj.NatMappingEntry == nil || len(natTblObj.NatMappingEntry) < 1 {
            //All entries from db needs to be deleted
            allKeys, keyErr :=  getAllTableKeys(inParams.d, &db.TableSpec{Name:tblName})
            if keyErr != nil {
                log.Info("YangToDb_nat_mapping_subtree_xfmr - GetallKeys failed for table : ", tblName)
                return natMap, errors.New("GetallKeys failed for table " + tblName)
            }
            for _, entKey := range allKeys {
                if len(entKey.Comp) < 1 {
                    continue
                }
                if _, ok := natMap[tblName]; !ok {
                    natMap[tblName] = make (map[string]db.Value)
                }

                entKeyStr := entKey.Comp[0]
                var emtData db.Value
                natMap[tblName][entKeyStr] = emtData
            }
        }
    }


    for key, data := range natTblObj.NatMappingEntry {
        if data.Config == nil && inParams.oper != DELETE {
            errStr := "NAT [Config DATA]  invalid."
            log.Info("YangToDb_nat_mapping_subtree_xfmr : " + errStr)
            return natMap, errors.New(errStr)
        }

        dbkey := key
        if _, ok := natMap[tblName]; !ok {
            natMap[tblName] = make (map[string]db.Value)
        }

        dbData := make(map[string]string)
        entry := db.Value{Field: dbData}
        if data.Config != nil {

            if data.Config.InternalAddress != nil {
                entry.Set("local_ip",  *data.Config.InternalAddress)
            }
            if data.Config.Type != 0 {
                natType :=  findInMap(NAT_TYPE_MAP, strconv.FormatInt(int64(data.Config.Type), 10))
                entry.Set(NAT_TYPE, natType)
            }
            if data.Config.TwiceNatId != nil {
                entry.SetInt("twice_nat_id", int(*data.Config.TwiceNatId))
            }
        }
        log.Info("YangToDb_nat_mapping_subtree_xfmr : dbkey - ", dbkey, " data - ", entry)

        natMap[tblName][dbkey] = entry
    }

    log.Info("YangToDb_nat_mapping_subtree_xfmr : Map -: ", natMap)

    return natMap, err
}

func nat_mapping_Cfg_attr_get (attrUri string, natKey string, natCfgObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NatMappingTable_NatMappingEntry_Config, entry *db.Value) error {

    var err error

    log.Info("nat_mapping_Cfg_attr_get - entry", natKey)
    if natCfgObj == nil || entry == nil {
        errStr := "Invalid params for NAT Config attr get."
        log.Info("nat_mapping_Cfg_attr_get: " + errStr)
        return errors.New(errStr)
    }

    switch (attrUri) {
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/config":
        attrList := []string {"external-address", "internal-address", "twice-nat-id", "type"}
        for _, val := range attrList {
            curAttrUri := attrUri + "/" + val
            nat_mapping_Cfg_attr_get (curAttrUri, natKey, natCfgObj, entry)
        }
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/config/external-address":
        natCfgObj.ExternalAddress = new(string)
        *natCfgObj.ExternalAddress = natKey
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/config/internal-address":
        if entry.Has("local_ip") {
            natCfgObj.InternalAddress = new(string)
            *natCfgObj.InternalAddress = entry.Get("local_ip")
        } else {
            err = errors.New("URI data (local_ip) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/config/twice-nat-id":
        if entry.Has("twice_nat_id") {
            natCfgObj.TwiceNatId = new(uint16)
            valInt, _ := strconv.Atoi(entry.Get("twice_nat_id"))
            *natCfgObj.TwiceNatId = uint16(valInt)
        } else {
            err = errors.New("URI data (twice-nat-id) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/config/type":
        if entry.Has(NAT_TYPE) {
            if entry.Get(NAT_TYPE) == "dnat" {
                natCfgObj.Type = ocbinds.OpenconfigNat_NAT_TYPE_DNAT
            } else  if entry.Get(NAT_TYPE) == "snat" {
                 natCfgObj.Type = ocbinds.OpenconfigNat_NAT_TYPE_SNAT
            } else {
                 err = errors.New("Invalid data in db (nat_type), (" +  entry.Get(NAT_TYPE) + ") " + attrUri)
            }

        } else {
            err = errors.New("URI data (nat_type) not found in db , " + attrUri)
        }
    default:
        errStr := "Invalid Uri " + attrUri
        log.Info("nat_mapping_Cfg_attr_get : " + errStr)
        return errors.New(errStr)
    }
    if err != nil {
        log.Info("nat_mapping_Cfg_attr_get : ", err)
    }
    return nil
}

func nat_mapping_State_attr_get (inParams XfmrParams, attrUri string, natKey string, natStateObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NatMappingTable_NatMappingEntry_State, entry *db.Value) error {
    var err error

    log.Info("nat_mapping_State_attr_get - entry", natKey)
    if natStateObj == nil || entry == nil {
        errStr := "Invalid params for NAT State attr get."
        log.Info("nat_mapping_State_attr_get: " + errStr)
        return errors.New(errStr)
    }

    switch (attrUri) {
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state":
        attrList := []string {"external-address", "translated-ip", "entry-type", "type", "counters"}
        for _, val := range attrList {
            curAttrUri := attrUri + "/" + val
            nat_mapping_State_attr_get(inParams, curAttrUri, natKey, natStateObj, entry)
        }
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state/external-address":
        natStateObj.ExternalAddress = new(string)
        *natStateObj.ExternalAddress = natKey
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state/translated-ip":
        if entry.Has("translated_ip") {
            natStateObj.TranslatedIp = new(string)
            *natStateObj.TranslatedIp = entry.Get("translated_ip")
        } else {
            err = errors.New("URI data (translated_ip) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state/entry-type":
        if entry.Has(NAT_ENTRY_TYPE) {
            if entry.Get(NAT_ENTRY_TYPE) == "static" {
                natStateObj.EntryType = ocbinds.OpenconfigNat_NAT_ENTRY_TYPE_STATIC
            } else if entry.Get(NAT_ENTRY_TYPE) ==  "dynamic" {
                natStateObj.EntryType = ocbinds.OpenconfigNat_NAT_ENTRY_TYPE_DYNAMIC
            } else {
                err = errors.New("Invalid data in db (entry_type), (" +  entry.Get(NAT_ENTRY_TYPE) + ") " + attrUri)
            }
        } else {
            err = errors.New("URI data (entry_type) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state/type":
        if entry.Has(NAT_TYPE) {
            if entry.Get(NAT_TYPE) == "dnat" {
                natStateObj.Type = ocbinds.OpenconfigNat_NAT_TYPE_DNAT
            } else  if entry.Get(NAT_TYPE) == "snat" {
                natStateObj.Type = ocbinds.OpenconfigNat_NAT_TYPE_SNAT
            } else {
                err = errors.New("Invalid data in db (nat_type), (" +  entry.Get(NAT_TYPE) + ") " + attrUri)
            }

        } else {
            err = errors.New("URI data (nat_type) not found in db , " + attrUri)
        }
    default:
        if strings.HasPrefix(attrUri, "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state/counters") {
            dbkey := db.Key{Comp:[]string{natKey}}
            entry, dbErr := inParams.dbs[db.CountersDB].GetEntry(&db.TableSpec{Name:"COUNTERS_NAT"}, dbkey)
            if dbErr != nil {
                log.Info("nat_mapping_State_attr_get Counter DB entry not found ", dbkey)
                return nil
            }
            err = nat_mapping_Counters_attr_get(attrUri, natStateObj.Counters, &entry)

        } else {
            errStr := "Invalid Uri " + attrUri
            log.Info("nat_mapping_State_attr_get : " + errStr)
            return errors.New(errStr)
        }
    }
    if err != nil {
        log.Info("nat_mapping_State_attr_get : ", err)
    }
    return nil
}

func nat_mapping_Counters_attr_get (attrUri string, natCntObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NatMappingTable_NatMappingEntry_State_Counters, entry *db.Value) error {

    var err error

    log.Info("nat_mapping_Counters_attr_get - entry")
    if natCntObj == nil || entry == nil {
        errStr := "Invalid params for NAT counters get."
        log.Info("nat_mapping_Counters_attr_get : " + errStr)
        return errors.New(errStr)
    }

    switch (attrUri) {
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state/counters":
        attrList := []string {"nat-translations-bytes", "nat-translations-pkts"}
        for _, val := range attrList {
            curAttrUri := attrUri + "/" + val
            nat_mapping_Counters_attr_get(curAttrUri, natCntObj, entry)
        }
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state/counters/nat-translations-bytes":
        if entry.Has("NAT_TRANSLATIONS_BYTES") {
            natCntObj.NatTranslationsBytes = new(uint64)
            *natCntObj.NatTranslationsBytes, _ = strconv.ParseUint(entry.Get("NAT_TRANSLATIONS_BYTES"), 10, 64)
        } else {
            err = errors.New("URI data (NAT_TRANSLATIONS_BYTES) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state/counters/nat-translations-pkts":
        if entry.Has("NAT_TRANSLATIONS_PKTS") {
            natCntObj.NatTranslationsPkts = new(uint64)
            *natCntObj.NatTranslationsPkts , _ = strconv.ParseUint(entry.Get("NAT_TRANSLATIONS_PKTS"), 10, 64)
        } else {
            err = errors.New("URI data (NAT_TRANSLATIONS_PKTS) not found in db , " + attrUri)
        }
    default:
        errStr := "Invalid Uri " + attrUri
        log.Info("nat_mapping_Counters_attr_get : " + errStr)
        return errors.New(errStr)
    }
    if err != nil {
        log.Info("nat_mapping_Counters_attr_get : ", err)
    }
    return nil
}

func natMappingTableGetAll(inParams XfmrParams) error {
    var err error
    cfgTbl := STATIC_NAT
    aptTbl := NAT_TABLE
    natTblObj := getNatTblRoot(inParams.ygRoot, true)

    log.Info("natMappingTableGetAll - entry")
    cfgKeys, cfgErr := getAllTableKeys(inParams.dbs[db.ConfigDB], &db.TableSpec{Name:cfgTbl})
    if cfgErr == nil {
        for _, cfgKey := range cfgKeys {
            if len(cfgKey.Comp) < 1 {
                continue
            }
            extAddress := cfgKey.Comp[0]

            var natObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NatMappingTable_NatMappingEntry
            if _, ok := natTblObj.NatMappingEntry[extAddress]; !ok {
                natObj, err = natTblObj.NewNatMappingEntry(extAddress)
                if err != nil {
                    log.Info("natMappingTableGetAll: NewNatMappingEntry failed for - ", extAddress)
                    return err
                }
            } else {
                natObj = natTblObj.NatMappingEntry[extAddress]
            }
            ygot.BuildEmptyTree(natObj)

            entry, dbErr := inParams.dbs[db.ConfigDB].GetEntry(&db.TableSpec{Name:cfgTbl}, cfgKey)
            if dbErr != nil {
                log.Info("natMappingTableGetAll: db.GetEntry entry failed for tbl " + cfgTbl + " dbKey :", cfgKey)
                continue
            }
            targetUriPath := "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/config"
            err = nat_mapping_Cfg_attr_get(targetUriPath, extAddress, natObj.Config, &entry)
        }
    }
    appKeys, stateErr := getAllTableKeys(inParams.dbs[db.ApplDB], &db.TableSpec{Name:aptTbl})
    if stateErr == nil {
        for _, appKey := range appKeys {
            if len(appKey.Comp) < 1 {
                continue
            }
            extAddress := appKey.Comp[0]
            var natObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NatMappingTable_NatMappingEntry
            if _, ok := natTblObj.NatMappingEntry[extAddress]; !ok {
                natObj, err = natTblObj.NewNatMappingEntry(extAddress)
                if err != nil {
                    log.Info("natMappingTableGetAll: NewNatMappingEntry failed for - ", extAddress)
                    return err
                }
            } else {
                natObj = natTblObj.NatMappingEntry[extAddress]
            }
            ygot.BuildEmptyTree(natObj)

            entry, dbErr := inParams.dbs[db.ApplDB].GetEntry(&db.TableSpec{Name:aptTbl}, appKey)
            if dbErr != nil {
                log.Info("natMappingTableGetAll: db.GetEntry entry failed for tbl " + aptTbl + " dbKey :", appKey)
                continue
            }
            targetUriPath := "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state"
            err = nat_mapping_State_attr_get(inParams, targetUriPath, extAddress, natObj.State, &entry)
        }
    }
    return err

}

var DbToYang_nat_mapping_subtree_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) (error) {
    return _DbToYang_nat_mapping_subtree_xfmr(inParams)
}

func _DbToYang_nat_mapping_subtree_xfmr(inParams XfmrParams) (error) {
    var err error
    natTblObj := getNatTblRoot(inParams.ygRoot, true)
    pathInfo := NewPathInfo(inParams.uri)
    extAddress := pathInfo.Var("external-address")
    targetUriPath, err := getYangPathFromUri(inParams.uri)
    log.Info("targetUriPath is ", targetUriPath)
    cfgTbl := STATIC_NAT
    aptTbl := NAT_TABLE


    if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/config") {
        var natObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NatMappingTable_NatMappingEntry
        if _, ok := natTblObj.NatMappingEntry[extAddress]; !ok {
            natObj, err = natTblObj.NewNatMappingEntry(extAddress)
            if err != nil {
                log.Info("DbToYang_nat_mapping_subtree_xfmr : NewNatMappingEntry failed for - ", extAddress)
                return err
            }
        } else {
            natObj = natTblObj.NatMappingEntry[extAddress]
        }
        ygot.BuildEmptyTree(natObj)
        dbKey := db.Key{Comp: []string{extAddress}}
        entry, dbErr := inParams.dbs[db.ConfigDB].GetEntry(&db.TableSpec{Name:cfgTbl}, dbKey)
        if dbErr != nil {
            log.Info("DbToYang_nat_mapping_subtree_xfmr : db.GetEntry entry failed for tbl " + cfgTbl + " dbKey :", dbKey)
            return nil
        }
        return nat_mapping_Cfg_attr_get(targetUriPath, extAddress, natObj.Config, &entry)

    } else if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state") {
        var natObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NatMappingTable_NatMappingEntry
        if _, ok := natTblObj.NatMappingEntry[extAddress]; !ok {
            natObj, err = natTblObj.NewNatMappingEntry(extAddress)
            if err != nil {
                log.Info("DbToYang_nat_mapping_subtree_xfmr : NewNatMappingEntry failed for - ", extAddress)
                return err
            }
        } else {
            natObj = natTblObj.NatMappingEntry[extAddress]
        }
        ygot.BuildEmptyTree(natObj)
        dbKey := db.Key{Comp: []string{extAddress}}
        entry, dbErr := inParams.dbs[db.ApplDB].GetEntry(&db.TableSpec{Name:aptTbl}, dbKey)
        if dbErr != nil {
            log.Info("DbToYang_nat_mapping_subtree_xfmr : db.GetEntry entry failed for tbl " + aptTbl + " dbKey :", dbKey)
            return nil
        }
        return nat_mapping_State_attr_get(inParams, targetUriPath, extAddress, natObj.State, &entry)

    } else if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry") {
        if extAddress == "" {
            err = natMappingTableGetAll(inParams)
        } else {
            curParams := inParams
            curParams.uri = inParams.uri + "/" + "config"
            err = _DbToYang_nat_mapping_subtree_xfmr(curParams)
            curParams.uri = inParams.uri + "/" + "state"
            err = _DbToYang_nat_mapping_subtree_xfmr(curParams)
        }
    } else {
        err = natMappingTableGetAll(inParams)
    }

    return err
}


var YangToDb_napt_mapping_subtree_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    var err error
    naptMap := make(map[string]map[string]db.Value)
    tblName := STATIC_NAPT

    naptTblObj := getNaptTblRoot(inParams.ygRoot, false)
    if naptTblObj == nil || naptTblObj.NaptMappingEntry == nil || len(naptTblObj.NaptMappingEntry) < 1 {
        errStr := "NAPT [container/list] not populated."
        log.Info("YangToDb_napt_mapping_subtree_xfmr: " + errStr)
        return naptMap, errors.New(errStr)
    }

    if inParams.oper == DELETE {
        if naptTblObj.NaptMappingEntry == nil || len(naptTblObj.NaptMappingEntry) < 1 {
            //All entries from db needs to be deleted
            allKeys, keyErr :=  getAllTableKeys(inParams.d, &db.TableSpec{Name:tblName})
            if keyErr != nil {
                log.Info("YangToDb_napt_mapping_subtree_xfmr - GetallKeys failed for table : ", tblName)
                return naptMap, errors.New("GetallKeys failed for table " + tblName)
            }
            for _, entKey := range allKeys {
                if len(entKey.Comp) < 3 {
                    continue
                }
                if _, ok := naptMap[tblName]; !ok {
                    naptMap[tblName] = make (map[string]db.Value)
                }

                entKeyStr := strings.Join(entKey.Comp, "|")
                var emtData db.Value
                naptMap[tblName][entKeyStr] = emtData
            }
        }
    }


    for key, data := range naptTblObj.NaptMappingEntry {
        if data.Config == nil && inParams.oper != DELETE {
            errStr := "NAPT [Config DATA]  invalid."
            log.Info("YangToDb_napt_mapping_subtree_xfmr : " + errStr)
            return naptMap, errors.New(errStr)
        }
        if _, ok := protocol_map[key.Protocol]; !ok {
            log.Info("YangToDb_napt_mapping_subtree_xfmr : Invalid protocol key for NAPT config entry.")
            errStr := "NAPT [keys] not valid."
            return naptMap, errors.New(errStr)
        }

        dbkey := key.ExternalAddress + "|" + protocol_map[key.Protocol] + "|" + strconv.FormatInt(int64(key.ExternalPort), 10)
        if _, ok := naptMap[tblName]; !ok {
            naptMap[tblName] = make (map[string]db.Value)
        }

        dbData := make(map[string]string)
        entry := db.Value{Field: dbData}

        if data.Config != nil {
            if data.Config.InternalAddress != nil {
                entry.Set("local_ip",  *data.Config.InternalAddress)
            }
            if data.Config.InternalPort != nil {
                entry.SetInt("local_port", int(*data.Config.InternalPort))
            }
            if data.Config.Type != 0 {
                natType :=  findInMap(NAT_TYPE_MAP, strconv.FormatInt(int64(data.Config.Type), 10))
                entry.Set(NAT_TYPE, natType)
            }
            if data.Config.TwiceNatId != nil {
                entry.SetInt("twice_nat_id", int(*data.Config.TwiceNatId))
            }
        }
        log.Info("YangToDb_napt_mapping_subtree_xfmr : dbkey - ", dbkey, " data - ", entry)

        naptMap[tblName][dbkey] = entry
    }

    log.Info("YangToDb_napt_mapping_subtree_xfmr : Map -: ", naptMap)

    return naptMap, err
}

func napt_mapping_Cfg_attr_get (attrUri string, naptKey ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry_Key, naptCfgObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry_Config, entry *db.Value) error {

    var err error

    log.Info("napt_mapping_Cfg_attr_get - entry", naptKey)
    if naptCfgObj == nil || entry == nil {
        errStr := "Invalid params for NAPT Config attr get."
        log.Info("napt_mapping_Cfg_attr_get: " + errStr, " ", naptCfgObj, " ", naptKey)
        return errors.New(errStr)
    }

    switch (attrUri) {
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config":
        attrList := []string {"external-address", "external-port", "protocol", "internal-address", "internal-port", "twice-nat-id", "type"}
        for _, val := range attrList {
            curAttrUri := attrUri + "/" + val
            napt_mapping_Cfg_attr_get (curAttrUri, naptKey, naptCfgObj, entry)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config/external-address":
        naptCfgObj.ExternalAddress = new(string)
        *naptCfgObj.ExternalAddress = naptKey.ExternalAddress
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config/external-port":
        naptCfgObj.ExternalPort = new(uint16)
        *naptCfgObj.ExternalPort = naptKey.ExternalPort
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config/protocol":
        naptCfgObj.Protocol = new(uint8)
        *naptCfgObj.Protocol =  naptKey.Protocol
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config/internal-address":
        if entry.Has("local_ip") {
            naptCfgObj.InternalAddress = new(string)
            *naptCfgObj.InternalAddress = entry.Get("local_ip")
        } else {
            err = errors.New("URI data (local_ip) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config/internal-port":
        if entry.Has("local_port") {
            naptCfgObj.InternalPort = new(uint16)
            valInt, _ := strconv.Atoi(entry.Get("local_port"))
            *naptCfgObj.InternalPort = uint16(valInt)
        } else {
            err = errors.New("URI data (local_port) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config/twice-nat-id":
        if entry.Has("twice_nat_id") {
            naptCfgObj.TwiceNatId = new(uint16)
            valInt, _ := strconv.Atoi(entry.Get("twice_nat_id"))
            *naptCfgObj.TwiceNatId = uint16(valInt)
        } else {
            err = errors.New("URI data (twice-nat-id) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config/type":
        if entry.Has(NAT_TYPE) {
            if entry.Get(NAT_TYPE) == "dnat" {
                naptCfgObj.Type = ocbinds.OpenconfigNat_NAT_TYPE_DNAT
            } else  if entry.Get(NAT_TYPE) == "snat" {
                 naptCfgObj.Type = ocbinds.OpenconfigNat_NAT_TYPE_SNAT
            } else {
                 err = errors.New("Invalid data in db (nat_type), (" +  entry.Get(NAT_TYPE) + ") " + attrUri)
            }

        } else {
            err = errors.New("URI data (nat_type) not found in db , " + attrUri)
        }
    default:
        errStr := "Invalid Uri " + attrUri
        log.Info("napt_mapping_Cfg_attr_get : " + errStr)
        return errors.New(errStr)
    }
    log.Info("napt_mapping_Cfg_attr_get : ", err)
    return nil
}

func napt_mapping_State_attr_get (inParams XfmrParams, attrUri string, naptKey ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry_Key, naptStateObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry_State, entry *db.Value) error {
    var err error

    log.Info("napt_mapping_State_attr_get - entry", naptKey)
    if naptStateObj == nil || entry == nil {
        errStr := "Invalid params for NAPT State attr get."
        log.Info("napt_mapping_State_attr_get: " + errStr)
        return errors.New(errStr)
    }

    switch (attrUri) {
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state":
        attrList := []string {"external-address", "external-port", "protocol", "translated-ip", "translated-port", "entry-type", "type", "counters"}
        for _, val := range attrList {
            curAttrUri := attrUri + "/" + val
            napt_mapping_State_attr_get(inParams, curAttrUri, naptKey, naptStateObj, entry)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/external-address":
        naptStateObj.ExternalAddress = new(string)
        *naptStateObj.ExternalAddress = naptKey.ExternalAddress
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/external-port":
        naptStateObj.ExternalPort = new(uint16)
        *naptStateObj.ExternalPort = naptKey.ExternalPort
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/protocol":
        naptStateObj.Protocol = new(uint8)
        *naptStateObj.Protocol =  naptKey.Protocol
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/translated-ip":
        if entry.Has("translated_ip") {
            naptStateObj.TranslatedIp = new(string)
            *naptStateObj.TranslatedIp = entry.Get("translated_ip")
        } else {
            err = errors.New("URI data (translated_ip) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/translated-port":
        if entry.Has("translated_l4_port") {
            naptStateObj.TranslatedPort = new(uint16)
            valInt, _ := strconv.Atoi(entry.Get("translated_l4_port"))
            *naptStateObj.TranslatedPort = uint16(valInt)
        } else {
            err = errors.New("URI data (translated_l4_port) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/entry-type":
        if entry.Has(NAT_ENTRY_TYPE) {
            if entry.Get(NAT_ENTRY_TYPE) == "static" {
                naptStateObj.EntryType = ocbinds.OpenconfigNat_NAT_ENTRY_TYPE_STATIC
            } else if entry.Get(NAT_ENTRY_TYPE) ==  "dynamic" {
                naptStateObj.EntryType = ocbinds.OpenconfigNat_NAT_ENTRY_TYPE_DYNAMIC
            } else {
                err = errors.New("Invalid data in db (entry_type), (" +  entry.Get(NAT_ENTRY_TYPE) + ") " + attrUri)
            }
        } else {
            err = errors.New("URI data (entry_type) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/type":
        if entry.Has(NAT_TYPE) {
            if entry.Get(NAT_TYPE) == "dnat" {
                naptStateObj.Type = ocbinds.OpenconfigNat_NAT_TYPE_DNAT
            } else  if entry.Get(NAT_TYPE) == "snat" {
                naptStateObj.Type = ocbinds.OpenconfigNat_NAT_TYPE_SNAT
            } else {
                err = errors.New("Invalid data in db (nat_type), (" +  entry.Get(NAT_TYPE) + ") " + attrUri)
            }

        } else {
            err = errors.New("URI data (nat_type) not found in db , " + attrUri)
        }
    default:
        if strings.HasPrefix(attrUri, "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/counters") {
            dbKey, _ := YangToDb_napt_mapping_key("counters", naptKey.ExternalAddress, strconv.Itoa(int(naptKey.ExternalPort)), strconv.Itoa(int(naptKey.Protocol)))
            entry, dbErr := inParams.dbs[db.CountersDB].GetEntry(&db.TableSpec{Name:"COUNTERS_NAPT"}, dbKey)
            if dbErr != nil {
                log.Info("napt_mapping_State_attr_get Counter DB entry not found ", dbKey)
                return nil
            }
            err = napt_mapping_Counters_attr_get(attrUri, naptStateObj.Counters, &entry)

        } else {
            errStr := "Invalid Uri " + attrUri
            log.Info("napt_mapping_State_attr_get : " + errStr)
            return errors.New(errStr)
        }
    }
    log.Info("napt_mapping_State_attr_get : ", err);

    return nil
}

func napt_mapping_Counters_attr_get (attrUri string, naptCntObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry_State_Counters, entry *db.Value) error {

    var err error

    log.Info("napt_mapping_Counters_attr_get - entry")
    if naptCntObj == nil || entry == nil {
        errStr := "Invalid params for NAPT counters get."
        log.Info("napt_mapping_Counters_attr_get : " + errStr)
        return errors.New(errStr)
    }

    switch (attrUri) {
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/counters":
        attrList := []string {"nat-translations-bytes", "nat-translations-pkts"}
        for _, val := range attrList {
            curAttrUri := attrUri + "/" + val
            napt_mapping_Counters_attr_get(curAttrUri, naptCntObj, entry)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/counters/nat-translations-bytes":
        if entry.Has("NAT_TRANSLATIONS_BYTES") {
            naptCntObj.NatTranslationsBytes = new(uint64)
            *naptCntObj.NatTranslationsBytes, _ = strconv.ParseUint(entry.Get("NAT_TRANSLATIONS_BYTES"), 10, 64)
        } else {
            err = errors.New("URI data (NAT_TRANSLATIONS_BYTES) not found in db , " + attrUri)
        }
    case "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state/counters/nat-translations-pkts":
        if entry.Has("NAT_TRANSLATIONS_PKTS") {
            naptCntObj.NatTranslationsPkts = new(uint64)
            *naptCntObj.NatTranslationsPkts , _ = strconv.ParseUint(entry.Get("NAT_TRANSLATIONS_PKTS"), 10, 64)
        } else {
            err = errors.New("URI data (NAT_TRANSLATIONS_PKTS) not found in db , " + attrUri)
        }
    default:
        errStr := "Invalid Uri " + attrUri
        log.Info("napt_mapping_Counters_attr_get : " + errStr)
        return errors.New(errStr)
    }
    log.Info("napt_mapping_Counters_attr_get - ", err)
    return nil
}

func createOCNaptKey (extAddr string, extPort string, protocol string) (ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry_Key, error) {
    var naptKey ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry_Key
    if extAddr == "" || extPort == "" || protocol == "" {
        errStr := "invali params " + extAddr + " " + extPort + " " + protocol
        log.Info("createOCNaptKey : " + errStr)
        return naptKey, errors.New(errStr)
    }

    naptKey.ExternalAddress = extAddr
    valInt, _ := strconv.Atoi(extPort)
    naptKey.ExternalPort =  uint16(valInt)
    valInt, _ = strconv.Atoi(protocol)
    naptKey.Protocol = uint8(valInt)

    return naptKey, nil
}

func naptMappingTableGetAll(inParams XfmrParams) error {
    var err error
    cfgTbl := STATIC_NAPT
    aptTbl := NAPT_TABLE
    naptTblObj := getNaptTblRoot(inParams.ygRoot, true)

    log.Info("naptMappingTableGetAll - entry")
    cfgKeys, cfgErr := getAllTableKeys(inParams.dbs[db.ConfigDB], &db.TableSpec{Name:cfgTbl})
    if cfgErr == nil {
        for _, cfgKey := range cfgKeys {
            if len(cfgKey.Comp) < 3 {
                continue
            }
            extAddress, extPort, protocol, _:= DbToYang_napt_mapping_key("config", cfgKey)
            naptKey, keyErr := createOCNaptKey(extAddress, extPort, protocol)
            if keyErr != nil {
                log.Info("naptMappingTableGetAll: Invalid key attributes for Config obj")
                return keyErr
            }

            var naptObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry
            if _, ok := naptTblObj.NaptMappingEntry[naptKey]; !ok {
                naptObj, err = naptTblObj.NewNaptMappingEntry(naptKey.ExternalAddress, naptKey.Protocol, naptKey.ExternalPort)
                if err != nil {
                    log.Info("naptMappingTableGetAll: NewNaptMappingEntry failed for - ", naptKey)
                    return err
                }
            } else {
                naptObj = naptTblObj.NaptMappingEntry[naptKey]
            }
            ygot.BuildEmptyTree(naptObj)

            entry, dbErr := inParams.dbs[db.ConfigDB].GetEntry(&db.TableSpec{Name:cfgTbl}, cfgKey)
            if dbErr != nil {
                log.Info("naptMappingTableGetAll: db.GetEntry entry failed for tbl " + cfgTbl + " dbKey :", cfgKey)
                continue
            }
            targetUriPath := "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config"
            err = napt_mapping_Cfg_attr_get(targetUriPath, naptKey, naptObj.Config, &entry)
        }
    }
    appKeys, stateErr := getAllTableKeys(inParams.dbs[db.ApplDB], &db.TableSpec{Name:aptTbl})
    if stateErr == nil {
        for _, appKey := range appKeys {
            if len(appKey.Comp) < 3 {
                continue
            }
            extAddress, extPort, protocol,_ := DbToYang_napt_mapping_key("state", appKey)
            naptKey, keyErr := createOCNaptKey(extAddress, extPort, protocol)
            if keyErr != nil {
                log.Info("naptMappingTableGetAll: Invalid key attributes for Config obj")
                return keyErr
            }

            var naptObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry
            if _, ok := naptTblObj.NaptMappingEntry[naptKey]; !ok {
                naptObj, err = naptTblObj.NewNaptMappingEntry(naptKey.ExternalAddress, naptKey.Protocol, naptKey.ExternalPort)
                if err != nil {
                    log.Info("naptMappingTableGetAll: NewNaptMappingEntry failed for - ", naptKey)
                    return err
                }
            } else {
                naptObj = naptTblObj.NaptMappingEntry[naptKey]
            }
            ygot.BuildEmptyTree(naptObj)

            entry, dbErr := inParams.dbs[db.ApplDB].GetEntry(&db.TableSpec{Name:aptTbl}, appKey)
            if dbErr != nil {
                log.Info("naptMappingTableGetAll: db.GetEntry entry failed for tbl " + aptTbl + " dbKey :", appKey)
                continue
            }
            targetUriPath := "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state"
            err = napt_mapping_State_attr_get(inParams, targetUriPath, naptKey, naptObj.State, &entry)
        }
    }
    return err

}

var DbToYang_napt_mapping_subtree_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) (error) {
    return _DbToYang_napt_mapping_subtree_xfmr(inParams)
}

func _DbToYang_napt_mapping_subtree_xfmr(inParams XfmrParams) (error) {
    var err error
    naptTblObj := getNaptTblRoot(inParams.ygRoot, true)
    pathInfo := NewPathInfo(inParams.uri)
    extAddress := pathInfo.Var("external-address")
    extPort := pathInfo.Var("external-port")
    protocol := pathInfo.Var("protocol")
    targetUriPath, err := getYangPathFromUri(inParams.uri)
    log.Info("targetUriPath is ", targetUriPath)
    cfgTbl := STATIC_NAPT
    aptTbl := NAPT_TABLE


    if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config") {
        naptKey, keyErr := createOCNaptKey(extAddress, extPort, protocol)
        if keyErr != nil {
            log.Info("DbToYang_napt_mapping_subtree_xfmr : Invalid key attributes for Config obj")
            return keyErr
        }
        var naptObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry
        if _, ok := naptTblObj.NaptMappingEntry[naptKey]; !ok {
            naptObj, err = naptTblObj.NewNaptMappingEntry(naptKey.ExternalAddress, naptKey.Protocol, naptKey.ExternalPort)
            if err != nil {
                log.Info("DbToYang_napt_mapping_subtree_xfmr : NewNaptMappingEntry failed for - ", naptKey)
                return err
            }
        } else {
            naptObj = naptTblObj.NaptMappingEntry[naptKey]
        }
        ygot.BuildEmptyTree(naptObj)
        dbKey, ykeyErr := YangToDb_napt_mapping_key("config", extAddress, extPort, protocol)
        if ykeyErr != nil {
            log.Info("DbToYang_napt_mapping_subtree_xfmr : YangToDb_napt_mapping_key failed for - ", naptKey)
            return ykeyErr
        }
        entry, dbErr := inParams.dbs[db.ConfigDB].GetEntry(&db.TableSpec{Name:cfgTbl}, dbKey)
        if dbErr != nil {
            log.Info("DbToYang_napt_mapping_subtree_xfmr : db.GetEntry entry failed for tbl " + cfgTbl + " dbKey :", dbKey)
            return nil
        }
        return napt_mapping_Cfg_attr_get(targetUriPath, naptKey, naptObj.Config, &entry)

    } else if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state") {
        naptKey, keyErr := createOCNaptKey(extAddress, extPort, protocol)
        if keyErr != nil {
            log.Info("DbToYang_napt_mapping_subtree_xfmr : Invalid key attributes for Config obj")
            return keyErr
        }
        var naptObj *ocbinds.OpenconfigNat_Nat_Instances_Instance_NaptMappingTable_NaptMappingEntry
        if _, ok := naptTblObj.NaptMappingEntry[naptKey]; !ok {
            naptObj, err = naptTblObj.NewNaptMappingEntry(naptKey.ExternalAddress, naptKey.Protocol, naptKey.ExternalPort)
            if err != nil {
                log.Info("DbToYang_napt_mapping_subtree_xfmr : NewNaptMappingEntry failed for - ", naptKey)
                return err
            }
        } else {
            naptObj = naptTblObj.NaptMappingEntry[naptKey]
        }
        ygot.BuildEmptyTree(naptObj)
        dbKey, ykeyErr := YangToDb_napt_mapping_key("state", extAddress, extPort, protocol)
        if ykeyErr != nil {
            log.Info("DbToYang_napt_mapping_subtree_xfmr : YangToDb_napt_mapping_key failed for - ", naptKey)
            return ykeyErr
        }
        entry, dbErr := inParams.dbs[db.ApplDB].GetEntry(&db.TableSpec{Name:aptTbl}, dbKey)
        if dbErr != nil {
            log.Info("DbToYang_napt_mapping_subtree_xfmr : db.GetEntry entry failed for tbl " + aptTbl + " dbKey :", dbKey)
            return nil
        }
        return napt_mapping_State_attr_get(inParams, targetUriPath, naptKey, naptObj.State, &entry)

    } else if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry") {
        if extAddress == "" || extPort == "" || protocol == "" {
            err = naptMappingTableGetAll(inParams)
        } else {
            curParams := inParams
            curParams.uri = inParams.uri + "/" + "config"
            err = _DbToYang_napt_mapping_subtree_xfmr(curParams)
            curParams.uri = inParams.uri + "/" + "state"
            err = _DbToYang_napt_mapping_subtree_xfmr(curParams)
        }
    } else {
        err = naptMappingTableGetAll(inParams)
    }

    return err
}


func YangToDb_napt_mapping_key (objType string, extAddress string, extPort string, proto string) (db.Key, error) {
    var err error
    var dbkey db.Key

    log.Info("YangToDb_napt_mapping_key ", extAddress, " ", extPort, " ", proto)
    protocol,_ := strconv.Atoi(proto)
    if _, ok := protocol_map[uint8(protocol)]; !ok {
        log.Info("YangToDb_napt_mapping_key_xfmr - Invalid protocol : ", protocol);
        return dbkey, errors.New("Invalid protocol :" + proto)
    }
    switch (objType) {
    case "config":
        dbkey = db.Key{Comp: []string{extAddress,  protocol_map[uint8(protocol)], extPort}}
    case "state", "counters":
        dbkey = db.Key{Comp: []string{ protocol_map[uint8(protocol)], extAddress, extPort}}
    default:
        log.Info("YangToDb_napt_mapping_key : Invalid objType " + objType)
        err = errors.New("Invalid objType " + objType)
    }

    log.Info("YangToDb_napt_mapping_key : Key : ", dbkey)
    return dbkey, err
}

func DbToYang_napt_mapping_key (objType string, key db.Key) (string, string, string, error) {
    var err error
    var extAddr, extPort, protocol string

    if len(key.Comp) < 3 {
        err = errors.New("Invalid key for NAPT ampping entry.")
        log.Info("Invalid Keys, NAPT Mapping entry", key)
        return extAddr, extPort, protocol, err
    }

    switch (objType) {
    case "config":
        oc_protocol := findProtocolByValue(protocol_map, key.Comp[1])
        protocol = strconv.Itoa(int(oc_protocol))
        extAddr = key.Comp[0]
        extPort = key.Comp[2]
    case "state", "counters":
        oc_protocol := findProtocolByValue(protocol_map, key.Comp[0])
        protocol = strconv.Itoa(int(oc_protocol))
        extAddr = key.Comp[1]
        extPort = key.Comp[2]
    default:
        log.Info("DbToYang_napt_mapping_key : Invalid objType " + objType)
        err = errors.New("Invalid objType " + objType)
    }

    log.Info(" DbToYang_napt_mapping_key : - ", extAddr, " ", extPort, " ", protocol)
    return extAddr, extPort, protocol, err
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

var YangToDb_nat_ip_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    ipPtr, _ := inParams.param.(*string)
    res_map["nat_ip"] = *ipPtr;
    return res_map, nil
}

var DbToYang_nat_ip_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    tblName := "NAT_POOL"
    if _, ok := data[tblName]; ok {
        if _, entOk := data[tblName][inParams.key]; entOk {
            entry := data[tblName][inParams.key]
            fldOk := entry.Has("nat_ip")
            if fldOk == true {
                ipStr := entry.Get("nat_ip")
                ipRange := strings.Contains(ipStr, "-")
                if ipRange == true {
                    result["IP-ADDRESS-RANGE"] = ipStr
                } else {
                    result["IP-ADDRESS"] = ipStr
                }
            }
        }
    }
    return result, err
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
    rmap["zone-id"],_ = strconv.Atoi(key)
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

    if srcIp == "" || dstIp == "" {
        log.Info("YangToDb_nat_twice_mapping_key_xfmr : Invalid key params.")
        return nat_key, err
    }
    key_sep = ":"

    nat_key = srcIp + key_sep + dstIp
    log.Info("YangToDb_nat_twice_mapping_key_xfmr : Key : ", nat_key)
    return nat_key, err
}

var DbToYang_nat_twice_mapping_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    nat_key := inParams.key
    var key_sep string
    key_sep = ":"

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
    proto    := pathInfo.Var("protocol")
    srcIp    := pathInfo.Var("src-ip")
    srcPort  := pathInfo.Var("src-port")
    dstIp    := pathInfo.Var("dst-ip")
    dstPort  := pathInfo.Var("dst-port")

    if proto == "" || srcIp == "" || srcPort == "" || dstIp == "" || dstPort == "" {
        log.Info("YangToDb_napt_twice_mapping_key_xfmr : Invalid key params.")
        return napt_key, nil
    }

    protocol, _ := strconv.Atoi(proto)
    if _, ok := protocol_map[uint8(protocol)]; !ok {
        log.Info("YangToDb_napt_twice_mapping_key_xfmr - Invalid protocol : ", protocol);
        return napt_key, nil
    }

    key_sep = ":"

    napt_key = protocol_map[uint8(protocol)] + key_sep + srcIp + key_sep + srcPort + key_sep + dstIp + key_sep + dstPort
    log.Info("YangToDb_napt_twice_mapping_key_xfmr : Key : ", napt_key)
    return napt_key, err
}

var DbToYang_napt_twice_mapping_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    var key_sep string
    nat_key := inParams.key
    key_sep = ":"

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

    targetUriPath, err := getYangPathFromUri(inParams.uri)
    var tblName string

    if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/config") {
        tblName = STATIC_NAPT
    } else if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry/state") {
        tblName = NAPT_TABLE
    } else if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/config") {
        tblName = STATIC_NAT
    } else if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/nat-mapping-table/nat-mapping-entry/state") {
        tblName = NAT_TABLE
    } else if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/nat-acl-pool-binding/nat-acl-pool-binding-entry") {
        tblName = NAT_BINDINGS
    }else {
        log.Info("DbToYang_nat_type_field_xfmr: Invalid URI: %s\n", targetUriPath)
        return result, errors.New("Invalid URI " + targetUriPath)
    }

    if _, ok := data[tblName]; ok {
        if _, entOk := data[tblName][inParams.key]; entOk {
            entry := data[tblName][inParams.key]
            fldOk := entry.Has(NAT_TYPE)
            if fldOk == true {
                t := findInMap(NAT_TYPE_MAP, data[tblName][inParams.key].Field[NAT_TYPE])
                var n int64
                n, err = strconv.ParseInt(t, 10, 64)
                if err == nil {
                    result["type"] = ocbinds.E_OpenconfigNat_NAT_TYPE(n).ΛMap()["E_OpenconfigNat_NAT_TYPE"][n].Name
                }
            }
        }
    }

    return result, err
}

var YangToDb_nat_entry_type_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    result := make(map[string]string)
    var err error

    if inParams.param == nil {
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
    targetUriPath, err := getYangPathFromUri(inParams.uri)
    var tblName string

    if strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/napt-mapping-table/napt-mapping-entry") {
        tblName = NAPT_TABLE
    } else if  strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/napt-twice-mapping-table/napt-twice-entry") {
        tblName = NAPT_TWICE_TABLE
    } else if  strings.HasPrefix(targetUriPath, "/openconfig-nat:nat/instances/instance/nat-twice-mapping-table/nat-twice-entry") {
        tblName = NAT_TWICE_TABLE
    } else {
        tblName = NAT_TABLE
    }
    if _, ok := data[tblName]; ok {
        if _, entOk := data[tblName][inParams.key]; entOk {
            entry := data[tblName][inParams.key]
            fldOk := entry.Has(NAT_ENTRY_TYPE)
            if fldOk == true {
                t := findInMap(NAT_ENTRY_TYPE_MAP, data[tblName][inParams.key].Field[NAT_ENTRY_TYPE])
                var n int64
                n, err = strconv.ParseInt(t, 10, 64)
                if err == nil {
                    result["entry-type"] = ocbinds.E_OpenconfigNat_NAT_ENTRY_TYPE(n).ΛMap()["E_OpenconfigNat_NAT_ENTRY_TYPE"][n].Name
                }
            }
        }
    }

    return result, err
}

