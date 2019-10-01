package transformer

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"strconv"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"
)

func init() {
	XlateFuncBind("YangToDb_acl_set_name_xfmr", YangToDb_acl_set_name_xfmr)
	XlateFuncBind("DbToYang_acl_set_name_xfmr", DbToYang_acl_set_name_xfmr)
	XlateFuncBind("YangToDb_acl_type_field_xfmr", YangToDb_acl_type_field_xfmr)
	XlateFuncBind("DbToYang_acl_type_field_xfmr", DbToYang_acl_type_field_xfmr)
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
        XlateFuncBind("YangToDb_acl_forwarding_action_xfmr", YangToDb_acl_forwarding_action_xfmr)
	XlateFuncBind("DbToYang_acl_forwarding_action_xfmr", DbToYang_acl_forwarding_action_xfmr)
	XlateFuncBind("validate_ipv4", validate_ipv4)
	XlateFuncBind("validate_ipv6", validate_ipv6)
	XlateFuncBind("acl_post_xfmr", acl_post_xfmr)
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

/* E_OpenconfigAcl_ACL_TYPE */
var ACL_TYPE_MAP = map[string]string{
	strconv.FormatInt(int64(ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4), 10): SONIC_ACL_TYPE_IPV4,
	strconv.FormatInt(int64(ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6), 10): SONIC_ACL_TYPE_IPV6,
	strconv.FormatInt(int64(ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2), 10):   SONIC_ACL_TYPE_L2,
}

/* E_OpenconfigAcl_FORWARDING_ACTION */
var ACL_FORWARDING_ACTION_MAP = map[string]string{
	strconv.FormatInt(int64(ocbinds.OpenconfigAcl_FORWARDING_ACTION_ACCEPT), 10): "FORWARD",
	strconv.FormatInt(int64(ocbinds.OpenconfigAcl_FORWARDING_ACTION_DROP), 10): "DROP",
	strconv.FormatInt(int64(ocbinds.OpenconfigAcl_FORWARDING_ACTION_REJECT), 10): "REDIRECT",
}

/* E_OpenconfigPacketMatchTypes_IP_PROTOCOL */
var IP_PROTOCOL_MAP = map[string]string{
	strconv.FormatInt(int64(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_ICMP), 10): "1",
	strconv.FormatInt(int64(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_IGMP), 10): "2",
	strconv.FormatInt(int64(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_TCP), 10):  "6",
	strconv.FormatInt(int64(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_UDP), 10):  "17",
	strconv.FormatInt(int64(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_RSVP), 10): "46",
	strconv.FormatInt(int64(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_GRE), 10):  "47",
	strconv.FormatInt(int64(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_AUTH), 10): "51",
	strconv.FormatInt(int64(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_PIM), 10):  "103",
	strconv.FormatInt(int64(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_L2TP), 10): "115",
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

func getAclRoot(s *ygot.GoStruct) *ocbinds.OpenconfigAcl_Acl {
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

////////////////////////////////////////////
// Validate callpoints
////////////////////////////////////////////
var validate_ipv4 ValidateCallpoint = func(inParams XfmrParams) (bool) {
	if strings.Contains(inParams.key, "ACL_IPV4") {
		return true
	}
	return false
}
var validate_ipv6 ValidateCallpoint = func(inParams XfmrParams) (bool) {
	if strings.Contains(inParams.key, "ACL_IPV6") {
		return true
	}
	return false
}

////////////////////////////////////////////
// Post Transformer
////////////////////////////////////////////
var acl_post_xfmr PostXfmrFunc = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
	log.Info("In Post transformer")
	//TODO: check if a default ACL Rule exists, else create one and update the resultMap with default rule
	// Return will be the updated result map
	return (*inParams.dbDataMap)[inParams.curDb], nil
}

////////////////////////////////////////////
// Bi-directoonal overloaded methods
////////////////////////////////////////////
var YangToDb_acl_forwarding_action_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	if inParams.param == nil {
	    res_map["PACKET_ACTION"] = ""
	    return res_map, err
	}
	action, _ := inParams.param.(ocbinds.E_OpenconfigAcl_FORWARDING_ACTION)
	log.Info("YangToDb_acl_forwarding_action_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " forwarding_action: ", action)
	res_map["PACKET_ACTION"] = findInMap(ACL_FORWARDING_ACTION_MAP, strconv.FormatInt(int64(action), 10))
	return res_map, err
}
var DbToYang_acl_forwarding_action_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_acl_forwarding_action_xfmr", data, inParams.ygRoot)
	oc_action := findInMap(ACL_FORWARDING_ACTION_MAP, data[RULE_TABLE][inParams.key].Field["PACKET_ACTION"])
	n, err := strconv.ParseInt(oc_action, 10, 64)
	result["forwarding-action"] = ocbinds.E_OpenconfigAcl_FORWARDING_ACTION(n).ΛMap()["E_OpenconfigAcl_FORWARDING_ACTION"][n].Name
	return result, err
}

var YangToDb_acl_type_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	if inParams.param == nil {
	    res_map[ACL_TYPE] = ""
	    return res_map, err
	}

	acltype, _ := inParams.param.(ocbinds.E_OpenconfigAcl_ACL_TYPE)
	log.Info("YangToDb_acl_type_field_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " acltype: ", acltype)
	res_map[ACL_TYPE] = findInMap(ACL_TYPE_MAP, strconv.FormatInt(int64(acltype), 10))
	return res_map, err
}
var DbToYang_acl_type_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_acl_type_field_xfmr", data, inParams.ygRoot)
	oc_acltype := findInMap(ACL_TYPE_MAP, data[ACL_TABLE][inParams.key].Field[ACL_TYPE])
	n, err := strconv.ParseInt(oc_acltype, 10, 64)
	result[ACL_TYPE] = ocbinds.E_OpenconfigAcl_ACL_TYPE(n).ΛMap()["E_OpenconfigAcl_ACL_TYPE"][n].Name
	return result, err
}

