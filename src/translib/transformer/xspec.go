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
    "os"
    "strings"
    log "github.com/golang/glog"
    "translib/db"

    "github.com/openconfig/goyang/pkg/yang"
)

/* Data needed to construct lookup table from yang */
type yangXpathInfo  struct {
    yangDataType   string
    tableName      *string
    xfmrTbl        *string
    childTable      []string
    dbEntry        *yang.Entry
    yangEntry      *yang.Entry
    keyXpath       map[int]*[]string
    delim          string
    fieldName      string
    xfmrFunc       string
    xfmrPost       string
    validateFunc   string
    xfmrKey        string
    keyName        *string
    dbIndex        db.DBNum
    keyLevel       int
    isKey          bool
}

type dbInfo  struct {
    dbIndex      db.DBNum
    keyName      *string
    fieldType    string
    dbEntry      *yang.Entry
    yangXpath    []string
}

var xYangSpecMap  map[string]*yangXpathInfo
var xDbSpecMap    map[string]*dbInfo
var xDbSpecOrdTblMap map[string][]string //map of module-name to ordered list of db tables { "sonic-acl" : ["ACL_TABLE", "ACL_RULE"] }

/* update transformer spec with db-node */
func updateDbTableData (xpath string, xpathData *yangXpathInfo, tableName string) {
	_, ok := xDbSpecMap[tableName]
	if ok {
		xDbSpecMap[tableName].yangXpath = append(xDbSpecMap[tableName].yangXpath, xpath)
		xpathData.dbEntry = xDbSpecMap[tableName].dbEntry
	}
}

/* Recursive api to fill the map with yang details */
func yangToDbMapFill (keyLevel int, xYangSpecMap map[string]*yangXpathInfo, entry *yang.Entry, xpathPrefix string) {
	xpath := ""
	/* create the yang xpath */
	if xYangSpecMap[xpathPrefix] != nil  && xYangSpecMap[xpathPrefix].yangDataType == "module" {
		/* module name is separated from the rest of xpath with ":" */
		xpath = xpathPrefix + ":" + entry.Name
	} else {
		xpath = xpathPrefix + "/" + entry.Name
	}

	xpathData, ok := xYangSpecMap[xpath]
	if !ok {
		xpathData = new(yangXpathInfo)
		xYangSpecMap[xpath] = xpathData
		xpathData.dbIndex = db.ConfigDB // default value
	} else {
		xpathData = xYangSpecMap[xpath]
	}

	xpathData.yangDataType = entry.Node.Statement().Keyword
	if entry.Node.Statement().Keyword == "list"  && xpathData.tableName != nil {
		childToUpdateParent(xpath, *xpathData.tableName)
	}

	parentXpathData, ok := xYangSpecMap[xpathPrefix]
	/* init current xpath table data with its parent data, change only if needed. */
	if ok {
		if xpathData.tableName == nil && parentXpathData.tableName != nil && xpathData.xfmrTbl == nil {
			xpathData.tableName = parentXpathData.tableName
		} else if xpathData.xfmrTbl == nil && parentXpathData.xfmrTbl != nil {
			xpathData.xfmrTbl = parentXpathData.xfmrTbl
		}
	}

	if ok && xpathData.dbIndex == db.ConfigDB && parentXpathData.dbIndex != db.ConfigDB {
		// If DB Index is not annotated and parent DB index is annotated inherit the DB Index of the parent
		xpathData.dbIndex = parentXpathData.dbIndex
	}

	if ok && len(parentXpathData.validateFunc) > 0 {
		xpathData.validateFunc = parentXpathData.validateFunc
	}

	if ok && len(parentXpathData.xfmrFunc) > 0 && len(xpathData.xfmrFunc) == 0 {
		xpathData.xfmrFunc = parentXpathData.xfmrFunc
	}

	if xpathData.yangDataType == "leaf" && len(xpathData.fieldName) == 0 {
		if xpathData.tableName != nil && xDbSpecMap[*xpathData.tableName] != nil {
			if xDbSpecMap[*xpathData.tableName].dbEntry.Dir[entry.Name] != nil {
				xpathData.fieldName = entry.Name
			} else if xDbSpecMap[*xpathData.tableName].dbEntry.Dir[strings.ToUpper(entry.Name)] != nil {
				xpathData.fieldName = strings.ToUpper(entry.Name)
			}
		} else if xpathData.xfmrTbl != nil {
			/* table transformer present */
			xpathData.fieldName = entry.Name
		}
	}

	if xpathData.yangDataType == "leaf" && len(xpathData.fieldName) > 0 && xpathData.tableName != nil {
		dbPath := *xpathData.tableName + "/" + xpathData.fieldName
		if xDbSpecMap[dbPath] != nil {
			xDbSpecMap[dbPath].yangXpath = append(xDbSpecMap[dbPath].yangXpath, xpath)
		}
	}

	/* fill table with key data. */
	curKeyLevel := keyLevel
	if len(entry.Key) != 0 {
		parentKeyLen := 0

		/* create list with current keys */
		keyXpath        := make([]string, len(strings.Split(entry.Key, " ")))
		for id, keyName := range(strings.Split(entry.Key, " ")) {
			keyXpath[id] = xpath + "/" + keyName
			keyXpathData := new(yangXpathInfo)
			xYangSpecMap[xpath + "/" + keyName] = keyXpathData
			xYangSpecMap[xpath + "/" + keyName].isKey = true
		}

		xpathData.keyXpath = make(map[int]*[]string, (parentKeyLen + 1))
		k := 0
		for ; k < parentKeyLen; k++ {
			/* copy parent key-list to child key-list*/
			xpathData.keyXpath[k] = parentXpathData.keyXpath[k]
		}
		xpathData.keyXpath[k] = &keyXpath
		xpathData.keyLevel    = curKeyLevel
		curKeyLevel++
	} else if parentXpathData != nil && parentXpathData.keyXpath != nil {
		xpathData.keyXpath = parentXpathData.keyXpath
	}

	/* get current obj's children */
	var childList []string
	for k := range entry.Dir {
		childList = append(childList, k)
	}

	xpathData.yangEntry = entry
	/* now recurse, filling the map with current node's children info */
	for _, child := range childList {
		yangToDbMapFill(curKeyLevel, xYangSpecMap, entry.Dir[child], xpath)
	}
}

