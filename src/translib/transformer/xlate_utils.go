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
    "strings"
    "reflect"
	"regexp"
    "translib/db"
    "github.com/openconfig/goyang/pkg/yang"
    "github.com/openconfig/gnmi/proto/gnmi"
    "github.com/openconfig/ygot/ygot"
    log "github.com/golang/glog"
)

/* Create db key from data xpath(request) */
func keyCreate(keyPrefix string, xpath string, data interface{}, dbKeySep string) string {
	_, ok := xYangSpecMap[xpath]
	if ok {
		if xYangSpecMap[xpath].yangEntry != nil {
			yangEntry := xYangSpecMap[xpath].yangEntry
			delim := dbKeySep
			if len(xYangSpecMap[xpath].delim) > 0 {
				delim = xYangSpecMap[xpath].delim
				log.Infof("key concatenater(\"%v\") found for xpath %v ", delim, xpath)
			}

			if len(keyPrefix) > 0 { keyPrefix += delim }
			keyVal := ""
			for i, k := range (strings.Split(yangEntry.Key, " ")) {
				if i > 0 { keyVal = keyVal + delim }
				val := fmt.Sprint(data.(map[string]interface{})[k])
				if strings.Contains(val, ":") {
					val = strings.Split(val, ":")[1]
				}
				keyVal += val
			}
			keyPrefix += string(keyVal)
		}
	}
	return keyPrefix
}

/* Copy redis-db source to destn map */
func mapCopy(destnMap map[string]map[string]db.Value, srcMap map[string]map[string]db.Value) {
   for table, tableData := range srcMap {
        _, ok := destnMap[table]
        if !ok {
            destnMap[table] = make(map[string]db.Value)
        }
        for rule, ruleData := range tableData {
            _, ok = destnMap[table][rule]
            if !ok {
                 destnMap[table][rule] = db.Value{Field: make(map[string]string)}
            }
            for field, value := range ruleData.Field {
                destnMap[table][rule].Field[field] = value
            }
        }
   }
}

func parentXpathGet(xpath string) string {
    path := ""
    if len(xpath) > 0 {
		p := strings.Split(xpath, "/")
		path = strings.Join(p[:len(p)-1], "/")
	}
    return path
}

func isYangResType(ytype string) bool {
    if ytype == "choose" || ytype == "case" {
        return true
    }
    return false
}

func yangTypeGet(entry *yang.Entry) string {
    if entry != nil && entry.Node != nil {
        return entry.Node.Statement().Keyword
    }
    return ""
}

