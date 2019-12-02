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

package translib

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"

	"github.com/facette/natsort"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/util"
	"github.com/openconfig/ygot/ygot"
)

const (
	STP_GLOBAL_TABLE         = "STP"
	STP_VLAN_TABLE           = "STP_VLAN"
	STP_VLAN_INTF_TABLE      = "STP_VLAN_INTF"
	STP_INTF_TABLE           = "STP_INTF"
	STP_VLAN_OPER_TABLE      = "_STP_VLAN_TABLE"
	STP_VLAN_INTF_OPER_TABLE = "_STP_VLAN_INTF_TABLE"
	STP_INTF_OPER_TABLE      = "_STP_INTF_TABLE"
	STP_MODE                 = "mode"
	OC_STP_APP_MODULE_NAME   = "/openconfig-spanning-tree:stp"
	OC_STP_YANG_PATH_PREFIX  = "/device/stp"
	PVST_MAX_INSTANCES       = 255

	STP_DEFAULT_ROOT_GUARD_TIMEOUT = "30"
	STP_DEFAULT_FORWARD_DELAY      = "15"
	STP_DEFAULT_HELLO_INTERVAL     = "2"
	STP_DEFAULT_MAX_AGE            = "20"
	STP_DEFAULT_BRIDGE_PRIORITY    = "32768"
	STP_DEFAULT_BPDU_FILTER        = "false"
)

type StpApp struct {
	pathInfo   *PathInfo
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}

	globalTable    *db.TableSpec
	vlanTable      *db.TableSpec
	vlanIntfTable  *db.TableSpec
	interfaceTable *db.TableSpec

	vlanOperTable     *db.TableSpec
	vlanIntfOperTable *db.TableSpec
	intfOperTable     *db.TableSpec

	globalInfo       db.Value
	vlanTableMap     map[string]db.Value
	vlanIntfTableMap map[string]map[string]db.Value
	intfTableMap     map[string]db.Value

	vlanOperTableMap     map[string]db.Value
	vlanIntfOperTableMap map[string]map[string]db.Value
	intfOperTableMap     map[string]db.Value

	appDB *db.DB
}

func init() {
	err := register("/openconfig-spanning-tree:stp",
		&appInfo{appType: reflect.TypeOf(StpApp{}),
			ygotRootType:  reflect.TypeOf(ocbinds.OpenconfigSpanningTree_Stp{}),
			isNative:      false,
			tablesToWatch: []*db.TableSpec{&db.TableSpec{Name: STP_GLOBAL_TABLE}, &db.TableSpec{Name: STP_VLAN_TABLE}, &db.TableSpec{Name: STP_VLAN_INTF_TABLE}, &db.TableSpec{Name: STP_INTF_TABLE}}})

	if err != nil {
		log.Fatal("Register STP app module with App Interface failed with error=", err)
	}

	err = addModel(&ModelData{Name: "openconfig-spanning-tree", Org: "OpenConfig working group", Ver: "0.3.0"})
	if err != nil {
		log.Fatal("Adding model data to appinterface failed with error=", err)
	}
}

func (app *StpApp) initialize(data appData) {
	log.Info("initialize:stp:path =", data.path)
	app.pathInfo = NewPathInfo(data.path)
	app.ygotRoot = data.ygotRoot
	app.ygotTarget = data.ygotTarget

	app.globalTable = &db.TableSpec{Name: STP_GLOBAL_TABLE}
	app.vlanTable = &db.TableSpec{Name: STP_VLAN_TABLE}
	app.vlanIntfTable = &db.TableSpec{Name: STP_VLAN_INTF_TABLE}
	app.interfaceTable = &db.TableSpec{Name: STP_INTF_TABLE}

	app.vlanOperTable = &db.TableSpec{Name: STP_VLAN_OPER_TABLE}
	app.vlanIntfOperTable = &db.TableSpec{Name: STP_VLAN_INTF_OPER_TABLE}
	app.intfOperTable = &db.TableSpec{Name: STP_INTF_OPER_TABLE}

	app.globalInfo = db.Value{Field: map[string]string{}}
	app.vlanTableMap = make(map[string]db.Value)
	app.intfTableMap = make(map[string]db.Value)
	app.vlanIntfTableMap = make(map[string]map[string]db.Value)

	app.vlanOperTableMap = make(map[string]db.Value)
	app.vlanIntfOperTableMap = make(map[string]map[string]db.Value)
	app.intfOperTableMap = make(map[string]db.Value)
}

func (app *StpApp) getAppRootObject() *ocbinds.OpenconfigSpanningTree_Stp {
	deviceObj := (*app.ygotRoot).(*ocbinds.Device)
	return deviceObj.Stp
}

func (app *StpApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCreate:stp:path =", app.pathInfo.Template)

	keys, err = app.translateCRUCommon(d, CREATE)
	return keys, err
}

func (app *StpApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateUpdate:stp:path =", app.pathInfo.Template)

	keys, err = app.translateCRUCommon(d, UPDATE)
	return keys, err
}

func (app *StpApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateReplace:stp:path =", app.pathInfo.Template)

	keys, err = app.translateCRUCommon(d, REPLACE)
	return keys, err
}

func (app *StpApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateDelete:stp:path =", app.pathInfo.Template)

	return keys, err
}

func (app *StpApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	var err error
	log.Info("translateGet:stp:path =", app.pathInfo.Template)
	return err
}

func (app *StpApp) translateAction(dbs [db.MaxDB]*db.DB) error {
	err := errors.New("Not supported")
	return err
}

func (app *StpApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
	var err error
	return nil, nil, err
}

func (app *StpApp) processCreate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	if err = app.processCommon(d, CREATE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}
	return resp, err
}

func (app *StpApp) processUpdate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	if err = app.processCommon(d, UPDATE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}
	return resp, err
}

func (app *StpApp) processReplace(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	if err = app.processCommon(d, REPLACE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}
	return resp, err
}

func (app *StpApp) processDelete(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	if err = app.processCommon(d, DELETE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}
	return resp, err
}

func (app *StpApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
	var err error
	var payload []byte

	app.appDB = dbs[db.ApplDB]

	configDb := dbs[db.ConfigDB]
	err = app.processCommon(configDb, GET)
	if err != nil {
		return GetResponse{Payload: payload, ErrSrc: AppErr}, err
	}

	payload, err = generateGetResponsePayload(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device), app.ygotTarget)
	if err != nil {
		return GetResponse{Payload: payload, ErrSrc: AppErr}, err
	}

	return GetResponse{Payload: payload}, err
}

func (app *StpApp) processAction(dbs [db.MaxDB]*db.DB) (ActionResponse, error) {
	var resp ActionResponse
	err := errors.New("Not implemented")

	return resp, err
}

func (app *StpApp) translateCRUCommon(d *db.DB, opcode int) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCRUCommon:STP:path =", app.pathInfo.Template)

	err = app.convertOCStpGlobalConfToInternal(opcode)
	if err != nil {
		return keys, err
	}
	err = app.convertOCPvstToInternal(opcode)
	if err != nil {
		return keys, err
	}
	err = app.convertOCRpvstConfToInternal(opcode)
	if err != nil {
		return keys, err
	}
	app.convertOCStpInterfacesToInternal()

	return keys, err
}

