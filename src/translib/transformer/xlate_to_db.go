package transformer

import (
	"errors"
	"fmt"
	"github.com/openconfig/ygot/ygot"
	"os"
	"reflect"
	"regexp"
	"strings"
	"translib/db"

	log "github.com/golang/glog"
)

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

	result[tableName][dbKey].Field[field] = value
	return
}

/* Fill the redis-db map with data */
func mapFillData(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, dbKey string, result map[string]map[string]db.Value, xpathPrefix string, name string, value interface{}) error {
	xpath := xpathPrefix + "/" + name
	xpathInfo := xSpecMap[xpath]
	log.Info("name: \"%v\", xpathPrefix(\"%v\").", name, xpathPrefix)

	if xpathInfo == nil {
		log.Errorf("Yang path(\"%v\") not found.", xpath)
		return errors.New("Invalid URI")
	}

	if xpathInfo.tableName == nil {
		log.Errorf("Table for yang-path(\"%v\") not found.", xpath)
		return errors.New("Invalid table name")
	}

	if len(dbKey) == 0 {
		log.Errorf("Table key for yang path(\"%v\") not found.", xpath)
		return errors.New("Invalid table key")
	}

	if len(xpathInfo.xfmrFunc) > 0 {
		/* field transformer present */
		log.Info("Transformer function(\"%v\") invoked for yang path(\"%v\").", xpathInfo.xfmrFunc, xpath)
		ret, err := XlateFuncCall(yangToDbXfmrFunc(xSpecMap[xpath].xfmrFunc), d, ygRoot, oper, uri, value)
		if err != nil {
			return err
		}
		retData := ret[0].Interface().(map[string]string)
		log.Info("Transformer function \"%v\" for \"%v\" returned(%v).", xpathInfo.xfmrFunc, xpath, retData)
		for f, v := range retData {
			dataToDBMapAdd(*xpathInfo.tableName, dbKey, result, f, v)
		}
		return nil
	}

	if len(xpathInfo.fieldName) == 0 {
		log.Info("Field for yang-path(\"%v\") not found in DB.", xpath)
		return errors.New("Invalid field name")
	}
	fieldName := xpathInfo.fieldName
    valueStr  := fmt.Sprintf("%v", value)
	if strings.Contains(valueStr, ":") {
		valueStr = strings.Split(valueStr, ":")[1]
	}

	dataToDBMapAdd(*xpathInfo.tableName, dbKey, result, fieldName, valueStr)
	log.Info("TblName: \"%v\", key: \"%v\", field: \"%v\", valueStr: \"%v\".",
		*xpathInfo.tableName, dbKey, fieldName, valueStr)
	return nil
}

func cvlYangReqToDbMapCreate(jsonData interface{}, result map[string]map[string]db.Value) error {
	if reflect.ValueOf(jsonData).Kind() == reflect.Map {
		data := reflect.ValueOf(jsonData)
		for _, key := range data.MapKeys() {
			_, ok := xDbSpecMap[key.String()]
			if ok {
				directDbMapData(key.String(), data.MapIndex(key).Interface(), result)
			} else {
				cvlYangReqToDbMapCreate(data.MapIndex(key).Interface(), result)
			}
		}
	}
	return nil
}

func directDbMapData(tableName string, jsonData interface{}, result map[string]map[string]db.Value) bool {
	_, ok := xDbSpecMap[tableName]

	if ok && xDbSpecMap[tableName].dbEntry != nil {
		dbSpecData := xDbSpecMap[tableName].dbEntry
		tblKeyName := strings.Split(dbSpecData.Key, " ")
		data := reflect.ValueOf(jsonData)
		result[tableName] = make(map[string]db.Value)

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

			result[tableName][keyName] = db.Value{Field: make(map[string]string)}
			for field, value := range d {
				result[tableName][keyName].Field[field] = fmt.Sprintf("%v", value)
			}
		}
		return true
	}
	return false
}

