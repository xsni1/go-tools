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
			leftWeight:     server.Weight,
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

	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		msg := ""
		for _, v := range serverPool.servers {
			msg += fmt.Sprintf("address: %s, weight: %d, leftWeight: %d, alive: %t, conns: %d\n", v.addr, v.weight, v.leftWeight, v.alive, v.connections)
		}
		w.Write([]byte(msg))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Processing request...")
		resp, err := lb.Balance(r)
		if err != nil {
			slog.Error("", "error", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(resp.StatusCode)
	})

	go serverPool.RunHeartBeats()
	slog.Info("Server running", "port", cfg.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)

	if err != nil {
		slog.Error("Error ListenAndServe", "error", err)
		os.Exit(1)
	}
}