/* Build lookup table based of yang xpath */
func yangToDbMapBuild(entries map[string]*yang.Entry) {
    if entries == nil {
        return
    }

    if xYangSpecMap == nil {
        xYangSpecMap = make(map[string]*yangXpathInfo)
    }

    for module, e := range entries {
        if e == nil || len(e.Dir) == 0 {
            continue
        }

	/* Start to fill xpath based map with yang data */
    keyLevel := 0
    yangToDbMapFill(keyLevel, xYangSpecMap, e, "")

	// Fill the ordered map of child tables list for oc yangs
	updateSchemaOrderedMap(module, e)
    }
    mapPrint(xYangSpecMap, "/tmp/fullSpec.txt")
    dbMapPrint("/tmp/dbSpecMap.txt")
}

/* Fill the map with db details */
func dbMapFill(tableName string, curPath string, moduleNm string, trkTpCnt bool, xDbSpecMap map[string]*dbInfo, entry *yang.Entry) {
	entryType := entry.Node.Statement().Keyword

	if entry.Name != moduleNm {
		if entryType == "container" {
			tableName = entry.Name
		}

		if !isYangResType(entryType) {
			dbXpath := tableName
			if entryType != "container" {
				dbXpath = tableName + "/" + entry.Name
			}
			xDbSpecMap[dbXpath] = new(dbInfo)
			xDbSpecMap[dbXpath].dbIndex   = db.MaxDB
			xDbSpecMap[dbXpath].dbEntry   = entry
			xDbSpecMap[dbXpath].fieldType = entryType
			if entryType == "container" {
				xDbSpecMap[dbXpath].dbIndex = db.ConfigDB
				if entry.Exts != nil && len(entry.Exts) > 0 {
					for _, ext := range entry.Exts {
						dataTagArr := strings.Split(ext.Keyword, ":")
						tagType := dataTagArr[len(dataTagArr)-1]
						switch tagType {
						case "key-name" :
							if xDbSpecMap[dbXpath].keyName == nil {
								xDbSpecMap[dbXpath].keyName = new(string)
							}
							*xDbSpecMap[dbXpath].keyName = ext.NName()
						default :
							log.Infof("Unsupported ext type(%v) for xpath(%v).", tagType, dbXpath)
						}
					}
				}
			}
		}
	} else {
		moduleXpath := "/" + moduleNm + ":" + entry.Name
		xDbSpecMap[moduleXpath] = new(dbInfo)
		xDbSpecMap[moduleXpath].dbEntry   = entry
		xDbSpecMap[moduleXpath].fieldType = entryType
	}

	var childList []string
	for _, k := range entry.DirOKeys {
		childList = append(childList, k)
	}

	if entryType == "container" &&  trkTpCnt {
		xDbSpecOrdTblMap[moduleNm] = childList
		log.Info("xDbSpecOrdTblMap after appending ", xDbSpecOrdTblMap)
		trkTpCnt = false
	}

	for _, child := range childList {
		childPath := tableName + "/" + entry.Dir[child].Name
		dbMapFill(tableName, childPath, moduleNm, trkTpCnt, xDbSpecMap, entry.Dir[child])
	}
}

