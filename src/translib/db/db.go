/*
Copyright 2019 Broadcom. All rights reserved.
The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
*/

/*
Package db implements a wrapper over the go-redis/redis.

There may be an attempt to mimic sonic-py-swsssdk to ease porting of
code written in python using that SDK to Go Language.

Example:

 * Initialization:

        d := db.NewDB(db.Options {
                        DBNo              : db.ConfigDB,
                        InitIndicator     : "CONFIG_DB_INITIALIZED",
                        TableNameSeparator: "|",
                        KeySeparator      : "|",
                      })

 * Close:

        d.DeleteDB()


 * No-Transaction SetEntry

        tsa := db.TableSpec { Name: "ACL_TABLE" }
        tsr := db.TableSpec { Name: "ACL_RULE" }

        ca := make([]string, 1, 1)

        ca[0] = "MyACL1_ACL_IPV4"
        akey := db.Key { Comp: ca}
        avalue := db.Value {map[string]string {"ports":"eth0","type":"mirror" }}

        d.SetEntry(&tsa, akey, avalue)

 * GetEntry

        avalue, _ := d.GetEntry(&tsa, akey)

 * GetKeys

        keys, _ := d.GetKeys(&tsa);

 * No-Transaction DeleteEntry

        d.DeleteEntry(&tsa, akey)

 * GetTable

        ta, _ := d.GetTable(&tsa)

 * No-Transaction DeleteTable

        d.DeleteTable(&ts)

 * Transaction

        rkey := db.Key { Comp: []string { "MyACL2_ACL_IPV4", "RULE_1" }}
        rvalue := db.Value { Field: map[string]string {
                "priority" : "0",
                "packet_action" : "eth1",
                        },
                }

        d.StartTx([]db.WatchKeys { {Ts: &tsr, Key: &rkey} })

        d.SetEntry( &tsa, akey, avalue)
        d.SetEntry( &tsr, rkey, rvalue)

        e := d.CommitTx()

 * Transaction Abort

        d.StartTx([]db.WatchKeys { {Ts: &tsr, Key: &rkey} })
        d.DeleteEntry( &tsa, rkey)
        d.AbortTx()


*/
package db

import (
	"fmt"
	"strconv"

	//	"reflect"
	"errors"
	"strings"

	"github.com/go-redis/redis"
	"github.com/golang/glog"
)

const (
	DefaultRedisUNIXSocket  string = "/var/run/redis/redis.sock"
	DefaultRedisLocalTCPEP  string = "localhost:6379"
	DefaultRedisRemoteTCPEP string = "127.0.0.1:6379"
)

func init() {
}

// DBNum type indicates the type of DB (Eg: ConfigDB, ApplDB, ...).
type DBNum int

const (
	ApplDB        DBNum = iota // 0
	AsicDB                     // 1
	CountersDB                 // 2
	LogLevelDB                 // 3
	ConfigDB                   // 4
	FlexCounterDB              // 5
	StateDB                    // 6

	// All DBs added above this line, please ----
	MaxDB // 7 The Number of DBs
)

// Options gives parameters for opening the redis client.
type Options struct {
	DBNo               DBNum
	InitIndicator      string
	TableNameSeparator string
	KeySeparator       string
}

func (o Options) String() string {
	return fmt.Sprintf(
		"{ DBNo: %v, InitIndicator: %v, TableNameSeparator: %v, KeySeparator: %v }",
		o.DBNo, o.InitIndicator, o.TableNameSeparator, o.KeySeparator)
}

type _txState int

const (
	txStateNone      _txState = iota // Idle (No transaction)
	txStateWatch                     // WATCH issued
	txStateSet                       // At least one SET|DEL done.
	txStateMultiExec                 // Between MULTI & EXEC
)

func (s _txState) String() string {
	var state string
	switch s {
	case txStateNone:
		state = "txStateNone"
	case txStateWatch:
		state = "txStateWatch"
	case txStateSet:
		state = "txStateSet"
	case txStateMultiExec:
		state = "txStateMultiExec"
	default:
		state = "Unknown _txState"
	}
	return state
}

