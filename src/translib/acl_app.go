package translib

import (
    "fmt"
    "bytes"
    "reflect"
    "strconv"
    "strings"
    "errors"
    "translib/db"
    "translib/ocbinds"
    "github.com/openconfig/ygot/ygot"
    "github.com/openconfig/ygot/util"
    log "github.com/golang/glog"
    "github.com/openconfig/gnmi/proto/gnmi"
)

const (
    TABLE_SEPARATOR  = "|"
    KEY_SEPARATOR = "|"
    ACL_TABLE = "ACL_TABLE"
    RULE_TABLE = "ACL_RULE"
    ACL_TYPE = "type"
    ACL_DESCRIPTION = "policy_desc"
    SONIC_ACL_TYPE_IPV4 = "L3"
    SONIC_ACL_TYPE_IPV6 = "L3V6"
    OPENCONFIG_ACL_TYPE_IPV4 = "ACL_IPV4"
    OPENCONFIG_ACL_TYPE_IPV6 = "ACL_IPV6"
    OPENCONFIG_ACL_TYPE_L2 = "ACL_L2"

    MIN_PRIORITY = 1
    MAX_PRIORITY = 65535
)

var OLD_IP_PROTOCOL_MAP = map[string]uint8 {
    "IP_ICMP": 1,
    "IP_IGMP": 2,
    "IP_TCP":  6,
    "IP_UDP":  17,
    "IP_RSVP": 46,
    "IP_GRE":  47,
    "IP_AUTH": 51,
    "IP_PIM":  103,
    "IP_L2TP": 115,
}

var IP_PROTOCOL_MAP = map[ocbinds.E_OpenconfigPacketMatchTypes_IP_PROTOCOL]uint8 {
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_ICMP:   1,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_IGMP:   2,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_TCP:    6,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_UDP:   17,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_RSVP:  46,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_GRE:   47,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_AUTH:  51,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_PIM:  103,
    ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_L2TP: 115,
}

var ETHERTYPE_MAP = map[ocbinds.E_OpenconfigPacketMatchTypes_ETHERTYPE]uint32 {
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_LLDP: 0x88CC,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_VLAN: 0x8100,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_ROCE: 0x8915,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_ARP:  0x0806,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_IPV4: 0x0800,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_IPV6: 0x86DD,
    ocbinds.OpenconfigPacketMatchTypes_ETHERTYPE_ETHERTYPE_MPLS: 0x8847,
}

var aclTs db.TableSpec
var ruleTs db.TableSpec
var xlated_data map[string]map[string]map[string]map[string]string
var acl_table_map map[string]map[string]string
var acl_rule_map map[string]map[string]map[string]string


type AclApp struct {
	path       string
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}

	aclTableMap map[string]db.Value
    ruleTableMap map[string]map[string]db.Value
}

func init() {
	log.Info("Init called for ACL module")
	err := register("/openconfig-acl:acl",
		&appInfo{appType: reflect.TypeOf(AclApp{}),
			ygotRootType: reflect.TypeOf(ocbinds.OpenconfigAcl_Acl{}),
			isNative:     false})
	if err != nil {
		log.Fatal("Register ACL app module with App Interface failed with error=", err)
	}

	err = addModel(&ModelData{Name: "openconfig-acl",
		Org: "OpenConfig working group",
		Ver:      "1.0.2"})
	if err != nil {
		log.Fatal("Adding model data to appinterface failed with error=", err)
	}
}

func (acl *AclApp) initialize(data appData) {
	log.Info("initialize:acl:path =", data.path)
    *acl = AclApp{path: data.path, ygotRoot: data.ygotRoot, ygotTarget: data.ygotTarget}

    xlated_data = make(map[string]map[string]map[string]map[string]string)
    acl_table_map = make(map[string]map[string]string)
    xlated_data[ACL_TABLE] = map[string]map[string]map[string]string{}
    xlated_data[ACL_TABLE][ACL_TABLE] = acl_table_map
    acl_rule_map = make(map[string]map[string]map[string]string)
    xlated_data[RULE_TABLE] = acl_rule_map

    aclTs = db.TableSpec {Name: ACL_TABLE}
    ruleTs = db.TableSpec {Name: RULE_TABLE}
}

func (acl *AclApp) translateCreate(d *db.DB) ([]db.WatchKeys, error)  {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCreate:acl:path =", acl.path)

    aclObj := (*acl.ygotRoot).(*ocbinds.OpenconfigAcl_Acl)
    acl.aclTableMap = convert_oc_acls_to_internal(aclObj)
    acl.ruleTableMap = convert_oc_acl_rules_to_internal(aclObj)
    convert_oc_acl_bindings_to_internal(acl.aclTableMap, aclObj)

    for aclName,_ := range acl.aclTableMap {
        keys = append(keys, db.WatchKeys{ &aclTs, &db.Key{Comp:[]string{aclName}} })
        for ruleName,_ := range acl.ruleTableMap[aclName] {
            keys = append(keys, db.WatchKeys{ &ruleTs, &db.Key{Comp:[]string{aclName, ruleName}} })
        }
    }

	//err = errors.New("Not implemented")
	return keys, err
}

func (acl *AclApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error)  {
	var err error
    var keys []db.WatchKeys
    log.Info("translateUpdate:acl:path =", acl.path)
    err = errors.New("Not implemented")
    return keys, err
}

func (acl *AclApp) translateReplace(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys
    log.Info("translateReplace:acl:path =", acl.path)
    err = errors.New("Not implemented")
    return keys, err
}

