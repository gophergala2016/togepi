package server

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

// Server contains server's settings.
type Server struct {
	endpoint string
	port     int
	listener net.Listener
}

// New returns new server.
func New(endpoint string, port int) *Server {
	return &Server{
		endpoint: endpoint,
		port:     port,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() (err error) {
	http.HandleFunc(s.endpoint, func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("incomming connection")
	})

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
