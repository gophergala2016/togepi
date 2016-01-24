package tcp

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/gophergala2016/togepi/meta"
)

// Listener contains TCP listener's data.
type Listener struct {
	tcpListener *net.TCPListener
	done        chan bool
	connections map[string]*ClientConn
	md          *meta.Data
}

// ClientConn contains client info.
type ClientConn struct {
	Conn       *net.TCPConn
	SocketPort string
}

// NewListener returns new TCP listener.
func NewListener(port int, md *meta.Data) (l *Listener, err error) {
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		return
	}

	var tcpListener *net.TCPListener
	tcpListener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return
	}

	l = &Listener{
		tcpListener: tcpListener,
		done:        make(chan bool),
		connections: make(map[string]*ClientConn),
		md:          md,
	}

	return
}

// GetConnection returns client connection.
func (l *Listener) GetConnection(uID string) (conn *ClientConn, err error) {
	conn, e := l.connections[uID]
	if !e {
		err = errors.New("client is not connected")
		return
	}
	return
}

// Start makes the TCP listener to start accepting incoming connections.
func (l *Listener) Start() {
	go func() {
		var closed bool

		for {
			tcpConn, tcpErr := l.tcpListener.AcceptTCP()

			select {
			case <-l.done:
				closed = true
			default:
			}
			if closed {
				break
			}

			if tcpErr != nil {
				log.Println("failed to establish TCP client connection")
				continue
			}

			result, resErr := bufio.NewReader(tcpConn).ReadString('\n')
			if resErr != nil {
				log.Println("failed to process data from client:" + resErr.Error())
				continue
			}

			resultSl := strings.Split(strings.Split(result, "\n")[0], "::")
			if len(resultSl) < 2 {
				log.Println("failed to process data from client:" + resErr.Error())
				tcpConn.Close()
				continue
			}

			clientID := resultSl[0]
			socketPort := resultSl[1]

			log.Printf("client %s connected\n", clientID)

			l.connections[clientID] = &ClientConn{
				Conn:       tcpConn,
				SocketPort: socketPort,
			}
		}
	}()
}

// Stop stops active TCP listener.
func (l *Listener) Stop() {
	l.tcpListener.Close()
	l.done <- true
}
