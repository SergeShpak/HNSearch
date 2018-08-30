package query_handler

import (
	"bufio"
	"fmt"
	"os"
	"sort"
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
	sort.Slice(qd.entries, func(i int, j int) bool {
		return qd.entries[i].Date.Before(*qd.entries[j].Date)
	})
	fmt.Println("QD created")
	return qd, nil
}

func (qh *simpleQueryHandler) CountDistinctQueries(from *time.Time, to *time.Time) *DistinctQueriesCount {
	interval := qh.getQueiresInterval(from, to)
	distinctQueries := qh.getDistinctQueries(interval)
	distinctQueriesCount := &DistinctQueriesCount{
		Count: len(distinctQueries),
	}
	return distinctQueriesCount
}

func (qh *simpleQueryHandler) GetTopQueries(from *time.Time, to *time.Time, size int) []*QueryCount {
	interval := qh.getQueiresInterval(from, to)
	queriesCount := qh.getTopQueries(interval, size)
	return queriesCount
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

func (qh *simpleQueryHandler) checkFromToOrder(from *time.Time, to *time.Time) (*time.Time, *time.Time) {
	if from.After(*to) {
		return to, from
	}
	return from, to
}

func (qh *simpleQueryHandler) getQueiresInterval(from *time.Time, to *time.Time) []*simpleQueryEntry {
	from, to = qh.checkFromToOrder(from, to)
	low := 0
	high := len(qh.entries) - 1
	// from and to before the lowest date
	if from.Before(*qh.entries[low].Date) && to.Before(*qh.entries[low].Date) {
		return nil
	}
	// from and to after the lowest date
	if from.After(*qh.entries[high].Date) && to.After(*qh.entries[high].Date) {
		return nil
	}
	// the interval includes the dump
	if qh.timeBeforeOrEqual(from, qh.entries[low].Date) && qh.timeAfterOrEqual(to, qh.entries[high].Date) {
		return qh.entries
	}
	result, ok := qh.cutDumpFromLeft(qh.entries, from)
	if !ok {
		return nil
	}
	result, ok = qh.cutDumpFromRight(result, to)
	if !ok {
		return nil
	}
	return result
}

func (qh *simpleQueryHandler) timeBeforeOrEqual(t *time.Time, u *time.Time) bool {
	return t.Before(*u) || t.Equal(*u)
}

func (qh *simpleQueryHandler) timeAfterOrEqual(t *time.Time, u *time.Time) bool {
	return t.After(*u) || t.Equal(*u)
}

func (qh *simpleQueryHandler) cutDumpFromLeft(dump []*simpleQueryEntry, from *time.Time) ([]*simpleQueryEntry, bool) {
	low := sort.Search(len(dump), func(i int) bool {
		return qh.timeAfterOrEqual(dump[i].Date, from)
	})
	if low == len(dump) {
		return nil, false
	}
	return dump[low:], true
}

func (qh *simpleQueryHandler) cutDumpFromRight(dump []*simpleQueryEntry, to *time.Time) ([]*simpleQueryEntry, bool) {
	high := sort.Search(len(dump), func(i int) bool {
		return qh.timeAfterOrEqual(dump[i].Date, to)
	})
	if high == len(dump) || high == 0 {
		return nil, false
	}
	return dump[:high], true
}

func (qh *simpleQueryHandler) getDistinctQueries(interval []*simpleQueryEntry) map[string]byte {
	dictSeen := make(map[string]byte)
	for _, el := range interval {
		dictSeen[el.Query] = '1'
	}
	return dictSeen
}

func (qh *simpleQueryHandler) countQueries(interval []*simpleQueryEntry) map[string]int {
	dictSeen := make(map[string]int)
	for _, el := range interval {
		if _, ok := dictSeen[el.Query]; !ok {
			dictSeen[el.Query] = 0
		}
		dictSeen[el.Query]++
	}
	return dictSeen
}

func (qh *simpleQueryHandler) getTopQueries(interval []*simpleQueryEntry, size int) []*QueryCount {
	queriesCount := qh.countQueries(interval)
	queries := make([]*QueryCount, len(queriesCount))
	i := 0
	for q, count := range queriesCount {
		qc := &QueryCount{
			Query: q,
			Count: count,
		}
		queries[i] = qc
		i++
	}
	sort.Slice(queries, func(i int, j int) bool {
		return queries[i].Count > queries[j].Count
	})
	if size > len(queries) {
		return queries
	}
	return queries[:size]
}
