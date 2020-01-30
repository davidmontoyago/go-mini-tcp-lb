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
	services := []string{":9001"}
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

func handleReq(src net.Conn, dst net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	buf := make([]byte, 1024)
	for {
		bytesRead, err := src.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Fatalln("failed to read from downstream:", err)
			}
			break
		}
		log.Println("read", bytesRead, "bytes from src.")

		bytesWtn, err := dst.Write(buf[:bytesRead])
		if err != nil {
			if err != io.EOF {
				log.Fatalln("failed to write to upstream", err)
			}
			break
		}
		log.Println("wrote", bytesWtn, "bytes to dst")
	}
}

func handleConnection(clientConn net.Conn, upstreams *upstreamManager, waitgroup *sync.WaitGroup) {
	// get next upstream connection (round robin)
	upstreamConn := upstreams.next()

	defer waitgroup.Done()
	// release upstream connection to the pool
	defer upstreams.release(upstreamConn)
	// release client
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go handleReq(clientConn, upstreamConn, &wg)
	go handleReq(upstreamConn, clientConn, &wg)
	wg.Wait()
}
