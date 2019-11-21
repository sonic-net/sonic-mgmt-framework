////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Dell, Inc.                                                 //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//  http://www.apache.org/licenses/LICENSE-2.0                                //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

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
	"cvl"
)

var ()

type CommonApp struct {
	pathInfo       *PathInfo
	body           []byte
	ygotRoot       *ygot.GoStruct
	ygotTarget     *interface{}
	cmnAppTableMap map[int]map[db.DBNum]map[string]map[string]db.Value
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
	*app = CommonApp{pathInfo: pathInfo, body: data.payload, ygotRoot: data.ygotRoot, ygotTarget: data.ygotTarget}

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

func (app *CommonApp) translateAction(dbs [db.MaxDB]*db.DB) error {
    var err error
    log.Info("translateAction:path =", app.pathInfo.Path, app.body)
    return err
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
    var resPayload []byte
    log.Info("processGet:path =", app.pathInfo.Path)
    var txCache interface{}

    for {
	    payload, err = transformer.GetAndXlateFromDB(app.pathInfo.Path, app.ygotRoot, dbs, txCache)
	    if err != nil {
		    log.Error("transformer.transformer.GetAndXlateFromDB failure. error:", err)
		    resPayload = payload
		    break
	    }

	    targetObj, tgtObjCastOk := (*app.ygotTarget).(ygot.GoStruct)
	    if tgtObjCastOk == false {
		    /*For ygotTarget populated by tranlib, for query on leaf level and list(without instance) level, 
		      casting to GoStruct fails so use the parent node of ygotTarget to Unmarshall the payload into*/
		    log.Infof("Use GetParentNode() since casting ygotTarget to GoStruct failed(uri - %v", app.pathInfo.Path)
		    targetUri := app.pathInfo.Path
		    parentTargetObj, _, getParentNodeErr := getParentNode(&targetUri, (*app.ygotRoot).(*ocbinds.Device))
		    if getParentNodeErr != nil {
			    log.Warningf("getParentNode() failure for uri %v", app.pathInfo.Path)
			    resPayload = payload
			    break
		    }
		    if parentTargetObj != nil {
			    targetObj, tgtObjCastOk = (*parentTargetObj).(ygot.GoStruct)
			    if tgtObjCastOk == false {
				    log.Warningf("Casting of parent object returned from getParentNode() to GoStruct failed(uri - %v)", app.pathInfo.Path)
				    resPayload = payload
				    break
			    }
		    } else {
			    log.Warningf("getParentNode() returned a nil Object for uri %v", app.pathInfo.Path)
                            resPayload = payload
                            break
		    }
	    }
	    if targetObj != nil {
		    err = ocbinds.Unmarshal(payload, targetObj)
		    if err != nil {
			    log.Error("ocbinds.Unmarshal()  failed. error:", err)
			    resPayload = payload
			    break
		    }

		    resPayload, err = generateGetResponsePayload(app.pathInfo.Path, (*app.ygotRoot).(*ocbinds.Device), app.ygotTarget)
		    if err != nil {
			    log.Error("generateGetResponsePayload()  failed")
			    resPayload = payload
		    }
		    break
	    } else {
		log.Warning("processGet. targetObj is null. Unable to Unmarshal payload")
		resPayload = payload
		break
	    }
    }

    return GetResponse{Payload: resPayload}, err
}

func (app *CommonApp) processAction(dbs [db.MaxDB]*db.DB) (ActionResponse, error) {
    var resp ActionResponse
	err := errors.New("Not implemented")

	resp.Payload, err = transformer.CallRpcMethod(app.pathInfo.Path, app.body, dbs)
	log.Info("transformer.XlateToDb() returned", resp.Payload)

	return resp, err
}

func (app *CommonApp) translateCRUDCommon(d *db.DB, opcode int) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	var tblsToWatch []*db.TableSpec
	var txCache interface{}
	log.Info("translateCRUDCommon:path =", app.pathInfo.Path)

	// translate yang to db
	result, err := transformer.XlateToDb(app.pathInfo.Path, opcode, d, (*app).ygotRoot, (*app).ygotTarget, txCache)
	fmt.Println(result)
	log.Info("transformer.XlateToDb() returned", result)

	if err != nil {
		log.Error(err)
		return keys, err
	}
	app.cmnAppTableMap = result
	if len(result) == 0 {
		log.Error("XlatetoDB() returned empty map")
		//Note: Get around for no redis ABNF Schema for set(temporary)
		//`err = errors.New("transformer.XlatetoDB() returned empty map")
		return keys, err
	}

	moduleNm, err := transformer.GetModuleNmFromPath(app.pathInfo.Path)
        if (err != nil) || (len(moduleNm) == 0) {
                log.Error("GetModuleNmFromPath() failed")
                return keys, err
        }

	var resultTblList []string
        for _, dbMap := range result { //Get dependency list for all tables in result
		for _, resMap := range dbMap { //Get dependency list for all tables in result
		        for tblnm, _ := range resMap { //Get dependency list for all tables in result
				resultTblList = append(resultTblList, tblnm)
			}
		}
	}
        log.Info("Result Tables List", resultTblList)

	// Get list of tables to watch
	depTbls := transformer.GetTablesToWatch(resultTblList, moduleNm)
	if len(depTbls) == 0 {
		log.Errorf("Failure to get Tables to watch for module %v", moduleNm)
		err = errors.New("GetTablesToWatch returned empty slice")
                return keys, err
	}
	for _, tbl := range depTbls {
		tblsToWatch = append(tblsToWatch, &db.TableSpec{Name: tbl})
	}
        log.Info("Tables to watch", tblsToWatch)
        cmnAppInfo.tablesToWatch = tblsToWatch

	keys, err = app.generateDbWatchKeys(d, false)
	return keys, err
}

