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
    "reflect"
    "strconv"
    "errors"
    "translib/db"
    "translib/ocbinds"
    "github.com/openconfig/ygot/ygot"
    log "github.com/golang/glog"
)

type PlatformApp struct {
    path        *PathInfo
    reqData     []byte
    ygotRoot    *ygot.GoStruct
    ygotTarget  *interface{}
    eepromTs    *db.TableSpec
    eepromTable map[string]dbEntry

}

func init() {
    log.Info("Init called for Platform module")
    err := register("/openconfig-platform:components",
    &appInfo{appType: reflect.TypeOf(PlatformApp{}),
    ygotRootType: reflect.TypeOf(ocbinds.OpenconfigPlatform_Components{}),
    isNative:     false})
    if err != nil {
        log.Fatal("Register Platform app module with App Interface failed with error=", err)
    }

    err = addModel(&ModelData{Name: "openconfig-platform",
    Org: "OpenConfig working group",
    Ver:      "1.0.2"})
    if err != nil {
        log.Fatal("Adding model data to appinterface failed with error=", err)
    }
}

func (app *PlatformApp) initialize(data appData) {
    log.Info("initialize:if:path =", data.path)

    app.path = NewPathInfo(data.path)
    app.reqData = data.payload
    app.ygotRoot = data.ygotRoot
    app.ygotTarget = data.ygotTarget
    app.eepromTs = &db.TableSpec{Name: "EEPROM_INFO"}

}

func (app *PlatformApp) getAppRootObject() (*ocbinds.OpenconfigPlatform_Components) {
    deviceObj := (*app.ygotRoot).(*ocbinds.Device)
    return deviceObj.Components
}

