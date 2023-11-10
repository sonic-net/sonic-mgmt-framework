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
	"reflect"
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
		if p.content != exp.content {
			t.Errorf("'content' mismatch; expecting %s, found %s", exp.content, p.content)
		}
		if !reflect.DeepEqual(p.fields, exp.fields) {
			t.Errorf("fields mismatch; expecting %s, found %s", exp.fields, p.fields)
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

func testGetQuery(t *testing.T, name, queryStr string, exp *translibArgs) {
	t.Run("GET/"+name, testQuery("GET", queryStr, exp))
	t.Run("HEAD/"+name, testQuery("HEAD", queryStr, exp))
}

func TestQuery_depth(t *testing.T) {
	rcCaps := restconfCapabilities
	defer func() { restconfCapabilities = rcCaps }()

	restconfCapabilities.depth = true

	// run depth test cases for GET and HEAD
	testGetQuery(t, "=unbounded", "depth=unbounded", &translibArgs{depth: 0})
	testGetQuery(t, "=0", "depth=0", nil)
	testGetQuery(t, "=1", "depth=1", &translibArgs{depth: 1})
	testGetQuery(t, "=101", "depth=101", &translibArgs{depth: 101})
	testGetQuery(t, "=65535", "depth=65535", &translibArgs{depth: 65535})
	testGetQuery(t, "=65536", "depth=65536", nil)
	testGetQuery(t, "=junk", "depth=junk", nil)
	testGetQuery(t, "extra", "depth=1&extra=1", nil)
	testGetQuery(t, "dup", "depth=1&depth=2", nil)

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

	testGetQuery(t, "100", "depth=100", nil)
	restconfCapabilities.depth = true
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

func TestQuery_content(t *testing.T) {
	rcCaps := restconfCapabilities
	defer func() { restconfCapabilities = rcCaps }()

	restconfCapabilities.content = true

	// run content query test cases for GET and HEAD
	testGetQuery(t, "=all", "content=all", &translibArgs{content: "all"})
	testGetQuery(t, "=ALL", "content=ALL", nil)
	testGetQuery(t, "=config", "content=config", &translibArgs{content: "config"})
	testGetQuery(t, "=Config", "content=Config", nil)
	testGetQuery(t, "=nonconfig", "content=nonconfig", &translibArgs{content: "nonconfig"})
	testGetQuery(t, "=NonConfig", "content=NonConfig", nil)
	testGetQuery(t, "=getall", "content=getall", nil)
	testGetQuery(t, "=operational", "content=operational", nil)
	testGetQuery(t, "=state", "content=state", nil)
	testGetQuery(t, "=0", "content=0", nil)
	testGetQuery(t, "dup", "content=config&content=nonconfig", nil)

	// check for other methods
	t.Run("OPTIONS", testQuery("OPTIONS", "content=config", nil))
	t.Run("PUT", testQuery("PUT", "content=config", nil))
	t.Run("POST", testQuery("POST", "content=config", nil))
	t.Run("PATCH", testQuery("PATCH", "content=config", nil))
	t.Run("DELETE", testQuery("DELETE", "content=config", nil))
}

func TestQuery_content_disabled(t *testing.T) {
	rcCaps := restconfCapabilities
	defer func() { restconfCapabilities = rcCaps }()

	restconfCapabilities.content = false
	testGetQuery(t, "config", "content=config", nil)
	restconfCapabilities.content = true
}

func TestQuery_fields(t *testing.T) {
	rcCaps := restconfCapabilities
	defer func() { restconfCapabilities = rcCaps }()

	restconfCapabilities.depth = true

	// run depth test cases for GET and HEAD
	testGetQuery(t, "testfield1", "fields=description", &translibArgs{fields: []string{"description"}})
	testGetQuery(t, "testfield2", "fields=description;mtu", &translibArgs{fields: []string{"description", "mtu"}})
	testGetQuery(t, "testfield3", "fields=description,mtu", &translibArgs{fields: []string{"description,mtu"}})
	testGetQuery(t, "testfield4", "fields=config/description;mtu", &translibArgs{fields: []string{"config/description", "mtu"}})
	testGetQuery(t, "testfield4", "fields=config/description,mtu", &translibArgs{fields: []string{"config/description,mtu"}})
	testGetQuery(t, "testfield5", "fields=config(description;mtu)", &translibArgs{fields: []string{"config/description", "config/mtu"}})
	testGetQuery(t, "testfield6", "fields=config(description;mtu);state", &translibArgs{fields: []string{"config/description", "config/mtu", "state"}})
	testGetQuery(t, "testfield7", "fields=config(description;mtu),state", &translibArgs{fields: []string{"config/description", "config/mtu", ",state"}})
	testGetQuery(t, "testfield8", "fields=config(description,mtu),state", &translibArgs{fields: []string{"config/description,mtu", ",state"}})
	testGetQuery(t, "testfield9", "fields=config(description;mtu);state/mtu", &translibArgs{fields: []string{"config/description", "config/mtu", "state/mtu"}})
	testGetQuery(t, "testfield10", "fields=config(description;mtu);state(mtu)", &translibArgs{fields: []string{"config/description", "config/mtu", "state/mtu"}})
	testGetQuery(t, "testfield11", "fields=config(description;mtu);state(mtu;counters)", &translibArgs{fields: []string{"config/description", "config/mtu", "state/mtu", "state/counters"}})
	testGetQuery(t, "testfield12", "fields=config(description;mtu)&state", nil)
	testGetQuery(t, "testfield13", "fields=config(description,mtu)&state=test", nil)
	testGetQuery(t, "testfield14", "fields=config/mtu@state", &translibArgs{fields: []string{"config/mtu@state"}})
	testGetQuery(t, "testfield15", "fields=config(description,mtu)@state", &translibArgs{fields: []string{"config/description,mtu", "@state"}})
	testGetQuery(t, "testfield16", "fields=mtu&depth=2", nil)
	testGetQuery(t, "testfield17", "fields=mtu&content=all", nil)
	testGetQuery(t, "testfield18", "fields=mtu&depth=2&content=all", nil)

	// check for other methods
	t.Run("OPTIONS", testQuery("OPTIONS", "fields=mtu", nil))
	t.Run("PUT", testQuery("PUT", "fields=mtu", nil))
	t.Run("POST", testQuery("POST", "fields=mtu", nil))
	t.Run("PATCH", testQuery("PATCH", "fields=mtu", nil))
	t.Run("DELETE", testQuery("DELETE", "fields=mtu", nil))
}

func TestQuery_fields_disabled(t *testing.T) {
	rcCaps := restconfCapabilities
	defer func() { restconfCapabilities = rcCaps }()

	restconfCapabilities.fields = false

	testGetQuery(t, "100", "fields=100", nil)
	restconfCapabilities.fields = true
}

func TestQuery_DepthContent(t *testing.T) {
	rcCaps := restconfCapabilities
	defer func() { restconfCapabilities = rcCaps }()

	restconfCapabilities.content = true
	restconfCapabilities.depth = true

	// run content and depth query test cases for GET and HEAD
	testGetQuery(t, "testDepConQuery1", "depth=unbounded&content=all", &translibArgs{content: "all", depth: 0})
	testGetQuery(t, "testDepConQuery2", "depth=1&content=all", &translibArgs{content: "all", depth: 1})
	testGetQuery(t, "testDepConQuery3", "depth=3&content=config", &translibArgs{content: "config", depth: 3})
	testGetQuery(t, "testDepConQuery4", "depth=4&content=nonconfig", &translibArgs{content: "nonconfig", depth: 4})
	testGetQuery(t, "testDepConQuery5", "depth=65535&content=nonconfig", &translibArgs{content: "nonconfig", depth: 65535})
	testGetQuery(t, "testDepConQuery6", "depth=65536&content=all", nil)
	testGetQuery(t, "testDepConQuery7", "depth=5&content=ALL", nil)
	testGetQuery(t, "testDepConQuery8", "depth=5&content=9", nil)
	testGetQuery(t, "testDepConQuery9", "depth=%$&content=state", nil)
	testGetQuery(t, "testDepConQuery10", "depth=3$&depth=2&content=all", nil)
	testGetQuery(t, "testDepConQuery11", "depth=3$&content=all&content=config", nil)
	testGetQuery(t, "testDepConQuery12", "depth=3$&content=all&content=config", nil)
	testGetQuery(t, "testDepConQuery13", "depth=3;content=all", nil)

	// check for other methods
	t.Run("OPTIONS", testQuery("OPTIONS", "depth=3&content=config", nil))
	t.Run("PUT", testQuery("PUT", "depth=3&content=config", nil))
	t.Run("POST", testQuery("POST", "depth=3&content=config", nil))
	t.Run("PATCH", testQuery("PATCH", "depth=3&content=config", nil))
	t.Run("DELETE", testQuery("DELETE", "depth=3&content=config", nil))
}

func TestQuery_depth_content_disabled(t *testing.T) {
	rcCaps := restconfCapabilities
	defer func() { restconfCapabilities = rcCaps }()

	restconfCapabilities.content = false
	restconfCapabilities.depth = true
	testGetQuery(t, "config", "depth=3&content=config", nil)

	restconfCapabilities.content = true
	restconfCapabilities.depth = false
	testGetQuery(t, "config", "depth=3&content=config", nil)

	restconfCapabilities.content = false
	restconfCapabilities.depth = false
	testGetQuery(t, "config", "depth=3&content=config", nil)

}
