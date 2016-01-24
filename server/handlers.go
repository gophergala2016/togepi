package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gophergala2016/togepi/util"
)

func (s *Server) regHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	uID, uIDErr := util.RandomString(16)
	if uIDErr != nil {
		returnError(uIDErr.Error(), http.StatusInternalServerError, w)
		return
	}

	uKey, uKeyErr := util.RandomString(16)
	if uKeyErr != nil {
		returnError(uKeyErr.Error(), http.StatusInternalServerError, w)
		return
	}

	respB, respBErr := json.Marshal(RegResp{
		UserID:  uID,
		UserKey: uKey,
	})

	if respBErr != nil {
		returnError(respBErr.Error(), http.StatusInternalServerError, w)
		return
	}

	addErr := s.r.AddUser(uID, uKey)
	if addErr != nil {
		returnError(addErr.Error(), http.StatusInternalServerError, w)
		return
	}

	log.Printf("registering new user: %s\n", uID)

	w.WriteHeader(http.StatusOK)
	w.Write(respB)
}

func (s *Server) validateHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("uid")
	userKey := r.FormValue("ukey")

	e, eErr := s.r.KeyExists(userID)
	if eErr != nil {
		returnError("failed to validate user ID", http.StatusBadRequest, w)
		return
	}
	if !e {
		returnError("user doesn't exist", http.StatusBadRequest, w)
		return
	}

	redisKey, redisErr := s.r.GetHashValue(userID, "key")
	if redisErr != nil {
		returnError("failed to validate user ID", http.StatusBadRequest, w)
		return
	}

	if redisKey != userKey {
		returnError("user doesn't exist", http.StatusBadRequest, w)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) fileHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	hash := r.FormValue("hash")
	user := r.FormValue("user")

	var failed bool

	switch action {
	case "add":
		redisErr := s.r.AddFileHash(user, hash)
		if redisErr != nil {
			returnError(redisErr.Error(), http.StatusBadRequest, w)
			failed = true
			break
		}
	case "request":
		shareHash := r.FormValue("sh")

		uID := shareHash[:32]
		ePath := shareHash[32:]

		rData, rErr := s.r.GetHashValue(uID, "files")
		if rErr != nil {
			returnError(rErr.Error(), http.StatusBadRequest, w)
			failed = true
			break
		}

		filesSl := strings.Split(rData, ",")
		var exists bool
		for _, v := range filesSl {
			if ePath == v {
				exists = true
				break
			}
		}

		if !exists {
			returnError("requested file doesn't exist", http.StatusBadRequest, w)
			return
		}

		conn, connErr := s.tcpListener.GetConnection(uID)
		if connErr != nil {
			returnError(connErr.Error(), http.StatusBadRequest, w)
			return
		}

		clientIP := strings.Split(conn.Conn.RemoteAddr().String(), ":")[0]

		var tcpAddr *net.TCPAddr
		tcpAddr, AddrErr := net.ResolveTCPAddr("tcp4", clientIP+":"+conn.SocketPort)
		if AddrErr != nil {
			returnError(AddrErr.Error(), http.StatusBadRequest, w)
			return
		}

		connected := make(chan bool)
		var tcpErr error
		var tcpConn *net.TCPConn
		go func() {
			tcpConn, tcpErr = net.DialTCP("tcp", nil, tcpAddr)
			connected <- true
		}()

		timeout := make(chan bool)
		go func() {
			time.Sleep(2 * time.Second)
			timeout <- true
		}()

		handleFirewall := func() {
			fmt.Println("++++>> firewall!")
			//returnError(tcpErr.Error(), http.StatusBadRequest, w)
		}

		handleDirect := func() {
			tcpConn.Write(append([]byte("PING::"), '\n'))
			tcpConn.Close()

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("NOFW::" + clientIP + ":" + conn.ProviderPort))
		}

		select {
		case <-timeout:
			handleFirewall()
		case <-connected:
			if tcpErr != nil {
				handleFirewall()
			} else {
				handleDirect()
			}
		}

	default:
		returnError("invalid action", http.StatusBadRequest, w)
		failed = true
	}
	if failed {
		return
	}
}
