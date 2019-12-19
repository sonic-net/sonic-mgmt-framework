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

package transformer

import (
    "fmt"
    "github.com/openconfig/ygot/ygot"
    "os"
    "reflect"
    "strings"
    "translib/db"
    "translib/ocbinds"
    "translib/tlerr"
    "github.com/openconfig/ygot/ytypes"
    "github.com/openconfig/goyang/pkg/yang"
    log "github.com/golang/glog"
)

var ocbSch, _ = ocbinds.Schema()

func leafXfmrHandler(inParams XfmrParams, xfmrFieldFuncNm string) (map[string]string, error) {
        log.Info("Received inParams ", inParams, " Field transformer name ", xfmrFieldFuncNm)
    ret, err := XlateFuncCall(yangToDbXfmrFunc(xfmrFieldFuncNm), inParams)
    if err != nil {
        return nil, err
    }
    if ((ret != nil) && (len(ret)>0)) {
            if len(ret) == 2 {
                    // field xfmr returns err as second value in return data list from <xfmr_func>.Call()
                    if ret[1].Interface() != nil {
                            err = ret[1].Interface().(error)
                            if err != nil {
                                    log.Warningf("Transformer function(\"%v\") returned error - %v.", xfmrFieldFuncNm, err)
                                    return nil, err
                            }
                    }
            }

            if ret[0].Interface() != nil {
                    fldValMap := ret[0].Interface().(map[string]string)
                    return fldValMap, nil
            }
    } else {
            retFldValMap := map[string]string{"NULL":"NULL"}
            return retFldValMap, nil
    }

    return nil, nil
}

func xfmrHandler(inParams XfmrParams, xfmrFuncNm string) (map[string]map[string]db.Value, error) {
        log.Info("Received inParams ", inParams, "Subtree function name ", xfmrFuncNm)
        ret, err := XlateFuncCall(yangToDbXfmrFunc(xfmrFuncNm), inParams)
        if err != nil {
                return nil, err
        }

        if ((ret != nil) && (len(ret)>0)) {
                if len(ret) == 2 {
                        // subtree xfmr returns err as second value in return data list from <xfmr_func>.Call()
                        if ret[1].Interface() != nil {
                                err = ret[1].Interface().(error)
                                if err != nil {
                                        log.Warningf("Transformer function(\"%v\") returned error - %v.", xfmrFuncNm, err)
                                        return nil, err
                                }
                        }
                }
                if ret[0].Interface() != nil {
                        retMap := ret[0].Interface().(map[string]map[string]db.Value)
                        return retMap, nil
                }
        }
        return nil, nil
}

/* Invoke the post tansformer */
func postXfmrHandlerFunc(xfmrPost string, inParams XfmrParams) (map[string]map[string]db.Value, error) {
    retData := make(map[string]map[string]db.Value)
    ret, err := XlateFuncCall(xfmrPost, inParams)
    if err != nil {
        return nil, err
    }
    if ((ret != nil) && (len(ret)>0)) {
        retData = ret[0].Interface().(map[string]map[string]db.Value)
        log.Info("Post Transformer function :", xfmrPost, " retData: ", retData)
    }
    return retData, err
}

/* Fill redis-db map with field & value info */
func dataToDBMapAdd(tableName string, dbKey string, result map[string]map[string]db.Value, field string, value string) {
    _, ok := result[tableName]
    if !ok {
        result[tableName] = make(map[string]db.Value)
    }

    _, ok = result[tableName][dbKey]
    if !ok {
        result[tableName][dbKey] = db.Value{Field: make(map[string]string)}
    }

	if field == XFMR_NONE_STRING {
		if len(result[tableName][dbKey].Field) == 0 {
			result[tableName][dbKey].Field["NULL"] = "NULL"
		}
		return
	}

    result[tableName][dbKey].Field[field] = value
    return
}

func tblNameFromTblXfmrGet(xfmrTblFunc string, inParams XfmrParams) (string, error){
	tblList := xfmrTblHandlerFunc(xfmrTblFunc, inParams)
	if len(tblList) != 1 {
		logStr := fmt.Sprintf("Invalid return value(%v) from table transformer for (%v)", tblList, inParams.uri)
		log.Error(logStr)
		err := tlerr.InternalError{Format: logStr}
		return "", err
	}
	return tblList[0], nil
}

