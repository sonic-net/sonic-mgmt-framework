//////////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom, Inc.
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
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
	"strings"
	"translib/ocbinds"
	"translib/tlerr"
)

func init() {
	XlateFuncBind("YangToDb_vlan_nd_suppress_key_xfmr", YangToDb_vlan_nd_suppress_key_xfmr)
	XlateFuncBind("DbToYang_vlan_nd_suppress_key_xfmr", DbToYang_vlan_nd_suppress_key_xfmr)
	XlateFuncBind("YangToDb_vlan_nd_suppress_fld_xfmr", YangToDb_vlan_nd_suppress_fld_xfmr)
	XlateFuncBind("DbToYang_vlan_nd_suppress_fld_xfmr", DbToYang_vlan_nd_suppress_fld_xfmr)
}

var YangToDb_vlan_nd_suppress_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	pathInfo := NewPathInfo(inParams.uri)
	vlanIdStr := pathInfo.Var("name")

	if !strings.HasPrefix(vlanIdStr, "Vlan") {
		return "", tlerr.InvalidArgs("Invalid key: %v", vlanIdStr)
	}
	return vlanIdStr, nil
}

var DbToYang_vlan_nd_suppress_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	res_map := make(map[string]interface{})

	log.Info("Vlan Name = ", inParams.key)
	res_map["name"] = inParams.key
	return res_map, nil
}

var YangToDb_vlan_nd_suppress_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)

	pathInfo := NewPathInfo(inParams.uri)
	vlanIdStr := pathInfo.Var("name")

	if !strings.HasPrefix(vlanIdStr, "Vlan") {
		return res_map, tlerr.InvalidArgs("Invalid key: %v", vlanIdStr)
	}
	log.Infof("YangToDb_vlan_nd_suppress_fld_xfmr: Params: %v", inParams.param)

	if inParams.param != nil {
		val, _ := inParams.param.(ocbinds.E_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_NeighbourSuppress_Config_ArpAndNdSuppress)
		if val == ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_NeighbourSuppress_Config_ArpAndNdSuppress_enable {
			res_map["suppress"] = "on"
		}
	}

	return res_map, nil
}

var DbToYang_vlan_nd_suppress_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	res_map := make(map[string]interface{})

	pathInfo := NewPathInfo(inParams.uri)
	vlanIdStr := pathInfo.Var("name")
	data := (*inParams.dbDataMap)[inParams.curDb]

	log.Infof("vlan_nd_suppress_fld_xfmr: key: %v, data: %v", vlanIdStr, data)
	if data != nil && len(data) > 0 {
		val := data["SUPPRESS_VLAN_NEIGH"][vlanIdStr]
		if val.Get("suppress") == "on" {
			res_map["arp-and-nd-suppress"], _ = ygot.EnumName(ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_NeighbourSuppress_Config_ArpAndNdSuppress_enable)
		}
	}

	return res_map, nil
}
