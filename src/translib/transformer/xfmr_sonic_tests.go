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
//	"bytes"
//	"errors"
//	"fmt"
	"encoding/json"
	"translib/tlerr"
	"translib/db"
	"github.com/golang/glog"
)

func init() {
	XlateFuncBind("rpc_sum_cb", rpc_sum_cb)
}

var rpc_sum_cb RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
	var err error
	var operand struct {
		Input struct {
			Left int32 `json:"left"`
			Right int32 `json:"right"`
		} `json:"sonic-tests:input"`
	}

	err = json.Unmarshal(body, &operand)
	if err != nil {
		glog.Errorf("Failed to parse rpc input; err=%v", err)
		return nil,tlerr.InvalidArgs("Invalid rpc input")
	}

	var sum struct {
		Output struct {
			Result int32 `json:"result"`
		} `json:"sonic-tests:output"`
	}

	sum.Output.Result = operand.Input.Left + operand.Input.Right
	result, err := json.Marshal(&sum)
	return result, err
}
