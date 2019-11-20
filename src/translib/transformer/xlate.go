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
	"encoding/json"
	"errors"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"
)

const (
	GET = 1 + iota
	CREATE
	REPLACE
	UPDATE
	DELETE
)

type KeySpec struct {
	dbNum db.DBNum
	Ts    db.TableSpec
	Key   db.Key
	Child []KeySpec
	ignoreParentKey bool
}

var XlateFuncs = make(map[string]reflect.Value)

var (
	ErrParamsNotAdapted = errors.New("The number of params is not adapted.")
)

func XlateFuncBind(name string, fn interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(name + " is not valid Xfmr function.")
		}
	}()

	if _, ok := XlateFuncs[name]; !ok {
		v := reflect.ValueOf(fn)
		v.Type().NumIn()
		XlateFuncs[name] = v
	} else {
		log.Info("Duplicate entry found in the XlateFunc map " + name)
	}
	return
}

func XlateFuncCall(name string, params ...interface{}) (result []reflect.Value, err error) {
	if _, ok := XlateFuncs[name]; !ok {
		err = errors.New(name + " Xfmr function does not exist.")
		return nil, err
	}
	if len(params) != XlateFuncs[name].Type().NumIn() {
		err = ErrParamsNotAdapted
		return nil, nil
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = XlateFuncs[name].Call(in)
	return result, nil
}

func TraverseDb(dbs [db.MaxDB]*db.DB, spec KeySpec, result *map[db.DBNum]map[string]map[string]db.Value, parentKey *db.Key) error {
	var err error
	var dbOpts db.Options

	dbOpts = getDBOptions(spec.dbNum)
	separator := dbOpts.KeySeparator
	//log.Infof("key separator for table %v in Db %v is %v", spec.Ts.Name, spec.dbNum, separator)

	if spec.Key.Len() > 0 {
		// get an entry with a specific key
		data, err := dbs[spec.dbNum].GetEntry(&spec.Ts, spec.Key)
		if err != nil {
			return err
		}

		if (*result)[spec.dbNum][spec.Ts.Name] == nil {
			(*result)[spec.dbNum][spec.Ts.Name] = map[string]db.Value{strings.Join(spec.Key.Comp, separator): data}
		} else {
			(*result)[spec.dbNum][spec.Ts.Name][strings.Join(spec.Key.Comp, separator)] = data
		}

		if len(spec.Child) > 0 {
			for _, ch := range spec.Child {
				err = TraverseDb(dbs, ch, result, &spec.Key)
			}
		}
	} else {
		// TODO - GetEntry support with regex patten, 'abc*' for optimization
		keys, err := dbs[spec.dbNum].GetKeys(&spec.Ts)
		if err != nil {
			return err
		}
		log.Infof("keys for table %v in Db %v are %v", spec.Ts.Name, spec.dbNum, keys)
		for i, _ := range keys {
			if parentKey != nil && (spec.ignoreParentKey == false) {
				// TODO - multi-depth with a custom delimiter
				if strings.Index(strings.Join(keys[i].Comp, separator), strings.Join((*parentKey).Comp, separator)) == -1 {
					continue
				}
			}
			spec.Key = keys[i]
			err = TraverseDb(dbs, spec, result, parentKey)
		}
	}
	return err
}

func XlateUriToKeySpec(uri string, requestUri string, ygRoot *ygot.GoStruct, t *interface{}, txCache interface{}) (*[]KeySpec, error) {

	var err error
	var retdbFormat = make([]KeySpec, 0)

	// In case of SONIC yang, the tablename and key info is available in the xpath
	if isSonicYang(uri) {
		/* Extract the xpath and key from input xpath */
		xpath, keyStr, tableName := sonicXpathKeyExtract(uri)
		retdbFormat = fillSonicKeySpec(xpath, tableName, keyStr)
	} else {
		/* Extract the xpath and key from input xpath */
		xpath, keyStr, _ := xpathKeyExtract(nil, ygRoot, 0, uri, requestUri, nil, txCache)
		retdbFormat = FillKeySpecs(xpath, keyStr, &retdbFormat)
	}

	return &retdbFormat, err
}

func FillKeySpecs(yangXpath string , keyStr string, retdbFormat *[]KeySpec) ([]KeySpec){
	if xYangSpecMap == nil {
		return *retdbFormat
	}
	_, ok := xYangSpecMap[yangXpath]
	if ok {
		xpathInfo := xYangSpecMap[yangXpath]
		if xpathInfo.tableName != nil {
			dbFormat := KeySpec{}
			dbFormat.Ts.Name = *xpathInfo.tableName
			dbFormat.dbNum = xpathInfo.dbIndex
			if len(xYangSpecMap[yangXpath].xfmrKey) > 0 || xYangSpecMap[yangXpath].keyName != nil {
				dbFormat.ignoreParentKey = true
			} else {
				dbFormat.ignoreParentKey = false
			}
			if keyStr != "" {
				dbFormat.Key.Comp = append(dbFormat.Key.Comp, keyStr)
			}
			for _, child := range xpathInfo.childTable {
				if xDbSpecMap != nil {
					if _, ok := xDbSpecMap[child]; ok {
						chlen := len(xDbSpecMap[child].yangXpath)
						if chlen > 0 {
							children := make([]KeySpec, 0)
							for _, childXpath := range xDbSpecMap[child].yangXpath {
								children = FillKeySpecs(childXpath, "", &children)
								dbFormat.Child = append(dbFormat.Child, children...)
							}
						}
					}
				}
			}
			*retdbFormat = append(*retdbFormat, dbFormat)
		} else {
			for _, child := range xpathInfo.childTable {
				if xDbSpecMap != nil {
					if _, ok := xDbSpecMap[child]; ok {
						chlen := len(xDbSpecMap[child].yangXpath)
						if chlen > 0 {
							for _, childXpath := range xDbSpecMap[child].yangXpath {
								*retdbFormat = FillKeySpecs(childXpath, "", retdbFormat)
							}
						}
					}
				}
			}
		}
	}
	return *retdbFormat
}

func fillSonicKeySpec(xpath string , tableName string, keyStr string) ( []KeySpec ) {

	var retdbFormat = make([]KeySpec, 0)

	if tableName != "" {
		dbFormat := KeySpec{}
		dbFormat.Ts.Name = tableName
                cdb := db.ConfigDB
                if _, ok := xDbSpecMap[tableName]; ok {
			cdb = xDbSpecMap[tableName].dbIndex
                }
		dbFormat.dbNum = cdb
		if keyStr != "" {
			dbFormat.Key.Comp = append(dbFormat.Key.Comp, keyStr)
		}
		retdbFormat = append(retdbFormat, dbFormat)
	} else {
		// If table name not available in xpath get top container name
		container := xpath
		if xDbSpecMap != nil {
			if _, ok := xDbSpecMap[container]; ok {
				dbInfo := xDbSpecMap[container]
				if dbInfo.fieldType == "container" {
					for dir, _ := range dbInfo.dbEntry.Dir {
						_, ok := xDbSpecMap[dir]
						if ok && xDbSpecMap[dir].dbEntry.Node.Statement().Keyword == "container" {
						cdb := xDbSpecMap[dir].dbIndex
						dbFormat := KeySpec{}
						dbFormat.Ts.Name = dir
						dbFormat.dbNum = cdb
						retdbFormat = append(retdbFormat, dbFormat)
						}
					}
				}
			}
		}
	}
	return retdbFormat
}

func XlateToDb(path string, opcode int, d *db.DB, yg *ygot.GoStruct, yt *interface{}, txCache interface{}) (map[int]map[db.DBNum]map[string]map[string]db.Value, error) {

	var err error
	requestUri := path

	device := (*yg).(*ocbinds.Device)
	jsonStr, err := ygot.EmitJSON(device, &ygot.EmitJSONConfig{
		Format:         ygot.RFC7951,
		Indent:         "  ",
		SkipValidation: true,
		RFC7951Config: &ygot.RFC7951JSONConfig{
			AppendModuleName: true,
		},
	})

	jsonData := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &jsonData)
	if err != nil {
		log.Errorf("Error: failed to unmarshal json.")
		return nil, err
	}

	// Map contains table.key.fields
	var result = make(map[int]map[db.DBNum]map[string]map[string]db.Value)
	switch opcode {
	case CREATE:
		log.Info("CREATE case")
		err = dbMapCreate(d, yg, opcode, path, requestUri, jsonData, result, txCache)
		if err != nil {
			log.Errorf("Error: Data translation from yang to db failed for create request.")
		}

	case UPDATE:
		log.Info("UPDATE case")
		err = dbMapUpdate(d, yg, opcode, path, requestUri, jsonData, result, txCache)
		if err != nil {
			log.Errorf("Error: Data translation from yang to db failed for update request.")
		}

	case REPLACE:
		log.Info("REPLACE case")
		err = dbMapUpdate(d, yg, opcode, path, requestUri, jsonData, result, txCache)
		if err != nil {
			log.Errorf("Error: Data translation from yang to db failed for replace request.")
		}

	case DELETE:
		log.Info("DELETE case")
		err = dbMapDelete(d, yg, opcode, path, requestUri, jsonData, result, txCache)
		if err != nil {
			log.Errorf("Error: Data translation from yang to db failed for delete request.")
		}
	}
	return result, err
}

