package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type WeightedRoundRobinBalancer struct {
	count      int
	httpClient http.Client
	serverPool ServerPool
}

func (rrb *WeightedRoundRobinBalancer) Balance(r *http.Request) (*http.Response, error) {
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
	resp, err := rrb.httpClient.Do(r)
	elapsed := time.Since(start)
	slog.Info("Balancing request...", "address", parsedAddr, "strategy", "weighted-round-robin", "execution time", elapsed)

	if err != nil {
		return nil, fmt.Errorf("balance err: %s", err)
	}
	return resp, nil
}

func (rrb *WeightedRoundRobinBalancer) getNext() *Server {
	for i := 0; i < len(rrb.serverPool.servers); i++ {
		server := &rrb.serverPool.servers[rrb.count]
		if server.alive {
			server.leftWeight--
			if server.leftWeight <= 0 {
				server.leftWeight = server.weight
				rrb.count = (rrb.count + 1) % len(rrb.serverPool.servers)
				slog.Info("", "COUNT", rrb.count)
			}
			return server
		}
		rrb.count = (rrb.count + 1) % len(rrb.serverPool.servers)
	}
	return nil
}
