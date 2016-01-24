package meta

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Data contains local user's data
type Data struct {
	ConfPath   string
	TCPServer  string
	HTTPServer string
	UserID     string
	UserKey    string
	Files      map[string]File
}

// File contains shared file's info.
type File struct {
	Path      string
	Timestamp int64
}

// NewData returns new Data.
func NewData() *Data {
	return &Data{
		Files: make(map[string]File),
	}
}

// SetServerData sets server info.
func (d *Data) SetServerData(tcpHost, httpHost string) {
	d.TCPServer = tcpHost
	d.HTTPServer = httpHost
}

// SetUserData sets user id and secret key.
func (d *Data) SetUserData(id, key string) {
	d.UserID = id
	d.UserKey = key
}

// AddFile adds a new shared file.
func (d *Data) AddFile(hash, path string) {
	d.Files[hash] = File{
		Path:      path,
		Timestamp: time.Now().UnixNano() / 1000000,
	}

	d.RewriteDataFile()
}

// RemoveFile removes a shared file.
func (d *Data) RemoveFile(hash string) {
	delete(d.Files, hash)
	d.RewriteDataFile()
}

// RewriteDataFile rewrites the user's data file.
func (d *Data) RewriteDataFile() {
	b, marshalErr := json.Marshal(*d)
	if marshalErr != nil {
		log.Println("failed to rewrite the data file:", marshalErr.Error())
		return
	}

	writeErr := ioutil.WriteFile(d.ConfPath, b, 0660)
	if writeErr != nil {
		log.Println("failed to rewrite the data file:", writeErr.Error())
		return
	}
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
