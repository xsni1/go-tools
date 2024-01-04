package main

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// I think it would be nicer if implementation-specific state was not stored here
// but with this project's structure it cannot be changed now
type Server struct {
	addr           string
	alive          bool
	enabled        bool
	weight         int
	leftWeight     int
	connections    int
	healthEndpoint string
	healthInterval int
	mux            sync.RWMutex
}

func (s *Server) SetAlive(alive bool) {
    s.mux.Lock()
	s.alive = alive
    s.mux.Unlock()
}

func (s *Server) IsAlive() bool {
    defer s.mux.RUnlock()
    s.mux.RLock()
    return s.alive
}

func (s *Server) DecrementLeftWeight() {
    s.mux.Lock()
    s.leftWeight--
    s.mux.Unlock()
}

func (s *Server) ResetLeftWeight() {
    s.mux.Lock()
    s.leftWeight = s.weight
    s.mux.Unlock()
}

func (s *Server) GetLeftWeight() int {
    defer s.mux.RUnlock()
    s.mux.RLock()
    return s.leftWeight
}

type ServerPool struct {
    // Has to be a slice of pointers because of mux
    // Is there a nicer way?
	servers []*Server
}

func (sp *ServerPool) RunHeartBeats() {
	for i := range sp.servers {
		go sp.heartBeat(sp.servers[i])
	}
}

func (sp *ServerPool) heartBeat(server *Server) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(server.healthInterval))

	for {
		select {
		case <-ticker.C:
			slog.Debug("Heart beat...")
			resp, err := http.Get(server.addr + server.healthEndpoint)

			if err != nil {
				slog.Debug("heart beat to: %s, err: %v", server.addr, err)
				server.SetAlive(false)
			} else if resp.StatusCode == 200 {
				server.SetAlive(true)
			} else {
				server.SetAlive(false)
			}
		}
	}
}
