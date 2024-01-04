package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type WeightedRoundRobinBalancer struct {
	count      int
	httpClient http.Client
	serverPool ServerPool
	mux        sync.RWMutex
}

func (lb *WeightedRoundRobinBalancer) Balance(r *http.Request) (*http.Response, error) {
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
	resp, err := lb.httpClient.Do(r)
	elapsed := time.Since(start)
	slog.Info("Balancing request...", "address", parsedAddr, "strategy", "weighted-round-robin", "execution time", elapsed)

	if err != nil {
		return nil, fmt.Errorf("balance err: %s", err)
	}
	return resp, nil
}

func (lb *WeightedRoundRobinBalancer) getNext() *Server {
	for i := 0; i < len(lb.serverPool.servers); i++ {
		server := lb.serverPool.servers[lb.getCount()]
		if server.IsAlive() {
			server.DecrementLeftWeight()
			if server.GetLeftWeight() <= 0 {
				server.ResetLeftWeight()
				lb.updateCount()
			}
			return server
		}
		lb.updateCount()
	}
	return nil
}

func (lb *WeightedRoundRobinBalancer) getCount() int {
    defer lb.mux.RUnlock()
    lb.mux.RLock()
    return lb.count
}

func (lb *WeightedRoundRobinBalancer) updateCount() {
	lb.mux.Lock()
	lb.count = (lb.count + 1) % len(lb.serverPool.servers)
	lb.mux.Unlock()
}
