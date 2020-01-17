////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Dell, Inc.                                                 //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//  http://www.apache.org/licenses/LICENSE-2.0                                //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package transformer

import (
    "encoding/json"
    "errors"
    "strconv"
    "translib/db"
    "translib/ocbinds"
    "translib/tlerr"
    "strings"
	"github.com/openconfig/ygot/ygot"
    log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("YangToDb_lag_min_links_xfmr", YangToDb_lag_min_links_xfmr)
    XlateFuncBind("YangToDb_lag_fallback_xfmr", YangToDb_lag_fallback_xfmr)
    XlateFuncBind("DbToYang_intf_lag_state_xfmr", DbToYang_intf_lag_state_xfmr)
    XlateFuncBind("YangToDb_lag_type_xfmr", YangToDb_lag_type_xfmr)
    XlateFuncBind("DbToYang_lag_type_xfmr", DbToYang_lag_type_xfmr)
}

const (
       LAG_TYPE           = "lag-type"
       PORTCHANNEL_TABLE  = "PORTCHANNEL"
)

var LAG_TYPE_MAP = map[string]string{
    strconv.FormatInt(int64(ocbinds.OpenconfigIfAggregate_AggregationType_LACP), 10): "false",
    strconv.FormatInt(int64(ocbinds.OpenconfigIfAggregate_AggregationType_STATIC), 10): "true",
}


/* Validate whether LAG exists in DB */
func validateLagExists(d *db.DB, lagTs *string, lagName *string) error {
    if len(*lagName) == 0 {
        return errors.New("Length of PortChannel name is zero")
    }
    entry, err := d.GetEntry(&db.TableSpec{Name:*lagTs}, db.Key{Comp: []string{*lagName}})
    if err != nil || !entry.IsPopulated() {
        errStr := "Invalid PortChannel:" + *lagName
        return errors.New(errStr)
    }
    return nil
}

func get_min_links(d *db.DB, lagName *string, links *uint16) error {
    intTbl := IntfTypeTblMap[IntfTypePortChannel]
    curr, err := d.GetEntry(&db.TableSpec{Name:intTbl.cfgDb.portTN}, db.Key{Comp: []string{*lagName}})
    if err != nil {
        errStr := "Failed to Get PortChannel details"
        log.Info(errStr)
        return errors.New(errStr)
    }
    if val, ok := curr.Field["min_links"]; ok {
        min_links, err := strconv.ParseUint(val, 10, 16)
        if err != nil {
            errStr := "Conversion of string to int failed: " + val
            log.Info(errStr)
            return errors.New(errStr)
        }
        *links = uint16(min_links)
    } else {
        log.Info("Minlinks set to 1 (dafault value)")
        *links = 1
    }
    log.Infof("Got min links from DB : %d\n", *links)
    return nil
}


func get_lag_type(d *db.DB, lagName *string, mode *string) error {
    intTbl := IntfTypeTblMap[IntfTypePortChannel]
    curr, err := d.GetEntry(&db.TableSpec{Name:intTbl.cfgDb.portTN}, db.Key{Comp: []string{*lagName}})
    if err != nil {
        errStr := "Failed to Get PortChannel details"
        log.Info(errStr)
        return errors.New(errStr)
    }
    if val, ok := curr.Field["static"]; ok {
        *mode = val
        log.Infof("Mode from DB: %s\n", *mode)
    } else {
        log.Info("Default LACP mode (static false)")
        *mode = "false"
        log.Infof("Default Mode: %s\n", *mode)
    }
    return nil
}

/* Handle min-links config */
var YangToDb_lag_min_links_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    ifKey := pathInfo.Var("name")

    log.Infof("Received Min links config for path: %s; template: %s vars: %v ifKey: %s", pathInfo.Path, pathInfo.Template, pathInfo.Vars, ifKey)

    var links uint16
    err = get_min_links(inParams.d, &ifKey, &links)

    if err == nil && links != *(inParams.param.(*uint16))  {
        errStr := "Cannot reconfigure min links for an existing PortChannel: " + ifKey
        log.Info(errStr)
        err = tlerr.InvalidArgsError{Format: errStr}
        return res_map, err
    }

    minLinks, _ := inParams.param.(*uint16)
    res_map["min_links"] = strconv.Itoa(int(*minLinks))
    return res_map, nil
}

