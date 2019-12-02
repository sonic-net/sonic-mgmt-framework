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
    "errors"
    "strconv"
    "translib/db"
    "translib/ocbinds"
    "translib/tlerr"
    "reflect"
    "strings"
    log "github.com/golang/glog"
)

type intfModeType int

const (
  MODE_UNSET intfModeType = iota
  ACCESS
  TRUNK
)

type intfModeReq struct {
  ifName string
  mode   intfModeType
}

type ifVlan struct {
    ifName     *string
    mode       intfModeType
    accessVlan *string
    trunkVlans []string
}

type swVlanMemberPort_t struct {
  swEthMember *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan
  swPortChannelMember *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan
}

func init () {
    XlateFuncBind("YangToDb_sw_vlans_xfmr", YangToDb_sw_vlans_xfmr)
    XlateFuncBind("DbToYang_sw_vlans_xfmr", DbToYang_sw_vlans_xfmr)
}

/* Validate whether VLAN exists in DB */
func validateVlanExists(d *db.DB, vlanName *string) error {
    if len(*vlanName) == 0 {
        return errors.New("Length of VLAN name is zero")
    }
    entry, err := d.GetEntry(&db.TableSpec{Name:VLAN_TN}, db.Key{Comp: []string{*vlanName}})
    if err != nil || !entry.IsPopulated() {
        errStr := "Invalid Vlan:" + *vlanName
    log.Error(errStr)
        return errors.New(errStr)
    }
    return nil
}

/* Generate Member Ports string from Slice to update VLAN table in CONFIG DB */
func generateMemberPortsStringFromSlice(memberPortsList []string) *string {
    if len(memberPortsList) == 0 {
        return nil
    }
    var memberPortsStr strings.Builder
    idx := 1

    for _, memberPort := range memberPortsList {
        if idx != len(memberPortsList) {
            memberPortsStr.WriteString(memberPort + ",")
        } else {
            memberPortsStr.WriteString(memberPort)
        }
        idx = idx + 1
    }
    memberPorts := memberPortsStr.String()
    return &(memberPorts)
}

/* Generate list of member-ports from string */
func generateMemberPortsSliceFromString(memberPortsStr *string) []string {
    if len(*memberPortsStr) == 0 {
        return nil
    }
    memberPorts := strings.Split(*memberPortsStr, ",")
    return memberPorts
}

/* Check member port exists in the list and get Interface mode */
func checkMemberPortExistsInListAndGetMode(d *db.DB, memberPortsList []string, memberPort *string, vlanName *string, ifMode *intfModeType) bool {
    for _, port := range memberPortsList {
        if *memberPort == port {
            tagModeEntry, err := d.GetEntry(&db.TableSpec{Name: VLAN_MEMBER_TN}, db.Key{Comp: []string{*vlanName, *memberPort}})
            if err != nil {
                return false
            }
            tagMode := tagModeEntry.Field["tagging_mode"]
            convertTaggingModeToInterfaceModeType(&tagMode, ifMode)
            return true
        }
    }
    return false
}

/* Convert tagging mode to Interface Mode type */
func convertTaggingModeToInterfaceModeType(tagMode *string, ifMode *intfModeType) {
    switch *tagMode {
    case "untagged":
        *ifMode = ACCESS
    case "tagged":
        *ifMode = TRUNK
    }
}

/* Validate whether Port has any Untagged VLAN Config existing */
func validateUntaggedVlanCfgredForIf(d *db.DB, vlanMemberTs *string, ifName *string, accessVlan *string) (bool, error) {
    var err error

    var vlanMemberKeys []db.Key
    vlanMemberTable, err := d.GetTable(&db.TableSpec{Name:*vlanMemberTs})
    if err != nil {
        return false, err
    }

    vlanMemberKeys, err = vlanMemberTable.GetKeys()
    log.Infof("Found %d Vlan Member table keys", len(vlanMemberKeys))

    for _, vlanMember := range vlanMemberKeys {
        if len(vlanMember.Comp) < 2 {
            continue
        }
        if vlanMember.Get(1) != *ifName {
            continue
        }
        memberPortEntry, err := d.GetEntry(&db.TableSpec{Name:*vlanMemberTs}, vlanMember)
        if err != nil || !memberPortEntry.IsPopulated() {
            errStr := "Get from VLAN_MEMBER table for Vlan: + " + vlanMember.Get(0) + " Interface:" + *ifName + " failed!"
      log.Error(errStr)
            return false, errors.New(errStr)
        }
        tagMode, ok := memberPortEntry.Field["tagging_mode"]
        if !ok {
            errStr := "tagging_mode entry is not present for VLAN: " + vlanMember.Get(0) + " Interface: " + *ifName
      log.Error(errStr)
            return false, errors.New(errStr)
        }
        if tagMode == "untagged" {
            *accessVlan = vlanMember.Get(0)
            return true, nil
        }
    }
    return false, nil
}

/* Removes the Interface name from Members list of VLAN table and updates it */
func removeFromMembersListForVlan(d *db.DB, vlan *string, ifName *string, vlanMap map[string]db.Value) error {

    vlanEntry, err := d.GetEntry(&db.TableSpec{Name:VLAN_TN}, db.Key{Comp: []string{*vlan}})
    if err != nil {
        log.Errorf("Get Entry for VLAN table with Vlan:%s failed!", *vlan)
        return err
    }
    memberPortsInfo, ok := vlanEntry.Field["members@"]
    if ok {
        memberPortsList := generateMemberPortsSliceFromString(&memberPortsInfo)
        if memberPortsList == nil {
            return nil
        }
        memberFound := false

        for _, memberName := range memberPortsList {
            if memberName == *ifName {
                memberFound = true
                break
            }
        }
        if memberFound {
      updatedVlanEntry := db.Value{Field:make(map[string]string)}
            updatedVlanEntry.Field["members@"] = *ifName
            vlanMap[*vlan] = updatedVlanEntry
        } else {
            return nil
        }
    }
    return nil
}

