///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package transformer

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
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
			fmt.Fprintf(&template, "{%s}", name)
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

func RemoveXPATHPredicates(s string) (string, error) {
	var b bytes.Buffer
	for i := 0; i < len(s); {
		ss := s[i:]
		si, ei := strings.Index(ss, "["), strings.Index(ss, "]")
		switch {
		case si == -1 && ei == -1:
			// This substring didn't contain a [] pair, therefore write it
			// to the buffer.
			b.WriteString(ss)
			// Move to the last character of the substring.
			i += len(ss)
		case si == -1 || ei == -1:
			// This substring contained a mismatched pair of []s.
			return "", fmt.Errorf("Mismatched brackets within substring %s of %s, [ pos: %d, ] pos: %d", ss, s, si, ei)
		case si > ei:
			// This substring contained a ] before a [.
			return "", fmt.Errorf("Incorrect ordering of [] within substring %s of %s, [ pos: %d, ] pos: %d", ss, s, si, ei)
		default:
			// This substring contained a matched set of []s.
			b.WriteString(ss[0:si])
			i += ei + 1
		}
	}

	return b.String(), nil
}

// stripPrefix removes the prefix from a YANG path element. For example, removing
// foo from "foo:bar". Such qualified paths are used in YANG modules where remote
// paths are referenced.
func stripPrefix(name string) (string, error) {
        ps := strings.Split(name, ":")
        switch len(ps) {
        case 1:
                return name, nil
        case 2:
                return ps[1], nil
        }
        return "", fmt.Errorf("path element did not form a valid name (name, prefix:name): %v", name)
}
/*
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
*/
