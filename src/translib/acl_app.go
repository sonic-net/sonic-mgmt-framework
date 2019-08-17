///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package translib

import (
	"reflect"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"
	"translib/transformer"

	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
)

const (
	TABLE_SEPARATOR = "|"
	KEY_SEPARATOR   = "|"
	ACL_TABLE       = "ACL_TABLE"
	RULE_TABLE      = "ACL_RULE"
)

type configHandler func(*db.DB, *AclApp, int) error

type AclApp struct {
	pathInfo   *PathInfo
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}

	aclTs  *db.TableSpec
	ruleTs *db.TableSpec

	aclTableMap  map[string]db.Value
	ruleTableMap map[string]db.Value
	//ruleTableMap map[string]map[string]db.Value
	callpoints map[string]configHandler
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
	app.ruleTableMap = make(map[string]db.Value)

	app.callpoints = map[string]configHandler{
		"/openconfig-acl:acl":                                        handleAcl,
		"/openconfig-acl:acl/acl-sets":                               handleAclSet,
		"/openconfig-acl:acl/acl-sets/acl-set":                       handleAclSet,
		"/openconfig-acl:acl/acl-sets/acl-set/acl-entries":           handleAclEntry,
		"/openconfig-acl:acl/acl-sets/acl-set/acl-entries/acl-entry": handleAclEntry,
		"/openconfig-acl:acl/interfaces":                             handleAclInterface,
	}
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

	keys, err = app.translateCRUCommon(d, DELETE)
	return keys, err
}

func (app *AclApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	var err error
	log.Info("translateGet:acl:path =", app.pathInfo.Template)
	return err
}

func (app *AclApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {

	return nil, nil, nil
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

	keyspec, err := transformer.XlateUriToKeySpec(app.pathInfo.Path, app.ygotRoot, app.ygotTarget)

	// table.key.fields
	var result = make(map[string]map[string]db.Value)

	for dbnum, specs := range *keyspec {
		for _, spec := range specs {
			err := transformer.TraverseDb(dbs[dbnum], spec, &result, nil)
			if err != nil {
				return GetResponse{Payload: payload}, err
			}
		}
	}

	payload, err = transformer.XlateFromDb(app.pathInfo.Path, result)
	if err != nil {
		return GetResponse{Payload: payload, ErrSrc: AppErr}, err
	}

	return GetResponse{Payload: payload}, err
}

func (app *AclApp) translateCRUCommon(d *db.DB, opcode int) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCRUCommon:acl:path =", app.pathInfo.Template)

	result, err := transformer.XlateToDb(app.pathInfo.Path, opcode, d, app.ygotRoot, app.ygotTarget)
	if err != nil {
		return nil, err
	}

	app.aclTableMap = result["ACL_TABLE"]
	app.ruleTableMap = result["ACL_RULE"]
	// add app specific as needed, e.g. default rule

	return keys, err
}

func handleAclInterface(d *db.DB, app *AclApp, opcode int) error {
	var err error
	for key, data := range app.aclTableMap {
		existingEntry, err := d.GetEntry(app.aclTs, db.Key{Comp: []string{key}})
		if !existingEntry.IsPopulated() {
			return tlerr.AlreadyExists("Acl %s already exists", key)
		}
		// !!! Overloaded xfmr methods tale care for bindings, to set the data
		switch opcode {
		case CREATE:
		case REPLACE:
		case UPDATE:
		case DELETE:
			err = d.SetEntry(app.aclTs, db.Key{Comp: []string{key}}, data)
		}
		if err != nil {
			break
		}
	}
	return err
}

func handleAclSet(d *db.DB, app *AclApp, opcode int) error {
	var err error

	// acl table
	for key, data := range app.aclTableMap {
		existingEntry, err := d.GetEntry(app.aclTs, db.Key{Comp: []string{key}})
		if opcode == CREATE && existingEntry.IsPopulated() {
			return tlerr.AlreadyExists("Acl %s already exists", key)
		}
		switch opcode {
		case CREATE:
			err = d.CreateEntry(app.aclTs, db.Key{Comp: []string{key}}, data)
		case REPLACE:
			if !existingEntry.IsPopulated() {
				err = d.CreateEntry(app.aclTs, db.Key{Comp: []string{key}}, data)
			} else {
				// !!! delete & add an entry?? Here just shows the set operation as showcase
				err = d.SetEntry(app.aclTs, db.Key{Comp: []string{key}}, data)
			}
		case UPDATE:
			err = d.ModEntry(app.aclTs, db.Key{Comp: []string{key}}, data)
		case DELETE:
			err = d.DeleteKeys(app.aclTs, db.Key{Comp: []string{key}})
		}
		if err != nil {
			break
		}
	}
	// acl rule
	if err == nil {
		err = handleAclEntry(d, app, opcode)
	}

	return err
}
func handleAclEntry(d *db.DB, app *AclApp, opcode int) error {
	var err error

	for key, data := range app.ruleTableMap {
		ruleName := strings.Split(key, KEY_SEPARATOR)[1]
		existingEntry, err := d.GetEntry(app.ruleTs, db.Key{Comp: []string{key}})
		if opcode == CREATE && existingEntry.IsPopulated() {
			return tlerr.AlreadyExists("Acl rule %s already exists", ruleName)
		}
		switch opcode {
		case CREATE:
			err = d.CreateEntry(app.ruleTs, db.Key{Comp: []string{key}}, data)
		case REPLACE:
			if !existingEntry.IsPopulated() {
				err = d.CreateEntry(app.ruleTs, db.Key{Comp: []string{key}}, data)
			} else {
				// !!! delete & add an entry?? Here just shows the merge operation as showcase
				err = d.SetEntry(app.ruleTs, db.Key{Comp: []string{key}}, data)
			}
		case UPDATE:
			err = d.ModEntry(app.ruleTs, db.Key{Comp: []string{key}}, data)
		case DELETE:
			err = d.DeleteKeys(app.ruleTs, db.Key{Comp: []string{key}})
		}
		if err != nil {
			break
		}
	}
	return err
}
func handleAcl(d *db.DB, app *AclApp, opcode int) error {
	var err error
	//err = app.processCommonToplevelPath(d, acl, opcode, true)
	err = handleAclSet(d, app, opcode)
	if err == nil {
		err = handleAclInterface(d, app, opcode)
	}
	return err
}

func (app *AclApp) processCommon(d *db.DB, opcode int) error {
	var err error

	log.Infof("processCommon--Path Received: %s", app.pathInfo.Template)
	// test
	d.Opts.DisableCVLCheck = true

	targetType := reflect.TypeOf(*app.ygotTarget)

	xpath, err := RemoveXPATHPredicates(app.pathInfo.Path)
	if err != nil {
		log.Errorf("processCommon: Failed to remove Xpath Predicates from path %s", app.pathInfo.Template)
	}

	if _, ok := app.callpoints[xpath]; ok {
		if err = app.callpoints[xpath](d, app, opcode); err != nil {
			log.Errorf("processCommon: Given path %s not handled", app.pathInfo.Template)
		}
	} else {
		// fallback callpoints
		if strings.Contains(xpath, "/openconfig-acl:acl/acl-sets/acl-set/acl-entries/acl-entry") {
			err = handleAclEntry(d, app, opcode)
		} else if strings.Contains(targetType.Elem().Name(), "/openconfig-acl:acl/interfaces") {
			err = handleAclInterface(d, app, opcode)
		}
		if err == nil {
			err = tlerr.NotSupported("URL %s is not supported", app.pathInfo.Template)
		}
	}

	return err
}