const (
	InitialTxPipelineSize int = 100
)

// TableSpec gives the name of the table, and other per-table customizations.
// (Eg: { Name: ACL_TABLE" }).
type TableSpec struct {
	Name string
}

// Key gives the key components.
// (Eg: { Comp : [] string { "acl1", "rule1" } } ).
type Key struct {
	Comp []string
}

// Value gives the fields as a map.
// (Eg: { Field: map[string]string { "type" : "l3v6", "ports" : "eth0" } } ).
type Value struct {
	Field map[string]string
}

// Table gives the entire table a a map.
// (Eg: { ts: &TableSpec{ Name: "ACL_TABLE" },
//        entry: map[string]Value {
//            "ACL_TABLE|acl1|rule1_1":  Value {
//                            Field: map[string]string {
//                              "type" : "l3v6", "ports" : "Ethernet0",
//                            }
//                          },
//            "ACL_TABLE|acl1|rule1_2":  Value {
//                            Field: map[string]string {
//                              "type" : "l3v6", "ports" : "eth0",
//                            }
//                          },
//                          }
//        })
type Table struct {
	ts    *TableSpec
	entry map[string]Value
	db    *DB
}

type _txCmd struct {
	ts    *TableSpec
	isSet bool
	key   *Key
	value *Value
}

// DB is the main type.
type DB struct {
	client *redis.Client
	Opts   *Options

	txState _txState
	txCmds  []_txCmd
}

func (d DB) String() string {
	return fmt.Sprintf("{ client: %v, Opts: %v, txState: %v, tsCmds: %v }",
		d.client, d.Opts, d.txState, d.txCmds)
}

// NewDB is the factory method to create new DB's.
func NewDB(opt Options) *DB {
	if glog.V(3) {
		glog.Info("NewDB: Begin: opt: ", opt)
	}

	d := DB{client: redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    DefaultRedisLocalTCPEP,
		//Addr:     DefaultRedisRemoteTCPEP,
		Password: "", /* TBD */
		// DB:       int(4), /* CONFIG_DB DB No. */
		DB:          int(opt.DBNo),
		DialTimeout: 0,
		// For Transactions, limit the pool
		PoolSize: 1,
		// Each DB gets it own (single) connection.
	}),
		Opts:    &opt,
		txState: txStateNone,
		txCmds:  make([]_txCmd, 0, InitialTxPipelineSize),
	}

	if d.client == nil {
		glog.Error("NewDB: Could not create redis client")
		goto NewDBExit
	}

	if opt.DBNo != ConfigDB {
		if glog.V(3) {
			glog.Info("NewDB: ! ConfigDB. Skip init. check.")
		}
		goto NewDBSkipInitIndicatorCheck
	}

	if len(d.Opts.InitIndicator) == 0 {

		glog.Info("NewDB: Init indication not requested")

	} else if init, _ := d.client.Get(d.Opts.InitIndicator).Int(); init != 1 {

		glog.Error("NewDB: Database not inited")
		goto NewDBExit
	}

NewDBSkipInitIndicatorCheck:

NewDBExit:

	if glog.V(3) {
		glog.Info("NewDB: End: d: ", d)
	}

	return &d
}

// DeleteDB is the gentle way to close the DB connection.
func (d *DB) DeleteDB() error {

	if glog.V(3) {
		glog.Info("DeleteDB: Begin: d: ", d)
	}

	if d.txState != txStateNone {
		glog.Warning("DeleteDB: not txStateNone, txState: ", d.txState)
	}

	return d.client.Close()
}

func (d *DB) key2redis(ts *TableSpec, key Key) string {

	if glog.V(5) {
		glog.Info("key2redis: Begin: ",
			ts.Name+
				d.Opts.TableNameSeparator+
				strings.Join(key.Comp, d.Opts.KeySeparator))
	}

	return ts.Name +
		d.Opts.TableNameSeparator +
		strings.Join(key.Comp, d.Opts.KeySeparator)
}