func (app *PlatformApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {

    var err error
    return nil, nil, err

}
func (app *PlatformApp) translateCreate(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys

    err = errors.New("PlatformApp Not implemented, translateCreate")
    return keys, err
}

func (app *PlatformApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys
    err = errors.New("PlatformApp Not implemented, translateUpdate")
    return keys, err
}

func (app *PlatformApp) translateReplace(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys

    err = errors.New("Not implemented PlatformApp translateReplace")
    return keys, err
}

func (app *PlatformApp) translateDelete(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys

    err = errors.New("Not implemented PlatformApp translateDelete")
    return keys, err
}

func (app *PlatformApp) translateGet(dbs [db.MaxDB]*db.DB) error  {
    var err error
    log.Info("PlatformApp: translateGet - path: ", app.path.Path)
    return err
}

func (app *PlatformApp) processCreate(d *db.DB) (SetResponse, error)  {
    var err error
    var resp SetResponse

    err = errors.New("Not implemented PlatformApp processCreate")
    return resp, err
}

func (app *PlatformApp) processUpdate(d *db.DB) (SetResponse, error)  {
    var err error
    var resp SetResponse

    err = errors.New("Not implemented PlatformApp processUpdate")
    return resp, err
}

func (app *PlatformApp) processReplace(d *db.DB) (SetResponse, error)  {
    var err error
    var resp SetResponse
    log.Info("processReplace:intf:path =", app.path)
    err = errors.New("Not implemented, PlatformApp processReplace")
    return resp, err
}

func (app *PlatformApp) processDelete(d *db.DB) (SetResponse, error)  {
    var err error
    var resp SetResponse

    err = errors.New("Not implemented PlatformApp processDelete")
    return resp, err
}

func (app *PlatformApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error)  {
    pathInfo := app.path
    log.Infof("Received GET for PlatformApp Template: %s ,path: %s, vars: %v",
    pathInfo.Template, pathInfo.Path, pathInfo.Vars)

    stateDb := dbs[db.StateDB]
    
    var payload []byte

    // Read eeprom info from DB
    app.eepromTable = make(map[string]dbEntry)

    tbl, derr := stateDb.GetTable(app.eepromTs)
    if derr != nil {
        log.Error("EEPROM_INFO table get failed!")
        return GetResponse{Payload: payload}, derr
    }

    keys, _ := tbl.GetKeys()
    for _, key := range keys {
        e, kerr := tbl.GetEntry(key)
        if kerr != nil {
            log.Error("EEPROM_INFO entry get failed!")
            return GetResponse{Payload: payload}, kerr
        }

        app.eepromTable[key.Get(0)] = dbEntry{entry: e}
    }

    targetUriPath, perr := getYangPathFromUri(app.path.Path)
    if perr != nil {
        log.Infof("getYangPathFromUri failed.")
        return GetResponse{Payload: payload}, perr
    }

    if isSubtreeRequest(targetUriPath, "/openconfig-platform:components") {
        return app.doGetSysEeprom()
    }
    err := errors.New("Not supported component")
    return GetResponse{Payload: payload}, err
}

///////////////////////////


/**
Structures to read syseeprom from redis-db
*/
type EepromDb  struct {
    Product_Name        string
    Part_Number         string
    Serial_Number       string
    Base_MAC_Address    string
    Manufacture_Date    string
    Device_Version      string
    Label_Revision      string
    Platform_Name       string
    ONIE_Version        string
    MAC_Addresses       int
    Manufacturer        string
    Manufacture_Country  string
    Vendor_Name         string
    Diag_Version        string
    Service_Tag         string
    Vendor_Extension    string
    Magic_Number        int
    Card_Type           string
    Hardware_Version    string
    Software_Version    string
    Model_Name          string

}

func (app *PlatformApp) getEepromDbObj () (EepromDb){
    log.Infof("parseEepromDb Enter")

    var eepromDbObj EepromDb

    for epItem, _ := range app.eepromTable {
        e := app.eepromTable[epItem].entry
        name := e.Get("Name")

        switch name {
        case "Device Version":
            eepromDbObj.Device_Version = e.Get("Value")
        case "Service Tag":
            eepromDbObj.Service_Tag = e.Get("Value")
        case "Vendor Extension":
            eepromDbObj.Vendor_Extension = e.Get("Value")
        case "Magic Number":
            mag, _ := strconv.ParseInt(e.Get("Value"), 10, 64)
            eepromDbObj.Magic_Number = int(mag)
        case "Card Type":
            eepromDbObj.Card_Type = e.Get("Value")
        case "Hardware Version":
            eepromDbObj.Hardware_Version = e.Get("Value")
        case "Software Version":
            eepromDbObj.Software_Version = e.Get("Value")
        case "Model Name":
            eepromDbObj.Model_Name = e.Get("Value")
        case "ONIE Version":
            eepromDbObj.ONIE_Version = e.Get("Value")
        case "Serial Number":
            eepromDbObj.Serial_Number = e.Get("Value")
        case "Vendor Name":
            eepromDbObj.Vendor_Name = e.Get("Value")
        case "Manufacturer":
            eepromDbObj.Manufacturer = e.Get("Value")
        case "Manufacture Country":
            eepromDbObj.Manufacture_Country = e.Get("Value")
        case "Platform Name":
            eepromDbObj.Platform_Name = e.Get("Value")
        case "Diag Version":
            eepromDbObj.Diag_Version = e.Get("Value")
        case "Label Revision":
            eepromDbObj.Label_Revision = e.Get("Value")
        case "Part Number":
            eepromDbObj.Part_Number = e.Get("Value")
        case "Product Name":
            eepromDbObj.Product_Name = e.Get("Value")
        case "Base MAC Address":
            eepromDbObj.Base_MAC_Address = e.Get("Value")
        case "Manufacture Date":
            eepromDbObj.Manufacture_Date = e.Get("Value")
        case "MAC Addresses":
            mac, _ := strconv.ParseInt(e.Get("Value"), 10, 16)
            eepromDbObj.MAC_Addresses = int(mac)
        }
    }

    return eepromDbObj
}

func (app *PlatformApp) getSysEepromFromDb (eeprom *ocbinds.OpenconfigPlatform_Components_Component_State, all bool) (error) {

    log.Infof("getSysEepromFromDb Enter")

    eepromDb := app.getEepromDbObj()

    empty := false
    removable := false
    name := "System Eeprom"
    location  :=  "Slot 1"

    if all == true {
        eeprom.Empty = &empty
        eeprom.Removable = &removable
        eeprom.Name = &name
        eeprom.OperStatus = ocbinds.OpenconfigPlatformTypes_COMPONENT_OPER_STATUS_ACTIVE
        eeprom.Location = &location

        if eepromDb.Product_Name != "" {
            eeprom.Id = &eepromDb.Product_Name
        }
        if eepromDb.Part_Number != "" {
            eeprom.PartNo = &eepromDb.Part_Number
        }
        if eepromDb.Serial_Number != "" {
            eeprom.SerialNo = &eepromDb.Serial_Number
        }
        if eepromDb.Base_MAC_Address != "" {
        }
        if eepromDb.Manufacture_Date != "" {
            eeprom.MfgDate = &eepromDb.Manufacture_Date
        }
        if eepromDb.Label_Revision != "" {
            eeprom.HardwareVersion = &eepromDb.Label_Revision
        }
        if eepromDb.Platform_Name != "" {
            eeprom.Description = &eepromDb.Platform_Name
        }
        if eepromDb.ONIE_Version != "" {
        }
        if eepromDb.MAC_Addresses != 0 {
        }
        if eepromDb.Manufacturer != "" {
            eeprom.MfgName = &eepromDb.Manufacturer
        }
        if eepromDb.Manufacture_Country != "" {
        }
        if eepromDb.Vendor_Name != "" {
            if eeprom.MfgName == nil {
                eeprom.MfgName = &eepromDb.Vendor_Name
            }
        }
        if eepromDb.Diag_Version != "" {
        }
        if eepromDb.Service_Tag != "" {
            if eeprom.SerialNo == nil {
                eeprom.SerialNo = &eepromDb.Service_Tag
            }
        }
        if eepromDb.Hardware_Version != "" {
            eeprom.HardwareVersion = &eepromDb.Hardware_Version
        }
        if eepromDb.Software_Version != "" {
            eeprom.SoftwareVersion = &eepromDb.Software_Version
        }
    } else {
        targetUriPath, _ := getYangPathFromUri(app.path.Path)
        switch targetUriPath {
        case "/openconfig-platform:components/component/state/name":
            eeprom.Name = &name
        case "/openconfig-platform:components/component/state/location":
            eeprom.Location = &location
        case "/openconfig-platform:components/component/state/empty":
            eeprom.Empty = &empty
        case "/openconfig-platform:components/component/state/removable":
            eeprom.Removable = &removable
        case "/openconfig-platform:components/component/state/oper-status":
            eeprom.OperStatus = ocbinds.OpenconfigPlatformTypes_COMPONENT_OPER_STATUS_ACTIVE
        case "/openconfig-platform:components/component/state/id":
            if eepromDb.Product_Name != "" {
                eeprom.Id = &eepromDb.Product_Name
            }
        case "/openconfig-platform:components/component/state/part-no":
            if eepromDb.Part_Number != "" {
                eeprom.PartNo = &eepromDb.Part_Number
            }
        case "/openconfig-platform:components/component/state/serial-no":
            if eepromDb.Serial_Number != "" {
                eeprom.SerialNo = &eepromDb.Serial_Number
            }
            if eepromDb.Service_Tag != "" {
                if eeprom.SerialNo == nil {
                    eeprom.SerialNo = &eepromDb.Service_Tag
                }
            }
        case "/openconfig-platform:components/component/state/mfg-date":
            if eepromDb.Manufacture_Date != "" {
                eeprom.MfgDate = &eepromDb.Manufacture_Date
            }
        case "/openconfig-platform:components/component/state/hardware-version":
            if eepromDb.Label_Revision != "" {
                eeprom.HardwareVersion = &eepromDb.Label_Revision
            }
            if eepromDb.Hardware_Version != "" {
                if eeprom.HardwareVersion == nil {
                    eeprom.HardwareVersion = &eepromDb.Hardware_Version
                }
            }
        case "/openconfig-platform:components/component/state/description":
            if eepromDb.Platform_Name != "" {
                eeprom.Description = &eepromDb.Platform_Name
            }
        case "/openconfig-platform:components/component/state/mfg-name":
            if eepromDb.Manufacturer != "" {
                eeprom.MfgName = &eepromDb.Manufacturer
            }
            if eepromDb.Vendor_Name != "" {
                if eeprom.MfgName == nil {
                    eeprom.MfgName = &eepromDb.Vendor_Name
                }
            }
        case "/openconfig-platform:components/component/state/software-version":
            if eepromDb.Software_Version != "" {
                eeprom.SoftwareVersion = &eepromDb.Software_Version
            }
        }
    }
    return nil 
}

func (app *PlatformApp) doGetSysEeprom () (GetResponse, error) {

    log.Infof("Preparing collection for system eeprom");

    var payload []byte
    var err error
    pf_cpts := app.getAppRootObject()

    targetUriPath, _ := getYangPathFromUri(app.path.Path)
    switch targetUriPath {
    case "/openconfig-platform:components":
        pf_comp,_ := pf_cpts.NewComponent("System Eeprom")
        ygot.BuildEmptyTree(pf_comp)
        err = app.getSysEepromFromDb(pf_comp.State, true)
        if err != nil {
            return GetResponse{Payload: payload}, err
        }
        payload, err = dumpIetfJson((*app.ygotRoot).(*ocbinds.Device), true)
    case "/openconfig-platform:components/component":
        compName := app.path.Var("name")
        if compName == "" {
            pf_comp,_ := pf_cpts.NewComponent("System Eeprom")
            ygot.BuildEmptyTree(pf_comp)
            err = app.getSysEepromFromDb(pf_comp.State, true)
            if err != nil {
                return GetResponse{Payload: payload}, err
            }
            payload, err = dumpIetfJson(pf_cpts, false)
        } else {
            if compName != "System Eeprom" {
                err = errors.New("Invalid component name")
            }
            pf_comp := pf_cpts.Component[compName]
            if pf_comp != nil {
                ygot.BuildEmptyTree(pf_comp)
                err = app.getSysEepromFromDb(pf_comp.State, true)
                if err != nil {
                    return GetResponse{Payload: payload}, err
                }
                payload, err = dumpIetfJson(pf_cpts.Component[compName], false)
            } else {
                err = errors.New("Invalid input component name")
            }
        }
    case "/openconfig-platform:components/component/state":
        compName := app.path.Var("name")
        if compName != "" && compName == "System Eeprom" {
            pf_comp := pf_cpts.Component[compName]
            if pf_comp != nil {
                ygot.BuildEmptyTree(pf_comp)
                err = app.getSysEepromFromDb(pf_comp.State, true)
                if err != nil {
                    return GetResponse{Payload: payload}, err
                }
                payload, err = dumpIetfJson(pf_cpts.Component[compName], false)
            } else {
                err = errors.New("Invalid input component name")
            }
        } else {
            err = errors.New("Invalid component name ")
        }

    default:
        if isSubtreeRequest(targetUriPath, "/openconfig-platform:components/component/state") {
            compName := app.path.Var("name")
            if compName == "" || compName != "System Eeprom" {
                err = errors.New("Invalid input component name")
            } else {
                pf_comp := pf_cpts.Component[compName]
                if pf_comp != nil {
                    ygot.BuildEmptyTree(pf_comp)
                    err = app.getSysEepromFromDb(pf_comp.State, false)
                    if err != nil {
                        return GetResponse{Payload: payload}, err
                    }
                    payload, err = dumpIetfJson(pf_cpts.Component[compName].State, false)
                } else {
                    err = errors.New("Invalid input component name")
                }
            }
        } else {
            err = errors.New("Invalid Path")
        }
    }
    return  GetResponse{Payload: payload}, err
}