func (acl *AclApp) translateDelete(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys
    log.Info("translateDelete:acl:path =", acl.path)

    fmt.Println("translateDelete: Target Type: " + reflect.TypeOf(*acl.ygotTarget).Elem().Name())

    acl_subtree := false
    //intf_subtree := false
    if *acl.ygotRoot == *acl.ygotTarget {
        acl_subtree = true
        //intf_subtree = true
    }
    targetUriPath, _ := getYangPathFromUri(acl.path)
    if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets") || acl_subtree {
        if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets/acl-set") {
            fmt.Println("Building Watch keys for Specific ACL")
        } else {
            fmt.Println("Building Watch keys for Delete Request for All ACLs")
            aclKeys,_ := d.GetKeys(&aclTs)
            ruleKeys,_ := d.GetKeys(&ruleTs)

            for _,aclkey := range aclKeys {
                keys = append(keys, db.WatchKeys{ &aclTs, &aclkey})
            }
            for _,rulekey := range ruleKeys {
                keys = append(keys, db.WatchKeys{ &ruleTs, &rulekey})
            }
        }
    }

	//err = errors.New("Not implemented")
	return keys, err
}

func (acl *AclApp) translateGet(dbs [db.MaxDB]*db.DB) error  {
	var err error
	log.Info("translateGet:acl:path =", acl.path)
    return err
}

func (acl *AclApp) processCreate(d *db.DB) (SetResponse, error)  {
	var resp SetResponse
	// below are some sample code for CREATE
	//aclObj := (*acl.ygotRoot).(*ocbinds.OpenconfigAcl_Acl)

    /*
	if *acl.ygotRoot == *acl.ygotTarget {
		fmt.Println("CREATE ACL request")
	} else if aclObj.AclSets == *acl.ygotTarget {
		fmt.Println("CREATE ACL SETS request")
		// insert the data in the config db present under the aclSets since aclSets is the target object
		fmt.Println("aclObj.AclSets.AclSet => ", aclObj.AclSets.AclSet)
	}
    */

    fmt.Println("ProcessCreate: Target Type is " + reflect.TypeOf(*acl.ygotTarget).Elem().Name())

	var err error
	log.Info("processCreate:acl:path =", acl.path)

    set_acl_data_in_config_db(d, acl.aclTableMap)
    set_acl_rule_data_in_config_db(d, acl.ruleTableMap)

	//err = errors.New("Not implemented")
	return resp, err
}

func (acl *AclApp) processUpdate(d *db.DB) (SetResponse, error)  {
	var err error
	var resp SetResponse
	log.Info("processUpdate:acl:path =", acl.path)
    err = errors.New("Not implemented")
    return resp, err
}

func (acl *AclApp) processReplace(d *db.DB) (SetResponse, error)  {
    var err error
	var resp SetResponse
    log.Info("processReplace:acl:path =", acl.path)
    err = errors.New("Not implemented")
    return resp, err
}

func (acl *AclApp) processDelete(d *db.DB) (SetResponse, error)  {
    var err error
	var resp SetResponse
    log.Info("processDelete:acl:path =", acl.path)

    acl_subtree := false
    //intf_subtree := false
    if *acl.ygotRoot == *acl.ygotTarget {
        acl_subtree = true
        //intf_subtree = true
    }
    targetUriPath, _ := getYangPathFromUri(acl.path)
    if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets") || acl_subtree {
        if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets/acl-set") {
            fmt.Println("Request is for specific ACL(s)")

            if reflect.ValueOf(*acl.ygotTarget).Kind() == reflect.Ptr {
                element := reflect.ValueOf(*acl.ygotTarget).Elem()
                fmt.Println("element type: " + element.Type().String())
                if "OpenconfigAcl_Acl_AclSets_AclSet" == element.Type().Name() {
                    aclSet := (*acl.ygotTarget).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet)
                    aclName := *aclSet.Name
                    var aclType string
                    switch aclSet.Type {
                    case ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4:
                        aclType = OPENCONFIG_ACL_TYPE_IPV4
                        break
                    case ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6:
                        aclType = OPENCONFIG_ACL_TYPE_IPV6
                        break
                    case ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2:
                        aclType = OPENCONFIG_ACL_TYPE_L2
                        break
                    /*case ocbinds.OpenconfigAcl_ACL_TYPE_ACL_MIXED:
                        break*/
                    }
                    aclKey := aclName + "_" + aclType

                    if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets/acl-set/acl-entries/acl-entry") {
                        fmt.Println("Request is for specific Rule(s)")
                        if "OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry" == element.Type().Name() {
                            entrySet := (*acl.ygotTarget).(*ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry)
                            ruleKey := aclKey + TABLE_SEPARATOR + "RULE_" + strconv.FormatInt(int64(*entrySet.SequenceId), 10)
                            fmt.Printf("delete request is for ACL: %s and Rule: %s\n", aclKey, ruleKey)
                            //d.DeleteEntry(&ruleTs, d.Key{Comp: []string {ruleKey} })
                        }
                    } else {
                        fmt.Println("Delete request is for acl key: " + aclKey)
                        //d.DeleteEntry(&aclTs, d.Key{Comp: []string {aclKey} })
                        //d.DeleteKeys(&ruleTs, aclKey + TABLE_SEPARATOR + "*")
                    }
                }
            }
        } else {
            fmt.Println("Request is for All ACLs")
            d.DeleteTable(&aclTs)
            d.DeleteTable(&ruleTs)
        }
    }

    //err = errors.New("Not implemented")
    return resp, err
}

