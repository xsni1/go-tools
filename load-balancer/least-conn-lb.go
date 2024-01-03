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

func (rrb *LeastConnBalancer) Balance(r *http.Request) (*http.Response, error) {
	server := rrb.getNext()
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
    server.connections++
	resp, err := rrb.httpClient.Do(r)
    server.connections--
	elapsed := time.Since(start)
	slog.Info("Balancing request...", "address", parsedAddr, "strategy", "least-conn", "execution time", elapsed)

	if err != nil {
		return nil, fmt.Errorf("balance err: %s", err)
	}
	return resp, nil
}

func (rrb *LeastConnBalancer) getNext() *Server {
	var server *Server

	// Instead of linear search it would probably be better to use min-heap?
	// No need to worry about perf though (for now)
	for idx := range rrb.serverPool.servers {
        v := &rrb.serverPool.servers[idx]
		if v.alive && server == nil {
			server = v
			continue
		}
		if server != nil && v.alive && v.connections < server.connections {
			server = v
			continue
		}
	}

	return server
}
