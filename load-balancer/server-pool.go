package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	addr           string
	alive          bool
	enabled        bool
	weight         int
	healthEndpoint string
	healthInterval int
}

func (s *Server) SetAlive(alive bool) {
    s.alive = alive
}

type ServerPool struct {
	servers []*Server
}

func (sp *ServerPool) RunHeartBeats() {
	for _, server := range sp.servers {
        server := server
        // TODO: Passing index is so silly
		go sp.heartBeat(server)
	}
}

func (sp *ServerPool) heartBeat(server *Server) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(server.healthInterval))

	for {
		select {
		case <-ticker.C:
            fmt.Println("hb", server.addr, server.alive)
			slog.Debug("Heart beat")
			resp, err := http.Get(server.addr + server.healthEndpoint)

			if err != nil {
				slog.Debug("heart beat to: %s, err: %v", server.addr, err)
                server.SetAlive(false)
				// server.alive = false
			} else if resp.StatusCode == 200 {
				// server.alive = true
                server.SetAlive(true)
			} else {
				// server.alive = false
                server.SetAlive(false)
			}
			// sp.servers[idx] = server
		}
	}
}
