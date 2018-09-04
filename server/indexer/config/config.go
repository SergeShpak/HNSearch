package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Sorter *Sorter
}

type Sorter struct {
	Simple *SimpleSorter
}

type SimpleSorter struct {
	Buffer             uint64
	OutDir             string
	Parser             *Parser
	TmpCreationRetries int
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
		Sorter: &Sorter{
			Simple: &SimpleSorter{
				Buffer: 52428800,
				OutDir: "../sorted",
				Parser: &Parser{
					Simple: &SimpleParser{},
				},
				TmpCreationRetries: 10,
			},
		},
	}
	return c
}
