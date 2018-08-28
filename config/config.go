package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server       *Server
	QueryHandler *QueryHandler
}

type Server struct {
	Port int
}

type QueryHandler struct {
	File string
	Type string
}

func Read(path string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(file).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
