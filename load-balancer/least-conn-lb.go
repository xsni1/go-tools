package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type LeastConnBalancer struct {
	count      int
	httpClient http.Client
	serverPool ServerPool
}

func (lb *LeastConnBalancer) Balance(r *http.Request) (*http.Response, error) {
	server := lb.getNext()
	if server == nil {
		return nil, fmt.Errorf("No available servers")
	}

	parsedAddr, err := url.Parse(server.addr + r.URL.Path)
	if err != nil {
		return nil, fmt.Errorf("parsing addr err: %v", err)
	}

	r.RequestURI = ""
	r.URL = parsedAddr
	start := time.Now()
    server.IncConnections()
	resp, err := lb.httpClient.Do(r)
    server.DecConnections()
	elapsed := time.Since(start)
	slog.Info("Balancing request...", "address", parsedAddr, "strategy", "least-conn", "execution time", elapsed)

	if err != nil {
		return nil, fmt.Errorf("balance err: %s", err)
	}
	return resp, nil
}

func (lb *LeastConnBalancer) getNext() *Server {
	var server *Server

	// Instead of linear search it would probably be better to use min-heap?
	// No need to worry about perf though (for now)
	for idx := range lb.serverPool.servers {
		v := lb.serverPool.servers[idx]
		if v.IsAlive() && server == nil {
			server = v
			continue
		}
		if server != nil && v.IsAlive() && v.GetConnections() < server.GetConnections() {
			server = v
			continue
		}
	}

	return server
}
