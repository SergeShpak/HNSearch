package simple

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SergeyShpak/HNSearch/indexer/server/types"
)

// Returns the number of distinct queries made during the specified time period
func (indexer *simpleIndexer) CountDistinctQueries(from *time.Time, to *time.Time) (int, error) {
	diff := newDatesDiff(from, to)
	paths, minSecPaths := diff.getIndexesPaths()
	// if the necessary data in the fast-fetch index, the search operation = the retrieve operation
	if len(minSecPaths) == 0 && len(paths) == 1 && len(paths[0].Indexes) == 1 {
		count, err := indexer.getSingleIndexDistinctQueriesCount(paths[0])
		if err != nil {
			return -1, err
		}
		return count, nil
	}
	// if the necessary data is not in the fast-fetch index, we need to combine the total indexes
	acc, err := indexer.gatherIndexesData(paths, minSecPaths)
	if err != nil {
		return -1, err
	}
	return len(acc.QueriesDict), nil
}

// Returns the queries that were made the most and the number of times they were made.
// The resulting queries are sorted in the descending order.
func (indexer *simpleIndexer) GetTopQueries(from *time.Time, to *time.Time, size int) (*types.TopQueriesResponse, error) {
	diff := newDatesDiff(from, to)
	paths, minSecPaths := diff.getIndexesPaths()
	// if the necessary data in the fast-fetch index, the search operation = the retrieve operation
	if len(minSecPaths) == 0 && len(paths) == 1 && len(paths[0].Indexes) == 1 {
		topQueries, err := indexer.getSingleIndexTopQueries(paths[0], size)
		if err != nil {
			return nil, err
		}
		resp := &types.TopQueriesResponse{
			Queries: topQueries,
		}
		return resp, nil
	}
	// if the necessary data is not in the fast-fetch index, we need to combine the total indexes
	acc, err := indexer.gatherIndexesData(paths, minSecPaths)
	if err != nil {
		return nil, err
	}
	if len(acc.QueriesCount) < size {
		size = len(acc.QueriesCount)
	}
	topQueries := acc.QueriesCount[:size]
	resp := &types.TopQueriesResponse{
		Queries: topQueries,
	}
	return resp, nil
}

type indexesPath struct {
	Path    []int
	Indexes []int
}

func newIndexesPath() *indexesPath {
	path := &indexesPath{
		Indexes: make([]int, 0),
	}
	return path
}

func (p *indexesPath) SetPath(parts ...int) {
	p.Path = make([]int, 0, len(parts))
	for _, part := range parts {
		p.Path = append(p.Path, part)
	}
}

func (p *indexesPath) AddIndex(i int) {
	p.Indexes = append(p.Indexes, i)
}

type minutesSecondsIndexesPaths indexesPath

type addDateUnit struct {
	Years  int
	Months int
	Days   int
}

type datesDiff struct {
	From *time.Time
	To   *time.Time
	Diff [6]int
}

func newDatesDiff(from *time.Time, to *time.Time) *datesDiff {
	diffDate := &datesDiff{
		From: from,
		To:   to,
	}
	diffDate.Diff[0] = to.Year() - from.Year()
	diffDate.Diff[1] = int(to.Month()) - int(from.Month())
	diffDate.Diff[2] = to.Day() - from.Day()
	diffDate.Diff[3] = to.Hour() - from.Hour()
	diffDate.Diff[4] = to.Minute() - from.Minute()
	diffDate.Diff[5] = to.Second() - from.Second()
	return diffDate
}

