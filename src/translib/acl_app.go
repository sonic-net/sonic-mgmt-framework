////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package translib

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"

	log "github.com/golang/glog"
	"github.com/openconfig/ygot/util"
	"github.com/openconfig/ygot/ygot"
)

const (
	TABLE_SEPARATOR          = "|"
	KEY_SEPARATOR            = "|"
	ACL_TABLE                = "ACL_TABLE"
	RULE_TABLE               = "ACL_RULE"
	ACL_TYPE                 = "type"
	ACL_DESCRIPTION          = "policy_desc"
	SONIC_ACL_TYPE_L2        = "L2"
	SONIC_ACL_TYPE_IPV4      = "L3"
	SONIC_ACL_TYPE_IPV6      = "L3V6"
	OPENCONFIG_ACL_TYPE_IPV4 = "ACL_IPV4"
	OPENCONFIG_ACL_TYPE_IPV6 = "ACL_IPV6"
	OPENCONFIG_ACL_TYPE_L2   = "ACL_L2"
	OC_ACL_APP_MODULE_NAME   = "/openconfig-acl:acl"
	OC_ACL_YANG_PATH_PREFIX  = "/device/acl"

	MIN_PRIORITY = 1
	MAX_PRIORITY = 65536
)

var IP_PROTOCOL_MAP = map[ocbinds.E_OpenconfigPacketMatchTypes_IP_PROTOCOL]uint8{
	ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_ICMP: 1,
	ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_IGMP: 2,
	ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_TCP:  6,
	ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_UDP:  17,
	ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_RSVP: 46,
	ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_GRE:  47,
	ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_AUTH: 51,
	ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_PIM:  103,
	ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_L2TP: 115,
}

var ETHERTYPE_MAP = map[ocbinds.E_OpenconfigPacketMatchTypes_ETHERTYPE]uint32{
	ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_LLDP: 0x88CC,
	ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_VLAN: 0x8100,
	ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_ROCE: 0x8915,
	ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_ARP:  0x0806,
	ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_IPV4: 0x0800,
	ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_IPV6: 0x86DD,
	ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_MPLS: 0x8847,
}

type AclApp struct {
	pathInfo   *PathInfo
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}

	aclTs  *db.TableSpec
	ruleTs *db.TableSpec

	aclTableMap  map[string]db.Value
	ruleTableMap map[string]map[string]db.Value
}

func init() {

	err := register("/openconfig-acl:acl",
		&appInfo{appType: reflect.TypeOf(AclApp{}),
			ygotRootType:  reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
			isNative:      false,
			tablesToWatch: []*db.TableSpec{&db.TableSpec{Name: ACL_TABLE}, &db.TableSpec{Name: RULE_TABLE}}})

	if err != nil {
		log.Fatal("Register ACL app module with App Interface failed with error=", err)
	}

	err = addModel(&ModelData{Name: "openconfig-acl",
		Org: "OpenConfig working group",
		Ver: "1.0.2"})
	if err != nil {
		log.Fatal("Adding model data to appinterface failed with error=", err)
	}
}

func (app *AclApp) initialize(data appData) {
	log.Info("initialize:acl:path =", data.path)
	pathInfo := NewPathInfo(data.path)
	*app = AclApp{pathInfo: pathInfo, ygotRoot: data.ygotRoot, ygotTarget: data.ygotTarget}

	app.aclTs = &db.TableSpec{Name: ACL_TABLE}
	app.ruleTs = &db.TableSpec{Name: RULE_TABLE}

	app.aclTableMap = make(map[string]db.Value)
	app.ruleTableMap = make(map[string]map[string]db.Value)
}

func (app *AclApp) getAppRootObject() *ocbinds.OpenconfigAcl_Acl {
	deviceObj := (*app.ygotRoot).(*ocbinds.Device)
	return deviceObj.Acl
}

func (app *AclApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCreate:acl:path =", app.pathInfo.Template)

	keys, err = app.translateCRUCommon(d, CREATE)
	return keys, err
}

func (app *AclApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateUpdate:acl:path =", app.pathInfo.Template)

	keys, err = app.translateCRUCommon(d, UPDATE)
	return keys, err
}

func (app *AclApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateReplace:acl:path =", app.pathInfo.Template)

	keys, err = app.translateCRUCommon(d, REPLACE)
	return keys, err
}

func (app *AclApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateDelete:acl:path =", app.pathInfo.Template)

	return keys, err
}

func (app *AclApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	var err error
	log.Info("translateGet:acl:path =", app.pathInfo.Template)
	return err
}

func (app *AclApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
	pathInfo := NewPathInfo(path)
	notifInfo := notificationInfo{dbno: db.ConfigDB}
	notSupported := tlerr.NotSupportedError{
		Format: "Subscribe not supported", Path: path}

	if isSubtreeRequest(pathInfo.Template, "/openconfig-acl:acl/acl-sets") {
		// Subscribing to top level ACL record is not supported. It requires listening
		// to 2 tables (ACL and ACL_RULE); TransLib does not support it yet
		if pathInfo.HasSuffix("/acl-sets") ||
			pathInfo.HasSuffix("/acl-set") ||
			pathInfo.HasSuffix("/acl-set{}{}") {
			log.Errorf("Subscribe not supported for top level ACL %s", pathInfo.Template)
			return nil, nil, notSupported
		}

		t, err := getAclTypeOCEnumFromName(pathInfo.Var("type"))
		if err != nil {
			return nil, nil, err
		}

		aclkey := getAclKeyStrFromOCKey(pathInfo.Var("name"), t)

		if strings.Contains(pathInfo.Template, "/acl-entry{}") {
			// Subscribe for one rule
			rulekey := "RULE_" + pathInfo.Var("sequence-id")
			notifInfo.table = db.TableSpec{Name: RULE_TABLE}
			notifInfo.key = asKey(aclkey, rulekey)
			notifInfo.needCache = !pathInfo.HasSuffix("/acl-entry{}")

		} else if pathInfo.HasSuffix("/acl-entries") || pathInfo.HasSuffix("/acl-entry") {
			// Subscribe for all rules of an ACL
			notifInfo.table = db.TableSpec{Name: RULE_TABLE}
			notifInfo.key = asKey(aclkey, "*")

		} else {
			// Subscibe for ACL fields only
			notifInfo.table = db.TableSpec{Name: ACL_TABLE}
			notifInfo.key = asKey(aclkey)
			notifInfo.needCache = true
		}

	} else if isSubtreeRequest(pathInfo.Template, "/openconfig-acl:acl/interfaces") {
		// Right now interface binding config is maintained within ACL
		// table itself. Multiple ACLs can be bound to one intf; one
		// inname can occur in multiple ACL entries. So we cannot map
		// interface binding xpaths to specific ACL table entry keys.
		// For now subscribe for full ACL table!!
		notifInfo.table = db.TableSpec{Name: ACL_TABLE}
		notifInfo.key = asKey("*")
		notifInfo.needCache = true

	} else {
		log.Errorf("Unknown path %s", pathInfo.Template)
		return nil, nil, notSupported
	}

	return nil, &notifInfo, nil
}

func (app *AclApp) processCreate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	if err = app.processCommon(d, CREATE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}
	return resp, err
}

func (app *AclApp) processUpdate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	if err = app.processCommon(d, UPDATE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}
	return resp, err
}

func (app *AclApp) processReplace(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	if err = app.processCommon(d, REPLACE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}
	return resp, err
}

func (app *AclApp) processDelete(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	if err = app.processCommon(d, DELETE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}
	return resp, err
}

func (app *AclApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
	var err error
	var payload []byte

	configDb := dbs[db.ConfigDB]
	err = app.processCommon(configDb, GET)
	if err != nil {
		return GetResponse{Payload: payload, ErrSrc: AppErr}, err
	}

	payload, err = generateGetResponsePayload(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device), app.ygotTarget)
	if err != nil {
		return GetResponse{Payload: payload, ErrSrc: AppErr}, err
	}

	return GetResponse{Payload: payload}, err
}

func (app *AclApp) translateCRUCommon(d *db.DB, opcode int) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCRUCommon:acl:path =", app.pathInfo.Template)

	app.convertOCAclsToInternal()
	app.convertOCAclRulesToInternal(d)
	app.convertOCAclBindingsToInternal()

	return keys, err
}

