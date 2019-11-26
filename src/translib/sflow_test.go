//////////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Dell, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//////////////////////////////////////////////////////////////////////////

package translib

import (
	"errors"
	"fmt"
	"testing"
	"time"
	db "translib/db"
)

func init() {
	fmt.Println("+++++  Init sFlow test  +++++")
}

func Test_sFlowOperations(t *testing.T) {

	fmt.Println("+++++  Start sFlow testing  +++++")

	url := "/sonic-sflow:sonic-sflow/SFLOW/SFLOW_LIST[sflow_key=global]"

        //Set admin state
	adminUrl := url + "/admin_state"
	t.Run("Enable sFlow", processSetRequest(adminUrl, globalAdminJson, "PATCH", false))
	time.Sleep(1 * time.Second)

        //Set polling interval
        pollingUrl := url + "/polling_interval"
	t.Run("Set sFlow polling interval", processSetRequest(pollingUrl, pollingJson, "PATCH", false))
	time.Sleep(1 * time.Second)

        //Set AgentID
        agentUrl := url + "/agent_id"
	t.Run("Set sFlow agent ID", processSetRequest(agentUrl, agentIdJson, "PATCH", false))
	time.Sleep(1 * time.Second)

        // Verify global configurations
	t.Run("Verify global configurations", processGetRequest(url, globalConfigGetJsonResp, false))

        //Add collector
        url = "/sonic-sflow:sonic-sflow/SFLOW_COLLECTOR/SFLOW_COLLECTOR_LIST[collector_name=col1]"
	t.Run("Add sFlow collector col1", processSetRequest(url, col1Json, "PATCH", false))
	time.Sleep(1 * time.Second)

        // Verify collector configurations
	t.Run("Verify sFlow collector col1", processGetRequest(url, col1Json, false))

        // Set collector ip
        ipUrl := url + "/collector_ip"
	t.Run("Set sFlow collector col1 ip", processSetRequest(ipUrl, colIPJson, "PATCH", false))
	time.Sleep(1 * time.Second)

        // Set collector port
        portUrl := url + "/collector_port"
	t.Run("Set sFlow collector col1 port", processSetRequest(portUrl, colPortJson, "PATCH", false))
	time.Sleep(2 * time.Second)

        // Verify collector configurations
	t.Run("Verify_sFlow_collector", processGetRequest(url, col1ModJson, false))
}

func clearsFlowDataFromDb() error {
	var err error
	sFlowTable := db.TableSpec{Name: "SFLOW|global"}
	colTable := db.TableSpec{Name: "SFLOW_COLLECTOR|col1"}
	intfTable := db.TableSpec{Name: "SFLOW_SESSION|Ethernet0"}

	d := getConfigDb()
	if d == nil {
		err = errors.New("Failed to connect to config Db")
		return err
	}
	if err = d.DeleteTable(&sFlowTable); err != nil {
		err = errors.New("Failed to clear SFLOW|global table")
		return err
	}
	if err = d.DeleteTable(&colTable); err != nil {
		err = errors.New("Failed to clear SFLOW_COLLECTOR|col1 Table")
		return err
	}
	if err = d.DeleteTable(&intfTable); err != nil {
		err = errors.New("Failed to clear SFLOW_SESSION|Ethernet0 Table")
		return err
	}
	return err
}
/***************************************************************************/
///////////                  JSON Data for Tests              ///////////////
/***************************************************************************/

var globalAdminJson string = "{\"sonic-sflow:admin_state\": \"up\"}"
var pollingJson string = "{\"sonic-sflow:polling_interval\": 10}"
var agentIdJson string = "{\"sonic-sflow:agent_id\": \"Ethernet0\"}"
var globalConfigGetJsonResp string = "{\"sonic-sflow:SFLOW_LIST\":[{\"admin_state\":\"up\",\"agent_id\":\"Ethernet0\",\"polling_interval\":10,\"sflow_key\":\"global\"}]}"

var col1Json string = "{\"sonic-sflow:SFLOW_COLLECTOR_LIST\":[{\"collector_ip\":\"1.1.1.1\",\"collector_name\":\"col1\",\"collector_port\":4444}]}"
var col1ModJson string = "{\"sonic-sflow:SFLOW_COLLECTOR_LIST\":[{\"collector_ip\":\"2.2.2.2\",\"collector_name\":\"col1\",\"collector_port\":1234}]}"

var colIPJson string = "{\"sonic-sflow:collector_ip\": \"2.2.2.2\"}"
var colPortJson string = "{\"sonic-sflow:collector_port\": 1234}"
