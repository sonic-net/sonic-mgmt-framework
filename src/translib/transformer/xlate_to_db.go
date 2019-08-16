package transformer

import (
    "fmt"
    "translib/db"
    "os"
    "reflect"
    "regexp"
    "strings"
    "errors"

    log "github.com/golang/glog"
)

/* Fill the redis-db map with data */
func mapFillData(dbKey string, result map[string]map[string]db.Value, xpathPrefix string, name string, value string) error {
    xpath := xpathPrefix + "/"  + name
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

    if len(xpathInfo.fieldName) == 0 {
        log.Info("Field for yang-path(\"%v\") not found in DB.", xpath)
        return errors.New("Invalid field name")
    }
    fieldName := xpathInfo.fieldName
    if strings.Contains(value, ":") {
        value = strings.Split(value, ":")[1]
    }

    if len(xpathInfo.xfmrFunc) > 0 {
        log.Info("Transformer function(\"%v\") invoked for yang path(\"%v\").", xpathInfo.xfmrFunc, xpath)
        // map[string]string
        //fieldName := XlateFuncCall(xpathInfo.xfmrFunc, name, value)
        //return errors.New("Invalid field name")
        return nil
    }

    _, ok := result[*xpathInfo.tableName]
    if !ok {
        result[*xpathInfo.tableName] = make(map[string]db.Value)
    }

    _, ok = result[*xpathInfo.tableName][dbKey]
    if !ok {
       result[*xpathInfo.tableName][dbKey] = db.Value{Field: make(map[string]string)}
    }

    result[*xpathInfo.tableName][dbKey].Field[fieldName] = value
    log.Info("TblName: \"%v\", key: \"%v\", field: \"%v\", value: \"%v\".",
              *xpathInfo.tableName, dbKey, fieldName, value)
    return nil
}

func callXfmr() map[string]map[string]db.Value {
    result := make(map[string]map[string]db.Value)
    result["ACL_TABLE"] = make(map[string]db.Value)
    result["ACL_TABLE"]["MyACL1_ACL_IPV4"] = db.Value{Field: make(map[string]string)}
    result["ACL_TABLE"]["MyACL1_ACL_IPV4"].Field["stage"]  = "INGRESS"
    result["ACL_TABLE"]["MyACL1_ACL_IPV4"].Field["ports@"] = "Ethernet0"
    result["ACL_TABLE"]["MyACL2_ACL_IPV4"] = db.Value{Field: make(map[string]string)}
    result["ACL_TABLE"]["MyACL2_ACL_IPV4"].Field["stage"]  = "INGRESS"
    result["ACL_TABLE"]["MyACL2_ACL_IPV4"].Field["ports@"] = "Ethernet4"
    return result
}