func (app *StpApp) processCommon(d *db.DB, opcode int) error {
	var err error
	var topmostPath bool = false
	stp := app.getAppRootObject()

	log.Infof("processCommon--Path Received: %s", app.pathInfo.Template)
	targetType := reflect.TypeOf(*app.ygotTarget)
	if !util.IsValueScalar(reflect.ValueOf(*app.ygotTarget)) && util.IsValuePtr(reflect.ValueOf(*app.ygotTarget)) {
		log.Infof("processCommon: Target object is a <%s> of Type: %s", targetType.Kind().String(), targetType.Elem().Name())
		if targetType.Elem().Name() == "OpenconfigSpanningTree_Stp" {
			topmostPath = true
		}
	}

	targetUriPath, _ := getYangPathFromUri(app.pathInfo.Path)
	log.Infof("processCommon -- isTopmostPath: %t and Uri: %s", topmostPath, targetUriPath)
	if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/global") {
		switch opcode {
		case CREATE:
			err = app.enableStpMode(d)
		case REPLACE, UPDATE:
			mode, _ := app.getStpModeFromConfigDB(d)
			if *app.ygotTarget == stp.Global || *app.ygotTarget == stp.Global.Config || targetUriPath == "/openconfig-spanning-tree:stp/global/config/enabled-protocol" {
				if len(mode) == 0 {
					err = app.enableStpMode(d)
				} else {
					if mode != app.convertOCStpModeToInternal(stp.Global.Config.EnabledProtocol[0]) {
						return tlerr.InvalidArgs("STP mode is configured as %s", mode)
					} else {
						app.handleStpGlobalFieldsUpdation(d, opcode)
					}
				}
			} else {
				app.handleStpGlobalFieldsUpdation(d, opcode)
			}
		case DELETE:
			if *app.ygotTarget == stp.Global || *app.ygotTarget == stp.Global.Config || targetUriPath == "/openconfig-spanning-tree:stp/global/config/enabled-protocol" {
				if app.pathInfo.Template == "/openconfig-spanning-tree:stp/global/config/enabled-protocol{}" {
					mode, _ := app.getStpModeFromConfigDB(d)
					if mode != app.convertOCStpModeToInternal(stp.Global.Config.EnabledProtocol[0]) {
						return tlerr.InvalidArgs("STP mode is configured as %s", mode)
					}
				}
				err = app.disableStpMode(d)
			} else {
				err = app.handleStpGlobalFieldsDeletion(d)
			}
		case GET:
			err = app.convertDBStpGlobalConfigToInternal(d)
			if err != nil {
				return err
			}
			ygot.BuildEmptyTree(stp)
			app.convertInternalToOCStpGlobalConfig(stp.Global)
		}
	} else if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/openconfig-spanning-tree-ext:pvst") {
		mode, _ := app.getStpModeFromConfigDB(d)
		if mode != "pvst" {
			return tlerr.InvalidArgs("STP mode is configured as %s", mode)
		}
		if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/openconfig-spanning-tree-ext:pvst/vlan{}") {
			for vlanId, _ := range stp.Pvst.Vlan {
				pvstVlan := stp.Pvst.Vlan[vlanId]
				vlanName := "Vlan" + strconv.Itoa(int(vlanId))
				if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/openconfig-spanning-tree-ext:pvst/vlan{}/interfaces/interface{}") {
					// Subtree of one interface under a vlan
					for intfId, _ := range pvstVlan.Interfaces.Interface {
						pvstVlanIntf := pvstVlan.Interfaces.Interface[intfId]
						switch opcode {
						case CREATE, REPLACE, UPDATE, DELETE:
							err = app.handleRpvstCRUDOperationsAtVlanInterfaceLevel(d, opcode, vlanName, intfId, pvstVlan, pvstVlanIntf)
							if err != nil {
								return err
							}
						case GET:
							err = app.convertDBRpvstVlanInterfaceToInternal(d, vlanName, intfId, asKey(vlanName, intfId), true)
							if err != nil {
								return err
							}
							ygot.BuildEmptyTree(pvstVlanIntf)
							app.convertInternalToOCPvstVlanInterface(vlanName, intfId, pvstVlan, pvstVlanIntf)
							// populate operational data
							app.convertOperInternalToOCVlanInterface(vlanName, intfId, pvstVlan, pvstVlanIntf)
						}
					}
				} else {
					isInterfacesSubtree := isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/openconfig-spanning-tree-ext:pvst/vlan{}/interfaces")
					switch opcode {
					case CREATE, REPLACE, UPDATE, DELETE:
						err = app.handleRpvstCRUDOperationsAtVlanLevel(d, opcode, vlanName, isInterfacesSubtree, stp.Pvst, pvstVlan)
						if err != nil {
							return err
						}
					case GET:
						err = app.convertDBRpvstVlanConfigToInternal(d, asKey(vlanName))
						if err != nil {
							return err
						}
						ygot.BuildEmptyTree(pvstVlan)
						app.convertInternalToOCPvstVlan(vlanName, stp.Pvst, pvstVlan)
					}
				}
			}
		} else {
			// Handle top PVST
			switch opcode {
			case CREATE, REPLACE, UPDATE, DELETE:
				log.Infof("Implementation in progress for URL: %s", app.pathInfo.Template)
				return tlerr.NotSupported("Implementation in progress")
			case GET:
				ygot.BuildEmptyTree(stp.Pvst)
				err = app.convertDBRpvstVlanConfigToInternal(d, db.Key{})
				if err != nil {
					return err
				}
				app.convertInternalToOCPvstVlan("", stp.Pvst, nil)
			}
		}
	} else if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/rapid-pvst") {
		mode, _ := app.getStpModeFromConfigDB(d)
		if mode != "rpvst" {
			return tlerr.InvalidArgs("STP mode is configured as %s", mode)
		}
		if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/rapid-pvst/vlan{}") {
			for vlanId, _ := range stp.RapidPvst.Vlan {
				rpvstVlanConf := stp.RapidPvst.Vlan[vlanId]
				vlanName := "Vlan" + strconv.Itoa(int(vlanId))
				if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/rapid-pvst/vlan{}/interfaces/interface{}") {
					// Subtree of one interface under a vlan
					for intfId, _ := range rpvstVlanConf.Interfaces.Interface {
						rpvstVlanIntfConf := rpvstVlanConf.Interfaces.Interface[intfId]
						switch opcode {
						case CREATE, REPLACE, UPDATE, DELETE:
							err = app.handleRpvstCRUDOperationsAtVlanInterfaceLevel(d, opcode, vlanName, intfId, rpvstVlanConf, rpvstVlanIntfConf)
							if err != nil {
								return err
							}
						case GET:
							err = app.convertDBRpvstVlanInterfaceToInternal(d, vlanName, intfId, asKey(vlanName, intfId), true)
							if err != nil {
								return err
							}
							ygot.BuildEmptyTree(rpvstVlanIntfConf)
							app.convertInternalToOCRpvstVlanInterface(vlanName, intfId, rpvstVlanConf, rpvstVlanIntfConf)
							// populate operational data
							app.convertOperInternalToOCVlanInterface(vlanName, intfId, rpvstVlanConf, rpvstVlanIntfConf)
						}
					}
				} else {
					isInterfacesSubtree := isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/rapid-pvst/vlan{}/interfaces")
					switch opcode {
					case CREATE, REPLACE, UPDATE, DELETE:
						err = app.handleRpvstCRUDOperationsAtVlanLevel(d, opcode, vlanName, isInterfacesSubtree, stp.RapidPvst, rpvstVlanConf)
						if err != nil {
							return err
						}
					case GET:
						err = app.convertDBRpvstVlanConfigToInternal(d, asKey(vlanName))
						if err != nil {
							return err
						}
						ygot.BuildEmptyTree(rpvstVlanConf)
						app.convertInternalToOCRpvstVlanConfig(vlanName, stp.RapidPvst, rpvstVlanConf)
					}
				}
			}
		} else {
			// Handle both rapid-pvst and rapid-pvst/vlan
			switch opcode {
			case CREATE, REPLACE, UPDATE, DELETE:
				log.Infof("Implementation in progress for URL: %s", app.pathInfo.Template)
				return tlerr.NotSupported("Implementation in progress")
			case GET:
				ygot.BuildEmptyTree(stp.RapidPvst)
				err = app.convertDBRpvstVlanConfigToInternal(d, db.Key{})
				if err != nil {
					return err
				}
				app.convertInternalToOCRpvstVlanConfig("", stp.RapidPvst, nil)
			}
		}
	} else if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/mstp") {
		mode, _ := app.getStpModeFromConfigDB(d)
		if mode != "mstp" {
			return tlerr.InvalidArgs("STP mode is configured as %s", mode)
		}
	} else if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/interfaces") {
		if isSubtreeRequest(app.pathInfo.Template, "/openconfig-spanning-tree:stp/interfaces/interface{}") {
			for intfId, _ := range stp.Interfaces.Interface {
				intfData := stp.Interfaces.Interface[intfId]
				switch opcode {
				case CREATE:
					err = app.setStpInterfacesDataInDB(d, true)
				case REPLACE:
					if *app.ygotTarget == intfData {
						err = d.DeleteEntry(app.interfaceTable, asKey(intfId))
						if err != nil {
							return err
						}
						err = app.setStpInterfacesDataInDB(d, true)
					} else {
						err = app.setStpInterfacesDataInDB(d, false)
					}
				case UPDATE:
					err = app.setStpInterfacesDataInDB(d, false)
				case DELETE:
					if *app.ygotTarget == intfData {
						err = d.DeleteEntry(app.interfaceTable, asKey(intfId))
					} else {
						err = app.handleInterfacesFieldsDeletion(d, intfId)
					}
				case GET:
					err = app.convertDBStpInterfacesToInternal(d, asKey(intfId))
					if err != nil {
						return err
					}
					ygot.BuildEmptyTree(intfData)
					app.convertInternalToOCStpInterfaces(intfId, stp.Interfaces, intfData)
				}
			}
		} else {
			switch opcode {
			case CREATE, REPLACE, UPDATE, DELETE:
				log.Infof("Implementation in progress for URL: %s", app.pathInfo.Template)
				return tlerr.NotSupported("Implementation in progress")
			case GET:
				ygot.BuildEmptyTree(stp.Interfaces)
				err = app.convertDBStpInterfacesToInternal(d, db.Key{})
				if err != nil {
					return err
				}
				app.convertInternalToOCStpInterfaces("", stp.Interfaces, nil)
			}
		}
	} else if topmostPath {
		switch opcode {
		case CREATE:
			err = app.enableStpMode(d)
			if err != nil {
				return err
			}
			err = app.setRpvstVlanDataInDB(d, true)
			if err != nil {
				return err
			}
			err = app.setRpvstVlanInterfaceDataInDB(d, true)
			if err != nil {
				return err
			}
			err = app.setStpInterfacesDataInDB(d, true)
		case REPLACE:
			err = app.disableStpMode(d)
			if err != nil {
				return err
			}
			err = app.enableStpMode(d)
			if err != nil {
				return err
			}
			err = app.setRpvstVlanDataInDB(d, true)
			if err != nil {
				return err
			}
			err = app.setRpvstVlanInterfaceDataInDB(d, true)
			if err != nil {
				return err
			}
			err = app.setStpInterfacesDataInDB(d, true)
		case DELETE:
			err = app.disableStpMode(d)
		case GET:
			ygot.BuildEmptyTree(stp)
			//////////////////////
			err = app.convertDBStpGlobalConfigToInternal(d)
			if err != nil {
				return err
			}
			app.convertInternalToOCStpGlobalConfig(stp.Global)

			//////////////////////
			err = app.convertDBRpvstVlanConfigToInternal(d, db.Key{})
			if err != nil {
				return err
			}
			stpMode := (&app.globalInfo).Get(STP_MODE)
			switch stpMode {
			case "pvst":
				ygot.BuildEmptyTree(stp.Pvst)
				app.convertInternalToOCPvstVlan("", stp.Pvst, nil)
			case "rpvst":
				ygot.BuildEmptyTree(stp.RapidPvst)
				app.convertInternalToOCRpvstVlanConfig("", stp.RapidPvst, nil)
			case "mstp":
			}
			//////////////////////
			ygot.BuildEmptyTree(stp.Interfaces)
			err = app.convertDBStpInterfacesToInternal(d, db.Key{})
			if err != nil {
				return err
			}
			app.convertInternalToOCStpInterfaces("", stp.Interfaces, nil)
		}
	}

	return err
}

func (app *StpApp) processCommonRpvstVlanToplevelPath(d *db.DB, stp *ocbinds.OpenconfigSpanningTree_Stp, opcode int) error {
	var err error

	switch opcode {
	case CREATE:
	case REPLACE:
	case UPDATE:
	case DELETE:
	case GET:
	}

	return err
}

func (app *StpApp) handleRpvstCRUDOperationsAtVlanLevel(d *db.DB, opcode int, vlanName string, isInterfacesSubtree bool, mode interface{}, vlan interface{}) error {
	var err error

	log.Infof("handleRpvstCRUDOperationsAtVlanLevel --> Perform CRUD for %s", reflect.TypeOf(vlan).Elem().Name())
	switch opcode {
	case CREATE:
		if *app.ygotTarget == vlan {
			err = app.setRpvstVlanDataInDB(d, true)
			if err != nil {
				return err
			}
			err = app.setRpvstVlanInterfaceDataInDB(d, true)
		} else if isInterfacesSubtree {
			err = app.setRpvstVlanInterfaceDataInDB(d, true)
		} else {
			err = d.SetEntry(app.vlanTable, asKey(vlanName), app.vlanTableMap[vlanName])
		}
	case REPLACE:
		if *app.ygotTarget == vlan {
			err = d.DeleteKeys(app.vlanIntfTable, asKey(vlanName+TABLE_SEPARATOR+"*"))
			if err != nil {
				return err
			}
			err = d.DeleteEntry(app.vlanTable, asKey(vlanName))
			if err != nil {
				return err
			}
			err = app.setRpvstVlanDataInDB(d, true)
			if err != nil {
				return err
			}
			err = app.setRpvstVlanInterfaceDataInDB(d, true)
		} else if isInterfacesSubtree {
			err = d.DeleteKeys(app.vlanIntfTable, asKey(vlanName+TABLE_SEPARATOR+"*"))
			if err != nil {
				return err
			}
			err = app.setRpvstVlanInterfaceDataInDB(d, true)
		} else {
			err = d.SetEntry(app.vlanTable, asKey(vlanName), app.vlanTableMap[vlanName])
		}
	case UPDATE:
		if *app.ygotTarget == vlan {
			err = app.setRpvstVlanDataInDB(d, false)
			if err != nil {
				return err
			}
			err = app.setRpvstVlanInterfaceDataInDB(d, false)
		} else if isInterfacesSubtree {
			err = app.setRpvstVlanInterfaceDataInDB(d, false)
		} else {
			err = d.ModEntry(app.vlanTable, asKey(vlanName), app.vlanTableMap[vlanName])
		}
	case DELETE:
		if *app.ygotTarget == vlan {
			err = d.DeleteKeys(app.vlanIntfTable, asKey(vlanName+TABLE_SEPARATOR+"*"))
			if err != nil {
				return err
			}
			err = d.DeleteEntry(app.vlanTable, asKey(vlanName))
		} else if isInterfacesSubtree {
			err = d.DeleteKeys(app.vlanIntfTable, asKey(vlanName+TABLE_SEPARATOR+"*"))
		} else {
			err = app.handleVlanFieldsDeletion(d, vlanName)
		}
	}

	return err
}

func (app *StpApp) handleRpvstCRUDOperationsAtVlanInterfaceLevel(d *db.DB, opcode int, vlanName string, intfId string, vlan interface{}, vlanIntf interface{}) error {
	var err error

	log.Infof("handleRpvstCRUDOperationsAtVlanInterfaceLevel --> Perform CRUD for %s", reflect.TypeOf(vlanIntf).Elem().Name())
	switch opcode {
	case CREATE:
		if *app.ygotTarget == vlanIntf {
			err = app.setRpvstVlanInterfaceDataInDB(d, true)
		} else {
			err = app.setRpvstVlanInterfaceDataInDB(d, false)
		}
	case REPLACE:
		if *app.ygotTarget == vlanIntf {
			err = d.DeleteEntry(app.vlanIntfTable, asKey(vlanName, intfId))
			if err != nil {
				return err
			}
			err = app.setRpvstVlanInterfaceDataInDB(d, true)
		} else {
			err = app.handleVlanInterfaceFieldsDeletion(d, vlanName, intfId)
			if err != nil {
				return err
			}
			err = app.setRpvstVlanInterfaceDataInDB(d, false)
		}
	case UPDATE:
		err = app.setRpvstVlanInterfaceDataInDB(d, false)
	case DELETE:
		if *app.ygotTarget == vlanIntf {
			err = d.DeleteEntry(app.vlanIntfTable, asKey(vlanName, intfId))
		} else {
			err = app.handleVlanInterfaceFieldsDeletion(d, vlanName, intfId)
		}
	}

	return err
}

