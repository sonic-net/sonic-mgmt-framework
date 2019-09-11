package transformer

import (
    "fmt"
    "bytes"
    "errors"
    "strings"
    "github.com/openconfig/ygot/ygot"
    "strconv"
    "translib/db"
    "reflect"
    log "github.com/golang/glog"
    "translib/ocbinds"
    "translib/tlerr"
)

func init () {
    XlateFuncBind("YangToDb_acl_set_name_xfmr", YangToDb_acl_set_name_xfmr)
    XlateFuncBind("DbToYang_acl_set_name_xfmr", DbToYang_acl_set_name_xfmr)
    XlateFuncBind("YangToDb_acl_entry_key_xfmr", YangToDb_acl_entry_key_xfmr)
    XlateFuncBind("DbToYang_acl_entry_key_xfmr", DbToYang_acl_entry_key_xfmr)
    XlateFuncBind("YangToDb_acl_entry_sequenceid_xfmr", YangToDb_acl_entry_sequenceid_xfmr)
    XlateFuncBind("DbToYang_acl_entry_sequenceid_xfmr", DbToYang_acl_entry_sequenceid_xfmr)
    XlateFuncBind("YangToDb_acl_l2_ethertype_xfmr", YangToDb_acl_l2_ethertype_xfmr)
    XlateFuncBind("DbToYang_acl_l2_ethertype_xfmr", DbToYang_acl_l2_ethertype_xfmr)
    XlateFuncBind("YangToDb_acl_ip_protocol_xfmr", YangToDb_acl_ip_protocol_xfmr)
    XlateFuncBind("DbToYang_acl_ip_protocol_xfmr", DbToYang_acl_ip_protocol_xfmr)
    XlateFuncBind("YangToDb_acl_source_port_xfmr", YangToDb_acl_source_port_xfmr)
    XlateFuncBind("DbToYang_acl_source_port_xfmr", DbToYang_acl_source_port_xfmr)
    XlateFuncBind("YangToDb_acl_destination_port_xfmr", YangToDb_acl_destination_port_xfmr)
    XlateFuncBind("DbToYang_acl_destination_port_xfmr", DbToYang_acl_destination_port_xfmr)
    XlateFuncBind("YangToDb_acl_tcp_flags_xfmr", YangToDb_acl_tcp_flags_xfmr)
    XlateFuncBind("DbToYang_acl_tcp_flags_xfmr", DbToYang_acl_tcp_flags_xfmr)
    XlateFuncBind("YangToDb_acl_port_bindings_xfmr", YangToDb_acl_port_bindings_xfmr)
    XlateFuncBind("DbToYang_acl_port_bindings_xfmr", DbToYang_acl_port_bindings_xfmr)
}

const (
    ACL_TABLE                = "ACL_TABLE"
    RULE_TABLE               = "ACL_RULE"
    SONIC_ACL_TYPE_IPV4      = "L3"
    SONIC_ACL_TYPE_L2        = "L2"
    SONIC_ACL_TYPE_IPV6      = "L3V6"
    OPENCONFIG_ACL_TYPE_IPV4 = "ACL_IPV4"
    OPENCONFIG_ACL_TYPE_IPV6 = "ACL_IPV6"
    OPENCONFIG_ACL_TYPE_L2   = "ACL_L2"
    ACL_TYPE                 = "type"

     MIN_PRIORITY = 1
     MAX_PRIORITY = 65535
)

var IP_PROTOCOL_MAP = map[ocbinds.E_OpenconfigPacketMatchTypes_IP_PROTOCOL]uint8{
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_ICMP: 1,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_IGMP: 2,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_TCP:  6,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_UDP:  17,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_RSVP: 46,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_GRE:  47,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_AUTH: 51,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_PIM:  103,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_L2TP: 115,
}

var ETHERTYPE_MAP = map[ocbinds.E_OpenconfigPacketMatchTypes_ETHERTYPE]uint32{
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_LLDP: 0x88CC,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_VLAN: 0x8100,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_ROCE: 0x8915,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_ARP:  0x0806,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_IPV4: 0x0800,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_IPV6: 0x86DD,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_MPLS: 0x8847,
}


func getAclRoot (s *ygot.GoStruct) *ocbinds.OpenconfigAcl_Acl {
    deviceObj := (*s).(*ocbinds.Device)
    return deviceObj.Acl
}


func getAclTypeOCEnumFromName(val string) (ocbinds.E_OpenconfigAcl_ACL_TYPE, error) {
    switch val {
    case "ACL_IPV4", "openconfig-acl:ACL_IPV4":
        return ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4, nil
    case "ACL_IPV6", "openconfig-acl:ACL_IPV6":
        return ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6, nil
    case "ACL_L2", "openconfig-acl:ACL_L2":
        return ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2, nil
    default:
        return ocbinds.OpenconfigAcl_ACL_TYPE_UNSET,
        tlerr.NotSupported("ACL Type '%s' not supported", val)
    }
}
func getAclKeyStrFromOCKey(aclname string, acltype ocbinds.E_OpenconfigAcl_ACL_TYPE) string {
    aclN := strings.Replace(strings.Replace(aclname, " ", "_", -1), "-", "_", -1)
    aclT := acltype.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(acltype)].Name
    return aclN + "_" + aclT
}

