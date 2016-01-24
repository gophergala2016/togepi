package tcp

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gophergala2016/togepi/util"
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
func (l *Listener) AcceptConnections(httpServerAddress, userID, userKey string) {
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
				log.Println("failed to process local command:" + resErr.Error())
				tcpConn.Close()
				continue
			}

			data := strings.Split(strings.Split(result, "\n")[0], "::")

			if len(data) < 2 {
				log.Println("failed to process local command:" + resErr.Error())
				tcpConn.Close()
				continue
			}

			command := data[0]
			loadData := data[1]

			log.Printf("local %s command received: %s\n", command, loadData)

			var invalidCommand, procFailed bool
			var procErr error
			switch command {
			case "SHARE":
				pathHash := util.Encrypt(loadData, userKey)
				l.md.AddFile(pathHash, loadData)
				resp, procErr := http.Get(httpServerAddress + "/file?action=add&hash=" + pathHash + "&user=" + userID)
				if procErr != nil {
					procFailed = true
					break
				}
				if resp.StatusCode != http.StatusOK {
					procErr = errors.New("received " + strconv.Itoa(resp.StatusCode) + " status")
				}
			case "PING":
			default:
				invalidCommand = true
			}

			if invalidCommand {
				log.Println("invalid command:" + command)
				tcpConn.Close()
				continue
			} else if procFailed {
				log.Println("failed to process command:" + procErr.Error())
				tcpConn.Close()
				continue
			}

			tcpConn.Close()
		}
	}()
}
