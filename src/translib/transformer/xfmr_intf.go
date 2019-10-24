package transformer

import (
    //"fmt"
    //"bytes"
    "errors"
    "strings"
    "strconv"
    "reflect"
    "regexp"
    "net"
    //"translib/tlerr"
    "github.com/openconfig/ygot/ygot"
    "translib/db"
    log "github.com/golang/glog"
    "translib/ocbinds"
    "bufio"
    "os"
    //"github.com/openconfig/ygot/ytypes"
)


func init () {
    XlateFuncBind("intf_table_xfmr", intf_table_xfmr)
    XlateFuncBind("YangToDb_intf_name_xfmr", YangToDb_intf_name_xfmr)
    XlateFuncBind("DbToYang_intf_name_xfmr", DbToYang_intf_name_xfmr)
    XlateFuncBind("YangToDb_intf_enabled_xfmr", YangToDb_intf_enabled_xfmr)
    XlateFuncBind("DbToYang_intf_enabled_xfmr", DbToYang_intf_enabled_xfmr)
    XlateFuncBind("DbToYang_intf_admin_status_xfmr", DbToYang_intf_admin_status_xfmr)
    XlateFuncBind("DbToYang_intf_oper_status_xfmr", DbToYang_intf_oper_status_xfmr)
    XlateFuncBind("YangToDb_intf_eth_auto_neg_xfmr", YangToDb_intf_eth_auto_neg_xfmr)
    XlateFuncBind("DbToYang_intf_eth_auto_neg_xfmr", DbToYang_intf_eth_auto_neg_xfmr)
    XlateFuncBind("YangToDb_intf_eth_port_speed_xfmr", YangToDb_intf_eth_port_speed_xfmr)
    XlateFuncBind("DbToYang_intf_eth_port_speed_xfmr", DbToYang_intf_eth_port_speed_xfmr)
    XlateFuncBind("YangToDb_intf_tbl_ip_key", YangToDb_intf_tbl_ip_key)
    XlateFuncBind("DbToYang_intf_tbl_ip_key", DbToYang_intf_tbl_ip_key)
    XlateFuncBind("DbToYang_intf_get_counters_xfmr", DbToYang_intf_get_counters_xfmr)
}

const (
    PORT_INDEX         = "index"
    PORT_MTU           = "mtu"
    PORT_ADMIN_STATUS  = "admin_status"
    PORT_SPEED         = "speed"
    PORT_DESC          = "description"
    PORT_OPER_STATUS   = "oper_status"
    PORT_AUTONEG       = "autoneg"
)

const (
    PIPE                     =  "|"
    COLON                    =  ":"

    ETHERNET                 = "Ethernet"
    MGMT                     = "eth"
    VLAN                     = "Vlan"
    PORTCHANNEL              = "PortChannel"
)

type TblData  struct  {
    portTN           string
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
        cfgDb:TblData{portTN:"PORTCHANNEL", intfTN:"PORTCHANNEL_INTERFACE", keySep: PIPE},
        appDb:TblData{portTN:"LAG_TABLE", intfTN:"INTF_TABLE", keySep: COLON},
        stateDb:TblData{portTN:"LAG_TABLE", intfTN:"INTERFACE_TABLE", keySep: PIPE},
        CountersHdl:CounterData{OIDTN: "COUNTERS_PORT_NAME_MAP", CountersTN:"COUNTERS", PopulateCounters: populatePortCounters},
    },
}
var dbIdToTblMap = map[db.DBNum][]string {
    db.ConfigDB: {"PORT", "INTERFACE", "MGMT_PORT", "MGMT_INTERFACE", , "PORTCHANNEL", "PORTCHANNEL_INTERFACE"},
    db.ApplDB  : {"PORT_TABLE", "INTF_TABLE", "MGMT_PORT_TABLE", "MGMT_INTF_TABLE", "LAG_TABLE"},
    db.StateDB : {"PORT_TABLE", "INTERFACE_TABLE", "MGMT_PORT_TABLE", "MGMT_INTERFACE_TABLE", "LAG_TABLE"},
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
    IntfTypePortChannel        E_InterfaceType = 4

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
    }
    err = errors.New("Interface name prefix not matched with supported types")
    return IntfTypeUnset, IntfSubTypeUnset, err
}

