package transformer

import (
    "fmt"
    "bytes"
    "errors"
    "strings"
    //	"os"
    //	"sort"
    //	"github.com/openconfig/goyang/pkg/yang"
    "github.com/openconfig/ygot/ygot"
    "strconv"
    "translib/db"
    "reflect"
    log "github.com/golang/glog"
    "translib/ocbinds"
    "translib/tlerr"
)

func init () {
    XlateFuncBind("YangToDb_acl_entry_key_xfmr", YangToDb_acl_entry_key_xfmr)
    XlateFuncBind("DbToYang_acl_entry_key_xfmr", DbToYang_acl_entry_key_xfmr)
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




var YangToDb_acl_entry_key_xfmr KeyXfmrYangToDb = func (d *db.DB, ygRoot *ygot.GoStruct, xpath string) (string, error) {
    var entry_key string
    var err error
    var oc_aclType ocbinds.E_OpenconfigAcl_ACL_TYPE
    log.Info("YangToDb_acl_entry_key_xfmr: ", ygRoot, xpath)
    pathInfo := NewPathInfo(xpath)

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

var DbToYang_acl_entry_key_xfmr KeyXfmrDbToYang = func (d *db.DB, entry_key string) (map[string]map[string]string, error) {
    res_map := make(map[string]map[string]string)
    var err error
    log.Info("DbToYang_acl_entry_key_xfmr: ", entry_key)

    key := strings.Split(entry_key, "|")
    if len(key) < 2 {
        err = errors.New("Invalid key for acl entries.")
        log.Info("Invalid Keys for acl enmtries", entry_key)
        return res_map, err
    }

    dbAclName := key[0]
    dbAclRule := key[1]
    aclName, aclType := getOCAclKeysFromStrDBKey(dbAclName)
    res_map["oc-acl:acl-set"]["name"] = aclName
    res_map["oc-acl:acl-set"]["type"] = aclType.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(aclType)].Name
    seqId := strings.Replace(dbAclRule, "RULE_", "", 1)
    res_map["acl-entry"]["sequence-id"] = seqId
    log.Info("DbToYang_acl_entry_key_xfmr - res_map: ", res_map)
    return res_map, err
}

func getAclRule(acl *ocbinds.OpenconfigAcl_Acl, aclName string, aclType ocbinds.E_OpenconfigAcl_ACL_TYPE, seqId uint32) (*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry, error) {
    var err error
    var aclSetKey ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_Key
    aclSetKey.Name = aclName
    aclSetKey.Type = aclType

    if _, ok := acl.AclSets.AclSet[aclSetKey]; !ok {
        err = errors.New("AclSet not found " + aclName)
        return nil, err
    }

    aclSet := acl.AclSets.AclSet[aclSetKey]
    if _, ok :=  aclSet.AclEntries.AclEntry[seqId]; !ok {
        err = errors.New("Acl Rule not found " + aclName + strconv.FormatInt(int64(seqId), 10))
        return nil, err
    }

    rule := aclSet.AclEntries.AclEntry[seqId]
    return rule, err
}

var YangToDb_acl_l2_ethertype_xfmr FieldXfmrYangToDb = func (d *db.DB, ygRoot *ygot.GoStruct, xpath string, ethertype interface {}) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    log.Info("YangToDb_acl_l2_ethertype_xfmr :", ygRoot, xpath)

    ethertypeType := reflect.TypeOf(ethertype).Elem()
    var b bytes.Buffer
    switch ethertypeType {
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_E_OpenconfigPacketMatchTypes_ETHERTYPE{}):
        v := (ethertype).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_E_OpenconfigPacketMatchTypes_ETHERTYPE)
        fmt.Fprintf(&b, "0x%0.4x", ETHERTYPE_MAP[v.E_OpenconfigPacketMatchTypes_ETHERTYPE])
        res_map["ETHER_TYPE"] = b.String()
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_Uint16{}):
        v := (ethertype).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_Uint16)
        fmt.Fprintf(&b, "0x%0.4x", v.Uint16)
        res_map["ETHER_TYPE"] = b.String()
        break
    }
    return res_map, err
}

var DbToYang_acl_l2_ethertype_xfmr FieldXfmrDbtoYang = func (d *db.DB, data map[string]map[string]db.Value, ygRoot *ygot.GoStruct)  (error) {
    var err error
    log.Info("DbToYang_acl_l2_ethertype_xfmr", data, ygRoot)

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
}

var YangToDb_acl_ip_protocol_xfmr FieldXfmrYangToDb = func (d *db.DB, ygRoot *ygot.GoStruct, xpath string, protocol interface {}) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error
    log.Info("YangToDb_acl_ip_protocol_xfmr: ", ygRoot, xpath)

    protocolType := reflect.TypeOf(protocol).Elem()
    switch (protocolType) {
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL{}):
        v := (protocol).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
        res_map["IP_PROTOCOL"] = strconv.FormatInt(int64(IP_PROTOCOL_MAP[v.E_OpenconfigPacketMatchTypes_IP_PROTOCOL]), 10)
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8{}):
        v := (protocol).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8)
        res_map["IP_PROTOCOL"] = strconv.FormatInt(int64(v.Uint8), 10)
        break
    }
    return res_map, err
}