/////////////////    STP GLOBAL   //////////////////////
func (app *StpApp) setStpGlobalConfigInDB(d *db.DB) error {
	var err error

	err = d.CreateEntry(app.globalTable, asKey("GLOBAL"), app.globalInfo)

	return err
}

func (app *StpApp) convertOCStpGlobalConfToInternal(opcode int) error {
	var err error
	stp := app.getAppRootObject()
	setDefaultFlag := (opcode == CREATE || opcode == REPLACE)
	if stp != nil {
		if stp.Global != nil && stp.Global.Config != nil {
			if stp.Global.Config.BridgePriority != nil {
				priorityVal := int(*stp.Global.Config.BridgePriority)
				if (priorityVal % 4096) != 0 {
					return tlerr.InvalidArgs("Priority value should be multiple of 4096")
				}
				(&app.globalInfo).Set("priority", strconv.Itoa(priorityVal))
			} else if setDefaultFlag {
				(&app.globalInfo).Set("priority", STP_DEFAULT_BRIDGE_PRIORITY)
			}
			if stp.Global.Config.ForwardingDelay != nil {
				(&app.globalInfo).Set("forward_delay", strconv.Itoa(int(*stp.Global.Config.ForwardingDelay)))
			} else if setDefaultFlag {
				(&app.globalInfo).Set("forward_delay", STP_DEFAULT_FORWARD_DELAY)
			}
			if stp.Global.Config.HelloTime != nil {
				(&app.globalInfo).Set("hello_time", strconv.Itoa(int(*stp.Global.Config.HelloTime)))
			} else if setDefaultFlag {
				(&app.globalInfo).Set("hello_time", STP_DEFAULT_HELLO_INTERVAL)
			}
			if stp.Global.Config.MaxAge != nil {
				(&app.globalInfo).Set("max_age", strconv.Itoa(int(*stp.Global.Config.MaxAge)))
			} else if setDefaultFlag {
				(&app.globalInfo).Set("max_age", STP_DEFAULT_MAX_AGE)
			}
			if stp.Global.Config.RootguardTimeout != nil {
				(&app.globalInfo).Set("rootguard_timeout", strconv.Itoa(int(*stp.Global.Config.RootguardTimeout)))
			} else if setDefaultFlag {
				(&app.globalInfo).Set("rootguard_timeout", STP_DEFAULT_ROOT_GUARD_TIMEOUT)
			}
			if stp.Global.Config.BpduFilter != nil {
				if *stp.Global.Config.BpduFilter == true {
					(&app.globalInfo).Set("bpdu_filter", "true")
				} else {
					(&app.globalInfo).Set("bpdu_filter", "false")
				}
			} else if setDefaultFlag {
				(&app.globalInfo).Set("bpdu_filter", STP_DEFAULT_BPDU_FILTER)
			}

			if len(stp.Global.Config.EnabledProtocol) > 0 {
				mode := app.convertOCStpModeToInternal(stp.Global.Config.EnabledProtocol[0])
				if len(mode) > 0 {
					(&app.globalInfo).Set(STP_MODE, mode)
				}
			}

			log.Infof("convertOCStpGlobalConfToInternal -- Internal Stp global config: %v", app.globalInfo)
		}
	}
	return err
}

func (app *StpApp) convertDBStpGlobalConfigToInternal(d *db.DB) error {
	var err error

	app.globalInfo, err = d.GetEntry(app.globalTable, asKey("GLOBAL"))
	if err != nil {
		return err
	}
	return err
}

func (app *StpApp) convertInternalToOCStpGlobalConfig(stpGlobal *ocbinds.OpenconfigSpanningTree_Stp_Global) {
	if stpGlobal != nil {
		var priority uint32
		var forDelay, helloTime, maxAge uint8
		var rootGTimeout uint16
		var bpduFilter bool
		ygot.BuildEmptyTree(stpGlobal)

		if stpGlobal.Config != nil {
			stpGlobal.Config.EnabledProtocol = app.convertInternalStpModeToOC((&app.globalInfo).Get(STP_MODE))

			var num uint64
			num, _ = strconv.ParseUint((&app.globalInfo).Get("priority"), 10, 32)
			priority = uint32(num)
			stpGlobal.Config.BridgePriority = &priority

			num, _ = strconv.ParseUint((&app.globalInfo).Get("forward_delay"), 10, 8)
			forDelay = uint8(num)
			stpGlobal.Config.ForwardingDelay = &forDelay

			num, _ = strconv.ParseUint((&app.globalInfo).Get("hello_time"), 10, 8)
			helloTime = uint8(num)
			stpGlobal.Config.HelloTime = &helloTime

			num, _ = strconv.ParseUint((&app.globalInfo).Get("max_age"), 10, 8)
			maxAge = uint8(num)
			stpGlobal.Config.MaxAge = &maxAge

			num, _ = strconv.ParseUint((&app.globalInfo).Get("rootguard_timeout"), 10, 16)
			rootGTimeout = uint16(num)
			stpGlobal.Config.RootguardTimeout = &rootGTimeout

			bpduFilter, _ = strconv.ParseBool((&app.globalInfo).Get("bpdu_filter"))
			stpGlobal.Config.BpduFilter = &bpduFilter
		}
		if stpGlobal.State != nil {
			stpGlobal.State.EnabledProtocol = app.convertInternalStpModeToOC((&app.globalInfo).Get(STP_MODE))
			stpGlobal.State.BridgePriority = &priority
			stpGlobal.State.ForwardingDelay = &forDelay
			stpGlobal.State.HelloTime = &helloTime
			stpGlobal.State.MaxAge = &maxAge
			stpGlobal.State.RootguardTimeout = &rootGTimeout
			stpGlobal.State.BpduFilter = &bpduFilter
		}
	}
}

/////////////////    RPVST //////////////////////
func (app *StpApp) convertOCRpvstConfToInternal(opcode int) error {
	var err error
	stp := app.getAppRootObject()
	setDefaultFlag := (opcode == CREATE || opcode == REPLACE)
	if stp != nil && stp.RapidPvst != nil && len(stp.RapidPvst.Vlan) > 0 {
		for vlanId, _ := range stp.RapidPvst.Vlan {
			vlanName := "Vlan" + strconv.Itoa(int(vlanId))
			app.vlanTableMap[vlanName] = db.Value{Field: map[string]string{}}
			rpvstVlanConf := stp.RapidPvst.Vlan[vlanId]
			if rpvstVlanConf.Config != nil {
				dbVal := app.vlanTableMap[vlanName]
				(&dbVal).Set("vlanid", strconv.Itoa(int(vlanId)))
				if rpvstVlanConf.Config.BridgePriority != nil {
					priorityVal := int(*rpvstVlanConf.Config.BridgePriority)
					if (priorityVal % 4096) != 0 {
						return tlerr.InvalidArgs("Priority value should be multiple of 4096")
					}
					(&dbVal).Set("priority", strconv.Itoa(priorityVal))
				} else if setDefaultFlag {
					(&dbVal).Set("priority", "32768")
				}
				if rpvstVlanConf.Config.ForwardingDelay != nil {
					(&dbVal).Set("forward_delay", strconv.Itoa(int(*rpvstVlanConf.Config.ForwardingDelay)))
				} else if setDefaultFlag {
					(&dbVal).Set("forward_delay", "15")
				}
				if rpvstVlanConf.Config.HelloTime != nil {
					(&dbVal).Set("hello_time", strconv.Itoa(int(*rpvstVlanConf.Config.HelloTime)))
				} else if setDefaultFlag {
					(&dbVal).Set("hello_time", "2")
				}
				if rpvstVlanConf.Config.MaxAge != nil {
					(&dbVal).Set("max_age", strconv.Itoa(int(*rpvstVlanConf.Config.MaxAge)))
				} else if setDefaultFlag {
					(&dbVal).Set("max_age", "20")
				}
				if rpvstVlanConf.Config.SpanningTreeEnable != nil {
					if *rpvstVlanConf.Config.SpanningTreeEnable {
						(&dbVal).Set("enabled", "true")
					} else {
						(&dbVal).Set("enabled", "false")
					}
				} else if setDefaultFlag {
					(&dbVal).Set("enabled", "false")
				}
			}
			if rpvstVlanConf.Interfaces != nil && len(rpvstVlanConf.Interfaces.Interface) > 0 {
				app.vlanIntfTableMap[vlanName] = make(map[string]db.Value)
				for intfId, _ := range rpvstVlanConf.Interfaces.Interface {
					rpvstVlanIntfConf := rpvstVlanConf.Interfaces.Interface[intfId]
					app.vlanIntfTableMap[vlanName][intfId] = db.Value{Field: map[string]string{}}
					if rpvstVlanIntfConf.Config != nil {
						dbVal := app.vlanIntfTableMap[vlanName][intfId]
						if rpvstVlanIntfConf.Config.Cost != nil {
							(&dbVal).Set("path_cost", strconv.Itoa(int(*rpvstVlanIntfConf.Config.Cost)))
						} else if setDefaultFlag {
							(&dbVal).Set("path_cost", "200")
						}
						if rpvstVlanIntfConf.Config.PortPriority != nil {
							(&dbVal).Set("priority", strconv.Itoa(int(*rpvstVlanIntfConf.Config.PortPriority)))
						} else if setDefaultFlag {
							(&dbVal).Set("priority", "128")
						}
					}
				}
			}
		}
	}
	return err
}

func (app *StpApp) setRpvstVlanDataInDB(d *db.DB, createFlag bool) error {
	var err error
	for vlanName := range app.vlanTableMap {
		existingEntry, err := d.GetEntry(app.vlanTable, asKey(vlanName))
		if createFlag && existingEntry.IsPopulated() {
			return tlerr.AlreadyExists("Vlan %s already configured", vlanName)
		}
		if createFlag || (!createFlag && err != nil && !existingEntry.IsPopulated()) {
			err = d.CreateEntry(app.vlanTable, asKey(vlanName), app.vlanTableMap[vlanName])
		} else {
			if existingEntry.IsPopulated() {
				err = d.ModEntry(app.vlanTable, asKey(vlanName), app.vlanTableMap[vlanName])
			}
		}
	}
	return err
}

func (app *StpApp) setRpvstVlanInterfaceDataInDB(d *db.DB, createFlag bool) error {
	var err error
	for vlanName := range app.vlanIntfTableMap {
		for intfId := range app.vlanIntfTableMap[vlanName] {
			existingEntry, err := d.GetEntry(app.vlanIntfTable, asKey(vlanName, intfId))
			if createFlag && existingEntry.IsPopulated() {
				return tlerr.AlreadyExists("Interface %s already configured", intfId)
			}
			if createFlag || (!createFlag && err != nil && !existingEntry.IsPopulated()) {
				err = d.CreateEntry(app.vlanIntfTable, asKey(vlanName, intfId), app.vlanIntfTableMap[vlanName][intfId])
			} else {
				if existingEntry.IsPopulated() {
					err = d.ModEntry(app.vlanIntfTable, asKey(vlanName, intfId), app.vlanIntfTableMap[vlanName][intfId])
				}
			}
		}
	}
	return err
}

func (app *StpApp) convertDBRpvstVlanConfigToInternal(d *db.DB, vlanKey db.Key) error {
	var err error
	if vlanKey.Len() > 0 {
		entry, err := d.GetEntry(app.vlanTable, vlanKey)
		if err != nil {
			return err
		}
		vlanName := vlanKey.Get(0)
		if entry.IsPopulated() {
			app.vlanTableMap[vlanName] = entry
			app.vlanIntfTableMap[vlanName] = make(map[string]db.Value)
			err = app.convertDBRpvstVlanInterfaceToInternal(d, vlanName, "", db.Key{}, false)
			if err != nil {
				return err
			}
			// Collect operational info from application DB
			err = app.convertApplDBRpvstVlanToInternal(vlanName)
		} else {
			return tlerr.NotFound("Vlan %s is not configured", vlanName)
		}
	} else {
		tbl, err := d.GetTable(app.vlanTable)
		if err != nil {
			return err
		}
		keys, err := tbl.GetKeys()
		if err != nil {
			return err
		}
		for i, _ := range keys {
			app.convertDBRpvstVlanConfigToInternal(d, keys[i])
		}
	}

	return err
}

