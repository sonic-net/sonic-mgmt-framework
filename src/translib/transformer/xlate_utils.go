package transformer

import (
    "fmt"
    "strings"
    "translib/db"
)

/* Create db key from datd xpath(request) */
func keyFromXpathCreate(keys string) string {
    keyOut := ""
    for i, k := range (strings.Split(keys, " ")) {
        if i > 0 { keyOut += "_" }
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

