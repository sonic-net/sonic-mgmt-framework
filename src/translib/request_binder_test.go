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
	"fmt"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"strings"
	"testing"
	"translib/ocbinds"
)

func TestInitSchema(t *testing.T) {
	initSchema()
}

func TestGetRequestBinder(t *testing.T) {
	tests := []struct {
		uri         string
		opcode      int
		payload     *[]byte
		appRootType reflect.Type
	}{{
		uri:         "/openconfig-acl:acl/acl-sets/",
		opcode:      1,
		payload:     nil,
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
	}}

	for _, tt := range tests {
		rb := getRequestBinder(&tt.uri, tt.payload, tt.opcode, &tt.appRootType)
		if rb == nil {
			t.Error("Error in creating the request binder object")
		} else if &tt.uri != rb.uri || tt.opcode != rb.opcode || tt.payload != rb.payload || &tt.appRootType != rb.appRootNodeType {
			t.Error("Error in creating the request binder object")
		}
	}
}

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		tid         int
		uri         string
		opcode      int
		payload     []byte
		appRootType reflect.Type
		want        string //target object name
	}{{
		tid:         1,
		uri:         "",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "Path is empty",
	}, {
		tid:         2,
		uri:         "/openconfig-acl:cpu/acl-sets/",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "no match found",
	}}

	for _, tt := range tests {
		deviceObj := ocbinds.Device{}
		deviceObj.Acl = &ocbinds.OpenconfigAcl_Acl{}
		deviceObj.Acl.AclSets = &ocbinds.OpenconfigAcl_Acl_AclSets{}
		deviceObj.Acl.AclSets.NewAclSet("SampleACL", ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4)

		binder := getRequestBinder(&tt.uri, &tt.payload, tt.opcode, &tt.appRootType)

		path, err := binder.getUriPath()
		if err != nil {
			tmpPath := gnmi.Path{}
			binder.pathTmp = &tmpPath
		} else {
			binder.pathTmp = path
		}

		err = binder.validateRequest(&deviceObj)

		if err != nil {
			// Negative test case
			if strings.Contains(err.Error(), tt.want) == false {
				t.Error("Error in validating the object: didn't get the expected error, and the error string is", err)
			}
		}
	}
}

