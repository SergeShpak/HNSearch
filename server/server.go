package server

import (
	"net/http"
	"time"

	"github.com/SergeyShpak/HNSearch/server/handlers"
	"github.com/gorilla/mux"
)

func InitServer() *http.Server {
	r := newRouter()
	s := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s
}

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.Path("/1/queries/count/{date}").Methods("GET").Queries("size", "{size}").HandlerFunc(handlers.DatePopularHandler)
	r.Path("/1/queries/count/{date}").Methods("GET").HandlerFunc(handlers.DateDistinctHandler)
	return r
}
