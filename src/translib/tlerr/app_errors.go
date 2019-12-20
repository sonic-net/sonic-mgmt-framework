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

package tlerr

// "app errors" are used to return user displayable error messages
// to NB agents.

// errordata holds user displayable message template, arguments
// and optional path string.
type errordata struct {
	Format string        // message format string
	Args   []interface{} // message format arguments
	Path   string        // error path (optional)
}

// InvalidArgsError indicates bad request error.
type InvalidArgsError errordata

// NotFoundError indicates resource not found error.
type NotFoundError errordata

// AlreadyExistsError indicates resource exists error.
type AlreadyExistsError errordata

// NotSupportedError indicates unspported operation error.
type NotSupportedError errordata

// InternalError indicates a generic error during app execution.
type InternalError errordata

// AuthorizationError indicates the user is not authorized for an operation.
type AuthorizationError errordata

/////////////

func (e InvalidArgsError) Error() string {
	return p.Sprintf(e.Format, e.Args...)
}

// InvalidArgs creates a InvalidArgsError
func InvalidArgs(msg string, args ...interface{}) InvalidArgsError {
	return InvalidArgsError{Format: msg, Args: args}
}
func (e NotFoundError) Error() string {
	return p.Sprintf(e.Format, e.Args...)
}

// NotFound creates a NotFoundError
func NotFound(msg string, args ...interface{}) NotFoundError {
	return NotFoundError{Format: msg, Args: args}
}

func (e AlreadyExistsError) Error() string {
	return p.Sprintf(e.Format, e.Args...)
}

// AlreadyExists creates a AlreadyExistsError
func AlreadyExists(msg string, args ...interface{}) AlreadyExistsError {
	return AlreadyExistsError{Format: msg, Args: args}
}

func (e NotSupportedError) Error() string {
	return p.Sprintf(e.Format, e.Args...)
}

// NotSupported creates a NotSupportedError
func NotSupported(msg string, args ...interface{}) NotSupportedError {
	return NotSupportedError{Format: msg, Args: args}
}

func (e InternalError) Error() string {
	return p.Sprintf(e.Format, e.Args...)
}

// New creates an InternalError
func New(msg string, args ...interface{}) InternalError {
	return InternalError{Format: msg, Args: args}
}

func (e AuthorizationError) Error() string {
	return p.Sprintf(e.Format, e.Args...)
}
