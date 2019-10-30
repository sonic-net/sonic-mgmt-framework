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

type intfModeCfgAlone struct {
  ifName string
  mode   intfModeType
}

func init () {
    XlateFuncBind("YangToDb_sw_vlans_xfmr", YangToDb_sw_vlans_xfmr)
}

/* Validate whether VLAN exists in DB */
func validateVlanExists(d *db.DB, vlanTs *string, vlanName *string) error {
    if len(*vlanName) == 0 {
        return errors.New("Length of VLAN name is zero")
    }
    entry, err := d.GetEntry(&db.TableSpec{Name:*vlanTs}, db.Key{Comp: []string{*vlanName}})
    if err != nil || !entry.IsPopulated() {
        errStr := "Invalid Vlan:" + *vlanName
		log.Error(errStr)
        return errors.New(errStr)
    }
    return nil
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

/* Adding member to VLAN requires updation of VLAN Table and VLAN Member Table */
func processIntfVlanAdd(d *db.DB, vlanMembersMap map[string]map[string]db.Value, vlanMap map[string]db.Value, vlanMemberMap map[string]db.Value) error {
  log.Info("processIntfVlanAdd() called")
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
            vlanMap[vlanName].Field["members@"] = memberPortsListStrB.String()

            vlanId := vlanName[len("Vlan"):len(vlanName)]
            var errStr string
            switch existingIfMode {
            case ACCESS:
              errStr = "Untagged VLAN: " + vlanId + " configuration exists for Interface: " + ifName
            case TRUNK:
              errStr = "Tagged VLAN: " + vlanId + " configuration exists for Interface: " + ifName
            }
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
    log.Info("Updated VLAN Map with VLAN: %s and Member-ports: %s", vlanName, memberPortsListStrB.String())
  }
  return err
}

var YangToDb_sw_vlans_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    var err error
    res_map := make(map[string]map[string]db.Value)
  vlanMap := make(map[string]db.Value)
    vlanMemberMap := make(map[string]db.Value)
  vlanMembersListMap := make(map[string]map[string]db.Value)

    log.Info("YangToDb_sw_vlans_xfmr: ", inParams.uri)

    pathInfo := NewPathInfo(inParams.uri)
    ifName := pathInfo.Var("name")
    intTbl := IntfTypeTblMap[IntfTypeVlan]

    deviceObj := (*inParams.ygRoot).(*ocbinds.Device)
    intfObj := deviceObj.Interfaces

    log.Info("Switched vlans requrest for ", ifName)
    intf := intfObj.Interface[ifName]

    if intf.Ethernet == nil || intf.Ethernet.SwitchedVlan == nil || intf.Ethernet.SwitchedVlan.Config == nil {
        return nil, errors.New("Wrong config request!")
    }
    swVlanConfig := intf.Ethernet.SwitchedVlan.Config

    var accessVlanId uint16 = 0
    var trunkVlanSlice []string
    accessVlanFound := false
    trunkVlanFound := false

    /* Retrieve the Access VLAN Id */
    if swVlanConfig.AccessVlan != nil {
        accessVlanId = *swVlanConfig.AccessVlan
        log.Infof("Vlan id : %d observed for Untagged Member port addition configuration!", accessVlanId)
        accessVlanFound = true
    }

    /* Retrieve the list of trunk-vlans */
    if swVlanConfig.TrunkVlans != nil {
        vlanUnionList := swVlanConfig.TrunkVlans
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

    /* Update the DS based on access-vlan/trunk-vlans config */
    if accessVlanFound {
        accessVlan := "Vlan" + strconv.Itoa(int(accessVlanId))

        err = validateVlanExists(inParams.d, &intTbl.cfgDb.portTN, &accessVlan)
        if err != nil {
            errStr := "Invalid VLAN: " + strconv.Itoa(int(accessVlanId))
            err = tlerr.InvalidArgsError{Format: errStr}
			log.Error(err)
            return res_map, err
        }
        var cfgredAccessVlan string
        exists, err := validateUntaggedVlanCfgredForIf(inParams.d, &intTbl.cfgDb.memberTN, &ifName, &cfgredAccessVlan)
        if err != nil {
            return res_map, err
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
            return res_map, err
        }
    if vlanMembersListMap[accessVlan] == nil {
      vlanMembersListMap[accessVlan] = make(map[string]db.Value)
    }
        vlanMembersListMap[accessVlan][ifName] = db.Value{Field:make(map[string]string)}
    vlanMembersListMap[accessVlan][ifName].Field["tagging_mode"] = "untagged"
    }

    TRUNKCONFIG:
    if trunkVlanFound {
        memberPortEntryMap := make(map[string]string)
        memberPortEntry := db.Value{Field: memberPortEntryMap}
        memberPortEntry.Field["tagging_mode"] = "tagged"
        for _, vlanId := range trunkVlanSlice {

            err = validateVlanExists(inParams.d, &intTbl.cfgDb.portTN, &vlanId)
            if err != nil {
                id := vlanId[len("Vlan"):len(vlanId)]
                errStr := "Invalid VLAN: " + id
				log.Error(errStr)
                err = tlerr.InvalidArgsError{Format: errStr}
                return res_map, err
            }

      if vlanMembersListMap[vlanId] == nil {
        vlanMembersListMap[vlanId] = make(map[string]db.Value)
      }
      vlanMembersListMap[vlanId][ifName] = db.Value{Field:make(map[string]string)}
      vlanMembersListMap[vlanId][ifName].Field["tagging_mode"] = "tagged"
        }
    }
  err = processIntfVlanAdd(inParams.d, vlanMembersListMap, vlanMap, vlanMemberMap)
  if err != nil {
    log.Info("Processing Interface VLAN addition failed!")
    return res_map, err
  }

    res_map[VLAN_TN] = vlanMap
    res_map[VLAN_MEMBER_TN] = vlanMemberMap

    log.Info("YangToDb_sw_vlans_xfmr: vlan res map:", res_map)
    return res_map, err
}
