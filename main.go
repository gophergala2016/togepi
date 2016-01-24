package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gophergala2016/togepi/meta"
	"github.com/gophergala2016/togepi/redis"
	"github.com/gophergala2016/togepi/server"
	"github.com/gophergala2016/togepi/tcp"
	"github.com/gophergala2016/togepi/util"
)

var (
	serverMode        = flag.Bool("server", false, "run in server mode")
	httpServerAddress = flag.String("http-host", "http://127.0.0.1:8011", "togepi server's host")
	tcpServerAddress  = flag.String("tcp-host", "127.0.0.1:8012", "togepi server's host")
	socketPort        = flag.Int("socket-port", 8013, "a port to be used for local inter-process communication")
	httpPort          = flag.Int("http-port", 8011, "HTTP server's port")
	tcpPort           = flag.Int("tcp-port", 8012, "TCP server's port")
	redisHost         = flag.String("redis-host", "127.0.0.1:6379", "Redis host address")
	redisDB           = flag.Int("redis-db", 0, "Redis DB")
)

var (
	srv *server.Server
	r   *redis.Redis
	md  *meta.Data
	l   *tcp.Listener
	cl  *tcp.Client
)

func init() {
	flag.Parse()
}

func shutdown() {
	if srv != nil {
		srv.Stop()
	}

	if r != nil {
		r.Close()
	}

	if l != nil {
		l.Stop()
	}

	if cl != nil {
		cl.Close()
	}

	os.Exit(0)
}

func startServer() {
	log.Println("starting server")
	var redisErr error
	r, redisErr = redis.NewClient(*redisHost, *redisDB)
	util.CheckError(redisErr, shutdown)

	sExists, sErr := r.KeyExists("secret")
	util.CheckError(sErr, shutdown)

	if !sExists {
		log.Println("running server for the first time")
		setErr := r.GenerateGlobalSecret()
		util.CheckError(setErr, shutdown)
	}

	getErr := r.RetrieveGlobalSecret()
	util.CheckError(getErr, shutdown)

	srv = server.New("/register", "/validate", "/file", *httpPort, r)
	startErr := srv.Start()
	util.CheckError(startErr, shutdown)

	var lErr error
	l, lErr = tcp.NewListener(*tcpPort, nil)
	util.CheckError(lErr, shutdown)

	l.Start()
}

func startDaemon() {
	log.Println("starting daemon")

	configPath := os.Getenv("HOME") + "/.togepi/data"
	configStat, configStatErr := os.Stat(configPath)

	md.ConfPath = configPath

	switch {
	case os.IsNotExist(configStatErr):
		log.Println("first start, generating configuration")

		resp, respErr := http.Get(*httpServerAddress + "/register")
		util.CheckError(respErr, shutdown)
		body, bodyErr := ioutil.ReadAll(resp.Body)
		util.CheckError(bodyErr, shutdown)
		resp.Body.Close()

		var respStruct server.RegResp
		jsonRespErr := json.Unmarshal(body, &respStruct)
		util.CheckError(jsonRespErr, shutdown)

		md.SetUserData(respStruct.UserID, respStruct.UserKey)
		dataErr := md.CreateDataFile(configPath)
		util.CheckError(dataErr, shutdown)
	case configStat.IsDir():
		util.CheckError(errors.New(configPath+" is a directory"), shutdown)
	default:
		readDataErr := md.ReadDataFile(configPath)
		util.CheckError(readDataErr, shutdown)

		resp, respErr := http.Get(*httpServerAddress + "/validate?uid=" + md.UserID + "&ukey=" + md.UserKey)
		util.CheckError(respErr, shutdown)

		if resp.StatusCode != http.StatusOK {
			util.CheckError(errors.New("invalid user"), shutdown)
		}
	}

	var clErr error
	cl, clErr = tcp.NewClient(md.UserID, *tcpServerAddress)
	util.CheckError(clErr, shutdown)

	cl.HandleServerCommands()

	var lErr error
	l, lErr = tcp.NewListener(*socketPort, md)
	util.CheckError(lErr, shutdown)

	l.AcceptConnections(*httpServerAddress, md.UserID, md.UserKey)
}

func shareFile(filePath string) (err error) {
	configPath := os.Getenv("HOME") + "/.togepi/data"
	_, err = os.Stat(configPath)
	if err != nil {
		return
	}

	err = md.ReadDataFile(configPath)
	if err != nil {
		return
	}

	err = tcp.SendAndClose(*socketPort, []byte("SHARE::"+filePath))
	if err != nil {
		return
	}

	pathHash := util.Encrypt(filePath, md.UserKey)
	fmt.Println(md.UserID + pathHash)

	return
}

func main() {
	if *serverMode {
		startServer()
	} else {
		if len(os.Args) > 1 {
			md = meta.NewData()

			if os.Args[1] == "start" {
				startDaemon()
			} else {
				filePath := os.Args[1]

				fileStat, fileStatErr := os.Stat(filePath)
				util.CheckError(fileStatErr, shutdown)

				if fileStat.IsDir() {
					util.CheckError(errors.New(filePath+" is a directory"), shutdown)
				}

				currentDir, currentDirErr := os.Getwd()
				util.CheckError(currentDirErr, shutdown)

				if string(filePath[0]) != "/" {
					filePath = currentDir + "/" + filePath
				}

				shareErr := shareFile(filePath)
				util.CheckError(shareErr, shutdown)
				shutdown()
			}
		} else {
			util.CheckError(errors.New("please provide required arguments"), shutdown)
		}
	}

	// Shutting down on SIGINT.
	go func() {
		intChan := make(chan os.Signal)
		signal.Notify(intChan, os.Interrupt)

		<-intChan
		log.Println("Shutting down gracefully")
		go shutdown()

		fmt.Println("send SIGINT again to kill")
		<-intChan

		os.Exit(1)
	}()

	select {}
}
