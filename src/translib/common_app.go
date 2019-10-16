package translib

import (
	"errors"
	"fmt"
	"strings"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"translib/db"
    "translib/ocbinds"
	"translib/tlerr"
	"translib/transformer"
	"encoding/json"
)

var ()

type CommonApp struct {
	pathInfo       *PathInfo
	ygotRoot       *ygot.GoStruct
	ygotTarget     *interface{}
	cmnAppTableMap map[string]map[string]db.Value
	cmnAppOrdTbllist []string
}

var cmnAppInfo = appInfo{appType: reflect.TypeOf(CommonApp{}),
	ygotRootType:  nil,
	isNative:      false,
	tablesToWatch: nil}

func init() {

	register_model_path := []string{"/sonic-", "*"} // register yang model path(s) to be supported via common app
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

	keys, err = app.translateCRUDCommon(d, CREATE)

	return keys, err
}

func (app *CommonApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateUpdate:path =", app.pathInfo.Path)

	keys, err = app.translateCRUDCommon(d, UPDATE)

	return keys, err
}

func (app *CommonApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateReplace:path =", app.pathInfo.Path)

	keys, err = app.translateCRUDCommon(d, REPLACE)

	return keys, err
}

func (app *CommonApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateDelete:path =", app.pathInfo.Path)
	keys, err = app.translateCRUDCommon(d, DELETE)

	return keys, err
}

func (app *CommonApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	var err error
	log.Info("translateGet:path =", app.pathInfo.Path)
	return err
}

func (app *CommonApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
	err := errors.New("Not supported")
	notifInfo := notificationInfo{dbno: db.ConfigDB}
	return nil, &notifInfo, err
}

func (app *CommonApp) processCreate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processCreate:path =", app.pathInfo.Path)
	targetType := reflect.TypeOf(*app.ygotTarget)
	log.Infof("processCreate: Target object is a <%s> of Type: %s", targetType.Kind().String(), targetType.Elem().Name())
	if err = app.processCommon(d, CREATE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}

	return resp, err
}

func (app *CommonApp) processUpdate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processUpdate:path =", app.pathInfo.Path)
	if err = app.processCommon(d, UPDATE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}

	return resp, err
}

func (app *CommonApp) processReplace(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processReplace:path =", app.pathInfo.Path)
	if err = app.processCommon(d, REPLACE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}
	return resp, err
}

func (app *CommonApp) processDelete(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processDelete:path =", app.pathInfo.Path)

	if err = app.processCommon(d, DELETE); err != nil {
		log.Error(err)
		resp = SetResponse{ErrSrc: AppErr}
	}

	return resp, err
}

func (app *CommonApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
    var err error
    var payload []byte
    log.Info("processGet:path =", app.pathInfo.Path)

    payload, err = transformer.GetAndXlateFromDB(app.pathInfo.Path, app.ygotRoot, dbs)
    if err != nil {
	    log.Error("transformer.transformer.GetAndXlateFromDB failure. error:", err)
        return GetResponse{Payload: payload, ErrSrc: AppErr}, err
    }

    targetObj, _ := (*app.ygotTarget).(ygot.GoStruct)
    if targetObj != nil {
	    err = ocbinds.Unmarshal(payload, targetObj)
	    if err != nil {
		    log.Error("ocbinds.Unmarshal()  failed. error:", err)
		    return GetResponse{Payload: payload, ErrSrc: AppErr}, err
	    }
    }

   payload, err = generateGetResponsePayload(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device), app.ygotTarget)
    if err != nil {
        log.Error("generateGetResponsePayload()  failed")
        return GetResponse{Payload: payload, ErrSrc: AppErr}, err
    }
    var dat map[string]interface{}
    err = json.Unmarshal(payload, &dat)

	return GetResponse{Payload: payload}, err
}

func (app *CommonApp) translateCRUDCommon(d *db.DB, opcode int) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	var tblsToWatch []*db.TableSpec
	var OrdTblList []string
	var moduleNm string
	log.Info("translateCRUDCommon:path =", app.pathInfo.Path)

	/* retrieve schema table order for incoming module name request */
	moduleNm, err = transformer.GetModuleNmFromPath(app.pathInfo.Path)
	if (err != nil) || (len(moduleNm) == 0) {
		log.Error("GetModuleNmFromPath() failed")
		return keys, err
	}
	log.Info("getModuleNmFromPath() returned module name = ", moduleNm)
	OrdTblList, err = transformer.GetOrdDBTblList(moduleNm)
	if (err != nil) || (len(OrdTblList) == 0) {
		log.Error("GetOrdDBTblList() failed")
		return keys, err
	}

	log.Info("GetOrdDBTblList() returned ordered table list = ", OrdTblList)
	app.cmnAppOrdTbllist = OrdTblList

	/* enhance this to handle dependent tables - need CVL to provide list of such tables for a given request */
	for _, tblnm := range OrdTblList { // OrdTblList already has has all tables corresponding to a module
		tblsToWatch = append(tblsToWatch, &db.TableSpec{Name: tblnm})
	}
	log.Info("Tables to watch", tblsToWatch)

	cmnAppInfo.tablesToWatch = tblsToWatch

	// translate yang to db
	result, err := transformer.XlateToDb(app.pathInfo.Path, opcode, d, (*app).ygotRoot, (*app).ygotTarget)
	fmt.Println(result)
	log.Info("transformer.XlateToDb() returned", result)

	if err != nil {
		log.Error(err)
		return keys, err
	}
	if len(result) == 0 {
		log.Error("XlatetoDB() returned empty map")
		err = errors.New("transformer.XlatetoDB() returned empty map")
		return keys, err
	}
	app.cmnAppTableMap = result

	keys, err = app.generateDbWatchKeys(d, false)

	return keys, err
}

