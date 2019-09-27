package transformer

import (
    "fmt"
    "translib/db"
    "strings"
    "encoding/json"
    "os"
    "strconv"
    "errors"
    "translib/ocbinds"
    "github.com/openconfig/goyang/pkg/yang"
    "github.com/openconfig/ygot/ygot"
    "github.com/openconfig/ygot/ytypes"

    log "github.com/golang/glog"
)

type typeMapOfInterface map[string]interface{}

func xfmrHandlerFunc(inParams XfmrParams) (map[string]interface{}, error) {
    result := make(map[string]interface{})
    xpath, _ := RemoveXPATHPredicates(inParams.uri)
    log.Infof("Subtree transformer function(\"%v\") invoked for yang path(\"%v\").", xYangSpecMap[xpath].xfmrFunc, xpath)
    _, err := XlateFuncCall(dbToYangXfmrFunc(xYangSpecMap[xpath].xfmrFunc), inParams)
    if err != nil {
        log.Infof("Failed to retrieve data for xpath(\"%v\") err(%v).", inParams.uri, err)
        return result, err
    }

    ocbSch, _  := ocbinds.Schema()
    schRoot    := ocbSch.RootSchema()
    device     := (*inParams.ygRoot).(*ocbinds.Device)

    path, _ := ygot.StringToPath(inParams.uri, ygot.StructuredPath, ygot.StringSlicePath)
    for _, p := range path.Elem {
        pathSlice := strings.Split(p.Name, ":")
        p.Name = pathSlice[len(pathSlice)-1]
        if len(p.Key) > 0 {
            for ekey, ent := range p.Key {
                eslice := strings.Split(ent, ":")
                p.Key[ekey] = eslice[len(eslice)-1]
            }
        }
    }

    nodeList, nErr := ytypes.GetNode(schRoot, device, path)
    if nErr != nil {
        log.Infof("Failed to get node for xpath(\"%v\") err(%v).", inParams.uri, err)
        return result, err
    }
    node := nodeList[0].Data
    nodeYgot, _:= (node).(ygot.ValidatedGoStruct)
    payload, err := ygot.EmitJSON(nodeYgot, &ygot.EmitJSONConfig{ Format: ygot.RFC7951,
                                  Indent: "  ", SkipValidation: true,
                                  RFC7951Config: &ygot.RFC7951JSONConfig{ AppendModuleName: false, },
                                  })
    err = json.Unmarshal([]byte(payload), result)
    return result, err
}

func leafXfmrHandlerFunc(inParams XfmrParams) (map[string]interface{}, error) {
    xpath, _ := RemoveXPATHPredicates(inParams.uri)
    ret, err := XlateFuncCall(dbToYangXfmrFunc(xYangSpecMap[xpath].xfmrFunc), inParams)
    if err != nil {
        return nil, err
    }
    fldValMap := ret[0].Interface().(map[string]interface{})
    return fldValMap, nil
}

func validateHandlerFunc(inParams XfmrParams) (bool) {
    xpath, _ := RemoveXPATHPredicates(inParams.uri)
    ret, err := XlateFuncCall(xYangSpecMap[xpath].validateFunc, inParams)
    if err != nil {
        return false
    }
    return ret[0].Interface().(bool)
}

func xfmrTblHandlerFunc(xfmrTblFunc string, inParams XfmrParams) []string {
    ret, err := XlateFuncCall(xfmrTblFunc, inParams)
    if err != nil {
        return []string{}
    }
    return ret[0].Interface().([]string)
}


func DbValToInt(dbFldVal string, base int, size int, isUint bool) (interface{}, error) {
	var res interface{}
	var err error
	if isUint {
		if res, err = strconv.ParseUint(dbFldVal, base, size); err != nil {
			log.Warningf("Non Yint%v type for yang leaf-list item %v", size, dbFldVal)
		}
	} else {
		if res, err = strconv.ParseInt(dbFldVal, base, size); err != nil {
			log.Warningf("Non Yint %v type for yang leaf-list item %v", size, dbFldVal)
		}
	}
	return res, err
}

