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
	"time"
	// "translib/tlerr"
)


func handler(d *db.DB, skey *db.SKey, key *db.Key, event db.SEvent) error {
	fmt.Println("***handler: d: ", d, " skey: ", *skey, " key: ", *key,
		" event: ", event)
	return nil
}


func main() {
	// var avalue,rvalue db.Value
	var akey db.Key
        // var rkey db.Key
	// var e error

	defer glog.Flush()

	flag.Parse()

	tsa := db.TableSpec { Name: "ACL_TABLE" }
	// tsr := db.TableSpec { Name: "ACL_RULE" }

	ca := make([]string, 1, 1)
	ca[0] = "MyACL1_ACL_IPVNOTEXIST*"
	akey = db.Key { Comp: ca}
	var skeys [] *db.SKey = make([]*db.SKey, 1)
        skeys[0] = & (db.SKey { Ts: &tsa, Key: &akey,
		SEMap: map[db.SEvent]bool {
			db.SEventHSet:	true,
			db.SEventHDel:	true,
			db.SEventDel:	true,
		}})

	fmt.Println("Creating the SubscribeDB ==============")
	d,e := db.SubscribeDB(db.Options {
	                DBNo              : db.ConfigDB,
	                InitIndicator     : "CONFIG_DB_INITIALIZED",
	                TableNameSeparator: "|",
	                KeySeparator      : "|",
                      }, skeys, handler)

	if e != nil {
		fmt.Println("Subscribe() returns error e: ", e)
	}

	fmt.Println("Sleeping 15 ==============")
	time.Sleep(15 * time.Second)


	fmt.Println("Testing UnsubscribeDB ==============")

	d.UnsubscribeDB()

	fmt.Println("Sleeping 5 ==============")
	time.Sleep(5 * time.Second)


}
