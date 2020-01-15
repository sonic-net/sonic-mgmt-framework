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
	"flag"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

  "translib"
	"sort"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

// Root directory for UI files
var swaggerUIDir = "./rest_ui"

func init() {
	flag.StringVar(&swaggerUIDir, "ui", "/rest_ui", "UI directory")
}

// Router dispatches http request to corresponding handlers.
type Router struct {
	// rcRoutes holds the RESTCONF routes tree. It matches paths
	// starting with "/restconf/" only.
	rcRoutes routeTree

	// muxRoutes holds all non-RESTCONF routes; including OpenAPI
	// defined routes and internal routes (like UI, yang download).
	muxRoutes *mux.Router

	// optionsHandler is the common handler for OPTIONS requests.
	optionsHandler http.Handler
}

// ServeHTTP resolves and invokes the handler for http request r.
// RESTCONF paths are served from the routeTree; rest from mux router.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := cleanPath(r.URL.EscapedPath())
	if isServeFromTree(path) {
		router.serveFromTree(path, r, w)
	} else {
		router.muxRoutes.ServeHTTP(w, r)
	}
}

// serveFromTree finds and invokes the handler for a path from routeTree
func (router *Router) serveFromTree(path string, r *http.Request, w http.ResponseWriter) {
	var routeInfo routeMatchInfo
	node := router.rcRoutes.match(path, &routeInfo)

	// Node not found..
	if node == nil {
		notFound(w, r)
		return
	}

	handler := node.handlers[r.Method]
	if handler == nil && r.Method == "OPTIONS" {
		handler = router.optionsHandler
	}

	// Node found, but no handler for the method
	if handler == nil {
		notAllowed(w, r)
		return
	}

	// Set route match info in context
	rc, r := GetContext(r)
	rc.route = &routeInfo

	handler.ServeHTTP(w, r)
}

// cleanPath returns the canonical path for p
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}

	return path.Clean(p)
}

// isServeFromTree checks if a path will be served from routeTree.
func isServeFromTree(path string) bool {
	return strings.HasPrefix(path, restconfPathPrefix)
}

// routeMatchInfo holds the matched route information in
// request context.
type routeMatchInfo struct {
	path string
	vars map[string]string
	node *routeNode // only valid for tree based matches
}

// routeTree is the mapping of path prefix to routeNode object.
// It holds REST API route infomation in tree format.
type routeTree map[string]*routeNode

// routeNode is the node in routeTree. Each node represents one
// element in the path. Eg, path "/one/two={id},{name}" has 2
// nodes - "/one" and "/two={id},{name}".
//
// routeNode holds handlers for current path and a routeTree
// for child paths.
type routeNode struct {
	path   string   // full path from root
	name   string   // node name
	params []string // path params for this node, in order

	handlers map[string]http.Handler // method to handler mapping
	subpaths routeTree               // sub paths
}

// add function registers a REST API info into routeTree.
func (t *routeTree) add(parentPrefix, path string, rr *routeRegInfo) {
	root, next := pathSplit(path)
	node := (*t)[root]
	if node == nil {
		node = newRouteNode(root, parentPrefix)
		if node == nil {
			glog.Errorf("Failed to parse path node '%s'", root)
			glog.Errorf("Ignoring route %s, %s %s", rr.name, rr.method, rr.path)
			return
		}

		(*t)[root] = node
	}

	if len(next) == 0 {
		// target node found, register handler
		if node.handlers == nil {
			node.handlers = make(map[string]http.Handler)
		}

		node.handlers[rr.method] = withMiddleware(rr.handler, rr.name)

	} else {
		if node.subpaths == nil {
			node.subpaths = make(routeTree)
		}

		node.subpaths.add(node.path, next, rr)
	}
}

// getNode returns the routeNode object for given path in routeTree.
// Returns nil if path is not found.
func (t *routeTree) getNode(path string) *routeNode {
	root, next := pathSplit(path)
	node := (*t)[root]
	if node == nil || len(next) == 0 {
		return node
	}

	return node.subpaths.getNode(next)
}

