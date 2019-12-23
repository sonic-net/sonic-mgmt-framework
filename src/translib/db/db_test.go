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

package db


import (
	// "fmt"
	// "errors"
	// "flag"
	// "github.com/golang/glog"
	"time"
	// "translib/tlerr"
	// "os/exec"
	"os"
	"testing"
	"strconv"
	"reflect"
)

func TestMain(m * testing.M) {

	exitCode := 0

/* Apparently, on an actual switch the swss container will have
 * a redis-server running, which will be in a different container than
 * mgmt, thus this pkill stuff to find out it is running will not work.
 *

	redisServerAttemptedStart := false

TestMainRedo:
	o, e := exec.Command("/usr/bin/pkill", "-HUP", "redis-server").Output()

	if e == nil {

	} else if redisServerAttemptedStart {

		exitCode = 1

	} else {

		fmt.Printf("TestMain: No redis server: pkill: %v\n", o)
		fmt.Println("TestMain: Starting redis-server")
		e = exec.Command("/tools/bin/redis-server").Start()
		time.Sleep(3 * time.Second)
		redisServerAttemptedStart = true
		goto TestMainRedo
	}
*/

	if exitCode == 0 {
		exitCode = m.Run()
	}


	os.Exit(exitCode)
	
}

/*

1.  Create, and close a DB connection. (NewDB(), DeleteDB())

*/

func TestNewDB(t * testing.T) {

	d,e := NewDB(Options {
	                DBNo              : ConfigDB,
	                InitIndicator     : "",
	                TableNameSeparator: "|",
	                KeySeparator      : "|",
			DisableCVLCheck   : true,
                      })

	if d == nil {
		t.Errorf("NewDB() fails e = %v", e)
	} else if e = d.DeleteDB() ; e != nil {
		t.Errorf("DeleteDB() fails e = %v", e)
	}
}


/*

2.  Get an entry (GetEntry())
3.  Set an entry without Transaction (SetEntry())
4.  Delete an entry without Transaction (DeleteEntry())

20. NT: GetEntry() EntryNotExist.

*/

func TestNoTransaction(t * testing.T) {

	var pid int = os.Getpid()

        d,e := NewDB(Options {
                        DBNo              : ConfigDB,
                        InitIndicator     : "",
                        TableNameSeparator: "|",
                        KeySeparator      : "|",
                        DisableCVLCheck   : true,
                      })

	if d == nil {
		t.Errorf("NewDB() fails e = %v", e)
		return
	}

	ts := TableSpec { Name: "TEST_" + strconv.FormatInt(int64(pid), 10) }

	ca := make([]string, 1, 1)
	ca[0] = "MyACL1_ACL_IPVNOTEXIST"
	akey := Key { Comp: ca}
	avalue := Value { map[string]string {"ports@":"Ethernet0","type":"MIRROR" }}
        e = d.SetEntry(&ts, akey, avalue)

	if e != nil {
		t.Errorf("SetEntry() fails e = %v", e)
		return
	}

	v, e := d.GetEntry(&ts, akey)

	if (e != nil) || (!reflect.DeepEqual(v,avalue)) {
		t.Errorf("GetEntry() fails e = %v", e)
		return
	}

        e = d.DeleteEntry(&ts, akey)

	if e != nil {
		t.Errorf("DeleteEntry() fails e = %v", e)
		return
	}

	v, e = d.GetEntry(&ts, akey)

	if e == nil {
		t.Errorf("GetEntry() after DeleteEntry() fails e = %v", e)
		return
	}

	if e = d.DeleteDB() ; e != nil {
		t.Errorf("DeleteDB() fails e = %v", e)
	}
}


/*

5.  Get a Table (GetTable())

9.  Get multiple keys (GetKeys())
10. Delete multiple keys (DeleteKeys())
11. Delete Table (DeleteTable())

*/

