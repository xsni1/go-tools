package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func (b *backend) Set(val string) error {
	*b = append(*b, val)
	return nil
}

func (b *backend) String() string {
	return strings.Join(*b, " ")
}

type backend []string

var backends backend

type BackendPool struct {
	backends   backend
	strategy   string
	lastServed int
}

func (bp *BackendPool) Balance() {
	backend := bp.backends[bp.lastServed]
	log.Printf("Routing to: %s", backend)
	bp.lastServed++
    if bp.lastServed >= len(bp.backends) {
        bp.lastServed = 0
    }
}

func main() {
	port := flag.String("p", "8888", "lb port")
	strategy := flag.String("strategy", "round-robin", "lb strategy")
	flag.Var(&backends, "backend", "backends to load balance")
	flag.Parse()

	backendPool := BackendPool{backends: backends, strategy: *strategy}

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received")
		backendPool.Balance()
	})

	log.Printf("Server running, %s", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)
	if err != nil {
		log.Fatalf("err ListenAndServer: %s", err)
	}
}
