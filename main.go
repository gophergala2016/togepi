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
		if len(os.Args) > 1 && os.Args[1] == "start" {
			log.Println("starting daemon..")
			configPath := os.Getenv("HOME") + "/.togepi/data"
			configStat, configStatErr := os.Stat(configPath)
			switch {
			case os.IsNotExist(configStatErr):
				log.Println("first start, generating configuration..")
				// get tokens and save to configuration
			case configStat.IsDir():
				log.Fatal(configPath + " is a directory")
			}
		} else {
			//share
		}
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
