///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package translib

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/util"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"strconv"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/transformer"
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
	MAX_PRIORITY = 65535
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
	path       string
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}

	aclTs  *db.TableSpec
	ruleTs *db.TableSpec

	aclTableMap  map[string]db.Value
	ruleTableMap map[string]map[string]db.Value

	createAclFlag  bool
	createRuleFlag bool
	bindAclFlag    bool
}

func init() {

	err := register("/openconfig-acl:acl",
        &appInfo{appType:  reflect.TypeOf(AclApp{}),
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

	yangFiles := []string{"openconfig-acl.yang", "sonic-acl.yang"}
        log.Info("Init transformer yang files :", yangFiles)
	err = transformer.LoadYangModules(yangFiles...)
	if err != nil {
		log.Fatal("Loading Yang modules failed with error=", err)
	}
}

func (app *AclApp) initialize(data appData) {
	log.Info("initialize:acl:path =", data.path)
	*app = AclApp{path: data.path, ygotRoot: data.ygotRoot, ygotTarget: data.ygotTarget}

	app.aclTs = &db.TableSpec{Name: ACL_TABLE}
	app.ruleTs = &db.TableSpec{Name: RULE_TABLE}

	app.aclTableMap = make(map[string]db.Value)
	app.ruleTableMap = make(map[string]map[string]db.Value)

	app.createAclFlag = false
	app.createRuleFlag = false
	app.bindAclFlag = false
}

func (app *AclApp) getAppRootObject() *ocbinds.OpenconfigAcl_Acl {
	deviceObj := (*app.ygotRoot).(*ocbinds.Device)
	return deviceObj.Acl
}

func (app *AclApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCreate:acl:path =", app.path)

	keys, err = app.translateCRUCommon(d, CREATE)

	return keys, err
}

func (app *AclApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateUpdate:acl:path =", app.path)

	keys, err = app.translateCRUCommon(d, UPDATE)

	return keys, err
}

func (app *AclApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateReplace:acl:path =", app.path)

	//keys, err = app.translateCRUCommon(d, REPLACE)

	err = errors.New("Not implemented")
	return keys, err
}

func (app *AclApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateDelete:acl:path =", app.path)

	keys, err = app.generateDbWatchKeys(d, true)

	return keys, err
}

func (app *AclApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	var err error
	log.Info("translateGet:acl:path =", app.path)
	return err
}

func (app *AclApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
    err := errors.New("Not supported")
    configDb := dbs[db.ConfigDB]
    pathInfo := NewPathInfo(path)
    notifInfo := notificationInfo{dbno: db.ConfigDB}

    if isSubtreeRequest(pathInfo.Template, "/openconfig-acl:acl/acl-sets") {
        if isSubtreeRequest(pathInfo.Template, "/openconfig-acl:acl/acl-sets/acl-set{name}{type}") {
            aclN := strings.Replace(strings.Replace(pathInfo.Var("name"), " ", "_", -1), "-", "_", -1)
            aclT := pathInfo.Var("type")
            if OPENCONFIG_ACL_TYPE_IPV4 != aclT && OPENCONFIG_ACL_TYPE_IPV6 != aclT && OPENCONFIG_ACL_TYPE_L2 != aclT {
                err = errors.New("Invalid ACL Type")
                return nil, nil, err
            }
            aclkey := aclN + "_" + aclT
            if isSubtreeRequest(pathInfo.Template, "/openconfig-acl:acl/acl-sets/acl-set{name}{type}/acl-entries/acl-entry{sequence-id}") {
                rulekey := "RULE_" + pathInfo.Var("sequence-id")
                notifInfo.table = db.TableSpec{Name: RULE_TABLE}
                notifInfo.key = db.Key{Comp: []string{aclkey, rulekey}}
            } else {
                // All Rules of a given Acl
                if pathInfo.Template == "/openconfig-acl:acl/acl-sets/acl-set{name}{type}/acl-entries" {
                    notifInfo.table = db.TableSpec{Name: RULE_TABLE}
                } else {
                    notifInfo.table = db.TableSpec{Name: ACL_TABLE}
                    notifInfo.key = db.Key{Comp: []string{aclkey}}
                }
            }
        } else {
            // All Acls and their rules
            notifInfo.table = db.TableSpec{Name: ACL_TABLE}
        }
    } else if isSubtreeRequest(pathInfo.Template, "/openconfig-acl:acl/interfaces") {
        if isSubtreeRequest(pathInfo.Template, "/openconfig-acl:acl/interfaces/interface{id}") {
            // With one interface, multiple ACLs can be binded. Need mehanism to pass multiple Keys
            var notifKeys []db.Key
            intfId := pathInfo.Var("id")
            aclKeys, _ := configDb.GetKeys(app.aclTs)
            for i, _ := range aclKeys {
                aclEntry, _ := configDb.GetEntry(app.aclTs, aclKeys[i])
                aclIntfs := aclEntry.GetList("ports")
                if contains(aclIntfs, intfId) {
                    notifKeys = append(notifKeys, aclKeys[i])
                }
            }
        }
        notifInfo.table = db.TableSpec{Name: ACL_TABLE}
    } else {
        // Topmost path
        notifInfo.table = db.TableSpec{Name: ACL_TABLE}
    }

    return nil, &notifInfo, err
}

func (app *AclApp) processCreate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processCreate:acl:path =", app.path)
	targetType := reflect.TypeOf(*app.ygotTarget)
	log.Infof("processCreate: Target object is a <%s> of Type: %s", targetType.Kind().String(), targetType.Elem().Name())

	if app.createAclFlag {
		err = app.setAclDataInConfigDb(d, app.aclTableMap, true)
		if err != nil {
			log.Error(err)
			return resp, err
		}
	}
	if app.createRuleFlag {
		err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, true)
		if err != nil {
			log.Error(err)
			return resp, err
		}
	}
	if app.bindAclFlag && !app.createAclFlag {
		err = app.setAclBindDataInConfigDb(d, app.aclTableMap)
	}

	return resp, err
}

