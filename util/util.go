package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
)

// RandomString returns random HEX of the predefined length.
func RandomString(len int) (str string, err error) {
	b := make([]byte, len)
	_, err = io.ReadFull(rand.Reader, b)
	if err != nil {
		return
	}
	str = hex.EncodeToString(b)
	return
}

type errorHandler func()

// CheckError executes the passed function if error is not nil.
func CheckError(err error, handler errorHandler) {
	if err != nil {
		log.Println(err)
		handler()
	}
}

// Encrypt encodes the data using SHA-256 + HMAC.
func Encrypt(data, key string) string {
	secretKey := []byte(key)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
