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

package server

import (
	"net/http"
	"strings"
	"time"
        "translib"
	"sort"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

// Root directory for UI files
var swaggerUIDir = "./ui"

// SetUIDirectory functions sets directiry where Swagger UI
// resources are maintained.
func SetUIDirectory(directory string) {
	swaggerUIDir = directory
}

// routeTree is the mapping of path prefix to routeNode object.
// It holds REST API route infomation in tree format.
type routeTree map[string]*routeNode

// routeNode is the node is routeTree. It holds handler info
// and child path info for a path prefix.
type routeNode struct {
	path     string
	handlers []handlerInfo
	subpaths routeTree
}

// handlerInfo has one REST API handler info. It holds API name,
// HTTP method name and the handler function.
type handlerInfo struct {
	name    string
	method  string
	handler http.HandlerFunc
}

// getMethods retruns HTTP methods supported for a routeNode.
func (n *routeNode) getMethods() []string {
	var methods []string
	for _, h := range n.handlers {
		methods = append(methods, h.method)
	}

	return methods
}

// add function registers a REST API info into routeTree.
func (t *routeTree) add(prefix, name, method, pattern string, handler http.HandlerFunc) {
	var root, next string
	var lastNode bool

	if k := strings.Index(pattern[1:], "/") + 1; k > 0 {
		root = pattern[0:k]
		next = pattern[k:]
	} else {
		root = pattern
		lastNode = true
	}

	node := (*t)[root]
	if node == nil {
		node = &routeNode{path: prefix + root}
		(*t)[root] = node
	}

	if lastNode {
		h := handlerInfo{name: name, method: method, handler: handler}
		node.handlers = append(node.handlers, h)

	} else {
		if node.subpaths == nil {
			node.subpaths = make(routeTree)
		}

		node.subpaths.add(node.path, name, method, next, handler)
	}
}

// getNode returns the routeNode object for given path in routeTree.
// Returns nil if path is not found.
func (t *routeTree) getNode(path string) *routeNode {
	var root, next string
	if k := strings.Index(path[1:], "/") + 1; k > 0 {
		root = path[0:k]
		next = path[k:]
	} else {
		root = path
	}

	node := (*t)[root]
	if node == nil { // not found
		return nil
	}
	if len(next) == 0 { // found
		return node
	}

	return node.subpaths.getNode(next)
}

// getMethods returns list of REST API HTTP methods supported for
// a given path. Returns nil if path is not valid or no APIs allowed.
func (t *routeTree) getMethods(path string) []string {
	if node := t.getNode(path); node != nil {
		return node.getMethods()
	}
	return nil
}

// muxContext maintains context info during route installation
type muxContext struct {
	minPathsPerRouter int
	prefix            string
	paths             []string
}

// install function fills route information into mux.Router
func (t *routeTree) install(router *mux.Router, mc *muxContext) {
	parentPrefix := mc.prefix
	for pp, node := range *t {
		mc.prefix = parentPrefix + pp
		if len(node.handlers) != 0 {
			mc.paths = append(mc.paths, node.path)
		}

		// Add route for each handler
		for _, h := range node.handlers {
			glog.V(2).Infof("Adding %s, %s %s",
				h.name, h.method, node.path)

			router.
				Methods(h.method).
				Path(mc.prefix).
				//Name(h.name).
				Handler(withMiddleware(h.handler, h.name))
		}

		// Re-use the router for subpaths if there are no methods
		// at this node..
		if len(node.handlers) == 0 {
			node.subpaths.install(router, mc)
			continue
		}

		// Create a sub-router to match child paths
		if node.subpaths != nil {
			subrouter := router.PathPrefix(mc.prefix).Subrouter()
			mc.prefix = "" // subrouter has the prefix
			node.subpaths.install(subrouter, mc)
		}
	}
}

/*
// Print function dumps the routes tree to stdout
func (t *routeTree) Print(indent int) {
	var buff strings.Builder
	for i := 1; i < indent; i++ {
		buff.WriteString(" |  ")
	}
	if indent > 0 {
		buff.WriteString(" +--")
	}

	padding := buff.String()
	indent++

	for pp, node := range *t {
		fmt.Printf("%s%s\n", padding, pp)
		if node.subpaths != nil {
			node.subpaths.Print(indent)
		}
	}
} */

// allRoutes is a collection of all routes
var allRoutes routeTree = make(routeTree)

// numRoutes is the number of routes added via AddRoute() function
var numRoutes uint

// AddRoute appends specified routes to the routes collection.
// Called by init functions of swagger generated router.go files.
func AddRoute(name, method, pattern string, handler http.HandlerFunc) {
	method = strings.ToUpper(method)
	allRoutes.add("", name, method, pattern, handler)
	numRoutes++
}

// NewRouter creates a new http router instance for the REST server.
// Includes all routes registered via AddRoute API as well as few
// in-built routes.
func NewRouter() http.Handler {
	router := NewMuxRouter()
	return withStat(router)
}

// NewMuxRouter creates a new github.com/gorilla/mux router instance
// using the registered routes.
func NewMuxRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true).UseEncodedPath()

	glog.Infof("Server has %d paths", numRoutes)

	// Collect swagger generated route information
	mc := muxContext{paths: make([]string, 0, numRoutes)}
	allRoutes.install(router, &mc)

	// Register common OPTIONS handler for every path template
	// New sub-router is used to avoid extra match during regular APIs
	sr := router.Methods("OPTIONS").Subrouter()
	oh := withMiddleware(http.HandlerFunc(commonOptionsHandler), "optionsHandler")
	for _, p := range mc.paths {
		sr.Path(p).Handler(oh)
	}

	// Documentation and test UI
	uiHandler := http.StripPrefix("/ui/", http.FileServer(http.Dir(swaggerUIDir)))
	router.Methods("GET").PathPrefix("/ui/").Handler(uiHandler)

	// Redirect "/ui" to "/ui/index.html"
	router.Methods("GET").Path("/ui").
		Handler(http.RedirectHandler("/ui/index.html", 301))

	if ClientAuth.Enabled("jwt") {
		router.Methods("POST").Path("/authenticate").Handler(http.HandlerFunc(Authenticate))
		router.Methods("POST").Path("/refresh").Handler(http.HandlerFunc(Refresh))
	}

	// To download yang models
	ydirHandler := http.FileServer(http.Dir(translib.GetYangPath()))
	router.Methods("GET").PathPrefix("/models/yang/").
		Handler(http.StripPrefix("/models/yang/", ydirHandler))

	return router
}