func TestUnMarshallUri(t *testing.T) {

	tests := []struct {
		tid         int
		uri         string
		opcode      int
		payload     []byte
		appRootType reflect.Type
		want        string //target object name
	}{{
		tid:         1,
		uri:         "/openconfig-acl:acl/acl-sets/",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets",
	}, {
		tid:         2,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample][type=ACL_IPV4]/state/description",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "Description",
	}, {
		tid:         3,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "AclSet",
	}, {
		tid:         4,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample][type=ACL_IPV4]/",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets_AclSet",
	}, {
		tid:         5,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=4]",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry",
	}, {
		//Negative test case
		tid:         6,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample]/config",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "failed to create map value for insert, root map[ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_Key]*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet, keys map[name:Sample]: missing type key in map[name:Sample]",
	}, {
		//Negative test case
		tid:         7,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample][type=ACL_IPV4]/cnfig",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "rpc error: code = InvalidArgument desc = no match found in *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet",
	}, {
		//Negative test case
		tid:         8,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample][type=ACL_IPVX]/",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "no enum matching with ACL_IPVX: <nil>",
	}, {
		//Negative test case
		tid:         9,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample][type=ACL_IPV4]/state/descriptXX",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "rpc error: code = InvalidArgument desc = no match found in *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_State, for path elem:<name:\"descriptXX\"",
	}, {
		tid:         10,
		uri:         "openconfig-system:system/cpus/cpu[index=3]/",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigSystem_System{}),
		want:        "OpenconfigSystem_System_Cpus_Cpu",
	}, {
		tid:         11,
		uri:         "openconfig-system:system/cpus/cpu[index=ALL]/",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigSystem_System{}),
		want:        "OpenconfigSystem_System_Cpus_Cpu",
	}, {
		tid:         12,
		uri:         "",
		opcode:      1,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigSystem_System{}),
		want:        "Error: URI is empty",
	}, {
		//Negative test case
		tid:         13,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=Sample][type=ACL_IPV4]/",
		opcode:      3,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets_AclSet",
	}, {
		tid:         14,
		uri:         "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL1][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=8]/transport/config/tcp-flags[tcp-flags=TCP_PSH]",
		opcode:      4,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "5",
		//	}, {
		//		tid:         15,
		//		uri:         "/openconfig-bgp:bgp/neighbors/neighbor[neighbor-address=30.0.0.1]/",
		//		opcode:      2,
		//		payload:     []byte{},
		//		appRootType: reflect.TypeOf(ocbinds.OpenconfigBgp_Bgp{}),
		//		want:        "5",
	}}

	for _, tt := range tests {
		var deviceObj ocbinds.Device = ocbinds.Device{}
		workObj, err := getRequestBinder(&tt.uri, &tt.payload, tt.opcode, &tt.appRootType).unMarshallUri(&deviceObj)

		if err != nil {
			// Negative test case
			if strings.Contains(err.Error(), tt.want) == false {
				t.Error("Error in unmarshalling the URI: didn't get the expected error, and the error string is", err)
			}
			//		} else if tt.tid == 15 {
			//			if neighborEntry, ok := (*workObj).(*ocbinds.OpenconfigBgp_Bgp_Neighbors_Neighbor); ok && *neighborEntry.NeighborAddress == "30.0.0.1" {
			//				fmt.Println("PASSED testcase 15 : neighborEntry.NeighborAddress => ", *neighborEntry.NeighborAddress)
			//			} else {
			//				t.Error("Error in unmarshalling the URI: OpenconfigBgp_Bgp_Neighbors_Neighbor - object casting failed")
			//			}
		} else if tt.tid == 14 {
			leafList, ok := (*workObj).([]ocbinds.E_OpenconfigPacketMatchTypes_TCP_FLAGS)
			fmt.Println("leaf list target node value is :", leafList)
			if ok == true {
				if len(leafList) != 1 {
					t.Error("Error in unmarshalling the URI with the target node as leaf-list - type is ", reflect.ValueOf(leafList).Type())
				}
			} else {
				t.Error("Error in unmarshalling the URI with the target node as leaf-list - faile for test case id : ", tt.tid)
			}
		} else {
			_, ok := (*workObj).(ygot.GoStruct)
			if ok == false {
				//				objFieldName, err := getObjectFieldName(&tt.uri, &deviceObj, workObj)
				//				if err != nil {
				//					t.Error("Error in unmarshalling the URI: ", err)
				//				} else if objFieldName != tt.want {
				//					t.Error("Error in unmarshalling the URI: Invalid target node: ", objFieldName)
				//				}
			} else if tt.tid == 4 {
				aclSet, ok := (*workObj).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet)
				if ok == true {
					if *aclSet.Name != "Sample" && aclSet.Type != ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 {
						t.Error("Error in unmarshalling the URI: Invalid target key leaf node value: ", aclSet.Name)
					}
				} else {
					t.Error("Error in unmarshalling the URI: OpenconfigAcl_Acl_AclSets_AclSet object casting failed")
				}
			} else if tt.tid == 5 {
				aclEntry, ok := (*workObj).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry)
				if ok == true {
					if *aclEntry.SequenceId != 4 {
						t.Error("Error in unmarshalling the URI: Invalid target key leaf node value: ", *aclEntry.SequenceId)
					}
				} else {
					t.Error("Error in unmarshalling the URI: OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry object casting failed")
				}
			} else if tt.tid == 10 {
				cpuObj, ok := (*workObj).(*ocbinds.OpenconfigSystem_System_Cpus_Cpu)
				if ok == false {
					t.Error("Error in unmarshalling the URI: OpenconfigSystem_System_Cpus_Cpu failed")
				}
				val := cpuObj.Index.(*ocbinds.OpenconfigSystem_System_Cpus_Cpu_State_Index_Union_Uint32)
				cIdx, _ := cpuObj.To_OpenconfigSystem_System_Cpus_Cpu_State_Index_Union(uint32(3))
				val2 := cIdx.(*ocbinds.OpenconfigSystem_System_Cpus_Cpu_State_Index_Union_Uint32)
				if *val != *val2 {
					t.Error("Error in unmarshalling the URI: OpenconfigSystem_System_Cpus_Cpu failed")
				}
			} else if tt.tid == 11 {
				cpuObj, ok := (*workObj).(*ocbinds.OpenconfigSystem_System_Cpus_Cpu)
				if ok == false {
					t.Error("Error in unmarshalling the URI: OpenconfigSystem_System_Cpus_Cpu failed")
				}
				val := cpuObj.Index.(*ocbinds.OpenconfigSystem_System_Cpus_Cpu_State_Index_Union_E_OpenconfigSystem_System_Cpus_Cpu_State_Index)
				cIdx, err := cpuObj.To_OpenconfigSystem_System_Cpus_Cpu_State_Index_Union(ocbinds.E_OpenconfigSystem_System_Cpus_Cpu_State_Index(1))
				if err != nil {
					t.Errorf("Error in unmarshalling the URI: OpenconfigSystem_System_Cpus_Cpu failed %s", fmt.Sprint(err))
				}
				val2 := cIdx.(*ocbinds.OpenconfigSystem_System_Cpus_Cpu_State_Index_Union_E_OpenconfigSystem_System_Cpus_Cpu_State_Index)
				if *val != *val2 {
					t.Error("Error in unmarshalling the URI: OpenconfigSystem_System_Cpus_Cpu failed")
				}
			} else if reflect.TypeOf(*workObj).Elem().Name() != tt.want {
				t.Error("Error in unmarshalling the URI: Invalid target node: ", reflect.TypeOf(*workObj).Elem().Name())
			}
		}
	}
}

