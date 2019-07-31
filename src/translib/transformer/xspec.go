package main
import (
    "fmt"
)

/* Data needed to construct lookup table from yang */
type yangXpathInfo  struct {
    children       *[]string
    yangDataType   string
    tableName      *string
    keyXpath       map[int]*[]string
    delim          *[]string
    fieldName      string
    xfmrFunc       string
}

var xSpecMap map[string]*yangXpathInfo

/* Recursive api to fill the map with yang details */
func fillMap (xSpecMap map[string]*yangXpathInfo, entry *yang.Entry, xpathPrefix string) {

    xpath := xpathPrefix + "/" + entry.Name
    xpathData := new(yangXpathInfo)

    parentXpathData, ok := xSpecMap[xpathPrefix]
    /* init current xpath table data with its parent data, change only if needed. */
    if ok {
        xpathData.tableName = parentXpathData.tableName
    }

    /* fill yang data type i.e. module, container, list, leaf etc. */
    xpathData.yangDataType = entry.Node.Statement().Keyword

    inheritParentKey := true
    /* fill table with yang extension data. */
    if len(entry.Exts) > 0 {
        for _, ext := range entry.Exts {
            dataTagArr := strings.Split(ext.Kind(), ":")
            tagType := dataTagArr[len(dataTagArr)-1]
            switch tagType {
                case "redis-table-name" :
                    xpathData.tableName = new(string)
                    *xpathData.tableName = ext.NName()
                case "redis-field-name" :
                    xpathData.fieldName = ext.NName()
                case "xfmr" :
                    xpathData.xfmrFunc = ext.NName()
                case "use-self-key" :
                    xpathData.keyXpath = nil
                    inheritParentKey = false
            }
        }
    }

    /* fill table with key data. */
    if len(entry.Key) != 0 {
        parentKeyLen := 0

        /* create list with current keys */
        keyXpath  := make([]string, len(strings.Split(entry.Key, " ")))
        for id, keyName := range(strings.Split(entry.Key, " ")) {
            keyXpath[id] = xpath + "/" + keyName
        }
        if inheritParentKey {
            /* init parentKeyLen to the number of parent-key list */
            parentKeyLen = len(parentXpathData.keyXpath)
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

    /* copy children of current obj */
    if len(childList) > 0 {
        children := make([]string, len(entry.Dir))
        copy(children, childList)
        xpathData.children = &children
    }
    /* update yang-map with xpath and table info. */
    xSpecMap[xpath] = xpathData

    /* now recurse, filling the map with current node's children info */
    for _, child := range childList {
        fillMap(xSpecMap, entry.Dir[child], xpath)
    }
}

/* Build lookup hash table based of yang xpath */
func mapBuild(entries map[string]*yang.Entry) {
    if entries == nil {
        return
    }
    xSpecMap = make(map[string]*yangXpathInfo)
    for _, e := range entries {
        if len(e.Dir) == 0 {
            continue
        }

        /* Start to fill xpath based map with yang data */
        fillMap(xSpecMap, e, "")
    }
    mapPrint(xSpecMap)
}

/* Debug function to print the map data into file */
func mapPrint(inMap map[string]*yangXpathInfo) {
    fp, err := os.Create("/tmp/xspec.txt")
        if err != nil {
            return
        }
    defer fp.Close()

    for k, d := range inMap {
        fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")
        fmt.Fprintf(fp, "%v:\r\n", k)
        fmt.Fprintf(fp, "    yangDataType: %v", d.yangDataType)
        fmt.Fprintf(fp, "    tableName: ")
        if d.tableName != nil {
            fmt.Fprintf(fp, "%v", *d.tableName)
        }
        fmt.Fprintf(fp, "\r\n    keyXpath: %d\r\n", d.keyXpath)
        for i, kd := range d.keyXpath {
            fmt.Fprintf(fp, "        %d. %#v\r\n", i, kd)
        }
        fmt.Fprintf(fp, "    children (%p : %v)\r\n", d.children, d.children)
    }
    fmt.Fprintf (fp, "-----------------------------------------------------------------\r\n")

}
