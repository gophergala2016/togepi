package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
)

var (
	serverMode = flag.Bool("server", false, "run in server mode")
)

func init() {
	flag.Parse()
}

func shutdown() {
	log.Println("Shutting down gracefully..")

	log.Println("terminating process")
	os.Exit(0)
}

func main() {
	if *serverMode {
		fmt.Println("server")
	} else {
		fmt.Println("client")
	}

	// Shutting down on SIGINT.
	go func() {
		intChan := make(chan os.Signal)
		signal.Notify(intChan, os.Interrupt)
		<-intChan
		shutdown()
	}()

	select {}
}