func (d *DB) redis2key(ts *TableSpec, redisKey string) Key {

	splitTable := strings.SplitN(redisKey, d.Opts.TableNameSeparator, 2)

	return Key{strings.Split(splitTable[1], d.Opts.KeySeparator)}

}

// GetEntry retrieves an entry(row) from the table.
func (d *DB) GetEntry(ts *TableSpec, key Key) (Value, error) {

	if glog.V(3) {
		glog.Info("GetEntry: Begin: ", "ts: ", ts, " key: ", key)
	}

	var value Value

	/*
		m := make(map[string]string)
		m["f0.0"] = "v0.0"
		m["f0.1"] = "v0.1"
		m["f0.2"] = "v0.2"
		v := Value{Field: m}
	*/

	v, e := d.client.HGetAll(d.key2redis(ts, key)).Result()

	if len(v) != 0 {
		value = Value{Field: v}
	} else {
		if glog.V(4) {
			glog.Info("GetEntry: HGetAll(): empty map")
		}
		e = errors.New("Entry does not exist")
	}

	if glog.V(3) {
		glog.Info("GetEntry: End: ", "value: ", value, " e: ", e)
	}

	return value, e
}

// GetKeys retrieves all entry/row keys.
func (d *DB) GetKeys(ts *TableSpec) ([]Key, error) {

	if glog.V(3) {
		glog.Info("GetKeys: Begin: ", "ts: ", ts)
	}

	/*
		k := []Key{
			{[]string{"k0.0", "k0.1"}},
			{[]string{"k1.0", "k1.1"}},
		}
	*/
	redisKeys, e := d.client.Keys(d.key2redis(ts,
		Key{Comp: []string{"*"}})).Result()
	if glog.V(4) {
		glog.Info("GetKeys: redisKeys: ", redisKeys, " e: ", e)
	}

	keys := make([]Key, 0, len(redisKeys))
	for i := 0; i < len(redisKeys); i++ {
		keys = append(keys, d.redis2key(ts, redisKeys[i]))
	}

	if glog.V(3) {
		glog.Info("GetKeys: End: ", "keys: ", keys, " e: ", e)
	}

	return keys, e
}

// DeleteKeys deletes all entry/row keys matching a pattern.
func (d *DB) DeleteKeys(ts *TableSpec, key Key) error {
	if glog.V(3) {
		glog.Info("DeleteKeys: Begin: ", "ts: ", ts, " key: ", key)
	}

	// This can be done via a LUA script as well. For now do this. TBD
	redisKeys, e := d.client.Keys(d.key2redis(ts, key)).Result()
	if glog.V(4) {
		glog.Info("DeleteKeys: redisKeys: ", redisKeys, " e: ", e)
	}

	for i := 0; i < len(redisKeys); i++ {
		if glog.V(4) {
			glog.Info("DeleteKeys: Deleting redisKey: ", redisKeys[i])
		}
		e = d.DeleteEntry(ts, d.redis2key(ts, redisKeys[i]))
		if e != nil {
			glog.Warning("DeleteKeys: Deleting: ts: ", ts, " key",
				d.redis2key(ts, redisKeys[i]), " : ", e)
		}
	}

	if glog.V(3) {
		glog.Info("DeleteKeys: End: e: ", e)
	}
	return e
}

