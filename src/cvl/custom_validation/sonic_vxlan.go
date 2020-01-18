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
	util "cvl/internal/util"
	)

type VxlanMap struct {
	vlanMap map[string]bool
	vniMap map[string]bool
}

func fetchVlanVNIMappingFromRedis(vc *CustValidationCtxt) {
	//Store map in the session
	pVxlanMap := &VxlanMap{}
	pVxlanMap.vlanMap = make(map[string]bool)
	pVxlanMap.vniMap = make(map[string]bool)
	vc.SessCache.Data = pVxlanMap

	//Get all VXLAN keys
	tableKeys, err:= vc.RClient.Keys("VXLAN_TUNNEL_MAP|*").Result()

	if (err != nil) || (vc.SessCache == nil) {
		util.TRACE_LEVEL_LOG(util.TRACE_SEMANTIC, "VXLAN_TUNNEL_MAP is empty or invalid argument")
		return
	}

	mCmd := map[string]*redis.SliceCmd{}
	//Get VLAN and VNI data, store
	pipe := vc.RClient.Pipeline()
	for _, dbKey := range tableKeys {
		mCmd[dbKey] = pipe.HMGet(dbKey, "vlan", "vni")
	}

	_, err = pipe.Exec()
	pipe.Close()

	for _, val := range mCmd {
		res, err := val.Result()
		if (err != nil || len(res) != 2) {
			continue
		}

		//Store data vlan-vni from Redis
		pVxlanMap.vlanMap[res[0].(string)] = true
		pVxlanMap.vniMap[res[1].(string)] = true
	}
}

//Validate unique vlan across all vlan-vni mappings
func (t *CustomValidation) ValidateUniqueVlan(vc *CustValidationCtxt) CVLErrorInfo {

	if (vc.CurCfg.VOp == OP_DELETE) {
		 return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	vlan, hasVlan := vc.CurCfg.Data["vlan"]
	if hasVlan == false {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	if (vc.SessCache.Data == nil) {
		fetchVlanVNIMappingFromRedis(vc)
	}

	vxlanMap := (vc.SessCache.Data).(*VxlanMap)

	//Loop up in session cache, if the vlan is already used
	if _, exists := vxlanMap.vlanMap[vlan]; exists {
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "VXLAN_TUNNEL_MAP",
			Keys: strings.Split(vc.CurCfg.Key, "|"),
			ErrAppTag:  "not-unique-vlanid",
		}
	}

	//Mark that Vlan is already used
	vxlanMap.vlanMap[vlan] = true

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}

//Validate unique vni across all vlan-vni mappings
func (t *CustomValidation) ValidateUniqueVNI(vc *CustValidationCtxt) CVLErrorInfo {
	if (vc.CurCfg.VOp == OP_DELETE) {
		 return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	vni, hasVni := vc.CurCfg.Data["vni"]
	if hasVni == false {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	if (vc.SessCache.Data == nil) {
		fetchVlanVNIMappingFromRedis(vc)
	}

	vxlanMap := (vc.SessCache.Data).(*VxlanMap)

	//Loop up in session cache, if the VNI is already used
	if _, exists := vxlanMap.vniMap[vni]; exists {
		return CVLErrorInfo{
			ErrCode: CVL_SEMANTIC_ERROR,
			TableName: "VXLAN_TUNNEL_MAP",
			Keys: strings.Split(vc.CurCfg.Key, "|"),
			ErrAppTag:  "not-unique-vni",
		}
	}

	//Mark that VNI is already used
	vxlanMap.vniMap[vni] = true

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}
