package main

import (
	"log"
	"net"
	"strings"
)

type upstream struct {
	addr  string
	conns []net.Conn
}

type upstreamManager struct {
	roundRobinIndex int
	connections     []*upstream
}

func newUpstreamManager(addresses []string) *upstreamManager {
	conns := make([]*upstream, len(addresses))
	for i, addr := range addresses {
		log.Println("registering upstream ", addr)
		conns[i] = &upstream{
			addr:  addr,
			conns: nil,
		}
	}
	u := &upstreamManager{roundRobinIndex: 0, connections: conns}
	return u
}

func (u *upstreamManager) next() net.Conn {
	var conn net.Conn

	nextIndex := u.roundRobinIndex % len(u.connections)
	nextUpstream := u.connections[nextIndex]
	log.Println("next upstream is", nextIndex, nextUpstream.addr)

	if nextUpstream.conns != nil && len(nextUpstream.conns) > 0 {
		log.Println("capturing connection for ", nextUpstream.addr)
		conn = nextUpstream.conns[0]
		nextUpstream.conns = nextUpstream.conns[1:]
	} else {
		conn = newConn(nextUpstream.addr)
	}
	u.roundRobinIndex++
	return conn
}

func newConn(addr string) net.Conn {
	log.Println("creating new connection for ", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln("unable to connect to upstream:", err)
	}
	return conn
}

func (u *upstreamManager) close(conn net.Conn) {
	isSameAddr := func(addr string) bool {
		return strings.HasSuffix(conn.RemoteAddr().String(), addr)
	}
	for _, upstream := range u.connections {
		if isSameAddr(upstream.addr) {
			log.Println("releasing connection for ", upstream.addr)
			log.Println()
			upstream.conns = append(upstream.conns, conn)
			break
		}
	}
}