/* Build redis db lookup map */
func dbMapBuild(entries []*yang.Entry) {
	if entries == nil {
		return
	}
	xDbSpecMap = make(map[string]*dbInfo)
	xDbSpecOrdTblMap = make(map[string][]string)

	for _, e := range entries {
		if e == nil || len(e.Dir) == 0 {
			continue
		}
		moduleNm := e.Name
		log.Infof("Module name(%v)", moduleNm)
		trkTpCnt := true
		dbMapFill("", "", moduleNm, trkTpCnt, xDbSpecMap, e)
	}
}

func childToUpdateParent( xpath string, tableName string) {
	var xpathData *yangXpathInfo
	parent := parentXpathGet(xpath)
	if len(parent) == 0  || parent == "/" {
		return
	}

	_, ok := xYangSpecMap[parent]
	if !ok {
		xpathData = new(yangXpathInfo)
		xYangSpecMap[parent] = xpathData
	}
	xYangSpecMap[parent].childTable = append(xYangSpecMap[parent].childTable, tableName)
	if xYangSpecMap[parent].yangEntry != nil &&
	   xYangSpecMap[parent].yangEntry.Node.Statement().Keyword == "list" {
		return
	}
	childToUpdateParent(parent, tableName)
}

/* Build lookup map based on yang xpath */
func annotEntryFill(xYangSpecMap map[string]*yangXpathInfo, xpath string, entry *yang.Entry) {
	xpathData := new(yangXpathInfo)
	_, ok := xYangSpecMap[xpath]
	if !ok {
		fmt.Printf("Xpath not found(%v) \r\n", xpath)
	}

	xpathData.dbIndex = db.ConfigDB // default value
	/* fill table with yang extension data. */
	if entry != nil && len(entry.Exts) > 0 {
		for _, ext := range entry.Exts {
			dataTagArr := strings.Split(ext.Keyword, ":")
			tagType := dataTagArr[len(dataTagArr)-1]
			switch tagType {
			case "table-name" :
				if xpathData.tableName == nil {
					xpathData.tableName = new(string)
				}
				*xpathData.tableName = ext.NName()
				updateDbTableData(xpath, xpathData, *xpathData.tableName)
				//childToUpdateParent(xpath, *xpathData.tableName)
			case "key-name" :
				if xpathData.keyName == nil {
					xpathData.keyName = new(string)
				}
				*xpathData.keyName = ext.NName()
			case "table-transformer" :
				if xpathData.xfmrTbl == nil {
					xpathData.xfmrTbl = new(string)
				}
				*xpathData.xfmrTbl  = ext.NName()
			case "field-name" :
				xpathData.fieldName = ext.NName()
			case "subtree-transformer" :
				xpathData.xfmrFunc  = ext.NName()
			case "key-transformer" :
				xpathData.xfmrKey   = ext.NName()
			case "key-delimiter" :
				xpathData.delim     = ext.NName()
			case "field-transformer" :
				xpathData.xfmrFunc  = ext.NName()
			case "post-transformer" :
				xpathData.xfmrPost  = ext.NName()
			case "get-validate" :
				xpathData.validateFunc  = ext.NName()
			case "use-self-key" :
				xpathData.keyXpath  = nil
			case "db-name" :
				if ext.NName() == "APPL_DB" {
					xpathData.dbIndex  = db.ApplDB
				} else if ext.NName() == "ASIC_DB" {
					xpathData.dbIndex  = db.AsicDB
				} else if ext.NName() == "COUNTERS_DB" {
					xpathData.dbIndex  = db.CountersDB
				} else if ext.NName() == "LOGLEVEL_DB" {
					xpathData.dbIndex  = db.LogLevelDB
				} else if ext.NName() == "CONFIG_DB" {
					xpathData.dbIndex  = db.ConfigDB
				} else if ext.NName() == "FLEX_COUNTER_DB" {
					xpathData.dbIndex  = db.FlexCounterDB
				} else if ext.NName() == "STATE_DB" {
					xpathData.dbIndex  = db.StateDB
				} else {
					xpathData.dbIndex  = db.ConfigDB
				}
			}
		}
	}
	xYangSpecMap[xpath] = xpathData
}

