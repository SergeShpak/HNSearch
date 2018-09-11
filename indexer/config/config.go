package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server  *Server
	Indexer *Indexer
	Sorter  *Sorter
	Parser  *Parser
}

type Server struct {
	Port int
}

type Indexer struct {
	IndexesDir string
	DataDir    string
	Simple     *SimpleIndexer
}

type SimpleIndexer struct {
}

type Sorter struct {
	Simple *SimpleSorter
}

type SimpleSorter struct {
	Buffer             uint64
	OutDir             string
	TmpCreationRetries int
	TmpDir             string
}

type Parser struct {
	Simple *SimpleParser
}

type SimpleParser struct{}

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

func GetDefaultConfig() *Config {
	c := &Config{
		Server: &Server{
			Port: 8081,
		},
		Indexer: &Indexer{
			Simple: &SimpleIndexer{},
		},
		Sorter: &Sorter{
			Simple: &SimpleSorter{
				Buffer:             52428800,
				OutDir:             "../sorted",
				TmpCreationRetries: 10,
			},
		},
		Parser: &Parser{
			Simple: &SimpleParser{},
		},
	}
	return c
}
