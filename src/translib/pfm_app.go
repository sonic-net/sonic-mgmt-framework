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
    "encoding/json"
    "errors"
    "translib/db"
    "translib/ocbinds"
    "github.com/openconfig/ygot/ygot"
    "os"
    "translib/tlerr"
    "io/ioutil"
    log "github.com/golang/glog"
)

type PlatformApp struct {
    path        *PathInfo
    reqData     []byte
    ygotRoot    *ygot.GoStruct
    ygotTarget  *interface{}


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
    var payload []byte
    log.Infof("Received GET for PlatformApp Template: %s ,path: %s, vars: %v",
    pathInfo.Template, pathInfo.Path, pathInfo.Vars)
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
func (app *PlatformApp) doGetSysEeprom() (GetResponse, error) {

    return app.getSysEepromJson()
}


/**
Structures to read syseeprom from json file
*/
type JSONEeprom  struct {
    Product_Name        string `json:"Product Name"`
    Part_Number         string `json:"Part Number"`
    Serial_Number       string `json:"Serial Number"`
    Base_MAC_Address    string `json:"Base MAC Address"`
    Manufacture_Date    string `json:"Manufacture Date"`
    Device_Version      string `json:"Device Version"`
    Label_Revision      string `json:"Label Revision"`
    Platform_Name       string `json:"Platform Name"`
    ONIE_Version        string `json:"ONIE Version"`
    MAC_Addresses       int    `json:"MAC Addresses"`
    Manufacturer        string `json:"Manufacturer"`
    Manufacture_Country  string `json:"Manufacture Country"`
    Vendor_Name         string `json:"Vendor Name"`
    Diag_Version        string `json:"Diag Version"`
    Service_Tag         string `json:"Service Tag"`
    Vendor_Extension    string `json:"Vendor Extension"`
    Magic_Number        int    `json:"Magic Number"`
    Card_Type           string `json:"Card Type"`
    Hardware_Version    string `json:"Hardware Version"`
    Software_Version    string `json:"Software Version"`
    Model_Name          string `json:"Model Name"`

}

func (app *PlatformApp) getSysEepromFromFile (eeprom *ocbinds.OpenconfigPlatform_Components_Component_State, all bool) (error) {

    log.Infof("getSysEepromFromFile Enter")
    jsonFile, err := os.Open("/mnt/platform/syseeprom")
    if err != nil {
        log.Infof("syseeprom.json open failed")
        errStr := "Information not available or Not supported"
        terr := tlerr.NotFoundError{Format: errStr}
        return terr
    }

    defer jsonFile.Close()

    byteValue, _ := ioutil.ReadAll(jsonFile)
    var jsoneeprom JSONEeprom

    json.Unmarshal(byteValue, &jsoneeprom)
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

        if jsoneeprom.Product_Name != "" {
            eeprom.Id = &jsoneeprom.Product_Name
        }
        if jsoneeprom.Part_Number != "" {
            eeprom.PartNo = &jsoneeprom.Part_Number
        }
        if jsoneeprom.Serial_Number != "" {
            eeprom.SerialNo = &jsoneeprom.Serial_Number
        }
        if jsoneeprom.Base_MAC_Address != "" {
        }
        if jsoneeprom.Manufacture_Date != "" {
            eeprom.MfgDate = &jsoneeprom.Manufacture_Date
        }
        if jsoneeprom.Label_Revision != "" {
            eeprom.HardwareVersion = &jsoneeprom.Label_Revision
        }
        if jsoneeprom.Platform_Name != "" {
            eeprom.Description = &jsoneeprom.Platform_Name
        }
        if jsoneeprom.ONIE_Version != "" {
        }
        if jsoneeprom.MAC_Addresses != 0 {
        }
        if jsoneeprom.Manufacturer != "" {
            eeprom.MfgName = &jsoneeprom.Manufacturer
        }
        if jsoneeprom.Manufacture_Country != "" {
        }
        if jsoneeprom.Vendor_Name != "" {
            if eeprom.MfgName == nil {
                eeprom.MfgName = &jsoneeprom.Vendor_Name
            }
        }
        if jsoneeprom.Diag_Version != "" {
        }
        if jsoneeprom.Service_Tag != "" {
            if eeprom.SerialNo == nil {
                eeprom.SerialNo = &jsoneeprom.Service_Tag
            }
        }
        if jsoneeprom.Hardware_Version != "" {
            eeprom.HardwareVersion = &jsoneeprom.Hardware_Version
        }
        if jsoneeprom.Software_Version != "" {
            eeprom.SoftwareVersion = &jsoneeprom.Software_Version
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
            if jsoneeprom.Product_Name != "" {
                eeprom.Id = &jsoneeprom.Product_Name
            }
        case "/openconfig-platform:components/component/state/part-no":
            if jsoneeprom.Part_Number != "" {
                eeprom.PartNo = &jsoneeprom.Part_Number
            }
        case "/openconfig-platform:components/component/state/serial-no":
            if jsoneeprom.Serial_Number != "" {
                eeprom.SerialNo = &jsoneeprom.Serial_Number
            }
            if jsoneeprom.Service_Tag != "" {
                if eeprom.SerialNo == nil {
                    eeprom.SerialNo = &jsoneeprom.Service_Tag
                }
            }
        case "/openconfig-platform:components/component/state/mfg-date":
            if jsoneeprom.Manufacture_Date != "" {
                eeprom.MfgDate = &jsoneeprom.Manufacture_Date
            }
        case "/openconfig-platform:components/component/state/hardware-version":
            if jsoneeprom.Label_Revision != "" {
                eeprom.HardwareVersion = &jsoneeprom.Label_Revision
            }
            if jsoneeprom.Hardware_Version != "" {
                if eeprom.HardwareVersion == nil {
                    eeprom.HardwareVersion = &jsoneeprom.Hardware_Version
                }
            }
        case "/openconfig-platform:components/component/state/description":
            if jsoneeprom.Platform_Name != "" {
                eeprom.Description = &jsoneeprom.Platform_Name
            }
        case "/openconfig-platform:components/component/state/mfg-name":
            if jsoneeprom.Manufacturer != "" {
                eeprom.MfgName = &jsoneeprom.Manufacturer
            }
            if jsoneeprom.Vendor_Name != "" {
                if eeprom.MfgName == nil {
                    eeprom.MfgName = &jsoneeprom.Vendor_Name
                }
            }
        case "/openconfig-platform:components/component/state/software-version":
            if jsoneeprom.Software_Version != "" {
                eeprom.SoftwareVersion = &jsoneeprom.Software_Version
            }
        }
    }
    return nil 
}

func (app *PlatformApp) getSysEepromJson () (GetResponse, error) {

    log.Infof("Preparing json for system eeprom");

    var payload []byte
    var err error
    pf_cpts := app.getAppRootObject()

    targetUriPath, _ := getYangPathFromUri(app.path.Path)
    switch targetUriPath {
    case "/openconfig-platform:components":
        pf_comp,_ := pf_cpts.NewComponent("System Eeprom")
        ygot.BuildEmptyTree(pf_comp)
        err = app.getSysEepromFromFile(pf_comp.State, true)
        if err != nil {
            return GetResponse{Payload: payload}, err
        }
        payload, err = dumpIetfJson((*app.ygotRoot).(*ocbinds.Device), true)
    case "/openconfig-platform:components/component":
        compName := app.path.Var("name")
        if compName == "" {
            pf_comp,_ := pf_cpts.NewComponent("System Eeprom")
            ygot.BuildEmptyTree(pf_comp)
            err = app.getSysEepromFromFile(pf_comp.State, true)
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
                err = app.getSysEepromFromFile(pf_comp.State, true)
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
                err = app.getSysEepromFromFile(pf_comp.State, true)
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
                    err = app.getSysEepromFromFile(pf_comp.State, false)
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

