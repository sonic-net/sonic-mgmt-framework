package transformer
 
import (
  "translib/db"
  log "github.com/golang/glog"
)
 
func init() {
  XlateFuncBind("YangToDb_snmp_engine_key_xfmr", YangToDb_snmp_engine_key_xfmr)
  XlateFuncBind("YangToDb_snmp_target_params_key_xfmr", YangToDb_snmp_target_params_key_xfmr)
  XlateFuncBind("DbToYang_snmp_target_params_key_xfmr", DbToYang_snmp_target_params_key_xfmr)
}
 
var YangToDb_snmp_engine_key_xfmr = func(inParams XfmrParams) (string, error) {
  log.Info("YangToDb_snmp_global_key_xfmr: ", inParams.ygRoot, inParams.uri)
  return "GLOBAL", nil
}

var YangToDb_snmp_target_params_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
  log.Info("YangToDb_snmp_target_params_key_xfmr: ", inParams.ygRoot, inParams.uri)
  curDb := db.ConfigDB
  val, _ := inParams.param.(*string)
  subOpDataMap := make(map[db.DBNum]map[string]map[string]db.Value)
  subOpDataMap[curDb] = make(map[string]map[string]db.Value)
  subOpDataMap[curDb]["SNMP_SERVER_PARAMS"] = make(map[string]db.Value)
  subOpDataMap[curDb]["SNMP_SERVER_PARAMS"][*val] = db.Value{Field: make(map[string]string)}
  inParams.subOpDataMap[inParams.oper] = &subOpDataMap
  return "", nil
}
 
var DbToYang_snmp_target_params_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
  log.Info("DbToYang_snmp_target_params_key_xfmr: ", inParams.ygRoot, inParams.uri)
  result := make(map[string]interface{})
  result["name"] = inParams.param
  return result, nil
}
