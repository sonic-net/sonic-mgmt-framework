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

/*
UT for
https://github.com/project-arlo/sonic-mgmt-framework/issues/29
*/

package main

import (
	"fmt"
	// "errors"
	"flag"
	"github.com/golang/glog"
	"translib/db"
	// "time"
	// "translib/tlerr"
)

func main() {
	var avalue db.Value
	var akey db.Key
	var e error

	defer glog.Flush()

	flag.Parse()

	fmt.Println("https://github.com/project-arlo/sonic-mgmt-framework/issues/29")
	fmt.Println("Creating the DB ==============")
	d,_ := db.NewDB(db.Options {
	                DBNo              : db.ApplDB,
	                InitIndicator     : "",
	                TableNameSeparator: ":",
	                KeySeparator      : ":",
                      })

	tsi := db.TableSpec { Name: "INTF_TABLE", CompCt: 2 }

	ca := make([]string, 2, 2)

	fmt.Println("Testing SetEntry ==============")
	ca[0] = "Ethernet20"
	ca[1] = "a::b/64"
	akey = db.Key { Comp: ca}
	avalue = db.Value { Field: map[string]string {
								"scope" : "global",
								"family" : "IPv4",
								} }

	e = d.SetEntry(&tsi, akey, avalue)
	if e != nil {
		fmt.Println("SetEntry() ERROR: e: ", e)
		return
	}

	fmt.Println("Testing GetEntry ==============")

	avalue, e = d.GetEntry(&tsi, akey)
	if e != nil {
		fmt.Println("GetEntry() ERROR: e: ", e)
		return
	}

	fmt.Println("ts: ", tsi, " ", akey, ": ", avalue)

	d.DeleteDB()
}
