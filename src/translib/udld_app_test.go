////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package translib

import (
	"errors"
	"fmt"
	"testing"
	db "translib/db"
)

func init() {
	fmt.Println("+++++  Init udld_app_test  +++++")

	if err := clearUdldDataFromConfigDb(); err == nil {
		fmt.Println("+++++  Removed All UDLD Data from Db  +++++")
	} else {
		fmt.Printf("Failed to remove All UDLD Data from Db: %v", err)
	}
}

func Test_UdldApp_Udld_Global_Enable_Disable(t *testing.T) {
	topUdldUrl := "/sonic-udld:sonic-udld"

	t.Run("Empty_Response_Top_Level", processGetRequest(topUdldUrl, emptyJson, true))

	t.Run("Enable_UDLD_Global_Level", processSetRequest(topUdldUrl+"/UDLD", enableUdldGlobalJsonRequest, "POST", false))
	t.Run("Verify_UDLD_Global_Enabled", processGetRequest(topUdldUrl+"/UDLD", udldGlobalEnabledJsonResponse, false))
	t.Run("Disable_UDLD_Global_Level", processDeleteRequest(topUdldUrl))

	t.Run("Verify_UDLD_Disabled_Global_Level", processGetRequest(topUdldUrl, emptyJson, true))
}

func Test_UdldApp_Udld_Port_Level_Enable_Disable(t *testing.T) {
	topUdldUrl := "/sonic-udld:sonic-udld"
	portLevelUdldUrl := "/sonic-udld:sonic-udld/UDLD_PORT"

	t.Run("Empty_Response_Port_Level", processGetRequest(portLevelUdldUrl, emptyJson, true))
	t.Run("Enable_UDLD_Global_Level", processSetRequest(topUdldUrl+"/UDLD", enableUdldGlobalJsonRequest, "POST", false))

	t.Run("Enable_UDLD_Port_Level", processSetRequest(portLevelUdldUrl, enableUdldPortJsonRequest, "POST", false))
	t.Run("Verify_UDLD_Port_Level_Enabled", processGetRequest(portLevelUdldUrl, udldPortEnabledJsonResponse, false))
	t.Run("Disable_UDLD_Port_Level", processDeleteRequest(portLevelUdldUrl+"/UDLD_PORT_LIST[ifname=Ethernet28]"))
	t.Run("Verify_UDLD_Disabled_Port_Level", processGetRequest(portLevelUdldUrl, emptyJson, true))

	t.Run("Disable_UDLD_Global_Level", processDeleteRequest(topUdldUrl))
	t.Run("Verify_UDLD_Disabled_Global_Level", processGetRequest(topUdldUrl, emptyJson, true))
}

func clearUdldDataFromConfigDb() error {
	var err error
	udldGlobalTbl := db.TableSpec{Name: "UDLD"}
	udldPortTbl := db.TableSpec{Name: "UDLD_PORT"}

	d := getConfigDb()
	if d == nil {
		err = errors.New("Failed to connect to config Db")
		return err
	}

	if err = d.DeleteTable(&udldPortTbl); err != nil {
		err = errors.New("Failed to delete UDLD Port Table")
		return err
	}

	if err = d.DeleteTable(&udldGlobalTbl); err != nil {
		err = errors.New("Failed to delete UDLD Global Table")
		return err
	}

	return err
}

/***************************************************************************/
///////////                  JSON Data for Tests              ///////////////
/***************************************************************************/

var enableUdldGlobalJsonRequest string = "{\"sonic-udld:UDLD_LIST\": [{\"id\": \"GLOBAL\", \"admin_enable\": true, \"aggressive\": false, \"msg_time\": 1, \"multiplier\": 3}]}"

var udldGlobalEnabledJsonResponse string = "{\"sonic-udld:UDLD\":{\"UDLD_LIST\":[{\"admin_enable\":true,\"aggressive\":false,\"id\":\"GLOBAL\",\"msg_time\":1,\"multiplier\":3}]}}"

var enableUdldPortJsonRequest string = "{\"sonic-udld:UDLD_PORT_LIST\": [{\"ifname\": \"Ethernet28\", \"admin_enable\": true, \"aggressive\": false}]}"

var udldPortEnabledJsonResponse string = "{\"sonic-udld:UDLD_PORT\":{\"UDLD_PORT_LIST\":[{\"admin_enable\":true,\"aggressive\":false,\"ifname\":\"Ethernet28\"}]}}"
