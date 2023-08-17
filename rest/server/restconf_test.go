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
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
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
	newDefaultRouter().ServeHTTP(w, r)

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
	newDefaultRouter().ServeHTTP(w, r)

	// Parse capability response
	var cap interface{}
	top := make(map[string]interface{})
	parseResponseJSON(t, w, &top)

	if c := top["ietf-restconf-monitoring:capabilities"]; c != nil {
		cap = c.(map[string]interface{})["capability"]
	} else {
		cap = top["ietf-restconf-monitoring:capability"]
	}

	if c, ok := cap.([]interface{}); !ok || len(c) == 0 {
		log.Fatalf("Could not parse capability info: %s", w.Body.String())
	}

	var curCap []interface{}
	curCap = append(curCap, "urn:ietf:params:restconf:capability:defaults:1.0?basic-mode=report-all",
		"urn:ietf:params:restconf:capability:depth:1.0",
		"urn:ietf:params:restconf:capability:content:1.0",
		"urn:ietf:params:restconf:capability:fields:1.0")

	if !reflect.DeepEqual(cap.([]interface{}), curCap) {
		t.Fatalf("Response does not include expected capabilities \n"+
			"expected: %v\nfound: %v", curCap, cap)
	}
}

func TestOpsDiscovery_none(t *testing.T) {
	testOpsDiscovery(t, nil)
}

func TestOpsDiscovery_one(t *testing.T) {
	testOpsDiscovery(t, []string{"testing:system-restart"})
}

func TestOpsDiscovery(t *testing.T) {
	testOpsDiscovery(t, []string{"testing:cpu", "testing:clock", "hello:world", "foo:bar"})
}

func testOpsDiscovery(t *testing.T, ops []string) {
	s := newEmptyRouter()
	f := func(w http.ResponseWriter, r *http.Request) {}
	s.addRoute("opsDiscovery", "GET",
		"/restconf/operations", operationsDiscoveryHandler)
	for _, op := range ops {
		s.addRoute(op, "POST", "/restconf/operations/"+op, f)
	}

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", "/restconf/operations", nil))

	var resp struct {
		Ops map[string][]interface{} `json:"operations"`
	}

	parseResponseJSON(t, w, &resp)
	if resp.Ops == nil {
		t.Fatal("Response does not contain 'operations' object:", resp)
	}

	var respOps []string
	for op := range resp.Ops {
		respOps = append(respOps, op)
	}

	sort.Strings(ops)
	sort.Strings(respOps)
	if !reflect.DeepEqual(respOps, ops) {
		t.Fatalf("Response does not include expected operations list\n"+
			"expected: %v\nfound: %v", ops, respOps)
	}
}

func parseResponseJSON(t *testing.T, w *httptest.ResponseRecorder, resp interface{}) {
	if w.Code != 200 {
		t.Fatalf("Request failed with status %d", w.Code)
	}
	if len(w.Body.Bytes()) == 0 {
		t.Fatalf("No response body")
	}
	if w.Header().Get("Content-Type") != mimeYangDataJSON {
		t.Fatalf("Expected content-type=%s, found=%s", mimeYangDataJSON, w.Header().Get("Content-Type"))
	}

	err := json.Unmarshal(w.Body.Bytes(), resp)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
}