/* Fill the redis-db map with data */
func mapFillData(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, requestUri string, dbKey string, result map[string]map[string]db.Value, subOpDataMap map[int]*RedisDbMap, xpathPrefix string, name string, value interface{}, tblXpathMap map[string][]string, txCache interface{}, xfmrErr *error) error {
	var dbs [db.MaxDB]*db.DB
	var err error
    xpath := xpathPrefix + "/" + name
    xpathInfo, ok := xYangSpecMap[xpath]
    log.Infof("name: \"%v\", xpathPrefix(\"%v\").", name, xpathPrefix)

    if !ok || xpathInfo == nil {
        log.Errorf("Yang path(\"%v\") not found.", xpath)
	return nil
    }

    if xpathInfo.tableName == nil && xpathInfo.xfmrTbl == nil{
        log.Errorf("Table for yang-path(\"%v\") not found.", xpath)
	return nil
    }

    if xpathInfo.tableName != nil && *xpathInfo.tableName == XFMR_NONE_STRING {
        log.Errorf("Table for yang-path(\"%v\") NONE.", xpath)
	return nil
    }

    if len(dbKey) == 0 {
        log.Errorf("Table key for yang path(\"%v\") not found.", xpath)
	return nil
    }

    tableName := ""
    if xpathInfo.xfmrTbl != nil {
	    inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, uri, requestUri, oper, "", nil, subOpDataMap, "", txCache)
	    // expecting only one table name from tbl-xfmr
	    tableName, _ = tblNameFromTblXfmrGet(*xYangSpecMap[xpath].xfmrTbl, inParams)
	    if tableName == "" {
		    return err
	    }
		tblXpathMap[tableName] = append(tblXpathMap[tableName], xpathPrefix)
    } else {
	    tableName = *xpathInfo.tableName
    }

	err = mapFillDataUtil(d, ygRoot, oper, uri, requestUri, xpath, tableName, dbKey, result, subOpDataMap, name, value, txCache, xfmrErr);
	return err
}

func mapFillDataUtil(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, requestUri string, xpath string, tableName string, dbKey string, result map[string]map[string]db.Value, subOpDataMap map[int]*RedisDbMap, name string, value interface{}, txCache interface{}, xfmrErr *error) error {
	var dbs [db.MaxDB]*db.DB
	xpathInfo := xYangSpecMap[xpath]

	if len(xpathInfo.xfmrField) > 0 {
		uri = uri + "/" + name

		/* field transformer present */
		log.Infof("Transformer function(\"%v\") invoked for yang path(\"%v\"). uri: %v", xpathInfo.xfmrField, xpath, uri)
		curYgotNodeData, nodeErr := yangNodeForUriGet(uri, ygRoot)
		if nodeErr != nil {
			return nil
		}
		inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, uri, requestUri, oper, "", nil, subOpDataMap, curYgotNodeData, txCache)
		retData, err := leafXfmrHandler(inParams, xpathInfo.xfmrField)
		if err != nil {
			if xfmrErr != nil && *xfmrErr == nil {
				*xfmrErr = err
			}
			return err
		}
		if retData != nil {
			log.Info("Transformer function :", xpathInfo.xfmrField, " Xpath: ", xpath, " retData: ", retData)
			for f, v := range retData {
				dataToDBMapAdd(tableName, dbKey, result, f, v)
			}
		}
		return nil
	}

	if len(xpathInfo.fieldName) == 0 {
		log.Infof("Field for yang-path(\"%v\") not found in DB.", xpath)
		errStr := fmt.Sprintf("Field for yang-path(\"%v\") not found in DB.", xpath)
		return tlerr.InternalError{Format: errStr}
	}
	fieldName := xpathInfo.fieldName
	valueStr := ""
	if xpathInfo.yangEntry.IsLeafList() {
		/* Both yang side and Db side('@' suffix field) the data type is leaf-list */
		log.Info("Yang type and Db type is Leaflist for field  = ", xpath)
		fieldName += "@"
		if reflect.ValueOf(value).Kind() != reflect.Slice {
			logStr := fmt.Sprintf("Value for yang xpath %v which is a leaf-list should be a slice", xpath)
			log.Error(logStr)
			return nil
		}
		valData := reflect.ValueOf(value)
		for fidx := 0; fidx < valData.Len(); fidx++ {
			if fidx > 0 {
				valueStr += ","
			}
			fVal := fmt.Sprintf("%v", valData.Index(fidx).Interface())
			if ((strings.Contains(fVal, ":")) && (strings.HasPrefix(fVal, OC_MDL_PFX) || strings.HasPrefix(fVal, IETF_MDL_PFX) || strings.HasPrefix(fVal, IANA_MDL_PFX))) {
				// identity-ref/enum has module prefix
				fVal = strings.SplitN(fVal, ":", 2)[1]
			}
			valueStr = valueStr + fVal
		}
		log.Infof("leaf-list value after conversion to DB format %v  :  %v", fieldName, valueStr)

	} else { // xpath is a leaf
		valueStr  = fmt.Sprintf("%v", value)
		if ((strings.Contains(valueStr, ":")) && (strings.HasPrefix(valueStr, OC_MDL_PFX) || strings.HasPrefix(valueStr, IETF_MDL_PFX) || strings.HasPrefix(valueStr, IANA_MDL_PFX))) {
			// identity-ref/enum might has module prefix
			valueStr = strings.SplitN(valueStr, ":", 2)[1]
		}
	}

	dataToDBMapAdd(tableName, dbKey, result, fieldName, valueStr)
	log.Infof("TblName: \"%v\", key: \"%v\", field: \"%v\", valueStr: \"%v\".", tableName, dbKey,
	fieldName, valueStr)
	return nil
}

func sonicYangReqToDbMapCreate(jsonData interface{}, result map[string]map[string]db.Value) error {
    if reflect.ValueOf(jsonData).Kind() == reflect.Map {
        data := reflect.ValueOf(jsonData)
        for _, key := range data.MapKeys() {
            _, ok := xDbSpecMap[key.String()]
            if ok {
                directDbMapData("", key.String(), data.MapIndex(key).Interface(), result)
            } else {
                sonicYangReqToDbMapCreate(data.MapIndex(key).Interface(), result)
            }
        }
    }
    return nil
}