func getOCAclKeysFromStrDBKey(aclKey string) (string, ocbinds.E_OpenconfigAcl_ACL_TYPE) {
    var aclOrigName string
    var aclOrigType ocbinds.E_OpenconfigAcl_ACL_TYPE

    if strings.Contains(aclKey, "_"+OPENCONFIG_ACL_TYPE_IPV4) {
        aclOrigName = strings.Replace(aclKey, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
        aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4
    } else if strings.Contains(aclKey, "_"+OPENCONFIG_ACL_TYPE_IPV6) {
        aclOrigName = strings.Replace(aclKey, "_"+OPENCONFIG_ACL_TYPE_IPV6, "", 1)
        aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6
    } else if strings.Contains(aclKey, "_"+OPENCONFIG_ACL_TYPE_L2) {
        aclOrigName = strings.Replace(aclKey, "_"+OPENCONFIG_ACL_TYPE_L2, "", 1)
        aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2
    }

    return aclOrigName, aclOrigType
}

func getIpProtocol(proto int64) interface{} {
    for k, v := range IP_PROTOCOL_MAP {
        if uint8(proto) == v {
            return k
        }
    }
    return uint8(proto)
}

func getTransportSrcDestPorts(portVal string, portType string) interface{} {
    var portRange string = ""

    portNum, err := strconv.Atoi(portVal)
    if err != nil && strings.Contains(portVal, "-") {
        portRange = portVal
    }

    if len(portRange) > 0 {
        return portRange
    } else if portNum > 0 {
        return uint16(portNum)
    } else {
        if "src" == portType {
            return ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_ANY
        } else if "dest" == portType {
            return ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_ANY
        }
    }
    return nil
}

func getTransportConfigTcpFlags(tcpFlags string) []ocbinds.E_OpenconfigPacketMatchTypes_TCP_FLAGS {
    var flags []ocbinds.E_OpenconfigPacketMatchTypes_TCP_FLAGS
    if len(tcpFlags) > 0 {
        flagStr := strings.Split(tcpFlags, "/")[0]
        flagNumber, _ := strconv.ParseUint(strings.Replace(flagStr, "0x", "", -1), 16, 32)
        for i := 0; i < 8; i++ {
            mask := 1 << uint(i)
            if (int(flagNumber) & mask) > 0 {
                switch int(flagNumber) & mask {
                case 0x01:
                    flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_FIN)
                case 0x02:
                    flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_SYN)
                case 0x04:
                    flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_RST)
                case 0x08:
                    flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_PSH)
                case 0x10:
                    flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ACK)
                case 0x20:
                    flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_URG)
                case 0x40:
                    flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ECE)
                case 0x80:
                    flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_CWR)
                default:
                }
            }
        }
    }
    return flags
}

func getL2EtherType(etherType uint64) interface{} {
    for k, v := range ETHERTYPE_MAP {
        if uint32(etherType) == v {
            return k
        }
    }
    return uint16(etherType)
}

var YangToDb_acl_set_name_xfmr FieldXfmrYangToDb = func (inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    log.Info("YangToDb_acl_set_name_xfmr: ")
    /*no-op since there is no redis table field to be filled corresponding to name attribute since its part of key */
    return res_map, err
}

var DbToYang_acl_set_name_xfmr FieldXfmrDbtoYang = func (inParams XfmrParams)  (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_acl_set_name_xfmr: ", inParams.key)
    /*name attribute corresponds to key in redis table*/
    aclName, _ := getOCAclKeysFromStrDBKey(inParams.key)
    res_map["name"] = aclName
    log.Info("acl-set/config/name  ", res_map)
    return res_map, err
}

