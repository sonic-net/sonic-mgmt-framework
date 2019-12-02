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
	XlateFuncBind("rpc_image_install", rpc_image_install)
  XlateFuncBind("rpc_image_remove", rpc_image_remove)
  XlateFuncBind("rpc_image_default", rpc_image_default)
}


var rpc_image_install RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
    return image_mgmt_operation("install", body)
}

var rpc_image_remove RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
    return image_mgmt_operation("remove", body)
}

var rpc_image_default RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
    return image_mgmt_operation("set_default", body)
}

func image_mgmt_operation(command string, body []byte) ([]byte, error) {

    var query_result  hostResult
   var result struct { 
    Output struct {
          Status int32 `json:"status"`
          Status_detail string`json:"status-detail"`
      } `json:"sonic-image-management:output"`
    }

    var mapData map[string]interface{}

    err:= json.Unmarshal(body, &mapData)
    var imagename, url string
    if err == nil || command == "remove" {

      input, image_present := mapData["sonic-image-management:input"]
      if image_present == true {
        mapData, image_present = input.(map[string]interface{})
        if image_present == true {
            var v interface{}
            v, image_present = mapData["imagename"]
            if (image_present) {
              imagename = v.(string)
            }
        }
      }
      err = nil
    
      log.Info("image_present:", image_present, "image:", imagename)
      if command == "remove" && image_present == false {
            command = "cleanup"
      }
      if command != "cleanup" && image_present == false {
          log.Error("Config input not provided.")
          err = errors.New("Image name not provided.") 
      }

      if err == nil {
        var options []string
   
        if (command == "install") {
          url = imagename
          if strings.HasPrefix(imagename, "file://") == true {
            imagename = strings.TrimPrefix(imagename, "file:")
          } else if ((strings.HasPrefix(imagename, "http:") == false) &&
            (strings.HasPrefix(imagename, "https:") == false)) {
            err = errors.New("ERROR:Invalid image url.") 
          }
        }

        if err == nil {
          options = append(options, command)
          if command == "install" || command == "remove" || command =="cleanup" {
                options = append(options, "-y")
          }

          if len(imagename) > 0 {
            options = append(options, imagename)
          }
          log.Info("Command:", options)
          query_result = hostQuery("image_mgmt.action", options)
        }
      }
    }

    result.Output.Status = 1
    if err != nil {
        result.Output.Status_detail  =  err.Error()
    } else if query_result.Err != nil {
        result.Output.Status_detail = "ERROR:Internal SONiC Hostservice communication failure."
    } else if query_result.Body[0].(int32) == 1 {
        if command == "install" {
          result.Output.Status_detail = fmt.Sprintf("ERROR:Invalid image URL %s.", url)
        } else {
          result.Output.Status_detail = fmt.Sprintf("ERROR:Invalid image name %s.", imagename)
        }
    } else if query_result.Body[0].(int32) != 0 {
         result.Output.Status_detail = "ERROR:Command Failed."
    } else {
        result.Output.Status = 0
        result.Output.Status_detail = "SUCCESS" 
    }
    return json.Marshal(&result)
}