func (acl *AclApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error)  {
    var err error
    var payload []byte
    var aclSubtree bool = false
    //var intfSubtree bool = false

    configDb := dbs[db.ConfigDB]
    aclObj := (*acl.ygotRoot).(*ocbinds.OpenconfigAcl_Acl)

    if *acl.ygotRoot == *acl.ygotTarget {
        aclSubtree = true
        //intfSubtree = true
    }

    fmt.Println("processGet: Target Type: " + reflect.TypeOf(*acl.ygotTarget).Elem().Name())

    targetUriPath, err := getYangPathFromUri(acl.path)
    if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/acl-sets") || aclSubtree {
        if aclObj.AclSets != nil && len(aclObj.AclSets.AclSet) > 0 {
            // Request for specific ACL
            for aclSetKey,_ := range aclObj.AclSets.AclSet {
                aclName := strings.ReplaceAll(strings.ReplaceAll(aclSetKey.Name, " ", "_"), "-", "_")
                aclType := aclSetKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(aclSetKey.Type)].Name
                aclSet := aclObj.AclSets.AclSet[aclSetKey]
                aclKey := aclName + "_" + aclType

                if aclSet.AclEntries != nil && len(aclSet.AclEntries.AclEntry) > 0 {
                    // Request for specific Rule
                    for seqId,_ := range aclSet.AclEntries.AclEntry {
                        //ruleKey := "RULE_" + strconv.FormatInt(int64(seqId), 10)
                        entrySet := aclSet.AclEntries.AclEntry[seqId]
                        err = convert_db_acl_rules_to_internal(configDb, aclKey, int64(seqId), db.Key{})
                        if (err != nil) {
                            return GetResponse{Payload:payload, ErrSrc:AppErr}, err
                        }
                        ygot.BuildEmptyTree(entrySet)
                        convert_internal_to_oc_acl_rule(aclKey, aclSetKey.Type, int64(seqId), nil, entrySet)

                        var jsonStr string
                        if *acl.ygotTarget == entrySet {
                            jsonStr, err = dumpIetfJson(aclSet.AclEntries)
                        } else {
                            dummyEntrySet := &ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry{}
                            if *acl.ygotTarget == entrySet.Config {
                                dummyEntrySet.Config = entrySet.Config
                                jsonStr, err = dumpIetfJson(dummyEntrySet)
                            } else if *acl.ygotTarget == entrySet.State {
                                dummyEntrySet.State = entrySet.State
                                jsonStr, err = dumpIetfJson(dummyEntrySet)
                            } else if *acl.ygotTarget == entrySet.Actions {
                                dummyEntrySet.Actions = entrySet.Actions
                                jsonStr, err = dumpIetfJson(dummyEntrySet)
                            } else if *acl.ygotTarget == entrySet.InputInterface {
                                dummyEntrySet.InputInterface = entrySet.InputInterface
                                jsonStr, err = dumpIetfJson(dummyEntrySet)
                            } else if *acl.ygotTarget == entrySet.Ipv4 {
                                dummyEntrySet.Ipv4 = entrySet.Ipv4
                                jsonStr, err = dumpIetfJson(dummyEntrySet)
                            } else if *acl.ygotTarget == entrySet.Ipv6 {
                                dummyEntrySet.Ipv6 = entrySet.Ipv6
                                jsonStr, err = dumpIetfJson(dummyEntrySet)
                            } else if *acl.ygotTarget == entrySet.L2 {
                                dummyEntrySet.L2 = entrySet.L2
                                jsonStr, err = dumpIetfJson(dummyEntrySet)
                            } else if *acl.ygotTarget == entrySet.Transport {
                                dummyEntrySet.Transport = entrySet.Transport
                                jsonStr, err = dumpIetfJson(dummyEntrySet)
                            } else {
                            }
                        }
                        payload = []byte(jsonStr)
                    }
                } else {
                    err = convert_db_acl_to_internal(configDb, db.Key{Comp: []string {aclKey} })
                    if (err != nil) {
                        return GetResponse{Payload:payload, ErrSrc:AppErr}, err
                    }

                    ygot.BuildEmptyTree(aclSet)
                    convert_internal_to_oc_acl(aclKey, aclObj.AclSets, aclSet)

                    var jsonStr string
                    if *acl.ygotTarget == aclSet {
                        jsonStr, err = dumpIetfJson(aclObj.AclSets)
                    } else {
                        dummyAclSet := &ocbinds.OpenconfigAcl_Acl_AclSets_AclSet{}
                        if *acl.ygotTarget == aclSet.Config {
                            dummyAclSet.Config = aclSet.Config
                            jsonStr, err = dumpIetfJson(dummyAclSet)
                        } else if *acl.ygotTarget == aclSet.State {
                            dummyAclSet.State = aclSet.State
                            jsonStr, err = dumpIetfJson(dummyAclSet)
                        } else if *acl.ygotTarget == aclSet.AclEntries {
                            dummyAclSet.AclEntries = aclSet.AclEntries
                            jsonStr, err = dumpIetfJson(dummyAclSet)
                        } else {
                            if targetUriPath == "/openconfig-acl:acl/acl-sets/acl-set/acl-entries/acl-entry" {
                                dummyAclSet.AclEntries = aclSet.AclEntries
                                jsonStr, err = dumpIetfJson(dummyAclSet.AclEntries)
                            } else if targetUriPath == "/openconfig-acl:acl/acl-sets/acl-set/config/description" {
                                dummyAclSet.Config = &ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_Config{}
                                dummyAclSet.Config.Description = aclSet.Config.Description
                                jsonStr, err = dumpIetfJson(dummyAclSet.Config)
                            }
                        }
                    }

                    payload = []byte(jsonStr)
                }
            }
        } else {
            // Request for all ACLs
            ygot.BuildEmptyTree(aclObj)
            err = convert_db_acl_to_internal(configDb, db.Key{})
            if (err != nil) {
                return GetResponse{Payload:payload, ErrSrc:AppErr}, err
            }

            convert_internal_to_oc_acl("", aclObj.AclSets, nil)
            if (err != nil) {
                return GetResponse{Payload:payload, ErrSrc:AppErr}, err
            }

            var jsonStr string
            if *acl.ygotTarget == *acl.ygotRoot {
                jsonStr, err = dumpIetfJson(&ocbinds.Device{Acl:aclObj})
            } else if *acl.ygotTarget == aclObj.AclSets {
                jsonStr, err = dumpIetfJson(aclObj)
            } else {
                jsonStr, err = dumpIetfJson(aclObj.AclSets)
            }
            payload = []byte(jsonStr)
        }
    }

    /*
    if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces") || intfSubtree {
        if aclObj.Interfaces != nil && len(aclObj.Interfaces.Interface) > 0 {
            for intfId,_ := range aclObj.Interfaces.Interface {
                intfData := aclObj.Interfaces.Interface[intfId]
                if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces/interface/ingress-acl-sets") {
                    // Ingress ACL Specific
                    get_acl_binding_info_for_subtree(configDb, intfData, intfId, "INGRESS")
                } else if isSubtreeRequest(targetUriPath, "/openconfig-acl:acl/interfaces/interface/egress-acl-sets") {
                    // Egress ACL Specific
                    get_acl_binding_info_for_subtree(configDb, intfData, intfId, "EGRESS")
                } else {
                    // Direction unknown. Check ACL Table for binding information.
                    fmt.Println("Request is for specific interface, ingress and egress ACLs")
                    get_acl_binding_info_for_subtree(configDb, intfData, intfId, "INGRESS")
                    get_acl_binding_info_for_subtree(configDb, intfData, intfId, "EGRESS")
                }
            }
        } else {
            fmt.Println("Request is for all interfaces and all directions on which ACL is applied")
            if len(xlated_data[ACL_TABLE][ACL_TABLE]) == 0 {
                // Get all ACLs
                tbl,_ := configDb.GetTable(&aclTs)
                keys, _ := tbl.GetKeys()
                for _, key := range keys {
                    convert_db_acl_to_internal(configDb, key)
                }

                for aclName := range xlated_data[ACL_TABLE][ACL_TABLE] {
                    aclData := xlated_data[ACL_TABLE][ACL_TABLE][aclName]
                    if len(aclData["ports@"]) > 0 {
                    }
                }
            }
        }
    }
    */

	return GetResponse{Payload:payload}, err
}