func cvlYangReqToDbMapCreate(uri string, jsonData interface{}, result map[string]map[string]db.Value) error {
    if reflect.ValueOf(jsonData).Kind() == reflect.Map {
        data := reflect.ValueOf(jsonData)
        for _, key := range data.MapKeys() {
            _, ok := xDbSpecMap[key.String()]
            if ok {
                directDbMapData(key.String(), data.MapIndex(key).Interface(), result)
            } else {
                cvlYangReqToDbMapCreate(key.String(), data.MapIndex(key).Interface(), result)
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
        data       := reflect.ValueOf(jsonData)
        result[tableName] = make(map[string]db.Value)

        for idx := 0; idx < data.Len(); idx++ {
            keyName    := ""
            d := data.Index(idx).Interface().(map[string]interface{})
            for i, k := range tblKeyName {
                if i > 0 { keyName += "|" }
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
func dbMapDelete(path string, jsonData interface{}, result map[string]map[string]db.Value) error {
    xpathPrefix, keyName := xpathKeyExtract(path)
    log.Info("Delete req: path(\"%v\"), key(\"%v\"), xpathPrefix(\"%v\").", path, keyName, xpathPrefix)
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
    log.Info("Delete req: path(\"%v\") result(\"%v\").", path, result)
    return nil
}

/* Get the data from incoming update/replace request, create map and fill with dbValue(ie. field:value 
   to write into redis-db */
func dbMapUpdate(path string, jsonData interface{}, result map[string]map[string]db.Value) error {
    xpathPrefix, keyName := xpathKeyExtract(path)
    log.Info("Update/replace req: path(\"%v\"), key(\"%v\"), xpathPrefix(\"%v\").", path, keyName, xpathPrefix)
    dbMapCreate(parentXpathGet(xpathPrefix), jsonData, result)
    log.Info("Update/replace req: path(\"%v\") result(\"%v\").", path, result)
    return nil
}

/* Get the data from incoming create request, create map and fill with dbValue(ie. field:value 
   to write into redis-db */
func dbMapCreate(uri string, jsonData interface{}, result map[string]map[string]db.Value) error {
    xpathTmplt, keyName := xpathKeyExtract(uri)
    if isCvlYang(uri) {
        cvlYangReqToDbMapCreate(uri, jsonData, result)
    } else {
        yangReqToDbMapCreate(uri, parentXpathGet(xpathTmplt), keyName, jsonData, result)
    }
    printDbData(result, "/tmp/yangToDbData.txt")
    return nil
}

func yangReqToDbMapCreate(uri string, xpathPrefix string, keyName string, jsonData interface{}, result map[string]map[string]db.Value) error {
    log.Info("key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)

    if reflect.ValueOf(jsonData).Kind() == reflect.Slice {
        log.Info("slice data: key(\"%v\"), xpathPrefix(\"%v\").", keyName, xpathPrefix)
        jData   := reflect.ValueOf(jsonData)
        dataMap := make([]interface{}, jData.Len())
        for idx := 0; idx < jData.Len(); idx++ {
            dataMap[idx] = jData.Index(idx).Interface()
        }
        // string
        for _, data := range dataMap {
            keyName := keyCreate(keyName, xpathPrefix, data)
            yangReqToDbMapCreate(uri, xpathPrefix, keyName, data, result)
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
                    }

                    if xSpecMap[xpath] != nil && len(xSpecMap[xpath].xfmrFunc) > 0 {
                        subMap := callXfmr()
                        // map[string]map[string]db.Value
                        //subMap := XlateFuncCall(xpathInfo.xfmrFunc, name, value)
                        mapCopy(result, subMap)
                        return nil
                    } else {
                        yangReqToDbMapCreate(uri, xpath, keyName, jData.MapIndex(key).Interface(), result)
                    }
                } else {
                    pathAttr := key.String()
                    if strings.Contains(pathAttr, ":") {
                        pathAttr = strings.Split(pathAttr, ":")[1]
                    }
                    value := jData.MapIndex(key).Interface()
                    log.Info("data field: key(\"%v\"), value(\"%v\").", key, value)
                    err := mapFillData(keyName, result, xpathPrefix, pathAttr, fmt.Sprintf("%v", value))
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

/* Extract key vars, create db key and xpath */
func xpathKeyExtract(path string) (string, string) {
    yangXpath := ""
    keyStr    := ""
    rgp       := regexp.MustCompile(`\[([^\[\]]*)\]`)

    for i, k := range (strings.Split(path, "/")) {
        if i > 0 { yangXpath += "/" }
        xpath := k
        if strings.Contains(k, "[") {
            if len(keyStr) > 0 { keyStr += "|" }
            xpath = strings.Split(k, "[")[0]
            var keyl []string
            for _, kname := range rgp.FindAllString(k, -1) {
                keyl = append(keyl, strings.TrimRight(strings.TrimLeft(kname, "["), "]"))
            }
            keyStr += keyFromXpathCreate(keyl)
        }
        yangXpath += xpath
    }
    return yangXpath, keyStr
}

/* Debug function to print the map data into file */
func printDbData (db map[string]map[string]db.Value, fileName string) {
    fp, err := os.Create(fileName)
    if err != nil {
        return
    }
    defer fp.Close()

    for k, v := range db {
        fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")
        fmt.Fprintf(fp, "table name : %v\r\n", k)
        for ik, iv := range v {
            fmt.Fprintf(fp, "  key : %v\r\n", ik)
            for k, d := range iv.Field {
                fmt.Fprintf(fp, "    %v :%v\r\n", k, d)
            }
        }
    }
    fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")
    return
}
