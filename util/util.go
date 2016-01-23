package util

import (
	"crypto/rand"
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
