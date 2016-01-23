package util

import (
	"crypto/rand"
	"encoding/hex"
	"io"
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
