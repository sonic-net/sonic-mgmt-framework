////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Dell, Inc.                                                 //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//  http://www.apache.org/licenses/LICENSE-2.0                                //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package transformer

import (
    log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("YangToDb_sag_global_key_xfmr", YangToDb_sag_global_key_xfmr)
    XlateFuncBind("DbToYang_sag_global_key_xfmr", DbToYang_sag_global_key_xfmr)	
}


var YangToDb_sag_global_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    /* var err error */

    log.Info("YangToDb_sag_global_key_xfmr ***", inParams.uri)
    /* pathInfo := NewPathInfo(inParams.uri) */

    /* Key should contain, <network-instance-name> */

    var sagTableKey string

    sagTableKey = "IP"

    log.Info("YangToDb_sag_global_key_xfmr: sagTableKey:", sagTableKey)
    return sagTableKey, nil
}

var DbToYang_sag_global_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_sag_global_key_xfmr: ", entry_key)

    rmap["name"] = "default"

    log.Info("DbToYang_sag_global_key_xfmr")

    return rmap, nil
}