func getIntfsRoot (s *ygot.GoStruct) *ocbinds.OpenconfigInterfaces_Interfaces {
    deviceObj := (*s).(*ocbinds.Device)
    return deviceObj.Interfaces
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
    /*
    switch (targetUriPath) {
    case "/openconfig-interfaces:interfaces/interface/config",
        "/openconfig-interfaces:interfaces/interface/ethernet/config",
        "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/config":
        tblList = append(tblList, intTbl.cfgDb.portTN)
    case "/openconfig-interfaces:interfaces/interface/state/counters":
        tblList = append(tblList, intTbl.CountersHdl.CountersTN)
    case "/openconfig-interfaces:interfaces/interface/state",
        "/openconfig-interfaces:interfaces/interface/ethernet/state",
        "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/state":
        tblList = append(tblList, intTbl.appDb.portTN)
    case "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses/address/config",
        "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv4/addresses/address/config",
        "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses/address/config",
        "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses/address/config":
        tblList = append(tblList, intTbl.cfgDb.intfTN)
    case "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses/address/state",
        "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv4/addresses/address/state",
        "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses/address/state",
        "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses/address/state":
        tblList = append(tblList, intTbl.appDb.intfTN)
    case "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface":
        tblList = append(tblList, intTbl.cfgDb.intfTN)
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet",
        "/openconfig-interfaces:interfaces/interface/ethernet":
        tblList = append(tblList, intTbl.cfgDb.portTN)
    case "/openconfig-interfaces:interfaces/interface":
        tblList = append(tblList, intTbl.cfgDb.portTN)
    default:
        err = errors.New("Invalid URI")
    }*/
    if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/config") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/ethernet/config") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/config") {
        tblList = append(tblList, intTbl.cfgDb.portTN)
    } else if  strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/state/counters") {
        tblList = append(tblList, intTbl.CountersHdl.CountersTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/state") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/ethernet/state") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/state") {
        tblList = append(tblList, intTbl.appDb.portTN)
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
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface") {
        tblList = append(tblList, intTbl.cfgDb.intfTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/ethernet") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet") {
        tblList = append(tblList, intTbl.cfgDb.portTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface") {
        tblList = append(tblList, intTbl.cfgDb.portTN)
    } else {
        err = errors.New("Invalid URI")
    }


    /*
    if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/config") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/ethernet/config") {
        tblList = append(tblList, intTbl.cfgDb.portTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/state/counters") {
        tblList = append(tblList, intTbl.CountersHdl.CountersTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/state") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/ethernet/state") {
        tblList = append(tblList, intTbl.appDb.portTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses/address/config") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses/address/config") {
        tblList = append(tblList, intTbl.cfgDb.intfTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv4/addresses/address/state") ||
        strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/ipv6/addresses/address/state") {
        tblList = append(tblList, intTbl.appDb.intfTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/subinterfaces/subinterface") {
        tblList = append(tblList, intTbl.cfgDb.intfTN)
    } else if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/ethernet") {
        tblList = append(tblList, intTbl.cfgDb.portTN)
    } else if pathInfo.Template ==  "/openconfig-interfaces:interfaces/interface{}" {
        tblList = append(tblList, intTbl.cfgDb.portTN)
    } else  {
        err = errors.New("Invalid URI")
    }
    */
    log.Infof("TableXfmrFunc - uri(%v), tblList(%v)\r\n", inParams.uri, tblList);
    return tblList, err
}

var YangToDb_intf_name_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    log.Info("YangToDb_intf_name_xfmr: ")
    /*no-op since there is no redis table field to be filled corresponding to name attribute since its part of key */
    return res_map, err
}

var DbToYang_intf_name_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    res_map := make(map[string]interface{})
    log.Info("DbToYang_intf_name_xfmr: ", inParams.key)
    res_map["name"] =  inParams.key
    log.Info("DbToYang_intf_name_xfmr: interface/config|state/name ", res_map)
    return res_map, err
}

var YangToDb_intf_enabled_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_intf_enabled_xfmr Entry - ", reflect.ValueOf(inParams.param), "Type of : ", reflect.TypeOf(inParams.param));
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
    log.Info("DbToYang_intf_enabled_xfmr ", data, "inParams : ", inParams)
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
    log.Info("DbToYang_intf_admin_status_xfmr ", data, "inParams : ", inParams)
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
        //status = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_UNSET
        log.Info("Admin status field not found in DB")
    }

    //result["admin-status"] = ocbinds.E_OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus.ΛMap(status)["E_OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus"][int64(status)].Name
    return result, err
}