func dbMapDataFill(uri string, tableName string, keyName string, d map[string]interface{}, result map[string]map[string]db.Value) {
	result[tableName][keyName] = db.Value{Field: make(map[string]string)}
	for field, value := range d {
		fieldXpath := tableName + "/" + field
		if _, fieldOk := xDbSpecMap[fieldXpath]; (fieldOk  && (xDbSpecMap[fieldXpath].dbEntry != nil)) {
			log.Info("Found non-nil yang entry in xDbSpecMap for field xpath = ", fieldXpath)
			if xDbSpecMap[fieldXpath].dbEntry.IsLeafList() {
				log.Info("Yang type is Leaflist for field  = ", field)
				field += "@"
				fieldDt := reflect.ValueOf(value)
				fieldValue := ""
				for fidx := 0; fidx < fieldDt.Len(); fidx++ {
					if fidx > 0 {
						fieldValue += ","
					}
					fVal := fmt.Sprintf("%v", fieldDt.Index(fidx).Interface())
					fieldValue = fieldValue + fVal
				}
				result[tableName][keyName].Field[field] = fieldValue
				continue
			}
		} else {
			// should ideally never happen , just adding for safety
			log.Info("Did not find entry in xDbSpecMap for field xpath = ", fieldXpath)
		}
		dbval, err := unmarshalJsonToDbData(xDbSpecMap[fieldXpath].dbEntry, field, value)
		if err != nil {
			log.Errorf("Failed to unmashal Json to DbData: path(\"%v\") error (\"%v\").", fieldXpath, err)
		} else {
			result[tableName][keyName].Field[field] = dbval
		}
	}
	return
}

func dbMapListDataFill(uri string, tableName string, dbEntry *yang.Entry, jsonData interface{}, result map[string]map[string]db.Value) {
	data := reflect.ValueOf(jsonData)
	tblKeyName := strings.Split(dbEntry.Key, " ")
	for idx := 0; idx < data.Len(); idx++ {
		keyName := ""
		d := data.Index(idx).Interface().(map[string]interface{})
		for i, k := range tblKeyName {
			if i > 0 {
				keyName += "|"
			}
			keyName += fmt.Sprintf("%v", d[k])
			delete(d, k)
		}
		dbMapDataFill(uri, tableName, keyName, d, result)
	}
	return
}

func directDbMapData(uri string, tableName string, jsonData interface{}, result map[string]map[string]db.Value) bool {
	_, ok := xDbSpecMap[tableName]
	if ok && xDbSpecMap[tableName].dbEntry != nil {
		data := reflect.ValueOf(jsonData).Interface().(map[string]interface{})
		key  := ""
		dbSpecData := xDbSpecMap[tableName]
		result[tableName] = make(map[string]db.Value)

		if dbSpecData.keyName != nil {
			key = *dbSpecData.keyName
			log.Infof("Fill data for container uri(%v), key(%v)", uri, key)
			dbMapDataFill(uri, tableName, key, data, result)
			return true
		}

		for k, v := range data {
			xpath := tableName + "/" + k
			curDbSpecData, ok := xDbSpecMap[xpath]
			if ok && curDbSpecData.dbEntry != nil {
				eType := yangTypeGet(curDbSpecData.dbEntry)
				switch eType {
				case "list":
					log.Infof("Fill data for list uri(%v)", uri)
					dbMapListDataFill(uri, tableName, curDbSpecData.dbEntry, v, result)
				default:
					log.Infof("Invalid node type for uri(%v)", uri)
				}
			}
		}
	}
	return true
}

