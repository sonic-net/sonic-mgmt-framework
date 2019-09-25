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
    "translib/ocbinds"
    "github.com/openconfig/ygot/ytypes"

    log "github.com/golang/glog"
)

const SONIC_TABLE_INDEX = 2
const SONIC_FIELD_INDEX = 3

/* Invoke the post tansformer */
func postXfmrHandlerFunc(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    xpath, _ := RemoveXPATHPredicates(inParams.uri)
    ret, err := XlateFuncCall(xSpecMap[xpath].xfmrPost, inParams)
    if err != nil {
        return nil, err
    }
    retData := ret[0].Interface().(map[string]map[string]db.Value)
    log.Info("Post Transformer function :", xSpecMap[xpath].xfmrPost, " Xpath: ", xpath, " retData: ", retData)
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

    result[tableName][dbKey].Field[field] = value
    return
}

/* Fill the redis-db map with data */
func mapFillData(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, dbKey string, result map[string]map[string]db.Value, xpathPrefix string, name string, value interface{}) error {
    xpath := xpathPrefix + "/" + name
    xpathInfo := xSpecMap[xpath]
    log.Infof("name: \"%v\", xpathPrefix(\"%v\").", name, xpathPrefix)

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

    if xpathInfo.isKey {
        return nil
    }
    if len(xpathInfo.xfmrFunc) > 0 {
        uri = uri + "/" + name

        /* field transformer present */
        log.Infof("Transformer function(\"%v\") invoked for yang path(\"%v\").", xpathInfo.xfmrFunc, xpath)
        path, _ := ygot.StringToPath(uri, ygot.StructuredPath, ygot.StringSlicePath)
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
        ocbSch, _ := ocbinds.Schema()
        schRoot := ocbSch.RootSchema()
        node, nErr := ytypes.GetNode(schRoot, (*ygRoot).(*ocbinds.Device), path)
        log.Info("GetNode data: ", node[0].Data, " nErr :", nErr)
        if nErr != nil {
            return nErr
        }
	    var dbs [db.MaxDB]*db.DB
	    inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, uri, oper, "", nil, node[0].Data)
        ret, err := XlateFuncCall(yangToDbXfmrFunc(xSpecMap[xpath].xfmrFunc), inParams)
        if err != nil {
            return err
        }
        retData := ret[0].Interface().(map[string]string)
        log.Info("Transformer function :", xpathInfo.xfmrFunc, " Xpath: ", xpath, " retData: ", retData)
        for f, v := range retData {
            dataToDBMapAdd(*xpathInfo.tableName, dbKey, result, f, v)
        }
        return nil
    }

    if len(xpathInfo.fieldName) == 0 {
        log.Infof("Field for yang-path(\"%v\") not found in DB.", xpath)
        return errors.New("Invalid field name")
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
	    err := errors.New(logStr)
	    return err
	}
	valData := reflect.ValueOf(value)
	for fidx := 0; fidx < valData.Len(); fidx++ {
	    if fidx > 0 {
		valueStr += ","
	    }
	    fVal := fmt.Sprintf("%v", valData.Index(fidx).Interface())
	    valueStr = valueStr + fVal
	}
	log.Infof("leaf-list value after conversion to DB format %v  :  %v", fieldName, valueStr)

    } else { // xpath is a leaf
	valueStr  = fmt.Sprintf("%v", value)
	if strings.Contains(valueStr, ":") {
	    valueStr = strings.Split(valueStr, ":")[1]
	}
    }

    dataToDBMapAdd(*xpathInfo.tableName, dbKey, result, fieldName, valueStr)
    log.Infof("TblName: \"%v\", key: \"%v\", field: \"%v\", valueStr: \"%v\".",
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
		log.Infof("Delete req: path(\"%v\"), key(\"%v\"), xpathPrefix(\"%v\"), tableName(\"%v\").", path, keyName, xpathPrefix, tableName)
		err = cvlYangReqToDbMapDelete(xpathPrefix, tableName, keyName, result)
	} else {
		xpathPrefix, keyName, tableName := xpathKeyExtract(d, ygRoot, oper, path)
		log.Infof("Delete req: path(\"%v\"), key(\"%v\"), xpathPrefix(\"%v\"), tableName(\"%v\").", path, keyName, xpathPrefix, tableName)
		spec, ok := xSpecMap[xpathPrefix]
		if ok {
			if  spec.tableName != nil {
				result[*spec.tableName] = make(map[string]db.Value)
				if len(keyName) > 0 {
					result[*spec.tableName][keyName] = db.Value{Field: make(map[string]string)}
					if spec.yangEntry != nil && spec.yangEntry.Node.Statement().Keyword == "leaf" {
						result[*spec.tableName][keyName].Field[spec.fieldName] = ""
					}
				}
			} else if len(spec.childTable) > 0 {
				for _, child := range spec.childTable {
					result[child] = make(map[string]db.Value)
				}
			}
		}
	}
	log.Infof("Delete req: path(\"%v\") result(\"%v\").", path, result)
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
        }
    } else {
        // Get all table entries
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
    log.Infof("Update/replace req: path(\"%v\").", path)
    dbMapCreate(d, ygRoot, oper, path, jsonData, result)
    log.Infof("Update/replace req: path(\"%v\") result(\"%v\").", path, result)
    printDbData(result, "/tmp/yangToDbDataUpRe.txt")
    return nil
}

