package main

import (
        "flag"
        "io/ioutil"
        "translib"
		"strings"
		//"sync"
		"github.com/Workiva/go-datastructures/queue"
		log "github.com/golang/glog"
)

func main() {
    var err error
    operationPtr := flag.String("o", "get", "Operation: create,update,replace,delete,get,getmodels,subscribe,supportsubscribe")
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
	} else if *operationPtr == "supportsubscribe" {
		paths := strings.Split(*uriPtr, ",")
		log.Info("Paths =", paths)

		resp, _ := translib.IsSubscribeSupported(paths)

		for i, path := range paths {
			log.Info("Response returned for path=", path)
			log.Info(resp[i])
		}

	} else if *operationPtr == "subscribe" {
		paths := strings.Split(*uriPtr, ",")
		log.Info("Paths =", paths)
		var q         *queue.PriorityQueue
		var stop       chan struct{}

		q = queue.NewPriorityQueue(1, false)
		stop = make(chan struct{}, 1)
		translib.Subscribe(paths, q, stop)
		log.Info("Subscribe completed")
		for {
			log.Info("Before calling Get")
			items, err := q.Get(1)
			log.Info("After calling Get")

			if items == nil {
				log.V(1).Infof("%v", err)
				break
			}
			if err != nil {
				log.V(1).Infof("%v", err)
				break
			}

			resp, _ := (items[0]).(*translib.SubscribeResponse)
			log.Info("SubscribeResponse received", resp)
		}

	} else {
		log.Info("Invalid Operation")
	}
}
