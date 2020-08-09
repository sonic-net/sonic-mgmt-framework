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
	depth bool // depth query parameter
}

func init() {

	// Metadata discovery handler
	AddRoute("hostMetadataHandler", "GET", "/.well-known/host-meta", hostMetadataHandler)

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
