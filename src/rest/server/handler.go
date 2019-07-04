///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"translib"

	"github.com/gorilla/mux"
)

////
// Request Id generator
var requestCounter uint64

// Process function is the common landing place for all REST requests.
// Swagger code-gen should be configured to invoke this function
// from all generated stub functions.
func Process(w http.ResponseWriter, r *http.Request) {
	reqID := fmt.Sprintf("REST-%d", atomic.AddUint64(&requestCounter, 1))

	var body []byte

	log.Printf("[%s] Received %s %s; content-len=%d", reqID, r.Method, r.URL.Path, r.ContentLength)
	if r.ContentLength > 0 {
		contentType := r.Header.Get("Content-Type")
		body, _ = ioutil.ReadAll(r.Body)

		log.Printf("[%s] Content-type=%s; data=%s", reqID, contentType, body)
	}

	path := getPathForTranslib(r)
	log.Printf("[%s] Translated path = %s", reqID, path)

	status, data := invokeTranslib(reqID, r.Method, path, body)

	log.Printf("[%s] Sending response %d, data=%s", reqID, status, data)

	// Write http response.. Following strict order should be
	// maintained to form proper response.
	//	1. Set custom headers via w.Header().Set("N", "V")
	//	2. Set status code via w.WriteHeader(code)
	//	3. Finally, write response body via w.Write(bytes)
	if len(data) != 0 {
		w.Header().Set("Content-Type", "application/yang-data+json; charset=UTF-8")
		w.WriteHeader(status)
		w.Write([]byte(data))
	} else {
		// No data, status only
		w.WriteHeader(status)
	}
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
		log.Printf("No path template for this route")
		return trimRestconfPrefix(r.URL.Path)
	}

	//log.Printf("vars = %v", vars)
	//log.Printf("path = %v", path)

	// Path is a template.. Convert it into GNMI style path
	// WARNING: does not handle duplicate key attribute names
	//
	// Template   = /openconfig-acl:acl/acl-sets/acl-set={name},{type}
	// REST style = /openconfig-acl:acl/acl-sets/acl-set=TEST,ACL_IPV4
	// GNMI style = /openconfig-acl:acl/acl-sets/acl-set[name=TEST][type=ACL_IPV4]
	//
	// Conversion logic:
	// 1) Remove all "=" and "," from the template
	//    "acl-set={name},{type}" becomes "acl-set{name}{type}"
	// 2) Replace all "{ATTR}" patterns with "[ATTR=VALUE]". Attribute
	//    name value mapping is provided by mux.Vars() API
	//    "acl-set{name}{type}" becomes "acl-set[name=TEST][type=ACL_IPV4]"
	path = trimRestconfPrefix(path)
	path = strings.Replace(path, "={", "{", -1)
	path = strings.Replace(path, "},{", "}{", -1)

	for k, v := range vars {
		restStyle := fmt.Sprintf("{%v}", k)
		gnmiStyle := fmt.Sprintf("[%v=%v]", k, v)
		path = strings.Replace(path, restStyle, gnmiStyle, 1)
	}

	return path
}

// trimRestconfPrefix removes "/<version>/restconf/data" prefix
// from the path.
func trimRestconfPrefix(path string) string {
	pattern := "/restconf/data/"
	k := strings.Index(path, pattern)
	if k > 0 {
		path = path[k+len(pattern)-1:]
	}

	return path
}

// isTranslibSuccess checks if the error object returned by TransLib
// is indeed an error!!!!
func isTranslibSuccess(err error) bool {
	if err != nil && err.Error() != "Success" {
		return false
	}

	return true
}

// invokeTranslib calls appropriate TransLib API for the given HTTP
// method. Returns response status code and content.
func invokeTranslib(reqID, method, path string, payload []byte) (int, []byte) {
	var status = 400
	var content []byte
	var err error

	switch method {
	case "GET":
		req := translib.GetRequest{Path: path}
		resp, err1 := translib.Get(req)
		if isTranslibSuccess(err1) {
			data := resp.Payload
			status = 200
			content = []byte(data)
		}
		if err1 != nil {
			log.Printf("Printing the ErrSrc =%v", resp.ErrSrc)
		}
		err = err1

	case "POST":
		//TODO return 200 for operations request
		status = 201
		req := translib.SetRequest{Path: path, Payload: payload}
		resp, err1 := translib.Create(req)
		if err1 != nil {
			log.Printf("Printing the ErrSrc =%v", resp.ErrSrc)
		}
		err = err1

	case "PUT":
		//TODO send 201 if PUT resulted in creation
		status = 204
		req := translib.SetRequest{Path: path, Payload: payload}
		resp, err1 := translib.Replace(req)
		if err1 != nil {
			log.Printf("Printing the ErrSrc =%v", resp.ErrSrc)
		}
		err = err1

	case "PATCH":
		status = 204
		req := translib.SetRequest{Path: path, Payload: payload}
		resp, err1 := translib.Update(req)
		if err1 != nil {
			log.Printf("Printing the ErrSrc =%v", resp.ErrSrc)
		}
		err = err1

	case "DELETE":
		status = 204
		req := translib.SetRequest{Path: path}
		resp, err1 := translib.Delete(req)
		if err1 != nil {
			log.Printf("Printing the ErrSrc =%v", resp.ErrSrc)
		}
		err = err1
	default:
		log.Printf("[%s] Unknown method '%v'", reqID, method)
		status = 400
	}

	if !isTranslibSuccess(err) {
		log.Printf("[%s] Translib returned error - %v", reqID, err)
		status = 400
		content = []byte(err.Error())
	}

	return status, content
}