var YangToDb_acl_entry_key_xfmr KeyXfmrYangToDb = func (inParams XfmrParams) (string, error) {
    var entry_key string
    var err error
    var oc_aclType ocbinds.E_OpenconfigAcl_ACL_TYPE
    log.Info("YangToDb_acl_entry_key_xfmr: ", inParams.ygRoot, inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    if len(pathInfo.Vars) < 3 {
        err = errors.New("Invalid xpath, key attributes not found")
        return entry_key, err
    }

    oc_aclType, err = getAclTypeOCEnumFromName(pathInfo.Var("type"))
    if err != nil {
        err = errors.New("OC Acl type name to OC Acl Enum failed")
        return entry_key, err
    }

    aclkey := getAclKeyStrFromOCKey(pathInfo.Var("name"), oc_aclType)
    var rulekey string
    if strings.Contains(pathInfo.Template, "/acl-entry{sequence-id}") {
        rulekey = "RULE_" + pathInfo.Var("sequence-id")
    }
    entry_key = aclkey + "|" + rulekey

    log.Info("YangToDb_acl_entry_key_xfmr - entry_key : ", entry_key)

    return entry_key, err
}

var DbToYang_acl_entry_key_xfmr KeyXfmrDbToYang = func (inParams XfmrParams) (map[string]string, error) {
    rmap := make(map[string]string)
    var err error
    entry_key := inParams.key
    log.Info("DbToYang_acl_entry_key_xfmr: ", entry_key)

    key := strings.Split(entry_key, "|")
    if len(key) < 2 {
        err = errors.New("Invalid key for acl entries.")
        log.Info("Invalid Keys for acl enmtries", entry_key)
        return rmap, err
    }

    dbAclRule := key[1]
    seqId := strings.Replace(dbAclRule, "RULE_", "", 1)
    rmap["sequence-id"] = seqId
    return rmap, err
}

var YangToDb_acl_entry_sequenceid_xfmr FieldXfmrYangToDb = func (inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    log.Info("YangToDb_acl_entry_sequenceid_xfmr: ")
    /*no-op since there is no redis table field to be filled corresponding to sequenec-id attribute since its part of key */
    return res_map, err
}

var DbToYang_acl_entry_sequenceid_xfmr FieldXfmrDbtoYang = func (inParams XfmrParams)  (map[string]interface{}, error) {
    res_map := make(map[string]interface{})
    var err error
    log.Info("DbToYang_acl_entry_sequenceid_xfmr: ", inParams.key)
    /*sequenec-id attribute corresponds to key in redis table*/
    res, err := DbToYang_acl_entry_key_xfmr(inParams)
    log.Info("acl-entry/config/sequence-id ", res)
    if err != nil {
	    return res_map, err
    }
    if seqId, ok := res["sequence-id"]; !ok {
	    log.Error("sequence-id not found in acl entry")
	    return res_map, err
    } else {
	    res_map["sequence-id"] = seqId
    }
    return res_map, err
}

var YangToDb_acl_l2_ethertype_xfmr FieldXfmrYangToDb = func (inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error

    ethertypeType := reflect.TypeOf(inParams.param).Elem()
    log.Info("YangToDb_acl_ip_protocol_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " ethertypeType: ", ethertypeType)
    var b bytes.Buffer
    switch ethertypeType {
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_E_OpenconfigPacketMatchTypes_ETHERTYPE{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_E_OpenconfigPacketMatchTypes_ETHERTYPE)
        fmt.Fprintf(&b, "0x%0.4x", ETHERTYPE_MAP[v.E_OpenconfigPacketMatchTypes_ETHERTYPE])
        res_map["ETHER_TYPE"] = b.String()
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_Uint16{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_Uint16)
        fmt.Fprintf(&b, "0x%0.4x", v.Uint16)
        res_map["ETHER_TYPE"] = b.String()
        break
    }
    return res_map, err
}

var DbToYang_acl_l2_ethertype_xfmr FieldXfmrDbtoYang = func (inParams XfmrParams)  (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})
    data:= (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_acl_l2_ethertype_xfmr", data, inParams.ygRoot)
    if _, ok := data[RULE_TABLE]; !ok {
        err = errors.New("RULE_TABLE entry not found in the input param")
        return result, err
    }

    ruleTbl   := data[RULE_TABLE]
    ruleInst  := ruleTbl[inParams.key]
    etype, ok := ruleInst.Field["ETHER_TYPE"]

    if ok {
        etypeVal, _ := strconv.ParseUint(strings.Replace(etype, "0x", "", -1), 16, 32)
        result["protocol"] = getL2EtherType(etypeVal)
    } else {
        err = errors.New("ETHER_TYPE field not found in DB")
    }
    return result, nil

    /*
    if _, ok := data[RULE_TABLE]; !ok {
        err = errors.New("RULE_TABLE entry not found in the input param")
        return err
    }
    ruleTbl := data[RULE_TABLE]

    for aclRuleKey := range ruleTbl {
        ruleData := ruleTbl[aclRuleKey]
        var entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry
        entrySet, err = getAclSetEntry(aclRuleKey, ygRoot)
        if err != nil {
            log.Info("getAclSetEntry failed for :", aclRuleKey)
            continue // If its not map doesnt need to loop just return from here.
        }
        ruleKey := "ETHER_TYPE"
        if !ruleData.Has(ruleKey) {
            log.Info("No entry found for the field ", ruleKey)
            err = errors.New("ETHER_TYPE field not found in DB")
            continue
        }
        ethType, _ := strconv.ParseUint(strings.Replace(ruleData.Get(ruleKey), "0x", "", -1), 16, 32)
        ethertype := getL2EtherType(ethType)
        entrySet.L2.Config.Ethertype, _ = entrySet.L2.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union(ethertype)
        entrySet.L2.State.Ethertype, _ = entrySet.L2.State.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_State_Ethertype_Union(ethertype)

    }
    return err
    */
}

var YangToDb_acl_ip_protocol_xfmr FieldXfmrYangToDb = func (inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error

    protocolType := reflect.TypeOf(inParams.param).Elem()
    log.Info("YangToDb_acl_ip_protocol_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " protocolType: ", protocolType)
    switch (protocolType) {
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
        res_map["IP_PROTOCOL"] = strconv.FormatInt(int64(IP_PROTOCOL_MAP[v.E_OpenconfigPacketMatchTypes_IP_PROTOCOL]), 10)
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8)
        res_map["IP_PROTOCOL"] = strconv.FormatInt(int64(v.Uint8), 10)
        break
    }
    return res_map, err
}

var DbToYang_acl_ip_protocol_xfmr FieldXfmrDbtoYang = func (inParams XfmrParams)  (map[string]interface{}, error) {
    var err error
    result   := make(map[string]interface{})
    data:= (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_acl_ip_protocol_xfmr ", data, inParams.ygRoot)
    if _, ok := data[RULE_TABLE]; !ok {
        err = errors.New("RULE_TABLE entry not found in the input param")
        return result, err
    }

    ruleTbl  := data[RULE_TABLE]
    ruleInst := ruleTbl[inParams.key]
    prot, ok := ruleInst.Field["IP_PROTOCOL"]

    if ok {
        ipProto, _  := strconv.ParseInt(prot, 10, 64)
        result["protocol"] = getIpProtocol(ipProto)
    } else {
        err = errors.New("IP_PROTOCOL field not found in DB")
    }
    return result, err

    /*
    for aclRuleKey := range ruleTbl {
        ruleData := ruleTbl[aclRuleKey]
        var entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry
        entrySet, err = getAclSetEntry(aclRuleKey, ygRoot)
        if err != nil {
            log.Info("getAclSetEntry failed for :", aclRuleKey)
            continue // If its not map doesnt need to loop just return from here.
        }
        ruleKey := "IP_PROTOCOL"
        if !ruleData.Has(ruleKey) {
            log.Info("No entry found for the field ", ruleKey)
            err = errors.New("IP_PROTOCOL field not found in DB")
            continue
        }

        ipProto, _ := strconv.ParseInt(ruleData.Get(ruleKey), 10, 64)
        protocolVal := getIpProtocol(ipProto)
        entrySet.Ipv6.Config.Protocol, _ = entrySet.Ipv6.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union(protocolVal)
        entrySet.Ipv6.State.Protocol, _ = entrySet.Ipv6.State.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_State_Protocol_Union(protocolVal)
    }
    */
}

var YangToDb_acl_source_port_xfmr FieldXfmrYangToDb = func (inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error;
    sourceportType := reflect.TypeOf(inParams.param).Elem()
    log.Info("YangToDb_acl_ip_protocol_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " sourceportType: ", sourceportType)
    switch sourceportType {
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort)
        res_map["L4_SRC_PORT"] = v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort.ΛMap()["E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort"][int64(v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort)].Name
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_String{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_String)
        res_map["L4_SRC_PORT_RANGE"] = strings.Replace(v.String, "..", "-", 1)
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16)
        res_map["L4_SRC_PORT"] = strconv.FormatInt(int64(v.Uint16), 10)
        break
    }

    return res_map, err
}

func getAclSetEntry (aclRuleKey string, ygRoot *ygot.GoStruct) (*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry, error) {
    var err error
    var entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry
    aclObj := getAclRoot(ygRoot)

    key := strings.Split(aclRuleKey, "|")
    if len(key) < 2 {
        log.Info("Invalid Keys for acl entries", aclRuleKey)
        err = errors.New("Invalid Keys for acl entries")
        return entrySet, err
    }
    dbAclName := key[0]
    dbAclRule := key[1]
    var aclSetKey ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_Key
    aclSetKey.Name, aclSetKey.Type = getOCAclKeysFromStrDBKey(dbAclName)
    seqId,_ := strconv.Atoi(strings.Replace(dbAclRule, "RULE_", "", 1))

    log.Info("Accessing Ygot tree for ACL rule", aclSetKey.Name, aclSetKey.Type, seqId)
    var aclSet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet
     _, ok := aclObj.AclSets.AclSet[aclSetKey]
    if !ok {
        log.Info("ACL set not allocated")
        aclSet, _ = aclObj.AclSets.NewAclSet(aclSetKey.Name, aclSetKey.Type)
        ygot.BuildEmptyTree(aclSet)
    } else {
        aclSet = aclObj.AclSets.AclSet[aclSetKey]
    }
    if _, ok := aclSet.AclEntries.AclEntry[uint32(seqId)]; !ok {
        log.Info("ACL rule not allocated")
        entrySet_, _ := aclSet.AclEntries.NewAclEntry(uint32(seqId))
        entrySet = entrySet_
        ygot.BuildEmptyTree(entrySet)
    } else {
        entrySet = aclSet.AclEntries.AclEntry[uint32(seqId)]
    }

    return entrySet, err

}

var DbToYang_acl_source_port_xfmr FieldXfmrDbtoYang = func (inParams XfmrParams)  (map[string]interface{}, error) {
    var err error
    data:= (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_acl_source_port_xfmr: ", data, inParams.ygRoot)
    result := make(map[string]interface{})
    if _, ok := data[RULE_TABLE]; !ok {
        err = errors.New("RULE_TABLE entry not found in the input param")
        return result, err
    }
    ruleTbl  := data[RULE_TABLE]
    ruleInst := ruleTbl[inParams.key]
    port, ok := ruleInst.Field["L4_SRC_PORT"]
    if ok {
        result["source-port"] = port
        return result, nil
    }

    portRange, ok := ruleInst.Field["L4_SRC_PORT_RANGE"]
    if ok {
        result["source-port"] = portRange
        return result, nil
    } else {
        err = errors.New("PORT/PORT_RANGE field not found in DB")
    }
    return result, err
        /*

    if _, ok := data[RULE_TABLE]; !ok {
        err = errors.New("RULE_TABLE entry not found in the input param")
        return err
    }
    ruleTbl := data[RULE_TABLE]

    for aclRuleKey := range ruleTbl {
        ruleData := ruleTbl[aclRuleKey]
        sp :=  ruleData.Has("L4_SRC_PORT")
        spr := ruleData.Has("L4_SRC_PORT_RANGE")

        if !sp && !spr {
            log.Info("Src port field is not present in the field.")
            continue
        }
        var entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry
        entrySet, err = getAclSetEntry(aclRuleKey, ygRoot)
        if err != nil {
            log.Info("getAclSetEntry failed for :", aclRuleKey)
            continue // If its not map doesnt need to loop just return from here.
        }
        var ruleKey string
        if sp  {
            ruleKey = "L4_SRC_PORT"
        } else {
            ruleKey = "L4_SRC_PORT_RANGE"
        }
        port := ruleData.Get(ruleKey)
        srcPort := getTransportSrcDestPorts(port, "src")
        entrySet.Transport.Config.SourcePort, _ = entrySet.Transport.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union(srcPort)
        entrySet.Transport.State.SourcePort, _ = entrySet.Transport.State.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_State_SourcePort_Union(srcPort)
    }
    return err
    */
}

var YangToDb_acl_destination_port_xfmr FieldXfmrYangToDb = func (inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error;
    destportType := reflect.TypeOf(inParams.param).Elem()
    log.Info("YangToDb_acl_ip_protocol_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " destportType: ", destportType)
    switch destportType {
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort)
        res_map["L4_DST_PORT"] = v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort.ΛMap()["E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort"][int64(v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort)].Name
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_String{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_String)
        res_map["L4_DST_PORT_RANGE"] = strings.Replace(v.String, "..", "-", 1)
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16{}):
        v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16)
        res_map["L4_DST_PORT"] = strconv.FormatInt(int64(v.Uint16), 10)
        break
    }
    return res_map, err
}

