package date

import (
	"fmt"
	"strings"
	t "time"

	"github.com/SergeyShpak/HNSearch/server/types"
)

type Date struct {
	Year  *string
	Month *string
	Day   *string
}

func NewDate(dateParts []string) (*Date, error) {
	d := &Date{}
	if len(dateParts) == 0 {
		return d, nil
	}
	datePartsCount := 3
	if len(dateParts) == 0 || len(dateParts) > datePartsCount {
		return nil, fmt.Errorf("expected from 1 to %d date parts, actual number: %d", datePartsCount, len(dateParts))
	}
	d.Year = &dateParts[0]
	if len(dateParts) >= 2 {
		monthStr := getTimeUnitWithHeadingZeroes(dateParts[1])
		d.Month = &monthStr
	}
	if len(dateParts) >= 3 {
		dayStr := getTimeUnitWithHeadingZeroes(dateParts[2])
		d.Day = &dayStr
	}
	return d, nil
}

func (d *Date) String() string {
	dateParts := []string{"1900", "01", "01"}
	if d.Year != nil {
		dateParts[0] = *d.Year
	}
	if d.Month != nil {
		dateParts[1] = *d.Month
	}
	if d.Day != nil {
		dateParts[2] = *d.Day
	}
	str := strings.Join(dateParts, "-")
	return str
}

type Time struct {
	Hour   *string
	Minute *string
	Second *string
}

func NewTime(timeParts []string) (*Time, error) {
	t := &Time{}
	if len(timeParts) == 0 {
		return t, nil
	}
	timePartsCount := 3
	if len(timeParts) == 0 || len(timeParts) > timePartsCount {
		return nil, fmt.Errorf("expected from 1 to %d time parts, actual number: %d", timePartsCount, len(timeParts))
	}
	hourStr := getTimeUnitWithHeadingZeroes(timeParts[0])
	t.Hour = &hourStr
	if len(timeParts) >= 2 {
		minuteStr := getTimeUnitWithHeadingZeroes(timeParts[1])
		t.Minute = &minuteStr
	}
	if len(timeParts) >= 3 {
		secondStr := getTimeUnitWithHeadingZeroes(timeParts[2])
		t.Second = &secondStr
	}
	return t, nil
}

func (t *Time) String() string {
	strParts := []string{"00", "00", "00"}
	if t.Hour != nil {
		strParts[0] = *t.Hour
	}
	if t.Minute != nil {
		strParts[1] = *t.Minute
	}
	if t.Second != nil {
		strParts[2] = *t.Second
	}
	str := strings.Join(strParts, ":")
	return str
}

func getTimeUnitWithHeadingZeroes(timeUnit string) string {
	if len(timeUnit) == 2 {
		return timeUnit
	}
	return "0" + timeUnit
}

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
