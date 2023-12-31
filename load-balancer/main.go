package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	port := flag.String("p", "8888", "lb port")
	strategy := flag.String("strategy", "round-robin", "lb strategy")
	heartBeatInterval := flag.Int("heart-beat-interval", 10000, "heart beat interval in ms")
	heartBeatAddr := flag.String("heart-beat-addr", "/health", "heart beat endpoint")
	flag.Var(&servers, "backend", "backends to load balance")
	flag.Parse()

	client := http.Client{}
	serverPool := NewServerPool(ServerPoolConfig{
		heartBeatInterval: *heartBeatInterval,
		heartBeatAddr:     *heartBeatAddr,
		addrs:             servers,
	})
	lb := NewLoadBalancer(LoadBalancerConfig{
		serverPool: serverPool,
		strategy:   *strategy,
		client:     client,
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Request received")
		err := lb.Balance(r)
		if err != nil {
			slog.Error("", "error", err)
		}
	})

	go serverPool.HeartBeat()
	slog.Info("Server running", "port", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)

	if err != nil {
        slog.Error("Error ListenAndServe", "error", err)
        os.Exit(1)
	}
}
