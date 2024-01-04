package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type RoundRobinBalancer struct {
	count      int
	httpClient http.Client
	serverPool ServerPool
	mux        sync.RWMutex
}

func (lb *RoundRobinBalancer) Balance(r *http.Request) (*http.Response, error) {
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
	slog.Info("Balancing request...", "address", parsedAddr, "strategy", "round-robin", "execution time", elapsed)

	if err != nil {
		return nil, fmt.Errorf("balance err: %s", err)
	}
	return resp, nil
}

func (lb *RoundRobinBalancer) getNext() *Server {
	for i := 0; i < len(lb.serverPool.servers); i++ {
		server := lb.serverPool.servers[lb.getCount()]
		lb.updateCount()
		if server.IsAlive() {
			return server
		}
	}
	return nil
}

func (lb *RoundRobinBalancer) getCount() int {
	defer lb.mux.RUnlock()
	lb.mux.RLock()
	return lb.count
}

func (lb *RoundRobinBalancer) updateCount() {
	lb.mux.Lock()
	lb.count = (lb.count + 1) % len(lb.serverPool.servers)
	lb.mux.Unlock()
}
