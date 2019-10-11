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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"translib"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

// Process function is the common landing place for all REST requests.
// Swagger code-gen should be configured to invoke this function
// from all generated stub functions.
func Process(w http.ResponseWriter, r *http.Request) {
	rc, r := GetContext(r)
	reqID := rc.ID
	path := r.URL.Path

	var status int
	var data []byte
	var rtype string

	glog.Infof("[%s] %s %s; content-len=%d", reqID, r.Method, path, r.ContentLength)
	_, body, err := getRequestBody(r, rc)
	if err != nil {
		status, data, rtype = prepareErrorResponse(err, r)
		goto write_resp
	}

	path = getPathForTranslib(r)
	glog.Infof("[%s] Translated path = %s", reqID, path)

	status, data, err = invokeTranslib(r, path, body)
	if err != nil {
		glog.Errorf("[%s] Translib error %T - %v", reqID, err, err)
		status, data, rtype = prepareErrorResponse(err, r)
		goto write_resp
	}

	rtype, err = resolveResponseContentType(data, r, rc)
	if err != nil {
		glog.Errorf("[%s] Failed to resolve response content-type, err=%v", rc.ID, err)
		status, data, rtype = prepareErrorResponse(err, r)
		goto write_resp
	}

write_resp:
	glog.Infof("[%s] Sending response %d, type=%s, data=%s", reqID, status, rtype, data)

	// Write http response.. Following strict order should be
	// maintained to form proper response.
	//	1. Set custom headers via w.Header().Set("N", "V")
	//	2. Set status code via w.WriteHeader(code)
	//	3. Finally, write response body via w.Write(bytes)
	if len(data) != 0 {
		w.Header().Set("Content-Type", rtype)
		w.WriteHeader(status)
		w.Write([]byte(data))
	} else {
		// No data, status only
		w.WriteHeader(status)
	}
}

// getRequestBody returns the validated request body
func getRequestBody(r *http.Request, rc *RequestContext) (*MediaType, []byte, error) {
	if r.ContentLength == 0 {
		glog.Infof("[%s] No body", rc.ID)
		return nil, nil, nil
	}

	// read body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("[%s] Failed to read body; err=%v", rc.ID, err)
		return nil, nil, httpError(http.StatusInternalServerError, "")
	}

	// Parse content-type header value
	ctype := r.Header.Get("Content-Type")

	// Guess the contet type if client did not provide it
	if ctype == "" {
		glog.Infof("[%s] Content-type not provided in request. Guessing it...", rc.ID)
		ctype = http.DetectContentType(body)
	}

	ct, err := parseMediaType(ctype)
	if err != nil {
		glog.Errorf("[%s] Bad content-type '%s'; err=%v",
			rc.ID, r.Header.Get("Content-Type"), err)
		return nil, nil, httpBadRequest("Bad content-type")
	}

	// Check if content type is one of the acceptable types specified
	// in "consumes" section in OpenAPI spec.
	if !rc.Consumes.Contains(ct.Type) {
		glog.Errorf("[%s] Content-type '%s' not supported. Valid types %v", rc.ID, ct.Type, rc.Consumes)
		return nil, nil, httpError(http.StatusUnsupportedMediaType, "Unsupported content-type")
	}

	// Do payload validation if model info is set in the context.
	if rc.Model != nil {
		body, err = RequestValidate(body, ct, rc)
		if err != nil {
			return nil, nil, err
		}
	}

	glog.Infof("[%s] Content-type=%s; data=%s", rc.ID, ctype, body)
	return ct, body, nil
}

// resolveResponseContentType
func resolveResponseContentType(data []byte, r *http.Request, rc *RequestContext) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	// If OpenAPI spec has only one "produces" option, assume that
	// app module will return that exact type data!!
	if len(rc.Produces) == 1 {
		return rc.Produces[0].Format(), nil
	}

	//TODO validate against Accept header

	return http.DetectContentType(data), nil
}

