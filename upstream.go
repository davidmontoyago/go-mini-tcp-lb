package main

import (
	"log"
	"net"
	"sync"
	"time"
)

type upstream struct {
	addr   string
	active []net.Conn
	mux    sync.Mutex
}

func newUpstream(addr string) *upstream {
	return &upstream{
		addr:   addr,
		active: nil,
	}
}

func (u *upstream) connect() net.Conn {
	u.mux.Lock()
	defer u.mux.Unlock()

	conn := newConn(u.addr)
	u.active = append(u.active, conn)
	log.Println("active connections for", u.addr, u.active)
	return conn
}

func (u *upstream) disconnect(conn net.Conn) {
	u.mux.Lock()
	defer u.mux.Unlock()
	defer conn.Close()

	for i, c := range u.active {
		if c == conn {
			u.active = append(u.active[:i], u.active[i+1:]...)
			break
		}
	}
}

func (u *upstream) disconnectAll() {
	u.mux.Lock()
	defer u.mux.Unlock()

	for _, c := range u.active {
		if err := c.Close(); err != nil {
			log.Println("failed closing connection", err)
		}
	}
}

func newConn(addr string) net.Conn {
	log.Println("creating new connection for ", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln("unable to connect to upstream", addr, err)
	}
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	return conn
}
