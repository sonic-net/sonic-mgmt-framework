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
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

const mimeYangDataJSON = "application/yang-data+json"
const mimeYangDataXML = "application/yang-data+xml"
const restconfPathPrefix = "/restconf/"
const restconfDataPathPrefix = "/restconf/data/"
const restconfOperPathPrefix = "/restconf/operations/"

func init() {

	// Metadata discovery handler
	AddRoute("hostMetadataHandler", "GET", "/.well-known/host-meta", hostMetadataHandler)

	// yanglib version handler
	AddRoute("yanglibVersionHandler", "GET", "/restconf/yang-library-version", yanglibVersionHandler)

	// RESTCONF capability handler
	AddRoute("capabilityHandler", "GET",
		"/restconf/data/ietf-restconf-monitoring:restconf-state/capabilities", capabilityHandler)
	AddRoute("capabilityHandler", "GET",
		"/restconf/data/ietf-restconf-monitoring:restconf-state/capabilities/capability", capabilityHandler)

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

// yanglibVersionHandler handles "GET /restconf/yang-library-version"
// request as per RFC8040. Yanglib version supported is "2016-06-21"
func yanglibVersionHandler(w http.ResponseWriter, r *http.Request) {
	var data bytes.Buffer
	var contentType string
	accept := r.Header.Get("Accept")

	// Rudimentary content negotiation
	if strings.Contains(accept, mimeYangDataXML) {
		contentType = mimeYangDataXML
		data.WriteString("<yang-library-version xmlns='urn:ietf:params:xml:ns:yang:ietf-restconf'>")
		data.WriteString("2016-06-21</yang-library-version>")
	} else {
		contentType = mimeYangDataJSON
		data.WriteString("{\"ietf-restconf:yang-library-version\": \"2016-06-21\"}")
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(data.Bytes())
}

// capabilityHandler serves RESTCONF capability requests -
// "GET /restconf/data/ietf-restconf-monitoring:restconf-state/capabilities"
func capabilityHandler(w http.ResponseWriter, r *http.Request) {
	var c struct {
		Capabilities struct {
			Capability []string `json:"capability"`
		} `json:"capabilities"`
	}

	c.Capabilities.Capability = append(c.Capabilities.Capability,
		"urn:ietf:params:restconf:capability:defaults:1.0?basic-mode=report-all")

	c.Capabilities.Capability = append(c.Capabilities.Capability,
		"urn:ietf:params:restconf:capability:depth:1.0")

	var data []byte
	if strings.HasSuffix(r.URL.Path, "/capabilities") {
		data, _ = json.Marshal(&c)
	} else {
		data, _ = json.Marshal(&c.Capabilities)
	}

	// A hack to add module prefix
	// TODO find better method.. My be ygot?
	if bytes.HasPrefix(data, []byte("{\"")) {
		var buff bytes.Buffer
		buff.WriteString("{\"ietf-restconf-monitoring:")
		buff.Write(data[2:])
		data = buff.Bytes()
	}

	w.Header().Set("Content-Type", mimeYangDataJSON)
	w.Write(data)
}

// parseRestconfQueryParams parses query parameters of a request 'r' to
// fill translibArgs object 'args'. Only RESTCONF standard parameters
// are accepted. Returns httpError with status 400 if any parameter is
// unsupported or has invalid value.
func parseRestconfQueryParams(args *translibArgs, r *http.Request) error {
	var err error
	qParams := r.URL.Query()

	for name, val := range qParams {
		switch name {
		case "depth":
			args.depth, err = parseDepthParam(val, r)
		default:
			err = httpError(http.StatusBadRequest, "Unsupported query parameter '%s'", name)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// parseDepthParam parses query parameter value for "depth" parameter.
// See https://tools.ietf.org/html/rfc8040#section-4.8.2
func parseDepthParam(v []string, r *http.Request) (uint, error) {
	if r.Method != "GET" && r.Method != "HEAD" {
		return 0, httpError(http.StatusBadRequest,
			"%s does not support 'depth' query parameter", r.Method)
	}

	if len(v) != 1 {
		glog.Errorf("Expecting only 1 depth param.. found %d", len(v))
		return 0, httpError(http.StatusBadRequest,
			"Invalid 'depth' query parameter")
	}

	if v[0] == "unbounded" {
		return 0, nil
	}

	d, err := strconv.ParseUint(v[0], 10, 16)
	if err != nil || d == 0 {
		glog.Errorf("Bad depth value '%s', err=%v", v[0], err)
		return 0, httpError(http.StatusBadRequest,
			"Invalid 'depth' query parameter")
	}

	return uint(d), nil
}
