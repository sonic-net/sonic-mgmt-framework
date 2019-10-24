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
	"reflect"
	"strconv"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"
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

type ifVlan struct {
	ifName     *string
	mode       intfModeType
	accessVlan *string
	trunkVlans []string
}

/******* CONFIG FUNCTIONS ********/

func (app *IntfApp) translateUpdatePhyIntfSubInterfaces(d *db.DB, ifKey *string, intf *ocbinds.OpenconfigInterfaces_Interfaces_Interface) error {
	var err error
	if intf.Subinterfaces == nil {
		return err
	}
	subIf := intf.Subinterfaces.Subinterface[0]
	if subIf != nil {
		if subIf.Ipv4 != nil && subIf.Ipv4.Addresses != nil {
			for ip, _ := range subIf.Ipv4.Addresses.Address {
				addr := subIf.Ipv4.Addresses.Address[ip]
				if addr.Config != nil {
					log.Info("Ip:=", *addr.Config.Ip)
					log.Info("prefix:=", *addr.Config.PrefixLength)
					if !validIPv4(*addr.Config.Ip) {
						errStr := "Invalid IPv4 address " + *addr.Config.Ip
						err = tlerr.InvalidArgsError{Format: errStr}
						return err
					}
					err = app.translateIpv4(d, *ifKey, *addr.Config.Ip, int(*addr.Config.PrefixLength))
					if err != nil {
						return err
					}
				}
			}
		}
		if subIf.Ipv6 != nil && subIf.Ipv6.Addresses != nil {
			for ip, _ := range subIf.Ipv6.Addresses.Address {
				addr := subIf.Ipv6.Addresses.Address[ip]
				if addr.Config != nil {
					log.Info("Ip:=", *addr.Config.Ip)
					log.Info("prefix:=", *addr.Config.PrefixLength)
					if !validIPv6(*addr.Config.Ip) {
						errStr := "Invalid IPv6 address " + *addr.Config.Ip
						err = tlerr.InvalidArgsError{Format: errStr}
						return err
					}
					err = app.translateIpv4(d, *ifKey, *addr.Config.Ip, int(*addr.Config.PrefixLength))
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		err = errors.New("Only subinterface index 0 is supported")
		return err
	}
	return err
}

func (app *IntfApp) translateUpdatePhyIntfEthernet(d *db.DB, ifKey *string, intf *ocbinds.OpenconfigInterfaces_Interfaces_Interface) error {
	var err error

	if intf.Ethernet == nil {
		return err
	}
	if intf.Ethernet.SwitchedVlan == nil {
		return err
	}

	switchedVlanIntf := intf.Ethernet.SwitchedVlan
	if switchedVlanIntf.Config == nil {
		return err
	}

	if !app.validateIpCfgredForInterface(d, ifKey) {
		errStr := "Interface: " + *ifKey + ", IP address cannot exist with L2 modes"
		err = tlerr.InvalidArgsError{Format: errStr}
		return err
	}

	var accessVlanId uint16 = 0
	var trunkVlanSlice []string
	accessVlanFound := false
	trunkVlanFound := false

	/* Retrieve the Access VLAN Id */
	if switchedVlanIntf.Config.AccessVlan != nil {
		accessVlanId = *switchedVlanIntf.Config.AccessVlan
		log.Infof("Vlan id : %d observed for Member port addition configuration!", accessVlanId)
		accessVlanFound = true
	}

	/* Retrieve the list of trunk-vlans */
	if switchedVlanIntf.Config.TrunkVlans != nil {
		vlanUnionList := switchedVlanIntf.Config.TrunkVlans
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
		err = app.validateVlanExists(d, &accessVlan)
		if err != nil {
			errStr := "Invalid VLAN: " + strconv.Itoa(int(accessVlanId))
			err = tlerr.InvalidArgsError{Format: errStr}
			return err
		}
		var cfgredAccessVlan string
		exists, err := app.validateUntaggedVlanCfgredForIf(d, ifKey, &cfgredAccessVlan)
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
			err = tlerr.InvalidArgsError{Format: errStr}
			return err
		}
		memberPortEntryMap := make(map[string]string)
		memberPortEntry := db.Value{Field: memberPortEntryMap}
		memberPortEntry.Field["tagging_mode"] = "untagged"

		if app.vlanD.vlanMembersTableMap[accessVlan] == nil {
			app.vlanD.vlanMembersTableMap[accessVlan] = make(map[string]dbEntry)
		}
		app.vlanD.vlanMembersTableMap[accessVlan][*ifKey] = dbEntry{entry: memberPortEntry, op: opCreate}
		log.Info("Untagged Port added to cache!")
	}

TRUNKCONFIG:
	if trunkVlanFound {
		memberPortEntryMap := make(map[string]string)
		memberPortEntry := db.Value{Field: memberPortEntryMap}
		memberPortEntry.Field["tagging_mode"] = "tagged"
		for _, vlanId := range trunkVlanSlice {
			err = app.validateVlanExists(d, &vlanId)
			if err != nil {
				id := vlanId[len("Vlan"):len(vlanId)]
				errStr := "Invalid VLAN: " + id
				err = tlerr.InvalidArgsError{Format: errStr}
				return err
			}
			if app.vlanD.vlanMembersTableMap[vlanId] == nil {
				app.vlanD.vlanMembersTableMap[vlanId] = make(map[string]dbEntry)
			}
			app.vlanD.vlanMembersTableMap[vlanId][*ifKey] = dbEntry{entry: memberPortEntry, op: opCreate}
			log.Info("Tagged Port added to cache!")
		}
	}
	if accessVlanFound || trunkVlanFound {
		return err
	}

	log.Info("Request is for Configuring just the Mode for Interface: ", *ifKey)
	ifMode := switchedVlanIntf.Config.InterfaceMode

	switch ifMode {
	case ocbinds.OpenconfigVlan_VlanModeType_ACCESS:
		/* Configuring Interface Mode as ACCESS only without VLAN info*/
		app.mode = intfModeCfgAlone{ifName: *ifKey, mode: ACCESS}
		log.Info("Access Mode Config for Interface: ", *ifKey)
	case ocbinds.OpenconfigVlan_VlanModeType_TRUNK:
	}

	return err
}

func (app *IntfApp) translateUpdatePhyIntf(d *db.DB, ifKey *string, inpOp reqType) ([]db.WatchKeys, error) {

	var err error
	var keys []db.WatchKeys

	app.allIpKeys, _ = app.doGetAllIpKeys(d, app.intfD.intfIPTs)
	intfObj := app.getAppRootObject()
	intf := intfObj.Interface[*ifKey]
	curr, err := d.GetEntry(app.intfD.portTs, db.Key{Comp: []string{*ifKey}})
	if err != nil {
		errStr := "Invalid Interface:" + *ifKey
		ifValidErr := tlerr.InvalidArgsError{Format: errStr}
		return keys, ifValidErr
	}
	if !curr.IsPopulated() {
		log.Error("Interface ", *ifKey, " doesn't exist in DB")
		err = errors.New("Interface: " + *ifKey + " doesn't exist in DB")
		return keys, err
	}
	/* Handling Interface Config updates */
	app.translateUpdateIntfConfig(ifKey, intf, &curr)

	/* Handling Interface Ethernet updates */
	err = app.translateUpdatePhyIntfEthernet(d, ifKey, intf)
	if err != nil {
		return keys, err
	}

	/* Handling Interface SubInterfaces updates */
	err = app.translateUpdatePhyIntfSubInterfaces(d, ifKey, intf)
	if err != nil {
		return keys, err
	}
	return keys, err
}

func (app *IntfApp) processUpdatePhyIntfConfig(d *db.DB) error {
	var err error
	/* Updating the Interface Table */
	for ifName, ifEntry := range app.ifTableMap {
		if ifEntry.op == opUpdate {
			log.Info("Updating entry for ", ifName)
			err = d.SetEntry(app.intfD.portTs, db.Key{Comp: []string{ifName}}, ifEntry.entry)
			if err != nil {
				errStr := "Updating Interface table for Interface : " + ifName + " failed"
				return errors.New(errStr)
			}
		}
	}
	return err
}

func (app *IntfApp) processUpdatePhyIntfSubInterfaces(d *db.DB) error {
	var err error
	/* Updating the Interface IP table */
	for ifName, ipEntries := range app.intfD.ifIPTableMap {
		/* Note: when the interface comes as L2, before IP config, an INTERFACE table entry with
		   ifName alone has to be created, otherwise IP config wont take place. */
		ifEntry, err := d.GetEntry(app.intfD.intfIPTs, db.Key{Comp: []string{ifName}})
		if err != nil || !ifEntry.IsPopulated() {
			m := make(map[string]string)
			m["NULL"] = "NULL"
			err = d.CreateEntry(app.intfD.intfIPTs, db.Key{Comp: []string{ifName}}, db.Value{Field: m})
			if err != nil {
				return err
			}
			log.Infof("Created Interface entry with Interface name : %s alone!", ifName)
		}
		for ip, ipEntry := range ipEntries {
			if ipEntry.op == opCreate {
				log.Info("Creating entry for ", ifName, ":", ip)
				err = d.CreateEntry(app.intfD.intfIPTs, db.Key{Comp: []string{ifName, ip}}, ipEntry.entry)
				if err != nil {
					errStr := "Creating entry for " + ifName + ":" + ip + " failed"
					return errors.New(errStr)
				}
			} else if ipEntry.op == opDelete {
				log.Info("Deleting entry for ", ifName, ":", ip)
				err = d.DeleteEntry(app.intfD.intfIPTs, db.Key{Comp: []string{ifName, ip}})
				if err != nil {
					errStr := "Deleting entry for " + ifName + ":" + ip + " failed"
					return errors.New(errStr)
				}
			}
		}
	}
	return err
}

/* Adding member to VLAN requires updation of VLAN Table and VLAN Member Table */
func (app *IntfApp) processUpdatePhyIntfVlanAdd(d *db.DB) error {
	var err error
	var isMembersListUpdate bool

	/* Updating the VLAN member table */

	for vlanName, ifEntries := range app.vlanD.vlanMembersTableMap {
		var memberPortsListStrB strings.Builder
		var memberPortsList []string
		isMembersListUpdate = false

		vlanEntry, err := d.GetEntry(app.vlanD.vlanTs, db.Key{Comp: []string{vlanName}})
		if !vlanEntry.IsPopulated() {
			errStr := "Failed to retrieve memberPorts info of VLAN : " + vlanName
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
			/* Adding the following validation, just to avoid an another db-get in translate fn */
			/* Reason why it's ignored is, if we return, it leads to sync data issues between VlanT and VlanMembT */
			if memberPortsExists {
				var existingIfMode intfModeType
				if checkMemberPortExistsInListAndGetMode(d, memberPortsList, &ifName, &vlanName, &existingIfMode) {
					/* Since translib doesn't support rollback, we need to keep the DB consistent at this point,
					and throw the error message */
					var cfgReqIfMode intfModeType
					tagMode := ifEntry.entry.Field["tagging_mode"]
					convertTaggingModeToInterfaceModeType(&tagMode, &cfgReqIfMode)

					if cfgReqIfMode == existingIfMode {
						continue
					} else {
						vlanEntry.Field["members@"] = memberPortsListStrB.String()
						err = d.SetEntry(app.vlanD.vlanTs, db.Key{Comp: []string{vlanName}}, vlanEntry)

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
			switch ifEntry.op {
			case opCreate:
				err = d.CreateEntry(app.vlanD.vlanMemberTs, db.Key{Comp: []string{vlanName, ifName}}, ifEntry.entry)
				if err != nil {
					errStr := "Creating entry for VLAN member table with vlan : " + vlanName + " If : " + ifName + " failed"
					return errors.New(errStr)
				}
			case opUpdate:
				err = d.SetEntry(app.vlanD.vlanMemberTs, db.Key{Comp: []string{vlanName, ifName}}, ifEntry.entry)
				if err != nil {
					errStr := "Set entry for VLAN member table with vlan : " + vlanName + " If : " + ifName + " failed"
					return errors.New(errStr)
				}
			}
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
		vlanEntry.Field["members@"] = memberPortsListStrB.String()

		err = d.SetEntry(app.vlanD.vlanTs, db.Key{Comp: []string{vlanName}}, vlanEntry)
		if err != nil {
			return errors.New("Updating VLAN table with member ports failed")
		}
	}
	return err
}

func (app *IntfApp) updateAccessModeConfig(d *db.DB, ifName *string) error {
	var err error

	if len(*ifName) == 0 {
		return errors.New("Empty Interface name received!")
	}

	vlanList, err := app.removeAllVlanMembrsForIfAndGetVlans(d, ifName, ACCESS)
	if err != nil {
		return err
	}

	err = app.removeFromMembersListForAllVlans(d, ifName, vlanList)
	if err != nil {
		return err
	}
	return err
}

func (app *IntfApp) processUpdateInterfaceModeConfig(d *db.DB, ifName *string) error {
	var err error
	switch app.mode.mode {
	case ACCESS:
		err := app.updateAccessModeConfig(d, &app.mode.ifName)
		if err != nil {
			return err
		}
	case TRUNK:
	case MODE_UNSET:
		break
	}
	return err
}

func (app *IntfApp) processUpdatePhyIntf(d *db.DB) error {
	var err error
	err = app.processUpdatePhyIntfConfig(d)
	if err != nil {
		return err
	}

	err = app.processUpdatePhyIntfSubInterfaces(d)
	if err != nil {
		return err
	}

	err = app.processUpdatePhyIntfVlanAdd(d)
	if err != nil {
		return err
	}

	/* Switchport access/trunk mode config without VLAN */
	/* This mode will be set in the translate fn, when request is just for mode without VLAN info. */
	if app.mode.mode != MODE_UNSET {
		err = app.processUpdateInterfaceModeConfig(d, &app.mode.ifName)
		if err != nil {
			return err
		}
	}
	return err
}

/******* DELETE FUNCTIONS ********/

/* Note: Reason why we don't use multi-map, which we use for config is because RESTCONF doesn't supply the access-vlan value
 * or it will give only the single instance of trunk-vlan for deletion */
func (app *IntfApp) translateDeletePhyIntfEthernetSwitchedVlan(d *db.DB, switchedVlanIntf *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan, ifName *string) error {
	var err error
	var ifVlanInfo ifVlan

	if switchedVlanIntf.Config != nil {
		if switchedVlanIntf.Config.AccessVlan != nil {
			ifVlanInfo.mode = ACCESS
		}
		if switchedVlanIntf.Config.TrunkVlans != nil {
			trunkVlansUnionList := switchedVlanIntf.Config.TrunkVlans
			ifVlanInfo.mode = TRUNK
			for _, trunkVlanUnion := range trunkVlansUnionList {
				trunkVlanUnionType := reflect.TypeOf(trunkVlanUnion).Elem()

				switch trunkVlanUnionType {

				case reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_String{}):
					val := (trunkVlanUnion).(*ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_Config_TrunkVlans_Union_String)
					vlanName := "Vlan" + val.String
					err = app.validateVlanExists(d, &vlanName)
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
		if ifVlanInfo.mode != MODE_UNSET {
			ifVlanInfo.ifName = ifName
			app.intfD.ifVlanInfoList = append(app.intfD.ifVlanInfoList, &ifVlanInfo)
		}
	}
	return err
}

func (app *IntfApp) translateDeletePhyIntfEthernet(d *db.DB, intf *ocbinds.OpenconfigInterfaces_Interfaces_Interface, ifName *string) error {
	var err error
	if intf.Ethernet == nil {
		return err
	}
	if intf.Ethernet.SwitchedVlan == nil {
		return err
	}
	switchedVlanIntf := intf.Ethernet.SwitchedVlan
	err = app.translateDeletePhyIntfEthernetSwitchedVlan(d, switchedVlanIntf, ifName)
	if err != nil {
		return err
	}
	return err
}

func (app *IntfApp) translateDeletePhyIntfSubInterfaces(d *db.DB, intf *ocbinds.OpenconfigInterfaces_Interfaces_Interface, ifName *string) error {
	var err error
	if intf.Subinterfaces == nil {
		return err
	}
	subIf := intf.Subinterfaces.Subinterface[0]
	if subIf != nil {
		if subIf.Ipv4 != nil && subIf.Ipv4.Addresses != nil {
			for ip, _ := range subIf.Ipv4.Addresses.Address {
				addr := subIf.Ipv4.Addresses.Address[ip]
				if addr != nil {
					ipAddr := addr.Ip
					log.Info("IPv4 address = ", *ipAddr)
					if !validIPv4(*ipAddr) {
						errStr := "Invalid IPv4 address " + *ipAddr
						ipValidErr := tlerr.InvalidArgsError{Format: errStr}
						return ipValidErr
					}
					err = app.validateIp(d, *ifName, *ipAddr)
					if err != nil {
						errStr := "Invalid IPv4 address " + *ipAddr
						ipValidErr := tlerr.InvalidArgsError{Format: errStr}
						return ipValidErr
					}
				}
			}
		}
		if subIf.Ipv6 != nil && subIf.Ipv6.Addresses != nil {
			for ip, _ := range subIf.Ipv6.Addresses.Address {
				addr := subIf.Ipv6.Addresses.Address[ip]
				if addr != nil {
					ipAddr := addr.Ip
					log.Info("IPv6 address = ", *ipAddr)
					if !validIPv6(*ipAddr) {
						errStr := "Invalid IPv6 address " + *ipAddr
						ipValidErr := tlerr.InvalidArgsError{Format: errStr}
						return ipValidErr
					}
					err = app.validateIp(d, *ifName, *ipAddr)
					if err != nil {
						errStr := "Invalid IPv6 address:" + *ipAddr
						ipValidErr := tlerr.InvalidArgsError{Format: errStr}
						return ipValidErr
					}
				}
			}
		}
	} else {
		err = errors.New("Only subinterface index 0 is supported")
		return err
	}
	return err
}

func (app *IntfApp) translateDeletePhyIntf(d *db.DB, ifName *string) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys

	intfObj := app.getAppRootObject()
	intf := intfObj.Interface[*ifName]

	err = app.translateDeletePhyIntfSubInterfaces(d, intf, ifName)
	if err != nil {
		return keys, err
	}

	err = app.translateDeletePhyIntfEthernet(d, intf, ifName)
	if err != nil {
		return keys, err
	}

	return keys, err
}

func (app *IntfApp) processDeletePhyIntfSubInterfaces(d *db.DB) error {
	var err error

	for ifKey, entrylist := range app.intfD.ifIPTableMap {
		for ip, _ := range entrylist {
			err = d.DeleteEntry(app.intfD.intfIPTs, db.Key{Comp: []string{ifKey, ip}})
			if err != nil {
				return err
			}
			log.Infof("Deleted IP : %s for Interface : %s", ip, ifKey)
		}
	}
	return err
}

func (app *IntfApp) processDeletePhyIntfVlanRemoval(d *db.DB) error {
	var err error

	if len(app.intfD.ifVlanInfoList) == 0 {
		log.Info("No VLAN Info present for membership removal!")
		return nil
	}

	for _, ifVlanInfo := range app.intfD.ifVlanInfoList {
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
			untagdVlan, err := app.removeUntaggedVlanAndUpdateVlanMembTbl(d, ifName)
			if err != nil {
				return err
			}
			if untagdVlan != nil {
				app.removeFromMembersListForVlan(d, untagdVlan, ifName)
			}

		case TRUNK:
			/* Handling trunk-vlans delete */
			log.Info("Trunk VLAN Delete!")
			if trunkVlans != nil {
				for _, trunkVlan := range trunkVlans {
					err = app.removeTaggedVlanAndUpdateVlanMembTbl(d, &trunkVlan, ifName)
					if err != nil {
						return err
					}
					app.removeFromMembersListForVlan(d, &trunkVlan, ifName)
				}
			}
		}
	}
	return err
}

func (app *IntfApp) processDeletePhyIntf(d *db.DB) error {
	var err error

	err = app.processDeletePhyIntfSubInterfaces(d)
	if err != nil {
		return err
	}

	err = app.processDeletePhyIntfVlanRemoval(d)
	if err != nil {
		return err
	}
	return err
}

/******** SUBSCRIBE FUNCTIONS ******/

func (app *IntfApp) translateSubscribePhyIntf(ifKey *string, pInfo *PathInfo) (*notificationOpts, *notificationInfo, error) {
	var err error
	notifInfo := notificationInfo{dbno: db.ApplDB}

	err = app.validateInterface(app.appDB, *ifKey, db.Key{Comp: []string{*ifKey}})
	if err != nil {
		return nil, nil, err
	}
	if pInfo.HasSuffix("/state/oper-status") {
		notifInfo.table = db.TableSpec{Name: "PORT_TABLE"}
		notifInfo.key = asKey(*ifKey)
		notifInfo.needCache = true
		return &notificationOpts{pType: OnChange}, &notifInfo, nil
	}
	return nil, nil, err
}