func DbToYangType(yngTerminalNdDtType yang.TypeKind, fldXpath string, dbFldVal string) (interface{}, error) {
	log.Infof("Received FieldXpath %v, yngTerminalNdDtType %v and Db field value %v to be converted to yang data-type.", fldXpath, yngTerminalNdDtType, dbFldVal)
	var res interface{}
	var err error
	const INTBASE = 10
	switch yngTerminalNdDtType {
        case yang.Ynone:
                log.Warning("Yang node data-type is non base yang type")
		//TODO - enhance to handle non base data types depending on future use case
		err = errors.New("Yang node data-type is non base yang type")
        case yang.Yint8:
                res, err = DbValToInt(dbFldVal, INTBASE, 8, false)
        case yang.Yint16:
                res, err = DbValToInt(dbFldVal, INTBASE, 16, false)
        case yang.Yint32:
                res, err = DbValToInt(dbFldVal, INTBASE, 32, false)
        case yang.Yuint8:
                res, err = DbValToInt(dbFldVal, INTBASE, 8, true)
        case yang.Yuint16:
                res, err = DbValToInt(dbFldVal, INTBASE, 16, true)
        case yang.Yuint32:
                res, err = DbValToInt(dbFldVal, INTBASE, 32, true)
        case yang.Ybool:
		if res, err = strconv.ParseBool(dbFldVal); err != nil {
			log.Warningf("Non Bool type for yang leaf-list item %v", dbFldVal)
		}
        case yang.Ybinary, yang.Ydecimal64, yang.Yenum, yang.Yidentityref, yang.Yint64, yang.Yuint64,     yang.Ystring, yang.Yunion:
                // TODO - handle the union type
                // Make sure to encode as string, expected by util_types.go: ytypes.yangToJSONType
                log.Info("Yenum/Ystring/Yunion(having all members as strings) type for yangXpath ", fldXpath)
                res = dbFldVal
	case yang.Yempty:
		logStr := fmt.Sprintf("Yang data type for xpath %v is Yempty.", fldXpath)
		log.Error(logStr)
		err = errors.New(logStr)
        default:
		logStr := fmt.Sprintf("Unrecognized/Unhandled yang-data type(%v) for xpath %v.", fldXpath, yang.TypeKindToName[yngTerminalNdDtType])
                log.Error(logStr)
                err = errors.New(logStr)
        }
	return res, err
}

/*convert leaf-list in Db to leaf-list in yang*/
func processLfLstDbToYang(fieldXpath string, dbFldVal string) []interface{} {
	valLst := strings.Split(dbFldVal, ",")
	var resLst []interface{}
	const INTBASE = 10
	yngTerminalNdDtType := xDbSpecMap[fieldXpath].dbEntry.Type.Kind
	switch  yngTerminalNdDtType {
	case yang.Yenum, yang.Ystring, yang.Yunion:
		log.Info("DB leaf-list and Yang leaf-list are of same data-type")
		for _, fldVal := range valLst {
			resLst = append(resLst, fldVal)
		}
	default:
		for _, fldVal := range valLst {
			resVal, err := DbToYangType(yngTerminalNdDtType, fieldXpath, fldVal)
			if err == nil {
				resLst = append(resLst, resVal)
			}
		}
	}
	return resLst
}