func GetAndXlateFromDB(uri string, ygRoot *ygot.GoStruct, dbs [db.MaxDB]*db.DB, txCache interface{}) ([]byte, error) {
	var err error
	var payload []byte
	log.Info("received xpath =", uri)

	requestUri := uri
	keySpec, err := XlateUriToKeySpec(uri, requestUri, ygRoot, nil, txCache)
	var dbresult = make(RedisDbMap)
        for i := db.ApplDB; i < db.MaxDB; i++ {
                dbresult[i] = make(map[string]map[string]db.Value)
	}

	for _, spec := range *keySpec {
		err := TraverseDb(dbs, spec, &dbresult, nil)
		if err != nil {
			log.Error("TraverseDb() failure")
			return payload, err
		}
	}

	payload, err = XlateFromDb(uri, ygRoot, dbs, dbresult, txCache)
	if err != nil {
		log.Error("XlateFromDb() failure.")
		return payload, err
	}

	return payload, err
}

func XlateFromDb(uri string, ygRoot *ygot.GoStruct, dbs [db.MaxDB]*db.DB, data RedisDbMap, txCache interface{}) ([]byte, error) {

	var err error
	var result []byte
	var dbData = make(RedisDbMap)
	var cdb db.DBNum = db.ConfigDB

	dbData = data
	if isSonicYang(uri) {
		xpath, keyStr, tableName := sonicXpathKeyExtract(uri)
		if (tableName != "") {
			dbInfo, ok := xDbSpecMap[tableName]
			if !ok {
				log.Warningf("No entry in xDbSpecMap for xpath %v", tableName)
			} else {
				cdb =  dbInfo.dbIndex
			}
			tokens:= strings.Split(xpath, "/")
			// Format /module:container/tableName/listname[key]/fieldName
			if tokens[SONIC_TABLE_INDEX] == tableName {
		                fieldName := tokens[len(tokens)-1]
				dbSpecField := tableName + "/" + fieldName
				_, ok := xDbSpecMap[dbSpecField]
				if ok && xDbSpecMap[dbSpecField].fieldType == "leaf" {
					dbData[cdb] = extractFieldFromDb(tableName, keyStr, fieldName, data[cdb])
				}
			}
		}
	} else {
	        xpath, _ := XfmrRemoveXPATHPredicates(uri)
		if _, ok := xYangSpecMap[xpath]; ok {
			cdb = xYangSpecMap[xpath].dbIndex
		}
	}
	payload, err := dbDataToYangJsonCreate(uri, ygRoot, dbs, &dbData, cdb, txCache)
	log.Info("Payload generated:", payload)

	if err != nil {
		log.Errorf("Error: failed to create json response from DB data.")
		return nil, err
	}

	result = []byte(payload)
	return result, err

}

