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
// Tests for the local REST Unix domain socket authorization path added for
// sonic-buildimage#29504: skipping the Basic-Auth password challenge for
// requests that arrive over /var/run/rest-local.sock, while still applying
// the same write/admin authorization as remote Basic-Auth users.

package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// newUDSAuthTestRouter builds a router with AuthEnable=true and a dummy
// 200-on-success handler registered for GET and PUT/POST/PATCH/DELETE, the
// same shape pamAuth_test.go uses for its Basic-Auth tests.
func newUDSAuthTestRouter() *Router {
	r := newEmptyRouter()
	r.config.AuthEnable = true
	r.addRoute("uds_test_get", "GET", "/api-tests:uds-auth", authTestHandler)
	r.addRoute("uds_test_put", "PUT", "/api-tests:uds-auth", authTestHandler)
	r.addRoute("uds_test_post", "POST", "/api-tests:uds-auth", authTestHandler)
	return r
}

// withPeerCred returns a shallow copy of r carrying cred as its verified
// UDS peer credential, exactly as ConnContext would attach it for a real
// Unix domain socket connection. A nil cred simulates a plain TCP request.
func withPeerCred(r *http.Request, cred *PeerCred) *http.Request {
	if cred == nil {
		return r
	}
	return r.WithContext(context.WithValue(r.Context(), peerCredContextKey, cred))
}

// useHostIdentFixtures points HostPasswdPath/HostGroupPath at temporary
// fixture files for the duration of the test.
func useHostIdentFixtures(t *testing.T, passwd, group string) {
	t.Helper()
	dir := t.TempDir()

	origPasswd, origGroup := HostPasswdPath, HostGroupPath
	t.Cleanup(func() { HostPasswdPath, HostGroupPath = origPasswd, origGroup })

	HostPasswdPath = writeFixture(t, dir, "passwd", passwd)
	HostGroupPath = writeFixture(t, dir, "group", group)
}

const udsTestPasswd = `root:x:0:0:root:/root:/bin/bash
admin:x:1000:1000:SONiC admin:/home/admin:/bin/bash
operator:x:1001:1001:read-only operator:/home/operator:/bin/bash
`

const udsTestGroup = `root:x:0:
admin:x:1000:
operator:x:1001:
`

// --- TCP behavior must be unchanged ---

func TestUDS_TCPWithoutBasicAuth_Remains401(t *testing.T) {
	useHostIdentFixtures(t, udsTestPasswd, udsTestGroup)
	router := newUDSAuthTestRouter()

	r := httptest.NewRequest("GET", "/api-tests:uds-auth", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r) // no PeerCred attached => plain TCP request

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("TCP request without Basic Auth: got %d; want 401", w.Code)
	}
}

func TestUDS_TCPWriteWithoutBasicAuth_Remains401(t *testing.T) {
	useHostIdentFixtures(t, udsTestPasswd, udsTestGroup)
	router := newUDSAuthTestRouter()

	r := httptest.NewRequest("PUT", "/api-tests:uds-auth", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("TCP write without Basic Auth: got %d; want 401 (must not be silently authorized)", w.Code)
	}
}

// A forged PeerCred is impossible for a real TCP connection (ConnContext
// only ever attaches one for *net.UnixConn), but this test documents that
// the TCP code path is chosen purely by the *absence* of PeerCred in the
// request context, not by any inspectable request property a client could
// spoof (like RemoteAddr==127.0.0.1).
func TestUDS_LoopbackTCPWithoutPeerCred_StillRequiresBasicAuth(t *testing.T) {
	useHostIdentFixtures(t, udsTestPasswd, udsTestGroup)
	router := newUDSAuthTestRouter()

	r := httptest.NewRequest("GET", "/api-tests:uds-auth", nil)
	r.RemoteAddr = "127.0.0.1:54321"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("loopback-looking TCP request without PeerCred: got %d; want 401", w.Code)
	}
}

// --- UDS behavior ---

func TestUDS_RecognizedNonAdminGet_Succeeds(t *testing.T) {
	useHostIdentFixtures(t, udsTestPasswd, udsTestGroup)
	router := newUDSAuthTestRouter()

	r := httptest.NewRequest("GET", "/api-tests:uds-auth", nil)
	r = withPeerCred(r, &PeerCred{UID: 1001, GID: 1001, PID: 4242}) // operator, non-admin
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("UDS GET from recognized non-admin user: got %d; want 200", w.Code)
	}
}

func TestUDS_RecognizedNonAdminWrite_Returns403(t *testing.T) {
	useHostIdentFixtures(t, udsTestPasswd, udsTestGroup)
	router := newUDSAuthTestRouter()

	r := httptest.NewRequest("PUT", "/api-tests:uds-auth", nil)
	r = withPeerCred(r, &PeerCred{UID: 1001, GID: 1001, PID: 4242}) // operator, non-admin
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("UDS write from recognized non-admin user: got %d; want 403", w.Code)
	}
}

