////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package translib

import (
	"net"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"translib/db"
	"translib/ocbinds"
	errors "translib/tlerr"
	"translib/transformer"

	"github.com/golang/glog"
	"github.com/openconfig/goyang/pkg/yang"
)

// yanglibApp implements app interface for the
// ietf-yang-library module
type yanglibApp struct {
	pathInfo   *PathInfo
	ygotRoot   *ocbinds.Device
	ygotTarget *interface{}
}

// theYanglibMutex synchronizes all cache loads
var theYanglibMutex sync.Mutex

// theYanglibCache holds parsed yanglib info. Populated on first
// request.
var theYanglibCache *ocbinds.IETFYangLibrary_ModulesState

func init() {
	err := register("/ietf-yang-library:modules-state",
		&appInfo{
			appType:      reflect.TypeOf(yanglibApp{}),
			ygotRootType: reflect.TypeOf(ocbinds.IETFYangLibrary_ModulesState{}),
			isNative:     false,
		})
	if err != nil {
		glog.Fatal("register() failed for yanglibApp;", err)
	}

	err = addModel(&ModelData{
		Name: "ietf-yang-library",
		Org:  "IETF NETCONF (Network Configuration) Working Group",
		Ver:  "2016-06-21",
	})
	if err != nil {
		glog.Fatal("addModel() failed for yanglibApp;", err)
	}
}

/*
 * App interface functions
 */

func (app *yanglibApp) initialize(data appData) {
	app.pathInfo = NewPathInfo(data.path)
	app.ygotRoot = (*data.ygotRoot).(*ocbinds.Device)
	app.ygotTarget = data.ygotTarget
}

func (app *yanglibApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	return nil // NOOP! everyting is in processGet
}

func (app *yanglibApp) translateAction(dbs [db.MaxDB]*db.DB) error {
	return errors.NotSupported("Unsupported")
}

func (app *yanglibApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
	return nil, nil, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) processCreate(d *db.DB) (SetResponse, error) {
	return SetResponse{}, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) processUpdate(d *db.DB) (SetResponse, error) {
	return SetResponse{}, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) processReplace(d *db.DB) (SetResponse, error) {
	return SetResponse{}, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) processDelete(d *db.DB) (SetResponse, error) {
	return SetResponse{}, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) processAction(dbs [db.MaxDB]*db.DB) (ActionResponse, error) {
	return ActionResponse{}, errors.NotSupported("Unsupported")
}

