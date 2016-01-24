package tcp

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gophergala2016/togepi/meta"
)

// Client contains TCP client's data.
type Client struct {
	TCPConn *net.TCPConn
	close   chan bool
	ackChan chan bool
	reader  *bufio.Reader
	md      *meta.Data
}

// NewClient returns new TCP client connection.
func NewClient(md *meta.Data, serverAddress string, sockerPort, providerPort int) (client *Client, err error) {
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp4", serverAddress)
	if err != nil {
		return
	}

	var tcpConn *net.TCPConn
	tcpConn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return
	}

	tcpConn.Write(append([]byte(md.UserID+"::"+strconv.Itoa(sockerPort)+"::"+strconv.Itoa(providerPort)), '\n'))

	client = &Client{
		TCPConn: tcpConn,
		close:   make(chan bool),
		ackChan: make(chan bool),
		reader:  bufio.NewReader(tcpConn),
		md:      md,
	}

	return
}

// HandleServerCommands receives and handles server's commands.
func (c *Client) HandleServerCommands() {
	go func() {
		var closed bool
		for {
			result, err := c.reader.ReadBytes('\n')

			select {
			case <-c.close:
				closed = true
			default:
			}
			if closed {
				c.ackChan <- true
				break
			}

			if err != nil {
				if err == io.EOF {
					log.Println("connection is closed by server")
					break
					// TODO: re-connect
				} else {
					log.Println("failed to process server's command")
					continue
				}
			}

			resSl := strings.Split(strings.Split(string(result), "\n")[0], "::")

			command := resSl[0]
			loadData := resSl[1]

			switch command {
			case "GET":
				//TODO: handle errors
				filePath := c.md.Files[loadData].Path
				log.Println("uploading file", filePath)

				fileStat, fileStatErr := os.Stat(filePath)
				if fileStatErr != nil {
					c.TCPConn.Write(append([]byte("ERROR"), '\n'))
					break
				}

				var data []byte
				data, dataErr := ioutil.ReadFile(filePath)
				if dataErr != nil {
					c.TCPConn.Write(append([]byte("ERROR"), '\n'))
					break
				}

				c.TCPConn.Write(append([]byte(filepath.Base(filePath)+"\n"+strconv.FormatInt(fileStat.Size(), 10)+"\n"), data...))
			}
		}
	}()
}

// Close closes the client connection.
func (c *Client) Close() {
	c.TCPConn.Close()
	c.close <- true
	<-c.ackChan
}