/* Removes Interface name from Members-list for all VLANs from VLAN table and updates it */
func removeFromMembersListForAllVlans(d *db.DB, ifName *string, vlanMemberMap map[string]db.Value,
                                      vlanMap map[string]db.Value) error {
  var err error

  for vlan, _ := range vlanMemberMap {
    err = removeFromMembersListForVlan(d, &vlan, ifName, vlanMap)
    if err != nil {
      return err
    }
  }
  return err
}

/* Remove tagged port associated with VLAN and update VLAN_MEMBER table */
func removeTaggedVlanAndUpdateVlanMembTbl(d *db.DB, trunkVlan *string, ifName *string, vlanMemberMap map[string]db.Value) error {
    var err error
    memberPortEntry, err := d.GetEntry(&db.TableSpec{Name:VLAN_MEMBER_TN}, db.Key{Comp: []string{*trunkVlan, *ifName}})
    if err != nil || !memberPortEntry.IsPopulated() {
        errStr := "Trunk Vlan: " + *trunkVlan + " not configured for Interface: " + *ifName
        return errors.New(errStr)
    }
    tagMode, ok := memberPortEntry.Field["tagging_mode"]
    if !ok {
        errStr := "tagging_mode entry is not present for VLAN: " + *trunkVlan + " Interface: " + *ifName
        return errors.New(errStr)
    }
    vlanName := *trunkVlan
    if tagMode == "tagged" {
        vlanMemberKey := *trunkVlan + "|" + *ifName
        vlanMemberMap[vlanMemberKey] = memberPortEntry
    } else {
        vlanId := vlanName[len("Vlan"):len(vlanName)]
        errStr := "Tagged VLAN: " + vlanId + " configuration doesn't exist for Interface: " + *ifName
        return tlerr.InvalidArgsError{Format: errStr}
    }
    return err
}

/* Remove untagged port associated with VLAN and update VLAN_MEMBER table */
func removeUntaggedVlanAndUpdateVlanMembTbl(d *db.DB, ifName *string, vlanMemberMap map[string]db.Value) (*string, error) {
    if len(*ifName) == 0 {
        return nil, errors.New("Interface name is empty for fetching list of VLANs!")
    }

    var vlanMemberKeys []db.Key
    vlanMemberTable, err := d.GetTable(&db.TableSpec{Name:VLAN_MEMBER_TN})
    if err != nil {
        return nil, err
    }

    vlanMemberKeys, err = vlanMemberTable.GetKeys()
    log.Infof("Found %d Vlan Member table keys", len(vlanMemberKeys))

    for _, vlanMember := range vlanMemberKeys {
        if len(vlanMember.Comp) < 2 {
            continue
        }
        if vlanMember.Get(1) != *ifName {
            continue
        }
        memberPortEntry, err := d.GetEntry(&db.TableSpec{Name: VLAN_MEMBER_TN}, vlanMember)
        if err != nil || !memberPortEntry.IsPopulated() {
            errStr := "Get from VLAN_MEMBER table for Vlan: + " + vlanMember.Get(0) + " Interface:" + *ifName + " failed!"
            return nil, errors.New(errStr)
        }
        tagMode, ok := memberPortEntry.Field["tagging_mode"]
        if !ok {
            errStr := "tagging_mode entry is not present for VLAN: " + vlanMember.Get(0) + " Interface: " + *ifName
            return nil, errors.New(errStr)
        }

        vlanName := vlanMember.Get(0)
        vlanMemberKey := vlanName + "|" + *ifName
        if tagMode == "untagged" {
            vlanMemberMap[vlanMemberKey] = memberPortEntry 
            return &vlanName, nil
        }
    }
    errStr := "Untagged VLAN configuration doesn't exist for Interface: " + *ifName
    return nil, tlerr.InvalidArgsError{Format: errStr}
}

func removeAllVlanMembrsForIfAndGetVlans(d *db.DB, ifName *string, ifMode intfModeType, vlanMemberMap map[string]db.Value) (error) {
    var err error
    var vlanKeys []db.Key
    vlanTable, err := d.GetTable(&db.TableSpec{Name: VLAN_MEMBER_TN})
    if err != nil {
        return err
    }

    vlanKeys, err = vlanTable.GetKeys()

    for _, vlanKey := range vlanKeys {
        if len(vlanKeys) < 2 {
            continue
        }
        if vlanKey.Get(1) == *ifName {
            memberPortEntry, err := d.GetEntry(&db.TableSpec{Name:VLAN_MEMBER_TN}, vlanKey)
            if err != nil {
                log.Errorf("Error found on fetching Vlan member info from App DB for Interface Name : %s", *ifName)
                return err
            }
            tagInfo, ok := memberPortEntry.Field["tagging_mode"]
            if ok {
                switch ifMode {
                case ACCESS:
                    if tagInfo != "tagged" {
                        continue
                    }
                case TRUNK:
                    if tagInfo != "untagged" {
                        continue
                    }
                }
                vlanMemberKey := vlanKey.Get(0) + "|" + *ifName
                vlanMemberMap[vlanMemberKey] = db.Value{Field: make(map[string]string)}
                vlanMemberMap[vlanMemberKey] = memberPortEntry
            }
        }
    }
    return err
}

func intfAccessModeReqConfig(d *db.DB, ifName *string,
                             vlanMap map[string]db.Value,
                             vlanMemberMap map[string]db.Value) error {
    var err error
    if len(*ifName) == 0 {
        return errors.New("Empty Interface name received!")
    }

    err = removeAllVlanMembrsForIfAndGetVlans(d, ifName, ACCESS, vlanMemberMap)
    if err != nil {
        return err
    }

    err = removeFromMembersListForAllVlans(d, ifName, vlanMemberMap, vlanMap)
    if err != nil {
        return err
    }
    return err
}

