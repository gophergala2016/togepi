package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gophergala2016/togepi/redis"
	"github.com/gophergala2016/togepi/server"
	"github.com/gophergala2016/togepi/tcp"
	"github.com/gophergala2016/togepi/util"
)

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
	err = readConfig()
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

func readConfig() (err error) {
	configPath := os.Getenv("HOME") + "/.togepi/data"
	_, err = os.Stat(configPath)
	if err != nil {
		return
	}

	err = md.ReadDataFile(configPath)
	if err != nil {
		return
	}

	return
}