func (app *StpApp) convertInternalToOCRpvstVlanConfig(vlanName string, rpvst *ocbinds.OpenconfigSpanningTree_Stp_RapidPvst, rpvstVlanConf *ocbinds.OpenconfigSpanningTree_Stp_RapidPvst_Vlan) {
	if len(vlanName) > 0 {
		var num uint64
		if rpvstVlanConf != nil {
			if rpvstVlanData, ok := app.vlanTableMap[vlanName]; ok {
				vlanId, _ := strconv.Atoi(strings.Replace(vlanName, "Vlan", "", 1))
				vlan := uint16(vlanId)
				rpvstVlanConf.VlanId = &vlan
				rpvstVlanConf.Config.VlanId = &vlan
				rpvstVlanConf.State.VlanId = &vlan

				stpEnabled, _ := strconv.ParseBool((&rpvstVlanData).Get("enabled"))
				rpvstVlanConf.Config.SpanningTreeEnable = &stpEnabled
				//rpvstVlanConf.State.SpanningTreeEnable = &stpEnabled

				num, _ = strconv.ParseUint((&rpvstVlanData).Get("priority"), 10, 32)
				priority := uint32(num)
				rpvstVlanConf.Config.BridgePriority = &priority
				rpvstVlanConf.State.BridgePriority = &priority

				num, _ = strconv.ParseUint((&rpvstVlanData).Get("forward_delay"), 10, 8)
				forDelay := uint8(num)
				rpvstVlanConf.Config.ForwardingDelay = &forDelay

				num, _ = strconv.ParseUint((&rpvstVlanData).Get("hello_time"), 10, 8)
				helloTime := uint8(num)
				rpvstVlanConf.Config.HelloTime = &helloTime

				num, _ = strconv.ParseUint((&rpvstVlanData).Get("max_age"), 10, 8)
				maxAge := uint8(num)
				rpvstVlanConf.Config.MaxAge = &maxAge
			}

			// populate operational information
			//ygot.BuildEmptyTree(rpvstVlanConf.State)
			operDbVal := app.vlanOperTableMap[vlanName]
			if operDbVal.IsPopulated() {
				num, _ = strconv.ParseUint((&operDbVal).Get("max_age"), 10, 8)
				opMaxAge := uint8(num)
				rpvstVlanConf.State.MaxAge = &opMaxAge

				num, _ = strconv.ParseUint((&operDbVal).Get("hello_time"), 10, 8)
				opHelloTime := uint8(num)
				rpvstVlanConf.State.HelloTime = &opHelloTime

				num, _ = strconv.ParseUint((&operDbVal).Get("forward_delay"), 10, 8)
				opForwardDelay := uint8(num)
				rpvstVlanConf.State.ForwardingDelay = &opForwardDelay

				num, _ = strconv.ParseUint((&operDbVal).Get("hold_time"), 10, 8)
				opHoldTime := uint8(num)
				rpvstVlanConf.State.HoldTime = &opHoldTime

				/*num, _ = strconv.ParseUint((&operDbVal).Get("root_max_age"), 10, 8)
				opRootMaxAge := uint8(num)
				rpvstVlanConf.State.RootMaxAge = &opRootMaxAge

				num, _ = strconv.ParseUint((&operDbVal).Get("root_hello_time"), 10, 8)
				opRootHelloTime := uint8(num)
				rpvstVlanConf.State.RootHelloTime = &opRootHelloTime

				num, _ = strconv.ParseUint((&operDbVal).Get("root_forward_delay"), 10, 8)
				opRootForwardDelay := uint8(num)
				rpvstVlanConf.State.RootForwardingDelay = &opRootForwardDelay  */

				num, _ = strconv.ParseUint((&operDbVal).Get("stp_instance"), 10, 16)
				opStpInstance := uint16(num)
				rpvstVlanConf.State.StpInstance = &opStpInstance

				num, _ = strconv.ParseUint((&operDbVal).Get("root_path_cost"), 10, 32)
				opRootCost := uint32(num)
				rpvstVlanConf.State.RootCost = &opRootCost

				num, _ = strconv.ParseUint((&operDbVal).Get("last_topology_change"), 10, 64)
				opLastTopologyChange := num
				rpvstVlanConf.State.LastTopologyChange = &opLastTopologyChange

				num, _ = strconv.ParseUint((&operDbVal).Get("topology_change_count"), 10, 64)
				opTopologyChanges := num
				rpvstVlanConf.State.TopologyChanges = &opTopologyChanges

				bridgeId := (&operDbVal).Get("bridge_id")
				rpvstVlanConf.State.BridgeAddress = &bridgeId

				desigRootAddr := (&operDbVal).Get("desig_bridge_id")
				rpvstVlanConf.State.DesignatedRootAddress = &desigRootAddr

				rootPortStr := (&operDbVal).Get("root_port")
				rpvstVlanConf.State.RootPortName = &rootPortStr
			}

			app.convertInternalToOCRpvstVlanInterface(vlanName, "", rpvstVlanConf, nil)
			// populate operational information
			app.convertOperInternalToOCVlanInterface(vlanName, "", rpvstVlanConf, nil)
		}
	} else {
		for vlanName := range app.vlanTableMap {
			vlanId, _ := strconv.Atoi(strings.Replace(vlanName, "Vlan", "", 1))
			vlan := uint16(vlanId)

			rpvstVlanConfPtr, _ := rpvst.NewVlan(vlan)
			ygot.BuildEmptyTree(rpvstVlanConfPtr)
			app.convertInternalToOCRpvstVlanConfig(vlanName, rpvst, rpvstVlanConfPtr)
		}
	}
}

func (app *StpApp) convertDBRpvstVlanInterfaceToInternal(d *db.DB, vlanName string, intfId string, vlanInterfaceKey db.Key, doGetOperData bool) error {
	var err error
	if vlanInterfaceKey.Len() > 1 {
		rpvstVlanIntfConf, err := d.GetEntry(app.vlanIntfTable, asKey(vlanName, intfId))
		if app.vlanIntfTableMap[vlanName] == nil {
			app.vlanIntfTableMap[vlanName] = make(map[string]db.Value)
		}
		if err == nil {
			app.vlanIntfTableMap[vlanName][intfId] = rpvstVlanIntfConf
		}
		// Collect operational info from application DB
		if doGetOperData {
			err = app.convertApplDBRpvstVlanInterfaceToInternal(vlanName, intfId)
		}
		if err != nil {
			return err
		}
	} else {
		keys, err := d.GetKeys(app.vlanIntfTable)
		if err != nil {
			return err
		}
		for i, _ := range keys {
			if vlanName == keys[i].Get(0) {
				err = app.convertDBRpvstVlanInterfaceToInternal(d, vlanName, keys[i].Get(1), keys[i], doGetOperData)
			}
		}
	}
	return err
}

func (app *StpApp) convertInternalToOCRpvstVlanInterface(vlanName string, intfId string, rpvstVlanConf *ocbinds.OpenconfigSpanningTree_Stp_RapidPvst_Vlan, rpvstVlanIntfConf *ocbinds.OpenconfigSpanningTree_Stp_RapidPvst_Vlan_Interfaces_Interface) {
	var num uint64

	if len(intfId) == 0 {
		for intf, _ := range app.vlanIntfTableMap[vlanName] {
			app.convertInternalToOCRpvstVlanInterface(vlanName, intf, rpvstVlanConf, rpvstVlanIntfConf)
		}
	} else {
		dbVal := app.vlanIntfTableMap[vlanName][intfId]

		if rpvstVlanIntfConf == nil {
			if rpvstVlanConf != nil {
				rpvstVlanIntfConf_, _ := rpvstVlanConf.Interfaces.NewInterface(intfId)
				rpvstVlanIntfConf = rpvstVlanIntfConf_
				ygot.BuildEmptyTree(rpvstVlanIntfConf)
			}
		}

		var err error
		num, err = strconv.ParseUint((&dbVal).Get("path_cost"), 10, 32)
		if err == nil {
			cost := uint32(num)
			rpvstVlanIntfConf.Config.Cost = &cost
		}

		num, err = strconv.ParseUint((&dbVal).Get("priority"), 10, 8)
		if err == nil {
			portPriority := uint8(num)
			rpvstVlanIntfConf.Config.PortPriority = &portPriority
		}

		rpvstVlanIntfConf.Config.Name = &intfId
	}
}

///////////   PVST   //////////////////////
func (app *StpApp) convertOCPvstToInternal(opcode int) error {
	var err error
	stp := app.getAppRootObject()
	setDefaultFlag := (opcode == CREATE || opcode == REPLACE)
	if stp != nil && stp.Pvst != nil && len(stp.Pvst.Vlan) > 0 {
		for vlanId, _ := range stp.Pvst.Vlan {
			vlanName := "Vlan" + strconv.Itoa(int(vlanId))
			app.vlanTableMap[vlanName] = db.Value{Field: map[string]string{}}
			pvstVlan := stp.Pvst.Vlan[vlanId]
			if pvstVlan.Config != nil {
				dbVal := app.vlanTableMap[vlanName]
				(&dbVal).Set("vlanid", strconv.Itoa(int(vlanId)))
				if pvstVlan.Config.BridgePriority != nil {
					priorityVal := int(*pvstVlan.Config.BridgePriority)
					if (priorityVal % 4096) != 0 {
						return tlerr.InvalidArgs("Priority value should be multiple of 4096")
					}
					(&dbVal).Set("priority", strconv.Itoa(priorityVal))
				} else if setDefaultFlag {
					(&dbVal).Set("priority", "32768")
				}
				if pvstVlan.Config.ForwardingDelay != nil {
					(&dbVal).Set("forward_delay", strconv.Itoa(int(*pvstVlan.Config.ForwardingDelay)))
				} else if setDefaultFlag {
					(&dbVal).Set("forward_delay", "15")
				}
				if pvstVlan.Config.HelloTime != nil {
					(&dbVal).Set("hello_time", strconv.Itoa(int(*pvstVlan.Config.HelloTime)))
				} else if setDefaultFlag {
					(&dbVal).Set("hello_time", "2")
				}
				if pvstVlan.Config.MaxAge != nil {
					(&dbVal).Set("max_age", strconv.Itoa(int(*pvstVlan.Config.MaxAge)))
				} else if setDefaultFlag {
					(&dbVal).Set("max_age", "20")
				}
				if pvstVlan.Config.SpanningTreeEnable != nil {
					if *pvstVlan.Config.SpanningTreeEnable {
						(&dbVal).Set("enabled", "true")
					} else {
						(&dbVal).Set("enabled", "false")
					}
				} else if setDefaultFlag {
					(&dbVal).Set("enabled", "false")
				}
			}
			if pvstVlan.Interfaces != nil && len(pvstVlan.Interfaces.Interface) > 0 {
				app.vlanIntfTableMap[vlanName] = make(map[string]db.Value)
				for intfId, _ := range pvstVlan.Interfaces.Interface {
					pvstVlanIntf := pvstVlan.Interfaces.Interface[intfId]
					app.vlanIntfTableMap[vlanName][intfId] = db.Value{Field: map[string]string{}}
					if pvstVlanIntf.Config != nil {
						dbVal := app.vlanIntfTableMap[vlanName][intfId]
						if pvstVlanIntf.Config.Cost != nil {
							(&dbVal).Set("path_cost", strconv.Itoa(int(*pvstVlanIntf.Config.Cost)))
						} else if setDefaultFlag {
							(&dbVal).Set("path_cost", "200")
						}
						if pvstVlanIntf.Config.PortPriority != nil {
							(&dbVal).Set("priority", strconv.Itoa(int(*pvstVlanIntf.Config.PortPriority)))
						} else if setDefaultFlag {
							(&dbVal).Set("priority", "128")
						}
					}
				}
			}
		}
	}
	return err
}