func intfModeReqConfig(d *db.DB, mode intfModeReq,
                       vlanMap map[string]db.Value,
                       vlanMemberMap map[string]db.Value) error {
    var err error
    switch mode.mode {
    case ACCESS:
        err := intfAccessModeReqConfig(d, &mode.ifName, vlanMap, vlanMemberMap)
        if err != nil {
            return err
        }
    case TRUNK:
    case MODE_UNSET:
        break
    }
    return err
}

/* Adding member to VLAN requires updation of VLAN Table and VLAN Member Table */
func processIntfVlanMemberAdd(d *db.DB, vlanMembersMap map[string]map[string]db.Value, vlanMap map[string]db.Value, vlanMemberMap map[string]db.Value) error {
    var err error
    var isMembersListUpdate bool

    /* Updating the VLAN member table */
    for vlanName, ifEntries := range vlanMembersMap {
        log.Info("Processing VLAN: ", vlanName)
        var memberPortsListStrB strings.Builder
        var memberPortsList []string
        isMembersListUpdate = false

        vlanEntry, _ := d.GetEntry(&db.TableSpec{Name:VLAN_TN}, db.Key{Comp: []string{vlanName}})
        if !vlanEntry.IsPopulated() {
            errStr := "Failed to retrieve memberPorts info of VLAN : " + vlanName
            log.Error(errStr)
            return errors.New(errStr)
        }
        memberPortsExists := false
        memberPortsListStr, ok := vlanEntry.Field["members@"]
        if ok {
            if len(memberPortsListStr) != 0 {
                memberPortsListStrB.WriteString(vlanEntry.Field["members@"])
                memberPortsList = generateMemberPortsSliceFromString(&memberPortsListStr)
                memberPortsExists = true
            }
        }

        for ifName, ifEntry := range ifEntries {
            log.Infof("Processing Interface: %s for VLAN: %s", ifName, vlanName)
            /* Adding the following validation, just to avoid an another db-get in translate fn */
            /* Reason why it's ignored is, if we return, it leads to sync data issues between VlanT and VlanMembT */
            if memberPortsExists {
                var existingIfMode intfModeType
                if checkMemberPortExistsInListAndGetMode(d, memberPortsList, &ifName, &vlanName, &existingIfMode) {
                    /* Since translib doesn't support rollback, we need to keep the DB consistent at this point,
                    and throw the error message */
                    var cfgReqIfMode intfModeType
                    tagMode := ifEntry.Field["tagging_mode"]
                    convertTaggingModeToInterfaceModeType(&tagMode, &cfgReqIfMode)

                    if cfgReqIfMode == existingIfMode {
                        continue
                    } else {
                        vlanId := vlanName[len("Vlan"):len(vlanName)]
                        var errStr string
                        switch existingIfMode {
                        case ACCESS:
                            errStr = "Untagged VLAN: " + vlanId + " configuration exists for Interface: " + ifName
                        case TRUNK:
                            errStr = "Tagged VLAN: " + vlanId + " configuration exists for Interface: " + ifName
                        }
                  log.Error(errStr)
                        return tlerr.InvalidArgsError{Format: errStr}
                    }
                }
            }

            isMembersListUpdate = true
            vlanMemberKey := vlanName + "|" + ifName
            vlanMemberMap[vlanMemberKey] = db.Value{Field:make(map[string]string)}
            vlanMemberMap[vlanMemberKey].Field["tagging_mode"] = ifEntry.Field["tagging_mode"] 
            log.Infof("Updated Vlan Member Map with vlan member key: %s and tagging-mode: %s", vlanMemberKey, ifEntry.Field["tagging_mode"])

            if len(memberPortsList) == 0 && len(ifEntries) == 1 {
                memberPortsListStrB.WriteString(ifName)
            } else {
                memberPortsListStrB.WriteString("," + ifName)
            }
        }
        log.Infof("Member ports = %s", memberPortsListStrB.String())
        if !isMembersListUpdate {
            continue
        }
        vlanMap[vlanName] = db.Value{Field:make(map[string]string)}
        vlanMap[vlanName].Field["members@"] = memberPortsListStrB.String()
        log.Infof("Updated VLAN Map with VLAN: %s and Member-ports: %s", vlanName, memberPortsListStrB.String())
    }
    return err
}

func processIntfVlanMemberRemoval(d *db.DB, ifVlanInfoList []*ifVlan, vlanMap map[string]db.Value, vlanMemberMap map[string]db.Value) error {
    var err error

    if len(ifVlanInfoList) == 0 {
        log.Info("No VLAN Info present for membership removal!")
        return nil
    }

    for _, ifVlanInfo := range ifVlanInfoList {
        if ifVlanInfo.ifName == nil {
            return errors.New("No Interface name present for membership removal from VLAN!")
        }

        ifName := ifVlanInfo.ifName
        ifMode := ifVlanInfo.mode
        trunkVlans := ifVlanInfo.trunkVlans

        switch ifMode {
        case ACCESS:
            /* Handling Access Vlan delete */
            log.Info("Access VLAN Delete!")
            untagdVlan, err := removeUntaggedVlanAndUpdateVlanMembTbl(d, ifName, vlanMemberMap)
            if err != nil {
                return err
            }
            if untagdVlan != nil {
                removeFromMembersListForVlan(d, untagdVlan, ifName, vlanMap)
            }

        case TRUNK:
            /* Handling trunk-vlans delete */
            log.Info("Trunk VLAN Delete!")
            if trunkVlans != nil {
                for _, trunkVlan := range trunkVlans {
                    err = removeTaggedVlanAndUpdateVlanMembTbl(d, &trunkVlan, ifName, vlanMemberMap)
                    if err != nil {
                        return err
                    }
                    removeFromMembersListForVlan(d, &trunkVlan, ifName, vlanMap)
                }
            }
        }
    }
    return err
}