// This function is not pretty, and I could not really prettify it due to the complexity of the operations with *time.Time.
// It returns the indexes that are need to be analysed to fetch the data on the given time period (which is represented by the *datesDiff object).
//
// The indexes are created for every time unit, i.e. second, minute, hour, day, month and year. Let's say that we need to get the data on the period
// 12.07.2013 12:34:56 - 10.06.2015 14:20:11. The method will do the following:
//		1.	"Ceil" the from date to the largest time unit (from_unit), so that it differs from the corresponding time unit of the to date (to_unit) in this way:
//				to_unit - from_unit >= 2
//				In this case, the from date becomes: 12.07.2013 12:34:56 -> 01.01.2014 00:00:00. The result of this operation, are the paths of the indexes,
//				that contain the date of the time period 12.07.2013 12:34:56 - 01.01.2014 00:00:00.
//		2.	For each time unit, starting from the largest (the year), get to the corresponding time unit of the to date, by fetching the paths of the indexes.
//				In our example, the method will get the index of the year 2014 first, and move the from date to: 01.01.2014 00:00:00 -> 01.01.2015 00:00:00.
//				Then it will get the index of all the months, that separate the current from date and the to date,
//				so the date becomes: 01.01.2015 00:00:00 -> 01.06.2015 00:00:00. This process continues till the current from date becomes equal to the to date,
//				and till all the necessary index paths are collected.
func (diff *datesDiff) getIndexesPaths() ([]*indexesPath, []*minutesSecondsIndexesPaths) {
	paths := make([]*indexesPath, 0)
	minSecPaths := make([]*minutesSecondsIndexesPaths, 0)
	ceiled := *diff.From

	ceilTimeStepFn := func(ceil time.Time, path []int, ceiledUnit int, maxCeiledUnit int, toUnit int, interval time.Duration) *indexesPath {
		idxPath := newIndexesPath()
		idxPath.SetPath(path...)
		if diff.To.Before(ceil) || diff.To.Equal(ceil) {
			for i := ceiledUnit; i < toUnit; i++ {
				idxPath.AddIndex(i)
			}
			ceiled = ceiled.Add(interval * time.Duration((toUnit - ceiledUnit)))
			if len(idxPath.Indexes) == 0 {
				return nil
			}
			return idxPath
		}
		for i := ceiledUnit; i <= maxCeiledUnit; i++ {
			idxPath.AddIndex(i)
		}
		ceiled = ceiled.Add(interval * time.Duration((maxCeiledUnit - ceiledUnit + 1)))
		if len(idxPath.Indexes) == 0 {
			return nil
		}
		return idxPath
	}

	ceilDateStepFn := func(ceil time.Time, path []int, ceiledUnit int, maxCeiledUnit int, toUnit int, intervalUnit *addDateUnit) *indexesPath {
		idxPath := newIndexesPath()
		idxPath.SetPath(path...)
		if diff.To.Before(ceil) || diff.To.Equal(ceil) {
			for i := ceiledUnit; i < toUnit; i++ {
				idxPath.AddIndex(i)
				ceiled = ceiled.AddDate(intervalUnit.Years, intervalUnit.Months, intervalUnit.Days)
			}
			if len(idxPath.Indexes) == 0 {
				return nil
			}
			return idxPath
		}
		for i := ceiledUnit; i <= maxCeiledUnit; i++ {
			idxPath.AddIndex(i)
			ceiled = ceiled.AddDate(intervalUnit.Years, intervalUnit.Months, intervalUnit.Days)
		}
		if len(idxPath.Indexes) == 0 {
			return nil
		}
		return idxPath
	}

	// ceiling seconds
	if ceiled.Second() != 0 {
		ceil := time.Date(
			ceiled.Year(),
			ceiled.Month(),
			ceiled.Day(),
			ceiled.Hour(),
			ceiled.Minute()+1,
			0,
			0,
			diff.From.Location(),
		)
		indexes := ceilTimeStepFn(
			ceil,
			[]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day(), ceiled.Hour(), ceiled.Minute()},
			ceiled.Second(),
			59,
			diff.To.Second(),
			time.Second,
		)
		if indexes != nil {
			minSecIndexes := minutesSecondsIndexesPaths(*indexes)
			minSecPaths = append(minSecPaths, &minSecIndexes)
		}
	}
	// ceiling minutes
	if ceiled.Minute() != 0 {
		ceil := time.Date(
			ceiled.Year(),
			ceiled.Month(),
			ceiled.Day(),
			ceiled.Hour()+1,
			0,
			0,
			0,
			diff.From.Location(),
		)
		indexes := ceilTimeStepFn(
			ceil,
			[]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day(), ceiled.Hour()},
			ceiled.Minute(),
			59,
			diff.To.Minute(),
			time.Minute,
		)
		if indexes != nil {
			minSecIndexes := minutesSecondsIndexesPaths(*indexes)
			minSecPaths = append(minSecPaths, &minSecIndexes)
		}

	}
	// ceiling hours
	if ceiled.Hour() != 0 {
		ceil := time.Date(
			ceiled.Year(),
			ceiled.Month(),
			ceiled.Day()+1,
			0,
			0,
			0,
			0,
			ceiled.Location(),
		)
		indexes := ceilTimeStepFn(
			ceil,
			[]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day()},
			ceiled.Hour(),
			23,
			diff.To.Hour(),
			time.Hour,
		)
		if indexes != nil {
			paths = append(paths, indexes)
		}
	}
	if ceiled.Day() != 1 {
		ceil := time.Date(ceiled.Year(), ceiled.Month()+1, 0, 0, 0, 0, 0, ceiled.Location())
		interval := &addDateUnit{
			Days: 1,
		}
		indexes := ceilDateStepFn(ceil, []int{ceiled.Year(), int(ceiled.Month())}, ceiled.Day(), 31, diff.To.Day(), interval)
		if indexes != nil {
			paths = append(paths, indexes)
		}
	}
	if int(ceiled.Month()) != 1 {
		ceil := time.Date(ceiled.Year()+1, ceiled.Month(), 0, 0, 0, 0, 0, ceiled.Location())
		interval := &addDateUnit{
			Months: 1,
		}
		indexes := ceilDateStepFn(ceil, []int{ceiled.Year()}, int(ceiled.Month()), 11, int(diff.To.Month()), interval)
		if indexes != nil {
			paths = append(paths, indexes)
		}
	}
	getPathsFromCeiled := func(parts []int, ceiledUnit int, toUnit int) (*indexesPath, int) {
		p := newIndexesPath()
		p.SetPath(parts...)
		ctr := 0
		for i := ceiledUnit; i < toUnit; i++ {
			p.AddIndex(i)
			ctr++
		}
		if ctr == 0 {
			return nil, 0
		}
		return p, ctr
	}
	p, yearsAdded := getPathsFromCeiled(nil, ceiled.Year(), diff.To.Year())
	if p != nil {
		paths = append(paths, p)
		ceiled = ceiled.AddDate(yearsAdded, 0, 0)
	}
	p, monthsAdded := getPathsFromCeiled([]int{ceiled.Year()}, int(ceiled.Month()), int(diff.To.Month()))
	if p != nil {
		paths = append(paths, p)
		ceiled = ceiled.AddDate(0, monthsAdded, 0)
	}
	p, daysAdded := getPathsFromCeiled([]int{ceiled.Year(), int(ceiled.Month())}, ceiled.Day(), diff.To.Day())
	if p != nil {
		paths = append(paths, p)
		ceiled = ceiled.AddDate(0, 0, daysAdded)
	}
	p, hoursAdded := getPathsFromCeiled([]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day()}, ceiled.Hour(), diff.To.Hour())
	if p != nil {
		paths = append(paths, p)
		ceiled = ceiled.Add(time.Hour * time.Duration(hoursAdded))
	}
	p, minutesAdded := getPathsFromCeiled([]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day(), ceiled.Hour()}, ceiled.Minute(), diff.To.Minute())
	if p != nil {
		minP := minutesSecondsIndexesPaths(*p)
		minSecPaths = append(minSecPaths, &minP)
		ceiled = ceiled.Add(time.Minute * time.Duration(minutesAdded))
	}
	p, secondsAdded := getPathsFromCeiled([]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day(), ceiled.Hour(), ceiled.Minute()}, ceiled.Second(), diff.To.Second())
	if p != nil {
		secP := minutesSecondsIndexesPaths(*p)
		minSecPaths = append(minSecPaths, &secP)
		ceiled = ceiled.Add(time.Second * time.Duration(secondsAdded))
	}
	return paths, minSecPaths
}

