package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
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

// SaveFile saves the data into file.
func SaveFile(path string, data []byte) (err error) {
	var f *os.File
	f, err = os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return
	}

	return
}
