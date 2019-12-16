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
	"github.com/openconfig/ygot/ygot"
	"translib/ocbinds"
	"fmt"
	"path/filepath"
	b64 "encoding/base64"
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
}

/* E_IETFPtp_DelayMechanismEnumeration */
var PTP_DELAY_MECH_MAP = map[string]string{
	strconv.FormatInt(int64(ocbinds.IETFPtp_DelayMechanismEnumeration_e2e), 10): "E2E",
	strconv.FormatInt(int64(ocbinds.IETFPtp_DelayMechanismEnumeration_p2p), 10): "P2P",
	strconv.FormatInt(int64(ocbinds.IETFPtp_DelayMechanismEnumeration_UNSET), 10): "Auto",
}

type ptp_id_bin [8]byte

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
	var err error
	log.Info("YangToDb_ptp_port_entry_key_xfmr root, uri: ", inParams.ygRoot, inParams.uri)
	pathInfo := NewPathInfo(inParams.uri)

	log.Info("YangToDb_ptp_port_entry_key_xfmr len(pathInfo.Vars): ", len(pathInfo.Vars))
	if len(pathInfo.Vars) < 2 {
		err = errors.New("Invalid xpath, key attributes not found")
		return entry_key, err
	}

	log.Info("YangToDb_ptp_port_entry_key_xfmr pathInfo.Var:port-number: ", pathInfo.Var("port-number"))
	entry_key = "GLOBAL|Ethernet" + pathInfo.Var("port-number")


	log.Info("YangToDb_ptp_port_entry_key_xfmr - entry_key : ", entry_key)

	return entry_key, err
}

var DbToYang_ptp_port_entry_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	var err error

	log.Info("DbToYang_ptp_port_entry_key_xfmr root, uri: ", inParams.ygRoot, inParams.uri)
	entry_key := inParams.key
	log.Info("DbToYang_ptp_port_entry_key_xfmr: ", entry_key)

	portName := entry_key
	port_num := strings.Replace(portName, "GLOBAL|Ethernet", "", 1)
	// rmap["instance-number"] = 0
	rmap["port-number"], _ = strconv.ParseInt(port_num, 10, 16)
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
	} else if strings.Contains(inParams.uri, "transparent-clock-default-ds") {
		identity = ptpObj.TransparentClockDefaultDs.ClockIdentity
		field = "clock-identity"
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
	log.Info("YangToDb_ptp_boolean_xfmr inval: ", inval, " field: ", field)

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
