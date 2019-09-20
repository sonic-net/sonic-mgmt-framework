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
Package translib defines the interface for all the app modules 

It exposes register function for all the app modules to register

It stores all the app module information in a map and presents it

to the tranlib infra when it asks for the same.
*/

package translib

import (
	"errors"
	"reflect"
	"strings"
	"translib/db"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
)

//Structure containing app module information
type appInfo struct {
	appType			reflect.Type
	ygotRootType	reflect.Type
	isNative		bool
	tablesToWatch   []*db.TableSpec
}

//Structure containing the app data coming from translib infra
type appData struct {
	path       string
	payload    []byte
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}
}

//map containing the base path to app module info
var appMap map[string]*appInfo

//array containing all the supported models
var models []ModelData

//Interface for all App Modules
type appInterface interface {
	initialize(data appData)
	translateCreate(d *db.DB) ([]db.WatchKeys, error)
	translateUpdate(d *db.DB) ([]db.WatchKeys, error)
	translateReplace(d *db.DB) ([]db.WatchKeys, error)
	translateDelete(d *db.DB) ([]db.WatchKeys, error)
	translateGet(dbs [db.MaxDB]*db.DB) error
	translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error)
	processCreate(d *db.DB) (SetResponse, error)
	processUpdate(d *db.DB) (SetResponse, error)
	processReplace(d *db.DB) (SetResponse, error)
	processDelete(d *db.DB) (SetResponse, error)
	processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error)
}

//App modules will use this function to register with App interface during boot up
func register(path string, info *appInfo) error {
	var err error
    log.Info("Registering for path =", path)

	if appMap == nil {
		appMap = make(map[string]*appInfo)
	}

	if _, ok := appMap[path]; ok == false {

		appMap[path] = info

	} else {
		log.Fatal("Duplicate path being registered. Path =", path)
		err = errors.New("Duplicate path")
	}

	return err
}

//Adds the model information to the supported models array
func addModel(model *ModelData) error {
	var err error

	models = append(models, *model)

	//log.Info("Models = ", models)
	return err
}

//App modules can use this function to unregister itself from the app interface
func unregister(path string) error {
	var err error
	log.Info("Unregister for path =", path)

	_, ok := appMap[path]

	if ok {
		log.Info("deleting for path =", path)
		delete(appMap, path)
	}

	return err
}

//Translib infra will use this function get the app info for a given path
func getAppModuleInfo(path string) (*appInfo, error) {
	var err error
	log.Info("getAppModule called for path =", path)

	for pattern, app := range appMap {
		if !strings.HasPrefix(path, pattern) {
			continue
		}

		log.Info("found the entry in the map for path =", pattern)

		return app, err
	}

	/* If no specific app registered fallback to default/common app */
	log.Infof("No app module registered for path %s hence fallback to default/common app", path)
	app := appMap["*"]

	return app, err
}

//Get all the supported models
func getModels() []ModelData {

	return models
}

//Creates a new app from the appType and returns it as an appInterface
func getAppInterface(appType reflect.Type) (appInterface, error) {
    var err error
    appInstance := reflect.New(appType)
    app, ok := appInstance.Interface().(appInterface)

    if !ok {
        err = errors.New("Invalid appType")
        log.Fatal("Appmodule does not confirm to appInterface method conventions for appType=", appType)
    } else {
        log.Info("cast to appInterface worked", app)
    }

	return app, err
}
