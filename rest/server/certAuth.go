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
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
)

// hostEtcDir is where the host /etc is mounted inside the mgmt-framework
// container. The container's own /etc does not have the host accounts.
var hostEtcDir = "/host_etc"

var errNoSuchUser = errors.New("no such user")

// hostAccount is a host user resolved from passwd and group files.
type hostAccount struct {
	name string
	gids []string // primary gid first, then supplementary gids
}

// CertAuthenAndAuthor authenticates a request using the subject CN of the
// verified TLS client certificate. The CN must name an existing host account;
// write operations additionally require membership in the admin group.
func CertAuthenAndAuthor(r *http.Request, rc *RequestContext) error {
	username := clientCertUsername(r)
	if username == "" {
		glog.Warningf("[%s] Client certificate has no subject common name", rc.ID)
		return httpError(http.StatusUnauthorized, "")
	}

	acct, err := lookupHostAccount(username)
	if err != nil {
		glog.Warningf("[%s] Client certificate user=%s rejected; %v", rc.ID, username, err)
		return httpError(http.StatusUnauthorized, "")
	}

	glog.Infof("[%s] Certificate authentication passed. user=%s", rc.ID, username)

	if isWriteOperation(r) {
		admin, err := isHostAdmin(acct)
		if err != nil {
			glog.Warningf("[%s] Admin group lookup failed; %v", rc.ID, err)
		}
		if !admin {
			glog.Warningf("[%s] Not an admin; cannot allow %s", rc.ID, r.Method)
			return httpError(http.StatusForbidden, "Not an admin user")
		}
	}

	glog.Infof("[%s] Authorization passed", rc.ID)
	return nil
}

// clientCertUsername returns the subject CN of the verified client
// certificate, or "" if the connection has no verified certificate.
func clientCertUsername(r *http.Request) string {
	if r.TLS == nil || len(r.TLS.VerifiedChains) == 0 || len(r.TLS.VerifiedChains[0]) == 0 {
		return ""
	}
	return r.TLS.VerifiedChains[0][0].Subject.CommonName
}

// lookupHostAccount resolves a user and its group ids from the host passwd
// and group files.
func lookupHostAccount(name string) (*hostAccount, error) {
	var acct *hostAccount
	err := scanColonFile(filepath.Join(hostEtcDir, "passwd"), func(f []string) bool {
		if len(f) >= 4 && f[0] == name {
			acct = &hostAccount{name: name, gids: []string{f[3]}}
			return false
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, errNoSuchUser
	}

	err = scanColonFile(filepath.Join(hostEtcDir, "group"), func(f []string) bool {
		if len(f) >= 4 && f[2] != acct.gids[0] {
			for _, m := range strings.Split(f[3], ",") {
				if m == name {
					acct.gids = append(acct.gids, f[2])
					break
				}
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return acct, nil
}

// isHostAdmin reports whether the account belongs to the primary group of
// the host "admin" user, matching the check IsAdminGroup applies.
func isHostAdmin(acct *hostAccount) (bool, error) {
	admin, err := lookupHostAccount("admin")
	if err != nil {
		return false, fmt.Errorf("admin user: %v", err)
	}
	for _, gid := range acct.gids {
		if gid == admin.gids[0] {
			return true, nil
		}
	}
	return false, nil
}

// scanColonFile calls fn with the fields of each entry in a colon separated
// file like /etc/passwd. Scanning stops when fn returns false.
func scanColonFile(path string, fn func([]string) bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || line[0] == '#' || line[0] == '+' || line[0] == '-' {
			continue
		}
		if !fn(strings.Split(line, ":")) {
			break
		}
	}
	return s.Err()
}

// CRLChecker rejects client certificates revoked by a CRL that is signed
// by one of the trusted client CAs.
type CRLChecker struct {
	issuer     []byte
	revoked    map[string]struct{}
	nextUpdate time.Time
}

// LoadCRLChecker loads a PEM or DER encoded CRL and verifies that it is
// signed by a certificate in caFile.
func LoadCRLChecker(crlFile, caFile string) (*CRLChecker, error) {
	data, err := os.ReadFile(crlFile)
	if err != nil {
		return nil, err
	}
	if block, _ := pem.Decode(data); block != nil {
		if block.Type != "X509 CRL" {
			return nil, fmt.Errorf("unexpected PEM block %q in %s", block.Type, crlFile)
		}
		data = block.Bytes
	}
	crl, err := x509.ParseRevocationList(data)
	if err != nil {
		return nil, err
	}

	cas, err := readPEMCertificates(caFile)
	if err != nil {
		return nil, err
	}
	signed := false
	for _, ca := range cas {
		if crl.CheckSignatureFrom(ca) == nil {
			signed = true
			break
		}
	}
	if !signed {
		return nil, fmt.Errorf("%s is not signed by a CA in %s", crlFile, caFile)
	}

	if !crl.NextUpdate.IsZero() && time.Now().After(crl.NextUpdate) {
		return nil, fmt.Errorf("%s expired at %v", crlFile, crl.NextUpdate)
	}

	c := &CRLChecker{
		issuer:     crl.RawIssuer,
		revoked:    make(map[string]struct{}),
		nextUpdate: crl.NextUpdate,
	}
	for _, rc := range crl.RevokedCertificates {
		c.revoked[rc.SerialNumber.String()] = struct{}{}
	}
	glog.Infof("Loaded CRL %s with %d revoked certificates", crlFile, len(c.revoked))
	return c, nil
}

// VerifyPeerCertificate can be used as tls.Config.VerifyPeerCertificate.
func (c *CRLChecker) VerifyPeerCertificate(_ [][]byte, verifiedChains [][]*x509.Certificate) error {
	if !c.nextUpdate.IsZero() && time.Now().After(c.nextUpdate) {
		return fmt.Errorf("CRL expired at %v", c.nextUpdate)
	}
	for _, chain := range verifiedChains {
		for _, cert := range chain {
			if !bytes.Equal(cert.RawIssuer, c.issuer) {
				continue
			}
			if _, ok := c.revoked[cert.SerialNumber.String()]; ok {
				glog.Warningf("Rejected revoked client certificate CN=%s serial=%s",
					cert.Subject.CommonName, cert.SerialNumber)
				return fmt.Errorf("certificate serial %s is revoked", cert.SerialNumber)
			}
		}
	}
	return nil
}

func readPEMCertificates(path string) ([]*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var certs []*x509.Certificate
	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates in %s", path)
	}
	return certs, nil
}