/* Get the data from incoming create request, create map and fill with dbValue(ie. field:value
to write into redis-db */
func dbMapCreate(d *db.DB, ygRoot *ygot.GoStruct, oper int, path string, jsonData interface{}, result map[string]map[string]db.Value) error {
	var err error
	root := xpathRootNameGet(path)
	if isCvlYang(path) {
		err = cvlYangReqToDbMapCreate(jsonData, result)
	} else {
		err = yangReqToDbMapCreate(d, ygRoot, oper, root, "", "", jsonData, result)
	}
	if err == nil {
		if oper == CREATE {
			moduleNm := "/" + strings.Split(path, "/")[1]
			log.Infof("Module name for path %s is %s", path, moduleNm)
			if _, ok := xSpecMap[moduleNm]; ok {
				if xSpecMap[moduleNm].yangDataType == "container" {
					log.Info("Invoke post transformer: ", xSpecMap[moduleNm].xfmrPost)
					dbDataMap := make(map[db.DBNum]map[string]map[string]db.Value)
					dbDataMap[db.ConfigDB] = result
					var dbs [db.MaxDB]*db.DB
					inParams := formXfmrInputRequest(d, dbs, db.ConfigDB, ygRoot, path, oper, "", &dbDataMap, nil)
					result, err = postXfmrHandlerFunc(inParams)
				}
			} else {
				log.Errorf("No Entry exists for module %s in XSpecMap. Unable to process post xfmr (\"%v\") path(\"%v\") error (\"%v\").", oper, path, err)
			}
		}
		printDbData(result, "/tmp/yangToDbDataCreate.txt")
	} else {
		log.Errorf("DBMapCreate req failed for oper (\"%v\") path(\"%v\") error (\"%v\").", oper, path, err)
	}
	return err
}

