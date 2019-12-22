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
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestStats_basic(t *testing.T) {
	var st opStat

	st.add(time.Second)
	st.add(time.Minute)
	for i := 1; i < 60; i++ {
		st.add(time.Second)
	}

	expTime := (2 * time.Minute)
	if st.Hits != 61 || st.Time != expTime || st.Peak != time.Minute {
		t.Fatalf("Found incorrect stats %v; expecting {%d %s %s}", st, 61, expTime, time.Minute)
	}

	st.clear()
	if st.Hits != 0 || st.Time != 0 || st.Peak != 0 {
		t.Fatalf("Found incorrect stats %v; expecting {0 0s 0s}", st)
	}
}

func TestStats_mware(t *testing.T) {
	router := mux.NewRouter()
	server := withStat(router)

	fiftyMillis := 50 * time.Millisecond
	hundredMillis := 100 * time.Millisecond

	// API handler - sets both handlerTime and translibTime
	router.Methods("GET").Path("/test/1").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getApiStats(r).handlerTime = time.Minute
		getApiStats(r).translibTime = time.Second
		time.Sleep(hundredMillis)
	})

	// API handler - sets only handlerTime
	router.Methods("GET").Path("/test/2").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getApiStats(r).handlerTime = time.Hour
		time.Sleep(fiftyMillis)
	})

	// Service handler
	router.Methods("GET").Path("/test/3").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
	})

	// Service handler - sets handlerTime, but explicitly marks internal flag
	router.Methods("GET").Path("/test/4").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getApiStats(r).internal = true
		getApiStats(r).handlerTime = time.Minute
		time.Sleep(hundredMillis)
	})

	// Make calls
	clearAllStats()
	tryGet(t, server, 10, "/test/1")
	tryGet(t, server, 20, "/test/2")
	tryGet(t, server, 2, "/test/3")
	tryGet(t, server, 5, "/test/4")

	// Check Service API stats
	exp := opStat{Hits: 7, Time: (2*time.Second + 5*hundredMillis), Peak: time.Second}
	if approx(svcRequestStat, hundredMillis) != exp {
		t.Fatalf("svcRequestStat %v does not match %v", svcRequestStat, exp)
	}

	// Check REST API stats
	exp = opStat{Hits: 30, Time: (10*hundredMillis + 20*fiftyMillis), Peak: hundredMillis}
	if approx(apiRequestStat, hundredMillis) != exp {
		t.Fatalf("apiRequestStat %v does not match %v", apiRequestStat, exp)
	}

	exp = opStat{Hits: 30, Time: (10*time.Minute + 20*time.Hour), Peak: time.Hour}
	if handlerStat != exp {
		t.Fatalf("handlerStat %v does not match %v", handlerStat, exp)
	}

	exp = opStat{Hits: 10, Time: 10 * time.Second, Peak: time.Second}
	if translibStat != exp {
		t.Fatalf("translibStat %v does not match %v", translibStat, exp)
	}
}

func tryGet(t *testing.T, server http.Handler, count int, path string) {
	for i := 0; i < count; i++ {
		r := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("GET %s returned error: %d %s", path, w.Code, w.Body)
		}
	}
}

func approx(s opStat, d time.Duration) opStat {
	return opStat{Hits: s.Hits, Time: s.Time.Truncate(d), Peak: s.Peak.Round(d)}
}

func TestStats_get(t *testing.T) {
	t.Run("text", testGetStats("text/plain", "text/plain"))
	t.Run("json", testGetStats("application/json", "application/json"))
	t.Run("default", testGetStats("", "text/plain"))
	t.Run("unknown", testGetStats("application/yang-data+json", "text/plain"))
}

func testGetStats(reqType, rspType string) func(*testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/debug/stats", nil)
		if reqType != "" {
			r.Header.Set("Accept", reqType)
		}

		NewMuxRouter().ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Unexpected response code %d", w.Code)
		}

		ctype := w.Header().Get("Content-Type")
		if ctype != rspType {
			t.Fatalf("Unexpected content-type '%s'; expected '%s'", ctype, rspType)
		}

		//TODO check contents
	}
}

func TestStats_del(t *testing.T) {
	w := httptest.NewRecorder()
	NewMuxRouter().ServeHTTP(w, httptest.NewRequest("DELETE", "/debug/stats", nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("Unexpected response code %d", w.Code)
	}

	var null opStat
	if svcRequestStat != null {
		t.Fatalf("Unexpected svcRequestStat %v", svcRequestStat)
	}
	if apiRequestStat != null {
		t.Fatalf("Unexpected apiRequestStat %v", apiRequestStat)
	}
	if handlerStat != null {
		t.Fatalf("Unexpected handlerStat %v", handlerStat)
	}
	if translibStat != null {
		t.Fatalf("Unexpected translibStat %v", translibStat)
	}
}