func dbKeyToYangDataConvert(uri string, requestUri string, xpath string, dbKey string, dbKeySep string, txCache interface{}) (map[string]interface{}, string, error) {
	var err error
	if len(uri) == 0 && len(xpath) == 0 && len(dbKey) == 0 {
		err = fmt.Errorf("Insufficient input")
		return nil, "", err
	}

	if _, ok := xYangSpecMap[xpath]; ok {
		if xYangSpecMap[xpath].yangEntry == nil {
			err = fmt.Errorf("Yang Entry not available for xpath ", xpath)
			return nil, "", nil
		}
	}

	keyNameList := yangKeyFromEntryGet(xYangSpecMap[xpath].yangEntry)
	id          := xYangSpecMap[xpath].keyLevel
	keyDataList := strings.SplitN(dbKey, dbKeySep, id)
	uriWithKey  := fmt.Sprintf("%v", xpath)
	uriWithKeyCreate := true
	if len(keyDataList) == 0 {
		keyDataList = append(keyDataList, dbKey)
	}

	/* if uri contins key, use it else use xpath */
	if strings.Contains(uri, "[") {
		uriXpath, _ := XfmrRemoveXPATHPredicates(uri)
		if (uriXpath == xpath  && (strings.HasSuffix(uri, "]") || strings.HasSuffix(uri, "]/"))) {
                        uriWithKeyCreate = false
                }
		uriWithKey  = fmt.Sprintf("%v", uri)
	}

	if len(xYangSpecMap[xpath].xfmrKey) > 0 {
		var dbs [db.MaxDB]*db.DB
		inParams := formXfmrInputRequest(nil, dbs, db.MaxDB, nil, uri, requestUri, GET, dbKey, nil, nil, nil, txCache)
		ret, err := XlateFuncCall(dbToYangXfmrFunc(xYangSpecMap[xpath].xfmrKey), inParams)
		if err != nil {
			return nil, "", err
		}
		rmap := ret[0].Interface().(map[string]interface{})
		if uriWithKeyCreate {
			for k, v := range rmap {
				uriWithKey += fmt.Sprintf("[%v=%v]", k, v)
			}
		}
		return rmap, uriWithKey, nil
	}

	if len(keyNameList) == 0 {
		return nil, "", nil
	}


	rmap := make(map[string]interface{})
	if len(keyNameList) > 1 {
		log.Infof("No key transformer found for multi element yang key mapping to a single redis key string.")
	        return rmap, uriWithKey, nil
	}
	keyXpath := xpath + "/" + keyNameList[0]
	xyangSpecInfo, ok := xYangSpecMap[keyXpath]
	if !ok  || xyangSpecInfo == nil {
		errStr := fmt.Sprintf("Failed to find key xpath %v in xYangSpecMap or is nil, needed to fetch the yangEntry data-type", keyXpath)
		err = fmt.Errorf("%v", errStr)
		return rmap, uriWithKey, err
	}
	yngTerminalNdDtType := xyangSpecInfo.yangEntry.Type.Kind
	resVal, _, err := DbToYangType(yngTerminalNdDtType, keyXpath, keyDataList[0])
	if err != nil {
		errStr := fmt.Sprintf("Failure in converting Db value type to yang type for field", keyXpath)
		err = fmt.Errorf("%v", errStr)
		return rmap, uriWithKey, err
	} else {
		 rmap[keyNameList[0]] = resVal
	}
	if uriWithKeyCreate {
		uriWithKey += fmt.Sprintf("[%v=%v]", keyNameList[0], resVal)
	}

	return rmap, uriWithKey, nil
}

func contains(sl []string, str string) bool {
    for _, v := range sl {
        if v == str {
            return true
        }
    }
    return false
}

func isSubtreeRequest(targetUriPath string, nodePath string) bool {
    if len(targetUriPath) > 0 && len(nodePath) > 0 {
        return strings.HasPrefix(targetUriPath, nodePath)
    }
    return false
}

func getYangPathFromUri(uri string) (string, error) {
    var path *gnmi.Path
    var err error

    path, err = ygot.StringToPath(uri, ygot.StructuredPath, ygot.StringSlicePath)
    if err != nil {
        log.Errorf("Error in uri to path conversion: %v", err)
        return "", err
    }

    yangPath, yperr := ygot.PathToSchemaPath(path)
    if yperr != nil {
        log.Errorf("Error in Gnmi path to Yang path conversion: %v", yperr)
        return "", yperr
    }

    return yangPath, err
}

func yangKeyFromEntryGet(entry *yang.Entry) []string {
    var keyList []string
    for _, key := range strings.Split(entry.Key, " ") {
        keyList = append(keyList, key)
    }
    return keyList
}

func isSonicYang(path string) bool {
    if strings.HasPrefix(path, "/sonic") {
        return true
    }
    return false
}

func sonicKeyDataAdd(dbIndex db.DBNum, keyNameList []string, xpathPrefix string, keyStr string, resultMap map[string]interface{}) {
	var dbOpts db.Options
	dbOpts = getDBOptions(dbIndex)
	keySeparator := dbOpts.KeySeparator
    keyValList := strings.SplitN(keyStr, keySeparator, len(keyNameList))
    log.Infof("yang keys list - %v, xpathprefix - %v, DB-key string - %v, DB-key list after db key separator split - %v, dbIndex - %v", keyNameList, xpathPrefix, keyStr, keyValList, dbIndex) 

    if len(keyNameList) != len(keyValList) {
        return
    }

    for i, keyName := range keyNameList {
	    keyXpath := xpathPrefix + "/" + keyName
	    dbInfo, ok := xDbSpecMap[keyXpath]
	    var resVal interface{}
	    resVal = keyValList[i]
	    if !ok || dbInfo == nil {
		    log.Warningf("xDbSpecMap entry not found or is nil for xpath %v, hence data-type conversion cannot happen", keyXpath)
	    } else {
		    yngTerminalNdDtType := dbInfo.dbEntry.Type.Kind
		    var err error
		    resVal, _, err = DbToYangType(yngTerminalNdDtType, keyXpath, keyValList[i])
		    if err != nil {
			    log.Warningf("Data-type conversion unsuccessfull for xpath %v", keyXpath)
			    resVal = keyValList[i]
		    }
	    }

        resultMap[keyName] = resVal
    }
}