/* Get the db table, key and field name for the incoming delete request */
func dbMapDelete(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, requestUri string, jsonData interface{}, resultMap map[int]map[db.DBNum]map[string]map[string]db.Value, txCache interface{}, skipOrdTbl *bool) error {
	var err error
	var result = make(map[string]map[string]db.Value)
	subOpDataMap := make(map[int]*RedisDbMap)
	var xfmrErr error
	*skipOrdTbl = false

    for i := 0; i < MAXOPER; i++ {
        resultMap[i] = make(map[db.DBNum]map[string]map[string]db.Value)
    }
	if isSonicYang(uri) {
		*skipOrdTbl = true
		xpathPrefix, keyName, tableName := sonicXpathKeyExtract(uri)
		log.Infof("Delete req: uri(\"%v\"), key(\"%v\"), xpathPrefix(\"%v\"), tableName(\"%v\").", uri, keyName, xpathPrefix, tableName)
		resultMap[oper][db.ConfigDB] = result
		err = sonicYangReqToDbMapDelete(requestUri, xpathPrefix, tableName, keyName, result)
	} else {
		xpathPrefix, keyName, tableName := xpathKeyExtract(d, ygRoot, oper, uri, requestUri, subOpDataMap, txCache)
		log.Infof("Delete req: uri(\"%v\"), key(\"%v\"), xpathPrefix(\"%v\"), tableName(\"%v\").", uri, keyName, xpathPrefix, tableName)
		spec, ok := xYangSpecMap[xpathPrefix]
		if ok {
			specYangType := yangTypeGet(spec.yangEntry)
			moduleNm := "/" + strings.Split(uri, "/")[1]
			log.Infof("Module name for uri %s is %s", uri, moduleNm)
			if len(spec.xfmrFunc) > 0 {
				var dbs [db.MaxDB]*db.DB
				cdb := spec.dbIndex
				inParams := formXfmrInputRequest(d, dbs, cdb, ygRoot, uri, requestUri, oper, "", nil, subOpDataMap, nil, txCache)
				stRetData, err := xfmrHandler(inParams, spec.xfmrFunc)
				if err == nil {
					mapCopy(result, stRetData)
				} else {
					return err
				}
			} else if len(tableName) > 0 {
				result[tableName] = make(map[string]db.Value)
				if len(keyName) > 0 {
					result[tableName][keyName] = db.Value{Field: make(map[string]string)}
					xpath := xpathPrefix
					uriItemList := strings.Split(strings.TrimSuffix(uri, "/"), "/")
					uriItemListLen := len(uriItemList)
					var terminalNode, luri string
					if uriItemListLen > 0 {
						terminalNode = uriItemList[uriItemListLen-1]
						luri = strings.Join(uriItemList[:uriItemListLen-1], "/") //strip off the leaf/leaf-list for mapFillDataUtil takes uri without it

					}
					if specYangType == YANG_LEAF {
						if len(xYangSpecMap[xpath].defVal) > 0 {
							err = mapFillDataUtil(d, ygRoot, oper, luri, requestUri, xpath, tableName, keyName, result, subOpDataMap, spec.fieldName, xYangSpecMap[xpath].defVal, txCache, &xfmrErr)
							if xfmrErr != nil {
								return xfmrErr
							}
							if err != nil {
								return err
							}
							if len(subOpDataMap) > 0 && subOpDataMap[UPDATE] != nil {
								subOperMap := subOpDataMap[UPDATE]
								mapCopy((*subOperMap)[db.ConfigDB], result)
							} else {
								var redisMap = new(RedisDbMap)
								var dbresult = make(RedisDbMap)
                                for i := db.ApplDB; i < db.MaxDB; i++ {
                                    dbresult[i] = make(map[string]map[string]db.Value)
                                }
                                redisMap = &dbresult
                                (*redisMap)[db.ConfigDB] = result
								subOpDataMap[UPDATE]     = redisMap
							}
							result = make(map[string]map[string]db.Value)
						} else {
							result[tableName][keyName].Field[spec.fieldName] = ""
						}
					} else if specYangType == YANG_LEAF_LIST {
						var fieldVal []interface{}
						if strings.Contains(terminalNode, "[") {
							terminalNodeData := strings.TrimSuffix(strings.SplitN(terminalNode, "[", 2)[1], "]")
							terminalNodeDataLst := strings.SplitN(terminalNodeData, "=", 2)
							terminalNodeVal := terminalNodeDataLst[1]
							fieldVal = append(fieldVal, terminalNodeVal)
						}
						err = mapFillDataUtil(d, ygRoot, oper, luri, requestUri, xpath, tableName, keyName, result, subOpDataMap, spec.yangEntry.Name, fieldVal, txCache, &xfmrErr)

						if xfmrErr != nil {
							return xfmrErr
						}
						if err != nil {
							return err
						}
					}
				}
			} else if len(spec.childTable) > 0 {
				for _, child := range spec.childTable {
					result[child] = make(map[string]db.Value)
				}
			}

			if len(xYangSpecMap[moduleNm].xfmrPost) > 0 {
				log.Info("Invoke post transformer: ", xYangSpecMap[moduleNm].xfmrPost)
				var dbs [db.MaxDB]*db.DB
				var dbresult = make(RedisDbMap)
				dbresult[db.ConfigDB] = result
				inParams := formXfmrInputRequest(d, dbs, db.ConfigDB, ygRoot, uri, requestUri, oper, "", &dbresult, subOpDataMap, nil, txCache)
				result, err = postXfmrHandlerFunc(xYangSpecMap[moduleNm].xfmrPost, inParams)
				if inParams.skipOrdTblChk != nil {
					*skipOrdTbl = *(inParams.skipOrdTblChk)
					log.Info("skipOrdTbl flag: ", *skipOrdTbl)
				}
			}

			if len(result) > 0 {
				resultMap[oper][db.ConfigDB] = result
			}

			if len(subOpDataMap) > 0 {
				for op, data := range subOpDataMap {
					if len(*data) > 0 {
						for dbType, dbData := range (*data) {
							if len(dbData) > 0 {
                                if _, ok := resultMap[op][dbType]; !ok {
                                    resultMap[op][dbType] = make(map[string]map[string]db.Value)
                                }
                                mapCopy(resultMap[op][dbType], (*subOpDataMap[op])[dbType])
							}
						}
					}
				}

			}
			/* for container/list delete req , it should go through, even if there are any leaf default-yang-values */
		}
	}

    printDbData(resultMap, "/tmp/yangToDbDataDel.txt")
	log.Infof("Delete req: uri(\"%v\") resultMap(\"%v\").", uri, resultMap)
	return err
}