func yangReqToDbMapCreate(d *db.DB, ygRoot *ygot.GoStruct, oper int, uri string, xpathPrefix string, keyName string, jsonData interface{}, result map[string]map[string]db.Value) error {
    log.Infof("key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)
    var dbs [db.MaxDB]*db.DB

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
            if len(xSpecMap[xpathPrefix].xfmrKey) > 0 {
                /* key transformer present */
		        inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, curUri, oper, "", nil, nil)
                ret, err := XlateFuncCall(yangToDbXfmrFunc(xSpecMap[xpathPrefix].xfmrKey), inParams)
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

                log.Infof("slice/map data: key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)
                xpath    := uri
                curUri   := uri
                pathAttr := key.String()
                if len(xpathPrefix) > 0 {
                    if strings.Contains(pathAttr, ":") {
                         pathAttr = strings.Split(pathAttr, ":")[1]
                    }
                    xpath  = xpathPrefix + "/" + pathAttr
                    curUri = uri + "/" + pathAttr
                }

                if (typeOfValue == reflect.Map || typeOfValue == reflect.Slice) && xSpecMap[xpath].yangDataType != "leaf-list" {
                    if xSpecMap[xpath] != nil && len(xSpecMap[xpath].xfmrFunc) > 0 {
                        /* subtree transformer present */
			inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, curUri, oper, "", nil, nil)
                        ret, err := XlateFuncCall(yangToDbXfmrFunc(xSpecMap[xpath].xfmrFunc), inParams)
                        if err != nil {
                            return nil
                        }
                        mapCopy(result, ret[0].Interface().(map[string]map[string]db.Value))
                    } else {
                        yangReqToDbMapCreate(d, ygRoot, oper, curUri, xpath, keyName, jData.MapIndex(key).Interface(), result)
                    }
                } else {
                    pathAttr := key.String()
                    if strings.Contains(pathAttr, ":") {
                        pathAttr = strings.Split(pathAttr, ":")[1]
                    }
                    value := jData.MapIndex(key).Interface()
                    log.Infof("data field: key(\"%v\"), value(\"%v\").", key, value)
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

func sonicXpathKeyExtract(path string) (string, string, string) {
    xpath, keyStr, tableName := "", "", ""
    var err error
    xpath, err = RemoveXPATHPredicates(path)
    if err != nil {
        return xpath, keyStr, tableName
    }
    rgp := regexp.MustCompile(`\[([^\[\]]*)\]`)
    pathsubStr := strings.Split(path , "/")
    if len(pathsubStr) > SONIC_TABLE_INDEX  {
        tableName = strings.Split(pathsubStr[SONIC_TABLE_INDEX], "[")[0]
        for i, kname := range rgp.FindAllString(path, -1) {
            if i > 0 { keyStr += "|" }
            val := strings.Split(kname, "=")[1]
            keyStr += strings.TrimRight(val, "]")
        }
    }
    return xpath, keyStr, tableName
}

/* Extract key vars, create db key and xpath */
func xpathKeyExtract(d *db.DB, ygRoot *ygot.GoStruct, oper int, path string) (string, string, string) {
    keyStr    := ""
    tableName := ""
    rgp       := regexp.MustCompile(`\[([^\[\]]*)\]`)
    curPathWithKey := ""
    var dbs [db.MaxDB]*db.DB

    for _, k := range strings.Split(path, "/") {
        curPathWithKey += k
        if strings.Contains(k, "[") {
            if len(keyStr) > 0 {
                keyStr += "|"
            }
            yangXpath, _ := RemoveXPATHPredicates(curPathWithKey)
	        _, ok := xSpecMap[yangXpath]
	    if ok {
            if len(xSpecMap[yangXpath].xfmrKey) > 0 {
                xfmrFuncName := yangToDbXfmrFunc(xSpecMap[yangXpath].xfmrKey)
		inParams := formXfmrInputRequest(d, dbs, db.MaxDB, ygRoot, curPathWithKey, oper, "", nil, nil)
                ret, err := XlateFuncCall(xfmrFuncName, inParams)
                if err != nil {
                    return "", "", ""
                }
                keyStr = ret[0].Interface().(string)
            } else {
                var keyl []string
                for _, kname := range rgp.FindAllString(k, -1) {
                    keyl = append(keyl, strings.TrimRight(strings.TrimLeft(kname, "["), "]"))
                }
                keyStr += keyFromXpathCreate(keyl)
            }
	    }
        }
        curPathWithKey += "/"
    }
    pfxPath, _ := RemoveXPATHPredicates(path)
    tblPtr     := xSpecMap[pfxPath].tableName
    if tblPtr != nil {
        tableName = *tblPtr
    }

    return pfxPath, keyStr, tableName
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
