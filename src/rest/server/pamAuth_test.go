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
//
// Test cases for REST Server PAM Authentication module.
//
// Runs various combinations with local and TACACS+ user credentials.
// Test users should be already configured in the system. Below table
// lists various default user name and passwords and corresponding
// command line parameters to override them.
//
// Test user type          User name      Password    Command line param
// ----------------------  -------------  ----------  -------------------
// Local admin user        testadmin      password    -ladmname -ladmpass
// Local non-admin user    testuser       password    -lusrname -lusrpass
// TACACS+ admin user      tactestadmin   password    -tadmname -tadmpass
// TACACS+ non-admin user  tactestuser    password    -tusrname -tusrpass
//
// By default all test cases are skipped!! This is to avoid seeing test
// failures if target system is not ready. Command line param -authtest
// should be passed to enable the test cases. Valid values are "local"
// or "tacacs" or comma separated list of them.
//
// -authtest=local         ==> Tests with only local user credentials
// -authtest=tacacs        ==> Tests with only TACACS+ user credentials
// -authtest=local,tacacs  ==> Tests with both local and TACACS+ users
//
///////////////////////////////////////////////////////////////////////

package server

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var authTest map[string]bool
var lusrName = flag.String("lusrname", "testuser", "Local non-admin username")
var lusrPass = flag.String("lusrpass", "password", "Local non-admin password")
var ladmName = flag.String("ladmname", "testadmin", "Local admin username")
var ladmPass = flag.String("ladmpass", "password", "Local admin password")
var tusrName = flag.String("tusrname", "tactestuser", "TACACS+ non-admin username")
var tusrPass = flag.String("tusrpass", "password", "TACACS+ non-admin password")
var tadmName = flag.String("tadmname", "tactestadmin", "TACACS+ admin username")
var tadmPass = flag.String("tadmpass", "password", "TACACS+ admin password")

func init() {
	fmt.Println("+++++ pamAuth_test +++++")
}

func TestMain(m *testing.M) {

	t := flag.String("authtest", "", "Comma separated auth types to test (local tacacs)")
	flag.Parse()

	var tlist []string
	if *t != "" {
		authTest = make(map[string]bool)
		for _, x := range strings.Split(*t, ",") {
			v := strings.ToLower(strings.TrimSpace(x))
			if v == "local" || v == "tacacs" {
				authTest[v] = true
				tlist = append(tlist, v)
			}
		}
		if len(authTest) != 0 {
			authTest[""] = true // Special key for any auth
		}
	}

	fmt.Println("+++++ Enabled auth test types", tlist)

	os.Exit(m.Run())
}

// Dummy test handler which returns 200 on success; 401 on
// authentication failure and 403 on authorization failure
var authTestHandler = authMiddleware(http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

func TestAuthLocalUser_Get(t *testing.T) {
	ensureAuthTestEnabled(t, "local")
	testAuthGet(t, *lusrName, *lusrPass, 200)
}

func TestAuthLocalUser_Set(t *testing.T) {
	ensureAuthTestEnabled(t, "local")
	testAuthSet(t, *lusrName, *lusrPass, 403)
}

func TestAuthLocalAdmin_Get(t *testing.T) {
	ensureAuthTestEnabled(t, "local")
	testAuthGet(t, *ladmName, *ladmPass, 200)
}

func TestAuthLocalAdmin_Set(t *testing.T) {
	ensureAuthTestEnabled(t, "local")
	testAuthSet(t, *ladmName, *ladmPass, 200)
}

func TestAuthTacacsUser_Get(t *testing.T) {
	ensureAuthTestEnabled(t, "tacacs")
	testAuthGet(t, *tusrName, *tusrPass, 200)
}

func TestAuthTacacsUser_Set(t *testing.T) {
	ensureAuthTestEnabled(t, "tacacs")
	testAuthSet(t, *tusrName, *tusrPass, 403)
}

func TestAuthTacacsAdmin_Get(t *testing.T) {
	ensureAuthTestEnabled(t, "tacacs")
	testAuthGet(t, *tadmName, *tadmPass, 200)
}

func TestAuthTacacsAdmin_Set(t *testing.T) {
	ensureAuthTestEnabled(t, "tacacs")
	testAuthSet(t, *tadmName, *tadmPass, 200)
}

func TestAuth_NoUser(t *testing.T) {
	ensureAuthTestEnabled(t, "")
	testAuthGet(t, "", "", 401)
	testAuthSet(t, "", "", 401)
}

func TestAuth_BadUser(t *testing.T) {
	ensureAuthTestEnabled(t, "")
	testAuthGet(t, "baduserbaduserbaduser", "password", 401)
	testAuthSet(t, "baduserbaduserbaduser", "password", 401)
}

func TestAuth_BadPass(t *testing.T) {
	ensureAuthTestEnabled(t, "")
	testAuthGet(t, *lusrName, "Hello,world!", 401)
	testAuthSet(t, *ladmName, "Hello,world!", 401)
}

func ensureAuthTestEnabled(t *testing.T, authtype string) {
	if _, ok := authTest[authtype]; !ok {
		t.Skipf("%s auth tests not enabled.. Rerun with -authtest flag", authtype)
	}
}

func testAuthGet(t *testing.T, username, password string, expStatus int) {
	t.Run("GET", testAuth("GET", username, password, expStatus))
	t.Run("HEAD", testAuth("HEAD", username, password, expStatus))
	t.Run("OPTIONS", testAuth("OPTIONS", username, password, expStatus))
}

func testAuthSet(t *testing.T, username, password string, expStatus int) {
	t.Run("PUT", testAuth("PUT", username, password, expStatus))
	t.Run("POST", testAuth("POST", username, password, expStatus))
	t.Run("PATCH", testAuth("PATCH", username, password, expStatus))
	t.Run("DELETE", testAuth("DELETE", username, password, expStatus))
}

func testAuth(method, username, password string, expStatus int) func(*testing.T) {
	return func(t *testing.T) {
		// Temporariliy enable password auth if not enabled already
		if !ClientAuth.Enabled("password") {
			ClientAuth.Set("password")
			defer ClientAuth.Unset("password")
		}

		r := httptest.NewRequest(method, "/auth", nil)
		w := httptest.NewRecorder()

		if username != "" {
			r.SetBasicAuth(username, password)
		}

		authTestHandler.ServeHTTP(w, r)

		if w.Code != expStatus {
			t.Fatalf("Expected response %d; got %d", expStatus, w.Code)
		}
	}
}
