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
	"errors"
	"fmt"
	"strings"
	"testing"

	"cvl"
	"translib/tlerr"
)

func init() {
	fmt.Println("+++++ init error_test +++++")
}

func TestProtoError(t *testing.T) {
	t.Run("400", testProtoError(
		httpBadRequest("Bad %s", "json"), 400, "Bad json"))

	t.Run("500", testProtoError(
		httpServerError("Failed"), 500, "Failed"))

	t.Run("XXX", testProtoError(
		httpError(401, "Invalid user"), 401, "Invalid user"))
}

func testProtoError(err error, status int, msg string) func(*testing.T) {
	return func(t *testing.T) {
		e, ok := err.(httpErrorType)
		if !ok {
			t.Error("not a httpErrorType")
		} else if e.status != status || chkmsg(e.message, msg) == false {
			t.Errorf("expecting %d/'%s'; found %d/'%s'",
				status, msg, e.status, e.message)
		}
	}
}

func TestErrorEntry(t *testing.T) {

	// errorEntry mapping for server errors

	t.Run("RequestReadError", testErrorEntry(
		httpBadRequest("hii"),
		400, "protocol", "invalid-value", "", "hii"))

	t.Run("GenericServerError", testErrorEntry(
		httpServerError("hii"),
		500, "protocol", "operation-failed", "", "hii"))

	t.Run("AuthenticationFailed", testErrorEntry(
		httpError(401, "hii"),
		401, "protocol", "access-denied", "", "hii"))

	t.Run("AuthorizationFailed", testErrorEntry(
		httpError(403, "hii"),
		403, "protocol", "access-denied", "", "hii"))

	t.Run("NotFound", testErrorEntry(
		httpError(404, "404 NotFound."),
		404, "protocol", "invalid-value", "", "404 NotFound."))

	t.Run("NotSupported", testErrorEntry(
		httpError(405, "405 NotSupported."),
		405, "protocol", "operation-not-supported", "", "405 NotSupported."))

	t.Run("UnknownMediaType", testErrorEntry(
		httpError(415, "hii"),
		415, "protocol", "invalid-value", "", "hii"))

	// errorEntry mapping for unknown errors

	t.Run("UnknownError", testErrorEntry(
		errors.New("hii"),
		500, "application", "operation-failed", "", ""))

	// errorEntry mapping for app errors

	t.Run("InvalidArgs", testErrorEntry(
		tlerr.InvalidArgsError{Format: "hii", Path: "xyz"},
		400, "application", "invalid-value", "xyz", "hii"))

	t.Run("ResourceNotFound", testErrorEntry(
		tlerr.NotFoundError{Format: "hii", Path: "xyz"},
		404, "application", "invalid-value", "xyz", "hii"))

	t.Run("AlreadyExists", testErrorEntry(
		tlerr.AlreadyExistsError{Format: "hii", Path: "xyz"},
		409, "application", "resource-denied", "xyz", "hii"))

	t.Run("UnsupportedOper", testErrorEntry(
		tlerr.NotSupportedError{Format: "hii", Path: "xyz"},
		405, "application", "operation-not-supported", "xyz", "hii"))

	t.Run("AppGenericErr", testErrorEntry(
		tlerr.InternalError{Format: "hii", Path: "xyz"},
		500, "application", "operation-failed", "xyz", "hii"))

	// errorEntry mapping for DB errors

	t.Run("DB_EntryNotExist", testErrorEntry(
		tlerr.TranslibRedisClientEntryNotExist{},
		404, "application", "invalid-value", "", "Entry not found"))

	t.Run("TransactionFailed", testErrorEntry(
		tlerr.TranslibTransactionFail{},
		409, "protocol", "in-use", "", "*"))

	t.Run("DB_CannotOpen", testErrorEntry(
		tlerr.TranslibDBCannotOpen{},
		500, "application", "operation-failed", "", ""))

	t.Run("DB_NotInit", testErrorEntry(
		tlerr.TranslibDBNotInit{},
		500, "application", "operation-failed", "", ""))

	t.Run("DB_SubscribeFailed", testErrorEntry(
		tlerr.TranslibDBSubscribeFail{},
		500, "application", "operation-failed", "", ""))

	// errorEntry mapping for CVL errors

	t.Run("CVL_KeyNotExists", testErrorEntry(
		cvlError(cvl.CVL_SEMANTIC_KEY_NOT_EXIST, "hii"),
		404, "application", "invalid-value", "", "Entry not found"))

	t.Run("CVL_KeyExists", testErrorEntry(
		cvlError(cvl.CVL_SEMANTIC_KEY_ALREADY_EXIST, "hii"),
		409, "application", "resource-denied", "", "Entry already exists"))

	t.Run("CVL_KeyDup", testErrorEntry(
		cvlError(cvl.CVL_SEMANTIC_KEY_DUPLICATE, "hii"),
		409, "application", "resource-denied", "", "Entry already exists"))

	t.Run("CVL_SemanticErr", testErrorEntry(
		cvlError(cvl.CVL_SEMANTIC_ERROR, "hii"),
		500, "application", "invalid-value", "", "hii"))

	// errorEntry mapping for YGOT errors
	t.Run("YGOT_400", testErrorEntry(
		tlerr.TranslibSyntaxValidationError{ErrorStr: errors.New("ygot")},
		400, "protocol", "invalid-value", "", "ygot"))

}

