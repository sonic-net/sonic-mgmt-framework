package transformer

import (
    "encoding/json"
    "translib/ocbinds"
    "translib/tlerr"
    "time"
    "io/ioutil"
    "syscall"
    "strconv"
    "os"
    log "github.com/golang/glog"
    ygot "github.com/openconfig/ygot/ygot"
)

func init () {
    XlateFuncBind("DbToYang_sys_state_xfmr", DbToYang_sys_state_xfmr)
    XlateFuncBind("DbToYang_sys_memory_xfmr", DbToYang_sys_memory_xfmr)
    XlateFuncBind("DbToYang_sys_cpus_xfmr", DbToYang_sys_cpus_xfmr)
    XlateFuncBind("DbToYang_sys_procs_xfmr", DbToYang_sys_procs_xfmr)
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

type CpuState struct {
    user uint8
    system uint8
    idle   uint8
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

func getAppRootObject(inParams XfmrParams) (*ocbinds.OpenconfigSystem_System) {                                                                  
    deviceObj := (*inParams.ygRoot).(*ocbinds.Device)
    return deviceObj.System                                                                                                                 
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
    err = json.Unmarshal(byteValue, &jsonsystem)
	if err != nil {
        errStr := "json.Unmarshal failed"
        terr := tlerr.InternalError{Format: errStr}
        return jsonsystem, terr
	
	}
    return jsonsystem, nil
}

func getSystemState (sys *JSONSystem, sysstate *ocbinds.OpenconfigSystem_System_State) () {
    log.Infof("getSystemState Entry")

    crtime := time.Now().Format(time.RFC3339) + "+00:00"

    sysstate.Hostname = &sys.Hostname
    sysstate.CurrentDatetime = &crtime;
    sysinfo := syscall.Sysinfo_t{}

    sys_err := syscall.Sysinfo(&sysinfo)
    if sys_err == nil {
        boot_time := uint64 (time.Now().Unix() - sysinfo.Uptime)
        sysstate.BootTime = &boot_time
    }
}


var DbToYang_sys_state_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error

    sysObj := getAppRootObject(inParams)

    jsonsystem, err := getSystemInfoFromFile()
    if err != nil {
        log.Infof("getSystemInfoFromFile failed")
        return err
    }
    ygot.BuildEmptyTree(sysObj)
    getSystemState(&jsonsystem, sysObj.State)
    return err;
}

func getSystemMemory (sys *JSONSystem, sysmem *ocbinds.OpenconfigSystem_System_Memory_State) () {
    log.Infof("getSystemMemory Entry")

    sysmem.Physical = &sys.Total
    sysmem.Reserved = &sys.Used
}

var DbToYang_sys_memory_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error

    sysObj := getAppRootObject(inParams)

    jsonsystem, err := getSystemInfoFromFile()
    if err != nil {
        log.Infof("getSystemInfoFromFile failed")
        return err
    }
    ygot.BuildEmptyTree(sysObj)

    sysObj.Memory.State = &ocbinds.OpenconfigSystem_System_Memory_State{}
    getSystemMemory(&jsonsystem, sysObj.Memory.State)
    return err;
}

func getSystemCpu (idx int, cpu Cpu, syscpus *ocbinds.OpenconfigSystem_System_Cpus) {
    log.Infof("getSystemCpu Entry idx ", idx)

    sysinfo := syscall.Sysinfo_t{}
    sys_err := syscall.Sysinfo(&sysinfo)
    if sys_err != nil {
        log.Infof("syscall.Sysinfo failed.")
    }

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

func getSystemCpus (sys *JSONSystem, syscpus *ocbinds.OpenconfigSystem_System_Cpus) {
    log.Infof("getSystemCpus Entry")

    sysinfo := syscall.Sysinfo_t{}
    sys_err := syscall.Sysinfo(&sysinfo)
    if sys_err != nil {
        log.Infof("syscall.Sysinfo failed.")
    }

    for  idx, cpu := range sys.Cpus {
        getSystemCpu(idx, cpu, syscpus)
    }
}

var DbToYang_sys_cpus_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error

    sysObj := getAppRootObject(inParams)

    jsonsystem, err := getSystemInfoFromFile()
    if err != nil {
        log.Infof("getSystemInfoFromFile failed")
        return err
    }

    ygot.BuildEmptyTree(sysObj)

    path := NewPathInfo(inParams.uri) 
    val := path.Vars["index"]
    if len(val) != 0 {
        cpu, _ := strconv.Atoi(val)
        log.Info("Cpu id: ", cpu, ", max is ", len(jsonsystem.Cpus))
        if cpu >=0 && cpu < len(jsonsystem.Cpus) {
            getSystemCpu(cpu, jsonsystem.Cpus[cpu], sysObj.Cpus)	
        } else {
            log.Info("Cpu id: ", cpu, "is invalid, max is ", len(jsonsystem.Cpus))
        }
    } else {
        getSystemCpus(&jsonsystem, sysObj.Cpus)
    }
    return err;
}

func getSystemProcess (proc *Proc, sysproc *ocbinds.OpenconfigSystem_System_Processes_Process, pid uint64) {

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

func getSystemProcesses (sys *JSONSystem, sysprocs *ocbinds.OpenconfigSystem_System_Processes, pid uint64) {
    log.Infof("getSystemProcesses Entry")

    if pid != 0 {
        proc := sys.Procs[strconv.Itoa(int(pid))]
        sysproc := sysprocs.Process[pid]

        getSystemProcess(&proc, sysproc, pid)
    } else {

        for  pidstr,  proc := range sys.Procs {
            idx, _:= strconv.Atoi(pidstr)

            sysproc, err := sysprocs.NewProcess(uint64 (idx))
            if err != nil {
                log.Infof("sysprocs.NewProcess failed")
                return
            }

            getSystemProcess(&proc, sysproc, uint64 (idx))
        }
    }
    return
}
var DbToYang_sys_procs_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error

    sysObj := getAppRootObject(inParams)

    jsonsystem, err := getSystemInfoFromFile()
    if err != nil {
        log.Infof("getSystemInfoFromFile failed")
        return err
    }

    ygot.BuildEmptyTree(sysObj)
    path := NewPathInfo(inParams.uri) 
    val := path.Vars["pid"]
    pid := 0
    if len(val) != 0 {
        pid, _ = strconv.Atoi(val)
    }
    getSystemProcesses(&jsonsystem, sysObj.Processes, uint64(pid))
    return err;
}