var DbToYang_acl_ip_protocol_xfmr FieldXfmrDbtoYang = func (d *db.DB, data map[string]map[string]db.Value, ygRoot *ygot.GoStruct)  (error) {
    var err error
    log.Info("DbToYang_acl_ip_protocol_xfmr ", data, ygRoot)
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
    return err
}

var YangToDb_acl_source_port_xfmr FieldXfmrYangToDb = func (d *db.DB, ygRoot *ygot.GoStruct, xpath string, value interface {}) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error;
    log.Info("YangToDb_acl_source_port_xfmr: ", ygRoot, xpath)
    sourceportType := reflect.TypeOf(value).Elem()
    switch sourceportType {
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort{}):
        v := (value).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort)
        res_map["L4_SRC_PORT"] = v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort.ΛMap()["E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort"][int64(v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort)].Name
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_String{}):
        v := (value).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_String)
        res_map["L4_SRC_PORT_RANGE"] = strings.Replace(v.String, "..", "-", 1)
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16{}):
        v := (value).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16)
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
    if _, ok := aclObj.AclSets.AclSet[aclSetKey]; !ok {
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

var DbToYang_acl_source_port_xfmr FieldXfmrDbtoYang = func (d *db.DB, data map[string]map[string]db.Value, ygRoot *ygot.GoStruct)  (error) {
    var err error
    log.Info("DbToYang_acl_source_port_xfmr: ", data, ygRoot)

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
}

var YangToDb_acl_destination_port_xfmr FieldXfmrYangToDb = func (d *db.DB, ygRoot *ygot.GoStruct, xpath string, value interface{}) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error;
    log.Info("YangToDb_acl_destination_port_xfmr: ", ygRoot, xpath)
    destportType := reflect.TypeOf(value).Elem()
    switch destportType {
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort{}):
        v := (value).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort)
        res_map["L4_DST_PORT"] = v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort.ΛMap()["E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort"][int64(v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort)].Name
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_String{}):
        v := (value).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_String)
        res_map["L4_DST_PORT_RANGE"] = strings.Replace(v.String, "..", "-", 1)
        break
    case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16{}):
        v := (value).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16)
        res_map["L4_DST_PORT"] = strconv.FormatInt(int64(v.Uint16), 10)
        break
    }
    return res_map, err
}

var DbToYang_acl_destination_port_xfmr FieldXfmrDbtoYang = func (d *db.DB, data map[string]map[string]db.Value, ygRoot *ygot.GoStruct)  (error) {
    var err error
    log.Info("DbToYang_acl_destination_port_xfmr: ", data, ygRoot)
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
}

var YangToDb_acl_tcp_flags_xfmr FieldXfmrYangToDb = func (d *db.DB, ygRoot *ygot.GoStruct, xpath string, value interface {}) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error;
    log.Info("YangToDb_acl_tcp_flags_xfmr: ", ygRoot, xpath)
    var tcpFlags uint32 = 0x00
    v := reflect.ValueOf(value)

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

var DbToYang_acl_tcp_flags_xfmr FieldXfmrDbtoYang = func (d *db.DB, data map[string]map[string]db.Value, ygRoot *ygot.GoStruct)  (error) {
    var err error
    log.Info("DbToYang_acl_tcp_flags_xfmr: ", data, ygRoot)

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
}

var YangToDb_acl_port_bindings_xfmr SubTreeXfmrYangToDb = func (d *db.DB, ygRoot *ygot.GoStruct, xpath string) (map[string]map[string]db.Value, error) {
    res_map := make(map[string]map[string]db.Value)
    aclTableMap := make(map[string]db.Value)
    var err error;
    log.Info("YangToDb_acl_port_bindings_xfmr: ", ygRoot, xpath)

    aclObj := getAclRoot(ygRoot)

    if aclObj.Interfaces != nil && len(aclObj.Interfaces.Interface) > 0 {
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
                        if len(aclTableMap) == 0 {
                            aclTableMap[aclName] = db.Value{Field: map[string]string{}}
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
                        if len(aclTableMap) == 0 {
                            aclTableMap[aclName] = db.Value{Field: map[string]string{}}
                        }
                        aclTableMap[aclName].Field["stage"] = "EGRESS"
                    }
                }
            }
        }
        for k, _ := range aclInterfacesMap {
            val := aclTableMap[k]
            (&val).SetList("ports", aclInterfacesMap[k])
        }
    }
    res_map[ACL_TABLE] = aclTableMap
    return res_map, err
}

var DbToYang_acl_port_bindings_xfmr SubTreeXfmrDbToYang = func (d *db.DB, data map[string]map[string]db.Value, ygRoot *ygot.GoStruct) (error) {
    var err error
    log.Info("DbToYang_acl_port_bindings_xfmr: ", data, ygRoot)
    return err
}