func TestUnMarshallPayload(t *testing.T) {
	tests := []struct {
		tid     int
		objIntf interface{}
		uri     string
		opcode  int
		payload []byte
		want    string //target object name
	}{{
		tid:     1,
		objIntf: "TestObj",
		uri:     "/openconfig-acl:acl/acl-sets/",
		opcode:  2,
		payload: []byte{},
		want:    "Error in casting the target object",
	}, {
		tid:     2,
		objIntf: ocbinds.OpenconfigAcl_Acl{},
		uri:     "/openconfig-acl:acl/acl-sets/",
		opcode:  3,
		payload: []byte{},
		want:    "Request payload is empty",
	}}

	for _, tt := range tests {
		objType := reflect.TypeOf(tt.objIntf)
		reqBinder := getRequestBinder(&tt.uri, &tt.payload, tt.opcode, &objType)
		var deviceObj ocbinds.Device = ocbinds.Device{}
		var workObj *interface{}
		var err error
		workObj, err = reqBinder.unMarshallUri(&deviceObj)
		if err != nil {
			t.Error(err)
		}

		if tt.tid == 1 {
			workObj = &tt.objIntf
		}

		err = reqBinder.unMarshallPayload(workObj)
		if err != nil {
			if strings.Contains(err.Error(), tt.want) == false {
				t.Error("Negative test case failed: ", err)
			}
		}
	}
}

func TestGetUriPath(t *testing.T) {

	tests := []struct {
		tid         int
		uri         string
		opcode      int
		payload     []byte
		appRootType reflect.Type
		want        string //target object name
	}{{
		tid:         1,
		uri:         "////",
		opcode:      2,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "error formatting path",
	}}

	for _, tt := range tests {
		reqBinder := getRequestBinder(&tt.uri, &tt.payload, tt.opcode, &tt.appRootType)
		_, err := reqBinder.getUriPath()
		if err != nil {
			if strings.Contains(err.Error(), tt.want) == false {
				t.Error("Negative test case failed: ", err)
			}
		}
	}
}

