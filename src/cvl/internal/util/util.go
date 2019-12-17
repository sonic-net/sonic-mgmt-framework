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

package util

/*
#cgo LDFLAGS: -lyang
#include <libyang/libyang.h>

extern void customLogCallback(LY_LOG_LEVEL, char* msg, char* path);

static void customLogCb(LY_LOG_LEVEL level, const char* msg, const char* path) {
	customLogCallback(level, (char*)msg, (char*)path);
}

static void ly_set_log_callback(int enable) {
	if (enable == 1) {
		ly_verb(LY_LLDBG);
		ly_set_log_clb(customLogCb, 0);
	} else {
		ly_verb(LY_LLERR);
		ly_set_log_clb(NULL, 0);
	}
}

*/
import "C"
import (
	"os"
	"fmt"
	"io"
	"runtime"
	 "encoding/json"
        "io/ioutil"
        "os/signal"
        "syscall"
	"strings"
	"flag"
	"strconv"
	log "github.com/golang/glog"
	fileLog "log"
	"sync"
)

var CVL_SCHEMA string = "schema/"
var CVL_CFG_FILE string = "/usr/sbin/cvl_cfg.json"
const CVL_LOG_FILE = "/tmp/cvl.log"

//package init function 
func init() {
	if (os.Getenv("CVL_SCHEMA_PATH") != "") {
		CVL_SCHEMA = os.Getenv("CVL_SCHEMA_PATH") + "/"
	}

	if (os.Getenv("CVL_CFG_FILE") != "") {
		CVL_CFG_FILE = os.Getenv("CVL_CFG_FILE")
	}

	//Initialize mutex
	logFileMutex = &sync.Mutex{}
}

var cvlCfgMap map[string]string
var isLogToFile bool
var logFileSize int
var pLogFile *os.File
var logFileMutex *sync.Mutex

/* Logging Level for CVL global logging. */
type CVLLogLevel uint8
const (
        INFO  = 0 + iota
        WARNING
        ERROR
        FATAL
	INFO_DEBUG
        INFO_API
	INFO_DATA
	INFO_DETAIL
	INFO_TRACE
	INFO_ALL
)

var cvlTraceFlags uint32

/* Logging levels for CVL Tracing. */
type CVLTraceLevel uint32 
const (
	TRACE_MIN = 0
	TRACE_MAX = 8 
        TRACE_CACHE  = 1 << TRACE_MIN 
        TRACE_LIBYANG = 1 << 1
        TRACE_YPARSER = 1 << 2
        TRACE_CREATE = 1 << 3
        TRACE_UPDATE = 1 << 4
        TRACE_DELETE = 1 << 5
        TRACE_SEMANTIC = 1 << 6
        TRACE_ONERROR = 1 << 7 
        TRACE_SYNTAX = 1 << TRACE_MAX 

)


var traceLevelMap = map[int]string {
	/* Caching operation traces */
	TRACE_CACHE : "TRACE_CACHE",
	/* Libyang library traces. */
	TRACE_LIBYANG: "TRACE_LIBYANG",
	/* Yang Parser traces. */
	TRACE_YPARSER : "TRACE_YPARSER", 
	/* Create operation traces. */
	TRACE_CREATE : "TRACE_CREATE", 
	/* Update operation traces. */
	TRACE_UPDATE : "TRACE_UPDATE", 
	/* Delete operation traces. */
	TRACE_DELETE : "TRACE_DELETE", 
	/* Semantic Validation traces. */
	TRACE_SEMANTIC : "TRACE_SEMANTIC",
	/* Syntax Validation traces. */
	TRACE_SYNTAX : "TRACE_SYNTAX", 
	/* Trace on Error. */
	TRACE_ONERROR : "TRACE_ONERROR",
}

var Tracing bool = false

var traceFlags uint16 = 0

func SetTrace(on bool) {
	if (on == true) {
		Tracing = true
		traceFlags = 1
	} else {
		Tracing = false 
		traceFlags = 0
	}
}

func IsTraceSet() bool {
	if (traceFlags == 0) {
		return false
	} else {
		return true
	}
}

/* The following function enbles the libyang logging by
changing libyang's global log setting */

//export customLogCallback
func customLogCallback(level C.LY_LOG_LEVEL, msg *C.char, path *C.char)  {
	TRACE_LEVEL_LOG(TRACE_YPARSER, "[libyang] %s (path: %s)",
	C.GoString(msg), C.GoString(path))
}

func IsTraceLevelSet(tracelevel CVLTraceLevel) bool {
	if (cvlTraceFlags & (uint32)(tracelevel)) != 0 {
		return true
	}

	return false
}

func TRACE_LEVEL_LOG(tracelevel CVLTraceLevel, fmtStr string, args ...interface{}) {

	/*
	if (IsTraceSet() == false) {
		return
	}

	level = (level - INFO_API) + 1;
	*/

	traceEnabled := false
	if ((cvlTraceFlags & (uint32)(tracelevel)) != 0) {
		traceEnabled = true
	}
	if (traceEnabled == true) && (isLogToFile == true) {
		logToCvlFile(fmtStr, args...)
		return
	}

	if IsTraceSet() == true && traceEnabled == true {
		pc := make([]uintptr, 10)
		runtime.Callers(2, pc)
		f := runtime.FuncForPC(pc[0])
		file, line := f.FileLine(pc[0])

		fmt.Printf("%s:%d [CVL] : %s(): ", file, line, f.Name())
		fmt.Printf(fmtStr+"\n", args...)
	} else {
		if (traceEnabled == true) {
			fmtStr = "[CVL] : " + fmtStr
			//Trace logs has verbose level INFO_TRACE
			log.V(INFO_TRACE).Infof(fmtStr, args...)
		}
	}
}

