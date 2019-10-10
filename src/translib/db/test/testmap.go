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


func handler(d *db.DB, skey *db.SKey, key *db.Key, event db.SEvent) error {
	fmt.Println("***handler: d: ", d, " skey: ", *skey, " key: ", *key,
		" event: ", event)
	return nil
}


func main() {
	defer glog.Flush()

	flag.Parse()

	tsc := db.TableSpec { Name: "COUNTERS_PORT_NAME_MAP" }

	fmt.Println("Creating the SubscribeDB ==============")
	d,e := db.NewDB(db.Options {
	                DBNo              : db.CountersDB,
	                InitIndicator     : "",
	                TableNameSeparator: ":",
	                KeySeparator      : ":",
                      })

	if e != nil {
		fmt.Println("NewDB() returns error e: ", e)
	}

	fmt.Println("Setting Some Maps ==============")
	d.SetMap(&tsc, "Ethernet2", "oid:0x1000000000002")
	d.SetMap(&tsc, "Ethernet5", "oid:0x1000000000005")
	d.SetMap(&tsc, "Ethernet3", "oid:0x1000000000003")

	fmt.Println("GetMapAll ==============")
	v, e := d.GetMapAll(&tsc)
	if e != nil {
		fmt.Println("GetMapAll() returns error e: ", e)
	}
	fmt.Println("v: ", v)

	fmt.Println("GetMap ==============")
	r2, e := d.GetMap(&tsc, "Ethernet2")
	if e != nil {
		fmt.Println("GetMap() returns error e: ", e)
	}
	r5, e := d.GetMap(&tsc, "Ethernet5")
	if e != nil {
		fmt.Println("GetMap() returns error e: ", e)
	}
	r3, e := d.GetMap(&tsc, "Ethernet3")
	if e != nil {
		fmt.Println("GetMap() returns error e: ", e)
	}

	fmt.Println("r2, r5, r3", r2, r5, r3)


	fmt.Println("GetMap NotExist mapKey ==============")
	rN, e := d.GetMap(&tsc, "EthernetN")
	if e == nil {
		fmt.Println("GetMap() NotExist mapKey returns nil !!! ", rN)
	}

	vN, e := d.GetMapAll(& db.TableSpec { Name: "NOTEXITMAP" } )
	if e == nil {
		fmt.Println("GetMapAll() NotExist returns nil !!! ", vN)
	}

	d.DeleteMapAll(&tsc)

	d.DeleteDB()
}