// match resolves the routeNode for a request path.
func (t *routeTree) match(path string, m *routeMatchInfo) *routeNode {
	var matchedNode *routeNode
	root, next := pathSplit(path)
	name, values := pathNameValues(root)
	numValues := len(values)

	for _, node := range *t {
		if node.name != name || len(node.params) != numValues {
			continue
		}

		// Node matched.. Collect variables
		if numValues != 0 && m.vars == nil {
			m.vars = make(map[string]string)
		}
		for i, p := range node.params {
			m.vars[p] = values[i]
		}

		matchedNode = node
		break
	}

	// No match
	if matchedNode == nil {
		return nil
	}

	// Full path match
	if len(next) == 0 {
		m.node = matchedNode
		m.path = matchedNode.path
		return matchedNode
	}

	// Paths matched so far.. Continue searching sub tree
	return matchedNode.subpaths.match(next, m)
}

// pathSplit splits a path into 2 parts - root and remaining.
// For the last node remaining part will be an empty string.
//
// Examples:
// pathSplit("/one") returns ("/one", "").
// pathSplit("/one/two/3") returns ("/one", "/two/3").
func pathSplit(p string) (string, string) {
	if len(p) > 1 {
		if k := strings.Index(p[1:], "/") + 1; k > 0 {
			return p[0:k], p[k:]
		}
	}

	return p, "" // last node
}

// pathNameValues splits the path element name and value list.
// Path element is expected to be in "/name[=value[,value]*]" syntax.
//
// Exampels:
// pathNameValues("/xx") returns ("/xx", nil)
// pathNameValues("/xx=") returns ("/xx", [])
// pathNameValues("/xx=yy") returns ("/xx", ["yy"])
// pathNameValues("/xx=yy,zz") returns ("/xx", ["yy", "zz"])
func pathNameValues(p string) (string, []string) {
	if k := strings.Index(p, "="); k != -1 {
		return p[0:k], strings.Split(p[k+1:], ",")
	}

	return p, nil
}

// pathParamExpr is the regex to parse path parameter in "{xyz}" syntax.
// Parameter should be a yang identifier as defined in YANG RFC6020; it
// can include alphabets, digits, dot, hypen or underscore.
var pathParamExpr = regexp.MustCompile(`{([a-zA-Z0-9_.-]+)}`)

// newRouteNode creates a routeNode object for a path element p.
func newRouteNode(p, parentPrefix string) *routeNode {
	node := &routeNode{path: parentPrefix + p}
	node.name, node.params = pathNameValues(p)

	// remove { } around the parameter name
	for i, param := range node.params {
		m := pathParamExpr.FindStringSubmatch(param)
		if len(m) != 2 { // invalid path param
			return nil
		}
		node.params[i] = m[1]
	}

	return node
}

// routeRegInfo holds route registration information.
type routeRegInfo struct {
	name    string
	method  string
	path    string
	handler http.HandlerFunc
}

// allRoutes is a collection of all routes
var allRoutes []routeRegInfo

// AddRoute appends specified routes to the routes collection.
// Called by init functions of swagger generated router.go files.
func AddRoute(name, method, pattern string, handler http.HandlerFunc) {
	rr := routeRegInfo{
		name:    name,
		method:  strings.ToUpper(method),
		path:    pattern,
		handler: handler,
	}

	allRoutes = append(allRoutes, rr)
}

// NewRouter creates a new http router instance for the REST server.
// Includes all routes registered via AddRoute API as well as few
// in-built routes.
func NewRouter() http.Handler {
	router := newRouter()
	return withStat(router)
}

// newRouter creates a new Router instance from the registered routes.
func newRouter() *Router {
	var mb muxBuilder
	mb.init()

	router := &Router{
		rcRoutes:  make(routeTree),
		muxRoutes: mb.router,
		optionsHandler: withMiddleware(
			http.HandlerFunc(commonOptionsHandler), "optionsHandler"),
	}

	glog.Infof("Server has %d routes", len(allRoutes))
	var rcRouteCount, muxRouteCount uint

	for _, rr := range allRoutes {
		if isServeFromTree(rr.path) {
			router.rcRoutes.add("", rr.path, &rr)
			rcRouteCount++
		} else {
			mb.add(&rr)
			muxRouteCount++
		}
	}

	glog.Infof("Installed %d routes on routeTree and %d on mux", rcRouteCount, muxRouteCount)

	mb.finish()
	return router
}

// muxBuilder is a utility to build a mux.Router object
// from the registered routes.
type muxBuilder struct {
	router *mux.Router
	paths  map[string]bool
}

