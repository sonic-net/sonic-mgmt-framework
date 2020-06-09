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

package server

import (
	"encoding/json"
	"reflect"

	"github.com/golang/glog"
	"gopkg.in/go-playground/validator.v9"
)

func isSkipValidation(t reflect.Type) bool {
	if t == reflect.TypeOf([]int32{}) {
		return true
	}

	return false
}

// RequestValidate performas payload validation of request body.
func RequestValidate(payload []byte, ctype *MediaType, rc *RequestContext) ([]byte, error) {
	if ctype.isJSON() {
		return validateRequestJSON(payload, rc)
	}

	glog.Infof("[%s] Skipping payload validation for content-type '%v'", rc.ID, ctype.Type)
	return payload, nil
}

// validateRequestJSON performs payload validation for JSON data
func validateRequestJSON(jsn []byte, rc *RequestContext) ([]byte, error) {
	var err error
	v := rc.Model
	glog.Infof("[%s] Unmarshalling %d bytes into %T", rc.ID, len(jsn), v)

	err = json.Unmarshal(jsn, v)
	if err != nil {
		glog.Errorf("[%s] json decoding error; %v", rc.ID, err)
		return nil, httpBadRequest("Invalid json")
	}

	//log.Printf("Received data: %s\n", jsn)
	//log.Printf("Type is: %T, Value is:%v\n", v, v)
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	if !isSkipValidation(val.Type()) {
		glog.Infof("[%s] Going to validate request", rc.ID)
		validate := validator.New()
		if val.Kind() == reflect.Slice {
			//log.Println("Validate using Var")
			err = validate.Var(v, "dive")
		} else {
			//log.Println("Validate using Struct")
			err = validate.Struct(v)
		}
		if err != nil {
			glog.Errorf("[%s] validation failed: %v", rc.ID, err)
			return nil, httpBadRequest("Content not as per schema")
		}
	} else {
		glog.Infof("[%s] Skipping payload validation for dataType %v", rc.ID, val.Type())
	}

	// Get sanitized json by marshalling validated body. Removes
	// extra fields if any..
	newBody, err := json.Marshal(v)
	if err != nil {
		glog.Errorf("[%s] Failed to marshall; %v", rc.ID, err)
		return nil, httpServerError("Internal error")
	}

	return newBody, nil
}
