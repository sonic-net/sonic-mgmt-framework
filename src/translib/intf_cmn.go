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
	"github.com/openconfig/ygot/ygot"
	"net"
	"strconv"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"
	"unsafe"
)

const (
	PORT         = "PORT"
	INDEX        = "index"
	MTU          = "mtu"
	ADMIN_STATUS = "admin_status"
	SPEED        = "speed"
	DESC         = "description"
	OPER_STATUS  = "oper_status"
	NAME         = "name"
	ACTIVE       = "active"
)

type Table int

const (
	IF_TABLE_MAP Table = iota
	PORT_STAT_MAP
)

func (app *IntfApp) translateUpdateIntfConfig(ifKey *string, intf *ocbinds.OpenconfigInterfaces_Interfaces_Interface, curr *db.Value) {
	if intf.Config != nil {
		if intf.Config.Description != nil {
			curr.Field["description"] = *intf.Config.Description
		} else if intf.Config.Mtu != nil {
			curr.Field["mtu"] = strconv.Itoa(int(*intf.Config.Mtu))
		} else if intf.Config.Enabled != nil {
			if *intf.Config.Enabled == true {
				curr.Field["admin_status"] = "up"
			} else {
				curr.Field["admin_status"] = "down"
			}
		}
	}
	log.Info("Writing to db for ", *ifKey)
	app.ifTableMap[*ifKey] = dbEntry{op: opUpdate, entry: *curr}
}

