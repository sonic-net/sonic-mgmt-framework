package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestMain(t *testing.T) {
	go main()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	fmt.Println("Listening on sig kill from TestMain")
	<-sigs
	fmt.Println("Returning from TestMain on sig kill")
}

