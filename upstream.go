package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type upstream struct {
	addr  string
	conns []net.Conn
	mux   sync.Mutex
}

func (u *upstream) releaseConn(conn net.Conn) {
	u.mux.Lock()
	defer u.mux.Unlock()

	u.conns = append(u.conns, conn)
	log.Printf("%s pool status %#v", u.addr, u.conns)
}

func (u *upstream) takeConn() net.Conn {
	u.mux.Lock()
	defer u.mux.Unlock()

	if len(u.conns) > 0 {
		log.Println("capturing connection for ", u.addr)
		conn := u.conns[0]
		u.conns = u.conns[1:]
		return conn
	}
	return newConn(u.addr)
}

func (u *upstream) close() {
	for _, conn := range u.conns {
		log.Printf("closing conn %#v\n", conn)
		err := conn.Close()
		log.Println(err)
	}
}

type upstreamManager struct {
	roundRobinIndex int
	upstreams       []*upstream
	mux             sync.Mutex
}

func newUpstreamManager(addresses []string) *upstreamManager {
	upstreams := make([]*upstream, len(addresses))
	for i, addr := range addresses {
		log.Println("registering upstream ", addr)
		upstreams[i] = &upstream{
			addr:  addr,
			conns: nil,
		}
	}
	u := &upstreamManager{roundRobinIndex: 0, upstreams: upstreams}
	return u
}

func (u *upstreamManager) next() net.Conn {
	u.mux.Lock()
	nextIndex := u.roundRobinIndex % len(u.upstreams)
	nextUpstream := u.upstreams[nextIndex]
	log.Println("next upstream is", nextIndex, nextUpstream.addr)
	u.roundRobinIndex++
	u.mux.Unlock()
	return nextUpstream.takeConn()
}

func newConn(addr string) net.Conn {
	log.Println("creating new connection for ", addr)
	// TODO pass IP
	port, _ := strconv.Atoi(addr[1:])
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{Port: port})
	if err != nil {
		log.Fatalln("unable to connect to upstream:", err)
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(5 * time.Second)
	conn.SetDeadline(time.Time{})
	return conn
}

func (u *upstreamManager) release(conn net.Conn) {
	isSameAddr := func(addr string) bool {
		return strings.HasSuffix(conn.RemoteAddr().String(), addr)
	}
	for _, upstream := range u.upstreams {
		if isSameAddr(upstream.addr) {
			log.Println("releasing connection for ", upstream.addr)
			upstream.releaseConn(conn)
			break
		}
	}
}

func (u *upstreamManager) cleanUp() {
	for _, upstream := range u.upstreams {
		upstream.close()
	}
}
