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
	"time"
	"translib/db"
	"github.com/Workiva/go-datastructures/queue"
)

type notificationInfo struct{
	table               db.TableSpec 
	key					db.Key
	dbno				db.DBNum
	isLeaf				bool
	path				string
	app					*appInterface
	cache				[]byte
	sDB					*db.DB
}

type subscribeInfo struct{
	syncDone			bool
	nInfo				[]notificationInfo
	stop				chan struct{}
}

type notificationMap map[*db.DB][]*subscribeInfo
type stopMap map[chan struct{}][]*db.DB

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