func (app *StpApp) convertInternalToOCPvstVlan(vlanName string, pvst *ocbinds.OpenconfigSpanningTree_Stp_Pvst, pvstVlan *ocbinds.OpenconfigSpanningTree_Stp_Pvst_Vlan) {
	if len(vlanName) > 0 {
		var num uint64
		if pvstVlan != nil {
			if pvstVlanData, ok := app.vlanTableMap[vlanName]; ok {
				vlanId, _ := strconv.Atoi(strings.Replace(vlanName, "Vlan", "", 1))
				vlan := uint16(vlanId)
				pvstVlan.VlanId = &vlan
				pvstVlan.Config.VlanId = &vlan
				pvstVlan.State.VlanId = &vlan

				stpEnabled, _ := strconv.ParseBool((&pvstVlanData).Get("enabled"))
				pvstVlan.Config.SpanningTreeEnable = &stpEnabled
				//pvstVlan.State.SpanningTreeEnable = &stpEnabled

				num, _ = strconv.ParseUint((&pvstVlanData).Get("priority"), 10, 32)
				priority := uint32(num)
				pvstVlan.Config.BridgePriority = &priority
				pvstVlan.State.BridgePriority = &priority

				num, _ = strconv.ParseUint((&pvstVlanData).Get("forward_delay"), 10, 8)
				forDelay := uint8(num)
				pvstVlan.Config.ForwardingDelay = &forDelay

				num, _ = strconv.ParseUint((&pvstVlanData).Get("hello_time"), 10, 8)
				helloTime := uint8(num)
				pvstVlan.Config.HelloTime = &helloTime

				num, _ = strconv.ParseUint((&pvstVlanData).Get("max_age"), 10, 8)
				maxAge := uint8(num)
				pvstVlan.Config.MaxAge = &maxAge
			}

			// populate operational information
			operDbVal := app.vlanOperTableMap[vlanName]
			if operDbVal.IsPopulated() {
				num, _ = strconv.ParseUint((&operDbVal).Get("max_age"), 10, 8)
				opMaxAge := uint8(num)
				pvstVlan.State.MaxAge = &opMaxAge

				num, _ = strconv.ParseUint((&operDbVal).Get("hello_time"), 10, 8)
				opHelloTime := uint8(num)
				pvstVlan.State.HelloTime = &opHelloTime

				num, _ = strconv.ParseUint((&operDbVal).Get("forward_delay"), 10, 8)
				opForwardDelay := uint8(num)
				pvstVlan.State.ForwardingDelay = &opForwardDelay

				num, _ = strconv.ParseUint((&operDbVal).Get("hold_time"), 10, 8)
				opHoldTime := uint8(num)
				pvstVlan.State.HoldTime = &opHoldTime

				/*num, _ = strconv.ParseUint((&operDbVal).Get("root_max_age"), 10, 8)
				opRootMaxAge := uint8(num)
				pvstVlan.State.RootMaxAge = &opRootMaxAge

				num, _ = strconv.ParseUint((&operDbVal).Get("root_hello_time"), 10, 8)
				opRootHelloTime := uint8(num)
				pvstVlan.State.RootHelloTime = &opRootHelloTime

				num, _ = strconv.ParseUint((&operDbVal).Get("root_forward_delay"), 10, 8)
				opRootForwardDelay := uint8(num)
				pvstVlan.State.RootForwardingDelay = &opRootForwardDelay  */

				num, _ = strconv.ParseUint((&operDbVal).Get("stp_instance"), 10, 16)
				opStpInstance := uint16(num)
				pvstVlan.State.StpInstance = &opStpInstance

				num, _ = strconv.ParseUint((&operDbVal).Get("root_path_cost"), 10, 32)
				opRootCost := uint32(num)
				pvstVlan.State.RootCost = &opRootCost

				num, _ = strconv.ParseUint((&operDbVal).Get("last_topology_change"), 10, 64)
				opLastTopologyChange := num
				pvstVlan.State.LastTopologyChange = &opLastTopologyChange

				num, _ = strconv.ParseUint((&operDbVal).Get("topology_change_count"), 10, 64)
				opTopologyChanges := num
				pvstVlan.State.TopologyChanges = &opTopologyChanges

				bridgeId := (&operDbVal).Get("bridge_id")
				pvstVlan.State.BridgeAddress = &bridgeId

				desigRootAddr := (&operDbVal).Get("desig_bridge_id")
				pvstVlan.State.DesignatedRootAddress = &desigRootAddr

				rootPortStr := (&operDbVal).Get("root_port")
				pvstVlan.State.RootPortName = &rootPortStr
			}

			app.convertInternalToOCPvstVlanInterface(vlanName, "", pvstVlan, nil)
			// populate operational information
			app.convertOperInternalToOCVlanInterface(vlanName, "", pvstVlan, nil)
		}
	} else {
		for vlanName := range app.vlanTableMap {
			vlanId, _ := strconv.Atoi(strings.Replace(vlanName, "Vlan", "", 1))
			vlan := uint16(vlanId)

			pvstVlanPtr, _ := pvst.NewVlan(vlan)
			ygot.BuildEmptyTree(pvstVlanPtr)
			app.convertInternalToOCPvstVlan(vlanName, pvst, pvstVlanPtr)
		}
	}
}

func (app *StpApp) convertInternalToOCPvstVlanInterface(vlanName string, intfId string, pvstVlan *ocbinds.OpenconfigSpanningTree_Stp_Pvst_Vlan, pvstVlanIntf *ocbinds.OpenconfigSpanningTree_Stp_Pvst_Vlan_Interfaces_Interface) {
	var num uint64

	if len(intfId) == 0 {
		for intf, _ := range app.vlanIntfTableMap[vlanName] {
			app.convertInternalToOCPvstVlanInterface(vlanName, intf, pvstVlan, pvstVlanIntf)
		}
	} else {
		dbVal := app.vlanIntfTableMap[vlanName][intfId]

		if pvstVlanIntf == nil {
			if pvstVlan != nil {
				pvstVlanIntf_, _ := pvstVlan.Interfaces.NewInterface(intfId)
				pvstVlanIntf = pvstVlanIntf_
				ygot.BuildEmptyTree(pvstVlanIntf)
			}
		}

		var err error
		num, err = strconv.ParseUint((&dbVal).Get("path_cost"), 10, 32)
		if err == nil {
			cost := uint32(num)
			pvstVlanIntf.Config.Cost = &cost
		}

		num, err = strconv.ParseUint((&dbVal).Get("priority"), 10, 8)
		if err == nil {
			portPriority := uint8(num)
			pvstVlanIntf.Config.PortPriority = &portPriority
		}

		pvstVlanIntf.Config.Name = &intfId
	}
}

///////////  Interfaces   //////////
func (app *StpApp) convertOCStpInterfacesToInternal() {
	stp := app.getAppRootObject()
	if stp != nil && stp.Interfaces != nil && len(stp.Interfaces.Interface) > 0 {
		for intfId, _ := range stp.Interfaces.Interface {
			app.intfTableMap[intfId] = db.Value{Field: map[string]string{}}

			stpIntfConf := stp.Interfaces.Interface[intfId]
			if stpIntfConf.Config != nil {
				dbVal := app.intfTableMap[intfId]

				if stpIntfConf.Config.BpduGuard != nil {
					if *stpIntfConf.Config.BpduGuard == true {
						(&dbVal).Set("bpdu_guard", "true")
					} else {
						(&dbVal).Set("bpdu_guard", "false")
					}
				}

				if stpIntfConf.Config.BpduFilter != nil {
					if *stpIntfConf.Config.BpduFilter == true {
						(&dbVal).Set("bpdu_filter", "enable")
					} else {
						(&dbVal).Set("bpdu_filter", "disable")
					}
				} else {
					(&dbVal).Set("bpdu_filter", "global")
				}

				if stpIntfConf.Config.BpduGuardPortShutdown != nil {
					if *stpIntfConf.Config.BpduGuardPortShutdown == true {
						(&dbVal).Set("bpdu_guard_do_disable", "true")
					} else {
						(&dbVal).Set("bpdu_guard_do_disable", "false")
					}
				}

				if stpIntfConf.Config.Portfast != nil {
					if *stpIntfConf.Config.Portfast == true {
						(&dbVal).Set("portfast", "true")
					} else {
						(&dbVal).Set("portfast", "false")
					}
				}

				if stpIntfConf.Config.UplinkFast != nil {
					if *stpIntfConf.Config.UplinkFast == true {
						(&dbVal).Set("uplink_fast", "true")
					} else {
						(&dbVal).Set("uplink_fast", "false")
					}
				}

				if stpIntfConf.Config.SpanningTreeEnable != nil {
					if *stpIntfConf.Config.SpanningTreeEnable == true {
						(&dbVal).Set("enabled", "true")
					} else {
						(&dbVal).Set("enabled", "false")
					}
				}

				if stpIntfConf.Config.Cost != nil {
					(&dbVal).Set("path_cost", strconv.Itoa(int(*stpIntfConf.Config.Cost)))
				} else {
					//(&dbVal).Set("path_cost", "200")
				}
				if stpIntfConf.Config.PortPriority != nil {
					(&dbVal).Set("priority", strconv.Itoa(int(*stpIntfConf.Config.PortPriority)))
				} else {
					//(&dbVal).Set("priority", "128")
				}

				if stpIntfConf.Config.Guard == ocbinds.OpenconfigSpanningTree_StpGuardType_ROOT {
					(&dbVal).Set("root_guard", "true")
				} else if stpIntfConf.Config.Guard == ocbinds.OpenconfigSpanningTree_StpGuardType_NONE {
					(&dbVal).Set("root_guard", "false")
				}
				////   For RPVST+   /////
				if stpIntfConf.Config.EdgePort == ocbinds.OpenconfigSpanningTreeTypes_STP_EDGE_PORT_EDGE_ENABLE {
					(&dbVal).Set("edge_port", "true")
				} else if stpIntfConf.Config.EdgePort == ocbinds.OpenconfigSpanningTreeTypes_STP_EDGE_PORT_EDGE_DISABLE {
					(&dbVal).Set("edge_port", "false")
				}

				if stpIntfConf.Config.LinkType == ocbinds.OpenconfigSpanningTree_StpLinkType_P2P {
					(&dbVal).Set("link_type", "point-to-point")
				} else if stpIntfConf.Config.LinkType == ocbinds.OpenconfigSpanningTree_StpLinkType_SHARED {
					(&dbVal).Set("link_type", "shared")
				} else {
					(&dbVal).Set("link_type", "auto")
				}
			}
		}
	}
}

func (app *StpApp) setStpInterfacesDataInDB(d *db.DB, createFlag bool) error {
	var err error
	// Fetch list of interfaces which are L2 i.e. members of any Vlan
	l2IntfList, err1 := app.getAllInterfacesFromVlanMemberTable(d)
	if err1 != nil {
		return tlerr.InvalidArgs("There are no L2 interfaces configured")
	}
	// Fetch list of interfaces which are members of port-channel
	poMemberIntfList, _ := app.getAllInterfacesFromPortChannelMemberTable(d)

	for intfName := range app.intfTableMap {
		if !contains(l2IntfList, intfName) {
			return tlerr.InvalidArgs("%s has no VLAN configured - It's not a L2 interface", intfName)
		}
		if contains(poMemberIntfList, intfName) {
			return tlerr.InvalidArgs("%s is a portchannel member port - STP can't be configured", intfName)
		}
		existingEntry, err := d.GetEntry(app.interfaceTable, asKey(intfName))
		if createFlag && existingEntry.IsPopulated() {
			return tlerr.AlreadyExists("Stp Interface %s already configured", intfName)
		}
		if createFlag || (!createFlag && err != nil && !existingEntry.IsPopulated()) {
			err = d.CreateEntry(app.interfaceTable, asKey(intfName), app.intfTableMap[intfName])
		} else {
			if existingEntry.IsPopulated() {
				err = d.ModEntry(app.interfaceTable, asKey(intfName), app.intfTableMap[intfName])
			}
		}
	}
	return err
}