/* Traverse db map and create json for cvl yang */
func directDbToYangJsonCreate(dbDataMap *map[db.DBNum]map[string]map[string]db.Value, jsonData string, resultMap map[string]interface{}) error {
    var err error
	for curDbIdx := db.ApplDB; curDbIdx < db.MaxDB; curDbIdx++ {
		dbTblData := (*dbDataMap)[curDbIdx]
		for tblName, tblData := range dbTblData {
			var mapSlice []typeMapOfInterface
			for keyStr, dbFldValData := range tblData {
				curMap := make(map[string]interface{})
				for field, value := range dbFldValData.Field {
					resField := field
					if strings.HasSuffix(field, "@") {
						fldVals := strings.Split(field, "@")
						resField = fldVals[0]
					}
					fieldXpath := tblName + "/" + resField
					xDbSpecMapEntry, ok := xDbSpecMap[fieldXpath]
					if !ok {
						log.Warningf("No entry found in xDbSpecMap for xpath %v", fieldXpath)
						continue
					}
					if xDbSpecMapEntry.dbEntry == nil {
						log.Warningf("Yang entry is nil in xDbSpecMap for xpath %v", fieldXpath)
						continue
					}
					yangType := yangTypeGet(xDbSpecMapEntry.dbEntry)
					if yangType == "leaf-list" {
						/* this should never happen but just adding for safetty */
						if !strings.HasSuffix(field, "@") {
							log.Warningf("Leaf-list in Sonic yang should also be a leaf-list in DB, its not for xpath %v", fieldXpath)
							continue
						}
						resLst := processLfLstDbToYang(fieldXpath, value)
						curMap[resField] = resLst
					} else { /* yangType is leaf - there are only 2 types of yang terminal node leaf and leaf-list */
					yngTerminalNdDtType := xDbSpecMapEntry.dbEntry.Type.Kind
					resVal, err := DbToYangType(yngTerminalNdDtType, fieldXpath, value)
					if err != nil {
						log.Warningf("Failure in converting Db value type to yang type for xpath", fieldXpath)
					} else {
						curMap[resField] = resVal
					}
				}
			} //end of for
			dbSpecData, ok := xDbSpecMap[tblName]
			dbIndex := db.ConfigDB
			if ok {
				dbIndex = dbSpecData.dbIndex
			}
			yangKeys := yangKeyFromEntryGet(xDbSpecMap[tblName].dbEntry)
			sonicKeyDataAdd(dbIndex, yangKeys, keyStr, curMap)
			if curMap != nil {
				mapSlice = append(mapSlice, curMap)
			}
		}
		resultMap[tblName] = mapSlice
	}
}
	return err
}

func tableNameAndKeyFromDbMapGet(dbDataMap map[string]map[string]db.Value) (string, string, error) {
    tableName := ""
    tableKey  := ""
    for tn, tblData := range dbDataMap {
        tableName = tn
        for kname, _ := range tblData {
            tableKey = kname
        }
    }
    return tableName, tableKey, nil
}

func fillDbDataMapForTbl(uri string, xpath string, tblName string, tblKey string, cdb db.DBNum, dbs [db.MaxDB]*db.DB) (map[db.DBNum]map[string]map[string]db.Value, error) {
	var err error
	dbresult  := make(map[db.DBNum]map[string]map[string]db.Value)
	dbresult[cdb] = make(map[string]map[string]db.Value)
	dbFormat := KeySpec{}
	dbFormat.Ts.Name = tblName
	dbFormat.dbNum = cdb
	/* TODO - handle key 
	if tblKey != "" {
		dbFormat.Key.Comp = append(dbFormat.Key.Comp, tblKey)
	} */
	err = TraverseDb(dbs, dbFormat, &dbresult, nil)
	if err != nil {
		log.Errorf("TraverseDb() failure for tbl(DB num) %v(%v) for xpath %v", tblName, cdb, xpath)
		return nil, err
	}
	if _, ok := dbresult[cdb]; !ok {
		logStr := fmt.Sprintf("TraverseDb() did not populate Db data for tbl(DB num) %v(%v) for xpath %v", tblName, cdb, xpath)
		err = fmt.Errorf("%v", logStr)
		return nil, err
	}
	return dbresult, err

}

// Assumption: All tables are from the same DB
func dbDataFromTblXfmrGet(tbl string, inParams XfmrParams, dbDataMap *map[db.DBNum]map[string]map[string]db.Value) error {
    xpath, _ := RemoveXPATHPredicates(inParams.uri)
    curDbDataMap, err := fillDbDataMapForTbl(inParams.uri, xpath, tbl, "", inParams.curDb, inParams.dbs)
    if err == nil {
        mapCopy((*dbDataMap)[inParams.curDb], curDbDataMap[inParams.curDb])
    }
    return nil
}

