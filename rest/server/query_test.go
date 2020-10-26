////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2020 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
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
	"net/http/httptest"
	"testing"
)

func testQuery(method, queryStr string, exp *translibArgs) func(*testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, "/restconf/data/querytest?"+queryStr, nil)
		_, r = GetContext(r)

		p := translibArgs{}
		err := p.parseQueryParams(r)

		errCode := 0
		if he, ok := err.(httpErrorType); ok {
			errCode = he.status
		}

		if exp == nil && errCode == 400 {
			return // success
		}
		if err != nil {
			t.Fatalf("Failed to process query '%s'; err=%d/%v", r.URL.RawQuery, errCode, err)
		}

		// compare parsed translibArgs
		if p.depth != exp.depth {
			t.Errorf("'depth' mismatch; expecting %d, found %d", exp.depth, p.depth)
		}
		if p.deleteEmpty != exp.deleteEmpty {
			t.Errorf("'deleteEmptyEntry' mismatch; expting %v, found %v", exp.deleteEmpty, p.deleteEmpty)
		}
		if t.Failed() {
			t.Errorf("Testcase failed for query '%s'", r.URL.RawQuery)
		}
	}
}

func TestQuery(t *testing.T) {
	t.Run("none", testQuery("GET", "", &translibArgs{}))
	t.Run("unknown", testQuery("GET", "one=1", nil))
}

func TestQuery_depth(t *testing.T) {
	rcCaps := restconfCapabilities
	defer func() { restconfCapabilities = rcCaps }()

	restconfCapabilities.depth = true

	// run depth test cases for GET and HEAD
	testDepth(t, "=unbounded", "depth=unbounded", &translibArgs{depth: 0})
	testDepth(t, "=0", "depth=0", nil)
	testDepth(t, "=1", "depth=1", &translibArgs{depth: 1})
	testDepth(t, "=101", "depth=101", &translibArgs{depth: 101})
	testDepth(t, "=65535", "depth=65535", &translibArgs{depth: 65535})
	testDepth(t, "=65536", "depth=65536", nil)
	testDepth(t, "=junk", "depth=junk", nil)
	testDepth(t, "extra", "depth=1&extra=1", nil)
	testDepth(t, "dup", "depth=1&depth=2", nil)

	// check for other methods
	t.Run("OPTIONS", testQuery("OPTIONS", "depth=1", nil))
	t.Run("PUT", testQuery("PUT", "depth=1", nil))
	t.Run("POST", testQuery("POST", "depth=1", nil))
	t.Run("PATCH", testQuery("PATCH", "depth=1", nil))
	t.Run("DELETE", testQuery("DELETE", "depth=1", nil))
}

func TestQuery_depth_disabled(t *testing.T) {
	rcCaps := restconfCapabilities
	defer func() { restconfCapabilities = rcCaps }()

	restconfCapabilities.depth = false

	testDepth(t, "100", "depth=100", nil)
}

func testDepth(t *testing.T, name, queryStr string, exp *translibArgs) {
	t.Run("GET/"+name, testQuery("GET", queryStr, exp))
	t.Run("HEAD/"+name, testQuery("HEAD", queryStr, exp))
}

func TestQuery_deleteEmptyEntry(t *testing.T) {
	t.Run("=true", testQuery("DELETE", "deleteEmptyEntry=true", &translibArgs{deleteEmpty: true}))
	t.Run("=True", testQuery("DELETE", "deleteEmptyEntry=True", &translibArgs{deleteEmpty: true}))
	t.Run("=TRUE", testQuery("DELETE", "deleteEmptyEntry=TRUE", &translibArgs{deleteEmpty: true}))
	t.Run("=false", testQuery("DELETE", "deleteEmptyEntry=false", &translibArgs{deleteEmpty: false}))
	t.Run("=1", testQuery("DELETE", "deleteEmptyEntry=1", &translibArgs{deleteEmpty: false}))
	t.Run("GET", testQuery("GET", "deleteEmptyEntry=true", nil))
	t.Run("HEAD", testQuery("HEAD", "deleteEmptyEntry=true", nil))
	t.Run("OPTIONS", testQuery("OPTIONS", "deleteEmptyEntry=true", nil))
	t.Run("PUT", testQuery("PUT", "deleteEmptyEntry=true", nil))
	t.Run("POST", testQuery("POST", "deleteEmptyEntry=true", nil))
	t.Run("PATCH", testQuery("PATCH", "deleteEmptyEntry=true", nil))
}
