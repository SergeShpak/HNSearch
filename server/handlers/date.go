package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/SergeyShpak/HNSearch/server/utils"
)

type date struct {
	Year  *string
	Month *string
	Day   *string
}

func newDate(dateStr *string) (*date, error) {
	if dateStr == nil {
		return nil, nil
	}
	dParts := utils.ReExtractNumbers(*dateStr)
	datePartsCount := 3
	if len(dParts) == 0 || len(dParts) > datePartsCount {
		return nil, fmt.Errorf("expected from 1 to %d date parts, actual number: %d", datePartsCount, len(dParts))
	}
	d := &date{}
	if len(dParts) >= 1 {
		d.Year = &dParts[0]
	}
	if len(dParts) >= 2 {
		d.Month = &dParts[1]
	}
	if len(dParts) >= 3 {
		d.Day = &dParts[2]
	}
	return d, nil
}

type time struct {
	Hour   *string
	Minute *string
	Second *string
}

func newTime(timeStr *string) (*time, error) {
	if timeStr == nil {
		return nil, nil
	}
	tParts := utils.ReExtractNumbers(*timeStr)
	timePartsCount := 3
	if len(tParts) == 0 || len(tParts) > timePartsCount {
		return nil, fmt.Errorf("expected from 1 to %d date parts, actual number: %d", timePartsCount, len(tParts))
	}
	t := &time{}
	if len(tParts) >= 1 {
		t.Hour = &tParts[0]
	}
	if len(tParts) >= 2 {
		t.Minute = &tParts[1]
	}
	if len(tParts) >= 3 {
		t.Second = &tParts[2]
	}
	return t, nil
}

type dateTime struct {
	date *date
	time *time
}

func newDateTime(dateTimeStr string) (*dateTime, error) {
	date, time, err := utils.ReSplitDateTime(dateTimeStr)
	if err != nil {
		return nil, err
	}
	d, err := newDate(date)
	if err != nil {
		return nil, err
	}
	t, err := newTime(time)
	if err != nil {
		return nil, err
	}
	dt := &dateTime{
		date: d,
		time: t,
	}
	return dt, nil
}

func DateDistinctHandler(w http.ResponseWriter, r *http.Request) {
	date, ok := mux.Vars(r)["date"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("NoDate!"))
		return
	}
	dt, err := newDateTime(date)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	msg := fmt.Sprintf("Date: %v", dt.date)
	w.Write([]byte(msg))
}

func DatePopularHandler(w http.ResponseWriter, r *http.Request) {
	date, ok := mux.Vars(r)["date"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("NoDate!"))
		return
	}
	dateMsg := fmt.Sprintf("Date: %s", date)
	size, err := getSizeParam(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	sizeMsg := fmt.Sprintf("Size: %d", size)
	msg := dateMsg + "\n" + sizeMsg
	w.Write([]byte(msg))
	return
}

func getSizeParam(r *http.Request) (int, error) {
	var size int
	sizeParams, err := getParameter("size", r)
	if err != nil {
		return size, err
	}
	if len(sizeParams) > 1 {
		return size, fmt.Errorf("multiple size parameters passed")
	}
	size, err = parseSize(sizeParams[0])
	if err != nil {
		return size, err
	}
	return size, nil
}

func parseSize(size string) (int, error) {
	s, err := strconv.Atoi(size)
	if err != nil {
		return s, err
	}
	if s <= 0 {
		return s, fmt.Errorf("size %d is invalid", s)
	}
	return s, nil
}
