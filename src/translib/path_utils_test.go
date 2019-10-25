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

package translib

import (
	"reflect"
	"strings"
	"testing"

	"translib/ocbinds"
)

func TestGetParentNode(t *testing.T) {

	tests := []struct {
		tid         int
		targetUri   string
		appRootType reflect.Type
		want        string
	}{{
		tid:         1,
		targetUri:   "/openconfig-acl:acl/acl-sets/",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl",
	}, {
		tid:         2,
		targetUri:   "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=MyACL1][type=ACL_IPV4]/",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets",
	}, {
		tid:         3,
		targetUri:   "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample][type=ACL_IPV4]/state/description",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets_AclSet_State",
	}}

	for _, tt := range tests {
		var deviceObj ocbinds.Device = ocbinds.Device{}
		_, err := getRequestBinder(&tt.targetUri, nil, 1, &tt.appRootType).unMarshallUri(&deviceObj)
		if err != nil {
			t.Error("TestGetParentNode: Error in unmarshalling the URI", err)
		} else {
			parentNode, _, err := getParentNode(&tt.targetUri, &deviceObj)
			if err != nil {
				t.Error("TestGetParentNode: Error in getting the parent node: ", err)
			} else if parentNode == nil {
				t.Error("TestGetParentNode: Error in getting the parent node")
			} else if reflect.TypeOf(*parentNode).Elem().Name() != tt.want {
				t.Error("TestGetParentNode: Error in getting the parent node: ", reflect.TypeOf(*parentNode).Elem().Name())
			}
		}
	}
}

func TestGetNodeName(t *testing.T) {

	tests := []struct {
		tid         int
		targetUri   string
		appRootType reflect.Type
		want        string
	}{{
		tid:         1,
		targetUri:   "/openconfig-acl:acl/acl-sets/",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "acl-sets",
	}, {
		tid:         2,
		targetUri:   "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=MyACL1][type=ACL_IPV4]/",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "acl-set",
	}, {
		tid:         3,
		targetUri:   "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample][type=ACL_IPV4]/state/description",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "description",
	}, {
		tid:         4, // Negative test case
		targetUri:   "/openconfig-acl:acl/acl-sets/acl-set[name=Sample][type=ACL_IPV4]/state/descriptXX",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "rpc error: code = InvalidArgument desc = no match found in *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_State, for path elem:<name:\"descriptXX",
	}}

	for _, tt := range tests {
		var deviceObj ocbinds.Device = ocbinds.Device{}
		_, err := getRequestBinder(&tt.targetUri, nil, 1, &tt.appRootType).unMarshallUri(&deviceObj)
		if err != nil {
			if strings.Contains(err.Error(), tt.want) == false {
				t.Error("TestGetNodeName: Error in unmarshalling the URI", err)
			}
		} else {
			nodeName, err := getNodeName(&tt.targetUri, &deviceObj)
			if err != nil {
				t.Error("TestGetNodeName: Error in getting the yang node name: ", err)
			} else if nodeName != tt.want {
				t.Error("TestGetNodeName: Error in getting the yang node name: ", nodeName)
			}
		}
	}
}

func TestGetObjectFieldName(t *testing.T) {

	tests := []struct {
		tid         int
		targetUri   string
		appRootType reflect.Type
		want        string
	}{{
		tid:         1,
		targetUri:   "/openconfig-acl:acl/acl-sets/",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "AclSets",
	}, {
		tid:         2,
		targetUri:   "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL1][type=ACL_IPV4]/state",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "State",
	}, {
		tid:         3,
		targetUri:   "/openconfig-acl:acl/acl-sets/acl-set[name=Sample][type=ACL_IPV4]/state/description",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "Description",
	}, {
		tid:         4, // Negative test case
		targetUri:   "/openconfig-acl:acl/acl-sets/acl-set[name=Sample][type=ACL_IPV4]/state/descriptXX",
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "rpc error: code = InvalidArgument desc = no match found in *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_State, for path elem:<name:\"descriptXX",
	}}

	for _, tt := range tests {
		var deviceObj ocbinds.Device = ocbinds.Device{}
		targetObj, err := getRequestBinder(&tt.targetUri, nil, 1, &tt.appRootType).unMarshallUri(&deviceObj)
		if err != nil {
			if strings.Contains(err.Error(), tt.want) == false {
				t.Error("TestGetObjectFieldName: Error in unmarshalling the URI", err)
			}
		} else {
			objFieldName, err := getObjectFieldName(&tt.targetUri, &deviceObj, targetObj)
			if err != nil {
				t.Error("TestGetObjectFieldName: Error in getting the ygot struct object field name: ", err)
			} else if objFieldName != tt.want {
				t.Error("TestGetObjectFieldName: Error in getting the ygot struct object field name:  ", objFieldName)
			}
		}
	}
}

func TestNewPathInfo_empty(t *testing.T) {
	testPathInfo(t, "", "", mkmap())
}

func TestNewPathInfo_novar(t *testing.T) {
	testPathInfo(t, "/test/simple", "/test/simple", mkmap())
}

func TestNewPathInfo_var1(t *testing.T) {
	testPathInfo(t, "/test/xx[one=1]", "/test/xx{}", mkmap("one", "1"))
}

func TestNewPathInfo_vars(t *testing.T) {
	testPathInfo(t, "/test/xx[one=1][two=2]/new[three=3]", "/test/xx{}{}/new{}",
		mkmap("one", "1", "two", "2", "three", "3"))
}

func TestNewPathInfo_dup1(t *testing.T) {
	testPathInfo(t, "/test/xx[one=1][two=2]/new[one=0001]", "/test/xx{}{}/new{}",
		mkmap("one", "1", "two", "2", "one#2", "0001"))
}

func TestNewPathInfo_dups(t *testing.T) {
	testPathInfo(t, "/test/one[xx=1]/two[yy=2]/three[xx=3]/four[zz=4]/five[yy=5]/six[xx=6]",
		"/test/one{}/two{}/three{}/four{}/five{}/six{}",
		mkmap("xx", "1", "yy", "2", "xx#2", "3", "zz", "4", "yy#2", "5", "xx#3", "6"))
}

func testPathInfo(t *testing.T, path, expTemplate string, expVars map[string]string) {
	info := NewPathInfo(path)
	if info == nil {
		t.Errorf("NewPathInfo() returned null!")
	} else if info.Path != path {
		t.Errorf("Expected info.Path = %s", path)
		t.Errorf("Actual info.Path   = %s", info.Path)
	} else if info.Template != expTemplate {
		t.Errorf("Expected info.Template = %s", expTemplate)
		t.Errorf("Actual info.Template   = %s", info.Template)
	} else if reflect.DeepEqual(info.Vars, expVars) == false {
		t.Errorf("Expected info.Vars = %v", expVars)
		t.Errorf("Actual info.Vars   = %v", info.Vars)
	}
	if t.Failed() {
		t.Fatalf("NewPathInfo() failed to parse \"%s\"", path)
	}
}

func mkmap(args ...string) map[string]string {
	m := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		m[args[i]] = args[i+1]
	}
	return m
}
