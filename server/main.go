package main

import (
	"flag"
	"log"

	"github.com/SergeyShpak/HNSearch/server/config"
	"github.com/SergeyShpak/HNSearch/server/initialization"
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
	s, err := initialization.InitServer(config)
	if err != nil {
		return err
	}
	err = s.ListenAndServe()
	if err != nil {
		return err
	}
	return err
}