/* Get the db table, key and field name for the incoming delete request */
func dbMapDelete(d *db.DB, ygRoot *ygot.GoStruct, oper int, path string, jsonData interface{}, result map[string]map[string]db.Value) error {
	var err error
	if isCvlYang(path) {
		xpathPrefix, keyName, tableName := sonicXpathKeyExtract(path)
		log.Info("Delete req: path(\"%v\"), key(\"%v\"), xpathPrefix(\"%v\"), tableName(\"%v\").", path, keyName, xpathPrefix, tableName)
		err = cvlYangReqToDbMapDelete(xpathPrefix, tableName, keyName, result)
	} else {
		xpathPrefix, keyName, tableName := xpathKeyExtract(path)
		log.Info("Delete req: path(\"%v\"), key(\"%v\"), xpathPrefix(\"%v\"), tableName(\"%v\").", path, keyName, xpathPrefix, tableName)
		spec, ok := xSpecMap[xpathPrefix]
		if ok && spec.tableName != nil {
			result[*spec.tableName] = make(map[string]db.Value)
			if len(keyName) > 0 {
				result[*spec.tableName][keyName] = db.Value{Field: make(map[string]string)}
				if spec.yangEntry != nil && spec.yangEntry.Node.Statement().Keyword == "leaf" {
					result[*spec.tableName][keyName].Field[spec.fieldName] = ""
				}
			}
		}
	}
	log.Info("Delete req: path(\"%v\") result(\"%v\").", path, result)
	return err
}

func cvlYangReqToDbMapDelete(xpathPrefix string, tableName string, keyName string, result map[string]map[string]db.Value) error {
	if (tableName != "") {
		// Specific table entry case
		result[tableName] = make(map[string]db.Value)
		if (keyName != "") {
			// Specific key case
			var dbVal db.Value
			tokens:= strings.Split(xpathPrefix, "/")
			// Format /module:container/tableName[key]/fieldName
			if tokens[len(tokens)-2] == tableName {
				// Specific leaf case
				fieldName := tokens[len(tokens)-1]
				dbVal.Field = make(map[string]string)
				dbVal.Field[fieldName] = ""
			}
			result[tableName][keyName] = dbVal
		} else {
			// Get all keys
			fmt.Println("No Key. Return table name")
		}
	} else {
		// Get all table entries
		fmt.Println("No table name. Delete all entries")
		// If table name not available in xpath get top container name
		tokens:= strings.Split(xpathPrefix, ":")
		container := "/" + tokens[len(tokens)-1]
		if xDbSpecMap[container] != nil {
			dbInfo := xDbSpecMap[container]
			if dbInfo.fieldType == "container" {
				for dir, _ := range dbInfo.dbEntry.Dir {
					result[dir] = make(map[string]db.Value)
				}
                       }
		}
	}
	return nil
}

/* Get the data from incoming update/replace request, create map and fill with dbValue(ie. field:value
   to write into redis-db */
func dbMapUpdate(d *db.DB, ygRoot *ygot.GoStruct, oper int, path string, jsonData interface{}, result map[string]map[string]db.Value) error {
	log.Info("Update/replace req: path(\"%v\").", path)
	dbMapCreate(d, ygRoot, oper, path, jsonData, result)
	log.Info("Update/replace req: path(\"%v\") result(\"%v\").", path, result)
	printDbData(result, "/tmp/yangToDbDataUpRe.txt")
	return nil
}

/* Get the data from incoming create request, create map and fill with dbValue(ie. field:value
   to write into redis-db */
func dbMapCreate(d *db.DB, ygRoot *ygot.GoStruct, oper int, path string, jsonData interface{}, result map[string]map[string]db.Value) error {
	root := xpathRootNameGet(path)
	if isCvlYang(path) {
		cvlYangReqToDbMapCreate(jsonData, result)
	} else {
		yangReqToDbMapCreate(d, ygRoot, oper, root, "", "", jsonData, result)
	}
	printDbData(result, "/tmp/yangToDbDataCreate.txt")
	return nil
}

