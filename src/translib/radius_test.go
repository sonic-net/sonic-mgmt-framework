
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
	"os"
)

const (
	RADIUS_TABLE               = "RADIUS_TABLE"
	RADIUS_SERVER_TABLE        = "RADIUS_SERVER_TABLE"
)

func clearRADIUSDb() {
	fmt.Println("---------  Init RADIUS Go test  --------")

	if err := clearRADIUSDataFromConfigDb(); err == nil {
		fmt.Println("----- Removed All RADIUS Data from Db  -------")
		createRADIUSData ()
	} else {
		fmt.Printf("Failed to remove All RADIUS Data from Db: %v", err)
		os.Exit(1) // Cancel any further tests.
	}
}

func createRADIUSData () {
}

func TestRADIUSConfigPatchDeleteGetAPIs(t *testing.T) {

	clearRADIUSDb()		

        // Global ==================================================

	//PATCH - global source-address
	t.Run("PATCH - RADIUS global source-address", processSetRequest(radiusGCSourceAddressUrl, radiusGCSourceAddressReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS global source-address", processGetRequest(radiusGCSourceAddressUrl, radiusGCSourceAddressReq, false))
	t.Run("Delete - RADIUS global source-address", processDeleteRequest(radiusGCSourceAddressUrl))
	t.Run("Verify: Delete - RADIUS global source-address", processGetRequest(radiusGCSourceAddressUrl, radiusGCSourceAddressEmptyReq, false))

	//PATCH - global timeout
	t.Run("PATCH - RADIUS global timeout", processSetRequest(radiusGCTimeoutUrl, radiusGCTimeoutReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS global timeout", processGetRequest(radiusGCTimeoutUrl, radiusGCTimeoutReq, false))
	t.Run("Delete - RADIUS global timeout", processDeleteRequest(radiusGCTimeoutUrl))
	t.Run("Verify: Delete - RADIUS global timeout", processGetRequest(radiusGCTimeoutUrl, radiusGCTimeoutEmptyReq, false))

	//PATCH - global retransmit-attempts
	t.Run("PATCH - RADIUS global retransmit-attempts", processSetRequest(radiusGCRetransmitUrl, radiusGCRetransmitReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS global retransmit-attempts", processGetRequest(radiusGCRetransmitUrl, radiusGCRetransmitReq, false))
	t.Run("Delete - RADIUS global retransmit-attempts", processDeleteRequest(radiusGCRetransmitUrl))
	t.Run("Verify: Delete - RADIUS global retransmit-attempts", processGetRequest(radiusGCRetransmitUrl, radiusGCRetransmitEmptyReq, false))

	//PATCH - global secret-key
	t.Run("PATCH - RADIUS global secret-key", processSetRequest(radiusGCSecretUrl, radiusGCSecretReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS global secret-key", processGetRequest(radiusGCSecretUrl, radiusGCSecretReq, false))
	t.Run("Delete - RADIUS global secret-key", processDeleteRequest(radiusGCSecretUrl))
	t.Run("Verify: Delete - RADIUS global secret-key", processGetRequest(radiusGCSecretUrl, radiusGCSecretEmptyReq, false))

	//PATCH - global auth-type
	t.Run("PATCH - RADIUS global auth-type", processSetRequest(radiusGCAuthTypeUrl, radiusGCAuthTypeReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS global auth-type", processGetRequest(radiusGCAuthTypeUrl, radiusGCAuthTypeReq, false))
	t.Run("Delete - RADIUS global auth-type", processDeleteRequest(radiusGCAuthTypeUrl))
	t.Run("Verify: Delete - RADIUS global auth-type", processGetRequest(radiusGCAuthTypeUrl, radiusGCAuthTypeEmptyReq, false))

        // Host ==================================================

	//PATCH - host auth-port
	t.Run("PATCH - RADIUS host auth-port", processSetRequest(radiusHRCAuthPortUrl, radiusHRCAuthPortReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS host auth-port", processGetRequest(radiusHRCAuthPortUrl, radiusHRCAuthPortReq, false))
	t.Run("Delete - RADIUS host auth-port", processDeleteRequest(radiusHRCAuthPortUrl))
	t.Run("Verify: Delete - RADIUS host auth-port", processGetRequest(radiusHRCAuthPortUrl, radiusHRCAuthPortEmptyReq, false))

	//PATCH - host secret-key
	t.Run("PATCH - RADIUS host secret-key", processSetRequest(radiusHRCSecretUrl, radiusHRCSecretReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS host secret-key", processGetRequest(radiusHRCSecretUrl, radiusHRCSecretReq, false))
	t.Run("Delete - RADIUS host secret-key", processDeleteRequest(radiusHRCSecretUrl))
	t.Run("Verify: Delete - RADIUS host secret-key", processGetRequest(radiusHRCSecretUrl, radiusHRCSecretEmptyReq, false))

	//PATCH - host retransmit-attempts
	t.Run("PATCH - RADIUS host retransmit-attempts", processSetRequest(radiusHRCRetransmitUrl, radiusHRCRetransmitReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS host retransmit-attempts", processGetRequest(radiusHRCRetransmitUrl, radiusHRCRetransmitReq, false))
	t.Run("Delete - RADIUS host retransmit-attempts", processDeleteRequest(radiusHRCRetransmitUrl))
	t.Run("Verify: Delete - RADIUS host retransmit-attempts", processGetRequest(radiusHRCRetransmitUrl, radiusHRCRetransmitEmptyReq, false))

	//PATCH - host timeout
	t.Run("PATCH - RADIUS host timeout", processSetRequest(radiusHCTimeoutUrl, radiusHCTimeoutReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS host timeout", processGetRequest(radiusHCTimeoutUrl, radiusHCTimeoutReq, false))
	t.Run("Delete - RADIUS host timeout", processDeleteRequest(radiusHCTimeoutUrl))
	t.Run("Verify: Delete - RADIUS host timeout", processGetRequest(radiusHCTimeoutUrl, radiusHCTimeoutEmptyReq, false))

	//PATCH - host auth-type
	t.Run("PATCH - RADIUS host auth-type", processSetRequest(radiusHCAuthTypeUrl, radiusHCAuthTypeReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS host auth-type", processGetRequest(radiusHCAuthTypeUrl, radiusHCAuthTypeReq, false))
	t.Run("Delete - RADIUS host auth-type", processDeleteRequest(radiusHCAuthTypeUrl))
	t.Run("Verify: Delete - RADIUS host auth-type", processGetRequest(radiusHCAuthTypeUrl, radiusHCAuthTypeEmptyReq, false))

	//PATCH - host priority
	t.Run("PATCH - RADIUS host priority", processSetRequest(radiusHCPriorityUrl, radiusHCPriorityReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS host priority", processGetRequest(radiusHCPriorityUrl, radiusHCPriorityReq, false))
	t.Run("Delete - RADIUS host priority", processDeleteRequest(radiusHCPriorityUrl))
	t.Run("Verify: Delete - RADIUS host priority", processGetRequest(radiusHCPriorityUrl, radiusHCPriorityEmptyReq, false))

	//PATCH - host vrf
	t.Run("PATCH - RADIUS host vrf", processSetRequest(radiusHCVrfUrl, radiusHCVrfReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS host vrf", processGetRequest(radiusHCVrfUrl, radiusHCVrfReq, false))
	t.Run("Delete - RADIUS host vrf", processDeleteRequest(radiusHCVrfUrl))
	t.Run("Verify: Delete - RADIUS host vrf", processGetRequest(radiusHCVrfUrl, radiusHCVrfEmptyReq, false))

}

func clearRADIUSDataFromConfigDb() error {
	var err error
	
	radius_ts := db.TableSpec{Name: "RADIUS"}
	radius_server_ts := db.TableSpec{Name: "RADIUS_SERVER"}

	d := getConfigDb()
	
	if d == nil {
		err = errors.New("Failed to connect to config Db")
		return err
	}

	if err = d.DeleteTable(&radius_ts); err != nil {
		err = errors.New("Failed to delete RADIUS Table")
		return err
	}

	if err = d.DeleteTable(&radius_server_ts); err != nil {
		err = errors.New("Failed to delete RADIUS_SERVER Table")
		return err
	}

	return err
}

//config-URL

var systemUrl string = "/openconfig-system:system/"
var aaaUrl string = systemUrl + "aaa/"
var serverGroupsUrl = aaaUrl + "server-groups/"
var radiusGUrl string = serverGroupsUrl + "server-group[name=RADIUS]/"
var radiusGCUrl string = radiusGUrl + "config/"

var radiusGCSourceAddressUrl string = radiusGCUrl + "openconfig-system-ext:source-address"

var radiusGCTimeoutUrl string = radiusGCUrl + "openconfig-system-ext:timeout"

var radiusGCRetransmitUrl string = radiusGCUrl + "openconfig-system-ext:retransmit-attempts"

var radiusGCSecretUrl string = radiusGCUrl + "openconfig-system-ext:secret-key"

var radiusGCAuthTypeUrl string = radiusGCUrl + "openconfig-system-ext:auth-type"


// Host config-URL

var radiusHUrl string = radiusGUrl + "servers/server[address=1.1.1.1]/"
var radiusHRCUrl string = radiusHUrl + "radius/config/"
var radiusHRCAuthPortUrl string = radiusHRCUrl + "auth-port"
var radiusHRCSecretUrl string = radiusHRCUrl + "secret-key"
var radiusHRCRetransmitUrl string = radiusHRCUrl + "retransmit-attempts"

var radiusHCUrl string = radiusHUrl + "config/"
var radiusHCTimeoutUrl string = radiusHCUrl + "timeout"
var radiusHCAuthTypeUrl string = radiusHCUrl + "openconfig-system-ext:auth-type"
var radiusHCPriorityUrl string = radiusHCUrl + "openconfig-system-ext:priority"
var radiusHCVrfUrl string = radiusHCUrl + "openconfig-system-ext:vrf"

//JSON data

// Global JSON data

var radiusGCSourceAddressReq string = "{\"source-address\":\"1.1.1.1\"}"
var radiusGCSourceAddressEmptyReq string = "{\"openconfig-system-ext:source-address\":\"\"}"

var radiusGCTimeoutReq string = "{\"timeout\":6}"
var radiusGCTimeoutEmptyReq string = "{\"openconfig-system-ext:timeout\":0}"

var radiusGCRetransmitReq string = "{\"retransmit-attempts\":7}"
var radiusGCRetransmitEmptyReq string = "{\"openconfig-system-ext:retransmit-attempts\":0}"

var radiusGCSecretReq string = "{\"secret-key\":\"sharedsecret\"}"
var radiusGCSecretEmptyReq string = "{\"openconfig-system-ext:secret-key\":\"\"}"

var radiusGCAuthTypeReq string = "{\"auth-type\":\"mschapv2\"}"
var radiusGCAuthTypeEmptyReq string = "{\"openconfig-system-ext:auth-type\":\"\"}"

// Host JSON data

var radiusHRCAuthPortReq string = "{\"auth-port\":1912}"
var radiusHRCAuthPortEmptyReq string = "{\"auth-port\":1812}"

var radiusHRCSecretReq string = "{\"secret-key\":\"sharedsecret\"}"
var radiusHRCSecretEmptyReq string = "{\"openconfig-system:secret-key\":\"\"}"

var radiusHRCRetransmitReq string = "{\"retransmit-attempts\":8}"
var radiusHRCRetransmitEmptyReq string = "{\"openconfig-system:retransmit-attempts\":0}"

var radiusHCTimeoutReq string = "{\"timeout\":9}"
var radiusHCTimeoutEmptyReq string = "{\"openconfig-system:timeout\":0}"

var radiusHCAuthTypeReq string = "{\"auth-type\":\"chap\"}"
var radiusHCAuthTypeEmptyReq string = "{\"openconfig-system-ext:auth-type\":\"\"}"

var radiusHCPriorityReq string = "{\"priority\":2}"
var radiusHCPriorityEmptyReq string = "{\"openconfig-system-ext:priority\":0}"

var radiusHCVrfReq string = "{\"vrf\":\"mgmt\"}"
var radiusHCVrfEmptyReq string = "{\"openconfig-system-ext:vrf\":\"\"}"

