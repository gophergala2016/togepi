package meta

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Data contains local user's data
type Data struct {
	UserID  string
	UserKey string
	Files   map[string]File
}

// File contains shared file's info.
type File struct {
	Time int64
}

// NewData returns new Data.
func NewData() *Data {
	return &Data{
		Files: make(map[string]File),
	}
}

// SetUserData sets user id and secret key.
func (d *Data) SetUserData(id, key string) {
	d.UserID = id
	d.UserKey = key
}

// ReadDataFile writes a content of the data file into the structure.
func (d *Data) ReadDataFile(path string) (err error) {
	var data []byte
	data, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, d)
	if err != nil {
		return
	}

	return
}

// CreateDataFile creates user's data file.
func (d *Data) CreateDataFile(path string) (err error) {
	err = os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return
	}

	var f *os.File
	f, err = os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	var b []byte
	b, err = json.Marshal(*d)
	if err != nil {
		return
	}

	_, err = f.Write(b)
	if err != nil {
		return
	}

	return
}
