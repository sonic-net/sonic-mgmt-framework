///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package translib

import (
	"fmt"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
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

func dumpIetfJson(s ygot.ValidatedGoStruct, skipValidation bool) ([]byte, error) {
	jsonStr, err := ygot.EmitJSON(s, &ygot.EmitJSONConfig{
		Format:         ygot.RFC7951,
		Indent:         "  ",
		SkipValidation: skipValidation,
		RFC7951Config: &ygot.RFC7951JSONConfig{
			AppendModuleName: true,
		},
	})
	return []byte(jsonStr), err
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
