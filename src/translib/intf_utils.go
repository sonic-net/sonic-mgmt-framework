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
	"fmt"
	log "github.com/golang/glog"
	"net"
	"regexp"
	"strconv"
	"strings"
	"translib/db"
	"translib/tlerr"
)

/* Extract Interface type from Interface name */
func (app *IntfApp) getIntfTypeFromIntf(ifName *string) error {
	var err error

	if len(*ifName) == 0 {
		return errors.New("Interface name received is empty! Fetching if-type from interface failed!")
	}
	if strings.HasPrefix(*ifName, "Ethernet") {
		app.intfType = ETHERNET
	} else if strings.HasPrefix(*ifName, "Vlan") {
		app.intfType = VLAN
	} else if strings.HasPrefix(*ifName, "PortChannel") {
		app.intfType = LAG
	} else {
		return errors.New("Fetching Interface type from Interface name failed!")
	}
	return err
}

/* Validates whether the specific IP exists in the DB for an Interface*/
func (app *IntfApp) validateIp(dbCl *db.DB, ifName string, ip string, ts *db.TableSpec) error {
	app.allIpKeys, _ = app.doGetAllIpKeys(dbCl, ts)

	for _, key := range app.allIpKeys {
		if len(key.Comp) < 2 {
			continue
		}
		if key.Get(0) != ifName {
			continue
		}
		ipAddr, _, _ := net.ParseCIDR(key.Get(1))
		ipStr := ipAddr.String()
		if ipStr == ip {
			log.Infof("IP address %s exists, updating the DS for deletion!", ipStr)
			ipInfo, err := dbCl.GetEntry(ts, key)
			if err != nil {
				log.Error("Error found on fetching Interface IP info from App DB for Interface Name : ", ifName)
				return err
			}
			if len(app.ifIPTableMap[key.Get(0)]) == 0 {
				app.ifIPTableMap[key.Get(0)] = make(map[string]dbEntry)
				app.ifIPTableMap[key.Get(0)][key.Get(1)] = dbEntry{entry: ipInfo}
			} else {
				app.ifIPTableMap[key.Get(0)][key.Get(1)] = dbEntry{entry: ipInfo}
			}
			return nil
		}
	}
	return errors.New(fmt.Sprintf("IP address : %s doesn't exist!", ip))
}

/* Validate whether the Interface has IP configuration */
func (app *IntfApp) validateIpCfgredForInterface(dbCl *db.DB, ifName *string) bool {
	app.allIpKeys, _ = app.doGetAllIpKeys(dbCl, app.intfD.intfIPTs)

	for _, key := range app.allIpKeys {
		if len(key.Comp) < 2 {
			continue
		}
		if key.Get(0) == *ifName {
			return false
		}
	}
	return true
}

/* Check for IP overlap */
func (app *IntfApp) translateIpv4(d *db.DB, intf string, ip string, prefix int) error {
	var err error
	var ifsKey db.Key

	ifsKey.Comp = []string{intf}

	ipPref := ip + "/" + strconv.Itoa(prefix)
	ifsKey.Comp = []string{intf, ipPref}

	log.Info("ifsKey:=", ifsKey)

	log.Info("Checking for IP overlap ....")
	ipA, ipNetA, _ := net.ParseCIDR(ipPref)

	for _, key := range app.allIpKeys {
		if len(key.Comp) < 2 {
			continue
		}
		ipB, ipNetB, _ := net.ParseCIDR(key.Get(1))

		if ipNetA.Contains(ipB) || ipNetB.Contains(ipA) {
			log.Info("IP ", ipPref, "overlaps with ", key.Get(1), " of ", key.Get(0))

			if intf != key.Get(0) {
				//IP overlap across different interface, reject
				log.Error("IP ", ipPref, " overlaps with ", key.Get(1), " of ", key.Get(0))

				errStr := "IP " + ipPref + " overlaps with IP " + key.Get(1) + " of Interface " + key.Get(0)
				err = tlerr.InvalidArgsError{Format: errStr}
				return err
			} else {
				//IP overlap on same interface, replace
				var entry dbEntry
				entry.op = opDelete

				log.Info("Entry ", key.Get(1), " on ", intf, " needs to be deleted")
				if app.ifIPTableMap[intf] == nil {
					app.ifIPTableMap[intf] = make(map[string]dbEntry)
				}
				app.ifIPTableMap[intf][key.Get(1)] = entry
			}
		}
	}

	//At this point, we need to add the entry to db
	{
		var entry dbEntry
		entry.op = opCreate

		m := make(map[string]string)
		m["NULL"] = "NULL"
		value := db.Value{Field: m}
		entry.entry = value
		if app.ifIPTableMap[intf] == nil {
			app.ifIPTableMap[intf] = make(map[string]dbEntry)
		}
		app.ifIPTableMap[intf][ipPref] = entry
	}
	return err
}

