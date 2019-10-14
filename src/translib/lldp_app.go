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
    "strconv"
    "reflect"
    "errors"
    "translib/db"
    "translib/ocbinds"
    "github.com/openconfig/ygot/ygot"
    log "github.com/golang/glog"
    "strings"
    "encoding/hex"
    "translib/tlerr"
)

const (
    LLDP_REMOTE_CAP_ENABLED     = "lldp_rem_sys_cap_enabled"
    LLDP_REMOTE_SYS_NAME        = "lldp_rem_sys_name"
    LLDP_REMOTE_PORT_DESC       = "lldp_rem_port_desc"
    LLDP_REMOTE_CHASS_ID        = "lldp_rem_chassis_id"
    LLDP_REMOTE_CAP_SUPPORTED   = "lldp_rem_sys_cap_supported"
    LLDP_REMOTE_PORT_ID_SUBTYPE = "lldp_rem_port_id_subtype"
    LLDP_REMOTE_SYS_DESC        = "lldp_rem_sys_desc"
    LLDP_REMOTE_REM_TIME        = "lldp_rem_time_mark"
    LLDP_REMOTE_PORT_ID         = "lldp_rem_port_id"
    LLDP_REMOTE_REM_ID          = "lldp_rem_index"
    LLDP_REMOTE_CHASS_ID_SUBTYPE = "lldp_rem_chassis_id_subtype"
    LLDP_REMOTE_MAN_ADDR        = "lldp_rem_man_addr"
)

type lldpApp struct {
    path        *PathInfo
    ygotRoot    *ygot.GoStruct
    ygotTarget  *interface{}
    appDb       *db.DB
    neighTs     *db.TableSpec
    lldpTableMap map[string]db.Value
    lldpNeighTableMap map[string]map[string]string
    lldpCapTableMap map[string]map[string]bool
}

func init() {
    log.Info("Init called for LLDP modules module")
    err := register("/openconfig-lldp:lldp",
                    &appInfo{appType: reflect.TypeOf(lldpApp{}),
                    ygotRootType: reflect.TypeOf(ocbinds.OpenconfigLldp_Lldp{}),
                    isNative: false})
    if err != nil {
        log.Fatal("Register LLDP app module with App Interface failed with error=", err)
    }

    err = addModel(&ModelData{Name: "openconfig-lldp",
    Org: "OpenConfig working group",
    Ver:      "1.0.2"})
    if err != nil {
        log.Fatal("Adding model data to appinterface failed with error=", err)
    }
}

func (app *lldpApp) initialize(data appData) {
    log.Info("initialize:lldp:path =", data.path)
    *app = lldpApp{path: NewPathInfo(data.path), ygotRoot: data.ygotRoot, ygotTarget: data.ygotTarget}
    app.neighTs = &db.TableSpec{Name: "LLDP_ENTRY_TABLE"}
    app.lldpTableMap = make(map[string]db.Value)
    app.lldpNeighTableMap = make(map[string]map[string]string)
    app.lldpCapTableMap = make(map[string]map[string]bool)
}

func (app *lldpApp) getAppRootObject() (*ocbinds.OpenconfigLldp_Lldp) {
       deviceObj := (*app.ygotRoot).(*ocbinds.Device)
       return deviceObj.Lldp
}

func (app *lldpApp) translateCreate(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys
    log.Info("translateCreate:lldp:path =", app.path)

    err = errors.New("Not implemented")
    return keys, err
}

func (app *lldpApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys
    log.Info("translateUpdate:lldp:path =", app.path)

    err = errors.New("Not implemented")
    return keys, err
}

func (app *lldpApp) translateReplace(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys
    log.Info("translateReplace:lldp:path =", app.path)

    err = errors.New("Not implemented")
    return keys, err
}

func (app *lldpApp) translateDelete(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys
    log.Info("translateDelete:lldp:path =", app.path)

    err = errors.New("Not implemented")
    return keys, err
}

func (app *lldpApp) translateGet(dbs [db.MaxDB]*db.DB) error  {
    var err error
    log.Info("translateGet:lldp:path = ", app.path)

    return err
}