var DbToYang_acl_destination_port_xfmr FieldXfmrDbtoYang = func (inParams XfmrParams)  (map[string]interface{}, error) {
    var err error
    data:= (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_acl_destination_port_xfmr: ", data, inParams.ygRoot)
    result := make(map[string]interface{})
    if _, ok := data[RULE_TABLE]; !ok {
        err = errors.New("RULE_TABLE entry not found in the input param")
        return result, err
    }
    ruleTbl  := data[RULE_TABLE]
    ruleInst := ruleTbl[inParams.key]
    port, ok := ruleInst.Field["L4_DST_PORT"]
    if ok {
        result["destination-port"] = port
        return result, nil
    }

    portRange, ok := ruleInst.Field["L4_DST_PORT_RANGE"]
    if ok {
        result["destination-port"] = portRange
        return result, nil
    } else {
        err = errors.New("DST PORT/PORT_RANGE field not found in DB")
    }
    return result, err
        /*
    if _, ok := data[RULE_TABLE]; !ok {
        err = errors.New("RULE_TABLE entry not found in the input param")
        return err
    }
    ruleTbl := data[RULE_TABLE]
    for aclRuleKey := range ruleTbl {
        ruleData := ruleTbl[aclRuleKey]
        dp :=  ruleData.Has("L4_DST_PORT")
        dpr := ruleData.Has("L4_DST_PORT_RANGE")
        if !dp && !dpr {
            log.Info("DST port field is not present in the field.")
            continue
        }

        var entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry
        entrySet, err = getAclSetEntry(aclRuleKey, ygRoot)
        if err != nil {
            log.Info("getAclSetEntry failed for :", aclRuleKey)
            continue // If its not map doesnt need to loop just return from here.
            var ruleKey string
            if dp  {
                ruleKey = "L4_DST_PORT"
            } else {
                ruleKey = "L4_DST_PORT_RANGE"
            }
            port := ruleData.Get(ruleKey)
            destPort := getTransportSrcDestPorts(port, "dest")
            entrySet.Transport.Config.DestinationPort, _ = entrySet.Transport.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union(destPort)
            entrySet.Transport.State.DestinationPort, _ = entrySet.Transport.State.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_State_DestinationPort_Union(destPort)
        }

    }
    return err
    */
}

var YangToDb_acl_tcp_flags_xfmr FieldXfmrYangToDb = func (inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error;
    log.Info("YangToDb_acl_tcp_flags_xfmr: ", inParams.ygRoot, inParams.uri)
    var tcpFlags uint32 = 0x00
    v := reflect.ValueOf(inParams.param)

    flags := v.Interface().([]ocbinds.E_OpenconfigPacketMatchTypes_TCP_FLAGS)
    for _, flag := range flags {
        fmt.Println("TCP Flag name: " + flag.ΛMap()["E_OpenconfigPacketMatchTypes_TCP_FLAGS"][int64(flag)].Name)
        switch flag {
        case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_FIN:
            tcpFlags |= 0x01
            break
        case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_SYN:
            tcpFlags |= 0x02
            break
        case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_RST:
            tcpFlags |= 0x04
            break
        case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_PSH:
            tcpFlags |= 0x08
            break
        case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ACK:
            tcpFlags |= 0x10
            break
        case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_URG:
            tcpFlags |= 0x20
            break
        case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ECE:
            tcpFlags |= 0x40
            break
        case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_CWR:
            tcpFlags |= 0x80
            break
        }
    }
    var b bytes.Buffer
    fmt.Fprintf(&b, "0x%0.2x/0x%0.2x", tcpFlags, tcpFlags)
    res_map["TCP_FLAGS"] = b.String()
    return res_map, err
}

var DbToYang_acl_tcp_flags_xfmr FieldXfmrDbtoYang = func (inParams XfmrParams)  (map[string]interface{}, error) {
    var err error
    data:= (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_acl_tcp_flags_xfmr: ", data, inParams.ygRoot)
    result := make(map[string]interface{})
    if _, ok := data[RULE_TABLE]; !ok {
        err = errors.New("RULE_TABLE entry not found in the input param")
        return result, err
    }
    ruleTbl  := data[RULE_TABLE]
    ruleInst := ruleTbl[inParams.key]
    tcpFlag, ok := ruleInst.Field["TCP_FLAGS"]
    if ok {
        result["tcp-flags"] = getTransportConfigTcpFlags(tcpFlag)
        return result, nil
    }
    return result, nil
        /*

    if _, ok := data[RULE_TABLE]; !ok {
        err = errors.New("RULE_TABLE entry not found in the input param")
        return err
    }
    ruleTbl := data[RULE_TABLE]

    for aclRuleKey := range ruleTbl {
        ruleData := ruleTbl[aclRuleKey]
        var entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry
        entrySet, err = getAclSetEntry(aclRuleKey, ygRoot)
        if err != nil {
            log.Info("getAclSetEntry failed for :", aclRuleKey)
            continue // If its not map doesnt need to loop just return from here.
        }
        ruleKey := "TCP_FLAGS"
        if !ruleData.Has(ruleKey) {
            log.Info("No entry found for the field ", ruleKey)
            err = errors.New("TCP_FLAGS field not found in DB")
            continue
        }
        tcpFlags := ruleData.Get(ruleKey)
        entrySet.Transport.Config.TcpFlags = getTransportConfigTcpFlags(tcpFlags)
        entrySet.Transport.State.TcpFlags = getTransportConfigTcpFlags(tcpFlags)
    }
    return err
    */
}

func convertDBAclRulesToInternal(dbCl *db.DB, aclName string, seqId int64, ruleKey db.Key) (ruleTableMap map[string]map[string]db.Value, ferr error) {
    ruleTs := &db.TableSpec{Name: RULE_TABLE}
    if seqId != -1 {
        ruleKey.Comp = []string{aclName, "RULE_" + strconv.FormatInt(int64(seqId), 10)}
    }
    if ruleKey.Len() > 1 {
        ruleName := ruleKey.Get(1)
        if ruleName != "DEFAULT_RULE" {
            ruleData, err := dbCl.GetEntry(ruleTs, ruleKey)
            if err != nil {
                ferr = err
                return
            }
            if len(ruleTableMap) == 0 {
                ruleTableMap = make(map[string]map[string]db.Value)
            }
            _, ok := ruleTableMap[aclName]
            if !ok {
                ruleTableMap[aclName] = make(map[string]db.Value)
                ruleTableMap[aclName][ruleName] =  db.Value{Field: make(map[string]string)}
            }
            ruleTableMap[aclName][ruleName] = ruleData
        }
    } else {
        ruleKeys, err := dbCl.GetKeys(ruleTs)
        if err != nil {
            ferr = err
            return
        }
        for i, _ := range ruleKeys {
            if aclName == ruleKeys[i].Get(0) {
                ruleTableMap, ferr = convertDBAclRulesToInternal(dbCl, aclName, -1, ruleKeys[i])
            }
        }
    }
    return
}

func convertDBAclToInternal(dbCl *db.DB, aclkey db.Key) (aclTableMap map[string]db.Value, ruleTableMap map[string]map[string]db.Value, ferr error) {
    aclTs := &db.TableSpec{Name: ACL_TABLE}
    if aclkey.Len() > 0 {
        // Get one particular ACL
        entry, err := dbCl.GetEntry(aclTs, aclkey)
        if err != nil {
            ferr = err
            return
        }
        if entry.IsPopulated() {
            log.Info("convertDBAclToInternal : ", aclkey, aclTableMap, " ", aclTableMap[aclkey.Get(0)], " ", entry)
            _, ok := aclTableMap[aclkey.Get(0)]
            if !ok {
                if len(aclTableMap) == 0 {
                    aclTableMap = make(map[string]db.Value)
                }
                aclTableMap[aclkey.Get(0)] = db.Value{Field: make(map[string]string)}
            }

            aclTableMap[aclkey.Get(0)] = entry
            if len(ruleTableMap) == 0 {
                ruleTableMap = make(map[string]map[string]db.Value)
            }
            _, rok := ruleTableMap[aclkey.Get(0)]
            if !rok {
                ruleTableMap[aclkey.Get(0)] = make(map[string]db.Value)
            }
            ruleTableMap, ferr  = convertDBAclRulesToInternal(dbCl, aclkey.Get(0), -1, db.Key{})
            if err != nil {
                ferr = err
                return
            }
        } else {
            ferr = tlerr.NotFound("Acl %s is not configured", aclkey.Get(0))
            return
        }
    } else {
        // Get all ACLs
        tbl, err := dbCl.GetTable(aclTs)
        if err != nil {
            ferr = err
            return
        }
        keys, _ := tbl.GetKeys()
        for i, _ := range keys {
            aclTableMap, ruleTableMap, ferr = convertDBAclToInternal(dbCl, keys[i])
        }
    }
    return
}

func getDbAlcTblsData (d *db.DB) (map[string]db.Value, map[string]map[string]db.Value, error) {
    var err error

    aclTableMap := make(map[string]db.Value)
    ruleTableMap := make(map[string]map[string]db.Value)

    aclTableMap, ruleTableMap, err = convertDBAclToInternal (d, db.Key{})

    return aclTableMap, ruleTableMap, err
}

var YangToDb_acl_port_bindings_xfmr SubTreeXfmrYangToDb = func (inParams XfmrParams) (map[string]map[string]db.Value, error) {
    res_map := make(map[string]map[string]db.Value)
    aclTableMap := make(map[string]db.Value)
    log.Info("YangToDb_acl_port_bindings_xfmr: ", inParams.ygRoot, inParams.uri)

    aclObj := getAclRoot(inParams.ygRoot)

    aclTableMapDb, _, err := getDbAlcTblsData(inParams.d)
    if err != nil {
        log.Info("YangToDb_acl_port_bindings_xfmr: getDbAlcTblsData not able to populate acl tables.")
    }

    if aclObj.Interfaces != nil {
        if len(aclObj.Interfaces.Interface) > 0 {
            aclInterfacesMap := make(map[string][]string)
            for intfId, _ := range aclObj.Interfaces.Interface {
                intf := aclObj.Interfaces.Interface[intfId]
                if intf != nil {
                    if intf.IngressAclSets != nil && len(intf.IngressAclSets.IngressAclSet) > 0 {
                        for inAclKey, _ := range intf.IngressAclSets.IngressAclSet {
                            aclName := getAclKeyStrFromOCKey(inAclKey.SetName, inAclKey.Type)
                            if intf.InterfaceRef != nil && intf.InterfaceRef.Config.Interface != nil {
                                aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.InterfaceRef.Config.Interface)
                            } else {
                                aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.Id)
                            }
                            _, ok := aclTableMap[aclName]
                            if !ok {
                                aclTableMap[aclName] = db.Value{Field: make(map[string]string)}
                            }
                            aclTableMap[aclName].Field["stage"] = "INGRESS"
                        }
                    }
                    if intf.EgressAclSets != nil && len(intf.EgressAclSets.EgressAclSet) > 0 {
                        for outAclKey, _ := range intf.EgressAclSets.EgressAclSet {
                            aclName := getAclKeyStrFromOCKey(outAclKey.SetName, outAclKey.Type)
                            if intf.InterfaceRef != nil && intf.InterfaceRef.Config.Interface != nil {
                                aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.InterfaceRef.Config.Interface)
                            } else {
                                aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.Id)
                            }
                            _, ok := aclTableMap[aclName]
                            if !ok {
                                aclTableMap[aclName] = db.Value{Field: make(map[string]string)}
                            }
                            aclTableMap[aclName].Field["stage"] = "EGRESS"
                        }
                    }
                    if intf.IngressAclSets == nil && intf.EgressAclSets == nil {
                        for aclName := range aclTableMapDb {
                            _, ok := aclTableMap[aclName]
                            if !ok {
                                aclTableMap[aclName] = db.Value{Field: make(map[string]string)}
                            }
                            aclEntryDb := aclTableMapDb[aclName]
                            intfsDb := aclEntryDb.GetList("ports")
                            if contains(intfsDb, intfId) {
                                var intfs []string
                                intfs = append(intfs, intfId)
                                aclTableMap[aclName].Field["stage"] = aclEntryDb.Get("stage")
                                val := aclTableMap[aclName]
                                (&val).SetList("ports", intfs)
                            }
                        }
                    }
                }
            }
            for k, _ := range aclInterfacesMap {
                val := aclTableMap[k]
                (&val).SetList("ports", aclInterfacesMap[k])
            }
        } else {
            for aclName := range aclTableMapDb {
                _, ok := aclTableMap[aclName]
                if !ok {
                    aclTableMap[aclName] = db.Value{Field: make(map[string]string)}
                }
                aclEntryDb := aclTableMapDb[aclName]
                aclTableMap[aclName].Field["stage"] = aclEntryDb.Get("stage")
                val := aclTableMap[aclName]
                (&val).SetList("ports", aclEntryDb.GetList("ports"))
            }
        }
    }
    res_map[ACL_TABLE] = aclTableMap
    return res_map, err
}

