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
    "encoding/json"
	"translib/ocbinds"
    "translib/db"
    "os/exec"

	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
)

func init () {
    XlateFuncBind("DbToYang_lacp_get_xfmr", DbToYang_lacp_get_xfmr)
}

func getLacpRoot (s *ygot.GoStruct) *ocbinds.OpenconfigLacp_Lacp {
    deviceObj := (*s).(*ocbinds.Device)
    return deviceObj.Lacp
}

func populateLacpData(ifKey string, state *ocbinds.OpenconfigLacp_Lacp_Interfaces_Interface_State,
                                    members *ocbinds.OpenconfigLacp_Lacp_Interfaces_Interface_Members) error {
	var err error
    cmd := exec.Command("docker", "exec", "teamd", "teamdctl", ifKey, "state", "dump")
	out_stream, e := cmd.StdoutPipe()
    if e != nil {
        log.Fatalf("Can't get stdout pipe: %s\n", err)
        return e
	}
    err = cmd.Start()
    if err != nil {
        log.Fatalf("cmd.Start() failed with %s\n", err)
        return err
    }

    var TeamdJson map[string]interface{}

    err = json.NewDecoder(out_stream).Decode(&TeamdJson)
    if err != nil {
        log.Fatalf("Not able to decode teamd json output: %s\n", err)
        return err
    }

    err = cmd.Wait()
    if err != nil {
       log.Fatalf("Command execution completion failed with %s\n", err)
        return err
    }

    runner_map := TeamdJson["runner"].(map[string]interface{})

    prio := runner_map["sys_prio"].(float64)
    sys_prio := uint16(prio)
    state.SystemPriority = &sys_prio

    fast_rate := runner_map["fast_rate"].(bool)
    if fast_rate {
	state.Interval = ocbinds.OpenconfigLacp_LacpPeriodType_FAST
    } else {
        state.Interval = ocbinds.OpenconfigLacp_LacpPeriodType_SLOW
    }

    active := runner_map["active"].(bool)
    if active {
        state.LacpMode = ocbinds.OpenconfigLacp_LacpActivityType_ACTIVE
    } else {
        state.LacpMode = ocbinds.OpenconfigLacp_LacpActivityType_PASSIVE
    }

    team_device := TeamdJson["team_device"].(map[string]interface{})
    team_device_ifinfo := team_device["ifinfo"].(map[string]interface{})
    SystemIdMac := team_device_ifinfo["dev_addr"].(string)
    state.SystemIdMac = &SystemIdMac

    var lacpMemberObj *ocbinds.OpenconfigLacp_Lacp_Interfaces_Interface_Members_Member

    ygot.BuildEmptyTree(members)

    if ports_map,ok := TeamdJson["ports"].(map[string]interface{}); ok {
        for ifKey := range ports_map {
			log.Infof("----------------------------Build Empty Tree for %s \n", ifKey)
			if lacpMemberObj, ok = members.Member[ifKey]; !ok {
				lacpMemberObj, err = members.NewMember(ifKey)
				if err != nil {
					log.Error("Creation of portchannel member subtree failed")
					return err
				}
				ygot.BuildEmptyTree(lacpMemberObj)
			}
            
           member_map := ports_map[ifKey].(map[string]interface{})
           port_runner := member_map["runner"].(map[string]interface{})

   			selected := port_runner["selected"].(bool)
			lacpMemberObj.State.Selected = &selected

           actor := port_runner["actor_lacpdu_info"].(map[string]interface{})

           port_num := actor["port"].(float64)
           pport_num := uint16(port_num)
           lacpMemberObj.State.PortNum = &pport_num

           system_id := actor["system"].(string)
           lacpMemberObj.State.SystemId = &system_id

           oper_key := actor["key"].(float64)
           ooper_key := uint16(oper_key)
           lacpMemberObj.State.OperKey = &ooper_key

           partner := port_runner["partner_lacpdu_info"].(map[string]interface{})
           partner_port_num := partner["port"].(float64)
           ppartner_num := uint16(partner_port_num)
           lacpMemberObj.State.PartnerPortNum = &ppartner_num

           partner_system_id := partner["system"].(string)
           lacpMemberObj.State.PartnerId = &partner_system_id

           partner_oper_key := partner["key"].(float64)
           ppartner_key := uint16(partner_oper_key)
           lacpMemberObj.State.PartnerKey = &ppartner_key

        }
    }

    log.Infof("----------------------------Successfully populated portchannel data for %s\n", ifKey)

    return err
}

var DbToYang_lacp_get_xfmr  SubTreeXfmrDbToYang = func(inParams XfmrParams) error {

    lacpIntfsObj := getLacpRoot(inParams.ygRoot)
    pathInfo := NewPathInfo(inParams.uri)
    ifKey := pathInfo.Var("name")

    targetUriPath, err := getYangPathFromUri(pathInfo.Path)

    log.Infof("------------Received GET for path: %s; template: %s vars: %v targetUriPath: %s ifKey: %s", pathInfo.Path, pathInfo.Template, pathInfo.Vars, targetUriPath, ifKey)

    var ok bool
    var lacpintfObj *ocbinds.OpenconfigLacp_Lacp_Interfaces_Interface

    if isSubtreeRequest(targetUriPath, "/openconfig-lacp:lacp/interfaces/interface") {
        log.Infof("----------------------Inside specific portchannel request")

        /* Request for a specific portchannel */
        if lacpIntfsObj.Interfaces.Interface != nil && len(lacpIntfsObj.Interfaces.Interface) > 0 && ifKey != "" {
            lacpintfObj, ok = lacpIntfsObj.Interfaces.Interface[ifKey]
            if !ok {
                lacpintfObj, _ = lacpIntfsObj.Interfaces.NewInterface(ifKey)
            }
             ygot.BuildEmptyTree(lacpintfObj)

             log.Infof("---------------------------About to populate LACP data for %s\n", ifKey)
             return populateLacpData(ifKey, lacpintfObj.State, lacpintfObj.Members)
        }
     } else if isSubtreeRequest(targetUriPath, "/openconfig-lacp:lacp/interfaces") {
        log.Infof("-------------------------------Inside all portchannel request")

        ygot.BuildEmptyTree(lacpIntfsObj)

        var lagTblTs = &db.TableSpec{Name: "LAG_TABLE"}
        var appDb = inParams.dbs[db.ApplDB]
        tbl, err := appDb.GetTable(lagTblTs)

        if err != nil {
            log.Error("App-DB get for list of portchannels failed!")
            return err
        }
        keys, _ := tbl.GetKeys()
		log.Infof("-------------------ALL KEYS: ", keys)
        for _, key := range keys {
		   log.Infof("--------------KEYS: ", key)
           ifKey := key.Get(0)
           log.Infof("PortChannel: %s\n", ifKey)

           lacpintfObj, ok = lacpIntfsObj.Interfaces.Interface[ifKey]
           if !ok {
              lacpintfObj, _ = lacpIntfsObj.Interfaces.NewInterface(ifKey)
           }
           ygot.BuildEmptyTree(lacpintfObj)

           log.Infof("---------------------------About to populate LACP data for %s\n", ifKey)
           populateLacpData(ifKey, lacpintfObj.State, lacpintfObj.Members)
        }
     }

    return err

}



