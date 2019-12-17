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

package transformer

import (
        "strings"
        "errors"
        log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("YangToDb_global_sg_name_xfmr", YangToDb_global_sg_name_xfmr)
    XlateFuncBind("YangToDb_global_sg_key_xfmr", YangToDb_global_sg_key_xfmr)
    XlateFuncBind("DbToYang_global_sg_key_xfmr", DbToYang_global_sg_key_xfmr)
    XlateFuncBind("global_sg_tbl_xfmr", global_sg_tbl_xfmr)
    XlateFuncBind("YangToDb_auth_set_key_xfmr", YangToDb_auth_set_key_xfmr)
    XlateFuncBind("YangToDb_server_key_xfmr", YangToDb_server_key_xfmr)
    XlateFuncBind("DbToYang_server_key_xfmr", DbToYang_server_key_xfmr)
    XlateFuncBind("server_table_xfmr", server_table_xfmr)
    XlateFuncBind("YangToDb_server_name_xfmr", YangToDb_server_name_xfmr)
}

var YangToDb_auth_set_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    return "authentication", nil
}

var YangToDb_server_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    if log.V(3) {
        log.Info( "YangToDb_server_key_xfmr: root: ", inParams.ygRoot,
            ", uri: ", inParams.uri)
    }
    pathInfo := NewPathInfo(inParams.uri)
    serverkey := pathInfo.Var("address")

    return serverkey, nil
}

var DbToYang_server_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{}, 1)
        var err error

        log.Info("DbToYang_server_key_xfmr: ", inParams.key)

        res_map["address"] = inParams.key

        return  res_map, err
}


var server_table_xfmr TableXfmrFunc = func(inParams XfmrParams) ([]string, error) {
    var err error;
    if log.V(3) {
        log.Info( "server_table_xfmr: root: ", inParams.ygRoot,
            ", uri: ", inParams.uri)
    }

    pathInfo := NewPathInfo(inParams.uri)
    servergroupname := pathInfo.Var("name")
    tables := make([]string, 0, 2)
    if strings.Contains(servergroupname, "RADIUS") {
        tables = append(tables, "RADIUS_SERVER")
    } else if strings.Contains(servergroupname, "TACACS") {
        tables = append(tables, "TACPLUS_SERVER")
    } else if inParams.oper == GET {
        tables = append(tables, "RADIUS_SERVER")
        tables = append(tables, "TACPLUS_SERVER")
    } else {
        err = errors.New("Invalid server group name")
    }

    return tables, err
}

var YangToDb_server_name_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    if log.V(3) {
        log.Info( "YangToDb_server_name_xfmr: root: ", inParams.ygRoot,
            ", uri: ", inParams.uri)
    }

    res_map :=  make(map[string]string)
    res_map["NULL"] = "NULL"
    return res_map, nil
}

var YangToDb_global_sg_name_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    if log.V(3) {
        log.Info( "YangToDb_global_sg_name_xfmr: root: ", inParams.ygRoot,
            ", uri: ", inParams.uri)
    }

    res_map :=  make(map[string]string)
    res_map["NULL"] = "NULL"
    return res_map, nil
}

var YangToDb_global_sg_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    if log.V(3) {
        log.Info( "YangToDb_global_sg_key_xfmr: root: ", inParams.ygRoot,
            ", uri: ", inParams.uri)
    }

    return "global", nil
}

var DbToYang_global_sg_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_global_sg_key_xfmr: ", inParams.key)

        return  res_map, err
}

var global_sg_tbl_xfmr TableXfmrFunc = func(inParams XfmrParams) ([]string, error) {
    var err error

    if log.V(3) {
        log.Info( "global_sg_tbl_xfmr: root: ", inParams.ygRoot,
            ", uri: ", inParams.uri)
    }

    pathInfo := NewPathInfo(inParams.uri)
    servergroupname := pathInfo.Var("name")
    tables := make([]string, 0, 2)
    if strings.Contains(servergroupname, "RADIUS") {
        tables = append(tables, "RADIUS")
    } else if strings.Contains(servergroupname, "TACACS") {
        tables = append(tables, "TACPLUS")
    } else if inParams.oper == GET {
        tables = append(tables, "RADIUS")
        tables = append(tables, "TACPLUS")
    } else {
        err = errors.New("Invalid server group name")
    }

    if log.V(3) {
        log.Info( "global_sg_tbl_xfmr: tables: ", tables,
            " err: ", err)
    }

    return tables, err
}
