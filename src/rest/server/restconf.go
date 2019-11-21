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
	"net/http"
	"strings"
)

const mimeYangDataJSON = "application/yang-data+json"
const mimeYangDataXML = "application/yang-data+xml"

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