func can_configure_fallback(inParams XfmrParams) error {
    device := (*inParams.ygRoot).(*ocbinds.Device)
    user_config_json, e := ygot.EmitJSON(device, &ygot.EmitJSONConfig{
    Format: ygot.RFC7951,
    Indent: "  ",
    RFC7951Config: &ygot.RFC7951JSONConfig{
        AppendModuleName: true,
    }})

    if e != nil {
        log.Infof("EmitJSON error: %v", e)
        return e
    }

    type intf struct {
        IntfS map[string]interface{}  `json:"openconfig-interfaces:interfaces"`
    }
    var res intf
    e = json.Unmarshal([]byte(user_config_json), &res)
    if e != nil {
        log.Infof("UnMarshall Error %v\n", e)
        return e
    }

    i := res.IntfS["interface"].([]interface{})
    po_map := i[0].(map[string]interface{})

    if agg,ok := po_map["openconfig-if-aggregate:aggregation"]; ok {
       a := agg.(map[string]interface{})
       agg_conf := a["config"].(map[string]interface{})
       if lag_type,k := agg_conf["lag-type"]; k {
           if lag_type == "STATIC" {
               errStr := "Fallback is not supported for Static LAGs"
               return tlerr.InvalidArgsError{Format:errStr}
           }
       } else {
           //User did not specify LAG Type; Check from DB
           pathInfo := NewPathInfo(inParams.uri)
           ifKey := pathInfo.Var("name")
           var mode string
           e = get_lag_type(inParams.d, &ifKey, &mode)
           if e == nil {
               errStr := "Fallback option cannot be re-configured for an already existing PortChannel: " + ifKey
               return tlerr.InvalidArgsError{Format:errStr}
           }
       }
    }

    return nil
}

/* Handle fallback config */
var YangToDb_lag_fallback_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)
    var err error

    err = can_configure_fallback(inParams)
    if err != nil {
        return res_map, err
    }

    fallback, _ := inParams.param.(*bool)
    res_map["fallback"] = strconv.FormatBool(*fallback)
    return res_map, nil
}

func getLagStateAttr(attr *string, ifName *string, lagInfoMap  map[string]db.Value,
                          oc_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_State) (error) {
    lagEntries, ok := lagInfoMap[*ifName]
    if !ok {
        errStr := "Cannot find info for Interface: " + *ifName
        return errors.New(errStr)
    }
    switch *attr {
    case "mode":
        oc_val.LagType = ocbinds.OpenconfigIfAggregate_AggregationType_LACP

        lag_type,ok := lagEntries.Field["static"]
        if ok {
            if lag_type == "true" {
                oc_val.LagType = ocbinds.OpenconfigIfAggregate_AggregationType_STATIC
            }
        }
    case "min-links":
        links, _ := strconv.Atoi(lagEntries.Field["min-links"])
        minlinks := uint16(links)
        oc_val.MinLinks = &minlinks
    case "fallback":
        fallbackVal, _:= strconv.ParseBool(lagEntries.Field["fallback"])
        oc_val.Fallback = &fallbackVal
    case "member":
        lagMembers := strings.Split(lagEntries.Field["member@"], ",")
        oc_val.Member = lagMembers
    }
    return nil
}

func getLagState(ifName *string, lagInfoMap  map[string]db.Value,
                          oc_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_State) (error) {
    log.Info("getIntfVlanAttr() called")
    lagEntries, ok := lagInfoMap[*ifName]
    if !ok {
        errStr := "Cannot find info for Interface: " + *ifName
        return errors.New(errStr)
    }
    links, _ := strconv.Atoi(lagEntries.Field["min-links"])
    minlinks := uint16(links)
    oc_val.MinLinks = &minlinks
    fallbackVal, _:= strconv.ParseBool(lagEntries.Field["fallback"])
    oc_val.Fallback = &fallbackVal

    oc_val.LagType = ocbinds.OpenconfigIfAggregate_AggregationType_LACP
    lag_type,ok := lagEntries.Field["static"]
    if ok {
        if lag_type == "true" {
            oc_val.LagType = ocbinds.OpenconfigIfAggregate_AggregationType_STATIC
        }
    }

    lagMembers := strings.Split(lagEntries.Field["member@"], ",")
    oc_val.Member = lagMembers
    return nil
}