func (app *StpApp) convertDBStpInterfacesToInternal(d *db.DB, intfKey db.Key) error {
	var err error
	if intfKey.Len() > 0 {
		entry, err := d.GetEntry(app.interfaceTable, intfKey)
		if err != nil {
			return err
		}
		intfName := intfKey.Get(0)
		if entry.IsPopulated() {
			app.intfTableMap[intfName] = entry
		} else {
			return tlerr.NotFound("STP interface %s is not configured", intfName)
		}
		err = app.convertApplDBStpInterfacesToInternal(intfName)
	} else {
		tbl, err := d.GetTable(app.interfaceTable)
		if err != nil {
			return err
		}
		keys, err := tbl.GetKeys()
		if err != nil {
			return err
		}
		for i, _ := range keys {
			app.convertDBStpInterfacesToInternal(d, keys[i])
		}
	}

	return err
}

func (app *StpApp) convertInternalToOCStpInterfaces(intfName string, interfaces *ocbinds.OpenconfigSpanningTree_Stp_Interfaces, intf *ocbinds.OpenconfigSpanningTree_Stp_Interfaces_Interface) {
	var err error
	if len(intfName) > 0 {
		if stpIntfData, ok := app.intfTableMap[intfName]; ok {
			if intf != nil {
				intf.Config.Name = &intfName
				intf.State.Name = &intfName

				stpEnabled, _ := strconv.ParseBool((&stpIntfData).Get("enabled"))
				intf.Config.SpanningTreeEnable = &stpEnabled
				intf.State.SpanningTreeEnable = &stpEnabled

				bpduGuardEnabled, _ := strconv.ParseBool((&stpIntfData).Get("bpdu_guard"))
				intf.Config.BpduGuard = &bpduGuardEnabled
				intf.State.BpduGuard = &bpduGuardEnabled

				var bpduFilterEnabled bool
				bpduFilterVal := (&stpIntfData).Get("bpdu_filter")
				if bpduFilterVal == "enable" {
					bpduFilterEnabled = true
					intf.Config.BpduFilter = &bpduFilterEnabled
					intf.State.BpduFilter = &bpduFilterEnabled
				} else if bpduFilterVal == "disable" {
					bpduFilterEnabled = false
					intf.Config.BpduFilter = &bpduFilterEnabled
					intf.State.BpduFilter = &bpduFilterEnabled
				}

				bpduGuardPortShut, _ := strconv.ParseBool((&stpIntfData).Get("bpdu_guard_do_disable"))
				intf.Config.BpduGuardPortShutdown = &bpduGuardPortShut
				intf.State.BpduGuardPortShutdown = &bpduGuardPortShut

				uplinkFast, _ := strconv.ParseBool((&stpIntfData).Get("uplink_fast"))
				intf.Config.UplinkFast = &uplinkFast
				intf.State.UplinkFast = &uplinkFast

				portFast, _ := strconv.ParseBool((&stpIntfData).Get("portfast"))
				intf.Config.Portfast = &portFast

				rootGuardEnabled, _ := strconv.ParseBool((&stpIntfData).Get("root_guard"))
				if rootGuardEnabled {
					intf.Config.Guard = ocbinds.OpenconfigSpanningTree_StpGuardType_ROOT
					intf.State.Guard = ocbinds.OpenconfigSpanningTree_StpGuardType_ROOT
				} else {
					intf.Config.Guard = ocbinds.OpenconfigSpanningTree_StpGuardType_NONE
					intf.State.Guard = ocbinds.OpenconfigSpanningTree_StpGuardType_NONE
				}

				if edgePortEnabled, err := strconv.ParseBool((&stpIntfData).Get("edge_port")); err == nil {
					if edgePortEnabled {
						intf.Config.EdgePort = ocbinds.OpenconfigSpanningTreeTypes_STP_EDGE_PORT_EDGE_ENABLE
						intf.State.EdgePort = ocbinds.OpenconfigSpanningTreeTypes_STP_EDGE_PORT_EDGE_ENABLE
					} else {
						intf.Config.EdgePort = ocbinds.OpenconfigSpanningTreeTypes_STP_EDGE_PORT_EDGE_DISABLE
						intf.State.EdgePort = ocbinds.OpenconfigSpanningTreeTypes_STP_EDGE_PORT_EDGE_DISABLE
					}
				}

				linkTypeVal := (&stpIntfData).Get("link_type")
				switch linkTypeVal {
				case "shared":
					intf.Config.LinkType = ocbinds.OpenconfigSpanningTree_StpLinkType_SHARED
					intf.State.LinkType = ocbinds.OpenconfigSpanningTree_StpLinkType_SHARED
				case "point-to-point":
					intf.Config.LinkType = ocbinds.OpenconfigSpanningTree_StpLinkType_P2P
					intf.State.LinkType = ocbinds.OpenconfigSpanningTree_StpLinkType_P2P
				}

				var num uint64
				if num, err = strconv.ParseUint((&stpIntfData).Get("priority"), 10, 8); err == nil {
					priority := uint8(num)
					intf.Config.PortPriority = &priority
					intf.State.PortPriority = &priority
				}

				if num, err = strconv.ParseUint((&stpIntfData).Get("path_cost"), 10, 32); err == nil {
					cost := uint32(num)
					intf.Config.Cost = &cost
					intf.State.Cost = &cost
				}
			}
		}

		operDbVal := app.intfOperTableMap[intfName]
		if operDbVal.IsPopulated() && intf != nil {
			var boolVal bool

			bpduGuardShut := (&operDbVal).Get("bpdu_guard_shutdown")
			if bpduGuardShut == "yes" {
				boolVal = true
			} else if bpduGuardShut == "no" {
				boolVal = false
			}
			intf.State.BpduGuardShutdown = &boolVal

			opPortfast := (&operDbVal).Get("port_fast")
			if opPortfast == "yes" {
				boolVal = true
			} else if opPortfast == "no" {
				boolVal = false
			}
			intf.State.Portfast = &boolVal

			opBpduFilter := (&operDbVal).Get("bpdu_filter")
			if opBpduFilter == "yes" {
				boolVal = true
			} else if opBpduFilter == "no" {
				boolVal = false
			}
			intf.State.BpduFilter = &boolVal

			opEdgePortType := (&operDbVal).Get("edge_port")
			if opEdgePortType == "yes" {
				intf.State.EdgePort = ocbinds.OpenconfigSpanningTreeTypes_STP_EDGE_PORT_EDGE_ENABLE
			} else if opEdgePortType == "no" {
				intf.State.EdgePort = ocbinds.OpenconfigSpanningTreeTypes_STP_EDGE_PORT_EDGE_DISABLE
			}

			opLinkType := (&operDbVal).Get("link_type")
			if opLinkType == "shared" {
				intf.State.LinkType = ocbinds.OpenconfigSpanningTree_StpLinkType_SHARED
			} else if opLinkType == "point-to-point" {
				intf.State.LinkType = ocbinds.OpenconfigSpanningTree_StpLinkType_P2P
			}
		}
	} else {
		for intfName := range app.intfTableMap {
			intfPtr, _ := interfaces.NewInterface(intfName)
			ygot.BuildEmptyTree(intfPtr)
			app.convertInternalToOCStpInterfaces(intfName, interfaces, intfPtr)
		}
	}
}

func (app *StpApp) convertOperInternalToOCVlanInterface(vlanName string, intfId string, vlan interface{}, vlanIntf interface{}) {
	if len(intfId) > 0 {
		var pvstVlan *ocbinds.OpenconfigSpanningTree_Stp_Pvst_Vlan
		var pvstVlanIntf *ocbinds.OpenconfigSpanningTree_Stp_Pvst_Vlan_Interfaces_Interface
		var rpvstVlan *ocbinds.OpenconfigSpanningTree_Stp_RapidPvst_Vlan
		var rpvstVlanIntf *ocbinds.OpenconfigSpanningTree_Stp_RapidPvst_Vlan_Interfaces_Interface
		var num uint64

		switch reflect.TypeOf(vlan).Elem().Name() {
		case "OpenconfigSpanningTree_Stp_Pvst_Vlan":
			pvstVlan, _ = vlan.(*ocbinds.OpenconfigSpanningTree_Stp_Pvst_Vlan)
			if vlanIntf == nil {
				pvstVlanIntf = pvstVlan.Interfaces.Interface[intfId]
				if pvstVlanIntf == nil {
					pvstVlanIntf, _ = pvstVlan.Interfaces.NewInterface(intfId)
				}
				ygot.BuildEmptyTree(pvstVlanIntf)
			} else {
				pvstVlanIntf, _ = vlanIntf.(*ocbinds.OpenconfigSpanningTree_Stp_Pvst_Vlan_Interfaces_Interface)
			}
			ygot.BuildEmptyTree(pvstVlanIntf.State)
		case "OpenconfigSpanningTree_Stp_RapidPvst_Vlan":
			rpvstVlan, _ = vlan.(*ocbinds.OpenconfigSpanningTree_Stp_RapidPvst_Vlan)
			if vlanIntf == nil {
				rpvstVlanIntf = rpvstVlan.Interfaces.Interface[intfId]
				if rpvstVlanIntf == nil {
					rpvstVlanIntf, _ = rpvstVlan.Interfaces.NewInterface(intfId)
				}
				ygot.BuildEmptyTree(rpvstVlanIntf)
			} else {
				rpvstVlanIntf, _ = vlanIntf.(*ocbinds.OpenconfigSpanningTree_Stp_RapidPvst_Vlan_Interfaces_Interface)
			}
			ygot.BuildEmptyTree(rpvstVlanIntf.State)
		}

		operDbVal := app.vlanIntfOperTableMap[vlanName][intfId]

		if operDbVal.IsPopulated() {
			num, _ = strconv.ParseUint((&operDbVal).Get("port_num"), 10, 16)
			opPortNum := uint16(num)

			num, _ = strconv.ParseUint((&operDbVal).Get("path_cost"), 10, 32)
			opcost := uint32(num)

			num, _ = strconv.ParseUint((&operDbVal).Get("priority"), 10, 8)
			opPortPriority := uint8(num)

			num, _ = strconv.ParseUint((&operDbVal).Get("desig_cost"), 10, 32)
			opDesigCost := uint32(num)

			num, _ = strconv.ParseUint((&operDbVal).Get("desig_port"), 10, 16)
			opDesigPortNum := uint16(num)

			num, _ = strconv.ParseUint((&operDbVal).Get("root_guard_timer"), 10, 16)
			opRootGuardTimer := uint16(num)

			num, _ = strconv.ParseUint((&operDbVal).Get("fwd_transitions"), 10, 64)
			opFwtrans := num

			desigRootAddr := (&operDbVal).Get("desig_root")

			desigBridgeAddr := (&operDbVal).Get("desig_bridge")

			portState := (&operDbVal).Get("port_state")

			//Counters
			num, _ = strconv.ParseUint((&operDbVal).Get("bpdu_sent"), 10, 64)
			opBpduSent := num

			num, _ = strconv.ParseUint((&operDbVal).Get("bpdu_received"), 10, 64)
			opBpduReceived := num

			num, _ = strconv.ParseUint((&operDbVal).Get("tcn_sent"), 10, 64)
			opTcnSent := num

			num, _ = strconv.ParseUint((&operDbVal).Get("tcn_received"), 10, 64)
			opTcnReceived := num

			// For RPVST+ only
			num, _ = strconv.ParseUint((&operDbVal).Get("config_bpdu_sent"), 10, 64)
			opConfigBpduSent := num

			num, _ = strconv.ParseUint((&operDbVal).Get("config_bpdu_received"), 10, 64)
			opConfigBpduReceived := num

			if pvstVlanIntf != nil && pvstVlanIntf.State != nil {
				pvstVlanIntf.State.Name = &intfId
				pvstVlanIntf.State.PortNum = &opPortNum
				pvstVlanIntf.State.Cost = &opcost
				pvstVlanIntf.State.PortPriority = &opPortPriority
				pvstVlanIntf.State.DesignatedCost = &opDesigCost
				pvstVlanIntf.State.DesignatedPortNum = &opDesigPortNum
				pvstVlanIntf.State.RootGuardTimer = &opRootGuardTimer
				pvstVlanIntf.State.ForwardTransisitions = &opFwtrans
				pvstVlanIntf.State.DesignatedRootAddress = &desigRootAddr
				pvstVlanIntf.State.DesignatedBridgeAddress = &desigBridgeAddr
				switch portState {
				case "DISABLED":
					pvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_DISABLED
				case "BLOCKING":
					pvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_BLOCKING
				case "LISTENING":
					pvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_LISTENING
				case "LEARNING":
					pvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_LEARNING
				case "FORWARDING":
					pvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_FORWARDING
				}
				if pvstVlanIntf.State.Counters != nil {
					pvstVlanIntf.State.Counters.BpduSent = &opBpduSent
					pvstVlanIntf.State.Counters.BpduReceived = &opBpduReceived
					pvstVlanIntf.State.Counters.TcnSent = &opTcnSent
					pvstVlanIntf.State.Counters.TcnReceived = &opTcnReceived
				}
			} else if rpvstVlanIntf != nil && rpvstVlanIntf.State != nil {
				rpvstVlanIntf.State.Name = &intfId
				rpvstVlanIntf.State.PortNum = &opPortNum
				rpvstVlanIntf.State.Cost = &opcost
				rpvstVlanIntf.State.PortPriority = &opPortPriority
				rpvstVlanIntf.State.DesignatedCost = &opDesigCost
				rpvstVlanIntf.State.DesignatedPortNum = &opDesigPortNum
				rpvstVlanIntf.State.RootGuardTimer = &opRootGuardTimer
				rpvstVlanIntf.State.ForwardTransisitions = &opFwtrans
				rpvstVlanIntf.State.DesignatedRootAddress = &desigRootAddr
				rpvstVlanIntf.State.DesignatedBridgeAddress = &desigBridgeAddr
				switch portState {
				case "DISABLED":
					rpvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_DISABLED
				case "BLOCKING":
					rpvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_BLOCKING
				case "LISTENING":
					rpvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_LISTENING
				case "LEARNING":
					rpvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_LEARNING
				case "FORWARDING":
					rpvstVlanIntf.State.PortState = ocbinds.OpenconfigSpanningTreeTypes_STP_PORT_STATE_FORWARDING
				}
				if rpvstVlanIntf.State.Counters != nil {
					rpvstVlanIntf.State.Counters.BpduSent = &opBpduSent
					rpvstVlanIntf.State.Counters.BpduReceived = &opBpduReceived
					rpvstVlanIntf.State.Counters.TcnSent = &opTcnSent
					rpvstVlanIntf.State.Counters.TcnReceived = &opTcnReceived
					rpvstVlanIntf.State.Counters.ConfigBpduSent = &opConfigBpduSent
					rpvstVlanIntf.State.Counters.ConfigBpduReceived = &opConfigBpduReceived
				}
			}
		}
	} else {
		vlanIntfOperKeys, _ := app.appDB.GetKeys(app.vlanIntfOperTable)
		for i, _ := range vlanIntfOperKeys {
			entryKey := vlanIntfOperKeys[i]
			if vlanName == (&entryKey).Get(0) {
				app.convertOperInternalToOCVlanInterface(vlanName, (&entryKey).Get(1), vlan, vlanIntf)
			}
		}
	}
}

