//////////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Dell, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//////////////////////////////////////////////////////////////////////////

package transformer

import (
    "errors"
    "strings"
    "strconv"
    "regexp"
    "net"
    "github.com/openconfig/ygot/ygot"
    "translib/db"
    log "github.com/golang/glog"
    "translib/ocbinds"
    "translib/tlerr"
    "bufio"
    "os"
    "fmt"
    "encoding/json"
    "time"
)

func init () {
    XlateFuncBind("intf_table_xfmr", intf_table_xfmr)
    XlateFuncBind("YangToDb_intf_name_xfmr", YangToDb_intf_name_xfmr)
    XlateFuncBind("DbToYang_intf_name_xfmr", DbToYang_intf_name_xfmr)
    XlateFuncBind("YangToDb_intf_enabled_xfmr", YangToDb_intf_enabled_xfmr)
    XlateFuncBind("DbToYang_intf_enabled_xfmr", DbToYang_intf_enabled_xfmr)
    XlateFuncBind("DbToYang_intf_admin_status_xfmr", DbToYang_intf_admin_status_xfmr)
    XlateFuncBind("DbToYang_intf_oper_status_xfmr", DbToYang_intf_oper_status_xfmr)
    XlateFuncBind("DbToYang_intf_eth_auto_neg_xfmr", DbToYang_intf_eth_auto_neg_xfmr)
    XlateFuncBind("DbToYang_intf_eth_port_speed_xfmr", DbToYang_intf_eth_port_speed_xfmr)
    XlateFuncBind("YangToDb_intf_eth_port_config_xfmr", YangToDb_intf_eth_port_config_xfmr)
    XlateFuncBind("YangToDb_intf_ip_addr_xfmr", YangToDb_intf_ip_addr_xfmr)
    XlateFuncBind("DbToYang_intf_ip_addr_xfmr", DbToYang_intf_ip_addr_xfmr)
    XlateFuncBind("YangToDb_intf_subintfs_xfmr", YangToDb_intf_subintfs_xfmr)
    XlateFuncBind("DbToYang_intf_subintfs_xfmr", DbToYang_intf_subintfs_xfmr)
    XlateFuncBind("DbToYang_intf_get_counters_xfmr", DbToYang_intf_get_counters_xfmr)
    XlateFuncBind("YangToDb_intf_tbl_key_xfmr", YangToDb_intf_tbl_key_xfmr)
    XlateFuncBind("DbToYang_intf_tbl_key_xfmr", DbToYang_intf_tbl_key_xfmr)
    XlateFuncBind("YangToDb_intf_name_empty_xfmr", YangToDb_intf_name_empty_xfmr)
    /*--show ip ARP/neighbors changes start--*/
    XlateFuncBind("DbToYang_neigh_tbl_get_all_ipv4_xfmr", DbToYang_neigh_tbl_get_all_ipv4_xfmr)
    XlateFuncBind("DbToYang_neigh_tbl_get_all_ipv6_xfmr", DbToYang_neigh_tbl_get_all_ipv6_xfmr)
    XlateFuncBind("DbToYang_neigh_tbl_key_xfmr", DbToYang_neigh_tbl_key_xfmr)
    XlateFuncBind("YangToDb_neigh_tbl_key_xfmr", YangToDb_neigh_tbl_key_xfmr)
    /*--show ip ARP/neighbors changes end--*/
    XlateFuncBind("rpc_clear_counters", rpc_clear_counters)
}

/*--show ip ARP/neighbors changes start--*/
const (
    NEIGH_IPv4_PREFIX = "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv4/neighbors"
    NEIGH_IPv4_PREFIX_IP = NEIGH_IPv4_PREFIX+"/neighbor"
    NEIGH_IPv4_PREFIX_STATE_IP = NEIGH_IPv4_PREFIX_IP+"/state/ip"
    NEIGH_IPv4_PREFIX_STATE_LL = NEIGH_IPv4_PREFIX_IP+"/state/link-layer-address"
    NEIGH_IPv6_PREFIX = "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/neighbors"
    NEIGH_IPv6_PREFIX_IP = NEIGH_IPv6_PREFIX+"/neighbor"
    NEIGH_IPv6_PREFIX_STATE_IP = NEIGH_IPv6_PREFIX_IP+"/state/ip"
    NEIGH_IPv6_PREFIX_STATE_LL = NEIGH_IPv6_PREFIX_IP+"/state/link-layer-address"
)
/*--show ip ARP/neighbors changes end--*/

const (
    PORT_INDEX         = "index"
    PORT_MTU           = "mtu"
    PORT_ADMIN_STATUS  = "admin_status"
    PORT_SPEED         = "speed"
    PORT_DESC          = "description"
    PORT_OPER_STATUS   = "oper_status"
    PORT_AUTONEG       = "autoneg"
    VLAN_TN            = "VLAN"
    VLAN_MEMBER_TN     = "VLAN_MEMBER"
    VLAN_INTERFACE_TN  = "VLAN_INTERFACE"
)

const (
    PIPE                     =  "|"
    COLON                    =  ":"

    ETHERNET                 = "Ethernet"
    MGMT                     = "eth"
    VLAN                     = "Vlan"
    PORTCHANNEL              = "PortChannel"
    LOOPBACK                 = "Loopback"
)

type TblData  struct  {
    portTN           string
    memberTN         string
    intfTN           string
    keySep           string
}

type PopulateIntfCounters func (inParams XfmrParams, counters *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters) (error)
type CounterData struct {
    OIDTN             string
    CountersTN        string
    PopulateCounters  PopulateIntfCounters
}

type IntfTblData struct {
    intfPrefix          string
    cfgDb               TblData
    appDb               TblData
    stateDb             TblData
    CountersHdl         CounterData
}

var IntfTypeTblMap = map[E_InterfaceType]IntfTblData {
    IntfTypeEthernet: IntfTblData{
        cfgDb:TblData{portTN:"PORT", intfTN: "INTERFACE", keySep:PIPE},
        appDb:TblData{portTN:"PORT_TABLE", intfTN: "INTF_TABLE", keySep: COLON},
        stateDb:TblData{portTN: "PORT_TABLE", intfTN: "INTERFACE_TABLE", keySep: PIPE},
        CountersHdl:CounterData{OIDTN: "COUNTERS_PORT_NAME_MAP", CountersTN: "COUNTERS", PopulateCounters: populatePortCounters},
    },
    IntfTypeMgmt : IntfTblData{
        cfgDb:TblData{portTN:"MGMT_PORT", intfTN:"MGMT_INTERFACE", keySep: PIPE},
        appDb:TblData{portTN:"MGMT_PORT_TABLE", intfTN:"MGMT_INTF_TABLE", keySep: COLON},
        stateDb:TblData{portTN:"MGMT_PORT_TABLE", intfTN:"MGMT_INTERFACE_TABLE", keySep: PIPE},
        CountersHdl:CounterData{OIDTN: "", CountersTN:"", PopulateCounters: populateMGMTPortCounters},
    },
    IntfTypePortChannel : IntfTblData{
        cfgDb:TblData{portTN:"PORTCHANNEL", intfTN:"PORTCHANNEL_INTERFACE", memberTN:"PORTCHANNEL_MEMBER", keySep: PIPE},
        appDb:TblData{portTN:"LAG_TABLE", intfTN:"INTF_TABLE", keySep: COLON, memberTN:"LAG_MEMBER_TABLE"},
        stateDb:TblData{portTN:"LAG_TABLE", intfTN:"INTERFACE_TABLE", keySep: PIPE},
        CountersHdl:CounterData{OIDTN: "COUNTERS_PORT_NAME_MAP", CountersTN:"COUNTERS", PopulateCounters: populatePortCounters},
    },
    IntfTypeVlan : IntfTblData{
        cfgDb:TblData{portTN:"VLAN", memberTN: "VLAN_MEMBER", intfTN:"VLAN_INTERFACE", keySep: PIPE},
        appDb:TblData{portTN:"VLAN_TABLE", memberTN: "VLAN_MEMBER_TABLE", intfTN:"INTF_TABLE", keySep: COLON},
    },
    IntfTypeLoopback : IntfTblData {
       cfgDb:TblData{portTN:"LOOPBACK_INTERFACE", intfTN: "LOOPBACK_INTERFACE", keySep: PIPE},
       appDb:TblData{intfTN: "INTF_TABLE", keySep: COLON},
   },
}

var dbIdToTblMap = map[db.DBNum][]string {
    db.ConfigDB: {"PORT", "MGMT_PORT", "VLAN", "PORTCHANNEL", "LOOPBACK_INTERFACE"},
    db.ApplDB  : {"PORT_TABLE", "MGMT_PORT_TABLE", "VLAN_TABLE", "LAG_TABLE"},
    db.StateDB : {"PORT_TABLE", "MGMT_PORT_TABLE", "LAG_TABLE"},
}

var intfOCToSpeedMap = map[ocbinds.E_OpenconfigIfEthernet_ETHERNET_SPEED] string {
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10MB: "10",
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_100MB: "100",
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_1GB: "1000",
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_2500MB: "2500",
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_5GB: "5000",
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB: "10000",
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_25GB: "25000",
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_40GB: "40000",
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_50GB: "50000",
    ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_100GB: "100000",

}


type E_InterfaceType  int64
const (
    IntfTypeUnset           E_InterfaceType = 0
    IntfTypeEthernet        E_InterfaceType = 1
    IntfTypeMgmt            E_InterfaceType = 2
    IntfTypeVlan            E_InterfaceType = 3
    IntfTypePortChannel     E_InterfaceType = 4
    IntfTypeLoopback        E_InterfaceType = 5
)
type E_InterfaceSubType int64
const (
    IntfSubTypeUnset        E_InterfaceSubType = 0
    IntfSubTypeVlanL2  E_InterfaceSubType = 1
    InterfaceSubTypeVlanL3  E_InterfaceSubType = 2
)

