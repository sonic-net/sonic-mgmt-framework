package transformer
 
import (
 "strings"
  log "github.com/golang/glog"
)
 
func init() {
  XlateFuncBind("YangToDb_snmp_engine_key_xfmr", YangToDb_snmp_engine_key_xfmr)
  XlateFuncBind("YangToDb_snmp_group_name_xfmr", YangToDb_snmp_group_name_xfmr)
  XlateFuncBind("YangToDb_snmp_member_key_xfmr", YangToDb_snmp_member_key_xfmr)
  XlateFuncBind("DbToYang_snmp_member_key_xfmr", DbToYang_snmp_member_key_xfmr)
}
 
var YangToDb_snmp_engine_key_xfmr = func(inParams XfmrParams) (string, error) {
  log.Info("YangToDb_snmp_engine_key_xfmr            uri: ", inParams.uri)
  return "GLOBAL", nil
}

func YangToDb_snmp_group_name_xfmr(inParams XfmrParams) (map[string]string, error) {
  data := map[string]string{ "NULL": "NULL" }
  log.Info("*** YangToDb_snmp_group_name_xfmr        map: ", data)
  return data, nil
}

var YangToDb_snmp_member_key_xfmr = func(inParams XfmrParams) (string, error) {
  var entry_key string
  log.Info("YangToDb_snmp_member_key_xfmr            uri: ", inParams.uri)
  log.Info("YangToDb_snmp_member_key_xfmr            key: ", inParams.key)

  pathInfo := NewPathInfo(inParams.uri)
  gName := pathInfo.Var("name")
  sName := pathInfo.Var("security-name")

  if len(sName) == 0 {
    entry_key = gName
  } else {
    entry_key = gName + "|" + sName
  }

  log.Info("YangToDb_snmp_member_key_xfmr   Key Returned: ", entry_key)
  return entry_key, nil
}

var DbToYang_snmp_member_key_xfmr = func(inParams XfmrParams) (map[string]interface{}, error) {
  rmap := make(map[string]interface{})
  log.Info("DbToYang_snmp_member_key_xfmr            uri: ", inParams.uri)
  log.Info("DbToYang_snmp_member_key_xfmr            key: ", inParams.key)

  keys := strings.Split(inParams.key, "|")
  secName := keys[1]
  rmap["security-name"] = secName
  log.Info("DbToYang_snmp_member_key_xfmr   Key Returned: ", rmap)
  return rmap, nil
}
