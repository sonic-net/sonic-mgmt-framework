package transformer

import (
    "fmt"
    "translib/db"
    "strings"
    "encoding/json"
    "os"
    "translib/ocbinds"
    "github.com/openconfig/ygot/ygot"
    "github.com/openconfig/ygot/ytypes"

    log "github.com/golang/glog"
)

type typeMapOfInterface map[string]interface{}

func xfmrHandlerFunc(d *db.DB, xpath string, uri string, ygRoot *ygot.GoStruct, dbDataMap map[string]map[string]db.Value) (string, error) {
    _, err := XlateFuncCall(dbToYangXfmrFunc(xSpecMap[xpath].xfmrFunc), d, GET, dbDataMap, ygRoot)
    if err != nil {
        return "", err
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

func leafXfmrHandlerFunc(d *db.DB, xpath string, uri string, ygRoot *ygot.GoStruct, dbDataMap map[string]map[string]db.Value) (map[string]interface{}, string, error) {
    _, keyName, _ := xpathKeyExtract(d, ygRoot, GET, uri)
    ret, err := XlateFuncCall(dbToYangXfmrFunc(xSpecMap[xpath].xfmrFunc), d, GET, dbDataMap, ygRoot, keyName)
    if err != nil {
        return nil, "", err
    }
    fldValMap := ret[0].Interface().(map[string]interface{})
    data      := ""
    for f, v  :=  range fldValMap {
        value := fmt.Sprintf("%v", v)
        data += fmt.Sprintf("\"%v\" : \"%v\",", f, value)
    }
    return fldValMap, data, nil
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

func yangListDataFill(d *db.DB, ygRoot *ygot.GoStruct, uri string, xpath string, dbDataMap map[string]map[string]db.Value, resultMap map[string]interface{}, tbl string, tblKey string) error {
    tblData, ok := dbDataMap[tbl]

    if ok {
        var mapSlice []typeMapOfInterface
        for dbKey, _ := range tblData {
            curMap := make(map[string]interface{})
			curKeyMap, curUri, _, _ := dbKeyToYangDataConvert(uri, xpath, dbKey)
            if len(xSpecMap[xpath].xfmrFunc) > 0 {
                jsonStr, _ := xfmrHandlerFunc(d, xpath, curUri, ygRoot, dbDataMap)
                fmt.Printf("From leaf-xfmr(%v)\r\n", jsonStr)
            } else {
                _, keyFromCurUri, _ := xpathKeyExtract(d, ygRoot, GET, curUri)
                if dbKey == keyFromCurUri {
                for k, kv := range curKeyMap {
                    curMap[k] = kv
                }
                curXpath, _ := RemoveXPATHPredicates(curUri)
                yangDataFill(d, ygRoot, curUri, curXpath, dbDataMap, curMap, tbl, dbKey)
                mapSlice = append(mapSlice, curMap)
                }
            }
        }
        if len(mapSlice) > 0 {
            resultMap[xSpecMap[xpath].yangEntry.Name] = mapSlice
        } else {
            fmt.Printf("Map slice is empty.\r\n ")
        }
    }
    return nil
}

func yangDataFill(d *db.DB, ygRoot *ygot.GoStruct, uri string, xpath string, dbDataMap map[string]map[string]db.Value, resultMap map[string]interface{}, tbl string, tblKey string) error {
    var err error
    yangNode, ok := xSpecMap[xpath]

    if ok  && yangNode.yangEntry != nil {
        for yangChldName := range yangNode.yangEntry.Dir {
            chldXpath := xpath+"/"+yangChldName
            chldUri   := uri+"/"+yangChldName
            if xSpecMap[chldXpath] != nil && xSpecMap[chldXpath].yangEntry != nil {
                chldYangType := yangTypeGet(xSpecMap[chldXpath].yangEntry)
                if chldYangType == "leaf" {
                    if len(xSpecMap[chldXpath].xfmrFunc) > 0 {
                        fldValMap, _, err := leafXfmrHandlerFunc(nil, chldXpath, chldUri, ygRoot, dbDataMap)
                        if err != nil {
                            return err
                        }
                        for lf, val := range fldValMap {
                            resultMap[lf] = val
                        }
                    } else {
                        dbFldName := xSpecMap[chldXpath].fieldName
                        if len(dbFldName) > 0  && !xSpecMap[chldXpath].isKey {
                            val, ok := dbDataMap[tbl][tblKey].Field[dbFldName]
                            if ok {
                                resultMap[xSpecMap[chldXpath].yangEntry.Name] = val
                            }
                        }
                    }
                } else if chldYangType == "container" {
                    if len(xSpecMap[chldXpath].xfmrFunc) > 0 {
                        jsonStr, _  := xfmrHandlerFunc(nil, chldXpath, chldUri, ygRoot, dbDataMap)
                        fmt.Printf("From container-xfmr(%v)\r\n", jsonStr)
                    } else {
                        cname := xSpecMap[chldXpath].yangEntry.Name
                        cmap  := make(map[string]interface{})
                        err    = yangDataFill(d, ygRoot, chldUri, chldXpath, dbDataMap, cmap, tbl, tblKey)
                        if len(cmap) > 0 {
                            resultMap[cname] = cmap
                        } else {
                            fmt.Printf("container : empty(%v) \r\n", cname)
                        }
                    }
                } else if chldYangType == "list" {
                    if len(xSpecMap[chldXpath].xfmrFunc) > 0 {
						jsonStr , _ := xfmrHandlerFunc(nil, chldXpath, chldUri, ygRoot, dbDataMap)
                        fmt.Printf("From list-xfmr(%v)\r\n", jsonStr)
                    } else {
                        ynode, ok := xSpecMap[chldXpath]
                        if ok && ynode.tableName != nil {
                            lTblName := *ynode.tableName
                            yangListDataFill(d, ygRoot, chldUri, chldXpath, dbDataMap, resultMap, lTblName, "")
                        }
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
func dbDataToYangJsonCreate(uri string, ygRoot *ygot.GoStruct, dbDataMap map[string]map[string]db.Value) (string, error) {
    jsonData := ""
	if isCvlYang(uri) {
		jsonData := directDbToYangJsonCreate(dbDataMap, jsonData)
		jsonDataPrint(jsonData)
		return jsonData, nil
	}

    reqXpath, keyName, tableName := xpathKeyExtract(nil, nil, GET, uri)
    ftype := yangTypeGet(xSpecMap[reqXpath].yangEntry)
    if ftype == "leaf" {
        fldName := xSpecMap[reqXpath].fieldName
        tbl, key, _ := tableNameAndKeyFromDbMapGet(dbDataMap)
        jsonData = fmt.Sprintf("{\r\n \"%v\" : \"%v\" \r\n }\r\n", xSpecMap[reqXpath].yangEntry.Name,
                               dbDataMap[tbl][key].Field[fldName])
        return jsonData, nil
    }

    resultMap := make(map[string]interface{})
    var d *db.DB
    yangDataFill(d, ygRoot, uri, reqXpath, dbDataMap, resultMap, tableName, keyName)
    jsonMapData, _ := json.Marshal(resultMap)
    jsonData        = fmt.Sprintf("%v", string(jsonMapData))
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