var DbToYang_intf_oper_status_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_intf_oper_status_xfmr ", data, inParams.ygRoot)
    intfType, _, ierr := getIntfTypeByName(inParams.key)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("DbToYang_intf_oper_status_xfmr - Invalid interface type IntfTypeUnset");
        return result, errors.New("Invalid interface type IntfTypeUnset");
    }
    intTbl := IntfTypeTblMap[intfType]

    tblName, _ := getPortTableNameByDBId(intTbl, inParams.curDb)
    pTbl := data[tblName]
    prtInst := pTbl[inParams.key]
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
        //status = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_UNSET
        log.Info("Admin status field not found in DB")
    }

    //result["oper-status"] = ocbinds.E_OpenconfigInterfaces_Interfaces_Interface_State_OperStatus.ΛMap(status)["E_OpenconfigInterfaces_Interfaces_Interface_State_OperStatus"][int64(status)].Name
    return result, err
}


var YangToDb_intf_eth_auto_neg_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_intf_eth_auto_neg_xfmr Entry");
    autoNeg, _ := inParams.param.(*bool)
    var enStr string
    if *autoNeg == true {
        enStr = "true"
    } else {
        enStr = "false"
    }
    res_map[PORT_AUTONEG] = enStr

    return res_map, nil
}

var DbToYang_intf_eth_auto_neg_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_intf_eth_auto_neg_xfmr ", data, inParams.ygRoot)
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


var YangToDb_intf_eth_port_speed_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error

    log.Info("YangToDb_intf_eth_port_speed_xfmr Entry");
    portSpeed, _ := inParams.param.(ocbinds.E_OpenconfigIfEthernet_ETHERNET_SPEED)
    val, ok := intfOCToSpeedMap[portSpeed]
    if ok {
        res_map[PORT_SPEED] = val
    } else {
        err = errors.New("Invalid/Unsupported speed.")
    }

    return res_map, err 
}

var DbToYang_intf_eth_port_speed_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_intf_eth_port_speed_xfmr ", data, inParams.ygRoot)
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

    //result["port-speed"] = ocbinds.E_OpenconfigIfEthernet_ETHERNET_SPEED.ΛMap(portSpeed)["E_OpenconfigIfEthernet_ETHERNET_SPEED"][int64(portSpeed)].Name
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

var YangToDb_intf_tbl_ip_key KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var entry_key string
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")

    var ip string
    var prefix int
    objType := reflect.TypeOf(inParams.param).Elem()

    log.Info("YangToDb_intf_tbl_ip_key ", objType)

    switch (objType) {
    case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address_Config{}):
        obj := inParams.param.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address_Config)
        if obj != nil {
            ip  = *obj.Ip
            prefix = int(*obj.PrefixLength)
            log.Info("YangToDb_intf_tbl_ip_key ", intfName, ip, prefix)
        } else {
            return entry_key, errors.New("config obj is not set")
        }
    case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address_State{}):
        obj := inParams.param.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address_State)
        if obj != nil {
            ip  = *obj.Ip
            prefix = int(*obj.PrefixLength)
            log.Info("YangToDb_intf_tbl_ip_key ", intfName, ip, prefix)
        } else {
            return entry_key, errors.New("config obj is not set")
        }
    case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address_Config{}):
        obj := inParams.param.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address_Config)
        if obj != nil {
            ip  = *obj.Ip
            prefix = int(*obj.PrefixLength)
            log.Info("YangToDb_intf_tbl_ip_key ", intfName, ip, prefix)
        } else {
            return entry_key, errors.New("config obj is not set")
        }
    case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address_State{}):
        obj := inParams.param.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address_State)
        if obj != nil {
            ip  = *obj.Ip
            prefix = int(*obj.PrefixLength)
            log.Info("YangToDb_intf_tbl_ip_key ", intfName, ip, prefix)
        } else {
            return entry_key, errors.New("config obj is not set")
        }
    }
    entry_key = intfName + "|" +  ip + "/" + strconv.Itoa(prefix)
    return entry_key, err
}

var DbToYang_intf_tbl_ip_key KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    entry_key := inParams.key
    log.Info("DbToYang_intf_tbl_ip_key: ", entry_key)

    key := strings.Split(entry_key, "|")
    if len(key) < 2 {
        err = errors.New("Invalid key for INTERFACE table entry.")
        log.Info("Invalid Keys for INTERFACE table entry : ", entry_key)
        return rmap, err
    }

    intfName := key[0]
    log.Info("DbToYang_intf_tbl_ip_key :- ", intfName)
    ipKey :=strings.Split(key[1], "/")
    if len(ipKey) < 2 {
        err = errors.New("Invalid key, prefix not specified.")
        log.Info("Invalid key, prefix not specified - ", entry_key)
        return rmap, err
    }

    rmap["ip"] = ipKey[0]
    rmap["prefix-length"] = ipKey[1]
    return rmap, err
}

