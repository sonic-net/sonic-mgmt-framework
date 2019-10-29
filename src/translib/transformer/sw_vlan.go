package transformer

import (
    "errors"
	"strconv"
    "translib/db"
    "translib/ocbinds"
	"translib/tlerr"
    "reflect"
    log "github.com/golang/glog"
)

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
		return errors.New(errStr)
	}
	return nil
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
			return false, errors.New(errStr)
		}
		tagMode, ok := memberPortEntry.Field["tagging_mode"]
		if !ok {
			errStr := "tagging_mode entry is not present for VLAN: " + vlanMember.Get(0) + " Interface: " + *ifName
			return false, errors.New(errStr)
		}
		if tagMode == "untagged" {
			*accessVlan = vlanMember.Get(0)
			return true, nil
		}
	}
	return false, nil
}

var YangToDb_sw_vlans_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    var err error
    res_map := make(map[string]map[string]db.Value)
    vlanMap := make(map[string]db.Value)
    vlanMemberMap := make(map[string]db.Value)

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

		err = validateVlanExists(inParams.dbs[db.ConfigDB], &intTbl.cfgDb.portTN, &accessVlan)
		if err != nil {
			errStr := "Invalid VLAN: " + strconv.Itoa(int(accessVlanId))
			err = tlerr.InvalidArgsError{Format: errStr}
			return res_map, err
		}
		var cfgredAccessVlan string
		exists, err := validateUntaggedVlanCfgredForIf(inParams.dbs[db.ConfigDB], &intTbl.cfgDb.memberTN, &ifName, &cfgredAccessVlan)
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
			err = tlerr.InvalidArgsError{Format: errStr}
			return res_map, err
		}

		key := accessVlan + "|" + ifName
		vlanMemberMap[key] =  db.Value{Field: make(map[string]string)}
		vlanMemberMap[key].Field["tagging_mode"] = "untagged"
		
    	vlanMap[accessVlan] = db.Value{Field: make(map[string]string)}
    	vlanMap[accessVlan].Field["members@"] = ifName

		log.Info("Untagged key: ", key)
	}

TRUNKCONFIG:
	if trunkVlanFound {
		memberPortEntryMap := make(map[string]string)
		memberPortEntry := db.Value{Field: memberPortEntryMap}
		memberPortEntry.Field["tagging_mode"] = "tagged"
		for _, vlanId := range trunkVlanSlice {

			err = validateVlanExists(inParams.dbs[db.ConfigDB], &intTbl.cfgDb.portTN, &vlanId)
			if err != nil {
				id := vlanId[len("Vlan"):len(vlanId)]
				errStr := "Invalid VLAN: " + id
				err = tlerr.InvalidArgsError{Format: errStr}
				return res_map, err
			}

		key := vlanId + "|" + ifName
		vlanMemberMap[key] =  db.Value{Field: make(map[string]string)}
		vlanMemberMap[key].Field["tagging_mode"] = "tagged"
		
    	vlanMap[vlanId] = db.Value{Field: make(map[string]string)}
    	vlanMap[vlanId].Field["members@"] = ifName

		log.Info("Tagged key ", key)
		}
	}
    res_map[VLAN_TABLE] = vlanMap
    res_map[VLAN_MEM_TABLE] = vlanMemberMap

    log.Info("YangToDb_sw_vlans_xfmr: vlan res map:", res_map)
    return res_map, err