func (d *DB) doSetDeleteEntry(ts *TableSpec, isSet bool, key Key, v interface{}) error {

	var e error = nil

	switch d.txState {
	case txStateNone:
		if glog.V(2) {
			glog.Info("doSetDeleteEntry: No Transaction.")
		}
		break
	case txStateWatch:
		if glog.V(2) {
			glog.Info("doSetDeleteEntry: Change to txStateSet, txState: ", d.txState)
		}
		d.txState = txStateSet
		break
	case txStateSet:
		if glog.V(5) {
			glog.Info("doSetDeleteEntry: Remain in txStateSet, txState: ", d.txState)
		}
	case txStateMultiExec:
		glog.Error("doSetDeleteEntry: Incorrect State, txState: ", d.txState)
		e = errors.New("Cannot issue {Set|Delete}Entry in txStateMultiExec")
	default:
		glog.Error("doSetDeleteEntry: Unknown, txState: ", d.txState)
		e = errors.New("Unknown State: " + string(d.txState))
	}

	if e != nil {
		goto doSetDeleteEntryExit
	}

	if d.txState == txStateNone {

		if isSet {

			val := make(map[string]interface{})
			for key, value := range v.(Value).Field {
				val[key] = value
			}
			e = d.client.Del(d.key2redis(ts, key)).Err()
			if e!= nil {
				glog.Error("doSetDeleteEntry: Del: ", key, " : ", e)
			}
			e = d.client.HMSet(d.key2redis(ts, key), val).Err()

		} else {

			e = d.client.Del(d.key2redis(ts, key)).Err()

		}

	} else if isSet {

		value := v.(Value)
		if glog.V(2) {
			glog.Info("doSetDeleteEntry: SetEntry ", key, " ", value)
		}
		d.txCmds = append(d.txCmds, _txCmd{
			ts:    ts,
			isSet: true,
			key:   &key,
			value: &value,
		})

	} else {

		if glog.V(2) {
			glog.Info("doSetDeleteEntry: DeleteEntry ", key)
		}
		d.txCmds = append(d.txCmds, _txCmd{
			ts:    ts,
			isSet: false,
			key:   &key,
			value: nil,
		})

	}

doSetDeleteEntryExit:

	if glog.V(3) {
		glog.Info("doSetDeleteEntry: End: e: ", e)
	}

	return e
}

// SetEntry sets an entry(row) in the table.
func (d *DB) SetEntry(ts *TableSpec, key Key, value Value) error {

	if glog.V(3) {
		glog.Info("SetEntry: Begin: ", "ts: ", ts, " key: ", key,
			" value: ", value)
	}

	return d.doSetDeleteEntry(ts, true, key, value)
}

// DeleteEntry deletes an entry(row) in the table.
func (d *DB) DeleteEntry(ts *TableSpec, key Key) error {

	if glog.V(3) {
		glog.Info("DeleteEntry: Begin: ", "ts: ", ts, " key: ", key)
	}

	return d.doSetDeleteEntry(ts, false, key, nil)
}

// ModEntry modifies an entry(row) in the table.
func (d *DB) ModEntry(ts *TableSpec, key Key, value Value) error {

	var e error

	if glog.V(3) {
		glog.Info("ModEntry: Begin: ", "ts: ", ts, " key: ", key,
			" value: ", value)
	}

	if len(value.Field) == 0 {
		glog.Info("ModEntry: Mapping to DeleteEntry()")
		e = d.DeleteEntry(ts, key)
	} else {
		// TBD return d.doSetDeleteEntry(ts, false, key, nil)
		glog.Error("ModEntry: Not Implemented!")
	}
	return e;
}

// DeleteEntryFields deletes some fields/columns in an entry(row) in the table.
func (d *DB) DeleteEntryFields(ts *TableSpec, key Key, value Value) error {

	if glog.V(3) {
		glog.Info("DeleteEntryFields: Begin: ", "ts: ", ts, " key: ", key,
			" value: ", value)
	}

	// TBD return d.doSetDeleteEntry(ts, false, key, nil)
	glog.Error("DeleteEntryFields: Not Implemented!")
	return nil;
}