func (app *AclApp) processCommon(d *db.DB, opcode int) error {
	var err error
	var topmostPath bool = false
	acl := app.getAppRootObject()

	log.Infof("processCommon--Path Received: %s", app.pathInfo.Template)
	targetType := reflect.TypeOf(*app.ygotTarget)
	if !util.IsValueScalar(reflect.ValueOf(*app.ygotTarget)) && util.IsValuePtr(reflect.ValueOf(*app.ygotTarget)) {
		log.Infof("processCommon: Target object is a <%s> of Type: %s", targetType.Kind().String(), targetType.Elem().Name())
		if targetType.Elem().Name() == "OpenconfigAcl_Acl" {
			topmostPath = true
		}
	}

	targetUriPath, _ := getYangPathFromUri(app.pathInfo.Path)
	if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/acl-sets") {
		if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/acl-sets/acl-set{}{}") {
			for aclSetKey, _ := range acl.AclSets.AclSet {
				aclSet := acl.AclSets.AclSet[aclSetKey]
				aclKey := getAclKeyStrFromOCKey(aclSetKey.Name, aclSetKey.Type)

				if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/acl-sets/acl-set{}{}/acl-entries/acl-entry{}") {
					// Subtree of one Rule
					for seqId, _ := range aclSet.AclEntries.AclEntry {
						ruleKey := "RULE_" + strconv.Itoa(int(seqId))
						entrySet := aclSet.AclEntries.AclEntry[seqId]

						ruleNodeYangPath := getYangPathFromYgotStruct(entrySet, OC_ACL_YANG_PATH_PREFIX, OC_ACL_APP_MODULE_NAME)
						isRuleNodeSubtree := len(targetUriPath) > len(ruleNodeYangPath)
						switch opcode {
						case CREATE:
							if isRuleNodeSubtree {
								err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, false)
							} else if *app.ygotTarget == entrySet {
								err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, true)
							} else {
								log.Errorf("processCommon: Given CREATE path %s not handled", targetUriPath)
							}
						case REPLACE:
							err = d.SetEntry(app.ruleTs, db.Key{Comp: []string{aclKey, ruleKey}}, app.ruleTableMap[aclKey][ruleKey])
						case UPDATE:
							err = d.ModEntry(app.ruleTs, db.Key{Comp: []string{aclKey, ruleKey}}, app.ruleTableMap[aclKey][ruleKey])
						case DELETE:
							if *app.ygotTarget == entrySet {
								err = d.DeleteEntry(app.ruleTs, db.Key{Comp: []string{aclKey, ruleKey}})
							} else if isRuleNodeSubtree {
								err = app.handleRuleFieldsDeletion(d, aclKey, ruleKey)
								if err != nil {
									return err
								}
								//err = d.SetEntry(app.ruleTs, db.Key{Comp: []string{aclKey, ruleKey}}, app.ruleTableMap[aclKey][ruleKey])
							} else {
								log.Errorf("processCommon: Given DELETE path %s not handled", targetUriPath)
							}
						case GET:
							err = app.convertDBAclRulesToInternal(d, aclKey, int64(seqId), db.Key{})
							ygot.BuildEmptyTree(entrySet)
							app.convertInternalToOCAclRule(aclKey, aclSetKey.Type, int64(seqId), nil, entrySet)
						}
					}
				} else {
					isAclEntriesSubtree := isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/acl-sets/acl-set{}{}/acl-entries")
					switch opcode {
					case CREATE:
						if *app.ygotTarget == aclSet {
							err = app.setAclDataInConfigDb(d, app.aclTableMap, true)
							if err != nil {
								return err
							}
							err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, true)
						} else if isAclEntriesSubtree {
							err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, true)
						} else {
							err = d.SetEntry(app.aclTs, db.Key{Comp: []string{aclKey}}, app.aclTableMap[aclKey])
						}
					case REPLACE:
						if *app.ygotTarget == aclSet || isAclEntriesSubtree {
							err = d.DeleteKeys(app.ruleTs, db.Key{Comp: []string{aclKey + TABLE_SEPARATOR + "RULE_*"}})
							if err != nil {
								return err
							}
							err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, true)
							if err != nil {
								return err
							}
						}
						if !isAclEntriesSubtree {
							err = d.ModEntry(app.aclTs, db.Key{Comp: []string{aclKey}}, app.aclTableMap[aclKey])
						}
					case UPDATE:
						if !isAclEntriesSubtree {
							err = app.setAclDataInConfigDb(d, app.aclTableMap, false)
							//err = d.ModEntry(app.aclTs, db.Key{Comp: []string{aclKey}}, app.aclTableMap[aclKey])
							if err != nil {
								return err
							}
						}
						if *app.ygotTarget == aclSet || isAclEntriesSubtree {
							err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, false)
						}
					case DELETE:
						if *app.ygotTarget == aclSet {
							err = d.DeleteKeys(app.ruleTs, db.Key{Comp: []string{aclKey + TABLE_SEPARATOR + "*"}})
							if err != nil {
								return err
							}
							err = d.DeleteEntry(app.aclTs, db.Key{Comp: []string{aclKey}})
						} else if isAclEntriesSubtree {
							err = d.DeleteKeys(app.ruleTs, db.Key{Comp: []string{aclKey + TABLE_SEPARATOR + "RULE_*"}})
						} else {
							nodeInfo, err := getTargetNodeYangSchema(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device))
							if err != nil {
								return err
							}
							if nodeInfo != nil && nodeInfo.IsLeaf() && nodeInfo.Name == "description" {
								err = d.DeleteEntryFields(app.aclTs, asKey(aclKey), createEmptyDbValue(ACL_DESCRIPTION))
							}
							//err = d.SetEntry(app.aclTs, db.Key{Comp: []string{aclKey}}, app.aclTableMap[aclKey])
						}
					case GET:
						err = app.convertDBAclToInternal(d, db.Key{Comp: []string{aclKey}})
						if err != nil {
							return err
						}
						ygot.BuildEmptyTree(aclSet)
						app.convertInternalToOCAcl(aclKey, acl.AclSets, aclSet)
					}
				}
			}
		} else {
			// All Acls and their rules
			err = app.processCommonToplevelPath(d, acl, opcode, false)
		}
	} else if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/interfaces") {
		switch opcode {
		case CREATE, REPLACE, UPDATE:
			err = app.setAclBindDataInConfigDb(d, app.aclTableMap, opcode)
		case DELETE:
			err = app.handleBindingsDeletion(d)
		case GET:
			if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/interfaces/interface{}") {
				for intfId := range acl.Interfaces.Interface {
					intfData := acl.Interfaces.Interface[intfId]
					ygot.BuildEmptyTree(intfData)
					if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces/interface/ingress-acl-sets") {
						err = app.getAclBindingInfoForInterfaceData(d, intfData, intfId, "INGRESS")
					} else if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces/interface/egress-acl-sets") {
						err = app.getAclBindingInfoForInterfaceData(d, intfData, intfId, "EGRESS")
					} else {
						// Direction unknown. Check ACL Table for binding information.
						err = app.getAclBindingInfoForInterfaceData(d, intfData, intfId, "INGRESS")
						if err != nil {
							return err
						}
						err = app.getAclBindingInfoForInterfaceData(d, intfData, intfId, "EGRESS")
					}
				}
			} else {
				err = app.getAllBindingsInfo(d)
			}
		}
	} else {
		err = app.processCommonToplevelPath(d, acl, opcode, true)
	}

	if !topmostPath && !isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets") && !isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces") {
		err = tlerr.NotSupported("URL %s is not supported", app.pathInfo.Template)
	}

	return err
}

func (app *AclApp) processCommonToplevelPath(d *db.DB, acl *ocbinds.OpenconfigAcl_Acl, opcode int, isTopmostPath bool) error {
	var err error
	switch opcode {
	case CREATE:
		err = app.setAclDataInConfigDb(d, app.aclTableMap, true)
		if err != nil {
			return err
		}
		err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, true)
	case REPLACE:
		err = d.DeleteTable(app.aclTs)
		if err != nil {
			return err
		}
		err = d.DeleteTable(app.ruleTs)
		if err != nil {
			return err
		}
		err = app.setAclDataInConfigDb(d, app.aclTableMap, true)
		if err != nil {
			return err
		}
		err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, true)
	case UPDATE:
		err = app.setAclDataInConfigDb(d, app.aclTableMap, false)
		if err != nil {
			return err
		}
		err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, false)
	case DELETE:
		err = d.DeleteTable(app.ruleTs)
		if err != nil {
			return err
		}
		err = d.DeleteTable(app.aclTs)
	case GET:
		ygot.BuildEmptyTree(acl)
		err = app.convertDBAclToInternal(d, db.Key{})
		if err != nil {
			return err
		}
		app.convertInternalToOCAcl("", acl.AclSets, nil)
		if isTopmostPath {
			err = app.getAllBindingsInfo(d)
		}
	}
	return err
}

