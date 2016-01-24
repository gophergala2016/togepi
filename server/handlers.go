package server

import (
	"encoding/json"
	"log"
	"net/http"

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