var DbToYang_acl_port_bindings_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) (error) {
    var err error
    data:= (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_acl_port_bindings_xfmr: ", data, inParams.ygRoot)

    aclTbl, ruleTbl, err := getDbAlcTblsData(inParams.d)

    if err != nil {
        log.Info("getDbAlcTblsData not able to populate acl tables.")
        err = errors.New("getDbAlcTblsData failed to populate tables.")
        return err
    }
    pathInfo := NewPathInfo(inParams.uri)
    acl := getAclRoot(inParams.ygRoot)
    targetUriPath, _ := getYangPathFromUri(pathInfo.Path)
    if isSubtreeRequest(pathInfo.Template, "/openconfig-acl:acl/interfaces/interface{}") {
        for intfId := range acl.Interfaces.Interface {
            intfData := acl.Interfaces.Interface[intfId]
            ygot.BuildEmptyTree(intfData)
            if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces/interface/ingress-acl-sets") {
                err = getAclBindingInfoForInterfaceData(aclTbl, ruleTbl, intfData, intfId, "INGRESS")
            } else if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces/interface/egress-acl-sets") {
                err = getAclBindingInfoForInterfaceData(aclTbl, ruleTbl, intfData, intfId, "EGRESS")
            } else {
                err = getAclBindingInfoForInterfaceData(aclTbl, ruleTbl, intfData, intfId, "INGRESS")
                if err != nil {
                    return err
                }
                err = getAclBindingInfoForInterfaceData(aclTbl, ruleTbl, intfData, intfId, "EGRESS")
            }
        }
    } else {
        err = getAllBindingsInfo(aclTbl, ruleTbl, inParams.ygRoot)
    }

    return err
}
func convertInternalToOCAclRuleBinding(aclTableMap map[string]db.Value, ruleTableMap map[string]map[string]db.Value, priority uint32, seqId int64, direction string, aclSet ygot.GoStruct, entrySet ygot.GoStruct) {
    if seqId == -1 {
        seqId = int64(MAX_PRIORITY - priority)
    }

    var num uint64
    num = 0
    var ruleId uint32 = uint32(seqId)

    if direction == "INGRESS" {
        var ingressEntrySet *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_AclEntries_AclEntry
        var ok bool
        if entrySet == nil {
            ingressAclSet := aclSet.(*ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet)
            if ingressEntrySet, ok = ingressAclSet.AclEntries.AclEntry[ruleId]; !ok {
                ingressEntrySet, _ = ingressAclSet.AclEntries.NewAclEntry(ruleId)
            }
        } else {
            ingressEntrySet = entrySet.(*ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_AclEntries_AclEntry)
        }
        if ingressEntrySet != nil {
            ygot.BuildEmptyTree(ingressEntrySet)
            ingressEntrySet.State.SequenceId = &ruleId
            ingressEntrySet.State.MatchedPackets = &num
            ingressEntrySet.State.MatchedOctets = &num
        }
    } else if direction == "EGRESS" {
        var egressEntrySet *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_AclEntries_AclEntry
        var ok bool
        if entrySet == nil {
            egressAclSet := aclSet.(*ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet)
            if egressEntrySet, ok = egressAclSet.AclEntries.AclEntry[ruleId]; !ok {
                egressEntrySet, _ = egressAclSet.AclEntries.NewAclEntry(ruleId)
            }
        } else {
            egressEntrySet = entrySet.(*ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_AclEntries_AclEntry)
        }
        if egressEntrySet != nil {
            ygot.BuildEmptyTree(egressEntrySet)
            egressEntrySet.State.SequenceId = &ruleId
            egressEntrySet.State.MatchedPackets = &num
            egressEntrySet.State.MatchedOctets = &num
        }
    }
}