func testErrorEntry(err error,
	expStatus int, expType, expTag, expPath, expMessage string) func(*testing.T) {
	return func(t *testing.T) {
		status, entry := toErrorEntry(err, nil)
		if status != expStatus || string(entry.Type) != expType ||
			string(entry.Tag) != expTag || entry.Path != expPath ||
			chkmsg(entry.Message, expMessage) == false {
			t.Errorf("%T: expecting %d/%s/%s/\"%s\"/\"%s\"; found %d/%s/%s/\"%s\"/\"%s\"",
				err, expStatus, expType, expTag, expPath, expMessage,
				status, entry.Type, entry.Tag, entry.Path, entry.Message)
		}
	}
}

func TestErrorResponse(t *testing.T) {
	t.Run("WithMsg", testErrorResponse(
		tlerr.NotFoundError{Format: "hii", Path: "xyz"},
		404, "{\"ietf-restconf:errors\":{\"error\":[{"+
			"\"error-type\":\"application\",\"error-tag\":\"invalid-value\","+
			"\"error-path\":\"xyz\",\"error-message\":\"hii\"}]}}"))

	t.Run("NoMsg", testErrorResponse(
		errors.New("hii"),
		500, "{\"ietf-restconf:errors\":{\"error\":[{"+
			"\"error-type\":\"application\",\"error-tag\":\"operation-failed\"}]}}"))
}

func testErrorResponse(err error, expStatus int, expData string) func(*testing.T) {
	return func(t *testing.T) {
		status, data, ctype := prepareErrorResponse(err, nil)

		if status != expStatus {
			t.Errorf("FAIL: bad status %d; expected %d", status, expStatus)
		} else if ctype != "application/yang-data+json" {
			t.Errorf("FAIL: bad content-type '%s'", ctype)
		} else if string(data) != expData {
			t.Errorf("FAIL: bad data %s", data)
			t.Errorf("expected %s", expData)
		}
	}
}

func chkmsg(actual, expected string) bool {
	if expected == "*" {
		return true
	}
	if strings.HasPrefix(expected, "!") {
		return actual != expected[1:]
	}
	return actual == expected
}

func cvlError(code cvl.CVLRetCode, msg string) error {
	return tlerr.TranslibCVLFailure{
		Code: int(code),
		CVLErrorInfo: cvl.CVLErrorInfo{
			ErrCode:          code,
			TableName:        "unknown",
			CVLErrDetails:    "blah blah blah",
			Msg:              "ignore me",
			ConstraintErrMsg: msg,
		},
	}
}