func yangListDataFill(dbs [db.MaxDB]*db.DB, ygRoot *ygot.GoStruct, uri string, xpath string, dbDataMap *map[db.DBNum]map[string]map[string]db.Value, resultMap map[string]interface{}, tbl string, tblKey string, cdb db.DBNum, validate bool) error {
	var tblList []string
	tblXfmr   := false

	if tbl == "" && xYangSpecMap[xpath].xfmrTbl != nil {
		xfmrTblFunc := *xYangSpecMap[xpath].xfmrTbl
		if len(xfmrTblFunc) > 0 {
			tblXfmr   = true
			inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, uri, GET, "", dbDataMap, nil)
			tblList   = xfmrTblHandlerFunc(xfmrTblFunc, inParams)
			if len(tblList) != 0 {
				for _, curTbl := range tblList {
					dbDataFromTblXfmrGet(curTbl, inParams, dbDataMap)
				}
			}
		}
	} else if tbl != "" && !tblXfmr {
		tblList = append(tblList, tbl)
	}

	for _, tbl = range(tblList) {
		tblData, ok := (*dbDataMap)[cdb][tbl]

		if ok {
			var mapSlice []typeMapOfInterface
			for dbKey, _ := range tblData {
				curMap := make(map[string]interface{})
				curKeyMap, curUri, _ := dbKeyToYangDataConvert(uri, xpath, dbKey, dbs[cdb].Opts.KeySeparator)
				if len(xYangSpecMap[xpath].xfmrFunc) > 0 {
					inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, curUri, GET, "", dbDataMap, nil)
					cmap, _  := xfmrHandlerFunc(inParams)
					if len(cmap) > 0 {
						mapSlice = append(mapSlice, curMap)
					} else {
						log.Infof("Empty container returned from overloaded transformer for(\"%v\")", curUri)
					}
				} else {
					_, keyFromCurUri, _ := xpathKeyExtract(dbs[cdb], ygRoot, GET, curUri)
					if dbKey == keyFromCurUri {
						for k, kv := range curKeyMap {
							curMap[k] = kv
						}
						curXpath, _ := RemoveXPATHPredicates(curUri)
						yangDataFill(dbs, ygRoot, curUri, curXpath, dbDataMap, curMap, tbl, dbKey, cdb, validate)
						mapSlice = append(mapSlice, curMap)
					}
				}
			}
			if len(mapSlice) > 0 {
				resultMap[xYangSpecMap[xpath].yangEntry.Name] = mapSlice
			} else {
				log.Infof("Empty slice for (\"%v\").\r\n", uri)
			}
		}
	}// end of tblList for
	return nil
}

func terminalNodeProcess(dbs [db.MaxDB]*db.DB, ygRoot *ygot.GoStruct, uri string, xpath string, dbDataMap *map[db.DBNum]map[string]map[string]db.Value, tbl string, tblKey string) (map[string]interface{}, error) {
	log.Infof("Received xpath - %v, uri - %v, dbDataMap - %v, table - %v, table key - %v", xpath, uri, (*dbDataMap), tbl, tblKey)
	var err error
	resFldValMap := make(map[string]interface{})
	if xYangSpecMap[xpath].yangEntry == nil {
		logStr := fmt.Sprintf("No yang entry found for xpath %v.", xpath)
		err = fmt.Errorf("%v", logStr)
		return resFldValMap, err
	}

	cdb := xYangSpecMap[xpath].dbIndex
	if len(xYangSpecMap[xpath].xfmrFunc) > 0 {
		_, key, _ := xpathKeyExtract(dbs[cdb], ygRoot, GET, uri)
		inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, uri, GET, key, dbDataMap, nil)
		fldValMap, err := leafXfmrHandlerFunc(inParams)
		if err != nil {
			logStr := fmt.Sprintf("%Failed to get data from overloaded function for %v -v.", uri, err)
			err = fmt.Errorf("%v", logStr)
			return resFldValMap, err
		}
		for lf, val := range fldValMap {
			resFldValMap[lf] = val
		}
	} else {
		dbFldName := xYangSpecMap[xpath].fieldName
		/* if there is no transformer extension/annotation then it means leaf-list in yang is also leaflist in db */
		if len(dbFldName) > 0  && !xYangSpecMap[xpath].isKey {
			yangType := yangTypeGet(xYangSpecMap[xpath].yangEntry)
			if yangType == "leaf-list" {
				dbFldName += "@"
				val, ok := (*dbDataMap)[cdb][tbl][tblKey].Field[dbFldName]
				if ok {
					resLst := processLfLstDbToYang(xpath, val)
					resFldValMap[xYangSpecMap[xpath].yangEntry.Name] = resLst
				}
			} else {
				val, ok := (*dbDataMap)[cdb][tbl][tblKey].Field[dbFldName]
				if ok {
					yngTerminalNdDtType := xYangSpecMap[xpath].yangEntry.Type.Kind
					resVal, err := DbToYangType(yngTerminalNdDtType, xpath, val)
					if err != nil {
						log.Error("Failure in converting Db value type to yang type for field", xpath)
					} else {
						resFldValMap[xYangSpecMap[xpath].yangEntry.Name] = resVal
					}
				}
			}
		}
	}
	return resFldValMap, err
}


