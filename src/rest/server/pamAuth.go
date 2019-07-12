package server

import (
	"log"
	"net/http"
    "os/user"
	"errors"
	"github.com/msteinert/pam"
    "golang.org/x/crypto/ssh"
)

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
/*	if err == nil {
		fmt.Println("PAM Authentication succeeded!")
	} else {
		fmt.Println("PAM Authentication failed!")
	} */
	return err
}

func IsAdminGroup (username string) bool {
    
	usr, err := user.Lookup(username)
    if err != nil {
        return false
    }
    gids, err := usr.GroupIds()
    if err != nil {
        return false 
    }
	log.Printf("User:%s, groups=%s", username, gids)
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

func PAMAuthenAndAuthor(w http.ResponseWriter, r *http.Request) error {

	username, passwd, authOK := r.BasicAuth()
    if authOK == false {
        http.Error(w, "Not authorized", 401)
        return errors.New("user not present")
    }
	log.Printf("Received user=%s, pass=%s", username, passwd)

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
        log.Printf("Failed to authenticate: ", err)
        return err
    }

	log.Printf("Authentication passed. user=%s ", username)

    //Check if user belong to admin group.
    adminGrp := IsAdminGroup(username)
	log.Printf("User:%s, isAdmGrp:%d", username, adminGrp)
    //Allow SET request only if user belong to admin group
    if adminGrp == false && r.Method == "SET" {
        http.Error(w, "Not authorized", 401)
        return errors.New("authorization error")
    }
    return nil 
}
