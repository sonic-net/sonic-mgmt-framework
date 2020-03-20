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
	"reflect"
	"strconv"
	"translib/db"
	"translib/ocbinds"
)

type SysApp struct {
	path       *PathInfo
	reqData    []byte
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}

	dockerTs *db.TableSpec
	procTs   *db.TableSpec

	dockerTable map[string]dbEntry
	procTable   map[uint64]dbEntry
}

func init() {
	log.Info("SysApp: Init called for System module")
	err := register("/openconfig-system:system",
		&appInfo{appType: reflect.TypeOf(SysApp{}),
			ygotRootType: reflect.TypeOf(ocbinds.OpenconfigSystem_System{}),
			isNative:     false})
	if err != nil {
		log.Fatal("SysApp:  Register System app module with App Interface failed with error=", err)
	}

	err = addModel(&ModelData{Name: "openconfig-system",
		Org: "OpenConfig working group",
		Ver: "1.0.2"})
	if err != nil {
		log.Fatal("SysApp:  Adding model data to appinterface failed with error=", err)
	}
}

func (app *SysApp) initialize(data appData) {
	log.Info("SysApp: initialize:if:path =", data.path)

	app.path = NewPathInfo(data.path)
	app.reqData = data.payload
	app.ygotRoot = data.ygotRoot
	app.ygotTarget = data.ygotTarget

	app.dockerTs = &db.TableSpec{Name: "DOCKER_STATS"}
	app.procTs = &db.TableSpec{Name: "PROCESS_STATS"}
}

func (app *SysApp) getAppRootObject() *ocbinds.OpenconfigSystem_System {
	deviceObj := (*app.ygotRoot).(*ocbinds.Device)
	return deviceObj.System
}

func (app *SysApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
	var err error

	return nil, nil, err
}

func (app *SysApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys

	err = errors.New("SysApp Not implemented, translateCreate")
	return keys, err
}

func (app *SysApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	err = errors.New("SysApp Not implemented, translateUpdate")
	return keys, err
}

func (app *SysApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	err = errors.New("Not implemented SysApp translateReplace")
	return keys, err
}

func (app *SysApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys

	err = errors.New("Not implemented SysApp translateDelete")
	return keys, err
}

func (app *SysApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	var err error
	log.Info("SysApp: translateGet:intf:path =", app.path)
	return err
}

func (app *SysApp) processCreate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	err = errors.New("Not implemented SysApp processCreate")
	return resp, err
}

func (app *SysApp) processUpdate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	err = errors.New("Not implemented SysApp processUpdate")
	return resp, err
}

func (app *SysApp) processReplace(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	err = errors.New("Not implemented, SysApp processReplace")
	return resp, err
}

func (app *SysApp) processDelete(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	err = errors.New("Not implemented SysApp processDelete")
	return resp, err
}

type ProcessState struct {
	Args              []string
	CpuUsageSystem    uint64
	CpuUsageUser      uint64
	CpuUtilization    uint8
	MemoryUsage       uint64
	MemoryUtilization uint8
	Name              string
	Pid               uint64
	StartTime         uint64
	Uptime            uint64
}

func (app *SysApp) getSystemProcess(sysproc *ocbinds.OpenconfigSystem_System_Processes_Process, pid uint64) {

	log.Info("getSystemProcess pid=", pid)

	e := app.procTable[pid].entry

	var procstate ProcessState

	procstate.CpuUsageUser = 0
	procstate.CpuUsageSystem = 0
	procstate.MemoryUsage = 0
	f, _ := strconv.ParseFloat(e.Get("%MEM"), 32)
	procstate.MemoryUtilization = uint8(f)
	f, _ = strconv.ParseFloat(e.Get("%CPU"), 32)
	procstate.CpuUtilization = uint8(f)
	procstate.Name = e.Get("CMD")
	procstate.Pid = pid
	procstate.StartTime = 0
	procstate.Uptime = 0

	targetUriPath, perr := getYangPathFromUri(app.path.Path)
	if perr != nil {
		log.Infof("getYangPathFromUri failed.")
		return
	}

	ygot.BuildEmptyTree(sysproc)

	switch targetUriPath {

	case "/openconfig-system:system/processes/process/state/name":
		sysproc.State.Name = &procstate.Name
	case "/openconfig-system:system/processes/process/state/args":
	case "/openconfig-system:system/processes/process/state/start-time":
		sysproc.State.StartTime = &procstate.StartTime
	case "/openconfig-system:system/processes/process/state/uptime":
		sysproc.State.Uptime = &procstate.Uptime
	case "/openconfig-system:system/processes/process/state/cpu-usage-user":
		sysproc.State.CpuUsageUser = &procstate.CpuUsageUser
	case "/openconfig-system:system/processes/process/state/cpu-usage-system":
		sysproc.State.CpuUsageSystem = &procstate.CpuUsageSystem
	case "/openconfig-system:system/processes/process/state/cpu-utilization":
		sysproc.State.CpuUtilization = &procstate.CpuUtilization
	case "/openconfig-system:system/processes/process/state/memory-usage":
		sysproc.State.MemoryUsage = &procstate.MemoryUsage
	case "/openconfig-system:system/processes/process/state/memory-utilization":
		sysproc.State.MemoryUtilization = &procstate.MemoryUtilization
	default:
		sysproc.Pid = &procstate.Pid
		sysproc.State.CpuUsageSystem = &procstate.CpuUsageSystem
		sysproc.State.CpuUsageUser = &procstate.CpuUsageUser
		sysproc.State.CpuUtilization = &procstate.CpuUtilization
		sysproc.State.MemoryUsage = &procstate.MemoryUsage
		sysproc.State.MemoryUtilization = &procstate.MemoryUtilization
		sysproc.State.Name = &procstate.Name
		sysproc.State.Pid = &procstate.Pid
		sysproc.State.StartTime = &procstate.StartTime
		sysproc.State.Uptime = &procstate.Uptime
	}
}

