package engine

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SergeyShpak/HNSearch/indexer/config"
	"github.com/SergeyShpak/HNSearch/indexer/parser"
	"github.com/SergeyShpak/HNSearch/indexer/sorter"
)

type hourData struct {
	Hour      *time.Time
	Partition [60][60]map[string]int
	Index     *index
}

type index struct {
	QueriesDict  map[string]int
	QueriesCount []*entry
}

func newIndex() *index {
	idx := &index{
		QueriesDict:  make(map[string]int),
		QueriesCount: make([]*entry, 0),
	}
	return idx
}

func (idx *index) CountQueries() {
	entries := make([]*entry, 0)
	for query, count := range idx.QueriesDict {
		e := &entry{
			Query: query,
			Count: count,
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i int, j int) bool {
		return entries[i].Count < entries[j].Count
	})
}

func (idx *index) Sub(other *index) {
	if len(other.QueriesDict) == 0 {
		return
	}
	for query, count := range other.QueriesDict {
		if _, ok := idx.QueriesDict[query]; !ok {
			idx.QueriesDict[query] = 0
		}
		idx.QueriesDict[query] -= count
	}
	idx.CountQueries()
}

func (idx *index) Add(other *index) {
	needRecalculation := idx.AddWithourRecalculation(other)
	if needRecalculation {
		idx.CountQueries()
	}
}

func (idx *index) AddWithourRecalculation(other *index) bool {
	if len(other.QueriesDict) == 0 {
		return false
	}
	for query, count := range other.QueriesDict {
		if _, ok := idx.QueriesDict[query]; !ok {
			idx.QueriesDict[query] = 0
		}
		idx.QueriesDict[query] += count
	}
	return true
}

func (idx *index) WriteTo(w io.Writer) error {
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(idx); err != nil {
		return err
	}
	return nil
}

type entry struct {
	Query string
	Count int
}

type hourIndexID struct {
	Day   int
	Month int
	Year  int
}

type toUpdate struct {
	paths map[int]map[int]map[int]byte
}

func newToUpdate() *toUpdate {
	toUpdate := &toUpdate{
		paths: make(map[int]map[int]map[int]byte),
	}
	return toUpdate
}

func (tu *toUpdate) add(year int, month int, day int) error {
	if tu == nil {
		return nil
	}
	if _, ok := tu.paths[year]; !ok {
		tu.paths[year] = make(map[int]map[int]byte)
	}
	if _, ok := tu.paths[year][month]; !ok {
		tu.paths[year][month] = make(map[int]byte)
	}
	tu.paths[year][month][day] = '1'
	return nil
}

type simpleIndexer struct {
	fileMux *sync.Mutex
	parser  parser.Parser
	sorter  sorter.Sorter
}

func newSimpleIndexer(c *config.SimpleIndexer) (*simpleIndexer, error) {
	if c == nil {
		return nil, fmt.Errorf("nil passed as simple indexer configuration")
	}
	indexer := &simpleIndexer{
		fileMux: &sync.Mutex{},
	}
	var err error
	indexer.parser, err = parser.NewParser(c.Parser)
	if err != nil {
		return nil, err
	}
	indexer.sorter, err = sorter.NewSorter(c.Sorter)
	if err != nil {
		return nil, err
	}
	return indexer, nil
}

func (indexer *simpleIndexer) UpdateIndices(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	var data *hourData
	var isFinished bool
	var err error
	indexesUpdated := make([]*hourIndexID, 0)
	for {
		data, isFinished, err = indexer.ScanHour(scanner)
		if err != nil {
			return err
		}
		indexer.finalizeHour(data)
		if isFinished {
			break
		}
		hourIdxID, err := indexer.updateHourIndex(data)
		if err != nil {
			return err
		}
		indexesUpdated = append(indexesUpdated, hourIdxID)
	}
	indexPath, err := indexer.updateHourIndex(data)
	if err != nil {
		return err
	}
	indexesUpdated = append(indexesUpdated, indexPath)
	indexesToUpdate, err := indexer.getParentIndexesToUpdate(indexesUpdated)
	if err != nil {
		return err
	}
	fmt.Println(indexesToUpdate)
	if err := indexer.updateParentsIndexes(indexesToUpdate); err != nil {
		return err
	}
	return nil
}

