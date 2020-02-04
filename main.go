package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	services := []string{":9001", ":9002", ":9003"}
	balancer := newRoundRobinBalancer(services)

	proxy := newProxy(":9000", balancer)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("waiting for in-flight connections...")
		proxy.gracefulShutdown()
		log.Println("shutting down now.")
		os.Exit(0)
	}()

	proxy.listen()
}
