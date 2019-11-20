package transformer

import (
	log "github.com/golang/glog"
)

func init() {
	XlateFuncBind("YangToDb_udld_global_key_xfmr", YangToDb_udld_global_key_xfmr)
}

var YangToDb_udld_global_key_xfmr = func(inParams XfmrParams) (string, error) {
	log.Info("YangToDb_udld_global_key_xfmr: ", inParams.ygRoot, inParams.uri)
	return "GLOBAL", nil
}
