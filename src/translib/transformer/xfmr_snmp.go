package transformer
 
import (
    "translib/db"
)
 
func init() {
    XlateFuncBind("YangToDb_snmp_target_params_key_xfmr", YangToDb_snmp_target_params_key_xfmr)
    XlateFuncBind("DbToYang_snmp_target_params_key_xfmr", DbToYang_snmp_target_params_key_xfmr)
}
 
var YangToDb_snmp_target_params_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
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
    result := make(map[string]interface{})
    result["name"] = inParams.param
    return result, nil
}
