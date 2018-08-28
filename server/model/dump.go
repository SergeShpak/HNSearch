package model

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type dumpEntry struct {
	time  *time.Time
	query string
}

type QueryDump struct {
	data []*dumpEntry
}

func NewQueryDump(fileName string) (*QueryDump, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	qd := &QueryDump{
		data: make([]*dumpEntry, 0),
	}
	scanner := bufio.NewScanner(file)
	ctr := 0
	for scanner.Scan() {
		if err := qd.addDumpEntry(scanner.Text()); err != nil {
			return nil, err
		}
		if ctr > 3 {
			break
		}
		ctr++
	}
	return qd, nil
}

func (qd *QueryDump) addDumpEntry(entry string) error {
	dParts := strings.Split(entry, "\t")
	if len(dParts) != 2 {
		return fmt.Errorf("entry contains only %d part (2 expected)", len(dParts))
	}
	date, err := qd.dateStrToTime(dParts[0])
	if err != nil {
		return err
	}
	e := &dumpEntry{
		time:  date,
		query: dParts[1],
	}
	qd.data = append(qd.data, e)
	return nil
}

func (qd *QueryDump) dateStrToTime(dateStr string) (*time.Time, error) {
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
