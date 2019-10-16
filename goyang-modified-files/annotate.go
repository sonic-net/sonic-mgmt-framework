// Copyright 2015 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

var allimports = make(map[string]string)
var modules = make(map[string]*yang.Module)

func init() {
	register(&formatter{
		name: "annotate",
		f:     genAnnotate,
		utilf: getFile,
		help: "generate template file for yang annotations",
	})
}

// Get the modules for which annotation file needs to be generated
func getFile(files []string, mods map[string]*yang.Module) {
    for _, name := range files {
        slash := strings.Split(name, "/")
        if strings.HasSuffix(name, ".yang") {
            modname := slash[len(slash)-1]
	    modname = strings.TrimSuffix(modname, ".yang");
	    /* Save the yang.Module entries we are interested in */
	    modules[modname] = mods[modname]
        }
    }
}

func genAnnotate(w io.Writer, entries []*yang.Entry) {
    /* Get all the imported modules in the entries */
    GetAllImports(entries)
    for _, e := range entries {
        if _, ok := modules[e.Name]; ok {
            var path string = ""
            generate(w, e, path)
            // { Add closing brace for each module 
            fmt.Fprintln(w, "}")
            fmt.Fprintln(w)
        }
    }
}

// generate writes to stdoutput a template annotation file entry for the selected modules.
func generate(w io.Writer, e *yang.Entry, path string) {
    if e.Parent == nil {
        if e.Name != "" {
            fmt.Fprintf(w, "module %s-annot {\n", e.Name) //}
            fmt.Fprintln(w)
            fmt.Fprintf(w, "    yang-version \"%s\"\n", getYangVersion(e.Name, modules))
            fmt.Fprintln(w)
            fmt.Fprintf(w, "    namespace \"http://openconfig.net/yang/annotation\";\n")
            if e.Prefix != nil {
                fmt.Fprintf(w, "    prefix \"%s-annot\" \n", e.Prefix.Name)
            }
            fmt.Fprintln(w)

	    var imports = make(map[string]string)
            imports = getImportModules(e.Name, modules)
	    for k := range imports {
		if e.Name != k {
                    fmt.Fprintf(w, "    import %s { prefix %s }\n", k, allimports[k])
	        }
            }

            fmt.Fprintln(w)
        }
    }

    name := e.Name
    if e.Prefix != nil {
	name = e.Prefix.Name + ":" + name
    }

    delim := ""
    if path != "" {
	delim = "/"
    }
    path = path + delim + name

    fmt.Fprintf(w, "    deviation %s {\n", path)
    fmt.Fprintf(w, "      deviate add {\n")
    fmt.Fprintf(w, "      }\n")
    fmt.Fprintf(w, "    }\n")
    fmt.Fprintln(w)

    var names []string
    for k := range e.Dir {
	names = append(names, k)
    }

    for _, k := range names {
        generate(w, e.Dir[k], path)
    }

}

// Save to map all imported modules
func GetAllImports(entries []*yang.Entry) {
    for _, e := range entries {
        allimports[e.Name] = e.Prefix.Name
    }
}

//Get Yang version from the yang.Modules
func getYangVersion(modname string, mods map[string]*yang.Module) string {
    if (mods[modname].YangVersion != nil) {
	    return mods[modname].YangVersion.Name
    }
    return ""

}

// Get imported modules for a given module from yang.Module
func getImportModules(modname string, mods map[string]*yang.Module) map[string]string {
    imports := map[string]string{}
    if (mods[modname].Import != nil) {
        for _, imp := range mods[modname].Import {
	    imports[imp.Name] = imp.Prefix.Name
        }
    }
    return imports
}
