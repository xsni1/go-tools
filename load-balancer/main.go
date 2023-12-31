package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.String("p", "8888", "lb port")
	strategy := flag.String("strategy", "round-robin", "lb strategy")
	heartBeatInterval := flag.Int("heart-beat-interval", 10000, "heart beat interval in ms")
	heartBeatAddr := flag.String("heart-beat-addr", "/health", "heart beat endpoint")
	flag.Var(&servers, "backend", "backends to load balance")
	flag.Parse()

	client := http.Client{}
	serverPool := ServerPool{
		servers:           servers,
		heartBeatInterval: *heartBeatInterval,
		heartBeatAddr:     *heartBeatAddr,
	}
	lb := NewLoadBalancer(LoadBalancerConfig{
		strategy:   *strategy,
		client:     client,
		serverPool: serverPool,
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received")
		err := lb.Balance(r)
		if err != nil {
			log.Printf("err: %v", err)
		}
	})

	go serverPool.HeartBeat()
	log.Printf("Server running, %s", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)
	if err != nil {
		log.Fatalf("err ListenAndServer: %s", err)
	}
}
