package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/SergeyShpak/HNSearch/server/config"
	ctxIDs "github.com/SergeyShpak/HNSearch/server/context"
	"github.com/SergeyShpak/HNSearch/server/handlers"
	"github.com/SergeyShpak/HNSearch/server/indexer"
	middleware_date "github.com/SergeyShpak/HNSearch/server/middleware/date"
	"github.com/SergeyShpak/HNSearch/server/utils/reqparser"
)

var indexerObj indexer.Indexer
var requestParser reqparser.Parser

func initServer(c *config.Config) (*http.Server, error) {
	if c == nil {
		c = config.GetDefaultConfig()
	}
	var err error
	indexerObj, err = indexer.NewIndexer(c)
	if err != nil {
		return nil, err
	}
	requestParser, err = reqparser.GetRequestsParser()
	if err != nil {
		return nil, err
	}
	r := newRouter()
	addr := ":8080"
	if c.Server != nil && c.Server.Port != 0 {
		addr = ":" + strconv.Itoa(c.Server.Port)
	}
	s := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s, nil
}

func newRouter() *mux.Router {
	r := mux.NewRouter()
	//r.Path("/1/queries/count/{date}").Methods("GET").Queries("size", "{size}").HandlerFunc(handlers.DatePopularHandler)
	r.Path("/1/queries/count/{date}").Methods("GET").HandlerFunc(handlers.DateDistinctHandler)
	r.Use(setRequestUtils)
	r.Use(middleware_date.ParseDateRequest)
	return r
}

func setRequestUtils(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxIDs.IndexerID, indexerObj)
		ctx = context.WithValue(ctx, ctxIDs.RequestParserID, requestParser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