func yangToDbXfmrFunc(funcName string) string {
    return ("YangToDb_" + funcName)
}

func uriWithKeyCreate (uri string, xpathTmplt string, data interface{}) (string, error) {
    var err error
    if _, ok := xYangSpecMap[xpathTmplt]; ok {
         yangEntry := xYangSpecMap[xpathTmplt].yangEntry
         if yangEntry != nil {
              for _, k := range (strings.Split(yangEntry.Key, " ")) {
                  uri += fmt.Sprintf("[%v=%v]", k, data.(map[string]interface{})[k])
              }
	 } else {
            err = fmt.Errorf("Yang Entry not available for xpath ", xpathTmplt)
	 }
    } else {
        err = fmt.Errorf("No entry in xYangSpecMap for xpath ", xpathTmplt)
    }
    return uri, err
}

func xpathRootNameGet(path string) string {
    if len(path) > 0 {
        pathl := strings.Split(path, "/")
        return ("/" + pathl[1])
    }
    return ""
}

func getDbNum(xpath string ) db.DBNum {
    _, ok := xYangSpecMap[xpath]
    if ok {
        xpathInfo := xYangSpecMap[xpath]
        return xpathInfo.dbIndex
    }
    // Default is ConfigDB
    return db.ConfigDB
}

func dbToYangXfmrFunc(funcName string) string {
    return ("DbToYang_" + funcName)
}

func uriModuleNameGet(uri string) (string, error) {
	var err error
	result := ""
	if len(uri) == 0 {
		log.Error("Empty uri string supplied")
                err = fmt.Errorf("Empty uri string supplied")
		return result, err
	}
	urislice := strings.Split(uri, ":")
	if len(urislice) == 1 {
		log.Errorf("uri string %s does not have module name", uri)
		err = fmt.Errorf("uri string does not have module name: ", uri)
		return result, err
	}
	moduleNm := strings.Split(urislice[0], "/")
	result = moduleNm[1]
	if len(strings.Trim(result, " ")) == 0 {
		log.Error("Empty module name")
		err = fmt.Errorf("No module name found in uri %s", uri)
        }
	log.Info("module name = ", result)
	return result, err
}

func recMap(rMap interface{}, name []string, id int, max int) {
    if id == max || id < 0 {
        return
    }
    val := name[id]
       if reflect.ValueOf(rMap).Kind() == reflect.Map {
               data := reflect.ValueOf(rMap)
               dMap := data.Interface().(map[string]interface{})
               _, ok := dMap[val]
               if ok {
                       recMap(dMap[val], name, id+1, max)
               } else {
                       dMap[val] = make(map[string]interface{})
                       recMap(dMap[val], name, id+1, max)
               }
       }
       return
}

func mapCreate(xpath string) map[string]interface{} {
    retMap   := make(map[string]interface{})
    if  len(xpath) > 0 {
        attrList := strings.Split(xpath, "/")
        alLen    := len(attrList)
        recMap(retMap, attrList, 1, alLen)
    }
    return retMap
}

func mapInstGet(name []string, id int, max int, inMap interface{}) map[string]interface{} {
    if inMap == nil {
        return nil
    }
    result := reflect.ValueOf(inMap).Interface().(map[string]interface{})
    if id == max {
        return result
    }
    val := name[id]
    if reflect.ValueOf(inMap).Kind() == reflect.Map {
        data := reflect.ValueOf(inMap)
        dMap := data.Interface().(map[string]interface{})
        _, ok := dMap[val]
        if ok {
            result = mapInstGet(name, id+1, max, dMap[val])
        } else {
            return result
        }
    }
    return result
}

func mapGet(xpath string, inMap map[string]interface{}) map[string]interface{} {
    attrList := strings.Split(xpath, "/")
    alLen    := len(attrList)
    recMap(inMap, attrList, 1, alLen)
    retMap := mapInstGet(attrList, 1, alLen, inMap)
    return retMap
}