func (app *lldpApp) translateAction(dbs [db.MaxDB]*db.DB) error {
    err := errors.New("Not supported")
    return err
}

func (app *lldpApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
    pathInfo := NewPathInfo(path)
    notifInfo := notificationInfo{dbno: db.ApplDB}
    notSupported := tlerr.NotSupportedError{Format: "Subscribe not supported", Path: path}

    if isSubtreeRequest(pathInfo.Template, "/openconfig-lldp:lldp/interfaces") {
        if pathInfo.HasSuffix("/neighbors") ||
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
        if pathInfo.HasSuffix("/interface{}") {
            notifInfo.table = db.TableSpec{Name: "LLDP_ENTRY_TABLE"}
            notifInfo.key = asKey(ifKey)
            notifInfo.needCache = true
            return &notificationOpts{pType: OnChange}, &notifInfo, nil
        }
    }
    return nil, nil, notSupported
}

func (app *lldpApp) processCreate(d *db.DB) (SetResponse, error)  {
    var err error

    err = errors.New("Not implemented")
    var resp SetResponse

    return resp, err
}

func (app *lldpApp) processUpdate(d *db.DB) (SetResponse, error)  {
    var err error

    err = errors.New("Not implemented")
    var resp SetResponse

    return resp, err
}

func (app *lldpApp) processReplace(d *db.DB) (SetResponse, error)  {
    var err error
    var resp SetResponse
    err = errors.New("Not implemented")

    return resp, err
}

func (app *lldpApp) processDelete(d *db.DB) (SetResponse, error)  {
    var err error
    err = errors.New("Not implemented")
    var resp SetResponse

    return resp, err
}

func (app *lldpApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error)  {
    var err error
    var payload []byte

    app.appDb = dbs[db.ApplDB]
    lldpIntfObj := app.getAppRootObject()

    targetUriPath, err := getYangPathFromUri(app.path.Path)
    log.Info("lldp processGet")
    log.Info("targetUriPath: ", targetUriPath)

    if targetUriPath == "/openconfig-lldp:lldp/interfaces" {
        log.Info("Requesting interfaces")
        app.getLldpInfoFromDB(nil)
        ygot.BuildEmptyTree(lldpIntfObj)
        ifInfo := lldpIntfObj.Interfaces
        ygot.BuildEmptyTree(ifInfo)
        for ifname,_  := range app.lldpNeighTableMap {
           oneIfInfo, err := ifInfo.NewInterface(ifname)
           if err != nil {
                log.Info("Creation of subinterface subtree failed!")
                return GetResponse{Payload: payload, ErrSrc: AppErr}, err
            }
            ygot.BuildEmptyTree(oneIfInfo)
            app.getLldpNeighInfoFromInternalMap(&ifname, oneIfInfo)
            if *app.ygotTarget == lldpIntfObj.Interfaces {
                payload, err = dumpIetfJson(lldpIntfObj, true)
            } else {
                log.Info("Wrong request!")
            }

        }
    } else if targetUriPath == "/openconfig-lldp:lldp/interfaces/interface" {
        intfObj := lldpIntfObj.Interfaces
        ygot.BuildEmptyTree(intfObj)
        if intfObj.Interface != nil && len(intfObj.Interface) > 0 {
            for ifname, _ := range intfObj.Interface {
                log.Info("if-name = ", ifname)
                app.getLldpInfoFromDB(&ifname)
                ifInfo := intfObj.Interface[ifname]
                ygot.BuildEmptyTree(ifInfo)
                app.getLldpNeighInfoFromInternalMap(&ifname, ifInfo)

                if *app.ygotTarget == intfObj.Interface[ifname] {
                    payload, err = dumpIetfJson(intfObj, true)
                    if err != nil {
                        log.Info("Creation of subinterface subtree failed!")
                        return GetResponse{Payload: payload, ErrSrc: AppErr}, err
                    }
                } else {
                    log.Info("Wrong request!")
                }
            }
        } else {
            log.Info("No data")
        }
   }

   return GetResponse{Payload:payload}, err
}

