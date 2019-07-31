package transformer

import (
	"fmt"
	"os"
	"sort"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
//	"translib/db"
//	"translib/ocbinds"
)

const YangPath = "/usr/models/yang/"   // OpenConfig-*.yang and sonic yang models path

var entries = map[string]*yang.Entry{}

//Interface for xfmr methods
type xfmrInterface interface {
	tableXfmr(s *ygot.GoStruct, t *interface{}) (string, error)
	keyXfmr(s *ygot.GoStruct, t *interface{}) (string, error)
	fieldXfmr(s *ygot.GoStruct, t *interface{}) (string, error)
}

func reportIfError(errs []error) {
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func init() {
    // TODO - define the path for YANG files from container
}

func LoadYangModules(files ...string) error {

	var err error

	paths := []string{YangPath}

	for _, path := range paths {
		expanded, err := yang.PathsWithModules(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		yang.AddPath(expanded...)
	}

	ms := yang.NewModules()

	for _, name := range files {
		if err := ms.Read(name); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}

	// Process the Yang files
	reportIfError(ms.Process())

	// Keep track of the top level modules we read in.
	// Those are the only modules we want to print below.
	mods := map[string]*yang.Module{}
	var names []string

	for _, m := range ms.Modules {
		if mods[m.Name] == nil {
			mods[m.Name] = m
			names = append(names, m.Name)
		}
	}
	sort.Strings(names)
	entries = make(map[string]*yang.Entry)
	for _, n := range names {
		if entries[n] == nil {
			entries[n] = yang.ToEntry(mods[n])
		}
	}
    mapBuild(entries)

	// TODO - build the inverse map for GET, from OC to Sonic

	return err
}