// GetTable gets the entire table.
func (d *DB) GetTable(ts *TableSpec) (Table, error) {
	if glog.V(3) {
		glog.Info("GetTable: Begin: ts: ", ts)
	}

	/*
		table := Table{
			ts: ts,
			entry: map[string]Value{
				"table1|k0.0|k0.1": Value{
					map[string]string{
						"f0.0": "v0.0",
						"f0.1": "v0.1",
						"f0.2": "v0.2",
					},
				},
				"table1|k1.0|k1.1": Value{
					map[string]string{
						"f1.0": "v1.0",
						"f1.1": "v1.1",
						"f1.2": "v1.2",
					},
				},
			},
		        db: d,
		}
	*/

	// Create Table
	table := Table{
		ts:    ts,
		entry: make(map[string]Value),
		db:    d,
	}

	// This can be done via a LUA script as well. For now do this. TBD
	// Read Keys
	keys, e := d.GetKeys(ts)
	if e != nil {
		glog.Error("GetTable: GetKeys: " + e.Error())
		goto GetTableExit
	}

	// For each key in Keys
	// 	Add Value into table.entry[key)]
	for i := 0; i < len(keys); i++ {
		value, e := d.GetEntry(ts, keys[i])
		if e != nil {
			glog.Warning("GetTable: GetKeys: " + e.Error())
			continue
		}
		table.entry[d.key2redis(ts, keys[i])] = value
	}

GetTableExit:

	if glog.V(3) {
		glog.Info("GetTable: End: table: ", table)
	}
	return table, e
}

// DeleteTable deletes the entire table.
func (d *DB) DeleteTable(ts *TableSpec) error {
	if glog.V(3) {
		glog.Info("DeleteTable: Begin: ts: ", ts)
	}

	// This can be done via a LUA script as well. For now do this. TBD
	// Read Keys
	keys, e := d.GetKeys(ts)
	if e != nil {
		glog.Error("GetTable: GetKeys: " + e.Error())
		goto DeleteTableExit
	}

	// For each key in Keys
	// 	Delete the entry
	for i := 0; i < len(keys); i++ {
		e := d.DeleteEntry(ts, keys[i])
		if e != nil {
			glog.Warning("GetTable: GetKeys: " + e.Error())
			continue
		}
	}
DeleteTableExit:
	if glog.V(3) {
		glog.Info("DeleteTable: End: ")
	}
	return e
}

// GetKeys method retrieves all entry/row keys from a previously read table.
func (t *Table) GetKeys() ([]Key, error) {
	if glog.V(3) {
		glog.Info("Table.GetKeys: Begin: t: ", t)
	}
	keys := make([]Key, 0, len(t.entry))
	for k, _ := range t.entry {
		keys = append(keys, t.db.redis2key(t.ts, k))
	}

	if glog.V(3) {
		glog.Info("Table.GetKeys: End: keys: ", keys)
	}
	return keys, nil
}

// GetEntry method retrieves an entry/row from a previously read table.
func (t *Table) GetEntry(key Key) (Value, error) {
	/*
		return Value{map[string]string{
			"f0.0": "v0.0",
			"f0.1": "v0.1",
			"f0.2": "v0.2",
		},
		}, nil
	*/
	if glog.V(3) {
		glog.Info("Table.GetEntry: Begin: t: ", t, " key: ", key)
	}
	v := t.entry[t.db.key2redis(t.ts, key)]
	if glog.V(3) {
		glog.Info("Table.GetEntry: End: entry: ", v)
	}
	return v, nil
}

//===== Functions for db.Key =====

// Len returns number of components in the Key
func (k *Key) Len() int {
	return len(k.Comp)
}

// Get returns the key component at given index
func (k *Key) Get(index int) string {
	return k.Comp[index]
}

//===== Functions for db.Value =====

// Has function checks if a field exists.
func (v *Value) Has(name string) bool {
	_, flag := v.Field[name]
	return flag
}

// Get returns the value of a field. Returns empty string if the field
// does not exists. Use Has() function to check existance of field.
func (v *Value) Get(name string) string {
	return v.Field[name]
}