func TestTable(t * testing.T) {

	var pid int = os.Getpid()

        d,e := NewDB(Options {
                        DBNo              : ConfigDB,
                        InitIndicator     : "",
                        TableNameSeparator: "|",
                        KeySeparator      : "|",
                        DisableCVLCheck   : true,
                      })

	if d == nil {
		t.Errorf("NewDB() fails e = %v", e)
		return
	}

	ts := TableSpec { Name: "TEST_" + strconv.FormatInt(int64(pid), 10) }

	ca := make([]string, 1, 1)
	ca[0] = "MyACL1_ACL_IPVNOTEXIST"
	akey := Key { Comp: ca}
	avalue := Value { map[string]string {"ports@":"Ethernet0","type":"MIRROR" }}
	ca2 := make([]string, 1, 1)
	ca2[0] = "MyACL2_ACL_IPVNOTEXIST"
	akey2 := Key { Comp: ca2}

        // Add the Entries for Get|DeleteKeys

        e = d.SetEntry(&ts, akey, avalue)

	if e != nil {
		t.Errorf("SetEntry() fails e = %v", e)
		return
	}

        e = d.SetEntry(&ts, akey2, avalue)

	if e != nil {
		t.Errorf("SetEntry() fails e = %v", e)
		return
	}

	keys, e := d.GetKeys(&ts)

	if (e != nil) || (len(keys) != 2) {
		t.Errorf("GetKeys() fails e = %v", e)
		return
	}

	e = d.DeleteKeys(&ts, Key {Comp: []string {"MyACL*_ACL_IPVNOTEXIST"}})

	if e != nil {
		t.Errorf("DeleteKeys() fails e = %v", e)
		return
	}

	v, e := d.GetEntry(&ts, akey)

	if e == nil {
		t.Errorf("GetEntry() after DeleteKeys() fails e = %v", e)
		return
	}



        // Add the Entries again for Table

        e = d.SetEntry(&ts, akey, avalue)

	if e != nil {
		t.Errorf("SetEntry() fails e = %v", e)
		return
	}

        e = d.SetEntry(&ts, akey2, avalue)

	if e != nil {
		t.Errorf("SetEntry() fails e = %v", e)
		return
	}

	tab, e := d.GetTable(&ts)

	if e != nil {
		t.Errorf("GetTable() fails e = %v", e)
		return
	}

	v, e = tab.GetEntry(akey)

	if (e != nil) || (!reflect.DeepEqual(v,avalue)) {
		t.Errorf("Table.GetEntry() fails e = %v", e)
		return
	}

	e = d.DeleteTable(&ts)

	if e != nil {
		t.Errorf("DeleteTable() fails e = %v", e)
		return
	}

	v, e = d.GetEntry(&ts, akey)

	if e == nil {
		t.Errorf("GetEntry() after DeleteTable() fails e = %v", e)
		return
	}

	if e = d.DeleteDB() ; e != nil {
		t.Errorf("DeleteDB() fails e = %v", e)
	}
}


/* Tests for 

6.  Set an entry with Transaction (StartTx(), SetEntry(), CommitTx())
7.  Delete an entry with Transaction (StartTx(), DeleteEntry(), CommitTx())
8.  Abort Transaction. (StartTx(), DeleteEntry(), AbortTx())

12. Set an entry with Transaction using WatchKeys Check-And-Set(CAS)
13. Set an entry with Transaction using Table CAS
14. Set an entry with Transaction using WatchKeys, and Table CAS

15. Set an entry with Transaction with empty WatchKeys, and Table CAS
16. Negative Test(NT): Fail a Transaction using WatchKeys CAS
17. NT: Fail a Transaction using Table CAS
18. NT: Abort an Transaction with empty WatchKeys/Table CAS

Cannot Automate 19 for now
19. NT: Check V logs, Error logs

 */

func TestTransaction(t * testing.T) {
	for transRun := TransRunBasic ; transRun < TransRunEnd ; transRun++ {
		testTransaction(t, transRun)
	}
}

type TransRun int

