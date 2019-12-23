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
	"translib/tlerr"
)

func main() {
	var avalue,rvalue db.Value
	var akey,rkey db.Key
	var e error

	defer glog.Flush()

	flag.Parse()

	fmt.Println("Creating the DB ==============")
	d,_ := db.NewDB(db.Options {
	                DBNo              : db.ConfigDB,
	                InitIndicator     : "CONFIG_DB_INITIALIZED",
	                TableNameSeparator: "|",
	                KeySeparator      : "|",
                      })

//	fmt.Println("key: CONFIG_DB_INITIALIZED value: ",
//		d.Client.Get("CONFIG_DB_INITIALIZED").String())

	tsa := db.TableSpec { Name: "ACL_TABLE" }
	tsr := db.TableSpec { Name: "ACL_RULE" }

	ca := make([]string, 1, 1)

	fmt.Println("Testing GetEntry error ==============")
	ca[0] = "MyACL1_ACL_IPVNOTEXIST"
	akey = db.Key { Comp: ca}
	avalue, e = d.GetEntry(&tsa, akey)
	fmt.Println("ts: ", tsa, " ", akey, ": ", avalue, " error: ", e)
	if _, ok := e.(tlerr.TranslibRedisClientEntryNotExist) ; ok {
	    fmt.Println("Type is TranslibRedisClientEntryNotExist")
	}


	fmt.Println("Testing NoTransaction SetEntry ==============")
	ca[0] = "MyACL1_ACL_IPV4"
	akey = db.Key { Comp: ca}
	avalue = db.Value { map[string]string {"ports@":"Ethernet0","type":"MIRROR" }}

        d.SetEntry(&tsa, akey, avalue)

	fmt.Println("Testing GetEntry ==============")
	avalue, _ = d.GetEntry(&tsa, akey)
	fmt.Println("ts: ", tsa, " ", akey, ": ", avalue)

	fmt.Println("Testing GetKeys ==============")
	keys, _ := d.GetKeys(&tsa);
	fmt.Println("ts: ", tsa, " keys: ", keys)

	fmt.Println("Testing NoTransaction DeleteEntry ==============")
	akey = db.Key { Comp: ca}

        d.DeleteEntry(&tsa, akey)

	avalue, e = d.GetEntry(&tsa, akey)
	if e == nil {
		fmt.Println("!!! ts: ", tsa, " ", akey, ": ", avalue)
	}

	fmt.Println("Testing 2 more ACLs ==============")
	ca[0] = "MyACL2_ACL_IPV4"
	avalue = db.Value { map[string]string {"ports@":"Ethernet0","type":"MIRROR" }}
        d.SetEntry(&tsa, akey, avalue)

	ca[0] = "MyACL3_ACL_IPV4"
        d.SetEntry(&tsa, akey, avalue)

	ta, _ := d.GetTable(&tsa)
	fmt.Println("ts: ", tsa, " table: ", ta)

	tr, _ := d.GetTable(&tsr)
	fmt.Println("ts: ", tsr, " table: ", tr)

	fmt.Println("Testing Transaction =================")
	rkey = db.Key { Comp: []string { "MyACL2_ACL_IPV4", "RULE_1" }}
	rvalue = db.Value { Field: map[string]string {
		"priority" : "0",
		"packet_action" : "DROP",
		 	},
		}

//	d.StartTx([]db.WatchKeys { {Ts: &tsr, Key: &rkey} })
	d.StartTx([]db.WatchKeys {{Ts: &tsr, Key: &rkey} },
		[]*db.TableSpec { &tsr, &tsa})

	fmt.Println("Sleeping 5...")
	time.Sleep(5 * time.Second)

	d.SetEntry( &tsr, rkey, rvalue)

	e = d.CommitTx()
	if e != nil {
		fmt.Println("Transaction Failed ======= e: ", e)
	}


	fmt.Println("Testing AbortTx =================")
//	d.StartTx([]db.WatchKeys { {Ts: &tsr, Key: &rkey} })
	d.StartTx([]db.WatchKeys {}, []*db.TableSpec { &tsr, &tsa})
	d.DeleteEntry( &tsa, rkey)
	d.AbortTx()
	avalue, e = d.GetEntry(&tsr, rkey)
	fmt.Println("ts: ", tsr, " ", akey, ": ", avalue)

	fmt.Println("Testing DeleteKeys =================")
	d.DeleteKeys(&tsr, db.Key { Comp: []string {"ToBeDeletedACLs*"} })

	fmt.Println("Testing GetTable")
	tr, _ = d.GetTable(&tsr)
	fmt.Println("ts: ", tsr, " table: ", tr)


//	d.DeleteTable(&ts)

	fmt.Println("Testing Tables2TableSpecs =================")
	var tables []string
	tables = []string { "ACL_TABLE", "ACL_RULE" }
	fmt.Println("Tables: ", tables)
	fmt.Println("TableSpecs: ")
	for _, tsi := range db.Tables2TableSpecs(tables) {
		fmt.Println("  ", *tsi)
	}

	fmt.Println("Empty TableSpecs: ")
	for _, tsi := range db.Tables2TableSpecs([]string { } ) {
		fmt.Println("  ", *tsi)
	}


	d.DeleteDB()
}