func formXfmrInputRequest(d *db.DB, dbs [db.MaxDB]*db.DB, cdb db.DBNum, ygRoot *ygot.GoStruct, uri string, requestUri string, oper int, key string, dbDataMap *RedisDbMap, subOpDataMap map[int]*RedisDbMap, param interface{}, txCache interface{}) XfmrParams {
	var inParams XfmrParams
	inParams.d = d
	inParams.dbs = dbs
	inParams.curDb = cdb
	inParams.ygRoot = ygRoot
	inParams.uri = uri
	inParams.requestUri = requestUri
	inParams.oper = oper
	inParams.key = key
	inParams.dbDataMap = dbDataMap
	inParams.subOpDataMap = subOpDataMap
	inParams.param = param // generic param
	inParams.txCache = txCache

	return inParams
}

func findByValue(m map[string]string, value string) string {
	for key, val := range m {
		if val == value {
			return key
		}
	}
	return ""
}
func findByKey(m map[string]string, key string) string {
	if val, found := m[key]; found {
		return val
	}
	return ""
}
func findInMap(m map[string]string, str string) string {
	// Check if str exists as a value in map m.
	if val := findByKey(m, str); val != "" {
		return val
	}

	// Check if str exists as a key in map m.
	if val := findByValue(m, str); val != "" {
		return val
	}

	// str doesn't exist in map m.
	return ""
}

func getDBOptions(dbNo db.DBNum) db.Options {
        var opt db.Options

        switch dbNo {
        case db.ApplDB, db.CountersDB:
                opt = getDBOptionsWithSeparator(dbNo, "", ":", ":")
                break
        case db.FlexCounterDB, db.AsicDB, db.LogLevelDB, db.ConfigDB, db.StateDB, db.ErrorDB:
                opt = getDBOptionsWithSeparator(dbNo, "", "|", "|")
                break
        }

        return opt
}

func getDBOptionsWithSeparator(dbNo db.DBNum, initIndicator string, tableSeparator string, keySeparator string) db.Options {
        return(db.Options {
                    DBNo              : dbNo,
                    InitIndicator     : initIndicator,
                    TableNameSeparator: tableSeparator,
                    KeySeparator      : keySeparator,
                      })
}

func stripAugmentedModuleNames(xpath string) string {
        pathList := strings.Split(xpath, "/")
        pathList = pathList[1:]
        for i, pvar := range pathList {
                if (i > 0) && strings.Contains(pvar, ":") {
                        pvar = strings.Split(pvar,":")[1]
                        pathList[i] = pvar
                }
        }
        path := "/" + strings.Join(pathList, "/")
        return path
}

func XfmrRemoveXPATHPredicates(xpath string) (string, error) {
	// Strip keys from xpath
	for {
		si, ei := strings.IndexAny(xpath, "["), strings.Index(xpath, "]")
		if si != -1 && ei != -1 {
			if si < ei {
				newpath := xpath[:si] + xpath[ei+1:]
				xpath = newpath
			} else {
				return "", fmt.Errorf("Incorrect ordering of [] in %s , [ pos: %d, ] pos: %d", xpath, si, ei)
			}
		} else if si != -1 || ei != -1 {
			return "", fmt.Errorf("Mismatched brackets within string %s, si:%d ei:%d", xpath, si, ei)
		} else {
			// No more keys available
			break
		}
	}
	path := stripAugmentedModuleNames(xpath)
	return path, nil
}

func replacePrefixWithModuleName(xpath string) (string) {
	//Input xpath is after removing the xpath Predicates
	var moduleNm string
	if _, ok := xYangSpecMap[xpath]; ok {
		moduleNm = xYangSpecMap[xpath].dbEntry.Prefix.Parent.NName()
		pathList := strings.Split(xpath, ":")
		if len(moduleNm) > 0 && len(pathList) == 2 {
			xpath = "/" + moduleNm + ":" + pathList[1]
		}
	}
	return xpath
}


