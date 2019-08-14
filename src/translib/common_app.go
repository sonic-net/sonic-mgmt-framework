package translib

import (
//	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
//	"github.com/openconfig/ygot/util"
	"github.com/openconfig/ygot/ygot"
	"reflect"
//	"strconv"
//	"strings"
	"translib/db"
	//"translib/ocbinds"
	"translib/transformer"
)

var (

)
type CommonApp struct {
	pathInfo *PathInfo
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}
}

var cmnAppInfo = appInfo{appType: reflect.TypeOf(CommonApp{}),
                        ygotRootType:  nil,
                        isNative:      false,
                        tablesToWatch: nil}


func init() {

    // @todo : Optimize to register supported paths/yang via common app and report error for unsupported
    register_model_path := []string{"/sonic-"} // register yang model path(s) to be supported via common app
    for _, mdl_pth := range register_model_path {
        err := register(mdl_pth, &cmnAppInfo)

        if err != nil {
		    log.Fatal("Register Common app module with App Interface failed with error=", err, "for path=", mdl_pth)
	    }
    }

}

func (app *CommonApp) initialize(data appData) {
	log.Info("initialize:path =", data.path)
	pathInfo := NewPathInfo(data.path)
	*app = CommonApp{pathInfo: pathInfo, ygotRoot: data.ygotRoot, ygotTarget: data.ygotTarget}

}

func (app *CommonApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCreate:path =", app.pathInfo.Path)

	keys, err = app.translateCRUCommon(d, CREATE)

	return keys, err
}

func (app *CommonApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateUpdate:path =", app.pathInfo.Path)

	keys, err = app.translateCRUCommon(d, UPDATE)

	return keys, err
}

func (app *CommonApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateReplace:path =", app.pathInfo.Path)

	//keys, err = app.translateCRUCommon(d, REPLACE)

	err = errors.New("Not implemented")
	return keys, err
}

func (app *CommonApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateDelete:path =", app.pathInfo.Path)

	keys, err = app.generateDbWatchKeys(d, true)

	return keys, err
}

func (app *CommonApp) translateGet(dbs [db.MaxDB]*db.DB) (*map[db.DBNum][]transformer.KeySpec, error) {
        var err error
        log.Info("translateGet:path =", app.pathInfo.Path)

        keySpec, err := transformer.XlateUriToKeySpec(app.pathInfo.Path, app.ygotRoot, app.ygotTarget)

        return keySpec, err
}

func (app *CommonApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
    err := errors.New("Not supported")
    //configDb := dbs[db.ConfigDB]
    //pathInfo := NewPathInfo(path)
    notifInfo := notificationInfo{dbno: db.ConfigDB}
    return nil, &notifInfo, err
}

func (app *CommonApp) processCreate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processCreate:path =", app.pathInfo.Path)
	targetType := reflect.TypeOf(*app.ygotTarget)
	log.Infof("processCreate: Target object is a <%s> of Type: %s", targetType.Kind().String(), targetType.Elem().Name())


	return resp, err
}

func (app *CommonApp) processUpdate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processUpdate:path =", app.pathInfo.Path)

	return resp, err
}

func (app *CommonApp) processReplace(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processReplace:path =", app.pathInfo.Path)
	err = errors.New("Not implemented")
	return resp, err
}

func (app *CommonApp) processDelete(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processDelete:path =", app.pathInfo.Path)

	//aclObj := app.getAppRootObject()

	//targetUriPath, err := getYangPathFromUri(app.pathInfo.Path)


	return resp, err
}

func (app *CommonApp) processGet(dbs [db.MaxDB]*db.DB, keyspec *map[db.DBNum][]transformer.KeySpec) (GetResponse, error) {
    var err error
    var payload []byte
    log.Info("processGet:path =", app.pathInfo.Path)

    // table.key.fields
    var result = make(map[string]map[string]db.Value)

    for dbnum, specs := range *keyspec {
        for _, spec := range specs {
            err := transformer.TraverseDb(dbs[dbnum], spec, &result, nil)
            if err != nil {
                return GetResponse{Payload: payload}, err
            }
        }
    }

    payload, err = transformer.XlateFromDb(result)
    if err != nil {
        return GetResponse{Payload: payload, ErrSrc: AppErr}, err
    }

    return GetResponse{Payload: payload}, err
}


func (app *CommonApp) translateCRUCommon(d *db.DB, opcode int) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	var tblsToWatch []*db.TableSpec
	log.Info("translateCRUCommon:path =", app.pathInfo.Path)

	// translate yang to db
	result, err := transformer.XlateToDb(app.pathInfo.Path, (*app).ygotRoot, (*app).ygotTarget)
	fmt.Println(result)
	log.Info("transformer.XlateToDb() returned", result)

	if err != nil {
		log.Error(err)
		return keys, err
	}
	if len(result) == 0 {
		log.Error("XlatetoDB() returned empty map")
		fmt.Println("XlatetoDB() returned empty map")
	}
	for tblnm, _  := range result {
           log.Error("Table name ", tblnm)
           tblsToWatch = append(tblsToWatch, &db.TableSpec{Name: tblnm})
        }
        log.Info("Tables to watch", tblsToWatch)

        cmnAppInfo.tablesToWatch = tblsToWatch

	keys, err = app.generateDbWatchKeys(d, false)

	return keys, err
}

func (app *CommonApp) generateDbWatchKeys(d *db.DB, isDeleteOp bool) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys

	return keys, err
}
