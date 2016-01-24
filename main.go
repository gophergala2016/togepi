package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/gophergala2016/togepi/meta"
	"github.com/gophergala2016/togepi/redis"
	"github.com/gophergala2016/togepi/server"
	"github.com/gophergala2016/togepi/tcp"
	"github.com/gophergala2016/togepi/util"
)

var (
	httpServerAddress = flag.String("http-host", "127.0.0.1:8011", "togepi server's host (without protocol)")
	tcpServerAddress  = flag.String("tcp-host", "127.0.0.1:8012", "togepi server's host")
	socketPort        = flag.Int("socket-port", 8013, "a port to be used for local inter-process communication")
	httpPort          = flag.Int("http-port", 8011, "HTTP server's port")
	tcpPort           = flag.Int("tcp-port", 8012, "TCP server's port")
	providerPort      = flag.Int("provider-port", 8014, "provider port")
	redisHost         = flag.String("redis-host", "127.0.0.1:6379", "Redis host address")
	redisDB           = flag.Int("redis-db", 0, "Redis DB")

	serverMode = flag.Bool("server", false, "run in server mode")
	daemonMode = flag.Bool("start", false, "run in daemon mode")
	showShared = flag.Bool("a", false, "List all shared files")
)

var (
	srv *server.Server
	p   *server.Provider
	r   *redis.Redis
	md  *meta.Data
	l   *tcp.Listener
	cl  *tcp.Client
)

func init() {
	flag.Parse()
}

func shutdown() {
	if srv != nil {
		srv.Stop()
	}

	if r != nil {
		r.Close()
	}

	if l != nil {
		l.Stop()
	}

	if cl != nil {
		cl.Close()
	}

	if p != nil {
		p.Stop()
	}

	os.Exit(0)
}

func main() {
	if *serverMode {
		startServer()
	} else if *showShared {
		md = meta.NewData()
		err := readConfig()
		if err != nil {
			util.CheckError(err, shutdown)
		}

		for k, v := range md.Files {
			fmt.Println(md.UserID+k, v.Path)
		}

		shutdown()
	} else {
		if len(os.Args) > 1 {
			md = meta.NewData()

			if *daemonMode {
				startDaemon()
			} else {
				filePath := os.Args[1]

				err := readConfig()
				if err != nil {
					util.CheckError(err, shutdown)
				}

				*tcpServerAddress = md.TCPServer
				*httpServerAddress = md.HTTPServer

				if len(filePath) == 96 && !strings.Contains(filePath, "/") {
					requestFile(filePath)
					shutdown()
				} else {
					fileStat, fileStatErr := os.Stat(filePath)
					util.CheckError(fileStatErr, shutdown)

					if fileStat.IsDir() {
						util.CheckError(errors.New(filePath+" is a directory"), shutdown)
					}

					currentDir, currentDirErr := os.Getwd()
					util.CheckError(currentDirErr, shutdown)

					if string(filePath[0]) != "/" {
						filePath = currentDir + "/" + filePath
					}

					shareErr := shareFile(filePath)
					util.CheckError(shareErr, shutdown)
					shutdown()
				}
			}
		} else {
			util.CheckError(errors.New("please provide required arguments"), shutdown)
		}
	}

	// Shutting down on SIGINT.
	go func() {
		intChan := make(chan os.Signal)
		signal.Notify(intChan, os.Interrupt)

		<-intChan
		log.Println("Shutting down gracefully")
		go shutdown()

		fmt.Println("send SIGINT again to kill")
		<-intChan

		os.Exit(1)
	}()

	select {}
}