func (app *StpApp) convertApplDBRpvstVlanToInternal(vlanName string) error {
	var err error

	rpvstVlanOperState, err := app.appDB.GetEntry(app.vlanOperTable, asKey(vlanName))
	if err != nil {
		return err
	}
	app.vlanOperTableMap[vlanName] = rpvstVlanOperState

	app.convertApplDBRpvstVlanInterfaceToInternal(vlanName, "")

	return err
}

func (app *StpApp) convertApplDBRpvstVlanInterfaceToInternal(vlanName string, intfId string) error {
	var err error

	if app.vlanIntfOperTableMap[vlanName] == nil {
		app.vlanIntfOperTableMap[vlanName] = make(map[string]db.Value)
	}

	if len(intfId) > 0 {
		rpvstVlanIntfOperState, err := app.appDB.GetEntry(app.vlanIntfOperTable, asKey(vlanName, intfId))
		if err != nil {
			return err
		}
		app.vlanIntfOperTableMap[vlanName][intfId] = rpvstVlanIntfOperState
	} else {
		vlanIntfOperKeys, _ := app.appDB.GetKeys(app.vlanIntfOperTable)
		for i, _ := range vlanIntfOperKeys {
			entryKey := vlanIntfOperKeys[i]
			if vlanName == (&entryKey).Get(0) {
				app.convertApplDBRpvstVlanInterfaceToInternal(vlanName, (&entryKey).Get(1))
			}
		}
	}

	return err
}

func (app *StpApp) convertApplDBStpInterfacesToInternal(intfId string) error {
	var err error

	intfOperState, err := app.appDB.GetEntry(app.intfOperTable, asKey(intfId))
	if err != nil {
		return err
	}
	app.intfOperTableMap[intfId] = intfOperState

	return err
}

func (app *StpApp) convertOCStpModeToInternal(mode ocbinds.E_OpenconfigSpanningTreeTypes_STP_PROTOCOL) string {
	switch mode {
	case ocbinds.OpenconfigSpanningTreeTypes_STP_PROTOCOL_MSTP:
		return "mstp"
	case ocbinds.OpenconfigSpanningTreeTypes_STP_PROTOCOL_PVST:
		return "pvst"
	case ocbinds.OpenconfigSpanningTreeTypes_STP_PROTOCOL_RAPID_PVST:
		return "rpvst"
	case ocbinds.OpenconfigSpanningTreeTypes_STP_PROTOCOL_RSTP:
		return "rstp"
	default:
		return ""
	}
}

func (app *StpApp) convertInternalStpModeToOC(mode string) []ocbinds.E_OpenconfigSpanningTreeTypes_STP_PROTOCOL {
	var stpModes []ocbinds.E_OpenconfigSpanningTreeTypes_STP_PROTOCOL
	if len(mode) > 0 {
		switch mode {
		case "pvst":
			stpModes = append(stpModes, ocbinds.OpenconfigSpanningTreeTypes_STP_PROTOCOL_PVST)
		case "rpvst":
			stpModes = append(stpModes, ocbinds.OpenconfigSpanningTreeTypes_STP_PROTOCOL_RAPID_PVST)
		case "mstp":
			stpModes = append(stpModes, ocbinds.OpenconfigSpanningTreeTypes_STP_PROTOCOL_MSTP)
		case "rstp":
			stpModes = append(stpModes, ocbinds.OpenconfigSpanningTreeTypes_STP_PROTOCOL_RSTP)
		}
	}
	return stpModes
}

func (app *StpApp) getStpModeFromConfigDB(d *db.DB) (string, error) {
	stpGlobalDbEntry, err := d.GetEntry(app.globalTable, asKey("GLOBAL"))
	if err != nil {
		return "", err
	}
	return (&stpGlobalDbEntry).Get(STP_MODE), nil
}

func (app *StpApp) getAllInterfacesFromVlanMemberTable(d *db.DB) ([]string, error) {
	var intfList []string

	keys, err := d.GetKeys(&db.TableSpec{Name: "VLAN_MEMBER"})
	if err != nil {
		return intfList, err
	}
	for i, _ := range keys {
		key := keys[i]
		if !contains(intfList, (&key).Get(1)) {
			intfList = append(intfList, (&key).Get(1))
		}
	}
	return intfList, err
}

func (app *StpApp) getAllInterfacesFromPortChannelMemberTable(d *db.DB) ([]string, error) {
	var intfList []string

	keys, err := d.GetKeys(&db.TableSpec{Name: "PORTCHANNEL_MEMBER"})
	if err != nil {
		return intfList, err
	}
	for i, _ := range keys {
		key := keys[i]
		if !contains(intfList, (&key).Get(1)) {
			intfList = append(intfList, (&key).Get(1))
		}
	}
	return intfList, err
}

func (app *StpApp) enableStpForInterfaces(d *db.DB) error {
	defaultDBValues := db.Value{Field: map[string]string{}}
	(&defaultDBValues).Set("enabled", "true")
	(&defaultDBValues).Set("root_guard", "false")
	(&defaultDBValues).Set("bpdu_guard", "false")
	(&defaultDBValues).Set("bpdu_filter", "global")
	(&defaultDBValues).Set("bpdu_guard_do_disable", "false")
	(&defaultDBValues).Set("portfast", "true")
	(&defaultDBValues).Set("uplink_fast", "false")

	intfList, err := app.getAllInterfacesFromVlanMemberTable(d)
	if err != nil {
		return err
	}

	portKeys, err := d.GetKeys(&db.TableSpec{Name: "PORT"})
	if err != nil {
		return err
	}
	for i, _ := range portKeys {
		portKey := portKeys[i]
		if contains(intfList, (&portKey).Get(0)) {
			d.CreateEntry(app.interfaceTable, portKey, defaultDBValues)
		}
	}

	// For portchannels
	portchKeys, err := d.GetKeys(&db.TableSpec{Name: "PORTCHANNEL"})
	if err != nil {
		return err
	}
	for i, _ := range portchKeys {
		portchKey := portchKeys[i]
		if contains(intfList, (&portchKey).Get(0)) {
			d.CreateEntry(app.interfaceTable, portchKey, defaultDBValues)
		}
	}
	return err
}

func (app *StpApp) enableStpForVlans(d *db.DB) error {
	stpGlobalVal := app.globalInfo
	fDelay := (&stpGlobalVal).Get("forward_delay")
	helloTime := (&stpGlobalVal).Get("hello_time")
	maxAge := (&stpGlobalVal).Get("max_age")
	priority := (&stpGlobalVal).Get("priority")

	vlanKeys, err := d.GetKeys(&db.TableSpec{Name: "VLAN"})
	if err != nil {
		return err
	}

	var vlanList []string
	for i, _ := range vlanKeys {
		vlanKey := vlanKeys[i]
		vlanList = append(vlanList, (&vlanKey).Get(0))
	}

	// Sort vlanList in natural order such that 'Vlan2' < 'Vlan10'
	natsort.Sort(vlanList)

	for i, _ := range vlanList {
		if i < PVST_MAX_INSTANCES {
			defaultDBValues := db.Value{Field: map[string]string{}}
			(&defaultDBValues).Set("enabled", "true")
			(&defaultDBValues).Set("forward_delay", fDelay)
			(&defaultDBValues).Set("hello_time", helloTime)
			(&defaultDBValues).Set("max_age", maxAge)
			(&defaultDBValues).Set("priority", priority)

			vlanId := strings.Replace(vlanList[i], "Vlan", "", 1)
			(&defaultDBValues).Set("vlanid", vlanId)
			d.CreateEntry(app.vlanTable, asKey(vlanList[i]), defaultDBValues)
		}
	}
	return err
}

func enableStpOnVlanCreation(d *db.DB, vlanList []string) {
	if len(vlanList) == 0 {
		return
	}
	log.Infof("enableStpOnVlanCreation --> Enable Stp on Vlans: %v", vlanList)
	vlanKeys, _ := d.GetKeys(&db.TableSpec{Name: STP_VLAN_TABLE})
	existingEntriesCount := len(vlanKeys)
	if existingEntriesCount < PVST_MAX_INSTANCES {
		stpGlobalDBEntry, err := d.GetEntry(&db.TableSpec{Name: STP_GLOBAL_TABLE}, asKey("GLOBAL"))
		if err != nil {
			return
		}
		fDelay := (&stpGlobalDBEntry).Get("forward_delay")
		helloTime := (&stpGlobalDBEntry).Get("hello_time")
		maxAge := (&stpGlobalDBEntry).Get("max_age")
		priority := (&stpGlobalDBEntry).Get("priority")

		// Sort vlanList in natural order such that 'Vlan2' < 'Vlan10'
		natsort.Sort(vlanList)

		for i, _ := range vlanList {
			if (existingEntriesCount + i) < PVST_MAX_INSTANCES {
				defaultDBValues := db.Value{Field: map[string]string{}}
				(&defaultDBValues).Set("enabled", "true")
				(&defaultDBValues).Set("forward_delay", fDelay)
				(&defaultDBValues).Set("hello_time", helloTime)
				(&defaultDBValues).Set("max_age", maxAge)
				(&defaultDBValues).Set("priority", priority)

				vlanId := strings.Replace(vlanList[i], "Vlan", "", 1)
				(&defaultDBValues).Set("vlanid", vlanId)
				d.CreateEntry(&db.TableSpec{Name: STP_VLAN_TABLE}, asKey(vlanList[i]), defaultDBValues)
			}
		}
	}
}