func (app *CommonApp) processCommon(d *db.DB, opcode int) error {

	var err error

	log.Info("Processing DB operation for ", app.cmnAppTableMap)
	switch opcode {
		case CREATE:
			log.Info("CREATE case")
			err = app.cmnAppCRUCommonDbOpn(d, opcode)
		case UPDATE:
			log.Info("UPDATE case")
			err = app.cmnAppCRUCommonDbOpn(d, opcode)
		case REPLACE:
			log.Info("REPLACE case")
			err = app.cmnAppCRUCommonDbOpn(d, opcode)
		case DELETE:
			log.Info("DELETE case")
			err = app.cmnAppDelDbOpn(d, opcode)
	}
	if err != nil {
		log.Info("Returning from processCommon() - fail")
	} else {
		log.Info("Returning from processCommon() - success")
	}
	return err
}

func (app *CommonApp) cmnAppCRUCommonDbOpn(d *db.DB, opcode int) error {
	var err error
	var cmnAppTs *db.TableSpec

	/* currently ordered by schema table order needs to be discussed */
	for _, tblNm := range app.cmnAppOrdTbllist {
		log.Info("In Yang to DB map returned from transformer looking for table = ", tblNm)
		if tblVal, ok := app.cmnAppTableMap[tblNm]; ok {
			cmnAppTs = &db.TableSpec{Name: tblNm}
			log.Info("Found table entry in yang to DB map")
			for tblKey, tblRw := range tblVal {
				log.Info("Processing Table key and row ", tblKey, tblRw)
				existingEntry, _ := d.GetEntry(cmnAppTs, db.Key{Comp: []string{tblKey}})
				switch opcode {
				case CREATE:
					if existingEntry.IsPopulated() {
						log.Info("Entry already exists hence return.")
						return tlerr.AlreadyExists("Entry %s already exists", tblKey)
					} else {
						err = d.CreateEntry(cmnAppTs, db.Key{Comp: []string{tblKey}}, tblRw)
						if err != nil {
							log.Error("CREATE case - d.CreateEntry() failure")
							return err
						}
					}
				case UPDATE:
					if existingEntry.IsPopulated() {
						log.Info("Entry already exists hence modifying it.")
						/* Handle leaf-list merge 
						   A leaf-list field in redis has "@" suffix as per swsssdk convention.
						 */
						resTblRw := db.Value{Field: map[string]string{}}
						resTblRw = processLeafList(existingEntry, tblRw, UPDATE, d, tblNm, tblKey)
						err = d.ModEntry(cmnAppTs, db.Key{Comp: []string{tblKey}}, resTblRw)
						if err != nil {
							log.Error("UPDATE case - d.ModEntry() failure")
							return err
						}
					} else {
                                                // workaround to patch operation from CLI
                                                log.Info("Create(pathc) an entry.")
                                                err = d.CreateEntry(cmnAppTs, db.Key{Comp: []string{tblKey}}, tblRw)
						if err != nil {
							log.Error("UPDATE case - d.CreateEntry() failure")
							return err
						}
					}
				case REPLACE:
					if existingEntry.IsPopulated() {
						log.Info("Entry already exists hence execute db.SetEntry")
						err := d.SetEntry(cmnAppTs, db.Key{Comp: []string{tblKey}}, tblRw)
						if err != nil {
							log.Error("REPLACE case - d.SetEntry() failure")
							return err
						}
					} else {
						log.Info("Entry doesn't exist hence create it.")
						err = d.CreateEntry(cmnAppTs, db.Key{Comp: []string{tblKey}}, tblRw)
						if err != nil {
							log.Error("REPLACE case - d.CreateEntry() failure")
							return err
						}
					}
				}
			}
		}
	}
	return err
}