/* Function performs VLAN Member removal from Interface */
func intfVlanMemberRemoval(swVlanConfig *swVlanMemberPort_t,
                           inParams *XfmrParams, ifName *string,
                           vlanMap map[string]db.Value,
                           vlanMemberMap map[string]db.Value, intfType E_InterfaceType) error {
    var err error
    var ifVlanInfo ifVlan
    var ifVlanInfoList []*ifVlan

    switch intfType {
    case IntfTypeEthernet:
        if swVlanConfig.swEthMember.Config.AccessVlan != nil {
            ifVlanInfo.mode = ACCESS
        }
        if swVlanConfig.swEthMember.Config.TrunkVlans != nil {
            trunkVlansUnionList := swVlanConfig.swEthMember.Config.TrunkVlans
            ifVlanInfo.mode = TRUNK
            for _, trunkVlanUnion := range trunkVlansUnionList {
                trunkVlanUnionType := reflect.TypeOf(trunkVlanUnion).Elem()

                switch trunkVlanUnionType {

                case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_String{}):
                    val := (trunkVlanUnion).(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_String)
                    vlanName := "Vlan" + val.String
                    err = validateVlanExists(inParams.d, &vlanName)
                    if err != nil {
                        errStr := "Invalid VLAN: " + val.String
                        err = tlerr.InvalidArgsError{Format: errStr}
                        return err
                    }
                    ifVlanInfo.trunkVlans = append(ifVlanInfo.trunkVlans, vlanName)
                case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_Uint16{}):
                    val := (trunkVlanUnion).(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_Uint16)
                    ifVlanInfo.trunkVlans = append(ifVlanInfo.trunkVlans, "Vlan"+strconv.Itoa(int(val.Uint16)))
                }
            }
        }
    case IntfTypePortChannel:
        if swVlanConfig.swPortChannelMember.Config.AccessVlan != nil {
            ifVlanInfo.mode = ACCESS
        }
        if swVlanConfig.swPortChannelMember.Config.TrunkVlans != nil {
            trunkVlansUnionList := swVlanConfig.swPortChannelMember.Config.TrunkVlans
            ifVlanInfo.mode = TRUNK
            for _, trunkVlanUnion := range trunkVlansUnionList {
                trunkVlanUnionType := reflect.TypeOf(trunkVlanUnion).Elem()

                switch trunkVlanUnionType {

                case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_Config_TrunkVlans_Union_String{}):
                    val := (trunkVlanUnion).(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_Config_TrunkVlans_Union_String)
                    vlanName := "Vlan" + val.String
                    err = validateVlanExists(inParams.d, &vlanName)
                    if err != nil {
                        errStr := "Invalid VLAN: " + val.String
                        err = tlerr.InvalidArgsError{Format: errStr}
                        return err
                    }
                    ifVlanInfo.trunkVlans = append(ifVlanInfo.trunkVlans, vlanName)
                case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_Config_TrunkVlans_Union_Uint16{}):
                    val := (trunkVlanUnion).(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_Config_TrunkVlans_Union_Uint16)
                    ifVlanInfo.trunkVlans = append(ifVlanInfo.trunkVlans, "Vlan"+strconv.Itoa(int(val.Uint16)))
                }
            }
        }
    }
    if ifVlanInfo.mode != MODE_UNSET {
        ifVlanInfo.ifName = ifName
        ifVlanInfoList = append(ifVlanInfoList, &ifVlanInfo)
    }
    err = processIntfVlanMemberRemoval(inParams.d, ifVlanInfoList, vlanMap, vlanMemberMap)
    if(err != nil) {
        log.Errorf("Interface VLAN member removal for Interface: %s failed!", *ifName)
        return err
    }
    return err
}