func getIntfTypeByName (name string) (E_InterfaceType, E_InterfaceSubType, error) {

    var err error
    if strings.HasPrefix(name, ETHERNET) == true {
        return IntfTypeEthernet, IntfSubTypeUnset, err
    } else if strings.HasPrefix(name, MGMT) == true {
        return IntfTypeMgmt, IntfSubTypeUnset, err
    } else if strings.HasPrefix(name, VLAN) == true {
        return IntfTypeVlan, IntfSubTypeUnset, err
    } else if strings.HasPrefix(name, PORTCHANNEL) == true {
        return IntfTypePortChannel, IntfSubTypeUnset, err
    } else if strings.HasPrefix(name, LOOPBACK) == true {
        return IntfTypeLoopback, IntfSubTypeUnset, err
    } else {
        err = errors.New("Interface name prefix not matched with supported types")
        return IntfTypeUnset, IntfSubTypeUnset, err
    }
}

func getIntfsRoot (s *ygot.GoStruct) *ocbinds.OpenconfigInterfaces_Interfaces {
    deviceObj := (*s).(*ocbinds.Device)
    return deviceObj.Interfaces
}

/* RPC for clear counters */
var rpc_clear_counters RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
    var err error
    var result struct {
        Output struct {
            Status int32 `json:"status"`
            Status_detail string`json:"status-detail"`
        } `json:"sonic-interface:output"`
    }
    result.Output.Status = 1
    /* Get input data */
    var mapData map[string]interface{}
    err = json.Unmarshal(body, &mapData)
    if err != nil {
        log.Info("Failed to unmarshall given input data")
        result.Output.Status_detail = fmt.Sprintf("Error: Failed to unmarshall given input data")
        return json.Marshal(&result)
    }
    input, _ := mapData["sonic-interface:input"]
    mapData = input.(map[string]interface{})
    input = mapData["interface-param"]
    input_str := fmt.Sprintf("%v", input)
    input_str = strings.ToUpper(string(input_str))

    portOidmapTs := &db.TableSpec{Name: "COUNTERS_PORT_NAME_MAP"}
    ifCountInfo, err := dbs[db.CountersDB].GetMapAll(portOidmapTs)
    if err != nil {
        result.Output.Status_detail = fmt.Sprintf("Error: Port-OID (Counters) get for all the interfaces failed!")
        return json.Marshal(&result)
    }

    if input_str == "ALL" {
        log.Info("rpc_clear_counters : Clear Counters for all interfaces")
        for  intf, oid := range ifCountInfo.Field {
            verr, cerr := resetCounters(dbs[db.CountersDB], oid)
            if verr != nil || cerr != nil {
                log.Info("Failed to reset counters for ", intf)
            } else {
                log.Info("Counters reset for " + intf)
            }
        }
    } else if input_str == "ETHERNET" || input_str == "PORTCHANNEL" {
        log.Info("rpc_clear_counters : Reset counters for given interface type")
        for  intf, oid := range ifCountInfo.Field {
            if strings.HasPrefix(strings.ToUpper(intf), input_str) {
                verr, cerr := resetCounters(dbs[db.CountersDB], oid)
                if verr != nil || cerr != nil {
                    log.Error("Failed to reset counters for: ", intf)
                } else {
                    log.Info("Counters reset for " + intf)
                }
            }
        }
    } else {
        log.Info("rpc_clear_counters: Clear counters for given interface name")
        id := getIdFromIntfName(&input_str)
        if strings.HasPrefix(input_str, "ETHERNET") {
            input_str = "Ethernet" + id
        } else if strings.HasPrefix(input_str, "PORTCHANNEL") {
            input_str = "PortChannel" + id
        } else {
            log.Info("Invalid Interface")
            result.Output.Status_detail = fmt.Sprintf("Error: Clear Counters not supported for %s", input_str)
            return json.Marshal(&result)
        }
        oid, ok := ifCountInfo.Field[input_str]
        if !ok {
            result.Output.Status_detail = fmt.Sprintf("Error: OID info not found in COUNTERS_PORT_NAME_MAP for %s", input_str)
            return json.Marshal(&result)
        }
        verr, cerr := resetCounters(dbs[db.CountersDB], oid)
        if verr != nil {
            result.Output.Status_detail = fmt.Sprintf("Error: Failed to get counter values from COUNTERS table for %s", input_str)
            return json.Marshal(&result)
        }
        if cerr != nil {
            log.Info("Failed to reset counters values")
            result.Output.Status_detail = fmt.Sprintf("Error: Failed to reset counters values for %s.", input_str)
            return json.Marshal(&result)
        }
        log.Info("Counters reset for " + input_str)
    }
    result.Output.Status = 0
    result.Output.Status_detail = "Success: Cleared Counters"
    return json.Marshal(&result)
}

/* Reset counter values in COUNTERS_BACKUP table for given OID */
func resetCounters(d *db.DB, oid string) (error,error) {
    var verr,cerr error
    CountrTblTs := db.TableSpec {Name: "COUNTERS"}
    CountrTblTsCp := db.TableSpec { Name: "COUNTERS_BACKUP" }
    value, verr := d.GetEntry(&CountrTblTs, db.Key{Comp: []string{oid}})
    if verr == nil {
        secs := time.Now().Unix()
        timeStamp := strconv.FormatInt(secs, 10)
        value.Field["LAST_CLEAR_TIMESTAMP"] = timeStamp
        cerr = d.CreateEntry(&CountrTblTsCp, db.Key{Comp: []string{oid}}, value)
    }
    return verr, cerr
}

/* Extract ID from Intf String */
func getIdFromIntfName(intfName *string) (string) {
    var re = regexp.MustCompile("[0-9]+")
    id := re.FindStringSubmatch(*intfName)
    return id[0]
}

var YangToDb_intf_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    log.Info("Entering YangToDb_intf_tbl_key_xfmr")
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    ifName := pathInfo.Var("name")

    log.Info("Intf name: ", ifName)
    log.Info("Exiting YangToDb_intf_tbl_key_xfmr")
    intfType, _, ierr := getIntfTypeByName(ifName)
    if ierr != nil {
        log.Errorf("Extracting Interface type for Interface: %s failed!", ifName)
        return "", ierr
    }
    requestUriPath, err := getYangPathFromUri(inParams.requestUri)
    log.Info("inParams.requestUri: ", requestUriPath)
    if inParams.oper == DELETE && requestUriPath == "/openconfig-interfaces:interfaces/interface" {
        switch intfType {
        case IntfTypeVlan:
            /* VLAN Interface Delete Handling */
            /* Update the map for VLAN and VLAN MEMBER table */
            err := deleteVlanIntfAndMembers(&inParams, &ifName)
            if err != nil {
                log.Errorf("Deleting VLAN: %s failed!", ifName)
                return "", err
            }
        case IntfTypePortChannel:
            err := deleteLagIntfAndMembers(&inParams, &ifName)
            if err != nil {
                log.Errorf("Deleting LAG: %s failed!", ifName)
                return "", err
            }
        case IntfTypeLoopback:
            err := deleteLoopbackIntf(&inParams, &ifName)
            if err != nil {
                log.Errorf("Deleting Loopback: %s failed!", ifName)
                return "", err
            }
        }
        log.Errorf("Invalid interface for delete:%s", ifName)
        return "", err

    }
    return ifName, err
}

// Code for DBToYang - Key xfmr
var DbToYang_intf_tbl_key_xfmr  KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    log.Info("Entering DbToYang_intf_tbl_key_xfmr")
    res_map := make(map[string]interface{})

    log.Info("Interface Name = ", inParams.key)
    res_map["name"] = inParams.key 
    return res_map, nil
}

var intf_table_xfmr TableXfmrFunc = func (inParams XfmrParams) ([]string, error) {

    var tblList []string
    var err error

    log.Info("TableXfmrFunc - Uri: ", inParams.uri);
    pathInfo := NewPathInfo(inParams.uri)

    targetUriPath, err := getYangPathFromUri(pathInfo.Path)

    ifName := pathInfo.Var("name");
    if ifName == "" {
        log.Info("TableXfmrFunc - intf_table_xfmr Intf key is not present")

        if _, ok := dbIdToTblMap[inParams.curDb]; !ok {
            log.Info("TableXfmrFunc - intf_table_xfmr db id entry not present")
            return tblList, errors.New("Key not present")
        } else {
            return dbIdToTblMap[inParams.curDb], nil
        }
    }

    intfType, _, ierr := getIntfTypeByName(ifName)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("TableXfmrFunc - Invalid interface type IntfTypeUnset");
        return tblList, errors.New("Invalid interface type IntfTypeUnset");
    }
    intTbl := IntfTypeTblMap[intfType]
    log.Info("TableXfmrFunc - targetUriPath : ", targetUriPath)

    if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/config"){ 
        tblList = append(tblList, intTbl.cfgDb.portTN)
    } else if  strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/state/counters") {
        tblList = append(tblList, intTbl.CountersHdl.CountersTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/state") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/ethernet/state") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/state") {
        tblList = append(tblList, intTbl.appDb.portTN)
    } else if strings.HasPrefix(targetUriPath,"/openconfig-interfaces:interfaces/interface/openconfig-interfaces-ext:nat-zone/config")||
        strings.HasPrefix(targetUriPath,"/openconfig-interfaces:interfaces/interface/nat-zone/config") {
        tblList = append(tblList, intTbl.cfgDb.intfTN)
    } else if strings.HasPrefix(targetUriPath,"/openconfig-interfaces:interfaces/interface/openconfig-interfaces-ext:nat-zone/state")||
        strings.HasPrefix(targetUriPath,"/openconfig-interfaces:interfaces/interface/nat-zone/state") {
        tblList = append(tblList, intTbl.appDb.intfTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses/address/config") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv4/addresses/address/config") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses/address/config") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses/address/config") {
        tblList = append(tblList, intTbl.cfgDb.intfTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses/address/state") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv4/addresses/address/state") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses/address/state") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses/address/state") {
        tblList = append(tblList, intTbl.appDb.intfTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv4/addresses") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses") {
        tblList = append(tblList, intTbl.cfgDb.intfTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/ethernet") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet") {
        tblList = append(tblList, intTbl.cfgDb.portTN)
    } else if strings.HasPrefix(targetUriPath,"/openconfig-interfaces:interfaces/interface/openconfig-interfaces-ext:nat-zone") ||
        strings.HasPrefix(targetUriPath,"/openconfig-interfaces:interfaces/interface/nat-zone") {
        tblList = append(tblList, intTbl.cfgDb.intfTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface") {
        tblList = append(tblList, intTbl.cfgDb.portTN)
    } else {       err = errors.New("Invalid URI")
    }

    log.Infof("TableXfmrFunc - uri(%v), tblList(%v)\r\n", inParams.uri, tblList);
    return tblList, err
}

var YangToDb_intf_name_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    ifName := pathInfo.Var("name")

    if strings.HasPrefix(ifName, VLAN) == true {
        vlanId := ifName[len("Vlan"):len(ifName)]
        res_map["vlanid"] = vlanId
    } else if strings.HasPrefix(ifName, PORTCHANNEL) == true {
        res_map["NULL"] = "NULL"
    } else if strings.HasPrefix(ifName, LOOPBACK) == true {
        res_map["NULL"] = "NULL"
    }
    log.Info("YangToDb_intf_name_xfm: res_map:", res_map)
    return res_map, err
}

var DbToYang_intf_name_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    log.Info("Entering DbToYang_intf_tbl_key_xfmr")
    res_map := make(map[string]interface{})

    pathInfo := NewPathInfo(inParams.uri)
    ifName:= pathInfo.Var("name")
    log.Info("Interface Name = ", ifName)
    res_map["name"] = ifName
    return res_map, nil
}

