///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package translib

import (
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
