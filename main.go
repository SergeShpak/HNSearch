package main

import (
	"log"

	"github.com/SergeyShpak/HNSearch/server"
)

func main() {
	if err := run(); err != nil {
		log.Println("ERROR: ", err)
	}
	return
}

func run() error {
	s := server.InitServer()
	err := s.ListenAndServe()
	return err
}
