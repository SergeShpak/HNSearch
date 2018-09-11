package main

import (
	"flag"
	"log"

	"github.com/SergeyShpak/HNSearch/indexer/config"
	"github.com/SergeyShpak/HNSearch/indexer/server"
)

var flagConfigFile string

func main() {
	log.Println("Start of main")
	flag.StringVar(&flagConfigFile, "config", "/etc/HNIndexer/config.json", "path for HNIndexer config file")
	flag.Parse()
	config, err := config.Read(flagConfigFile)
	if err != nil {
		log.Println("Error during configuration reading: ", err)
		return
	}
	if err := run(config); err != nil {
		log.Println("ERROR: ", err)
	}
	return
}

func run(config *config.Config) error {
	s, err := server.InitServer(config)
	if err != nil {
		return err
	}
	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return err
}