// init initializes the mux router in muxBuilder
func (mb *muxBuilder) init() {
	mb.router = mux.NewRouter().StrictSlash(true).UseEncodedPath()
	mb.router.NotFoundHandler = http.HandlerFunc(notFound)
	mb.router.MethodNotAllowedHandler = http.HandlerFunc(notAllowed)

	mb.paths = make(map[string]bool)
}

// add creates a new route in mux router
func (mb *muxBuilder) add(rr *routeRegInfo) {
	glog.V(2).Infof("Adding %s, %s %s", rr.name, rr.method, rr.path)

	mb.paths[rr.path] = true
	h := withMiddleware(rr.handler, rr.name)
	mb.router.Name(rr.name).Methods(rr.method).Path(rr.path).Handler(h)
}

// finish creates routes for all internal service API handlers in
// mux router. Should be called after all REST API routes are added.
func (mb *muxBuilder) finish() {
	router := mb.router

	// Register common OPTIONS handler for every path template
	// New sub-router is used to avoid extra match during regular APIs
	sr := router.Methods("OPTIONS").Subrouter()
	oh := withMiddleware(http.HandlerFunc(commonOptionsHandler), "optionsHandler")
	for p := range mb.paths {
		sr.Path(p).Handler(oh)
	}

	// Documentation and test UI
	uiHandler := http.StripPrefix("/ui/", http.FileServer(http.Dir(swaggerUIDir)))
	router.Methods("GET").PathPrefix("/ui/").Handler(uiHandler)

	// Redirect "/ui" to "/ui/index.html"
	router.Methods("GET").Path("/ui").
		Handler(http.RedirectHandler("/ui/index.html", 301))

	if ClientAuth.Enabled("jwt") {
		//Allow POST for user/pass auth and or GET for cert auth.
		router.Methods("POST","GET").Path("/authenticate").Handler(http.HandlerFunc(Authenticate))
		router.Methods("POST","GET").Path("/refresh").Handler(http.HandlerFunc(Refresh))
	}

	// To download yang models
	ydirHandler := http.FileServer(http.Dir(translib.GetYangPath()))
	router.Methods("GET").PathPrefix("/models/yang/").
		Handler(http.StripPrefix("/models/yang/", ydirHandler))
}

// GetContextWithRouteInfo returns the RequestContext object for a
// http request r. Uses GetContext() to resolve the context object.
// If the request was serviced by mux router, fills the route match
// info into the context.
func GetContextWithRouteInfo(r *http.Request) (*RequestContext, *http.Request) {
	rc, r := GetContext(r)
	if rc.route != nil { // route info exists
		return rc, r
	}

	rc.route = new(routeMatchInfo)

	// Fill route info from mux "current route"
	if curr := mux.CurrentRoute(r); curr != nil {
		rc.route.path, _ = curr.GetPathTemplate()
		rc.route.vars = mux.Vars(r)
	}
	if len(rc.route.path) == 0 {
		rc.route.path = cleanPath(r.URL.Path)
	}

	return rc, r
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
		ts := time.Now()

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

		if rc.stats != nil {
			rc.stats.authTime = time.Since(ts)
		}

		if !success {
			writeErrorResponse(w, r, err)
		} else {
			inner.ServeHTTP(w, r)
		}
	})
}

// notFound responds with HTTP 404 status
func notFound(w http.ResponseWriter, r *http.Request) {
	writeErrorResponse(w, r,
		httpError(http.StatusNotFound, "Not Found"))
}

// notAllowed responds with HTTP 405 status
func notAllowed(w http.ResponseWriter, r *http.Request) {
	writeErrorResponse(w, r,
		httpError(http.StatusMethodNotAllowed, "%s Not Allowed", r.Method))
}

// writeErrorResponse writes HTTP error response for a error object
func writeErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	status, data, ctype := prepareErrorResponse(err, r)
	w.Header().Set("Content-Type", ctype)
	w.WriteHeader(status)
	w.Write(data)
}

// getAllMethodsForPath returns all registered HTTP methods for
// the path of a routeMatchInfo. For routeTree based matches the
// methods are readily available in matched node. For mux based
// matches it traverses allRoutes cache to find registered methods.
func (m *routeMatchInfo) getAllMethodsForPath() (methods []string) {
	if m.node != nil {
		for k := range m.node.handlers {
			methods = append(methods, k)
		}
	} else {
		for _, rr := range allRoutes {
			if rr.path == m.path {
				methods = append(methods, rr.method)
			}
		}
	}
	return
}