// Set function sets a string value for a field.
func (v *Value) Set(name, value string) {
	v.Field[name] = value
}

// GetInt returns value of a field as int. Returns 0 if the field does
// not exists. Returns an error if the field value is not a number.
func (v *Value) GetInt(name string) (int, error) {
	data, ok := v.Field[name]
	if ok {
		return strconv.Atoi(data)
	}
	return 0, nil
}

// SetInt sets an integer value for a field.
func (v *Value) SetInt(name string, value int) {
	v.Set(name, strconv.Itoa(value))
}

// GetList returns the value of a an array field. A "@" suffix is
// automatically appended to the field name if not present (as per
// swsssdk convention). Field value is split by comma and resulting
// slice is returned. Empty slice is returned if field not exists.
func (v *Value) GetList(name string) []string {
	var data string
	if strings.HasSuffix(name, "@") {
		data = v.Get(name)
	} else {
		data = v.Get(name + "@")
	}

	if len(data) == 0 {
		return []string{}
	}

	return strings.Split(data, ",")
}

// SetList function sets an list value to a field. Field name and
// value are formatted as per swsssdk conventions:
// - A "@" suffix is appended to key name
// - Field value is the comma separated string of list items
func (v *Value) SetList(name string, items []string) {
	if !strings.HasSuffix(name, "@") {
		name += "@"
	}

	if len(items) != 0 {
		data := strings.Join(items, ",")
		v.Set(name, data)
	} else {
		v.Remove(name)
	}
}

// Remove function removes a field from this Value.
func (v *Value) Remove(name string) {
	delete(v.Field, name)
}

//////////////////////////////////////////////////////////////////////////
// The Transaction API for translib infra
//////////////////////////////////////////////////////////////////////////

// WatchKeys is array of (TableSpec, Key) tuples to be watched in a Transaction.
type WatchKeys struct {
	Ts  *TableSpec
	Key *Key
}

func (w WatchKeys) String() string {
	return fmt.Sprintf("{ Ts: %v, Key: %v }", w.Ts, w.Key)
}

// StartTx method is used by infra to start a check-and-set Transaction.
func (d *DB) StartTx(w []WatchKeys) error {
	if glog.V(3) {
		glog.Info("StartTx: Begin: w: ", w)
	}

	var e error = nil
	var args []interface{}

	// Validate State
	if d.txState != txStateNone {
		glog.Error("StartTx: Incorrect State, txState: ", d.txState)
		e = errors.New("Transaction already in progress")
		goto StartTxExit
	}

	// For each watchkey
	//   If a pattern, Get the keys, appending results to Cmd args.
	//   Else append keys to the Cmd args
	//   Note: (LUA scripts do not support WATCH)

	if len(w) == 0 {
		glog.Warning("StartTx: Empty WatchKeys. Skipping WATCH")
		goto StartTxSkipWatch
	}

	args = make([]interface{}, 0, len(w)+1) // Init. est. with no wildcard
	args = append(args, "WATCH")
	for i := 0; i < len(w); i++ {

		redisKey := d.key2redis(w[i].Ts, *(w[i].Key))

		if !strings.Contains(redisKey, "*") {
			args = append(args, redisKey)
			continue
		}

		redisKeys, e := d.client.Keys(redisKey).Result()
		if e != nil {
			glog.Warning("StartTx: Keys: " + e.Error())
			continue
		}
		for j := 0; j < len(redisKeys); j++ {
			args = append(args, d.redis2key(w[i].Ts, redisKeys[j]))
		}
	}

	// Issue the WATCH
	_, e = d.client.Do(args...).Result()

	if e != nil {
		glog.Warning("StartTx: Do: WATCH ", args, " e: ", e.Error())
	}

StartTxSkipWatch:

	// Switch State
	d.txState = txStateWatch

StartTxExit:

	if glog.V(3) {
		glog.Info("StartTx: End: e: ", e)
	}
	return e
}

