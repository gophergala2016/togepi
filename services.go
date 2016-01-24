package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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

	var lErr error
	l, lErr = tcp.NewListener(*tcpPort, nil)
	util.CheckError(lErr, shutdown)

	srv = server.New("/register", "/validate", "/file", *httpPort, r, l)
	startErr := srv.Start()
	util.CheckError(startErr, shutdown)

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

		resp, respErr := http.Get("http://" + *httpServerAddress + "/register")
		util.CheckError(respErr, shutdown)
		body, bodyErr := ioutil.ReadAll(resp.Body)
		util.CheckError(bodyErr, shutdown)
		resp.Body.Close()

		var respStruct server.RegResp
		jsonRespErr := json.Unmarshal(body, &respStruct)
		util.CheckError(jsonRespErr, shutdown)

		md.SetUserData(respStruct.UserID, respStruct.UserKey)
		md.SetServerData(*tcpServerAddress, *httpServerAddress)
		dataErr := md.CreateDataFile(configPath)
		util.CheckError(dataErr, shutdown)
	case configStat.IsDir():
		util.CheckError(errors.New(configPath+" is a directory"), shutdown)
	default:
		readDataErr := md.ReadDataFile(configPath)
		util.CheckError(readDataErr, shutdown)

		*httpServerAddress = md.HTTPServer
		*tcpServerAddress = md.TCPServer

		resp, respErr := http.Get("http://" + *httpServerAddress + "/validate?uid=" + md.UserID + "&ukey=" + md.UserKey)
		util.CheckError(respErr, shutdown)

		if resp.StatusCode != http.StatusOK {
			util.CheckError(errors.New("invalid user"), shutdown)
		}
	}

	var clErr error
	cl, clErr = tcp.NewClient(md.UserID, *tcpServerAddress, *socketPort, *providerPort)
	util.CheckError(clErr, shutdown)

	cl.HandleServerCommands()

	var lErr error
	l, lErr = tcp.NewListener(*socketPort, md)
	util.CheckError(lErr, shutdown)

	p = server.NewProvider("/provide", *providerPort, md)
	pErr := p.Start()
	util.CheckError(pErr, shutdown)

	l.AcceptConnections(*httpServerAddress, md.UserID, md.UserKey)
}

func requestFile(shareHash string) (err error) {
	resp, respErr := http.Get("http://" + *httpServerAddress + "/file?action=request&sh=" + shareHash)
	util.CheckError(respErr, shutdown)
	body, bodyErr := ioutil.ReadAll(resp.Body)
	util.CheckError(bodyErr, shutdown)
	resp.Body.Close()

	bSl := strings.Split(string(body), "::")

	switch bSl[0] {
	case "NOFW":
		log.Println("pulling directly from client")
		fResp, respErr := http.Get("http://" + bSl[1] + "/provide?sh=" + shareHash)
		util.CheckError(respErr, shutdown)
		fBody, fBodyErr := ioutil.ReadAll(fResp.Body)
		util.CheckError(fBodyErr, shutdown)
		fResp.Body.Close()

		bodyStr := string(fBody)
		sep := "::"
		fileName := strings.Split(bodyStr, sep)[0]
		fileContents := bodyStr[len(fileName)+len(sep):]

		saveErr := util.SaveFile(fileName, []byte(fileContents))
		util.CheckError(saveErr, shutdown)

		log.Printf("file %s saved\n", fileName)
	}

	return
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
