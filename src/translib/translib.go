///////////////////////////////////////////////////
//
// Copyright 2019 Broadcom Inc.
//
///////////////////////////////////////////////////

/*
Package translib implements APIs like Create, Get, Subscribe etc.

to be consumed by the north bound management server implementations

This package take care of translating the incoming requests to 

Redis ABNF format and persisting them in the Redis DB.

It can also translate the ABNF format to YANG specific JSON IETF format

This package can also talk to non-DB clients.

Example: TBD

*/

package translib

import (
		//"errors"
		"sync"
		"translib/db"
        "github.com/openconfig/ygot/ygot"
        "github.com/Workiva/go-datastructures/queue"
        log "github.com/golang/glog"
)

//Write lock for all write operations to be synchronized
var writeMutex = &sync.Mutex{}

//minimum global interval for subscribe in secs
var minSubsInterval = 20

type ErrSource int

const(
	ProtoErr ErrSource = iota
	AppErr
)

type SetRequest struct{
    Path       string
    Payload    []byte
}

type SetResponse struct{
	ErrSrc     ErrSource
}

type GetRequest struct{
    Path       string
}

type GetResponse struct{
    Payload    []byte
	ErrSrc     ErrSource
}

type SubscribeResponse struct{
	Path		string
	Payload     []byte
	Timestamp	int64
}

type NotificationType int

const(
    Sample	NotificationType = iota
    OnChange
)

type IsSubscribeResponse struct{
	Path				string
	IsSupported			bool
	MinInterval			int
	Err					error
	PreferredType		NotificationType
}

type ModelData struct{
	Name      string
	Org		  string
	Ver		  string
}

type notificationOpts struct {
    mInterval		int
    pType			NotificationType  // for TARGET_DEFINED
}

//initializes logging and app modules
func init() {
    log.Flush()
}

//Creates entries in the redis DB pertaining to the path and payload
func Create(req SetRequest) (SetResponse, error){
    var ygotRoot *ygot.GoStruct
    var ygotTarget *interface{}
    var data appData
	var keys []db.WatchKeys
	var resp SetResponse

	path	:= req.Path
    payload := req.Payload

    log.Info("Create request received with path =", path)
    log.Info("Create request received with payload =", string(payload))

	app, appInfo, err := getAppModule(path)

	if err != nil {
		resp.ErrSrc = ProtoErr
        return resp, err
	}

    if appInfo.isNative {
        log.Info("Native MSFT format")
        data = appData{path: path, payload:payload}
        app.initialize(data)
    } else {
		ygotRoot, ygotTarget, err = getRequestBinder(&path, &payload, CREATE, &(appInfo.ygotRootType)).unMarshall()
		if err != nil {
			log.Info("Error in request binding in the create request: ", err)
			resp.ErrSrc = AppErr
			return resp, err
		}

		data = appData{path: path, ygotRoot: ygotRoot, ygotTarget: ygotTarget}
		app.initialize(data)
	}

	writeMutex.Lock()

	d, err := db.NewDB(getDBOptions(db.ConfigDB))

	if err != nil {
		writeMutex.Unlock()
		resp.ErrSrc = ProtoErr
		return resp, err
	}

	defer d.DeleteDB()

    keys, err = app.translateCreate(d)

	if err != nil {
		writeMutex.Unlock()
		resp.ErrSrc = AppErr
        return resp, err
	}

	err = d.StartTx(keys, appInfo.tablesToWatch)

	if err != nil {
		writeMutex.Unlock()
		resp.ErrSrc = AppErr
        return resp, err
	}

    resp, err = app.processCreate (d)

    if err != nil {
		writeMutex.Unlock()
		d.AbortTx()
		resp.ErrSrc = AppErr
        return resp, err
    }

	err = d.CommitTx()

    if err != nil {
        resp.ErrSrc = AppErr
    }

	writeMutex.Unlock()

    return resp, err
}