// getPathForTranslib converts REST URIs into GNMI paths
func getPathForTranslib(r *http.Request) string {
	// Return the URL path if no variables in the template..
	vars := mux.Vars(r)
	if len(vars) == 0 {
		return trimRestconfPrefix(r.URL.Path)
	}

	path, err := mux.CurrentRoute(r).GetPathTemplate()
	if err != nil {
		glog.Infof("No path template for this route")
		return trimRestconfPrefix(r.URL.Path)
	}

	// Path is a template.. Convert it into GNMI style path
	// WARNING: does not handle duplicate key attribute names
	//
	// Template   = /openconfig-acl:acl/acl-sets/acl-set={name},{type}
	// REST style = /openconfig-acl:acl/acl-sets/acl-set=TEST,ACL_IPV4
	// GNMI style = /openconfig-acl:acl/acl-sets/acl-set[name=TEST][type=ACL_IPV4]
	path = trimRestconfPrefix(path)
	path = strings.Replace(path, "={", "{", -1)
	path = strings.Replace(path, "},{", "}{", -1)
	rc, _ := GetContext(r)

	for k, v := range vars {
		restStyle := fmt.Sprintf("{%v}", k)
		gnmiStyle := fmt.Sprintf("[%v=%v]", rc.PMap.Get(k), v)
		path = strings.Replace(path, restStyle, gnmiStyle, 1)
	}

	return path
}

// trimRestconfPrefix removes "/restconf/data" prefix from the path.
func trimRestconfPrefix(path string) string {
	pattern := "/restconf/data/"
	k := strings.Index(path, pattern)
	if k < 0 {
		pattern = "/restconf/operations/"
		k = strings.Index(path, pattern)
	}
	if k >= 0 {
		path = path[k+len(pattern)-1:]
	}

	return path
}

// isOperationsRequest checks if a request is a RESTCONF operations
// request (rpc or action)
func isOperationsRequest(r *http.Request) bool {
	k := strings.Index(r.URL.Path, "/restconf/operations/")
	return k >= 0
	//FIXME URI pattern will not help identifying yang action APIs.
	//Use swagger generated API name instead???
}

// getRequestID returns the request id encoded in the context
func getRequestID(r *http.Request) string {
	rc, _ := GetContext(r)
	return rc.ID
}

// invokeTranslib calls appropriate TransLib API for the given HTTP
// method. Returns response status code and content.
func invokeTranslib(r *http.Request, path string, payload []byte) (int, []byte, error) {
	var status = 400
	var content []byte
	var err error

	switch r.Method {
	case "GET":
		req := translib.GetRequest{Path: path}
		resp, err1 := translib.Get(req)
		if err1 == nil {
			status = 200
			content = []byte(resp.Payload)
		} else {
			err = err1
		}

	case "POST":
		if isOperationsRequest(r) {
			req := translib.ActionRequest{Path: path, Payload: payload}
			res, err1 := translib.Action(req)
			if err1 == nil {
				status = 200
				content = res.Payload
			} else {
				err = err1
			}
		} else {
			status = 201
			req := translib.SetRequest{Path: path, Payload: payload}
			_, err = translib.Create(req)
		}

	case "PUT":
		//TODO send 201 if PUT resulted in creation
		status = 204
		req := translib.SetRequest{Path: path, Payload: payload}
		_, err = translib.Replace(req)

	case "PATCH":
		status = 204
		req := translib.SetRequest{Path: path, Payload: payload}
		_, err = translib.Update(req)

	case "DELETE":
		status = 204
		req := translib.SetRequest{Path: path}
		_, err = translib.Delete(req)

	default:
		glog.Errorf("[%s] Unknown method '%v'", getRequestID(r), r.Method)
		err = httpBadRequest("Invalid method")
	}

	return status, content, err
}

// hostMetadataHandler function handles "GET /.well-known/host-meta"
// request as per RFC6415. RESTCONF specification requires this for
// advertising the RESTCONF root path ("/restconf" in our case).
func hostMetadataHandler(w http.ResponseWriter, r *http.Request) {
	var data bytes.Buffer
	data.WriteString("<XRD xmlns='http://docs.oasis-open.org/ns/xri/xrd-1.0'>")
	data.WriteString("<Link rel='restconf' href='/restconf'/>")
	data.WriteString("</XRD>")

	w.Header().Set("Content-Type", "application/xrd+xml")
	w.Write(data.Bytes())
}
