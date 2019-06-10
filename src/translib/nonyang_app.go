///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package translib

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"translib/db"

	log "github.com/golang/glog"
)

// nonYangDemoApp holds all invocation and state information for
// the non-yang demo app
type nonYangDemoApp struct {
	// request information
	path    *PathInfo
	reqData []byte

	// DB client to operate on config_db
	confDB      *db.DB
	vlanTable   *db.TableSpec
	memberTable *db.TableSpec

	// Cahce for read operation
	respJSON interface{}

	// Cache for write operations
	watchKeys []db.WatchKeys
	vlans     []transVlanData
}

type transVlanData struct {
	id         int               // vlan id
	isVlanMod  bool              // indicate if vlan create/update
	isVlanDel  bool              // indicate if vlan delete operation
	allMembers map[string]bool   // all memebrs port names
	delMembers map[string]bool   // Member port names for delete
	modMembers map[string]string // Port name to tagging mode for create/update
}

type jsonObject map[string]interface{}
type jsonArray []interface{}

func init() {
	register(
		"/nonyang/",
		&appInfo{appType: reflect.TypeOf(nonYangDemoApp{}),
			isNative: true})
}

// initialize function prepares this nonYangDemoApp instance
// for a new request handling.
func (app *nonYangDemoApp) initialize(data appData) {
	app.path = NewPathInfo(data.path)
	app.reqData = data.payload

	app.vlanTable = &db.TableSpec{"VLAN"}
	app.memberTable = &db.TableSpec{"VLAN_MEMBER"}
}

func (app *nonYangDemoApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	app.confDB = d
	pathInfo := app.path
	var err error

	log.Infof("Received CREATE for path %s; vars=%v", pathInfo.Template, pathInfo.Vars)

	switch pathInfo.Template {
	case "/nonyang/vlan":
		err = app.translateCreateVlans()

	case "/nonyang/vlan/{id}/member":
		err = app.translateCreateVlanMembers()

	default:
		err = errors.New("Unknown path")
	}

	return app.watchKeys, err
}

func (app *nonYangDemoApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.New("Not implemented")
}

func (app *nonYangDemoApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.New("Not implemented")
}

func (app *nonYangDemoApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	app.confDB = d
	pathInfo := app.path
	var err error

	log.Infof("Received DELETE for path %s; vars=%v", pathInfo.Template, pathInfo.Vars)

	switch pathInfo.Template {
	case "/nonyang/vlan/{id}":
		err = app.translateDeleteVlan()

	case "/nonyang/vlan/{id}/member/{port}":
		err = app.translateDeleteVlanMember()

	default:
		err = errors.New("Unknown path")
	}

	return app.watchKeys, err
}

func (app *nonYangDemoApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	return nil
}

func (app *nonYangDemoApp) processCreate(d *db.DB) (SetResponse, error) {
	var resp SetResponse
	err := app.writeToDatabase()
	return resp, err
}

func (app *nonYangDemoApp) processUpdate(d *db.DB) (SetResponse, error) {
	var resp SetResponse
	return resp, errors.New("Not implemented")
}

func (app *nonYangDemoApp) processReplace(d *db.DB) (SetResponse, error) {
	var resp SetResponse
	return resp, errors.New("Not implemented")
}

func (app *nonYangDemoApp) processDelete(d *db.DB) (SetResponse, error) {
	var resp SetResponse
	err := app.writeToDatabase()
	return resp, err
}

func (app *nonYangDemoApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
	app.confDB = dbs[db.ConfigDB]
	pathInfo := app.path
	var err error

	log.Infof("Received GET for path %s; vars=%v", pathInfo.Template, pathInfo.Vars)

	switch pathInfo.Template {
	case "/nonyang/vlan":
		err = app.doGetAllVlans()

	case "/nonyang/vlan/{id}":
		err = app.doGetVlanByID()

	default:
		err = errors.New("Unknown path")
	}

	var respData []byte
	if err == nil && app.respJSON != nil {
		respData, err = json.Marshal(app.respJSON)
	}

	return GetResponse{Payload: respData}, err
}

// doGetAllVlans is the handler for "/nonyang/vlan" path
// Loads all vlans and member data from db and prepares
// a json array - each item being one vlan info.
func (app *nonYangDemoApp) doGetAllVlans() error {
	log.Infof("in GetAllVlans")

	// Get all vlans from db
	vlanTable, err := app.confDB.GetTable(app.vlanTable)
	if err != nil {
		return err
	}

	var allVlansJSON jsonArray

	keys, _ := vlanTable.GetKeys()
	log.Infof("Found %d VLAN table keys", len(keys))
	for _, key := range keys {
		log.Infof("Processing %v", key.Get(0))

		vlanInfo, _ := vlanTable.GetEntry(key)
		vlanJSON, err := app.getVlanJSON(&vlanInfo)
		if err != nil {
			return err
		}

		allVlansJSON = append(allVlansJSON, *vlanJSON)
	}

	app.respJSON = &allVlansJSON
	return nil
}

