package transformer

import (
	"fmt"
	//"os"
	//	"sort"
	//	"github.com/openconfig/goyang/pkg/yang"
	"encoding/json"
	"errors"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"strings"
	"translib/db"
	"translib/ocbinds"
)

const (
	GET = 1 + iota
	CREATE
	REPLACE
	UPDATE
	DELETE
)

type KeySpec struct {
	Ts    db.TableSpec
	Key   db.Key
	Child []KeySpec
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
		return
	}
	if len(params) != XlateFuncs[name].Type().NumIn() {
		err = ErrParamsNotAdapted
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = XlateFuncs[name].Call(in)
	return
}

func TraverseDb(d *db.DB, spec KeySpec, result *map[string]map[string]db.Value, parentKey *db.Key) error {
	var err error

	if spec.Key.Len() > 0 {
		// get an entry with a specific key
		data, err := d.GetEntry(&spec.Ts, spec.Key)
		if err != nil {
			return err
		}

		if (*result)[spec.Ts.Name] == nil {
			(*result)[spec.Ts.Name] = map[string]db.Value{strings.Join(spec.Key.Comp, "|"): data}
		} else {
			(*result)[spec.Ts.Name][strings.Join(spec.Key.Comp, "|")] = data
		}

		if len(spec.Child) > 0 {
			for _, ch := range spec.Child {
				err = TraverseDb(d, ch, result, &spec.Key)
			}
		}
	} else {
		// TODO - GetEntry suuport with regex patten, 'abc*' for optimization
		keys, err := d.GetKeys(&spec.Ts)
		if err != nil {
			return err
		}
		for i, _ := range keys {
			if parentKey != nil {
				// TODO - multi-depth with a custom delimiter
				if strings.Index(strings.Join(keys[i].Comp, "|"), strings.Join((*parentKey).Comp, "|")) == -1 {
					continue
				}
			}
			spec.Key = keys[i]
			err = TraverseDb(d, spec, result, parentKey)
		}
	}
	return err
}

func XlateUriToKeySpec(path string, uri *ygot.GoStruct, t *interface{}) (*map[db.DBNum][]KeySpec, error) {

	var err error
	var result = make(map[db.DBNum][]KeySpec)
	var retdbFormat = make([]KeySpec, 0)

	// In case of CVL yang, the tablename and key info is available in the xpath
	if isCvlYang(path) {
		/* Extract the xpath and key from input xpath */
		yangXpath, keyStr, tableName := sonicXpathKeyExtract(path)
		retdbFormat = fillCvlKeySpec(yangXpath, tableName, keyStr)
	} else {
		/* Extract the xpath and key from input xpath */
		yangXpath, keyStr, _ := xpathKeyExtract(nil, uri, 0, path)
		retdbFormat = FillKeySpecs(yangXpath, keyStr, &retdbFormat)
	}
	result[db.ConfigDB] = retdbFormat

	return &result, err
}

