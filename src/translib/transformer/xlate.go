package transformer

import (
    //"fmt"
    //"os"
    //	"sort"
    //	"github.com/openconfig/goyang/pkg/yang"
    "github.com/openconfig/ygot/ygot"
    "translib/db"
    "translib/ocbinds"
    "reflect"
    "errors"
    "strings"
    "encoding/json"
    log "github.com/golang/glog"
)

type KeySpec struct {
    Ts      db.TableSpec
    Key db.Key
    Child *KeySpec
}

var XlateFuncs = make(map[string]reflect.Value)

var (
    ErrParamsNotAdapted = errors.New("The number of params is not adapted.")
)

func XlateFuncBind (name string, fn interface{}) (err error) {
    defer func() {
        if e := recover(); e != nil {
            err = errors.New(name + " is not valid Xfmr function.")
        }
    }()

    if  _, ok := XlateFuncs[name]; !ok {
        v :=reflect.ValueOf(fn)
        v.Type().NumIn()
        XlateFuncs[name] = v
    } else {
        log.Info("Duplicate entry found in the XlateFunc map " + name)
    }
    return
}

func XlateFuncCall(name string, params ... interface{}) (result []reflect.Value, err error) {
    if _, ok := XlateFuncs[name]; !ok {
        err = errors.New(name + " Xfmr function does not exist.")
        return
    }
    if len(params) != XlateFuncs[name].Type().NumIn() {
        err = ErrParamsNotAdapted
        return
    }
    in := make([]reflect.Value, len(params))
    for k, param := range params {
        in[k] = reflect.ValueOf(param)
    }
    result = XlateFuncs[name].Call(in)
    return
}

func TraverseDb(d *db.DB, spec KeySpec, result *map[string]map[string]db.Value, parentKey *db.Key) error {
    var err error

    if spec.Key.Len() > 0 {
        // get an entry with a specific key
        data, err := d.GetEntry(&spec.Ts, spec.Key)
        if err != nil {
            return err
        }

        if (*result)[spec.Ts.Name] == nil {
            (*result)[spec.Ts.Name] = map[string]db.Value{strings.Join(spec.Key.Comp, "|"): data}
        } else {
            (*result)[spec.Ts.Name][strings.Join(spec.Key.Comp, "|")] = data
        }

        if spec.Child != nil {
            err = TraverseDb(d, *spec.Child, result, &spec.Key)
        }
    } else {
        // TODO - GetEntry suuport with regex patten, 'abc*' for optimization
        keys, err := d.GetKeys(&spec.Ts)
        if err != nil {
            return err
        }
        for i, _ := range keys {
            if parentKey != nil {
            // TODO - multi-depth with a custom delimiter
                if strings.Index(strings.Join(keys[i].Comp, "|"), strings.Join((*parentKey).Comp, "|")) == -1 {
                    continue
                }
            }
            spec.Key = keys[i]
            err = TraverseDb(d, spec, result, parentKey)
        }
    }
    return err
}

func XlateUriToKeySpec(path string, uri *ygot.GoStruct, t *interface{}) (*map[db.DBNum][]KeySpec, error) {

    var err error
    var result = make(map[db.DBNum][]KeySpec)
    var retdbFormat = make([]KeySpec, 1)
    var dbFormat KeySpec
    retdbFormat = append(retdbFormat, dbFormat)

    /* Extract the xpath and key from input xpath */
    yangXpath, keyStr := xpathKeyExtract(path);

    fillKeySpec(yangXpath, keyStr, &dbFormat)

    result[db.ConfigDB] = retdbFormat

    // 1 - mock data for a URI /openconfig-acl:acl/acl-sets/acl-set=MyACL1,ACL_IPV4
    //result[db.ConfigDB] = []KeySpec{
    //    {
    //    Ts: db.TableSpec{Name: "ACL_TABLE"},
    //    Key: db.Key{Comp: []string{"MyACL1_ACL_IPV4"}},
    //    Child: &KeySpec{
    //        Ts: db.TableSpec{Name: "ACL_RULE"},
    //        Key: db.Key{}}},
    //    }
        // 2 - mock data for a URI /openconfig-acl:acl/acl-sets/acl-set=MyACL1,ACL_IPV4/acl-entires/ecl-entry=1
//      result[db.ConfigDB] = []KeySpec{
//                      {
//                              Ts: db.TableSpec{Name: "ACL_RULE"},
//                              Key: db.Key{Comp: []string{"MyACL1_ACL_IPV4|RULE_1"}},
//                      }
//              }

        // 3 - mock data for a URI /openconfig-acl:acl
//      result[db.ConfigDB] = []KeySpec{
//                      {
//                              Ts: db.TableSpec{Name: "ACL_TABLE"},
//                              Key: db.Key{},
//                              Child: &KeySpec{
//                                      Ts: db.TableSpec{Name: "ACL_RULE"},
//                                      Key: db.Key{}}},
//                      }

    return &result, err
}

