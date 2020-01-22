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
    "errors"
    log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("YangToDb_sag_global_key_xfmr", YangToDb_sag_global_key_xfmr)
    XlateFuncBind("DbToYang_sag_global_key_xfmr", DbToYang_sag_global_key_xfmr)	
    XlateFuncBind("YangToDb_sag_ipv4_enable_xfmr", YangToDb_sag_ipv4_enable_xfmr)
    XlateFuncBind("DbToYang_sag_ipv4_enable_xfmr", DbToYang_sag_ipv4_enable_xfmr)
    XlateFuncBind("YangToDb_sag_ipv6_enable_xfmr", YangToDb_sag_ipv6_enable_xfmr)
    XlateFuncBind("DbToYang_sag_ipv6_enable_xfmr", DbToYang_sag_ipv6_enable_xfmr)		
}


var YangToDb_sag_global_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    /* var err error */

    log.Info("YangToDb_sag_global_key_xfmr ***", inParams.uri)
    /* pathInfo := NewPathInfo(inParams.uri) */

    /* Key should contain, <network-instance-name> */

    var sagTableKey string

    sagTableKey = "IP"

    log.Info("YangToDb_sag_global_key_xfmr: sagTableKey:", sagTableKey)
    return sagTableKey, nil
}

var DbToYang_sag_global_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_sag_global_key_xfmr: ", entry_key)

    rmap["name"] = "default"

    log.Info("DbToYang_sag_global_key_xfmr")

    return rmap, nil
}

var YangToDb_sag_ipv4_enable_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    enabled, _ := inParams.param.(*bool)
    var enStr string
    if *enabled == true {
        enStr = "enable"
    } else {
        enStr = "disable"
    }
    res_map["IPv4"] = enStr

    return res_map, nil
}

var YangToDb_sag_ipv6_enable_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    enabled, _ := inParams.param.(*bool)
    var enStr string
    if *enabled == true {
        enStr = "enable"
    } else {
        enStr = "disable"
    }
    res_map["IPv6"] = enStr

    return res_map, nil
}

var DbToYang_sag_ipv4_enable_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]

    tblName := "IP"
    if _, ok := data[tblName]; !ok {
        log.Info("DbToYang_sag_ipv4_enable_xfmr table not found : ", tblName)
        return result, errors.New("table not found : " + tblName)
    }
	
    pTbl := data[tblName]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_sag_ipv4_enable_xfmr SAG not found : ", inParams.key)
        return result, errors.New("SAG not found : " + inParams.key)
    }
    prtInst := pTbl[inParams.key]
    adminStatus, ok := prtInst.Field["IPv4"]
    if ok {
        if adminStatus == "enable" {
            result["ipv4-enable"] = true
        } else {
            result["ipv4-enable"] = false
        }
    } else {
        log.Info("Admin status field not found in DB")
    }
    return result, err
}

var DbToYang_sag_ipv6_enable_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]

    tblName := "SAG_GLOBAL|IP"
    if _, ok := data[tblName]; !ok {
        log.Info("DbToYang_sag_ipv6_enable_xfmr table not found : ", tblName)
        return result, errors.New("table not found : " + tblName)
    }

    pTbl := data[tblName]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_sag_ipv6_enable_xfmr SAG not found : ", inParams.key)
        return result, errors.New("SAG not found : " + inParams.key)
    }
    prtInst := pTbl[inParams.key]
    adminStatus, ok := prtInst.Field["IPv6"]
    if ok {
        if adminStatus == "enable" {
            result["ipv6-enable"] = true
        } else {
            result["ipv6-enable"] = false
        }
    } else {
        log.Info("Admin status field not found in DB")
    }
    return result, err
}