func yangDataFill(dbs [db.MaxDB]*db.DB, ygRoot *ygot.GoStruct, uri string, xpath string, dbDataMap *map[db.DBNum]map[string]map[string]db.Value, resultMap map[string]interface{}, tbl string, tblKey string, cdb db.DBNum, validate bool) error {
	var err error
	isValid := validate
	yangNode, ok := xYangSpecMap[xpath]

	if ok  && yangNode.yangEntry != nil {
		for yangChldName := range yangNode.yangEntry.Dir {
			chldXpath := xpath+"/"+yangChldName
			chldUri   := uri+"/"+yangChldName
			if xYangSpecMap[chldXpath] != nil && xYangSpecMap[chldXpath].yangEntry != nil {
				_, key, _ := xpathKeyExtract(dbs[cdb], ygRoot, GET, chldUri)
				if len(xYangSpecMap[chldXpath].validateFunc) > 0 && !validate {
					// TODO - handle non CONFIG-DB
					inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, chldUri, GET, key, dbDataMap, nil)
					res := validateHandlerFunc(inParams)
					if res != true {
						continue
					} else {
						isValid = res
					}
				}
				chldYangType := yangTypeGet(xYangSpecMap[chldXpath].yangEntry)
				cdb = xYangSpecMap[chldXpath].dbIndex
				if  chldYangType == "leaf" || chldYangType == "leaf-list" {
					fldValMap, err := terminalNodeProcess(dbs, ygRoot, chldUri, chldXpath, dbDataMap, tbl, tblKey)
					if err != nil {
						log.Infof("Failed to get data(\"%v\").", chldUri)
					}
					for lf, val := range fldValMap {
						resultMap[lf] = val
					}
				} else if chldYangType == "container" {
					cname := xYangSpecMap[chldXpath].yangEntry.Name
					if len(xYangSpecMap[chldXpath].xfmrFunc) > 0 {
						inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, chldUri, GET, "", dbDataMap, nil)
						cmap, _  := xfmrHandlerFunc(inParams)
						if len(cmap) > 0 {
							resultMap[cname] = cmap
						} else {
							log.Infof("Empty container(\"%v\").\r\n", chldUri)
						}
					} else if xYangSpecMap[chldXpath].xfmrTbl != nil {
						xfmrTblFunc := *xYangSpecMap[chldXpath].xfmrTbl
						if len(xfmrTblFunc) > 0 {
							inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, chldUri, GET, "", dbDataMap, nil)
							tblList := xfmrTblHandlerFunc(xfmrTblFunc, inParams)
							if len(tblList) > 1 {
								log.Warningf("Table transformer returned more than one table for container %v", chldXpath)
							}
							dbDataFromTblXfmrGet(tblList[0], inParams, dbDataMap)
						}
					}
					cmap := make(map[string]interface{})
					err  = yangDataFill(dbs, ygRoot, chldUri, chldXpath, dbDataMap, cmap, tbl, tblKey, cdb, isValid)
					if len(cmap) > 0 {
						resultMap[cname] = cmap
					} else {
						log.Infof("Empty container(\"%v\").\r\n", chldUri)
					}
				} else if chldYangType == "list" {
					cdb = xYangSpecMap[chldXpath].dbIndex
					if len(xYangSpecMap[chldXpath].xfmrFunc) > 0 {
						inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, chldUri, GET, "", dbDataMap, nil)
						cmap, _  := xfmrHandlerFunc(inParams)
						if len(cmap) > 0 {
							resultMap = cmap
						} else {
							log.Infof("Empty list(\"%v\").\r\n", chldUri)
						}
					} else {
						ynode, ok := xYangSpecMap[chldXpath]
						lTblName := ""
						if ok && ynode.tableName != nil {
							lTblName = *ynode.tableName
						}
						yangListDataFill(dbs, ygRoot, chldUri, chldXpath, dbDataMap, resultMap, lTblName, "", cdb, isValid)
					}
				} else {
					return err
				}
			}
		}
	}
	return err
}

