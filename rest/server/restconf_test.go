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
	"encoding/xml"
	"log"
	"net/http/httptest"
	"testing"
)

func TestMetaHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/.well-known/host-meta", nil)
	w := httptest.NewRecorder()

	newDefaultRouter().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("Request failed with status %d", w.Code)
	}

	ct, _ := parseMediaType(w.Header().Get("content-type"))
	if ct == nil || ct.Type != "application/xrd+xml" {
		t.Fatalf("Unexpected content-type '%s'", w.Header().Get("content-type"))
	}

	data := w.Body.Bytes()
	if len(data) == 0 {
		t.Fatalf("No response body")
	}

	var payload struct {
		XMLName xml.Name `xml:"XRD"`
		Links   []struct {
			Rel  string `xml:"rel,attr"`
			Href string `xml:"href,attr"`
		} `xml:"Link"`
	}

	err := xml.Unmarshal(data, &payload)
	if err != nil {
		t.Fatalf("Response parsing failed; err=%v", err)
	}

	if payload.XMLName.Local != "XRD" ||
		payload.XMLName.Space != "http://docs.oasis-open.org/ns/xri/xrd-1.0" {
		t.Fatalf("Invalid response '%s'", data)
	}

	var rcRoot string
	for _, x := range payload.Links {
		if x.Rel == "restconf" {
			rcRoot = x.Href
		}
	}

	t.Logf("Restconf root = '%s'", rcRoot)
	if rcRoot != "/restconf" {
		t.Fatalf("Invalid restconf root; expected '/restconf'")
	}
}

func TestCapability_1(t *testing.T) {
	testCapability(t, "/restconf/data/ietf-restconf-monitoring:restconf-state/capabilities")
}

func TestCapability_2(t *testing.T) {
	testCapability(t, "/restconf/data/ietf-restconf-monitoring:restconf-state/capabilities/capability")
}

func testCapability(t *testing.T, path string) {
	r := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	newDefaultRouter().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("Request failed with status %d", w.Code)
	}
	if len(w.Body.Bytes()) == 0 {
		t.Fatalf("No response body")
	}
	if w.Header().Get("Content-Type") != mimeYangDataJSON {
		t.Fatalf("Expected content-type=%s, found=%s", mimeYangDataJSON, w.Header().Get("Content-Type"))
	}

	// Parse capability response
	var cap interface{}
	top := make(map[string]interface{})
	json.Unmarshal(w.Body.Bytes(), &top)

	if c := top["ietf-restconf-monitoring:capabilities"]; c != nil {
		cap = c.(map[string]interface{})["capability"]
	} else {
		cap = top["ietf-restconf-monitoring:capability"]
	}

	if c, ok := cap.([]interface{}); !ok || len(c) == 0 {
		log.Fatalf("Could not parse capability info: %s", w.Body.String())
	}
}