func (app *yanglibApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {
	glog.Infof("path = %s", app.pathInfo.Template)
	glog.Infof("vars = %s", app.pathInfo.Vars)

	var resp GetResponse
	ylib, err := getYanglibInfo()
	if err != nil {
		return resp, err
	}

	switch {
	case app.pathInfo.HasSuffix("/module-set-id"): // only module-set-id
		app.ygotRoot.ModulesState.ModuleSetId = ylib.ModuleSetId

	case app.pathInfo.HasVar("name"): // only one module
		err = app.copyOneModuleInfo(ylib)

	default: // all modules
		app.ygotRoot.ModulesState = ylib
	}

	if err == nil {
		resp.Payload, err = generateGetResponsePayload(
			app.pathInfo.Path, app.ygotRoot, app.ygotTarget)
	}

	return resp, err
}

// copyOneModuleInfo fills one module from given ygot IETFYangLibrary_ModulesState
// object into app.ygotRoot.
func (app *yanglibApp) copyOneModuleInfo(fromMods *ocbinds.IETFYangLibrary_ModulesState) error {
	key := ocbinds.IETFYangLibrary_ModulesState_Module_Key{
		Name: app.pathInfo.Var("name"), Revision: app.pathInfo.Var("revision")}

	to := app.ygotRoot.ModulesState.Module[key]
	from := fromMods.Module[key]
	if from == nil {
		glog.Errorf("No module %s in yanglib", key)
		return errors.NotFound("Module %s@%s not found", key.Name, key.Revision)
	}

	switch pt := app.pathInfo.Template; {
	case strings.HasSuffix(pt, "/deviation"):
		// Copy only deviations.
		if len(from.Deviation) != 0 {
			to.Deviation = from.Deviation
		} else {
			return errors.NotFound("Module %s@%s has no deviations", key.Name, key.Revision)
		}

	case strings.Contains(pt, "/deviation{}{}"):
		// Copy only one deviation info
		devkey := ocbinds.IETFYangLibrary_ModulesState_Module_Deviation_Key{
			Name: app.pathInfo.Var("name#2"), Revision: app.pathInfo.Var("revision#2")}

		if devmod := from.Deviation[devkey]; devmod != nil {
			*to.Deviation[devkey] = *devmod
		} else {
			return errors.NotFound("Module %s@%s has no deviation %s@%s",
				key.Name, key.Revision, devkey.Name, devkey.Revision)
		}

	case strings.HasSuffix(pt, "/submodule"):
		// Copy only submodules..
		if len(from.Submodule) != 0 {
			to.Submodule = from.Submodule
		} else {
			return errors.NotFound("Module %s@%s has no submodules", key.Name, key.Revision)
		}

	case strings.Contains(pt, "/submodule{}{}"):
		// Copy only one submodule info
		subkey := ocbinds.IETFYangLibrary_ModulesState_Module_Submodule_Key{
			Name: app.pathInfo.Var("name#2"), Revision: app.pathInfo.Var("revision#2")}

		if submod := from.Submodule[subkey]; submod != nil {
			*to.Submodule[subkey] = *submod
		} else {
			return errors.NotFound("Module %s@%s has no submodule %s@%s",
				key.Name, key.Revision, subkey.Name, subkey.Revision)
		}

	default:
		// Copy full module
		glog.Infof("Copying module %s@%s", key.Name, key.Revision)
		app.ygotRoot.ModulesState.Module[key] = from
	}

	return nil
}

/*
 * Yang parsing utilities
 */

// yanglibBuilder is the utility for parsing and loading yang files into
// ygot IETFYangLibrary_ModulesState object.
type yanglibBuilder struct {
	// yangDir is the directory with all yang files
	yangDir string

	// baseURL is the base URL for downloading yang files. Yang schema URL
	// can be obtained by appending yang file name to this base URL.
	baseURL string

	// yangModules is the temporary cache of all parsed yang modules.
	// Populated by loadYangs() function.
	yangModules *yang.Modules

	// ygotModules is the output ygot object tree containing all
	// yang module info
	ygotModules *ocbinds.IETFYangLibrary_ModulesState
}

// getYanglibInfo returns the ygot IETFYangLibrary_ModulesState object
// with all yang library information.
func getYanglibInfo() (ylib *ocbinds.IETFYangLibrary_ModulesState, err error) {
	theYanglibMutex.Lock()
	if theYanglibCache == nil {
		glog.Infof("Building yanglib cache")
		theYanglibCache, err = newYanglibInfo()
		glog.Infof("Yanglib cache ready; err=%v", err)
	}

	ylib = theYanglibCache
	theYanglibMutex.Unlock()
	return
}

// newYanglibInfo loads all eligible yangs and fills yanglib info into the
// ygot IETFYangLibrary_ModulesState object
func newYanglibInfo() (*ocbinds.IETFYangLibrary_ModulesState, error) {
	var yb yanglibBuilder
	if err := yb.prepare(); err != nil {
		return nil, err
	}
	if err := yb.loadYangs(); err != nil {
		return nil, err
	}
	if err := yb.translate(); err != nil {
		return nil, err
	}

	return yb.ygotModules, nil
}

// prepare function initializes the yanglibBuilder object for
// parsing yangs and translating into ygot.
func (yb *yanglibBuilder) prepare() error {
	yb.yangDir = GetYangPath()

	// Yang schema URL base will be set only if we can resolve a management IP
	// for this device. Otherwise we wil not advertise the schema URL.
	if ip := findAManagementIP(); ip != "" {
		yb.baseURL = "https://" + ip + "/models/yang/"
	}

	glog.Infof("yanglibBuilder.prepare: yangDir = %s", yb.yangDir)
	glog.Infof("yanglibBuilder.prepare: baseURL = %s", yb.baseURL)

	yb.ygotModules = &ocbinds.IETFYangLibrary_ModulesState{}
	return nil
}

// loadYangs reads eligible yang files into yang.Modules object.
// Skips transformer annotation yangs.
func (yb *yanglibBuilder) loadYangs() error {
	glog.Infof("Loading yangs from %s directory", yb.yangDir)
	mods := yang.NewModules()
	start := time.Now()

	files, _ := filepath.Glob(filepath.Join(yb.yangDir, "*.yang"))
	for _, f := range files {
		if err := mods.Read(f); err != nil {
			glog.Errorf("Failed to parse %s; err=%v", f, err)
			return errors.New("System error")
		}
	}

	glog.Infof("%d yang files loaded in %s", len(files), time.Since(start))
	yb.yangModules = mods
	return nil
}

// translate function fills parsed yang.Modules info into the
// ygot IETFYangLibrary_ModulesState object.
func (yb *yanglibBuilder) translate() error {
	var modsWithDeviation []*yang.Module

	// First iteration -- create ygot module entry for each yang.Module
	for _, mod := range yb.yangModules.Modules {
		m, _ := yb.ygotModules.NewModule(mod.Name, mod.Current())
		if m == nil {
			// ignore; yang.Modules map contains dupicate entries - one for name and
			// other for name@rev. NewModule() will return nil if entry exists.
			continue
		}

		// Fill basic properties into ygot module
		yb.fillModuleInfo(m, mod)

		// Mark the yang.Module with "deviation" statements for 2nd iteration. We need reverse
		// mapping of deviation target -> current module in ygot. Hence 2nd iteration..
		if len(mod.Deviation) != 0 {
			modsWithDeviation = append(modsWithDeviation, mod)
		}
	}

	// 2nd iteration -- fill deviations.
	for _, mod := range modsWithDeviation {
		yb.translateDeviations(mod)
	}

	return nil
}

// fillModuleInfo yang module info from yang.Module to ygot IETFYangLibrary_ModulesState_Module
// object.. Deviation information is not filled.
func (yb *yanglibBuilder) fillModuleInfo(to *ocbinds.IETFYangLibrary_ModulesState_Module, from *yang.Module) {
	to.Namespace = &from.Namespace.Name
	to.Schema = yb.getSchemaURL(from)

	// Fill the "feature" info from yang even though we dont have full
	// support for yang features.
	for _, f := range from.Feature {
		to.Feature = append(to.Feature, f.Name)
	}

	// Iterate thru "include" statements to resolve submodules
	for _, inc := range from.Include {
		submod := yb.yangModules.FindModule(inc)
		if submod == nil { // should not happen
			glog.Errorf("No sub-module %s; @%s", inc.Name, inc.Statement().Location())
			continue
		}

		// NewSubmodule() returns nil if submodule entry already exists.. Ignore it.
		if sm, _ := to.NewSubmodule(submod.Name, submod.Current()); sm != nil {
			sm.Schema = yb.getSchemaURL(submod)
		}
	}

	//TODO is it okay to set "implement" always???
	to.ConformanceType = ocbinds.IETFYangLibrary_ModulesState_Module_ConformanceType_implement
}

// fillModuleDeviation creates a deviation module info in the ygot structure
// for a given main module.
func (yb *yanglibBuilder) fillModuleDeviation(main *yang.Module, deviation *yang.Module) {
	key := ocbinds.IETFYangLibrary_ModulesState_Module_Key{
		Name: main.Name, Revision: main.Current()}

	if m, ok := yb.ygotModules.Module[key]; ok {
		m.NewDeviation(deviation.Name, deviation.Current())
	} else {
		glog.Errorf("Ygot module entry %s not found", key)
	}
}

// translateDeviations function will process all "devaiation" statements of
// a yang.Module and fill deviation info into corresponding ygot module objects.
func (yb *yanglibBuilder) translateDeviations(mod *yang.Module) error {
	deviationTargets := make(map[string]bool)

	// Loop thru deviation statements and find modules deviated by current module
	for _, d := range mod.Deviation {
		if !strings.HasPrefix(d.Name, "/") {
			glog.Errorf("Deviation path \"%s\" is not absolute! @%s", d.Name, d.Statement().Location())
			continue
		}

		// Get prefix of root node from the deviation path. First split the path
		// by "/" char and then split 1st part by ":".
		// Eg, find "acl" from "/acl:scl-sets/config/something"
		root := strings.SplitN(strings.SplitN(d.Name, "/", 3)[1], ":", 2)
		if len(root) != 2 {
			glog.Errorf("Deviation path \"%s\" has no prefix for root element! @%s",
				d.Name, d.Statement().Location())
		} else {
			deviationTargets[root[0]] = true
		}
	}

	glog.V(2).Infof("Module %s has deviations for %d modules", mod.FullName(), len(deviationTargets))

	// Deviation target prefixes must be in the import list.. Find the target
	// modules by matching the prefix in imports.
	for _, imp := range mod.Import {
		prefix := imp.Name
		if imp.Prefix != nil {
			prefix = imp.Prefix.Name
		}
		if !deviationTargets[prefix] {
			continue
		}

		if m := yb.yangModules.FindModule(imp); m != nil {
			yb.fillModuleDeviation(m, mod)
		} else {
			glog.Errorf("No module for prefix \"%s\"", prefix)
		}
	}

	return nil
}

// getSchemaURL resolves the URL for downloading yang file from current
// device. Returns nil if yang URL could not be prepared.
func (yb *yanglibBuilder) getSchemaURL(m *yang.Module) *string {
	if len(yb.baseURL) == 0 {
		return nil // Base URL not resolved; hence no yang URL
	}

	// Ugly hack to get source file name from yang.Module. See implementation
	// of yang.Statement.Location() function.
	// TODO: any better way to get source file path from yang.Module??
	toks := strings.Split(m.Source.Location(), ":")
	if len(toks) != 1 && len(toks) != 3 {
		glog.Warningf("Could not resolve file path for module %s; location=%s",
			m.FullName(), m.Source.Location())
		return nil
	}

	uri := yb.baseURL + filepath.Base(toks[0])
	return &uri
}

/*
 * Other utilities..
 */

// findAManagementIP returns a valid IP address of management interface.
// Empty string is returned if no address could be resolved.
func findAManagementIP() string {
	var addrs []net.Addr
	eth0, err := net.InterfaceByName("eth0")
	if err == nil {
		addrs, err = eth0.Addrs()
	}
	if err != nil {
		glog.Errorf("Could not read eth0 info; err=%v", err)
		return ""
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err == nil && ip.To4() != nil {
			return ip.String()
		}
	}

	glog.Errorf("Could not find a management address!!")
	return ""
}

// GetYangPath returns directory containing yang files. Use
// transformer.YangPath for now.
func GetYangPath() string {
	return transformer.YangPath
}
