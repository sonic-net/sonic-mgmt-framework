///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package server

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Root directory for UI files
var swaggerUIDir = "./ui"

// SetUIDirectory functions sets directiry where Swagger UI
// resources are maintained.
func SetUIDirectory(directory string) {
	swaggerUIDir = directory
}

// Route registration information
type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

// Collection of all routes
var allRoutes []Route

// AddRoute appends specified routes to the routes collection.
// Called by init functions of swagger generated router.go files.
func AddRoute(name, method, pattern string, handler http.HandlerFunc) {
	route := Route{
		Name:    name,
		Method:  strings.ToUpper(method),
		Pattern: pattern,
		Handler: handler,
	}

	allRoutes = append(allRoutes, route)
}

// NewRouter function returns a new http router instance. Collects
// route information from swagger-codegen generated code and makes a
// github.com/gorilla/mux router object.
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	log.Printf("Server has %d paths", len(allRoutes))

	// Collect swagger generated route information
	for _, route := range allRoutes {
		handler := loggingWrapper(route.Handler, route.Name)

		//log.Printf(
		//	"++ %s %s %s\n",
		//	route.Method, route.Name, route.Pattern)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	// Documentation and test UI
	uiHandler := http.StripPrefix("/ui/", http.FileServer(http.Dir(swaggerUIDir)))
	router.Methods("GET").PathPrefix("/ui/").Handler(uiHandler)

	// Redirect "/ui" to "/ui/index.html"
	router.Methods("GET").Path("/ui").
		Handler(http.RedirectHandler("/ui/index.html", 301))

	router.Methods("GET").Path("/model").
		Handler(http.RedirectHandler("/ui/model.html", 301))

	// Metadata discovery handler
	metadataHandler := http.HandlerFunc(hostMetadataHandler)
	router.Methods("GET").Path("/.well-known/host-meta").
		Handler(loggingWrapper(metadataHandler, "hostMetadataHandler"))

	return router
}

// loggingWrapper creates a new http.HandlerFunc which wraps a handler
// function to log time taken by it.
func loggingWrapper(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s %s; %s took %s",
			r.Method, r.RequestURI, name, time.Since(start))
	})
}
