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

type CVLLogLevel uint8 

const (
        INFO  = 0 + iota
        WARNING
        ERROR
        FATAL
        INFO_API
	INFO_TRACE
	INFO_DEBUG
	INFO_DATA
	INFO_DETAIL
	INFO_ALL
)

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

func IsLogTraceSet() bool {
	return true
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

func CVL_LOG(level CVLLogLevel, format string, args ...interface{}) {

	if (IsLogTraceSet() == false) {
		return
	}

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

