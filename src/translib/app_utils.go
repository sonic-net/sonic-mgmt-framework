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
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"

	log "github.com/golang/glog"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
)

func getYangPathFromUri(uri string) (string, error) {
	var path *gnmi.Path
	var err error

	path, err = ygot.StringToPath(uri, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		log.Errorf("Error in uri to path conversion: %v", err)
		return "", err
	}

	yangPath, yperr := ygot.PathToSchemaPath(path)
	if yperr != nil {
		log.Errorf("Error in Gnmi path to Yang path conversion: %v", yperr)
		return "", yperr
	}

	return yangPath, err
}

func getYangPathFromYgotStruct(s ygot.GoStruct, yangPathPrefix string, appModuleName string) string {
	tn := reflect.TypeOf(s).Elem().Name()
	schema, ok := ocbinds.SchemaTree[tn]
	if !ok {
		log.Errorf("could not find schema for type %s", tn)
		return ""
	} else if schema != nil {
		yPath := schema.Path()
		//yPath = strings.Replace(yPath, "/device/acl", "/openconfig-acl:acl", 1)
		yPath = strings.Replace(yPath, yangPathPrefix, appModuleName, 1)
		return yPath
	}
	return ""
}

func generateGetResponsePayload(targetUri string, deviceObj *ocbinds.Device, ygotTarget *interface{}) ([]byte, error) {
	var err error
	var payload []byte

	if len(targetUri) == 0 {
		return payload, tlerr.InvalidArgs("GetResponse failed as target Uri is not valid")
	}
	path, err := ygot.StringToPath(targetUri, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		return payload, tlerr.InvalidArgs("URI to path conversion failed: %v", err)
	}

	// Get current node (corresponds to ygotTarget) and its parent node
	var pathList []*gnmi.PathElem = path.Elem
	parentPath := &gnmi.Path{}
	for i := 0; i < len(pathList); i++ {
		if log.V(3) {
			log.Infof("pathList[%d]: %s\n", i, pathList[i])
		}
		pathSlice := strings.Split(pathList[i].Name, ":")
		pathList[i].Name = pathSlice[len(pathSlice)-1]
		if i < (len(pathList) - 1) {
			parentPath.Elem = append(parentPath.Elem, pathList[i])
		}
	}
	parentNodeList, err := ytypes.GetNode(ygSchema.RootSchema(), deviceObj, parentPath)
	if err != nil {
		return payload, err
	}
	if len(parentNodeList) == 0 {
		return payload, tlerr.InvalidArgs("Invalid URI: %s", targetUri)
	}
	parentNode := parentNodeList[0].Data

	currentNodeList, err := ytypes.GetNode(ygSchema.RootSchema(), deviceObj, path, &(ytypes.GetPartialKeyMatch{}))
	if err != nil {
		return payload, err
	}
	if len(currentNodeList) == 0 {
		return payload, tlerr.NotFound("Resource not found")
	}
	//currentNode := currentNodeList[0].Data
	currentNodeYangName := currentNodeList[0].Schema.Name

	// Create empty clone of parent node
	parentNodeClone := reflect.New(reflect.TypeOf(parentNode).Elem())
	var parentCloneObj ygot.ValidatedGoStruct
	var ok bool
	if parentCloneObj, ok = (parentNodeClone.Interface()).(ygot.ValidatedGoStruct); ok {
		ygot.BuildEmptyTree(parentCloneObj)
		pcType := reflect.TypeOf(parentCloneObj).Elem()
		pcValue := reflect.ValueOf(parentCloneObj).Elem()

		var currentNodeOCFieldName string
		for i := 0; i < pcValue.NumField(); i++ {
			fld := pcValue.Field(i)
			fldType := pcType.Field(i)
			if fldType.Tag.Get("path") == currentNodeYangName {
				currentNodeOCFieldName = fldType.Name
				// Take value from original parent and set in parent clone
				valueFromParent := reflect.ValueOf(parentNode).Elem().FieldByName(currentNodeOCFieldName)
				fld.Set(valueFromParent)
				break
			}
		}
		if log.V(3) {
			log.Infof("Target yang name: %s  OC Field name: %s\n", currentNodeYangName, currentNodeOCFieldName)
		}
	}

	payload, err = dumpIetfJson(parentCloneObj, true)

	return payload, err
}

func getTargetNodeYangSchema(targetUri string, deviceObj *ocbinds.Device) (*yang.Entry, error) {
	if len(targetUri) == 0 {
		return nil, tlerr.InvalidArgs("GetResponse failed as target Uri is not valid")
	}
	path, err := ygot.StringToPath(targetUri, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		return nil, tlerr.InvalidArgs("URI to path conversion failed: %v", err)
	}
	// Get current node (corresponds to ygotTarget)
	var pathList []*gnmi.PathElem = path.Elem
	for i := 0; i < len(pathList); i++ {
		if log.V(3) {
			log.Infof("pathList[%d]: %s\n", i, pathList[i])
		}
		pathSlice := strings.Split(pathList[i].Name, ":")
		pathList[i].Name = pathSlice[len(pathSlice)-1]
	}
	targetNodeList, err := ytypes.GetNode(ygSchema.RootSchema(), deviceObj, path, &(ytypes.GetPartialKeyMatch{}))
	if err != nil {
		return nil, tlerr.InvalidArgs("Getting node information failed: %v", err)
	}
	if len(targetNodeList) == 0 {
		return nil, tlerr.NotFound("Resource not found")
	}
	targetNodeSchema := targetNodeList[0].Schema
	//targetNode := targetNodeList[0].Data
	if log.V(3) {
		log.Infof("Target node yang name: %s\n", targetNodeSchema.Name)
	}
	return targetNodeSchema, nil
}

func dumpIetfJson(s ygot.ValidatedGoStruct, skipValidation bool) ([]byte, error) {
	jsonStr, err := ygot.EmitJSON(s, &ygot.EmitJSONConfig{
		Format:         ygot.RFC7951,
		Indent:         "  ",
		SkipValidation: skipValidation,
		RFC7951Config: &ygot.RFC7951JSONConfig{
			AppendModuleName: true,
		},
	})
	var buf bytes.Buffer
	json.Compact(&buf, []byte(jsonStr))
	return []byte(buf.String()), err
}

func contains(sl []string, str string) bool {
	for _, v := range sl {
		if v == str {
			return true
		}
	}
	return false
}

func removeElement(sl []string, str string) []string {
	for i := 0; i < len(sl); i++ {
		if sl[i] == str {
			sl = append(sl[:i], sl[i+1:]...)
			i--
			sl = sl[:len(sl)]
			break
		}
	}
	return sl
}

// isNotFoundError return true if the error is a 'not found' error
func isNotFoundError(err error) bool {
	switch err.(type) {
	case tlerr.TranslibRedisClientEntryNotExist, tlerr.NotFoundError:
		return true
	default:
		return false
	}
}

// asKey cretaes a db.Key from given key components
func asKey(parts ...string) db.Key {
	return db.Key{Comp: parts}
}

func createEmptyDbValue(fieldName string) db.Value {
	return db.Value{Field: map[string]string{fieldName: ""}}
}

/* Check if targetUriPath is child (subtree) of nodePath
The return value can be used to decide if subtrees needs
to visited to fill the data or not.
*/
func isSubtreeRequest(targetUriPath string, nodePath string) bool {
	return strings.HasPrefix(targetUriPath, nodePath)
}