/***********    These are Translation Helper Function   ***********/
func (app *AclApp) convertDBAclRulesToInternal(dbCl *db.DB, aclName string, seqId int64, ruleKey db.Key) error {
	var err error
	if seqId != -1 {
		ruleKey.Comp = []string{aclName, "RULE_" + strconv.FormatInt(int64(seqId), 10)}
	}
	if ruleKey.Len() > 1 {
		ruleName := ruleKey.Get(1)
		if ruleName != "DEFAULT_RULE" {
			ruleData, err := dbCl.GetEntry(app.ruleTs, ruleKey)
			if err != nil {
				return err
			}
			if app.ruleTableMap[aclName] == nil {
				app.ruleTableMap[aclName] = make(map[string]db.Value)
			}
			app.ruleTableMap[aclName][ruleName] = ruleData
		}
	} else {
		ruleKeys, err := dbCl.GetKeys(app.ruleTs)
		if err != nil {
			return err
		}
		for i, _ := range ruleKeys {
			if aclName == ruleKeys[i].Get(0) {
				app.convertDBAclRulesToInternal(dbCl, aclName, -1, ruleKeys[i])
			}
		}
	}
	return err
}

func (app *AclApp) convertDBAclToInternal(dbCl *db.DB, aclkey db.Key) error {
	var err error
	if aclkey.Len() > 0 {
		// Get one particular ACL
		entry, err := dbCl.GetEntry(app.aclTs, aclkey)
		if err != nil {
			return err
		}
		if entry.IsPopulated() {
			app.aclTableMap[aclkey.Get(0)] = entry
			app.ruleTableMap[aclkey.Get(0)] = make(map[string]db.Value)
			err = app.convertDBAclRulesToInternal(dbCl, aclkey.Get(0), -1, db.Key{})
			if err != nil {
				return err
			}
		} else {
			return tlerr.NotFound("Acl %s is not configured", aclkey.Get(0))
		}
	} else {
		// Get all ACLs
		tbl, err := dbCl.GetTable(app.aclTs)
		if err != nil {
			return err
		}
		keys, _ := tbl.GetKeys()
		for i, _ := range keys {
			app.convertDBAclToInternal(dbCl, keys[i])
		}
	}
	return err
}

