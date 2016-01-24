package tcp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

// Client contains TCP client's data.
type Client struct {
	TCPConn *net.TCPConn
	close   chan bool
	ackChan chan bool
	reader  *bufio.Reader
}

// NewClient returns new TCP client connection.
func NewClient(clientID, serverAddress string) (client *Client, err error) {
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

	tcpConn.Write(append([]byte(clientID), '\n'))

	client = &Client{
		TCPConn: tcpConn,
		close:   make(chan bool),
		ackChan: make(chan bool),
		reader:  bufio.NewReader(tcpConn),
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

			fmt.Println("===>>", string(result))

		}
	}()
}

// Close closes the client connection.
func (c *Client) Close() {
	c.TCPConn.Close()
	c.close <- true
	<-c.ackChan
}