/* Function performs VLAN Member addition to Interface */
func intfVlanMemberAdd(swVlanConfig *swVlanMemberPort_t,
                       inParams *XfmrParams, ifName *string,
                       vlanMap map[string]db.Value,
                       vlanMemberMap map[string]db.Value, intfType E_InterfaceType) error {

    var err error
    var accessVlanId uint16 = 0
    var trunkVlanSlice []string
    var ifMode ocbinds.E_OpenconfigVlan_VlanModeType

    accessVlanFound := false
    trunkVlanFound := false

    intTbl := IntfTypeTblMap[IntfTypeVlan]
    vlanMembersListMap := make(map[string]map[string]db.Value)

    switch intfType {
    case IntfTypeEthernet:
        /* Retrieve the Access VLAN Id */
		if swVlanConfig.swEthMember == nil || swVlanConfig.swEthMember.Config == nil {
			errStr := "Not supported switched-vlan request for Interface: " + *ifName
			log.Error(errStr)
			return errors.New(errStr)
		}
        if swVlanConfig.swEthMember.Config.AccessVlan != nil {
            accessVlanId = *swVlanConfig.swEthMember.Config.AccessVlan
            log.Infof("Vlan id : %d observed for Untagged Member port addition configuration!", accessVlanId)
            accessVlanFound = true
        }

        /* Retrieve the list of trunk-vlans */
        if swVlanConfig.swEthMember.Config.TrunkVlans != nil {
            vlanUnionList := swVlanConfig.swEthMember.Config.TrunkVlans
            if len(vlanUnionList) != 0 {
                trunkVlanFound = true
            }
            for _, vlanUnion := range vlanUnionList {
                vlanUnionType := reflect.TypeOf(vlanUnion).Elem()

                switch vlanUnionType {

                case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_String{}):
                    val := (vlanUnion).(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_String)
                    trunkVlanSlice = append(trunkVlanSlice, val.String)
                case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_Uint16{}):
                    val := (vlanUnion).(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_Uint16)
                    trunkVlanSlice = append(trunkVlanSlice, "Vlan"+strconv.Itoa(int(val.Uint16)))
                }
            }
        }
        if swVlanConfig.swEthMember.Config.InterfaceMode != ocbinds.OpenconfigVlan_VlanModeType_UNSET {
            ifMode = swVlanConfig.swEthMember.Config.InterfaceMode
        }
    case IntfTypePortChannel:
        /* Retrieve the Access VLAN Id */
		if swVlanConfig.swPortChannelMember == nil || swVlanConfig.swPortChannelMember.Config == nil {
			errStr := "Not supported switched-vlan request for Interface: " + *ifName
			log.Error(errStr)
			return errors.New(errStr)
		}
        if swVlanConfig.swPortChannelMember.Config.AccessVlan != nil {
            accessVlanId = *swVlanConfig.swPortChannelMember.Config.AccessVlan
            log.Infof("Vlan id : %d observed for Untagged Member port addition configuration!", accessVlanId)
            accessVlanFound = true
        }

        /* Retrieve the list of trunk-vlans */
        if swVlanConfig.swPortChannelMember.Config.TrunkVlans != nil {
            vlanUnionList := swVlanConfig.swPortChannelMember.Config.TrunkVlans
            if len(vlanUnionList) != 0 {
                trunkVlanFound = true
            }
            for _, vlanUnion := range vlanUnionList {
                vlanUnionType := reflect.TypeOf(vlanUnion).Elem()

                switch vlanUnionType {

                case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_Config_TrunkVlans_Union_String{}):
                    val := (vlanUnion).(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_Config_TrunkVlans_Union_String)
                    trunkVlanSlice = append(trunkVlanSlice, val.String)
                case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_Config_TrunkVlans_Union_Uint16{}):
                    val := (vlanUnion).(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_Config_TrunkVlans_Union_Uint16)
                    trunkVlanSlice = append(trunkVlanSlice, "Vlan"+strconv.Itoa(int(val.Uint16)))
                }
            }
        }
        if swVlanConfig.swPortChannelMember.Config.InterfaceMode != ocbinds.OpenconfigVlan_VlanModeType_UNSET {
            ifMode = swVlanConfig.swPortChannelMember.Config.InterfaceMode
        }
    }

    /* Update the DS based on access-vlan/trunk-vlans config */
    if accessVlanFound {
        accessVlan := "Vlan" + strconv.Itoa(int(accessVlanId))

        err = validateVlanExists(inParams.d, &accessVlan)
        if err != nil {
            errStr := "Invalid VLAN: " + strconv.Itoa(int(accessVlanId))
            err = tlerr.InvalidArgsError{Format: errStr}
            log.Error(err)
            return err
        }
        var cfgredAccessVlan string
        exists, err := validateUntaggedVlanCfgredForIf(inParams.d, &intTbl.cfgDb.memberTN, ifName, &cfgredAccessVlan)
        if err != nil {
            return err
        }
        if exists {
            if cfgredAccessVlan == accessVlan {
                log.Infof("Untagged VLAN: %s already configured, not updating the cache!", accessVlan)
                goto TRUNKCONFIG
            }
            vlanId := cfgredAccessVlan[len("Vlan"):len(cfgredAccessVlan)]
            errStr := "Untagged VLAN: " + vlanId + " configuration exists"
            log.Error(errStr)
            err = tlerr.InvalidArgsError{Format: errStr}
            return err
        }
        if vlanMembersListMap[accessVlan] == nil {
            vlanMembersListMap[accessVlan] = make(map[string]db.Value)
        }
        vlanMembersListMap[accessVlan][*ifName] = db.Value{Field:make(map[string]string)}
        vlanMembersListMap[accessVlan][*ifName].Field["tagging_mode"] = "untagged"
    }

    TRUNKCONFIG:
    if trunkVlanFound {
        memberPortEntryMap := make(map[string]string)
        memberPortEntry := db.Value{Field: memberPortEntryMap}
        memberPortEntry.Field["tagging_mode"] = "tagged"
        for _, vlanId := range trunkVlanSlice {

            err = validateVlanExists(inParams.d, &vlanId)
            if err != nil {
                id := vlanId[len("Vlan"):len(vlanId)]
                errStr := "Invalid VLAN: " + id
                log.Error(errStr)
                err = tlerr.InvalidArgsError{Format: errStr}
                return err
            }

            if vlanMembersListMap[vlanId] == nil {
                vlanMembersListMap[vlanId] = make(map[string]db.Value)
            }
            vlanMembersListMap[vlanId][*ifName] = db.Value{Field:make(map[string]string)}
            vlanMembersListMap[vlanId][*ifName].Field["tagging_mode"] = "tagged"
        }
    }
    if accessVlanFound || trunkVlanFound {
        err = processIntfVlanMemberAdd(inParams.d, vlanMembersListMap, vlanMap, vlanMemberMap)
        if err != nil {
            log.Info("Processing Interface VLAN addition failed!")
            return err
        }
        return err
    }

    if ifMode == ocbinds.OpenconfigVlan_VlanModeType_UNSET {
	    return nil
	}
    /* Handling the request just for setting Interface Mode */
    log.Info("Request is for Configuring just the Mode for Interface: ", *ifName)
    var mode intfModeReq

    switch ifMode {
    case ocbinds.OpenconfigVlan_VlanModeType_ACCESS:
        /* Configuring Interface Mode as ACCESS only without VLAN info*/
        mode = intfModeReq{ifName: *ifName, mode: ACCESS}
        log.Info("Access Mode Config for Interface: ", *ifName)
    case ocbinds.OpenconfigVlan_VlanModeType_TRUNK:
    }
    /* Switchport access/trunk mode config without VLAN */
    /* This mode will be set in the translate fn, when request is just for mode without VLAN info. */
    if mode.mode != MODE_UNSET {
        err = intfModeReqConfig(inParams.d, mode, vlanMap, vlanMemberMap)
        if err != nil {
            return err
        }
    }
    return nil
}

