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

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

// Root directory for UI files
var swaggerUIDir = "/etc/rest_server/ui"

func init() {
	flag.StringVar(&swaggerUIDir, "rest_ui", "/etc/rest_server/ui", "REST UI resources directory")
}

// Router dispatches http request to corresponding handlers.
type Router struct {
	// config for this Router instance
	config RouterConfig

	// routes contains all registered route info
	routes *routeStore
}

// RouterConfig holds runtime configurations for a Router instance.
type RouterConfig struct {
	// AuthEnable indicates if client authentication is enabled
	AuthEnable bool
}

// ServeHTTP resolves and invokes the handler for http request r.
// RESTCONF paths are served from the routeTree; rest from mux router.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := cleanPath(r.URL.EscapedPath())
	r = setContextValue(r, routerObjContextKey, router)

	if isServeFromTree(path) {
		router.serveFromTree(path, r, w)
	} else {
		router.routes.muxRoutes.ServeHTTP(w, r)
	}
}

// serveFromTree finds and invokes the handler for a path from routeTree
func (router *Router) serveFromTree(path string, r *http.Request, w http.ResponseWriter) {
	var routeInfo routeMatchInfo
	node := router.routes.rcRoutes.match(path, &routeInfo)

	// Node not found..
	if node == nil {
		notFound(w, r)
		return
	}

	handler := node.handlers[r.Method]
	if handler == nil && r.Method == "OPTIONS" {
		handler = router.routes.rcOptsHandler
	}

	// Node found, but no handler for the method
	if handler == nil {
		notAllowed(w, r)
		return
	}

	// Set route match info in context
	r = setContextValue(r, routeMatchContextKey, &routeInfo)

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
	return strings.HasPrefix(path, "/restconf/")
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
var allRoutes = newRouteStore()

// AddRoute appends specified routes to the routes collection.
// Called by init functions of swagger generated router.go files.
func AddRoute(name, method, pattern string, handler http.HandlerFunc) {
	rr := routeRegInfo{
		name:    name,
		method:  strings.ToUpper(method),
		path:    pattern,
		handler: handler,
	}

	allRoutes.addRoute(&rr)
}

// NewRouter creates a new http router instance for the REST server.
// Includes all routes registered via AddRoute API as well as few
// internal service API routes.
// Router instance specific configurations are accepted through a
// RouterConfig object.
func NewRouter(config RouterConfig) *Router {
	glog.Infof("Server has %d routes on routeTree and %d on mux router",
		allRoutes.rcRouteCount, allRoutes.muxRouteCount)

	// Add internal service API routes if not added already
	allRoutes.addServiceRoutes()

	router := &Router{
		config: config,
		routes: allRoutes,
	}

	return router
}

// routeStore holds REST route information - which includes route name,
// HTTP method, path and the handler function. All RESTCONF routes (path
// starting with "/restconf") are maintained in a routeTree. Other routes
// are maintained in a mux router.
type routeStore struct {
	rcRoutes      routeTree    // restconf routes
	rcOptsHandler http.Handler // OPTIONS handler for restconf routes
	rcRouteCount  uint32       // number of restconf routes

	muxRoutes      *mux.Router         // non-restconf and internal routes (UI, yang)
	muxOptsRouter  *mux.Router         // subrouter for OPTIONS handlers
	muxOptsHandler http.Handler        // OPTIONS handler for mux routes
	muxOptsData    map[string][]string // path to operations map for mux routes
	muxRouteCount  uint32              // number of routes in mux router

	svcRoutesAdded bool // indicates if service routes have been registered
}

// newRouteStore creates an empty routeStore instance.
func newRouteStore() *routeStore {
	rs := new(routeStore)
	rs.rcRoutes = make(routeTree)
	rs.rcOptsHandler = withMiddleware(http.HandlerFunc(rcOptions), "optionsHandler")

	r := mux.NewRouter().StrictSlash(true).UseEncodedPath()
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.MethodNotAllowedHandler = http.HandlerFunc(notAllowed)

	rs.muxRoutes = r
	rs.muxOptsRouter = r.Methods("OPTIONS").Subrouter()
	rs.muxOptsData = make(map[string][]string)
	rs.muxOptsHandler = withMiddleware(http.HandlerFunc(muxOptions), "optionsHandler")

	return rs
}

func (rs *routeStore) addRoute(rr *routeRegInfo) {
	glog.V(2).Infof("Adding route %s, %s %s", rr.name, rr.method, rr.path)

	if isServeFromTree(rr.path) {
		rs.rcRoutes.add("", rr.path, rr)
		rs.rcRouteCount++
	} else {
		rs.addMuxRoute(rr)
	}
}

func (rs *routeStore) addMuxRoute(rr *routeRegInfo) {
	h := withMiddleware(rr.handler, rr.name)
	rs.muxRoutes.Methods(rr.method).Path(rr.path).Handler(h)
	rs.muxOptsRouter.Path(rr.path).Handler(rs.muxOptsHandler)
	rs.muxOptsData[rr.path] = append(rs.muxOptsData[rr.path], rr.method)
	rs.muxRouteCount++
}

// finish creates routes for all internal service API handlers in
// mux router. Should be called after all REST API routes are added.
func (rs *routeStore) addServiceRoutes() {
	if rs.svcRoutesAdded {
		return
	}

	rs.svcRoutesAdded = true
	router := rs.muxRoutes

	// Documentation and test UI
	uiHandler := http.StripPrefix("/ui/", http.FileServer(http.Dir(swaggerUIDir)))
	router.Methods("GET").PathPrefix("/ui/").Handler(uiHandler)

	// Redirect "/ui" to "/ui/index.html"
	router.Methods("GET").Path("/ui").
		Handler(http.RedirectHandler("/ui/index.html", http.StatusMovedPermanently))
}

// getRouteMatchInfo returns routeMatchInfo from request context.
func getRouteMatchInfo(r *http.Request) *routeMatchInfo {
	m, _ := getContextValue(r, routeMatchContextKey).(*routeMatchInfo)
	if m != nil {
		return m
	}

	m = new(routeMatchInfo)

	// Fill route info from mux "current route"
	if curr := mux.CurrentRoute(r); curr != nil {
		m.path, _ = curr.GetPathTemplate()
		m.vars = mux.Vars(r)
	}
	if len(m.path) == 0 {
		m.path = cleanPath(r.URL.Path)
	}

	return m
}

// getRouterConfig returns the RouterConfig from current HTTP
// requests's context. Returns nil if the RouterConfig was not set.
func getRouterConfig(r *http.Request) *RouterConfig {
	rr := getContextValue(r, routerObjContextKey).(*Router)
	return &rr.config
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

		glog.Infof("[%s] %s took %s", rc.ID, name, time.Since(start))
	})
}

// withMiddleware function prepares the default middleware chain for
// REST APIs.
func withMiddleware(h http.Handler, name string) http.Handler {
	h = authMiddleware(h)
	return loggingMiddleware(h, name)
}

// notFound responds with HTTP 404 status
func notFound(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("NOT FOUND: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	writeErrorResponse(w, r,
		httpError(http.StatusNotFound, "Not Found"))
}

// notAllowed responds with HTTP 405 status
func notAllowed(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("NOT ALLOWED: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	writeErrorResponse(w, r,
		httpError(http.StatusMethodNotAllowed, "%s Not Allowed", r.Method))
}

// rcOptions handles OPTIONS for routeTree based paths
func rcOptions(w http.ResponseWriter, r *http.Request) {
	match := getRouteMatchInfo(r)
	var methods []string
	for k := range match.node.handlers {
		methods = append(methods, k)
	}
	writeOptionsResponse(w, r, match.path, methods)
}

// muxOptions handles OPTIONS for mux matched paths
func muxOptions(w http.ResponseWriter, r *http.Request) {
	match := getRouteMatchInfo(r)
	routr := getContextValue(r, routerObjContextKey).(*Router)
	writeOptionsResponse(w, r, match.path, routr.routes.muxOptsData[match.path])
}
