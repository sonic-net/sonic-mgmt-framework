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
var allmodules = make(map[string]*yang.Module)

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
    allmodules = mods
    for _, name := range files {
        slash := strings.Split(name, "/")
            modname := slash[len(slash)-1]
	    modname = strings.TrimSuffix(modname, ".yang");
	    /* Save the yang.Module entries we are interested in */
	    modules[modname] = mods[modname]
    }
}

func genAnnotate(w io.Writer, entries []*yang.Entry) {
    /* Get all the imported modules in the entries */
    GetAllImports(entries)
    for _, e := range entries {
        if _, ok := modules[e.Name]; ok {
            var path string = ""
            var prefix string = ""
            generate(w, e, path, prefix)
            // { Add closing brace for each module 
            fmt.Fprintln(w, "}")
            fmt.Fprintln(w)
        }
    }
}

// generate writes to stdoutput a template annotation file entry for the selected modules.
func generate(w io.Writer, e *yang.Entry, path string, prefix string) {
    if e.Parent == nil {
        if e.Name != "" {
            fmt.Fprintf(w, "module %s-annot {\n", e.Name) //}
            fmt.Fprintln(w)
            fmt.Fprintf(w, "    yang-version \"%s\";\n", getYangVersion(e.Name, modules))
            fmt.Fprintln(w)
            fmt.Fprintf(w, "    namespace \"http://openconfig.net/yang/annotation/%s-annot\";\n", e.Prefix.Name)
            if e.Prefix != nil {
                fmt.Fprintf(w, "    prefix \"%s-annot\";\n", e.Prefix.Name)
            }
            fmt.Fprintln(w)

	    var imports = make(map[string]string)
            imports = getImportModules(e.Name, modules)
	    for k := range imports {
		if e.Name != k {
                    fmt.Fprintf(w, "    import %s { prefix %s; }\n", k, allimports[k])
	        }
            }
	    // Include the module for which annotation is being generated
            fmt.Fprintf(w, "    import %s { prefix %s; }\n", e.Name, e.Prefix.Name)

            fmt.Fprintln(w)
        }
    }

    name := e.Name
    if prefix == "" && e.Prefix != nil {
	prefix = e.Prefix.Name
    }
    name = prefix + ":" + name

    if (e.Node.Kind() != "module") {
        path = path + "/" + name
        printDeviation(w, path)
    }

    var names []string
    for k := range e.Dir {
	names = append(names, k)
    }

    if (e.Node.Kind() == "module") {
	    if len(e.Node.(*yang.Module).Augment) > 0 {
		    for _,a := range e.Node.(*yang.Module).Augment {
			    pathList := strings.Split(a.Name, "/")
			    pathList = pathList[1:]
			    for i, pvar := range pathList {
				    if len(pvar) > 0 && !strings.Contains(pvar, ":") {
					    pvar = e.Prefix.Name + ":" + pvar
					    pathList[i] = pvar
				    }
			    }
			    path = "/" + strings.Join(pathList, "/")
			    handleAugments(w, a, e.Node.(*yang.Module).Grouping, e.Prefix.Name, path)
		    }
	    }
    }

    for _, k := range names {
        generate(w, e.Dir[k], path, prefix)
    }

}

func printDeviation(w io.Writer, path string){
    fmt.Fprintf(w, "    deviation %s {\n", path)
    fmt.Fprintf(w, "      deviate add {\n")
    fmt.Fprintf(w, "      }\n")
    fmt.Fprintf(w, "    }\n")
    fmt.Fprintln(w)
}


// Save to map all imported modules
func GetAllImports(entries []*yang.Entry) {
    for _, e := range entries {
        allimports[e.Name] = e.Prefix.Name
    }
}

