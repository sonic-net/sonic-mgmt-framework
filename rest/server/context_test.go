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
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func init() {
	fmt.Println("+++++ init context_test +++++")
}

func TestGetContext(t *testing.T) {
	r, err := http.NewRequest("GET", "/index.html", nil)
	if err != nil {
		t.Fatalf("Unexpected error; %v", err)
	}

	idSufix := fmt.Sprintf("-%v", requestCounter+1)

	rc1, r := GetContext(r)
	rc2, r := GetContext(r)
	rc3, r := GetContext(r)

	if rc1 != rc2 || rc1 != rc3 {
		t.Fatalf("Got duplicate contexts!!")
	}

	if !strings.HasSuffix(rc1.ID, idSufix) {
		t.Fatalf("Unexpected id '%s'; expected suffix '%s", rc1.ID, idSufix)
	}
}

func TestParseEmptyMtype(t *testing.T) {
	m, e := parseMediaType("")
	if m != nil || e != nil {
		t.Errorf("Unexpected return values; m=%v, e=%v", m, e)
	}
}

func TestParseMtype(t *testing.T) {
	t.Run("X/Y", testParseMtype("application/json",
		"application/json", "application", "", "json", mkmap()))

	t.Run("X/Y+Z", testParseMtype("application/yang-data+json",
		"application/yang-data+json", "application", "yang-data", "json", mkmap()))

	t.Run("X/Z; A=B", testParseMtype("application/xml; q=5",
		"application/xml", "application", "", "xml", mkmap("q", "5")))

	t.Run("X/Z; A=B; C=D", testParseMtype("application/xml; q=5; ver=0.1",
		"application/xml", "application", "", "xml", mkmap("q", "5", "ver", "0.1")))

	t.Run("*/*", testParseMtype("*/*",
		"*/*", "*", "*", "*", mkmap()))

	t.Run("text/*", testParseMtype("text/*",
		"text/*", "text", "*", "*", mkmap()))

	t.Run("*/xml", testParseMtype("*/xml",
		"*/xml", "*", "", "xml", mkmap()))

	t.Run("text/*+xml", testParseMtype("text/*+xml",
		"text/*+xml", "text", "*", "xml", mkmap()))

	// invalid media types
	t.Run("Partial", testBadMtype("application/"))
	t.Run("WithSpace", testBadMtype("app lication/json"))
	t.Run("WithSpace2", testBadMtype("application/ json"))
	t.Run("BadParam", testBadMtype("application/json;q=10 x=20"))
}

func testParseMtype(v, mimeType, prefix, middle, suffix string, params map[string]string) func(*testing.T) {
	return func(t *testing.T) {
		m := toMtype(t, v)
		if m.Type != mimeType || m.TypePrefix != prefix ||
			m.TypeMiddle != middle || m.TypeSuffix != suffix ||
			reflect.DeepEqual(m.Params, params) == false {
			t.Fatalf("\"%s\" did not tokenize into mime=\"%s\", prefix=\"%s\", middle=\"%s\", suffix=\"%s\", params=%v",
				v, mimeType, prefix, middle, suffix, params)
		}
		if m.Format() != v {
			t.Logf("Could not reconstruct \"%s\" from %v", v, m)
		}
	}
}

func testBadMtype(v string) func(*testing.T) {
	return func(t *testing.T) {
		_, e := parseMediaType(v)
		if e == nil {
			t.Errorf("Invalid media type \"%s\" not flagged", v)
		}
	}
}

func toMtype(t *testing.T, v string) *MediaType {
	m, e := parseMediaType(v)
	if e != nil {
		t.Fatalf("Bad media type \"%s\"; err=%v", v, e)
	}
	return m
}

func mkmap(args ...string) map[string]string {
	m := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		m[args[i]] = args[i+1]
	}
	return m
}

