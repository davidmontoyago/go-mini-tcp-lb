package main

import (
	"io"
	"log"
	"net"
	"sync"
)

type proxy struct {
	addr     string
	listener net.Listener
	balancer *roundRobinBalancer
	clientWg *sync.WaitGroup
}

func newProxy(addr string, balancer *roundRobinBalancer) *proxy {
	var wg sync.WaitGroup
	return &proxy{
		addr:     addr,
		balancer: balancer,
		clientWg: &wg,
	}
}

func (p *proxy) gracefulShutdown() {
	p.listener.Close()
	p.clientWg.Wait()
	p.balancer.shutdown()
}

func (p *proxy) listen() {
	var err error
	p.listener, err = net.Listen("tcp", p.addr)
	if err != nil {
		log.Fatalln("unable to listen on host interface(s):", err)
	}

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			log.Println("listener no longer accepting connections.", err)
			break
		}
		p.clientWg.Add(1)
		go p.proxyConnection(conn)
	}
}

func (p *proxy) proxyConnection(clientConn net.Conn) {
	upstream := p.balancer.next()
	upstreamConn := upstream.connect()

	defer p.clientWg.Done()
	defer upstream.disconnect(upstreamConn)
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go stream(clientConn, upstreamConn, &wg)
	go stream(upstreamConn, clientConn, &wg)
	wg.Wait()
}

func stream(src net.Conn, dst net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	wtn, err := io.Copy(dst, src)
	if err != nil {
		log.Println("failed proxy'ing tcp stream", err)
	}
	log.Println("wrote", wtn, "bytes to dst.")
}