/* Function to delete VLAN and all its member ports */
func deleteVlanIntfAndMembers(inParams *XfmrParams, vlanName *string) error {
    var err error
    subOpMap := make(map[db.DBNum]map[string]map[string]db.Value)
    resMap := make(map[string]map[string]db.Value)
    vlanMap := make(map[string]db.Value)
    vlanMemberMap := make(map[string]db.Value)

    vlanMap[*vlanName] = db.Value{Field:map[string]string{}}
    subOpMap[db.ConfigDB] = resMap
    inParams.subOpDataMap[DELETE] = &subOpMap

    vlanEntry, err := inParams.d.GetEntry(&db.TableSpec{Name:VLAN_TN}, db.Key{Comp: []string{*vlanName}})
    if err != nil {
        log.Errorf("Retrieving data from VLAN table for VLAN: %s failed!", *vlanName)
        return err
    }
    /* Handle VLAN_INTERFACE TABLE */
    ipCnt := 0
    _ = interfaceIPcount(VLAN_INTERFACE_TN, inParams.d, vlanName, &ipCnt)
    if ipCnt > 0 {
        errStr := "Need to first remove IP address entry"
        log.Error(errStr)
        return errors.New(errStr)
    }

    memberPortsVal, ok := vlanEntry.Field["members@"]
    if ok {
        memberPorts := generateMemberPortsSliceFromString(&memberPortsVal)
        if memberPorts == nil {
            return nil
        }
        log.Infof("MemberPorts for VLAN: %s = %s", *vlanName, memberPortsVal)

        for _, memberPort := range memberPorts {
            log.Infof("Member Port:%s part of vlan:%s to be deleted!", memberPort, *vlanName)
            if err != nil {
                log.Errorf("Get for VLAN_MEMBER table for VLAN: %s and Interface: %s failed!", *vlanName, memberPort)
                return err
            }
            vlanMemberKey := *vlanName + "|" + memberPort
            vlanMemberMap[vlanMemberKey] = db.Value{Field:map[string]string{}} 
            if err != nil {
                return err
            }
        }
        resMap[VLAN_MEMBER_TN] = vlanMemberMap
    }
    resMap[VLAN_TN] = vlanMap

    subOpMap[db.ConfigDB] = resMap
    inParams.subOpDataMap[DELETE] = &subOpMap
    return err
}

/* Subtree transformer supports CREATE, UPDATE and DELETE operations */
var YangToDb_sw_vlans_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    var err error
    res_map := make(map[string]map[string]db.Value)
    vlanMap := make(map[string]db.Value)
    vlanMemberMap := make(map[string]db.Value)

    log.Info("YangToDb_sw_vlans_xfmr: ", inParams.uri)

    var swVlanConfig swVlanMemberPort_t 
    pathInfo := NewPathInfo(inParams.uri)
    ifName := pathInfo.Var("name")

    deviceObj := (*inParams.ygRoot).(*ocbinds.Device)
    intfObj := deviceObj.Interfaces

    log.Info("Switched vlans request for ", ifName)
    intf := intfObj.Interface[ifName]

    if intf.Ethernet == nil && intf.Aggregation == nil {
        return nil, errors.New("Wrong Config Request")
    }    
    if intf.Ethernet != nil {
        if intf.Ethernet.SwitchedVlan == nil || intf.Ethernet.SwitchedVlan.Config == nil {
            return nil, errors.New("Wrong config request for Ethernet!")
        }
        swVlanConfig.swEthMember = intf.Ethernet.SwitchedVlan
    }
    if intf.Aggregation != nil {
        if intf.Aggregation.SwitchedVlan == nil || intf.Aggregation.SwitchedVlan.Config == nil {
            return nil, errors.New("Wrong Config Request for Port Channel")
        }
        swVlanConfig.swPortChannelMember = intf.Aggregation.SwitchedVlan
    }

    intfType, _, err := getIntfTypeByName(ifName)
    if err != nil {
        errStr := "Extraction of Interface type from Interface: " + ifName + " failed!"
        return nil, errors.New(errStr)
    }
    switch inParams.oper {
    case CREATE:
    case UPDATE:
        err = intfVlanMemberAdd(&swVlanConfig, &inParams, &ifName, vlanMap, vlanMemberMap, intfType)
        if err != nil {
            log.Errorf("Interface VLAN member port addition failed for Interface: %s!", ifName)
            return nil, err
        }
        res_map[VLAN_TN] = vlanMap
        res_map[VLAN_MEMBER_TN] = vlanMemberMap
    case DELETE:
        err = intfVlanMemberRemoval(&swVlanConfig, &inParams, &ifName, vlanMap, vlanMemberMap, intfType)
        if err != nil {
            log.Errorf("Interface VLAN member port removal failed for Interface: %s!", ifName)
            return nil, err
        }

        res_map[VLAN_MEMBER_TN] = vlanMemberMap
        res_map[VLAN_TN] = vlanMap
    }

    log.Info("YangToDb_sw_vlans_xfmr: vlan res map:", res_map)
    return res_map, err
}

func fillDBSwitchedVlanInfoForIntf(d *db.DB, ifName *string, vlanMemberMap map[string]map[string]db.Value) error {
    log.Info("fillDBSwitchedVlanInfoForIntf() called!")
    var err error

    vlanMemberTable, err := d.GetTable(&db.TableSpec{Name: VLAN_MEMBER_TN})
    if err != nil {
        return err
    }
    vlanMemberKeys, err := vlanMemberTable.GetKeys()
    if err != nil {
        return err
    }
    log.Infof("Found %d vlan-member-table keys", len(vlanMemberKeys))

    for _, vlanMember := range vlanMemberKeys {
        if len(vlanMember.Comp) < 2 {
            continue
        }
        vlanId := vlanMember.Get(0)
        ifName := vlanMember.Get(1)
        log.Infof("Received Vlan: %s for Interface: %s", vlanId, ifName)

        memberPortEntry, err := d.GetEntry(&db.TableSpec{Name: VLAN_MEMBER_TN}, vlanMember)
        if err != nil {
            return err
        }
        if !memberPortEntry.IsPopulated() {
            errStr := "Tagging Info not present for Vlan: " + vlanId + " Interface: " + ifName + " from VLAN_MEMBER_TABLE"
            return errors.New(errStr)
        }

        /* vlanMembersTableMap is used as DS for ifName to list of VLANs */
        if vlanMemberMap[ifName] == nil {
            vlanMemberMap[ifName] = make(map[string]db.Value)
            vlanMemberMap[ifName][vlanId] = memberPortEntry
        } else {
            vlanMemberMap[ifName][vlanId] = memberPortEntry
        }
    }
    log.Infof("Updated the vlan-member-table ds for Interface: %s", *ifName)
    return err
}

