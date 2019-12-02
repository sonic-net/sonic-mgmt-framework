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
     "fmt"
    "strings"
	"encoding/json"
	"translib/db"
	log "github.com/golang/glog"
)

func init() {
	XlateFuncBind("rpc_config_copy", rpc_config_copy)
}


var rpc_config_copy RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
    return cfg_copy_action(body)
}

type Config struct {
    source  string
    destination string
    overwrite bool
}

var cfg_op_map map[Config]string


func init() {

    cfg_op_map = make(map[Config]string)

    cfg_op_map[Config{"running-configuration", "startup-configuration", false}] = "cfg_mgmt.save"
    cfg_op_map[Config{"running-configuration", "filename", false}]= "cfg_mgmt.save"
    cfg_op_map[Config{"filename", "running-configuration", false}] ="cfg_mgmt.load"
    cfg_op_map[Config{"startup-configuration", "running-configuration", false}] ="cfg_mgmt.load"
    cfg_op_map[Config{"filename", "running-configuration", true}] ="cfg_mgmt.reload"
    cfg_op_map[Config{"startup-configuration", "running-configuration", true}] ="cfg_mgmt.reload"
}


func validate_filename(filename string) (fname string, err error ) {
   
    if  (strings.HasPrefix(filename, "file://etc/sonic/") == false)  ||
        (strings.Contains(filename, "/..") == true)  {
            return filename, errors.New("ERROR:Invalid filename " + filename)
    }

    filename = strings.TrimPrefix(filename, "file:/")
    return filename, nil
}

func cfg_copy_action(body []byte) ([]byte, error) {
    var err error
    var result []byte
    var options []string
    var query_result  hostResult
    var source,destination,filename string

    var operand struct {
		Input struct {
			Source string `json:"source"`
            Destination string `json:"destination"`
            Overwrite bool `json:"overwrite"`
		} `json:"sonic-config-mgmt:input"`
	}

    var sum struct {
		Output struct {
			Status int32 `json:"status"`
            Status_detail string`json:"status-detail"`
		} `json:"sonic-config-mgmt:output"`
	}

    err = json.Unmarshal(body, &operand)
	if err != nil {
        /* Unmarshall failed, no input provided.
         * set to default */
       log.Error("Config input not provided.")
       err = errors.New("Input parameters missing.")
	} else {
       source = operand.Input.Source
       destination = operand.Input.Destination

       if (source != "running-configuration") &&
          (source != "startup-configuration") {
             filename, err = validate_filename(source)
             source = "filename"
       }

       if destination != "running-configuration" &&
          destination != "startup-configuration" {
             filename, err = validate_filename(destination)
             destination = "filename"
       }

       if (err == nil ) {
            config := Config{source, destination, operand.Input.Overwrite}
            cfg_cmd, ok := cfg_op_map[config]
            if ok == true {
               if (source == "filename")  ||
                   (destination == "filename") {
                       options = append(options, filename)
                       log.Info("filename", filename)
               }
               query_result = hostQuery(cfg_cmd, options)
            } else {
               log.Error("Invalid command src %s, dest %s, overwrite %t\n",
                source, destination, operand.Input.Overwrite) 
               err = errors.New("Invalid command.")
            }
       }
    }
    sum.Output.Status = 1
    if err != nil {
        sum.Output.Status_detail  =  err.Error()
    } else if query_result.Err != nil {
        sum.Output.Status_detail = "ERROR:Internal SONiC Hostservice communication failure."
    } else if query_result.Body[0].(int32) ==2 {
            sum.Output.Status_detail = fmt.Sprintf("ERROR:Invalid filename %s.", filename)
    } else if query_result.Body[0].(int32) != 0 {
            sum.Output.Status_detail = "ERROR:Command Failed."
    } else {
            sum.Output.Status = 0
            sum.Output.Status_detail = "SUCCESS."
    }
    result, err = json.Marshal(&sum)

	return result, err
}
