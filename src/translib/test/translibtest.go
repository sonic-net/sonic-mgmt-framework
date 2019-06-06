package main

import (
        "flag"
        "io/ioutil"
        "translib"
		log "github.com/golang/glog"
)

func main() {
    var err error
    operationPtr := flag.String("o", "get", "Operation: create,update,replace,delete,get,getmodels")
    uriPtr := flag.String("u", "", "URI string")
    payloadFilePtr := flag.String("p", "", "JSON payload file")
    flag.Parse()
    log.Info("operation =", *operationPtr)
    log.Info("uri =", *uriPtr)
    log.Info("payload =", *payloadFilePtr)

    payloadFromFile, err := ioutil.ReadFile(*payloadFilePtr)
    if err != nil {
        log.Fatal(err)
    }

	if *operationPtr == "create" {
		req := translib.SetRequest{Path:*uriPtr, Payload:[]byte(payloadFromFile)}
		translib.Create(req)
	} else if *operationPtr == "update" {
		req := translib.SetRequest{Path:*uriPtr, Payload:[]byte(payloadFromFile)}
		translib.Update(req)
	} else if *operationPtr == "replace" {
		req := translib.SetRequest{Path:*uriPtr, Payload:[]byte(payloadFromFile)}
		translib.Replace(req)
	} else if *operationPtr == "delete" {
		req := translib.SetRequest{Path:*uriPtr}
		translib.Delete(req)
	} else if *operationPtr == "get" {
		req := translib.GetRequest{Path:*uriPtr}
		translib.Get(req)
	} else if *operationPtr == "getmodels" {
		models,_ := translib.GetModels()
		log.Info("Models =", models)
	} else {
		log.Info("Invalid Operation")
	}
}