// doGetVlanByID is the handler for "/nonyang/vlan/{id}" path.
// Loads data for one vlan and its members and prepares a json
// object.
func (app *nonYangDemoApp) doGetVlanByID() error {
	log.Infof("in GetVlanByID()")

	vlanID, _ := app.path.IntVar("id")
	if !isValidVlan(vlanID) {
		log.Errorf("Got invalid vlan param \"%s\"", app.path.Var("id"))
		return errors.New("Invalid vlan id")
	}

	vlanName := toVlanName(vlanID)
	log.Infof("Processing %v", vlanName)

	vlanEntry, err := app.confDB.GetEntry(app.vlanTable, asKey(vlanName))
	if err == nil {
		app.respJSON, err = app.getVlanJSON(&vlanEntry)
	}

	return err
}

// getVlanJSON prepares a raw json object for given VLAN table
// entry. Member information are fetched from VLAN_MEMBER table.
func (app *nonYangDemoApp) getVlanJSON(vlanEntry *db.Value) (*jsonObject, error) {
	vlanJSON := make(jsonObject)
	var memberListJSON jsonArray

	vlanID, _ := vlanEntry.GetInt("vlanid")
	vlanName := toVlanName(vlanID)

	log.Infof("Preparing json for vlan %d", vlanID)

	memberPorts := vlanEntry.GetList("members")
	log.Infof("%s members = %v", vlanName, memberPorts)

	for _, portName := range memberPorts {
		memberJSON := make(jsonObject)
		memberJSON["port"] = portName

		memberEntry, err := app.confDB.GetEntry(app.memberTable, asKey(vlanName, portName))
		if err != nil {
			// ignore "not exists" error; don't fill tagging mode
			log.Warningf("Failed to load VLAN_MEMBER %s,%s; err=%v",
				vlanName, portName, err)
		} else {
			memberJSON["mode"] = memberEntry.Get("tagging_mode")
		}

		memberListJSON = append(memberListJSON, memberJSON)
	}

	vlanJSON["id"] = vlanID
	vlanJSON["name"] = vlanName
	vlanJSON["members"] = memberListJSON

	return &vlanJSON, nil
}

// translateCreateVlans handles CREATE operation on "/nonyang/vlan" path.
func (app *nonYangDemoApp) translateCreateVlans() error {
	log.Infof("in translateCreateVlans()")

	// vlan creation expects array of vlan ids.
	var vlansJSON []int
	err := json.Unmarshal(app.reqData, &vlansJSON)
	if err != nil {
		log.Errorf("Failed to parse input.. err=%v", err)
		return errors.New("Invalid input")
	}

	log.Infof("Receieved %d vlan ids; %v", len(vlansJSON), vlansJSON)

	for _, vid := range vlansJSON {
		log.Infof("Processing vlan id %d", vid)
		if !isValidVlan(vid) {
			return errors.New("Invalid vlan id")
		}

		vlanData := transVlanData{
			id:        vid,
			isVlanMod: true,
		}

		app.vlans = append(app.vlans, vlanData)
		app.addWatchKey(app.vlanTable, toVlanName(vid))
	}

	return nil
}

// doCreateVlanMembers handles CREATE operation on path
// "/nonyang/vlan/{id}/member"
func (app *nonYangDemoApp) translateCreateVlanMembers() error {
	log.Infof("in translateCreateVlanMembers()")

	vlanID, _ := app.path.IntVar("id")
	if !isValidVlan(vlanID) {
		log.Errorf("Got invalid vlan param \"%s\"", app.path.Var("id"))
		return errors.New("Invalid vlan id")
	}

	var memberListJSON []map[string]string
	err := json.Unmarshal(app.reqData, &memberListJSON)
	if err != nil {
		log.Errorf("Failed to parse input.. err=%v", err)
		return errors.New("Invalid input")
	}

	currentMembers, err := app.getMemberPortsAsMap(vlanID)
	if err != nil {
		return err
	}

	vlanData := transVlanData{
		id:         vlanID,
		allMembers: currentMembers,
		modMembers: make(map[string]string),
	}

	vlanName := toVlanName(vlanID)
	app.addWatchKey(app.vlanTable, vlanName)

	for _, memberJSON := range memberListJSON {
		log.Infof("Processing member info %v", memberJSON)

		portName, ok := memberJSON["port"]
		if !ok || len(portName) == 0 {
			log.Infof("Invalid input - 'port' missing")
			return errors.New("Invalid input")
		}

		taggingMode, ok := memberJSON["mode"]
		if !ok {
			taggingMode = "tagged"
		} else if taggingMode != "tagged" && taggingMode != "untagged" {
			log.Infof("Invalid input - bad tagging mode '%s'", taggingMode)
			return errors.New("Invalid input")
		}

		vlanData.isVlanMod = true
		vlanData.allMembers[portName] = true
		vlanData.modMembers[portName] = taggingMode
		app.addWatchKey(app.memberTable, vlanName, portName)
	}

	app.vlans = append(app.vlans, vlanData)
	return nil
}

