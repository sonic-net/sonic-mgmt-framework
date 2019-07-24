/*
Copyright 2019 Broadcom. All rights reserved.
The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
*/

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
	"golang.org/x/text/message"
	"golang.org/x/text/language"
	"cvl"
//	"errors"
//	"strings"
)

var p * message.Printer

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
	Entry    string
}

func (e TranslibRedisClientEntryNotExist) Error() string {
	return p.Sprintf("Translib Redis Error: Entry does not exist: %s", e.Entry)
}

type TranslibCVLFailure struct {
	Code    int
	CVLErrorInfo  cvl.CVLErrorInfo
}

func (e TranslibCVLFailure) Error() string {
	return p.Sprintf("Translib Redis Error: CVL Failure: %d: %v", e.Code,
		e.CVLErrorInfo)
}

type TranslibDBSubscribeFail struct {
}

func (e TranslibDBSubscribeFail) Error() string {
	return p.Sprintf("Translib Redis Error: DB Subscribe Fail")
}

