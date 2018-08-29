package query_handler

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type simpleQueryEntry struct {
	Date  *time.Time
	Query string
}

type simpleQueryHandler struct {
	entries []*simpleQueryEntry
}

func newSimleQueryHanlder(queryDumpFile string) (*simpleQueryHandler, error) {
	file, err := os.Open(queryDumpFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	qd := &simpleQueryHandler{
		entries: make([]*simpleQueryEntry, 0),
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := qd.addQueryEntry(scanner.Text()); err != nil {
			return nil, err
		}
	}
	return qd, nil
}

func (qh *simpleQueryHandler) addQueryEntry(entry string) error {
	dParts := strings.Split(entry, "\t")
	if len(dParts) != 2 {
		return fmt.Errorf("entry contains only %d part (2 expected)", len(dParts))
	}
	date, err := qh.dateStrToTime(dParts[0])
	if err != nil {
		return err
	}
	e := &simpleQueryEntry{
		Date:  date,
		Query: dParts[1],
	}
	qh.entries = append(qh.entries, e)
	return nil
}

func (qh *simpleQueryHandler) dateStrToTime(dateStr string) (*time.Time, error) {
	dParts := strings.Split(dateStr, " ")
	if len(dParts) != 2 {
		return nil, fmt.Errorf("expected number of date parts: 2, actual number of date parts: %d", len(dParts))
	}
	dateRFCStr := dParts[0] + "T" + dParts[1] + "Z"
	date, err := time.Parse(time.RFC3339, dateRFCStr)
	if err != nil {
		return nil, err
	}
	return &date, nil
}

func (qh *simpleQueryHandler) CountQueries(from *time.Time, to *time.Time) int {
	count := len(qh.entries)
	return count
}

func (qh *simpleQueryHandler) GetTopQueries(from *time.Time, to *time.Time, size int) int {
	return 0
}
