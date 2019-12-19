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

	if err := PopulateAuthStruct(username, &rc.Auth); err != nil {
		glog.Infof("[%s] Failed to retrieve authentication information; %v", rc.ID, err)
		return httpError(http.StatusUnauthorized, "")
	}

	glog.Infof("[%s] Authorization passed", rc.ID)
	return nil
}
