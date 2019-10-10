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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"translib/ocbinds"

	log "github.com/golang/glog"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
)

// PathInfo structure contains parsed path information.
type PathInfo struct {
	Path     string
	Template string
	Vars     map[string]string
}

// Var returns the string value for a path variable. Returns
// empty string if no such variable exists.
func (p *PathInfo) Var(name string) string {
	return p.Vars[name]
}

// IntVar returns the value for a path variable as an int.
// Returns 0 if no such variable exists. Returns an error
// if the value is not an integer.
func (p *PathInfo) IntVar(name string) (int, error) {
	val := p.Vars[name]
	if len(val) == 0 {
		return 0, nil
	}

	return strconv.Atoi(val)
}

// HasPrefix checks if this path template starts with given
// prefix.. Shorthand for strings.HasPrefix(p.Template, s)
func (p *PathInfo) HasPrefix(s string) bool {
	return strings.HasPrefix(p.Template, s)
}

// HasSuffix checks if this path template ends with given
// suffix.. Shorthand for strings.HasSuffix(p.Template, s)
func (p *PathInfo) HasSuffix(s string) bool {
	return strings.HasSuffix(p.Template, s)
}

// NewPathInfo parses given path string into a PathInfo structure.
func NewPathInfo(path string) *PathInfo {
	var info PathInfo
	info.Path = path
	info.Vars = make(map[string]string)

	//TODO optimize using regexp
	var template strings.Builder
	r := strings.NewReader(path)

	for r.Len() > 0 {
		c, _ := r.ReadByte()
		if c != '[' {
			template.WriteByte(c)
			continue
		}

		name := readUntil(r, '=')
		value := readUntil(r, ']')
		if len(name) != 0 {
			fmt.Fprintf(&template, "{}")
			info.Vars[name] = value
		}
	}

	info.Template = template.String()

	return &info
}

func readUntil(r *strings.Reader, delim byte) string {
	var buff strings.Builder
	for {
		c, err := r.ReadByte()
		if err == nil && c != delim {
			buff.WriteByte(c)
		} else {
			break
		}
	}

	return buff.String()
}

func getParentNode(targetUri *string, deviceObj *ocbinds.Device) (*interface{}, *yang.Entry, error) {
	path, err := ygot.StringToPath(*targetUri, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		return nil, nil, err
	}

	var pathList []*gnmi.PathElem = path.Elem

	parentPath := &gnmi.Path{}

	for i := 0; i < (len(pathList) - 1); i++ {
		pathSlice := strings.Split(pathList[i].Name, ":")
		pathList[i].Name = pathSlice[len(pathSlice)-1]
		parentPath.Elem = append(parentPath.Elem, pathList[i])
	}

	treeNodeList, err2 := ytypes.GetNode(ygSchema.RootSchema(), deviceObj, parentPath)
	if err2 != nil {
		return nil, nil, err2
	}

	if len(treeNodeList) == 0 {
		return nil, nil, errors.New("Invalid URI")
	}

	return &(treeNodeList[0].Data), treeNodeList[0].Schema, nil
}

func getNodeName(targetUri *string, deviceObj *ocbinds.Device) (string, error) {
	path, err := ygot.StringToPath(*targetUri, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		log.Error("Error in uri to path conversion: ", err)
		return "", err
	}

	pathList := path.Elem
	for i := 0; i < len(pathList); i++ {
		pathSlice := strings.Split(pathList[i].Name, ":")
		pathList[i].Name = pathSlice[len(pathSlice)-1]
	}

	treeNodeList, err := ytypes.GetNode(ygSchema.RootSchema(), deviceObj, path)
	if err != nil {
		log.Error("Error in uri to path conversion: ", err)
		return "", err
	}

	if len(treeNodeList) == 0 {
		return "", errors.New("Invalid URI")
	}

	return treeNodeList[0].Schema.Name, nil
}

func getObjectFieldName(targetUri *string, deviceObj *ocbinds.Device, ygotTarget *interface{}) (string, error) {
	parentObjIntf, _, err := getParentNode(targetUri, deviceObj)
	if err != nil {
		return "", err
	}
	valObj := reflect.ValueOf(*parentObjIntf).Elem()
	parentObjType := reflect.TypeOf(*parentObjIntf).Elem()

	for i := 0; i < parentObjType.NumField(); i++ {
		if reflect.ValueOf(*ygotTarget).Kind() == reflect.Ptr && valObj.Field(i).Kind() == reflect.Ptr {
			if valObj.Field(i).Pointer() == reflect.ValueOf(*ygotTarget).Pointer() {
				return parentObjType.Field(i).Name, nil
			}
		} else if valObj.Field(i).String() == reflect.ValueOf(*ygotTarget).String() {
			return parentObjType.Field(i).Name, nil
		}
	}
	return "", errors.New("Target object not found")
}