/***********    These are Translation Helper Function   ***********/
func convert_db_acl_rules_to_internal(dbCl *db.DB, aclName string , seqId int64, ruleKey db.Key) error {
    var err error
    if seqId != -1 {
        ruleKey.Comp = []string {aclName, "RULE_" + strconv.FormatInt(int64(seqId), 10)}
    }
    if len(ruleKey.Comp) > 1 {
        ruleName := ruleKey.Comp[1]
        if ruleName != "DEFAULT_RULE" {
            ruleData, err := dbCl.GetEntry(&ruleTs, ruleKey)
            if err != nil {
                return err
            }
            if acl_rule_map[aclName] == nil {
                acl_rule_map[aclName] = map[string]map[string]string{}
            }
            acl_rule_map[aclName][ruleName] = ruleData.Field
        }
    } else {
        ruleKeys, err := dbCl.GetKeys(&ruleTs)
        if err != nil {
            return err
        }
        for _, rkey := range ruleKeys {
            convert_db_acl_rules_to_internal(dbCl, aclName, -1, rkey)
        }
    }
    return err
}

func convert_db_acl_to_internal(dbCl *db.DB, acl db.Key) error {
    var err error
    if len(acl.Comp) > 0 {
        // Get one particular ACL
        entry, err := dbCl.GetEntry(&aclTs, acl)
        if err != nil {
            return err
        }
        if len(entry.Field) > 0 {
            acl_table_map[acl.Comp[0]] = entry.Field
            acl_rule_map[acl.Comp[0]] = map[string]map[string]string{}
            err = convert_db_acl_rules_to_internal(dbCl, acl.Comp[0], -1, db.Key{})
            if err != nil {
                return err
            }
        } else {
            return errors.New("ACL is not configured")
        }
    } else {
        // Get all ACLs
        tbl,err := dbCl.GetTable(&aclTs)
        if err != nil {
            return err
        }
        keys, _ := tbl.GetKeys()
        for _, key := range keys {
            convert_db_acl_to_internal(dbCl, key)
        }
    }
    return err
}

