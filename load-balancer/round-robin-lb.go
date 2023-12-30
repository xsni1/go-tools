package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type RoundRobinBalancer struct {
	servers           []string
	count             int
	client            http.Client
	heartBeatInterval int
	heartBeatAddr     string
}

func (rrb *RoundRobinBalancer) Balance(r *http.Request) error {
	addr := servers[rrb.count] + "/health"
	rrb.count = (rrb.count + 1) % len(rrb.servers)

	parsedAddr, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("parsing addr err: %v", err)
	}

	r.RequestURI = ""
	r.URL = parsedAddr
	res, err := rrb.client.Do(r)

	if err != nil {
		return fmt.Errorf("balance err: %s", err)
	}
	log.Printf("%s\n%v\n", addr, res)
	return nil
}
