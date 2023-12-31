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
	// server := rrb.serverPool.servers[rrb.count]
	addr := rrb.serverPool.servers[rrb.count].addr + "/health"
	rrb.count = (rrb.count + 1) % len(rrb.serverPool.servers)
	parsedAddr, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("parsing addr err: %v", err)
	}

	for _, v := range rrb.serverPool.servers {
		fmt.Println(v.addr, v.alive)
	}
	// fmt.Println(rrb.serverPool.servers)
	slog.Info("Balancing request", "target", parsedAddr, "strategy", "round-robin")

	r.RequestURI = ""
	r.URL = parsedAddr
	_, err = rrb.client.Do(r)

	if err != nil {
		return fmt.Errorf("balance err: %s", err)
	}
	return nil
}