func convertInternalToOCAclBinding(aclTableMap map[string]db.Value, ruleTableMap map[string]map[string]db.Value, aclName string, intfId string, direction string, intfAclSet ygot.GoStruct) error {
    var err error
    if _, ok := aclTableMap[aclName]; !ok {
        err = errors.New("Acl entry not found, convertInternalToOCAclBinding")
        return err
    } else {
        aclEntry := aclTableMap[aclName]
        if !contains(aclEntry.GetList("ports"), intfId) {
            return tlerr.InvalidArgs("Acl %s not binded with %s", aclName, intfId)
        }
    }

    for ruleName := range ruleTableMap[aclName] {
        if ruleName != "DEFAULT_RULE" {
            seqId, _ := strconv.Atoi(strings.Replace(ruleName, "RULE_", "", 1))
            convertInternalToOCAclRuleBinding(aclTableMap, ruleTableMap, 0, int64(seqId), direction, intfAclSet, nil)
        }
    }

    return err
}

func getAllBindingsInfo(aclTableMap map[string]db.Value, ruleTableMap map[string]map[string]db.Value, ygRoot *ygot.GoStruct) error {
    var err error
    acl := getAclRoot(ygRoot)

    var interfaces []string
    for aclName := range aclTableMap {
        aclData := aclTableMap[aclName]
        if len(aclData.Get("ports@")) > 0 {
            aclIntfs := aclData.GetList("ports")
            for i, _ := range aclIntfs {
                if !contains(interfaces, aclIntfs[i]) && aclIntfs[i] != "" {
                    interfaces = append(interfaces, aclIntfs[i])
                }
            }
        }
    }
    ygot.BuildEmptyTree(acl)
    for _, intfId := range interfaces {
        var intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface
        intfData, ok := acl.Interfaces.Interface[intfId]
        if !ok {
            intfData, _ = acl.Interfaces.NewInterface(intfId)
        }
        ygot.BuildEmptyTree(intfData)
        err = getAclBindingInfoForInterfaceData(aclTableMap, ruleTableMap, intfData, intfId, "INGRESS")
        err = getAclBindingInfoForInterfaceData(aclTableMap, ruleTableMap, intfData, intfId, "EGRESS")
    }
    return err
}