var validate_intf_ip_address = func(inParams XfmrParams) (bool) {
    var ret bool = true
    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")

    var ip string
    var prefix int
    var isIpV4 bool = false
    objType := reflect.TypeOf(inParams.param)

    log.Info("YangToDb_intf_tbl_ip_key ", objType.Name())

    switch (objType.Name()) {
    case "*OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address_Config":
        obj := inParams.param.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address_Config)
        if obj != nil {
            ip  = *obj.Ip
            prefix = int(*obj.PrefixLength)
            log.Info("YangToDb_intf_tbl_ip_key ", intfName, ip, prefix)
            isIpV4 = true
        }
    case "*OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address_State":
        obj := inParams.param.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address_State)
        if obj != nil {
            ip  = *obj.Ip
            prefix = int(*obj.PrefixLength)
            log.Info("YangToDb_intf_tbl_ip_key ", intfName, ip, prefix)
            isIpV4 = true
        }
    case "*OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address_Config":
        obj := inParams.param.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address_Config)
        if obj != nil {
            ip  = *obj.Ip
            prefix = int(*obj.PrefixLength)
            log.Info("YangToDb_intf_tbl_ip_key ", intfName, ip, prefix)
            isIpV4 = false
        }
    case "*OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address_State":
        obj := inParams.param.(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address_State)
        if obj != nil {
            ip  = *obj.Ip
            prefix = int(*obj.PrefixLength)
            log.Info("YangToDb_intf_tbl_ip_key ", intfName, ip, prefix)
            isIpV4 = false
        }
    }

    intfType, _, ierr := getIntfTypeByName(intfName)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("DbToYang_intf_enabled_xfmr - Invalid interface type IntfTypeUnset");
        return false
    }
    intTbl := IntfTypeTblMap[intfType]

    tblName, _ := getIntfTableNameByDBId(intTbl, inParams.curDb)
    if isIpV4 == true {
        if !validIPv4(ip) {
            log.Info("YangToDb_intf_tbl_ip_key validIPv4 failed")
            return false
        }

    } else {
        if !validIPv6(ip) {
            log.Info("YangToDb_intf_tbl_ip_key validIPv6 failed")
            return false
        }
    }
    keys,_ := doGetAllIpKeys(inParams.d,  &db.TableSpec{Name: tblName})
    err := translateIpv4(inParams.d, keys, intfName, ip, prefix)
    if err != nil { 
        log.Info(err)
        ret = false
    }
    return ret
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