func (app *CommonApp) processCommon(d *db.DB, opcode int) error {

	var err error
	if len(app.cmnAppTableMap) == 0 {
		return err
	}

	log.Info("Processing DB operation for ", app.cmnAppTableMap)
	switch opcode {
		case CREATE:
			log.Info("CREATE case")
		case UPDATE:
			log.Info("UPDATE case")
		case REPLACE:
			log.Info("REPLACE case")
		case DELETE:
			log.Info("DELETE case")
	}

	// Handle delete first if any available
	if _, ok := app.cmnAppTableMap[DELETE][db.ConfigDB]; ok {
		err = app.cmnAppDelDbOpn(d, DELETE, app.cmnAppTableMap[DELETE][db.ConfigDB])
		if err != nil {
			log.Info("Process delete fail. cmnAppDelDbOpn error:", err)
			return err
		}
	}
	// Handle create operation next
	if _, ok := app.cmnAppTableMap[CREATE][db.ConfigDB]; ok {
		err = app.cmnAppCRUCommonDbOpn(d, CREATE, app.cmnAppTableMap[CREATE][db.ConfigDB])
		if err != nil {
			log.Info("Process create fail. cmnAppCRUCommonDbOpn error:", err)
			return err
		}
	}
	// Handle update and replace operation next
	if _, ok := app.cmnAppTableMap[UPDATE][db.ConfigDB]; ok {
		err = app.cmnAppCRUCommonDbOpn(d, UPDATE, app.cmnAppTableMap[UPDATE][db.ConfigDB])
		if err != nil {
			log.Info("Process update fail. cmnAppCRUCommonDbOpn error:", err)
			return err
		}
	}
	if _, ok := app.cmnAppTableMap[REPLACE][db.ConfigDB]; ok {
		err = app.cmnAppCRUCommonDbOpn(d, REPLACE, app.cmnAppTableMap[REPLACE][db.ConfigDB])
		if err != nil {
			log.Info("Process replace fail. cmnAppCRUCommonDbOpn error:", err)
			return err
		}
	}
	log.Info("Returning from processCommon() - success")
	return err
}

