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
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// parseQueryParams parses the http request's query parameters
// into a translibArgs args.
func (args *translibArgs) parseQueryParams(r *http.Request) error {
	if r.URL.RawQuery == "" {
		return nil
	}
	if strings.Contains(r.URL.Path, restconfDataPathPrefix) {
		return args.parseRestconfQueryParams(r)
	}

	return nil
}

// parseRestconfQueryParams parses query parameters of a request 'r' to
// fill translibArgs object 'args'. Returns httpError with status 400
// if any parameter is unsupported or has invalid value.
func (args *translibArgs) parseRestconfQueryParams(r *http.Request) error {
	var err error
	qParams := extractQuery(r.URL.RawQuery)

	for name, vals := range qParams {
		switch name {
		case "depth":
			args.depth, err = parseDepthParam(vals, r)
		case "content":
			args.content, err = parseContentParam(vals, r)
		case "fields":
			args.fields, err = parseFieldsParam(vals, r)
		case "deleteEmptyEntry":
			args.deleteEmpty, err = parseDeleteEmptyEntryParam(vals, r)
		default:
			err = newUnsupportedParamError(name, r)
		}
		if err != nil {
			return err
		}
	}
	if len(args.fields) > 0 {
		if len(args.content) > 0 || args.depth > 0 {
			return httpError(http.StatusBadRequest, "Fields query parameter is not supported along with other query parameters")
		}
	}

	return nil
}

func extractQuery(rawQuery string) map[string][]string {
	queryParamsMap := make(map[string][]string)
	if len(rawQuery) == 0 {
		return queryParamsMap
	}
	// The query parameters are seperated by &
	qpList := strings.Split(rawQuery, "&")
	for _, each := range qpList {
		var valList []string
		if strings.Contains(each, "=") {
			eqIdx := strings.Index(each, "=")
			key := each[:eqIdx]
			val := each[eqIdx+1:]
			if _, ok := queryParamsMap[key]; ok {
				queryParamsMap[key] = append(queryParamsMap[key], val)
			} else {
				valList = append(valList, val)
				queryParamsMap[key] = valList
			}
		} else {
			queryParamsMap[each] = valList
		}
	}
	return queryParamsMap
}

func newUnsupportedParamError(name string, r *http.Request) error {
	return httpError(http.StatusBadRequest, "query parameter '%s' not supported", name)
}

func newInvalidParamError(name string, r *http.Request) error {
	return httpError(http.StatusBadRequest, "invalid '%s' query parameter", name)
}

// parseDepthParam parses query parameter value for "depth" parameter.
// See https://tools.ietf.org/html/rfc8040#section-4.8.2
func parseDepthParam(v []string, r *http.Request) (uint, error) {
	if !restconfCapabilities.depth {
		glog.V(1).Infof("[%s] 'depth' support disabled", getRequestID(r))
		return 0, newUnsupportedParamError("depth", r)
	}

	if r.Method != "GET" && r.Method != "HEAD" {
		glog.V(1).Infof("[%s] 'depth' not supported for %s", getRequestID(r), r.Method)
		return 0, newUnsupportedParamError("depth supported only for GET/HEAD requests", r)
	}

	if len(v) != 1 {
		glog.V(1).Infof("[%s] Expecting only 1 depth param; found %d", getRequestID(r), len(v))
		return 0, newInvalidParamError("depth", r)
	}

	if v[0] == "unbounded" {
		return 0, nil
	}

	d, err := strconv.ParseUint(v[0], 10, 16)
	if err != nil || d == 0 {
		glog.V(1).Infof("[%s] Bad depth value '%s', err=%v", getRequestID(r), v[0], err)
		return 0, newInvalidParamError("depth", r)
	}

	return uint(d), nil
}

// parseContentParam parses query parameter value for "content" parameter.
// See https://tools.ietf.org/html/rfc8040#section-4.8.1
func parseContentParam(v []string, r *http.Request) (string, error) {
	if !restconfCapabilities.content {
		glog.V(1).Infof("'content' support disabled")
		return "", newUnsupportedParamError("content", r)
	}

	if r.Method != "GET" && r.Method != "HEAD" {
		glog.V(1).Infof("'content' not supported for %s", r.Method)
		return "", newUnsupportedParamError("content", r)
	}

	if len(v) != 1 {
		glog.V(1).Infof("Expecting only 1 content param; found %d", len(v))
		return "", newInvalidParamError("content", r)
	}

	if v[0] == "all" || v[0] == "config" || v[0] == "nonconfig" {
		return v[0], nil
	} else {
		glog.V(1).Infof("Bad content value '%s'", v[0])
		return "", newInvalidParamError("content", r)
	}

	return v[0], nil
}

func extractFields(s string) []string {
	prefix := ""
	cur := ""
	res := make([]string, 0)
	for i, c := range s {
		if c == '(' {
			prefix = cur
			cur = ""
		} else if c == ')' {
			res = append(res, prefix+"/"+cur)
			prefix = ""
			cur = ""
		} else if c == ';' {
			fullpath := prefix
			if len(prefix) > 0 {
				fullpath += "/"
			}
			if len(fullpath+cur) > 0 {
				res = append(res, fullpath+cur)
			}
			cur = ""
		} else if c == ' ' {
			continue
		} else {
			cur += string(c)
		}
		if i == (len(s) - 1) {
			fullpath := prefix
			if len(prefix) > 0 {
				fullpath += "/"
			}
			if len(fullpath+cur) > 0 {
				res = append(res, fullpath+cur)
			}
		}
	}
	return res
}

// parseFieldsParam parses query parameter value for "fields" parameter.
// See https://tools.ietf.org/html/rfc8040#section-4.8.3
func parseFieldsParam(v []string, r *http.Request) ([]string, error) {
	if !restconfCapabilities.fields {
		glog.V(1).Infof("'fields' support disabled")
		return v, newUnsupportedParamError("fields", r)
	}

	if r.Method != "GET" && r.Method != "HEAD" {
		glog.V(1).Infof("'fields' not supported for %s", r.Method)
		return v, newUnsupportedParamError("fields supported only for GET/HEAD query", r)
	}

	if len(v) != 1 {
		glog.V(1).Infof("Expecting atleast 1 fields param; found %d", len(v))
		return v, newInvalidParamError("fields", r)
	}

	res := extractFields(v[0])
	return res, nil
}

// parseDeleteEmptyEntryParam parses the custom "deleteEmptyEntry" query parameter.
func parseDeleteEmptyEntryParam(v []string, r *http.Request) (bool, error) {
	if r.Method != "DELETE" {
		glog.V(1).Infof("[%s] deleteEmptyEntry not supported for %s", getRequestID(r), r.Method)
		return false, newUnsupportedParamError("deleteEmptyEntry", r)
	}

	if len(v) != 1 {
		glog.V(1).Infof("[%s] expecting only 1 deleteEmptyEntry; found %d", getRequestID(r), len(v))
		return false, newInvalidParamError("deleteEmptyEntry", r)
	}

	return strings.EqualFold(v[0], "true"), nil
}
