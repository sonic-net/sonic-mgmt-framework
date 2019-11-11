package server

import (
	"time"
	"encoding/json"
	jwt "github.com/dgrijalva/jwt-go"
	"net/http"
	"crypto/rand"
	"github.com/golang/glog"
	"strings"
)
var (
	JwtRefreshInt time.Duration
	JwtValidInt   time.Duration
	hmacSampleSecret = make([]byte, 16)
)
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}


type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type jwtToken struct {
	Token string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpIn int64 `json:"expires_in"`
}

func generateJWT(username string, expire_dt time.Time) string {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expire_dt.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(hmacSampleSecret)

	return tokenString
}
func GenerateJwtSecretKey() {
	rand.Read(hmacSampleSecret)
}

func tokenResp(w http.ResponseWriter, r *http.Request, username string) {
	exp_tm := time.Now().Add(JwtValidInt)
	token := jwtToken{Token: generateJWT(username, exp_tm), TokenType: "Bearer", ExpIn: int64(JwtValidInt/time.Second)}
	resp,err := json.Marshal(token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		status, data, ctype := prepareErrorResponse(httpError(http.StatusUnauthorized, err.Error()), r)
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(status)
		w.Write(data)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth_success, err := UserPwAuth(creds.Username, creds.Password)
	if  auth_success {
		tokenResp(w, r, creds.Username)
		return

		
	} else {
		status, data, ctype := prepareErrorResponse(httpError(http.StatusUnauthorized, ""), r)
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(status)
		w.Write(data)
		return
	}
}

func Refresh(w http.ResponseWriter, r *http.Request) {

	ctx,_ := GetContext(r)
	token, err := JwtAuthenAndAuthor(r, ctx)
	if err != nil {
		status, data, ctype := prepareErrorResponse(httpError(http.StatusUnauthorized, ""), r)
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(status)
		w.Write(data)
		return	
	}

	claims := &Claims{}
	jwt.ParseWithClaims(token.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > JwtRefreshInt {
		status, data, ctype := prepareErrorResponse(httpError(http.StatusBadRequest, ""), r)
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(status)
		w.Write(data)
		return
	}
	tokenResp(w, r, claims.Username)

}

func JwtAuthenAndAuthor(r *http.Request, rc *RequestContext) (jwtToken, error) {
	var token jwtToken
	auth_hdr := r.Header.Get("Authorization")
	if len(auth_hdr) == 0 {
		glog.Errorf("[%s] JWT Token not present", rc.ID)
		return token, httpError(http.StatusUnauthorized, "JWT Token not present")
	}
	auth_parts := strings.Split(auth_hdr, " ")
	if len(auth_parts) != 2 || auth_parts[0] != "Bearer" {
		glog.Errorf("[%s] Bad Request", rc.ID)
		return token, httpError(http.StatusBadRequest, "Bad Request")
	}

	token.Token = auth_parts[1]

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			glog.Errorf("[%s] Failed to authenticate, Invalid JWT Signature", rc.ID)
			return token, httpError(http.StatusUnauthorized, "Invalid JWT Signature")
			
		}
		glog.Errorf("[%s] Bad Request", rc.ID)
		return token, httpError(http.StatusBadRequest, "Bad Request")
	}
	if !tkn.Valid {
		glog.Errorf("[%s] Failed to authenticate, Invalid JWT Token", rc.ID)
		return token, httpError(http.StatusUnauthorized, "Invalid JWT Token")
	}
	return token, nil
}

