package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	t "time"

	ctxIDs "github.com/SergeyShpak/HNSearch/server/context"
	"github.com/SergeyShpak/HNSearch/server/indexer"
	"github.com/SergeyShpak/HNSearch/server/types"
)

type dateTime struct {
	Date *types.Date
	Time *types.Time
}

func newDateTime(date *types.Date, time *types.Time) *dateTime {
	dt := &dateTime{
		Date: date,
		Time: time,
	}
	return dt
}

func (dt dateTime) String() string {
	repr := dt.Date.String() + "T" + dt.Time.String() + "Z"
	return repr
}

type timePeriod struct {
	from *t.Time
	to   *t.Time
}

func newTimePeriod(fromDT *dateTime) (*timePeriod, error) {
	if fromDT == nil {
		return nil, fmt.Errorf("nil dateTime passed")
	}
	from, err := t.Parse(t.RFC3339, fromDT.String())
	if err != nil {
		return nil, err
	}
	to := timePeriodGetTo(fromDT, from)
	timePeriod := &timePeriod{
		from: &from,
		to:   &to,
	}
	return timePeriod, nil
}

func timePeriodGetTo(fromDT *dateTime, from t.Time) t.Time {
	if fromDT.Date.Month == nil {
		return from.AddDate(1, 0, 0)
	}
	if fromDT.Date.Day == nil {
		return from.AddDate(0, 1, 0)
	}
	if fromDT.Time.Hour == nil {
		return from.AddDate(0, 0, 1)
	}
	if fromDT.Time.Minute == nil {
		return from.Add(t.Hour)
	}
	if fromDT.Time.Second == nil {
		return from.Add(t.Minute)
	}
	return from.Add(t.Second)
}

func (tp *timePeriod) String() string {
	repr := fmt.Sprintf("From: %v\nTo: %v", tp.from, tp.to)
	return repr
}

func DateDistinctHandler(w http.ResponseWriter, r *http.Request) {
	tp, err := getTimePeriod(r)
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
	queiresCount, err := indexer.CountDistinctQueries(tp.from, tp.to)
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

/*
func DatePopularHandler(w http.ResponseWriter, r *http.Request) {
	tp, err := getTimePeriod(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("An error occurred during time period parsing: %v", err)))
		return
	}
	size, err := getSize(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	qdVal := r.Context().Value(ctxIDs.QueryHandlerID)
	qd, ok := qdVal.(query_handler.QueryHandler)
	if !ok {
		msg := "could not cast dump in the request context to QueryDump"
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		returnhttps://stackoverflow.com/questions/33238518/what-could-happen-if-i-dont-close-response-body-in-golang
	}
	queriesCount := qd.GetTopQueries(tp.from, tp.to, size)
	msg, err := json.Marshal(queriesCount)
	if err != nil {
		msg := "could not marshal QueriesCount object"
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(msg))
}
*/

func getTimePeriod(r *http.Request) (*timePeriod, error) {
	date, ok := r.Context().Value(ctxIDs.DateParamID).(*types.Date)
	if !ok {
		return nil, fmt.Errorf("cannot get Date from the request context")
	}
	time, ok := r.Context().Value(ctxIDs.TimeParamID).(*types.Time)
	if !ok {
		return nil, fmt.Errorf("cannot get Time from the request context")
	}
	dt := newDateTime(date, time)
	tp, err := newTimePeriod(dt)
	if err != nil {
		return nil, err
	}
	return tp, nil
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
	w.Write(respB)
}
