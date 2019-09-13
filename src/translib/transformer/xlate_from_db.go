package transformer

import (
    "fmt"
    "translib/db"
    "strings"
    "encoding/json"
    "os"
    "strconv"
    "errors"
    "translib/ocbinds"
    "github.com/openconfig/goyang/pkg/yang"
    "github.com/openconfig/ygot/ygot"
    "github.com/openconfig/ygot/ytypes"

    log "github.com/golang/glog"
)

type typeMapOfInterface map[string]interface{}

const (
    Yuint8 = 5
)

func xfmrHandlerFunc(inParams XfmrParams) (string, error) {
    xpath, _ := RemoveXPATHPredicates(inParams.uri)
    _, err := XlateFuncCall(dbToYangXfmrFunc(xSpecMap[xpath].xfmrFunc), inParams)
    if err != nil {
        return "", err
    }

    ocbSch, _  := ocbinds.Schema()
    schRoot    := ocbSch.RootSchema()
    device     := (*inParams.ygRoot).(*ocbinds.Device)

    log.Info("Subtree transformer function(\"%v\") invoked for yang path(\"%v\").", xSpecMap[xpath].xfmrFunc, xpath)
    path, _ := ygot.StringToPath(inParams.uri, ygot.StructuredPath, ygot.StringSlicePath)
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

func leafXfmrHandlerFunc(inParams XfmrParams) (map[string]interface{}, string, error) {
    xpath, _ := RemoveXPATHPredicates(inParams.uri)
    ret, err := XlateFuncCall(dbToYangXfmrFunc(xSpecMap[xpath].xfmrFunc), inParams)
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

func validateHandlerFunc(inParams XfmrParams) (bool) {
    xpath, _ := RemoveXPATHPredicates(inParams.uri)
    ret, err := XlateFuncCall(xSpecMap[xpath].validateFunc, inParams)
    if err != nil {
        return false
    }
    return ret[0].Interface().(bool)
}

func DbValToInt(dbFldVal string, base int, size int, isUint bool) (interface{}, error) {
	var res interface{}
	var err error
	if res, err = strconv.ParseInt(dbFldVal, base, size); err != nil {
		log.Warning("Non Yint%v type for yang leaf-list item %v", size, dbFldVal)
	}
	return res, err
}

func DbToYangType(yngTerminalNdDtType yang.TypeKind, fldXpath string, dbFldVal string) (interface{}, error) {
	log.Infof("Received FieldXpath %v, yngTerminalNdDtType %v and Db field value %v to be converted to yang data-type.", fldXpath, yngTerminalNdDtType, dbFldVal)
	var res interface{}
	var err error
	const INTBASE = 10
	switch yngTerminalNdDtType {
        case yang.Ynone:
                log.Warning("Yang node data-type is non base yang type")
		//TODO - enhance to handle non base data types depending on future use case
		err = errors.New("Yang node data-type is non base yang type")
        case yang.Yint8:
                res, err = DbValToInt(dbFldVal, INTBASE, 8, false)
        case yang.Yint16:
                res, err = DbValToInt(dbFldVal, INTBASE, 16, false)
        case yang.Yint32:
                res, err = DbValToInt(dbFldVal, INTBASE, 32, false)
        case yang.Yint64:
                res, err = DbValToInt(dbFldVal, INTBASE, 64, false)
        case yang.Yuint8:
                res, err = DbValToInt(dbFldVal, INTBASE, 8, true)
        case yang.Yuint16:
                res, err = DbValToInt(dbFldVal, INTBASE, 16, true)
        case yang.Yuint32:
                res, err = DbValToInt(dbFldVal, INTBASE, 32, true)
        case yang.Yuint64:
                res, err = DbValToInt(dbFldVal, INTBASE, 64, true)
        case yang.Ybool:
		if res, err = strconv.ParseBool(dbFldVal); err != nil {
			log.Warning("Non Bool type for yang leaf-list item %v", dbFldVal)
		}
        //TODO enhance Yunion using yangEntry way
        case yang.Yenum, yang.Ystring, yang.Yidentityref, yang.Yunion:
                log.Info("Yenum/Ystring/Yunion(having all members as strings) type for yangXpath ", fldXpath)
                res = dbFldVal
	case yang.Yempty:
		logStr := fmt.Sprintf("Yang data type for xpath %v is Yempty.", fldXpath)
		log.Error(logStr)
		err = errors.New(logStr)
        default:
		logStr := fmt.Sprintf("Unrecognized/Unhandled yang-data type(%v) for xpath %v.", fldXpath, yang.TypeKindToName[yngTerminalNdDtType])
                log.Error(logStr)
                err = errors.New(logStr)
        }
	return res, err
}

/*convert leaf-list in Db to leaf-list in yang*/
func processLfLstDbToYang(fieldXpath string, dbFldVal string) []interface{} {
	valLst := strings.Split(dbFldVal, ",")
	var resLst []interface{}
	const INTBASE = 10
	yngTerminalNdDtType := xDbSpecMap[fieldXpath].dbEntry.Type.Kind
	switch  yngTerminalNdDtType {
	case yang.Yenum, yang.Ystring, yang.Yunion:
		log.Info("DB leaf-list and Yang leaf-list are of same data-type")
		for _, fldVal := range valLst {
			resLst = append(resLst, fldVal)
		}
	default:
		for _, fldVal := range valLst {
			resVal, err := DbToYangType(yngTerminalNdDtType, fieldXpath, fldVal)
			if err == nil {
				resLst = append(resLst, resVal)
			}
		}
	}
	return resLst
}

/* Traverse db map and create json for cvl yang */
func directDbToYangJsonCreate(dbDataMap map[string]map[string]db.Value, jsonData string) string {
    for tblName, tblData := range dbDataMap {
        dataInst := ""
        for keyStr, dbFldValData := range tblData {
            fldValPair := ""
            for field, value := range dbFldValData.Field {
			if strings.HasSuffix(field, "@") {
                                fldVals := strings.Split(field, "@")
                                field = fldVals[0]
                                fieldXpath := tblName + "/" + field
                                if xDbSpecMap[fieldXpath].dbEntry.IsLeafList() {
                                        resLst := processLfLstDbToYang(fieldXpath, value)
                                        fldValPair += fmt.Sprintf("\"%v\" : \"%v\",\r\n", field, resLst)
                                        continue
                                }
                        } else {
                                fieldXpath := tblName + "/" + field
                                yngTerminalNdDtType := xDbSpecMap[fieldXpath].dbEntry.Type.Kind
                                resVal, err := DbToYangType(yngTerminalNdDtType, fieldXpath, value)
                                if err != nil {
                                        log.Warning("Failure in converting Db value type to yang type for field", field)
                                } else {
                                        fldValPair += fmt.Sprintf("\"%v\" : \"%v\",\r\n", field, resVal)
                                }
                        }
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

func yangListDataFill(dbs [db.MaxDB]*db.DB, ygRoot *ygot.GoStruct, uri string, xpath string, dbDataMap *map[db.DBNum]map[string]map[string]db.Value, resultMap map[string]interface{}, tbl string, tblKey string, cdb db.DBNum) error {
    tblData, ok := (*dbDataMap)[cdb][tbl]

    if ok {
        var mapSlice []typeMapOfInterface
        for dbKey, _ := range tblData {
            curMap := make(map[string]interface{})
			curKeyMap, curUri, _, _ := dbKeyToYangDataConvert(uri, xpath, dbKey)
            if len(xSpecMap[xpath].xfmrFunc) > 0 {
		inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, curUri, GET, "", dbDataMap, nil)
                jsonStr, _ := xfmrHandlerFunc(inParams)
                fmt.Printf("From leaf-xfmr(%v)\r\n", jsonStr)
            } else {
                _, keyFromCurUri, _ := xpathKeyExtract(dbs[cdb], ygRoot, GET, curUri)
                if dbKey == keyFromCurUri {
                    for k, kv := range curKeyMap {
                        curMap[k] = kv
                    }
                    curXpath, _ := RemoveXPATHPredicates(curUri)
                    yangDataFill(dbs, ygRoot, curUri, curXpath, dbDataMap, curMap, tbl, dbKey, cdb)
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

func yangDataFill(dbs [db.MaxDB]*db.DB, ygRoot *ygot.GoStruct, uri string, xpath string, dbDataMap *map[db.DBNum]map[string]map[string]db.Value, resultMap map[string]interface{}, tbl string, tblKey string, cdb db.DBNum) error {
    var err error
    yangNode, ok := xSpecMap[xpath]

    if ok  && yangNode.yangEntry != nil {
        for yangChldName := range yangNode.yangEntry.Dir {
            chldXpath := xpath+"/"+yangChldName
            chldUri   := uri+"/"+yangChldName
            if xSpecMap[chldXpath] != nil && xSpecMap[chldXpath].yangEntry != nil {
                if len(xSpecMap[chldXpath].validateFunc) > 0 {
                   // TODO - handle non CONFIG-DB
                   inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, chldUri, GET, "", dbDataMap, nil)
                   res := validateHandlerFunc(inParams)
                   if res != true {
                      continue
                   }
                }
                chldYangType := yangTypeGet(xSpecMap[chldXpath].yangEntry)
		cdb = xSpecMap[chldXpath].dbIndex
                if xSpecMap[chldXpath].yangEntry.IsLeaf() || xSpecMap[chldXpath].yangEntry.IsLeafList() {
                    if len(xSpecMap[chldXpath].xfmrFunc) > 0 {
			_, key, _ := xpathKeyExtract(dbs[cdb], ygRoot, GET, chldUri)
			inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, chldUri, GET, key, dbDataMap, nil)
                        fldValMap, _, err := leafXfmrHandlerFunc(inParams)
                        if err != nil {
                            return err
                        }
                        for lf, val := range fldValMap {
                            resultMap[lf] = val
                        }
                    } else {
                        dbFldName := xSpecMap[chldXpath].fieldName
			/* if there is no transformer extension/annotation then it means leaf-list in yang is also leaflist in db */
                        if len(dbFldName) > 0  && !xSpecMap[chldXpath].isKey {
				if xSpecMap[chldXpath].yangEntry.IsLeafList() {
				    dbFldName += "@"
                                    val, ok := (*dbDataMap)[cdb][tbl][tblKey].Field[dbFldName]
				    if ok {
					resLst := processLfLstDbToYang(chldXpath, val)
					resultMap[xSpecMap[chldXpath].yangEntry.Name] = resLst
				    }
				} else {
				    val, ok := (*dbDataMap)[cdb][tbl][tblKey].Field[dbFldName]
				    if ok {
					/* this will be enhanced to support all yang data types */
					yngTerminalNdDtType := xSpecMap[chldXpath].yangEntry.Type.Kind
                                        resVal, err := DbToYangType(yngTerminalNdDtType, chldXpath, val)
					if err != nil {
					    log.Warning("Failure in converting Db value type to yang type for field", chldXpath)
					} else {
					    resultMap[xSpecMap[chldXpath].yangEntry.Name] = resVal
					}
				    }
				}
                        }
                    }
                } else if chldYangType == "container" {
                    if len(xSpecMap[chldXpath].xfmrFunc) > 0 {
			inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, chldUri, GET, "", dbDataMap, nil)
                        jsonStr, _  := xfmrHandlerFunc(inParams)
                        fmt.Printf("From container-xfmr(%v)\r\n", jsonStr)
                    } else {
                        cname := xSpecMap[chldXpath].yangEntry.Name
                        cmap  := make(map[string]interface{})
                        err    = yangDataFill(dbs, ygRoot, chldUri, chldXpath, dbDataMap, cmap, tbl, tblKey, cdb)
                        if len(cmap) > 0 {
                            resultMap[cname] = cmap
                        } else {
                            fmt.Printf("container : empty(%v) \r\n", cname)
                        }
                    }
                } else if chldYangType == "list" {
		    cdb = xSpecMap[chldXpath].dbIndex
                    if len(xSpecMap[chldXpath].xfmrFunc) > 0 {
			inParams := formXfmrInputRequest(dbs[cdb], dbs, cdb, ygRoot, chldUri, GET, "", dbDataMap, nil)
			jsonStr , _ := xfmrHandlerFunc(inParams)
                        fmt.Printf("From list-xfmr(%v)\r\n", jsonStr)
                    } else {
                        ynode, ok := xSpecMap[chldXpath]
                        if ok && ynode.tableName != nil {
                            lTblName := *ynode.tableName
                            yangListDataFill(dbs, ygRoot, chldUri, chldXpath, dbDataMap, resultMap, lTblName, "", cdb)
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
func dbDataToYangJsonCreate(uri string, ygRoot *ygot.GoStruct, dbs [db.MaxDB]*db.DB, dbDataMap *map[db.DBNum]map[string]map[string]db.Value, cdb db.DBNum) (string, error) {
    jsonData := ""
	if isCvlYang(uri) {
		jsonData := directDbToYangJsonCreate((*dbDataMap)[cdb], jsonData)
		jsonDataPrint(jsonData)
		return jsonData, nil
	}

    var d *db.DB
    resultMap := make(map[string]interface{})
    reqXpath, keyName, tableName := xpathKeyExtract(d, ygRoot, GET, uri)
    yangNode, ok := xSpecMap[reqXpath]
    if ok {
        yangType := yangTypeGet(yangNode.yangEntry)
        if yangType == "leaf" {
            fldName := xSpecMap[reqXpath].fieldName
            tbl, key, _ := tableNameAndKeyFromDbMapGet((*dbDataMap)[cdb])
            jsonData = fmt.Sprintf("{\r\n \"%v\" : \"%v\" \r\n }\r\n", xSpecMap[reqXpath].yangEntry.Name,
                               (*dbDataMap)[cdb][tbl][key].Field[fldName])
            return jsonData, nil
        } else {
            yangDataFill(dbs, ygRoot, uri, reqXpath, dbDataMap, resultMap, tableName, keyName, cdb)
        }
    }

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

