package transformer

import (
    "fmt"
    "translib/db"
    "strings"
    "os"
    "translib/ocbinds"
    "github.com/openconfig/ygot/ygot"
    "github.com/openconfig/ygot/ytypes"

    log "github.com/golang/glog"
)

func xfmrHandlerFunc(d *db.DB, xpath string, uri string, ygRoot *ygot.GoStruct, dbDataMap map[string]map[string]db.Value) (string, error) {
    //var root ygot.ValidatedGoStruct
    //ygRoot  := &root
    /*ygot.BuildEmptyTree(root)*/
    ret, err := XlateFuncCall(dbToYangXfmrFunc(xSpecMap[xpath].xfmrFunc), d, ygRoot, GET, uri, dbDataMap)
    if err != nil {
        return "", err
    }

    errData    := ret[0].Interface().(error)
    if errData != nil {
        return "", errData
    }
    ocbSch, _  := ocbinds.Schema()
    schRoot    := ocbSch.RootSchema()
    device     := (*ygRoot).(*ocbinds.Device)

    log.Info("Subtree transformer function(\"%v\") invoked for yang path(\"%v\").", xSpecMap[xpath].xfmrFunc, xpath)
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

    nodeList, nErr := ytypes.GetNode(schRoot, device, path)
    if nErr != nil {
        return "", nErr
    }
    node := nodeList[0].Data
    nodeYgot, _:= (node).(ygot.ValidatedGoStruct)
    payload, err := ygot.EmitJSON(nodeYgot, &ygot.EmitJSONConfig{ Format: ygot.RFC7951,
                                  Indent: "  ", SkipValidation: true,
                                  RFC7951Config: &ygot.RFC7951JSONConfig{ AppendModuleName: false, },
                                  })
    return payload, err
}

/* Traverse db map and add data to json */
func dataToJsonAdd(uri string, xpath string, ygRoot *ygot.GoStruct, fieldData map[string]string, key string, dbDataMap map[string]map[string]db.Value) (string, error) {
    spec, ok := xSpecMap[xpath]
    jsonData := ""

    if ok {
        for fld := range spec.yangEntry.Dir {
            fldXpath := xpath+"/"+fld
			curUri   := uri+"/"+fld
            if xSpecMap[fldXpath] != nil && xSpecMap[fldXpath].yangEntry != nil {
                ftype := yangTypeGet(xSpecMap[fldXpath].yangEntry)
                if ftype == "leaf" {
                    if len(xSpecMap[fldXpath].xfmrFunc) > 0 {
                        /* field transformer present */
                        jsonStr, err  := xfmrHandlerFunc(nil, fldXpath, curUri, ygRoot, dbDataMap)
                        if err != nil {
                            return "", err
                        }
                        jsonData += jsonStr
                    } else {
                        /* Add db field and value to json, call xfmr if needed */
                        fldName := xSpecMap[fldXpath].fieldName
                        if len(fldName) > 0 {
                            val, ok := fieldData[fldName]
                            if ok {
                                jsonData += fmt.Sprintf("\"%v\" : \"%v\",", xSpecMap[fldXpath].yangEntry.Name, val)
                            }
                        }
                    }
                } else if ftype == "container" && xSpecMap[fldXpath].yangEntry.Name != "state" {
                    if len(xSpecMap[fldXpath].xfmrFunc) > 0 {
                        jsonStr, _  := xfmrHandlerFunc(nil, fldXpath, curUri, ygRoot, dbDataMap)
                        jsonData += jsonStr
                    } else {
                        /* Create container enclosure and attach container name and add to json */
                        data, _:= dataToJsonAdd(curUri, fldXpath, ygRoot, fieldData, key, dbDataMap)
                        if len(data) > 0 {
                            jsonData += fmt.Sprintf("\"%v\" : { \r\n %v \r\n },",
                                                    xSpecMap[fldXpath].yangEntry.Name, data)
						}
					}
                } else if ftype == "list" {
					if len(xSpecMap[fldXpath].xfmrFunc) > 0 {
						jsonStr , _ := xfmrHandlerFunc(nil, fldXpath, curUri, ygRoot, dbDataMap)
						jsonData += jsonStr
					} else {
						/* Inner(child) list, traverse this list */
                        childMap, ok := dbDataMap[*xSpecMap[fldXpath].tableName]
                        if ok {
                            var xpathl []string
                            xpathl = append(xpathl, fldXpath)
                            jsonData += listDataToJsonAdd(curUri, ygRoot, xpathl, childMap, key, dbDataMap)
						}
					}
				}
			}
        }
		/* Last node in json data in current context, trim extra "," in data, so that json data is valid */
		jsonData = strings.TrimRight(jsonData, ",")
    }
    return jsonData, nil
}

/* Traverse list data and add to json */
func listDataToJsonAdd(uri string, ygRoot *ygot.GoStruct, xpathl []string, dataMap map[string]db.Value, key string, dbDataMap map[string]map[string]db.Value) string {
    jsonData := ""

    for _, xpath := range xpathl {
        for kval, data := range dataMap {
            if len(key) > 0 && !strings.HasPrefix(kval, key) {
                continue
            }
			curUri, kdata, _ := dbKeyToYangDataConvert(uri, xpath, kval)
			/* Traverse list members and add to json */
			data, _ := dataToJsonAdd(curUri, xpath, ygRoot, data.Field, kval, dbDataMap)
			data    += kdata
            if len(data) > 0 {
				/* Enclose all list instances with {} */
                jsonData += fmt.Sprintf("{\r\n %v },", data)
            }
			/* Added data to json, so delete current instance data */
			delete(dataMap, kval)
        }
        if len(jsonData) > 0 {
		    /* Last data in list,so trim extra "," in data, so that the json is valid */
			jsonData = strings.TrimRight(jsonData, ",")
			/* Create list enclosure, attach list-name and add to json */
            jsonData = fmt.Sprintf("\"%v\" : [\r\n %v\r\n ]\r\n", xSpecMap[xpath].yangEntry.Name, jsonData)
        }
    }
    jsonData = strings.TrimRight(jsonData, ",")
    return jsonData
}

