package main

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
)

const maxConnections = 10000
const maxConcurrent = 1000

type proxy struct {
	addr        string
	listener    net.Listener
	balancer    *roundRobinBalancer
	clientWg    *sync.WaitGroup
	clientConns chan net.Conn
}

func newProxy(addr string, balancer *roundRobinBalancer) *proxy {
	var wg sync.WaitGroup
	return &proxy{
		addr:        addr,
		balancer:    balancer,
		clientWg:    &wg,
		clientConns: make(chan net.Conn, maxConnections),
	}
}

func (p *proxy) gracefulShutdown() {
	p.listener.Close()
	p.clientWg.Wait()
	p.balancer.shutdown()
}

func (p *proxy) listen(ctx context.Context) {
	p.spawnConnectionHandlers()

	// serves as proxy for client connections chan and allows shutting down gracefully
	newClientChan := make(chan net.Conn, 1)
	go func() {
		defer close(p.clientConns)
		for {
			select {
			case conn := <-newClientChan:
				p.clientConns <- conn
			case <-ctx.Done():
				log.Println("waiting for in-flight connections...")
				p.gracefulShutdown()
				return
			}
		}
	}()

	var err error
	p.listener, err = net.Listen("tcp", p.addr)
	if err != nil {
		log.Fatalln("unable to listen on host interface", err)
	}

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			log.Println("listener no longer accepting connections", err)
			return
		}
		p.clientWg.Add(1)
		newClientChan <- conn
	}
}

func (p *proxy) spawnConnectionHandlers() {
	for i := 0; i < maxConcurrent; i++ {
		go func() {
			for conn := range p.clientConns {
				p.proxyConnection(conn)
			}
		}()
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