/* Build xpath from yang-annotation */
func xpathFromDevCreate(path string) string {
	p := strings.Split(path, "/")
	for i, k := range p {
		if len(k) > 0 { p[i] = strings.Split(k, ":")[1] }
	}
	return strings.Join(p[1:], "/")
}

/* Build lookup map based on yang xpath */
func annotToDbMapBuild(annotEntries []*yang.Entry) {
    if annotEntries == nil {
        return
    }
    if xYangSpecMap == nil {
        xYangSpecMap = make(map[string]*yangXpathInfo)
    }

    for _, e := range annotEntries {
        if e != nil && len(e.Deviations) > 0 {
            for _, d := range e.Deviations {
                xpath := xpathFromDevCreate(d.Name)
                xpath = "/" + strings.Replace(e.Name, "-annot", "", -1) + ":" + xpath
                for i, deviate := range d.Deviate {
                    if i == 2 {
                        for _, ye := range deviate {
                            annotEntryFill(xYangSpecMap, xpath, ye)
                        }
                    }
                }
            }
        }
    }
    mapPrint(xYangSpecMap, "/tmp/annotSpec.txt")
}

func annotDbSpecMapFill(xDbSpecMap map[string]*dbInfo, dbXpath string, entry *yang.Entry) error {
	var err error
	var dbXpathData *dbInfo
	var ok bool

	//Currently sonic-yang annotation is supported for "list" type only.
	listName := strings.Split(dbXpath, "/")
	if len(listName) < 3 {
		log.Errorf("Invalid list xpath length(%v) \r\n", dbXpath)
		return err
	}
	dbXpathData, ok = xDbSpecMap[listName[2]]
	if !ok {
		log.Errorf("DB spec-map data not found(%v) \r\n", dbXpath)
		return err
	}
	log.Infof("Annotate dbSpecMap for (%v)(listName:%v)\r\n", dbXpath, listName[2])
	dbXpathData.dbIndex = db.ConfigDB // default value 

	/* fill table with cvl yang extension data. */
	if entry != nil && len(entry.Exts) > 0 {
		for _, ext := range entry.Exts {
			dataTagArr := strings.Split(ext.Keyword, ":")
			tagType := dataTagArr[len(dataTagArr)-1]
			switch tagType {
			case "key-name" :
				if dbXpathData.keyName == nil {
					dbXpathData.keyName = new(string)
				}
				*dbXpathData.keyName = ext.NName()
			case "db-name" :
				if ext.NName() == "APPL_DB" {
					dbXpathData.dbIndex  = db.ApplDB
				} else if ext.NName() == "ASIC_DB" {
					dbXpathData.dbIndex  = db.AsicDB
				} else if ext.NName() == "COUNTERS_DB" {
					dbXpathData.dbIndex  = db.CountersDB
				} else if ext.NName() == "LOGLEVEL_DB" {
					dbXpathData.dbIndex  = db.LogLevelDB
				} else if ext.NName() == "CONFIG_DB" {
					dbXpathData.dbIndex  = db.ConfigDB
				} else if ext.NName() == "FLEX_COUNTER_DB" {
					dbXpathData.dbIndex  = db.FlexCounterDB
				} else if ext.NName() == "STATE_DB" {
					dbXpathData.dbIndex  = db.StateDB
				} else {
					dbXpathData.dbIndex  = db.ConfigDB
				}
			default :
			}
		}
	}

    dbMapPrint("/tmp/dbSpecMapFull.txt")
	return err
}

func annotDbSpecMap(annotEntries []*yang.Entry) {
	if annotEntries == nil || xDbSpecMap == nil {
		return
	}
	for _, e := range annotEntries {
		if e != nil && len(e.Deviations) > 0 {
			for _, d := range e.Deviations {
				xpath := xpathFromDevCreate(d.Name)
				xpath = "/" + strings.Replace(e.Name, "-annot", "", -1) + ":" + xpath
				for i, deviate := range d.Deviate {
					if i == 2 {
						for _, ye := range deviate {
							annotDbSpecMapFill(xDbSpecMap, xpath, ye)
						}
					}
				}
			}
		}
	}
}