func yangReqToDbMapCreate(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, xpathPrefix string, keyName string, jsonData interface{}, result map[string]map[string]db.Value) error {
	log.Info("key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)

	if reflect.ValueOf(jsonData).Kind() == reflect.Slice {
		log.Info("slice data: key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)
		jData := reflect.ValueOf(jsonData)
		dataMap := make([]interface{}, jData.Len())
		for idx := 0; idx < jData.Len(); idx++ {
			dataMap[idx] = jData.Index(idx).Interface()
		}
		for _, data := range dataMap {
			curKey := ""
			curUri := uriWithKeyCreate(uri, xpathPrefix, data)
			if len(xSpecMap[xpathPrefix].xfmrKey) > 0 {
				/* key transformer present */
				ret, err := XlateFuncCall(yangToDbXfmrFunc(xSpecMap[xpathPrefix].xfmrKey), d, ygRoot, oper, curUri)
				if err != nil {
					return err
				}
				curKey = ret[0].Interface().(string)
			} else {
				curKey = keyCreate(keyName, xpathPrefix, data)
			}
			yangReqToDbMapCreate(d, ygRoot, oper, curUri, xpathPrefix, curKey, data, result)
		}
	} else {
		if reflect.ValueOf(jsonData).Kind() == reflect.Map {
			jData := reflect.ValueOf(jsonData)
			for _, key := range jData.MapKeys() {
				typeOfValue := reflect.TypeOf(jData.MapIndex(key).Interface()).Kind()

				if typeOfValue == reflect.Map || typeOfValue == reflect.Slice {
					log.Info("slice/map data: key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)
                    xpath    := uri
                    pathAttr := key.String()
                    if len(xpathPrefix) > 0 {
                         if strings.Contains(pathAttr, ":") {
                             pathAttr = strings.Split(pathAttr, ":")[1]
                         }
                         xpath = xpathPrefix + "/" + pathAttr
                         uri   = uri + "/" + pathAttr
                    }

					if xSpecMap[xpath] != nil && len(xSpecMap[xpath].xfmrFunc) > 0 {
						/* subtree transformer present */
						ret, err := XlateFuncCall(yangToDbXfmrFunc(xSpecMap[xpath].xfmrFunc), d, ygRoot, oper, uri)
						if err != nil {
							return nil
						}
						mapCopy(result, ret[0].Interface().(map[string]map[string]db.Value))
					} else {
						yangReqToDbMapCreate(d, ygRoot, oper, uri, xpath, keyName, jData.MapIndex(key).Interface(), result)
					}
				} else {
					pathAttr := key.String()
					if strings.Contains(pathAttr, ":") {
						pathAttr = strings.Split(pathAttr, ":")[1]
					}
					value := jData.MapIndex(key).Interface()
					log.Info("data field: key(\"%v\"), value(\"%v\").", key, value)
					err := mapFillData(d, ygRoot, oper, uri, keyName, result, xpathPrefix,
						pathAttr, value)
					if err != nil {
						log.Errorf("Failed constructing data for db write: key(\"%v\"), value(\"%v\"), path(\"%v\").",
							pathAttr, value, xpathPrefix)
					}
				}
			}
		}
	}
	return nil
}

func sonicXpathKeyExtract(path string) (string, string, string){
	rgp := regexp.MustCompile(`\[([^\[\]]*)\]`)
	tableName := strings.Split(strings.Split(path , "/")[2], "[")[0]
	xpath, err := RemoveXPATHPredicates(path)
    if err != nil {
        return "", "", ""
    }
	keyStr := ""
	for i, kname := range rgp.FindAllString(path, -1) {
		if i > 0 { keyStr += "|" }
		val := strings.Split(kname, "=")[1]
		keyStr += strings.TrimRight(val, "]")
	}
	return xpath, keyStr, tableName
}

/* Extract key vars, create db key and xpath */
func xpathKeyExtract(path string) (string, string, string) {
	yangXpath := ""
	keyStr := ""
	tableName := ""
	rgp := regexp.MustCompile(`\[([^\[\]]*)\]`)

	for i, k := range strings.Split(path, "/") {
		if i > 0 {
			yangXpath += "/"
		}
		xpath := k
		if strings.Contains(k, "[") {
			if len(keyStr) > 0 {
				keyStr += "|"
			}
			xpath = strings.Split(k, "[")[0]
			var keyl []string
			for _, kname := range rgp.FindAllString(k, -1) {
				keyl = append(keyl, strings.TrimRight(strings.TrimLeft(kname, "["), "]"))
			}
			keyStr += keyFromXpathCreate(keyl)
			if isCvlYang(path) {
				//Format- /module:container/table[key]/field
				// table name extracted from the string token having key entry
				tableName = xpath
			}
		}
		yangXpath += xpath
	}

	return yangXpath, keyStr, tableName
}

/* Debug function to print the map data into file */
func printDbData(db map[string]map[string]db.Value, fileName string) {
	fp, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer fp.Close()

	for k, v := range db {
		fmt.Fprintf(fp, "-----------------------------------------------------------------\r\n")
		fmt.Fprintf(fp, "table name : %v\r\n", k)
		for ik, iv := range v {
			fmt.Fprintf(fp, "  key : %v\r\n", ik)
			for k, d := range iv.Field {
				fmt.Fprintf(fp, "    %v :%v\r\n", k, d)
			}
		}
	}
	fmt.Fprintf(fp, "-----------------------------------------------------------------\r\n")
	return
}
