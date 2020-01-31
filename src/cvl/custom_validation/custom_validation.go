////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package custom_validation

import (
	"github.com/go-redis/redis"
	"strings"
	"fmt"
	log "github.com/golang/glog"
	"net"
	"reflect"
	"os"
	"bufio"
	)

//Custom validation code for sonic-acl.yang//
/////////////////////////////////////////////
//Path : /sonic-acl/ACL_RULE/ACL_RULE_LIST
//Purpose: Allow maximum 65536 ACL rules 
//vc : Custom Validation Context
//Returns -  CVL Error object
const MAX_ACL_RULE_INSTANCES = 65536
func (t *CustomValidation) ValidateMaxAclRule(
	vc *CustValidationCtxt) CVLErrorInfo {

	var nokey []string
	ls := redis.NewScript(`return #redis.call('KEYS', "ACL_RULE|*")`)

	//Get current coutnt from Redis
	redisEntries, err := ls.Run(vc.RClient, nokey).Result()
	if err != nil {
		return CVLErrorInfo{ErrCode: CVL_SEMANTIC_ERROR}
	}

	aclTblCount := int(redisEntries.(int64))
	//Get count from user request
	for idx := 0; idx < len(vc.ReqData); idx++ {
		if (vc.ReqData[idx].VOp == OP_CREATE) &&
		(strings.HasPrefix(vc.ReqData[idx].Key, "ACL_RULE|")) {
			aclTblCount = aclTblCount + 1
		}
	}

	if (aclTblCount > MAX_ACL_RULE_INSTANCES) {
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			ErrAppTag: "too-many-elements",
			Msg: fmt.Sprintf("Max elements limit %d reached", MAX_ACL_RULE_INSTANCES),
			CVLErrDetails: "Config Validation Syntax Error",
			TableName: "ACL_RULE",
		}
	}

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}

