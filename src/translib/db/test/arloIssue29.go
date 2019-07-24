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