func getAclBindingInfoForInterfaceData(aclTableMap map[string]db.Value, ruleTableMap map[string]map[string]db.Value, intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface, intfId string, direction string) error {
    var err error
    if intfData != nil {
        intfData.Config.Id = intfData.Id
        intfData.State.Id = intfData.Id
    }
    if direction == "INGRESS" {
        if intfData.IngressAclSets != nil && len(intfData.IngressAclSets.IngressAclSet) > 0 {
            for ingressAclSetKey, _ := range intfData.IngressAclSets.IngressAclSet {
                aclName := strings.Replace(strings.Replace(ingressAclSetKey.SetName, " ", "_", -1), "-", "_", -1)
                aclType := ingressAclSetKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(ingressAclSetKey.Type)].Name
                aclKey := aclName + "_" + aclType

                ingressAclSet := intfData.IngressAclSets.IngressAclSet[ingressAclSetKey]
                if ingressAclSet != nil && ingressAclSet.AclEntries != nil && len(ingressAclSet.AclEntries.AclEntry) > 0 {
                    for seqId, _ := range ingressAclSet.AclEntries.AclEntry {
                        rulekey := "RULE_" + strconv.Itoa(int(seqId))
                        entrySet := ingressAclSet.AclEntries.AclEntry[seqId]
                        _, ok := ruleTableMap[aclKey][rulekey]
                        if !ok {
                            log.Info("Acl Rule not found ", aclKey, rulekey)
                            err = errors.New("Acl Rule not found ingress, getAclBindingInfoForInterfaceData")
                            return err
                        }
                        convertInternalToOCAclRuleBinding(aclTableMap, ruleTableMap, 0, int64(seqId), direction, nil, entrySet)
                    }
                } else {
                    ygot.BuildEmptyTree(ingressAclSet)
                    ingressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_Config{SetName: &aclName, Type: ingressAclSetKey.Type}
                    ingressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_State{SetName: &aclName, Type: ingressAclSetKey.Type}
                    err = convertInternalToOCAclBinding(aclTableMap, ruleTableMap, aclKey, intfId, direction, ingressAclSet)
                }
            }
        } else {
            err = findAndGetAclBindingInfoForInterfaceData(aclTableMap, ruleTableMap, intfId, direction, intfData)
        }
    } else if direction == "EGRESS" {
        if intfData.EgressAclSets != nil && len(intfData.EgressAclSets.EgressAclSet) > 0 {
            for egressAclSetKey, _ := range intfData.EgressAclSets.EgressAclSet {
                aclName := strings.Replace(strings.Replace(egressAclSetKey.SetName, " ", "_", -1), "-", "_", -1)
                aclType := egressAclSetKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(egressAclSetKey.Type)].Name
                aclKey := aclName + "_" + aclType

                egressAclSet := intfData.EgressAclSets.EgressAclSet[egressAclSetKey]
                if egressAclSet != nil && egressAclSet.AclEntries != nil && len(egressAclSet.AclEntries.AclEntry) > 0 {
                    for seqId, _ := range egressAclSet.AclEntries.AclEntry {
                        rulekey := "RULE_" + strconv.Itoa(int(seqId))
                        entrySet := egressAclSet.AclEntries.AclEntry[seqId]
                        _, ok := ruleTableMap[aclKey][rulekey]
                        if !ok {
                            log.Info("Acl Rule not found ", aclKey, rulekey)
                            err = errors.New("Acl Rule not found egress, getAclBindingInfoForInterfaceData")
                            return err
                        }
                        convertInternalToOCAclRuleBinding(aclTableMap, ruleTableMap, 0, int64(seqId), direction, nil, entrySet)
                    }
                } else {
                    ygot.BuildEmptyTree(egressAclSet)
                    egressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_Config{SetName: &aclName, Type: egressAclSetKey.Type}
                    egressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_State{SetName: &aclName, Type: egressAclSetKey.Type}
                    err = convertInternalToOCAclBinding(aclTableMap, ruleTableMap, aclKey, intfId, direction, egressAclSet)
                }
            }
        } else {
            err = findAndGetAclBindingInfoForInterfaceData(aclTableMap, ruleTableMap, intfId, direction, intfData)
        }
    } else {
        log.Error("Unknown direction")
    }
    return err
}

