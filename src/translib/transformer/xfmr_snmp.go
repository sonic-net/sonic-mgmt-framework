package transformer
 
import (
//  "translib/db"
  log "github.com/golang/glog"
)
 
func init() {
  XlateFuncBind("YangToDb_snmp_engine_key_xfmr", YangToDb_snmp_engine_key_xfmr)
  XlateFuncBind("YangToDb_snmp_group_name_xfmr", YangToDb_snmp_group_name_xfmr)
}
 
var YangToDb_snmp_engine_key_xfmr = func(inParams XfmrParams) (string, error) {
  log.Info("YangToDb_snmp_global_key_xfmr: ", inParams.ygRoot, inParams.uri)
  return "GLOBAL", nil
}

func YangToDb_snmp_group_name_xfmr(inParams XfmrParams) (map[string]string, error) {
  data := map[string]string{ "NULL": "NULL" }
  return data, nil
}
