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

	NewRouter().ServeHTTP(w, r)

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

func TestYanglibVer_json(t *testing.T) {
	testYanglibVer(t, mimeYangDataJSON, mimeYangDataJSON)
}

func TestYanglibVer_xml(t *testing.T) {
	testYanglibVer(t, mimeYangDataXML, mimeYangDataXML)
}

func TestYanglibVer_default(t *testing.T) {
	testYanglibVer(t, "", mimeYangDataJSON)
}

func TestYanglibVer_unknown(t *testing.T) {
	testYanglibVer(t, "text/plain", mimeYangDataJSON)
}

func testYanglibVer(t *testing.T, requestAcceptType, expectedContentType string) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/restconf/yang-library-version", nil)
	if requestAcceptType != "" {
		r.Header.Set("Accept", requestAcceptType)
	}

	t.Logf("GET /restconf/yang-library-version with accept=%s", requestAcceptType)
	NewRouter().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("Request failed with status %d", w.Code)
	}
	if len(w.Body.Bytes()) == 0 {
		t.Fatalf("No response body")
	}
	if w.Header().Get("Content-Type") != expectedContentType {
		t.Fatalf("Expected content-type=%s, found=%s", expectedContentType, w.Header().Get("Content-Type"))
	}

	var err error
	var resp struct {
		XMLName xml.Name `json:"-" xml:"urn:ietf:params:xml:ns:yang:ietf-restconf yang-library-version"`
		Version string   `json:"ietf-restconf:yang-library-version" xml:",chardata"`
	}

	if expectedContentType == mimeYangDataXML {
		err = xml.Unmarshal(w.Body.Bytes(), &resp)
	} else {
		err = json.Unmarshal(w.Body.Bytes(), &resp)
	}
	if err != nil {
		t.Fatalf("Response parsing failed; err=%v", err)
	}

	t.Logf("GOT yang-library-version %s; content-type=%s", resp.Version, w.Header().Get("Content-Type"))
	if resp.Version != "2016-06-21" {
		t.Fatalf("Expected yanglib version 2016-06-21; received=%s", resp.Version)
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

	NewRouter().ServeHTTP(w, r)

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

	if c, ok := cap.([]interface{}); !ok || len(c) != 1 {
		log.Fatalf("Could not parse capability info: %v", w.Body.String())
	}
}
