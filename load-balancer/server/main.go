package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.String("p", "8080", "http server port")
	flag.Parse()

	s := http.NewServeMux()
	s.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received HEALTH request")
	})

	log.Printf("Server running, :%s", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), s)
	if err != nil {
		log.Fatalf("err ListenAndServe: %s", err)
	}
}