func convert_internal_to_oc_acl(aclName string, aclSets *ocbinds.OpenconfigAcl_Acl_AclSets, aclSet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet) {
    if len(aclName) > 0 {
        aclData := xlated_data[ACL_TABLE][ACL_TABLE][aclName]
        if aclSet != nil {
            aclSet.Config.Name = aclSet.Name
            aclSet.Config.Type = aclSet.Type
            aclSet.State.Name = aclSet.Name
            aclSet.State.Type = aclSet.Type

            for k := range aclData {
                if ACL_DESCRIPTION == k {
                    descr := aclData[k]
                    aclSet.Config.Description = &descr
                    aclSet.State.Description = &descr
                } else if ACL_TYPE == k {
                } else if "ports@" == k {
                    continue
                    //convert_db_to_oc_acl_bindings
                }
            }

            convert_internal_to_oc_acl_rule(aclName, aclSet.Type, -1, aclSet, nil)
        }
    } else {
        aclTable := xlated_data[ACL_TABLE][ACL_TABLE]
        for acln := range aclTable {
            acldata := xlated_data[ACL_TABLE][ACL_TABLE][acln]
            var aclNameStr string
            var aclType ocbinds.E_OpenconfigAcl_ACL_TYPE
            if acldata[ACL_TYPE] == SONIC_ACL_TYPE_IPV4 {
                aclNameStr = strings.Replace(acln, "_"+OPENCONFIG_ACL_TYPE_IPV4, "", 1)
                aclType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4
            } else if acldata[ACL_TYPE] == SONIC_ACL_TYPE_IPV6 {
                aclNameStr = strings.Replace(acln, "_"+OPENCONFIG_ACL_TYPE_IPV6, "", 1)
                aclType = ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6
            }
            aclSetPtr, aclErr := aclSets.NewAclSet(aclNameStr, aclType) ; if (aclErr != nil) {
                fmt.Println("Error handling: ", aclErr)
            }
            ygot.BuildEmptyTree(aclSetPtr)
            convert_internal_to_oc_acl(acln, nil, aclSetPtr)
        }
    }
}

func convert_internal_to_oc_acl_rule(aclName string, aclType ocbinds.E_OpenconfigAcl_ACL_TYPE, seqId int64, aclSet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet, entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
    if seqId != -1 {
        ruleName := "RULE_" + strconv.FormatInt(int64(seqId), 10)
        if aclType == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 {
            convert_internal_to_oc_acl_rule_ipv4(acl_rule_map[aclName][ruleName], nil, entrySet)
        } else if aclType == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6 {
            convert_internal_to_oc_acl_rule_ipv6(acl_rule_map[aclName][ruleName], nil, entrySet)
        }
    } else {
        for ruleName := range xlated_data[RULE_TABLE][aclName] {
            if aclType == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 {
                convert_internal_to_oc_acl_rule_ipv4(acl_rule_map[aclName][ruleName], aclSet, nil)
            } else if aclType == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6 {
                convert_internal_to_oc_acl_rule_ipv6(acl_rule_map[aclName][ruleName], aclSet, nil)
            }
        }
    }
}

func convert_internal_to_oc_acl_rule_ipv4(ruleData map[string]string, aclSet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet, entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
    priority,_ := strconv.ParseInt(ruleData["PRIORITY"], 10, 32)
    seqId := uint32(MAX_PRIORITY - priority)
    ruleDescr := ruleData["RULE_DESCRIPTION"]

    if entrySet == nil {
        if aclSet != nil {
            entrySet_, _ := aclSet.AclEntries.NewAclEntry(seqId)
            entrySet = entrySet_
            ygot.BuildEmptyTree(entrySet)
        }
    }

    entrySet.Config.SequenceId = &seqId
    entrySet.Config.Description = &ruleDescr
    entrySet.State.SequenceId = &seqId
    entrySet.State.Description = &ruleDescr

    var num uint64
    num = 0
    entrySet.State.MatchedOctets = &num
    entrySet.State.MatchedPackets = &num

    ygot.BuildEmptyTree(entrySet.Ipv4)
    ygot.BuildEmptyTree(entrySet.Transport)
    ygot.BuildEmptyTree(entrySet.Actions)
    for ruleKey := range ruleData {
        if "IP_PROTOCOL" == ruleKey {
            ipProto, _ := strconv.ParseInt(ruleData[ruleKey], 10, 64)
            entrySet.Ipv4.Config.Protocol = getIpProtocolConfig(ipProto)
            //entrySet.Ipv4.State.Protocol = getIpProtocolState(ipProto)
        } else if "DSCP" == ruleKey {
            var dscp uint8
            dscpData, _ := strconv.ParseInt(ruleData[ruleKey], 10, 64)
            dscp = uint8(dscpData)
            entrySet.Ipv4.Config.Dscp = &dscp
            entrySet.Ipv4.State.Dscp = &dscp
        } else if "SRC_IP" == ruleKey {
            addr := ruleData[ruleKey]
            entrySet.Ipv4.Config.SourceAddress = &addr
            entrySet.Ipv4.State.SourceAddress = &addr
        } else if "DST_IP" == ruleKey {
            addr := ruleData[ruleKey]
            entrySet.Ipv4.Config.DestinationAddress = &addr
            entrySet.Ipv4.State.DestinationAddress = &addr
        } else if "L4_SRC_PORT" == ruleKey {
            port := ruleData[ruleKey]
            entrySet.Transport.Config.SourcePort = getTransportConfigSrcPort(port)
            //entrySet.Transport.State.SourcePort = &addr
        } else if "L4_DST_PORT" == ruleKey {
            port := ruleData[ruleKey]
            entrySet.Transport.Config.DestinationPort = getTransportConfigDestPort(port)
            //entrySet.Transport.State.DestinationPort = &addr
        } else if "PACKET_ACTION" == ruleKey {
            if "FORWARD" == ruleData[ruleKey] {
                entrySet.Actions.Config.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_ACCEPT
                //entrySet.Actions.State.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_ACCEPT
            } else {
                entrySet.Actions.Config.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_DROP
                //entrySet.Actions.State.ForwardingAction = ocbinds.OpenconfigAcl_FORWARDING_ACTION_DROP
            }
        }
    }
}

func convert_internal_to_oc_acl_rule_ipv6(ruleData map[string]string, aclSet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet, entrySet *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
}