func fillKeySpec(yangXpath string , keyStr string, dbFormat *KeySpec) {

    if xSpecMap == nil {
	return;
    }
    _, ok := xSpecMap[yangXpath]
    if ok {
        xpathInfo := xSpecMap[yangXpath]
        if xpathInfo.tableName != nil {
            dbFormat.Ts.Name = *xpathInfo.tableName
            dbFormat.Key.Comp = append(dbFormat.Key.Comp, keyStr)
        }
        for _, child := range xpathInfo.childTable {
           /* Current support for one child. Should change the KeySpec.Child
              to array of pointers later when we support all children */
	    if xDbSpecMap != nil {
                if  len(xDbSpecMap[child].yangXpath) > 0 {
                    var childXpath = xDbSpecMap[child].yangXpath[0]
                    dbFormat.Child =  new(KeySpec)
                    fillKeySpec(childXpath, "", dbFormat.Child)
                }
            }
        }
    } else {
        return;
    }
}


func XlateToDb(path string, yg *ygot.GoStruct, yt *interface{}) (map[string]map[string]db.Value, error) {

    var err error

    device := (*yg).(*ocbinds.Device)
    jsonStr, err := ygot.EmitJSON(device, &ygot.EmitJSONConfig{
        Format:         ygot.RFC7951,
        Indent:         "  ",
        SkipValidation: true,
        RFC7951Config: &ygot.RFC7951JSONConfig{
            AppendModuleName: true,
        },
    })

    jsonData := make(map[string]interface{})
    err = json.Unmarshal([]byte(jsonStr), &jsonData)
    if err != nil {
	    log.Errorf("Error: failed to unmarshal json.")
        return nil,err
    }

    // table.key.fields
    var result = make(map[string]map[string]db.Value)
    err = dbMapCreate(path, jsonData, result)

    if err != nil {
	    log.Errorf("Error: Data translation from yang to db failed.")
        return result, err
    }

    return result, err
}

