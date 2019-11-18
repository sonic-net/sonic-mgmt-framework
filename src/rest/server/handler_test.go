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
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func init() {
	fmt.Println("+++++ init handler_test +++++")
}

var testRouter *mux.Router

// Basic mux.Router tests
func TestRoutes(t *testing.T) {
	initCount := countRoutes(NewRouter())

	// Add couple of test handlers

	AddRoute("one", "GET", "/test/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	})

	AddRoute("two", "GET", "/test/2", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(2)
	})

	SetUIDirectory("/tmp/ui") // !!?
	testRouter = NewRouter()
	newCount := countRoutes(testRouter)

	if newCount != initCount+2 {
		t.Fatalf("Expected route count %d; found %d", initCount+2, newCount)
	}

	// Try the test URLs and an unknown URL. The unknonw path
	// should return 404
	t.Run("Get1", testGet("/test/1", 1))
	t.Run("Get2", testGet("/test/2", 2))
	t.Run("GetUnknown", testGet("/test/unknown", 404))
	t.Run("Meta", testGet("/.well-known/host-meta", 200))

	// Try the test URLs with authentication enabled.. This should
	// fail the requests with 401 error. Unknown path should still
	// return 404.
	ClientAuth.Set("password")
	testRouter = NewRouter()
	t.Run("Get1_auth", testGet("/test/1", 401))
	t.Run("Get2_auth", testGet("/test/2", 401))
	t.Run("GetUnknown_auth", testGet("/test/unknown", 404))

	// Meta handler should not be affected by user auth
	t.Run("Meta_auth", testGet("/.well-known/host-meta", 200))

	// Cleanup for next tests
	ClientAuth.Unset("password")
	testRouter = nil
}

// countRoutes counts the registered routes in a mux.Router
// object by walking it
func countRoutes(r *mux.Router) int {
	var count int
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		count++
		return nil
	})

	return count
}

// Try the url and check response code
func testGet(url string, expStatus int) func(*testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, httptest.NewRequest("GET", url, nil))
		if w.Code != expStatus {
			t.Fatalf("Expected response code %d; found %d", expStatus, w.Code)
		}
	}
}