/* Handling IP address updates for given interface */
func (app *IntfApp) translateUpdateIntfSubInterfaces(d *db.DB, ifKey *string, intf *ocbinds.OpenconfigInterfaces_Interfaces_Interface) error {
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

/* Handling IP address configuration for given Interface */
func (app *IntfApp) processUpdateIntfSubInterfaces(d *db.DB) error {
	var err error
	/* Updating the table */
	for ifName, ipEntries := range app.ifIPTableMap {
		ts := app.intfD.intfIPTs
		if app.intfType == LAG {
			ts = app.lagD.lagIPTs
		}
		m := make(map[string]string)
		m["NULL"] = "NULL"
		ifEntry, err := d.GetEntry(ts, db.Key{Comp: []string{ifName}})
		if err != nil || !ifEntry.IsPopulated() {
			err = d.CreateEntry(ts, db.Key{Comp: []string{ifName}}, db.Value{Field: m})
			if err != nil {
				log.Info("Failed to create Interface entry with Interface name")
				return err
			}
			log.Infof("Created Interface entry with Interface name : %s alone!", ifName)
		}
		for ip, ipEntry := range ipEntries {
			if ipEntry.op == opCreate {
				log.Info("Creating entry for ", ifName, ":", ip)
				err = d.CreateEntry(ts, db.Key{Comp: []string{ifName, ip}}, ipEntry.entry)
				if err != nil {
					errStr := "Creating entry for " + ifName + ":" + ip + " failed"
					return errors.New(errStr)
				}
			} else if ipEntry.op == opDelete {
				log.Info("Deleting entry for ", ifName, ":", ip)
				err = d.DeleteEntry(ts, db.Key{Comp: []string{ifName, ip}})
				if err != nil {
					errStr := "Deleting entry for " + ifName + ":" + ip + " failed"
					return errors.New(errStr)
				}
			}
		}
	}
	return err
}

func (app *IntfApp) translateDeleteIntfSubInterfaces(d *db.DB, intf *ocbinds.OpenconfigInterfaces_Interfaces_Interface, ifName *string) error {
	log.Info("Inside translateDeleteIntfSubInterfaces")
	var err error
	if intf.Subinterfaces == nil {
		return err
	}
	err = app.getIntfTypeFromIntf(ifName)
	if err != nil {
		return err
	}
	/* Find the type of Interface*/
	ts := app.intfD.intfIPTs
	if app.intfType == LAG {
		ts = app.lagD.lagIPTs
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
					err = app.validateIp(d, *ifName, *ipAddr, ts)
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
					err = app.validateIp(d, *ifName, *ipAddr, ts)
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

func (app *IntfApp) getSpecificIfStateAttr(targetUriPath *string, ifKey *string, oc_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State) (bool, error) {
	switch *targetUriPath {

	case "/openconfig-interfaces:interfaces/interface/state/oper-status":
		val, e := app.getIntfAttr(ifKey, OPER_STATUS, IF_TABLE_MAP)
		if len(val) > 0 {
			switch val {
			case "up":
				oc_val.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_UP
			case "down":
				oc_val.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_DOWN
			default:
				oc_val.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_UNSET
			}
			return true, nil
		} else {
			return true, e
		}
	case "/openconfig-interfaces:interfaces/interface/state/admin-status":
		val, e := app.getIntfAttr(ifKey, ADMIN_STATUS, IF_TABLE_MAP)
		if len(val) > 0 {
			switch val {
			case "up":
				oc_val.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_UP
			case "down":
				oc_val.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_DOWN
			default:
				oc_val.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_UNSET
			}
			return true, nil
		} else {
			return true, e
		}
	case "/openconfig-interfaces:interfaces/interface/state/mtu":
		val, e := app.getIntfAttr(ifKey, MTU, IF_TABLE_MAP)
		if len(val) > 0 {
			v, e := strconv.ParseUint(val, 10, 16)
			if e == nil {
				oc_val.Mtu = (*uint16)(unsafe.Pointer(&v))
				return true, nil
			}
		}
		return true, e
	case "/openconfig-interfaces:interfaces/interface/state/ifindex":
		val, e := app.getIntfAttr(ifKey, INDEX, IF_TABLE_MAP)
		if len(val) > 0 {
			v, e := strconv.ParseUint(val, 10, 32)
			if e == nil {
				oc_val.Ifindex = (*uint32)(unsafe.Pointer(&v))
				return true, nil
			}
		}
		return true, e
	case "/openconfig-interfaces:interfaces/interface/state/description":
		val, e := app.getIntfAttr(ifKey, DESC, IF_TABLE_MAP)
		if e == nil {
			oc_val.Description = &val
			return true, nil
		}
		return true, e
	default:
		log.Infof(*targetUriPath + " - Not an interface state attribute")
	}
	return false, nil
}

func (app *IntfApp) getSpecificIfVlanAttr(targetUriPath *string, ifKey *string, oc_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_State) (bool, error) {
	switch *targetUriPath {
	case "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/state/access-vlan":
		_, accessVlanName, e := app.getIntfVlanAttr(ifKey, ACCESS)
		if e != nil {
			return true, e
		}
		if accessVlanName == nil {
			return true, nil
		}
		vlanName := *accessVlanName
		vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
		vlanId, err := strconv.Atoi(vlanIdStr)
		if err != nil {
			errStr := "Conversion of string to int failed for " + vlanIdStr
			return true, errors.New(errStr)
		}
		vlanIdCast := uint16(vlanId)

		oc_val.AccessVlan = &vlanIdCast
		return true, nil
	case "/openconfig-interfaces:interfaces/interface/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/state/trunk-vlans":
		trunkVlans, _, e := app.getIntfVlanAttr(ifKey, TRUNK)
		if e != nil {
			return true, e
		}
		for _, vlanName := range trunkVlans {
			vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
			vlanId, err := strconv.Atoi(vlanIdStr)
			if err != nil {
				errStr := "Conversion of string to int failed for " + vlanIdStr
				return true, errors.New(errStr)
			}
			vlanIdCast := uint16(vlanId)

			trunkVlan, _ := oc_val.To_OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_State_TrunkVlans_Union(vlanIdCast)
			oc_val.TrunkVlans = append(oc_val.TrunkVlans, trunkVlan)
		}
		return true, nil
	}
	return false, nil
}

func (app *IntfApp) getSpecificIfLagAttr(d *db.DB, targetUriPath *string, ifKey *string, oc_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_State) (bool, error) {
	switch *targetUriPath {
	case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/state/min-links":
		curr, err := d.GetEntry(app.lagD.lagTs, db.Key{Comp: []string{*ifKey}})
		if err != nil {
			errStr := "Failed to Get PortChannel details"
			return true, errors.New(errStr)
		}
		if val, ok := curr.Field["min_links"]; ok {
			log.Info("curr.Field['min_links']", val)
			min_links, err := strconv.Atoi(curr.Field["min_links"])
			if err != nil {
				errStr := "Conversion of string to int failed"
				return true, errors.New(errStr)
			}
			links := uint16(min_links)
			oc_val.MinLinks = &links
		} else {
			log.Info("Minlinks set to 0 (dafault value)")
			links := uint16(0)
			oc_val.MinLinks = &links
		}
		return true, nil
	case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/state/member":
		lagKeys, err := d.GetKeys(app.lagD.lagMemberTs)
		if err != nil {
			log.Info("No entries in PORTCHANNEL_MEMBER TABLE")
			return true, err
		}
		var flag bool = false
		for i, _ := range lagKeys {
			if *ifKey == lagKeys[i].Get(0) {
				log.Info("Found lagKey")
				flag = true
				ethName := lagKeys[i].Get(1)
				oc_val.Member = append(oc_val.Member, ethName)
			}
		}
		if flag == false {
			log.Info("Given PortChannel has no members")
			errStr := "Given PortChannel has no members"
			return true, errors.New(errStr)
		}
		return true, nil
	case "/openconfig-interfaces:interfaces/interface/openconfig-if-aggregate:aggregation/state/dell-intf-augments:fallback":
		curr, err := d.GetEntry(app.lagD.lagTs, db.Key{Comp: []string{*ifKey}})
		if err != nil {
			errStr := "Failed to Get PortChannel details"
			return true, errors.New(errStr)
		}
		if val, ok := curr.Field["fallback"]; ok {
			log.Info("curr.Field['fallback']", val)
			fallbackVal, err := strconv.ParseBool(val)
			if err != nil {
				errStr := "Conversion of string to bool failed"
				return true, errors.New(errStr)
			}
			oc_val.Fallback = &fallbackVal
		} else {
			log.Info("Fallback set to False, default value")
			fallbackVal := false
			oc_val.Fallback = &fallbackVal
		}
		return true, nil

	default:
		log.Infof(*targetUriPath + " - Not an supported Get attribute")
	}
	return false, nil
}

func (app *IntfApp) getSpecificCounterAttr(targetUriPath *string, ifKey *string, counter_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters) (bool, error) {

	var e error

	switch *targetUriPath {
	case "/openconfig-interfaces:interfaces/interface/state/counters/in-octets":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_OCTETS", &counter_val.InOctets)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-unicast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_UCAST_PKTS", &counter_val.InUnicastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-broadcast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_BROADCAST_PKTS", &counter_val.InBroadcastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-multicast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_MULTICAST_PKTS", &counter_val.InMulticastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-errors":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_ERRORS", &counter_val.InErrors)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-discards":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_DISCARDS", &counter_val.InDiscards)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-pkts":
		var inNonUCastPkt, inUCastPkt *uint64
		var in_pkts uint64

		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_NON_UCAST_PKTS", &inNonUCastPkt)
		if e == nil {
			e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_UCAST_PKTS", &inUCastPkt)
			if e != nil {
				return true, e
			}
			in_pkts = *inUCastPkt + *inNonUCastPkt
			counter_val.InPkts = &in_pkts
			return true, e
		} else {
			return true, e
		}

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-octets":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_OCTETS", &counter_val.OutOctets)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-unicast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_UCAST_PKTS", &counter_val.OutUnicastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-broadcast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_BROADCAST_PKTS", &counter_val.OutBroadcastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-multicast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_MULTICAST_PKTS", &counter_val.OutMulticastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-errors":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_ERRORS", &counter_val.OutErrors)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-discards":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_DISCARDS", &counter_val.OutDiscards)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-pkts":
		var outNonUCastPkt, outUCastPkt *uint64
		var out_pkts uint64

		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_NON_UCAST_PKTS", &outNonUCastPkt)
		if e == nil {
			e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_UCAST_PKTS", &outUCastPkt)
			if e != nil {
				return true, e
			}
			out_pkts = *outUCastPkt + *outNonUCastPkt
			counter_val.OutPkts = &out_pkts
			return true, e
		} else {
			return true, e
		}

	default:
		log.Infof(*targetUriPath + " - Not an interface state counter attribute")
	}
	return false, nil
}