func (app *nonYangDemoApp) translateDeleteVlan() error {
	vlanID, _ := app.path.IntVar("id")
	log.Infof("in translateDeleteVlan(); vid=%d", vlanID)

	members, err := app.getMemberPortsAsMap(vlanID)
	if err != nil {
		return err
	}

	vlanData := transVlanData{
		id:         vlanID,
		isVlanDel:  true,
		delMembers: make(map[string]bool),
	}

	vlanName := toVlanName(vlanID)
	app.addWatchKey(app.vlanTable, vlanName)

	for portName := range members {
		vlanData.delMembers[portName] = true
		app.addWatchKey(app.memberTable, vlanName, portName)
	}

	app.vlans = append(app.vlans, vlanData)
	return nil
}

func (app *nonYangDemoApp) translateDeleteVlanMember() error {
	vlanID, _ := app.path.IntVar("id")
	portName := app.path.Var("port")
	log.Infof("in translateDeleteVlan(); vid=%d, member=%s", vlanID, portName)

	members, err := app.getMemberPortsAsMap(vlanID)
	if err != nil {
		return err
	}

	_, ok := members[portName]
	if !ok {
		log.Errorf("%s is not a member of vlan %d", portName, vlanID)
		return errors.New("Entry does not exist")
	}

	vlanData := transVlanData{
		id:         vlanID,
		isVlanMod:  true,
		allMembers: members,
		delMembers: make(map[string]bool),
	}

	delete(vlanData.allMembers, portName)
	vlanData.delMembers[portName] = true

	vlanName := toVlanName(vlanID)
	app.addWatchKey(app.vlanTable, vlanName)
	app.addWatchKey(app.memberTable, vlanName, portName)

	app.vlans = append(app.vlans, vlanData)
	return nil
}

func (app *nonYangDemoApp) getMemberPortsAsMap(vid int) (map[string]bool, error) {
	vlanKey := asKey(toVlanName(vid))
	vlanEntry, err := app.confDB.GetEntry(app.vlanTable, vlanKey)
	if err != nil {
		return nil, err
	}

	portsMap := make(map[string]bool)
	for _, portName := range vlanEntry.GetList("members") {
		portsMap[portName] = true
	}

	return portsMap, nil
}

func (app *nonYangDemoApp) addWatchKey(ts *db.TableSpec, keyParts ...string) {
	key := asKey(keyParts...)
	app.watchKeys = append(app.watchKeys, db.WatchKeys{Ts: ts, Key: &key})
}

func (app *nonYangDemoApp) writeToDatabase() error {
	if app.vlans == nil {
		log.Infof("No operations found")
		return nil
	}

	for _, vlanData := range app.vlans {
		vlanName := toVlanName(vlanData.id)
		vlanKey := asKey(vlanName)

		if vlanData.isVlanDel {
			log.Infof("DEL vlan entry '%s'", vlanName)
			err := app.confDB.DeleteEntry(app.vlanTable, vlanKey)
			if err != nil {
				return err
			}
		}

		if vlanData.isVlanMod {
			var memberNames []string
			for port := range vlanData.allMembers {
				memberNames = append(memberNames, port)
			}

			log.Infof("SET vlan entry '%s', members=%v", vlanName, memberNames)
			vlanValue := db.Value{Field: make(map[string]string)}
			vlanValue.SetInt("vlanid", vlanData.id)
			vlanValue.SetList("members", memberNames)

			err := app.confDB.SetEntry(app.vlanTable, vlanKey, vlanValue)
			if err != nil {
				return err
			}
		}

		for port := range vlanData.delMembers {
			log.Infof("DEL vlan_member entry '%s|%s'", vlanName, port)
			memberKey := asKey(vlanName, port)
			err := app.confDB.DeleteEntry(app.memberTable, memberKey)
			if err != nil {
				return err
			}
		}

		for port, mode := range vlanData.modMembers {
			log.Infof("SET vlan_member entry '%s|%s'; mode=%s", vlanName, port, mode)
			memberKey := asKey(vlanName, port)
			memberValue := db.Value{Field: make(map[string]string)}
			memberValue.Set("tagging_mode", mode)

			err := app.confDB.SetEntry(app.memberTable, memberKey, memberValue)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// asKey cretaes a db.Key from given key components
func asKey(parts ...string) db.Key {
	return db.Key{Comp: parts}
}

// toVlanName returns the vlan name for given vlan id.
func toVlanName(vid int) string {
	return fmt.Sprintf("Vlan%d", vid)
}

// isValidVlan checks if given number is within valid
// valn id range of 1-4095.
func isValidVlan(num int) bool {
	return (num > 0 && num < 4096)
}
