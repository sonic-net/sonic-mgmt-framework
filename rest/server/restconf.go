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
	"strings"

	"github.com/golang/glog"
)

const (
	mimeYangDataJSON = "application/yang-data+json"
	mimeYangDataXML  = "application/yang-data+xml"

	restconfPathPrefix     = "/restconf/"
	restconfDataPathPrefix = "/restconf/data/"
	restconfOperPathPrefix = "/restconf/operations/"
)

// restconfCapabilities defines server capabilities
var restconfCapabilities struct {
	depth   bool // depth query parameter
	content bool // content query parameter
	fields  bool // fields query parameter
}

func init() {

	// Metadata discovery handler
	AddRoute("hostMetadataHandler", "GET", "/.well-known/host-meta", hostMetadataHandler)

	// yanglib version handler
	AddRoute("yanglibVersionHandler", "GET", "/restconf/yang-library-version", yanglibVersionHandler)

	// RESTCONF capability handler
	restconfCapabilities.depth = true
	restconfCapabilities.content = true
	restconfCapabilities.fields = true
	AddRoute("capabilityHandler", "GET",
		"/restconf/data/ietf-restconf-monitoring:restconf-state/capabilities", capabilityHandler)
	AddRoute("capabilityHandler", "GET",
		"/restconf/data/ietf-restconf-monitoring:restconf-state/capabilities/capability", capabilityHandler)

	// RESTCONF operations discovery
	AddRoute("operationsDiscovery", "GET",
		"/restconf/operations", operationsDiscoveryHandler)
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

	if restconfCapabilities.depth {
		c.Capabilities.Capability = append(c.Capabilities.Capability,
			"urn:ietf:params:restconf:capability:depth:1.0")
	}
	if restconfCapabilities.content {
		c.Capabilities.Capability = append(c.Capabilities.Capability,
			"urn:ietf:params:restconf:capability:content:1.0")
	}
	if restconfCapabilities.fields {
		c.Capabilities.Capability = append(c.Capabilities.Capability,
			"urn:ietf:params:restconf:capability:fields:1.0")
	}
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

// operationsDiscoveryHandler serves "GET /restconf/operations" request
// and returns all registered operations info -- RFC8040, section 3.3.2.
func operationsDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	emptyValue := []interface{}{nil}
	operations := make(map[string]interface{})

	match := getRouteMatchInfo(r)
	for name := range match.node.subpaths {
		name = strings.TrimPrefix(name, "/")
		operations[name] = emptyValue
	}

	glog.Infof("Found %d operation nodes", len(operations))
	dataJSON := map[string]interface{}{
		"operations": operations,
	}

	data, err := json.Marshal(dataJSON)
	if err == nil {
		w.Header().Set("Content-Type", mimeYangDataJSON)
		w.Write(data)
	} else {
		glog.Warning("Marshal error:", err)
		writeErrorResponse(w, r, err)
	}
}