func removeStpConfigOnVlanDeletion(d *db.DB, vlanList []string) {
	if len(vlanList) == 0 {
		return
	}
	log.Infof("removeStpConfigOnVlanDeletion --> Disable Stp on Vlans: %v", vlanList)
	for i, _ := range vlanList {
		err := d.DeleteEntry(&db.TableSpec{Name: STP_VLAN_INTF_TABLE}, asKey(vlanList[i], "*"))
		if err != nil {
			log.Error(err)
		}
		err = d.DeleteEntry(&db.TableSpec{Name: STP_VLAN_TABLE}, asKey(vlanList[i]))
		if err != nil {
			log.Error(err)
		}
	}
}

// This function accepts list of Interface names (i.e. Eth or Portchannel)
// on which switchport is enabled
func enableStpOnInterfaceVlanMembership(d *db.DB, intfList []string) {
	if len(intfList) == 0 {
		return
	}
	_, serr := d.GetEntry(&db.TableSpec{Name: STP_GLOBAL_TABLE}, asKey("GLOBAL"))
	if serr != nil {
		return
	}
	log.Infof("enableStpOnInterfaceVlanMembership --> Enable Stp on Interfaces: %v", intfList)
	defaultDBValues := db.Value{Field: map[string]string{}}
	(&defaultDBValues).Set("enabled", "true")
	(&defaultDBValues).Set("root_guard", "false")
	(&defaultDBValues).Set("bpdu_guard", "false")
	(&defaultDBValues).Set("bpdu_filter", "global")
	(&defaultDBValues).Set("bpdu_guard_do_disable", "false")
	(&defaultDBValues).Set("portfast", "true")
	(&defaultDBValues).Set("uplink_fast", "false")

	var stpEnabledIntfList []string
	intfKeys, err := d.GetKeys(&db.TableSpec{Name: STP_INTF_TABLE})
	if err != nil {
		log.Error(err)
	} else {
		for i, _ := range intfKeys {
			dbKey := intfKeys[i]
			stpEnabledIntfList = append(stpEnabledIntfList, (&dbKey).Get(0))
		}

		for i, _ := range intfList {
			if !contains(stpEnabledIntfList, intfList[i]) {
				d.CreateEntry(&db.TableSpec{Name: STP_INTF_TABLE}, asKey(intfList[i]), defaultDBValues)
			}
		}
	}
}

// This function accepts list of Interface names (i.e. Eth or Portchannel)
// on which switchport is disabled
func removeStpOnInterfaceSwitchportDeletion(d *db.DB, intfList []string) {
	if len(intfList) == 0 {
		return
	}
	log.Infof("removeStpOnInterfaceSwitchportDeletion --> Disable Stp on Interfaces: %v", intfList)
	for i, _ := range intfList {
		err := d.DeleteEntry(&db.TableSpec{Name: STP_VLAN_INTF_TABLE}, asKey("*", intfList[i]))
		if err != nil {
			log.Error(err)
		}
		err = d.DeleteEntry(&db.TableSpec{Name: STP_INTF_TABLE}, asKey(intfList[i]))
		if err != nil {
			log.Error(err)
		}
	}
}

func (app *StpApp) updateGlobalFieldsToStpVlanTable(d *db.DB, fldValuePair map[string]string, stpGlobalDbEntry db.Value) error {
	vlanKeys, err := d.GetKeys(app.vlanTable)
	if err != nil {
		return err
	}
	for i, _ := range vlanKeys {
		vlanEntry, _ := d.GetEntry(app.vlanTable, vlanKeys[i])

		for fldName := range fldValuePair {
			valStr := fldValuePair[fldName]
			globalFldVal := (&stpGlobalDbEntry).Get(fldName)

			if (&vlanEntry).Get(fldName) == globalFldVal {
				(&vlanEntry).Set(fldName, valStr)
				err := d.ModEntry(app.vlanTable, vlanKeys[i], vlanEntry)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (app *StpApp) enableStpMode(d *db.DB) error {
	var err error
	err = app.setStpGlobalConfigInDB(d)
	if err != nil {
		return err
	}
	err = app.enableStpForInterfaces(d)
	if err != nil {
		return err
	}
	err = app.enableStpForVlans(d)

	return err
}

func (app *StpApp) disableStpMode(d *db.DB) error {
	var err error
	err = d.DeleteTable(app.vlanIntfTable)
	if err != nil {
		return err
	}
	err = d.DeleteTable(app.vlanTable)
	if err != nil {
		return err
	}
	err = d.DeleteTable(app.interfaceTable)
	if err != nil {
		return err
	}
	err = d.DeleteTable(app.globalTable)

	return err
}

func (app *StpApp) handleStpGlobalFieldsUpdation(d *db.DB, opcode int) error {
	stpGlobalDBEntry, err := d.GetEntry(app.globalTable, asKey("GLOBAL"))
	if err != nil {
		return err
	}
	// Make a copy of StpGlobalDBEntry to modify fields value.
	tmpDbEntry := db.Value{Field: map[string]string{}}
	for field, value := range stpGlobalDBEntry.Field {
		tmpDbEntry.Field[field] = value
	}
	fldValuePair := make(map[string]string)

	for fld := range app.globalInfo.Field {
		valStr := app.globalInfo.Field[fld]
		(&tmpDbEntry).Set(fld, valStr)

		if fld != "rootguard_timeout" && fld != "bpdu_filter" {
			fldValuePair[fld] = valStr
		}
	}

	err = d.ModEntry(app.globalTable, asKey("GLOBAL"), tmpDbEntry)
	if err != nil {
		return err
	}

	err = app.updateGlobalFieldsToStpVlanTable(d, fldValuePair, stpGlobalDBEntry)
	if err != nil {
		return err
	}

	return nil
}

func (app *StpApp) handleStpGlobalFieldsDeletion(d *db.DB) error {
	stpGlobalDBEntry, err := d.GetEntry(app.globalTable, asKey("GLOBAL"))
	if err != nil {
		return err
	}

	nodeInfo, err := getTargetNodeYangSchema(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device))
	if err != nil {
		return err
	}
	log.Infof("Node received for deletion: %s", nodeInfo.Name)
	if nodeInfo.IsLeaf() {
		var fldName, valStr string
		switch nodeInfo.Name {
		case "rootguard-timeout":
			fldName = "rootguard_timeout"
			valStr = STP_DEFAULT_ROOT_GUARD_TIMEOUT
		case "hello-time":
			fldName = "hello_time"
			valStr = STP_DEFAULT_HELLO_INTERVAL
		case "max-age":
			fldName = "max_age"
			valStr = STP_DEFAULT_MAX_AGE
		case "forwarding-delay":
			fldName = "forward_delay"
			valStr = STP_DEFAULT_FORWARD_DELAY
		case "bridge-priority":
			fldName = "priority"
			valStr = STP_DEFAULT_BRIDGE_PRIORITY
		case "bpdu-filter":
			fldName = "bpdu_filter"
			valStr = STP_DEFAULT_BPDU_FILTER
		}

		// Make a copy of StpGlobalDBEntry to modify fields value.
		tmpDbEntry := db.Value{Field: map[string]string{}}
		for field, value := range stpGlobalDBEntry.Field {
			tmpDbEntry.Field[field] = value
		}
		fldValuePair := make(map[string]string)

		(&tmpDbEntry).Set(fldName, valStr)
		err = d.ModEntry(app.globalTable, asKey("GLOBAL"), tmpDbEntry)
		if err != nil {
			return err
		}

		if fldName != "rootguard_timeout" && fldName != "bpdu_filter" {
			fldValuePair[fldName] = valStr
		}

		err = app.updateGlobalFieldsToStpVlanTable(d, fldValuePair, stpGlobalDBEntry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *StpApp) handleVlanInterfaceFieldsDeletion(d *db.DB, vlanName string, intfId string) error {
	dbEntry, err := d.GetEntry(app.vlanIntfTable, asKey(vlanName, intfId))
	if err != nil {
		return err
	}

	nodeInfo, err := getTargetNodeYangSchema(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device))
	if err != nil {
		return err
	}
	log.Infof("Node received for deletion: %s", nodeInfo.Name)
	if nodeInfo.IsLeaf() {
		switch nodeInfo.Name {
		case "cost":
			(&dbEntry).Remove("path_cost")
		case "port-priority":
			(&dbEntry).Remove("priority")
		}
	}

	err = d.SetEntry(app.vlanIntfTable, asKey(vlanName, intfId), dbEntry)
	if err != nil {
		return err
	}

	return nil
}

func (app *StpApp) handleVlanFieldsDeletion(d *db.DB, vlanName string) error {
	stpGlobalDBEntry, err := d.GetEntry(app.globalTable, asKey("GLOBAL"))
	if err != nil {
		return err
	}
	fDelay := (&stpGlobalDBEntry).Get("forward_delay")
	helloTime := (&stpGlobalDBEntry).Get("hello_time")
	maxAge := (&stpGlobalDBEntry).Get("max_age")
	priority := (&stpGlobalDBEntry).Get("priority")

	dbEntry, err := d.GetEntry(app.vlanTable, asKey(vlanName))
	if err != nil {
		return err
	}

	nodeInfo, err := getTargetNodeYangSchema(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device))
	if err != nil {
		return err
	}
	log.Infof("Node received for deletion: %s", nodeInfo.Name)
	if nodeInfo.IsLeaf() {
		switch nodeInfo.Name {
		case "hello-time":
			(&dbEntry).Set("hello_time", helloTime)
		case "max-age":
			(&dbEntry).Set("max_age", maxAge)
		case "bridge-priority":
			(&dbEntry).Set("priority", priority)
		case "forwarding-delay":
			(&dbEntry).Set("forward_delay", fDelay)
		case "spanning-tree-enable":
			(&dbEntry).Set("enabled", "false")
		}
	}

	err = d.ModEntry(app.vlanTable, asKey(vlanName), dbEntry)
	if err != nil {
		return err
	}

	return nil
}

func (app *StpApp) handleInterfacesFieldsDeletion(d *db.DB, intfId string) error {
	dbEntry, err := d.GetEntry(app.interfaceTable, asKey(intfId))
	if err != nil {
		return err
	}

	nodeInfo, err := getTargetNodeYangSchema(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device))
	if err != nil {
		return err
	}
	log.Infof("Node received for deletion: %s", nodeInfo.Name)
	if nodeInfo.IsLeaf() {
		switch nodeInfo.Name {
		case "guard":
			(&dbEntry).Remove("root_guard")
		case "bpdu-guard":
			(&dbEntry).Remove("bpdu_guard")
		case "bpdu-filter":
			(&dbEntry).Set("bpdu_filter", "global")
		case "portfast":
			(&dbEntry).Remove("portfast")
		case "uplink-fast":
			(&dbEntry).Remove("uplink_fast")
		case "bpdu-guard-port-shutdown":
			(&dbEntry).Remove("bpdu_guard_do_disable")
		case "cost":
			(&dbEntry).Remove("path_cost")
		case "port-priority":
			(&dbEntry).Remove("priority")
		case "spanning-tree-enable":
			(&dbEntry).Remove("enabled")
		case "edge-port":
			(&dbEntry).Remove("edge_port")
		case "link-type":
			(&dbEntry).Set("link_type", "auto")
		}
	}

	err = d.SetEntry(app.interfaceTable, asKey(intfId), dbEntry)
	if err != nil {
		return err
	}

	return nil
}
