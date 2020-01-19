package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	services := []string{":9001", ":9002", ":9003"}
	upstreams := newUpstreamManager(services)

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalln("unable to listen on host interface(s):", err)
	}

	// wait for in-flight connections before terminating
	var waitgroup sync.WaitGroup
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		waitgroup.Wait()
		upstreams.cleanUp()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("unable to accept connection:", err)
		}
		waitgroup.Add(1)
		go handleConnection(conn, upstreams, &waitgroup)
	}
}

func handleConnection(conn net.Conn, upstreams *upstreamManager, waitgroup *sync.WaitGroup) {
	// release client when done
	defer conn.Close()

	// get next upstream connection (round robin)
	upstreamConn := upstreams.next()

	// defers are LIFO stacked
	defer waitgroup.Done()
	defer upstreams.close(upstreamConn)

	inputBuf := make([]byte, 1024)
	for {
		bytesRead, err := conn.Read(inputBuf)
		if err != nil {
			if err != io.EOF {
				log.Fatalln("failed to read from downstream:", err)
			}
			break
		}
		log.Println("read", bytesRead, "bytes.")
		upstreamConn.Write(inputBuf[:bytesRead])
	}
}
