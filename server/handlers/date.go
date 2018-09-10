package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	ctxIDs "github.com/SergeyShpak/HNSearch/server/context"
	"github.com/SergeyShpak/HNSearch/server/indexer"
	"github.com/SergeyShpak/HNSearch/server/types"
)

func CountDistinctQueries(w http.ResponseWriter, r *http.Request) {
	from, to, err := getTimePeriod(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("An error occurred during time period parsing: %v", err)))
		return
	}
	indexer, err := getIndexer(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
	queiresCount, err := indexer.CountDistinctQueries(from, to)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
	resp := &types.DistinctQueriesCountResponse{
		Count: queiresCount,
	}
	sendJSONResponse(w, resp)
}

func GetTopQueries(w http.ResponseWriter, r *http.Request) {
	from, to, err := getTimePeriod(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("An error occurred during time period parsing: %v", err)))
		return
	}
	size, err := getSize(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("An error occurred during size parsing: %v", err)))
		return
	}
	indexer, err := getIndexer(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%v", err)))
	}
	resp, err := indexer.GetTopQueries(from, to, size)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%v", err)))
	}
	sendJSONResponse(w, resp)
}

func getTimePeriod(r *http.Request) (*time.Time, *time.Time, error) {
	from, ok := r.Context().Value(ctxIDs.FromDateID).(*time.Time)
	if !ok {
		return nil, nil, fmt.Errorf("cannot get Date from the request context")
	}
	to, ok := r.Context().Value(ctxIDs.ToDateID).(*time.Time)
	if !ok {
		return nil, nil, fmt.Errorf("cannot get Time from the request context")
	}
	return from, to, nil
}

func getSize(r *http.Request) (int, error) {
	size, ok := r.Context().Value(ctxIDs.SizeParamID).(int)
	if !ok {
		return 0, fmt.Errorf("cannot get Size from the request context")
	}
	return size, nil
}

func getIndexer(r *http.Request) (indexer.Indexer, error) {
	indexer, ok := r.Context().Value(ctxIDs.IndexerID).(indexer.Indexer)
	if !ok {
		return nil, fmt.Errorf("could not retrieve indexer from the context")
	}
	return indexer, nil
}

func sendJSONResponse(w http.ResponseWriter, resp interface{}) {
	respB, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("could not marshal the response: %v", err)))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respB)
}
