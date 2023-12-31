package main

import (
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	addr    string
	alive   bool
	enabled bool
}

type ServerPool struct {
	servers           []Server
	heartBeatAddr     string
	heartBeatInterval int
}

type ServerPoolConfig struct {
	addrs             []string
	heartBeatAddr     string
	heartBeatInterval int
}

func NewServerPool(config ServerPoolConfig) ServerPool {
	servers := []Server{}
	for _, addr := range config.addrs {
		servers = append(servers, Server{addr: addr})
	}

	return ServerPool{
		servers:           servers,
		heartBeatAddr:     config.heartBeatAddr,
		heartBeatInterval: config.heartBeatInterval,
	}
}

func (sp *ServerPool) HeartBeat() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(sp.heartBeatInterval))
	for {
		select {
		case <-ticker.C:
            slog.Debug("Heart beat")
			for idx, server := range sp.servers {
				resp, err := http.Get(server.addr + sp.heartBeatAddr)
				if err != nil {
					slog.Debug("heart beat to: %s, err: %v", server.addr, err)
					server.alive = false
				} else if resp.StatusCode == 200 {
					server.alive = true
				} else {
					server.alive = false
				}
				sp.servers[idx] = server
			}
		}
	}
}
