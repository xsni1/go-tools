package main

import (
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	addr           string
	alive          bool
	enabled        bool
	weight         int
	leftWeight     int
	healthEndpoint string
	healthInterval int
}

func (s *Server) SetAlive(alive bool) {
	s.alive = alive
}

type ServerPool struct {
	servers []Server
}

func (sp *ServerPool) RunHeartBeats() {
	for i := range sp.servers {
		go sp.heartBeat(&sp.servers[i])
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
