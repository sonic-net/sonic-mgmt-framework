package transformer

import (
    "fmt"
    "os"
    "sort"
    "strings"

    "github.com/openconfig/goyang/pkg/yang"
)

/* Data needed to construct lookup table from yang */
type yangXpathInfo  struct {
    yangDataType   string
    tableName      *string
    childTable      []string
    dbEntry        *yang.Entry
    yangEntry      *yang.Entry
    keyXpath       map[int]*[]string
    delim          string
    fieldName      string
    xfmrFunc       string
    xfmrKey        string
}

var xSpecMap map[string]*yangXpathInfo
var xDbSpecMap map[string]*yang.Entry

/* update transformer spec with db-node */
func updateDbTableData (xpathData *yangXpathInfo, tableName string) {
    _, ok := xDbSpecMap[tableName]
    if ok {
        xpathData.dbEntry = xDbSpecMap[tableName]
    }
}

/* Recursive api to fill the map with yang details */
func yangToDbMapFill (xSpecMap map[string]*yangXpathInfo, entry *yang.Entry, xpathPrefix string) {
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
    } else {
        xpathData = xSpecMap[xpath]
    }
    xpathData.yangDataType = entry.Node.Statement().Keyword

    parentXpathData, ok := xSpecMap[xpathPrefix]
    /* init current xpath table data with its parent data, change only if needed. */
    if ok && xpathData.tableName == nil && parentXpathData.tableName != nil {
        xpathData.tableName = parentXpathData.tableName
    }

    if xpathData.yangDataType == "leaf" && len(xpathData.fieldName) == 0 {
        if xpathData.tableName != nil && xDbSpecMap[*xpathData.tableName] != nil &&
           (xDbSpecMap[*xpathData.tableName].Dir[entry.Name] != nil ||
            xDbSpecMap[*xpathData.tableName].Dir[strings.ToUpper(entry.Name)] != nil) {
                xpathData.fieldName = strings.ToUpper(entry.Name)
       }
    }

    /* fill table with key data. */
    if len(entry.Key) != 0 {
        parentKeyLen := 0

        /* create list with current keys */
        keyXpath      := make([]string, len(strings.Split(entry.Key, " ")))
        for id, keyName := range(strings.Split(entry.Key, " ")) {
            keyXpath[id] = xpath + "/" + keyName
        }
        xpathData.keyXpath = make(map[int]*[]string, (parentKeyLen + 1))
        k := 0
        for ; k < parentKeyLen; k++ {
            /* copy parent key-list to child key-list*/
            xpathData.keyXpath[k] = parentXpathData.keyXpath[k]
        }
        xpathData.keyXpath[k] = &keyXpath
    } else if parentXpathData != nil && parentXpathData.keyXpath != nil {
        xpathData.keyXpath = parentXpathData.keyXpath
    }

    /* get current obj's children */
    var childList []string
    for k := range entry.Dir {
        childList = append(childList, k)
    }

    sort.Strings(childList)
    xpathData.yangEntry = entry
    /* now recurse, filling the map with current node's children info */
    for _, child := range childList {
        yangToDbMapFill(xSpecMap, entry.Dir[child], xpath)
    }
}

/* Build lookup table based of yang xpath */
func yangToDbMapBuild(entries map[string]*yang.Entry) {
    if entries == nil {
        return
    }
    for _, e := range entries {
        if e == nil || len(e.Dir) == 0 {
            continue
        }

        /* Start to fill xpath based map with yang data */
        yangToDbMapFill(xSpecMap, e, "")
    }
    mapPrint(xSpecMap, "/tmp/fullSpec.txt")
}

/* Fill the map with db details */
func dbMapFill(xDbSpecMap map[string]*yang.Entry, entry *yang.Entry) {
    entryType := entry.Node.Statement().Keyword
    if entryType == "list" {
        xDbSpecMap[entry.Name] = entry
    }

    var childList []string
    for k := range entry.Dir {
        childList = append(childList, k)
    }
    sort.Strings(childList)
    for _, child := range childList {
        dbMapFill(xDbSpecMap, entry.Dir[child])
    }
}