func findAndGetAclBindingInfoForInterfaceData(aclTableMap map[string]db.Value, ruleTableMap map[string]map[string]db.Value, intfId string, direction string, intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface) error {
    var err error
    for aclName, _ := range aclTableMap {
        aclData := aclTableMap[aclName]
        aclIntfs := aclData.GetList("ports")
        aclType := aclData.Get(ACL_TYPE)
        var aclOrigName string
        var aclOrigType ocbinds.E_OpenconfigAcl_ACL_TYPE
        if SONIC_ACL_TYPE_IPV4 == aclType {
            aclOrigName = strings.Replace(aclName, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
            aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4
        } else if SONIC_ACL_TYPE_IPV6 == aclType {
            aclOrigName = strings.Replace(aclName, "_"+OPENCONFIG_ACL_TYPE_IPV6, "", 1)
            aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6
        } else if SONIC_ACL_TYPE_L2 == aclType {
            aclOrigName = strings.Replace(aclName, "_"+OPENCONFIG_ACL_TYPE_L2, "", 1)
            aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2
        }

        if contains(aclIntfs, intfId) && direction == aclData.Get("stage") {
            if direction == "INGRESS" {
                if intfData.IngressAclSets != nil {
                    aclSetKey := ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_Key{SetName: aclOrigName, Type: aclOrigType}
                    ingressAclSet, ok := intfData.IngressAclSets.IngressAclSet[aclSetKey]
                    if !ok {
                        ingressAclSet, _ = intfData.IngressAclSets.NewIngressAclSet(aclOrigName, aclOrigType)
                        ygot.BuildEmptyTree(ingressAclSet)
                        ingressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_Config{SetName: &aclOrigName, Type: aclOrigType}
                        ingressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_State{SetName: &aclOrigName, Type: aclOrigType}
                    }
                    err = convertInternalToOCAclBinding(aclTableMap, ruleTableMap, aclName, intfId, direction, ingressAclSet)
                    if err != nil {
                        return err
                    }
                }
            } else if direction == "EGRESS" {
                if intfData.EgressAclSets != nil {
                    aclSetKey := ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_Key{SetName: aclOrigName, Type: aclOrigType}
                    egressAclSet, ok := intfData.EgressAclSets.EgressAclSet[aclSetKey]
                    if !ok {
                        egressAclSet, _ = intfData.EgressAclSets.NewEgressAclSet(aclOrigName, aclOrigType)
                        ygot.BuildEmptyTree(egressAclSet)
                        egressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_Config{SetName: &aclOrigName, Type: aclOrigType}
                        egressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_State{SetName: &aclOrigName, Type: aclOrigType}
                    }
                    err = convertInternalToOCAclBinding(aclTableMap, ruleTableMap, aclName, intfId, direction, egressAclSet)
                    if err != nil {
                        return err
                    }
                }
            }
        }
    }
    return err
}
