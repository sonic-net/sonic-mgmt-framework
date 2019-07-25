///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package translib

import (
    "bytes"
    "encoding/json"
    "errors"
	"fmt"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
	"reflect"
	"strings"
	"translib/ocbinds"
)

func getYangPathFromUri(uri string) (string, error) {
	var path *gnmi.Path
	var err error

	path, err = ygot.StringToPath(uri, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		fmt.Println("Error in uri to path conversion: ", err)
		return "", err
	}

	yangPath, yperr := ygot.PathToSchemaPath(path)
	if yperr != nil {
		fmt.Println("Error in Gnmi path to Yang path conversion: ", yperr)
		return "", yperr
	}

	return yangPath, err
}

func getYangPathFromYgotStruct(s ygot.GoStruct, yangPathPrefix string, appModuleName string) string {
	tn := reflect.TypeOf(s).Elem().Name()
	schema, ok := ocbinds.SchemaTree[tn]
	if !ok {
		fmt.Errorf("could not find schema for type %s", tn)
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
		return payload, errors.New("GetResponse failed as target Uri is not valid")
	}
	path, err := ygot.StringToPath(targetUri, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		fmt.Println("Error in uri to path conversion: ", err)
		return payload, err
	}

	// Get current node (corresponds to ygotTarget) and its parent node
	var pathList []*gnmi.PathElem = path.Elem
	parentPath := &gnmi.Path{}
	for i := 0; i < len(pathList); i++ {
		fmt.Printf("pathList[%d]: %s\n", i, pathList[i])
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
		return payload, errors.New("Invalid URI")
	}
	parentNode := parentNodeList[0].Data

	currentNodeList, err := ytypes.GetNode(ygSchema.RootSchema(), deviceObj, path, &(ytypes.GetPartialKeyMatch{}))
	if err != nil {
		return payload, err
	}
	if len(currentNodeList) == 0 {
		return payload, errors.New("Invalid URI")
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
		fmt.Printf("Target yang name: %s  OC Field name: %s\n", currentNodeYangName, currentNodeOCFieldName)
	}

	payload, err = dumpIetfJson(parentCloneObj, true)

	return payload, err
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