// CommitTx method is used by infra to commit a check-and-set Transaction.
func (d *DB) CommitTx() error {
	if glog.V(3) {
		glog.Info("CommitTx: Begin:")
	}

	var e error = nil

	// Validate State
	switch d.txState {
	case txStateNone:
		glog.Error("CommitTx: No WATCH done, txState: ", d.txState)
		e = errors.New("StartTx() not done. No Transaction active.")
	case txStateWatch:
		if glog.V(1) {
			glog.Info("CommitTx: No SET|DEL done, txState: ", d.txState)
		}
	case txStateSet:
		break
	case txStateMultiExec:
		glog.Error("CommitTx: Incorrect State, txState: ", d.txState)
		e = errors.New("Cannot issue MULTI in txStateMultiExec")
	default:
		glog.Error("CommitTx: Unknown, txState: ", d.txState)
		e = errors.New("Unknown State: " + string(d.txState))
	}

	if e != nil {
		goto CommitTxExit
	}

	// Issue MULTI
	_, e = d.client.Do("MULTI").Result()

	if e != nil {
		glog.Warning("CommitTx: Do: MULTI e: ", e.Error())
	}

	// For each cmd in txCmds
	//   Invoke it
	for i := 0; i < len(d.txCmds); i++ {
		// First Del
		redisKey := d.key2redis(d.txCmds[i].ts, *(d.txCmds[i].key))

		if glog.V(4) {
			glog.Info("CommitTx: Do: DEL ", redisKey)
		}

		d.client.Do("DEL", redisKey)

		if !d.txCmds[i].isSet {
			continue
		}

		// Second HMSet, if isSet
		args := make([]interface{}, 0, len(d.txCmds[i].value.Field)*2+2)
		args = append(args, "HMSET", redisKey)

		for key, value := range d.txCmds[i].value.Field {
			args = append(args, key, value)
		}

		if glog.V(4) {
			glog.Info("CommitTx: Do: ", args)
		}

		_, e = d.client.Do(args...).Result()

		if e != nil {
			glog.Warning("CommitTx: Do: ", args, " e: ", e.Error())
		}
	}

	// Issue EXEC
	_, e = d.client.Do("EXEC").Result()

	if e != nil {
		glog.Warning("CommitTx: Do: EXEC e: ", e.Error())
	}

	// Switch State, Clear Command list
	d.txState = txStateNone
	d.txCmds = d.txCmds[:0]

CommitTxExit:
	if glog.V(3) {
		glog.Info("CommitTx: End: e: ", e)
	}
	return e
}

// AbortTx method is used by infra to abort a check-and-set Transaction.
func (d *DB) AbortTx() error {
	if glog.V(3) {
		glog.Info("AbortTx: Begin:")
	}

	var e error = nil

	// Validate State
	switch d.txState {
	case txStateNone:
		glog.Error("AbortTx: No WATCH done, txState: ", d.txState)
		e = errors.New("StartTx() not done. No Transaction active.")
	case txStateWatch:
		if glog.V(1) {
			glog.Info("AbortTx: No SET|DEL done, txState: ", d.txState)
		}
	case txStateSet:
		break
	case txStateMultiExec:
		glog.Error("AbortTx: Incorrect State, txState: ", d.txState)
		e = errors.New("Cannot issue UNWATCH in txStateMultiExec")
	default:
		glog.Error("AbortTx: Unknown, txState: ", d.txState)
		e = errors.New("Unknown State: " + string(d.txState))
	}

	if e != nil {
		goto AbortTxExit
	}

	// Issue UNWATCH
	_, e = d.client.Do("UNWATCH").Result()

	if e != nil {
		glog.Warning("AbortTx: Do: UNWATCH e: ", e.Error())
	}

	// Switch State, Clear Command list
	d.txState = txStateNone
	d.txCmds = d.txCmds[:0]

AbortTxExit:
	if glog.V(3) {
		glog.Info("AbortTx: End: e: ", e)
	}
	return e
}