func extractFieldFromDb(tableName string, keyStr string, fieldName string, data map[string]map[string]db.Value) (map[string]map[string]db.Value) {

	var dbVal db.Value
	var dbData = make(map[string]map[string]db.Value)

	if tableName != "" && keyStr != "" && fieldName != "" {
		if data[tableName][keyStr].Field != nil {
			dbData[tableName] = make(map[string]db.Value)
			dbVal.Field = make(map[string]string)
			dbVal.Field[fieldName] = data[tableName][keyStr].Field[fieldName]
			dbData[tableName][keyStr] = dbVal
		}
	}
	return dbData
}

func GetModuleNmFromPath(uri string) (string, error) {
	log.Infof("received uri %s to extract module name from ", uri)
	moduleNm, err := uriModuleNameGet(uri)
	return moduleNm, err
}

func GetOrdDBTblList(ygModuleNm string) ([]string, error) {
        var result []string
	var err error
        if dbTblList, ok := xDbSpecOrdTblMap[ygModuleNm]; ok {
                result = dbTblList
		if len(dbTblList) == 0 {
			log.Error("Ordered DB Table list is empty for module name = ", ygModuleNm)
			err = fmt.Errorf("Ordered DB Table list is empty for module name %v", ygModuleNm)

		}
        } else {
                log.Error("No entry found in the map of module names to ordered list of DB Tables for module = ", ygModuleNm)
                err = fmt.Errorf("No entry found in the map of module names to ordered list of DB Tables for module = %v", ygModuleNm)
        }
        return result, err
}

