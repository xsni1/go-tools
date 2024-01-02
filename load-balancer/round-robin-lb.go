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

func (rrb *RoundRobinBalancer) Balance(r *http.Request) (*http.Response, error) {
	server := rrb.getNext()
	if server == nil {
		return nil, fmt.Errorf("No available servers")
	}

	parsedAddr, err := url.Parse(server.addr + r.URL.Path)
	if err != nil {
		return nil, fmt.Errorf("parsing addr err: %v", err)
	}

    slog.Info("Balancing request...", "address", parsedAddr, "strategy", "round-robin")
	r.RequestURI = ""
	r.URL = parsedAddr
	resp, err := rrb.client.Do(r)

	if err != nil {
		return nil, fmt.Errorf("balance err: %s", err)
	}
	return resp, nil
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