//Updates entries in the redis DB pertaining to the path and payload
func Update(req SetRequest) (SetResponse, error){
    var ygotRoot *ygot.GoStruct
    var ygotTarget *interface{}
    var data appData
    var keys []db.WatchKeys
	var resp SetResponse

    path    := req.Path
    payload := req.Payload

    log.Info("Update request received with path =", path)
    log.Info("Update request received with payload =", string(payload))

	app, appInfo, err := getAppModule(path)

    if err != nil {
		resp.ErrSrc = ProtoErr
        return resp, err
    }

    if appInfo.isNative {
        log.Info("Native MSFT format")
        data = appData{path: path, payload:payload}
        app.initialize(data)
    } else {
        ygotRoot, ygotTarget, err = getRequestBinder(&path, &payload, UPDATE, &(appInfo.ygotRootType)).unMarshall()
        if err != nil {
            log.Info("Error in request binding in the update request: ", err)
			resp.ErrSrc = AppErr
            return resp, err
        }

        data = appData{path: path, ygotRoot: ygotRoot, ygotTarget: ygotTarget}
        app.initialize(data)
    }

    writeMutex.Lock()

    d, err := db.NewDB(getDBOptions(db.ConfigDB))

    if err != nil {
        writeMutex.Unlock()
        resp.ErrSrc = ProtoErr
        return resp, err
    }

	defer d.DeleteDB()

    keys, err = app.translateUpdate(d)

    if err != nil {
        writeMutex.Unlock()
		resp.ErrSrc = AppErr
        return resp, err
    }

    err = d.StartTx(keys, appInfo.tablesToWatch)

    if err != nil {
        writeMutex.Unlock()
		resp.ErrSrc = AppErr
        return resp, err
    }

    resp, err = app.processUpdate (d)

    if err != nil {
        writeMutex.Unlock()
        d.AbortTx()
		resp.ErrSrc = AppErr
        return resp, err
    }

    err = d.CommitTx()

    if err != nil {
        resp.ErrSrc = AppErr
    }

    writeMutex.Unlock()
    return resp, err
}

//Replaces entries in the redis DB pertaining to the path and payload
func Replace(req SetRequest) (SetResponse, error){
    var err error
    var app appInterface
    var ygotRoot *ygot.GoStruct
    var ygotTarget *interface{}
    var data appData
    var keys []db.WatchKeys
	var resp SetResponse

    path    := req.Path
    payload := req.Payload

    log.Info("Replace request received with path =", path)
    log.Info("Replace request received with payload =", string(payload))

	app, appInfo, err := getAppModule(path)

    if err != nil {
		resp.ErrSrc = ProtoErr
        return resp, err
    }

    if appInfo.isNative {
        log.Info("Native MSFT format")
        data = appData{path: path, payload:payload}
        app.initialize(data)
    } else {
        ygotRoot, ygotTarget, err = getRequestBinder(&path, &payload, REPLACE, &(appInfo.ygotRootType)).unMarshall()
        if err != nil {
            log.Info("Error in request binding in the replace request: ", err)
			resp.ErrSrc = AppErr
            return resp, err
        }

        data = appData{path: path, ygotRoot: ygotRoot, ygotTarget: ygotTarget}
        app.initialize(data)
    }

    writeMutex.Lock()

    d, err := db.NewDB(getDBOptions(db.ConfigDB))

    if err != nil {
        writeMutex.Unlock()
        resp.ErrSrc = ProtoErr
        return resp, err
    }

	defer d.DeleteDB()

    keys, err = app.translateReplace(d)

    if err != nil {
        writeMutex.Unlock()
		resp.ErrSrc = AppErr
        return resp, err
    }

    err = d.StartTx(keys, appInfo.tablesToWatch)

    if err != nil {
        writeMutex.Unlock()
		resp.ErrSrc = AppErr
        return resp, err
    }

    resp, err = app.processReplace (d)

    if err != nil {
        writeMutex.Unlock()
        d.AbortTx()
		resp.ErrSrc = AppErr
        return resp, err
    }

    err = d.CommitTx()

	if err != nil {
		resp.ErrSrc = AppErr
    }

    writeMutex.Unlock()
    return resp, err
}

//Deletes entries in the redis DB pertaining to the path
func Delete(req SetRequest) (SetResponse, error){
    var err error
    var app appInterface
    var ygotRoot *ygot.GoStruct
    var ygotTarget *interface{}
    var data appData
    var keys []db.WatchKeys
	var resp SetResponse

    path    := req.Path

    log.Info("Delete request received with path =", path)

	app, appInfo, err := getAppModule(path)

    if err != nil {
		resp.ErrSrc = ProtoErr
        return resp, err
    }

    if appInfo.isNative {
        log.Info("Native MSFT format")
        data = appData{path: path}
        app.initialize(data)
    } else {
        ygotRoot, ygotTarget, err = getRequestBinder(&path, nil, DELETE, &(appInfo.ygotRootType)).unMarshall()
        if err != nil {
            log.Info("Error in request binding in the delete request: ", err)
			resp.ErrSrc = AppErr
            return resp, err
        }

        data = appData{path: path, ygotRoot: ygotRoot, ygotTarget: ygotTarget}
        app.initialize(data)
    }

    writeMutex.Lock()

    d, err := db.NewDB(getDBOptions(db.ConfigDB))

    if err != nil {
        writeMutex.Unlock()
        resp.ErrSrc = ProtoErr
        return resp, err
    }

	defer d.DeleteDB()

    keys, err = app.translateDelete(d)

    if err != nil {
        writeMutex.Unlock()
		resp.ErrSrc = AppErr
        return resp, err
    }

    err = d.StartTx(keys, appInfo.tablesToWatch)

    if err != nil {
        writeMutex.Unlock()
		resp.ErrSrc = AppErr
        return resp, err
    }

    resp, err = app.processDelete(d)

    if err != nil {
        writeMutex.Unlock()
        d.AbortTx()
		resp.ErrSrc = AppErr
        return resp, err
    }

    err = d.CommitTx()

    if err != nil {
        resp.ErrSrc = AppErr
    }

    writeMutex.Unlock()
	return resp, err
}

