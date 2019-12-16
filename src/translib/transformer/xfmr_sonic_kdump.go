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
)

func init() {
	XlateFuncBind("rpc_kdump_config_cb", rpc_kdump_config_cb)
	XlateFuncBind("rpc_kdump_state_cb",  rpc_kdump_state_cb)
}

var rpc_kdump_config_cb RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {

	var err error
	var operand struct {
		Input struct {
            Enabled bool `json:"enabled"`
			Num_Dumps int32 `json:"num_dumps"`
			Memory string `json:"memory"`
		} `json:"sonic-kdump:input"`
	}

	err = json.Unmarshal(body, &operand)
	if err != nil {
		glog.Errorf("Failed to parse rpc input; err=%v", err)
		return nil,tlerr.InvalidArgs("Invalid rpc input")
	}

	var exec struct {
		Output struct {
			Result string `json:"result"`
		} `json:"sonic-kdump:output"`
	}

    host_output := hostQuery("kdump.command", operand.Input.Enabled, operand.Input.Num_Dumps, operand.Input.Memory)
    if host_output.Err != nil {
        glog.Errorf("host Query failed: err=%v", host_output.Err)
        glog.Flush()
        return nil, host_output.Err
    }

    var output string
    output, _ = host_output.Body[1].(string)

	exec.Output.Result = output
	result, err := json.Marshal(&exec)

	return result, err
}

var rpc_kdump_state_cb RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {

	var err error
	var operand struct {
		Input struct {
			Param string `json:"param"`
		} `json:"sonic-kdump:input"`
	}

	err = json.Unmarshal(body, &operand)
	if err != nil {
		glog.Errorf("Failed to parse rpc input; err=%v", err)
		return nil,tlerr.InvalidArgs("Invalid rpc input")
	}

	var exec struct {
		Output struct {
			Result string `json:"result"`
		} `json:"sonic-kdump:output"`
	}

    host_output := hostQuery("kdump.state", operand.Input.Param)
    if host_output.Err != nil {
        glog.Errorf("host Query failed: err=%v", host_output.Err)
        glog.Flush()
        return nil, host_output.Err
    }

    var output string
    output, _ = host_output.Body[1].(string)

	exec.Output.Result = output
	result, err := json.Marshal(&exec)

	return result, err
}
