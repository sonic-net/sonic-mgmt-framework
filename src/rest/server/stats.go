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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/gorilla/mux"
)

func init() {
	AddRoute("getStats", "GET", "/debug/stats", getStatsHandler)
	AddRoute("delStats", "DELETE", "/debug/stats", deleteStatsHandler)
}

// svcRequestStat tracks time taken by in-built service requests (non-REST).
// Includes mux and internal handler timings.
var svcRequestStat opStat

// apiRequestStat tracks time taken for processing REST API requests.
// Includes mux and REST handlers (handlerStat) timings.
var apiRequestStat opStat

// handlerStat tracks time taken by REST API handlers.
// Includes aut, handler core and translib (translibStat) timings.
var handlerStat opStat

// translibStat tracks time taken by translib.
// Includes translib infra, ygot, app and cvl timings.
var translibStat opStat

var theStatMutex sync.Mutex

// opStat tracks number of times an operation is invoked
// and time taken for it.
type opStat struct {
	Hits uint          `json:"hits"`       // Number of hits/calls
	Time time.Duration `json:"total-time"` // Total time taken
	Peak time.Duration `json:"peak-time"`  // Max time taken aby any call
}

func (s *opStat) add(t time.Duration) {
	s.Hits++
	s.Time += t
	if t > s.Peak {
		s.Peak = t
	}
}

func (s *opStat) clear() {
	s.Hits = 0
	s.Time = 0
	s.Peak = 0
}

// clearAllStats resets all stats maintained by REST server
func clearAllStats() {
	theStatMutex.Lock()

	svcRequestStat.clear()
	apiRequestStat.clear()
	handlerStat.clear()
	translibStat.clear()

	theStatMutex.Unlock()
}

// apiStats holds different stats for a REST API.
// Each request will have an instance of apiStats object maintained
// in its context. Can be accessed thru getApiStats function.
// For convenience it will also be available thru RequestContext.
type apiStats struct {
	handlerTime  time.Duration
	translibTime time.Duration
	internal     bool
}

// getApiStats returns the apiStats from a request's contxt.
// Returns nil if stat object not found in the context.
func getApiStats(r *http.Request) *apiStats {
	s, _ := r.Context().Value(statsContextKey).(*apiStats)
	return s
}

// markInternal tags a request as internal service API request.
// Statistics will be added to non-REST bucket.
func markInternal(r *http.Request) {
	if s := getApiStats(r); s != nil {
		s.internal = true
	}
}

// withStat creates a stats tracking for given mux.Router.
func withStat(router *mux.Router) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var stats apiStats
		r = r.WithContext(context.WithValue(r.Context(), statsContextKey, &stats))

		ts := time.Now()

		router.ServeHTTP(w, r)

		tt := time.Since(ts)

		theStatMutex.Lock()

		if stats.internal || stats.handlerTime == 0 {
			svcRequestStat.add(tt)
		} else {
			apiRequestStat.add(tt)
			handlerStat.add(stats.handlerTime)
			if stats.translibTime > 0 {
				translibStat.add(stats.translibTime)
			}
		}

		theStatMutex.Unlock()
	})
}

// getStatsHandler is the HTTP handler for "GET /debug/stats"
func getStatsHandler(w http.ResponseWriter, r *http.Request) {
	markInternal(r)
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		writeStatsJSON(w)
	} else {
		writeStatsText(w)
	}
}

// writeStatsText dumps all stats into a http.ResponseWriter in tabular format.
func writeStatsText(w http.ResponseWriter) {
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
	theStatMutex.Lock()

	fmt.Fprintln(tw, "API TYPE\tTIMER TYPE\tNUM HITS\tTOTAL TIME\tPEAK")
	fmt.Fprintln(tw, "============\t============\t==========\t============\t============")

	fmt.Fprintf(tw, "REST APIs")
	fmt.Fprintf(tw, "\tServer\t%d\t%s\t%s\n", apiRequestStat.Hits, apiRequestStat.Time, apiRequestStat.Peak)
	fmt.Fprintf(tw, "\tHandler\t%d\t%s\t%s\n", handlerStat.Hits, handlerStat.Time, handlerStat.Peak)
	fmt.Fprintf(tw, "\tTranslib\t%d\t%s\t%s\n", translibStat.Hits, translibStat.Time, translibStat.Peak)

	fmt.Fprintf(tw, "Service APIs")
	fmt.Fprintf(tw, "\tServer\t%d\t%s\t%s\n", svcRequestStat.Hits, svcRequestStat.Time, svcRequestStat.Peak)

	theStatMutex.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	tw.Flush()
}

// writeStatsText dumps all stats into a http.ResponseWriter as json.
func writeStatsJSON(w http.ResponseWriter) {
	data := map[string]interface{}{}
	theStatMutex.Lock()

	data["rest-api"] = map[string]interface{}{
		"server":   apiRequestStat,
		"handler":  handlerStat,
		"translib": translibStat,
	}

	data["service-api"] = map[string]interface{}{
		"server": svcRequestStat,
	}

	theStatMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	jw := json.NewEncoder(w)
	jw.SetIndent("", "  ")
	jw.Encode(data)
}

// deleteStatsHandler is the HTTP handler for "DELETE /debug/stats"
// Clears all statistics.
func deleteStatsHandler(w http.ResponseWriter, r *http.Request) {
	clearAllStats()
	markInternal(r)
	w.WriteHeader(http.StatusNoContent)
}
