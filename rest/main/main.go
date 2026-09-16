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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azure/sonic-mgmt-framework/build/rest_server/dist/openapi"
	"github.com/Azure/sonic-mgmt-framework/rest/server"
	"github.com/golang/glog"
	"github.com/pkg/profile"
)

// restLocalSocketPath is the local Unix domain socket used by sonic-cli
// (and other local processes in the same container) to reach the REST
// server without TLS or a password challenge. Its path is fixed rather
// than a flag because the native Klish REST client (rest_cl.cpp) already
// hardcodes this exact path.
const restLocalSocketPath = "/var/run/rest-local.sock"

// Command line parameters
var (
	port       int    // Server port
	certFile   string // Server certificate file path
	keyFile    string // Server private key file path
	caFile     string // Client CA certificate file path
	clientAuth string // Client auth mode

	// readTimeout is the deadline for receiving a full request (TLS+header+body)
	// once the connection is made. Value 0 indicates no timeout.
	readTimeout time.Duration = 15 * time.Second
)

func init() {
	// Parse command line
	flag.IntVar(&port, "port", 443, "Listen port")
	flag.String("ui", "", "UI directory - deprecated")
	flag.StringVar(&certFile, "cert", "", "Server certificate file path")
	flag.StringVar(&keyFile, "key", "", "Server private key file path")
	flag.StringVar(&caFile, "cacert", "", "CA certificate for client certificate validation")
	flag.StringVar(&clientAuth, "client_auth", "none", "Client auth mode - none|cert|user")
	flag.DurationVar(&readTimeout, "readtimeout", readTimeout, "Maximum duration for reading entire request")
	flag.Parse()
}

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
		<-sigs
		prof.Stop()
	}()

	openapi.Load()

	rtrConfig := server.RouterConfig{}
	if clientAuth == "user" {
		rtrConfig.AuthEnable = true
	}
	if ip := findAManagementIP(); ip != "" {
		rtrConfig.ServerAddr = fmt.Sprintf("https://%s:%d", ip, port)
	}

	router := server.NewRouter(rtrConfig)

	address := fmt.Sprintf(":%d", port)

	// Prepare TLSConfig from the parameters
	tlsConfig := tls.Config{
		ClientAuth:               getTLSClientAuthType(),
		Certificates:             prepareServerCertificate(),
		ClientCAs:                prepareCACertificates(),
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}

	// Prepare HTTPS server
	restServer := &http.Server{
		Addr:        address,
		Handler:     router,
		TLSConfig:   &tlsConfig,
		ReadTimeout: readTimeout,
		ErrorLog:    serverLog,
		ConnContext: server.ConnContext,
	}

	if glog.V(1) {
		glog.Infof("Read timeout = %v", readTimeout)
		glog.Infof("Authentication modes = %v", clientAuth)
	}

	// Start the local Unix domain socket listener alongside the TCP/TLS
	// listener, serving the same router/handler. A failure here must not
	// prevent the remote TCP/TLS listener (and its Basic Auth requirement)
	// from starting -- it is only logged.
	if udsListener, err := prepareUnixSocketListener(restLocalSocketPath); err != nil {
		glog.Errorf("Local REST unix socket disabled: %v", err)
	} else {
		glog.Infof("Local REST unix socket listening on %s", restLocalSocketPath)
		go func() {
			if serveErr := restServer.Serve(udsListener); serveErr != nil && serveErr != http.ErrServerClosed {
				glog.Errorf("Local REST unix socket listener stopped: %v", serveErr)
			}
		}()
	}

	glog.Infof("Server started on %v", address)

	// Start HTTPS server
	glog.Fatal(restServer.ListenAndServeTLS("", ""))
}

// prepareUnixSocketListener removes a stale rest-local.sock left behind by
// a previous process (if any) and binds a fresh listener at path. Only a
// pre-existing Unix domain socket special file is ever removed -- any
// other file type found at path is left untouched, and listener setup
// fails instead of clobbering it. The socket is created with mode 0666;
// filesystem permissions are intentionally permissive here because
// authorization for requests on this socket is based on the kernel
// verified SO_PEERCRED uid (see server.ConnContext / IsHostAdminGroup),
// not on who can open the socket file.
func prepareUnixSocketListener(path string) (net.Listener, error) {
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, fmt.Errorf("%s exists and is not a socket; refusing to remove it", path)
		}
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("failed to remove stale socket %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat %s: %w", path, err)
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(path, 0666); err != nil {
		listener.Close()
		os.Remove(path)
		return nil, fmt.Errorf("failed to chmod %s: %w", path, err)
	}

	return listener, nil
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
	default:
		glog.Fatalf("Invalid '--client_auth' value '%s'. "+
			"Expecting one of 'none', 'cert' or 'user'", clientAuth)
		return tls.RequireAndVerifyClientCert // dummy
	}
}

// findAManagementIP returns a valid IPv4 address of eth0.
// Empty string is returned if no address could be resolved.
func findAManagementIP() string {
	var addrs []net.Addr
	eth0, err := net.InterfaceByName("eth0")
	if err == nil {
		addrs, err = eth0.Addrs()
	}
	if err != nil {
		glog.Errorf("Could not read eth0 info; err=%v", err)
		return ""
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err == nil && ip.To4() != nil {
			return ip.String()
		}
	}

	glog.Warning("Could not find a management address!!")
	return ""
}

// errorMsgPrefixes identifies the important error messages logged by
// the standard http library that should not be missed
var errorMsgPrefixes = [][]byte{
	[]byte("http: Accept error"),
	[]byte("http: panic"),
	[]byte("http2: panic"),
}

// serverLog forwards the http library's logs to glog
var serverLog = log.New(logWriter{}, "", 0)

type logWriter struct{}

func (logWriter) Write(p []byte) (int, error) {
	for _, pfx := range errorMsgPrefixes {
		if bytes.HasPrefix(p, pfx) {
			glog.Warning(string(p))
			return len(p), nil
		}
	}
	if glog.V(2) {
		glog.Info(string(p))
	}
	return len(p), nil
}