const (
	TransRunBasic         TransRun = iota // 0
	TransRunWatchKeys                     // 1
	TransRunTable                         // 2
	TransRunWatchKeysAndTable             // 3
	TransRunEmptyWatchKeysAndTable        // 4
	TransRunFailWatchKeys                 // 5
	TransRunFailTable                     // 6

	// Nothing after this.
	TransRunEnd
)

func testTransaction(t * testing.T, transRun TransRun) {

	var pid int = os.Getpid()

        d,e := NewDB(Options {
                        DBNo              : ConfigDB,
                        InitIndicator     : "",
                        TableNameSeparator: "|",
                        KeySeparator      : "|",
                        DisableCVLCheck   : true,
                      })

	if d == nil {
		t.Errorf("NewDB() fails e = %v, transRun = %v", e, transRun)
		return
	}

	ts := TableSpec { Name: "TEST_" + strconv.FormatInt(int64(pid), 10) }

	ca := make([]string, 1, 1)
	ca[0] = "MyACL1_ACL_IPVNOTEXIST"
	akey := Key { Comp: ca}
	avalue := Value { map[string]string {"ports@":"Ethernet0","type":"MIRROR" }}

	var watchKeys []WatchKeys
	var table []*TableSpec

	switch transRun {
	case TransRunBasic, TransRunWatchKeysAndTable:
		watchKeys = []WatchKeys{{Ts: &ts, Key: &akey}}
		table = []*TableSpec { &ts }
	case TransRunWatchKeys, TransRunFailWatchKeys:
		watchKeys = []WatchKeys{{Ts: &ts, Key: &akey}}
		table = []*TableSpec { }
	case TransRunTable, TransRunFailTable:
		watchKeys = []WatchKeys{}
		table = []*TableSpec { &ts }
	}

	e = d.StartTx(watchKeys, table)

	if e != nil {
		t.Errorf("StartTx() fails e = %v", e)
		return
	}

        e = d.SetEntry(&ts, akey, avalue)

	if e != nil {
		t.Errorf("SetEntry() fails e = %v", e)
		return
	}

	e = d.CommitTx()

	if e != nil {
		t.Errorf("CommitTx() fails e = %v", e)
		return
	}

	v, e := d.GetEntry(&ts, akey)

	if (e != nil) || (!reflect.DeepEqual(v,avalue)) {
		t.Errorf("GetEntry() after Tx fails e = %v", e)
		return
	}

	e = d.StartTx(watchKeys, table)

	if e != nil {
		t.Errorf("StartTx() fails e = %v", e)
		return
	}

        e = d.DeleteEntry(&ts, akey)

	if e != nil {
		t.Errorf("DeleteEntry() fails e = %v", e)
		return
	}

	e = d.AbortTx()

	if e != nil {
		t.Errorf("AbortTx() fails e = %v", e)
		return
	}

	v, e = d.GetEntry(&ts, akey)

	if (e != nil) || (!reflect.DeepEqual(v,avalue)) {
		t.Errorf("GetEntry() after Abort Tx fails e = %v", e)
		return
	}

	e = d.StartTx(watchKeys, table)

	if e != nil {
		t.Errorf("StartTx() fails e = %v", e)
		return
	}

        e = d.DeleteEntry(&ts, akey)

	if e != nil {
		t.Errorf("DeleteEntry() fails e = %v", e)
		return
	}

	switch transRun {
	case TransRunFailWatchKeys, TransRunFailTable:
        	d2,_ := NewDB(Options {
                        DBNo              : ConfigDB,
                        InitIndicator     : "",
                        TableNameSeparator: "|",
                        KeySeparator      : "|",
                        DisableCVLCheck   : true,
                      })

		d2.StartTx(watchKeys, table);
        	d2.DeleteEntry(&ts, akey)
		d2.CommitTx();
		d2.DeleteDB();
	default:
	}

	e = d.CommitTx()

	switch transRun {
	case TransRunFailWatchKeys, TransRunFailTable:
		if e == nil {
			t.Errorf("NT CommitTx() tr: %v fails e = %v",
				transRun, e)
			return
		}
	default:
		if e != nil {
			t.Errorf("CommitTx() fails e = %v", e)
			return
		}
	}

	v, e = d.GetEntry(&ts, akey)

	if e == nil {
		t.Errorf("GetEntry() after Tx DeleteEntry() fails e = %v", e)
		return
	}

	d.DeleteMapAll(&ts)

	if e = d.DeleteDB() ; e != nil {
		t.Errorf("DeleteDB() fails e = %v", e)
	}
}