func (app *AclApp) processUpdate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processUpdate:acl:path =", app.path)

	if app.createAclFlag {
		err = app.setAclDataInConfigDb(d, app.aclTableMap, false)
		if err != nil {
			log.Error(err)
			return resp, err
		}
	}
	if app.createRuleFlag {
		err = app.setAclRuleDataInConfigDb(d, app.ruleTableMap, false)
		if err != nil {
			log.Error(err)
			return resp, err
		}
	}
	if app.bindAclFlag && !app.createAclFlag {
		err = app.setAclBindDataInConfigDb(d, app.aclTableMap)
	}

	return resp, err
}

func (app *AclApp) processReplace(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processReplace:acl:path =", app.path)
	err = errors.New("Not implemented")
	return resp, err
}

func (app *AclApp) processDelete(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	var aclSubtree = false
	log.Info("processDelete:acl:path =", app.path)

	aclObj := app.getAppRootObject()
	if reflect.TypeOf(*app.ygotTarget).Elem().Name() == "OpenconfigAcl_Acl" {
		aclSubtree = true
	}
	targetUriPath, err := getYangPathFromUri(app.path)
	if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets") || aclSubtree {
		if aclObj.AclSets != nil && len(aclObj.AclSets.AclSet) > 0 {
			// Deletion of a specific ACL
			for aclSetKey, _ := range aclObj.AclSets.AclSet {
				aclKey := getAclKeyStrFromOCKey(aclSetKey.Name, aclSetKey.Type)
				aclSet := aclObj.AclSets.AclSet[aclSetKey]
				if aclSet.AclEntries != nil && len(aclSet.AclEntries.AclEntry) > 0 {
					// Deletion of a specific Rule
					for seqId, _ := range aclSet.AclEntries.AclEntry {
						ruleName := "RULE_" + strconv.FormatInt(int64(seqId), 10)
						err = d.DeleteEntry(app.ruleTs, db.Key{Comp: []string{aclKey, ruleName}})
						if err != nil {
							log.Error(err)
							resp = SetResponse{ErrSrc: AppErr}
							return resp, err
						}
					}
				} else {
					// Deletion of a specific Acl and all its rule
					if *app.ygotTarget == aclSet {
						err = d.DeleteKeys(app.ruleTs, db.Key{Comp: []string{aclKey + TABLE_SEPARATOR + "*"}})
						if err != nil {
							log.Error(err)
							resp = SetResponse{ErrSrc: AppErr}
							return resp, err
						}
						err = d.DeleteEntry(app.aclTs, db.Key{Comp: []string{aclKey}})
						if err != nil {
							log.Error(err)
							resp = SetResponse{ErrSrc: AppErr}
							return resp, err
						}
					}
					// Deletion of all rules for a specific ACL but NOT ACL
					if *app.ygotTarget == aclSet.AclEntries {
						err = d.DeleteKeys(app.ruleTs, db.Key{Comp: []string{aclKey + TABLE_SEPARATOR + "*"}})
						if err != nil {
							log.Error(err)
							resp = SetResponse{ErrSrc: AppErr}
							return resp, err
						}
					} else {
						err = d.DeleteKeys(app.ruleTs, db.Key{Comp: []string{aclKey + TABLE_SEPARATOR + "*"}})
						if err != nil {
							log.Error(err)
							resp = SetResponse{ErrSrc: AppErr}
							return resp, err
						}
					}
				}
			}
		} else {
			// Deletion of All ACLs and Rules
            err = d.DeleteTable(app.ruleTs)
			if err != nil {
				log.Error(err)
				resp = SetResponse{ErrSrc: AppErr}
				return resp, err
			}
            err = d.DeleteTable(app.aclTs)
			if err != nil {
				log.Error(err)
				resp = SetResponse{ErrSrc: AppErr}
				return resp, err
			}
		}
	} else if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces") {
		aclKeys, _ := d.GetKeys(app.aclTs)
		for i, _ := range aclKeys {
			aclEntry, _ := d.GetEntry(app.aclTs, aclKeys[i])
			var isRequestedAclFound = false
			if len(aclEntry.GetList("ports")) > 0 {
				if aclObj.Interfaces != nil && len(aclObj.Interfaces.Interface) > 0 {
					direction := aclEntry.Get("stage")
					for intfId := range aclObj.Interfaces.Interface {
						if targetUriPath == "/openconfig-acl:acl/interfaces/interface/ingress-acl-sets" && direction != "INGRESS" {
							resp = SetResponse{ErrSrc: AppErr}
							err = errors.New("Acl is not Ingress")
							return resp, err
						}
						if targetUriPath == "/openconfig-acl:acl/interfaces/interface/egress-acl-sets" && direction != "EGRESS" {
							resp = SetResponse{ErrSrc: AppErr}
							err = errors.New("Acl is not Egress")
							return resp, err
						}

						aclname, acltype := getAclKeysFromStrKey(aclKeys[i].Get(0), aclEntry.Get("type"))
						if targetUriPath == "/openconfig-acl:acl/interfaces/interface/ingress-acl-sets/ingress-acl-set" {
							intfData := aclObj.Interfaces.Interface[intfId]
							for k := range intfData.IngressAclSets.IngressAclSet {
								if aclname == k.SetName {
									if acltype == k.Type {
										isRequestedAclFound = true
									} else {
										err = errors.New("Acl Type is not maching")
										resp = SetResponse{ErrSrc: AppErr}
										return resp, err
									}
								} else {
									goto SkipDBProcessing
								}
							}
						} else if targetUriPath == "/openconfig-acl:acl/interfaces/interface/egress-acl-sets/egress-acl-set" {
							intfData := aclObj.Interfaces.Interface[intfId]
							for k := range intfData.EgressAclSets.EgressAclSet {
								if aclname == k.SetName {
									if acltype == k.Type {
										isRequestedAclFound = true
									} else {
										err = errors.New("Acl Type is not maching")
										resp = SetResponse{ErrSrc: AppErr}
										return resp, err
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
							log.Error(err)
							resp = SetResponse{ErrSrc: AppErr}
							return resp, err
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
						log.Error(err)
						resp = SetResponse{ErrSrc: AppErr}
						return resp, err
					}
				}
			}
			if isRequestedAclFound {
				break
			}
		}
	}

	return resp, err
}

func (app *AclApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
	var err error
	var payload []byte
	var aclSubtree bool = false
	var intfSubtree bool = false

	configDb := dbs[db.ConfigDB]
	aclObj := app.getAppRootObject()

	targetType := reflect.TypeOf(*app.ygotTarget)
	if !util.IsValueScalar(reflect.ValueOf(*app.ygotTarget)) && util.IsValuePtr(reflect.ValueOf(*app.ygotTarget)) {
		log.Infof("processGet: Target object is a <%s> of Type: %s", targetType.Kind().String(), targetType.Elem().Name())
		if targetType.Elem().Name() == "OpenconfigAcl_Acl" {
			aclSubtree = true
			intfSubtree = true
		}
	}

	targetUriPath, err := getYangPathFromUri(app.path)
	if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets") || aclSubtree {
		if aclObj.AclSets != nil && len(aclObj.AclSets.AclSet) > 0 {
			// Request for specific ACL
			for aclSetKey, _ := range aclObj.AclSets.AclSet {
				aclKey := getAclKeyStrFromOCKey(aclSetKey.Name, aclSetKey.Type)
				aclSet := aclObj.AclSets.AclSet[aclSetKey]

				if aclSet.AclEntries != nil && len(aclSet.AclEntries.AclEntry) > 0 {
					// Request for specific Rule
					for seqId, _ := range aclSet.AclEntries.AclEntry {
						//ruleKey := "RULE_" + strconv.FormatInt(int64(seqId), 10)
						entrySet := aclSet.AclEntries.AclEntry[seqId]
						err = app.convertDBAclRulesToInternal(configDb, aclKey, int64(seqId), db.Key{})
						if err != nil {
							return GetResponse{Payload: payload, ErrSrc: AppErr}, err
						}
						ygot.BuildEmptyTree(entrySet)
						app.convertInternalToOCAclRule(aclKey, aclSetKey.Type, int64(seqId), nil, entrySet)
					}
				} else {
					err = app.convertDBAclToInternal(configDb, db.Key{Comp: []string{aclKey}})
					if err != nil {
						return GetResponse{Payload: payload, ErrSrc: AppErr}, err
					}

					ygot.BuildEmptyTree(aclSet)
					app.convertInternalToOCAcl(aclKey, aclObj.AclSets, aclSet)
				}
			}
		} else {
			// Request for all ACLs
			ygot.BuildEmptyTree(aclObj)
			err = app.convertDBAclToInternal(configDb, db.Key{})
			if err != nil {
				return GetResponse{Payload: payload, ErrSrc: AppErr}, err
			}

			app.convertInternalToOCAcl("", aclObj.AclSets, nil)
			if err != nil {
				return GetResponse{Payload: payload, ErrSrc: AppErr}, err
			}
		}
	}

	if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces") || intfSubtree {
		if aclObj.Interfaces != nil && len(aclObj.Interfaces.Interface) > 0 {
			var intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface
			for intfId := range aclObj.Interfaces.Interface {
				intfData = aclObj.Interfaces.Interface[intfId]
				// Validate if given interface is bind with any ACL
				if !app.isInterfaceBindWithACL(configDb, intfId) {
					err = errors.New("Interface not bind with any ACL")
					return GetResponse{Payload: payload, ErrSrc: AppErr}, err
				}
				ygot.BuildEmptyTree(intfData)
				if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces/interface/ingress-acl-sets") {
					// Ingress ACL Specific
					app.getAclBindingInfoForInterfaceData(configDb, intfData, intfId, "INGRESS")
				} else if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces/interface/egress-acl-sets") {
					// Egress ACL Specific
					app.getAclBindingInfoForInterfaceData(configDb, intfData, intfId, "EGRESS")
				} else {
					// Direction unknown. Check ACL Table for binding information.
					fmt.Println("Request is for specific interface, ingress and egress ACLs")
					app.getAclBindingInfoForInterfaceData(configDb, intfData, intfId, "INGRESS")
					app.getAclBindingInfoForInterfaceData(configDb, intfData, intfId, "EGRESS")
				}
			}
		} else {
			fmt.Println("Request is for all interfaces and all directions on which ACL is applied")
			if len(app.aclTableMap) == 0 {
				// Get all ACLs
				app.convertDBAclToInternal(configDb, db.Key{})
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
				intfData, ok := aclObj.Interfaces.Interface[intfId]
				if !ok {
					intfData, _ = aclObj.Interfaces.NewInterface(intfId)
				}
				ygot.BuildEmptyTree(intfData)
				app.getAclBindingInfoForInterfaceData(configDb, intfData, intfId, "INGRESS")
				app.getAclBindingInfoForInterfaceData(configDb, intfData, intfId, "EGRESS")
			}
		}
	}

	payload, err = generateGetResponsePayload(app.path, (*app.ygotRoot).(*ocbinds.Device), app.ygotTarget)
	if err != nil {
		return GetResponse{Payload: payload, ErrSrc: AppErr}, err
	}

	return GetResponse{Payload: payload}, err
}

