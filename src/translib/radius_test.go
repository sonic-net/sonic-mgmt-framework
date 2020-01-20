
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

/*
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

	//PATCH - global source-address
	t.Run("PATCH - RADIUS global source-address", processSetRequest(radiusGCSourceAddressUrl, radiusGCSourceAddressReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS global source-address", processGetRequest(radiusGCSourceAddressUrl, radiusGCSourceAddressReq, false))
	t.Run("Delete - RADIUS global source-address", processDeleteRequest(radiusGCSourceAddressUrl))
	t.Run("Verify: Delete - RADIUS global source-address", processGetRequest(radiusGCSourceAddressUrl, radiusGCSourceAddressEmptyReq, false))

	//PATCH - global auth-type
	t.Run("PATCH - RADIUS global auth-type", processSetRequest(radiusGCAuthTypeUrl, radiusGCAuthTypeReq, "PATCH", false))
	t.Run("Verify: PATCH - RADIUS global auth-type", processGetRequest(radiusGCAuthTypeUrl, radiusGCAuthTypeReq, false))
	t.Run("Delete - RADIUS global auth-type", processDeleteRequest(radiusGCAuthTypeUrl))
	t.Run("Verify: Delete - RADIUS global auth-type", processGetRequest(radiusGCAuthTypeUrl, radiusGCAuthTypeEmptyReq, false))


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
var radiusGCAuthTypeUrl string = radiusGCUrl + "openconfig-system-ext:auth-type"
var radiusGCTimeoutUrl string = radiusGCUrl + "openconfig-system-ext:timeout"

//JSON data

var radiusGCSourceAddressReq string = "{\"openconfig-system-ext:source-address\":\"1.1.1.1\"}"
var radiusGCSourceAddressEmptyReq string = "{\"openconfig-system-ext:source-address\":\"\"}"
var radiusGCAuthTypeReq string = "{\"openconfig-system-ext:auth-type\":\"mschapv2\"}"
var radiusGCAuthTypeEmptyReq string = "{\"openconfig-system-ext:auth-type\":\"\"}"

*/

