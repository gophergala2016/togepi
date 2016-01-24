package server

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gophergala2016/togepi/redis"
)

// Server contains server's settings.
type Server struct {
	regEndpoint      string
	validateEndpoint string
	fileEndpoint     string
	port             int
	listener         net.Listener
	r                *redis.Redis
}

// New returns new server.
func New(regEndpoint, validateEndpoint, fileEndpoint string, port int, r *redis.Redis) *Server {
	return &Server{
		regEndpoint:      regEndpoint,
		validateEndpoint: validateEndpoint,
		fileEndpoint:     fileEndpoint,
		port:             port,
		r:                r,
	}
}

// RegResp defines registration response structure.
type RegResp struct {
	UserID  string
	UserKey string
}

func returnError(msg string, statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(`{"msg":"%s","status":%d}`, msg, statusCode)))
}

// Start starts the HTTP server.
func (s *Server) Start() (err error) {
	http.HandleFunc(s.regEndpoint, s.regHandler)
	http.HandleFunc(s.validateEndpoint, s.validateHandler)
	http.HandleFunc(s.fileEndpoint, s.fileHandler)

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