func (app *AclApp) translateCRUCommon(d *db.DB, opcode int) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCRUCommon:acl:path =", app.path)

	aclObj := app.getAppRootObject()
	
	// translate yang to db
	//payload, err := dumpIetfJson(aclObj, false)
	result, err := transformer.XlateToDb((*app).ygotRoot, (*app).ygotTarget)
	fmt.Println(result)
	
	app.aclTableMap = app.convertOCAclsToInternal(aclObj)
	app.ruleTableMap = app.convertOCAclRulesToInternal(aclObj)
	app.bindAclFlag, err = app.convertOCAclBindingsToInternal(d, app.aclTableMap, aclObj)

	if err != nil {
		log.Error(err)
		return keys, err
	}

	keys, err = app.generateDbWatchKeys(d, false)

	return keys, err
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
			return errors.New("ACL is not configured")
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
		if "L4_SRC_PORT" == ruleKey {
			port := ruleData.Get(ruleKey)
			entrySet.Transport.Config.SourcePort = getTransportConfigSrcPort(port)
			//entrySet.Transport.State.SourcePort = &addr
		} else if "L4_DST_PORT" == ruleKey {
			port := ruleData.Get(ruleKey)
			entrySet.Transport.Config.DestinationPort = getTransportConfigDestPort(port)
			//entrySet.Transport.State.DestinationPort = &addr
        } else if "TCP_FLAGS" == ruleKey {
            tcpFlags := ruleData.Get(ruleKey)
            entrySet.Transport.Config.TcpFlags = getTransportConfigTcpFlags(tcpFlags)
            entrySet.Transport.State.TcpFlags = getTransportConfigTcpFlags(tcpFlags)
		} else if "PACKET_ACTION" == ruleKey {
			if "FORWARD" == ruleData.Get(ruleKey) {
				entrySet.Actions.Config.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_ACCEPT
				//entrySet.Actions.State.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_ACCEPT
			} else {
				entrySet.Actions.Config.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_DROP
				//entrySet.Actions.State.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_DROP
			}
		}
	}

	if aclType == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 {
		ygot.BuildEmptyTree(entrySet.Ipv4)
		for ruleKey := range ruleData.Field {
			if "IP_PROTOCOL" == ruleKey {
				ipProto, _ := strconv.ParseInt(ruleData.Get(ruleKey), 10, 64)
				ipv4ProElem := getIpProtocol(ipProto, ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4, "config")
				entrySet.Ipv4.Config.Protocol = ipv4ProElem.(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)

				ipv4ProElem = getIpProtocol(ipProto, ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4, "state")
				entrySet.Ipv4.State.Protocol = ipv4ProElem.(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_State_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
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
				ipv6ProElem := getIpProtocol(ipProto, ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6, "config")
				entrySet.Ipv6.Config.Protocol = ipv6ProElem.(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)

				ipv6ProElem = getIpProtocol(ipProto, ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6, "state")
				entrySet.Ipv6.State.Protocol = ipv6ProElem.(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_State_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
			} else if "DSCP" == ruleKey {
				var dscp uint8
				dscpData, _ := strconv.ParseInt(ruleData.Get(ruleKey), 10, 64)
				dscp = uint8(dscpData)
				entrySet.Ipv6.Config.Dscp = &dscp
				entrySet.Ipv6.State.Dscp = &dscp
			}
		}
	} else if aclType == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2 {
		ygot.BuildEmptyTree(entrySet.L2)
		for ruleKey := range ruleData.Field {
			if "ETHER_TYPE" == ruleKey {
				ethType, _ := strconv.ParseInt(ruleData.Get(ruleKey), 10, 64)
				fmt.Println(ethType)
				//entrySet.L2.Config.Ethertype = ""
				//entrySet.Ipv6.State.Protocol = getIpProtocolState(ipProto)
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

func (app *AclApp) convertInternalToOCAclBinding(d *db.DB, aclName string, intfId string, direction string, intfAclSet ygot.GoStruct) {
	if _, ok := app.aclTableMap[aclName]; !ok {
		app.convertDBAclToInternal(d, db.Key{Comp: []string{aclName}})
	}

	if _, ok := app.ruleTableMap[aclName]; !ok {
		app.convertDBAclRulesToInternal(d, aclName, -1, db.Key{})
	}

	for ruleName, _ := range app.ruleTableMap[aclName] {
		ruleData := app.ruleTableMap[aclName][ruleName]
		priority, _ := strconv.ParseInt(ruleData.Get("PRIORITY"), 10, 32)
		convertInternalToOCAclRuleBinding(d, uint32(priority), -1, direction, intfAclSet, nil)
	}
}

func (app *AclApp) getAclBindingInfoForInterfaceData(d *db.DB, intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface, intfId string, direction string) {
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
						entrySet := ingressAclSet.AclEntries.AclEntry[seqId]
						convertInternalToOCAclRuleBinding(d, 0, int64(seqId), direction, nil, entrySet)
					}
				} else {
					ygot.BuildEmptyTree(ingressAclSet)
					ingressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_Config{SetName: &aclName, Type: ingressAclSetKey.Type}
					ingressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet_State{SetName: &aclName, Type: ingressAclSetKey.Type}
					app.convertInternalToOCAclBinding(d, aclKey, intfId, direction, ingressAclSet)
				}
			}
		} else {
			app.findAndGetAclBindingInfoForInterfaceData(d, intfId, direction, intfData)
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
						entrySet := egressAclSet.AclEntries.AclEntry[seqId]
						convertInternalToOCAclRuleBinding(d, 0, int64(seqId), direction, nil, entrySet)
					}
				} else {
					ygot.BuildEmptyTree(egressAclSet)
					egressAclSet.Config = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_Config{SetName: &aclName, Type: egressAclSetKey.Type}
					egressAclSet.State = &ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet_State{SetName: &aclName, Type: egressAclSetKey.Type}
					app.convertInternalToOCAclBinding(d, aclKey, intfId, direction, egressAclSet)
				}
			}
		} else {
			app.findAndGetAclBindingInfoForInterfaceData(d, intfId, direction, intfData)
		}
	} else {
		log.Error("Unknown direction")
	}
}

