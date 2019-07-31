package transformer

import (
	"fmt"
//	"os"
//	"sort"
//	"github.com/openconfig/goyang/pkg/yang"
//	"github.com/openconfig/ygot/ygot"
	"translib/db"
//	"translib/ocbinds"
)

func init () {
    XlateFuncBind("acl_set_key_xfmr", acl_set_key_xfmr)
    XlateFuncBind("port_bindings_xfmr", port_bindings_xfmr)
}

func acl_set_key_xfmr(json []byte) (map[string]map[string]db.Value, error) {

	var err error
	
	// table.key.fields
	var result = make(map[string]map[string]db.Value)
	fmt.Println(string(json))
	
	// TODO - traverse JSON with the metadata to translate to DB
	// xfmr method dynamically invoked 
		
	return result, err
}

func port_bindings_xfmr(data map[string]map[string]db.Value) ([]byte, error) {
	var err error
	
	// table.key.fields
	var result []byte
	
	//TODO - implement me
	return result, err
}