/* Get PortChannel Info */
func fillLagInfoForIntf(d *db.DB, ifName *string, lagInfoMap map[string]db.Value) error {
    var err error
    var lagMemKeys []db.Key
    intTbl := IntfTypeTblMap[IntfTypePortChannel]
    /* Get members list */
    lagMemKeys, err = d.GetKeys(&db.TableSpec{Name:intTbl.cfgDb.memberTN})
    if err != nil {
        return err
    }
    log.Infof("Found %d lag-member-table keys", len(lagMemKeys))
    log.Infof("lag-member-table keys", lagMemKeys)
    var lagMembers []string
    var memberPortsStr strings.Builder
    for i, _ := range lagMemKeys {
        if *ifName == lagMemKeys[i].Get(0) {
            log.Info("Found member")
            ethName := lagMemKeys[i].Get(1)
            lagMembers = append(lagMembers, ethName)
            memberPortsStr.WriteString(ethName + ",")
        }
    }
    lagInfoMap[*ifName] = db.Value{Field:make(map[string]string)}
    lagInfoMap[*ifName].Field["member@"] = strings.Join(lagMembers, ",")
    /* Get MinLinks value */
    curr, err := d.GetEntry(&db.TableSpec{Name:intTbl.cfgDb.portTN}, db.Key{Comp: []string{*ifName}})
    if err != nil {
        errStr := "Failed to Get PortChannel details"
        return errors.New(errStr)
    }
    var links int
    if val, ok := curr.Field["min_links"]; ok {
        min_links, err := strconv.Atoi(val)
        if err != nil {
            errStr := "Conversion of string to int failed"
            return errors.New(errStr)
        }
        links = min_links
    } else {
        log.Info("Minlinks set to 0 (dafault value)")
        links = 0
    }
    lagInfoMap[*ifName].Field["min-links"] = strconv.Itoa(links)
    /* Get fallback value */
    var fallbackVal string
    if val, ok := curr.Field["fallback"]; ok {
        fallbackVal = val
        if err != nil {
            errStr := "Conversion of string to bool failed"
            return errors.New(errStr)
        }
    } else {
        log.Info("Fallback set to False, default value")
        fallbackVal = "false"
    }
    lagInfoMap[*ifName].Field["fallback"] = fallbackVal

    if v, k := curr.Field["static"]; k {
        lagInfoMap[*ifName].Field["static"] = v
    } else {
        log.Info("Mode set to LACP, default value")
        lagInfoMap[*ifName].Field["static"] = "false"
    }
    log.Infof("Updated the lag-info-map for Interface: %s", *ifName)

    return err
}

/* PortChannel GET operation */
var DbToYang_intf_lag_state_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) (error) {
    var err error

    intfsObj := getIntfsRoot(inParams.ygRoot)
    if intfsObj == nil || intfsObj.Interface == nil {
        errStr := "Failed to Get root object!"
        log.Errorf(errStr)
        return errors.New(errStr)
    }
    pathInfo := NewPathInfo(inParams.uri)
    ifName := pathInfo.Var("name")
    if _, ok := intfsObj.Interface[ifName]; !ok  {
		obj, _ := intfsObj.NewInterface(ifName)
		ygot.BuildEmptyTree(obj)
	}
    intfObj := intfsObj.Interface[ifName]
    if intfObj.Aggregation == nil {
        return errors.New("Not a valid request")
    }
    if intfObj.Aggregation.State == nil {
        return errors.New("Not a valid PortChannel Get request")
    }
    intfType, _, err := getIntfTypeByName(ifName)
    if intfType != IntfTypePortChannel || err != nil {
        intfTypeStr := strconv.Itoa(int(intfType))
        errStr := "TableXfmrFunc - Invalid interface type" + intfTypeStr
        log.Error(errStr);
        return errors.New(errStr);
    }
    intTbl := IntfTypeTblMap[IntfTypePortChannel]
    /*Validate given PortChannel exists */
    err = validateLagExists(inParams.d, &intTbl.cfgDb.portTN, &ifName)
    if err != nil {
        errStr := "Invalid PortChannel: " + ifName
        err = tlerr.InvalidArgsError{Format: errStr}
        return err
    }

    targetUriPath, err := getYangPathFromUri(inParams.uri)
    log.Info("targetUriPath is ", targetUriPath)
    lagInfoMap := make(map[string]db.Value)
    ocAggregationStateVal := intfObj.Aggregation.State
    err = fillLagInfoForIntf(inParams.d, &ifName, lagInfoMap)
    if err != nil {
        log.Errorf("Failed to get info: %s failed!", ifName)
        return err
    }
    log.Info("Succesfully completed DB map population!", lagInfoMap)
    switch targetUriPath {
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/state/min-links":
        log.Info("Get is for min-links")
        attr := "min-links"
        err = getLagStateAttr(&attr, &ifName, lagInfoMap, ocAggregationStateVal)
        if err != nil {
            return err
        }
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/state/lag-type":
         log.Info("Get is for lag type")
         attr := "mode"
         err = getLagStateAttr(&attr, &ifName, lagInfoMap, ocAggregationStateVal)
         if err != nil {
             return err
         }
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/state/openconfig-interfaces-ext:fallback":
        log.Info("Get is for fallback")
        attr := "fallback"
        err = getLagStateAttr(&attr, &ifName, lagInfoMap, ocAggregationStateVal)
        if err != nil {
            return err
        }
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/state/member":
        log.Info("Get is for member")
        attr := "member"
        err = getLagStateAttr(&attr, &ifName, lagInfoMap, ocAggregationStateVal)
        if err != nil {
            return err
        }
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/state":
        log.Info("Get is for State Container!")
        err = getLagState(&ifName, lagInfoMap, ocAggregationStateVal)
        if err != nil {
            return err
        }
    default:
        log.Infof(targetUriPath + " - Not an supported Get attribute")
    }
    return err
}

