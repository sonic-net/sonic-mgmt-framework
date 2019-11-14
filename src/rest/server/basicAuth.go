
package server

import (
	"net/http"
	"github.com/golang/glog"
)
func BasicAuthenAndAuthor(r *http.Request, rc *RequestContext) error {

	username, passwd, authOK := r.BasicAuth()
	if authOK == false {
		glog.Errorf("[%s] User info not present", rc.ID)
		return httpError(http.StatusUnauthorized, "")
	}

	glog.Infof("[%s] Received user=%s", rc.ID, username)

	auth_success, err := UserPwAuth(username, passwd)
	if auth_success == false {
		glog.Infof("[%s] Failed to authenticate; %v", rc.ID, err)
		return httpError(http.StatusUnauthorized, "")	
	}


	glog.Infof("[%s] Authentication passed. user=%s ", rc.ID, username)

	//Allow SET request only if user belong to admin group
	if isWriteOperation(r) && IsAdminGroup(username) == false {
		glog.Errorf("[%s] Not an admin; cannot allow %s", rc.ID, r.Method)
		return httpError(http.StatusForbidden, "Not an admin user")
	}

	glog.Infof("[%s] Authorization passed", rc.ID)
	return nil
}