//Path : /sonic-acl/ACL_RULE/ACL_RULE_LIST/IP_TYPE
//Purpose: Check correct for IP address provided
//         based on type IP_TYPE
//vc : Custom Validation Context
//Returns -  CVL Error object
func (t *CustomValidation) ValidateAclRuleIPAddress(
	vc *CustValidationCtxt) CVLErrorInfo {

	if (vc.CurCfg.VOp == OP_DELETE) || (vc.CurCfg.VOp == OP_UPDATE) {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	if (vc.YNodeVal == "") {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	if  (vc.YNodeVal == "ANY" || vc.YNodeVal == "IP" ||
	vc.YNodeVal == "IPV4" || vc.YNodeVal == "IPV4ANY") {

		_, srcIpV4exists := vc.CurCfg.Data["SRC_IP"]
		_, dstIpV4exists := vc.CurCfg.Data["DST_IP"]

		if (srcIpV4exists == false) || (dstIpV4exists == false) {
			return CVLErrorInfo{
				ErrCode: CVL_SEMANTIC_ERROR,
				TableName: "ACL_RULE",
				CVLErrDetails : "IP address is missing for " +
				"IP_TYPE=" + vc.YNodeVal,
			}
		}

	} else if  (vc.YNodeVal == "ANY" || vc.YNodeVal == "IP" ||
	vc.YNodeVal == "IPV6" || vc.YNodeVal == "IPV6ANY") {

		_, srcIpV6exists := vc.CurCfg.Data["SRC_IPV6"]
		_, dstIpV6exists := vc.CurCfg.Data["DST_IPV6"]

		if (srcIpV6exists == false) || (dstIpV6exists == false) {
			return CVLErrorInfo{
				ErrCode: CVL_SEMANTIC_ERROR,
				TableName: "ACL_RULE",
				CVLErrDetails : "IP address is missing for " +
				"IP_TYPE=" + vc.YNodeVal,
			}
		}
	}

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}

//Path : /sonic-sflow/SFLOW/SFLOW_LIST/agent_id
//Purpose: Check correct for correct agent_id
//vc : Custom Validation Context
//Returns -  CVL Error object
func (t *CustomValidation) ValidateSflowAgentId(
	vc *CustValidationCtxt) CVLErrorInfo {

	log.Info("ValidateSflowAgentId operation: ", vc.CurCfg.VOp)
	if (vc.CurCfg.VOp == OP_DELETE) {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	log.Info("ValidateSflowAgentId YNodeVal: ", vc.YNodeVal)
	/*  allow empty or deleted agent_id */
	if vc.YNodeVal == "" {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	/* check if input passed is found in ConfigDB PORT|* */
	tableKeys, err:= vc.RClient.Keys("PORT|*").Result()

	if (err != nil) || (vc.SessCache == nil) {
		log.Info("ValidateSflowAgentId PORT is empty or invalid argument")
		errStr := "ConfigDB PORT list is empty"
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "SFLOW",
			CVLErrDetails : errStr,
			ConstraintErrMsg : errStr,
		}
	}

	for _, dbKey := range tableKeys {
		tmp := strings.Replace(dbKey, "PORT|", "", 1)
		log.Info("ValidateSflowAgentId dbKey ", tmp)
		if (tmp == vc.YNodeVal) {
			return CVLErrorInfo{ErrCode: CVL_SUCCESS}
		}
	}

	/* check if input passed is found in list of network interfaces (includes, network_if, mgmt_if, and loopback) */
	ifaces, err2 := net.Interfaces()
	if err2 != nil {
		log.Info("ValidateSflowAgentId Error getting network interfaces")
		errStr := "Error getting network interfaces"
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "SFLOW",
			CVLErrDetails : errStr,
			ConstraintErrMsg : errStr,
		}
	}
	for _, i := range ifaces {
		log.Info("ValidateSflowAgentId i.Name ", i.Name)
		if (i.Name == vc.YNodeVal) {
			return CVLErrorInfo{ErrCode: CVL_SUCCESS}
		}
	}

	errStr := "Invalid interface name"
	return CVLErrorInfo{
		ErrCode: CVL_SEMANTIC_ERROR,
		TableName: "SFLOW",
		CVLErrDetails : errStr,
		ConstraintErrMsg : errStr,
	}
}

//Path : /sonic-ptp/PTP_PORT/PTP_PORT_LIST/underlying-interface
//Purpose: Check correct for correct agent_id
//vc : Custom Validation Context
//Returns -  CVL Error object
func (t *CustomValidation) ValidatePtpUnderlyingInterface(
	vc *CustValidationCtxt) CVLErrorInfo {

	log.Info("ValidatePtpUnderlyingInterface operation: ", vc.CurCfg.VOp)
	if (vc.CurCfg.VOp == OP_DELETE) {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	log.Info("ValidatePtpUnderlyingInterface YNodeVal: ", vc.YNodeVal)

	/* check if input passed is found in ConfigDB PORT|* */
	tableKeys, err:= vc.RClient.Keys("PORT|*").Result()

	if (err != nil) || (vc.SessCache == nil) {
		log.Info("ValidatePtpUnderlyingInterface PORT is empty or invalid argument")
		errStr := "ConfigDB PORT list is empty"
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "SFLOW",
			CVLErrDetails : errStr,
			ConstraintErrMsg : errStr,
		}
	}

	for _, dbKey := range tableKeys {
		tmp := strings.Replace(dbKey, "PORT|", "", 1)
		log.Info("ValidatePtpUnderlyingInterface dbKey ", tmp)
		if (tmp == vc.YNodeVal) {
			return CVLErrorInfo{ErrCode: CVL_SUCCESS}
		}
	}

	/* check if input passed is found in list of network interfaces (includes, network_if, mgmt_if, and loopback) */
	ifaces, err2 := net.Interfaces()
	if err2 != nil {
		log.Info("ValidatePtpUnderlyingInterface Error getting network interfaces")
		errStr := "Error getting network interfaces"
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "SFLOW",
			CVLErrDetails : errStr,
			ConstraintErrMsg : errStr,
		}
	}
	for _, i := range ifaces {
		log.Info("ValidatePtpUnderlyingInterface i.Name ", i.Name)
		if (i.Name == vc.YNodeVal) {
			return CVLErrorInfo{ErrCode: CVL_SUCCESS}
		}
	}

	errStr := "Invalid interface name"
	return CVLErrorInfo{
		ErrCode: CVL_SEMANTIC_ERROR,
		TableName: "SFLOW",
		CVLErrDetails : errStr,
		ConstraintErrMsg : errStr,
	}
}

//Path : /sonic-ptp/PTP_CLOCK
//Purpose: Check correct platform
//Returns -  CVL Error object
func (t *CustomValidation) ValidatePtp(
	vc *CustValidationCtxt) CVLErrorInfo {
		
	log.Info("ValidatePtp operation: ", vc.CurCfg.VOp)

	/* validate software build version */
	file, err := os.Open("/etc/sonic/sonic_version.yml")
	if err != nil {
		errStr := "Error opening sonic_version.yml"
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "PTP_CLOCK",
			CVLErrDetails : errStr,
			ConstraintErrMsg : errStr,
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "build_version") {
			log.Info("ValidatePtp : ", scanner.Text())
			if !strings.Contains(scanner.Text(), "Enterprise_Advanced") &&
			   !strings.Contains(scanner.Text(), "Cloud_Advanced") {
				errStr := "This object is not supported in this build"
				return CVLErrorInfo{
					ErrCode: CVL_SEMANTIC_ERROR,
					TableName: "PTP_CLOCK",
					CVLErrDetails : errStr,
					ConstraintErrMsg : errStr,
				}
			}
		}
	}

	/* validate platform */
	ls := redis.NewScript(`return redis.call('HGETALL', "DEVICE_METADATA|localhost")`)

	redisEntries, err := ls.Run(vc.RClient, []string{"platform"}).Result()
	if err != nil {
		errStr := "Cannot retrieve platform information"
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "PTP_CLOCK",
			CVLErrDetails : errStr,
			ConstraintErrMsg : errStr,
		}
	}

	s := reflect.ValueOf(redisEntries)
	var platform string
	if (s.Len() >= 2) {
		platform = s.Index(1).Interface().(string)
	}
	log.Info("ValidatePtp platform : ", platform)

	/* SONIC-16692. For some reason, with the addition of docker_routing_config_mode = split
	in DEVICE_METADATA|localhost, redis returns the hwsku. So adding hwsku in list of strings
	*/
	if !strings.Contains(platform, "x86_64-accton_as7712_32x") && 
		!strings.Contains(platform, "x86_64-accton_as5712_54x") &&
		!strings.Contains(platform, "Accton-AS7712-32X") &&
		!strings.Contains(platform, "Accton-AS5712-54X") {
		errStr := "This object is not supported in this platform"
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "PTP_CLOCK",
			CVLErrDetails : errStr,
			ConstraintErrMsg : errStr,
		}
	}

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}
