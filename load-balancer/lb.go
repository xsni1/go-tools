package main

import "net/http"

type LoadBalancer interface {
	Balance(r *http.Request) (*http.Response, error)
}

type LoadBalancerConfig struct {
	strategy   string
	client     http.Client
	serverPool ServerPool
}

func NewLoadBalancer(config LoadBalancerConfig) LoadBalancer {
	switch config.strategy {
	case "round-robin":
		return &RoundRobinBalancer{
			httpClient: config.client,
			serverPool: config.serverPool,
		}
	case "weighted-round-robin":
		return &WeightedRoundRobinBalancer{
			httpClient: config.client,
			serverPool: config.serverPool,
		}
	default:
		return nil
	}
}