func (app *AclApp) findAndGetAclBindingInfoForInterfaceData(d *db.DB, intfId string, direction string, intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface) {
	if len(app.aclTableMap) == 0 {
		app.convertDBAclToInternal(d, db.Key{})
	}

	for aclName, aclData := range app.aclTableMap {
		aclIntfs := aclData.GetList("ports")
		aclType := aclData.Get(ACL_TYPE)
		var aclOrigName string
		var aclOrigType ocbinds.E_OpenconfigAcl_ACL_TYPE
		if SONIC_ACL_TYPE_IPV4 == aclType {
			aclOrigName = strings.Replace(aclName, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
			aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4
		} else if SONIC_ACL_TYPE_IPV6 == aclType {
			aclOrigName = strings.Replace(aclName, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
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
					app.convertInternalToOCAclBinding(d, aclName, intfId, direction, ingressAclSet)
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
					app.convertInternalToOCAclBinding(d, aclName, intfId, direction, egressAclSet)
				}
			}
		}
	}
}

func (app *AclApp) isInterfaceBindWithACL(d *db.DB, intfId string) bool {
	var isFound bool = false

	if len(app.aclTableMap) == 0 {
		app.convertDBAclToInternal(d, db.Key{})
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
}

/********************   CREATE related    *******************************/
func (app *AclApp) convertOCAclsToInternal(acl *ocbinds.OpenconfigAcl_Acl) map[string]db.Value {
	var aclInfo map[string]db.Value
	if acl != nil {
		aclInfo = make(map[string]db.Value)
		if acl.AclSets != nil && len(acl.AclSets.AclSet) > 0 {
			for aclSetKey, _ := range acl.AclSets.AclSet {
				aclSet := acl.AclSets.AclSet[aclSetKey]
				aclKey := getAclKeyStrFromOCKey(aclSetKey.Name, aclSetKey.Type)
				m := make(map[string]string)
				aclInfo[aclKey] = db.Value{Field: m}

				if aclSet.Config != nil {
					if aclSet.Config.Type == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 {
						aclInfo[aclKey].Field[ACL_TYPE] = SONIC_ACL_TYPE_IPV4
					} else if aclSet.Config.Type == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6 {
						aclInfo[aclKey].Field[ACL_TYPE] = SONIC_ACL_TYPE_IPV6
					} else if aclSet.Config.Type == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2 {
						aclInfo[aclKey].Field[ACL_TYPE] = SONIC_ACL_TYPE_L2
					}

					if aclSet.Config.Description != nil && len(*aclSet.Config.Description) > 0 {
						aclInfo[aclKey].Field[ACL_DESCRIPTION] = *aclSet.Config.Description
					}
				}
			}
		}
	}

	return aclInfo
}

func (app *AclApp) convertOCAclRulesToInternal(acl *ocbinds.OpenconfigAcl_Acl) map[string]map[string]db.Value {
	var rulesInfo map[string]map[string]db.Value
	if acl != nil {
		rulesInfo = make(map[string]map[string]db.Value)
		if acl.AclSets != nil && len(acl.AclSets.AclSet) > 0 {
			for aclSetKey, _ := range acl.AclSets.AclSet {
				aclSet := acl.AclSets.AclSet[aclSetKey]
				aclKey := getAclKeyStrFromOCKey(aclSetKey.Name, aclSetKey.Type)
				rulesInfo[aclKey] = make(map[string]db.Value)

				if aclSet.AclEntries != nil {
					for seqId, _ := range aclSet.AclEntries.AclEntry {
						entrySet := aclSet.AclEntries.AclEntry[seqId]
						ruleName := "RULE_" + strconv.FormatInt(int64(seqId), 10)
						m := make(map[string]string)
						rulesInfo[aclKey][ruleName] = db.Value{Field: m}
						convertOCAclRuleToInternalAclRule(rulesInfo[aclKey][ruleName], seqId, aclKey, aclSet.Type, entrySet)
					}
				}

				yangPathStr, _ := getYangPathFromUri(app.path)
				if yangPathStr != "/openconfig-acl:acl/acl-sets/acl-set/acl-entries" && yangPathStr != "/openconfig-acl:acl/acl-sets/acl-set/acl-entries/acl-entry" {
					app.createDefaultDenyAclRule(rulesInfo[aclKey])
				}
			}
		}
	}

	return rulesInfo
}

func (app *AclApp) convertOCAclBindingsToInternal(d *db.DB, aclData map[string]db.Value, aclObj *ocbinds.OpenconfigAcl_Acl) (bool, error) {
	var err error
	var ret bool = false

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
						if len(aclData) == 0 {
							aclData[aclName] = db.Value{Field: map[string]string{}}
						}
						aclData[aclName].Field["stage"] = "INGRESS"
						ret = true
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
						if len(aclData) == 0 {
							aclData[aclName] = db.Value{Field: map[string]string{}}
						}
						aclData[aclName].Field["stage"] = "EGRESS"
						ret = true
					}
				}
			}
		}
		for k, _ := range aclInterfacesMap {
			val := aclData[k]
			(&val).SetList("ports", aclInterfacesMap[k])
		}
	}
	return ret, err
}

