package transformer

import (
    "fmt"
    "strings"
    "translib/db"
    "github.com/openconfig/goyang/pkg/yang"
    "github.com/openconfig/gnmi/proto/gnmi"
    "github.com/openconfig/ygot/ygot"
    log "github.com/golang/glog"
)

/* Create db key from datd xpath(request) */
func keyFromXpathCreate(keyList []string) string {
    keyOut := ""
    for i, k := range keyList {
        if i > 0 { keyOut += "_" }
        if strings.Contains(k, ":") {
            k = strings.Split(k, ":")[1]
        }
        keyOut += strings.Split(k, "=")[1]
    }
    return keyOut
}

/* Create db key from datd xpath(request) */
func keyCreate(keyPrefix string, xpath string, data interface{}) string {
    yangEntry := xSpecMap[xpath].yangEntry
    if len(keyPrefix) > 0 { keyPrefix += "|" }

    keyVal := ""
    for i, k := range (strings.Split(yangEntry.Key, " ")) {
        if i > 0 { keyVal = keyVal + "_" }
        keyVal += fmt.Sprintf("%v", data.(map[string]interface{})[k])
    }
    keyPrefix += string(keyVal)
    return keyPrefix
}

/* Copy redis-db source to destn map */
func mapCopy(destnMap map[string]map[string]db.Value, srcMap map[string]map[string]db.Value) {
   for table, tableData := range srcMap {
        _, ok := destnMap[table]
        if !ok {
            destnMap[table] = make(map[string]db.Value)
        }
        for rule, ruleData := range tableData {
            _, ok = destnMap[table][rule]
            if !ok {
                 destnMap[table][rule] = db.Value{Field: make(map[string]string)}
            }
            for field, value := range ruleData.Field {
                destnMap[table][rule].Field[field] = value
            }
        }
   }
}

func parentXpathGet(xpath string) string {
    path := ""
    if len(xpath) > 0 {
		p := strings.Split(xpath, "/")
		path = strings.Join(p[:len(p)-1], "/")
	}
    return path
}

func isYangResType(ytype string) bool {
    if ytype == "choose" || ytype == "case" {
        return true
    }
    return false
}

func yangTypeGet(entry *yang.Entry) string {
    if entry != nil && entry.Node != nil {
        return entry.Node.Statement().Keyword
    }
    return ""
}

func dbKeyToYangDataConvert(dbKey string, xpath string) {
    return
}

func contains(sl []string, str string) bool {
    for _, v := range sl {
        if v == str {
            return true
        }
    }
    return false
}


func isSubtreeRequest(targetUriPath string, nodePath string) bool {
    return strings.HasPrefix(targetUriPath, nodePath)
}

func getYangPathFromUri(uri string) (string, error) {
    var path *gnmi.Path
    var err error

    path, err = ygot.StringToPath(uri, ygot.StructuredPath, ygot.StringSlicePath)
    if err != nil {
        log.Errorf("Error in uri to path conversion: %v", err)
        return "", err
    }

    yangPath, yperr := ygot.PathToSchemaPath(path)
    if yperr != nil {
        log.Errorf("Error in Gnmi path to Yang path conversion: %v", yperr)
        return "", yperr
    }

    return yangPath, err
}

