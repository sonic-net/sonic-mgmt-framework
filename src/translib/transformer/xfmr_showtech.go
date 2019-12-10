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
    "encoding/json"
    "translib/tlerr"
    "translib/db"
    "github.com/golang/glog"
    "regexp"
)

func init() {
    XlateFuncBind("rpc_showtech_cb", rpc_showtech_cb)
}

var rpc_showtech_cb RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
    var err error
    var matched bool
    var output string
    var operand struct {
        Input struct {
            Date string `json:"date"`
        } `json:"sonic-show-techsupport:input"`
    }

    err = json.Unmarshal(body, &operand)
    if err != nil {
        glog.Errorf("%Error: Failed to parse rpc input; err=%v", err)
        return nil,tlerr.InvalidArgs("Invalid rpc input")
    }

    if operand.Input.Date == "" {
        matched = true
    } else {
        matched, err = regexp.MatchString((`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?` +
                                           `(Z|[\+\-]\d{2}:\d{2})`), operand.Input.Date)
    }

    var showtech struct {
        Output struct {
        Result string `json:"output-filename"`
        } `json:"sonic-show-techsupport:output"`
    }

    if matched != true {
        showtech.Output.Result = "Invalid input: Incorrect DateTime format"
        result, _ := json.Marshal(&showtech)
        return result, nil
    }

    host_output := hostQuery("showtech.info", operand.Input.Date)
    if host_output.Err != nil {
        glog.Errorf("%Error: Showtech host Query failed: err=%v", host_output.Err)
        glog.Flush()
        return nil, host_output.Err
    }

    output, _ = host_output.Body[1].(string)
    showtech.Output.Result = output

    result, _ := json.Marshal(&showtech)

    return result, nil
}