func (app *lldpApp) processAction(dbs [db.MaxDB]*db.DB) (ActionResponse, error) {
    var resp ActionResponse
    err := errors.New("Not implemented")

    return resp, err
}

/** Helper function to populate JSON response for GET request **/
func (app *lldpApp) getLldpNeighInfoFromInternalMap(ifName *string, ifInfo *ocbinds.OpenconfigLldp_Lldp_Interfaces_Interface) {

    ngInfo, err := ifInfo.Neighbors.NewNeighbor(*ifName)
    if err != nil {
        log.Info("Creation of subinterface subtree failed!")
        return
    }
    ygot.BuildEmptyTree(ngInfo)
    neighAttrMap:= app.lldpNeighTableMap[*ifName]
    for attr, value := range neighAttrMap {
        switch attr {
            case LLDP_REMOTE_SYS_NAME:
                name  := new(string)
                *name  = value
                ngInfo.State.SystemName = name
            case LLDP_REMOTE_PORT_DESC:
                pdescr := new(string)
                *pdescr = value
                ngInfo.State.PortDescription = pdescr
            case LLDP_REMOTE_CHASS_ID:
                chId := new (string)
                *chId = value
                ngInfo.State.ChassisId = chId
            case LLDP_REMOTE_PORT_ID_SUBTYPE:
                remPortIdTypeVal, err :=  strconv.Atoi(value)
                if err == nil {
                        ngInfo.State.PortIdType =ocbinds.E_OpenconfigLldp_PortIdType(remPortIdTypeVal)
                }
            case LLDP_REMOTE_SYS_DESC:
                sdesc:= new(string)
                *sdesc = value
                ngInfo.State.SystemDescription = sdesc
            case LLDP_REMOTE_REM_TIME:
            /* Ignore Remote System time */
            case LLDP_REMOTE_PORT_ID:
                remPortIdPtr := new(string)
                *remPortIdPtr = value
                ngInfo.State.PortId = remPortIdPtr
            case LLDP_REMOTE_REM_ID:
                Id := new(string)
                *Id = value
                ngInfo.State.Id = Id
            case LLDP_REMOTE_CHASS_ID_SUBTYPE:
                remChassIdTypeVal , err:=strconv.Atoi(value)
                if err  == nil {
                        ngInfo.State.ChassisIdType =ocbinds.E_OpenconfigLldp_ChassisIdType(remChassIdTypeVal)
                }
            case LLDP_REMOTE_MAN_ADDR:
                mgmtAdr:= new(string)
                *mgmtAdr = value
                ngInfo.State.ManagementAddress = mgmtAdr
            default:
                log.Info("Not a valid attribute!")
        }
    }
    capLst := app.lldpCapTableMap[*ifName]
    for capName, enabled := range capLst {
        if capName == "Router" {
             capInfo, err :=  ngInfo.Capabilities.NewCapability(6)
             if err == nil  {
                 ygot.BuildEmptyTree(capInfo)
                 capInfo.State.Name = 6
                 capInfo.State.Enabled = &enabled
             }
        } else if capName == "Repeater" {
             capInfo, err :=  ngInfo.Capabilities.NewCapability(5)
             if err == nil {
                 ygot.BuildEmptyTree(capInfo)
                 capInfo.State.Name =  5
                 capInfo.State.Enabled =  &enabled
                }
        } else if capName == "Bridge" {
                capInfo, err :=  ngInfo.Capabilities.NewCapability(3)
                if err == nil {
                    ygot.BuildEmptyTree(capInfo)
                    capInfo.State.Name =  3
                    capInfo.State.Enabled = &enabled
                }
        } else {

        }
    }
}