func FillKeySpecs(yangXpath string , keyStr string, retdbFormat *[]KeySpec) ([]KeySpec){
    if xSpecMap == nil {
        return *retdbFormat
    }
    _, ok := xSpecMap[yangXpath]
    if ok {
        xpathInfo := xSpecMap[yangXpath]
        if xpathInfo.tableName != nil {
            dbFormat := KeySpec{}
            dbFormat.Ts.Name = *xpathInfo.tableName
	    if keyStr != "" {
		dbFormat.Key.Comp = append(dbFormat.Key.Comp, keyStr)
	    }
            for _, child := range xpathInfo.childTable {
                if xDbSpecMap != nil {
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
            *retdbFormat = append(*retdbFormat, dbFormat)
        } else {
            for _, child := range xpathInfo.childTable {
                if xDbSpecMap != nil {
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
    return *retdbFormat
}

func fillCvlKeySpec(yangXpath string , tableName string, keyStr string) ( []KeySpec ) {

	var retdbFormat = make([]KeySpec, 0)

	if tableName != "" {
		dbFormat := KeySpec{}
		dbFormat.Ts.Name = tableName
		if keyStr != "" {
			dbFormat.Key.Comp = append(dbFormat.Key.Comp, keyStr)
		}
		retdbFormat = append(retdbFormat, dbFormat)
	} else {
		// If table name not available in xpath get top container name
		tokens:= strings.Split(yangXpath, ":")
		container := "/" + tokens[len(tokens)-1]
		if xDbSpecMap[container] != nil {
			dbInfo := xDbSpecMap[container]
			if dbInfo.fieldType == "container" {
				for dir, _ := range dbInfo.dbEntry.Dir {
					dbFormat := KeySpec{}
					dbFormat.Ts.Name = dir
					retdbFormat = append(retdbFormat, dbFormat)
				}
			}
		}
	}
	return retdbFormat
}

func XlateToDb(path string, opcode int, d *db.DB, yg *ygot.GoStruct, yt *interface{}) (map[string]map[string]db.Value, error) {

	var err error

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

	// table.key.fields
	var result = make(map[string]map[string]db.Value)
	switch opcode {
	case CREATE:
		log.Info("CREATE case")
		err = dbMapCreate(d, yg, opcode, path, jsonData, result)
		if err != nil {
			log.Errorf("Error: Data translation from yang to db failed for create request.")
		}

	case UPDATE:
		log.Info("UPDATE case")
		err = dbMapUpdate(d, yg, opcode, path, jsonData, result)
		if err != nil {
			log.Errorf("Error: Data translation from yang to db failed for update request.")
		}

	case REPLACE:
		log.Info("REPLACE case")
		err = dbMapUpdate(d, yg, opcode, path, jsonData, result)
		if err != nil {
			log.Errorf("Error: Data translation from yang to db failed for replace request.")
		}

	case DELETE:
		log.Info("DELETE case")
		err = dbMapDelete(d, yg, opcode, path, jsonData, result)
		if err != nil {
			log.Errorf("Error: Data translation from yang to db failed for delete request.")
		}
	}
	return result, err
}

func XlateFromDb(xpath string, data map[string]map[string]db.Value) ([]byte, error) {
	var err error
	var fieldName, tableName, yangXpath string
	var dbData = make(map[string]map[string]db.Value)

	dbData = data
	if isCvlYang(xpath) {
		yXpath, keyStr, tableName := sonicXpathKeyExtract(xpath)
		yangXpath = yXpath
		if (tableName != "") {
			tokens:= strings.Split(yangXpath, "/")
			// Format /module:container/tableName[key]/fieldName
			if tokens[len(tokens)-2] == tableName {
				fieldName = tokens[len(tokens)-1]
				dbData = extractFieldFromDb(tableName, keyStr, fieldName, data)
			}
		}
	} else {
		yXpath, keyStr, _ := xpathKeyExtract(nil, nil, 0, xpath)
		yangXpath = yXpath
		if xSpecMap == nil {
			return nil, err
		}
		_, ok := xSpecMap[yangXpath]
		if !ok {
			return nil, err
		}
		if xSpecMap[yangXpath].yangDataType == "leaf" {
			fieldName = xSpecMap[yangXpath].fieldName
			tableName = *xSpecMap[yangXpath].tableName
			dbData = extractFieldFromDb(tableName, keyStr, fieldName, data)
		}
	}
	payload, err := dbDataToYangJsonCreate(yangXpath, dbData)

	if err != nil {
		log.Errorf("Error: failed to create json response from DB data.")
		return nil, err
	}

	result := []byte(payload)

	//TODO - implement me
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
        path, err := ygot.StringToPath(uri, ygot.StructuredPath, ygot.StringSlicePath)
        if err != nil {
                log.Error("Error in uri to path conversion: ", err)
                return "", err
        }
        pathSlice := strings.Split(path.Elem[0].Name, ":")
	moduleNm := pathSlice[0]
        log.Info("Module name for given path = ", moduleNm)
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
