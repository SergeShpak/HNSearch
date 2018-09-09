package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	fmt.Println("resp: ", msg)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(msg))
}

func getQueriesTimeInterval(r *http.Request) (*types.QueriesByDateRequest, error) {
	queriesTimeUnterval, ok := r.Context().Value(ctxIDs.QueryByDateRequestID).(*types.QueriesByDateRequest)
	if !ok {
		return nil, fmt.Errorf("could not retrived the queries time interval from the request's context")
	}
	return queriesTimeUnterval, nil
}

func getIndexer(r *http.Request) (engine.Indexer, error) {
	indexer, ok := r.Context().Value(ctxIDs.IndexerID).(engine.Indexer)
	if !ok {
		return nil, fmt.Errorf("could not retrived an indexer from the request's context")
	}
	return indexer, nil
}
