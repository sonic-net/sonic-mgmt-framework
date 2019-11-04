//////////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Dell, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//////////////////////////////////////////////////////////////////////////

package translib

import (
	"errors"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"
)

type reqType int

const (
	opCreate reqType = iota + 1
	opDelete
	opUpdate
)

type ifType int

const (
	ETHERNET ifType = iota
	VLAN
	LAG
)

type dbEntry struct {
	op    reqType
	entry db.Value
}

type vlanData struct {
	vlanTs          *db.TableSpec
	vlanMemberTs    *db.TableSpec
	vlanTblTs       *db.TableSpec
	vlanMemberTblTs *db.TableSpec

	vlanMembersTableMap map[string]map[string]dbEntry
}

type lagData struct {
	lagTs              *db.TableSpec
	lagMemberTs        *db.TableSpec
	lagTblTs           *db.TableSpec
	lagIPTs            *db.TableSpec
	lagMembersTableMap map[string]map[string]dbEntry
}

type intfData struct {
	portTs             *db.TableSpec
	portTblTs          *db.TableSpec
	portOidCountrTblTs *db.TableSpec

	portOidMap  dbEntry
	portStatMap map[string]dbEntry

	intfIPTs        *db.TableSpec
	intfIPTblTs     *db.TableSpec
	intfCountrTblTs *db.TableSpec

	ifVlanInfoList []*ifVlan
}

type IntfApp struct {
	path       *PathInfo
	reqData    []byte
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}

	respJSON  interface{}
	allIpKeys []db.Key

	appDB      *db.DB
	countersDB *db.DB
	configDB   *db.DB

	intfType ifType
	mode     intfModeCfgAlone

	intfD intfData
	vlanD vlanData
	lagD  lagData

	ifTableMap   map[string]dbEntry
	ifIPTableMap map[string]map[string]dbEntry
}

func init() {
	log.Info("Init called for INTF module")
	err := register("/openconfig-interfaces:interfaces",
		&appInfo{appType: reflect.TypeOf(IntfApp{}),
			ygotRootType: reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces{}),
			isNative:     false})
	if err != nil {
		log.Fatal("Register INTF app module with App Interface failed with error=", err)
	}

	err = addModel(&ModelData{Name: "openconfig-interfaces",
		Org: "OpenConfig working group",
		Ver: "1.0.2"})
	if err != nil {
		log.Fatal("Adding model data to appinterface failed with error=", err)
	}
}

func (app *IntfApp) initializeInterface() {
	app.intfD.portTs = &db.TableSpec{Name: "PORT"}
	app.intfD.portTblTs = &db.TableSpec{Name: "PORT_TABLE"}
	app.intfD.portOidCountrTblTs = &db.TableSpec{Name: "COUNTERS_PORT_NAME_MAP"}

	app.intfD.portStatMap = make(map[string]dbEntry)

	app.intfD.intfIPTs = &db.TableSpec{Name: "INTERFACE"}
	app.intfD.intfIPTblTs = &db.TableSpec{Name: "INTF_TABLE", CompCt: 2}
	app.intfD.intfCountrTblTs = &db.TableSpec{Name: "COUNTERS"}

}

func (app *IntfApp) initializeVlan() {
	app.vlanD.vlanTs = &db.TableSpec{Name: "VLAN"}
	app.vlanD.vlanMemberTs = &db.TableSpec{Name: "VLAN_MEMBER"}
	app.vlanD.vlanTblTs = &db.TableSpec{Name: "VLAN_TABLE"}
	app.vlanD.vlanMemberTblTs = &db.TableSpec{Name: "VLAN_MEMBER_TABLE"}

	app.vlanD.vlanMembersTableMap = make(map[string]map[string]dbEntry)
}

func (app *IntfApp) initializeLag() {
	app.lagD.lagTs = &db.TableSpec{Name: "PORTCHANNEL"}
	app.lagD.lagMemberTs = &db.TableSpec{Name: "PORTCHANNEL_MEMBER"}
	app.lagD.lagIPTs = &db.TableSpec{Name: "PORTCHANNEL_INTERFACE"}
	app.lagD.lagTblTs = &db.TableSpec{Name: "LAG_TABLE"}

	app.lagD.lagMembersTableMap = make(map[string]map[string]dbEntry)
}