func TestMetadataHandler(t *testing.T) {
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

// Test REST to Translib path conversions
func TestPathConv(t *testing.T) {

	t.Run("novar", testPathConv(
		"/simple/url/with/no/vars",
		"/simple/url/with/no/vars",
		"/simple/url/with/no/vars"))

	t.Run("1var", testPathConv(
		"/sample/id={name}",
		"/sample/id=TEST1",
		"/sample/id[name=TEST1]"))

	t.Run("1var_no=", testPathConv(
		"/sample/{name}",
		"/sample/TEST1",
		"/sample/[name=TEST1]"))

	t.Run("1var_middle", testPathConv(
		"/sample/id={name}/test/suffix",
		"/sample/id=TEST1/test/suffix",
		"/sample/id[name=TEST1]/test/suffix"))

	t.Run("2vars", testPathConv(
		"/sample/id={name},{type}",
		"/sample/id=TEST2,NEW",
		"/sample/id[name=TEST2][type=NEW]"))

	t.Run("2vars_middle", testPathConv(
		"/sample/id={name},{type}/hey",
		"/sample/id=TEST2,NEW/hey",
		"/sample/id[name=TEST2][type=NEW]/hey"))

	t.Run("5vars", testPathConv(
		"/sample/key={name},{type},{subtype},{color},{ver}",
		"/sample/key=TEST2,NEW,LATEST,RED,1.0",
		"/sample/key[name=TEST2][type=NEW][subtype=LATEST][color=RED][ver=1.0]"))

	t.Run("5vars_no=", testPathConv(
		"/sample/{name},{type},{subtype},{color},{ver}",
		"/sample/TEST2,NEW,LATEST,RED,1.0",
		"/sample/[name=TEST2][type=NEW][subtype=LATEST][color=RED][ver=1.0]"))

	t.Run("multi", testPathConv(
		"/sample/id={name},{type},{subtype}/data/color={colorname},{rgb}/{ver}",
		"/sample/id=TEST2,NEW,LATEST/data/color=RED,ff0000/1.0",
		"/sample/id[name=TEST2][type=NEW][subtype=LATEST]/data/color[colorname=RED][rgb=ff0000]/[ver=1.0]"))

	t.Run("rcdata_novar", testPathConv(
		"/restconf/data/no/vars",
		"/restconf/data/no/vars",
		"/no/vars"))

	t.Run("xrcdata_novar", testPathConv(
		"/myroot/restconf/data/no/vars",
		"/myroot/restconf/data/no/vars",
		"/no/vars"))

	t.Run("rcdata_1var", testPathConv(
		"/restconf/data/id={name}",
		"/restconf/data/id=TEST1",
		"/id[name=TEST1]"))

	t.Run("xrcdata_1var", testPathConv(
		"/myroot/restconf/data/id={name}",
		"/myroot/restconf/data/id=TEST1",
		"/id[name=TEST1]"))

	t.Run("no_template", testPathConv(
		"*",
		"/test/id=NOTEMPLATE",
		"/test/id=NOTEMPLATE"))

	t.Run("empty_params", testPathConv2(
		map[string]string{},
		"/test/id={name}",
		"/test/id=X",
		"/test/id[name=X]"))

	t.Run("1param", testPathConv2(
		map[string]string{"name1": "name"},
		"/test/id={name1}",
		"/test/id=X",
		"/test/id[name=X]"))

	t.Run("nparams", testPathConv2(
		map[string]string{"name1": "name", "name2": "name"},
		"/test/id={name1}/data/ref={name2}",
		"/test/id=X/data/ref=Y",
		"/test/id[name=X]/data/ref[name=Y]"))

	t.Run("extra_params", testPathConv2(
		map[string]string{"name1": "name", "name2": "name"},
		"/test/id={name1}",
		"/test/id=X",
		"/test/id[name=X]"))

	t.Run("escaped", testPathConv(
		"/test/interface={name}/ip={addr}",
		"/test/interface=Ethernet%200%2f1/ip=10.0.0.1%2f24",
		"/test/interface[name=Ethernet 0/1]/ip[addr=10.0.0.1/24]"))

	t.Run("escaped2", testPathConv(
		"/test/interface={name},{ip}",
		"/test/interface=Eth0%2f1%5b2%5c%5d,1::1",
		"/test/interface[name=Eth0/1[2\\\\\\]][ip=1::1]"))

	t.Run("escaped+param", testPathConv2(
		map[string]string{"name1": "name"},
		"/test/interface={name1},{type}",
		"/test/interface=Eth0%2f1:1,PHY",
		"/test/interface[name=Eth0/1:1][type=PHY]"))

}

// test handler to invoke getPathForTranslib and write the conveted
// path into response. Conversion logic depends on context values
// managed by mux router. Hence should be called from a handler.
var pathConvHandler = func(w http.ResponseWriter, r *http.Request) {
	// t, err := mux.CurrentRoute(r).GetPathTemplate()
	// fmt.Printf("Patt : %v (err=%v)\n", t, err)
	// fmt.Printf("Vars : %v\n", mux.Vars(r))

	w.Write([]byte(getPathForTranslib(r)))
}

func testPathConv(template, path, expPath string) func(*testing.T) {
	return testPathConv2(nil, template, path, expPath)
}

func testPathConv2(m map[string]string, template, path, expPath string) func(*testing.T) {
	return func(t *testing.T) {
		router := NewRouter() //mux.NewRouter()
		if template == "*" {
			t.Logf("No template...")
			router.Methods("GET").HandlerFunc(pathConvHandler)
		} else {
			router.HandleFunc(template, pathConvHandler)
		}

		r := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()

		if m != nil {
			rc, r1 := GetContext(r)
			rc.PMap = m
			r = r1
		}

		router.ServeHTTP(w, r)

		convPath := w.Body.String()
		if convPath != expPath {
			t.Logf("Conversion for template '%s' failed", template)
			t.Logf("Input path '%s'", path)
			t.Logf("Converted  '%s'", convPath)
			t.Logf("Expected   '%s'", expPath)
			t.FailNow()
		}
	}
}

type errReader string

func (er errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New(string(er))
}

func TestReqData_NoBody(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	rc := &RequestContext{ID: t.Name()}

	ct, data, err := getRequestBody(r, rc)
	if ct != nil || data != nil || err != nil {
		t.Fatalf("Expected nil response; found ct=%v, data=%v, err=%v", ct, data, err)
	}
}

func TestReqData_ReadFailure(t *testing.T) {
	r := httptest.NewRequest("PUT", "/test", errReader("e-r-r-o-r"))
	rc := &RequestContext{ID: t.Name()}

	testReqError(t, r, rc, 500)
}

func TestReqData_Unknown(t *testing.T) {
	r := httptest.NewRequest("PUT", "/test", strings.NewReader("Hello, world!"))
	rc := &RequestContext{ID: t.Name()}

	testReqError(t, r, rc, 415)
}

func TestReqData_Unknown2(t *testing.T) {
	r := httptest.NewRequest("PUT", "/test", strings.NewReader("Hello, world!"))
	rc := &RequestContext{ID: t.Name()}
	rc.Consumes.Add("text/html")

	testReqError(t, r, rc, 415)
}

func TestReqData_BadMime(t *testing.T) {
	r := httptest.NewRequest("PUT", "/test", strings.NewReader("Hello, world!"))
	r.Header.Set("content-type", "b a d")
	rc := &RequestContext{ID: t.Name()}
	rc.Consumes.Add("b a d")

	testReqError(t, r, rc, 400)
}

func TestReqData_Text(t *testing.T) {
	r := httptest.NewRequest("PUT", "/test", strings.NewReader("Hello, world!"))
	rc := &RequestContext{ID: t.Name()}
	rc.Consumes.Add("text/plain")

	testReqSuccess(t, r, rc, "text/plain", "Hello, world!")
}

func TestReqData_Json(t *testing.T) {
	input := "{\"one\":1}"
	r := httptest.NewRequest("PUT", "/test", strings.NewReader(input))
	r.Header.Set("content-type", "application/json")

	rc := &RequestContext{ID: t.Name()}
	rc.Consumes.Add("text/html")
	rc.Consumes.Add("text/plain")
	rc.Consumes.Add("application/json")

	testReqSuccess(t, r, rc, "application/json", input)
}

func TestReqData_BadJsonNoValidation(t *testing.T) {
	input := "{\"one:1}"
	r := httptest.NewRequest("PUT", "/test", strings.NewReader(input))
	r.Header.Set("content-type", "application/json")

	rc := &RequestContext{ID: t.Name()}
	rc.Consumes.Add("application/json")

	testReqSuccess(t, r, rc, "application/json", input)
}

func TestReqData_BadJsonWithValidation(t *testing.T) {
	input := "{\"one:1}"
	r := httptest.NewRequest("PUT", "/test", strings.NewReader(input))
	r.Header.Set("content-type", "application/json")

	model := make(map[string]int)
	rc := &RequestContext{ID: t.Name(), Model: &model}
	rc.Consumes.Add("application/json")

	testReqError(t, r, rc, 400)
}

func testReqSuccess(t *testing.T, r *http.Request, rc *RequestContext, expType, expData string) {
	ct, data, err := getRequestBody(r, rc)

	if ct == nil || ct.Type != expType {
		t.Fatalf("Expected %s; found %s", expType, ct.Type)
	}
	if data == nil || string(data) != expData {
		t.Fatalf("Expected data \"%s\"; found \"%s\"", expData, data)
	}
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testReqError(t *testing.T, r *http.Request, rc *RequestContext, expCode int) {
	ct, data, err := getRequestBody(r, rc)

	if ct != nil {
		t.Fatalf("Expected nil content-type; found %s", ct.Type)
	}
	if data != nil {
		t.Fatalf("Expected nil data; found \"%s\"", data)
	}

	he, ok := err.(httpErrorType)
	if !ok {
		t.Fatalf("Expecting httpErrorType; got %T", err)
	}
	if he.status != expCode {
		t.Fatalf("Expecting http status %d; got %d", expCode, he.status)
	}
}

func TestRespData_NoContent(t *testing.T) {
	rc := &RequestContext{ID: t.Name()}
	t.Run("nil", testRespData(nil, rc, nil, ""))
	t.Run("empty", testRespData(nil, rc, []byte(""), ""))
}

func TestRespData_NoProduces(t *testing.T) {
	rc := &RequestContext{ID: t.Name()}
	t.Run("txt", testRespData(nil, rc, []byte("Hello, world!"), "text/plain"))
	t.Run("bin", testRespData(nil, rc, make([]byte, 5), "application/octet-stream"))
}

func TestRespData_1Produces(t *testing.T) {
	rc := &RequestContext{ID: t.Name()}
	rc.Produces.Add("application/json")

	t.Run("jsn", testRespData(nil, rc, []byte("{}"), "application/json"))
	t.Run("bin", testRespData(nil, rc, make([]byte, 5), "application/json"))
}

func TestRespData_nProduces(t *testing.T) {
	rc := &RequestContext{ID: t.Name()}
	rc.Produces.Add("application/json")
	rc.Produces.Add("application/xml")
	rc.Produces.Add("text/plain")

	t.Run("jsn", testRespData(nil, rc, []byte("{}"), "text/plain"))
	t.Run("bin", testRespData(nil, rc, make([]byte, 5), "application/octet-stream"))
}

func testRespData(r *http.Request, rc *RequestContext, data []byte, expType string) func(*testing.T) {
	return func(t *testing.T) {
		if r == nil {
			r = httptest.NewRequest("GET", "/get", nil)
		}

		ctype, err := resolveResponseContentType(data, r, rc)
		ct, err := parseMediaType(ctype)

		if (expType == "" && ctype != "") || (ct != nil && ct.Type != expType) {
			t.Fatalf("Expected resp content-type \"%s\"; got \"%s\"", expType, ctype)
		}
		if err != nil {
			t.Fatalf("Unexpected error %v", err)
		}
	}
}

func TestProcessGET(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "GET", "/api-tests:sample", ""))
	verifyResponse(t, w, 200)
}

