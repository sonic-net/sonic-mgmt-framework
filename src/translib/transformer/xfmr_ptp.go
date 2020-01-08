////////////////////////////////////////////////////////////////////////////////
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
	log "github.com/golang/glog"
	"errors"
	"strings"
	"strconv"
	"regexp"
	"net"
	"github.com/openconfig/ygot/ygot"
	"translib/ocbinds"
	"fmt"
	"path/filepath"
	b64 "encoding/base64"
	"translib/db"
    "translib/tlerr"
)

func init() {
	XlateFuncBind("YangToDb_ptp_entry_key_xfmr", YangToDb_ptp_entry_key_xfmr)
	XlateFuncBind("DbToYang_ptp_entry_key_xfmr", DbToYang_ptp_entry_key_xfmr)
	XlateFuncBind("YangToDb_ptp_port_entry_key_xfmr", YangToDb_ptp_port_entry_key_xfmr)
	XlateFuncBind("DbToYang_ptp_port_entry_key_xfmr", DbToYang_ptp_port_entry_key_xfmr)
	XlateFuncBind("YangToDb_ptp_global_key_xfmr", YangToDb_ptp_global_key_xfmr)
	XlateFuncBind("DbToYang_ptp_global_key_xfmr", DbToYang_ptp_global_key_xfmr)

	XlateFuncBind("YangToDb_ptp_tcport_entry_key_xfmr", YangToDb_ptp_tcport_entry_key_xfmr)
	XlateFuncBind("DbToYang_ptp_tcport_entry_key_xfmr", DbToYang_ptp_tcport_entry_key_xfmr)
	XlateFuncBind("YangToDb_ptp_clock_identity_xfmr", YangToDb_ptp_clock_identity_xfmr)
	XlateFuncBind("DbToYang_ptp_clock_identity_xfmr", DbToYang_ptp_clock_identity_xfmr)
	XlateFuncBind("YangToDb_ptp_boolean_xfmr", YangToDb_ptp_boolean_xfmr)
	XlateFuncBind("DbToYang_ptp_boolean_xfmr", DbToYang_ptp_boolean_xfmr)
	XlateFuncBind("YangToDb_ptp_delay_mech_xfmr", YangToDb_ptp_delay_mech_xfmr)
	XlateFuncBind("DbToYang_ptp_delay_mech_xfmr", DbToYang_ptp_delay_mech_xfmr)
	XlateFuncBind("YangToDb_ptp_port_state_xfmr", YangToDb_ptp_port_state_xfmr)
	XlateFuncBind("DbToYang_ptp_port_state_xfmr", DbToYang_ptp_port_state_xfmr)
	XlateFuncBind("YangToDb_ptp_inst_number_xfmr:", YangToDb_ptp_inst_number_xfmr);
	XlateFuncBind("DbToYang_ptp_inst_number_xfmr:", DbToYang_ptp_inst_number_xfmr);

	XlateFuncBind("YangToDb_ptp_network_transport_xfmr", YangToDb_ptp_network_transport_xfmr )
	XlateFuncBind("DbToYang_ptp_network_transport_xfmr", DbToYang_ptp_network_transport_xfmr )
	XlateFuncBind("YangToDb_ptp_domain_number_xfmr", YangToDb_ptp_domain_number_xfmr )
	XlateFuncBind("DbToYang_ptp_domain_number_xfmr", DbToYang_ptp_domain_number_xfmr )
	XlateFuncBind("YangToDb_ptp_clock_type_xfmr", YangToDb_ptp_clock_type_xfmr )
	XlateFuncBind("DbToYang_ptp_clock_type_xfmr", DbToYang_ptp_clock_type_xfmr )
	XlateFuncBind("YangToDb_ptp_domain_profile_xfmr", YangToDb_ptp_domain_profile_xfmr )
	XlateFuncBind("DbToYang_ptp_domain_profile_xfmr", DbToYang_ptp_domain_profile_xfmr )
	XlateFuncBind("YangToDb_ptp_unicast_multicast_xfmr", YangToDb_ptp_unicast_multicast_xfmr )
	XlateFuncBind("DbToYang_ptp_unicast_multicast_xfmr", DbToYang_ptp_unicast_multicast_xfmr )
	XlateFuncBind("YangToDb_ptp_unicast_table_xfmr", YangToDb_ptp_unicast_table_xfmr )
	XlateFuncBind("DbToYang_ptp_unicast_table_xfmr", DbToYang_ptp_unicast_table_xfmr )
	XlateFuncBind("YangToDb_ptp_udp6_scope_xfmr", YangToDb_ptp_udp6_scope_xfmr )
	XlateFuncBind("DbToYang_ptp_udp6_scope_xfmr", DbToYang_ptp_udp6_scope_xfmr )
}

/* E_IETFPtp_DelayMechanismEnumeration */
var PTP_DELAY_MECH_MAP = map[string]string{
	strconv.FormatInt(int64(ocbinds.IETFPtp_DelayMechanismEnumeration_e2e), 10): "E2E",
	strconv.FormatInt(int64(ocbinds.IETFPtp_DelayMechanismEnumeration_p2p), 10): "P2P",
	strconv.FormatInt(int64(ocbinds.IETFPtp_DelayMechanismEnumeration_UNSET), 10): "Auto",
}

type ptp_id_bin [8]byte

type E_Ptp_AddressTypeEnumeration int64
const (
    PTP_ADDRESSTYPE_UNKNOWN E_Ptp_AddressTypeEnumeration = 0
    PTP_ADDRESSTYPE_IP_IPV4 E_Ptp_AddressTypeEnumeration = 2
    PTP_ADDRESSTYPE_IP_IPV6 E_Ptp_AddressTypeEnumeration = 3
    PTP_ADDRESSTYPE_IP_MAC E_Ptp_AddressTypeEnumeration = 4
)