var YangToDb_intf_name_empty_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    return res_map, err
}

var YangToDb_intf_enabled_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    enabled, _ := inParams.param.(*bool)
    var enStr string
    if *enabled == true {
        enStr = "up"
    } else {
        enStr = "down"
    }
    res_map[PORT_ADMIN_STATUS] = enStr

    return res_map, nil
}


func getPortTableNameByDBId (intftbl IntfTblData, curDb db.DBNum) (string, error) {

    var tblName string

    switch (curDb) {
    case db.ConfigDB:
        tblName = intftbl.cfgDb.portTN
    case db.ApplDB:
        tblName = intftbl.appDb.portTN
    case db.StateDB:
        tblName = intftbl.stateDb.portTN
    default:
        tblName = intftbl.cfgDb.portTN
    }

    return tblName, nil
}

var DbToYang_intf_enabled_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]

    intfType, _, ierr := getIntfTypeByName(inParams.key)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("DbToYang_intf_enabled_xfmr - Invalid interface type IntfTypeUnset");
        return result, errors.New("Invalid interface type IntfTypeUnset");
    }
    intTbl := IntfTypeTblMap[intfType]

    tblName, _ := getPortTableNameByDBId(intTbl, inParams.curDb)
    if _, ok := data[tblName]; !ok {
        log.Info("DbToYang_intf_enabled_xfmr table not found : ", tblName)
        return result, errors.New("table not found : " + tblName)
    }

    pTbl := data[tblName]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_intf_enabled_xfmr Interface not found : ", inParams.key)
        return result, errors.New("Interface not found : " + inParams.key)
    }
    prtInst := pTbl[inParams.key]
    adminStatus, ok := prtInst.Field[PORT_ADMIN_STATUS]
    if ok {
        if adminStatus == "up" {
            result["enabled"] = true
        } else {
            result["enabled"] = false
        }
    } else {
        log.Info("Admin status field not found in DB")
    }
    return result, err
}

var DbToYang_intf_admin_status_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]

    intfType, _, ierr := getIntfTypeByName(inParams.key)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("DbToYang_intf_admin_status_xfmr - Invalid interface type IntfTypeUnset");
        return result, errors.New("Invalid interface type IntfTypeUnset");
    }
    intTbl := IntfTypeTblMap[intfType]

    tblName, _ := getPortTableNameByDBId(intTbl, inParams.curDb)
    if _, ok := data[tblName]; !ok {
        log.Info("DbToYang_intf_admin_status_xfmr table not found : ", tblName)
        return result, errors.New("table not found : " + tblName)
    }
    pTbl := data[tblName]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_intf_admin_status_xfmr Interface not found : ", inParams.key)
        return result, errors.New("Interface not found : " + inParams.key)
    }
    prtInst := pTbl[inParams.key]
    adminStatus, ok := prtInst.Field[PORT_ADMIN_STATUS]
    var status ocbinds.E_OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus
    if ok {
        if adminStatus == "up" {
            status = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_UP
        } else {
            status = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_DOWN
        }
        result["admin-status"] = ocbinds.E_OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus.ΛMap(status)["E_OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus"][int64(status)].Name
    } else {
        log.Info("Admin status field not found in DB")
    }

    return result, err
}

var DbToYang_intf_oper_status_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})
    var prtInst db.Value

    data := (*inParams.dbDataMap)[inParams.curDb]
    intfType, _, ierr := getIntfTypeByName(inParams.key)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("DbToYang_intf_oper_status_xfmr - Invalid interface type IntfTypeUnset");
        return result, errors.New("Invalid interface type IntfTypeUnset");
    }
    intTbl := IntfTypeTblMap[intfType]
    if intfType == IntfTypeMgmt {
        pathInfo := NewPathInfo(inParams.uri)
        ifName := pathInfo.Var("name");
        entry, dbErr := inParams.dbs[db.StateDB].GetEntry(&db.TableSpec{Name:intTbl.stateDb.portTN}, db.Key{Comp: []string{ifName}})
        if dbErr != nil {
            log.Info("Failed to read mgmt port status from state DB, " + intTbl.stateDb.portTN + " " + ifName)
            return result, dbErr
        }
        prtInst = entry
    } else {
        tblName, _ := getPortTableNameByDBId(intTbl, inParams.curDb)
        pTbl := data[tblName]
        prtInst = pTbl[inParams.key]
    }

    operStatus, ok := prtInst.Field[PORT_OPER_STATUS]
    var status ocbinds.E_OpenconfigInterfaces_Interfaces_Interface_State_OperStatus
    if ok {
        if operStatus == "up" {
            status = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_UP
        } else {
            status = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_DOWN
        }
        result["oper-status"] = ocbinds.E_OpenconfigInterfaces_Interfaces_Interface_State_OperStatus.ΛMap(status)["E_OpenconfigInterfaces_Interfaces_Interface_State_OperStatus"][int64(status)].Name
    } else {
        log.Info("Oper status field not found in DB")
    }

    return result, err
}

var DbToYang_intf_eth_auto_neg_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    intfType, _, ierr := getIntfTypeByName(inParams.key)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("DbToYang_intf_eth_auto_neg_xfmr - Invalid interface type IntfTypeUnset");
        return result, errors.New("Invalid interface type IntfTypeUnset");
    }
    intTbl := IntfTypeTblMap[intfType]

    tblName, _ := getPortTableNameByDBId(intTbl, inParams.curDb)
    pTbl := data[tblName]
    prtInst := pTbl[inParams.key]
    autoNeg, ok := prtInst.Field[PORT_AUTONEG]
    if ok {
        if autoNeg == "true" {
            result["auto-negotiate"] = true
        } else {
            result["auto-negotiate"] = false
        }
    } else {
        log.Info("auto-negotiate field not found in DB")
    }
    return result, err
}

var DbToYang_intf_eth_port_speed_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    intfType, _, ierr := getIntfTypeByName(inParams.key)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("DbToYang_intf_eth_port_speed_xfmr - Invalid interface type IntfTypeUnset");
        return result, errors.New("Invalid interface type IntfTypeUnset");
    }
    intTbl := IntfTypeTblMap[intfType]

    tblName, _ := getPortTableNameByDBId(intTbl, inParams.curDb)
    pTbl := data[tblName]
    prtInst := pTbl[inParams.key]
    speed, ok := prtInst.Field[PORT_SPEED]
    portSpeed := ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
    if ok {
        portSpeed, err = getDbToYangSpeed(speed)
        result["port-speed"] = ocbinds.E_OpenconfigIfEthernet_ETHERNET_SPEED.ΛMap(portSpeed)["E_OpenconfigIfEthernet_ETHERNET_SPEED"][int64(portSpeed)].Name
    } else {
        log.Info("Speed field not found in DB")
    }

    return result, err
}



func getDbToYangSpeed (speed string) (ocbinds.E_OpenconfigIfEthernet_ETHERNET_SPEED, error) {
    portSpeed := ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_UNKNOWN
    var err error = errors.New("Not found in port speed map")
    for k, v := range intfOCToSpeedMap {
        if speed == v {
            portSpeed = k
            err = nil
        }
    }
    return portSpeed, err
}

func intf_intf_tbl_key_gen (intfName string, ip string, prefixLen int, keySep string) string {
    return intfName + keySep + ip + "/" + strconv.Itoa(prefixLen)
}

var YangToDb_intf_subintfs_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var subintf_key string
    var err error

    return subintf_key, err
}

var DbToYang_intf_subintfs_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    rmap["index"] = 0
    return rmap, err
}


func intf_ip_addr_del (d *db.DB , ifName string, tblName string, subIntf *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface) (map[string]map[string]db.Value, error) {
    var err error
    subIntfmap := make(map[string]map[string]db.Value)
    intfIpMap := make(map[string]db.Value)

    if subIntf.Ipv4 != nil && subIntf.Ipv4.Addresses != nil {
        if len(subIntf.Ipv4.Addresses.Address) < 1 {
            ipMap, _:= getIntfIpByName(d, tblName, ifName, true, false, "")
            if ipMap != nil && len(ipMap) > 0 {
                for k, v := range ipMap {
                    intfIpMap[k] = v
                }
            }
        } else {
            for ip, _ := range subIntf.Ipv4.Addresses.Address {
                ipMap, _ := getIntfIpByName(d, tblName, ifName, true, false, ip)

                if ipMap != nil && len(ipMap) > 0 {
                    for k, v := range ipMap {
                        intfIpMap[k] = v
                    }
                }
            }
        }
    }

    if subIntf.Ipv6 != nil && subIntf.Ipv6.Addresses != nil {
        if len(subIntf.Ipv6.Addresses.Address) < 1 {
            ipMap, _ := getIntfIpByName(d, tblName, ifName, false, true, "")
            if ipMap != nil && len(ipMap) > 0 {
                for k, v := range ipMap {
                    intfIpMap[k] = v
                }
            }
        } else {
            for ip, _ := range subIntf.Ipv6.Addresses.Address {
                ipMap, _ := getIntfIpByName(d, tblName, ifName, false, true, ip)

                if ipMap != nil && len(ipMap) > 0 {
                    for k, v := range ipMap {
                        intfIpMap[k] = v
                    }
                }
            }
        }
    }
    if len(intfIpMap) > 0 {
        if _, ok := subIntfmap[tblName]; !ok {
            subIntfmap[tblName] = make (map[string]db.Value)
        }
        var data db.Value
        for k, _ := range intfIpMap {
            ifKey := ifName + "|" + k
            subIntfmap[tblName][ifKey] = data
        }
        count := 0
        _ = interfaceIPcount(tblName, d, &ifName, &count)
        if (count - len(intfIpMap)) == 1 {
            _, dbErr := d.GetEntry(&db.TableSpec{Name:tblName}, db.Key{Comp: []string{ifName}})
            if dbErr == nil {
                subIntfmap[tblName][ifName] = data
            }
        }
    }
    log.Info("Delete IP address list ", subIntfmap,  " ", err)
    return subIntfmap, err
}

