package main

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int
	Strategy string
	Servers  []struct {
		Address   string
		Weight    int
		HeartBeat struct {
			Interval int
			Endpoint string
		} `yaml:"heart-beat"`
	}
}

func readConfig() Config {
	var cfg Config
	file, err := os.ReadFile("./config.yaml")
	if err != nil {
		slog.Error("reading file", "error", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		slog.Error("unmarshaling yaml", "error", err)
		os.Exit(1)
	}

	return cfg
}