func (app *SysApp) getSystemProcesses(sysprocs *ocbinds.OpenconfigSystem_System_Processes, ispid bool) {
	log.Infof("getSystemProcesses Entry")

	if ispid == true {
		pid, _ := app.path.IntVar("pid")
		sysproc := sysprocs.Process[uint64(pid)]

		app.getSystemProcess(sysproc, uint64(pid))

	} else {

		for pid, _ := range app.procTable {
			sysproc, err := sysprocs.NewProcess(pid)
			if err != nil {
				log.Infof("sysprocs.NewProcess failed")
				return
			}

			app.getSystemProcess(sysproc, pid)
		}
	}
	return
}

func (app *SysApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
	log.Info("SysApp: processGet Path: ", app.path.Path)

	stateDb := dbs[db.StateDB]

	var payload []byte
	empty_resp := GetResponse{Payload: payload}

	// Read docker info from DB

	app.dockerTable = make(map[string]dbEntry)

	tbl, err := stateDb.GetTable(app.dockerTs)
	if err != nil {
		log.Error("DOCKER_STATS table get failed!")
		return empty_resp, err
	}

	keys, _ := tbl.GetKeys()
	for _, key := range keys {
		e, err := tbl.GetEntry(key)
		if err != nil {
			log.Error("DOCKER_STATS entry get failed!")
			return empty_resp, err
		}

		app.dockerTable[key.Get(0)] = dbEntry{entry: e}
	}

	// Read process info from DB

	app.procTable = make(map[uint64]dbEntry)

	tbl, err = stateDb.GetTable(app.procTs)
	if err != nil {
		log.Error("PROCESS_STATS table get failed!")
		return empty_resp, err
	}

	keys, _ = tbl.GetKeys()
	for _, key := range keys {
		e, err := tbl.GetEntry(key)
		if err != nil {
			log.Error("PROCESS_STATS entry get failed!")
			return empty_resp, err
		}

		pid, _ := strconv.ParseUint(key.Get(0), 10, 64)
		app.procTable[pid] = dbEntry{entry: e}
	}

	sysObj := app.getAppRootObject()

	targetUriPath, perr := getYangPathFromUri(app.path.Path)
	if perr != nil {
		log.Infof("getYangPathFromUri failed.")
		return GetResponse{Payload: payload}, perr
	}

	log.Info("targetUriPath : ", targetUriPath, "Args: ", app.path.Vars)

	if isSubtreeRequest(targetUriPath, "/openconfig-system:system/processes") {
		if targetUriPath == "/openconfig-system:system/processes" {
			ygot.BuildEmptyTree(sysObj)
			app.getSystemProcesses(sysObj.Processes, false)
			payload, err = dumpIetfJson(sysObj, false)
		} else if targetUriPath == "/openconfig-system:system/processes/process" {
			pid, perr := app.path.IntVar("pid")
			if perr == nil {
				if pid == 0 {
					ygot.BuildEmptyTree(sysObj)
					app.getSystemProcesses(sysObj.Processes, false)
					payload, err = dumpIetfJson(sysObj.Processes, false)
				} else {
					app.getSystemProcesses(sysObj.Processes, true)
					payload, err = dumpIetfJson(sysObj.Processes, false)
				}
			}
		} else if targetUriPath == "/openconfig-system:system/processes/process/state" {
			pid, _ := app.path.IntVar("pid")
			app.getSystemProcesses(sysObj.Processes, true)
			payload, err = dumpIetfJson(sysObj.Processes.Process[uint64(pid)], true)
		} else if isSubtreeRequest(targetUriPath, "/openconfig-system:system/processes/process/state") {
			pid, _ := app.path.IntVar("pid")
			app.getSystemProcesses(sysObj.Processes, true)
			payload, err = dumpIetfJson(sysObj.Processes.Process[uint64(pid)].State, true)
		}
	} else {
		return GetResponse{Payload: payload}, errors.New("Not implemented processGet, path: ")
	}
	return GetResponse{Payload: payload}, err
}
