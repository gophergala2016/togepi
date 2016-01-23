package tcp

import (
	"bufio"
	"net"
)

// Client contains TCP client's data.
type Client struct {
	TCPConn *net.TCPConn
	close   chan bool
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
		reader:  bufio.NewReader(tcpConn),
	}

	return
}

// Close closes the client connection.
func (c *Client) Close() {
	c.TCPConn.Close()
	c.close <- true
}
