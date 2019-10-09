///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package translib

import (
	"encoding/json"
	"reflect"
	"strings"
	"translib/db"
	"translib/tlerr"

	"github.com/golang/glog"
)

type apiTests struct {
	path string
	body []byte

	echoMsg string
	echoErr string
}

func init() {
	err := register("/api-tests:",
		&appInfo{
			appType:       reflect.TypeOf(apiTests{}),
			isNative:      true,
			tablesToWatch: nil})

	if err != nil {
		glog.Fatalf("Failed to register ApiTest app; %v", err)
	}
}

func (app *apiTests) initialize(inp appData) {
	app.path = inp.path
	app.body = inp.payload
}

func (app *apiTests) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	return nil, app.translatePath()
}

func (app *apiTests) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	return nil, app.translatePath()
}

func (app *apiTests) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	return nil, app.translatePath()
}

func (app *apiTests) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	return nil, app.translatePath()
}

func (app *apiTests) translateGet(dbs [db.MaxDB]*db.DB) error {
	return app.translatePath()
}

func (app *apiTests) translateAction(dbs [db.MaxDB]*db.DB) error {
	var req struct {
		Input struct {
			Message string `json:"message"`
			ErrType string `json:"error-type"`
		} `json:"api-tests:input"`
	}

	err := json.Unmarshal(app.body, &req)
	if err != nil {
		glog.Errorf("Failed to parse rpc input; err=%v", err)
		return tlerr.InvalidArgs("Invalid rpc input")
	}

	app.echoMsg = req.Input.Message
	app.echoErr = req.Input.ErrType

	return nil
}

func (app *apiTests) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
	return nil, nil, nil
}

func (app *apiTests) processCreate(d *db.DB) (SetResponse, error) {
	return app.processSet()
}

func (app *apiTests) processUpdate(d *db.DB) (SetResponse, error) {
	return app.processSet()
}

func (app *apiTests) processReplace(d *db.DB) (SetResponse, error) {
	return app.processSet()
}

func (app *apiTests) processDelete(d *db.DB) (SetResponse, error) {
	return app.processSet()
}

func (app *apiTests) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
	var gr GetResponse
	err := app.getError()
	if err == nil {
		gr.Payload, err = json.Marshal(&app.echoMsg)
	}
	return gr, err
}

func (app *apiTests) processAction(dbs [db.MaxDB]*db.DB) (ActionResponse, error) {
	var ar ActionResponse

	err := app.getError()
	if err == nil {
		var respData struct {
			Output struct {
				Message string `json:"message"`
			} `json:"api-tests:output"`
		}

		respData.Output.Message = app.echoMsg
		ar.Payload, err = json.Marshal(&respData)
	}

	return ar, err
}

func (app *apiTests) translatePath() error {
	app.echoMsg = "Hello, world!"
	k := strings.Index(app.path, "error/")
	if k >= 0 {
		app.echoErr = app.path[k+6:]
	}
	return nil
}

func (app *apiTests) processSet() (SetResponse, error) {
	var sr SetResponse
	err := app.getError()
	return sr, err
}

func (app *apiTests) getError() error {
	switch strings.ToLower(app.echoErr) {
	case "invalid-args", "invalidargs":
		return tlerr.InvalidArgs(app.echoMsg)
	case "exists":
		return tlerr.AlreadyExists(app.echoMsg)
	case "not-found", "notfound":
		return tlerr.NotFound(app.echoMsg)
	case "not-supported", "notsupported", "unsupported":
		return tlerr.NotSupported(app.echoMsg)
	case "", "no", "none", "false":
		return nil
	default:
		return tlerr.New(app.echoMsg)
	}
}
