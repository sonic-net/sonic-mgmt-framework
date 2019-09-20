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
    "reflect"
    "encoding/json"
    "errors"
    "translib/db"
    "translib/ocbinds"
    "github.com/openconfig/ygot/ygot"
    "os"
    "strconv"
    "time"
    "io/ioutil"
    "syscall"
    "translib/tlerr"
    log "github.com/golang/glog"
)

type SysApp struct {
    path        *PathInfo
    reqData     []byte
    ygotRoot    *ygot.GoStruct
    ygotTarget  *interface{}


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
    Ver:      "1.0.2"})
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

}

func (app *SysApp) getAppRootObject() (*ocbinds.OpenconfigSystem_System) {
	deviceObj := (*app.ygotRoot).(*ocbinds.Device)
	return deviceObj.System
}

func (app *SysApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {

    var err error
    return nil, nil, err
}

func (app *SysApp) translateCreate(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys

    err = errors.New("SysApp Not implemented, translateCreate")
    return keys, err
}

func (app *SysApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys
    err = errors.New("SysApp Not implemented, translateUpdate")
    return keys, err
}

func (app *SysApp) translateReplace(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys
    err = errors.New("Not implemented SysApp translateReplace")
    return keys, err
}

func (app *SysApp) translateDelete(d *db.DB) ([]db.WatchKeys, error)  {
    var err error
    var keys []db.WatchKeys

    err = errors.New("Not implemented SysApp translateDelete")
    return keys, err
}

func (app *SysApp) translateGet(dbs [db.MaxDB]*db.DB) error  {
    var err error
    log.Info("SysApp: translateGet:intf:path =", app.path)
    return err
}

func (app *SysApp) processCreate(d *db.DB) (SetResponse, error)  {
    var err error
    var resp SetResponse

    err = errors.New("Not implemented SysApp processCreate")
    return resp, err
}

func (app *SysApp) processUpdate(d *db.DB) (SetResponse, error)  {
    var err error
    var resp SetResponse

    err = errors.New("Not implemented SysApp processUpdate")
    return resp, err
}

func (app *SysApp) processReplace(d *db.DB) (SetResponse, error)  {
    var err error
    var resp SetResponse
    err = errors.New("Not implemented, SysApp processReplace")
    return resp, err
}

func (app *SysApp) processDelete(d *db.DB) (SetResponse, error)  {
    var err error
    var resp SetResponse

    err = errors.New("Not implemented SysApp processDelete")
    return resp, err
}

func (app *SysApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error)  {
    log.Info("SysApp: processGet Path: ", app.path.Path)

    return app.doGetSystem(app.path.Path)
}

func (app *SysApp) doGetSystem(path string) (GetResponse, error)  {

    var payload []byte
    jsonsystem, err := getSystemInfoFromFile()
    if err != nil {
        log.Infof("getSystemInfoFromFile failed")
        return GetResponse{Payload: payload}, err
    }
    sysObj := app.getAppRootObject()

    targetUriPath, perr := getYangPathFromUri(app.path.Path)
    if perr != nil {
        log.Infof("getYangPathFromUri failed.")
        return GetResponse{Payload: payload}, perr
    }

    log.Info("targetUriPath : ", targetUriPath, "Args: ", app.path.Vars)

    if isSubtreeRequest(targetUriPath, "/openconfig-system:system/state") {
        ygot.BuildEmptyTree(sysObj)
        app.getSystemState(&jsonsystem, sysObj.State)
        if targetUriPath == "/openconfig-system:system/state" {
            payload, err = dumpIetfJson(sysObj, false)
        } else {
            payload, err = dumpIetfJson(sysObj.State, false)
        }
    } else if isSubtreeRequest(targetUriPath, "/openconfig-system:system/memory") {
        ygot.BuildEmptyTree(sysObj)
        sysObj.Memory.State = &ocbinds.OpenconfigSystem_System_Memory_State{}
        app.getSystemMemory(&jsonsystem, sysObj.Memory.State)
        if targetUriPath == "/openconfig-system:system/memory" {
            payload, err = dumpIetfJson(sysObj, false)
        } else if targetUriPath == "/openconfig-system:system/memory/state" {
            payload, err = dumpIetfJson(sysObj.Memory, false)
        } else {
            payload, err = dumpIetfJson(sysObj.Memory.State, false)
        }
    } else if isSubtreeRequest(targetUriPath, "/openconfig-system:system/cpus") {
        ygot.BuildEmptyTree(sysObj)
        var index ocbinds.OpenconfigSystem_System_Cpus_Cpu_State_Index_Union_Uint32
        app.getSystemCpus(&jsonsystem, sysObj.Cpus)
        if targetUriPath == "/openconfig-system:system/cpus" {
            payload, err = dumpIetfJson(sysObj, false)
        } else if targetUriPath == "/openconfig-system:system/cpus/cpu" {
            payload, err = dumpIetfJson(sysObj.Cpus, false)
        } else {
            idx := 0
            index.Uint32 = uint32(idx)
            payload, err = dumpIetfJson(sysObj.Cpus.Cpu[&index], false)
        }
    } else if isSubtreeRequest(targetUriPath, "/openconfig-system:system/processes"){
        if targetUriPath == "/openconfig-system:system/processes" {
            ygot.BuildEmptyTree(sysObj)
            app.getSystemProcesses(&jsonsystem, sysObj.Processes, false)
            payload, err = dumpIetfJson(sysObj, false)
        } else if targetUriPath == "/openconfig-system:system/processes/process" {
            pid, perr := app.path.IntVar("pid")
            if perr == nil {
                if pid == 0 {
                    ygot.BuildEmptyTree(sysObj)
                    app.getSystemProcesses(&jsonsystem, sysObj.Processes, false)
                    payload, err = dumpIetfJson(sysObj.Processes, false)
                } else {
                    app.getSystemProcesses(&jsonsystem, sysObj.Processes, true)
                    payload, err = dumpIetfJson(sysObj.Processes, false)
                }
            }
        } else if targetUriPath == "/openconfig-system:system/processes/process/state" {
            pid, _ := app.path.IntVar("pid")
            app.getSystemProcesses(&jsonsystem, sysObj.Processes, true)
            payload, err = dumpIetfJson(sysObj.Processes.Process[uint64(pid)], true)
        } else if isSubtreeRequest(targetUriPath, "/openconfig-system:system/processes/process/state"){
            pid, _ := app.path.IntVar("pid")
            app.getSystemProcesses(&jsonsystem, sysObj.Processes, true)
            payload, err = dumpIetfJson(sysObj.Processes.Process[uint64(pid)].State, true)
        }
    } else if targetUriPath == "/openconfig-system:system" {
        ygot.BuildEmptyTree(sysObj)
        sysObj.Memory.State = &ocbinds.OpenconfigSystem_System_Memory_State{}
        app.getSystemState(&jsonsystem, sysObj.State)
        app.getSystemMemory(&jsonsystem, sysObj.Memory.State)
        app.getSystemCpus(&jsonsystem, sysObj.Cpus)
        app.getSystemProcesses(&jsonsystem, sysObj.Processes, false)
        payload, err = dumpIetfJson((*app.ygotRoot).(*ocbinds.Device), true)
    } else {
        err = errors.New("Not implemented doGetSystem, path: ")
        return  GetResponse{Payload: payload}, err
    }
    return  GetResponse{Payload: payload}, err
}


type JSONSystem  struct {
    Hostname       string  `json:"hostname"`
    Total          uint64  `json:"total"`
    Used           uint64  `json:"used"`
    Free           uint64  `json:"free"`

    Cpus  []Cpu            `json:"cpus"`
    Procs map[string]Proc  `json:"procs"`

}

type Cpu struct {
    User     int64   `json:"user"`
    System   int64   `json:"system"`
    Idle     int64   `json:"idle"`
}

type Proc struct {
    Cmd        string     `json:"cmd"`
    Start      uint64     `json:"start"`
    User       uint64     `json:"user"`
    System     uint64     `json:"system"`
    Mem        uint64     `json:"mem"`
    Cputil   float32    `json:"cputil"`
    Memutil   float32    `json:"memutil"`
}

func getSystemInfoFromFile () (JSONSystem, error) {
    log.Infof("getSystemInfoFromFile Enter")

    var jsonsystem JSONSystem 
    jsonFile, err := os.Open("/mnt/platform/system")
    if err != nil {
        log.Infof("system json open failed")
        errStr := "Information not available or Platform support not added"
        terr := tlerr.NotFoundError{Format: errStr}
        return jsonsystem, terr
    }
    syscall.Flock(int(jsonFile.Fd()),syscall.LOCK_EX)
    log.Infof("syscall.Flock done")

    defer jsonFile.Close()
    defer log.Infof("jsonFile.Close called")
    defer syscall.Flock(int(jsonFile.Fd()), syscall.LOCK_UN);
    defer log.Infof("syscall.Flock unlock  called")

    byteValue, _ := ioutil.ReadAll(jsonFile)
    json.Unmarshal(byteValue, &jsonsystem)
    return jsonsystem, nil
}

func (app *SysApp) getSystemState (sys *JSONSystem, sysstate *ocbinds.OpenconfigSystem_System_State) () {
    log.Infof("getSystemState Entry")

    targetUriPath, perr := getYangPathFromUri(app.path.Path)
    if perr != nil {
        log.Infof("getYangPathFromUri failed.")
        return
    }

    crtime := time.Now().Format(time.RFC3339) + "+00:00"

    switch targetUriPath {
    case "/openconfig-system:system/state/hostname":
        sysstate.Hostname = &sys.Hostname
    case "/openconfig-system:system/state/current-datetime":
        sysstate.CurrentDatetime = &crtime;
    case "/openconfig-system:system/state/boot-time":
        sysinfo := syscall.Sysinfo_t{}
        sys_err := syscall.Sysinfo(&sysinfo)
        if sys_err == nil {
            boot_time := uint64 (time.Now().Unix() - sysinfo.Uptime)
            sysstate.BootTime = &boot_time
        }
    case "/openconfig-system:system/state":
        sysstate.Hostname = &sys.Hostname
        sysstate.CurrentDatetime = &crtime;
        sysinfo := syscall.Sysinfo_t{}
        sys_err := syscall.Sysinfo(&sysinfo)
        if sys_err == nil {
            boot_time := uint64 (time.Now().Unix() - sysinfo.Uptime)
            sysstate.BootTime = &boot_time
        }
    }
}

func (app *SysApp) getSystemMemory (sys *JSONSystem, sysmem *ocbinds.OpenconfigSystem_System_Memory_State) () {
    log.Infof("getSystemMemory Entry")

    targetUriPath, perr := getYangPathFromUri(app.path.Path)
    if perr != nil {
        log.Infof("getYangPathFromUri failed.")
        return
    }

    switch targetUriPath {
    case "/openconfig-system:system/memory":
        fallthrough
    case "/openconfig-system:system/memory/state":
        sysmem.Physical = &sys.Total
        sysmem.Reserved = &sys.Used
    case "/openconfig-system:system/memory/state/physical":
        sysmem.Physical = &sys.Total
    case "/openconfig-system:system/memory/state/reserved":
        sysmem.Reserved = &sys.Used
    }
}

type CpuState struct {
    user uint8
    system uint8
    idle   uint8
}

func (app *SysApp) getSystemCpus (sys *JSONSystem, syscpus *ocbinds.OpenconfigSystem_System_Cpus) {
    log.Infof("getSystemCpus Entry")

    sysinfo := syscall.Sysinfo_t{}
    sys_err := syscall.Sysinfo(&sysinfo)
    if sys_err != nil {
        log.Infof("syscall.Sysinfo failed.")
    }

    for  idx, cpu := range sys.Cpus {
        var index  ocbinds.OpenconfigSystem_System_Cpus_Cpu_State_Index_Union_Uint32
        index.Uint32 = uint32(idx)
        syscpu, err := syscpus.NewCpu(&index)
        if err != nil {
            log.Infof("syscpus.NewCpu failed")
            return
        }
        ygot.BuildEmptyTree(syscpu)
        syscpu.Index = &index
        var cpucur CpuState
        if idx == 0 {
            cpucur.user = uint8((cpu.User/4)/sysinfo.Uptime)
            cpucur.system = uint8((cpu.System/4)/sysinfo.Uptime)
            cpucur.idle = uint8((cpu.Idle/4)/sysinfo.Uptime)
        } else {
            cpucur.user = uint8(cpu.User/sysinfo.Uptime)
            cpucur.system = uint8(cpu.System/sysinfo.Uptime)
            cpucur.idle = uint8(cpu.Idle/sysinfo.Uptime)
        }
        syscpu.State.User.Instant = &cpucur.user
        syscpu.State.Kernel.Instant = &cpucur.system
        syscpu.State.Idle.Instant = &cpucur.idle
    }
}

type ProcessState struct {
    Args [] string
    CpuUsageSystem uint64
    CpuUsageUser   uint64
    CpuUtilization uint8
    MemoryUsage    uint64
    MemoryUtilization uint8
    Name              string
    Pid               uint64
    StartTime         uint64
    Uptime            uint64
}

func (app *SysApp) getSystemProcess (proc *Proc, sysproc *ocbinds.OpenconfigSystem_System_Processes_Process, pid uint64) {

    var procstate ProcessState

    ygot.BuildEmptyTree(sysproc)
    procstate.CpuUsageUser = proc.User
    procstate.CpuUsageSystem = proc.System
    procstate.MemoryUsage  = proc.Mem * 1024
    procstate.MemoryUtilization = uint8(proc.Memutil)
    procstate.CpuUtilization  = uint8(proc.Cputil)
    procstate.Name = proc.Cmd
    procstate.Pid = pid
    procstate.StartTime = proc.Start * 1000000000  // ns
    procstate.Uptime = uint64(time.Now().Unix()) - proc.Start

    targetUriPath, perr := getYangPathFromUri(app.path.Path)
    if perr != nil {
        log.Infof("getYangPathFromUri failed.")
        return
    }

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
        sysproc.State.CpuUtilization =  &procstate.CpuUtilization
    case "/openconfig-system:system/processes/process/state/memory-usage":
        sysproc.State.MemoryUsage = &procstate.MemoryUsage
    case "/openconfig-system:system/processes/process/state/memory-utilization":
        sysproc.State.MemoryUtilization = &procstate.MemoryUtilization
    default:
        sysproc.Pid = &procstate.Pid 
        sysproc.State.CpuUsageSystem = &procstate.CpuUsageSystem
        sysproc.State.CpuUsageUser = &procstate.CpuUsageUser
        sysproc.State.CpuUtilization =  &procstate.CpuUtilization
        sysproc.State.MemoryUsage = &procstate.MemoryUsage
        sysproc.State.MemoryUtilization = &procstate.MemoryUtilization
        sysproc.State.Name = &procstate.Name
        sysproc.State.Pid = &procstate.Pid
        sysproc.State.StartTime = &procstate.StartTime
        sysproc.State.Uptime = &procstate.Uptime
    }
}

func (app *SysApp) getSystemProcesses (sys *JSONSystem, sysprocs *ocbinds.OpenconfigSystem_System_Processes, ispid bool) {
    log.Infof("getSystemProcesses Entry")

    if ispid == true {
        proc := sys.Procs[app.path.Var("pid")]
        pid,_ := app.path.IntVar("pid")
        sysproc := sysprocs.Process[uint64(pid)]

        app.getSystemProcess(&proc, sysproc, uint64(pid))

    } else {

        for  pidstr,  proc := range sys.Procs {
            idx, _:= strconv.Atoi(pidstr)

            sysproc, err := sysprocs.NewProcess(uint64 (idx))
            if err != nil {
                log.Infof("sysprocs.NewProcess failed")
                return
            }

            app.getSystemProcess(&proc, sysproc, uint64 (idx))
        }
    }
    return
}