var YangToDb_acl_set_name_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	log.Info("YangToDb_acl_set_name_xfmr: ")
	/*no-op since there is no redis table field to be filled corresponding to name attribute since its part of key */
	return res_map, err
}

var DbToYang_acl_set_name_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	res_map := make(map[string]interface{})
	var err error
	log.Info("DbToYang_acl_set_name_xfmr: ", inParams.key)
	/*name attribute corresponds to key in redis table*/
	aclName, _ := getOCAclKeysFromStrDBKey(inParams.key)
	res_map["name"] = aclName
	log.Info("acl-set/config/name  ", res_map)
	return res_map, err
}

var YangToDb_acl_entry_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
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

var DbToYang_acl_entry_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
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
	rmap["sequence-id"], _ = strconv.ParseFloat(seqId, 64)
	return rmap, err
}

var YangToDb_acl_entry_sequenceid_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	log.Info("YangToDb_acl_entry_sequenceid_xfmr: ")
	/*no-op since there is no redis table field to be filled corresponding to sequenec-id attribute since its part of key */
	return res_map, err
}

var DbToYang_acl_entry_sequenceid_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
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

var YangToDb_acl_l2_ethertype_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error

	if inParams.param == nil {
	    res_map["ETHER_TYPE"] = ""
	    return res_map, err
	}
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

var DbToYang_acl_l2_ethertype_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_acl_l2_ethertype_xfmr", data, inParams.ygRoot)
	if _, ok := data[RULE_TABLE]; !ok {
		err = errors.New("RULE_TABLE entry not found in the input param")
		return result, err
	}

	ruleTbl := data[RULE_TABLE]
	ruleInst := ruleTbl[inParams.key]
	etype, ok := ruleInst.Field["ETHER_TYPE"]

	if ok {
		etypeVal, _ := strconv.ParseUint(strings.Replace(etype, "0x", "", -1), 16, 32)
		result["protocol"] = getL2EtherType(etypeVal)
	} else {
		err = errors.New("ETHER_TYPE field not found in DB")
	}
	return result, nil
}

var YangToDb_acl_ip_protocol_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error

	if inParams.param == nil {
	    res_map["IP_PROTOCOL"] = ""
	    return res_map, err
	}
	protocolType := reflect.TypeOf(inParams.param).Elem()
	log.Info("YangToDb_acl_ip_protocol_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " protocolType: ", protocolType)
	switch protocolType {
	case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL{}):
		v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
		res_map["IP_PROTOCOL"] = findInMap(IP_PROTOCOL_MAP, strconv.FormatInt(int64(v.E_OpenconfigPacketMatchTypes_IP_PROTOCOL), 10))
		v = nil
		break
	case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8{}):
		v := (inParams.param).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8)
		res_map["IP_PROTOCOL"] = strconv.FormatInt(int64(v.Uint8), 10)
		break
	}
	return res_map, err
}

var DbToYang_acl_ip_protocol_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_acl_ip_protocol_xfmr", data, inParams.ygRoot)
	oc_protocol := findByValue(IP_PROTOCOL_MAP, data[RULE_TABLE][inParams.key].Field["IP_PROTOCOL"])
	n, err := strconv.ParseInt(oc_protocol, 10, 64)
	result["protocol"] = ocbinds.E_OpenconfigPacketMatchTypes_IP_PROTOCOL(n).ΛMap()["E_OpenconfigPacketMatchTypes_IP_PROTOCOL"][n].Name
	return result, err
}

