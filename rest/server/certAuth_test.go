////////////////////////////////////////////////////////////////////////////////
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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testPasswd = `root:x:0:0:root:/root:/bin/bash
# comment
admin:x:1000:1000::/home/admin:/bin/bash
alice:x:1001:1001::/home/alice:/bin/bash
bob:x:1002:1002::/home/bob:/bin/bash
`

const testGroup = `root:x:0:
admin:x:1000:
alice:x:1001:
bob:x:1002:
sudo:x:27:admin,alice
docker:x:999:alice,bob
admins:x:1000:alice
`

func setupHostEtc(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "passwd"), []byte(testPasswd))
	writeFile(t, filepath.Join(dir, "group"), []byte(testGroup))
	saved := hostEtcDir
	hostEtcDir = dir
	t.Cleanup(func() { hostEtcDir = saved })
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLookupHostAccount(t *testing.T) {
	setupHostEtc(t)

	acct, err := lookupHostAccount("alice")
	if err != nil {
		t.Fatalf("lookup alice: %v", err)
	}
	if want := []string{"1001", "27", "999", "1000"}; !equalStrings(acct.gids, want) {
		t.Fatalf("alice gids = %v, want %v", acct.gids, want)
	}

	for _, name := range []string{"nobody", "", "alice:x", "#"} {
		if _, err := lookupHostAccount(name); err != errNoSuchUser {
			t.Errorf("lookup %q: err = %v, want errNoSuchUser", name, err)
		}
	}

	for name, want := range map[string]bool{"admin": true, "alice": true, "bob": false} {
		acct, err := lookupHostAccount(name)
		if err != nil {
			t.Fatalf("lookup %s: %v", name, err)
		}
		if got, _ := isHostAdmin(acct); got != want {
			t.Errorf("isHostAdmin(%s) = %v, want %v", name, got, want)
		}
	}
}

func TestLookupHostAccountMissingFiles(t *testing.T) {
	saved := hostEtcDir
	hostEtcDir = t.TempDir()
	defer func() { hostEtcDir = saved }()

	if _, err := lookupHostAccount("admin"); err == nil {
		t.Fatal("lookup succeeded without a passwd file")
	}
}

func TestCertAuthenAndAuthor(t *testing.T) {
	setupHostEtc(t)

	tests := []struct {
		name   string
		cn     string
		noTLS  bool
		method string
		status int // 0 means allowed
	}{
		{"no TLS", "", true, "GET", http.StatusUnauthorized},
		{"empty CN", "", false, "GET", http.StatusUnauthorized},
		{"unknown user read", "hl_nobody", false, "GET", http.StatusUnauthorized},
		{"unknown user write", "hl_nobody", false, "PATCH", http.StatusUnauthorized},
		{"non-admin read", "bob", false, "GET", 0},
		{"non-admin write", "bob", false, "PATCH", http.StatusForbidden},
		{"non-admin delete", "bob", false, "DELETE", http.StatusForbidden},
		{"admin group member write", "alice", false, "POST", 0},
		{"admin write", "admin", false, "PUT", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/restconf/data/x", nil)
			if !tt.noTLS {
				cert := &x509.Certificate{Subject: pkix.Name{CommonName: tt.cn}}
				r.TLS = &tls.ConnectionState{VerifiedChains: [][]*x509.Certificate{{cert}}}
			}
			err := CertAuthenAndAuthor(r, &RequestContext{ID: "t"})
			checkAuthStatus(t, err, tt.status)
		})
	}
}

func TestCertAuthenAndAuthorUnverifiedPeer(t *testing.T) {
	setupHostEtc(t)

	// PeerCertificates alone (no verified chain) must not authenticate.
	r := httptest.NewRequest("GET", "/restconf/data/x", nil)
	cert := &x509.Certificate{Subject: pkix.Name{CommonName: "admin"}}
	r.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}}
	checkAuthStatus(t, CertAuthenAndAuthor(r, &RequestContext{ID: "t"}), http.StatusUnauthorized)
}