/* Build redis db lookup map */
func dbMapBuild(entries []*yang.Entry) {
    if entries == nil {
        return
    }
    xDbSpecMap = make(map[string]*yang.Entry)

    for _, e := range entries {
        if e == nil || len(e.Dir) == 0 {
            continue
        }
        dbMapFill(xDbSpecMap, e)
    }
    dbMapPrint(xSpecMap)
}

func childToUpdateParent( xpath string, tableName string) {
    var xpathData *yangXpathInfo
    parent := parentXpathGet(xpath)
    if len(parent) == 0  || parent == "/" {
        return
    }

    fmt.Printf(" Parent Table: %v\r\n", parent)
    _, ok := xSpecMap[parent]
    if !ok {
        xpathData = new(yangXpathInfo)
        xSpecMap[parent] = xpathData
    }
    xSpecMap[parent].childTable = append(xSpecMap[parent].childTable, tableName)
    childToUpdateParent(parent, tableName)
}

/* Build lookup map based on yang xpath */
func annotEntryFill(xSpecMap map[string]*yangXpathInfo, xpath string, entry *yang.Entry) {
    xpathData := new(yangXpathInfo)
    _, ok := xSpecMap[xpath]
    if !ok {
        fmt.Printf("Xpath not found(%v) \r\n", xpath)
    }

    /* fill table with yang extension data. */
    if len(entry.Exts) > 0 {
        for _, ext := range entry.Exts {
            dataTagArr := strings.Split(ext.Keyword, ":")
            tagType := dataTagArr[len(dataTagArr)-1]
            switch tagType {
                case "table-name" :
                    if xpathData.tableName == nil {
                        xpathData.tableName = new(string)
                    }
                    *xpathData.tableName = ext.NName()
                    updateDbTableData(xpathData, *xpathData.tableName)
					childToUpdateParent(xpath, *xpathData.tableName)
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
                    xpathData.xfmrFunc  = ext.NName()
                case "use-self-key" :
                    xpathData.keyXpath  = nil
            }
        }
    }
    xSpecMap[xpath] = xpathData
}

/* Build xpath from yang-annotation */
func xpathFromDevCreate(path string) string {
    xpath := ""
    for _, k := range (strings.Split(path, "/")) {
        if len(k) > 0 {
            xpath += strings.Split(k, ":")[1] + "/"
        }
    }
    return xpath[:len(xpath)-1]
}

/* Build lookup map based on yang xpath */
func annotToDbMapBuild(annotEntries []*yang.Entry) {
    if annotEntries == nil {
        return
    }

    xSpecMap = make(map[string]*yangXpathInfo)
    for _, e := range annotEntries {
        if e != nil && len(e.Deviations) > 0 {
            for _, d := range e.Deviations {
                xpath := xpathFromDevCreate(d.Name)
                xpath = "/" + strings.Replace(e.Name, "-annot", "", -1) + ":" + xpath
                for i, deviate := range d.Deviate {
                    if i == 2 {
                        for _, ye := range deviate {
                            fmt.Println(ye.Name)
                            fmt.Printf(" Annot fill:(%v)\r\n", xpath)
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
        fmt.Fprintf(fp, "\r\n    FieldName: ")
        if len(d.fieldName) > 0 {
            fmt.Fprintf(fp, "%v", d.fieldName)
        }
        fmt.Fprintf(fp, "\r\n    keyXfmr: ")
        if d.dbEntry != nil {
            fmt.Fprintf(fp, "%v\r\n", d.xfmrKey)
        }
        fmt.Fprintf(fp, "\r\n    SubTreeXfmr: ")
        if d.dbEntry != nil {
            fmt.Fprintf(fp, "%v\r\n", d.xfmrFunc)
        }
        fmt.Fprintf(fp, "\r\n    yangEntry: ")
        if d.dbEntry != nil {
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
    }
    fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")

}

/* Debug function to print redis db lookup map */
func dbMapPrint(inMap map[string]*yangXpathInfo) {
    fp, err := os.Create("/tmp/dbTmplt.txt")
    if err != nil {
        return
    }
    defer fp.Close()
    fmt.Fprintf(fp, "%v", xDbSpecMap)
}