func GetModuleFromPrefix(prefix string) string {
    for m, p := range allimports {
	    if prefix == p {
		    return m
	    }
    }
    return ""
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

func handleAugments(w io.Writer, a *yang.Augment, grp []*yang.Grouping, prefix string, path string) {
    for _, u := range a.Uses {
	    grpN := u.Name
	    for _, g := range grp {
		if grpN == g.Name {
		    if len(g.Container) > 0 {
                        handleContainer(w, g.Container, grp, prefix, path)
		    }
		    if len(g.List) > 0 {
                        handleList(w, g.List, grp, prefix, path)
		    }
		    if len(g.LeafList) > 0 {
                        handleLeafList(w, g.LeafList, prefix, path)
		    }
		    if len(g.Leaf) > 0 {
                        handleLeaf(w, g.Leaf, prefix, path)
		    }
		    if len(g.Choice) > 0 {
                        handleChoice(w, g.Choice, grp, prefix, path)
		    }
		    if len(g.Uses) > 0 {
                        handleUses(w, g.Uses, grp, prefix, path)
		    }
		}
	    }
    }

}

func handleUses(w io.Writer, u []*yang.Uses, grp []*yang.Grouping, prefix string, path string) {
    for _, u := range u {
            grpN := u.Name
	    if  strings.Contains(grpN, ":") {
	        tokens := strings.Split(grpN, ":")
		nprefix := tokens[0]
		grpN = tokens[1]
		mod := GetModuleFromPrefix(nprefix)
	        grp = allmodules[mod].Grouping
	    }
            for _, g := range grp {
                if grpN == g.Name {
                    if len(g.Container) > 0 {
                        handleContainer(w, g.Container, grp, prefix, path)
                    }
                    if len(g.List) > 0 {
                        handleList(w, g.List, grp, prefix, path)
                    }
                    if len(g.LeafList) > 0 {
                        handleLeafList(w, g.LeafList, prefix, path)
                    }
                    if len(g.Leaf) > 0 {
                        handleLeaf(w, g.Leaf, prefix, path)
                    }
                    if len(g.Choice) > 0 {
                        handleChoice(w, g.Choice, grp, prefix, path)
                    }
                    if len(g.Uses) > 0 {
                        handleUses(w, g.Uses, grp, prefix, path)
                    }

                }
            }
    }

}

func handleContainer(w io.Writer, ctr []*yang.Container, grp []*yang.Grouping, prefix string, path string) {
	for _, c := range ctr {
		npath := path + "/" + prefix + ":" + c.Name
		printDeviation(w, npath)
		if len(c.Container) > 0 {
			handleContainer(w, c.Container, grp, prefix, npath)
		}
		if len(c.List) > 0 {
			handleList(w, c.List, grp, prefix, npath)
		}
		if len(c.LeafList) > 0 {
			handleLeafList(w, c.LeafList, prefix, npath)
		}
		if len(c.Leaf) > 0 {
			handleLeaf(w, c.Leaf, prefix, npath)
		}
		if len(c.Choice) > 0 {
			handleChoice(w, c.Choice, grp, prefix, npath)
		}
		if len(c.Grouping) > 0 {
			handleGrouping(w, c.Grouping, grp, prefix, npath)
		}
		if len(c.Uses) > 0 {
			handleUses(w, c.Uses, grp, prefix, npath)
		}
	}
}

func handleList(w io.Writer, lst []*yang.List, grp []*yang.Grouping, prefix string, path string) {
        for _, l := range lst {
		npath := path + "/" + prefix + ":" + l.Name
                printDeviation(w, npath)
                if len(l.Container) > 0 {
                        handleContainer(w, l.Container, grp, prefix, npath)
                }
                if len(l.List) > 0 {
                        handleList(w, l.List, grp, prefix, npath)
                }
                if len(l.LeafList) > 0 {
                        handleLeafList(w, l.LeafList, prefix, npath)
                }
                if len(l.Leaf) > 0 {
                        handleLeaf(w, l.Leaf, prefix, npath)
                }
                if len(l.Choice) > 0 {
                        handleChoice(w, l.Choice, grp, prefix, npath)
                }
                if len(l.Grouping) > 0 {
                        handleGrouping(w, l.Grouping, grp, prefix, npath)
                }
                if len(l.Uses) > 0 {
                        handleUses(w, l.Uses, grp, prefix, npath)
                }

        }
}

func handleGrouping(w io.Writer, grp []*yang.Grouping, grptop []*yang.Grouping, prefix string, path string) {
        for _, g := range grp {
		npath := path + "/" + prefix + ":" + g.Name
                printDeviation(w, npath)
                if len(g.Container) > 0 {
                        handleContainer(w, g.Container, grptop, prefix, npath)
                }
                if len(g.List) > 0 {
                        handleList(w, g.List, grptop, prefix, npath)
                }
                if len(g.LeafList) > 0 {
                        handleLeafList(w, g.LeafList, prefix, npath)
                }
                if len(g.Leaf) > 0 {
                        handleLeaf(w, g.Leaf, prefix, npath)
                }
                if len(g.Choice) > 0 {
                        handleChoice(w, g.Choice, grptop, prefix, npath)
                }
                if len(g.Grouping) > 0 {
                        handleGrouping(w, g.Grouping, grptop, prefix, npath)
                }
                if len(g.Uses) > 0 {
                        handleUses(w, g.Uses, grptop, prefix, npath)
                }

        }
}

func handleLeaf (w io.Writer, lf []*yang.Leaf, prefix string, path string) {
	if len(lf) > 0 {
		for _, l := range lf {
			npath := path + "/" + prefix + ":" + l.Name
			printDeviation(w, npath)
		}
	}

}

func handleLeafList (w io.Writer, llst []*yang.LeafList, prefix string, path string) {
	if len(llst) > 0 {
		for _, l := range llst {
			npath := path + "/" + prefix + ":" + l.Name
			printDeviation(w, npath)
		}
	}
}

func handleChoice (w io.Writer, ch []*yang.Choice, grp []*yang.Grouping, prefix string, path string) {
        for _, c := range ch {
		npath := path + "/" + prefix + ":" + c.Name
                printDeviation(w, npath)
                if len(c.Container) > 0 {
                        handleContainer(w, c.Container, grp, prefix, npath)
                }
                if len(c.List) > 0 {
                        handleList(w, c.List, grp, prefix, npath)
                }
                if len(c.LeafList) > 0 {
                        handleLeafList(w, c.LeafList, prefix, npath)
                }
                if len(c.Leaf) > 0 {
                        handleLeaf(w, c.Leaf, prefix, npath)
		}
		if len(c.Case) > 0 {
			handleCase(w, c.Case, grp, prefix, npath)
		}
	}
}

func handleCase (w io.Writer, ch []*yang.Case, grp []*yang.Grouping, prefix string, path string) {
        for _, c := range ch {
		npath := path + "/" + prefix + ":" + c.Name
                printDeviation(w, npath)
                if len(c.Container) > 0 {
                        handleContainer(w, c.Container, grp, prefix, npath)
                }
                if len(c.List) > 0 {
                        handleList(w, c.List, grp, prefix, npath)
                }
                if len(c.LeafList) > 0 {
                        handleLeafList(w, c.LeafList, prefix, npath)
                }
                if len(c.Leaf) > 0 {
                        handleLeaf(w, c.Leaf, prefix, npath)
                }
                if len(c.Choice) > 0 {
                        handleChoice(w, c.Choice, grp, prefix, npath)
                }
                if len(c.Uses) > 0 {
                        handleUses(w, c.Uses, grp, prefix, npath)
                }

        }
}