func (app *AclApp) createDefaultDenyAclRule(rulesInfo map[string]db.Value) {
	m := make(map[string]string)
	rulesInfo["DEFAULT_RULE"] = db.Value{Field: m}
	rulesInfo["DEFAULT_RULE"].Field["PRIORITY"] = strconv.FormatInt(int64(MIN_PRIORITY), 10)
	rulesInfo["DEFAULT_RULE"].Field["PACKET_ACTION"] = "DROP"
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
			break
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_Uint16{}):
			v := (rule.L2.Config.Ethertype).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_L2_Config_Ethertype_Union_Uint16)
			fmt.Fprintf(&b, "0x%0.4x", v.Uint16)
			ruleData.Field["ETHER_TYPE"] = b.String()
			break
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
			break
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8{}):
			v := (rule.Ipv4.Config.Protocol).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8)
			ruleData.Field["IP_PROTOCOL"] = strconv.FormatInt(int64(v.Uint8), 10)
			break
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
			break
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_Uint8{}):
			v := (rule.Ipv6.Config.Protocol).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_Uint8)
			ruleData.Field["IP_PROTOCOL"] = strconv.FormatInt(int64(v.Uint8), 10)
			break
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
			break
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_String{}):
			v := (rule.Transport.Config.SourcePort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_String)
			ruleData.Field["L4_SRC_PORT_RANGE"] = strings.Replace(v.String, "..", "-", 1)
			break
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16{}):
			v := (rule.Transport.Config.SourcePort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16)
			ruleData.Field["L4_SRC_PORT"] = strconv.FormatInt(int64(v.Uint16), 10)
			break
		}
	}

	if rule.Transport.Config.DestinationPort != nil && util.IsTypeStructPtr(reflect.TypeOf(rule.Transport.Config.DestinationPort)) {
		destportType := reflect.TypeOf(rule.Transport.Config.DestinationPort).Elem()
		switch destportType {
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort{}):
			v := (rule.Transport.Config.DestinationPort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort)
			ruleData.Field["L4_DST_PORT"] = v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort.ΛMap()["E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort"][int64(v.E_OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort)].Name
			break
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_String{}):
			v := (rule.Transport.Config.DestinationPort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_String)
			ruleData.Field["L4_DST_PORT_RANGE"] = strings.Replace(v.String, "..", "-", 1)
			break
		case reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16{}):
			v := (rule.Transport.Config.DestinationPort).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16)
			ruleData.Field["L4_DST_PORT"] = strconv.FormatInt(int64(v.Uint16), 10)
			break
		}
	}

	var tcpFlags uint32 = 0x00
	if len(rule.Transport.Config.TcpFlags) > 0 {
		for _, flag := range rule.Transport.Config.TcpFlags {
			fmt.Println("TCP Flag name: " + flag.ΛMap()["E_OpenconfigPacketMatchTypes_TCP_FLAGS"][int64(flag)].Name)
			switch flag {
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_FIN:
				tcpFlags |= 0x01
				break
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_SYN:
				tcpFlags |= 0x02
				break
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_RST:
				tcpFlags |= 0x04
				break
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_PSH:
				tcpFlags |= 0x08
				break
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ACK:
				tcpFlags |= 0x10
				break
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_URG:
				tcpFlags |= 0x20
				break
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_ECE:
				tcpFlags |= 0x40
				break
			case ocbinds.OpenconfigPacketMatchTypes_TCP_FLAGS_TCP_CWR:
				tcpFlags |= 0x80
				break
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
			break
		case ocbinds.OpenconfigAcl_FORWARDING_ACTION_DROP:
			ruleData.Field["PACKET_ACTION"] = "DROP"
			break
		case ocbinds.OpenconfigAcl_FORWARDING_ACTION_REJECT:
			ruleData.Field["PACKET_ACTION"] = "DROP"
			break
		default:
		}
	}
}

