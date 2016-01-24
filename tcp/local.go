package tcp

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
)

// SendAndClose sends data to specified port and closes the connection.
func SendAndClose(port int, data []byte) (err error) {
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp4", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return
	}

	var tcpConn *net.TCPConn
	tcpConn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return
	}

	tcpConn.Write(append(data, '\n'))

	tcpConn.Close()

	return
}

// AcceptConnections handles a connection and disconnects.
func (l *Listener) AcceptConnections() {
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
				log.Println("failed to process data from master:" + resErr.Error())
				continue
			}

			data := strings.Split(result, "\n")[0]

			log.Printf("local command received: %s\n", data)

			tcpConn.Close()
		}
	}()
}