func sonicYangReqToDbMapDelete(requestUri string, xpathPrefix string, tableName string, keyName string, result map[string]map[string]db.Value) error {
	var err error
    if (tableName != "") {
        // Specific table entry case
        result[tableName] = make(map[string]db.Value)
        if (keyName != "") {
            // Specific key case
            var dbVal db.Value
            tokens:= strings.Split(xpathPrefix, "/")
            if tokens[SONIC_TABLE_INDEX] == tableName {
		    fieldName := ""
		    if len(tokens) > SONIC_FIELD_INDEX {
			    fieldName = tokens[SONIC_FIELD_INDEX]
		    }

		     if fieldName != "" {
			     dbSpecField := tableName + "/" + fieldName
			     _, ok := xDbSpecMap[dbSpecField]
			     if ok {
				     yangType := xDbSpecMap[dbSpecField].fieldType
				     // terminal node case
				     if yangType == YANG_LEAF_LIST {
					     dbVal.Field = make(map[string]string)
					     //check if it is a specific item in leaf-list delete
					     uriItemList := strings.Split(strings.TrimSuffix(requestUri, "/"), "/")
					     uriItemListLen := len(uriItemList)
					     var terminalNode string
					     if uriItemListLen > 0 {
						     terminalNode = uriItemList[uriItemListLen-1]
						     dbFldVal := ""
						     if strings.Contains(terminalNode, "[") {
							     terminalNodeData := strings.TrimSuffix(strings.SplitN(terminalNode, "[", 2)[1], "]")
							     terminalNodeDataLst := strings.SplitN(terminalNodeData, "=", 2)
							     terminalNodeVal := terminalNodeDataLst[1]
							     dbFldVal, err = unmarshalJsonToDbData(xDbSpecMap[dbSpecField].dbEntry, fieldName, terminalNodeVal)
							     if err != nil {
								     log.Errorf("Failed to unmashal Json to DbData: path(\"%v\") error (\"%v\").", dbSpecField, err)
								     return err
							     }
						     }
						     fieldName = fieldName + "@"
						     dbVal.Field[fieldName] = dbFldVal
					     }
				     }
				     if yangType == YANG_LEAF {
					     dbVal.Field = make(map[string]string)
					     dbVal.Field[fieldName] = ""
				     }
			     }
		     }
	     }
	     result[tableName][keyName] = dbVal
        } else {
            // Get all keys
        }
    } else {
        // Get all table entries
        // If table name not available in xpath get top container name
	_, ok := xDbSpecMap[xpathPrefix]
        if ok && xDbSpecMap[xpathPrefix] != nil {
            dbInfo := xDbSpecMap[xpathPrefix]
            if dbInfo.fieldType == "container" {
                for dir, _ := range dbInfo.dbEntry.Dir {
                    if dbInfo.dbEntry.Dir[dir].Config != yang.TSFalse {
                       result[dir] = make(map[string]db.Value)
                    }
                }
            }
        }
    }
    return nil
}

/* Get the data from incoming update/replace request, create map and fill with dbValue(ie. field:value to write into redis-db */
func dbMapUpdate(d *db.DB, ygRoot *ygot.GoStruct, oper int, path string, requestUri string, jsonData interface{}, result map[int]map[db.DBNum]map[string]map[string]db.Value, txCache interface{}) error {
    log.Infof("Update/replace req: path(\"%v\").", path)
    var err error
    err = dbMapCreate(d, ygRoot, oper, path, requestUri, jsonData, result, txCache)
    log.Infof("Update/replace req: path(\"%v\") result(\"%v\").", path, result)
    printDbData(result, "/tmp/yangToDbDataUpRe.txt")
    return err
}