func (app *AclApp) setAclDataInConfigDb(d *db.DB, aclData map[string]db.Value, createFlag bool) error {
	var err error
	for key := range aclData {
		existingEntry, err := d.GetEntry(app.aclTs, db.Key{Comp: []string{key}})
		// If Create ACL request comes and ACL already exists, throw error
		if createFlag && existingEntry.IsPopulated() {
			return errors.New("Acl " + key + " already exists")
		}
		if createFlag || (!createFlag && err != nil && !existingEntry.IsPopulated()) {
			err := d.CreateEntry(app.aclTs, db.Key{Comp: []string{key}}, aclData[key])
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			if existingEntry.IsPopulated() {
                if existingEntry.Get(ACL_DESCRIPTION) != aclData[key].Field[ACL_DESCRIPTION] {
                    err := d.ModEntry(app.aclTs, db.Key{Comp: []string{key}}, aclData[key])
                    if err != nil {
                        log.Error(err)
                        return err
                    }
                }
				/*
					//Merge any ACL binds already present. Validate should take care of any checks so its safe to blindly merge here
					if len(existingEntry.Field) > 0  {
						value.Field["ports"] += "," + existingEntry.Field["ports@"]
					}
					fmt.Println(value)
				*/
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
				return errors.New("Rule " + ruleName + " already exists")
			}
			if createFlag || (!createFlag && err != nil && !existingRuleEntry.IsPopulated()) {
				err := d.CreateEntry(app.ruleTs, db.Key{Comp: []string{aclName, ruleName}}, ruleData[aclName][ruleName])
				if err != nil {
					log.Error(err)
					return err
				}
			} else {
				if existingRuleEntry.IsPopulated() && ruleName != "DEFAULT_RULE" {
					err := d.ModEntry(app.ruleTs, db.Key{Comp: []string{aclName, ruleName}}, ruleData[aclName][ruleName])
					if err != nil {
						log.Error(err)
						return err
					}
				}
			}
		}
	}
	return err
}