//sort transformer result table list based on dependenciesi(using CVL API) tables to be used for CRUD operations
func sortAsPerTblDeps(tblLst []string) ([]string, error) {
	var resultTblLst []string
	var err error
	logStr := "Failure in CVL API to sort table list as per dependencies."

	cvSess, cvlRetSess := cvl.ValidationSessOpen()
	if cvlRetSess != cvl.CVL_SUCCESS {

		log.Errorf("Failure in creating CVL validation session object required to use CVl API(sort table list as per dependencies) - %v", cvlRetSess)
		err = fmt.Errorf("%v", logStr)
		return resultTblLst, err
	}
	cvlSortDepTblList, cvlRetDepTbl := cvSess.SortDepTables(tblLst)
	if cvlRetDepTbl != cvl.CVL_SUCCESS {
		log.Warningf("Failure in cvlSess.SortDepTables: %v", cvlRetDepTbl)
		cvl.ValidationSessClose(cvSess)
		err = fmt.Errorf("%v", logStr)
		return resultTblLst, err
	}
	log.Info("cvlSortDepTblList = ", cvlSortDepTblList)
	resultTblLst = cvlSortDepTblList

	cvl.ValidationSessClose(cvSess)
	return resultTblLst, err

}

func (app *CommonApp) cmnAppCRUCommonDbOpn(d *db.DB, opcode int, dbMap map[string]map[string]db.Value) error {
	var err error
	var cmnAppTs *db.TableSpec
	var xfmrTblLst []string
	var resultTblLst []string

	for tblNm, _ := range(dbMap) {
		xfmrTblLst = append(xfmrTblLst, tblNm)
	}
	resultTblLst, err = sortAsPerTblDeps(xfmrTblLst)
	if err != nil {
		return err
	}

	/* CVL sorted order is in child first, parent later order. CRU ops from parent first order */
	for idx := len(resultTblLst)-1; idx >= 0; idx-- {
		tblNm := resultTblLst[idx]
		log.Info("In Yang to DB map returned from transformer looking for table = ", tblNm)
		if tblVal, ok := dbMap[tblNm]; ok {
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
						/* Handle leaf-list merge if any leaf-list exists 
						A leaf-list field in redis has "@" suffix as per swsssdk convention.
						*/
						resTblRw := db.Value{Field: map[string]string{}}
						resTblRw = checkAndProcessLeafList(existingEntry, tblRw, UPDATE, d, tblNm, tblKey)
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

func (app *CommonApp) cmnAppDelDbOpn(d *db.DB, opcode int, dbMap map[string]map[string]db.Value) error {
	var err error
	var cmnAppTs, dbTblSpec *db.TableSpec
	var moduleNm string
	var xfmrTblLst []string
	var resultTblLst []string

	for tblNm, _ := range(dbMap) {
		xfmrTblLst = append(xfmrTblLst, tblNm)
	}
	resultTblLst, err = sortAsPerTblDeps(xfmrTblLst)
	if err != nil {
		return err
	}


	/* Retrieve module Name */
	moduleNm, err = transformer.GetModuleNmFromPath(app.pathInfo.Path)
	if (err != nil) || (len(moduleNm) == 0) {
		log.Error("GetModuleNmFromPath() failed")
		return err
	}
	log.Info("getModuleNmFromPath() returned module name = ", moduleNm)

	/* resultTblLst has child first, parent later order */
	for _, tblNm := range resultTblLst {
		log.Info("In Yang to DB map returned from transformer looking for table = ", tblNm)
		if tblVal, ok := dbMap[tblNm]; ok {
			cmnAppTs = &db.TableSpec{Name: tblNm}
			log.Info("Found table entry in yang to DB map")
			ordTblList := transformer.GetOrdTblList(tblNm, moduleNm)
			if len(ordTblList) == 0 {
				log.Error("GetOrdTblList returned empty slice")
				err = errors.New("GetOrdTblList returned empty slice. Insufficient information to process request")
				return err
			}
			log.Infof("GetOrdTblList for table - %v, module %v returns %v", tblNm, moduleNm, ordTblList)
			if len(tblVal) == 0 {
				log.Info("DELETE case - No table instances/rows found hence delete entire table = ", tblNm)
				for _, ordtbl := range ordTblList {
					log.Info("Since parent table is to be deleted, first deleting child table = ", ordtbl)
					if ordtbl == tblNm {
						// Handle the child tables only till you reach the parent table entry
						break
					}
					dbTblSpec = &db.TableSpec{Name: ordtbl}
					err = d.DeleteTable(dbTblSpec)
					if err != nil {
						log.Warning("DELETE case - d.DeleteTable() failure for Table = ", ordtbl)
						return err
					}
				}
				err = d.DeleteTable(cmnAppTs)
				if err != nil {
					log.Warning("DELETE case - d.DeleteTable() failure for Table = ", tblNm)
					return err
				}
				log.Info("DELETE case - Deleted entire table = ", tblNm)
				// Continue to repeat ordered deletion for all tables
				continue

			}

			for tblKey, tblRw := range tblVal {
				if len(tblRw.Field) == 0 {
					log.Info("DELETE case - no fields/cols to delete hence delete the entire row.")
					log.Info("First, delete child table instances that correspond to parent table instance to be deleted = ", tblKey)
					for _, ordtbl := range ordTblList {
						if ordtbl == tblNm {
							// Handle the child tables only till you reach the parent table entry
							break;
						}
						dbTblSpec = &db.TableSpec{Name: ordtbl}
						keyPattern := tblKey + "|*"
						log.Info("Key pattern to be matched for deletion = ", keyPattern)
						err = d.DeleteKeys(dbTblSpec, db.Key{Comp: []string{keyPattern}})
						if err != nil {
							log.Warning("DELETE case - d.DeleteTable() failure for Table = ", ordtbl)
							return err
						}
						log.Info("Deleted keys matching parent table key pattern for child table = ", ordtbl)

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
					/* handle leaf-list merge if any leaf-list exists */
					resTblRw := checkAndProcessLeafList(existingEntry, tblRw, DELETE, d, tblNm, tblKey)
					if len(resTblRw.Field) > 0 {
						err := d.DeleteEntryFields(cmnAppTs, db.Key{Comp: []string{tblKey}}, resTblRw)
						if err != nil {
							log.Error("DELETE case - d.DeleteEntryFields() failure")
							return err
						}
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

/*check if any field is leaf-list , if yes perform merge*/
func checkAndProcessLeafList(existingEntry db.Value, tblRw db.Value, opcode int, d *db.DB, tblNm string, tblKey string) db.Value {
	dbTblSpec := &db.TableSpec{Name: tblNm}
	mergeTblRw := db.Value{Field: map[string]string{}}
	for field, value := range tblRw.Field {
		if strings.HasSuffix(field, "@") {
			exstLst := existingEntry.GetList(field)
			var valueLst []string
			if value != "" { //zero len string as leaf-list value is treated as delete all items in leaf-list
				valueLst = strings.Split(value, ",")
			}
			if len(exstLst) != 0 {
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
					if len(valueLst) > 0 {
						mergeTblRw.SetList(field, exstLst)
						delete(tblRw.Field, field)
					}
				}
			} else if opcode == UPDATE {
                                exstLst = valueLst
                        }
			if opcode == UPDATE {
				tblRw.SetList(field, exstLst)
			}
		}
	}
	/* delete specific item from leaf-list */
	if opcode == DELETE {
		if len(mergeTblRw.Field) == 0 {
			log.Infof("mergeTblRow is empty - Returning Table Row %v", tblRw)
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
