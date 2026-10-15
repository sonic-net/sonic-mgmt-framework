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
	"os/user"

	"github.com/golang/glog"
	//"github.com/msteinert/pam"
	"golang.org/x/crypto/ssh"
)

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

func PAMAuthenAndAuthor(r *http.Request, rc *RequestContext) error {

	// Requests arriving over the local REST Unix domain socket
	// (/var/run/rest-local.sock) carry a kernel-verified peer uid instead
	// of a Basic-Auth header. Klish/sonic-cli has no credentials to send,
	// but Klish also enforces no read/write privilege boundary of its own,
	// so the UDS path below bypasses only the password challenge -- it
	// still performs the same write-authorization check as remote users.
	if cred, ok := PeerCredFromRequest(r); ok {
		return udsAuthenAndAuthor(cred, r, rc)
	}

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

	//Use ssh for authentication.
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	_, err := ssh.Dial("tcp", "127.0.0.1:22", config)
	if err != nil {
		glog.Infof("[%s] Failed to authenticate; %v", rc.ID, err)
		return httpError(http.StatusUnauthorized, "")
	}

	glog.Infof("[%s] Authentication passed. user=%s ", rc.ID, username)

	//Allow SET request only if user belong to admin group
	if isWriteOperation(r) && IsAdminGroup(username) == false {
		glog.Warningf("[%s] Not an admin; cannot allow %s", rc.ID, r.Method)
		return httpError(http.StatusForbidden, "Not an admin user")
	}

	glog.Infof("[%s] Authorization passed", rc.ID)
	return nil
}

// udsAuthenAndAuthor authorizes a request that arrived over the local REST
// Unix domain socket. Identity comes solely from cred, the kernel-verified
// SO_PEERCRED uid captured by ConnContext -- no password challenge is
// performed, and no client-supplied value (header, env var, etc.) is ever
// consulted here. The uid is resolved to a host username via the host's
// bind-mounted /etc (see hostident.go), and write operations still require
// the same admin-group membership that PAMAuthenAndAuthor requires for
// authenticated remote users.
func udsAuthenAndAuthor(cred *PeerCred, r *http.Request, rc *RequestContext) error {

	username, ok := LookupHostUsername(cred.UID)
	if !ok {
		glog.Warningf("[%s] No host user found for local REST socket peer uid=%d", rc.ID, cred.UID)
		return httpError(http.StatusUnauthorized, "")
	}

	glog.Infof("[%s] Local REST socket request from uid=%d, user=%s", rc.ID, cred.UID, username)

	if isWriteOperation(r) && IsHostAdminGroup(username) == false {
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
