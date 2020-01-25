package transformer

import (
    "strings"
    "translib/db"
    log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("network_instance_protocols_ptotocol_table_name_xfmr", network_instance_protocols_ptotocol_table_name_xfmr)
    XlateFuncBind("YangToDb_network_instance_protocol_key_xfmr", YangToDb_network_instance_protocol_key_xfmr)
    XlateFuncBind("DbToYang_network_instance_protocol_key_xfmr", DbToYang_network_instance_protocol_key_xfmr)

}

var network_instance_protocols_ptotocol_table_name_xfmr TableXfmrFunc = func (inParams XfmrParams)  ([]string, error) {
    var tblList []string

    log.Info("network_instance_protocols_protocol_table_name_xfmr")
    if (inParams.oper == GET) {
        if(inParams.dbDataMap != nil) {
            (*inParams.dbDataMap)[db.ConfigDB]["CFG_PROTO_TBL"] = make(map[string]db.Value)
            (*inParams.dbDataMap)[db.ConfigDB]["CFG_PROTO_TBL"]["BGP|bgp"] = db.Value{Field: make(map[string]string)}
            (*inParams.dbDataMap)[db.ConfigDB]["CFG_PROTO_TBL"]["BGP|bgp"].Field["NULL"] = "NULL"
            tblList = append(tblList, "CFG_PROTO_TBL")
        }
    }
    return tblList, nil
}

var YangToDb_network_instance_protocol_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var key string
    log.Info("YangToDb_network_instance_protocol_key_xfmr - URI: ", inParams.uri)
    if (inParams.oper == GET) {
        pathInfo := NewPathInfo(inParams.uri)
        protoId := pathInfo.Var("identifier")
        protoName := pathInfo.Var("name#2")
        key = protoId+"|"+protoName
    }
    log.Info("returned Key: ", key)
    return key, nil
}

var DbToYang_network_instance_protocol_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    dynKey := strings.Split(entry_key, "|")
    rmap["identifier"] = dynKey[0]
    rmap["name"] = dynKey[1]
    return rmap, nil
}