func dbMapDefaultFieldValFill(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, requestUri string, result map[string]map[string]db.Value, subOpDataMap map[int]*RedisDbMap, yangXpathList []string, tblName string, dbKey string, txCache interface{}) error {
	tblData := result[tblName]
	var dbs [db.MaxDB]*db.DB
	for _, yangXpath := range yangXpathList {
		yangNode, ok := xYangSpecMap[yangXpath]
		if ok {
			for childName  := range yangNode.yangEntry.Dir {
				childXpath := yangXpath + "/" + childName
				childNode, ok := xYangSpecMap[childXpath]
				if ok {
					if childNode.yangDataType == YANG_LIST || childNode.yangDataType == YANG_CONTAINER {
						var tblList []string
						tblList = append(tblList, childXpath)
						err := dbMapDefaultFieldValFill(d, ygRoot, oper, uri, requestUri, result, subOpDataMap, tblList, tblName, dbKey, txCache)
						if err != nil {
							return err
						}
					}
					if (childNode.tableName != nil && *childNode.tableName == tblName) || (childNode.xfmrTbl != nil) {
						_, ok := tblData[dbKey].Field[childName]
						if !ok && len(childNode.defVal) > 0  && len(childNode.fieldName) > 0 {
							log.Infof("Update(\"%v\") default: tbl[\"%v\"]key[\"%v\"]fld[\"%v\"] = val(\"%v\").",
							childXpath, tblName, dbKey, childNode.fieldName, childNode.defVal)
							if len(childNode.xfmrField) > 0 {
								childYangType := childNode.yangEntry.Type.Kind
								_, defValPtr, err := DbToYangType(childYangType, childXpath, childNode.defVal)
								if err == nil && defValPtr != nil {
									inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, childXpath, requestUri, oper, "", nil, subOpDataMap, defValPtr, txCache)
									retData, err := leafXfmrHandler(inParams, childNode.xfmrField)
									if err != nil {
										return err
									}
									if retData != nil {
										log.Info("Transformer function :", childNode.xfmrField, " Xpath: ", childXpath, " retData: ", retData)
										for f, v := range retData {
											dataToDBMapAdd(tblName, dbKey, result, f, v)
										}
									}

								} else {
									log.Infof("Failed to update(\"%v\") default: tbl[\"%v\"]key[\"%v\"]fld[\"%v\"] = val(\"%v\").",
									childXpath, tblName, dbKey, childNode.fieldName, childNode.defVal)
								}
							} else {
								var xfmrErr error
								err := mapFillDataUtil(d, ygRoot, oper, uri, requestUri, childXpath, tblName, dbKey, result, subOpDataMap, childName, childNode.defVal, txCache, &xfmrErr)
								if err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func dbMapDefaultValFill(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, requestUri string, result map[string]map[string]db.Value, subOpDataMap map[int]*RedisDbMap, tblXpathMap map[string][]string, txCache interface{}) error {
	for tbl, tblData := range result {
		for dbKey, _ := range tblData {
			yxpathList := xDbSpecMap[tbl].yangXpath
			if _, ok := tblXpathMap[tbl]; ok {
				yxpathList = tblXpathMap[tbl]
			}
			err := dbMapDefaultFieldValFill(d, ygRoot, oper, uri, requestUri, result, subOpDataMap, yxpathList, tbl, dbKey, txCache)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/* Get the data from incoming create request, create map and fill with dbValue(ie. field:value to write into redis-db */
func dbMapCreate(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, requestUri string, jsonData interface{}, resultMap map[int]map[db.DBNum]map[string]map[string]db.Value, txCache interface{}) error {
	var err error
	tblXpathMap := make(map[string][]string)
	var result = make(map[string]map[string]db.Value)
	subOpDataMap := make(map[int]*RedisDbMap)
	root := xpathRootNameGet(uri)
	var xfmrErr error

	if isSonicYang(uri) {
		err = sonicYangReqToDbMapCreate(jsonData, result)
		resultMap[oper] = make(RedisDbMap)
		resultMap[oper][db.ConfigDB] = result
	} else {
		err = yangReqToDbMapCreate(d, ygRoot, oper, root, uri, "", "", jsonData, result, subOpDataMap, tblXpathMap, txCache, &xfmrErr)
		if xfmrErr != nil {
			return xfmrErr
		}
		if err != nil {
			return err
		}
	}
	if err == nil {
		if !isSonicYang(uri) {
			xpath, _ := XfmrRemoveXPATHPredicates(uri)
			yangNode, ok := xYangSpecMap[xpath]
			if ok && yangNode.yangDataType != YANG_LEAF && (oper == CREATE || oper == REPLACE) {
				log.Infof("Fill default value for %v, oper(%v)\r\n", uri, oper)
				err = dbMapDefaultValFill(d, ygRoot, oper, uri, requestUri, result, subOpDataMap, tblXpathMap, txCache)
				if err != nil {
					return err
				}
			}

			if ok && yangNode.yangDataType == YANG_LEAF && oper == REPLACE {
				log.Infof("Fill default value for %v, oper(%v)\r\n", uri, oper)
				resultMap[UPDATE] = make(RedisDbMap)
				resultMap[UPDATE][db.ConfigDB] = result
				result = make(map[string]map[string]db.Value)
			}

			moduleNm := "/" + strings.Split(uri, "/")[1]
			log.Infof("Module name for uri %s is %s", uri, moduleNm)
			if _, ok := xYangSpecMap[moduleNm]; ok {
				if xYangSpecMap[moduleNm].yangDataType == "container" && len(xYangSpecMap[moduleNm].xfmrPost) > 0 {
					log.Infof("Invoke post transformer: %v, result map: %v ", xYangSpecMap[moduleNm].xfmrPost, result)
					var dbDataMap = make(RedisDbMap)
					dbDataMap[db.ConfigDB] = result
					var dbs [db.MaxDB]*db.DB
					inParams := formXfmrInputRequest(d, dbs, db.ConfigDB, ygRoot, uri, requestUri, oper, "", &dbDataMap, subOpDataMap, nil, txCache)
					result, err = postXfmrHandlerFunc(xYangSpecMap[moduleNm].xfmrPost, inParams)
				}
			} else {
				log.Errorf("No Entry exists for module %s in xYangSpecMap. Unable to process post xfmr (\"%v\") uri(\"%v\") error (\"%v\").", oper, uri, err)
			}
                        if len(result) > 0 || len(subOpDataMap) > 0 {
                                  resultMap[oper] = make(RedisDbMap)
                                  resultMap[oper][db.ConfigDB] = result
                                  for op, redisMapPtr := range subOpDataMap {
                                         if redisMapPtr != nil {
                                                 if _,ok := resultMap[op]; !ok {
                                                       resultMap[op] = make(RedisDbMap)
                                               }
                                               for dbNum, dbMap := range *redisMapPtr {
                                                       if _,ok := resultMap[op][dbNum]; !ok {
                                                               resultMap[op][dbNum] = make(map[string]map[string]db.Value)
                                                       }
                                                       mapCopy(resultMap[op][dbNum],dbMap)
                                               }
                                         }
                                  }
                        }

		}
		printDbData(resultMap, "/tmp/yangToDbDataCreate.txt")
	} else {
		log.Errorf("DBMapCreate req failed for oper (\"%v\") uri(\"%v\") error (\"%v\").", oper, uri, err)
	}
	return err
}

func yangNodeForUriGet(uri string, ygRoot *ygot.GoStruct) (interface{}, error) {
	path, _ := ygot.StringToPath(uri, ygot.StructuredPath, ygot.StringSlicePath)
	for _, p := range path.Elem {
		pathSlice := strings.Split(p.Name, ":")
		p.Name = pathSlice[len(pathSlice)-1]
		if len(p.Key) > 0 {
			for ekey, ent := range p.Key {
				// SNC-2126: check the occurrence of ":"
				if ((strings.Contains(ent, ":")) && (strings.HasPrefix(ent, OC_MDL_PFX) || strings.HasPrefix(ent, IETF_MDL_PFX) || strings.HasPrefix(ent, IANA_MDL_PFX))) {
					// identity-ref/enum has module prefix
					eslice := strings.SplitN(ent, ":", 2)
					// TODO - exclude the prexix by checking enum type
					p.Key[ekey] = eslice[len(eslice)-1]
				} else {
					p.Key[ekey] = ent
				}
			}
		}
	}
	schRoot := ocbSch.RootSchema()
	node, nErr := ytypes.GetNode(schRoot, (*ygRoot).(*ocbinds.Device), path)
	if nErr != nil {
		log.Warningf("For uri %v - GetNode failure - %v", uri, nErr)
		errStr := fmt.Sprintf("%v", nErr)
		return nil, tlerr.InternalError{Format: errStr}
	}
	if ((node == nil) || (len(node) == 0) || (node[0].Data == nil)) {
		log.Warningf("GetNode returned nil for uri %v", uri)
		errStr := "GetNode returned nil for the given uri."
		return nil, tlerr.InternalError{Format: errStr}
	}
	log.Info("GetNode data: ", node[0].Data)
	return node[0].Data, nil
}

func yangReqToDbMapCreate(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, requestUri string, xpathPrefix string, keyName string, jsonData interface{}, result map[string]map[string]db.Value, subOpDataMap map[int]*RedisDbMap, tblXpathMap map[string][]string, txCache interface{}, xfmrErr *error) error {
	log.Infof("key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)
	var dbs [db.MaxDB]*db.DB
	var retErr error

	if reflect.ValueOf(jsonData).Kind() == reflect.Slice {
		log.Infof("slice data: key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)
		jData := reflect.ValueOf(jsonData)
		dataMap := make([]interface{}, jData.Len())
		for idx := 0; idx < jData.Len(); idx++ {
			dataMap[idx] = jData.Index(idx).Interface()
		}
		for _, data := range dataMap {
			curKey := ""
			curUri, _ := uriWithKeyCreate(uri, xpathPrefix, data)
			_, ok := xYangSpecMap[xpathPrefix]
			if ok && len(xYangSpecMap[xpathPrefix].xfmrKey) > 0 {
				/* key transformer present */
				curYgotNode, nodeErr := yangNodeForUriGet(curUri, ygRoot)
				if nodeErr != nil {
					curYgotNode = nil
				}
				inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, curUri, requestUri, oper, "", nil, subOpDataMap, curYgotNode, txCache)
				ret, err := XlateFuncCall(yangToDbXfmrFunc(xYangSpecMap[xpathPrefix].xfmrKey), inParams)
				if err != nil {
					return err
				}
				if ret != nil {
					curKey = ret[0].Interface().(string)
				}
			} else if xYangSpecMap[xpathPrefix].keyName != nil {
				curKey = *xYangSpecMap[xpathPrefix].keyName
			} else {
				curKey = keyCreate(keyName, xpathPrefix, data, d.Opts.KeySeparator)
			}
			retErr = yangReqToDbMapCreate(d, ygRoot, oper, curUri, requestUri, xpathPrefix, curKey, data, result, subOpDataMap, tblXpathMap, txCache, xfmrErr)
		}
	} else {
		if reflect.ValueOf(jsonData).Kind() == reflect.Map {
			jData := reflect.ValueOf(jsonData)
			for _, key := range jData.MapKeys() {
				typeOfValue := reflect.TypeOf(jData.MapIndex(key).Interface()).Kind()

				log.Infof("slice/map data: key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)
				xpath    := uri
				curUri   := uri
				curKey   := keyName
				pathAttr := key.String()
				if len(xpathPrefix) > 0 {
					if strings.Contains(pathAttr, ":") {
						pathAttr = strings.Split(pathAttr, ":")[1]
					}
					xpath  = xpathPrefix + "/" + pathAttr
					curUri = uri + "/" + pathAttr
				}
				_, ok := xYangSpecMap[xpath]
				log.Infof("slice/map data: curKey(\"%v\"), xpath(\"%v\"), curUri(\"%v\").",
				curKey, xpath, curUri)
				if ok && xYangSpecMap[xpath] != nil && len(xYangSpecMap[xpath].xfmrKey) > 0 {
					curYgotNode, nodeErr := yangNodeForUriGet(curUri, ygRoot)
					if nodeErr != nil {
						curYgotNode = nil
					}
					inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, curUri, requestUri, oper, "", nil, subOpDataMap, curYgotNode, txCache)
					ret, err := XlateFuncCall(yangToDbXfmrFunc(xYangSpecMap[xpath].xfmrKey), inParams)
					if err != nil {
						return err
					}
					if ret != nil {
						curKey = ret[0].Interface().(string)
					}
				} else if ok && xYangSpecMap[xpath].keyName != nil {
					curKey = *xYangSpecMap[xpath].keyName
				}

				if ok && (typeOfValue == reflect.Map || typeOfValue == reflect.Slice) && xYangSpecMap[xpath].yangDataType != "leaf-list" {
					log.Infof("CurUri: %v, requestUri: %v", curUri, requestUri)
					// Call subtree only if start processing for the requestUri. Skip for parent uri traversal
					if strings.HasPrefix(curUri,requestUri) {
						if xYangSpecMap[xpath] != nil && len(xYangSpecMap[xpath].xfmrFunc) > 0 &&
						(xYangSpecMap[xpathPrefix] != xYangSpecMap[xpath]) {
							/* subtree transformer present */
							curYgotNode, nodeErr := yangNodeForUriGet(curUri, ygRoot)
							if nodeErr != nil {
								curYgotNode = nil
							}
							inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, curUri, requestUri, oper, "", nil, subOpDataMap, curYgotNode, txCache)
							stRetData, err := xfmrHandler(inParams, xYangSpecMap[xpath].xfmrFunc)
							if err != nil {
								if xfmrErr != nil && *xfmrErr == nil {
									*xfmrErr = err
                                                                }
								return nil
							}
							if stRetData != nil {
								mapCopy(result, stRetData)
							}
						}
					}
					retErr = yangReqToDbMapCreate(d, ygRoot, oper, curUri, requestUri, xpath, curKey, jData.MapIndex(key).Interface(), result, subOpDataMap, tblXpathMap, txCache, xfmrErr)
				} else {
					pathAttr := key.String()
					if strings.Contains(pathAttr, ":") {
						pathAttr = strings.Split(pathAttr, ":")[1]
					}
					xpath := xpathPrefix + "/" + pathAttr
					log.Infof("LEAF Case: xpath: %v, xpathPrefix: %v, pathAttr: %v", xpath, xpathPrefix, pathAttr)
					/* skip processing for list key-leaf outside of config container(OC yang) directly under the list.
					   Inside full-spec isKey is set to true for list key-leaf dierctly under the list(outside of config container) 
					   For ietf yang(eg.ietf-ptp) list key-leaf might have a field transformer.
					 */
					if ((!xYangSpecMap[xpath].isKey) || (len(xYangSpecMap[xpath].xfmrField) > 0)) {
						if len(xYangSpecMap[xpath].xfmrFunc) == 0 {
							value := jData.MapIndex(key).Interface()
							log.Infof("data field: key(\"%v\"), value(\"%v\").", key, value)
							retErr = mapFillData(d, ygRoot, oper, uri, requestUri, curKey, result, subOpDataMap, xpathPrefix,
							pathAttr, value, tblXpathMap, txCache, xfmrErr)
							if retErr != nil {
								log.Errorf("Failed constructing data for db write: key(\"%v\"), value(\"%v\"), path(\"%v\").",
								pathAttr, value, xpathPrefix)
								return retErr
							}
						} else {
							log.Infof("write: key(\"%v\"), xpath(\"%v\"), uri(%v).",key, xpath, uri)
							curYgotNode, nodeErr := yangNodeForUriGet(uri, ygRoot)
							if nodeErr != nil {
								curYgotNode = nil
							}
							inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, uri, requestUri, oper, curKey, nil, subOpDataMap, curYgotNode, txCache)
							stRetData, err := xfmrHandler(inParams, xYangSpecMap[xpath].xfmrFunc)
							if err != nil {
								if xfmrErr != nil && *xfmrErr == nil {
                                                                        *xfmrErr = err
                                                                }
								return nil
							}
							if stRetData != nil {
                                                                mapCopy(result, stRetData)
							}
						}
					}
				}
			}
		}
	}

	return retErr
}

/* Debug function to print the map data into file */
func printDbData(resMap map[int]map[db.DBNum]map[string]map[string]db.Value, fileName string) {
	fp, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer fp.Close()
	for oper, dbRes := range resMap {
		fmt.Fprintf(fp, "-----------------------------------------------------------------\r\n")
		fmt.Fprintf(fp, "Oper Type : %v\r\n", oper)
		for d, dbMap := range dbRes {
			fmt.Fprintf(fp, "DB num : %v\r\n", d)
			for k, v := range dbMap {
				fmt.Fprintf(fp, "table name : %v\r\n", k)
				for ik, iv := range v {
					fmt.Fprintf(fp, "  key : %v\r\n", ik)
					for k, d := range iv.Field {
						fmt.Fprintf(fp, "    %v :%v\r\n", k, d)
					}
				}
			}
		}
	}
	fmt.Fprintf(fp, "-----------------------------------------------------------------\r\n")
	return
}