/* Validate IP exists in the INTERFACE table of corresponding Interface type */
func validateIPExists(tblName string, d *db.DB, ifName *string) error {
    ipCnt := 0
    _    = interfaceIPcount(tblName, d, ifName, &ipCnt)
    if ipCnt > 0 {
        errStr := "L3 Configuration exists for Interface: " + *ifName
        log.Error(errStr)
        return errors.New(errStr)
    }
    return nil
}

/* Note: This function can be extended for IP validations for all Interface types */
func validateIpForIntfType(ifType E_InterfaceType, ip *string, prfxLen *uint8, isIpv4 bool) error {
    var err error

    switch ifType {
    case IntfTypeLoopback:
        if(isIpv4) {
            if *prfxLen != 32 {
                errStr := "Not supported prefix length (32 is supported)"
                err = tlerr.InvalidArgsError{Format:errStr}
                return err
            }
        } else {
            if(*prfxLen != 128) {
                errStr := "Not supported prefix length (128 is supported)"
                err = tlerr.InvalidArgsError{Format:errStr}
                return err
            }
        }
    }
    return err
}


var YangToDb_intf_ip_addr_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    var err error
    subIntfmap := make(map[string]map[string]db.Value)

    intfsObj := getIntfsRoot(inParams.ygRoot)
    if intfsObj == nil || len(intfsObj.Interface) < 1 {
        log.Info("YangToDb_intf_subintf_ip_xfmr : IntfsObj/interface list is empty.")
        return subIntfmap, errors.New("IntfsObj/Interface is not specified")
    }
    pathInfo := NewPathInfo(inParams.uri)
    ifName := pathInfo.Var("name")

    if ifName == "" {
        errStr := "Interface KEY not present"
        log.Info("YangToDb_intf_subintf_ip_xfmr : " + errStr)
        return subIntfmap, errors.New(errStr)
    }

    intfType, _, ierr := getIntfTypeByName(ifName)
    if intfType == IntfTypeUnset || ierr != nil {
        errStr := "Invalid interface type IntfTypeUnset"
        log.Info("YangToDb_intf_subintf_ip_xfmr : " + errStr)
        return subIntfmap, errors.New(errStr)
    }
    /* Validate whether the Interface is configured as member-port associated with any vlan */
    if intfType == IntfTypeEthernet || intfType == IntfTypePortChannel {
        err = validateIntfAssociatedWithVlan(inParams.d, &ifName)
        if err != nil {
            return subIntfmap, err
        }
    }

    if _, ok := intfsObj.Interface[ifName]; !ok {
        errStr := "Interface entry not found in Ygot tree, ifname: " + ifName
        log.Info("YangToDb_intf_subintf_ip_xfmr : " + errStr)
        return subIntfmap, errors.New(errStr)
    }

    intfObj := intfsObj.Interface[ifName]

    if intfObj.Subinterfaces == nil || len(intfObj.Subinterfaces.Subinterface) < 1 {
        errStr := "SubInterface node is not set"
        log.Info("YangToDb_intf_subintf_ip_xfmr : " + errStr)
        return subIntfmap, errors.New(errStr)
    }
    if _, ok := intfObj.Subinterfaces.Subinterface[0]; !ok {
        log.Info("YangToDb_intf_subintf_ip_xfmr : No IP address handling required")
        return subIntfmap, err
    }

    intTbl := IntfTypeTblMap[intfType]
    tblName, _ := getIntfTableNameByDBId(intTbl, inParams.curDb)

    subIntfObj := intfObj.Subinterfaces.Subinterface[0]
    if inParams.oper == DELETE {
        return intf_ip_addr_del(inParams.d, ifName, tblName, subIntfObj)
    }

    entry, dbErr := inParams.d.GetEntry(&db.TableSpec{Name:intTbl.cfgDb.intfTN}, db.Key{Comp: []string{ifName}})
    if dbErr != nil || !entry.IsPopulated() {
        ifdb := make(map[string]string)
        ifdb["NULL"] = "NULL"
        value := db.Value{Field: ifdb}
        if _, ok := subIntfmap[tblName]; !ok {
            subIntfmap[tblName] = make(map[string]db.Value)
        }
        subIntfmap[tblName][ifName] = value

    }

    if subIntfObj.Ipv4 != nil && subIntfObj.Ipv4.Addresses != nil {
        for ip, _ := range subIntfObj.Ipv4.Addresses.Address {
            addr := subIntfObj.Ipv4.Addresses.Address[ip]
            if addr.Config != nil {
                if addr.Config.Ip == nil {
                    addr.Config.Ip = new(string)
                    *addr.Config.Ip = ip
                }
                log.Info("Ip:=", *addr.Config.Ip)
                log.Info("prefix:=", *addr.Config.PrefixLength)
                if !validIPv4(*addr.Config.Ip) {
                    errStr := "Invalid IPv4 address " + *addr.Config.Ip
                    err = tlerr.InvalidArgsError{Format: errStr}
                    return subIntfmap, err
                }
                /* Validate IP specific to Interface type */
                err = validateIpForIntfType(intfType, addr.Config.Ip, addr.Config.PrefixLength,  true)
                if err != nil {
                    return subIntfmap, err
                }

                intf_key := intf_intf_tbl_key_gen(ifName, *addr.Config.Ip, int(*addr.Config.PrefixLength), "|")
                m := make(map[string]string)
                if addr.Config.GwAddr != nil {
                    if intfType != IntfTypeMgmt {
                        errStr := "GwAddr config is not supported " + ifName
                        log.Info("GwAddr config is not supported for intfType: ", intfType, " " , ifName)
                        return subIntfmap, errors.New(errStr)
                    }
                    if !validIPv4(*addr.Config.GwAddr) {
                        errStr := "Invalid IPv4 Gateway address " + *addr.Config.GwAddr
                        err = tlerr.InvalidArgsError{Format: errStr}
                        return subIntfmap, err
                    }
                    m["gwaddr"] = *addr.Config.GwAddr
                } else {
                    m["NULL"] = "NULL"
                }
                value := db.Value{Field: m}
                if _, ok := subIntfmap[tblName]; !ok {
                    subIntfmap[tblName] = make(map[string]db.Value)
                }
                subIntfmap[tblName][intf_key] = value
                log.Info("tblName :", tblName, "intf_key: ", intf_key, "data : ", value)

            }
        }
    }
    if subIntfObj.Ipv6 != nil && subIntfObj.Ipv6.Addresses != nil {
        for ip, _ := range subIntfObj.Ipv6.Addresses.Address {
            addr := subIntfObj.Ipv6.Addresses.Address[ip]
            if addr.Config != nil {
                if addr.Config.Ip == nil {
                    addr.Config.Ip = new(string)
                    *addr.Config.Ip = ip
                }
                log.Info("Ipv6 IP:=", *addr.Config.Ip)
                log.Info("Ipv6 prefix:=", *addr.Config.PrefixLength)
                if !validIPv6(*addr.Config.Ip) {
                    errStr := "Invalid IPv6 address " + *addr.Config.Ip
                    err = tlerr.InvalidArgsError{Format: errStr}
                    return subIntfmap, err
                }
                /* Validate IP specific to Interface type */
                err = validateIpForIntfType(intfType, addr.Config.Ip, addr.Config.PrefixLength, false)
                if err != nil {
                    return subIntfmap, err
                }

                intf_key := intf_intf_tbl_key_gen(ifName, *addr.Config.Ip, int(*addr.Config.PrefixLength), "|")
                m := make(map[string]string)
                if addr.Config.GwAddr != nil {
                    if intfType != IntfTypeMgmt {
                        errStr := "GwAddr config is not supported " + ifName
                        log.Info("GwAddr config is not supported for intfType: ", intfType, " " , ifName)
                        return subIntfmap, errors.New(errStr)
                    }
                    if !validIPv6(*addr.Config.GwAddr) {
                        errStr := "Invalid IPv6 Gateway address " + *addr.Config.GwAddr
                        err = tlerr.InvalidArgsError{Format: errStr}
                        return subIntfmap, err
                    }
                    m["gwaddr"] = *addr.Config.GwAddr
                } else {
                    m["NULL"] = "NULL"
                }
                value := db.Value{Field: m}
                if _, ok := subIntfmap[tblName]; !ok {
                    subIntfmap[tblName] = make(map[string]db.Value)
                }
                subIntfmap[tblName][intf_key] = value
                log.Info("tblName :", tblName, "intf_key: ", intf_key, "data : ", value)
            }
        }
    }
    log.Info("YangToDb_intf_subintf_ip_xfmr : subIntfmap : ",  subIntfmap)

    return subIntfmap, err
}