/* Traverse db map and create json for cvl yang */
func directDbToYangJsonCreate(dbDataMap map[string]map[string]db.Value, jsonData string) string {
    for tblName, tblData := range dbDataMap {
        dataInst := ""
        for keyStr, dbFldValData := range tblData {
            fldValPair := ""
            for field, value := range dbFldValData.Field {
                fldValPair += fmt.Sprintf("\"%v\" : \"%v\",\r\n", field, value)
            }
            yangKeys := yangKeyFromEntryGet(xDbSpecMap[tblName].dbEntry)
            fldValPair = keyJsonDataAdd(yangKeys, keyStr, fldValPair)
            dataInst += fmt.Sprintf("{ \r\n %v \r\n },", fldValPair)
        }
        dataInst = strings.TrimRight(dataInst, ",")
        jsonData += fmt.Sprintf("\"%v\" : [\r\n %v\r\n ],", tblName, dataInst)
    }
    jsonData = strings.TrimRight(jsonData, ",")
    return jsonData
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

/* Traverse linear db-map data and add to nested json data */
func dbDataToYangJsonCreate(uri string, ygRoot *ygot.GoStruct, dbDataMap map[string]map[string]db.Value) (string, error) {
    jsonData := ""

	if isCvlYang(uri) {
		jsonData := directDbToYangJsonCreate(dbDataMap, jsonData)
		jsonDataPrint(jsonData)
		return jsonData, nil
	}

    reqXpath, _, _ := xpathKeyExtract(nil, nil, GET, uri)
    ftype := yangTypeGet(xSpecMap[reqXpath].yangEntry)
    if ftype == "leaf" {
        fldName := xSpecMap[reqXpath].fieldName
        tbl, key, _ := tableNameAndKeyFromDbMapGet(dbDataMap)
        jsonData = fmt.Sprintf("{\r\n \"%v\" : \"%v\" \r\n }\r\n", xSpecMap[reqXpath].yangEntry.Name,
                               dbDataMap[tbl][key].Field[fldName])
        return jsonData, nil
    }

    curXpath := ""
    tableOrder := [...]string{"ACL_TABLE", "ACL_RULE"}
    for tblId := range tableOrder {
        tblName := tableOrder[tblId]
        if dbDataMap[tblName] != nil {
            if len(curXpath) == 0 || strings.HasPrefix(curXpath, xDbSpecMap[tblName].yangXpath[0]) {
                curXpath = xDbSpecMap[tblName].yangXpath[0]
            }
            jsonData += listDataToJsonAdd(uri, ygRoot, xDbSpecMap[tblName].yangXpath, dbDataMap[tblName], "", dbDataMap)
         }
    }
    if strings.HasPrefix(reqXpath, curXpath) {
        if ftype != "leaf" {
            jsonData = fmt.Sprintf("{ \r\n %v \r\n }", jsonData)
        }
        return jsonData, nil
    }
    jsonData = parentJsonDataUpdate(reqXpath, curXpath, jsonData)
	jsonDataPrint(jsonData)
    return jsonData, nil
}

func xpathLastAttrGet(xpath string) string {
    attrList := strings.Split(xpath, "/")
    return attrList[len(attrList)-1]
}

func jsonPayloadComplete(reqXpath string, data string) string {
    entry     := xSpecMap[reqXpath].yangEntry
    entryType := entry.Node.Statement().Keyword
    name      := xpathLastAttrGet(reqXpath)
    switch entryType {
        case "container":
            data = fmt.Sprintf("\"%v\" : { \r\n %v \r\n }\r\n", name, data)
        case "list":
            data = fmt.Sprintf("\"%v\" : [\r\n %v\r\n ]\r\n", name, data)
    }
    data  = fmt.Sprintf("{\r\n %v }\r\n", data)
    return data
}

func parentJsonDataUpdate(reqXpath string, xpath string, data string) string {
    curXpath := parentXpathGet(xpath)
    if reqXpath == xpath {
        data  = fmt.Sprintf("{\r\n %v }\r\n", data)
        return data
    }
    if reqXpath == curXpath {
        data = jsonPayloadComplete(reqXpath, data)
        return data
    }
    if xSpecMap[curXpath] != nil {
        entry     := xSpecMap[curXpath].yangEntry
        entryType := entry.Node.Statement().Keyword
        switch entryType {
            case "container":
                data = fmt.Sprintf("\"%v\" : { \r\n %v \r\n }", xSpecMap[curXpath].yangEntry.Name, data)
                return parentJsonDataUpdate(reqXpath, curXpath, data)
            case "list":
                data = fmt.Sprintf("\"%v\" : [\r\n %v\r\n ]\r\n", xSpecMap[curXpath].yangEntry.Name, data)
                return parentJsonDataUpdate(reqXpath, curXpath, data)
            case "module":
                data = fmt.Sprintf("\"%v\" : { \r\n %v \r\n }", xSpecMap[curXpath].yangEntry.Name, data)
                return data
            default:
               return ""
        }
    }
    return ""
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