func (app *IntfApp) getCounters(ifKey *string, attr string, counter_val **uint64) error {
	val, e := app.getIntfAttr(ifKey, attr, PORT_STAT_MAP)
	if len(val) > 0 {
		v, e := strconv.ParseUint(val, 10, 64)
		if e == nil {
			*counter_val = &v
			return nil
		}
	}
	return e
}

func (app *IntfApp) getIntfAttr(ifName *string, attr string, table Table) (string, error) {

	var ok bool = false
	var entry dbEntry

	if table == IF_TABLE_MAP {
		entry, ok = app.ifTableMap[*ifName]
	} else if table == PORT_STAT_MAP {
		entry, ok = app.intfD.portStatMap[*ifName]
	} else {
		return "", errors.New("Unsupported table")
	}

	if ok {
		ifData := entry.entry

		if val, ok := ifData.Field[attr]; ok {
			return val, nil
		}
	}
	return "", errors.New("Attr " + attr + "doesn't exist in IF table Map!")
}

func (app *IntfApp) getIntfVlanAttr(ifName *string, ifMode intfModeType) ([]string, *string, error) {

	vlanEntries, ok := app.vlanD.vlanMembersTableMap[*ifName]
	if !ok {
		errStr := "Cannot find info for Interface: " + *ifName + " from VLAN_MEMBERS_TABLE!"
		return nil, nil, errors.New(errStr)
	}
	switch ifMode {
	case ACCESS:
		for vlanKey, tagEntry := range vlanEntries {
			tagMode, ok := tagEntry.entry.Field["tagging_mode"]
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
			tagMode, ok := tagEntry.entry.Field["tagging_mode"]
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

func (app *IntfApp) processGetSpecificAttr(targetUriPath *string, ifKey *string) (bool, *GetResponse, error) {
	var err error
	var payload []byte

	/*Check if the request is for a specific attribute in Interfaces state container*/
	ocStateVal := &ocbinds.OpenconfigInterfaces_Interfaces_Interface_State{}
	ok, e := app.getSpecificIfStateAttr(targetUriPath, ifKey, ocStateVal)
	if ok {
		if e != nil {
			return ok, &(GetResponse{Payload: payload, ErrSrc: AppErr}), e
		}
		payload, err = dumpIetfJson(ocStateVal, false)
		if err != nil {
			return ok, &(GetResponse{Payload: payload, ErrSrc: AppErr}), err
		}
		return ok, &(GetResponse{Payload: payload}), err

	}

	/*Check if the request is for a specific attribute in Interfaces state COUNTERS container*/
	counter_val := &ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters{}
	ok, e = app.getSpecificCounterAttr(targetUriPath, ifKey, counter_val)
	if ok {
		if e != nil {
			return ok, &(GetResponse{Payload: payload, ErrSrc: AppErr}), e
		}

		payload, err = dumpIetfJson(counter_val, false)
		if err != nil {
			return ok, &(GetResponse{Payload: payload, ErrSrc: AppErr}), err
		}
		return ok, &(GetResponse{Payload: payload}), err
	}

	/*Check if the request is for a specific attribute in Interfaces Ethernet container*/
	ocEthernetVlanStateVal := &ocbinds.OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_State{}
	ok, e = app.getSpecificIfVlanAttr(targetUriPath, ifKey, ocEthernetVlanStateVal)
	if ok {
		if e != nil {
			return ok, &(GetResponse{Payload: payload, ErrSrc: AppErr}), e
		}
		payload, err = dumpIetfJson(ocEthernetVlanStateVal, false)
		if err != nil {
			return ok, &(GetResponse{Payload: payload, ErrSrc: AppErr}), err
		}
		return ok, &(GetResponse{Payload: payload}), err
	}
	/*Check if the request is for a specific attribute in Interfaces Aggregation container*/
	ocEthernetLagStateVal := &ocbinds.OpenconfigInterfaces_Interfaces_Interface_Aggregation_State{}
	ok, e = app.getSpecificIfLagAttr(app.configDB, targetUriPath, ifKey, ocEthernetLagStateVal)
	if ok {
		if e != nil {
			return ok, &(GetResponse{Payload: payload, ErrSrc: AppErr}), e
		}
		payload, err = dumpIetfJson(ocEthernetLagStateVal, false)
		if err != nil {
			return ok, &(GetResponse{Payload: payload, ErrSrc: AppErr}), err
		}
		return ok, &(GetResponse{Payload: payload}), err
	}
	return ok, &(GetResponse{Payload: payload}), err
}

func (app *IntfApp) getPortOidMapForCounters(dbCl *db.DB) error {
	var err error
	ifCountInfo, err := dbCl.GetMapAll(app.intfD.portOidCountrTblTs)
	if err != nil {
		log.Error("Port-OID (Counters) get for all the interfaces failed!")
		return err
	}
	if ifCountInfo.IsPopulated() {
		app.intfD.portOidMap.entry = ifCountInfo
	} else {
		return errors.New("Get for OID info from all the interfaces from Counters DB failed!")
	}
	return err
}

func (app *IntfApp) convertDBIntfCounterInfoToInternal(dbCl *db.DB, ifName *string) error {
	var err error

	if len(*ifName) > 0 {
		oid := app.intfD.portOidMap.entry.Field[*ifName]
		log.Infof("OID : %s received for Interface : %s", oid, *ifName)

		/* Get the statistics for the port */
		var ifStatKey db.Key
		ifStatKey.Comp = []string{oid}

		ifStatInfo, err := dbCl.GetEntry(app.intfD.intfCountrTblTs, ifStatKey)
		if err != nil {
			log.Infof("Fetching port-stat for port : %s failed!", *ifName)
			return err
		}
		app.intfD.portStatMap[*ifName] = dbEntry{entry: ifStatInfo}
	} else {
		for ifName, _ := range app.ifTableMap {
			app.convertDBIntfCounterInfoToInternal(dbCl, &ifName)
		}
	}
	return err
}

func (app *IntfApp) convertDBIfVlanListInfoToInternal(dbCl *db.DB, ts *db.TableSpec, ifName *string) error {
	var err error

	vlanMemberTable, err := dbCl.GetTable(ts)
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

		memberPortEntry, err := dbCl.GetEntry(ts, vlanMember)
		if err != nil {
			return err
		}
		if !memberPortEntry.IsPopulated() {
			errStr := "Tagging Info not present for Vlan: " + vlanId + " Interface: " + ifName + " from VLAN_MEMBER_TABLE"
			return errors.New(errStr)
		}

		/* vlanMembersTableMap is used as DS for ifName to list of VLANs */
		if app.vlanD.vlanMembersTableMap[ifName] == nil {
			app.vlanD.vlanMembersTableMap[ifName] = make(map[string]dbEntry)
			app.vlanD.vlanMembersTableMap[ifName][vlanId] = dbEntry{entry: memberPortEntry}
		} else {
			app.vlanD.vlanMembersTableMap[ifName][vlanId] = dbEntry{entry: memberPortEntry}
		}
	}
	log.Infof("Updated the vlan-member-table ds for Interface: %s", *ifName)
	return err
}

func (app *IntfApp) convertDBIntfInfoToInternal(dbCl *db.DB, ts *db.TableSpec, ifName *string, ifKey db.Key) error {

	var err error
	/* Fetching DB data for a specific Interface */
	if len(*ifName) > 0 {
		log.Info("Updating Interface info from APP-DB to Internal DS for Interface name : ", *ifName)
		ifInfo, err := dbCl.GetEntry(ts, ifKey)
		if err != nil {
			log.Errorf("Error found on fetching Interface info from App DB for If Name : %s", *ifName)
			errStr := "Invalid Interface:" + *ifName
			err = tlerr.InvalidArgsError{Format: errStr}
			return err
		}
		if ifInfo.IsPopulated() {
			log.Info("Interface Info populated for ifName : ", *ifName)
			app.ifTableMap[*ifName] = dbEntry{entry: ifInfo}
		} else {
			return errors.New("Populating Interface info for " + *ifName + "failed")
		}
	} else {
		log.Info("App-DB get for all the interfaces")
		tbl, err := dbCl.GetTable(ts)
		if err != nil {
			log.Error("App-DB get for list of interfaces failed!")
			return err
		}
		keys, _ := tbl.GetKeys()
		for _, key := range keys {
			ifName := key.Get(0)
			app.convertDBIntfInfoToInternal(dbCl, ts, &(ifName), db.Key{Comp: []string{key.Get(0)}})
		}
	}
	return err
}

func (app *IntfApp) convertDBIntfIPInfoToInternal(dbCl *db.DB, ts *db.TableSpec, ifName *string) error {

	var err error
	log.Info("Updating Interface IP Info from APP-DB to Internal DS for Interface Name : ", *ifName)
	app.allIpKeys, _ = app.doGetAllIpKeys(dbCl, ts)

	for _, key := range app.allIpKeys {
		if len(key.Comp) <= 1 {
			continue
		}
		ipInfo, err := dbCl.GetEntry(ts, key)
		if err != nil {
			log.Errorf("Error found on fetching Interface IP info from App DB for Interface Name : %s", *ifName)
			return err
		}
		if len(app.ifIPTableMap[key.Get(0)]) == 0 {
			app.ifIPTableMap[key.Get(0)] = make(map[string]dbEntry)
			app.ifIPTableMap[key.Get(0)][key.Get(1)] = dbEntry{entry: ipInfo}
		} else {
			app.ifIPTableMap[key.Get(0)][key.Get(1)] = dbEntry{entry: ipInfo}
		}
	}
	return err
}

func (app *IntfApp) processGetConvertDBPhyIfInfoToDS(ifName *string) error {
	var err error

	err = app.convertDBIntfInfoToInternal(app.appDB, app.intfD.portTblTs, ifName, db.Key{Comp: []string{*ifName}})
	if err != nil {
		return err
	}

	err = app.convertDBIfVlanListInfoToInternal(app.appDB, app.vlanD.vlanMemberTblTs, ifName)
	if err != nil {
		return err
	}

	err = app.getPortOidMapForCounters(app.countersDB)
	if err != nil {
		return err
	}
	err = app.convertDBIntfCounterInfoToInternal(app.countersDB, ifName)
	if err != nil {
		return err
	}

	err = app.convertDBIntfIPInfoToInternal(app.appDB, app.intfD.intfIPTblTs, ifName)
	if err != nil {
		return err
	}
	return err
}

func (app *IntfApp) processGetConvertDBVlanIfInfoToDS(vlanName *string) error {
	var err error

	err = app.convertDBIntfInfoToInternal(app.appDB, app.vlanD.vlanTblTs, vlanName, db.Key{Comp: []string{*vlanName}})
	if err != nil {
		return err
	}
	return err
}

func (app *IntfApp) processGetConvertDBLagIfInfoToDS(lagName *string) error {
	var err error

	err = app.convertDBIntfInfoToInternal(app.appDB, app.lagD.lagTblTs, lagName, db.Key{Comp: []string{*lagName}})
	if err != nil {
		return err
	}
	return err
}

func (app *IntfApp) processGetConvertDBIfInfoToDS(ifName *string) error {
	var err error

	err = app.processGetConvertDBPhyIfInfoToDS(ifName)
	if err != nil {
		return err
	}

	err = app.processGetConvertDBVlanIfInfoToDS(ifName)
	if err != nil {
		return err
	}
	return err
}

func (app *IntfApp) convertInternalToOCIntfAttrInfo(ifName *string, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface) {

	/* Handling the Interface attributes */
	if ifInfo.Config == nil || ifInfo.State == nil {
		return
	}
	if entry, ok := app.ifTableMap[*ifName]; ok {
		ifData := entry.entry

		name := *ifName
		ifInfo.Config.Name = &name
		ifInfo.State.Name = &name

		for ifAttr := range ifData.Field {
			switch ifAttr {
			case ADMIN_STATUS:
				adminStatus := ifData.Get(ifAttr)
				ifInfo.State.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_DOWN
				if adminStatus == "up" {
					ifInfo.State.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_UP
				}
			case OPER_STATUS:
				operStatus := ifData.Get(ifAttr)
				ifInfo.State.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_DOWN
				if operStatus == "up" {
					ifInfo.State.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_UP
				}
			case DESC:
				descVal := ifData.Get(ifAttr)
				descr := new(string)
				*descr = descVal
				ifInfo.Config.Description = descr
				ifInfo.State.Description = descr
			case MTU:
				mtuStr := ifData.Get(ifAttr)
				mtuVal, err := strconv.Atoi(mtuStr)
				mtu := new(uint16)
				*mtu = uint16(mtuVal)
				if err == nil {
					ifInfo.Config.Mtu = mtu
					ifInfo.State.Mtu = mtu
				}
			case SPEED:
				speed := ifData.Get(ifAttr)
				var speedEnum ocbinds.E_OpenconfigIfEthernet_ETHERNET_SPEED

				switch speed {
				case "2500":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_2500MB
				case "1000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_1GB
				case "5000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_5GB
				case "10000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB
				case "25000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_25GB
				case "40000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_40GB
				case "50000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_50GB
				case "100000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_100GB
				default:
					log.Infof("Not supported speed: %s!", speed)
				}
				if ifInfo.Ethernet != nil {
					if ifInfo.Ethernet.Config != nil {
						ifInfo.Ethernet.Config.PortSpeed = speedEnum
					}
					if ifInfo.Ethernet.State != nil {
						ifInfo.Ethernet.State.PortSpeed = speedEnum
					}
				}
			case INDEX:
				ifIdxStr := ifData.Get(ifAttr)
				ifIdxNum, err := strconv.Atoi(ifIdxStr)
				if err == nil {
					ifIdx := new(uint32)
					*ifIdx = uint32(ifIdxNum)
					ifInfo.State.Ifindex = ifIdx
				}
			case ACTIVE:
			case NAME:
			default:
				log.Info("Not a valid attribute!")
			}
		}
	}
}

func (app *IntfApp) convertInternalToOCIntfVlanListInfo(ifName *string, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface) error {
	var err error
	if ifInfo.Ethernet == nil || ifInfo.Ethernet.SwitchedVlan == nil || ifInfo.Ethernet.SwitchedVlan.State == nil {
		return nil
	}
	taggedMemberPresent := false

	if len(*ifName) < 0 {
		return nil
	}
	vlanMap, ok := app.vlanD.vlanMembersTableMap[*ifName]
	if ok {
		for vlanName, tagEntry := range vlanMap {
			vlanIdStr := vlanName[len("Vlan"):len(vlanName)]
			vlanId, err := strconv.Atoi(vlanIdStr)
			vlanIdCast := uint16(vlanId)

			tagMode := tagEntry.entry.Field["tagging_mode"]
			if tagMode == "untagged" {
				if err != nil {
					errStr := "Translation of Vlan-name: " + vlanName + " to vlan-id failed!"
					return errors.New(errStr)
				}
				ifInfo.Ethernet.SwitchedVlan.State.AccessVlan = &(vlanIdCast)
				log.Infof("Adding access-vlan: %d succesful!", vlanIdCast)
			} else {
				taggedMemberPresent = true
				trunkVlan, _ := ifInfo.Ethernet.SwitchedVlan.State.To_OpenconfigInterfaces_Interfaces_Interface_Ethernet_SwitchedVlan_State_TrunkVlans_Union(vlanIdCast)
				ifInfo.Ethernet.SwitchedVlan.State.TrunkVlans = append(ifInfo.Ethernet.SwitchedVlan.State.TrunkVlans, trunkVlan)
				log.Infof("Adding trunk-vlan: %d succesful!", trunkVlan)
			}

		}
		if taggedMemberPresent {
			ifInfo.Ethernet.SwitchedVlan.State.InterfaceMode = ocbinds.OpenconfigVlan_VlanModeType_TRUNK
		} else {
			ifInfo.Ethernet.SwitchedVlan.State.InterfaceMode = ocbinds.OpenconfigVlan_VlanModeType_ACCESS
		}
	}
	return err
}

func (app *IntfApp) convertInternalToOCIntfIPAttrInfo(ifName *string, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface) {

	/* Handling the Interface IP attributes */
	subIntf, err := ifInfo.Subinterfaces.NewSubinterface(0)
	if err != nil {
		log.Error("Creation of subinterface subtree failed!")
		return
	}
	if subIntf == nil {
		return
	}
	ygot.BuildEmptyTree(subIntf)
	if ipMap, ok := app.ifIPTableMap[*ifName]; ok {
		for ipKey, _ := range ipMap {
			log.Info("IP address = ", ipKey)
			ipB, ipNetB, _ := net.ParseCIDR(ipKey)

			v4Flag := false
			v6Flag := false

			var v4Address *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address
			var v6Address *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address

			if validIPv4(ipB.String()) {
				v4Address, err = subIntf.Ipv4.Addresses.NewAddress(ipB.String())
				v4Flag = true
			} else if validIPv6(ipB.String()) {
				v6Address, err = subIntf.Ipv6.Addresses.NewAddress(ipB.String())
				v6Flag = true
			} else {
				log.Error("Invalid IP address " + ipB.String())
				continue
			}

			if err != nil {
				log.Error("Creation of address subtree failed!")
				return
			}
			if v4Flag {
				ygot.BuildEmptyTree(v4Address)

				ipStr := new(string)
				*ipStr = ipB.String()
				v4Address.Ip = ipStr
				v4Address.Config.Ip = ipStr
				v4Address.State.Ip = ipStr

				ipNetBNum, _ := ipNetB.Mask.Size()
				prfxLen := new(uint8)
				*prfxLen = uint8(ipNetBNum)
				v4Address.Config.PrefixLength = prfxLen
				v4Address.State.PrefixLength = prfxLen
			}
			if v6Flag {
				ygot.BuildEmptyTree(v6Address)

				ipStr := new(string)
				*ipStr = ipB.String()
				v6Address.Ip = ipStr
				v6Address.Config.Ip = ipStr
				v6Address.State.Ip = ipStr

				ipNetBNum, _ := ipNetB.Mask.Size()
				prfxLen := new(uint8)
				*prfxLen = uint8(ipNetBNum)
				v6Address.Config.PrefixLength = prfxLen
				v6Address.State.PrefixLength = prfxLen
			}
		}
	}
}

func (app *IntfApp) convertInternalToOCPortStatInfo(ifName *string, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface) {
	if len(app.intfD.portStatMap) == 0 {
		log.Info("Port stat info not present for interface :", *ifName)
		return
	}
	if ifInfo.State == nil || ifInfo.State.Counters == nil {
		return
	}
	if portStatInfo, ok := app.intfD.portStatMap[*ifName]; ok {
		log.Info("Entered Counters filling")

		inOctet := new(uint64)
		inOctetVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_OCTETS"])
		*inOctet = uint64(inOctetVal)
		ifInfo.State.Counters.InOctets = inOctet

		inUCastPkt := new(uint64)
		inUCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_UCAST_PKTS"])
		*inUCastPkt = uint64(inUCastPktVal)
		ifInfo.State.Counters.InUnicastPkts = inUCastPkt

		inNonUCastPkt := new(uint64)
		inNonUCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_NON_UCAST_PKTS"])
		*inNonUCastPkt = uint64(inNonUCastPktVal)

		inPkt := new(uint64)
		*inPkt = *inUCastPkt + *inNonUCastPkt
		ifInfo.State.Counters.InPkts = inPkt

		inBCastPkt := new(uint64)
		inBCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_BROADCAST_PKTS"])
		*inBCastPkt = uint64(inBCastPktVal)
		ifInfo.State.Counters.InBroadcastPkts = inBCastPkt

		inMCastPkt := new(uint64)
		inMCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_MULTICAST_PKTS"])
		*inMCastPkt = uint64(inMCastPktVal)
		ifInfo.State.Counters.InMulticastPkts = inMCastPkt

		inErrPkt := new(uint64)
		inErrPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_ERRORS"])
		*inErrPkt = uint64(inErrPktVal)
		ifInfo.State.Counters.InErrors = inErrPkt

		inDiscPkt := new(uint64)
		inDiscPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_DISCARDS"])
		*inDiscPkt = uint64(inDiscPktVal)
		ifInfo.State.Counters.InDiscards = inDiscPkt

		outOctet := new(uint64)
		outOctetVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_OCTETS"])
		*outOctet = uint64(outOctetVal)
		ifInfo.State.Counters.OutOctets = outOctet

		outUCastPkt := new(uint64)
		outUCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_UCAST_PKTS"])
		*outUCastPkt = uint64(outUCastPktVal)
		ifInfo.State.Counters.OutUnicastPkts = outUCastPkt

		outNonUCastPkt := new(uint64)
		outNonUCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_NON_UCAST_PKTS"])
		*outNonUCastPkt = uint64(outNonUCastPktVal)

		outPkt := new(uint64)
		*outPkt = *outUCastPkt + *outNonUCastPkt
		ifInfo.State.Counters.OutPkts = outPkt

		outBCastPkt := new(uint64)
		outBCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_BROADCAST_PKTS"])
		*outBCastPkt = uint64(outBCastPktVal)
		ifInfo.State.Counters.OutBroadcastPkts = outBCastPkt

		outMCastPkt := new(uint64)
		outMCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_MULTICAST_PKTS"])
		*outMCastPkt = uint64(outMCastPktVal)
		ifInfo.State.Counters.OutMulticastPkts = outMCastPkt

		outErrPkt := new(uint64)
		outErrPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_ERRORS"])
		*outErrPkt = uint64(outErrPktVal)
		ifInfo.State.Counters.OutErrors = outErrPkt

		outDiscPkt := new(uint64)
		outDiscPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_DISCARDS"])
		*outDiscPkt = uint64(outDiscPktVal)
		ifInfo.State.Counters.OutDiscards = outDiscPkt
	}
}

func (app *IntfApp) convertInternalToOCIntfInfo(ifName *string, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface) {
	app.convertInternalToOCIntfAttrInfo(ifName, ifInfo)
	app.convertInternalToOCIntfVlanListInfo(ifName, ifInfo)
	app.convertInternalToOCIntfIPAttrInfo(ifName, ifInfo)
	app.convertInternalToOCPortStatInfo(ifName, ifInfo)
}

/* Build tree for sending the response back to North bound */
func (app *IntfApp) processBuildTree(ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface, ifKey *string) {
	ygot.BuildEmptyTree(ifInfo)
	if *app.ygotTarget == ifInfo.State {
		ygot.BuildEmptyTree(ifInfo.State)
	}
	app.convertInternalToOCIntfInfo(ifKey, ifInfo)
}

func (app *IntfApp) processGetSpecificIntf(dbs [db.MaxDB]*db.DB, targetUriPath *string) (GetResponse, error) {
	var err error
	var payload []byte
	var ok bool
	var resp *GetResponse

	pathInfo := app.path
	intfObj := app.getAppRootObject()

	log.Infof("Received GET for path %s; template: %s vars=%v", pathInfo.Path, pathInfo.Template, pathInfo.Vars)

	if intfObj.Interface != nil && len(intfObj.Interface) > 0 {
		/* Interface name is the key */
		for ifKey, _ := range intfObj.Interface {
			log.Info("Interface Name = ", ifKey)
			err = app.getIntfTypeFromIntf(&ifKey)
			if err != nil {
				return GetResponse{Payload: payload, ErrSrc: AppErr}, err
			}
			switch app.intfType {
			case ETHERNET:
				/* First, convert the data to the DS and check for the request type */
				err = app.processGetConvertDBPhyIfInfoToDS(&ifKey)
				if err != nil {
					return GetResponse{Payload: payload, ErrSrc: AppErr}, err
				}

				ok, resp, err = app.processGetSpecificAttr(targetUriPath, &ifKey)
				if ok {
					return *resp, err
				}

			case VLAN:
				err = app.processGetConvertDBVlanIfInfoToDS(&ifKey)
				if err != nil {
					return GetResponse{Payload: payload, ErrSrc: AppErr}, err
				}
				ok, resp, err := app.processGetSpecificAttr(targetUriPath, &ifKey)
				if ok {
					return *resp, err
				}
			case LAG:
				err = app.processGetConvertDBLagIfInfoToDS(&ifKey)
				if err != nil {
					return GetResponse{Payload: payload, ErrSrc: AppErr}, err
				}
				ok, resp, err := app.processGetSpecificAttr(targetUriPath, &ifKey)
				if ok {
					return *resp, err
				}
			}

			ifInfo := intfObj.Interface[ifKey]
			/*  Attribute level handling is done at the top and returned. Any container level handling
			    needs to be updated here. This is done, just to avoid un-necessary building of tree for any
			    incoming request */
			/* TODO: Need to handle these conditions in a cleaner way */
			if *app.ygotTarget != ifInfo && *app.ygotTarget != ifInfo.Config && *app.ygotTarget != ifInfo.State &&
				*app.ygotTarget != ifInfo.State.Counters {
				return GetResponse{Payload: payload}, errors.New("Requested get type not supported!")
			}
			app.processBuildTree(ifInfo, &ifKey)

			if *app.ygotTarget == ifInfo {
				payload, err = dumpIetfJson(intfObj, false)
			} else {
				dummyifInfo := &ocbinds.OpenconfigInterfaces_Interfaces_Interface{}
				if *app.ygotTarget == ifInfo.Config {
					dummyifInfo.Config = ifInfo.Config
					payload, err = dumpIetfJson(dummyifInfo, false)
				} else if *app.ygotTarget == ifInfo.State {
					dummyifInfo.State = ifInfo.State
					payload, err = dumpIetfJson(dummyifInfo, false)
				} else if *app.ygotTarget == ifInfo.State.Counters {
					dummyifStateInfo := &ocbinds.OpenconfigInterfaces_Interfaces_Interface_State{}
					dummyifStateInfo.Counters = ifInfo.State.Counters
					payload, err = dumpIetfJson(dummyifStateInfo, false)
				} else {
					log.Info("Not supported get type!")
					err = errors.New("Requested get-type not supported!")
				}
			}
			resp = &(GetResponse{Payload: payload})
		}
	}
	return *resp, err
}

func (app *IntfApp) processGetAllInterfaces(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
	var err error
	var payload []byte
	var resp *GetResponse

	ifName := ""
	intfObj := app.getAppRootObject()

	log.Info("Get all Interfaces request!")

	err = app.processGetConvertDBIfInfoToDS(&ifName)
	if err != nil {
		return GetResponse{Payload: payload, ErrSrc: AppErr}, err
	}

	ygot.BuildEmptyTree(intfObj)
	for ifName, _ := range app.ifTableMap {
		ifInfo, err := intfObj.NewInterface(ifName)
		if err != nil {
			log.Errorf("Creation of interface subtree for %s failed!", ifName)
			return GetResponse{Payload: payload, ErrSrc: AppErr}, err
		}
		app.processBuildTree(ifInfo, &ifName)
	}
	if *app.ygotTarget == intfObj {
		payload, err = dumpIetfJson((*app.ygotRoot).(*ocbinds.Device), true)
		resp = &(GetResponse{Payload: payload})
	} else {
		log.Error("Wrong Request!")
	}
	return *resp, err
}