func (indexer *simpleIndexer) ScanHour(scanner *bufio.Scanner) (*hourData, bool, error) {
	if !scanner.Scan() {
		return nil, true, nil
	}
	line := scanner.Text()
	q, err := indexer.parser.Parse(line)
	if err != nil {
		return nil, false, err
	}
	thisHour := time.Date(q.Date.Year(), q.Date.Month(), q.Date.Day(), q.Date.Hour(), 0, 0, 0, q.Date.Location())
	nextHour := time.Date(q.Date.Year(), q.Date.Month(), q.Date.Day(), q.Date.Hour()+1, 0, 0, 0, q.Date.Location())
	currMinute := q.Date.Minute()
	currSecond := q.Date.Second()
	data := &hourData{
		Hour: &thisHour,
	}
	data.Partition[currMinute][currSecond] = make(map[string]int)
	for scanner.Scan() {
		line := scanner.Text()
		q, err := indexer.parser.Parse(line)
		if err != nil {
			return nil, false, err
		}
		if q.Date.Before(nextHour) {
			if q.Date.Minute() != currMinute || q.Date.Second() != currSecond {
				currMinute = q.Date.Minute()
				currSecond = q.Date.Second()
				data.Partition[currMinute][currSecond] = make(map[string]int)
			}
			if _, ok := data.Partition[currMinute][currSecond][q.Query]; !ok {
				data.Partition[currMinute][currSecond][q.Query] = 0
			}
			data.Partition[currMinute][currSecond][q.Query]++
			continue
		}
		return data, false, nil
	}
	return data, true, nil
}

func (indexer *simpleIndexer) finalizeHour(data *hourData) {
	data.Index = newIndex()
	for i := 0; i < 60; i++ {
		for j := 0; j < 60; j++ {
			currDict := data.Partition[i][j]
			if currDict == nil {
				continue
			}
			mergeDicts(data.Index.QueriesDict, currDict)
		}
	}
	data.Index.CountQueries()
}

func mergeDicts(acc map[string]int, src map[string]int) {
	for k, v := range src {
		if _, ok := acc[k]; !ok {
			acc[k] = 0
		}
		acc[k] += v
	}
}

func (indexer *simpleIndexer) updateHourIndex(data *hourData) (*hourIndexID, error) {
	id, err := indexer.getHourIndex(data.Hour)
	if err != nil {
		return nil, err
	}
	hourIndexPath, err := indexer.createHourIndexPath(id)
	if err != nil {
		return nil, err
	}
	indexer.fileMux.Lock()
	if err := createPath(hourIndexPath); err != nil {
		return nil, err
	}
	indexer.fileMux.Unlock()
	if err != nil {
		return nil, err
	}
	if _, err = indexer.writeIndexToFile(hourIndexPath, data); err != nil {
		return nil, err
	}
	return id, nil
}

func (indexer *simpleIndexer) getHourIndex(hour *time.Time) (*hourIndexID, error) {
	if hour == nil {
		return nil, fmt.Errorf("passed hour is nil")
	}
	id := &hourIndexID{
		Day:   hour.Day(),
		Month: int(hour.Month()),
		Year:  int(hour.Year()),
	}
	return id, nil
}

//TODO: move ./indexes to global, rewrite with getIndexFileDir
func (indexer *simpleIndexer) createHourIndexPath(id *hourIndexID) (string, error) {
	if id == nil {
		return "", fmt.Errorf("passed hour index is nil")
	}
	pathParts := []string{"./indexes", strconv.Itoa(id.Year), strconv.Itoa(id.Month), strconv.Itoa(id.Day)}
	path := strings.Join(pathParts, string(os.PathSeparator))
	return path, nil
}

func (indexer *simpleIndexer) getParentIndexesToUpdate(hourIndexesUpdated []*hourIndexID) (*toUpdate, error) {
	toUpdate := newToUpdate()
	for _, id := range hourIndexesUpdated {
		if err := toUpdate.add(id.Year, id.Month, id.Day); err != nil {
			return nil, err
		}
	}
	return toUpdate, nil
}

