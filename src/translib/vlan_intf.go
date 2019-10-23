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
	"translib/db"
	"translib/tlerr"
)

/******** CONFIG FUNCTIONS ********/

func (app *IntfApp) translateUpdateVlanIntf(d *db.DB, vlanName *string, inpOp reqType) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys

	intfObj := app.getAppRootObject()

	m := make(map[string]string)
	entryVal := db.Value{Field: m}
	entryVal.Field["vlanid"], err = getVlanIdFromVlanName(vlanName)
	if err != nil {
		return keys, err
	}

	vlan := intfObj.Interface[*vlanName]
	curr, _ := d.GetEntry(app.vlanD.vlanTs, db.Key{Comp: []string{*vlanName}})
	if !curr.IsPopulated() {
		log.Info("VLAN-" + *vlanName + " not present in DB, need to create it!!")
		app.ifTableMap[*vlanName] = dbEntry{op: opCreate, entry: entryVal}
		return keys, nil
	}
	app.translateUpdateIntfConfig(vlanName, vlan, &curr)
	return keys, err
}

func (app *IntfApp) processUpdateVlanIntfConfig(d *db.DB) error {
	var err error

	for vlanId, vlanEntry := range app.ifTableMap {
		switch vlanEntry.op {
		case opCreate:
			err = d.CreateEntry(app.vlanD.vlanTs, db.Key{Comp: []string{vlanId}}, vlanEntry.entry)
			if err != nil {
				errStr := "Creating VLAN entry for VLAN : " + vlanId + " failed"
				return errors.New(errStr)
			}
		case opUpdate:
			err = d.SetEntry(app.vlanD.vlanTs, db.Key{Comp: []string{vlanId}}, vlanEntry.entry)
			if err != nil {
				errStr := "Updating VLAN entry for VLAN : " + vlanId + " failed"
				return errors.New(errStr)
			}
		}
	}
	return err
}

func (app *IntfApp) processUpdateVlanIntf(d *db.DB) error {
	var err error
	err = app.processUpdateVlanIntfConfig(d)
	if err != nil {
		return err
	}

	return err
}

/********* DELETE FUNCTIONS ********/

func (app *IntfApp) translateDeleteVlanIntf(d *db.DB, vlan *string) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	curr, err := d.GetEntry(app.vlanD.vlanTs, db.Key{Comp: []string{*vlan}})
	if err != nil {
		vlanName := *vlan
		vlanId := vlanName[len("Vlan"):len(vlanName)]
		errStr := "Invalid Vlan: " + vlanId
		return keys, tlerr.InvalidArgsError{Format: errStr}
	}
	app.ifTableMap[*vlan] = dbEntry{entry: curr, op: opDelete}
	return keys, err
}

func (app *IntfApp) processDeleteVlanIntfAndMembers(d *db.DB) error {
	var err error

	for vlanKey, dbentry := range app.ifTableMap {
		memberPortsVal, ok := dbentry.entry.Field["members@"]
		if ok {
			memberPorts := generateMemberPortsSliceFromString(&memberPortsVal)
			if memberPorts == nil {
				return nil
			}
			log.Info("MemberPorts = ", memberPortsVal)

			for _, memberPort := range memberPorts {
				log.Infof("Member Port:%s part of vlan:%s to be deleted!", memberPort, vlanKey)
				err = d.DeleteEntry(app.vlanD.vlanMemberTs, db.Key{Comp: []string{vlanKey, memberPort}})
				if err != nil {
					return err
				}
			}
		}
		err = d.DeleteEntry(app.vlanD.vlanTs, db.Key{Comp: []string{vlanKey}})
		if err != nil {
			return err
		}
	}
	return err
}

func (app *IntfApp) processDeleteVlanIntf(d *db.DB) error {
	var err error

	err = app.processDeleteVlanIntfAndMembers(d)
	if err != nil {
		return err
	}
	return err
}