func XlateFromDb(data map[string]map[string]db.Value) ([]byte, error) {
    var err error

    // please implement me - data translated by transforme
    // here is a mock data
    payload := `{
    "acl-sets": {
        "acl-set": [
            {
                "name": "MyACL1",
                "type": "ACL_IPV4",
                "config": {
                    "name": "MyACL1",
                    "type": "ACL_IPV4",
                    "description": "Description for MyACL1"
                },
                "acl-entries": {
                    "acl-entry": [
                        {
                            "sequence-id": 1,
                            "config": {
                                "sequence-id": 1,
                                "description": "Description for MyACL1 Rule Seq 1"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "11.1.1.1/32",
                                    "destination-address": "21.1.1.1/32",
                                    "dscp": 1,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 101,
                                    "destination-port": 201
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "ACCEPT"
                                }
                            }
                        },
                        {
                            "sequence-id": 2,
                            "config": {
                                "sequence-id": 2,
                                "description": "Description for MyACL1 Rule Seq 2"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "11.1.1.2/32",
                                    "destination-address": "21.1.1.2/32",
                                    "dscp": 2,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 102,
                                    "destination-port": 202
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "DROP"
                                }
                            }
                        },
                        {
                            "sequence-id": 3,
                            "config": {
                                "sequence-id": 3,
                                "description": "Description for MyACL1 Rule Seq 3"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "11.1.1.3/32",
                                    "destination-address": "21.1.1.3/32",
                                    "dscp": 3,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 103,
                                    "destination-port": 203
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "ACCEPT"
                                }
                            }
                        },
                        {
                            "sequence-id": 4,
                            "config": {
                                "sequence-id": 4,
                                "description": "Description for MyACL1 Rule Seq 4"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "11.1.1.4/32",
                                    "destination-address": "21.1.1.4/32",
                                    "dscp": 4,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 104,
                                    "destination-port": 204
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "DROP"
                                }
                            }
                        },
                        {
                            "sequence-id": 5,
                            "config": {
                                "sequence-id": 5,
                                "description": "Description for MyACL1 Rule Seq 5"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "11.1.1.5/32",
                                    "destination-address": "21.1.1.5/32",
                                    "dscp": 5,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 105,
                                    "destination-port": 205
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "ACCEPT"
                                }
                            }
                        }
                    ]
                }
            },
            {
                "name": "MyACL2",
                "type": "ACL_IPV4",
                "config": {
                    "name": "MyACL2",
                    "type": "ACL_IPV4",
                    "description": "Description for MyACL2"
                },
                "acl-entries": {
                    "acl-entry": [
                        {
                            "sequence-id": 1,
                            "config": {
                                "sequence-id": 1,
                                "description": "Description for Rule Seq 1"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "12.1.1.1/32",
                                    "destination-address": "22.1.1.1/32",
                                    "dscp": 1,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 101,
                                    "destination-port": 201
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "ACCEPT"
                                }
                            }
                        },
                        {
                            "sequence-id": 2,
                            "config": {
                                "sequence-id": 2,
                                "description": "Description for Rule Seq 2"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "12.1.1.2/32",
                                    "destination-address": "22.1.1.2/32",
                                    "dscp": 2,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 102,
                                    "destination-port": 202
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "ACCEPT"
                                }
                            }
                        },
                        {
                            "sequence-id": 3,
                            "config": {
                                "sequence-id": 3,
                                "description": "Description for Rule Seq 3"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "12.1.1.3/32",
                                    "destination-address": "22.1.1.3/32",
                                    "dscp": 3,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 103,
                                    "destination-port": 203
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "ACCEPT"
                                }
                            }
                        },
                        {
                            "sequence-id": 4,
                            "config": {
                                "sequence-id": 4,
                                "description": "Description for Rule Seq 4"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "12.1.1.4/32",
                                    "destination-address": "22.1.1.4/32",
                                    "dscp": 4,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 104,
                                    "destination-port": 204
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "ACCEPT"
                                }
                            }
                        },
                        {
                            "sequence-id": 5,
                            "config": {
                                "sequence-id": 5,
                                "description": "Description for Rule Seq 5"
                            },
                            "ipv4": {
                                "config": {
                                    "source-address": "12.1.1.5/32",
                                    "destination-address": "22.1.1.5/32",
                                    "dscp": 5,
                                    "protocol": "IP_TCP"
                                }
                            },
                            "transport": {
                                "config": {
                                    "source-port": 105,
                                    "destination-port": 205
                                }
                            },
                            "actions": {
                                "config": {
                                    "forwarding-action": "ACCEPT"
                                }
                            }
                        }
                    ]
                }
            }
        ]
    },
    "interfaces": {
        "interface": [
            {
                "id": "Ethernet0",
                "config": {
                    "id": "Ethernet0"
                },
                "interface-ref": {
                    "config": {
                        "interface": "Ethernet0"
                    }
                },
                "ingress-acl-sets": {
                    "ingress-acl-set": [
                        {
                            "set-name": "MyACL1",
                            "type": "ACL_IPV4",
                            "config": {
                                "set-name": "MyACL1",
                                "type": "ACL_IPV4"
                            }
                        }
                    ]
                }
            },
            {
                "id": "Ethernet4",
                "config": {
                    "id": "Ethernet4"
                },
                "interface-ref": {
                    "config": {
                        "interface": "Ethernet4"
                    }
                },
                "ingress-acl-sets": {
                    "ingress-acl-set": [
                        {
                            "set-name": "MyACL2",
                            "type": "ACL_IPV4",
                            "config": {
                                "set-name": "MyACL2",
                                "type": "ACL_IPV4"
                            }
                        }
                    ]
                }
            }
        ]
    }
}`
    
    result := []byte(payload)

    //TODO - implement me
    return result, err

}
