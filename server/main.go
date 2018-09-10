package main

import (
	"log"

	"github.com/SergeyShpak/HNSearch/server/config"
	"github.com/SergeyShpak/HNSearch/server/initialization"
)

func main() {
	config, err := config.Read("config.json")
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