func (app *AclApp) convertInternalToOCAcl(aclName string, aclSets *ocbinds.OpenconfigAcl_Acl_AclSets, aclSet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet) {
	if len(aclName) > 0 {
		aclData := app.aclTableMap[aclName]
		if aclSet != nil {
			aclSet.Config.Name = aclSet.Name
			aclSet.Config.Type = aclSet.Type
			aclSet.State.Name = aclSet.Name
			aclSet.State.Type = aclSet.Type

			for k := range aclData.Field {
				if ACL_DESCRIPTION == k {
					descr := aclData.Get(k)
					aclSet.Config.Description = &descr
					aclSet.State.Description = &descr
				} else if "ports@" == k {
					continue
				}
			}

			app.convertInternalToOCAclRule(aclName, aclSet.Type, -1, aclSet, nil)
		}
	} else {
		for acln := range app.aclTableMap {
			acldata := app.aclTableMap[acln]
			var aclNameStr string
			var aclType ocbinds.E_OpenconfigAcl_ACL_TYPE
			if acldata.Get(ACL_TYPE) == SONIC_ACL_TYPE_IPV4 {
				aclNameStr = strings.Replace(acln, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
				aclType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4
			} else if acldata.Get(ACL_TYPE) == SONIC_ACL_TYPE_IPV6 {
				aclNameStr = strings.Replace(acln, "_"+OPENCONFIG_ACL_TYPE_IPV6, "", 1)
				aclType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6
			} else if acldata.Get(ACL_TYPE) == SONIC_ACL_TYPE_L2 {
				aclNameStr = strings.Replace(acln, "_"+OPENCONFIG_ACL_TYPE_L2, "", 1)
				aclType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2
			}
			aclSetPtr, aclErr := aclSets.NewAclSet(aclNameStr, aclType)
			if aclErr != nil {
				fmt.Println("Error handling: ", aclErr)
			}
			ygot.BuildEmptyTree(aclSetPtr)
			app.convertInternalToOCAcl(acln, nil, aclSetPtr)
		}
	}
}

func (app *AclApp) convertInternalToOCAclRule(aclName string, aclType ocbinds.E_OpenconfigAcl_ACL_TYPE, seqId int64, aclSet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet, entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
	if seqId != -1 {
		ruleName := "RULE_" + strconv.FormatInt(int64(seqId), 10)
		app.convertInternalToOCAclRuleProperties(app.ruleTableMap[aclName][ruleName], aclType, nil, entrySet)
	} else {
		for ruleName := range app.ruleTableMap[aclName] {
			app.convertInternalToOCAclRuleProperties(app.ruleTableMap[aclName][ruleName], aclType, aclSet, nil)
		}
	}
}

func (app *AclApp) convertInternalToOCAclRuleProperties(ruleData db.Value, aclType ocbinds.E_OpenconfigAcl_ACL_TYPE, aclSet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet, entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
	priority, _ := strconv.ParseInt(ruleData.Get("PRIORITY"), 10, 32)
	seqId := uint32(MAX_PRIORITY - priority)
	//ruleDescr := ruleData.Get("RULE_DESCRIPTION")

	if entrySet == nil {
		if aclSet != nil {
			entrySet_, _ := aclSet.AclEntries.NewAclEntry(seqId)
			entrySet = entrySet_
			ygot.BuildEmptyTree(entrySet)
		}
	}

	entrySet.Config.SequenceId = &seqId
	//entrySet.Config.Description = &ruleDescr
	entrySet.State.SequenceId = &seqId
	//entrySet.State.Description = &ruleDescr

	var num uint64
	num = 0
	entrySet.State.MatchedOctets = &num
	entrySet.State.MatchedPackets = &num

	ygot.BuildEmptyTree(entrySet.Transport)
	ygot.BuildEmptyTree(entrySet.Actions)

	for ruleKey := range ruleData.Field {
		if "L4_SRC_PORT" == ruleKey || "L4_SRC_PORT_RANGE" == ruleKey {
			port := ruleData.Get(ruleKey)
			srcPort := getTransportSrcDestPorts(port, "src")
			entrySet.Transport.Config.SourcePort, _ = entrySet.Transport.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union(srcPort)
			entrySet.Transport.State.SourcePort, _ = entrySet.Transport.State.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_State_SourcePort_Union(srcPort)
		} else if "L4_DST_PORT" == ruleKey || "L4_DST_PORT_RANGE" == ruleKey {
			port := ruleData.Get(ruleKey)
			destPort := getTransportSrcDestPorts(port, "dest")
			entrySet.Transport.Config.DestinationPort, _ = entrySet.Transport.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union(destPort)
			entrySet.Transport.State.DestinationPort, _ = entrySet.Transport.State.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_State_DestinationPort_Union(destPort)
		} else if "TCP_FLAGS" == ruleKey {
			tcpFlags := ruleData.Get(ruleKey)
			entrySet.Transport.Config.TcpFlags = getTransportConfigTcpFlags(tcpFlags)
			entrySet.Transport.State.TcpFlags = getTransportConfigTcpFlags(tcpFlags)
		} else if "PACKET_ACTION" == ruleKey {
			if "FORWARD" == ruleData.Get(ruleKey) {
				entrySet.Actions.Config.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_ACCEPT
				entrySet.Actions.State.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_ACCEPT
			} else {
				entrySet.Actions.Config.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_DROP
				entrySet.Actions.State.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_DROP
			}
		}
	}

	if aclType == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 {
		ygot.BuildEmptyTree(entrySet.Ipv4)
		for ruleKey := range ruleData.Field {
			if "IP_PROTOCOL" == ruleKey {
				ipProto, _ := strconv.ParseInt(ruleData.Get(ruleKey), 10, 64)
				protocolVal := getIpProtocol(ipProto)
				entrySet.Ipv4.Config.Protocol, _ = entrySet.Ipv4.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union(protocolVal)
				entrySet.Ipv4.State.Protocol, _ = entrySet.Ipv4.State.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_State_Protocol_Union(protocolVal)
			} else if "DSCP" == ruleKey {
				var dscp uint8
				dscpData, _ := strconv.ParseInt(ruleData.Get(ruleKey), 10, 64)
				dscp = uint8(dscpData)
				entrySet.Ipv4.Config.Dscp = &dscp
				entrySet.Ipv4.State.Dscp = &dscp
			} else if "SRC_IP" == ruleKey {
				addr := ruleData.Get(ruleKey)
				entrySet.Ipv4.Config.SourceAddress = &addr
				entrySet.Ipv4.State.SourceAddress = &addr
			} else if "DST_IP" == ruleKey {
				addr := ruleData.Get(ruleKey)
				entrySet.Ipv4.Config.DestinationAddress = &addr
				entrySet.Ipv4.State.DestinationAddress = &addr
			}
		}
	} else if aclType == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6 {
		ygot.BuildEmptyTree(entrySet.Ipv6)
		for ruleKey := range ruleData.Field {
			if "IP_PROTOCOL" == ruleKey {
				ipProto, _ := strconv.ParseInt(ruleData.Get(ruleKey), 10, 64)
				protocolVal := getIpProtocol(ipProto)
				entrySet.Ipv6.Config.Protocol, _ = entrySet.Ipv6.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union(protocolVal)
				entrySet.Ipv6.State.Protocol, _ = entrySet.Ipv6.State.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_State_Protocol_Union(protocolVal)
			} else if "DSCP" == ruleKey {
				var dscp uint8
				dscpData, _ := strconv.ParseInt(ruleData.Get(ruleKey), 10, 64)
				dscp = uint8(dscpData)
				entrySet.Ipv6.Config.Dscp = &dscp
				entrySet.Ipv6.State.Dscp = &dscp
			} else if "SRC_IPV6" == ruleKey {
				addr := ruleData.Get(ruleKey)
				entrySet.Ipv6.Config.SourceAddress = &addr
				entrySet.Ipv6.State.SourceAddress = &addr
			} else if "DST_IPV6" == ruleKey {
				addr := ruleData.Get(ruleKey)
				entrySet.Ipv6.Config.DestinationAddress = &addr
				entrySet.Ipv6.State.DestinationAddress = &addr
			}
		}
	} else if aclType == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2 {
		ygot.BuildEmptyTree(entrySet.L2)
		for ruleKey := range ruleData.Field {
			if "ETHER_TYPE" == ruleKey {
				ethType, _ := strconv.ParseUint(strings.Replace(ruleData.Get(ruleKey), "0x", "", -1), 16, 32)
				ethertype := getL2EtherType(ethType)
				entrySet.L2.Config.Ethertype, _ = entrySet.L2.Config.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union(ethertype)
				entrySet.L2.State.Ethertype, _ = entrySet.L2.State.To_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_State_Ethertype_Union(ethertype)
			}
		}
	}
}

func convertInternalToOCAclRuleBinding(d *db.DB, priority uint32, seqId int64, direction string, aclSet ygot.GoStruct, entrySet ygot.GoStruct) {
	if seqId == -1 {
		seqId = int64(MAX_PRIORITY - priority)
	}

	var num uint64
	num = 0
	var ruleId uint32 = uint32(seqId)

	if direction == "INGRESS" {
		var ingressEntrySet *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_AclEntries_AclEntry
		var ok bool
		if entrySet == nil {
			ingressAclSet := aclSet.(*ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet)
			if ingressEntrySet, ok = ingressAclSet.AclEntries.AclEntry[ruleId]; !ok {
				ingressEntrySet, _ = ingressAclSet.AclEntries.NewAclEntry(ruleId)
			}
		} else {
			ingressEntrySet = entrySet.(*ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_AclEntries_AclEntry)
		}
		if ingressEntrySet != nil {
			ygot.BuildEmptyTree(ingressEntrySet)
			ingressEntrySet.State.SequenceId = &ruleId
			ingressEntrySet.State.MatchedPackets = &num
			ingressEntrySet.State.MatchedOctets = &num
		}
	} else if direction == "EGRESS" {
		var egressEntrySet *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_AclEntries_AclEntry
		var ok bool
		if entrySet == nil {
			egressAclSet := aclSet.(*ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet)
			if egressEntrySet, ok = egressAclSet.AclEntries.AclEntry[ruleId]; !ok {
				egressEntrySet, _ = egressAclSet.AclEntries.NewAclEntry(ruleId)
			}
		} else {
			egressEntrySet = entrySet.(*ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_AclEntries_AclEntry)
		}
		if egressEntrySet != nil {
			ygot.BuildEmptyTree(egressEntrySet)
			egressEntrySet.State.SequenceId = &ruleId
			egressEntrySet.State.MatchedPackets = &num
			egressEntrySet.State.MatchedOctets = &num
		}
	}
}

func (app *AclApp) convertInternalToOCAclBinding(d *db.DB, aclName string, intfId string, direction string, intfAclSet ygot.GoStruct) error {
	var err error
	if _, ok := app.aclTableMap[aclName]; !ok {
		aclEntry, err1 := d.GetEntry(app.aclTs, db.Key{Comp: []string{aclName}})
		if err1 != nil {
			return err1
		}
		if !contains(aclEntry.GetList("ports"), intfId) {
			return tlerr.InvalidArgs("Acl %s not binded with %s", aclName, intfId)
		}
	}

	if _, ok := app.ruleTableMap[aclName]; !ok {
		ruleKeys, _ := d.GetKeys(app.ruleTs)
		for i, _ := range ruleKeys {
			rulekey := ruleKeys[i]
			// Rulekey has two keys, first aclkey and second rulename
			if rulekey.Get(0) == aclName && rulekey.Get(1) != "DEFAULT_RULE" {
				seqId, _ := strconv.Atoi(strings.Replace(rulekey.Get(1), "RULE_", "", 1))
				convertInternalToOCAclRuleBinding(d, 0, int64(seqId), direction, intfAclSet, nil)
			}
		}
	} else {
		for ruleName := range app.ruleTableMap[aclName] {
			if ruleName != "DEFAULT_RULE" {
				seqId, _ := strconv.Atoi(strings.Replace(ruleName, "RULE_", "", 1))
				convertInternalToOCAclRuleBinding(d, 0, int64(seqId), direction, intfAclSet, nil)
			}
		}
	}

	return err
}

func (app *AclApp) getAllBindingsInfo(d *db.DB) error {
	var err error
	acl := app.getAppRootObject()
	if len(app.aclTableMap) == 0 {
		aclKeys, _ := d.GetKeys(app.aclTs)
		for i, _ := range aclKeys {
			aclEntry, _ := d.GetEntry(app.aclTs, aclKeys[i])
			app.aclTableMap[(aclKeys[i]).Get(0)] = aclEntry
		}
	}
	var interfaces []string
	for aclName := range app.aclTableMap {
		aclData := app.aclTableMap[aclName]
		if len(aclData.Get("ports@")) > 0 {
			aclIntfs := aclData.GetList("ports")
			for i, _ := range aclIntfs {
				if !contains(interfaces, aclIntfs[i]) && aclIntfs[i] != "" {
					interfaces = append(interfaces, aclIntfs[i])
				}
			}
		}
	}

	for _, intfId := range interfaces {
		var intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface
		intfData, ok := acl.Interfaces.Interface[intfId]
		if !ok {
			intfData, _ = acl.Interfaces.NewInterface(intfId)
		}
		ygot.BuildEmptyTree(intfData)
		err = app.getAclBindingInfoForInterfaceData(d, intfData, intfId, "INGRESS")
		err = app.getAclBindingInfoForInterfaceData(d, intfData, intfId, "EGRESS")
	}
	return err
}

func (app *AclApp) getAclBindingInfoForInterfaceData(d *db.DB, intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface, intfId string, direction string) error {
	var err error
	if intfData != nil {
		intfData.Config.Id = intfData.Id
		intfData.State.Id = intfData.Id
	}
	if direction == "INGRESS" {
		if intfData.IngressAclSets != nil && len(intfData.IngressAclSets.IngressAclSet) > 0 {
			for ingressAclSetKey, _ := range intfData.IngressAclSets.IngressAclSet {
				aclName := strings.Replace(strings.Replace(ingressAclSetKey.SetName, " ", "_", -1), "-", "_", -1)
				aclType := ingressAclSetKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(ingressAclSetKey.Type)].Name
				aclKey := aclName + "_" + aclType

				ingressAclSet := intfData.IngressAclSets.IngressAclSet[ingressAclSetKey]
				if ingressAclSet != nil && ingressAclSet.AclEntries != nil && len(ingressAclSet.AclEntries.AclEntry) > 0 {
					for seqId, _ := range ingressAclSet.AclEntries.AclEntry {
						rulekey := "RULE_" + strconv.Itoa(int(seqId))
						entrySet := ingressAclSet.AclEntries.AclEntry[seqId]
						_, err := d.GetEntry(app.ruleTs, db.Key{Comp: []string{aclKey, rulekey}})
						if err != nil {
							return err
						}
						convertInternalToOCAclRuleBinding(d, 0, int64(seqId), direction, nil, entrySet)
					}
				} else {
					ygot.BuildEmptyTree(ingressAclSet)
					ingressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_Config{SetName: &aclName, Type: ingressAclSetKey.Type}
					ingressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_State{SetName: &aclName, Type: ingressAclSetKey.Type}
					err = app.convertInternalToOCAclBinding(d, aclKey, intfId, direction, ingressAclSet)
				}
			}
		} else {
			err = app.findAndGetAclBindingInfoForInterfaceData(d, intfId, direction, intfData)
		}
	} else if direction == "EGRESS" {
		if intfData.EgressAclSets != nil && len(intfData.EgressAclSets.EgressAclSet) > 0 {
			for egressAclSetKey, _ := range intfData.EgressAclSets.EgressAclSet {
				aclName := strings.Replace(strings.Replace(egressAclSetKey.SetName, " ", "_", -1), "-", "_", -1)
				aclType := egressAclSetKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(egressAclSetKey.Type)].Name
				aclKey := aclName + "_" + aclType

				egressAclSet := intfData.EgressAclSets.EgressAclSet[egressAclSetKey]
				if egressAclSet != nil && egressAclSet.AclEntries != nil && len(egressAclSet.AclEntries.AclEntry) > 0 {
					for seqId, _ := range egressAclSet.AclEntries.AclEntry {
						rulekey := "RULE_" + strconv.Itoa(int(seqId))
						entrySet := egressAclSet.AclEntries.AclEntry[seqId]
						_, err := d.GetEntry(app.ruleTs, db.Key{Comp: []string{aclKey, rulekey}})
						if err != nil {
							return err
						}
						convertInternalToOCAclRuleBinding(d, 0, int64(seqId), direction, nil, entrySet)
					}
				} else {
					ygot.BuildEmptyTree(egressAclSet)
					egressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_Config{SetName: &aclName, Type: egressAclSetKey.Type}
					egressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_State{SetName: &aclName, Type: egressAclSetKey.Type}
					err = app.convertInternalToOCAclBinding(d, aclKey, intfId, direction, egressAclSet)
				}
			}
		} else {
			err = app.findAndGetAclBindingInfoForInterfaceData(d, intfId, direction, intfData)
		}
	} else {
		log.Error("Unknown direction")
	}
	return err
}