// commonOptionsHandler is the common HTTP OPTIONS method handler
// for all path based routes. Resolves allowed methods for current
// path template by traversing allRoutes cache.
func commonOptionsHandler(w http.ResponseWriter, r *http.Request) {
	var methods []string
	var hasPatch bool
	t, _ := mux.CurrentRoute(r).GetPathTemplate()

	// Collect allowed method names
	methods = allRoutes.getMethods(t)
	for _, m := range methods {
		if m == "PATCH" {
			hasPatch = true
			break
		}
	}

	// "Allow" header
	if len(methods) != 0 {
		methods = append(methods, "OPTIONS") // OPTIONS will not be part of allRoutes
		sort.Strings(methods)
		w.Header().Set("Allow", strings.Join(methods, ", "))
	}

	// "Accept-Patch" header for RESTCONF data paths
	if hasPatch && strings.HasPrefix(t, restconfDataPathPrefix) {
		w.Header().Set("Accept-Patch", mimeYangDataJSON)
	}
}

// loggingMiddleware returns a handler which times and logs the request.
// It should be the top handler in the middleware chain.
func loggingMiddleware(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc, r := GetContext(r)
		rc.Name = name

		glog.Infof("[%s] Recevied %s request from %s", rc.ID, name, r.RemoteAddr)

		start := time.Now()

		inner.ServeHTTP(w, r)

		tt := time.Since(start)
		if rc.stats != nil {
			rc.stats.handlerTime = tt
		}

		glog.Infof("[%s] %s took %s", rc.ID, name, tt)
	})
}

// withMiddleware function prepares the default middleware chain for
// REST APIs.
func withMiddleware(h http.Handler, name string) http.Handler {
	if ClientAuth.Any() {
		h = authMiddleware(h)
	}

	return loggingMiddleware(h, name)
}

// authMiddleware function creates a middleware for request
// authentication and authorization. This middleware will return
// 401 response if authentication fails and 403 if authorization
// fails.
func authMiddleware(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc, r := GetContext(r)
		var err error
		success := false
		if ClientAuth.Enabled("password") {
			err = BasicAuthenAndAuthor(r, rc)
			if err == nil {
				success = true
			}
		}
		if !success && ClientAuth.Enabled("jwt") {
			_, err = JwtAuthenAndAuthor(r, rc)
			if err == nil {
				success = true
			}
		}
		if !success && (ClientAuth.Enabled("cert") || ClientAuth.Enabled("cliuser")) {
			err = ClientCertAuthenAndAuthor(r, rc)
			if err == nil {
				success = true
			}
		}

		if !success {
			status, data, ctype := prepareErrorResponse(err, r)
			w.Header().Set("Content-Type", ctype)
			w.WriteHeader(status)
			w.Write(data)
		} else {
			inner.ServeHTTP(w, r)
		}
	})
}