func (app *AclApp) setAclBindDataInConfigDb(d *db.DB, aclData map[string]db.Value) error {
	var err error
	for aclKey, aclInfo := range aclData {
		// Get ACL info from DB and merge ports from request with ports from DB
		dbAcl, err := d.GetEntry(app.aclTs, db.Key{Comp: []string{aclKey}})
		if err != nil {
			log.Error(err)
			return err
		}
		dbAclIntfs := dbAcl.GetList("ports")
		if len(dbAclIntfs) > 0 {
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

		err = d.SetEntry(app.aclTs, db.Key{Comp: []string{aclKey}}, dbAcl)
		//err = d.ModEntry(app.aclTs, db.Key{Comp: []string{aclKey}}, dbAcl)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return err
}

func getIpProtocol(proto int64, aclType ocbinds.E_OpenconfigAcl_ACL_TYPE, contType string) interface{} {
	foundInMap := false
	var ptype ocbinds.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_UNSET

	for k, v := range IP_PROTOCOL_MAP {
		if proto == int64(v) {
			foundInMap = true
			ptype = k
		}
	}

	switch aclType {
	case ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4:
		if "config" == contType {
			if foundInMap {
				var ipProCfg *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL
				ipProCfg = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
				ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ptype
				return ipProCfg
			} else {
				var ipProCfgUint8 *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8
				ipProCfgUint8 = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8)
				ipProCfgUint8.Uint8 = uint8(proto)
				return ipProCfgUint8
			}
		} else if "state" == contType {
			if foundInMap {
				var ipProSt *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_State_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL
				ipProSt = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_State_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
				ipProSt.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ptype
				return ipProSt
			} else {
				var ipProStUint8 *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_State_Protocol_Union_Uint8
				ipProStUint8 = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_State_Protocol_Union_Uint8)
				ipProStUint8.Uint8 = uint8(proto)
				return ipProStUint8
			}
		}
		break
	case ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6:
		if "config" == contType {
			if foundInMap {
				var ipv6ProCfg *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL
				ipv6ProCfg = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
				ipv6ProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ptype
				return ipv6ProCfg
			} else {
				var ipv6ProCfgUint8 *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_Uint8
				ipv6ProCfgUint8 = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_Config_Protocol_Union_Uint8)
				ipv6ProCfgUint8.Uint8 = uint8(proto)
				return ipv6ProCfgUint8
			}
		} else if "state" == contType {
			if foundInMap {
				var ipv6ProSt *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_State_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL
				ipv6ProSt = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_State_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
				ipv6ProSt.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ptype
				return ipv6ProSt
			} else {
				var ipv6ProStUint8 *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_State_Protocol_Union_Uint8
				ipv6ProStUint8 = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv6_State_Protocol_Union_Uint8)
				ipv6ProStUint8.Uint8 = uint8(proto)
				return ipv6ProStUint8
			}
		}
		break
	}
	return nil
}

func getTransportConfigDestPort(destPort string) ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union {
	portNum, _ := strconv.ParseInt(destPort, 10, 64)
	var destPortCfg *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16
	destPortCfg = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16)
	destPortCfg.Uint16 = uint16(portNum)
	return destPortCfg
}

func getTransportConfigSrcPort(srcPort string) ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union {
	portNum, _ := strconv.ParseInt(srcPort, 10, 64)
	var srcPortCfg *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16
	srcPortCfg = new(ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16)
	srcPortCfg.Uint16 = uint16(portNum)
	return srcPortCfg
}

