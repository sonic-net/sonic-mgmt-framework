///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// errorResponse defines the RESTCONF compliant error response
// payload. It includes a list of errorInfo object.
type errorResponse struct {
	Err struct {
		Arr []errorInfo `json:"error"`
	} `json:"ietf-restconf:errors"`
}

// errorInfo defines the RESTCONF compilant error information
// payload.
type errorInfo struct {
	Type    string `json:"error-type"`
	Tag     string `json:"error-tag"`
	AppTag  string `json:"error-app-tag,omitempty"`
	Path    string `json:"error-path,omitempty"`
	Message string `json:"error-message,omitempty"`
}

// httpErrorType is an error structure for indicating HTTP protocol
// errors. Includes HTTP status code and user displayable message.
type httpErrorType struct {
	status  int
	message string
}

func (e httpErrorType) Error() string {
	return e.message
}

func httpError(status int, msg string, args ...interface{}) error {
	return httpErrorType{
		status:  status,
		message: fmt.Sprintf(msg, args...)}
}

func httpBadRequest(msg string, args ...interface{}) error {
	return httpError(http.StatusBadRequest, msg, args...)
}

func httpServerError(msg string, args ...interface{}) error {
	return httpError(http.StatusInternalServerError, msg, args...)
}

// prepareErrorResponse returns HTTP status code and response payload
// for an error object. Response payalod is formatted as per RESTCONF
// specification (RFC8040, section 7.1). Uses json encoding.
func prepareErrorResponse(err error, r *http.Request) (status int, data []byte, mimeType string) {
	var errInfo errorInfo

	switch e := err.(type) {
	case httpErrorType:
		status = e.status
		errInfo.Type = "protocol"
		errInfo.Message = e.message

	default:
		//FIXME
		status = http.StatusBadRequest
		errInfo.Type = "application"
		errInfo.Message = err.Error()
	}

	var resp errorResponse
	resp.Err.Arr = append(resp.Err.Arr, errInfo)
	data, _ = json.Marshal(&resp)
	mimeType = "application/yang-data+json"

	return
}