/* Debug function to print the yang xpath lookup map */
func mapPrint(inMap map[string]*yangXpathInfo, fileName string) {
    fp, err := os.Create(fileName)
    if err != nil {
        return
    }
    defer fp.Close()

    for k, d := range inMap {
        fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")
        fmt.Fprintf(fp, "%v:\r\n", k)
        fmt.Fprintf(fp, "    yangDataType: %v\r\n", d.yangDataType)
        fmt.Fprintf(fp, "    tableName: ")
        if d.tableName != nil {
            fmt.Fprintf(fp, "%v", *d.tableName)
        }
        fmt.Fprintf(fp, "\r\n    xfmrTbl  : ")
        if d.xfmrTbl != nil {
            fmt.Fprintf(fp, "%v", *d.xfmrTbl)
        }
        fmt.Fprintf(fp, "\r\n    keyName  : ")
        if d.keyName != nil {
            fmt.Fprintf(fp, "%v", *d.keyName)
        }
        fmt.Fprintf(fp, "\r\n    childTbl : %v", d.childTable)
        fmt.Fprintf(fp, "\r\n    FieldName: %v", d.fieldName)
        fmt.Fprintf(fp, "\r\n    keyLevel : %v", d.keyLevel)
        fmt.Fprintf(fp, "\r\n    xfmrKeyFn: %v", d.xfmrKey)
        fmt.Fprintf(fp, "\r\n    xfmrFunc : %v", d.xfmrFunc)
        fmt.Fprintf(fp, "\r\n    dbIndex  : %v", d.dbIndex)
        fmt.Fprintf(fp, "\r\n    validateFunc  : %v", d.validateFunc)
        fmt.Fprintf(fp, "\r\n    yangEntry: ")
        if d.yangEntry != nil {
            fmt.Fprintf(fp, "%v", *d.yangEntry)
        }
        fmt.Fprintf(fp, "\r\n    dbEntry: ")
        if d.dbEntry != nil {
            fmt.Fprintf(fp, "%v", *d.dbEntry)
        }
        fmt.Fprintf(fp, "\r\n    keyXpath: %d\r\n", d.keyXpath)
        for i, kd := range d.keyXpath {
            fmt.Fprintf(fp, "        %d. %#v\r\n", i, kd)
        }
        fmt.Fprintf(fp, "\r\n    isKey   : %v\r\n", d.isKey)
    }
    fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")

}

/* Debug function to print redis db lookup map */
func dbMapPrint( fname string) {
    fp, err := os.Create(fname)
    if err != nil {
        return
    }
    defer fp.Close()
	fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")
    for k, v := range xDbSpecMap {
        fmt.Fprintf(fp, " field:%v \r\n", k)
        fmt.Fprintf(fp, "     type     :%v \r\n", v.fieldType)
        fmt.Fprintf(fp, "     db-type  :%v \r\n", v.dbIndex)
        fmt.Fprintf(fp, "     KeyName: ")
        if v.keyName != nil {
            fmt.Fprintf(fp, "%v", *v.keyName)
        }
        fmt.Fprintf(fp, "\r\n     oc-yang  :%v \r\n", v.yangXpath)
        fmt.Fprintf(fp, "     cvl-yang :%v \r\n", v.dbEntry)
        fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")

    }
}

func updateSchemaOrderedMap(module string, entry *yang.Entry) {
	var children []string
	if entry.Node.Statement().Keyword == "module" {
		for _, dir := range entry.DirOKeys {
			// Gives the yang xpath for the top level container
			xpath := "/" + module + ":" + dir
			_, ok := xYangSpecMap[xpath]
			if ok {
				yentry := xYangSpecMap[xpath].yangEntry
				if yentry.Node.Statement().Keyword == "container" {
					var keyspec = make([]KeySpec, 0)
					keyspec = FillKeySpecs(xpath, "" , &keyspec)
					children = updateChildTable(keyspec, &children)
				}
			}
		}
	}
}

func updateChildTable(keyspec []KeySpec, chlist *[]string) ([]string) {
	for _, ks := range keyspec {
		if (ks.Ts.Name != "") {
			if !contains(*chlist, ks.Ts.Name) {
				*chlist = append(*chlist, ks.Ts.Name)
			}
		}
		*chlist = updateChildTable(ks.Child, chlist)
	}
	return *chlist
}