//Gets data from the redis DB and converts it to northbound format
func Get(req GetRequest) (GetResponse, error){
    var payload []byte
    var data appData
	var resp GetResponse

	path := req.Path

    log.Info("Received Get request for path = ",path)

	app, appInfo, err := getAppModule(path)

    if err != nil {
        resp = GetResponse{Payload:payload, ErrSrc:ProtoErr}
        return resp, err
    }

    if appInfo.isNative {
        log.Info("Native MSFT format")
        data = appData{path: path}
        app.initialize(data)
    } else {
       ygotStruct, ygotTarget, err := getRequestBinder (&path, nil, GET, &(appInfo.ygotRootType)).unMarshall()
        if err != nil {
                log.Info("Error in request binding: ", err)
				resp = GetResponse{Payload:payload, ErrSrc:AppErr}
                return resp, err
        }

        data = appData{path: path, ygotRoot: ygotStruct, ygotTarget: ygotTarget}
        app.initialize(data)
    }

	dbs, err := getAllDbs()

	if err != nil {
		resp = GetResponse{Payload:payload, ErrSrc:ProtoErr}
        return resp, err
	}

	defer closeAllDbs(dbs[:])

    err = app.translateGet (dbs)

	if err != nil {
		resp = GetResponse{Payload:payload, ErrSrc:AppErr}
        return resp, err
	}

    resp, err = app.processGet(dbs)

    return resp, err
}

//Subscribes to the paths requested and sends notifications when the data changes in DB
func Subscribe(paths []string, q *queue.PriorityQueue, stop chan struct{}) ([]*IsSubscribeResponse, error) {
    var err error
	var sErr error
	//err = errors.New("Not implemented")

	dbNotificationMap := make(map[db.DBNum][]*notificationInfo)

	resp := make ([]*IsSubscribeResponse, len(paths))

    for i, _ := range resp {
        resp[i] = &IsSubscribeResponse{Path: paths[i],
                                IsSupported: false,
                                MinInterval: minSubsInterval}
    }

	dbs, err := getAllDbs()

    if err != nil {
        return resp, err
    }

	//Do NOT close the DBs here as we need to use them during subscribe notification

    for i, path := range paths {

		app, _, err := getAppModule(path)

        if err != nil {

            if sErr == nil {
				sErr = err
			}

			resp[i].Err = err
			continue
        }

        nOpts, nInfo, errApp := app.translateSubscribe (dbs, path)

        if errApp != nil {
            resp[i].Err = errApp

			if sErr == nil {
				sErr = errApp
			}

            continue
        } else {
            resp[i].IsSupported = true

            if nOpts.mInterval != 0 {
                resp[i].MinInterval = nOpts.mInterval
            }

            resp[i].PreferredType = nOpts.pType

			dbNotificationMap[nInfo.dbno] = append(dbNotificationMap[nInfo.dbno], nInfo)
        }

	}

	log.Info("map=", dbNotificationMap)

	if sErr != nil {
		return resp, sErr
	}

	sInfo := &subscribeInfo {syncDone:false,
					nInfo:make([]*notificationInfo,0),
					stop:stop,
					dbs:dbs}

	for _, value := range dbNotificationMap {
		sInfo.nInfo = append(sInfo.nInfo, value...)
		startSubscribe(value, sInfo)
	}

	go runSubscribe(q)

	return resp, sErr
}

