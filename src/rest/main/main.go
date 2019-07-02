///////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Broadcom. All rights reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
//
///////////////////////////////////////////////////////////////////////

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"rest/server"
	"swagger"
        "os"
        "os/signal"
        "syscall"
        "github.com/pkg/profile"
)

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

	var port int
	var uiDir string

	flag.IntVar(&port, "port", 8080, "Listen port")
	flag.StringVar(&uiDir, "ui", "/usr/sonic-mgmt/ui", "UI directory")
	flag.Parse()

	swagger.Load()

	server.SetUIDirectory(uiDir)

	router := server.NewRouter()

	address := fmt.Sprintf(":%d", port)
	log.Printf("Server started on %v", address)
	log.Printf("UI directory is %v", uiDir)

	log.Fatal(http.ListenAndServe(address, router))
}