//Logs to /tmp/cvl.log file
func logToCvlFile(format string, args ...interface{}) {
	if (pLogFile == nil) {
		return
	}

	logFileMutex.Lock()
	if (logFileSize == 0) {
		fileLog.Printf(format, args...)
		logFileMutex.Unlock()
		return
	}

	fStat, err := pLogFile.Stat()

	var curSize int64 = 0
	if (err == nil) && (fStat != nil) {
		curSize = fStat.Size()
	}

	// Roll over the file contents if size execeeds max defined limit
	if (curSize >= int64(logFileSize)) {
		//Write 70% contents from bottom and write to top
		//Truncate 30% of bottom

		//close the file first
		pLogFile.Close()

		pFile, err := os.OpenFile(CVL_LOG_FILE,
		os.O_RDONLY, 0666)
		pFileOut, errOut := os.OpenFile(CVL_LOG_FILE + ".tmp",
		os.O_WRONLY | os.O_CREATE, 0666)


		if (err != nil) && (errOut != nil) {
			fileLog.Printf("Failed to roll over the file, current size %v", curSize)
		} else {
			pFile.Seek(int64(logFileSize * 30/100), io.SeekStart)
			_, err := io.Copy(pFileOut, pFile)
			if err == nil {
				os.Rename(CVL_LOG_FILE + ".tmp", CVL_LOG_FILE)
			}
		}

		if (pFile != nil) {
			pFile.Close()
		}
		if (pFileOut != nil) {
			pFileOut.Close()
		}

		// Reopen the file 
		pLogFile, err := os.OpenFile(CVL_LOG_FILE,
		os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Error in opening log file %s, %v", CVL_LOG_FILE, err)
		} else {
			fileLog.SetOutput(pLogFile)
		}
	}


	fileLog.Printf(format, args...)

	logFileMutex.Unlock()
}

func CVL_LEVEL_LOG(level CVLLogLevel, format string, args ...interface{}) {

	if (isLogToFile == true) {
		logToCvlFile(format, args...)
		return
	}

	format = "[CVL] : " + format

	switch level {
		case INFO:
		       log.Infof(format, args...)
		case  WARNING:
		       log.Warningf(format, args...)
		case  ERROR:
		       log.Errorf(format, args...)
		case  FATAL:
		       log.Fatalf(format, args...)
		case INFO_API:
			log.V(1).Infof(format, args...)
		case INFO_TRACE:
			log.V(2).Infof(format, args...)
		case INFO_DEBUG:
			log.V(3).Infof(format, args...)
		case INFO_DATA:
			log.V(4).Infof(format, args...)
		case INFO_DETAIL:
			log.V(5).Infof(format, args...)
		case INFO_ALL:
			log.V(6).Infof(format, args...)
	}

}

// Function to check CVL log file related settings
func applyCvlLogFileConfig() {

	if (pLogFile != nil) {
		pLogFile.Close()
		pLogFile = nil
	}

	//Disable libyang trace log
	C.ly_set_log_callback(0)
	isLogToFile = false
	logFileSize = 0

	enabled, exists := cvlCfgMap["LOG_TO_FILE"]
	if exists == false {
		return
	}

	if fileSize, sizeExists := cvlCfgMap["LOG_FILE_SIZE"];
	sizeExists == true {
		logFileSize, _ = strconv.Atoi(fileSize)
	}

	if (enabled == "true") {
		pFile, err := os.OpenFile(CVL_LOG_FILE,
		os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)

		if err != nil {
			fmt.Printf("Error in opening log file %s, %v", CVL_LOG_FILE, err)
		} else {
			pLogFile = pFile
			fileLog.SetOutput(pLogFile)
			isLogToFile = true
		}

		//Enable libyang trace log
		C.ly_set_log_callback(1)
	}
}

func ConfigFileSyncHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR2)
	go func() {
		for {
			<-sigs
			cvlCfgMap := ReadConfFile()

			if cvlCfgMap == nil {
				return
			}

			CVL_LEVEL_LOG(INFO ,"Received SIGUSR2. Changed configuration values are %v", cvlCfgMap)


			flag.Set("v", cvlCfgMap["VERBOSITY"])
			if (strings.Compare(cvlCfgMap["LOGTOSTDERR"], "true") == 0) {
				SetTrace(true)
				flag.Set("logtostderr", "true")
				flag.Set("stderrthreshold", cvlCfgMap["STDERRTHRESHOLD"])
			}

		}
	}()

}

func ReadConfFile()  map[string]string{

	/* Return if CVL configuration file is not present. */
	if _, err := os.Stat(CVL_CFG_FILE); os.IsNotExist(err) {
		return nil
	}

	data, err := ioutil.ReadFile(CVL_CFG_FILE)

	err = json.Unmarshal(data, &cvlCfgMap)

	if err != nil {
		CVL_LEVEL_LOG(INFO ,"Error in reading cvl configuration file %v", err)
		return nil
	}

	CVL_LEVEL_LOG(INFO ,"Current Values of CVL Configuration File %v", cvlCfgMap)
	var index uint32

	for  index = TRACE_MIN ; index <= TRACE_MAX ; index++  {
		if (strings.Compare(cvlCfgMap[traceLevelMap[1 << index]], "true") == 0) {
			cvlTraceFlags = cvlTraceFlags |  (1 << index) 
		}
	}

	applyCvlLogFileConfig()

	return cvlCfgMap
}

func SkipValidation() bool {
	val, existing := cvlCfgMap["SKIP_VALIDATION"]
	if (existing == true) && (val == "true") {
		return true
	}

	return false
}

func SkipSemanticValidation() bool {
	val, existing := cvlCfgMap["SKIP_SEMANTIC_VALIDATION"]
	if (existing == true) && (val == "true") {
		return true
	}

	return false
}