func checkAuthStatus(t *testing.T, err error, status int) {
	t.Helper()
	if status == 0 {
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		return
	}
	e, ok := err.(httpErrorType)
	if !ok || e.status != status {
		t.Fatalf("err = %#v, want status %d", err, status)
	}
}

type testCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

func newTestCA(t *testing.T, cn string) *testCA {
	t.Helper()
	key := newKey(t)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return &testCA{cert: cert, key: key}
}

func (ca *testCA) issue(t *testing.T, cn string, serial int64, usage x509.ExtKeyUsage) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key := newKey(t)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{usage},
		DNSNames:     []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert, key
}

func (ca *testCA) crl(t *testing.T, serials ...int64) []byte {
	t.Helper()
	return ca.crlUntil(t, time.Now().Add(time.Hour), serials...)
}

func (ca *testCA) crlUntil(t *testing.T, next time.Time, serials ...int64) []byte {
	t.Helper()
	var revoked []pkix.RevokedCertificate
	for _, s := range serials {
		revoked = append(revoked, pkix.RevokedCertificate{SerialNumber: big.NewInt(s), RevocationTime: time.Now()})
	}
	der, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:              big.NewInt(1),
		ThisUpdate:          time.Now().Add(-time.Hour),
		NextUpdate:          next,
		RevokedCertificates: revoked,
	}, ca.cert, ca.key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "X509 CRL", Bytes: der})
}

func newKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func certPEM(c *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
}

func TestCRLChecker(t *testing.T) {
	dir := t.TempDir()
	ca := newTestCA(t, "TESTCA")
	other := newTestCA(t, "OTHERCA")
	good, _ := ca.issue(t, "admin", 10, x509.ExtKeyUsageClientAuth)
	bad, _ := ca.issue(t, "admin", 11, x509.ExtKeyUsageClientAuth)

	caFile := filepath.Join(dir, "ca.pem")
	writeFile(t, caFile, append(certPEM(other.cert), certPEM(ca.cert)...))

	crlFile := filepath.Join(dir, "ca.crl")
	writeFile(t, crlFile, ca.crl(t, 11))

	checker, err := LoadCRLChecker(crlFile, caFile)
	if err != nil {
		t.Fatalf("LoadCRLChecker: %v", err)
	}
	if err := checker.VerifyPeerCertificate(nil, [][]*x509.Certificate{{good, ca.cert}}); err != nil {
		t.Errorf("good certificate rejected: %v", err)
	}
	if err := checker.VerifyPeerCertificate(nil, [][]*x509.Certificate{{bad, ca.cert}}); err == nil {
		t.Error("revoked certificate accepted")
	}

	// Same serial from a different issuer is not revoked by this CRL.
	otherCert, _ := other.issue(t, "admin", 11, x509.ExtKeyUsageClientAuth)
	if err := checker.VerifyPeerCertificate(nil, [][]*x509.Certificate{{otherCert, other.cert}}); err != nil {
		t.Errorf("certificate from another issuer rejected: %v", err)
	}

	// DER encoded CRL is accepted too.
	block, _ := pem.Decode(ca.crl(t, 11))
	derFile := filepath.Join(dir, "ca.der")
	writeFile(t, derFile, block.Bytes)
	if _, err := LoadCRLChecker(derFile, caFile); err != nil {
		t.Errorf("DER CRL: %v", err)
	}
}

func TestLoadCRLCheckerRejectsExpiredCRL(t *testing.T) {
	dir := t.TempDir()
	ca := newTestCA(t, "TESTCA")
	good, _ := ca.issue(t, "admin", 10, x509.ExtKeyUsageClientAuth)

	caFile := filepath.Join(dir, "ca.pem")
	writeFile(t, caFile, certPEM(ca.cert))

	crlFile := filepath.Join(dir, "expired.crl")
	writeFile(t, crlFile, ca.crlUntil(t, time.Now().Add(-time.Minute)))
	if _, err := LoadCRLChecker(crlFile, caFile); err == nil {
		t.Error("expired CRL was accepted")
	}

	// A CRL that expires after load must stop accepting clients.
	writeFile(t, crlFile, ca.crl(t))
	checker, err := LoadCRLChecker(crlFile, caFile)
	if err != nil {
		t.Fatalf("LoadCRLChecker: %v", err)
	}
	checker.nextUpdate = time.Now().Add(-time.Second)
	if err := checker.VerifyPeerCertificate(nil, [][]*x509.Certificate{{good, ca.cert}}); err == nil {
		t.Error("client accepted after the CRL expired")
	}
}