func translateIpv4(d *db.DB, allIpKeys []db.Key, intf string, ip string, prefix int) error {
    var err error
    var ifsKey db.Key
    ifsKey.Comp = []string{intf}

    ipPref := ip + "/" + strconv.Itoa(prefix)
    ifsKey.Comp = []string{intf, ipPref}

    log.Info("ifsKey:=", ifsKey)

    log.Info("Checking for IP overlap ....")
    ipA, ipNetA, _ := net.ParseCIDR(ipPref)

    for _, key := range allIpKeys {
        if len(key.Comp) < 2 {
            continue
        }
        ipB, ipNetB, _ := net.ParseCIDR(key.Get(1))

        if ipNetA.Contains(ipB) || ipNetB.Contains(ipA) {
            log.Info("IP ", ipPref, "overlaps with ", key.Get(1), " of ", key.Get(0))

            if intf != key.Get(0) {
                //IP overlap across different interface, reject
                log.Error("IP ", ipPref, " overlaps with ", key.Get(1), " of ", key.Get(0))

                errStr := "IP " + ipPref + " overlaps with IP " + key.Get(1) + " of Interface " + key.Get(0)
                err =  errors.New(errStr)
                return err
            }
        }
    }
    return err
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

func getSpecificCounterAttr(targetUriPath string, entry *db.Value, counter_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters) (bool, error) {

    var e error

    switch targetUriPath {
    case "/openconfig-interfaces:interfaces/interface/state/counters/in-octets":
        e = getCounters(entry, "SAI_PORT_STAT_IF_IN_OCTETS", &counter_val.InOctets)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-unicast-pkts":
        e = getCounters(entry, "SAI_PORT_STAT_IF_IN_UCAST_PKTS", &counter_val.InUnicastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-broadcast-pkts":
        e = getCounters(entry, "SAI_PORT_STAT_IF_IN_BROADCAST_PKTS", &counter_val.InBroadcastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-multicast-pkts":
        e = getCounters(entry, "SAI_PORT_STAT_IF_IN_MULTICAST_PKTS", &counter_val.InMulticastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-errors":
        e = getCounters(entry, "SAI_PORT_STAT_IF_IN_ERRORS", &counter_val.InErrors)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-discards":
        e = getCounters(entry, "SAI_PORT_STAT_IF_IN_DISCARDS", &counter_val.InDiscards)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/in-pkts":
        var inNonUCastPkt, inUCastPkt *uint64
        var in_pkts uint64

        e = getCounters(entry, "SAI_PORT_STAT_IF_IN_NON_UCAST_PKTS", &inNonUCastPkt)
        if e == nil {
            e = getCounters(entry, "SAI_PORT_STAT_IF_IN_UCAST_PKTS", &inUCastPkt)
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
        e = getCounters(entry, "SAI_PORT_STAT_IF_OUT_OCTETS", &counter_val.OutOctets)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-unicast-pkts":
        e = getCounters(entry, "SAI_PORT_STAT_IF_OUT_UCAST_PKTS", &counter_val.OutUnicastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-broadcast-pkts":
        e = getCounters(entry, "SAI_PORT_STAT_IF_OUT_BROADCAST_PKTS", &counter_val.OutBroadcastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-multicast-pkts":
        e = getCounters(entry, "SAI_PORT_STAT_IF_OUT_MULTICAST_PKTS", &counter_val.OutMulticastPkts)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-errors":
        e = getCounters(entry, "SAI_PORT_STAT_IF_OUT_ERRORS", &counter_val.OutErrors)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-discards":
        e = getCounters(entry, "SAI_PORT_STAT_IF_OUT_DISCARDS", &counter_val.OutDiscards)
        return true, e

    case "/openconfig-interfaces:interfaces/interface/state/counters/out-pkts":
        var outNonUCastPkt, outUCastPkt *uint64
        var out_pkts uint64

        e = getCounters(entry, "SAI_PORT_STAT_IF_OUT_NON_UCAST_PKTS", &outNonUCastPkt)
        if e == nil {
            e = getCounters(entry, "SAI_PORT_STAT_IF_OUT_UCAST_PKTS", &outUCastPkt)
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

func getCounters(entry *db.Value, attr string, counter_val **uint64 ) error {

    var ok bool = false
    var val string
    var err error

    val, ok = entry.Field[attr]
    if !ok {
        return errors.New("Attr " + attr + "doesn't exist in IF table Map!")
    }

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

var portCntList [] string = []string {"in-octets", "in-unicast-pkts", "in-broadcast-pkts", "in-multicast-pkts",
"in-errors", "in-discards", "in-pkts", "out-octets", "out-unicast-pkts",
"out-broadcast-pkts", "out-multicast-pkts", "out-errors", "out-discards",
"out-pkts"}
var populatePortCounters PopulateIntfCounters = func (inParams XfmrParams, counter *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters) (error) {

    data := (*inParams.dbDataMap)[inParams.curDb]
    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)

    counterTbl := data["COUNTERS"]
    log.Info("PopulateIntfCounters : inParams.curDb : ", inParams.curDb, "D: ", inParams.d, "DB index : ", inParams.dbs[inParams.curDb])
    //oid, oiderr := getIntfCountersTblKey(inParams.d, intfName)
    oid, oiderr := getIntfCountersTblKey(inParams.dbs[inParams.curDb], intfName)
    if oiderr != nil {
        log.Info(oiderr)
        return oiderr
    }

    CounterData, ok := counterTbl[oid]
    if !ok {
        log.Info("oid not found in counters table ", oid)
        return errors.New("oid not found in counters table")
    }

    switch (targetUriPath) {
    case "/openconfig-interfaces:interfaces/interface/state/counters":
        for _, attr := range portCntList {
            uri := targetUriPath + "/" + attr
            if ok, err := getSpecificCounterAttr(uri, &CounterData, counter); !ok || err != nil {
                log.Info("Get Counter URI failed :", uri)
                err = errors.New("Get Counter URI failed")
            }
        }
    default:
        _, err = getSpecificCounterAttr(targetUriPath, &CounterData, counter)
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

var DbToYang_intf_get_counters_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error

    intfsObj := getIntfsRoot(inParams.ygRoot)
    pathInfo := NewPathInfo(inParams.uri)
    intfName := pathInfo.Var("name")

    intfType, _, ierr := getIntfTypeByName(intfName)
    if intfType == IntfTypeUnset || ierr != nil {
        log.Info("DbToYang_intf_get_counters_xfmr - Invalid interface type IntfTypeUnset");
        return errors.New("Invalid interface type IntfTypeUnset");
    }
    intTbl := IntfTypeTblMap[intfType]

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

    //state_counters := &ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters{}
    err = intTbl.CountersHdl.PopulateCounters(inParams, state_counters)
    log.Info("DbToYang_intf_get_counters_xfmr - ", state_counters)

    return err
}