var YangToDb_acl_source_port_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	if inParams.param == nil {
	    res_map["L4_SRC_PORT"] = ""
	    return res_map, err
	}
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

var DbToYang_acl_source_port_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_acl_source_port_xfmr: ", data, inParams.ygRoot)
	result := make(map[string]interface{})
	if _, ok := data[RULE_TABLE]; !ok {
		err = errors.New("RULE_TABLE entry not found in the input param")
		return result, err
	}
	ruleTbl := data[RULE_TABLE]
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
}

var YangToDb_acl_destination_port_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	if inParams.param == nil {
	    res_map["L4_DST_PORT_RANGE"] = ""
	    return res_map, err
	}
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

var DbToYang_acl_destination_port_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_acl_destination_port_xfmr: ", data, inParams.ygRoot)
	if _, ok := data[RULE_TABLE]; !ok {
		err = errors.New("RULE_TABLE entry not found in the input param")
		return result, err
	}
	ruleTbl := data[RULE_TABLE]
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
}

var YangToDb_acl_tcp_flags_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	log.Info("YangToDb_acl_tcp_flags_xfmr: ")
	var tcpFlags uint32 = 0x00
	var b bytes.Buffer
	if inParams.param == nil {
	    res_map["TCP_FLAGS"] = b.String()
	    return res_map, err
	}
	log.Info("YangToDb_acl_tcp_flags_xfmr: ", inParams.ygRoot, inParams.uri)
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
	fmt.Fprintf(&b, "0x%0.2x/0x%0.2x", tcpFlags, tcpFlags)
	res_map["TCP_FLAGS"] = b.String()
	return res_map, err
}

var DbToYang_acl_tcp_flags_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_acl_tcp_flags_xfmr: ", data, inParams.ygRoot)
	result := make(map[string]interface{})
	if _, ok := data[RULE_TABLE]; !ok {
		err = errors.New("RULE_TABLE entry not found in the input param")
		return result, err
	}
	ruleTbl := data[RULE_TABLE]
	ruleInst := ruleTbl[inParams.key]
	tcpFlag, ok := ruleInst.Field["TCP_FLAGS"]
	if ok {
		result["tcp-flags"] = getTransportConfigTcpFlags(tcpFlag)
		return result, nil
	}
	return result, nil
}

var YangToDb_acl_port_bindings_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
	var err error
	res_map := make(map[string]map[string]db.Value)
	aclTableMap := make(map[string]db.Value)
	log.Info("YangToDb_acl_port_bindings_xfmr: ", inParams.ygRoot, inParams.uri)

	aclObj := getAclRoot(inParams.ygRoot)
	if aclObj.Interfaces == nil {
		return res_map, err
	}
	//aclset := &ocbinds.OpenconfigAcl_Acl_AclSets_AclSet{}
	aclInterfacesMap := make(map[string][]string)
	for intfId, _ := range aclObj.Interfaces.Interface {
		intf := aclObj.Interfaces.Interface[intfId]
		if intf != nil {
			if intf.IngressAclSets != nil && len(intf.IngressAclSets.IngressAclSet) > 0 {
				for inAclKey, _ := range intf.IngressAclSets.IngressAclSet {
					aclName := getAclKeyStrFromOCKey(inAclKey.SetName, inAclKey.Type)
					aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.Id)
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
					aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.Id)
					_, ok := aclTableMap[aclName]
					if !ok {
						aclTableMap[aclName] = db.Value{Field: make(map[string]string)}
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
	res_map[ACL_TABLE] = aclTableMap
	return res_map, err
}

var DbToYang_acl_port_bindings_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
	var err error
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_acl_port_bindings_xfmr: ", data, inParams.ygRoot)

	aclTbl := data["ACL_TABLE"]
	var ruleTbl map[string]map[string]db.Value

	// repoluate to use existing code
	ruleTbl = make(map[string]map[string]db.Value)
	for key, element := range data["ACL_RULE"] {
		// split into aclKey and ruleKey
		tokens := strings.Split(key, "|")
		if ruleTbl[tokens[0]] == nil {
			ruleTbl[tokens[0]] = make(map[string]db.Value)
		}
		ruleTbl[tokens[0]][tokens[1]] = db.Value{Field: make(map[string]string)}
		ruleTbl[tokens[0]][tokens[1]] = element
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
						_, ok := ruleTableMap[aclKey+"|"+rulekey]
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
						_, ok := ruleTableMap[aclKey+"|"+rulekey]
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
