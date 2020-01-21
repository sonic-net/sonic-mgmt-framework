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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"translib"

	"github.com/golang/glog"
)

// Process function is the common landing place for all REST requests.
// Swagger code-gen should be configured to invoke this function
// from all generated stub functions.
func Process(w http.ResponseWriter, r *http.Request) {
	rc, r := GetContext(r)
	reqID := rc.ID
	args := translibArgs{reqID: reqID, method: r.Method}

	var err error
	var status int
	var data []byte
	var rtype string

	glog.Infof("[%s] %s %s; content-len=%d", reqID, r.Method, r.URL.Path, r.ContentLength)
	_, args.data, err = getRequestBody(r, rc)
	if err != nil {
		status, data, rtype = prepareErrorResponse(err, r)
		goto write_resp
	}

	args.path = getPathForTranslib(r)
	glog.V(1).Infof("[%s] Translated path = %s", reqID, args.path)

	err = parseQueryParams(&args, r)
	if err != nil {
		status, data, rtype = prepareErrorResponse(err, r)
		goto write_resp
	}

	status, data, err = invokeTranslib(&args, r, rc)
	if err != nil {
		glog.Errorf("[%s] Translib error %T - %v", reqID, err, err)
		status, data, rtype = prepareErrorResponse(err, r)
		goto write_resp
	}

	// Special handling for HEAD -- ignore the data but set content-length.
	// HTTP spec says HEAD can return content-length and content-type as if it was a GET.
	if r.Method == "HEAD" {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		data = nil
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
	rc, _ := GetContextWithRouteInfo(r)
	path := rc.route.path
	vars := rc.route.vars

	// Return the URL path if no variables in the template..
	if len(vars) == 0 {
		return trimRestconfPrefix(path)
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
	var err error

	for k, v := range vars {
		v, err = url.PathUnescape(v)
		if err != nil {
			glog.Warningf("Failed to unescape path var \"%s\". err=%v", v, err)
			v = vars[k]
		}

		restStyle := fmt.Sprintf("{%v}", k)
		gnmiStyle := fmt.Sprintf("[%v=%v]", rc.PMap.Get(k), escapeKeyValue(v))
		path = strings.Replace(path, restStyle, gnmiStyle, 1)
	}

	return path
}

// escapeKeyValue function escapes a path key's value as per gNMI path
// conventions -- prefixes '\' to ']' and '\'
func escapeKeyValue(val string) string {
	val = strings.Replace(val, "\\", "\\\\", -1)
	val = strings.Replace(val, "]", "\\]", -1)

	return val
}

// trimRestconfPrefix removes "/restconf/data" prefix from the path.
func trimRestconfPrefix(path string) string {
	pattern := restconfDataPathPrefix
	k := strings.Index(path, pattern)
	if k < 0 {
		pattern = restconfOperPathPrefix
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
	k := strings.Index(r.URL.Path, restconfOperPathPrefix)
	return k >= 0
	//FIXME URI pattern will not help identifying yang action APIs.
	//Use swagger generated API name instead???
}

// parseQueryParams parses the http request's query parameters
// into a translibArgs args.
func parseQueryParams(args *translibArgs, r *http.Request) error {
	if strings.HasPrefix(r.URL.Path, restconfPathPrefix) {
		return parseRestconfQueryParams(args, r)
	}

	return nil
}

// translibArgs holds arguments for invoking translib APIs.
type translibArgs struct {
	reqID  string // request id
	method string // method name
	path   string // Translib path
	data   []byte // payload

	depth uint // RESTCONF depth, for Get API only
}

// invokeTranslib calls appropriate TransLib API for the given HTTP
// method. Returns response status code and content.
func invokeTranslib(args *translibArgs, r *http.Request, rc *RequestContext) (int, []byte, error) {
	var status = 400
	var content []byte
	var err error

	ts := time.Now()

	switch r.Method {
	case "GET", "HEAD":
		req := translib.GetRequest{
			Path:  args.path,
			Depth: args.depth,
			User:  rc.Auth.User,
			Group: rc.Auth.Group,
		}

		resp, err1 := translib.Get(req)
		if err1 == nil {
			status = 200
			content = []byte(resp.Payload)
		} else {
			err = err1
		}

	case "POST":
		if isOperationsRequest(r) {
			req := translib.ActionRequest{
				Path:    args.path,
				Payload: args.data,
				User:    rc.Auth.User,
				Group:   rc.Auth.Group,
			}
			res, err1 := translib.Action(req)
			if err1 == nil {
				status = 200
				content = res.Payload
			} else {
				err = err1
			}
		} else {
			status = 201
			req := translib.SetRequest{
				Path:    args.path,
				Payload: args.data,
				User:    rc.Auth.User,
				Group:   rc.Auth.Group,
			}
			_, err = translib.Create(req)
		}

	case "PUT":
		//TODO send 201 if PUT resulted in creation
		status = 204
		req := translib.SetRequest{
			Path:    args.path,
			Payload: args.data,
			User:    rc.Auth.User,
			Group:   rc.Auth.Group,
		}
		_, err = translib.Replace(req)

	case "PATCH":
		status = 204
		req := translib.SetRequest{
			Path:    args.path,
			Payload: args.data,
			User:    rc.Auth.User,
			Group:   rc.Auth.Group,
		}
		_, err = translib.Update(req)

	case "DELETE":
		status = 204
		req := translib.SetRequest{
			Path:  args.path,
			User:  rc.Auth.User,
			Group: rc.Auth.Group,
		}
		_, err = translib.Delete(req)

	default:
		glog.Errorf("[%s] Unknown method '%v'", rc.ID, r.Method)
		err = httpBadRequest("Invalid method")
	}

	tt := time.Since(ts)
	if rc.stats != nil {
		rc.stats.translibTime = tt
		//TODO record per operation time
	}

	return status, content, err
}

// commonOptionsHandler is the common HTTP OPTIONS method handler
// for all REST API paths. Resolves allowed methods for current
// path by traversing allRoutes cache.
func commonOptionsHandler(w http.ResponseWriter, r *http.Request) {
	rc, _ := GetContextWithRouteInfo(r)
	methods := rc.route.getAllMethodsForPath()
	hasPatch := containsString(methods, "PATCH")

	// "Allow" header
	if len(methods) != 0 {
		if !containsString(methods, "OPTIONS") {
			methods = append(methods, "OPTIONS")
		}

		sort.Strings(methods)
		w.Header().Set("Allow", strings.Join(methods, ", "))
	}

	// "Accept-Patch" header for RESTCONF data paths
	if hasPatch && strings.HasPrefix(rc.route.path, restconfDataPathPrefix) {
		w.Header().Set("Accept-Patch", mimeYangDataJSON)
	}
}

// containsString checks if slice 'arr' contains the string value 's'
func containsString(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}