func getIntfVlanAttr(ifName *string, ifMode intfModeType, vlanMemberMap map[string]map[string]db.Value) ([]string, *string, error) {

    log.Info("getIntfVlanAttr() called")
    vlanEntries, ok := vlanMemberMap[*ifName]
    if !ok {
        errStr := "Cannot find info for Interface: " + *ifName + " from VLAN_MEMBERS_TABLE!"
        return nil, nil, errors.New(errStr)
    }
    switch ifMode {
    case ACCESS:
        for vlanKey, tagEntry := range vlanEntries {
            tagMode, ok := tagEntry.Field["tagging_mode"]
            if ok {
                if tagMode == "untagged" {
                    log.Info("Untagged VLAN found!")
                    return nil, &vlanKey, nil
                }
            }
        }
    case TRUNK:
        var trunkVlans []string
        for vlanKey, tagEntry := range vlanEntries {
            tagMode, ok := tagEntry.Field["tagging_mode"]
            if ok {
                if tagMode == "tagged" {
                    trunkVlans = append(trunkVlans, vlanKey)
                }
            }
        }
        return trunkVlans, nil, nil
    }
    return nil, nil, nil
}

func getSpecificSwitchedVlanStateAttr(targetUriPath *string, ifKey *string, 
                                      vlanMemberMap map[string]map[string]db.Value,
                                      swVlan *swVlanMemberPort_t, intfType E_InterfaceType) (bool, error) {
    log.Info("Specific Switched-vlan attribute!")
    switch *targetUriPath {
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/state/access-vlan":
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/openconfig-vlan:switched-vlan/state/access-vlan":
        _, accessVlanName, e := getIntfVlanAttr(ifKey, ACCESS, vlanMemberMap)
        if e != nil {
            return true, e
        }
        if accessVlanName == nil {
            return true, nil
        }
        log.Info("Access VLAN - ", accessVlanName)
        vlanName := *accessVlanName
        vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
        vlanId, err := strconv.Atoi(vlanIdStr)
        if err != nil {
            errStr := "Conversion of string to int failed for " + vlanIdStr
            return true, errors.New(errStr)
        }
        vlanIdCast := uint16(vlanId)

        switch intfType {
        case IntfTypeEthernet:
            swVlan.swEthMember.State.AccessVlan = &vlanIdCast
        case IntfTypePortChannel:
            swVlan.swPortChannelMember.State.AccessVlan = &vlanIdCast
        }
        return true, nil
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/state/trunk-vlans":
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/openconfig-vlan:switched-vlan/state/trunk-vlans":
        trunkVlans, _, e := getIntfVlanAttr(ifKey, TRUNK, vlanMemberMap)
        if e != nil {
            return true, e
        }

        switch intfType {
        case IntfTypeEthernet:

            for _, vlanName := range trunkVlans {
                log.Info("Trunk VLAN - ", vlanName)
                vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
                vlanId, err := strconv.Atoi(vlanIdStr)
                if err != nil {
                    errStr := "Conversion of string to int failed for " + vlanIdStr
                    return true, errors.New(errStr)
                }
                vlanIdCast := uint16(vlanId)

                trunkVlan, _ := swVlan.swEthMember.State.To_OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_State_TrunkVlans_Union(vlanIdCast)
                swVlan.swEthMember.State.TrunkVlans = append(swVlan.swEthMember.State.TrunkVlans, trunkVlan)
            }
        case IntfTypePortChannel:
            for _, vlanName := range trunkVlans {
                log.Info("Trunk VLAN - ", vlanName)
                vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
                vlanId, err := strconv.Atoi(vlanIdStr)
                if err != nil {
                    errStr := "Conversion of string to int failed for " + vlanIdStr
                    return true, errors.New(errStr)
                }
                vlanIdCast := uint16(vlanId)

                trunkVlan, _ := swVlan.swPortChannelMember.State.To_OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_State_TrunkVlans_Union(vlanIdCast)
                swVlan.swPortChannelMember.State.TrunkVlans = append(swVlan.swPortChannelMember.State.TrunkVlans, trunkVlan)
            }
        }
        return true, nil
    case "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/state/interface-mode":
        return true, errors.New("Interface mode attribute not supported!")
    }
    return false, nil
}

func getSwitchedVlanState(ifKey *string, vlanMemberMap map[string]map[string]db.Value,
                          swVlan *swVlanMemberPort_t, intfType E_InterfaceType) (error) {
    /* Get Access VLAN info for Interface */
    _, accessVlanName, e := getIntfVlanAttr(ifKey, ACCESS, vlanMemberMap)
    if e != nil {
        return e
    }

    /* Get Trunk VLAN info for Interface */
    trunkVlans, _, e := getIntfVlanAttr(ifKey, TRUNK, vlanMemberMap)
    if e != nil {
        return e
    }

    switch intfType {
    case IntfTypeEthernet:
        if accessVlanName != nil {
            vlanName := *accessVlanName
            vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
            vlanId, err := strconv.Atoi(vlanIdStr)
            if err != nil {
                errStr := "Conversion of string to int failed for " + vlanIdStr
                return errors.New(errStr)
            }
            vlanIdCast := uint16(vlanId)
            swVlan.swEthMember.State.AccessVlan = &vlanIdCast
        }
        for _, vlanName := range trunkVlans {
            vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
            vlanId, err := strconv.Atoi(vlanIdStr)
            if err != nil {
                errStr := "Conversion of string to int failed for " + vlanIdStr
                return errors.New(errStr)
            }
            vlanIdCast := uint16(vlanId)
            trunkVlan, _ := swVlan.swEthMember.State.To_OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_State_TrunkVlans_Union(vlanIdCast)
            swVlan.swEthMember.State.TrunkVlans = append(swVlan.swEthMember.State.TrunkVlans, trunkVlan)
        }
    case IntfTypePortChannel:
        if accessVlanName != nil {
            vlanName := *accessVlanName
            vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
            vlanId, err := strconv.Atoi(vlanIdStr)
            if err != nil {
                errStr := "Conversion of string to int failed for " + vlanIdStr
                return errors.New(errStr)
            }
            vlanIdCast := uint16(vlanId)
            swVlan.swPortChannelMember.State.AccessVlan = &vlanIdCast
        }
        for _, vlanName := range trunkVlans {
            vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
            vlanId, err := strconv.Atoi(vlanIdStr)
            if err != nil {
                errStr := "Conversion of string to int failed for " + vlanIdStr
                return errors.New(errStr)
            }
            vlanIdCast := uint16(vlanId)
            trunkVlan, _ := swVlan.swPortChannelMember.State.To_OpenconfigInterfaces_Interfaces_Interface_Aggregation_SwitchedVlan_State_TrunkVlans_Union(vlanIdCast)
            swVlan.swPortChannelMember.State.TrunkVlans = append(swVlan.swPortChannelMember.State.TrunkVlans, trunkVlan)
        }
    }
    return nil 
}

