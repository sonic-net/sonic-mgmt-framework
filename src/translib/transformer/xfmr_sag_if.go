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
    "strings"
    log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("YangToDb_sag_ipv4_if_key_xfmr", YangToDb_sag_ipv4_if_key_xfmr)
    XlateFuncBind("DbToYang_sag_ipv4_if_key_xfmr", DbToYang_sag_ipv4_if_key_xfmr)
    XlateFuncBind("YangToDb_sag_ipv6_if_key_xfmr", YangToDb_sag_ipv6_if_key_xfmr)
    XlateFuncBind("DbToYang_sag_ipv6_if_key_xfmr", DbToYang_sag_ipv6_if_key_xfmr)	
}


var YangToDb_sag_ipv4_if_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var ifname string

    log.Info("YangToDb_sag_ipv4_if_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <name, index> */

    ifname      = pathInfo.Var("name")
    ifindex    := pathInfo.Var("index")

    if len(pathInfo.Vars) <  2 {
        err = errors.New("Invalid Key length");
        log.Info("Invalid Key length", len(pathInfo.Vars))
        return ifname, err
    }

    if len(ifname) == 0 {
        err = errors.New("SAG interface name is missing");
        log.Info("SAG interface name is Missing")
        return ifname, err
    }
    if len(ifindex) == 0 {
        err = errors.New("SAG subinterface index is missing")
        log.Info("SAG subinterface index is missing")
        return ifindex, err
    }

    log.Info("URI Interface ", ifname)
    log.Info("URI Ifindex ", ifindex)

    var sagTableKey string

    sagTableKey = ifname + "|" + "IPv4"

    log.Info("YangToDb_sag_ipv4_if_key_xfmr: sagTableKey:", sagTableKey)
    return sagTableKey, nil
}

var DbToYang_sag_ipv4_if_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_sag_ipv4_if_key_xfmr: ", entry_key)

    sagKey := strings.Split(entry_key, "|")
    ifname := sagKey[1]


    rmap["name"]  = ifname
    rmap["index"] = 0

    log.Info("DbToYang_sag_ipv4_if_key_xfmr:  ifname ", ifname)

    return rmap, nil
}


var YangToDb_sag_ipv6_if_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var ifname string

    log.Info("YangToDb_sag_ipv6_if_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <name, index> */

    ifname      = pathInfo.Var("name")
    ifindex    := pathInfo.Var("index")

    if len(pathInfo.Vars) <  2 {
        err = errors.New("Invalid Key length");
        log.Info("Invalid Key length", len(pathInfo.Vars))
        return ifname, err
    }

    if len(ifname) == 0 {
        err = errors.New("SAG interface name is missing");
        log.Info("SAG interface name is Missing")
        return ifname, err
    }
    if len(ifindex) == 0 {
        err = errors.New("SAG subinterface index is missing")
        log.Info("SAG subinterface index is missing")
        return ifindex, err
    }

    log.Info("URI Interface ", ifname)
    log.Info("URI Ifindex ", ifindex)

    var sagTableKey string

    sagTableKey = ifname + "|" + "IPv6"

    log.Info("YangToDb_sag_ipv6_if_key_xfmr: sagTableKey:", sagTableKey)
    return sagTableKey, nil
}

var DbToYang_sag_ipv6_if_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_sag_ipv6_if_key_xfmr: ", entry_key)

    sagKey := strings.Split(entry_key, "|")
    ifname := sagKey[1]


    rmap["name"]  = ifname
    rmap["index"] = 0

    log.Info("DbToYang_sag_ipv6_if_key_xfmr:  ifname ", ifname)

    return rmap, nil
}