func (app *CommonApp) cmnAppDelDbOpn(d *db.DB, opcode int) error {
	var err error
	var cmnAppTs, dbTblSpec *db.TableSpec

	/* needs enhancements from CVL to give table dependencies, and grouping of related tables only 
	   if such a case where the sonic yang has unrelated tables */
	for tblidx, tblNm := range app.cmnAppOrdTbllist {
		log.Info("In Yang to DB map returned from transformer looking for table = ", tblNm)
		if tblVal, ok := app.cmnAppTableMap[tblNm]; ok {
			cmnAppTs = &db.TableSpec{Name: tblNm}
			log.Info("Found table entry in yang to DB map")
			if len(tblVal) == 0 {
				log.Info("DELETE case - No table instances/rows found hence delete entire table = ", tblNm)
				for idx := len(app.cmnAppOrdTbllist)-1; idx >= tblidx+1; idx-- {
					log.Info("Since parent table is to be  deleted, first deleting child table = ", app.cmnAppOrdTbllist[idx])
					dbTblSpec = &db.TableSpec{Name: app.cmnAppOrdTbllist[idx]}
					err = d.DeleteTable(dbTblSpec)
					if err != nil {
						log.Warning("DELETE case - d.DeleteTable() failure for Table = ", app.cmnAppOrdTbllist[idx])
						return err
					}
				}
				err = d.DeleteTable(cmnAppTs)
				if err != nil {
					log.Warning("DELETE case - d.DeleteTable() failure for Table = ", tblNm)
					return err
				}
				log.Info("DELETE case - Deleted entire table = ", tblNm)
				log.Info("Done processing all tables.")
				break

			}

			for tblKey, tblRw := range tblVal {
				if len(tblRw.Field) == 0 {
					log.Info("DELETE case - no fields/cols to delete hence delete the entire row.")
					log.Info("First, delete child table instances that correspond to parent table instance to be deleted = ", tblKey)
					for idx := len(app.cmnAppOrdTbllist)-1; idx >= tblidx+1; idx-- {
						dbTblSpec = &db.TableSpec{Name: app.cmnAppOrdTbllist[idx]}
						keyPattern := tblKey + "|*"
						log.Info("Key pattern to be matched for deletion = ", keyPattern)
						err = d.DeleteKeys(dbTblSpec, db.Key{Comp: []string{keyPattern}})
						if err != nil {
							log.Warning("DELETE case - d.DeleteTable() failure for Table = ", app.cmnAppOrdTbllist[idx])
							return err
						}
						log.Info("Deleted keys matching parent table key pattern for child table = ", app.cmnAppOrdTbllist[idx])

					}
					err = d.DeleteEntry(cmnAppTs, db.Key{Comp: []string{tblKey}})
                                        if err != nil {
                                                log.Warning("DELETE case - d.DeleteEntry() failure")
                                                return err
                                        }
					log.Info("Finally deleted the parent table row with key = ", tblKey)
				} else {
					log.Info("DELETE case - fields/cols to delete hence delete only those fields.")
					existingEntry, _ := d.GetEntry(cmnAppTs, db.Key{Comp: []string{tblKey}})
					if !existingEntry.IsPopulated() {
						log.Info("Table Entry from which the fields are to be deleted does not exist")
						return err
					}
					/*handle leaf-list merge*/
					resTblRw := processLeafList(existingEntry, tblRw, DELETE, d, tblNm, tblKey)
					err := d.DeleteEntryFields(cmnAppTs, db.Key{Comp: []string{tblKey}}, resTblRw)
					if err != nil {
						log.Error("DELETE case - d.DeleteEntryFields() failure")
						return err
					}
				}

			}
		}
	} /* end of ordered table list for loop */
	return err
}

func (app *CommonApp) generateDbWatchKeys(d *db.DB, isDeleteOp bool) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys

	return keys, err
}

func processLeafList(existingEntry db.Value, tblRw db.Value, opcode int, d *db.DB, tblNm string, tblKey string) db.Value {
	log.Info("process leaf-list Fields in table row.")
	dbTblSpec := &db.TableSpec{Name: tblNm}
	mergeTblRw := db.Value{Field: map[string]string{}}
	for field, value := range tblRw.Field {
		if strings.HasSuffix(field, "@") {
			exstLst := existingEntry.GetList(field)
			if len(exstLst) != 0 {
				valueLst := strings.Split(value, ",")
				for _, item := range valueLst {
					if !contains(exstLst, item) {
						if opcode == UPDATE {
							exstLst = append(exstLst, item)
						}
					} else {
						if opcode == DELETE {
                                                        exstLst = removeElement(exstLst, item)
                                                }

					}
				}
				log.Infof("For field %v value after merge %v", field, exstLst)
				if opcode == DELETE {
					mergeTblRw.SetList(field, exstLst)
					delete(tblRw.Field, field)
				}
			}
			tblRw.SetList(field, exstLst)
		}
	}
	/* delete specific item from leaf-list */
	if opcode == DELETE {
		if mergeTblRw.Field == nil {
			return tblRw
		}
		err := d.ModEntry(dbTblSpec, db.Key{Comp: []string{tblKey}}, mergeTblRw)
		if err != nil {
			log.Warning("DELETE case(merge leaf-list) - d.ModEntry() failure")
		}
	}
	log.Infof("Returning Table Row %v", tblRw)
	return tblRw
}

