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

func dbKeyToYangDataConvert(uri string, xpath string, dbKey string, dbKeySep string) (map[string]interface{}, string, error) {
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

	var kLvlValList []string
	keyDataList := strings.Split(dbKey, dbKeySep)
	keyNameList := yangKeyFromEntryGet(xYangSpecMap[xpath].yangEntry)
	id          := xYangSpecMap[xpath].keyLevel
	uriWithKey  := fmt.Sprintf("%v", xpath)

	/* if uri contins key, use it else use xpath */
	if strings.Contains(uri, "[") {
		uriWithKey  = fmt.Sprintf("%v", uri)
	}
	keyConcat := dbKeySep
	if len(xYangSpecMap[xpath].delim) > 0 {
		keyConcat = xYangSpecMap[xpath].delim
		log.Infof("Found key concatenater(\"%v\") for xpath %v", keyConcat, xpath)
	}

	if len(xYangSpecMap[xpath].xfmrKey) > 0 {
		var dbs [db.MaxDB]*db.DB
		inParams := formXfmrInputRequest(nil, dbs, db.MaxDB, nil, uri, GET, dbKey, nil, nil)
		ret, err := XlateFuncCall(dbToYangXfmrFunc(xYangSpecMap[xpath].xfmrKey), inParams)
		if err != nil {
			return nil, "", err
		}
		rmap := ret[0].Interface().(map[string]interface{})
		for k, v := range rmap {
			uriWithKey += fmt.Sprintf("[%v=%v]", k, v)
		}
		return rmap, uriWithKey, nil
	}

	if len(keyDataList) == 0 || len(keyNameList) == 0 {
		return nil, "", nil
	}

	kLvlValList = append(kLvlValList, keyDataList[id])

	if len(keyNameList) > 1 {
		kLvlValList = strings.Split(keyDataList[id], keyConcat)
	}

	/* TODO: Need to add leaf-ref related code in here and remove this code*/
	kvalExceedFlag := false
	chgId := -1
	if len(keyNameList) < len(kLvlValList) {
		kvalExceedFlag = true
		chgId = len(keyNameList) - 1
	}

	rmap := make(map[string]interface{})
	for i, kname := range keyNameList {
		kval := kLvlValList[i]

		/* TODO: Need to add leaf-ref related code in here and remove this code*/
		if kvalExceedFlag && (i == chgId) {
			kval = strings.Join(kLvlValList[chgId:], keyConcat)
		}

		uriWithKey += fmt.Sprintf("[%v=%v]", kname, kval)
		rmap[kname] = kval
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

func sonicKeyDataAdd(dbIndex db.DBNum, keyNameList []string, keyStr string, resultMap map[string]interface{}) {
	var dbOpts db.Options
	dbOpts = getDBOptions(dbIndex)
	keySeparator := dbOpts.KeySeparator
    keyValList := strings.Split(keyStr, keySeparator)
	
    if len(keyNameList) != len(keyValList) {
        return
    }

    for i, keyName := range keyNameList {
        resultMap[keyName] = keyValList[i]
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

func formXfmrInputRequest(d *db.DB, dbs [db.MaxDB]*db.DB, cdb db.DBNum, ygRoot *ygot.GoStruct, uri string, oper int, key string, dbDataMap *map[db.DBNum]map[string]map[string]db.Value, param interface{}) XfmrParams {
	var inParams XfmrParams
	inParams.d = d
	inParams.dbs = dbs
	inParams.curDb = cdb
	inParams.ygRoot = ygRoot
	inParams.uri = uri
	inParams.oper = oper
	inParams.key = key
	inParams.dbDataMap = dbDataMap
	inParams.param = param // generic param

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
        case db.FlexCounterDB, db.AsicDB, db.LogLevelDB, db.ConfigDB, db.StateDB:
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
        pathList := strings.Split(xpath, "/")
        pathList = pathList[1:]
        for i, pvar := range pathList {
                if strings.Contains(pvar, "[") && strings.Contains(pvar, "]") {
                        si, ei := strings.Index(pvar, "["), strings.Index(pvar, "]")
                        // substring contains [] entries
                        if (si < ei) {
                                pvar = strings.Split(pvar, "[")[0]
                                pathList[i] = pvar

                        } else {
                                // This substring contained a ] before a [.
                                return "", fmt.Errorf("Incorrect ordering of [] within substring %s of %s, [ pos: %d, ] pos: %d", pvar, xpath, si, ei)
                        }
                } else if strings.Contains(pvar, "[") || strings.Contains(pvar, "]") {
                        // This substring contained a mismatched pair of []s.
                        return "", fmt.Errorf("Mismatched brackets within substring %s of %s", pvar, xpath)
                }
                if (i > 0) && strings.Contains(pvar, ":") {
                        pvar = strings.Split(pvar,":")[1]
                        pathList[i] = pvar
                }
        }
        path := "/" + strings.Join(pathList, "/")
        return path,nil
}

