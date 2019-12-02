package transformer

import (
        log "github.com/golang/glog"
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
    XlateFuncBind("YangToDb_tacacs_global_set_key_xfmr", YangToDb_tacacs_global_set_key_xfmr)
    XlateFuncBind("server_table_xfmr", server_table_xfmr)
    XlateFuncBind("tacacs_global_table_xfmr", tacacs_global_table_xfmr)
}

var YangToDb_auth_set_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    return "authentication", nil
}

var YangToDb_server_set_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    pathInfo := NewPathInfo(inParams.uri)
    serverkey := pathInfo.Var("address")

    return serverkey, nil
}

var YangToDb_tacacs_global_set_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    return "global", nil
}

var tacacs_global_table_xfmr TableXfmrFunc = func (inParams XfmrParams) ([]string, error) {

    var tblList []string
    var err error

    log.Infof("tacacs_global_table_xfmr - Uri: ", inParams.uri);
    pathInfo := NewPathInfo(inParams.uri)

    targetUriPath, err := getYangPathFromUri(pathInfo.Path)

    aaaType := pathInfo.Var("name");
    log.Info("TableXfmrFunc - targetUriPath : ", targetUriPath)
    log.Info("TableXfmrFunc - type : ", aaaType)

    if (aaaType == "TACACS") {
        tblList = append(tblList, "TACPLUS")
    } else if (aaaType == "RADIUS") {
        tblList = append(tblList, "RADIUS")
    }

    log.Infof("TableXfmrFunc - uri(%v), tblList(%v)\r\n", inParams.uri, tblList);
    return tblList, err
}


var server_table_xfmr TableXfmrFunc = func (inParams XfmrParams) ([]string, error) {

    var tblList []string
    var err error

    log.Infof("server_global_table_xfmr - Uri: ", inParams.uri);
    pathInfo := NewPathInfo(inParams.uri)

    targetUriPath, err := getYangPathFromUri(pathInfo.Path)

    aaaType := pathInfo.Var("name");
    log.Info("TableXfmrFunc - targetUriPath : ", targetUriPath)
    log.Info("TableXfmrFunc - type : ", aaaType)

    if (aaaType == "TACACS") {
        tblList = append(tblList, "TACPLUS_SERVER")
    } else if (aaaType == "RADIUS") {
        tblList = append(tblList, "RADIUS")
    }
//   else {
//        err = errors.New("Invalid URI")
//    }

    log.Infof("TableXfmrFunc - uri(%v), tblList(%v)\r\n", inParams.uri, tblList);
    return tblList, err
}

