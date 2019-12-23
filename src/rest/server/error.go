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

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cvl"
	"translib/tlerr"
)

// errorResponse defines the RESTCONF compliant error response
// payload. It includes a list of errorEntry object.
type errorResponse struct {
	Err struct {
		Arr []errorEntry `json:"error"`
	} `json:"ietf-restconf:errors"`
}

// errorEntry defines the RESTCONF compilant error information
// payload.
type errorEntry struct {
	Type    errtype `json:"error-type"`
	Tag     errtag  `json:"error-tag"`
	AppTag  string  `json:"error-app-tag,omitempty"`
	Path    string  `json:"error-path,omitempty"`
	Message string  `json:"error-message,omitempty"`
}

type errtype string
type errtag string

const (
	// error-type values
	errtypeProtocol    errtype = "protocol"
	errtypeApplication errtype = "application"

	// error-tag values
	errtagInvalidValue          errtag = "invalid-value"
	errtagOperationFailed       errtag = "operation-failed"
	errtagOperationNotSupported errtag = "operation-not-supported"
	errtagAccessDenied          errtag = "access-denied"
	errtagResourceDenied        errtag = "resource-denied"
	errtagInUse                 errtag = "in-use"
	errtagMalformedMessage      errtag = "malformed-message"
)

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
	status, entry := toErrorEntry(err, r)
	var resp errorResponse
	resp.Err.Arr = append(resp.Err.Arr, entry)
	data, _ = json.Marshal(&resp)
	mimeType = "application/yang-data+json"
	return
}

// toErrorEntry translates an error object into HTTP status and an
// errorEntry object.
func toErrorEntry(err error, r *http.Request) (status int, errInfo errorEntry) {
	// By default everything is 500 Internal Server Error
	status = http.StatusInternalServerError
	errInfo.Type = errtypeApplication
	errInfo.Tag = errtagOperationFailed

	switch e := err.(type) {
	case httpErrorType:
		status = e.status
		errInfo.Type = errtypeProtocol
		errInfo.Message = e.message

		// Guess error app tag from http status code
		switch status {
		case http.StatusBadRequest: // 400
			errInfo.Tag = errtagInvalidValue
		case http.StatusUnauthorized: // 401
			errInfo.Tag = errtagAccessDenied
		case http.StatusForbidden: // 403
			errInfo.Tag = errtagAccessDenied
		case http.StatusNotFound: // 404
			errInfo.Tag = errtagInvalidValue
		case http.StatusMethodNotAllowed: // 405
			errInfo.Tag = errtagOperationNotSupported
		case http.StatusUnsupportedMediaType:
			errInfo.Tag = errtagInvalidValue
		default: // 5xx and others
			errInfo.Tag = errtagOperationFailed
		}

	case tlerr.TranslibSyntaxValidationError:
		status = http.StatusBadRequest
		errInfo.Type = errtypeProtocol
		errInfo.Tag = errtagInvalidValue
		errInfo.Message = e.ErrorStr.Error()

	case tlerr.TranslibRedisClientEntryNotExist:
		status = http.StatusNotFound
		errInfo.Tag = errtagInvalidValue
		errInfo.Message = "Entry not found"

	case tlerr.TranslibCVLFailure:
		status = http.StatusInternalServerError
		errInfo.Tag = errtagInvalidValue
		errInfo.Message = e.CVLErrorInfo.ConstraintErrMsg
		errInfo.AppTag = e.CVLErrorInfo.ErrAppTag

		switch cvl.CVLRetCode(e.Code) {
		case cvl.CVL_SEMANTIC_KEY_ALREADY_EXIST, cvl.CVL_SEMANTIC_KEY_DUPLICATE:
			status = http.StatusConflict
			errInfo.Tag = errtagResourceDenied
			errInfo.Message = "Entry already exists"

		case cvl.CVL_SEMANTIC_KEY_NOT_EXIST:
			status = http.StatusNotFound
			errInfo.Tag = errtagInvalidValue
			errInfo.Message = "Entry not found"
		}

	case tlerr.TranslibTransactionFail:
		status = http.StatusConflict
		errInfo.Type = errtypeProtocol
		errInfo.Tag = errtagInUse
		errInfo.Message = "Transaction failed. Please try again."

	case tlerr.InternalError:
		errInfo.Message = e.Error()
		errInfo.Path = e.Path

	case tlerr.NotSupportedError:
		status = http.StatusMethodNotAllowed
		errInfo.Tag = errtagOperationNotSupported
		errInfo.Message = e.Error()
		errInfo.Path = e.Path

	case tlerr.InvalidArgsError:
		status = http.StatusBadRequest
		errInfo.Tag = errtagInvalidValue
		errInfo.Message = e.Error()
		errInfo.Path = e.Path

	case tlerr.NotFoundError:
		status = http.StatusNotFound
		errInfo.Tag = errtagInvalidValue
		errInfo.Message = e.Error()
		errInfo.Path = e.Path

	case tlerr.AlreadyExistsError:
		status = http.StatusConflict
		errInfo.Tag = errtagResourceDenied
		errInfo.Message = e.Error()
		errInfo.Path = e.Path

	}

	return
}