/* Extract key vars, create db key and xpath */
func xpathKeyExtract(d *db.DB, ygRoot *ygot.GoStruct, oper int, path string, requestUri string, subOpDataMap map[int]*RedisDbMap, txCache interface{}) (string, string, string) {
	 keyStr    := ""
	 tableName := ""
	 pfxPath := ""
	 rgp       := regexp.MustCompile(`\[([^\[\]]*)\]`)
	 curPathWithKey := ""
	 cdb := db.ConfigDB
	 var dbs [db.MaxDB]*db.DB

	 pfxPath, _ = XfmrRemoveXPATHPredicates(path)
	 xpathInfo, ok := xYangSpecMap[pfxPath]
	 if !ok {
		 log.Errorf("No entry found in xYangSpecMap for xpath %v.", pfxPath)
		 return pfxPath, keyStr, tableName
	 }
	 cdb = xpathInfo.dbIndex
	 dbOpts := getDBOptions(cdb)
	 keySeparator := dbOpts.KeySeparator
	 if len(xpathInfo.delim) > 0 {
		 keySeparator = xpathInfo.delim
	 }

	 for _, k := range strings.Split(path, "/") {
		 curPathWithKey += k
		 yangXpath, _ := XfmrRemoveXPATHPredicates(curPathWithKey)
		 xpathInfo, ok := xYangSpecMap[yangXpath]
		 if ok {
			 yangType := yangTypeGet(xpathInfo.yangEntry)
			 /* when deleting a specific element from leaf-list query uri is of the form
			    /prefix-path/leafList-field-name[leafList-field-name=value].
			    Here the syntax is like a list-key instance enclosed in square 
			    brackets .So avoid list key instance like processing for such a case
			 */
			 if yangType == YANG_LEAF_LIST {
				 break
			 }
			 if strings.Contains(k, "[") {
				 if len(keyStr) > 0 {
					 keyStr += keySeparator
				 }
				 if len(xYangSpecMap[yangXpath].xfmrKey) > 0 {
					 xfmrFuncName := yangToDbXfmrFunc(xYangSpecMap[yangXpath].xfmrKey)
					 inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, curPathWithKey, requestUri, oper, "", nil, subOpDataMap, nil, txCache)
					 ret, err := XlateFuncCall(xfmrFuncName, inParams)
					 if err != nil {
						 return "", "", ""
					 }
					 if ret != nil {
						 keyStr = ret[0].Interface().(string)
					 }
				 } else if xYangSpecMap[yangXpath].keyName != nil {
					 keyStr += *xYangSpecMap[yangXpath].keyName
				 } else {
					 /* multi-leaf yang key together forms a single key-string in redis.
					 There should be key-transformer, if not then the yang key leaves
					 will be concatenated with respective default DB type key-delimiter
					 */
					 for idx, kname := range rgp.FindAllString(k, -1) {
						 if idx > 0 { keyStr += keySeparator }
						 keyl := strings.TrimRight(strings.TrimLeft(kname, "["), "]")
						 keys := strings.Split(keyl, "=")
						 keyStr += keys[1]
					 }
				 }
			 } else if len(xYangSpecMap[yangXpath].xfmrKey) > 0 {
				 xfmrFuncName := yangToDbXfmrFunc(xYangSpecMap[yangXpath].xfmrKey)
				 inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, curPathWithKey, requestUri, oper, "", nil, subOpDataMap, nil, txCache)
				 ret, err := XlateFuncCall(xfmrFuncName, inParams)
				 if err != nil {
					 return "", "", ""
				 }
				 if ret != nil {
					 keyStr = ret[0].Interface().(string)
				 }
			 } else if xYangSpecMap[yangXpath].keyName != nil {
				 keyStr += *xYangSpecMap[yangXpath].keyName
			 }
		 }
		 curPathWithKey += "/"
	 }
	 curPathWithKey = strings.TrimSuffix(curPathWithKey, "/")
	 tblPtr     := xpathInfo.tableName
	 if tblPtr != nil && *tblPtr != XFMR_NONE_STRING {
		 tableName = *tblPtr
	 } else if xpathInfo.xfmrTbl != nil {
		 inParams := formXfmrInputRequest(d, dbs, cdb, ygRoot, curPathWithKey, requestUri, oper, "", nil, subOpDataMap, nil, txCache)
		 tableName, _ = tblNameFromTblXfmrGet(*xpathInfo.xfmrTbl, inParams)
	 }
	 return pfxPath, keyStr, tableName
 }

 func sonicXpathKeyExtract(path string) (string, string, string) {
	 xpath, keyStr, tableName, fldNm := "", "", "", ""
	 var err error
	 lpath := path
	 xpath, err = XfmrRemoveXPATHPredicates(path)
	 if err != nil {
		 return xpath, keyStr, tableName
	 }
	 if xpath != "" {
		 fldPth := strings.Split(xpath, "/")
		 if len(fldPth) > SONIC_FIELD_INDEX {
			 fldNm = fldPth[SONIC_FIELD_INDEX]
			 log.Info("Field Name : ", fldNm)
		 }
	 }
	 rgp := regexp.MustCompile(`\[([^\[\]]*)\]`)
	 pathsubStr := strings.Split(path , "/")
	 if len(pathsubStr) > SONIC_TABLE_INDEX  {
		 if strings.Contains(pathsubStr[2], "[") {
			 tableName = strings.Split(pathsubStr[SONIC_TABLE_INDEX], "[")[0]
		 } else {
			 tableName = pathsubStr[SONIC_TABLE_INDEX]
		 }
		 dbInfo, ok := xDbSpecMap[tableName]
		 cdb := db.ConfigDB
		 if !ok {
			 log.Infof("No entry in xDbSpecMap for xpath %v in order to fetch DB index.", tableName)
			 return xpath, keyStr, tableName
		 }
		 cdb = dbInfo.dbIndex
		 dbOpts := getDBOptions(cdb)
		 if dbInfo.keyName != nil {
			 keyStr = *dbInfo.keyName
		 } else {
			 /* chomp off the field portion to avoid processing specific item delete in leaf-list
			    eg. /sonic-acl:sonic-acl/ACL_TABLE/ACL_TABLE_LIST[aclname=MyACL2_ACL_IPV4]/ports[ports=Ethernet12]
			 */
			 if fldNm != "" {
				 chompFld := strings.Split(path, "/")
				 lpath = strings.Join(chompFld[:SONIC_FIELD_INDEX], "/")
				 log.Info("path after removing the field portion ", lpath)

			 }
			 for i, kname := range rgp.FindAllString(lpath, -1) {
				 if i > 0 {
					 keyStr += dbOpts.KeySeparator
				 }
				 val := strings.Split(kname, "=")[1]
				 keyStr += strings.TrimRight(val, "]")
			 }
		 }
	 }
	 return xpath, keyStr, tableName
 }