func TestUnMarshall(t *testing.T) {

	tests := []struct {
		tid         int
		uri         string
		opcode      int
		payload     []byte
		appRootType reflect.Type
		want        string //target object name
	}{{
		tid:         1,
		uri:         "/openconfig-acl:acl/acl-sets/",
		opcode:      2,
		payload:     []byte("{    \"acl-set\": [    {      \"name\": \"MyACL3\",    \"type\": \"ACL_IPV4\",   \"config\": {   \"name\": \"MyACL3\",                 \"type\": \"ACL_IPV4\",                 \"description\": \"Description for MyACL3\" }    }   ] } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets",
	}, {
		//Negative test case
		tid:         2,
		uri:         "/openconfig-acl:acl/acl-sets/",
		opcode:      2,
		payload:     []byte("{    \"acl-set\": [    {      \"nname\": \"MyACL3\",    \"type\": \"ACL_IPV4\",   \"config\": {   \"name\": \"MyACL3\",                 \"type\": \"ACL_IPV4\",                 \"description\": \"Description for MyACL3\" }    }   ] } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "parent container acl-set (type *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet): JSON contains unexpected field nname",
	}, {
		//Negative test case
		tid:         3,
		uri:         "/openconfig-acl:acl/acl-sets/",
		opcode:      2,
		payload:     []byte("{    \"acl-set\": [    {      \"name\": \"MyACL3\",    \"type\": \"ACL_IPV7\",   \"config\": {   \"name\": \"MyACL3\",                 \"type\": \"ACL_IPV4\",                 \"description\": \"Description for MyACL3\" }    }   ] } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "ACL_IPV7 is not a valid value for enum field Type, type ocbinds.E_OpenconfigAcl_ACL_TYPE",
	}, {
		tid:         4,
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=MyACL1][type=ACL_IPV4]/",
		opcode:      2,
		payload:     []byte("{ \"acl-entries\": { \"acl-entry\": [  {  \"sequence-id\": 1, \"config\": { \"sequence-id\": 1, \"description\": \"Description for MyACL1 Rule Seq 1\"  },   \"ipv4\": {  \"config\": {  \"source-address\": \"11.1.1.1/32\",  \"destination-address\": \"21.1.1.1/32\", \"dscp\": 1, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 101, \"destination-port\": 201 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } } ] } } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets_AclSet",
	}, {
		tid:         5, //PUT
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=MyACL1][type=ACL_IPV4]/acl-entries",
		opcode:      3,
		payload:     []byte("{ \"acl-entries\": { \"acl-entry\": [  {  \"sequence-id\": 1, \"config\": { \"sequence-id\": 1, \"description\": \"Description for MyACL1 Rule Seq 1\"  },   \"ipv4\": {  \"config\": {  \"source-address\": \"11.1.1.1/32\",  \"destination-address\": \"21.1.1.1/32\", \"dscp\": 1, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 101, \"destination-port\": 201 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } } ] } } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets_AclSet",
	}, {
		tid:         6, //PATCH
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=MyACL1][type=ACL_IPV4]/acl-entries",
		opcode:      4,
		payload:     []byte("{ \"acl-entries\": { \"acl-entry\": [  {  \"sequence-id\": 1, \"config\": { \"sequence-id\": 1, \"description\": \"Description for MyACL1 Rule Seq 1\"  },   \"ipv4\": {  \"config\": {  \"source-address\": \"11.1.1.1/32\",  \"destination-address\": \"21.1.1.1/32\", \"dscp\": 1, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 101, \"destination-port\": 201 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } } ] } } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets_AclSet",
	}, {
		tid:         7, //DELETE
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=MyACL1][type=ACL_IPV4]/",
		opcode:      5,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "OpenconfigAcl_Acl_AclSets_AclSet",
	}, {
		tid:         8, //GET on leaf node
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=MyACL1][type=ACL_IPV4]/acl-entries/acl-entry[sequence-id=4]/ipv4/config/source-address",
		opcode:      5,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "SourceAddress",
	}, {
		tid:         9, //negative test case
		uri:         "/openconfig-acl:acl/",
		opcode:      3,
		payload:     []byte{},
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "Request payload is empty",
	}, {
		tid:         10, //GET  - negative test case - incorrect opcode
		uri:         "/openconfig-acl:acl/acl-sets/",
		opcode:      7,
		payload:     []byte("{    \"acl-set\": [    {      \"name\": \"MyACL3\",    \"type\": \"ACL_IPV4\",   \"config\": {   \"name\": \"MyACL3\",                 \"type\": \"ACL_IPV4\",                 \"description\": \"Description for MyACL3\" }    }   ] } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "Unknown HTTP METHOD in the request",
	}, {
		tid:         11, // negative
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set/",
		opcode:      2,
		payload:     []byte("{    \"acl-set\": [    {      \"name\": \"MyACL3\",    \"type\": \"ACL_IPV4\",   \"config\": {   \"name\": \"MyACL3\",                 \"type\": \"ACL_IPV4\",                 \"description\": \"Description for MyACL3\" }    }   ] } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "URI doesn't have keys in the request",
	}, {
		tid:         12, // negative
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=MyACL1][type=ACL_IPV4]/",
		opcode:      2,
		payload:     []byte("{ \"acl-entries\": { \"acl-entry\": [  {  \"sequence-id\": abc, \"config\": { \"sequence-id\": 1, \"description\": \"Description for MyACL1 Rule Seq 1\"  },   \"ipv4\": {  \"config\": {  \"source-address\": \"11.1.1.1/32\",  \"destination-address\": \"21.1.1.1/32\", \"dscp\": 1, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 101, \"destination-port\": 201 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } } ] } } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "invalid character",
	}, {
		tid:         13, // negative
		uri:         "/openconfig-acl:acl/acl-sets/acl-set[name=MyACL1]",
		opcode:      1,
		payload:     []byte("{ \"acl-entries\": { \"acl-entry\": [  {  \"sequence-id\": abc, \"config\": { \"sequence-id\": 1, \"description\": \"Description for MyACL1 Rule Seq 1\"  },   \"ipv4\": {  \"config\": {  \"source-address\": \"11.1.1.1/32\",  \"destination-address\": \"21.1.1.1/32\", \"dscp\": 1, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 101, \"destination-port\": 201 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } } ] } } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "failed to create map value for insert",
		//	}, {
		//		tid:         14, // PATCH - test bgp neighbor bind with update
		//		uri:         "/openconfig-bgp:bgp/neighbors/neighbor[neighbor-address=30.0.0.1]/ebgp-multihop/config/multihop-ttl",
		//		opcode:      4,
		//		payload:     []byte("{\"openconfig-bgp:multihop-ttl\": 3}"),
		//		appRootType: reflect.TypeOf(ocbinds.OpenconfigBgp_Bgp{}),
		//		want:        "15",
	}, {
		tid:         15, // PATCH - from the base node acl
		uri:         "/openconfig-acl:acl/",
		opcode:      4,
		payload:     []byte("{  \"openconfig-acl:acl\": {  \"acl-sets\": {  \"acl-set\": [    {      \"name\": \"MyACL3\",    \"type\": \"ACL_IPV4\",   \"config\": {   \"name\": \"MyACL3\",                 \"type\": \"ACL_IPV4\",                 \"description\": \"Description for MyACL3\" }    }   ] } } }"),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "XXX",
	}, {
		tid:         16, // PATCH - from the base node acl
		uri:         "/openconfig-acl:acl/acl-sets/acl-set",
		opcode:      4,
		payload:     []byte("{    \"acl-set\": [    {      \"name\": \"MyACL3\",    \"type\": \"ACL_IPV4\",   \"config\": {   \"name\": \"MyACL3\",                 \"type\": \"ACL_IPV4\",                 \"description\": \"Description for MyACL3\" }    }   ] } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "XXX",
	}, {
		tid:         17, // PATCH - from the base node acl
		uri:         "/openconfig-acl:acl/acl-sets/openconfig-acl:acl-set[name=MyACL333][type=ACL_IPV4]/",
		opcode:      4,
		payload:     []byte("{    \"acl-set\": [    {      \"name\": \"MyACL333\",    \"type\": \"ACL_IPV4\",   \"config\": {   \"name\": \"MyACL333\",                 \"type\": \"ACL_IPV4\",                 \"description\": \"Description for MyACL333\" }    }   ] } "),
		appRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
		want:        "XXX",
	}}

	for _, tt := range tests {
		_, workObj, err := getRequestBinder(&tt.uri, &tt.payload, tt.opcode, &tt.appRootType).unMarshall()

		if err != nil {
			if strings.Contains(err.Error(), tt.want) == false {
				t.Errorf("TestUnMarshall: Testcase id: %d Error in unmarshalling the payload: didn't get the expected error, and the error string is %s ", tt.tid, err)
			}
		} else {
			if tt.tid == 17 {
				fmt.Println("TestUnMarshall: Testcase 17 passed - workObj: ", *workObj)
				if aclSet, ok := (*workObj).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet); ok {
					fmt.Printf("TestUnMarshall: Testcase 17 passed - Acl Name: %s, and Acl Type: %d\n", *aclSet.Name, aclSet.Type)
				} else {
					t.Error("TestUnMarshall: Testcase 17 - Error in unmarshalling the payload: OpenconfigAcl_Acl_AclSets")
				}
			} else if tt.tid == 16 {
				if reflect.ValueOf(*workObj).Kind().String() == "map" {
					fmt.Println("TestUnMarshall: Testcase 16 passed - workobj: ", *workObj)
				} else {
					t.Error("TestUnMarshall: Testcase 16 - Error in unmarshalling the payload: OpenconfigAcl_AclSet")
				}
			} else if tt.tid == 15 {
				aclObj, ok := (*workObj).(*ocbinds.OpenconfigAcl_Acl)
				if ok == true {
					fmt.Println("TestUnMarshall: Testcase 15 passed - ", aclObj)
				} else {
					t.Error("TestUnMarshall: Testcase 15 - Error in unmarshalling the payload: OpenconfigAcl_Acl")
				}
			} else if tt.tid == 14 {
				ttlVal, ok := (*workObj).(*uint8)
				if ok == false || *ttlVal != 3 {
					t.Error("TestUnMarshall: Testcase 14 - Error in unmarshalling the payload: OpenconfigBgp_Bgp_Neighbors_Neighbor_EbgpMultihop_Config_multihop-ttl failed")
				} else {
					fmt.Println("TestUnMarshall: Testcase 14 passed - ptach method: => multihop-ttl value ==> ", *ttlVal)
				}
			} else if tt.tid == 4 {
				aclSet, ok := (*workObj).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet)
				if ok == true {
					if aclSet.AclEntries.AclEntry[1] != nil && *aclSet.AclEntries.AclEntry[1].SequenceId == 1 {
						_, err = aclSet.AclEntries.AclEntry[1].Transport.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union(uint16(201))
						if err != nil {
							t.Error("Error in unmarshalling the payload: ", err)
						}
					} else {
						t.Error("Error in unmarshalling the payload")
					}
				} else {
					t.Error("Error in unmarshalling the payload: OpenconfigAcl_Acl_AclSets_AclSet object casting failed")
				}
			} else if tt.tid == 5 || tt.tid == 6 {
				aclEntries, ok := (*workObj).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries)
				if ok == true {
					if aclEntries.AclEntry[1] != nil && *aclEntries.AclEntry[1].SequenceId == 1 {
						_, err = aclEntries.AclEntry[1].Ipv4.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union(ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_TCP)
						if err != nil {
							t.Error("Error in unmarshalling the payload: ", err)
						}
					} else {
						t.Error("Error in unmarshalling the payload")
					}
				} else {
					t.Error("Error in unmarshalling the payload: OpenconfigAcl_Acl_AclSets_AclSet object casting failed")
				}
			} else {
				_, ok := (*workObj).(ygot.GoStruct)
				if ok == false {
					//					objFieldName, err := getObjectFieldName(&tt.uri, (*rootObj).(*ocbinds.Device), workObj)
					//					if err != nil {
					//						t.Error("Error in unmarshalling the URI: ", err)
					//					} else if objFieldName != tt.want {
					//						t.Error("Error in unmarshalling the payload: Invalid target node: ", objFieldName)
					//					}
				} else if reflect.TypeOf(*workObj).Elem().Name() != tt.want {
					t.Error("Error in unmarshalling the payload: Invalid target node: ", reflect.TypeOf(*workObj).Elem().Name())
				}
			}
		}
	}
}