/*
func convert_internal_to_oc_acl_binding (d *db.DB, aclName string, intfId string, direction string, intfAclSet *ygot.GoStruct) {
    convert_db_acl_to_internal(d, db.Key{Comp: []string {aclName} })

    convert_db_acl_rules_to_internal(d, aclName, -1, db.Key{})

    if reflect.TypeOf(*intfAclSet).Elem() == reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet{}) {
        fmt.Println("Got Egress Acl set subtree")
    } else if reflect.TypeOf(*intfAclSet).Elem() == reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet{}) {
        fmt.Println("Got Ingress Acl set subtree")
    }

    for ruleName,_ := range xlated_data[RULE_TABLE][aclName] {
        ruleData := xlated_data[RULE_TABLE][aclName][ruleName]
        convert_internal_to_oc_acl_rule_binding(d, aclName, intfId, priority, intfAclSet)
    }
}

func convert_internal_to_oc_acl_rule_binding(d *db.DB, aclName string, intfId string, priority uint32, seqId int64, aclSet *ygot.GoStruct, entrySet *ygotGoStruct) {
    if seqId == -1 {
        seqId = MAX_PRIORITY - priority
    }
    if entrySet == nil {
        entrySet = aclSet.AclEntries.AclEntry[seqId]
    }
    entrySet.state.SequenceId(seqId)
    entrySet.state.MatchedPackets = 0
    entrySet.state.MatchedOctets = 0
}

func get_acl_binding_info_for_subtree(intfData *ocbinds.OpenconfigAcl_Acl_Interfaces_Interface, intfId string, direction string) {
    if direction == "INGRESS" {
        if intfData.IngressAclSets != nil && len(intfData.IngressAclSets.IngressAclSet) > 0 {
            for ingressAclSetKey,_ := range intfData.IngressAclSets.IngressAclSet {
                aclName := strings.ReplaceAll(strings.ReplaceAll(ingressAclSetKey.SetName, " ", "_"), "-", "_")
                aclType := aclSetKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(ingressAclSetKey.Type)].Name
                aclKey := aclName + "_" + aclType

                ingressAclSet := intfData.IngressAclSets.IngressAclSet[ingressAclSetKey]
                if ingressAclSet != nil && ingressAclSet.AclEntries != nil && len(ingressAclSet.AclEntries.AclEntry) > 0 {
                    for seqId,_ := range ingressAclSet.AclEntries.AclEntry {
                        entrySet := ingressAclSet.AclEntries.AclEntry[seqId]
                        convert_internal_to_oc_acl_rule_binding()
                    }
                } else {
                    convert_internal_to_oc_acl_binding(aclKey, intfId, direction, ingressAclSet)
                }
            }
        } else {
            find_and_get_acl_binding_info_for(intfId, direction, ingressAclSet)
        }
    } else if direction == "EGRESS" {
        if intfData.EgressAclSets != nil && len(intfData.EgressAclSets.EgressAclSet) > 0 {
            for egressAclSetKey,_ := range intfData.EgressAclSets.EgressAclSet {
                aclName := strings.ReplaceAll(strings.ReplaceAll(egressAclSetKey.SetName, " ", "_"), "-", "_")
                aclType := aclSetKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(egressAclSetKey.Type)].Name
                aclKey := aclName + "_" + aclType

                egressAclSet := intfData.EgressAclSets.EgressAclSet[egressAclSetKey]
                if egressAclSet != nil && egressAclSet.AclEntries != nil && len(egressAclSet.AclEntries.AclEntry) > 0 {
                    for seqId,_ := range egressAclSet.AclEntries.AclEntry {
                        entrySet := egressAclSet.AclEntries.AclEntry[seqId]
                        convert_internal_to_oc_acl_rule_binding()
                    }
                } else {
                    convert_internal_to_oc_acl_binding(aclKey, intfId, direction, egressAclSet)
                }
            }
        } else {
            find_and_get_acl_binding_info_for(intfId, direction, egressAclSet)
        }
    } else {
        fmt.Println("Unknown direction")
    }
}

func find_and_get_acl_binding_info_for(intfId string, direction string, aclSet *ygot.GoStruct) {
    if reflect.TypeOf(*aclSet).Elem() == reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_EgressAclSets_EgressAclSet{}) {
        fmt.Println("Got Egress Acl set subtree")
    } else if reflect.TypeOf(*aclSet).Elem() == reflect.TypeOf(ocbinds.OpenconfigAcl_Acl_Interfaces_Interface_IngressAclSets_IngressAclSet{}) {
        fmt.Println("Got Ingress Acl set subtree")
    }
}
*/


/********************   CREATE related    *******************************/
func convert_oc_acls_to_internal(acl *ocbinds.OpenconfigAcl_Acl) map[string]db.Value {
    var aclInfo map[string]db.Value
    if acl != nil {
        aclInfo = make(map[string]db.Value)
        for aclSetKey,_ := range acl.AclSets.AclSet {
            aclSet := acl.AclSets.AclSet[aclSetKey]
            aclName := aclSetKey.Name + "_" + aclSetKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(aclSetKey.Type)].Name
            m := make(map[string]string)
            aclInfo[aclName] = db.Value{Field: m}

            if aclSet.Config.Type == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 {
                aclInfo[aclName].Field[ACL_TYPE] = SONIC_ACL_TYPE_IPV4
            } else if aclSet.Config.Type == ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV6 {
                aclInfo[aclName].Field[ACL_TYPE] = SONIC_ACL_TYPE_IPV6
            }

            if len(*aclSet.Config.Description) > 0 {
                aclInfo[aclName].Field[ACL_DESCRIPTION] = *aclSet.Config.Description
            }
        }
    }

    return aclInfo
}

