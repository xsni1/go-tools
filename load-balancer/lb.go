package main

import "net/http"

type LoadBalancer interface {
	Balance(r *http.Request) error
}

type LoadBalancerConfig struct {
	strategy          string
	servers           []string
	client            http.Client
	heartBeatInterval int
	heartBeatAddr     string
}

func NewLoadBalancer(config LoadBalancerConfig) LoadBalancer {
	switch config.strategy {
	case "round-robin":
		return &RoundRobinBalancer{
			servers:           config.servers,
			client:            config.client,
			heartBeatAddr:     config.heartBeatAddr,
			heartBeatInterval: config.heartBeatInterval,
		}
	default:
		return nil
	}
}