func getTransportConfigTcpFlags(tcpFlags string) []ocbinds.E_OpenconfigPacketMatchTypes_TCP_FLAGS {
    var flags []ocbinds.E_OpenconfigPacketMatchTypes_TCP_FLAGS
    if len(tcpFlags) > 0 {
        flagStr := strings.Split(tcpFlags, "/")[0]
        flagNumber,_ := strconv.ParseUint(strings.Replace(flagStr, "0x", "", -1), 16, 32)
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

func (app *AclApp) generateDbWatchKeys(d *db.DB, isDeleteOp bool) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	var aclSubtree = false

	aclObj := app.getAppRootObject()
	if !util.IsValueScalar(reflect.ValueOf(*app.ygotTarget)) && util.IsValuePtr(reflect.ValueOf(*app.ygotTarget)) {
		if reflect.TypeOf(*app.ygotTarget).Elem().Name() == "OpenconfigAcl_Acl" {
			aclSubtree = true
		}
	}

	// These slices will store the yangPaths derived from the URI requested to help
	// determining when to create/update ACL or rule or both
	var ruleBasedTargets []string = []string{}
	var aclBasedTargets []string = []string{getYangPathFromYgotStruct(aclObj, OC_ACL_YANG_PATH_PREFIX, OC_ACL_APP_MODULE_NAME), getYangPathFromYgotStruct(aclObj.AclSets, OC_ACL_YANG_PATH_PREFIX, OC_ACL_APP_MODULE_NAME)}

	targetUriPath, err := getYangPathFromUri(app.path)

	if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets") || aclSubtree {
		if aclObj.AclSets != nil && len(aclObj.AclSets.AclSet) > 0 {
			// Build Watch keys for a specific ACL
			for aclSetKey, _ := range aclObj.AclSets.AclSet {
				aclKey := getAclKeyStrFromOCKey(aclSetKey.Name, aclSetKey.Type)
				keys = append(keys, db.WatchKeys{app.aclTs, &(db.Key{Comp: []string{aclKey}})})

				aclSet := aclObj.AclSets.AclSet[aclSetKey]
				aclBasedTargets = append(aclBasedTargets, getYangPathFromYgotStruct(aclSet, OC_ACL_YANG_PATH_PREFIX, OC_ACL_APP_MODULE_NAME))
				ruleBasedTargets = append(ruleBasedTargets, getYangPathFromYgotStruct(aclSet.AclEntries, OC_ACL_YANG_PATH_PREFIX, OC_ACL_APP_MODULE_NAME))

				if aclSet.AclEntries != nil && len(aclSet.AclEntries.AclEntry) > 0 {
					// Build Watch keys for a specific Rule
					for seqId, _ := range aclSet.AclEntries.AclEntry {
						ruleName := "RULE_" + strconv.FormatInt(int64(seqId), 10)
						keys = append(keys, db.WatchKeys{app.ruleTs, &(db.Key{Comp: []string{aclKey, ruleName}})})
						ruleBasedTargets = append(ruleBasedTargets, getYangPathFromYgotStruct(aclSet.AclEntries.AclEntry[seqId], OC_ACL_YANG_PATH_PREFIX, OC_ACL_APP_MODULE_NAME))
					}
				} else {
					// Build watch keys for all rules for a specific ACL
					if isDeleteOp {
						ruleKeys, _ := d.GetKeys(app.ruleTs)
						for i, rulekey := range ruleKeys {
							// Rulekey has two keys, first aclkey and second rulename
							if rulekey.Get(0) == aclKey {
								keys = append(keys, db.WatchKeys{app.ruleTs, &ruleKeys[i]})
							}
						}
					} else {
						for ruleName, _ := range app.ruleTableMap[aclKey] {
							keys = append(keys, db.WatchKeys{app.ruleTs, &db.Key{Comp: []string{aclKey, ruleName}}})
						}
					}
				}
			}
		} else {
			// Building Watch keys for All ACLs and Rules
			if isDeleteOp {
				aclKeys, _ := d.GetKeys(app.aclTs)
				ruleKeys, _ := d.GetKeys(app.ruleTs)

				for i, _ := range aclKeys {
					keys = append(keys, db.WatchKeys{app.aclTs, &aclKeys[i]})
				}
				for i, _ := range ruleKeys {
					keys = append(keys, db.WatchKeys{app.ruleTs, &ruleKeys[i]})
				}
			} else {
				for aclName, _ := range app.aclTableMap {
					keys = append(keys, db.WatchKeys{app.aclTs, &db.Key{Comp: []string{aclName}}})
					for ruleName, _ := range app.ruleTableMap[aclName] {
						keys = append(keys, db.WatchKeys{app.ruleTs, &db.Key{Comp: []string{aclName, ruleName}}})
					}
				}
			}
		}
	}

	if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces") {
		if aclObj.Interfaces != nil && len(aclObj.Interfaces.Interface) > 0 {
			// Request is for specific interface
			var intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface
			for intfId := range aclObj.Interfaces.Interface {
				intfData = aclObj.Interfaces.Interface[intfId]
				if intfData != nil {
					if intfData.IngressAclSets != nil && len(intfData.IngressAclSets.IngressAclSet) > 0 {
						for inAclKey, _ := range intfData.IngressAclSets.IngressAclSet {
							aclName := getAclKeyStrFromOCKey(inAclKey.SetName, inAclKey.Type)
							keys = append(keys, db.WatchKeys{app.aclTs, &db.Key{Comp: []string{aclName}}})
						}
					} else if intfData.EgressAclSets != nil && len(intfData.EgressAclSets.EgressAclSet) > 0 {
						for outAclKey, _ := range intfData.EgressAclSets.EgressAclSet {
							aclName := getAclKeyStrFromOCKey(outAclKey.SetName, outAclKey.Type)
							keys = append(keys, db.WatchKeys{app.aclTs, &db.Key{Comp: []string{aclName}}})
						}
					}
				}
			}
		} else {
			// Request for all interfaces
			if isDeleteOp {
				aclKeys, _ := d.GetKeys(app.aclTs)
				for i, _ := range aclKeys {
					aclEntry, _ := d.GetEntry(app.aclTs, aclKeys[i])
					if len(aclEntry.GetList("ports")) > 0 {
						keys = append(keys, db.WatchKeys{app.aclTs, &aclKeys[i]})
					}
				}
			}
		}
	}

	if contains(aclBasedTargets, targetUriPath) {
		app.createAclFlag = true
		app.createRuleFlag = true
	}
	if contains(ruleBasedTargets, targetUriPath) {
		app.createRuleFlag = true
	}

	log.Infof("Values of createAclFlag: %t and  createRuleFlag: %t", app.createAclFlag, app.createRuleFlag)

	return keys, err
}

func getAclKeysFromStrKey(aclKey string, aclType string) (string, ocbinds.E_OpenconfigAcl_ACL_TYPE) {
	var aclOrigName string
	var aclOrigType ocbinds.E_OpenconfigAcl_ACL_TYPE

	if SONIC_ACL_TYPE_IPV4 == aclType {
		aclOrigName = strings.Replace(aclKey, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
		aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4
	} else if SONIC_ACL_TYPE_IPV6 == aclType {
		aclOrigName = strings.Replace(aclKey, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
		aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6
	} else if SONIC_ACL_TYPE_L2 == aclType {
		aclOrigName = strings.Replace(aclKey, "_"+OPENCONFIG_ACL_TYPE_L2, "", 1)
		aclOrigType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2
	}
	return aclOrigName, aclOrigType
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
