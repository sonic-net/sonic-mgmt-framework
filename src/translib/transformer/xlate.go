package transformer

import (
	"fmt"
//	"os"
//	"sort"
//	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"translib/db"
	"translib/ocbinds"
)

	
	
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