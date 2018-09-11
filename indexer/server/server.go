package server

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/SergeyShpak/HNSearch/indexer/config"
	"github.com/SergeyShpak/HNSearch/indexer/engine"
	ctxIDs "github.com/SergeyShpak/HNSearch/indexer/server/context"
	"github.com/SergeyShpak/HNSearch/indexer/server/handlers"
	middleware_queries "github.com/SergeyShpak/HNSearch/indexer/server/middleware/queries"
)

var indexer engine.Indexer

func InitServer(c *config.Config) (*http.Server, error) {
	if c == nil {
		c = config.GetDefaultConfig()
	}
	r := newRouter()
	addr := ":8081"
	if c.Server != nil && c.Server.Port != 0 {
		addr = ":" + strconv.Itoa(c.Server.Port)
	}
	log.Println("Addr:", addr)
	s := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if err := initIndexer(c); err != nil {
		return nil, err
	}
	return s, nil
}

func initIndexer(c *config.Config) error {
	var err error
	indexer, err = engine.NewIndexer(c)
	if err != nil {
		return err
	}
	if err := indexer.IndexData(); err != nil {
		return err
	}
	log.Println("Data indexed")
	return nil
}

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.Path("/1/queries/count").Methods("POST").HandlerFunc(handlers.CountQueries)
	r.Path("/1/queries/top").Methods("POST").HandlerFunc(handlers.GetTopQueries)
	r.Use(injectIndexer)
	r.Use(middleware_queries.ParseQueriesByDateRequest)
	return r
}

func injectIndexer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxIDs.IndexerID, indexer)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