func check_address(address string) E_Ptp_AddressTypeEnumeration {
	trial := net.ParseIP(address)
	if (trial != nil) {
		if trial.To4() != nil {
			return PTP_ADDRESSTYPE_IP_IPV4
		}
		if strings.Contains(address, ":") {
			return PTP_ADDRESSTYPE_IP_IPV6
		}
	} else {
		matched, _ := regexp.Match(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`, []byte(address))
		if (matched) {
			return PTP_ADDRESSTYPE_IP_MAC
		}
		return PTP_ADDRESSTYPE_UNKNOWN
	}
	return PTP_ADDRESSTYPE_UNKNOWN
}

// ParseIdentity parses an s with the following format
// 010203.0405.060708
func ParseIdentity(s string) (ptp_id ptp_id_bin, err error) {
	if len(s) < 18 {
		return ptp_id, fmt.Errorf("Invalid input identity string %s", s)
	}
	fmt.Sscanf(s, "%02x%02x%02x.%02x%02x.%02x%02x%02x", &ptp_id[0], &ptp_id[1], &ptp_id[2], &ptp_id[3], &ptp_id[4], &ptp_id[5], &ptp_id[6], &ptp_id[7])
	return ptp_id, err
}

////////////////////////////////////////////
// Bi-directoonal overloaded methods
////////////////////////////////////////////
var YangToDb_ptp_entry_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	var entry_key string
	var err error
	log.Info("YangToDb_ptp_entry_key_xfmr: ", inParams.ygRoot, " XPath ", inParams.uri, " key: ", inParams.key)
	pathInfo := NewPathInfo(inParams.uri)
	log.Info("YangToDb_ptp_entry_key_xfmr len(pathInfo.Vars): ", len(pathInfo.Vars))
	if len(pathInfo.Vars) < 1 {
		err = errors.New("Invalid xpath, key attributes not found")
		return entry_key, err
	}

	inkey,_ := strconv.ParseUint(pathInfo.Var("instance-number"), 10, 64)
	if inkey > 0 {
		err = errors.New("Invalid input instance-number")
		return entry_key, err
	}

	entry_key = "GLOBAL"

	log.Info("YangToDb_ptp_entry_key_xfmr - entry_key : ", entry_key)

	return entry_key, err
}

var DbToYang_ptp_entry_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	var err error
	log.Info("DbToYang_ptp_entry_key_xfmr root, uri: ", inParams.ygRoot, inParams.uri)
	// rmap["instance-number"] = 0
	return rmap, err
}

var YangToDb_ptp_port_entry_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	var entry_key string
	var underlying_interface string
	var err error
	log.Info("YangToDb_ptp_port_entry_key_xfmr root, uri: ", inParams.ygRoot, inParams.uri)
	pathInfo := NewPathInfo(inParams.uri)

	log.Info("YangToDb_ptp_port_entry_key_xfmr len(pathInfo.Vars): ", len(pathInfo.Vars))
	log.Info("YangToDb_ptp_port_entry_key_xfmr pathInfo.Vars: ", pathInfo.Vars)
	if len(pathInfo.Vars) < 2 {
		err = errors.New("Invalid xpath, key attributes not found")
		return entry_key, err
	}

	instance_id, _  := strconv.ParseUint(pathInfo.Var("instance-number"), 10, 64)
	port_number_str  := pathInfo.Var("port-number")
	port_number, _  := strconv.ParseUint(port_number_str, 10, 64)
	ptpObj := getPtpRoot(inParams.ygRoot)
	pDsList := ptpObj.InstanceList[uint32(instance_id)].PortDsList
	log.Info("YangToDb_ptp_port_entry_key_xfmr len(pDsList) : ", len(pDsList))

	if (0 != len(pDsList) && nil != pDsList[uint16(port_number)].UnderlyingInterface &&
		"" != *pDsList[uint16(port_number)].UnderlyingInterface) {
		underlying_interface = *pDsList[uint16(port_number)].UnderlyingInterface
		log.Info("YangToDb_ptp_port_entry_key_xfmr underlying-interface: ", underlying_interface)
		// if (underlying_interface == "") {
			// underlying_interface = "Ethernet" + port_number_str
			// log.Info("YangToDb_ptp_port_entry_key_xfmr : underlying-interface is required on create")
			// return entry_key, errors.New("underlying-interface is required on create")
		// }
		if (port_number < 1000) {
			if (port_number_str != strings.Replace(underlying_interface, "Ethernet", "", 1)) {
				log.Info("YangToDb_ptp_port_entry_key_xfmr : underlying-interface port-number mismatch")
				return entry_key, errors.New("underlying-interface port-number mismatch")
			}
		} else {
			if (strconv.FormatInt(int64(port_number-1000), 10) != strings.Replace(underlying_interface, "Vlan", "", 1)) {
				log.Info("YangToDb_ptp_port_entry_key_xfmr : underlying-interface port-number mismatch")
				return entry_key, errors.New("underlying-interface port-number mismatch")
			}
		}
	} else {
		if (port_number < 1000) {
			underlying_interface = "Ethernet" + port_number_str
		} else {
			underlying_interface = "Vlan" + strconv.FormatInt(int64(port_number-1000), 10)
		}
	}

	log.Info("YangToDb_ptp_port_entry_key_xfmr pathInfo.Var:port-number: ", pathInfo.Var("port-number"))
	entry_key = "GLOBAL|" + underlying_interface


	log.Info("YangToDb_ptp_port_entry_key_xfmr - entry_key : ", entry_key)

	return entry_key, err
}

var DbToYang_ptp_port_entry_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	var err error
	var port_num string
	var is_vlan bool

	log.Info("DbToYang_ptp_port_entry_key_xfmr root, uri: ", inParams.ygRoot, inParams.uri)
	entry_key := inParams.key
	log.Info("DbToYang_ptp_port_entry_key_xfmr: ", entry_key)

	portName := entry_key
	if (strings.Contains(portName, "Ethernet")) {
		port_num = strings.Replace(portName, "GLOBAL|Ethernet", "", 1)
		is_vlan = false
	} else {
		port_num = strings.Replace(portName, "GLOBAL|Vlan", "", 1)
		is_vlan = true
	}
	// rmap["instance-number"] = 0
	port_num_int, _ := strconv.ParseInt(port_num, 10, 16)
	if (is_vlan) {
		port_num_int += 1000
	}
	rmap["port-number"] = port_num_int
	log.Info("DbToYang_ptp_port_entry_key_xfmr port-number: ", port_num)
	return rmap, err
}

var YangToDb_ptp_global_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	var entry_key string
	var err error
	log.Info("YangToDb_ptp_global_key_xfmr: ", inParams.ygRoot, inParams.uri)

	entry_key = "GLOBAL"

	log.Info("YangToDb_ptp_global_key_xfmr - entry_key : ", entry_key)

	return entry_key, err
}

var DbToYang_ptp_global_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	var err error
	log.Info("DbToYang_ptp_global_key_xfmr root, uri: ", inParams.ygRoot, inParams.uri)
	rmap["instance-number"] = 0
	return rmap, err
}

var YangToDb_ptp_tcport_entry_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	var entry_key string
	var err error
	log.Info("YangToDb_ptp_tcport_entry_key_xfmr root, uri: ", inParams.ygRoot, inParams.uri)
	pathInfo := NewPathInfo(inParams.uri)

	log.Info("YangToDb_ptp_tcport_entry_key_xfmr len(pathInfo.Vars): ", len(pathInfo.Vars))
	if len(pathInfo.Vars) < 1 {
		err = errors.New("Invalid xpath, key attributes not found")
		return entry_key, err
	}

	log.Info("YangToDb_ptp_tcport_entry_key_xfmr pathInfo.Var:port-number: ", pathInfo.Var("port-number"))
	entry_key = "Ethernet" + pathInfo.Var("port-number")


	log.Info("YangToDb_ptp_tcport_entry_key_xfmr - entry_key : ", entry_key)

	return entry_key, err
}

var DbToYang_ptp_tcport_entry_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	var err error
	log.Info("DbToYang_ptp_tcport_entry_key_xfmr root, uri: ", inParams.ygRoot, inParams.uri)

	entry_key := inParams.key
	log.Info("DbToYang_ptp_tcport_entry_key_xfmr: ", entry_key)

	portName := entry_key
	port_num := strings.Replace(portName, "Ethernet", "", 1)
	rmap["port-number"], _ = strconv.ParseInt(port_num, 10, 16)
	log.Info("DbToYang_ptp_tcport_entry_key_xfmr port-number: ", port_num)
	return rmap, err
}

func getPtpRoot(s *ygot.GoStruct) *ocbinds.IETFPtp_Ptp {
	deviceObj := (*s).(*ocbinds.Device)
	return deviceObj.Ptp
}

var YangToDb_ptp_clock_identity_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	var field string
	var identity []byte
	if inParams.param == nil {
		log.Info("YangToDb_ptp_clock_identity_xfmr Error: ")
	    return res_map, err
	}
	log.Info("YangToDb_ptp_clock_identity_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_clock_identity_xfmr inParams.key: ", inParams.key)

	ptpObj := getPtpRoot(inParams.ygRoot)

	if strings.Contains(inParams.uri, "grandmaster-identity") {
		identity = ptpObj.InstanceList[0].ParentDs.GrandmasterIdentity
		field = "grandmaster-identity"
	} else if strings.Contains(inParams.uri, "parent-port-identity") {
		identity = ptpObj.InstanceList[0].ParentDs.ParentPortIdentity.ClockIdentity
		field = "clock-identity"
	// } else if strings.Contains(inParams.uri, "transparent-clock-default-ds") {
		// identity = ptpObj.TransparentClockDefaultDs.ClockIdentity
		// field = "clock-identity"
	} else if strings.Contains(inParams.uri, "default-ds") {
		identity = ptpObj.InstanceList[0].DefaultDs.ClockIdentity
		field = "clock-identity"
	}


	enc := fmt.Sprintf("%02x%02x%02x.%02x%02x.%02x%02x%02x",
		identity[0], identity[1], identity[2], identity[3], identity[4], identity[5], identity[6], identity[7])

	log.Info("YangToDb_ptp_clock_identity_xfmr enc: ", enc, " field: ", field)
	res_map[field] = enc
	return res_map, err
}

var DbToYang_ptp_clock_identity_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	var ptp_id ptp_id_bin
	var field,identity,sEnc string
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_clock_identity_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)

	if strings.Contains(inParams.uri, "grandmaster-identity") {
		field = "grandmaster-identity"
		identity = data["PTP_PARENTDS"][inParams.key].Field[field]
	} else if strings.Contains(inParams.uri, "parent-port-identity") {
		field = "clock-identity"
		identity = data["PTP_PARENTDS"][inParams.key].Field[field]
	} else if strings.Contains(inParams.uri, "transparent-clock-default-ds") {
		field = "clock-identity"
		identity = data["PTP_TC_CLOCK"][inParams.key].Field[field]
	} else if strings.Contains(inParams.uri, "default-ds") {
		field = "clock-identity"
		identity = data["PTP_CLOCK"][inParams.key].Field[field]
	}
	if len(identity) >= 18 {
		ptp_id,err = ParseIdentity(identity)
		sEnc = b64.StdEncoding.EncodeToString(ptp_id[:])
		result[field] = sEnc
	} else {
		sEnc = ""
	}

	return result, err
}

var YangToDb_ptp_boolean_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	var outval string
	if inParams.param == nil {
		log.Info("YangToDb_ptp_boolean_xfmr Error: ")
	    return res_map, err
	}
	log.Info("YangToDb_ptp_boolean_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_boolean_xfmr inParams.key: ", inParams.key)
	log.Info("YangToDb_ptp_boolean_xfmr inParams.curDb: ", inParams.curDb)

	inval, _ := inParams.param.(*bool)
	_, field := filepath.Split(inParams.uri)
	log.Info("YangToDb_ptp_boolean_xfmr inval: ", *inval, " field: ", field)

	if (*inval) {
		outval = "1"
	} else {
		outval = "0"
	}

	log.Info("YangToDb_ptp_boolean_xfmr outval: ", outval, " field: ", field)
	res_map[field] = outval
	return res_map, err
}

var DbToYang_ptp_boolean_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	var inval string
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_boolean_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)

	_, field := filepath.Split(inParams.uri)
	if field == "two-step-flag" {
		inval = data["PTP_CLOCK"][inParams.key].Field[field]
	} else if field == "slave-only" {
		inval = data["PTP_CLOCK"][inParams.key].Field[field]
	} else if field == "parent-stats" {
		inval = data["PTP_PARENTDS"][inParams.key].Field[field]
	} else if field == "current-utc-offset-valid" {
		inval = data["PTP_TIMEPROPDS"][inParams.key].Field[field]
	} else if field == "leap59" {
		inval = data["PTP_TIMEPROPDS"][inParams.key].Field[field]
	} else if field == "leap61" {
		inval = data["PTP_TIMEPROPDS"][inParams.key].Field[field]
	} else if field == "time-traceable" {
		inval = data["PTP_TIMEPROPDS"][inParams.key].Field[field]
	} else if field == "frequency-traceable" {
		inval = data["PTP_TIMEPROPDS"][inParams.key].Field[field]
	} else if field == "ptp-timescale" {
		inval = data["PTP_TIMEPROPDS"][inParams.key].Field[field]
	} else if field == "faulty-flag" {
		inval = data["PTP_TC_PORT"][inParams.key].Field[field]
	}

	if (inval == "0") {
		result[field] = false
	} else if inval == "1" {
		result[field] = true
	}

	return result, err
}

var YangToDb_ptp_delay_mech_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	var outval string
	if inParams.param == nil {
		log.Info("YangToDb_ptp_delay_mech_xfmr Error: ")
	    return res_map, err
	}
	log.Info("YangToDb_ptp_delay_mech_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_delay_mech_xfmr inParams.key: ", inParams.key)

	inval, _ := inParams.param.(ocbinds.E_IETFPtp_DelayMechanismEnumeration)
	_, field := filepath.Split(inParams.uri)

	log.Info("YangToDb_ptp_delay_mech_xfmr outval: ", outval, " field: ", field)
	res_map[field] = findInMap(PTP_DELAY_MECH_MAP, strconv.FormatInt(int64(inval), 10))

	return res_map, err
}

var DbToYang_ptp_delay_mech_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	var inval string
	var outval string
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_delay_mech_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)

	_, field := filepath.Split(inParams.uri)

	if strings.Contains(inParams.uri, "port-ds-list") {
		inval = data["PTP_PORT"][inParams.key].Field[field]
	} else if strings.Contains(inParams.uri, "transparent-clock-default-ds") {
		inval = data["PTP_TC_CLOCK"][inParams.key].Field[field]
	}

	switch inval {
	case "E2E":
		outval = "e2e"
	case "P2P":
		outval = "p2p"
	default:
		outval = ""
	}
	log.Info("DbToYang_ptp_delay_mech_xfmr result: ", outval, " inval: ", inval)
	if outval != "" {
		result[field] = outval
	}

	return result, err
}

var YangToDb_ptp_port_state_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	var outval string
	if inParams.param == nil {
		log.Info("YangToDb_ptp_port_state_xfmr Error: ")
	    return res_map, err
	}
	log.Info("YangToDb_ptp_port_state_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_port_state_xfmr inParams.key: ", inParams.key)

	inval, _ := inParams.param.(ocbinds.E_IETFPtp_PortStateEnumeration)
	_, field := filepath.Split(inParams.uri)

	switch inval {
	case ocbinds.IETFPtp_PortStateEnumeration_initializing:
		outval = "1"
	case ocbinds.IETFPtp_PortStateEnumeration_faulty:
		outval = "2"
	case ocbinds.IETFPtp_PortStateEnumeration_disabled:
		outval = "3"
	case ocbinds.IETFPtp_PortStateEnumeration_listening:
		outval = "4"
	case ocbinds.IETFPtp_PortStateEnumeration_pre_master:
		outval = "5"
	case ocbinds.IETFPtp_PortStateEnumeration_master:
		outval = "6"
	case ocbinds.IETFPtp_PortStateEnumeration_passive:
		outval = "7"
	case ocbinds.IETFPtp_PortStateEnumeration_uncalibrated:
		outval = "8"
	case ocbinds.IETFPtp_PortStateEnumeration_slave:
		outval = "9"
	}

	log.Info("YangToDb_ptp_port_state_xfmr outval: ", outval, " field: ", field)
	res_map[field] = outval
	return res_map, err
}

var DbToYang_ptp_port_state_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	var inval string
	var outval string
	result := make(map[string]interface{})

	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_port_state_xfmr :", data, inParams.ygRoot)

	inval = data["PTP_PORT"][inParams.key].Field["port-state"]
	switch inval {
	case "1":
		outval = "initializing"
	case "2":
		outval = "faulty"
	case "3":
		outval = "disabled"
	case "4":
		outval = "listening"
	case "5":
		outval = "pre-master"
	case "6":
		outval = "master"
	case "7":
		outval = "passive"
	case "8":
		outval = "uncalibrated"
	case "9":
		outval = "slave"
	default:
		goto done
	}
	result["port-state"] = outval
done:
	return result, err
}

var YangToDb_ptp_inst_number_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map :=  make(map[string]string)
    res_map["NULL"] = "NULL"
    return res_map, nil
}


var DbToYang_ptp_inst_number_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	/* do nothing */
	var err error
	result := make(map[string]interface{})
	return result, err
}


var YangToDb_ptp_network_transport_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map :=  make(map[string]string)
	var outval string
	var err error
	if inParams.param == nil {
		log.Info("YangToDb_ptp_network_transport_xfmr Error: ")
		return res_map, err
	}
	log.Info("YangToDb_ptp_network_transport_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_network_transport_xfmr inParams.key: ", inParams.key)

	pathInfo := NewPathInfo(inParams.uri)
	instance_id, _  := strconv.ParseUint(pathInfo.Var("instance-number"), 10, 64)
	log.Info("YangToDb_ptp_network_transport_xfmr instance_number : ", instance_id)

	ptpObj := getPtpRoot(inParams.ygRoot)
	outval = *ptpObj.InstanceList[uint32(instance_id)].DefaultDs.NetworkTransport
	log.Info("YangToDb_ptp_network_transport_xfmr outval: ", outval)
	_, field := filepath.Split(inParams.uri)
	domain_profile := ""

	ts := db.TableSpec { Name: "PTP_CLOCK" }
	ca := make([]string, 1, 1)

	ca[0] = "GLOBAL"
	akey := db.Key { Comp: ca}
	entry, err := inParams.d.GetEntry(&ts, akey)
	if entry.Has("domain-profile") {
		domain_profile = entry.Get("domain-profile")
	}

	log.Info("YangToDb_ptp_network_transport_xfmr domain_profile : ", domain_profile)

	if outval == "L2" && domain_profile == "G.8275.x" {
		return res_map, tlerr.InvalidArgsError{Format:"L2 not supported with G.8275.2"}
	}

	log.Info("YangToDb_ptp_network_transport_xfmr outval: ", outval, " field: ", field)
	res_map[field] = outval
    return res_map, nil
}

var DbToYang_ptp_network_transport_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})

	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_network_transport_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)
	log.Info("DbToYang_ptp_network_transport_xfmr inParams.key: ", inParams.key)

	_, field := filepath.Split(inParams.uri)
	log.Info("DbToYang_ptp_network_transport_xfmr field: ", field)
	value := data["PTP_CLOCK"][inParams.key].Field[field]
	result[field] = value
	log.Info("DbToYang_ptp_network_transport_xfmr value: ", value)
	return result, err
}

var YangToDb_ptp_domain_number_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map :=  make(map[string]string)
	var outval uint8
	var err error
	if inParams.param == nil {
		log.Info("YangToDb_ptp_domain_number_xfmr Error: ")
		return res_map, err
	}
	log.Info("YangToDb_ptp_domain_number_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_domain_number_xfmr inParams.key: ", inParams.key)

	pathInfo := NewPathInfo(inParams.uri)
	instance_id, _  := strconv.ParseUint(pathInfo.Var("instance-number"), 10, 64)
	log.Info("YangToDb_ptp_domain_number_xfmr instance_number : ", instance_id)

	ptpObj := getPtpRoot(inParams.ygRoot)
	outval = *ptpObj.InstanceList[uint32(instance_id)].DefaultDs.DomainNumber
	log.Info("YangToDb_ptp_domain_number_xfmr outval: ", outval)
	_, field := filepath.Split(inParams.uri)
	domain_profile := ""

	ts := db.TableSpec { Name: "PTP_CLOCK" }
	ca := make([]string, 1, 1)

	ca[0] = "GLOBAL"
	akey := db.Key { Comp: ca}
	entry, err := inParams.d.GetEntry(&ts, akey)
	if entry.Has("domain-profile") {
		domain_profile = entry.Get("domain-profile")
	}

	log.Info("YangToDb_ptp_domain_number_xfmr domain_profile : ", domain_profile)

    if domain_profile == "G.8275.x" {
		if outval < 44 || outval > 63 {
			return res_map, tlerr.InvalidArgsError{Format:"domain must be in range 44-63 with G.8275.2"}
		}
	}

	log.Info("YangToDb_ptp_domain_number_xfmr outval: ", outval, " field: ", field)
	res_map[field] = strconv.FormatInt(int64(outval), 10)
    return res_map, nil
}

var DbToYang_ptp_domain_number_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})

	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_domain_number_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)
	log.Info("DbToYang_ptp_domain_number_xfmr inParams.key: ", inParams.key)

	_, field := filepath.Split(inParams.uri)
	log.Info("DbToYang_ptp_domain_number_xfmr field: ", field)
	value := data["PTP_CLOCK"][inParams.key].Field[field]
	result[field],_ = strconv.ParseUint(value, 10, 64)
	log.Info("DbToYang_ptp_domain_number_xfmr value: ", value)
	return result, err
}

var YangToDb_ptp_clock_type_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map :=  make(map[string]string)
	var outval string
	var err error
	if inParams.param == nil {
		log.Info("YangToDb_ptp_clock_type_xfmr Error: ")
		return res_map, err
	}
	log.Info("YangToDb_ptp_clock_type_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_clock_type_xfmr inParams.key: ", inParams.key)

	pathInfo := NewPathInfo(inParams.uri)
	instance_id, _  := strconv.ParseUint(pathInfo.Var("instance-number"), 10, 64)
	log.Info("YangToDb_ptp_clock_type_xfmr instance_number : ", instance_id)

	ptpObj := getPtpRoot(inParams.ygRoot)
	outval = *ptpObj.InstanceList[uint32(instance_id)].DefaultDs.ClockType

	if outval == "P2P_TC" {
		return res_map, tlerr.InvalidArgsError{Format:"peer-to-peer-transparent-clock is not supported"}
	}

	log.Info("YangToDb_ptp_clock_type_xfmr outval: ", outval)
	_, field := filepath.Split(inParams.uri)
	domain_profile := ""
	network_transport := ""
	unicast_multicast := ""

	ts := db.TableSpec { Name: "PTP_CLOCK" }
	ca := make([]string, 1, 1)

	ca[0] = "GLOBAL"
	akey := db.Key { Comp: ca}
	entry, err := inParams.d.GetEntry(&ts, akey)
	if entry.Has("domain-profile") {
		domain_profile = entry.Get("domain-profile")
	}
	if entry.Has("network-transport") {
		network_transport = entry.Get("network-transport")
	}
	if entry.Has("unicast-multicast") {
		unicast_multicast = entry.Get("unicast-multicast")
	}

	log.Info("YangToDb_ptp_clock_type_xfmr domain_profile : ", domain_profile, " network-transport : ", network_transport,
		" unicast-multicast : ", unicast_multicast)

	if outval == "P2P_TC" || outval == "E2E_TC" {
		if domain_profile == "G.8275.x" {
			return res_map, tlerr.InvalidArgsError{Format:"transparent-clock not supported with G.8275.2"}
		}
		if domain_profile == "ieee1588" && unicast_multicast == "unicast" {
			return res_map, tlerr.InvalidArgsError{Format:"transparent-clock not supported with default profile and unicast"}
		}
	}
	if outval == "BC" {
		if domain_profile == "G.8275.x" && network_transport == "L2" {
			return res_map, tlerr.InvalidArgsError{Format:"boundary-clock not supported with G.8275.2 and L2"}
		}
		if domain_profile == "G.8275.x" && unicast_multicast == "multicast" {
			return res_map, tlerr.InvalidArgsError{Format:"boundary-clock not supported with G.8275.2 and multicast"}
		}
		if domain_profile == "G.8275.x" && network_transport == "UDPv6" {
			return res_map, tlerr.InvalidArgsError{Format:"boundary-clock not supported with G.8275.2 and ipv6"}
		}
	}

	log.Info("YangToDb_ptp_clock_type_xfmr outval: ", outval, " field: ", field)
	res_map[field] = outval
    return res_map, nil
}

var DbToYang_ptp_clock_type_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})

	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_clock_type_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)
	log.Info("DbToYang_ptp_clock_type_xfmr inParams.key: ", inParams.key)

	_, field := filepath.Split(inParams.uri)
	log.Info("DbToYang_ptp_clock_type_xfmr field: ", field)
	value := data["PTP_CLOCK"][inParams.key].Field[field]
	result[field] = value
	log.Info("DbToYang_ptp_clock_type_xfmr value: ", value)
	return result, err
}

var YangToDb_ptp_domain_profile_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map :=  make(map[string]string)
	var outval string
	var err error
	if inParams.param == nil {
		log.Info("YangToDb_ptp_domain_profile_xfmr Error: ")
		return res_map, err
	}
	log.Info("YangToDb_ptp_domain_profile_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_domain_profile_xfmr inParams.key: ", inParams.key)

	pathInfo := NewPathInfo(inParams.uri)
	instance_id, _  := strconv.ParseUint(pathInfo.Var("instance-number"), 10, 64)
	log.Info("YangToDb_ptp_domain_profile_xfmr instance_number : ", instance_id)

	ptpObj := getPtpRoot(inParams.ygRoot)
	outval = *ptpObj.InstanceList[uint32(instance_id)].DefaultDs.DomainProfile

	if outval == "G.8275.1" {
		return res_map, tlerr.InvalidArgsError{Format:"g8275.1 is not supported"}
	}
	if outval == "G.8275.2" {
		outval = "G.8275.x"
	}

	log.Info("YangToDb_ptp_domain_profile_xfmr outval: ", outval)
	_, field := filepath.Split(inParams.uri)
	var domain_number uint64
	network_transport := ""
	unicast_multicast := ""
	clock_type := ""

	ts := db.TableSpec { Name: "PTP_CLOCK" }
	ca := make([]string, 1, 1)

	ca[0] = "GLOBAL"
	akey := db.Key { Comp: ca}
	entry, err := inParams.d.GetEntry(&ts, akey)
	if entry.Has("domain-number") {
		domain_number, _ = strconv.ParseUint(entry.Get("domain-number"), 10, 64)
	}
	if entry.Has("network-transport") {
		network_transport = entry.Get("network-transport")
	}
	if entry.Has("unicast-multicast") {
		unicast_multicast = entry.Get("unicast-multicast")
	}
	if entry.Has("clock-type") {
		clock_type = entry.Get("clock-type")
	}

	log.Info("YangToDb_ptp_domain_profile_xfmr domain_number : ", domain_number, " network-transport : ", network_transport,
		" unicast-multicast : ", unicast_multicast, " clock-type : ", clock_type)

	if outval == "G.8275.x" {
		if clock_type == "BC" && network_transport == "L2" {
			return res_map, tlerr.InvalidArgsError{Format:"G.8275.2 not supported with L2 transport"}
		}
		if clock_type == "BC" && unicast_multicast == "multicast" {
			return res_map, tlerr.InvalidArgsError{Format:"G.8275.2 not supported with multicast transport"}
		}
		if clock_type == "BC" && (domain_number < 44 || domain_number > 63) {
			return res_map, tlerr.InvalidArgsError{Format:"domain must be in range 44-63 with G.8275.2"}
		}
		if clock_type == "BC" && network_transport == "UDPv6" {
			return res_map, tlerr.InvalidArgsError{Format:"ipv6 not supported with boundary-clock and G.8275.2"}
		}
		if clock_type == "P2P_TC" || clock_type == "E2E_TC" {
			return res_map, tlerr.InvalidArgsError{Format:"G.8275.2 not supported with transparent-clock"}
		}
	}
	if outval == "ieee1588" {
		if unicast_multicast == "unicast" && (clock_type == "PTP_TC" || clock_type == "E2E_TC") {
			return res_map, tlerr.InvalidArgsError{Format:"default profile not supported with transparent-clock and unicast"}
		}
	}
			

	log.Info("YangToDb_ptp_domain_profile_xfmr outval: ", outval, " field: ", field)
	res_map[field] = outval
    return res_map, nil
}

var DbToYang_ptp_domain_profile_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})

	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_domain_profile_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)
	log.Info("DbToYang_ptp_domain_profile_xfmr inParams.key: ", inParams.key)

	_, field := filepath.Split(inParams.uri)
	log.Info("DbToYang_ptp_domain_profile_xfmr field: ", field)
	value := data["PTP_CLOCK"][inParams.key].Field[field]
	result[field] = value
	log.Info("DbToYang_ptp_domain_profile_xfmr value: ", value)
	return result, err
}

var YangToDb_ptp_unicast_multicast_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map :=  make(map[string]string)
	var outval string
	var err error
	if inParams.param == nil {
		log.Info("YangToDb_ptp_unicast_multicast_xfmr Error: ")
		return res_map, err
	}
	log.Info("YangToDb_ptp_unicast_multicast_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_unicast_multicast_xfmr inParams.key: ", inParams.key)

	pathInfo := NewPathInfo(inParams.uri)
	instance_id, _  := strconv.ParseUint(pathInfo.Var("instance-number"), 10, 64)
	log.Info("YangToDb_ptp_unicast_multicast_xfmr instance_number : ", instance_id)

	ptpObj := getPtpRoot(inParams.ygRoot)
	outval = *ptpObj.InstanceList[uint32(instance_id)].DefaultDs.UnicastMulticast
	log.Info("YangToDb_ptp_unicast_multicast_xfmr outval: ", outval)
	_, field := filepath.Split(inParams.uri)
	domain_profile := ""
	network_transport := ""
	clock_type := ""

	ts := db.TableSpec { Name: "PTP_CLOCK" }
	ca := make([]string, 1, 1)

	ca[0] = "GLOBAL"
	akey := db.Key { Comp: ca}
	entry, err := inParams.d.GetEntry(&ts, akey)
	if entry.Has("domain-profile") {
		domain_profile = entry.Get("domain-profile")
	}
	if entry.Has("network-transport") {
		network_transport = entry.Get("network-transport")
	}
	if entry.Has("clock-type") {
		clock_type = entry.Get("clock-type")
	}

	log.Info("YangToDb_ptp_unicast_multicast_xfmr domain_profile : ", domain_profile,
		" network_transport : ", network_transport, " clock_type : ", clock_type)

	if outval == "multicast" {
		if domain_profile == "G.8275.x" {
			return res_map, tlerr.InvalidArgsError{Format:"multicast not supported with G.8275.2"}
		}
		keys, tblErr := inParams.d.GetKeys(&db.TableSpec{Name:"PTP_PORT|GLOBAL"})
		if tblErr == nil {
			for _, key := range keys {
				entry2, err2 := inParams.d.GetEntry(&db.TableSpec{Name:"PTP_PORT"}, key)
				if err2 == nil {
					if entry2.Has("unicast-table") {
						log.Info("YangToDb_ptp_unicast_multicast_xfmr unicast-table : ", entry2.Get("unicast-table"))
						if entry2.Get("unicast-table") != "" {
							return res_map, tlerr.InvalidArgsError{Format:"master table must be removed from " + key.Comp[1]}
						}
					}
				}
			}
		}
	}
	if outval == "unicast" {
		if domain_profile == "ieee1588" && (clock_type == "PTP_TC" || clock_type == "E2E_TC") {
			return res_map, tlerr.InvalidArgsError{Format:"unicast not supported with transparent-clock and default profile"}
		}
		if network_transport == "UDPv6" {
			return res_map, tlerr.InvalidArgsError{Format:"ipv6 not supported with unicast"}
		}
	}

	log.Info("YangToDb_ptp_unicast_multicast_xfmr outval: ", outval, " field: ", field)
	res_map[field] = outval
    return res_map, nil
}

var DbToYang_ptp_unicast_multicast_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})

	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_unicast_multicast_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)
	log.Info("DbToYang_ptp_unicast_multicast_xfmr inParams.key: ", inParams.key)

	_, field := filepath.Split(inParams.uri)
	log.Info("DbToYang_ptp_unicast_multicast_xfmr field: ", field)
	value := data["PTP_CLOCK"][inParams.key].Field[field]
	result[field] = value
	log.Info("DbToYang_ptp_unicast_multicast_xfmr value: ", value)
	return result, err
}

var YangToDb_ptp_unicast_table_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map :=  make(map[string]string)
	var outval string
	var err error
	if inParams.param == nil {
		log.Info("YangToDb_ptp_unicast_table_xfmr Error: ")
		return res_map, err
	}
	log.Info("YangToDb_ptp_unicast_table_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_unicast_table_xfmr inParams.key: ", inParams.key)
	pathInfo := NewPathInfo(inParams.uri)
	instance_id, _  := strconv.ParseUint(pathInfo.Var("instance-number"), 10, 64)
	port_number, _  := strconv.ParseUint(pathInfo.Var("port-number"), 10, 64)
	log.Info("YangToDb_ptp_unicast_table_xfmr instance_number : ", instance_id, " port_number: ", port_number)

	ptpObj := getPtpRoot(inParams.ygRoot)
	outval = *ptpObj.InstanceList[uint32(instance_id)].PortDsList[uint16(port_number)].UnicastTable
	log.Info("YangToDb_ptp_unicast_table_xfmr outval: ", outval)
	_, field := filepath.Split(inParams.uri)
	unicast_multicast := ""

	ts := db.TableSpec { Name: "PTP_CLOCK" }
	ca := make([]string, 1, 1)

	ca[0] = "GLOBAL"
	akey := db.Key { Comp: ca}
	entry, err := inParams.d.GetEntry(&ts, akey)
	if entry.Has("unicast-multicast") {
		unicast_multicast = entry.Get("unicast-multicast")
	}

    if unicast_multicast == "multicast" {
		return res_map, tlerr.InvalidArgsError{Format:"master-table is not needed in with multicast transport"}
	}

	if (outval != "") {
		addresses := strings.Split(outval, ",")
		var prev_tmp E_Ptp_AddressTypeEnumeration 
		var tmp E_Ptp_AddressTypeEnumeration 
		var first bool
		first = true
		for _,address := range addresses {
			tmp = check_address(address)
			if (PTP_ADDRESSTYPE_UNKNOWN == tmp) {
				return res_map, tlerr.InvalidArgsError{Format:"Invalid value passed for unicast-table"}
			}
			if (!first && tmp != prev_tmp) {
				return res_map, tlerr.InvalidArgsError{Format:"Mismatched addresses passed in unicast-table"}
			}
			prev_tmp = tmp
			first = false
		}
	}

	log.Info("YangToDb_ptp_unicast_table_xfmr outval: ", outval, " field: ", field)
	res_map[field] = outval
    return res_map, nil
}

var DbToYang_ptp_unicast_table_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error

	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_unicast_table_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)
	log.Info("DbToYang_ptp_unicast_table_xfmr inParams.key: ", inParams.key)

	_, field := filepath.Split(inParams.uri)
	log.Info("DbToYang_ptp_unicast_table_xfmr field: ", field)
	value := data["PTP_PORT"][inParams.key].Field[field]
	result[field] = value
	log.Info("DbToYang_ptp_unicast_table_xfmr value: ", value)
	return result, err
}

var YangToDb_ptp_udp6_scope_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map :=  make(map[string]string)
	var inval uint8
	var err error


	if inParams.param == nil {
		log.Info("YangToDb_ptp_udp6_scope_xfmr Error: ")
		return res_map, err
	}
	log.Info("YangToDb_ptp_udp6_scope_xfmr : ", *inParams.ygRoot, " Xpath: ", inParams.uri)
	log.Info("YangToDb_ptp_udp6_scope_xfmr inParams.key: ", inParams.key)
	pathInfo := NewPathInfo(inParams.uri)
	instance_id, _  := strconv.ParseUint(pathInfo.Var("instance-number"), 10, 64)
	port_number, _  := strconv.ParseUint(pathInfo.Var("port-number"), 10, 64)
	log.Info("YangToDb_ptp_udp6_scope_xfmr instance_number : ", instance_id, " port_number: ", port_number)

	ptpObj := getPtpRoot(inParams.ygRoot)
	inval = *ptpObj.InstanceList[uint32(instance_id)].DefaultDs.Udp6Scope
	log.Info("YangToDb_ptp_udp6_scope_xfmr inval: ", inval)
	_, field := filepath.Split(inParams.uri)

	if (inval > 0xf) {
		return res_map, tlerr.InvalidArgsError{Format:"Invalid value passed for udp6-scope"}
	}
	outval := fmt.Sprintf("0x%x", inval);

	log.Info("YangToDb_ptp_udp6_scope_xfmr outval: ", outval, " field: ", field)
	res_map[field] = outval
    return res_map, nil
}

var DbToYang_ptp_udp6_scope_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error

	result := make(map[string]interface{})
	data := (*inParams.dbDataMap)[inParams.curDb]
	log.Info("DbToYang_ptp_udp6_scope_xfmr ygRoot: ", *inParams.ygRoot, " Xpath: ", inParams.uri, " data: ", data)
	log.Info("DbToYang_ptp_udp6_scope_xfmr inParams.key: ", inParams.key)

	_, field := filepath.Split(inParams.uri)
	log.Info("DbToYang_ptp_udp6_scope_xfmr field: ", field)
	log.Info("DbToYang_ptp_udp6_scope_xfmr data: ", data["PTP_CLOCK"][inParams.key].Field[field])
	value,_ := strconv.ParseInt(strings.Replace(data["PTP_CLOCK"][inParams.key].Field[field], "0x", "", -1), 16, 64)
	result[field] = uint8(value)
	log.Info("DbToYang_ptp_udp6_scope_xfmr value: ", value)
	return result, err
}
