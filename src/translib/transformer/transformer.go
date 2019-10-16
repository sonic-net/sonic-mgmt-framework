package transformer

import (
	"fmt"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"os"
	"strings"
        "bufio"
        "path/filepath"
        "io/ioutil"
)

const YangPath = "/usr/models/yang/" // OpenConfig-*.yang and sonic yang models path

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

func getOcModelsList () ([] string) {
    var fileList [] string
    file, err := os.Open(YangPath + "models_list")
    if err != nil {
        return fileList
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        fileEntry := scanner.Text()
        if strings.HasPrefix(fileEntry, "#") != true {
            _, err := os.Stat(YangPath + fileEntry)
            if err != nil {
                continue
            }
            fileList = append(fileList, fileEntry)
        }
    }
    return fileList
}

func getDefaultModelsList () ([] string) {
    var files []string
    fileInfo, err := ioutil.ReadDir(YangPath)
    if err != nil {
        return files
    }

    for _, file := range fileInfo {
        if strings.HasPrefix(file.Name(), "sonic-") && !strings.HasSuffix(file.Name(), "-dev.yang") &&  filepath.Ext(file.Name()) == ".yang" {
            files = append(files, file.Name())
        }
    }
    return files
}

func init() {
	yangFiles := []string{}
        ocList := getOcModelsList()
        yangFiles = getDefaultModelsList()
        yangFiles = append(yangFiles, ocList...)
        fmt.Println("Yang model List:", yangFiles)
	err := loadYangModules(yangFiles...)
    if err != nil {
	    fmt.Fprintln(os.Stderr, err)
    }
}

func loadYangModules(files ...string) error {

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

	sonic_entries := make([]*yang.Entry, len(names))
	oc_entries := make(map[string]*yang.Entry)
	annot_entries := make([]*yang.Entry, len(names))

	for _, n := range names {
		if strings.Contains(n, "annot") {
			annot_entries = append(annot_entries, yang.ToEntry(mods[n]))
		} else if strings.Contains(n, "sonic") {
			sonic_entries = append(sonic_entries, yang.ToEntry(mods[n]))
		} else if oc_entries[n] == nil {
			oc_entries[n] = yang.ToEntry(mods[n])
		}
	}

	dbMapBuild(sonic_entries)
	annotToDbMapBuild(annot_entries)
	yangToDbMapBuild(oc_entries)

	return err
}
