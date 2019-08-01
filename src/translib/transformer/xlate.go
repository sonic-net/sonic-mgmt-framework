package transformer

import (
    "fmt"
    //	"os"
    //	"sort"
    //	"github.com/openconfig/goyang/pkg/yang"
    "github.com/openconfig/ygot/ygot"
    "translib/db"
    "translib/ocbinds"
    "reflect"
    "errors"
    log "github.com/golang/glog"
)

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

func XlateToDb(s *ygot.GoStruct, t *interface{}) (map[string]map[string]db.Value, error) {

    var err error

    d := (*s).(*ocbinds.Device)
    jsonStr, err := ygot.EmitJSON(d, &ygot.EmitJSONConfig{
        Format:         ygot.RFC7951,
        Indent:         "  ",
        SkipValidation: true,
        RFC7951Config: &ygot.RFC7951JSONConfig{
            AppendModuleName: true,
        },
    })
    fmt.Println(jsonStr)

    // table.key.fields
    var result = make(map[string]map[string]db.Value)


    // TODO - traverse ygot/JSON with the metadata to translate to DB
    // xfmr method dynamically invoked 

    // use reflect to call the xfmr method from yang extension "key_xfmr", "tanle_xfmr", "field_xfmr"
    //var xfmr keyXfmr



    return result, err
}

func XlateFromDb(data map[string]map[string]db.Value) ([]byte, error) {
    var err error

    // table.key.fields
    var result []byte

    //TODO - implement me
    return result, err
}
