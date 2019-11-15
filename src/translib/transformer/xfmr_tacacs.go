package transformer

import (
        "github.com/openconfig/ygot/ygot"
        "translib/ocbinds"
)

func getSystemRoot(s *ygot.GoStruct) *ocbinds.OpenconfigSystem_System {
    deviceObj := (*s).(*ocbinds.Device)
    return deviceObj.System
}

func init() {
    XlateFuncBind("YangToDb_auth_set_key_xfmr", YangToDb_auth_set_key_xfmr)
    XlateFuncBind("YangToDb_server_set_key_xfmr", YangToDb_server_set_key_xfmr)
    XlateFuncBind("YangToDb_global_set_key_xfmr", YangToDb_global_set_key_xfmr)
}

var YangToDb_auth_set_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    return "authentication", nil
}

var YangToDb_server_set_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    pathInfo := NewPathInfo(inParams.uri)
    serverkey := pathInfo.Var("address")

    return serverkey, nil
}

var YangToDb_global_set_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    return "global", nil
}

