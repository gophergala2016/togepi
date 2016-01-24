package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gophergala2016/togepi/redis"
	"github.com/gophergala2016/togepi/util"
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
		}
	default:
		returnError("invalid action", http.StatusBadRequest, w)
		failed = true
	}
	if failed {
		return
	}

	w.WriteHeader(http.StatusOK)
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