func (app *AclApp) findAndGetAclBindingInfoForInterfaceData(d *db.DB, intfId string, direction string, intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface) error {
	var err error
	if len(app.aclTableMap) == 0 {
		aclKeys, _ := d.GetKeys(app.aclTs)
		for i, _ := range aclKeys {
			aclEntry, _ := d.GetEntry(app.aclTs, aclKeys[i])
			app.aclTableMap[aclKeys[i].Get(0)] = aclEntry
		}
	}

	for aclName, _ := range app.aclTableMap {
		aclData := app.aclTableMap[aclName]
		aclIntfs := aclData.GetList("ports")
		aclType := aclData.Get(ACL_TYPE)
		var aclOrigName string
		var aclOrigType ocbinds.E_OpenconfigAcl_ACL_TYPE
		if SONIC_ACL_TYPE_IPV4 == aclType {
			aclOrigName = strings.Replace(aclName, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
			aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4
		} else if SONIC_ACL_TYPE_IPV6 == aclType {
			aclOrigName = strings.Replace(aclName, "_"+OPENCONFIG_ACL_TYPE_IPV6, "", 1)
			aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6
		} else if SONIC_ACL_TYPE_L2 == aclType {
			aclOrigName = strings.Replace(aclName, "_"+OPENCONFIG_ACL_TYPE_L2, "", 1)
			aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2
		}

		if contains(aclIntfs, intfId) && direction == aclData.Get("stage") {
			if direction == "INGRESS" {
				if intfData.IngressAclSets != nil {
					aclSetKey := ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_Key{SetName: aclOrigName, Type: aclOrigType}
					ingressAclSet, ok := intfData.IngressAclSets.IngressAclSet[aclSetKey]
					if !ok {
						ingressAclSet, _ = intfData.IngressAclSets.NewIngressAclSet(aclOrigName, aclOrigType)
						ygot.BuildEmptyTree(ingressAclSet)
						ingressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_Config{SetName: &aclOrigName, Type: aclOrigType}
						ingressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_State{SetName: &aclOrigName, Type: aclOrigType}
					}
					err = app.convertInternalToOCAclBinding(d, aclName, intfId, direction, ingressAclSet)
					if err != nil {
						return err
					}
				}
			} else if direction == "EGRESS" {
				if intfData.EgressAclSets != nil {
					aclSetKey := ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_Key{SetName: aclOrigName, Type: aclOrigType}
					egressAclSet, ok := intfData.EgressAclSets.EgressAclSet[aclSetKey]
					if !ok {
						egressAclSet, _ = intfData.EgressAclSets.NewEgressAclSet(aclOrigName, aclOrigType)
						ygot.BuildEmptyTree(egressAclSet)
						egressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_Config{SetName: &aclOrigName, Type: aclOrigType}
						egressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_State{SetName: &aclOrigName, Type: aclOrigType}
					}
					err = app.convertInternalToOCAclBinding(d, aclName, intfId, direction, egressAclSet)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return err
}

/*func (app *AclApp) isInterfaceBindWithACL(d *db.DB, intfId string) bool {
	var isFound bool = false

	if len(app.aclTableMap) == 0 {
		aclKeys, _ := d.GetKeys(app.aclTs)
		for i, _ := range aclKeys {
			aclEntry, _ := d.GetEntry(app.aclTs, aclKeys[i])
			app.aclTableMap[(aclKeys[i]).Get(0)] = aclEntry
		}
	}

	var interfaces []string
	for aclName := range app.aclTableMap {
		aclData := app.aclTableMap[aclName]
		if len(aclData.Get("ports@")) > 0 {
			aclIntfs := aclData.GetList("ports")
			for i, _ := range aclIntfs {
				if !contains(interfaces, aclIntfs[i]) && aclIntfs[i] != "" {
					interfaces = append(interfaces, aclIntfs[i])
				}
			}
		}
	}

	isFound = contains(interfaces, intfId)
	return isFound
}*/

func (app *AclApp) handleBindingsDeletion(d *db.DB) error {
	var err error

	acl := app.getAppRootObject()
	aclKeys, _ := d.GetKeys(app.aclTs)
	for i, _ := range aclKeys {
		aclEntry, _ := d.GetEntry(app.aclTs, aclKeys[i])
		var isRequestedAclFound = false
		if len(aclEntry.GetList("ports")) > 0 {
			if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/interfaces/interface{}") {
				direction := aclEntry.Get("stage")
				if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/interfaces/interface{}/ingress-acl-sets") && direction != "INGRESS" {
					return tlerr.InvalidArgs("Acl %s is not Ingress", aclKeys[i].Get(0))
				}
				if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/interfaces/interface{}/egress-acl-sets") && direction != "EGRESS" {
					return tlerr.InvalidArgs("Acl %s is not Egress", aclKeys[i].Get(0))
				}
				for intfId := range acl.Interfaces.Interface {
					aclname, acltype := getAclKeysFromStrKey(aclKeys[i].Get(0), aclEntry.Get("type"))
					intfData := acl.Interfaces.Interface[intfId]
					if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/interfaces/interface{}/ingress-acl-sets/ingress-acl-set{}{}") {
						for k := range intfData.IngressAclSets.IngressAclSet {
							if aclname == k.SetName {
								if acltype == k.Type {
									isRequestedAclFound = true
								} else {
									return tlerr.InvalidArgs("Acl Type is not matching")
								}
							} else {
								goto SkipDBProcessing
							}
						}
					} else if isSubtreeRequest(app.pathInfo.Template, "/openconfig-acl:acl/interfaces/interface{}/egress-acl-sets/egress-acl-set{}{}") {
						for k := range intfData.EgressAclSets.EgressAclSet {
							if aclname == k.SetName {
								if acltype == k.Type {
									isRequestedAclFound = true
								} else {
									return tlerr.InvalidArgs("Acl Type is not matching")
								}
							} else {
								goto SkipDBProcessing
							}
						}
					}
					intfs := aclEntry.GetList("ports")
					intfs = removeElement(intfs, intfId)
					aclEntry.SetList("ports", intfs)
					err = d.SetEntry(app.aclTs, aclKeys[i], aclEntry)
					if err != nil {
						return err
					}
					// If last interface removed, then remove stage field also
					if len(intfs) == 0 {
						aclEntry.Remove("stage")
					}
				}
			SkipDBProcessing:
			} else {
				aclEntry.Remove("stage")
				aclEntry.SetList("ports", []string{})
				err = d.SetEntry(app.aclTs, aclKeys[i], aclEntry)
				if err != nil {
					return err
				}
			}
		}
		if isRequestedAclFound {
			break
		}
	}

	return err
}

/********************   CREATE related    *******************************/
func (app *AclApp) convertOCAclsToInternal() {
	acl := app.getAppRootObject()
	if acl != nil {
		app.aclTableMap = make(map[string]db.Value)
		if acl.AclSets != nil && len(acl.AclSets.AclSet) > 0 {
			for aclSetKey, _ := range acl.AclSets.AclSet {
				aclSet := acl.AclSets.AclSet[aclSetKey]
				aclKey := getAclKeyStrFromOCKey(aclSetKey.Name, aclSetKey.Type)
				app.aclTableMap[aclKey] = db.Value{Field: map[string]string{}}

				if aclSet.Config != nil {
					if aclSet.Config.Type == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 {
						app.aclTableMap[aclKey].Field[ACL_TYPE] = SONIC_ACL_TYPE_IPV4
					} else if aclSet.Config.Type == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6 {
						app.aclTableMap[aclKey].Field[ACL_TYPE] = SONIC_ACL_TYPE_IPV6
					} else if aclSet.Config.Type == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2 {
						app.aclTableMap[aclKey].Field[ACL_TYPE] = SONIC_ACL_TYPE_L2
					}

					if aclSet.Config.Description != nil && len(*aclSet.Config.Description) > 0 {
						app.aclTableMap[aclKey].Field[ACL_DESCRIPTION] = *aclSet.Config.Description
					}
				}
			}
		}
	}
}

func (app *AclApp) convertOCAclRulesToInternal(d *db.DB) {
	acl := app.getAppRootObject()
	if acl != nil {
		app.ruleTableMap = make(map[string]map[string]db.Value)
		if acl.AclSets != nil && len(acl.AclSets.AclSet) > 0 {
			for aclSetKey, _ := range acl.AclSets.AclSet {
				aclSet := acl.AclSets.AclSet[aclSetKey]
				aclKey := getAclKeyStrFromOCKey(aclSetKey.Name, aclSetKey.Type)
				app.ruleTableMap[aclKey] = make(map[string]db.Value)

				if aclSet.AclEntries != nil {
					for seqId, _ := range aclSet.AclEntries.AclEntry {
						entrySet := aclSet.AclEntries.AclEntry[seqId]
						ruleName := "RULE_" + strconv.Itoa(int(seqId))
						app.ruleTableMap[aclKey][ruleName] = db.Value{Field: map[string]string{}}
						convertOCAclRuleToInternalAclRule(app.ruleTableMap[aclKey][ruleName], seqId, aclKey, aclSet.Type, entrySet)
					}
				}

				yangPathStr, _ := getYangPathFromUri(app.pathInfo.Path)
				if yangPathStr != "/openconfig-acl:acl/acl-sets/acl-set/acl-entries" && yangPathStr != "/openconfig-acl:acl/acl-sets/acl-set/acl-entries/acl-entry" {
					app.createDefaultDenyAclRule(d, aclKey, app.ruleTableMap[aclKey])
				}
			}
		}
	}
}

func (app *AclApp) convertOCAclBindingsToInternal() {
	aclObj := app.getAppRootObject()

	if aclObj.Interfaces != nil && len(aclObj.Interfaces.Interface) > 0 {
		aclInterfacesMap := make(map[string][]string)
		// Below code assumes that an ACL can be either INGRESS or EGRESS but not both.
		for intfId, _ := range aclObj.Interfaces.Interface {
			intf := aclObj.Interfaces.Interface[intfId]
			if intf != nil {
				if intf.IngressAclSets != nil && len(intf.IngressAclSets.IngressAclSet) > 0 {
					for inAclKey, _ := range intf.IngressAclSets.IngressAclSet {
						aclName := getAclKeyStrFromOCKey(inAclKey.SetName, inAclKey.Type)
						// TODO: Need to handle Subinterface also
						if intf.InterfaceRef != nil && intf.InterfaceRef.Config.Interface != nil {
							aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.InterfaceRef.Config.Interface)
						} else {
							aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.Id)
						}
						if len(app.aclTableMap) == 0 {
							app.aclTableMap[aclName] = db.Value{Field: map[string]string{}}
						}
						app.aclTableMap[aclName].Field["stage"] = "INGRESS"
					}
				}

				if intf.EgressAclSets != nil && len(intf.EgressAclSets.EgressAclSet) > 0 {
					for outAclKey, _ := range intf.EgressAclSets.EgressAclSet {
						aclName := getAclKeyStrFromOCKey(outAclKey.SetName, outAclKey.Type)
						if intf.InterfaceRef != nil && intf.InterfaceRef.Config.Interface != nil {
							aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.InterfaceRef.Config.Interface)
						} else {
							aclInterfacesMap[aclName] = append(aclInterfacesMap[aclName], *intf.Id)
						}
						if len(app.aclTableMap) == 0 {
							app.aclTableMap[aclName] = db.Value{Field: map[string]string{}}
						}
						app.aclTableMap[aclName].Field["stage"] = "EGRESS"
					}
				}
			}
		}
		for k, _ := range aclInterfacesMap {
			val := app.aclTableMap[k]
			(&val).SetList("ports", aclInterfacesMap[k])
		}
	}
}