/** Helper function to get information from applDB **/
func (app *lldpApp) getLldpInfoFromDB(ifname *string) {

    lldpTbl, err := app.appDb.GetTable(app.neighTs)
    if err != nil {
        log.Info("Can't get lldp table")
        return
    }

    keys, err := lldpTbl.GetKeys()
    if err != nil {
        log.Info("Can't get lldp keys")
        return
    }


    for _, key := range keys {
        log.Info("lldp key = ", key.Get(0))

        lldpEntry, err := app.appDb.GetEntry(app.neighTs, db.Key{Comp: []string{key.Get(0)}})
        if err != nil {
            log.Info("can't access neighbor table for key: ", key.Get(0))
            return
        }

        if lldpEntry.IsPopulated() {
            log.Info("lldp entry populated for key: ", key.Get(0))
            app.lldpTableMap[key.Get(0)] = lldpEntry
        }
    }

    for _, key := range keys {
        if (ifname != nil && key.Get(0) != *ifname) {
            continue
        }
        entryData := app.lldpTableMap[key.Get(0)]
        if len(app.lldpNeighTableMap[key.Get(0)]) == 0 {
            app.lldpNeighTableMap[key.Get(0)] = make(map[string]string)
        }
        for lldpAttr := range entryData.Field {
            switch lldpAttr {
            case LLDP_REMOTE_CAP_ENABLED:
                app.getRemoteSysCap(entryData.Get(lldpAttr), key.Get(0), true)
            case LLDP_REMOTE_SYS_NAME:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_SYS_NAME] = entryData.Get(lldpAttr)
            case LLDP_REMOTE_PORT_DESC:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_PORT_DESC] = entryData.Get(lldpAttr)
            case LLDP_REMOTE_CHASS_ID:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_CHASS_ID] = entryData.Get(lldpAttr)
            case LLDP_REMOTE_CAP_SUPPORTED:
                app.getRemoteSysCap(entryData.Get(lldpAttr), key.Get(0), false)
            case LLDP_REMOTE_PORT_ID_SUBTYPE:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_PORT_ID_SUBTYPE] = entryData.Get(lldpAttr)
            case LLDP_REMOTE_SYS_DESC:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_SYS_DESC] = entryData.Get(lldpAttr)
            case LLDP_REMOTE_REM_TIME:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_REM_TIME] = entryData.Get(lldpAttr)
            case LLDP_REMOTE_PORT_ID:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_PORT_ID] = entryData.Get(lldpAttr)
            case LLDP_REMOTE_REM_ID:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_REM_ID] = entryData.Get(lldpAttr)
            case LLDP_REMOTE_CHASS_ID_SUBTYPE:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_CHASS_ID_SUBTYPE] = entryData.Get(lldpAttr)
            case LLDP_REMOTE_MAN_ADDR:
                app.lldpNeighTableMap[key.Get(0)][LLDP_REMOTE_MAN_ADDR] = entryData.Get(lldpAttr)
            default:
                log.Info("Unknown LLDP Attribute")
            }
        }
    }
}

/** Helper function to get remote system capabilities into a map **/
func (app *lldpApp) getRemoteSysCap(capb string, ifname string, setCap bool) {
    num_str := strings.Split(capb, " ")
    byte, _ := hex.DecodeString(num_str[0] + num_str[1])
    sysCap := byte[0]
    sysCap |= byte[1]

    log.Info("sysCap: ", sysCap)

    if (sysCap & (128 >> 1)) != 0  {
        if app.lldpCapTableMap[ifname] == nil {
            app.lldpCapTableMap[ifname] = make(map[string]bool)
            app.lldpCapTableMap[ifname]["Repeater"] = false
        }
        if (setCap) {
            log.Info("Repeater ENABLED")
            app.lldpCapTableMap[ifname]["Repeater"] = true
        }
    }

    if (sysCap & (128 >> 2)) != 0 {
        if app.lldpCapTableMap[ifname] == nil {
            app.lldpCapTableMap[ifname] = make(map[string]bool)
            app.lldpCapTableMap[ifname]["Bridge"] = false
        }
        if (setCap) {
            log.Info("Bridge  ENABLED")
            app.lldpCapTableMap[ifname]["Bridge"] = true
        }
    }

    if (sysCap & (128 >> 4)) != 0 {
        if app.lldpCapTableMap[ifname] == nil {
            app.lldpCapTableMap[ifname] = make(map[string]bool)
            app.lldpCapTableMap[ifname]["Router"] = false
        }
        if (setCap) {
            log.Info("Router ENABLED")
            app.lldpCapTableMap[ifname]["Router"] = true
        }
    }
}

