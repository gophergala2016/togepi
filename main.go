package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gophergala2016/togepi/server"
)

var (
	serverMode    = flag.Bool("server", false, "run in server mode")
	serverAddress = flag.String("server-host", "http://127.0.0.1:8011", "togepi server's host")
	httpPort      = flag.Int("http-port", 8011, "HTTP server's port")
	redisHost     = flag.String("redis-host", "127.0.0.1:6379", "Redis host address")
	redisDB       = flag.Int("redis-db", 0, "Redis DB")
)

var (
	srv *server.Server
)

func init() {
	flag.Parse()
}

func shutdown() {
	log.Println("Shutting down gracefully..")

	if srv != nil {
		srv.Stop()
	}

	log.Println("terminating process")
	os.Exit(0)
}

func main() {
	if *serverMode {
		log.Println("starting server")
		srv = server.New("/register", *httpPort)
		startErr := srv.Start()
		if startErr != nil {
			log.Fatal(startErr)
		}
	} else {
		if len(os.Args) > 1 && os.Args[1] == "start" {
			log.Println("starting daemon")

			configPath := os.Getenv("HOME") + "/.togepi/data"
			configStat, configStatErr := os.Stat(configPath)
			switch {
			case os.IsNotExist(configStatErr):
				log.Println("first start, generating configuration")

				resp, respErr := http.Get(*serverAddress + "/register")
				if respErr != nil {
					log.Fatal(respErr)
				}
				body, bodyErr := ioutil.ReadAll(resp.Body)
				if bodyErr != nil {
					log.Fatal(bodyErr)
				}
				resp.Body.Close()

				var respStruct server.RegResp
				jsonRespErr := json.Unmarshal(body, &respStruct)
				if jsonRespErr != nil {
					log.Fatal(jsonRespErr)
				}

				fmt.Println(respStruct)
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