/* Validate whether VLAN exists in DB */
func (app *IntfApp) validateVlanExists(d *db.DB, vlanName *string) error {
	if len(*vlanName) == 0 {
		return errors.New("Length of VLAN name is zero")
	}
	entry, err := d.GetEntry(app.vlanD.vlanTs, db.Key{Comp: []string{*vlanName}})
	if err != nil || !entry.IsPopulated() {
		errStr := "Invalid Vlan:" + *vlanName
		return errors.New(errStr)
	}
	return nil
}

/* Validate whether LAG exists in DB */
func (app *IntfApp) validateLagExists(d *db.DB, lagName *string) error {
	if len(*lagName) == 0 {
		return errors.New("Length of Lag name is zero")
	}
	entry, err := d.GetEntry(app.lagD.lagTs, db.Key{Comp: []string{*lagName}})
	log.Info("Lag Entry found:", entry)
	if err != nil || !entry.IsPopulated() {
		errStr := "Invalid Lag:" + *lagName
		return errors.New(errStr)
	}
	return nil
}

/* Validate whether physical interface is valid or not */
/* TODO: This needs to be extended based on Interface type */
func (app *IntfApp) validateInterface(dbCl *db.DB, ifName string, ifKey db.Key) error {
	var err error
	if len(ifName) == 0 {
		return errors.New("Empty Interface name")
	}

	_, err = dbCl.GetEntry(app.intfD.portTblTs, ifKey)
	if err != nil {
		log.Errorf("Error found on fetching Interface info from App DB for If Name : %s", ifName)
		errStr := "Invalid Interface:" + ifName
		err = tlerr.InvalidArgsError{Format: errStr}
		return err
	}
	return err
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

/* Extract VLAN-Id from Vlan String */
func getVlanIdFromVlanName(vlanName *string) (string, error) {
	if !strings.HasPrefix(*vlanName, "Vlan") {
		return "", errors.New("Not valid vlan name : " + *vlanName)
	}
	id := strings.SplitAfter(*vlanName, "Vlan")
	log.Info("Extracted VLAN-Id = ", id[1])
	return id[1], nil
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

/* Validate whether member port exists in the member ports list and return the configured Interface mode */
func checkMemberPortExistsInListAndGetMode(d *db.DB, memberPortsList []string, memberPort *string, vlanName *string, ifMode *intfModeType) bool {
	for _, port := range memberPortsList {
		if *memberPort == port {
			tagModeEntry, err := d.GetEntry(&db.TableSpec{Name: "VLAN_MEMBER"}, db.Key{Comp: []string{*vlanName, *memberPort}})
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

/* Validate IPv4 address */
func validIPv4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if re.MatchString(ipAddress) {
		return true
	}
	return false
}

/* Validate IPv6 address */
func validIPv6(ip6Address string) bool {
	ip6Address = strings.Trim(ip6Address, " ")
	re, _ := regexp.Compile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`)
	if re.MatchString(ip6Address) {
		return true
	}
	return false
}

/* Get all the IP keys from INTERFACE table */
func (app *IntfApp) doGetAllIpKeys(d *db.DB, dbSpec *db.TableSpec) ([]db.Key, error) {

	var keys []db.Key

	intfTable, err := d.GetTable(dbSpec)
	if err != nil {
		return keys, err
	}

	keys, err = intfTable.GetKeys()
	log.Infof("Found %d INTF table keys", len(keys))
	return keys, err
}

/* Removal of untagged-vlan associated with interface, updates VLAN_MEMBER table and returns vlan*/
func (app *IntfApp) removeUntaggedVlanAndUpdateVlanMembTbl(d *db.DB, ifName *string) (*string, error) {
	if len(*ifName) == 0 {
		return nil, errors.New("Interface name is empty for fetching list of VLANs!")
	}

	var vlanMemberKeys []db.Key
	vlanMemberTable, err := d.GetTable(app.vlanD.vlanMemberTs)
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
		memberPortEntry, err := d.GetEntry(app.vlanD.vlanMemberTs, vlanMember)
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
		if tagMode == "untagged" {
			err = d.DeleteEntry(app.vlanD.vlanMemberTs, db.Key{Comp: []string{vlanMember.Get(0), *ifName}})
			if err != nil {
				return nil, err
			}
			// Disable STP configuration for ports which are removed from VLan membership
			var memberPorts []string
			memberPorts = append(memberPorts, *ifName)
			removeStpOnInterfaceSwitchportDeletion(d, memberPorts)

			return &vlanName, nil
		}
	}
	errStr := "Untagged VLAN configuration doesn't exist for Interface: " + *ifName
	return nil, tlerr.InvalidArgsError{Format: errStr}
}

/* Removal of tagged-vlan associated with interface and update VLAN_MEMBER table */
func (app *IntfApp) removeTaggedVlanAndUpdateVlanMembTbl(d *db.DB, trunkVlan *string, ifName *string) error {
	var err error
	memberPortEntry, err := d.GetEntry(app.vlanD.vlanMemberTs, db.Key{Comp: []string{*trunkVlan, *ifName}})
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
		err = d.DeleteEntry(app.vlanD.vlanMemberTs, db.Key{Comp: []string{*trunkVlan, *ifName}})
		if err != nil {
			return err
		}
		// Disable STP configuration for ports which are removed from VLan membership
		var memberPorts []string
		memberPorts = append(memberPorts, *ifName)
		removeStpOnInterfaceSwitchportDeletion(d, memberPorts)
	} else {
		vlanId := vlanName[len("Vlan"):len(vlanName)]
		errStr := "Tagged VLAN: " + vlanId + " configuration doesn't exist for Interface: " + *ifName
		return tlerr.InvalidArgsError{Format: errStr}
	}
	return err
}

/* Validate whether Port has any Untagged VLAN Config existing */
func (app *IntfApp) validateUntaggedVlanCfgredForIf(d *db.DB, ifName *string, accessVlan *string) (bool, error) {
	var err error

	var vlanMemberKeys []db.Key
	vlanMemberTable, err := d.GetTable(app.vlanD.vlanMemberTs)
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
		memberPortEntry, err := d.GetEntry(app.vlanD.vlanMemberTs, vlanMember)
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

/* Removes all VLAN_MEMBER table entries for Interface and Get list of VLANs */
func (app *IntfApp) removeAllVlanMembrsForIfAndGetVlans(d *db.DB, ifName *string, ifMode intfModeType) ([]string, error) {
	var err error
	var vlanKeys []db.Key
	vlanTable, err := d.GetTable(app.vlanD.vlanMemberTs)
	if err != nil {
		return nil, err
	}

	vlanKeys, err = vlanTable.GetKeys()
	var vlanSlice []string

	for _, vlanKey := range vlanKeys {
		if len(vlanKeys) < 2 {
			continue
		}
		if vlanKey.Get(1) == *ifName {
			entry, err := d.GetEntry(app.vlanD.vlanMemberTs, vlanKey)
			if err != nil {
				log.Errorf("Error found on fetching Vlan member info from App DB for Interface Name : %s", *ifName)
				return vlanSlice, err
			}
			tagInfo, ok := entry.Field["tagging_mode"]
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
				vlanSlice = append(vlanSlice, vlanKey.Get(0))
				d.DeleteEntry(app.vlanD.vlanMemberTs, vlanKey)
			}
		}
	}
	return vlanSlice, err
}

/* Removes the Interface name from Members list of VLAN table and updates it */
func (app *IntfApp) removeFromMembersListForVlan(d *db.DB, vlan *string, ifName *string) error {

	vlanEntry, err := d.GetEntry(app.vlanD.vlanTs, db.Key{Comp: []string{*vlan}})
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
		idx := 0
		memberFound := false

		for idxVal, memberName := range memberPortsList {
			if memberName == *ifName {
				memberFound = true
				idx = idxVal
				break
			}
		}
		if memberFound {
			memberPortsList = append(memberPortsList[:idx], memberPortsList[idx+1:]...)
			if len(memberPortsList) == 0 {
				log.Info("Deleting the members@")
				delete(vlanEntry.Field, "members@")
			} else {
				memberPortsStr := generateMemberPortsStringFromSlice(memberPortsList)
				log.Infof("Updated Member ports = %s for VLAN: %s", *memberPortsStr, *vlan)
				vlanEntry.Field["members@"] = *memberPortsStr
			}
			d.SetEntry(app.vlanD.vlanTs, db.Key{Comp: []string{*vlan}}, vlanEntry)
		} else {
			return nil
		}
	}
	return nil
}

/* Removes Interface name from Members-list for all VLANs from VLAN table and updates it */
func (app *IntfApp) removeFromMembersListForAllVlans(d *db.DB, ifName *string, vlanSlice []string) error {
	var err error

	for _, vlan := range vlanSlice {
		err = app.removeFromMembersListForVlan(d, &vlan, ifName)
		if err != nil {
			return err
		}
	}
	return err
}