func convertIpMapToOC (intfIpMap map[string]db.Value, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface, isState bool) error {
    var subIntf *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface
    var err error

    if _, ok := ifInfo.Subinterfaces.Subinterface[0]; !ok {
        subIntf, err = ifInfo.Subinterfaces.NewSubinterface(0)
        if err != nil {
            log.Error("Creation of subinterface subtree failed!")
            return err
        }
    }

    subIntf = ifInfo.Subinterfaces.Subinterface[0]
    ygot.BuildEmptyTree(subIntf)

    for ipKey, ipdata := range intfIpMap {
        log.Info("IP address = ", ipKey)
        ipB, ipNetB, _ := net.ParseCIDR(ipKey)
        v4Flag := false
        v6Flag := false

        var v4Address *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address
        var v6Address *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address
        if validIPv4(ipB.String()) {
            if _, ok := subIntf.Ipv4.Addresses.Address[ipB.String()]; !ok {
                v4Address, err = subIntf.Ipv4.Addresses.NewAddress(ipB.String())
            }
            v4Address = subIntf.Ipv4.Addresses.Address[ipB.String()]
            v4Flag = true
        } else if validIPv6(ipB.String()) {
            if _, ok := subIntf.Ipv6.Addresses.Address[ipB.String()]; !ok {
                v6Address, err = subIntf.Ipv6.Addresses.NewAddress(ipB.String())
            }
            v6Address =  subIntf.Ipv6.Addresses.Address[ipB.String()]
            v6Flag = true
        } else {
            log.Error("Invalid IP address " + ipB.String())
            continue
        }
        if err != nil {
            log.Error("Creation of address subtree failed!")
            return err
        }
        if v4Flag {
            ygot.BuildEmptyTree(v4Address)
            ipStr := new(string)
            *ipStr = ipB.String()
            v4Address.Ip = ipStr
            ipNetBNum, _ := ipNetB.Mask.Size()
            prfxLen := new(uint8)
            *prfxLen = uint8(ipNetBNum)
            if isState {
                v4Address.State.Ip = ipStr
                v4Address.State.PrefixLength = prfxLen
                if ipdata.Has("gwaddr") {
                    gwaddr := new(string)
                    *gwaddr = ipdata.Get("gwaddr")
                    v4Address.State.GwAddr = gwaddr
                }
            } else {
                v4Address.Config.Ip = ipStr
                v4Address.Config.PrefixLength = prfxLen
                if ipdata.Has("gwaddr") {
                    gwaddr := new(string)
                    *gwaddr = ipdata.Get("gwaddr")
                    v4Address.Config.GwAddr = gwaddr
                }
            }
        }
        if v6Flag {
            ygot.BuildEmptyTree(v6Address)
            ipStr := new(string)
            *ipStr = ipB.String()
            v6Address.Ip = ipStr
            ipNetBNum, _ := ipNetB.Mask.Size()
            prfxLen := new(uint8)
            *prfxLen = uint8(ipNetBNum)
            if isState {
                v6Address.State.Ip = ipStr
                v6Address.State.PrefixLength = prfxLen
                if ipdata.Has("gwaddr") {
                    gwaddr := new(string)
                    *gwaddr = ipdata.Get("gwaddr")
                    v6Address.State.GwAddr = gwaddr
                }
            } else {
                v6Address.Config.Ip = ipStr
                v6Address.Config.PrefixLength = prfxLen
                if ipdata.Has("gwaddr") {
                    gwaddr := new(string)
                    *gwaddr = ipdata.Get("gwaddr")
                    v6Address.Config.GwAddr = gwaddr
                }
            }
        }
    }
    return err
}

func interfaceIPcount(tblName string, d *db.DB, intfName *string, ipCnt *int) error {
    intfIPKeys, _ := d.GetKeys(&db.TableSpec{Name:tblName})
    if len(intfIPKeys) > 0 {
        for i := range intfIPKeys {
            if *intfName == intfIPKeys[i].Get(0) {
                *ipCnt = *ipCnt+1
            }
        }
    }
    return nil
}

/* Function to delete Loopback Interface */
func deleteLoopbackIntf(inParams *XfmrParams, loName *string) error {
    var err error
    intTbl := IntfTypeTblMap[IntfTypeLoopback]
    subOpMap := make(map[db.DBNum]map[string]map[string]db.Value)
    resMap := make(map[string]map[string]db.Value)
    loMap := make(map[string]db.Value)

    loMap[*loName] = db.Value{Field:map[string]string{}}

    _, err = inParams.d.GetEntry(&db.TableSpec{Name:intTbl.cfgDb.portTN}, db.Key{Comp: []string{*loName}})
    if err != nil {
        log.Errorf("Retrieving data from LOOPBACK_INTERFACE table for Loopback: %s failed!", *loName)
        return err
    }
    err = validateIPExists(intTbl.cfgDb.intfTN, inParams.d, loName)
	if err != nil {
		return err
	}
    resMap[intTbl.cfgDb.intfTN] = loMap

    subOpMap[db.ConfigDB] = resMap
    inParams.subOpDataMap[DELETE] = &subOpMap
    return err
}

func getIntfIpByName(dbCl *db.DB, tblName string, ifName string, ipv4 bool, ipv6 bool, ip string) (map[string]db.Value, error) {
    var err error
    intfIpMap := make(map[string]db.Value)
    all := true
    if ipv4 == false || ipv6 == false {
        all = false
    }
    log.Info("Updating Interface IP Info from DB to Internal DS for Interface Name : ", ifName)

    keys,_ := doGetAllIpKeys(dbCl, &db.TableSpec{Name:tblName})

    for _, key := range keys {
        if len(key.Comp) < 2 {
            continue
        }
        if key.Get(0) != ifName {
            continue
        }
        if len(key.Comp) > 2 {
            for i, _ := range key.Comp {
                if i == 0 || i == 1 {
                    continue
                }
                key.Comp[1] = key.Comp[1] + ":" + key.Comp[i]
            }
        }
        if all == false {
            ipB, _, _ := net.ParseCIDR(key.Get(1))
            if ((validIPv4(ipB.String()) && (ipv4 == false)) ||
                (validIPv6(ipB.String()) && (ipv6 == false))) {
                continue
            }
            if ip != "" {
                if ipB.String() != ip {
                    continue
                }
            }
        }

        ipInfo, _ := dbCl.GetEntry(&db.TableSpec{Name:tblName}, db.Key{Comp: []string{key.Get(0), key.Get(1)}})
        intfIpMap[key.Get(1)]= ipInfo
    }
    return intfIpMap, err 
}

func handleIntfIPGetByTargetURI (inParams XfmrParams, targetUriPath string, ifName string, intfObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface) error {
    ipMap := make(map[string]db.Value)
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    ipAddr := pathInfo.Var("ip")
    intfType, _, ierr := getIntfTypeByName(ifName)
    if intfType == IntfTypeUnset || ierr != nil {
        errStr := "Invalid interface type IntfTypeUnset"
        log.Info("YangToDb_intf_subintf_ip_xfmr : " + errStr)
        return errors.New(errStr)
    }
    intTbl := IntfTypeTblMap[intfType]

    if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses/address/config") ||
       strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv4/addresses/address/config") {
           ipMap, err = getIntfIpByName(inParams.dbs[db.ConfigDB], intTbl.cfgDb.intfTN, ifName, true, false, ipAddr)
           log.Info("handleIntfIPGetByTargetURI : ipv4 config ipMap - : ", ipMap)
           convertIpMapToOC(ipMap, intfObj, false)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses/address/config") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses/address/config") {
           ipMap, err = getIntfIpByName(inParams.dbs[db.ConfigDB], intTbl.cfgDb.intfTN, ifName, false, true, ipAddr)
           log.Info("handleIntfIPGetByTargetURI : ipv6 config ipMap - : ", ipMap)
           convertIpMapToOC(ipMap, intfObj, false)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses/address/state") ||
         strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv4/addresses/address/state") {
           ipMap, err = getIntfIpByName(inParams.dbs[db.ApplDB], intTbl.appDb.intfTN, ifName, true, false, ipAddr)
           log.Info("handleIntfIPGetByTargetURI : ipv4 state ipMap - : ", ipMap)
           convertIpMapToOC(ipMap, intfObj, true)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses/address/state") ||
         strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses/address/state") {
           ipMap, err = getIntfIpByName(inParams.dbs[db.ApplDB], intTbl.appDb.intfTN, ifName, false, true, ipAddr)
           log.Info("handleIntfIPGetByTargetURI : ipv6 state ipMap - : ", ipMap)
           convertIpMapToOC(ipMap, intfObj, true)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv4/addresses") {
        ipMap, err = getIntfIpByName(inParams.dbs[db.ConfigDB], intTbl.cfgDb.intfTN, ifName, true, false, ipAddr)
           log.Info("handleIntfIPGetByTargetURI : ipv4 config ipMap - : ", ipMap)
        convertIpMapToOC(ipMap, intfObj, false)
        ipMap, err = getIntfIpByName(inParams.dbs[db.ApplDB], intTbl.appDb.intfTN, ifName, true, false, ipAddr)
           log.Info("handleIntfIPGetByTargetURI : ipv4 state ipMap - : ", ipMap)
        convertIpMapToOC(ipMap, intfObj, true)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses") {
        ipMap, err = getIntfIpByName(inParams.dbs[db.ConfigDB], intTbl.cfgDb.intfTN, ifName, false, true, ipAddr)
           log.Info("handleIntfIPGetByTargetURI : ipv6 config ipMap - : ", ipMap)
        convertIpMapToOC(ipMap, intfObj, false)
        ipMap, err = getIntfIpByName(inParams.dbs[db.ApplDB], intTbl.appDb.intfTN, ifName, false, true, ipAddr)
           log.Info("handleIntfIPGetByTargetURI : ipv6 state ipMap - : ", ipMap)
        convertIpMapToOC(ipMap, intfObj, true)
    }
    return err
}

var DbToYang_intf_ip_addr_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) (error) {
    var err error
    intfsObj := getIntfsRoot(inParams.ygRoot)
    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")
    targetUriPath, err := getYangPathFromUri(inParams.uri)
    log.Info("targetUriPath is ", targetUriPath)
    var intfObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface

    if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces") {
        if intfsObj != nil && intfsObj.Interface != nil && len(intfsObj.Interface) > 0 {
            var ok bool = false
            if intfObj, ok = intfsObj.Interface[intfName]; !ok {
                intfObj, _ = intfsObj.NewInterface(intfName)
            }
            ygot.BuildEmptyTree(intfObj)
            if intfObj.Subinterfaces == nil {
                ygot.BuildEmptyTree(intfObj.Subinterfaces)
            }
        } else {
            ygot.BuildEmptyTree(intfsObj)
            intfObj, _ = intfsObj.NewInterface(intfName)
            ygot.BuildEmptyTree(intfObj)
        }


    } else {
        err = errors.New("Invalid URI : " + targetUriPath)
    }
    err = handleIntfIPGetByTargetURI(inParams, targetUriPath, intfName, intfObj)

    return err
}