func convert_oc_acl_rules_to_internal(acl *ocbinds.OpenconfigAcl_Acl) map[string]map[string]db.Value {
    var rulesInfo map[string]map[string]db.Value
    if acl != nil {
        rulesInfo = make(map[string]map[string]db.Value)
        for aclSetKey,_ := range acl.AclSets.AclSet {
            aclSet := acl.AclSets.AclSet[aclSetKey]
            aclName := aclSetKey.Name + "_" + aclSetKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(aclSetKey.Type)].Name
            rulesInfo[aclName] = make(map[string]db.Value)

            if aclSet.AclEntries != nil {
                for seqId,_ := range aclSet.AclEntries.AclEntry {
                    entrySet := aclSet.AclEntries.AclEntry[seqId]
                    ruleName := "RULE_" + strconv.FormatInt(int64(seqId), 10)
                    m := make(map[string]string)
                    rulesInfo[aclName][ruleName] = db.Value{ Field:m }
                    convert_oc_to_internal_rule(rulesInfo[aclName][ruleName], aclName, aclSet.Type, entrySet)
                }
            }

            default_deny_rule(rulesInfo[aclName])
        }
    }

    return rulesInfo
}

func convert_oc_acl_bindings_to_internal(aclData map[string]db.Value, acl *ocbinds.OpenconfigAcl_Acl) bool {
    var ret bool = false
    if acl.Interfaces != nil && len(acl.Interfaces.Interface) > 0 {
        for intfId,_ := range acl.Interfaces.Interface {
            intf := acl.Interfaces.Interface[intfId]
            if intf != nil {
                fmt.Println("Interface Name: " + *intf.Id)
                if intf.IngressAclSets != nil && len(intf.IngressAclSets.IngressAclSet) > 0 {
                    for inAclKey,_ := range intf.IngressAclSets.IngressAclSet {
                        //ingressAclSet := intf.IngressAclSets.IngressAclSet[inAclKey]
                        aclName := inAclKey.SetName + "_" + inAclKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(inAclKey.Type)].Name
                        // TODO: Need to handle Subinterface also
                        if intf.InterfaceRef != nil && intf.InterfaceRef.Config.Interface != nil {
                            aclData[aclName].Field["ports@"] = *intf.InterfaceRef.Config.Interface
                        }
                        aclData[aclName].Field["stage"] = "INGRESS"
                        ret = true
                    }
                }

                if intf.EgressAclSets != nil && len(intf.EgressAclSets.EgressAclSet) > 0 {
                    for outAclKey,_ := range intf.EgressAclSets.EgressAclSet {
                        //egressAclSet := intf.EgressAclSets.EgressAclSet[outAclKey]
                        //aclName := strings.ReplaceAll(strings.ReplaceAll(*egressAclSet.SetName, " ", "_"), "-", "_")
                        aclName := outAclKey.SetName + "_" + outAclKey.Type.ΛMap()["E_OpenconfigAcl_ACL_TYPE"][int64(outAclKey.Type)].Name
                        if intf.InterfaceRef != nil && intf.InterfaceRef.Config.Interface != nil {
                            aclData[aclName].Field["ports@"] = *intf.InterfaceRef.Config.Interface
                        }
                        aclData[aclName].Field["stage"] = "EGRESS"
                        ret = true
                    }
                }
            }
        }
    }
    return ret
}

func default_deny_rule(rulesInfo map[string]db.Value) {
    m := make(map[string]string)
    rulesInfo["DEFAULT_RULE"] = db.Value{ Field: m }
    rulesInfo["DEFAULT_RULE"].Field["PRIORITY"] = strconv.FormatInt(int64(MIN_PRIORITY), 10)
    rulesInfo["DEFAULT_RULE"].Field["PACKET_ACTION"] = "DROP"
}

func convert_oc_to_internal_rule(ruleData db.Value, aclName string, aclType ocbinds.E_OpenconfigAcl_ACL_TYPE, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
    ruleIndex := *rule.Config.SequenceId
    ruleData.Field["PRIORITY"] = strconv.FormatInt(int64(MAX_PRIORITY - ruleIndex), 10)
    if rule.Config != nil && rule.Config.Description != nil {
        ruleData.Field["RULE_DESCRIPTION"] = *rule.Config.Description
    }

    if ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 == aclType {
        convert_oc_to_internal_ipv4(ruleData, aclName, ruleIndex, rule)
    } else if ocbinds.OpenconfigAcl_ACL_TYPE_ACL_IPV4 == aclType {
        convert_oc_to_internal_ipv6(ruleData, aclName, ruleIndex, rule)
    } else if ocbinds.OpenconfigAcl_ACL_TYPE_ACL_L2 == aclType {
        convert_oc_to_internal_l2(ruleData, aclName, ruleIndex, rule)
    } /*else if ocbinds.OpenconfigAcl_ACL_TYPE_ACL_MIXED == aclType {
    } */

    convert_oc_to_internal_transport(ruleData, aclName, ruleIndex, rule)
    convert_oc_to_internal_input_interface(ruleData, aclName, ruleIndex, rule)
    convert_oc_to_internal_action(ruleData, aclName, ruleIndex, rule)
}

