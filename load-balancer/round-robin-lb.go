package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type RoundRobinBalancer struct {
	count      int
	client     http.Client
	serverPool ServerPool
}

func (rrb *RoundRobinBalancer) Balance(r *http.Request) error {
	server := rrb.getNext()
	if server == nil {
		return fmt.Errorf("No available servers")
	}
	addr := server.addr + "/health"
	parsedAddr, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("parsing addr err: %v", err)
	}
	slog.Info("Balancing request", "target", parsedAddr, "strategy", "round-robin")

	r.RequestURI = ""
	r.URL = parsedAddr
	_, err = rrb.client.Do(r)

	if err != nil {
		return fmt.Errorf("balance err: %s", err)
	}
	return nil
}

func (rrb *RoundRobinBalancer) getNext() *Server {
	for i := 0; i < len(rrb.serverPool.servers); i++ {
		server := rrb.serverPool.servers[rrb.count]
        rrb.count = (rrb.count + 1) % len(rrb.serverPool.servers)
		if server.alive {
			// does it really have to be a pointer?
			return &server
		}
	}
	return nil
}
