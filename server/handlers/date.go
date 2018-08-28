package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	t "time"

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

	HasYear  bool
	HasMonth bool
	HasDay   bool

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
		year:    &yearStr,
		HasYear: true,
	}
	if len(dParts) >= 2 {
		monthStr := getTimeUnitWithHeadingZeroes(dParts[1])
		d.month = &monthStr
		d.HasMonth = true
		dStrParts[1] = monthStr
	}
	if len(dParts) >= 3 {
		dayStr := getTimeUnitWithHeadingZeroes(dParts[2])
		d.day = &dayStr
		d.HasDay = true
		dStrParts[2] = dayStr
	}
	d.repr = strings.Join(dStrParts, "-")
	return d, nil
}

func (d *date) String() string {
	return d.repr
}

type time struct {
	hour   *string
	minute *string
	second *string

	HasHour   bool
	HasMinute bool
	HasSecond bool

	repr string
}

func newTime(timeStr *string) (*time, error) {
	t := &time{}
	if timeStr == nil || *timeStr == "" {
		t.repr = "00:00:00"
		return t, nil
	}
	tParts := utils.ReExtractNumbers(*timeStr)
	timePartsCount := 3
	if len(tParts) == 0 || len(tParts) > timePartsCount {
		return nil, fmt.Errorf("expected from 1 to %d time parts, actual number: %d", timePartsCount, len(tParts))
	}
	hourStr := getTimeUnitWithHeadingZeroes(tParts[0])
	tStrParts := []string{hourStr, "00", "00"}
	t.hour = &hourStr
	t.HasHour = true
	if len(tParts) >= 2 {
		minuteStr := getTimeUnitWithHeadingZeroes(tParts[1])
		tStrParts[1] = minuteStr
		t.minute = &tParts[1]
		t.HasMinute = true
	}
	if len(tParts) >= 3 {
		secondStr := getTimeUnitWithHeadingZeroes(tParts[2])
		tStrParts[2] = secondStr
		t.second = &tParts[2]
		t.HasSecond = true
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

func (dt dateTime) String() string {
	repr := dt.date.String() + "T" + dt.time.String() + "Z"
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
	if !fromDT.date.HasMonth {
		return from.AddDate(1, 0, 0)
	}
	if !fromDT.date.HasDay {
		return from.AddDate(0, 1, 0)
	}
	if !fromDT.time.HasHour {
		return from.AddDate(0, 0, 1)
	}
	if !fromDT.time.HasMinute {
		return from.Add(t.Hour)
	}
	if !fromDT.time.HasSecond {
		return from.Add(t.Minute)
	}
	return from.Add(t.Second)
}

func (tp *timePeriod) String() string {
	repr := fmt.Sprintf("From: %v\nTo: %v", tp.from, tp.to)
	return repr
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
	tp, err := newTimePeriod(dt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	msg := fmt.Sprintf("TP: %v", tp)
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
