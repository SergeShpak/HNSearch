package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/SergeyShpak/HNSearch/server/utils"
)

func getTimeUnitWithHeadingZeroes(timeUnit string) string {
	if len(timeUnit) == 2 {
		return timeUnit
	}
	return "0" + timeUnit
}

type date struct {
	year  *string
	month *string
	day   *string

	repr string
}

func newDate(dateStr *string) (*date, error) {
	if dateStr == nil {
		return nil, fmt.Errorf("nil dateString passed")
	}
	dParts := utils.ReExtractNumbers(*dateStr)
	datePartsCount := 3
	if len(dParts) == 0 || len(dParts) > datePartsCount {
		return nil, fmt.Errorf("expected from 1 to %d date parts, actual number: %d", datePartsCount, len(dParts))
	}
	yearStr := dParts[0]
	dStrParts := []string{yearStr, "01", "01"}
	d := &date{
		year: &yearStr,
	}
	if len(dParts) >= 2 {
		monthStr := getTimeUnitWithHeadingZeroes(dParts[1])
		d.month = &monthStr
		dStrParts[1] = monthStr
	}
	if len(dParts) >= 3 {
		dayStr := getTimeUnitWithHeadingZeroes(dParts[2])
		d.day = &dayStr
		dStrParts[2] = dayStr
	}
	d.repr = strings.Join(dStrParts, "-")
	return d, nil
}

func (d *date) String() string {
	return d.repr
}

type time struct {
	Hour   *string
	Minute *string
	Second *string

	repr string
}

func newTime(timeStr *string) (*time, error) {
	t := &time{}
	if timeStr == nil {
		t.repr = "00:00:00"
		return t, nil
	}
	tParts := utils.ReExtractNumbers(*timeStr)
	timePartsCount := 3
	if len(tParts) == 0 || len(tParts) > timePartsCount {
		return nil, fmt.Errorf("expected from 1 to %d date parts, actual number: %d", timePartsCount, len(tParts))
	}
	hourStr := getTimeUnitWithHeadingZeroes(tParts[0])
	tStrParts := []string{hourStr, "00", "00"}
	t.Hour = &hourStr
	if len(tParts) >= 2 {
		minuteStr := getTimeUnitWithHeadingZeroes(tParts[1])
		tStrParts[1] = minuteStr
		t.Minute = &tParts[1]
	}
	if len(tParts) >= 3 {
		secondStr := getTimeUnitWithHeadingZeroes(tParts[2])
		tStrParts[2] = secondStr
		t.Second = &tParts[2]
	}
	t.repr = strings.Join(tStrParts, ":")
	return t, nil
}

func (t time) String() string {
	return t.repr
}

type dateTime struct {
	date *date
	time *time
}

func (dt dateTime) String() string {
	repr := dt.date.String() + "T" + dt.time.String() + "Z"
	return repr
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
	msg := fmt.Sprintf("Date: %v", dt)
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