/* Traverse linear db-map data and add to nested json data */
func dbDataToYangJsonCreate(uri string, ygRoot *ygot.GoStruct, dbs [db.MaxDB]*db.DB, dbDataMap *map[db.DBNum]map[string]map[string]db.Value, cdb db.DBNum) (string, error) {
	jsonData := ""
	resultMap := make(map[string]interface{})
	if isCvlYang(uri) {
		directDbToYangJsonCreate(dbDataMap, jsonData, resultMap)
	} else {
		var d *db.DB
		reqXpath, keyName, tableName := xpathKeyExtract(d, ygRoot, GET, uri)
		yangNode, ok := xYangSpecMap[reqXpath]
		if ok {
			yangType := yangTypeGet(yangNode.yangEntry)
			if yangType == "leaf" || yangType == "leaf-list" {
				//fldName := xYangSpecMap[reqXpath].fieldName
				yangName := xYangSpecMap[reqXpath].yangEntry.Name
				tbl, key, _ := tableNameAndKeyFromDbMapGet((*dbDataMap)[cdb])
				validateHandlerFlag := false
				if len(xYangSpecMap[reqXpath].validateFunc) > 0 {
					// TODO - handle non CONFIG-DB
					inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, uri, GET, key, dbDataMap, nil)
					res := validateHandlerFunc(inParams)
					if !res {
						validateHandlerFlag = true
						resultMap[yangName] = ""
					}
				}
				if !validateHandlerFlag {
					fldValMap, err := terminalNodeProcess(dbs, ygRoot, uri, reqXpath, dbDataMap, tbl, key)
					//err := terminalNodeProcess(dbs, ygRoot, uri, reqXpath, dbDataMap, tbl, key)
					if err != nil {
						log.Infof("Empty terminal node (\"%v\").", uri)
					}
					resultMap = fldValMap
				}
			} else if yangType == "container" {
				cname := xYangSpecMap[reqXpath].yangEntry.Name
				cmap  := make(map[string]interface{})
				if len(xYangSpecMap[reqXpath].xfmrFunc) > 0 {
					inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, uri, GET, "", dbDataMap, nil)
					cmap, _   = xfmrHandlerFunc(inParams)
					if len(cmap) > 0 {
						resultMap[cname] = cmap
					} else {
						err    := yangDataFill(dbs, ygRoot, uri, reqXpath, dbDataMap, resultMap, tableName, keyName, cdb, false)
						if err != nil {
							log.Infof("Empty container(\"%v\").\r\n", uri)
						}
					}
				} else {
					err    := yangDataFill(dbs, ygRoot, uri, reqXpath, dbDataMap, resultMap, tableName, keyName, cdb, false)
					if err != nil {
						log.Infof("Empty container(\"%v\").\r\n", uri)
					}
				}
			} else {
				yangDataFill(dbs, ygRoot, uri, reqXpath, dbDataMap, resultMap, tableName, keyName, cdb, false)
			}
		}
	}

	jsonMapData, _ := json.Marshal(resultMap)
	jsonData        = fmt.Sprintf("%v", string(jsonMapData))
	jsonDataPrint(jsonData)
	return jsonData, nil
}

func jsonDataPrint(data string) {
    fp, err := os.Create("/tmp/dbToYangJson.txt")
    if err != nil {
        return
    }
    defer fp.Close()

    fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")
    fmt.Fprintf (fp, "%v \r\n", data)
    fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")
}

