package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	services := []string{":9001", ":9002", ":9003"}
	balancer := newRoundRobinBalancer(services)

	proxy := newProxy(":9000", balancer)

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("received termination signal...")
		cancel()
		log.Println("shutting down now.")
		os.Exit(0)
	}()

	proxy.listen(ctx)
}