func convert_oc_to_internal_l2(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
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

func convert_oc_to_internal_ipv4(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
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

func convert_oc_to_internal_ipv6(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
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

func convert_oc_to_internal_transport(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
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
        for _,flag := range rule.Transport.Config.TcpFlags {
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

func convert_oc_to_internal_input_interface(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
    if rule.InputInterface != nil && rule.InputInterface.InterfaceRef != nil {
        ruleData.Field["IN_PORTS"] = *rule.InputInterface.InterfaceRef.Config.Interface
    }
}

func convert_oc_to_internal_action(ruleData db.Value, aclName string, ruleIndex uint32, rule *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry) {
    if rule.Actions.Config != nil {
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

func set_acl_data_in_config_db(d *db.DB, aclData map[string]db.Value) {
    for key, value := range aclData {

        /*
        existingEntry,_ := dbCl.GetEntry(&aclTs, db.Key{Comp: []string {key} })
        //Merge any ACL binds already present. Validate should take care of any checks so its safe to blindly merge here
        if len(existingEntry.Field) > 0  {
            value.Field["ports"] += "," + existingEntry.Field["ports@"]
        }
        fmt.Println(value)
        */
        err := d.SetEntry(&aclTs, db.Key{Comp: []string {key} }, value)
        if err != nil {
            fmt.Println(err)
        }
    }
}

func set_acl_rule_data_in_config_db(d *db.DB, ruleData map[string]map[string]db.Value) {
    for aclName,_ := range ruleData {
        for ruleName,rule := range ruleData[aclName] {
            err := d.SetEntry(&ruleTs, db.Key{Comp: []string {aclName, ruleName} }, rule)
            if err != nil {
                fmt.Println(err)
            }
        }
    }
}

/*
func set_acl_bind_data_in_config_db(dbCl *db.DB, aclData map[string]map[string]map) {
}
*/


func getIpProtocolConfig(proto int64) ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union {

    var protoCfg ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union

    foundInMap := false
    for _, v := range OLD_IP_PROTOCOL_MAP {
        if proto == int64(v) {
            foundInMap = true
        }
    }

    if foundInMap {
        var ipProCfg *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL
        ipProCfg = new (ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_E_OpenconfigPacketMatchTypes_IP_PROTOCOL)
        switch(uint8(proto)) {
        case OLD_IP_PROTOCOL_MAP["IP_ICMP"]:
            ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_ICMP
            return ipProCfg
        case OLD_IP_PROTOCOL_MAP["IP_IGMP"]:
            ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_IGMP
            return ipProCfg
        case OLD_IP_PROTOCOL_MAP["IP_TCP"]:
            ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_TCP
            return ipProCfg
        case OLD_IP_PROTOCOL_MAP["IP_UDP"]:
            ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_UDP
            return ipProCfg
        case OLD_IP_PROTOCOL_MAP["IP_RSVP"]:
            ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_RSVP
            return ipProCfg
        case OLD_IP_PROTOCOL_MAP["IP_GRE"]:
            ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_GRE
            return ipProCfg
        case OLD_IP_PROTOCOL_MAP["IP_AUTH"]:
            ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_AUTH
            return ipProCfg
        case OLD_IP_PROTOCOL_MAP["IP_PIM"]:
            ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_PIM
            return ipProCfg
        case OLD_IP_PROTOCOL_MAP["IP_L2TP"]:
            ipProCfg.E_OpenconfigPacketMatchTypes_IP_PROTOCOL = ocbinds.OpenconfigPacketMatchTypes_IP_PROTOCOL_IP_L2TP
            return ipProCfg
        }
    } else {
        var ipProCfgUint8 *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8
        ipProCfgUint8 = new (ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Ipv4_Config_Protocol_Union_Uint8)
        ipProCfgUint8.Uint8 = uint8(proto)
        return ipProCfgUint8
    }
    return protoCfg
}

func getTransportConfigDestPort(destPort string) ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union {
    portNum, _ := strconv.ParseInt(destPort, 10, 64)
    var destPortCfg *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16
    destPortCfg = new (ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_DestinationPort_Union_Uint16)
    destPortCfg.Uint16 = uint16(portNum)
    return destPortCfg
}

func getTransportConfigSrcPort(srcPort string) ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union {
    portNum, _ := strconv.ParseInt(srcPort, 10, 64)
    var srcPortCfg *ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16
    srcPortCfg = new (ocbinds.OpenconfigAcl_Acl_AclSets_AclSet_AclEntries_AclEntry_Transport_Config_SourcePort_Union_Uint16)
    srcPortCfg.Uint16 = uint16(portNum)
    return srcPortCfg
}


func getYangPathFromUri(uri string) (string, error) {
    var path *gnmi.Path
    var err error

    path, err = ygot.StringToPath(uri, ygot.StructuredPath, ygot.StringSlicePath)
    if err != nil {
        fmt.Println("Error in uri to path conversion: ", err)
        return "", err
    }

    yangPath, yperr := ygot.PathToSchemaPath(path)
    if yperr != nil {
        fmt.Println("Error in Gnmi path to Yang path conversion: ", yperr)
        return "", yperr
    }

    return yangPath, err
}

func getYangPathFromStruct(s ygot.GoStruct) string {
    tn := reflect.TypeOf(s).Elem().Name()
    schema, ok := ocbinds.SchemaTree[tn]
    if !ok {
        fmt.Errorf("could not find schema for type %s", tn )
        return ""
    } else if schema != nil {
        yPath := schema.Path()
        yPath = strings.Replace(yPath, "/device/acl", "/openconfig-acl:acl", 1)
        return yPath
    }
    return ""
}

/* Check if targetUriPath is child (subtree) of nodePath
The return value can be used to decide if subtrees needs
to visited to fill the data or not.
*/
func isSubtreeRequest(targetUriPath string, nodePath string) bool {
    return strings.HasPrefix(targetUriPath, nodePath)
}

func dumpIetfJson(s ygot.ValidatedGoStruct) (string, error) {
    jsonStr, err := ygot.EmitJSON(s, &ygot.EmitJSONConfig{
        Format: ygot.RFC7951,
        Indent: "  ",
        RFC7951Config: &ygot.RFC7951JSONConfig{
            AppendModuleName: true,
        },
    })
    return jsonStr, err
}

