package main

import (
	"io"
	"log"
	"net"
)

func main() {
	services := []string{":9001", ":9002", ":9003"}
	upstreams := newUpstreamManager(services)

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalln("unable to listen on host interface(s):", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("unable to accept connection:", err)
		}
		go handleConnection(conn, upstreams)
	}
}

func handleConnection(conn net.Conn, upstreams *upstreamManager) {
	// release client when done
	defer conn.Close()

	// get next upstream connection (round robin)
	upstreamConn := upstreams.next()
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
