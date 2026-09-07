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
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/golang/glog"
	//"github.com/msteinert/pam"
	"golang.org/x/crypto/ssh"
)

// sshAuthAddr is the address of the local sshd used to validate credentials.
const sshAuthAddr = "127.0.0.1:22"

// sshAuthTimeout bounds both the TCP connect and the SSH handshake/auth phase
// so a rogue listener cannot hang the auth path indefinitely.
const sshAuthTimeout = 5 * time.Second

// sshHostKeyPaths lists candidate public host key files for the local sshd, in
// preference order. The mgmt-framework container mounts the host /etc read-only
// at /host_etc (see docker-sonic-mgmt-framework.mk), so the host sshd keys live
// under /host_etc/ssh rather than the container's own /etc/ssh.
var sshHostKeyPaths = []string{
	"/host_etc/ssh/ssh_host_ed25519_key.pub",
	"/host_etc/ssh/ssh_host_ecdsa_key.pub",
	"/host_etc/ssh/ssh_host_rsa_key.pub",
}

/*
type UserCredential struct {
	Username string
	Password string
}

//PAM conversation handler.
func (u UserCredential) PAMConvHandler(s pam.Style, msg string) (string, error) {

	switch s {
	case pam.PromptEchoOff:
		return u.Password, nil
	case pam.PromptEchoOn:
		return u.Password, nil
	case pam.ErrorMsg:
		return "", nil
	case pam.TextInfo:
		return "", nil
	default:
		return "", errors.New("unrecognized conversation message style")
	}
}

// PAMAuthenticate performs PAM authentication for the user credentials provided
func (u UserCredential) PAMAuthenticate() error {
	tx, err := pam.StartFunc("login", u.Username, u.PAMConvHandler)
	if err != nil {
		return err
	}
	return tx.Authenticate(0)
}

func PAMAuthUser(u string, p string) error {

	cred := UserCredential{u, p}
	err := cred.PAMAuthenticate()
	return err
}
*/

func IsAdminGroup(username string) bool {

	usr, err := user.Lookup(username)
	if err != nil {
		return false
	}
	gids, err := usr.GroupIds()
	if err != nil {
		return false
	}
	glog.V(2).Infof("User:%s, groups=%s", username, gids)
	admin, err := user.Lookup("admin")
	if err != nil {
		return false
	}
	for _, x := range gids {
		if x == admin.Gid {
			return true
		}
	}
	return false
}

// sshdHostKeyCallback returns an ssh.HostKeyCallback pinned to the local
// sshd's public host key(s) read from paths. It fails closed if no key can be
// loaded, so a rogue process bound to 127.0.0.1:22 cannot intercept credentials.
func sshdHostKeyCallback(paths []string) (ssh.HostKeyCallback, error) {
	var pinned []ssh.PublicKey
	var problems []string
	for _, path := range paths {
		keyBytes, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(keyBytes)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: parse: %v", path, err))
			continue
		}
		if len(bytes.TrimSpace(rest)) != 0 {
			problems = append(problems, fmt.Sprintf("%s: unexpected trailing data", path))
			continue
		}
		pinned = append(pinned, pubKey)
	}
	if len(pinned) == 0 {
		return nil, fmt.Errorf("no usable sshd host keys found: %s", strings.Join(problems, "; "))
	}

	callback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		presented := key.Marshal()
		for _, p := range pinned {
			if bytes.Equal(p.Marshal(), presented) {
				return nil
			}
		}
		return fmt.Errorf("ssh: host key mismatch: presented key of type %s is not among the %d pinned sshd host keys", key.Type(), len(pinned))
	}
	return callback, nil
}

func PAMAuthenAndAuthor(r *http.Request, rc *RequestContext) error {

	username, passwd, authOK := r.BasicAuth()
	if authOK == false {
		glog.Warningf("[%s] User info not present", rc.ID)
		return httpError(http.StatusUnauthorized, "")
	}

	glog.Infof("[%s] Received user=%s", rc.ID, username)

	/*
	 * mgmt-framework container does not have access to /etc/passwd, /etc/group,
	 * /etc/shadow and /etc/tacplus_conf files of host. One option is to share
	 * /etc of host with /etc of container. For now disable this and use ssh
	 * for authentication.
	 */
	/* err := PAMAuthUser(username, passwd)
	    if err != nil {
			log.Printf("Authentication failed. user=%s, error:%s", username, err.Error())
	        return err
	    }*/

	hostKeyCallback, err := sshdHostKeyCallback(sshHostKeyPaths)
	if err != nil {
		glog.Errorf("[%s] Authentication unavailable: cannot load sshd host key: %v", rc.ID, err)
		return httpError(http.StatusUnauthorized, "")
	}

	//Use ssh for authentication.
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		HostKeyCallback: hostKeyCallback,
	}

	conn, err := net.DialTimeout("tcp", sshAuthAddr, sshAuthTimeout)
	if err != nil {
		glog.Infof("[%s] Failed to authenticate; %v", rc.ID, err)
		return httpError(http.StatusUnauthorized, "")
	}
	if err := conn.SetDeadline(time.Now().Add(sshAuthTimeout)); err != nil {
		conn.Close()
		glog.Infof("[%s] Failed to authenticate; %v", rc.ID, err)
		return httpError(http.StatusUnauthorized, "")
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, sshAuthAddr, config)
	if err != nil {
		conn.Close()
		glog.Infof("[%s] Failed to authenticate; %v", rc.ID, err)
		return httpError(http.StatusUnauthorized, "")
	}
	ssh.NewClient(sshConn, chans, reqs).Close()

	glog.Infof("[%s] Authentication passed. user=%s ", rc.ID, username)

	//Allow SET request only if user belong to admin group
	if isWriteOperation(r) && IsAdminGroup(username) == false {
		glog.Warningf("[%s] Not an admin; cannot allow %s", rc.ID, r.Method)
		return httpError(http.StatusForbidden, "Not an admin user")
	}

	glog.Infof("[%s] Authorization passed", rc.ID)
	return nil
}

// isWriteOperation checks if the HTTP request is a write operation
func isWriteOperation(r *http.Request) bool {
	m := r.Method
	return m == "POST" || m == "PUT" || m == "PATCH" || m == "DELETE"
}

// authMiddleware function creates a middleware for request
// authentication and authorization. This middleware will return
// 401 response if authentication fails and 403 if authorization
// fails.
func authMiddleware(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := getRouterConfig(r)
		if config == nil || !config.AuthEnable {
			inner.ServeHTTP(w, r)
			return
		}

		rc, r := GetContext(r)
		err := PAMAuthenAndAuthor(r, rc)
		if err != nil {
			writeErrorResponse(w, r, err)
		} else {
			inner.ServeHTTP(w, r)
		}
	})
}
