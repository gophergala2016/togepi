package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
)

// Server contains server's settings.
type Server struct {
	regEndpoint string
	port        int
	listener    net.Listener
}

// New returns new server.
func New(regEndpoint string, port int) *Server {
	return &Server{
		regEndpoint: regEndpoint,
		port:        port,
	}
}

// RegResp defines registration response structure.
type RegResp struct {
	UserID  string
	UserKey string
}

func randomString(len int) (str string, err error) {
	b := make([]byte, len)
	_, err = io.ReadFull(rand.Reader, b)
	if err != nil {
		return
	}
	str = hex.EncodeToString(b)
	return
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	returnError := func(msg string, statusCode int) {
		w.WriteHeader(statusCode)
		w.Write([]byte(fmt.Sprintf(`{"msg":"%s","status":%d}`, msg, statusCode)))
	}

	uID, uIDErr := randomString(16)
	if uIDErr != nil {
		returnError(uIDErr.Error(), http.StatusInternalServerError)
	}

	uKey, uKeyErr := randomString(16)
	if uKeyErr != nil {
		returnError(uKeyErr.Error(), http.StatusInternalServerError)
	}

	respB, respBErr := json.Marshal(RegResp{
		UserID:  uID,
		UserKey: uKey,
	})

	if respBErr != nil {
		returnError(respBErr.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respB)
}

// Start starts the HTTP server.
func (s *Server) Start() (err error) {
	http.HandleFunc(s.regEndpoint, regHandler)

	s.listener, err = net.Listen("tcp", ":"+strconv.Itoa(s.port))
	if err != nil {
		return
	}

	go http.Serve(s.listener, nil)

	return
}

// Stop stops the server.
func (s *Server) Stop() {
	s.listener.Close()
}
