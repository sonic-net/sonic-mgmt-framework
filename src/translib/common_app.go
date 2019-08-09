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
	path       string
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}
}

func init() {

    // @todo : Optimize to register supported paths/yang via common app and report error for unsupported
    register_model_path := []string{"/sonic-"} // register yang model path(s) to be supported via common app
    for _, mdl_pth := range register_model_path {
        err := register(mdl_pth,
                        &appInfo{appType:  reflect.TypeOf(CommonApp{}),
                        ygotRootType:  nil, //interface{},
                        isNative:      false,
                        tablesToWatch: nil})

        if err != nil {
		    log.Fatal("Register Common app module with App Interface failed with error=", err, "for path=", mdl_pth)
	    }
    }

    // @todo : Add support for all yang models that will use common app
    //err = addModel(&ModelData{})
    yangFiles := []string{"sonic-acl.yang"}
    log.Info("Init transformer yang files :", yangFiles)
    err := transformer.LoadYangModules(yangFiles...)
    if err != nil {
        log.Fatal("Common App - Transformer call for loading yang modules failed with error=", err)
    }
}

func (app *CommonApp) initialize(data appData) {
	log.Info("initialize:path =", data.path)
	*app = CommonApp{path: data.path, ygotRoot: data.ygotRoot, ygotTarget: data.ygotTarget}

}

func (app *CommonApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCreate:path =", app.path)

	keys, err = app.translateCRUCommon(d, CREATE)

	return keys, err
}

func (app *CommonApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateUpdate:path =", app.path)

	keys, err = app.translateCRUCommon(d, UPDATE)

	return keys, err
}

func (app *CommonApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateReplace:path =", app.path)

	//keys, err = app.translateCRUCommon(d, REPLACE)

	err = errors.New("Not implemented")
	return keys, err
}

func (app *CommonApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateDelete:path =", app.path)

	keys, err = app.generateDbWatchKeys(d, true)

	return keys, err
}

func (app *CommonApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	var err error
	log.Info("translateGet:path =", app.path)
	return err
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

	log.Info("processCreate:path =", app.path)
	targetType := reflect.TypeOf(*app.ygotTarget)
	log.Infof("processCreate: Target object is a <%s> of Type: %s", targetType.Kind().String(), targetType.Elem().Name())


	return resp, err
}

func (app *CommonApp) processUpdate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processUpdate:path =", app.path)

	return resp, err
}

func (app *CommonApp) processReplace(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processReplace:path =", app.path)
	err = errors.New("Not implemented")
	return resp, err
}

func (app *CommonApp) processDelete(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processDelete:path =", app.path)

	//aclObj := app.getAppRootObject()

	//targetUriPath, err := getYangPathFromUri(app.path)


	return resp, err
}

func (app *CommonApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
	var err error
	var payload []byte

	return GetResponse{Payload: payload}, err
}

func (app *CommonApp) translateCRUCommon(d *db.DB, opcode int) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCRUCommon:path =", app.path)

	// translate yang to db
	result, err := transformer.XlateToDb((*app).ygotRoot, (*app).ygotTarget)
	fmt.Println(result)

	if err != nil {
		log.Error(err)
		return keys, err
	}

	keys, err = app.generateDbWatchKeys(d, false)

	return keys, err
}

func (app *CommonApp) generateDbWatchKeys(d *db.DB, isDeleteOp bool) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys

	return keys, err
}