func TestMap(t * testing.T) {

	var pid int = os.Getpid()

	d,e := NewDB(Options {
	                DBNo              : ConfigDB,
	                InitIndicator     : "",
	                TableNameSeparator: "|",
	                KeySeparator      : "|",
			DisableCVLCheck   : true,
                      })

	if d == nil {
		t.Errorf("NewDB() fails e = %v", e)
		return
	}

	ts := TableSpec { Name: "TESTMAP_" + strconv.FormatInt(int64(pid), 10) }

	d.SetMap(&ts, "k1", "v1");
	d.SetMap(&ts, "k2", "v2");

	if v, e := d.GetMap(&ts, "k1"); v != "v1" {
		t.Errorf("GetMap() fails e = %v", e)
		return
	}

	if v, e := d.GetMapAll(&ts) ;
		(e != nil) ||
		(!reflect.DeepEqual(v,
			Value{ Field: map[string]string {
				"k1" : "v1", "k2" : "v2" }})) {
		t.Errorf("GetMapAll() fails e = %v", e)
		return
	}

	d.DeleteMapAll(&ts)

	if e = d.DeleteDB() ; e != nil {
		t.Errorf("DeleteDB() fails e = %v", e)
	}
}

func TestSubscribe(t * testing.T) {

	var pid int = os.Getpid()

	var hSetCalled, hDelCalled, delCalled bool

        d,e := NewDB(Options {
                        DBNo              : ConfigDB,
                        InitIndicator     : "",
                        TableNameSeparator: "|",
                        KeySeparator      : "|",
                        DisableCVLCheck   : true,
                      })

	if (d == nil) || (e != nil) {
		t.Errorf("NewDB() fails e = %v", e)
		return
	}

	ts := TableSpec { Name: "TEST_" + strconv.FormatInt(int64(pid), 10) }

	ca := make([]string, 1, 1)
	ca[0] = "MyACL1_ACL_IPVNOTEXIST"
	akey := Key { Comp: ca}
	avalue := Value { map[string]string {"ports@":"Ethernet0","type":"MIRROR" }}

	var skeys [] *SKey = make([]*SKey, 1)
        skeys[0] = & (SKey { Ts: &ts, Key: &akey,
		SEMap: map[SEvent]bool {
			SEventHSet:	true,
			SEventHDel:	true,
			SEventDel:	true,
		}})

	s,e := SubscribeDB(Options {
	                DBNo              : ConfigDB,
	                InitIndicator     : "CONFIG_DB_INITIALIZED",
	                TableNameSeparator: "|",
	                KeySeparator      : "|",
                        DisableCVLCheck   : true,
                      }, skeys, func (s *DB,
				skey *SKey, key *Key,
				event SEvent) error {
			switch event {
			case SEventHSet: hSetCalled = true
			case SEventHDel: hDelCalled = true
			case SEventDel: delCalled = true
			default:
			}
			return nil })

	if (s == nil) || (e != nil) {
		t.Errorf("Subscribe() returns error e: %v", e)
		return
	}

        d.SetEntry(&ts, akey, avalue)
        d.DeleteEntryFields(&ts, akey, avalue)

	time.Sleep(5 * time.Second)

	if !hSetCalled || !hDelCalled || !delCalled {
		t.Errorf("Subscribe() callbacks missed: %v %v %v", hSetCalled,
			hDelCalled, delCalled)
		return
	}

	s.UnsubscribeDB()

	time.Sleep(2 * time.Second)

	if e = d.DeleteDB() ; e != nil {
		t.Errorf("DeleteDB() fails e = %v", e)
	}
}