func (indexer *simpleIndexer) getSingleIndexDistinctQueriesCount(p *indexesPath) (int, error) {
	parts := make([]int, 0, len(p.Path)+len(p.Indexes))
	for _, pathPart := range p.Path {
		parts = append(parts, pathPart)
	}
	if len(p.Indexes) != 1 {
		return -1, fmt.Errorf("expected to have a single indexs in the indexes path object, actually got %d", len(p.Indexes))
	}
	parts = append(parts, p.Indexes[0])
	count, err := indexer.loadDistinctQueries(parts...)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (indexer *simpleIndexer) getSingleIndexTopQueries(p *indexesPath, size int) ([]*types.Query, error) {
	parts := make([]int, 0, len(p.Path)+len(p.Indexes))
	for _, pathPart := range p.Path {
		parts = append(parts, pathPart)
	}
	if len(p.Indexes) != 1 {
		return nil, fmt.Errorf("expected to have a single indexs in the indexes path object, actually got %d", len(p.Indexes))
	}
	parts = append(parts, p.Indexes[0])
	topQueries, err := indexer.loadTopQueries(size, parts...)
	if err != nil {
		return nil, err
	}
	return topQueries, nil
}

func (indexer *simpleIndexer) loadDistinctQueries(parts ...int) (int, error) {
	fastFetchFilePath := indexer.getIndexFilePath(fastFetchDataName, parts)

	_, err := os.Stat(fastFetchFilePath)
	if err != nil {
		return -1, err
	}

	f, err := os.Open(fastFetchFilePath)
	if err != nil {
		return -1, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	distinctQueriesCount, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return -1, err
	}
	return distinctQueriesCount, nil
}

func (indexer *simpleIndexer) loadTopQueries(size int, parts ...int) ([]*types.Query, error) {
	fastFetchFilePath := indexer.getIndexFilePath(fastFetchDataName, parts)
	_, err := os.Stat(fastFetchFilePath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(fastFetchFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	topQueries := make([]*types.Query, 0, size)
	processedCtr := 0
	for scanner.Scan() {
		if processedCtr >= size {
			break
		}
		l := scanner.Text()
		lParts := strings.Split(l, "\t")
		if len(lParts) != 2 {
			return nil, fmt.Errorf("bad top query line format: %v", l)
		}
		q := &types.Query{
			Query: lParts[0],
		}
		q.Count, err = strconv.Atoi(lParts[1])
		if err != nil {
			return nil, fmt.Errorf("cannot cast query count \"%s\" to an int", lParts[1])
		}
		topQueries = append(topQueries, q)
		processedCtr++
	}
	return topQueries, nil
}

func (indexer *simpleIndexer) addIndexes(acc *index, paths []*indexesPath) (*index, error) {
	for _, p := range paths {
		for _, i := range p.Indexes {
			fileIndexParts := append(p.Path, i)
			index, err := indexer.loadTotalIndex(fileIndexParts...)
			if err != nil {
				return nil, err
			}
			acc.AddWithoutRecalculation(index)
		}
	}
	return acc, nil
}

func (indexer *simpleIndexer) addMinSecIndexes(acc *index, paths []*minutesSecondsIndexesPaths) (*index, error) {
	for _, p := range paths {
		fileParts := p.Path[0:3]
		for _, idx := range p.Indexes {
			hourIndex, err := indexer.loadHourIndex(p.Path[3], fileParts)
			if err != nil {
				return nil, err
			}
			// if indexes are minutes
			if len(p.Path) == 4 {
				m := hourIndex.Partition[idx]
				for i := 0; i < 60; i++ {
					acc.AddMapWithoutRecalculation(m[i])
				}
				continue
			}
			// if indexes are seconds
			if len(p.Path) == 5 {
				s := hourIndex.Partition[p.Path[4]][idx]
				acc.AddMapWithoutRecalculation(s)
			}
		}
	}
	return acc, nil
}

func (indexer *simpleIndexer) gatherIndexesData(paths []*indexesPath, minSecPaths []*minutesSecondsIndexesPaths) (*index, error) {
	acc := newIndex()
	var err error
	acc, err = indexer.addIndexes(acc, paths)
	if err != nil {
		return nil, err
	}
	acc, err = indexer.addMinSecIndexes(acc, minSecPaths)
	if err != nil {
		return nil, err
	}
	acc.CountQueries()
	return acc, nil
}

func (indexer *simpleIndexer) getIndexFilePath(fileName string, pathParts []int) string {
	parts := []string{indexer.getIndexFileDir(pathParts), fileName}
	filePath := strings.Join(parts, string(os.PathSeparator))
	return filePath
}
