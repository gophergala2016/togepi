package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gophergala2016/togepi/meta"
	"github.com/gophergala2016/togepi/redis"
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
	r   *redis.Redis
	md  *meta.Data
)

func init() {
	flag.Parse()
}

func shutdown() {
	log.Println("Shutting down gracefully..")

	if srv != nil {
		srv.Stop()
	}

	if r != nil {
		r.Close()
	}

	log.Println("terminating process")
	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		log.Println(err)
		shutdown()
	}
}

func startServer() {
	log.Println("starting server")
	var redisErr error
	r, redisErr = redis.NewClient(*redisHost, *redisDB)
	checkError(redisErr)

	sExists, sErr := r.KeyExists("secret")
	checkError(sErr)

	if !sExists {
		log.Println("running server for the first time")
		setErr := r.GenerateGlobalSecret()
		checkError(setErr)
	}

	getErr := r.RetrieveGlobalSecret()
	checkError(getErr)

	srv = server.New("/register", "/validate", *httpPort, r)
	startErr := srv.Start()
	checkError(startErr)
}

func startDaemon() {
	log.Println("starting daemon")

	configPath := os.Getenv("HOME") + "/.togepi/data"
	configStat, configStatErr := os.Stat(configPath)
	switch {
	case os.IsNotExist(configStatErr):
		log.Println("first start, generating configuration")

		resp, respErr := http.Get(*serverAddress + "/register")
		checkError(respErr)
		body, bodyErr := ioutil.ReadAll(resp.Body)
		checkError(bodyErr)
		resp.Body.Close()

		var respStruct server.RegResp
		jsonRespErr := json.Unmarshal(body, &respStruct)
		checkError(jsonRespErr)

		md.SetUserData(respStruct.UserID, respStruct.UserKey)
		dataErr := md.CreateDataFile(configPath)
		checkError(dataErr)
	case configStat.IsDir():
		log.Fatal(configPath + " is a directory")
	default:
		readDataErr := md.ReadDataFile(configPath)
		checkError(readDataErr)

		resp, respErr := http.Get(*serverAddress + "/validate?uid=" + md.UserID + "&ukey=" + md.UserKey)
		checkError(respErr)

		if resp.StatusCode != http.StatusOK {
			checkError(errors.New("invalid user"))
		}
	}
}

func shareFile() {

}

func main() {
	md = meta.NewData()
	if *serverMode {
		startServer()
	} else {
		if len(os.Args) > 1 && os.Args[1] == "start" {
			startDaemon()
		} else {
			shareFile()
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
