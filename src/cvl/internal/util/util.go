package util

import (
	"os"
	"fmt"
	"runtime"
	log "github.com/golang/glog"
)

var CVL_SCHEMA string = "schema/"

//package init function 
func init() {
	if (os.Getenv("CVL_SCHEMA_PATH") != "") {
		CVL_SCHEMA = os.Getenv("CVL_SCHEMA_PATH") + "/"
	}
}

var Tracing bool = false

var traceFlags uint16 = 0

func SetTrace(on bool) {
	if (on == true) {
		traceFlags = 1
	} else {
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

func SetTraceLevel(level uint8) {
	traceFlags = traceFlags | (1 << level)
}

func ClearTraceLevel(level uint8) {
	traceFlags = traceFlags &^ (1 << level)
}

func TRACE_LOG(level log.Level, fmtStr string, args ...interface{}) {
	if (IsTraceSet() == false) {
		return
	}

	if IsTraceSet() == true {
		pc := make([]uintptr, 10)
		runtime.Callers(2, pc)
		f := runtime.FuncForPC(pc[0])
		file, line := f.FileLine(pc[0])

		fmt.Printf("%s:%d %s(): ", file, line, f.Name())
		fmt.Printf(fmtStr, args...)
	} else {
		log.V(level).Infof(fmtStr, args...)
	}
}

