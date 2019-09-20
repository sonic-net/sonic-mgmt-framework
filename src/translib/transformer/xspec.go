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
    dbIndex        db.DBNum
    keyLevel       int
    isKey          bool
}

type dbInfo  struct {
    fieldType    string
    dbEntry      *yang.Entry
    yangXpath     []string
}

var xSpecMap map[string]*yangXpathInfo
var xDbSpecMap map[string]*dbInfo
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
func yangToDbMapFill (keyLevel int, xSpecMap map[string]*yangXpathInfo, entry *yang.Entry, xpathPrefix string) {
    xpath := ""
    /* create the yang xpath */
    if xSpecMap[xpathPrefix] != nil  && xSpecMap[xpathPrefix].yangDataType == "module" {
        /* module name is separated from the rest of xpath with ":" */
        xpath = xpathPrefix + ":" + entry.Name
    } else {
        xpath = xpathPrefix + "/" + entry.Name
    }

    xpathData, ok := xSpecMap[xpath]
    if !ok {
        xpathData = new(yangXpathInfo)
        xSpecMap[xpath] = xpathData
        xpathData.dbIndex = db.ConfigDB // default value
    } else {
        xpathData = xSpecMap[xpath]
    }

    xpathData.yangDataType = entry.Node.Statement().Keyword
    if entry.Node.Statement().Keyword == "list"  && xpathData.tableName != nil {
        childToUpdateParent(xpath, *xpathData.tableName)
    }

    parentXpathData, ok := xSpecMap[xpathPrefix]
    /* init current xpath table data with its parent data, change only if needed. */
    if ok {
        if xpathData.tableName == nil && parentXpathData.tableName != nil && xpathData.xfmrTbl == nil {
            xpathData.tableName = parentXpathData.tableName
        } else if xpathData.xfmrTbl == nil && parentXpathData.xfmrTbl != nil {
            xpathData.xfmrTbl = parentXpathData.xfmrTbl
        }
    }

    if ok && parentXpathData.dbIndex != db.ConfigDB {
	    xpathData.dbIndex = parentXpathData.dbIndex
    }

    if ok && len(parentXpathData.validateFunc) > 0 {
            xpathData.validateFunc = parentXpathData.validateFunc
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
            xSpecMap[xpath + "/" + keyName] = keyXpathData
            xSpecMap[xpath + "/" + keyName].isKey = true
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
        yangToDbMapFill(curKeyLevel, xSpecMap, entry.Dir[child], xpath)
    }
}

/* Build lookup table based of yang xpath */
func yangToDbMapBuild(entries map[string]*yang.Entry) {
    if entries == nil {
        return
    }

    if xSpecMap == nil {
        xSpecMap = make(map[string]*yangXpathInfo)
    }

    for module, e := range entries {
        if e == nil || len(e.Dir) == 0 {
            continue
        }

	/* Start to fill xpath based map with yang data */
    keyLevel := 0
    yangToDbMapFill(keyLevel, xSpecMap, e, "")

	// Fill the ordered map of child tables list for oc yangs
	updateSchemaOrderedMap(module, e)
    }
    mapPrint(xSpecMap, "/tmp/fullSpec.txt")
    dbMapPrint()
}

/* Fill the map with db details */
func dbMapFill(prefixPath string, curPath string, moduleNm string, trkTpCnt bool, xDbSpecMap map[string]*dbInfo, entry *yang.Entry) {
    entryType := entry.Node.Statement().Keyword
    if entryType == "list" {
        prefixPath = entry.Name
    }

    if !isYangResType(entryType) {
        dbXpath := prefixPath
        if entryType != "list" {
            dbXpath = prefixPath + "/" + entry.Name
        }
        xDbSpecMap[dbXpath] = new(dbInfo)
        xDbSpecMap[dbXpath].dbEntry   = entry
        xDbSpecMap[dbXpath].fieldType = entryType
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
        childPath := prefixPath + "/" + entry.Dir[child].Name
        dbMapFill(prefixPath, childPath, moduleNm, trkTpCnt, xDbSpecMap, entry.Dir[child])
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
        log.Info("Module name", moduleNm)
        xDbSpecOrdTblMap[moduleNm] = []string{}
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

    _, ok := xSpecMap[parent]
    if !ok {
        xpathData = new(yangXpathInfo)
        xSpecMap[parent] = xpathData
    }
    xSpecMap[parent].childTable = append(xSpecMap[parent].childTable, tableName)
    if xSpecMap[parent].yangEntry != nil && xSpecMap[parent].yangEntry.Node.Statement().Keyword == "list" {
        return
    }
    childToUpdateParent(parent, tableName)
}

/* Build lookup map based on yang xpath */
func annotEntryFill(xSpecMap map[string]*yangXpathInfo, xpath string, entry *yang.Entry) {
    xpathData := new(yangXpathInfo)
    _, ok := xSpecMap[xpath]
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
                case "redis-db-name" :
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
    xSpecMap[xpath] = xpathData
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
    if xSpecMap == nil {
        xSpecMap = make(map[string]*yangXpathInfo)
    }

    for _, e := range annotEntries {
        if e != nil && len(e.Deviations) > 0 {
            for _, d := range e.Deviations {
                xpath := xpathFromDevCreate(d.Name)
                xpath = "/" + strings.Replace(e.Name, "-annot", "", -1) + ":" + xpath
                for i, deviate := range d.Deviate {
                    if i == 2 {
                        for _, ye := range deviate {
                            annotEntryFill(xSpecMap, xpath, ye)
                        }
                    }
                }
            }
        }
    }
    mapPrint(xSpecMap, "/tmp/annotSpec.txt")
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
func dbMapPrint() {
    fp, err := os.Create("/tmp/dbTmplt.txt")
    if err != nil {
        return
    }
    defer fp.Close()
	fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")
    for k, v := range xDbSpecMap {
        fmt.Fprintf(fp, " field:%v \r\n", k)
        fmt.Fprintf(fp, "     type :%v \r\n", v.fieldType)
        fmt.Fprintf(fp, "     Yang :%v \r\n", v.yangXpath)
        fmt.Fprintf(fp, "     DB   :%v \r\n", v.dbEntry)
        fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")

    }
}

func updateSchemaOrderedMap(module string, entry *yang.Entry) {
    var children []string
    if entry.Node.Statement().Keyword == "module" {
        for _, dir := range entry.DirOKeys {
            // Gives the yang xpath for the top level container
            xpath := "/" + module + ":" + dir
            _, ok := xSpecMap[xpath]
            if ok {
		yentry := xSpecMap[xpath].yangEntry
		if yentry.Node.Statement().Keyword == "container" {
                    var keyspec = make([]KeySpec, 0)
                    keyspec = FillKeySpecs(xpath, "" , &keyspec)
		    children = updateChildTable(keyspec, &children)
                }
            }
	}
    }
    xDbSpecOrdTblMap[module] = children
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