//Check if subscribe is supported on the given paths
func IsSubscribeSupported(paths []string) ([]*IsSubscribeResponse, error) {

	resp := make ([]*IsSubscribeResponse, len(paths))

	for i, _ := range resp {
		resp[i] = &IsSubscribeResponse{Path: paths[i],
								IsSupported: false,
								MinInterval: minSubsInterval}
	}

	dbs, err := getAllDbs()

    if err != nil {
        return resp, err
    }

	defer closeAllDbs(dbs[:])

	for i, path := range paths {

		app, _, err := getAppModule(path)

		if err != nil {
			resp[i].Err = err
			continue
		}

		nOpts, _, errApp := app.translateSubscribe (dbs, path)

		if errApp != nil {
			resp[i].Err = errApp
			err = errApp
            continue
        } else {
			resp[i].IsSupported = true

			if nOpts.mInterval != 0 {
				resp[i].MinInterval = nOpts.mInterval
			}
			resp[i].PreferredType = nOpts.pType
		}
	}

	return resp, err
}

//Gets all the models supported by Translib
func GetModels() ([]ModelData, error) {
	var err error

	return getModels(), err
}

//Creates connection will all the redis DBs. To be used for get request
func getAllDbs() ([db.MaxDB]*db.DB, error) {
	var dbs [db.MaxDB]*db.DB
    var err error

	//Create Application DB connection
    dbs[db.ApplDB], err = db.NewDB(getDBOptions(db.ApplDB))

	if err != nil {
		closeAllDbs(dbs[:])
		return dbs, err
	}

    //Create ASIC DB connection
    dbs[db.AsicDB], err = db.NewDB(getDBOptions(db.AsicDB))

	if err != nil {
		closeAllDbs(dbs[:])
		return dbs, err
	}

	//Create Counter DB connection
    dbs[db.CountersDB], err = db.NewDB(getDBOptions(db.CountersDB))

	if err != nil {
		closeAllDbs(dbs[:])
		return dbs, err
	}

	//Create Log Level DB connection
    dbs[db.LogLevelDB], err = db.NewDB(getDBOptions(db.LogLevelDB))

	if err != nil {
		closeAllDbs(dbs[:])
		return dbs, err
	}

	//Create Config DB connection
    dbs[db.ConfigDB], err = db.NewDB(getDBOptions(db.ConfigDB))

	if err != nil {
		closeAllDbs(dbs[:])
		return dbs, err
	}

	//Create State DB connection
    dbs[db.StateDB], err = db.NewDB(getDBOptions(db.StateDB))

	if err != nil {
		closeAllDbs(dbs[:])
		return dbs, err
	}

	return dbs, err
}

//Closes the dbs, and nils out the arr.
func closeAllDbs(dbs []*db.DB) {
	for dbsi, d := range dbs {
		if d != nil {
			d.DeleteDB()
			dbs[dbsi] = nil
		}
	}
}

// Implement Compare method for priority queue for SubscribeResponse struct
func (val SubscribeResponse) Compare(other queue.Item) int {
	o := other.(*SubscribeResponse)
	if val.Timestamp > o.Timestamp {
		return 1
	} else if val.Timestamp == o.Timestamp {
		return 0
	}
	return -1
}

func getDBOptions(dbNo db.DBNum) db.Options {

	var opt db.Options

	switch dbNo {
	case db.ApplDB:
		opt = db.Options {
                    DBNo              : dbNo,
                    InitIndicator     : "",
                    TableNameSeparator: ":",
                    KeySeparator      : ":",
                      }
	case db.AsicDB:
		opt = db.Options {
                    DBNo              : dbNo,
                    InitIndicator     : "",
                    TableNameSeparator: "|",
                    KeySeparator      : "|",
                      }
	case db.CountersDB:
		opt =  db.Options {
                    DBNo              : dbNo,
                    InitIndicator     : "",
                    TableNameSeparator: "|",
                    KeySeparator      : "|",
                      }
	case db.LogLevelDB:
		opt = db.Options {
                    DBNo              : dbNo,
                    InitIndicator     : "",
                    TableNameSeparator: "|",
                    KeySeparator      : "|",
                      }
	case db.ConfigDB:
		opt =  db.Options {
                    DBNo              : dbNo,
                    InitIndicator     : "",
                    TableNameSeparator: "|",
                    KeySeparator      : "|",
                      }
	case db.FlexCounterDB:
		opt = db.Options {
                    DBNo              : dbNo,
                    InitIndicator     : "",
                    TableNameSeparator: "|",
                    KeySeparator      : "|",
                      }
	case db.StateDB:
		opt = db.Options {
                    DBNo              : dbNo,
                    InitIndicator     : "",
                    TableNameSeparator: "|",
                    KeySeparator      : "|",
                      }
	}

	return opt
}

func getAppModule (path string) (appInterface, appInfo, error) {
	var app appInterface

    aInfo, err := getAppModuleInfo(path)

    if err != nil {
        return nil, aInfo, err
    }

    app, err = getAppInterface(aInfo.appType)

    if err != nil {
        return nil, aInfo, err
    }

	return app, aInfo, err
}
