package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gophergala2016/togepi/redis"
	"github.com/gophergala2016/togepi/util"
)

// Server contains server's settings.
type Server struct {
	regEndpoint string
	port        int
	listener    net.Listener
	r           *redis.Redis
}

// New returns new server.
func New(regEndpoint string, port int, r *redis.Redis) *Server {
	return &Server{
		regEndpoint: regEndpoint,
		port:        port,
		r:           r,
	}
}

// RegResp defines registration response structure.
type RegResp struct {
	UserID  string
	UserKey string
}

func (s *Server) regHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	returnError := func(msg string, statusCode int) {
		w.WriteHeader(statusCode)
		w.Write([]byte(fmt.Sprintf(`{"msg":"%s","status":%d}`, msg, statusCode)))
	}

	uID, uIDErr := util.RandomString(16)
	if uIDErr != nil {
		returnError(uIDErr.Error(), http.StatusInternalServerError)
	}

	uKey, uKeyErr := util.RandomString(16)
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

	addErr := s.r.AddUser(uID, uKey)
	if addErr != nil {
		returnError(addErr.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respB)
}

// Start starts the HTTP server.
func (s *Server) Start() (err error) {
	http.HandleFunc(s.regEndpoint, s.regHandler)

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