func (indexer *simpleIndexer) updateParentsIndexes(parents *toUpdate) error {
	for year, months := range parents.paths {
		yearIdx, err := indexer.loadTotalIndex(year)
		if err != nil {
			return err
		}
		for month, days := range months {
			monthIdx, err := indexer.loadTotalIndex(year, month)
			if err != nil {
				return err
			}
			//TODO: check if yearIdx exists
			yearIdx.Sub(monthIdx)
			for day := range days {
				dayIdx, err := indexer.loadTotalIndex(year, month, day)
				if err != nil {
					return err
				}
				monthIdx.Sub(dayIdx)
				newDayIdx, err := indexer.calculateDayIndex(year, month, day)
				if err != nil {
					return err
				}
				if err := indexer.writeTotalIndexToFile(newDayIdx, year, month, day); err != nil {
					return err
				}
				monthIdx.Add(newDayIdx)
			}
			if err := indexer.writeTotalIndexToFile(monthIdx, year, month); err != nil {
				return err
			}
			yearIdx.Add(monthIdx)
		}
		if err := indexer.writeTotalIndexToFile(yearIdx, year); err != nil {
			return err
		}
	}
	return nil
}

func (indexer *simpleIndexer) calculateDayIndex(parts ...int) (*index, error) {
	index := newIndex()
	for hour := 0; hour < 24; hour++ {
		hourData, err := indexer.loadHourIndex(hour, parts)
		if err != nil {
			return nil, err
		}
		if hourData == nil {
			continue
		}
		index.AddWithourRecalculation(hourData.Index)
	}
	index.CountQueries()
	return index, nil
}

func (indexer *simpleIndexer) loadHourIndex(hour int, parts []int) (*hourData, error) {
	hourIndexFileParts := []string{indexer.getIndexFileDir(parts), strconv.Itoa(hour)}
	hourIndexFilePath := strings.Join(hourIndexFileParts, string(os.PathSeparator))
	_, err := os.Stat(hourIndexFilePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	index := &hourData{}
	indexer.fileMux.Lock()
	defer indexer.fileMux.Unlock()

	f, err := os.Open(hourIndexFilePath)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	gobDecoder := gob.NewDecoder(f)
	if err := gobDecoder.Decode(index); err != nil {
		return nil, err
	}
	return index, nil
}

func (indexer *simpleIndexer) loadTotalIndex(parts ...int) (*index, error) {
	totalIndexFileParts := []string{indexer.getIndexFileDir(parts), "total"}
	totalIndexFilePath := strings.Join(totalIndexFileParts, string(os.PathSeparator))
	_, err := os.Stat(totalIndexFilePath)
	if os.IsNotExist(err) {
		idx := newIndex()
		return idx, nil
	}
	if err != nil {
		return nil, err
	}
	idx, err := indexer.loadIndex(totalIndexFilePath)
	if err != nil {
		return nil, err
	}
	return idx, err
}

func (indexer *simpleIndexer) getIndexFileDir(parts []int) string {
	pathsParts := []string{"./indexes"}
	for _, p := range parts {
		pathsParts = append(pathsParts, strconv.Itoa(p))
	}
	dir := strings.Join(pathsParts, string(os.PathSeparator))
	return dir
}

func (indexer *simpleIndexer) loadIndex(indexFilePath string) (*index, error) {
	idx := &index{}

	indexer.fileMux.Lock()
	defer indexer.fileMux.Unlock()

	f, err := os.Open(indexFilePath)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	gobDecoder := gob.NewDecoder(f)
	if err := gobDecoder.Decode(idx); err != nil {
		return nil, err
	}
	return idx, nil
}

func (indexer *simpleIndexer) writeTotalIndexToFile(idx *index, pathParts ...int) error {
	// TODO: move to function
	fileDir := indexer.getIndexFileDir(pathParts)
	filePath := strings.Join([]string{fileDir, "total"}, string(os.PathSeparator))

	indexer.fileMux.Lock()
	defer indexer.fileMux.Unlock()

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	if err := idx.WriteTo(f); err != nil {
		if err := f.Close(); err != nil {
			return err
		}
		return err
	}
	return nil
}

func (indexer *simpleIndexer) writeIndexToFile(dir string, data *hourData) (string, error) {
	fileName := strconv.Itoa(data.Hour.Hour())
	filePath := fmt.Sprintf("%s%c%s", dir, os.PathSeparator, fileName)

	indexer.fileMux.Lock()
	defer indexer.fileMux.Unlock()

	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	gobEncoder := gob.NewEncoder(f)
	if err := gobEncoder.Encode(data); err != nil {
		if err := f.Close(); err != nil {
			return "", err
		}
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return filePath, nil
}

func createPath(path string) error {
	if err := os.MkdirAll(path, 0766); err != nil {
		return err
	}
	return nil
}
