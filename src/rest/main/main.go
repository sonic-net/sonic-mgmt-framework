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

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"rest/server"
	"swagger"
	"syscall"
	"time"
	"github.com/golang/glog"
	"github.com/pkg/profile"
)

// Command line parameters
var (
	port       int    // Server port
	uiDir      string // SwaggerUI directory
	certFile   string // Server certificate file path
	keyFile    string // Server private key file path
	caFile     string // Client CA certificate file path
	clientAuth string // Client auth mode
	jwtValInt  uint64    // JWT Valid Interval
	jwtRefInt  uint64    // JWT Refresh seconds before expiry
)

func init() {
	// Parse command line
	flag.IntVar(&port, "port", 443, "Listen port")
	flag.StringVar(&uiDir, "ui", "/rest_ui", "UI directory")
	flag.StringVar(&certFile, "cert", "", "Server certificate file path")
	flag.StringVar(&keyFile, "key", "", "Server private key file path")
	flag.StringVar(&caFile, "cacert", "", "CA certificate for client certificate validation")
	flag.StringVar(&clientAuth, "client_auth", "none", "Client auth mode - none|cert|user|jwt")
	flag.Uint64Var(&jwtRefInt, "jwt_refresh_int", 30, "Seconds before JWT expiry the token can be refreshed.")
	flag.Uint64Var(&jwtValInt, "jwt_valid_int", 3600, "Seconds that JWT token is valid for.")
	flag.Parse()
	// Suppress warning messages related to logging before flag parse
	flag.CommandLine.Parse([]string{})
}

var profRunning bool = true

// Start REST server
func main() {

	/* Enable profiling by default. Send SIGUSR1 signal to rest_server to
	 * stop profiling and save data to /tmp/profile<xxxxx>/cpu.pprof file.
	 * Copy over the cpu.pprof file and rest_server to a Linux host and run
	 * any of the following commands to generate a report in needed format.
	 * go tool pprof --txt ./rest_server ./cpu.pprof > report.txt
	 * go tool pprof --pdf ./rest_server ./cpu.pprof > report.pdf
	 * Note: install graphviz to generate the graph on a pdf format
	 */
	prof := profile.Start()
	defer prof.Stop()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	go func() {
	  for {
		<-sigs
		if (profRunning) {
			prof.Stop()
			profRunning = false
		} else {
			prof = profile.Start()
			defer prof.Stop()
			profRunning = true
		}
	  }
	}()

	swagger.Load()

	server.SetUIDirectory(uiDir)

	server.JwtRefreshInt = time.Duration(jwtRefInt*uint64(time.Second))
	server.JwtValidInt = time.Duration(jwtValInt*uint64(time.Second))

	if clientAuth == "user" {
		server.UserAuth.User = true
	}
	if clientAuth == "jwt" {
		server.UserAuth.Jwt = true
		server.GenerateJwtSecretKey()
	}
	if clientAuth == "cert" {
		server.UserAuth.Cert = true
	}

	router := server.NewRouter()

	address := fmt.Sprintf(":%d", port)

	// Prepare TLSConfig from the parameters
	tlsConfig := tls.Config{
		ClientAuth:               getTLSClientAuthType(),
		Certificates:             prepareServerCertificate(),
		ClientCAs:                prepareCACertificates(),
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         getPreferredCurveIDs(),
		PreferServerCipherSuites: true,
		CipherSuites:             getPreferredCipherSuites(),
	}

	// Prepare HTTPS server
	restServer := &http.Server{
		Addr:      address,
		Handler:   router,
		TLSConfig: &tlsConfig,
	}

	glog.Infof("Server started on %v", address)
	glog.Infof("UI directory is %v", uiDir)

	// Start HTTPS server
	glog.Fatal(restServer.ListenAndServeTLS("", ""))
}

// prepareServerCertificate function parses --cert and --key parameter
// values. Both cert and private key PEM files are loaded  into a
// tls.Certificate objects. Exits the process if files are not
// specified or not found or corrupted.
func prepareServerCertificate() []tls.Certificate {
	if certFile == "" {
		glog.Fatal("Server certificate file not specified")
	}

	if keyFile == "" {
		glog.Fatal("Server private key file not specified")
	}

	glog.Infof("Server certificate file: %s", certFile)
	glog.Infof("Server private key file: %s", keyFile)

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		glog.Fatal("Failed to load server cert/key -- ", err)
	}

	return []tls.Certificate{certificate}
}

// prepareCACertificates function parses --ca parameter, which is the
// path to CA certificate file. Loads file contents to a x509.CertPool
// object. Returns nil if file name is empty (not specified). Exists
// the process if file path is invalid or file is corrupted.
func prepareCACertificates() *x509.CertPool {
	if caFile == "" { // no CA file..
		return nil
	}

	glog.Infof("Client CA certificate file: %s", caFile)

	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		glog.Fatal("Failed to load CA certificate file -- ", err)
	}

	caPool := x509.NewCertPool()
	ok := caPool.AppendCertsFromPEM(caCert)
	if !ok {
		glog.Fatal("Invalid CA certificate")
	}

	return caPool
}

// getTLSClientAuthType function parses the --client_auth parameter.
// Returns corresponding tls.ClientAuthType value. Exits the process
// if value is not valid ('none', 'cert' or 'auth')
func getTLSClientAuthType() tls.ClientAuthType {
	switch clientAuth {
	case "none":
		return tls.RequestClientCert
	case "user":
		return tls.RequestClientCert
	case "cert":
		if caFile == "" {
			glog.Fatal("--cacert option is mandatory when --client_auth is 'cert'")
		}
		return tls.RequireAndVerifyClientCert
	case "jwt":
		return tls.RequestClientCert
	default:
		glog.Fatalf("Invalid '--client_auth' value '%s'. "+
			"Expecting one of 'none', 'cert', 'user' or 'jwt'", clientAuth)
		return tls.RequireAndVerifyClientCert // dummy
	}
}

func getPreferredCurveIDs() []tls.CurveID {
	return []tls.CurveID{
		tls.CurveP521,
		tls.CurveP384,
		tls.CurveP256,
	}
}

func getPreferredCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
}