func (app *AclApp) createDefaultDenyAclRule(d *db.DB, aclName string, rulesInfo map[string]db.Value) {
	existingRuleEntry, err := d.GetEntry(app.ruleTs, db.Key{Comp: []string{aclName, "DEFAULT_RULE"}})
	// If Default Rule already exists, Do not add new Default Rule
	if existingRuleEntry.IsPopulated() && err == nil {
		return
	}
	m := make(map[string]string)
	rulesInfo["DEFAULT_RULE"] = db.Value{Field: m}
	rulesInfo["DEFAULT_RULE"].Field["PRIORITY"] = strconv.FormatInt(int64(MIN_PRIORITY), 10)
	rulesInfo["DEFAULT_RULE"].Field["PACKET_ACTION"] = "DROP"
	rulesInfo["DEFAULT_RULE"].Field["IP_TYPE"] = "ANY"
}

func convertOCAclRuleToInternalAclRule(ruleData db.Value, seqId uint32, aclName string, aclType ocbinds.E_OpenconfigAcl_ACL_TYPE, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
	ruleIndex := seqId
	ruleData.Field["PRIORITY"] = strconv.FormatInt(int64(MAX_PRIORITY-ruleIndex), 10)
	// Rule Description is not supported in Sonic. So commenting this out.
	/*
	   if rule.Config != nil && rule.Config.Description != nil {
	       ruleData.Field["RULE_DESCRIPTION"] = *rule.Config.Description
	   }
	*/

	if ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 == aclType {
		convertOCToInternalIPv4(ruleData, aclName, ruleIndex, rule)
	} else if ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6 == aclType {
		convertOCToInternalIPv6(ruleData, aclName, ruleIndex, rule)
	} else if ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2 == aclType {
		convertOCToInternalL2(ruleData, aclName, ruleIndex, rule)
	} /*else if ocbinds.OpenconfigAcl_ACL_TYPE_ACL_MIXED == aclType {
	  } */

	convertOCToInternalTransport(ruleData, aclName, ruleIndex, rule)
	convertOCToInternalInputInterface(ruleData, aclName, ruleIndex, rule)
	convertOCToInternalInputAction(ruleData, aclName, ruleIndex, rule)
}

func convertOCToInternalL2(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
	if rule.L2 == nil {
		return
	}
	if rule.L2.Config.Ethertype != nil && util.IsTypeStructPtr(reflect.TypeOf(rule.L2.Config.Ethertype)) {
		ethertypeType := reflect.TypeOf(rule.L2.Config.Ethertype).Elem()
		var b bytes.Buffer
		switch ethertypeType {
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_E_OpenconfigPacketMatchTypes_ETHERTYPE{}):
			v := (rule.L2.Config.Ethertype).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_E_OpenconfigPacketMatchTypes_ETHERTYPE)
			//ruleData["ETHER_TYPE"] = v.E_OpenconfigPacketMatchTypes_ETHERTYPE.ΛMap()["E_OpenconfigPacketMatchTypes_ETHERTYPE"][int64(v.E_OpenconfigPacketMatchTypes_ETHERTYPE)].Name
			fmt.Fprintf(&b, "0x%0.4x", ETHERTYPE_MAP[v.E_OpenconfigPacketMatchTypes_ETHERTYPE])
			ruleData.Field["ETHER_TYPE"] = b.String()
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_Uint16{}):
			v := (rule.L2.Config.Ethertype).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_Uint16)
			fmt.Fprintf(&b, "0x%0.4x", v.Uint16)
			ruleData.Field["ETHER_TYPE"] = b.String()
		}
	}
}

func convertOCToInternalIPv4(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
	if rule.Ipv4.Config.Protocol != nil && util.IsTypeStructPtr(reflect.TypeOf(rule.Ipv4.Config.Protocol)) {
		protocolType := reflect.TypeOf(rule.Ipv4.Config.Protocol).Elem()
		switch protocolType {
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL{}):
			v := (rule.Ipv4.Config.Protocol).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
			//ruleData["IP_PROTOCOL"] = v.E_OpenconfigPacketMatchTypes_IP_PROTOCOL.ΛMap()["E_OpenconfigPacketMatchTypes_IP_PROTOCOL"][int64(v.E_OpenconfigPacketMatchTypes_IP_PROTOCOL)].Name
			ruleData.Field["IP_PROTOCOL"] = strconv.FormatInt(int64(IP_PROTOCOL_MAP[v.E_OpenconfigPacketMatchTypes_IP_PROTOCOL]), 10)
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8{}):
			v := (rule.Ipv4.Config.Protocol).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8)
			ruleData.Field["IP_PROTOCOL"] = strconv.FormatInt(int64(v.Uint8), 10)
		}
	}

	if rule.Ipv4.Config.Dscp != nil {
		ruleData.Field["DSCP"] = strconv.FormatInt(int64(*rule.Ipv4.Config.Dscp), 10)
	}
	if rule.Ipv4.Config.SourceAddress != nil {
		ruleData.Field["SRC_IP"] = *rule.Ipv4.Config.SourceAddress
	}
	if rule.Ipv4.Config.DestinationAddress != nil {
		ruleData.Field["DST_IP"] = *rule.Ipv4.Config.DestinationAddress
	}
}

