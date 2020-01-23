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

var ylibRouter http.Handler

func TestYanglibHandler(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		rc, r := GetContext(r)
		rc.Produces.Add("application/yang-data+json")
		Process(w, r)
	}

	AddRoute("ylibTop", "GET", "/restconf/data/ietf-yang-library:modules-state", h)
	AddRoute("ylibMset", "GET", "/restconf/data/ietf-yang-library:modules-state/module-set-id", h)
	AddRoute("ylibOne", "GET", "/restconf/data/ietf-yang-library:modules-state/module={name},{revision}", h)
	AddRoute("ylibNS", "GET", "/restconf/data/ietf-yang-library:modules-state/module={name},{revision}/namespace", h)

	ylibRouter = NewRouter()

	t.Run("all", testYlibGetAll)
	t.Run("mset", testYlibGetMsetID)
	t.Run("1yang", testYlibGetOne)
	t.Run("1attr", testYlibGetOneAttr)
	t.Run("1bad", testYlibGetInvalid)

	ylibRouter = nil
}

func testYlibGetAll(t *testing.T) {
	status, data := getYanglibInfo("", "", "")
	verifyRespStatus(t, status, 200)
	v := data["ietf-yang-library:modules-state"]
	if v1, ok := v.(map[string]interface{}); ok {
		v = v1["module"]
	}
	if v1, ok := v.([]interface{}); !ok || len(v1) == 0 {
		t.Fatalf("Server returned incorrect info.. %v", data)
	}
}

func testYlibGetMsetID(t *testing.T) {
	status, data := getYanglibInfo("", "", "module-set-id")
	verifyRespStatus(t, status, 200)
	if len(data) != 1 || data["ietf-yang-library:module-set-id"] == nil {
		t.Fatalf("Server returned incorrect info.. %v", data)
	}
}

func testYlibGetOne(t *testing.T) {
	status, data := getYanglibInfo("ietf-yang-library", "2016-06-21", "")
	verifyRespStatus(t, status, 200)

	var m map[string]interface{}
	if v, ok := data["ietf-yang-library:module"].([]interface{}); ok && len(v) == 1 {
		m, ok = v[0].(map[string]interface{})
	}

	if m["name"] != "ietf-yang-library" ||
		m["revision"] != "2016-06-21" ||
		m["namespace"] != "urn:ietf:params:xml:ns:yang:ietf-yang-library" ||
		m["conformance-type"] != "implement" {
		t.Fatalf("Server returned incorrect info.. %v", data)
	}
}

func testYlibGetOneAttr(t *testing.T) {
	status, data := getYanglibInfo("ietf-yang-library", "2016-06-21", "namespace")
	verifyRespStatus(t, status, 200)
	if data["ietf-yang-library:namespace"] != "urn:ietf:params:xml:ns:yang:ietf-yang-library" {
		t.Fatalf("Server returned incorrect info.. %v", data)
	}
}

func testYlibGetInvalid(t *testing.T) {
	status, _ := getYanglibInfo("unknown-yang", "0000-00-00", "")
	verifyRespStatus(t, status, 404)
}

func verifyRespStatus(t *testing.T, status, expStatus int) {
	if status != expStatus {
		t.Fatalf("Expecting response status %d; got %d", expStatus, status)
	}
}

func getYanglibInfo(name, rev, attr string) (int, map[string]interface{}) {
	u := "/restconf/data/ietf-yang-library:modules-state"
	if name != "" && rev != "" {
		u += ("/module=" + name + "," + rev)
	}
	if attr != "" {
		u += ("/" + attr)
	}

	w := httptest.NewRecorder()
	ylibRouter.ServeHTTP(w, httptest.NewRequest("GET", u, nil))

	if w.Code != 200 {
		return w.Code, nil
	}

	data := make(map[string]interface{})
	json.Unmarshal(w.Body.Bytes(), &data)
	return w.Code, data
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

	if c, ok := cap.([]interface{}); !ok || len(c) != 2 {
		log.Fatalf("Could not parse capability info: %s", w.Body.String())
	}
}

func TestQuery(t *testing.T) {
	t.Run("none", testQuery("GET", "", 0, translibArgs{}))
	t.Run("unknown", testQuery("GET", "one=1", 400, translibArgs{}))
	t.Run("depth_def", testQuery("GET", "depth=unbounded", 0, translibArgs{depth: 0}))
	t.Run("depth_0", testQuery("GET", "depth=0", 400, translibArgs{}))
	t.Run("depth_1", testQuery("GET", "depth=1", 0, translibArgs{depth: 1}))
	t.Run("depth_101", testQuery("GET", "depth=101", 0, translibArgs{depth: 101}))
	t.Run("depth_65535", testQuery("GET", "depth=65535", 0, translibArgs{depth: 65535}))
	t.Run("depth_65536", testQuery("GET", "depth=65536", 400, translibArgs{}))
	t.Run("depth_bad", testQuery("GET", "depth=bad", 400, translibArgs{}))
	t.Run("depth_extra", testQuery("GET", "depth=1&extra=1", 400, translibArgs{}))
	t.Run("depth_head", testQuery("HEAD", "depth=5", 0, translibArgs{depth: 5}))
	t.Run("depth_head_bad", testQuery("HEAD", "depth=bad", 400, translibArgs{}))
	t.Run("depth_patch", testQuery("PATCH", "depth=1", 400, translibArgs{}))
}

func testQuery(method, queryStr string, expStatus int, expData translibArgs) func(*testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, "/test?"+queryStr, nil)
		p := translibArgs{}
		err := parseRestconfQueryParams(&p, r)

		if expStatus != 0 {
			if e1, ok := err.(httpErrorType); ok && e1.status == expStatus {
				return // success
			}
			t.Fatalf("Failed to process query '%s'; expected err %d, got=%v", queryStr, expStatus, err)
		}

		if err != nil {
			t.Fatalf("Failed to process query '%s'; err=%v", queryStr, err)
		}
		if expData.depth != p.depth {
			t.Errorf("Testcase failed for query '%s'", queryStr)
			t.Fatalf("'depth' mismatch; expecting %d, found %d", expData.depth, p.depth)
		}
	}
}