func GetOrdTblList(xfmrTbl string, uriModuleNm string) []string {
        var ordTblList []string
        processedTbl := false
        var sncMdlList []string

	sncMdlList = getYangMdlToSonicMdlList(uriModuleNm)
        for _, sonicMdlNm := range(sncMdlList) {
                sonicMdlTblInfo := xDbSpecTblSeqnMap[sonicMdlNm]
                for _, ordTblNm := range(sonicMdlTblInfo.OrdTbl) {
                                if xfmrTbl == ordTblNm {
                                        log.Infof("Found sonic module(%v) whose ordered table list contains table %v", sonicMdlNm, xfmrTbl)
                                        ordTblList = sonicMdlTblInfo.OrdTbl
                                        processedTbl = true
                                        break
                                }
                }
                if processedTbl {
                        break
                }
        }
        return ordTblList
}

func GetTablesToWatch(xfmrTblList []string, uriModuleNm string) []string {
        var depTblList []string
        depTblMap := make(map[string]bool) //create to avoid duplicates in depTblList, serves as a Set
        processedTbl := false
	var sncMdlList []string
	var lXfmrTblList []string

	sncMdlList = getYangMdlToSonicMdlList(uriModuleNm)

	// remove duplicates from incoming list of tables
	xfmrTblMap := make(map[string]bool) //create to avoid duplicates in xfmrTblList
	for _, xfmrTblNm :=range(xfmrTblList) {
		xfmrTblMap[xfmrTblNm] = true
	}
	for xfmrTblNm, _ := range(xfmrTblMap) {
		lXfmrTblList = append(lXfmrTblList, xfmrTblNm)
	}

        for _, xfmrTbl := range(lXfmrTblList) {
		processedTbl = false
                //can be optimized if there is a way to know all sonic modules, a given OC-Yang spans over
                for _, sonicMdlNm := range(sncMdlList) {
                        sonicMdlTblInfo := xDbSpecTblSeqnMap[sonicMdlNm]
                        for _, ordTblNm := range(sonicMdlTblInfo.OrdTbl) {
                                if xfmrTbl == ordTblNm {
                                        log.Infof("Found sonic module(%v) whose ordered table list contains table %v", sonicMdlNm, xfmrTbl)
                                        ldepTblList := sonicMdlTblInfo.DepTbl[xfmrTbl]
                                        for _, depTblNm := range(ldepTblList) {
                                                depTblMap[depTblNm] = true
                                        }
                                        //assumption that a table belongs to only one sonic module
                                        processedTbl = true
                                        break
                                }
                        }
                        if processedTbl {
                                break
                        }
                }
        }
        for depTbl := range(depTblMap) {
                depTblList = append(depTblList, depTbl)
        }
	return depTblList
}

func CallRpcMethod(path string, body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
	var err error
	var ret []byte

	// TODO - check module name
	rpcName := strings.Split(path, ":")
	if dbXpathData, ok := xDbSpecMap[rpcName[1]]; ok {
		log.Info("RPC callback invoked (%v) \r\n", rpcName)
		data, err := XlateFuncCall(dbXpathData.rpcFunc, body, dbs)
		if err != nil {
			return nil, err
		}
		ret = data[0].Interface().([]byte)
	} else {
		log.Error("No tsupported RPC", path)
		err = tlerr.NotSupported("Not supported RPC")
	}

	return ret, err
}