func (app *IntfApp) initialize(data appData) {
	log.Info("initialize:if:path =", data.path)

	app.path = NewPathInfo(data.path)
	app.reqData = data.payload
	app.ygotRoot = data.ygotRoot
	app.ygotTarget = data.ygotTarget

	app.ifTableMap = make(map[string]dbEntry)
	app.ifIPTableMap = make(map[string]map[string]dbEntry)

	app.initializeInterface()
	app.initializeVlan()
	app.initializeLag()
}

func (app *IntfApp) getAppRootObject() *ocbinds.OpenconfigInterfaces_Interfaces {
	deviceObj := (*app.ygotRoot).(*ocbinds.Device)
	return deviceObj.Interfaces
}

func (app *IntfApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCreate:intf:path =", app.path)

	err = errors.New("Not implemented")
	return keys, err
}

/* Reason why we don't have the Interface Type validation at beginning is due to,
   the fact that, you could get mixture of different interfaces from GNMI or rest.
   So Ideally, you need to initialize the DS for all the interface types. */
func (app *IntfApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	pathInfo := app.path

	log.Infof("Received UPDATE for path %s; vars=%v", pathInfo.Template, pathInfo.Vars)

	intfObj := app.getAppRootObject()

	targetUriPath, err := getYangPathFromUri(app.path.Path)
	log.Info("uripath:=", targetUriPath)
	log.Info("err:=", err)

	if intfObj.Interface != nil && len(intfObj.Interface) > 0 {
		log.Info("len:=", len(intfObj.Interface))
		for ifKey, _ := range intfObj.Interface {
			log.Info("Name:=", ifKey)
			err := app.getIntfTypeFromIntf(&ifKey)
			if err != nil {
				errStr := "Invalid Interface type:" + ifKey
				ifValidErr := tlerr.InvalidArgsError{Format: errStr}
				return keys, ifValidErr
			}
			switch app.intfType {
			case ETHERNET:
				keys, err = app.translateUpdatePhyIntf(d, &ifKey, opUpdate)
				if err != nil {
					return keys, err
				}
			case VLAN:
				keys, err = app.translateUpdateVlanIntf(d, &ifKey, opUpdate)
				if err != nil {
					return keys, err
				}
			case LAG:
				keys, err = app.translateUpdateLagIntf(d, &ifKey, opUpdate)
				if err != nil {
					return keys, err
				}
			}
		}
	}
	return keys, err
}

func (app *IntfApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateReplace:intf:path =", app.path)
	err = errors.New("Not implemented")
	return keys, err
}

func (app *IntfApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	pathInfo := app.path

	log.Infof("Received Delete for path %s; vars=%v", pathInfo.Template, pathInfo.Vars)

	intfObj := app.getAppRootObject()

	targetUriPath, err := getYangPathFromUri(app.path.Path)
	log.Info("uripath:=", targetUriPath)
	log.Info("err:=", err)

	if intfObj.Interface != nil && len(intfObj.Interface) > 0 {
		log.Info("len:=", len(intfObj.Interface))
		for ifKey, _ := range intfObj.Interface {
			log.Info("Name:=", ifKey)

			err := app.getIntfTypeFromIntf(&ifKey)
			if err != nil {
				errStr := "Invalid Interface type:" + ifKey
				ifValidErr := tlerr.InvalidArgsError{Format: errStr}
				return keys, ifValidErr
			}
			switch app.intfType {
			case ETHERNET:
				keys, err = app.translateDeletePhyIntf(d, &ifKey)
				if err != nil {
					return keys, err
				}
			case VLAN:
				keys, err = app.translateDeleteVlanIntf(d, &ifKey)
				if err != nil {
					return keys, err
				}
			case LAG:
				keys, err = app.translateDeleteLagIntf(d, &ifKey)
				if err != nil {
					return keys, err
				}
			}
		}
	} else {
		err = errors.New("Not implemented")
	}
	return keys, err
}