func validIPv4(ipAddress string) bool {
    ipAddress = strings.Trim(ipAddress, " ")

    re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
    if re.MatchString(ipAddress) {
        return true
    }
    return false
}

func validIPv6(ip6Address string) bool {
    ip6Address = strings.Trim(ip6Address, " ")
    re, _ := regexp.Compile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`)
    if re.MatchString(ip6Address) {
        return true
    }
    return false
}

func doGetAllIpKeys(d *db.DB, dbSpec *db.TableSpec) ([]db.Key, error) {

    var keys []db.Key

    intfTable, err := d.GetTable(dbSpec)
    if err != nil {
        return keys, err
    }

    keys, err = intfTable.GetKeys()
    log.Infof("Found %d INTF table keys", len(keys))
    return keys, err
}

func getMemTableNameByDBId (intftbl IntfTblData, curDb db.DBNum) (string, error) {

    var tblName string

    switch (curDb) {
    case db.ConfigDB:
        tblName = intftbl.cfgDb.memberTN
    case db.ApplDB:
        tblName = intftbl.appDb.memberTN
    case db.StateDB:
        tblName = intftbl.stateDb.memberTN
    default:
        tblName = intftbl.cfgDb.memberTN
    }

    return tblName, nil
}

func getIntfTableNameByDBId (intftbl IntfTblData, curDb db.DBNum) (string, error) {

    var tblName string

    switch (curDb) {
    case db.ConfigDB:
        tblName = intftbl.cfgDb.intfTN
    case db.ApplDB:
        tblName = intftbl.appDb.intfTN
    case db.StateDB:
        tblName = intftbl.stateDb.intfTN
    default:
        tblName = intftbl.cfgDb.intfTN
    }

    return tblName, nil
}



func getIntfCountersTblKey (d *db.DB, ifKey string) (string, error) {
    var oid string

    portOidCountrTblTs := &db.TableSpec{Name: "COUNTERS_PORT_NAME_MAP"}
    ifCountInfo, err := d.GetMapAll(portOidCountrTblTs)
    if err != nil {
        log.Error("Port-OID (Counters) get for all the interfaces failed!")
        return oid, err
    }

    if ifCountInfo.IsPopulated() {
        _, ok := ifCountInfo.Field[ifKey]
        if !ok {
            err = errors.New("OID info not found from Counters DB for interface " + ifKey)
        } else {
            oid = ifCountInfo.Field[ifKey]
        }
    } else {
        err = errors.New("Get for OID info from all the interfaces from Counters DB failed!")
    }

    return oid, err
}

func getSpecificCounterAttr(targetUriPath string, entry *db.Value, entry_backup *db.Value, counter_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters) (bool, error) {

    var e error

    switch targetUriPath {
    case "/openconfig-interfaces:interfaces/interface/state/counters/in-octets":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_IN_OCTETS", &counter_val.InOctets)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-unicast-pkts":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_IN_UCAST_PKTS", &counter_val.InUnicastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-broadcast-pkts":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_IN_BROADCAST_PKTS", &counter_val.InBroadcastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-multicast-pkts":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_IN_MULTICAST_PKTS", &counter_val.InMulticastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-errors":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_IN_ERRORS", &counter_val.InErrors)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-discards":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_IN_DISCARDS", &counter_val.InDiscards)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-pkts":
        var inNonUCastPkt, inUCastPkt *uint64
        var in_pkts uint64

        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_IN_NON_UCAST_PKTS", &inNonUCastPkt)
        if e == nil {
            e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_IN_UCAST_PKTS", &inUCastPkt)
            if e != nil {
                return true, e
            }
            in_pkts = *inUCastPkt + *inNonUCastPkt
            counter_val.InPkts = &in_pkts
            return true, e
        } else {
            return true, e
        }

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-octets":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_OUT_OCTETS", &counter_val.OutOctets)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-unicast-pkts":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_OUT_UCAST_PKTS", &counter_val.OutUnicastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-broadcast-pkts":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_OUT_BROADCAST_PKTS", &counter_val.OutBroadcastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-multicast-pkts":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_OUT_MULTICAST_PKTS", &counter_val.OutMulticastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-errors":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_OUT_ERRORS", &counter_val.OutErrors)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-discards":
        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_OUT_DISCARDS", &counter_val.OutDiscards)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/last-clear":
        timestampStr := (entry_backup.Field["LAST_CLEAR_TIMESTAMP"])
        timestamp, _ := strconv.ParseUint(timestampStr, 10, 64)
        counter_val.LastClear = &timestamp
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-pkts":
        var outNonUCastPkt, outUCastPkt *uint64
        var out_pkts uint64

        e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_OUT_NON_UCAST_PKTS", &outNonUCastPkt)
        if e == nil {
            e = getCounters(entry, entry_backup, "SAI_PORT_STAT_IF_OUT_UCAST_PKTS", &outUCastPkt)
            if e != nil {
                return true, e
            }
            out_pkts = *outUCastPkt + *outNonUCastPkt
            counter_val.OutPkts = &out_pkts
            return true, e
        } else {
            return true, e
        }


    default:
        log.Infof(targetUriPath + " - Not an interface state counter attribute")
    }
    return false, nil
}

func getCounters(entry *db.Value, entry_backup *db.Value, attr string, counter_val **uint64 ) error {

    var ok bool = false
    var err error
    val1, ok := entry.Field[attr]
    if !ok {
        return errors.New("Attr " + attr + "doesn't exist in IF table Map!")
    }
    val2, ok := entry_backup.Field[attr]
    if !ok {
        return errors.New("Attr " + attr + "doesn't exist in IF backup table Map!")
    }

    if len(val1) > 0 {
        v, _ := strconv.ParseUint(val1, 10, 64)
        v_backup, _ := strconv.ParseUint(val2, 10, 64)
        val := v-v_backup
        *counter_val = &val
        return nil
    }
    return err
}

var portCntList [] string = []string {"in-octets", "in-unicast-pkts", "in-broadcast-pkts", "in-multicast-pkts",
"in-errors", "in-discards", "in-pkts", "out-octets", "out-unicast-pkts",
"out-broadcast-pkts", "out-multicast-pkts", "out-errors", "out-discards",
"out-pkts","last-clear"}
var populatePortCounters PopulateIntfCounters = func (inParams XfmrParams, counter *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters) (error) {
    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)

    log.Info("PopulateIntfCounters : inParams.curDb : ", inParams.curDb, "D: ", inParams.d, "DB index : ", inParams.dbs[inParams.curDb])
    oid, oiderr := getIntfCountersTblKey(inParams.dbs[inParams.curDb], intfName)
    if oiderr != nil {
        log.Info(oiderr)
        return oiderr
    }
    cntTs := &db.TableSpec{Name: "COUNTERS"}
    entry, dbErr := inParams.dbs[inParams.curDb].GetEntry(cntTs, db.Key{Comp: []string{oid}})
    if dbErr != nil {
        log.Info("PopulateIntfCounters : not able find the oid entry in DB Counters table")
        return dbErr
    }
    CounterData := entry
    cntTs_cp := &db.TableSpec { Name: "COUNTERS_BACKUP" }
    entry_backup, dbErr := inParams.dbs[inParams.curDb].GetEntry(cntTs_cp, db.Key{Comp: []string{oid}})
    if dbErr != nil {
        m := make(map[string]string)
        log.Info("PopulateIntfCounters : not able find the oid entry in DB COUNTERS_BACKUP table")
        /* Frame backup data with 0 as counter values */
        for  attr,_ := range entry.Field {
            m[attr] = "0"
        }
        m["LAST_CLEAR_TIMESTAMP"] = "0"
        entry_backup = db.Value{Field: m}
    }
    CounterBackUpData := entry_backup

    switch (targetUriPath) {
    case "/openconfig-interfaces:interfaces/interface/state/counters":
        for _, attr := range portCntList {
            uri := targetUriPath + "/" + attr
            if ok, err := getSpecificCounterAttr(uri, &CounterData, &CounterBackUpData, counter); !ok || err != nil {
                log.Info("Get Counter URI failed :", uri)
                err = errors.New("Get Counter URI failed")
            }
        }
    default:
        _, err = getSpecificCounterAttr(targetUriPath, &CounterData, &CounterBackUpData, counter)
    }

    return err
}

var mgmtCounterIndexMap = map[string]int {
    "in-octets"            : 1,
    "in-pkts"              : 2,
    "in-errors"            : 3,
    "in-discards"          : 4,
    "in-multicast-pkts"    : 8,
    "out-octets"           : 9,
    "out-pkts"             : 10,
    "out-errors"           : 11,
    "out-discards"         : 12,
}

func getMgmtCounters(val string, counter_val **uint64 ) error {

    var err error
    if len(val) > 0 {
        v, e := strconv.ParseUint(val, 10, 64)
        if err == nil {
            *counter_val = &v
            return nil
        }
        err = e
    }
    return err
}
func getMgmtSpecificCounterAttr (uri string, cnt_data []string, counter *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters) (error) {

    var e error
    switch (uri) {
    case "/openconfig-interfaces:interfaces/interface/state/counters/in-octets":
        e = getMgmtCounters(cnt_data[mgmtCounterIndexMap["in-octets"]], &counter.InOctets)
        return e
    case "/openconfig-interfaces:interfaces/interface/state/counters/in-pkts":
        e = getMgmtCounters(cnt_data[mgmtCounterIndexMap["in-pkts"]], &counter.InPkts)
        return  e
    case "/openconfig-interfaces:interfaces/interface/state/counters/in-errors":
        e = getMgmtCounters(cnt_data[mgmtCounterIndexMap["in-errors"]], &counter.InErrors)
        return  e
    case "/openconfig-interfaces:interfaces/interface/state/counters/in-discards":
        e = getMgmtCounters(cnt_data[mgmtCounterIndexMap["in-discards"]], &counter.InDiscards)
        return e
    case "/openconfig-interfaces:interfaces/interface/state/counters/in-multicast-pkts":
        e = getMgmtCounters(cnt_data[mgmtCounterIndexMap["in-multicast-pkts"]], &counter.InMulticastPkts)
        return e
    case "/openconfig-interfaces:interfaces/interface/state/counters/out-octets":
        e = getMgmtCounters(cnt_data[mgmtCounterIndexMap["out-octets"]], &counter.OutOctets)
        return e
    case "/openconfig-interfaces:interfaces/interface/state/counters/out-pkts":
        e = getMgmtCounters(cnt_data[mgmtCounterIndexMap["out-pkts"]], &counter.OutPkts)
        return e
    case "/openconfig-interfaces:interfaces/interface/state/counters/out-errors":
        e = getMgmtCounters(cnt_data[mgmtCounterIndexMap["out-errors"]], &counter.OutErrors)
        return e
    case "/openconfig-interfaces:interfaces/interface/state/counters/out-discards":
        e = getMgmtCounters(cnt_data[mgmtCounterIndexMap["out-discards"]], &counter.OutDiscards)
        return e
    case "/openconfig-interfaces:interfaces/interface/state/counters":
        for key := range mgmtCounterIndexMap {
            xuri := uri + "/" + key
            e = getMgmtSpecificCounterAttr(xuri, cnt_data, counter)
        }
        return  nil
    }

    log.Info("getMgmtSpecificCounterAttr - Invalid counters URI : ", uri)
    return errors.New("Invalid counters URI")

}

var populateMGMTPortCounters PopulateIntfCounters = func (inParams XfmrParams, counter *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters) (error) {
    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)

    fileName := "/proc/net/dev"
    file, err := os.Open(fileName)
    if err != nil {
        log.Info("failed opening file: %s", err)
        return err
    }

    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)
    var txtlines []string
    for scanner.Scan() {
        txtlines = append(txtlines, scanner.Text())
    }
    file.Close()
    var entry string
    for _, eachline := range txtlines {
        ln := strings.TrimSpace(eachline)
        if strings.HasPrefix(ln, intfName) {
            entry = ln
            log.Info(" Interface stats : ", entry)
            break
        }
    }

    if entry  == "" {
        log.Info("Counters not found for Interface " + intfName)
        return errors.New("Counters not found for Interface " + intfName)
    }

    stats := strings.Fields(entry)
    log.Info(" Interface filds: ", stats)

    ret := getMgmtSpecificCounterAttr(targetUriPath, stats, counter)
    log.Info(" getMgmtCounters : ", *counter)
    return ret
}

var YangToDb_intf_counters_key KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var entry_key string
    var err error
    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")
    oid, oiderr := getIntfCountersTblKey(inParams.dbs[inParams.curDb], intfName)

    if oiderr == nil {
        entry_key = oid
    }
    return entry_key, err
}

var DbToYang_intf_counters_key KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    return rmap, err
}

var DbToYang_intf_get_counters_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error

    intfsObj := getIntfsRoot(inParams.ygRoot)
    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")

    targetUriPath, err := getYangPathFromUri(inParams.uri)
    log.Info("targetUriPath is ", targetUriPath)

    if  targetUriPath != "/openconfig-interfaces:interfaces/interface/state/counters" {
        log.Info("%s is redundant", targetUriPath)
        return err
    }

    intfType, _, ierr := getIntfTypeByName(intfName)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("DbToYang_intf_get_counters_xfmr - Invalid interface type IntfTypeUnset");
        return errors.New("Invalid interface type IntfTypeUnset");
    }
    intTbl := IntfTypeTblMap[intfType]
    if intTbl.CountersHdl.PopulateCounters == nil {
         log.Infof("Counters for Interface: %s not supported!", intfName)
		 return nil
 	}
    var state_counters * ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters

    if intfsObj != nil && intfsObj.Interface != nil && len(intfsObj.Interface) > 0 {
        var ok bool = false
        var intfObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface
        if intfObj, ok = intfsObj.Interface[intfName]; !ok {
            intfObj, _ = intfsObj.NewInterface(intfName)
            ygot.BuildEmptyTree(intfObj)
        }
        ygot.BuildEmptyTree(intfObj)
        if intfObj.State == nil  ||  intfObj.State.Counters == nil {
            ygot.BuildEmptyTree(intfObj.State)
        }
        state_counters = intfObj.State.Counters
    } else {
        ygot.BuildEmptyTree(intfsObj)
        intfObj, _:= intfsObj.NewInterface(intfName)
        ygot.BuildEmptyTree(intfObj)
        state_counters = intfObj.State.Counters
    }

    err = intTbl.CountersHdl.PopulateCounters(inParams, state_counters)
    log.Info("DbToYang_intf_get_counters_xfmr - ", state_counters)

    return err
}

/* Handle port-speed, auto-neg and aggregate-id config */
var YangToDb_intf_eth_port_config_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {

    pathInfo := NewPathInfo(inParams.uri)
    ifName := pathInfo.Var("name")
    intfsObj := getIntfsRoot(inParams.ygRoot)
    intfObj := intfsObj.Interface[ifName]
    if intfObj.Ethernet == nil  {
        return nil, errors.New("Invalid request")
    }
    if intfObj.Ethernet.Config == nil {
        return nil, errors.New("Invalid config request")
    }

    var err error

    intfType, _, ierr := getIntfTypeByName(ifName)
    if ierr != nil {
        errStr := "Invalid Interface"
        err = tlerr.InvalidArgsError{Format: errStr}
        return nil, err
    }

    memMap := make(map[string]map[string]db.Value)

    /* Handle AggregateId config */
    if intfObj.Ethernet.Config.AggregateId != nil {
        if strings.HasPrefix(ifName, ETHERNET) == false {
            return nil, errors.New("Invalid config request")
        }
        intTbl := IntfTypeTblMap[IntfTypePortChannel]
        tblName, _ := getMemTableNameByDBId(intTbl, inParams.curDb)
        var lagStr string
        switch inParams.oper {
            case CREATE:
            case UPDATE:
                log.Info("Add member port")
                lagId := intfObj.Ethernet.Config.AggregateId
                lagStr = "PortChannel" + (*lagId)
                /* Check if PortChannel exists */
                err = validateLagExists(inParams.d, &intTbl.cfgDb.portTN, &lagStr)
                if err != nil {
                    errStr := "Invalid PortChannel: " + lagStr
                    err = tlerr.InvalidArgsError{Format: errStr}
                    return nil, err
                }
                /* Check if given iface already part of a PortChannel */
                lagKeys, err := inParams.d.GetKeys(&db.TableSpec{Name:tblName})
                if err == nil {
                    for i, _ := range lagKeys {
                        if ifName == lagKeys[i].Get(1) {
                            errStr := "Given interface already part of " + lagKeys[i].Get(0)
                            err = tlerr.InvalidArgsError{Format: errStr}
                            return nil, err
                        }
                    }
                }

            case DELETE:
                log.Info("Delete member port")
                lagKeys, err := inParams.d.GetKeys(&db.TableSpec{Name:tblName})
                /* Find the port-channel the given ifname is part of */
                if err != nil {
                    log.Info("No entries in PORTCHANNEL_MEMBER TABLE")
                    return nil, errors.New("No entries in PORTCHANNEL_MEMBER TABLE")
                }
                var flag bool = false
                for i, _ := range lagKeys {
                    if ifName == lagKeys[i].Get(1) {
                        log.Info("Found Entry in PORTCHANNEL_MEMBER TABLE")
                        flag = true
                        lagStr = lagKeys[i].Get(0)
                        log.Info("Given interface part of PortChannel", lagStr)
                        break
                    }
                }
                if flag == false {
                    log.Info("Given Interface not part of any PortChannel")
                    err = errors.New("Given Interface not part of any PortChannel")
                    return nil, err
                }
        }/* End of switch case */
        m := make(map[string]string)
        value := db.Value{Field: m}
        m["NULL"] = "NULL"
        intfKey := lagStr + "|" + ifName
        if _, ok := memMap[tblName]; !ok {
            memMap[tblName] = make(map[string]db.Value)
        }
        memMap[tblName][intfKey] = value
    }
    /* Handle PortSpeed config */
    if intfObj.Ethernet.Config.PortSpeed != 0 {
        if intfType != IntfTypeMgmt {
            return nil, errors.New("PortSpeed config not supported for given interface type")
        }
        res_map := make(map[string]string)
        value := db.Value{Field: res_map}
        intTbl := IntfTypeTblMap[IntfTypeMgmt]

        portSpeed := intfObj.Ethernet.Config.PortSpeed
        val, ok := intfOCToSpeedMap[portSpeed]
        if ok {
            res_map[PORT_SPEED] = val
        } else {
            err = errors.New("Invalid/Unsupported speed.")
        }

        if _, ok := memMap[intTbl.cfgDb.portTN]; !ok {
            memMap[intTbl.cfgDb.portTN] = make(map[string]db.Value)
        }
        memMap[intTbl.cfgDb.portTN][ifName] = value

    }
    /* Handle AutoNegotiate config */
    if intfObj.Ethernet.Config.AutoNegotiate != nil {
        if intfType != IntfTypeMgmt {
            return nil, errors.New("AutoNegotiate config not supported for given Interface type")
        }
        res_map := make(map[string]string)
        value := db.Value{Field: res_map}
        intTbl := IntfTypeTblMap[IntfTypeMgmt]

        autoNeg := intfObj.Ethernet.Config.AutoNegotiate
        var enStr string
        if *autoNeg == true {
            enStr = "true"
        } else {
            enStr = "false"
        }
        res_map[PORT_AUTONEG] = enStr

        if _, ok := memMap[intTbl.cfgDb.portTN]; !ok {
            memMap[intTbl.cfgDb.portTN] = make(map[string]db.Value)
        }
        memMap[intTbl.cfgDb.portTN][ifName] = value
    }
    return memMap, err
}

/*--show ip ARP/neighbors changes start--*/
var YangToDb_neigh_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var neightbl_key string
    var err error

    log.Info("YangToDb_neigh_tbl_key_xfmr - inParams: ", inParams)
    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")
    ipAddr := pathInfo.Var("ip")

    neightbl_key = intfName + ":" +  ipAddr
    log.Info("YangToDb_neigh_tbl_key_xfmr - key returned: ", neightbl_key)

    return neightbl_key, err
}

var DbToYang_neigh_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    log.Info("DbToYang_neigh_tbl_key_xfmr - inParams: ", inParams)
    mykey := strings.Split(inParams.key,":")

    rmap["ip"] =  inParams.key[(len(mykey[0])+1):]
    return rmap, err
}


var DbToYang_neigh_tbl_get_all_ipv4_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) (error) {
    var err error
    var ok bool

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_neigh_tbl_get_all_ipv4_xfmr - data:", data)
    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)
    log.Info("DbToYang_neigh_tbl_get_all_ipv4_xfmr - targetUriPath: ", targetUriPath)

    var intfObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface
    var subIntfObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface
    var neighObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Neighbors_Neighbor

    intfsObj := getIntfsRoot(inParams.ygRoot)

    intfNameRcvd := pathInfo.Var("name")
    ipAddrRcvd := pathInfo.Var("ip")

    if intfObj, ok = intfsObj.Interface[intfNameRcvd]; !ok {
        intfObj, err = intfsObj.NewInterface(intfNameRcvd)
        if err != nil {
            log.Error("Creation of interface subtree failed!")
            return err
        }
    }
    ygot.BuildEmptyTree(intfObj)

    if subIntfObj, ok = intfObj.Subinterfaces.Subinterface[0]; !ok {
        subIntfObj, err = intfObj.Subinterfaces.NewSubinterface(0)
        if err != nil {
            log.Error("Creation of subinterface subtree failed!")
            return err
        }
    }
    ygot.BuildEmptyTree(subIntfObj)

    for key, entry := range data["NEIGH_TABLE"] {
        var ipAddr string

        /*separate ip and interface*/
        tokens := strings.Split(key, ":")
        intfName := tokens[0]
        ipAddr = key[len(intfName)+1:]

        linkAddr := data["NEIGH_TABLE"][key].Field["neigh"]
        if (linkAddr == "") {
            log.Info("No mac-address found for IP: ", ipAddr)
            continue;
        }

        addrFamily := data["NEIGH_TABLE"][key].Field["family"]
        if (addrFamily == "") {
            log.Info("No address family found for IP: ", ipAddr)
            continue;
        }

        /*The transformer returns complete table regardless of the interface.
          First check if the interface and IP of this redis entry matches one
          available in the received URI
        */
        if (strings.Contains(targetUriPath, "ipv4") && addrFamily != "IPv4") ||
            intfName != intfNameRcvd ||
            (ipAddrRcvd != "" && ipAddrRcvd != ipAddr) {
                log.Info("Skipping entry: ", entry, "for interface: ", intfName, " and IP:", ipAddr,
                         "interface received: ", intfNameRcvd, " IP received: ", ipAddrRcvd)
                continue
        } else if strings.HasPrefix(targetUriPath, NEIGH_IPv4_PREFIX_STATE_LL) {
            if neighObj, ok = subIntfObj.Ipv4.Neighbors.Neighbor[ipAddr]; !ok {
                neighObj, err = subIntfObj.Ipv4.Neighbors.NewNeighbor(ipAddr)
                if err != nil {
                    log.Error("Creation of neighbor subtree failed!")
                    return err
                }
            }
            ygot.BuildEmptyTree(neighObj)
            neighObj.State.LinkLayerAddress = &linkAddr
            break
        } else if strings.HasPrefix(targetUriPath, NEIGH_IPv4_PREFIX_STATE_IP) {
            if neighObj, ok = subIntfObj.Ipv4.Neighbors.Neighbor[ipAddr]; !ok {
                neighObj, err = subIntfObj.Ipv4.Neighbors.NewNeighbor(ipAddr)
                if err != nil {
                    log.Error("Creation of neighbor subtree failed!")
                    return err
                }
            }
            ygot.BuildEmptyTree(neighObj)
            neighObj.State.Ip = &ipAddr
            break
        } else if strings.HasPrefix(targetUriPath, NEIGH_IPv4_PREFIX_IP) {
            if neighObj, ok = subIntfObj.Ipv4.Neighbors.Neighbor[ipAddr]; !ok {
                neighObj, err = subIntfObj.Ipv4.Neighbors.NewNeighbor(ipAddr)
                if err != nil {
                    log.Error("Creation of neighbor subtree failed!")
                    return err
                }
            }
            ygot.BuildEmptyTree(neighObj)
            neighObj.State.Ip = &ipAddr
            neighObj.State.LinkLayerAddress = &linkAddr
            neighObj.State.Origin = 0
            break
        } else if strings.HasPrefix(targetUriPath, NEIGH_IPv4_PREFIX) {
            if neighObj, ok = subIntfObj.Ipv4.Neighbors.Neighbor[ipAddr]; !ok {
                neighObj, err = subIntfObj.Ipv4.Neighbors.NewNeighbor(ipAddr)
                if err != nil {
                    log.Error("Creation of neighbor subtree failed!")
                    return err
                }
            }
            ygot.BuildEmptyTree(neighObj)
            neighObj.State.Ip = &ipAddr
            neighObj.State.LinkLayerAddress = &linkAddr
            neighObj.State.Origin = 0
        }
    }
    return err
}

var DbToYang_neigh_tbl_get_all_ipv6_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) (error) {
    var err error
    var ok bool

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_neigh_tbl_get_all_ipv6_xfmr - data: ", data)
    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)
    log.Info("DbToYang_neigh_tbl_get_all_ipv6_xfmr - targetUriPath: ", targetUriPath)

    var intfObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface
    var subIntfObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface
    var neighObj *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Neighbors_Neighbor

    intfsObj := getIntfsRoot(inParams.ygRoot)

    intfNameRcvd := pathInfo.Var("name")
    ipAddrRcvd := pathInfo.Var("ip")

    if intfObj, ok = intfsObj.Interface[intfNameRcvd]; !ok {
        intfObj, err = intfsObj.NewInterface(intfNameRcvd)
        if err != nil {
            log.Error("Creation of interface subtree failed!")
            return err
        }
    }
    ygot.BuildEmptyTree(intfObj)

    if subIntfObj, ok = intfObj.Subinterfaces.Subinterface[0]; !ok {
        subIntfObj, err = intfObj.Subinterfaces.NewSubinterface(0)
        if err != nil {
            log.Error("Creation of subinterface subtree failed!")
            return err
        }
    }
    ygot.BuildEmptyTree(subIntfObj)

    for key, entry := range data["NEIGH_TABLE"] {
        var ipAddr string

        /*separate ip and interface*/
        tokens := strings.Split(key, ":")
        intfName := tokens[0]
        ipAddr = key[len(intfName)+1:]

        linkAddr := data["NEIGH_TABLE"][key].Field["neigh"]
        if (linkAddr == "") {
            log.Info("No mac-address found for IP: ", ipAddr)
            continue;
        }

        addrFamily := data["NEIGH_TABLE"][key].Field["family"]
        if (addrFamily == "") {
            log.Info("No address family found for IP: ", ipAddr)
            continue;
        }

        if (strings.Contains(targetUriPath, "ipv6") && addrFamily != "IPv6") ||
            intfName != intfNameRcvd ||
            (ipAddrRcvd != "" && ipAddrRcvd != ipAddr) {
                log.Info("Skipping entry: ", entry, "for interface: ", intfName, " and IP:", ipAddr,
                         "interface received: ", intfNameRcvd, " IP received: ", ipAddrRcvd)
                continue
        }else if strings.HasPrefix(targetUriPath, NEIGH_IPv6_PREFIX_STATE_LL) {
            if neighObj, ok = subIntfObj.Ipv6.Neighbors.Neighbor[ipAddr]; !ok {
                neighObj, err = subIntfObj.Ipv6.Neighbors.NewNeighbor(ipAddr)
                if err != nil {
                    log.Error("Creation of neighbor subtree failed!")
                    return err
                }
            }
            ygot.BuildEmptyTree(neighObj)
            neighObj.State.LinkLayerAddress = &linkAddr
            break
        } else if strings.HasPrefix(targetUriPath, NEIGH_IPv6_PREFIX_STATE_IP) {
            if neighObj, ok = subIntfObj.Ipv6.Neighbors.Neighbor[ipAddr]; !ok {
                neighObj, err = subIntfObj.Ipv6.Neighbors.NewNeighbor(ipAddr)
                if err != nil {
                    log.Error("Creation of neighbor subtree failed!")
                    return err
                }
            }
            ygot.BuildEmptyTree(neighObj)
            neighObj.State.Ip = &ipAddr
            break
        } else if strings.HasPrefix(targetUriPath, NEIGH_IPv6_PREFIX_IP) {
            if neighObj, ok = subIntfObj.Ipv6.Neighbors.Neighbor[ipAddr]; !ok {
                neighObj, err = subIntfObj.Ipv6.Neighbors.NewNeighbor(ipAddr)
                if err != nil {
                    log.Error("Creation of neighbor subtree failed!")
                    return err
                }
            }
            ygot.BuildEmptyTree(neighObj)
            neighObj.State.Ip = &ipAddr
            neighObj.State.LinkLayerAddress = &linkAddr
            neighObj.State.IsRouter = true
            neighObj.State.NeighborState = 0
            neighObj.State.Origin = 0
            break
        } else if strings.HasPrefix(targetUriPath, NEIGH_IPv6_PREFIX) {
            if neighObj, ok = subIntfObj.Ipv6.Neighbors.Neighbor[ipAddr]; !ok {
                neighObj, err = subIntfObj.Ipv6.Neighbors.NewNeighbor(ipAddr)
                if err != nil {
                    log.Error("Creation of neighbor subtree failed!")
                    return err
                }
            }
            ygot.BuildEmptyTree(neighObj)
            neighObj.State.Ip = &ipAddr
            neighObj.State.LinkLayerAddress = &linkAddr
            neighObj.State.IsRouter = true
            neighObj.State.NeighborState = 0
            neighObj.State.Origin = 0
        }
    }
    return err
}
/*--show ip ARP/neighbors changes end--*/