func getYangMdlToSonicMdlList(moduleNm string) []string {
	var sncMdlList []string
        if xDbSpecTblSeqnMap == nil || len(xDbSpecTblSeqnMap) == 0 {
                log.Info("xDbSpecTblSeqnMap is empty.")
                return sncMdlList
        }
        if strings.HasPrefix(moduleNm, SONIC_MDL_PFX) {
                sncMdlList = append(sncMdlList, moduleNm)
        } else {
                //can be optimized if there is a way to know sonic modules, a given OC-Yang spans over
                for sncMdl := range(xDbSpecTblSeqnMap) {
                        sncMdlList = append(sncMdlList, sncMdl)
                }
        }
	return sncMdlList
}

func yangFloatIntToGoType(t yang.TypeKind, v float64) (interface{}, error) {
        switch t {
        case yang.Yint8:
                return int8(v), nil
        case yang.Yint16:
                return int16(v), nil
        case yang.Yint32:
                return int32(v), nil
        case yang.Yuint8:
                return uint8(v), nil
        case yang.Yuint16:
                return uint16(v), nil
        case yang.Yuint32:
                return uint32(v), nil
        }
        return nil, fmt.Errorf("unexpected YANG type %v", t)
}

func unmarshalJsonToDbData(schema *yang.Entry, fieldName string, value interface{}) (string, error) {
        var data string
        ykind := schema.Type.Kind

        switch ykind {
        case yang.Ystring, yang.Ydecimal64, yang.Yint64, yang.Yuint64,
             yang.Yenum, yang.Ybool, yang.Ybinary, yang.Yidentityref, yang.Yunion:
                data = fmt.Sprintf("%v", value)

        case yang.Yint8, yang.Yint16, yang.Yint32,
             yang.Yuint8, yang.Yuint16, yang.Yuint32:
                pv, err := yangFloatIntToGoType(ykind, value.(float64))
                if err != nil {
                        return "", fmt.Errorf("error parsing %v for schema %s: %v", value, schema.Name, err)
                }
                data = fmt.Sprintf("%v", pv)
        default:
                // TODO - bitset, empty
                data = fmt.Sprintf("%v", value)
        }

        return data, nil
}