/* Subtree transformer supports GET operation */
var DbToYang_sw_vlans_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) (error) {
    var err error
    var swVlan swVlanMemberPort_t
    intfsObj := getIntfsRoot(inParams.ygRoot)
    if intfsObj == nil {
        errStr := "Nil root object received for Ethernet-Switched VLAN Get!"
        log.Errorf(errStr)
        return errors.New(errStr)
    }
    pathInfo := NewPathInfo(inParams.uri)

    ifName := pathInfo.Var("name")
    log.Infof("Ethernet-Switched Vlan Get observed for Interface: %s", ifName)
    intfType, _, err := getIntfTypeByName(ifName)
    if intfType != IntfTypeEthernet && intfType != IntfTypePortChannel || err != nil {
        intfTypeStr := strconv.Itoa(int(intfType))
        errStr := "TableXfmrFunc - Invalid interface type" + intfTypeStr
        log.Error(errStr);
        return errors.New(errStr);
    }

    targetUriPath, err := getYangPathFromUri(inParams.uri)
    log.Info("targetUriPath is ", targetUriPath)

    /* Transformer must have given a tree with everything filled up to switched-vlan container */
    if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan") ||
       strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/openconfig-vlan:switched-vlan") {
        intfObj := intfsObj.Interface[ifName]
        if intfObj.Ethernet == nil && intfObj.Aggregation == nil {
            return errors.New("Wrong GET request for switched-vlan!")
        }
        if intfObj.Ethernet != nil {
            if intfObj.Ethernet.SwitchedVlan == nil {
                errStr := "Required tree for Ethernet-Switched-Vlan request not built correctly!"
                return errors.New(errStr)
            }
            swVlan.swEthMember = intfObj.Ethernet.SwitchedVlan
        }
        if intfObj.Aggregation != nil {
            if intfObj.Aggregation.SwitchedVlan == nil {
                errStr := "Required tree for Aggregation-Switched-Vlan request not built correctly!"
                return errors.New(errStr)
            }
            swVlan.swPortChannelMember = intfObj.Aggregation.SwitchedVlan
        }
        if strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/state") ||
           strings.HasPrefix(targetUriPath, "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/openconfig-vlan:switched-vlan/state") {
            switch intfType {
            case IntfTypeEthernet:
                if intfObj.Ethernet == nil || intfObj.Ethernet.SwitchedVlan == nil || intfObj.Ethernet.SwitchedVlan.State == nil {
                    errStr := "Switched-vlan state tree not built correctly for Interface: " + ifName
                    log.Error(errStr)
                    return errors.New(errStr)
                }
                vlanMemberMap := make(map[string]map[string]db.Value)
                err = fillDBSwitchedVlanInfoForIntf(inParams.d, &ifName, vlanMemberMap)
                if err != nil {
                    log.Errorf("Filiing Switched Vlan Info for Interface: %s failed!", ifName)
                    return err
                }
                log.Info("Succesfully completed DB population for Ethernet!")
                attrPresent, err := getSpecificSwitchedVlanStateAttr(&targetUriPath, &ifName, vlanMemberMap, &swVlan, intfType)
                if(err != nil) {
                    return err
                }
                if(!attrPresent) {
                    log.Infof("Get is for Switched Vlan State Container!")
                    err = getSwitchedVlanState(&ifName, vlanMemberMap, &swVlan, intfType)
                    if err != nil {
                        return err
                    }
                }

            case IntfTypePortChannel:
                if intfObj.Aggregation == nil || intfObj.Aggregation.SwitchedVlan == nil || intfObj.Aggregation.SwitchedVlan.State == nil {
                    errStr := "Switched-vlan state tree not built correctly for Interface: " + ifName
                    log.Error(errStr)
                    return errors.New(errStr)
                }
                vlanMemberMap := make(map[string]map[string]db.Value)
                err = fillDBSwitchedVlanInfoForIntf(inParams.d, &ifName, vlanMemberMap)
                if err != nil {
                    log.Errorf("Filiing Switched Vlan Info for Interface: %s failed!", ifName)
                    return err
                }
                log.Info("Succesfully completed DB population for Port-Channel!")
                attrPresent, err := getSpecificSwitchedVlanStateAttr(&targetUriPath, &ifName, vlanMemberMap, &swVlan, intfType)
                if(err != nil) {
                    return err
                }
                if(!attrPresent) {
                    log.Infof("Get is for Switched Vlan State Container!")
                    err = getSwitchedVlanState(&ifName, vlanMemberMap, &swVlan, intfType)
                    if err != nil {
                        return err
                    }
                }
            } 
        } else {
            errStr := "Not Supported URI: " + targetUriPath
            log.Errorf(errStr)
            return errors.New(errStr)
        }
    } else {
        err = errors.New("Invalid URI : " + targetUriPath)
  }
    return err
}
