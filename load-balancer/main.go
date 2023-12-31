package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg := readConfig()
	servers := []*Server{}
	for _, server := range cfg.Servers {
		servers = append(servers, &Server{
			addr:           server.Address,
			weight:         server.Weight,
			healthEndpoint: server.HeartBeat.Endpoint,
			healthInterval: server.HeartBeat.Interval,
			alive:          false,
			enabled:        true,
		})
	}

	client := http.Client{}
	serverPool := ServerPool{
		servers: servers,
	}
	lb := NewLoadBalancer(LoadBalancerConfig{
		serverPool: serverPool,
		client:     client,
		strategy:   cfg.Strategy,
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Request received")
		err := lb.Balance(r)
		if err != nil {
			slog.Error("", "error", err)
		}
	})

	go serverPool.RunHeartBeats()
	slog.Info("Server running", "port", cfg.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)

	if err != nil {
		slog.Error("Error ListenAndServe", "error", err)
		os.Exit(1)
	}
}
