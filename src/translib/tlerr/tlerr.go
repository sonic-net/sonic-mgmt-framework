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

/*
Package tlerr defines the errors of the translib library.

The Error() method of the error interface for these functions
returns the English version, and are only meant for log files.

For message strings that are returned to the users, the localization
will happen at when the GNMI/REST client's locale is known.
Hence, it cannot occur here.

*/
package tlerr

import (
	//	"fmt"
	"cvl"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	//	"errors"
	//	"strings"
)

var p *message.Printer

func init() {
	p = message.NewPrinter(language.English)
}

// DB errors

type TranslibDBCannotOpen struct {
}

func (e TranslibDBCannotOpen) Error() string {
	return p.Sprintf("Translib Redis Error: Cannot open DB")
}

type TranslibDBNotInit struct {
}

func (e TranslibDBNotInit) Error() string {
	return p.Sprintf("Translib Redis Error: DB Not Initialized")
}

type TranslibRedisClientEntryNotExist struct {
	Entry string
}

func (e TranslibRedisClientEntryNotExist) Error() string {
	return p.Sprintf("Translib Redis Error: Entry does not exist: %s", e.Entry)
}

type TranslibCVLFailure struct {
	Code         int
	CVLErrorInfo cvl.CVLErrorInfo
}

func (e TranslibCVLFailure) Error() string {
	return p.Sprintf("Translib Redis Error: CVL Failure: %d: %v", e.Code,
		e.CVLErrorInfo)
}

type TranslibTransactionFail struct {
}

func (e TranslibTransactionFail) Error() string {
	return p.Sprintf("Translib Redis Error: Transaction Fails")
}

type TranslibDBSubscribeFail struct {
}

func (e TranslibDBSubscribeFail) Error() string {
	return p.Sprintf("Translib Redis Error: DB Subscribe Fail")
}

type TranslibSyntaxValidationError struct {
	StatusCode int   // status code
	ErrorStr   error // error message
}

func (e TranslibSyntaxValidationError) Error() string {
	return p.Sprintf("%s", e.ErrorStr)
}