func TestUDS_RecognizedAdminWrite_Succeeds(t *testing.T) {
	useHostIdentFixtures(t, udsTestPasswd, udsTestGroup)
	router := newUDSAuthTestRouter()

	r := httptest.NewRequest("PUT", "/api-tests:uds-auth", nil)
	r = withPeerCred(r, &PeerCred{UID: 1000, GID: 1000, PID: 4242}) // admin
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("UDS write from admin user: got %d; want 200", w.Code)
	}
}

func TestUDS_UnknownUID_FailsClosed(t *testing.T) {
	useHostIdentFixtures(t, udsTestPasswd, udsTestGroup)
	router := newUDSAuthTestRouter()

	r := httptest.NewRequest("GET", "/api-tests:uds-auth", nil)
	r = withPeerCred(r, &PeerCred{UID: 65534, GID: 65534, PID: 4242}) // no passwd entry
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("UDS request from unresolvable uid: got %d; want 401 (fail closed)", w.Code)
	}
}

// TestUDS_ForgedCliUserHeaderHasNoEffect proves the fix for the identified
// flaw in the original design: a client-supplied identity header must never
// influence authorization. A non-admin peer claiming to be "admin" via a
// header must still be denied write access.
func TestUDS_ForgedCliUserHeaderHasNoEffect(t *testing.T) {
	useHostIdentFixtures(t, udsTestPasswd, udsTestGroup)
	router := newUDSAuthTestRouter()

	r := httptest.NewRequest("PUT", "/api-tests:uds-auth", nil)
	r.Header.Set("X-Sonic-Cli-User", "admin")
	r.Header.Set("CLI_USER", "admin")
	r = withPeerCred(r, &PeerCred{UID: 1001, GID: 1001, PID: 4242}) // real peer is operator, non-admin
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("UDS write with forged admin header from non-admin peer: got %d; want 403", w.Code)
	}
}

// --- SO_PEERCRED acquisition over a real Unix domain socket ---

// TestConnContext_RealUnixSocketPeerCred verifies ConnContext against a
// real AF_UNIX socket rather than a mocked credential value, confirming
// SO_PEERCRED is actually read from the kernel. Both ends of the
// connection are this test process, so the observed uid/gid/pid must match
// the test process's own.
func TestConnContext_RealUnixSocketPeerCred(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test-rest-local.sock")

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("failed to listen on unix socket: %v", err)
	}
	defer ln.Close()

	type acceptResult struct {
		ctx context.Context
		err error
	}
	resultCh := make(chan acceptResult, 1)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			resultCh <- acceptResult{err: err}
			return
		}
		defer conn.Close()
		ctx := ConnContext(context.Background(), conn)
		resultCh <- acceptResult{ctx: ctx}
	}()

	client, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("failed to dial unix socket: %v", err)
	}
	defer client.Close()

	result := <-resultCh
	if result.err != nil {
		t.Fatalf("Accept failed: %v", result.err)
	}

	cred, ok := result.ctx.Value(peerCredContextKey).(*PeerCred)
	if !ok || cred == nil {
		t.Fatalf("expected a PeerCred value in the connection context, got %#v", result.ctx.Value(peerCredContextKey))
	}

	wantUID := uint32(os.Getuid())
	wantGID := uint32(os.Getgid())
	wantPID := int32(os.Getpid())

	if cred.UID != wantUID || cred.GID != wantGID || cred.PID != wantPID {
		t.Fatalf("PeerCred = %+v; want uid=%d gid=%d pid=%d (this test process)",
			cred, wantUID, wantGID, wantPID)
	}
}

// TestConnContext_TCPConnUnaffected confirms ConnContext leaves the context
// untouched for a non-Unix connection (verified against a real TCP loopback
// connection, not just a type assertion in isolation).
func TestConnContext_TCPConnUnaffected(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on tcp: %v", err)
	}
	defer ln.Close()

	type acceptResult struct {
		ctx context.Context
		err error
	}
	resultCh := make(chan acceptResult, 1)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			resultCh <- acceptResult{err: err}
			return
		}
		defer conn.Close()
		ctx := ConnContext(context.Background(), conn)
		resultCh <- acceptResult{ctx: ctx}
	}()

	client, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial tcp: %v", err)
	}
	defer client.Close()

	result := <-resultCh
	if result.err != nil {
		t.Fatalf("Accept failed: %v", result.err)
	}

	if v := result.ctx.Value(peerCredContextKey); v != nil {
		t.Fatalf("expected no PeerCred value for a TCP connection, got %#v", v)
	}
}
