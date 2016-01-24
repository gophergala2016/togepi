package server

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gophergala2016/togepi/meta"
)

// Provider contains server's settings.
type Provider struct {
	endpoint string
	port     int
	listener net.Listener
	md       *meta.Data
}

// NewProvider returns new server.
func NewProvider(endpoint string, port int, md *meta.Data) *Provider {
	return &Provider{
		endpoint: endpoint,
		port:     port,
		md:       md,
	}
}

func (p *Provider) providerHandler(w http.ResponseWriter, r *http.Request) {
	shareHash := r.FormValue("sh")

	filePath := p.md.Files[shareHash[32:]].Path

	log.Println("uploading file:", filePath)

	_, configStatErr := os.Stat(filePath)
	if configStatErr != nil {
		returnError("file doesn't exist", http.StatusBadRequest, w)
		return
	}

	var data []byte
	data, dataErr := ioutil.ReadFile(filePath)
	if dataErr != nil {
		returnError(dataErr.Error(), http.StatusBadRequest, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(append([]byte(filepath.Base(filePath)+"::"), data...))
}

// Start starts the HTTP server.
func (p *Provider) Start() (err error) {
	http.HandleFunc(p.endpoint, p.providerHandler)

	p.listener, err = net.Listen("tcp", ":"+strconv.Itoa(p.port))
	if err != nil {
		return
	}

	go http.Serve(p.listener, nil)

	return
}

// Stop stops the server.
func (p *Provider) Stop() {
	p.listener.Close()
}
