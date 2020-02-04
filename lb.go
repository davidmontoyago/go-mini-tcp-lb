package main

import (
	"log"
	"sync"
)

type roundRobinBalancer struct {
	roundRobinIndex int
	upstreams       []*upstream
	mux             sync.Mutex
}

func newRoundRobinBalancer(addresses []string) *roundRobinBalancer {
	upstreams := make([]*upstream, len(addresses))
	for i, addr := range addresses {
		log.Println("registering upstream", addr)
		upstreams[i] = newUpstream(addr)
	}
	return &roundRobinBalancer{roundRobinIndex: 0, upstreams: upstreams}
}

func (u *roundRobinBalancer) next() *upstream {
	u.mux.Lock()
	defer u.mux.Unlock()

	nextIndex := u.roundRobinIndex % len(u.upstreams)
	nextUpstream := u.upstreams[nextIndex]
	log.Println("next upstream is", nextIndex, nextUpstream.addr)

	u.roundRobinIndex++

	return nextUpstream
}

func (u *roundRobinBalancer) shutdown() {
	for _, upstream := range u.upstreams {
		upstream.disconnectAll()
	}
}