func (app *IntfApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	var err error
	log.Info("translateGet:intf:path =", app.path)
	return err
}

func (app *IntfApp) translateAction(dbs [db.MaxDB]*db.DB) error {
	err := errors.New("Not supported")
	return err
}

func (app *IntfApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
	app.appDB = dbs[db.ApplDB]
	pathInfo := NewPathInfo(path)
	notSupported := tlerr.NotSupportedError{Format: "Subscribe not supported", Path: path}

	if isSubtreeRequest(pathInfo.Template, "/openconfig-interfaces:interfaces") {
		if pathInfo.HasSuffix("/interface{name}") ||
			pathInfo.HasSuffix("/config") ||
			pathInfo.HasSuffix("/state") {
			log.Errorf("Subscribe not supported for %s!", pathInfo.Template)
			return nil, nil, notSupported
		}
		ifKey := pathInfo.Var("name")
		if len(ifKey) == 0 {
			return nil, nil, errors.New("ifKey given is empty!")
		}

		log.Info("Interface name = ", ifKey)

		err := app.getIntfTypeFromIntf(&ifKey)
		if err != nil {
			return nil, nil, err
		}

		switch app.intfType {
		case ETHERNET:
			return app.translateSubscribePhyIntf(&ifKey, pathInfo)
		case VLAN:
			break
		}
	}
	return nil, nil, notSupported
}

func (app *IntfApp) processCreate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processCreate:intf:path =", app.path)
	log.Info("ProcessCreate: Target Type is " + reflect.TypeOf(*app.ygotTarget).Elem().Name())

	err = errors.New("Not implemented")
	return resp, err
}

func (app *IntfApp) processUpdate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processUpdate:intf:path =", app.path)
	log.Info("ProcessUpdate: Target Type is " + reflect.TypeOf(*app.ygotTarget).Elem().Name())

	switch app.intfType {
	case ETHERNET:
		err = app.processUpdatePhyIntf(d)
		if err != nil {
			return resp, err
		}
	case VLAN:
		err = app.processUpdateVlanIntf(d)
		if err != nil {
			return resp, err
		}
	case LAG:
		err = app.processUpdateLagIntf(d)
		if err != nil {
			return resp, err
		}
	}
	return resp, err
}

func (app *IntfApp) processReplace(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processReplace:intf:path =", app.path)
	err = errors.New("Not implemented")
	return resp, err
}

func (app *IntfApp) processDelete(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processDelete:intf:path =", app.path)

	switch app.intfType {
	case ETHERNET:
		err = app.processDeletePhyIntf(d)
		if err != nil {
			return resp, err
		}
	case VLAN:
		err = app.processDeleteVlanIntf(d)
		if err != nil {
			return resp, err
		}
	case LAG:
		err = app.processDeleteLagIntf(d)
		if err != nil {
			return resp, err
		}
	}
	return resp, err
}

func (app *IntfApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {

	var err error
	var payload []byte
	pathInfo := app.path

	log.Infof("Received GET for path %s; template: %s vars=%v", pathInfo.Path, pathInfo.Template, pathInfo.Vars)
	app.appDB = dbs[db.ApplDB]
	app.countersDB = dbs[db.CountersDB]
	app.configDB = dbs[db.ConfigDB]

	targetUriPath, err := getYangPathFromUri(app.path.Path)
	log.Info("URI Path = ", targetUriPath)

	if isSubtreeRequest(targetUriPath, "/openconfig-interfaces:interfaces/interface") {
		return app.processGetSpecificIntf(dbs, &targetUriPath)
	}
	if isSubtreeRequest(targetUriPath, "/openconfig-interfaces:interfaces") {
		return app.processGetAllInterfaces(dbs)
	}
	return GetResponse{Payload: payload}, err
}

func (app *IntfApp) processAction(dbs [db.MaxDB]*db.DB) (ActionResponse, error) {
	var resp ActionResponse
	err := errors.New("Not implemented")

	return resp, err
}