func convertOCToInternalIPv6(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
	if rule.Ipv6.Config.Protocol != nil && util.IsTypeStructPtr(reflect.TypeOf(rule.Ipv6.Config.Protocol)) {
		protocolType := reflect.TypeOf(rule.Ipv6.Config.Protocol).Elem()
		switch protocolType {
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL{}):
			v := (rule.Ipv6.Config.Protocol).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
			//ruleData["IP_PROTOCOL"] = v.E_OpenconfigPacketMatchTypes_IP_PROTOCOL.ΛMap()["E_OpenconfigPacketMatchTypes_IP_PROTOCOL"][int64(v.E_OpenconfigPacketMatchTypes_IP_PROTOCOL)].Name
			ruleData.Field["IP_PROTOCOL"] = strconv.FormatInt(int64(IP_PROTOCOL_MAP[v.E_OpenconfigPacketMatchTypes_IP_PROTOCOL]), 10)
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_Uint8{}):
			v := (rule.Ipv6.Config.Protocol).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_Uint8)
			ruleData.Field["IP_PROTOCOL"] = strconv.FormatInt(int64(v.Uint8), 10)
		}
	}

	if rule.Ipv6.Config.Dscp != nil {
		ruleData.Field["DSCP"] = strconv.FormatInt(int64(*rule.Ipv6.Config.Dscp), 10)
	}
	if rule.Ipv6.Config.SourceAddress != nil {
		ruleData.Field["SRC_IPV6"] = *rule.Ipv6.Config.SourceAddress
	}
	if rule.Ipv6.Config.DestinationAddress != nil {
		ruleData.Field["DST_IPV6"] = *rule.Ipv6.Config.DestinationAddress
	}
	if rule.Ipv6.Config.SourceFlowLabel != nil {
		ruleData.Field["SRC_FLOWLABEL"] = strconv.FormatInt(int64(*rule.Ipv6.Config.SourceFlowLabel), 10)
	}
	if rule.Ipv6.Config.DestinationFlowLabel != nil {
		ruleData.Field["DST_FLOWLABEL"] = strconv.FormatInt(int64(*rule.Ipv6.Config.DestinationFlowLabel), 10)
	}
}

func convertOCToInternalTransport(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
	if rule.Transport == nil {
		return
	}
	if rule.Transport.Config.SourcePort != nil && util.IsTypeStructPtr(reflect.TypeOf(rule.Transport.Config.SourcePort)) {
		sourceportType := reflect.TypeOf(rule.Transport.Config.SourcePort).Elem()
		switch sourceportType {
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort{}):
			v := (rule.Transport.Config.SourcePort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort)
			ruleData.Field["L4_SRC_PORT"] = v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort.ΛMap()["E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort"][int64(v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort)].Name
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_String{}):
			v := (rule.Transport.Config.SourcePort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_String)
			ruleData.Field["L4_SRC_PORT_RANGE"] = strings.Replace(v.String, "..", "-", 1)
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16{}):
			v := (rule.Transport.Config.SourcePort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16)
			ruleData.Field["L4_SRC_PORT"] = strconv.FormatInt(int64(v.Uint16), 10)
		}
	}

	if rule.Transport.Config.DestinationPort != nil && util.IsTypeStructPtr(reflect.TypeOf(rule.Transport.Config.DestinationPort)) {
		destportType := reflect.TypeOf(rule.Transport.Config.DestinationPort).Elem()
		switch destportType {
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort{}):
			v := (rule.Transport.Config.DestinationPort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort)
			ruleData.Field["L4_DST_PORT"] = v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort.ΛMap()["E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort"][int64(v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort)].Name
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_String{}):
			v := (rule.Transport.Config.DestinationPort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_String)
			ruleData.Field["L4_DST_PORT_RANGE"] = strings.Replace(v.String, "..", "-", 1)
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16{}):
			v := (rule.Transport.Config.DestinationPort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16)
			ruleData.Field["L4_DST_PORT"] = strconv.FormatInt(int64(v.Uint16), 10)
		}
	}

	var tcpFlags uint32 = 0x00
	if len(rule.Transport.Config.TcpFlags) > 0 {
		for _, flag := range rule.Transport.Config.TcpFlags {
			switch flag {
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_FIN:
				tcpFlags |= 0x01
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_SYN:
				tcpFlags |= 0x02
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_RST:
				tcpFlags |= 0x04
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_PSH:
				tcpFlags |= 0x08
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ACK:
				tcpFlags |= 0x10
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_URG:
				tcpFlags |= 0x20
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ECE:
				tcpFlags |= 0x40
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_CWR:
				tcpFlags |= 0x80
			}
		}
		var b bytes.Buffer
		fmt.Fprintf(&b, "0x%0.2x/0x%0.2x", tcpFlags, tcpFlags)
		ruleData.Field["TCP_FLAGS"] = b.String()
	}
}

func convertOCToInternalInputInterface(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
	if rule.InputInterface != nil && rule.InputInterface.InterfaceRef != nil {
		ruleData.Field["IN_PORTS"] = *rule.InputInterface.InterfaceRef.Config.Interface
	}
}

func convertOCToInternalInputAction(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
	if rule.Actions != nil && rule.Actions.Config != nil {
		switch rule.Actions.Config.ForwardingAction {
		case ocbinds.OpenconfigAcl_FORWARDING_ACTION_ACCEPT:
			ruleData.Field["PACKET_ACTION"] = "FORWARD"
		case ocbinds.OpenconfigAcl_FORWARDING_ACTION_DROP, ocbinds.OpenconfigAcl_FORWARDING_ACTION_REJECT:
			ruleData.Field["PACKET_ACTION"] = "DROP"
		default:
		}
	}
}

func (app *AclApp) handleRuleFieldsDeletion(d *db.DB, aclKey string, ruleKey string) error {
	var err error

	ruleEntry, err := d.GetEntry(app.ruleTs, asKey(aclKey, ruleKey))
	if err != nil {
		return err
	}
	nodeInfo, err := getTargetNodeYangSchema(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device))
	if err != nil {
		return err
	}
	if nodeInfo.IsLeaf() {
		switch nodeInfo.Name {
		case "description":
			(&ruleEntry).Remove("RULE_DESCRIPTION")
		// L2
		case "ethertype":
			(&ruleEntry).Remove("ETHER_TYPE")
			// IPv4/IPv6
		case "source-address":
			if strings.Contains(app.pathInfo.Path, "ipv4/config") {
				(&ruleEntry).Remove("SRC_IP")
			} else if strings.Contains(app.pathInfo.Path, "ipv6/config") {
				(&ruleEntry).Remove("SRC_IPV6")
			}
		case "destination-address":
			if strings.Contains(app.pathInfo.Path, "ipv4/config") {
				(&ruleEntry).Remove("DST_IP")
			} else if strings.Contains(app.pathInfo.Path, "ipv6/config") {
				(&ruleEntry).Remove("DST_IPV6")
			}
		case "dscp":
			(&ruleEntry).Remove("DSCP")
		case "protocol":
			(&ruleEntry).Remove("IP_PROTOCOL")
			// transport
		case "source-port":
			(&ruleEntry).Remove("L4_SRC_PORT")
			(&ruleEntry).Remove("L4_SRC_PORT_RANGE")
		case "destination-port":
			(&ruleEntry).Remove("L4_DST_PORT")
			(&ruleEntry).Remove("L4_DST_PORT_RANGE")
			// actions
		case "forwarding-action":
			(&ruleEntry).Remove("PACKET_ACTION")
			//input-interface
		case "interface":
			(&ruleEntry).Remove("IN_PORTS")
			//case "subinterface":
		}
	} else if nodeInfo.IsContainer() {
		targetType := reflect.TypeOf(*app.ygotTarget)
		switch targetType.Elem().Name() {
		case "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2", "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config":
			(&ruleEntry).Remove("ETHER_TYPE")
		case "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4", "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config":
			(&ruleEntry).Remove("IP_PROTOCOL")
			(&ruleEntry).Remove("SRC_IP")
			(&ruleEntry).Remove("DST_IP")
			(&ruleEntry).Remove("DSCP")
		case "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6", "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config":
			(&ruleEntry).Remove("IP_PROTOCOL")
			(&ruleEntry).Remove("SRC_IPV6")
			(&ruleEntry).Remove("DST_IPV6")
			(&ruleEntry).Remove("DSCP")
		case "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport", "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config":
			(&ruleEntry).Remove("L4_SRC_PORT")
			(&ruleEntry).Remove("L4_SRC_PORT_RANGE")
			(&ruleEntry).Remove("L4_DST_PORT")
			(&ruleEntry).Remove("L4_DST_PORT_RANGE")
			(&ruleEntry).Remove("TCP_FLAGS")
		case "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_InputInterface", "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_InputInterface_InterfaceRef", "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_InputInterface_InterfaceRef_Config":
			(&ruleEntry).Remove("IN_PORTS")
		case "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Actions", "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Actions_Config":
			(&ruleEntry).Remove("PACKET_ACTION")
		}
	} else if nodeInfo.IsLeafList() {
		switch nodeInfo.Name {
		case "tcp-flags":
			(&ruleEntry).Remove("TCP_FLAGS")
		}
	} else {
		log.Error("This yang type is not handled currently")
	}
	err = d.SetEntry(app.ruleTs, asKey(aclKey, ruleKey), ruleEntry)

	return err
}