func TestLoadCRLCheckerRejectsUntrustedCRL(t *testing.T) {
	dir := t.TempDir()
	ca := newTestCA(t, "TESTCA")
	rogue := newTestCA(t, "ROGUE")

	caFile := filepath.Join(dir, "ca.pem")
	writeFile(t, caFile, certPEM(ca.cert))

	crlFile := filepath.Join(dir, "rogue.crl")
	writeFile(t, crlFile, rogue.crl(t))
	if _, err := LoadCRLChecker(crlFile, caFile); err == nil {
		t.Error("CRL signed by an untrusted CA was accepted")
	}

	writeFile(t, crlFile, certPEM(ca.cert))
	if _, err := LoadCRLChecker(crlFile, caFile); err == nil {
		t.Error("certificate PEM accepted as CRL")
	}

	if _, err := LoadCRLChecker(filepath.Join(dir, "missing.crl"), caFile); err == nil {
		t.Error("missing CRL file accepted")
	}
}

// TestCertAuthOverTLS runs a real TLS handshake to confirm the verified
// client certificate reaches CertAuthenAndAuthor and the CRL is enforced.
func TestCertAuthOverTLS(t *testing.T) {
	setupHostEtc(t)
	dir := t.TempDir()

	ca := newTestCA(t, "TESTCA")
	srvCert, srvKey := ca.issue(t, "localhost", 2, x509.ExtKeyUsageServerAuth)
	caFile := filepath.Join(dir, "ca.pem")
	writeFile(t, caFile, certPEM(ca.cert))
	crlFile := filepath.Join(dir, "ca.crl")
	writeFile(t, crlFile, ca.crl(t, 99))

	checker, err := LoadCRLChecker(crlFile, caFile)
	if err != nil {
		t.Fatal(err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca.cert)

	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := CertAuthenAndAuthor(r, &RequestContext{ID: "tls"}); err != nil {
			w.WriteHeader(err.(httpErrorType).status)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	srv.TLS = &tls.Config{
		Certificates:          []tls.Certificate{{Certificate: [][]byte{srvCert.Raw}, PrivateKey: srvKey}},
		ClientAuth:            tls.RequireAndVerifyClientCert,
		ClientCAs:             pool,
		VerifyPeerCertificate: checker.VerifyPeerCertificate,
		MinVersion:            tls.VersionTLS12,
	}
	srv.StartTLS()
	defer srv.Close()

	do := func(cn string, serial int64, method string) (int, error) {
		cert, key := ca.issue(t, cn, serial, x509.ExtKeyUsageClientAuth)
		client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			ServerName:   "localhost",
			Certificates: []tls.Certificate{{Certificate: [][]byte{cert.Raw}, PrivateKey: key}},
		}}}
		req, _ := http.NewRequest(method, srv.URL+"/restconf/data/x", nil)
		resp, err := client.Do(req)
		if err != nil {
			return 0, err
		}
		resp.Body.Close()
		return resp.StatusCode, nil
	}

	cases := []struct {
		cn     string
		serial int64
		method string
		want   int
	}{
		{"hl_nobody", 20, "GET", http.StatusUnauthorized},
		{"hl_nobody", 21, "PATCH", http.StatusUnauthorized},
		{"bob", 22, "GET", http.StatusNoContent},
		{"bob", 23, "PATCH", http.StatusForbidden},
		{"admin", 24, "PATCH", http.StatusNoContent},
	}
	for _, c := range cases {
		got, err := do(c.cn, c.serial, c.method)
		if err != nil || got != c.want {
			t.Errorf("%s %s: status=%d err=%v, want %d", c.cn, c.method, got, err, c.want)
		}
	}

	if _, err := do("admin", 99, "GET"); err == nil {
		t.Error("revoked certificate completed the request")
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