/* Function to delete PortChannel and all its member ports */
func deleteLagIntfAndMembers(inParams *XfmrParams, lagName *string) error {
    log.Info("Inside deleteLagIntfAndMembers")
    var err error

    subOpMap := make(map[db.DBNum]map[string]map[string]db.Value)
    resMap := make(map[string]map[string]db.Value)
    lagMap := make(map[string]db.Value)
    lagMemberMap := make(map[string]db.Value)
    lagMap[*lagName] = db.Value{Field:map[string]string{}}

    intTbl := IntfTypeTblMap[IntfTypePortChannel]
    subOpMap[db.ConfigDB] = resMap
    inParams.subOpDataMap[DELETE] = &subOpMap
    /* Validate given PortChannel exits */
    err = validateLagExists(inParams.d, &intTbl.cfgDb.portTN, lagName)
    if err != nil {
        errStr := "PortChannel does not exist: " + *lagName
        log.Error(errStr)
        return errors.New(errStr)
    }

    /* Handle PORTCHANNEL_INTERFACE TABLE */
    err = validateL3ConfigExists(inParams.d, lagName)
    if err != nil {
        return err
    }

    /* Handle PORTCHANNEL_MEMBER TABLE */
    var flag bool = false
    lagKeys, err := inParams.d.GetKeys(&db.TableSpec{Name:intTbl.cfgDb.memberTN})
    if err == nil {
        for key, _ := range lagKeys {
            if *lagName == lagKeys[key].Get(0) {
                flag = true
                log.Info("Member port", lagKeys[key].Get(1))
                memberKey := *lagName + "|" + lagKeys[key].Get(1)
                lagMemberMap[memberKey] = db.Value{Field:map[string]string{}}
            }
        }
        if flag == true {
            resMap["PORTCHANNEL_MEMBER"] = lagMemberMap
        }
    }
    /* Handle PORTCHANNEL TABLE */
    resMap["PORTCHANNEL"] = lagMap
    subOpMap[db.ConfigDB] = resMap
    log.Info("subOpMap: ", subOpMap)
    inParams.subOpDataMap[DELETE] = &subOpMap
    return nil
}


var YangToDb_lag_type_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    result := make(map[string]string)
    var err error

    if inParams.param == nil {
        return result, err
    }

    pathInfo := NewPathInfo(inParams.uri)
    ifKey := pathInfo.Var("name")

    log.Infof("Received Mode configuration for path: %s; template: %s vars: %v ifKey: %s", pathInfo.Path, pathInfo.Template, pathInfo.Vars, ifKey)

    var mode string
    err = get_lag_type(inParams.d, &ifKey, &mode)

    t, _ := inParams.param.(ocbinds.E_OpenconfigIfAggregate_AggregationType)
    user_mode := findInMap(LAG_TYPE_MAP, strconv.FormatInt(int64(t), 10))

    if err == nil && mode != user_mode  {
        errStr := "Cannot reconfigure Mode for an existing PortChannel: " + ifKey
        err = tlerr.InvalidArgsError{Format: errStr}
        return result, err
    }

    log.Info("YangToDb_lag_type_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " type: ", t)
    result["static"] = user_mode
    return result, nil

}

var DbToYang_lag_type_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})


    intfType, _, ierr := getIntfTypeByName(inParams.key)
    if ierr != nil || intfType != IntfTypePortChannel  {
        return result, err
    }


    data := (*inParams.dbDataMap)[inParams.curDb]
    var agg_type ocbinds.E_OpenconfigIfAggregate_AggregationType
    agg_type = ocbinds.OpenconfigIfAggregate_AggregationType_LACP

    lag_type,ok := data[PORTCHANNEL_TABLE][inParams.key].Field["static"]
    if ok {
        if lag_type == "true" {
            agg_type = ocbinds.OpenconfigIfAggregate_AggregationType_STATIC
        }
    }
        result[LAG_TYPE] = ocbinds.E_OpenconfigIfAggregate_AggregationType.ΛMap(agg_type)["E_OpenconfigIfAggregate_AggregationType"][int64(agg_type)].Name
    log.Infof("Lag Type returned from Field Xfmr: %v\n", result)
    return result, err
}

