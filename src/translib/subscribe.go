///////////////////////////////////////////////////
//
// Copyright 2019 Broadcom Inc.
//
///////////////////////////////////////////////////

/*
Package translib defines the functions to be used by the subscribe 

handler to subscribe for a key space notification. It also has

functions to handle the key space notification from redis and

call the appropriate app module to handle them.

*/

package translib

import (
	"sync"
	"time"
	"translib/db"
	log "github.com/golang/glog"
	"github.com/Workiva/go-datastructures/queue"
)

//Subscribe mutex for all the subscribe operations on the maps to be thread safe
var subscribeMutex = &sync.Mutex{}

type notificationInfo struct{
	table               db.TableSpec
	key					db.Key
	dbno				db.DBNum
	needCache			bool
	path				string
	app					*appInterface
	cache				[]byte
	sDB					*db.DB //Subscription DB should be used only for keyspace notification
}

type subscribeInfo struct{
	syncDone			bool
	q				   *queue.PriorityQueue
	nInfo				[]*notificationInfo
	stop				chan struct{}
	dbs [db.MaxDB]	   *db.DB //used to perform get operations
}

var notificationMap map[*db.SKey]*notificationInfo
var subscribeMap map[*notificationInfo]*subscribeInfo
var stopMap map[chan struct{}]*subscribeInfo

func init() {
	notificationMap = make(map[*db.SKey]*notificationInfo)
	subscribeMap = make(map[*notificationInfo]*subscribeInfo)
	stopMap	= make(map[chan struct{}]*subscribeInfo)
}

func runSubscribe(q *queue.PriorityQueue) error {
	var err error

	for i := 0; i < 10; i++ {
		time.Sleep(2 * time.Second)
		q.Put(&SubscribeResponse{
				Path:"/testPath",
				Payload:[]byte("test payload"),
				Timestamp:    time.Now().UnixNano(),
		})

	}

	return err
}

func startSubscribe(opt db.Options, nInfoList []*notificationInfo, sInfo *subscribeInfo) error {
	var skey *db.SKey
	var sKeyList []*db.SKey
	var nInfo *notificationInfo

	subscribeMutex.Lock()

	for _, nInfo = range nInfoList {
		skey = &db.SKey{ Ts: &nInfo.table, Key: &nInfo.key}
		sKeyList = append(sKeyList, skey)
		notificationMap[skey] = nInfo
		subscribeMap[nInfo] = sInfo
	}

	sDB, err := db.SubscribeDB(opt, sKeyList, notificationHandler)

	if err == nil {
		for _, nInfo = range nInfoList {
			nInfo.sDB = sDB
		}
	}

	subscribeMutex.Unlock()
	return err
}

func notificationHandler(d *db.DB, skey *db.SKey, key *db.Key, event db.SEvent) error {
    log.Info("notificationHandler: d: ", d, " skey: ", *skey, " key: ", *key,
        " event: ", event)
    return nil
}