func TestMtypeMatch(t *testing.T) {
	t.Run("A/B=~A/B", testMtypeMatch("text/json", "text/json", true))
	t.Run("A/B!~A/C", testMtypeMatch("text/json", "text/xml", false))
	t.Run("A/B!~C/B", testMtypeMatch("text/json", "new/json", false))
	t.Run("A/B=~*/*", testMtypeMatch("text/json", "*/*", true))
	t.Run("A/B=~A/*", testMtypeMatch("text/json", "text/*", true))
	t.Run("A/B=~*/B", testMtypeMatch("text/json", "*/json", true))
	t.Run("A/B!~*/C+B", testMtypeMatch("text/json", "*/new+json", false))
	t.Run("A/B!~*/B+*", testMtypeMatch("text/json", "*/json+*", false))
	t.Run("A/B=~A/*+B", testMtypeMatch("text/json", "text/*+json", true))
	t.Run("A/B=~A/*+*", testMtypeMatch("text/json", "text/*+*", true))
	t.Run("A/V+B=~A/V+B", testMtypeMatch("text/v1+json", "text/v1+json", true))
	t.Run("A/V+B=~A/V+*", testMtypeMatch("text/v1+json", "text/v1+*", true))
	t.Run("A/V+B=~A/*+B", testMtypeMatch("text/v1+json", "text/*+json", true))
	t.Run("A/V+B=~A/*", testMtypeMatch("text/v1+json", "text/*", true))
	t.Run("A/V+B=~*/*", testMtypeMatch("text/v1+json", "*/*", true))
	t.Run("A/V+B!~A/B", testMtypeMatch("text/v1+json", "text/json", false))
}

func testMtypeMatch(lhs, rhs string, exp bool) func(*testing.T) {
	return func(t *testing.T) {
		x := toMtype(t, lhs)
		y := toMtype(t, rhs)
		if x.Matches(y) != exp {
			t.Fatalf("condition failed: \"%s\" match \"%s\" == %v", lhs, rhs, exp)
		}
	}
}

func TestIsJSON(t *testing.T) {
	t.Run("A/json", testIsJSON("text/json", true))
	t.Run("A/V+json", testIsJSON("text/v1+json", true))
	t.Run("A/json+V", testIsJSON("text/json+V", false))
	t.Run("A/xml", testIsJSON("text/xml", false))
	t.Run("json/A", testIsJSON("json/text", false))
}

func testIsJSON(v string, exp bool) func(*testing.T) {
	return func(t *testing.T) {
		m := toMtype(t, v)
		if m.isJSON() != exp {
			t.Fatalf("condition failed: isJson(\"%s\") == %v", v, exp)
		}
	}
}

func TestMtypes(t *testing.T) {
	var m MediaTypes

	// add 3 values to MediaTypes
	m.Add("text/xml; q=5")
	m.Add("text/json; q=2; r=3")
	m.Add("text/v1+json")

	// check length
	if len(m) != 3 {
		t.Fatalf("Expected 3 items; found %d", len(m))
	}

	// check String()
	expStrValue := "[text/xml text/json text/v1+json]"
	if m.String() != expStrValue {
		t.Fatalf("String() check failed.. expected \"%s\"; found \"%s\"", expStrValue, m)
	}

	// check Contains()
	t.Run("Contains#1", testMtypesContains(m, "text/xml", true))
	t.Run("Contains#2", testMtypesContains(m, "text/json", true))
	t.Run("Contains#3", testMtypesContains(m, "text/v1+json", true))
	t.Run("NotContains", testMtypesContains(m, "text/plain", false))

	// check GetMatching()
	t.Run("Match 1", testMtypesGetMatching(m, "text/xml", "[text/xml]"))
	t.Run("Match 2", testMtypesGetMatching(m, "text/*+json", "[text/json text/v1+json]"))
	t.Run("Match 0", testMtypesGetMatching(m, "text/plain", "[]"))
	t.Run("Match Err", testMtypesGetMatching(m, "text plain", "[]"))
}

func testMtypesContains(m MediaTypes, v string, exp bool) func(*testing.T) {
	return func(t *testing.T) {
		if m.Contains(v) != exp {
			t.Fatalf("condition failed: %v.Contains(\"%s\") == %v", m, v, exp)
		}
	}
}

func testMtypesGetMatching(m MediaTypes, v, exp string) func(*testing.T) {
	return func(t *testing.T) {
		m1 := m.GetMatching(v)
		if m1.String() != exp {
			t.Logf("Items matching \"%s\" from %s are %s", v, m, m1)
			t.Fatalf("Expected %s", exp)
		}
	}
}