func (app *AclApp) setAclDataInConfigDb(d *db.DB, aclData map[string]db.Value, createFlag bool) error {
	var err error
	for key := range aclData {
		existingEntry, err := d.GetEntry(app.aclTs, db.Key{Comp: []string{key}})
		// If Create ACL request comes and ACL already exists, throw error
		if createFlag && existingEntry.IsPopulated() {
			return tlerr.AlreadyExists("Acl %s already exists", key)
		}
		if createFlag || (!createFlag && err != nil && !existingEntry.IsPopulated()) {
			err := d.CreateEntry(app.aclTs, db.Key{Comp: []string{key}}, aclData[key])
			if err != nil {
				return err
			}
		} else {
			if existingEntry.IsPopulated() {
				if existingEntry.Get(ACL_DESCRIPTION) != aclData[key].Field[ACL_DESCRIPTION] {
					err := d.ModEntry(app.aclTs, db.Key{Comp: []string{key}}, aclData[key])
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return err
}

func (app *AclApp) setAclRuleDataInConfigDb(d *db.DB, ruleData map[string]map[string]db.Value, createFlag bool) error {
	var err error
	for aclName := range ruleData {
		for ruleName := range ruleData[aclName] {
			existingRuleEntry, err := d.GetEntry(app.ruleTs, db.Key{Comp: []string{aclName, ruleName}})
			// If Create Rule request comes and Rule already exists, throw error
			if createFlag && existingRuleEntry.IsPopulated() {
				return tlerr.AlreadyExists("Rule %s already exists", ruleName)
			}
			if createFlag || (!createFlag && err != nil && !existingRuleEntry.IsPopulated()) {
				err := d.CreateEntry(app.ruleTs, db.Key{Comp: []string{aclName, ruleName}}, ruleData[aclName][ruleName])
				if err != nil {
					return err
				}
			} else {
				if existingRuleEntry.IsPopulated() && ruleName != "DEFAULT_RULE" {
					err := d.ModEntry(app.ruleTs, db.Key{Comp: []string{aclName, ruleName}}, ruleData[aclName][ruleName])
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return err
}

func (app *AclApp) setAclBindDataInConfigDb(d *db.DB, aclData map[string]db.Value, opcode int) error {
	var err error
	for aclKey, aclInfo := range aclData {
		// Get ACL info from DB
		dbAcl, err := d.GetEntry(app.aclTs, db.Key{Comp: []string{aclKey}})
		if err != nil {
			return err
		}
		if REPLACE == opcode {
			dbAcl.SetList("ports", aclInfo.GetList("ports"))
			dbAcl.Set("stage", aclInfo.Get("stage"))
		} else {
			dbAclIntfs := dbAcl.GetList("ports")
			if len(dbAclIntfs) > 0 {
				dbAclDirec := dbAcl.Get("stage")
				newDirec := aclInfo.Get("stage")
				if (UPDATE == opcode) && (len(dbAclDirec) > 0) && (len(newDirec) > 0) && (dbAclDirec != newDirec) {
					return tlerr.InvalidArgs("Acl direction of %s not allowed when it is already configured as %s", newDirec, dbAclDirec)
				}
				// Merge interfaces from DB to list in aclInfo and set back in DB
				intfs := aclInfo.GetList("ports")
				for _, ifId := range dbAclIntfs {
					if !contains(intfs, ifId) {
						intfs = append(intfs, ifId)
					}
				}
				dbAcl.SetList("ports", intfs)
			} else {
				dbAcl.SetList("ports", aclInfo.GetList("ports"))
			}

			if len(dbAcl.Get("stage")) == 0 {
				dbAcl.Set("stage", aclInfo.Get("stage"))
			}
		}
		err = d.SetEntry(app.aclTs, db.Key{Comp: []string{aclKey}}, dbAcl)
		//err = d.ModEntry(app.aclTs, db.Key{Comp: []string{aclKey}}, dbAcl)
		if err != nil {
			return err
		}
	}
	return err
}

func getIpProtocol(proto int64) interface{} {
	for k, v := range IP_PROTOCOL_MAP {
		if uint8(proto) == v {
			return k
		}
	}
	return uint8(proto)
}

func getTransportSrcDestPorts(portVal string, portType string) interface{} {
	var portRange string = ""

	portNum, err := strconv.Atoi(portVal)
	if err != nil && strings.Contains(portVal, "-") {
		portRange = portVal
	}

	if len(portRange) > 0 {
		return portRange
	} else if portNum > 0 {
		return uint16(portNum)
	} else {
		if "src" == portType {
			return ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_ANY
		} else if "dest" == portType {
			return ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_ANY
		}
	}
	return nil
}

func getTransportConfigTcpFlags(tcpFlags string) []ocbinds.E_OpenconfigPacketMatchTypes_TCP_FLAGS {
	var flags []ocbinds.E_OpenconfigPacketMatchTypes_TCP_FLAGS
	if len(tcpFlags) > 0 {
		flagStr := strings.Split(tcpFlags, "/")[0]
		flagNumber, _ := strconv.ParseUint(strings.Replace(flagStr, "0x", "", -1), 16, 32)
		for i := 0; i < 8; i++ {
			mask := 1 << uint(i)
			if (int(flagNumber) & mask) > 0 {
				switch int(flagNumber) & mask {
				case 0x01:
					flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_FIN)
				case 0x02:
					flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_SYN)
				case 0x04:
					flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_RST)
				case 0x08:
					flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_PSH)
				case 0x10:
					flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ACK)
				case 0x20:
					flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_URG)
				case 0x40:
					flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ECE)
				case 0x80:
					flags = append(flags, ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_CWR)
				default:
				}
			}
		}
	}
	return flags
}

func getL2EtherType(etherType uint64) interface{} {
	for k, v := range ETHERTYPE_MAP {
		if uint32(etherType) == v {
			return k
		}
	}
	return uint16(etherType)
}

func getAclKeysFromStrKey(aclKey string, aclType string) (string, ocbinds.E_OpenconfigAcl_ACL_TYPE) {
	var aclOrigName string
	var aclOrigType ocbinds.E_OpenconfigAcl_ACL_TYPE

	if SONIC_ACL_TYPE_IPV4 == aclType {
		aclOrigName = strings.Replace(aclKey, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
		aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4
	} else if SONIC_ACL_TYPE_IPV6 == aclType {
		aclOrigName = strings.Replace(aclKey, "_"+OPENCONFIG_ACL_TYPE_IPV6, "", 1)
		aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6
	} else if SONIC_ACL_TYPE_L2 == aclType {
		aclOrigName = strings.Replace(aclKey, "_"+OPENCONFIG_ACL_TYPE_L2, "", 1)
		aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2
	}
	return aclOrigName, aclOrigType
}

// getAclTypeOCEnumFromName returns the ACL_TYPE enum from name
func getAclTypeOCEnumFromName(val string) (ocbinds.E_OpenconfigAcl_ACL_TYPE, error) {
	switch val {
	case "ACL_IPV4", "openconfig-acl:ACL_IPV4":
		return ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4, nil
	case "ACL_IPV6", "openconfig-acl:ACL_IPV6":
		return ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6, nil
	case "ACL_L2", "openconfig-acl:ACL_L2":
		return ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2, nil
	default:
		return ocbinds.OpenconfigAcl_ACL_TYPE_UNSET,
			tlerr.NotSupported("ACL Type '%s' not supported", val)
	}
}

func getAclKeyStrFromOCKey(aclname string, acltype ocbinds.E_OpenconfigAcl_ACL_TYPE) string {
	aclN := strings.Replace(strings.Replace(aclname, " ", "_", -1), "-", "_", -1)
	aclT := acltype.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(acltype)].Name
	return aclN + "_" + aclT
}

/* Check if targetUriPath is child (subtree) of nodePath
The return value can be used to decide if subtrees needs
to visited to fill the data or not.
*/
func isSubtreeRequest(targetUriPath string, nodePath string) bool {
	return strings.HasPrefix(targetUriPath, nodePath)
}
