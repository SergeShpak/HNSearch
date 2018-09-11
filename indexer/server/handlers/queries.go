package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/SergeyShpak/HNSearch/indexer/engine"
	ctxIDs "github.com/SergeyShpak/HNSearch/indexer/server/context"
	"github.com/SergeyShpak/HNSearch/indexer/server/types"
)

func CountQueries(w http.ResponseWriter, r *http.Request) {
	queriesTimeInterval, err := getQueriesTimeInterval(r)
	if err != nil {
		msg := fmt.Sprintf("could not get the queries time interval: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}
	indexer, err := getIndexer(r)
	if err != nil {
		msg := fmt.Sprintf("could not get an indexer: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	from := time.Unix(queriesTimeInterval.FromDate, 0).UTC()
	fmt.Println("From: ", from)
	to := time.Unix(queriesTimeInterval.ToDate, 0).UTC()
	fmt.Println("To: ", to)
	count, err := indexer.CountDistinctQueries(&from, &to)
	if err != nil {
		msg := fmt.Sprintf("an error occurred duiring distinct queries counting: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	countResponse := &types.DistinctQueriesCountResponse{
		Count: count,
	}
	msg, err := json.Marshal(countResponse)
	if err != nil {
		msg := "could not marshal DistinctQueriesCountResponse object"
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(msg))
}

func GetTopQueries(w http.ResponseWriter, r *http.Request) {
	queriesTimeInterval, err := getQueriesTimeInterval(r)
	if err != nil {
		msg := fmt.Sprintf("could not get the queries time interval: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}
	from := time.Unix(queriesTimeInterval.FromDate, 0).UTC()
	to := time.Unix(queriesTimeInterval.ToDate, 0).UTC()
	size, err := getSize(r)
	if err != nil {
		msg := fmt.Sprintf("could not get top queries size: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	indexer, err := getIndexer(r)
	if err != nil {
		msg := fmt.Sprintf("could not get an indexer: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	topQueries, err := indexer.GetTopQueries(&from, &to, size)
	if err != nil {
		msg := fmt.Sprintf("error while getting the top queries: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	msg, err := json.Marshal(topQueries)
	if err != nil {
		msg := "could not marshal TopQueriesResponse object"
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(msg))
}

func getQueriesTimeInterval(r *http.Request) (*types.DistinctQueriesCountRequest, error) {
	req, ok := r.Context().Value(ctxIDs.QueryByDateRequestID).(*types.DistinctQueriesCountRequest)
	if !ok {
		return nil, fmt.Errorf("could not retrived the queries time interval from the request's context")
	}
	return req, nil
}

func getIndexer(r *http.Request) (engine.Indexer, error) {
	indexer, ok := r.Context().Value(ctxIDs.IndexerID).(engine.Indexer)
	if !ok {
		return nil, fmt.Errorf("could not retrived an indexer from the request's context")
	}
	return indexer, nil
}

func getSize(r *http.Request) (int, error) {
	sizeHeader := r.Header.Get("x-top-size")
	if len(sizeHeader) == 0 {
		return -1, fmt.Errorf("size header was not found")
	}
	size, err := strconv.Atoi(sizeHeader)
	if err != nil {
		return -1, err
	}
	return size, nil
}
