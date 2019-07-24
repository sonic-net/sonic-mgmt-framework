/*
Copyright 2019 Broadcom. All rights reserved.
The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
*/

/*
Package db implements a wrapper over the go-redis/redis.
*/
package db

import (
	// "fmt"
	// "strconv"

	//	"reflect"
	"errors"
	"strings"

	// "github.com/go-redis/redis"
	"github.com/golang/glog"
	// "cvl"
	"translib/tlerr"
)

// SKey is (TableSpec, Key, []SEvent) 3-tuples to be watched in a Transaction.
type SKey struct {
	Ts  *TableSpec
	Key *Key
	SEMap map[SEvent]bool	// nil map indicates subscribe to all
}

type SEvent int

const (
	SEventNone      SEvent = iota // No Op
	SEventHSet   // HSET, HMSET, and its variants
	SEventHDel   // HDEL, also SEventDel generated, if HASH is becomes empty
	SEventDel    // DEL, & also if key gets deleted (empty HASH, expire,..)
	SEventOther  // Some other command not covered above.
	SEventClose  // Close requested due to Unsubscribe() called.
	SEventErr    // Error condition. Call Unsubscribe, after return.
)

var redisPayload2sEventMap map[string]SEvent = map[string]SEvent {
	""      : SEventNone,
	"hset"  : SEventHSet,
	"hdel"  : SEventHDel,
	"del"   : SEventDel,
}


func init() {
    // Optimization: Start the goroutine that is scanning the SubscribeDB
    // channels. Instead of one goroutine per Subscribe.
}

// func map[DB]


// func handler(d *DB, skey SKey, key Key, event SEvent) error {
// }

// HFunc gives the name of the table, and other per-table customizations.
type HFunc func( *DB, *SKey, *Key, SEvent) (error)


// SubscribeDB is the factory method to create a subscription to the DB.
// The returned handle can only be used for Subscription.
func SubscribeDB(opt Options, skeys []SKey, handler HFunc) (*DB, error) {

// func SubscribeDB(opt Options, skeys []SKey, handler func (*DB, WatchKeys, Value) (error)) (*DB, error) {

	if glog.V(3) {
		glog.Info("SubscribeDB: Begin: opt: ", opt, " skeys: ", skeys,
			" handler: ", handler)
	}

	patterns := make([]string, len(skeys))
	patMap := make(map[string]int, len(skeys))
	var s string

	// NewDB
	d , e := NewDB(opt)

	if d.client == nil {
		goto SubscribeDBExit
	}

	// Make sure that the DB is configured for key space notifications
	// Optimize with LUA scripts to atomically add "Kgshxe".
	s, e = d.client.ConfigSet("notify-keyspace-events", "AKE").Result()

	if e != nil {
		glog.Error("SubscribeDB: ConfigSet(): e: ", e, " s: ", s)
		goto SubscribeDBExit
	}

	for i:=0 ; i < len(skeys); i++ {
		patterns[i] = d.key2redisChannel(skeys[i].Ts, *(skeys[i].Key))
		patMap[patterns[i]] = i
	}

	glog.Info("SubscribeDB: patterns: ", patterns)

	d.sPubSub = d.client.PSubscribe(patterns[:]...)

	if d.sPubSub == nil {
		glog.Error("SubscribeDB: PSubscribe() nil: pats: ", patterns)
		e = tlerr.TranslibDBSubscribeFail { }
		goto SubscribeDBExit
	}

	// Wait for confirmation, of channel creation
	_, e = d.sPubSub.Receive()

	if e != nil {
		glog.Error("SubscribeDB: Receive() fails: e: ", e)
		e = tlerr.TranslibDBSubscribeFail { }
		goto SubscribeDBExit
	}


	// Get channel for receiving messages

	d.sCh = d.sPubSub.Channel()

	// Keep a copy in our context.
	d.sKeys = skeys
	d.sHandler = handler

	// Start a goroutine to read messages and call handler.
	go func() {
		for msg := range d.sCh {
			glog.Info("SubscribeDB: msg: ", msg, " Pattern: ", msg.Pattern)

			// Should this be a goroutine, in case each notification CB
			// takes a long time to run ?
			skey := &skeys[patMap[msg.Pattern]]
			key := d.redisChannel2key(skey.Ts, msg.Channel)

			d.sHandler(d, skey, &key, d.redisPayload2sEvent(msg.Payload))
		}

		// Send the Close|Err notification.
		var sEvent = SEventClose
		if d.sCIP == false {
			sEvent = SEventErr
		}
		glog.Info("SubscribeDB: SEventClose|Err: ", sEvent)
		d.sHandler(d, & SKey{}, & Key {}, sEvent)
	} ()


SubscribeDBExit:

	if e != nil {
		if d.sPubSub != nil {
		}

		if d.client != nil {
			d.DeleteDB()
			d.client = nil
		}
	}

	if glog.V(3) {
		glog.Info("SubscribeDB: End: d: ", d, " e: ", e)
	}

	return d, e
}

// UnsubscribeDB is used to close a DB subscription
func (d * DB) UnsubscribeDB() error {

	var e error = nil

	if glog.V(3) {
		glog.Info("UnsubscribeDB: d:", d)
	}

	if d.sCIP {
		glog.Error("UnsubscribeDB: Close in Progress")
		e = errors.New("UnsubscribeDB: Close in Progress")
		goto UnsubscribeDBExit
	}

	// Mark close in progress.
	d.sCIP = true;

	// Do the close, ch gets closed too.
	d.sPubSub.Close()

	// Wait for the goroutine to complete ? TBD
	// Should not this happen because of the range statement on ch?

	// Close the DB
	d.DeleteDB()

UnsubscribeDBExit:

	if glog.V(3) {
		glog.Info("UnsubscribeDB: End: d: ", d, " e: ", e)
	}

	return e
}


func (d *DB) key2redisChannel(ts *TableSpec, key Key) string {

	if glog.V(5) {
		glog.Info("key2redisChannel: ", *ts, " key: " + key.String())
	}

	return "__keyspace@" + (d.Opts.DBNo).String() + "__:" + d.key2redis(ts, key)
}

func (d *DB) redisChannel2key(ts *TableSpec, redisChannel string) Key {

	if glog.V(5) {
		glog.Info("redisChannel2key: ", *ts, " redisChannel: " + redisChannel)
	}

	splitRedisKey := strings.SplitN(redisChannel, ":", 2)

	if len(splitRedisKey) > 1 {
		return d.redis2key(ts, splitRedisKey[1])
	}

	glog.Warning("redisChannel2key: Missing key: redisChannel: ", redisChannel)

	return Key{}
}

func (d *DB) redisPayload2sEvent(redisPayload string) SEvent {

	if glog.V(5) {
		glog.Info("redisPayload2sEvent: ", redisPayload)
	}

	sEvent := redisPayload2sEventMap[redisPayload]

	if sEvent == 0 {
		sEvent = SEventOther
	}

	if glog.V(3) {
		glog.Info("redisPayload2sEvent: ", sEvent)
	}

    return sEvent
}