func TestProcessGET_error(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "GET", "/api-tests:sample/error/not-found", ""))
	verifyResponse(t, w, 404)
}

func TestProcessPUT(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "PUT", "/api-tests:sample", "{}"))
	verifyResponse(t, w, 204)
}

func TestProcessPUT_error(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "PUT", "/api-tests:sample/error/not-supported", "{}"))
	verifyResponse(t, w, 405)
}

func TestProcessPOST(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "POST", "/api-tests:sample", "{}"))
	verifyResponse(t, w, 201)
}

func TestProcessPOST_error(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "POST", "/api-tests:sample/error/invalid-args", "{}"))
	verifyResponse(t, w, 400)
}

func TestProcessPATCH(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "PATCH", "/api-tests:sample", "{}"))
	verifyResponse(t, w, 204)
}

func TestProcessPATCH_error(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "PATCH", "/api-tests:sample/error/unknown", "{}"))
	verifyResponse(t, w, 500)
}

func TestProcessDELETE(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "DELETE", "/api-tests:sample", ""))
	verifyResponse(t, w, 204)
}

func TestProcessDELETE_error(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "DELETE", "/api-tests:sample/error/not-found", ""))
	verifyResponse(t, w, 404)
}

func TestProcessRPC(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "POST", "/restconf/operations/api-tests:my-echo",
		"{\"/api-tests:input\":{\"message\":\"Hii\"}}"))
	verifyResponse(t, w, 200)
}

