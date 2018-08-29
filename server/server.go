package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/SergeyShpak/HNSearch/config"
	ctx "github.com/SergeyShpak/HNSearch/server/context"
	"github.com/SergeyShpak/HNSearch/server/handlers"
	qh "github.com/SergeyShpak/HNSearch/server/model/query_handler"
)

var queryHandler qh.QueryHandler

func InitServer(c *config.Config) (*http.Server, error) {
	if c == nil {
		c = config.GetDefaultConfig()
	}
	var err error
	queryHandler, err = qh.GetQueryHandler(c.QueryHandler)
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
	r.Path("/1/queries/count/{date}").Methods("GET").Queries("size", "{size}").HandlerFunc(handlers.DatePopularHandler)
	r.Path("/1/queries/count/{date}").Methods("GET").HandlerFunc(handlers.DateDistinctHandler)
	r.Use(getDumpHandler)
	return r
}

func getDumpHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctx.QueryHandlerID, queryHandler)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