func TestProcessRPC_error(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "POST", "/restconf/operations/api-tests:my-echo",
		"{\"api-tests:input\":{\"error-type\":\"not-supported\"}}"))
	verifyResponse(t, w, 405)
}

func TestProcessBadMethod(t *testing.T) {
	w := httptest.NewRecorder()
	Process(w, prepareRequest(t, "TEST", "/test", "{}"))
	verifyResponse(t, w, 400)
}

func TestProcessBadContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := prepareRequest(t, "PUT", "/test", "{}")
	r.Header.Set("content-type", "bad/content")

	Process(w, r)
	verifyResponse(t, w, 415)
}

func TestProcessReadError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/test", errReader("simulated error"))
	r.Header.Set("content-type", "application/json")

	rc, r := GetContext(r)
	rc.ID = t.Name()
	rc.Consumes.Add("application/json")

	Process(w, r)
	verifyResponse(t, w, 500)
}

func prepareRequest(t *testing.T, method, path, data string) *http.Request {
	if !strings.Contains(path, "/restconf/") {
		path = "/restconf/data" + path
	}

	r := httptest.NewRequest(method, path, strings.NewReader(data))
	rc, r := GetContext(r)
	rc.ID = t.Name()

	if data != "" {
		r.Header.Set("content-type", "application/json")
		rc.Consumes.Add("application/json")
	} else {
		rc.Produces.Add("application/json")
	}

	return r
}

func verifyResponse(t *testing.T, w *httptest.ResponseRecorder, expCode int) {
	if w.Code != expCode {
		t.Fatalf("Expecting response status %d; got %d", expCode, w.Code)
	}
}
