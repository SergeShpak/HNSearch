package simple

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SergeyShpak/HNSearch/indexer/parser"
	"github.com/SergeyShpak/HNSearch/indexer/server/types"
)

var totalIndexName = "total"
var fastFetchDataName = "fast-fetch"

var indexedFilesLogName = "log"
var dataServiceFileNames = []string{indexedFilesLogName}

type indexedFilesLog struct {
	Files map[string]byte
}

func newIndexedFilesLog() *indexedFilesLog {
	l := &indexedFilesLog{}
	l.Files = make(map[string]byte)
	return l
}

func loadIndexedFilesLog(logFilePath string) (*indexedFilesLog, error) {
	logFileExists, err := doesFileExist(logFilePath)
	if err != nil {
		return nil, err
	}
	if !logFileExists {
		return nil, nil
	}
	fi, err := os.Stat(logFilePath)
	if err != nil {
		return nil, err
	}
	if fi.Size() == 0 {
		return nil, nil
	}
	f, err := os.Open(logFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := gob.NewDecoder(f)
	l := &indexedFilesLog{}
	if err := decoder.Decode(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *indexedFilesLog) Log(path string) {
	l.Files[path] = '1'
}

func (l *indexedFilesLog) StoreIndexedFilesLog(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(f)
	if err := encoder.Encode(l); err != nil {
		if err := f.Close(); err != nil {
			return err
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func (l *indexedFilesLog) IsFileIndexed(filePath string) bool {
	if _, ok := l.Files[filePath]; !ok {
		return false
	}
	return true
}

type hourData struct {
	Hour      *time.Time
	Partition [60][60]map[string]int
	Index     *index
}

type index struct {
	QueriesDict  map[string]int
	QueriesCount []*types.Query
}

func newIndex() *index {
	idx := &index{
		QueriesDict:  make(map[string]int),
		QueriesCount: make([]*types.Query, 0),
	}
	return idx
}

func (idx *index) CountQueries() {
	entries := make([]*types.Query, 0)
	for query, count := range idx.QueriesDict {
		e := &types.Query{
			Query: query,
			Count: count,
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i int, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	idx.QueriesCount = entries
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
	needRecalculation := idx.AddWithoutRecalculation(other)
	if needRecalculation {
		idx.CountQueries()
	}
}

func (idx *index) AddWithoutRecalculation(other *index) bool {
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

func (idx *index) AddMapWithoutRecalculation(m map[string]int) bool {
	other := newIndex()
	other.QueriesDict = m
	return idx.AddWithoutRecalculation(other)
}

func (idx *index) WriteTo(w io.Writer) error {
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(idx); err != nil {
		return err
	}
	return nil
}

func (idx *index) WriteFastFetchData(w io.Writer) error {
	writer := bufio.NewWriter(w)
	distinctQueriesCount := len(idx.QueriesDict)
	fmt.Fprintln(writer, distinctQueriesCount)
	for _, q := range idx.QueriesCount {
		l := fmt.Sprintf("%s\t%d", q.Query, q.Count)
		fmt.Fprintln(writer, l)
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}

type hourIndexID struct {
	Day   int
	Month int
	Year  int
}

type toUpdate struct {
	paths map[int]map[int]map[int]byte
}

func (indexer *simpleIndexer) IndexData() error {
	dataFilesPaths, err := indexer.getUnidexedFiles()
	if err != nil {
		return err
	}
	fmt.Println("Unindexed files: ", dataFilesPaths)
	l, err := indexer.loadIndexedFilesLog()
	if err != nil {
		return err
	}
	for _, path := range dataFilesPaths {
		if err := indexer.addDataToIndexes(path); err != nil {
			// TODO: log and continue
			return err
		}
		l.Log(path)
	}
	if err := indexer.storeIndexedFilesLog(l); err != nil {
		return err
	}
	return nil
}

func (indexer *simpleIndexer) addDataToIndexes(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	sortedPath, err := indexer.sorter.SortSet(f)
	f.Close()
	if err != nil {
		return err
	}

	fmt.Println("Sorted path: ", sortedPath)

	f, err = os.Open(sortedPath)
	if err != nil {
		if err := os.Remove(sortedPath); err != nil {
			return err
		}
		return err
	}
	if err := indexer.updateIndexes(f); err != nil {
		f.Close()
		if err := os.Remove(sortedPath); err != nil {
			return err
		}
		return err
	}
	f.Close()
	if err := os.Remove(sortedPath); err != nil {
		return err
	}
	return nil

}

func (indexer *simpleIndexer) loadIndexedFilesLog() (*indexedFilesLog, error) {
	logFilePath := strings.Join([]string{indexer.dataDir, indexedFilesLogName}, string(os.PathSeparator))
	log, err := loadIndexedFilesLog(logFilePath)
	if err != nil {
		return nil, err
	}
	if log == nil {
		log = newIndexedFilesLog()
	}
	return log, err
}

func (indexer *simpleIndexer) storeIndexedFilesLog(l *indexedFilesLog) error {
	logFilePath := strings.Join([]string{indexer.dataDir, indexedFilesLogName}, string(os.PathSeparator))
	if err := l.StoreIndexedFilesLog(logFilePath); err != nil {
		return err
	}
	return nil
}

func (indexer *simpleIndexer) getUnidexedFiles() ([]string, error) {
	files, err := ioutil.ReadDir(indexer.dataDir)
	if err != nil {
		return nil, err
	}
	log, err := indexer.loadIndexedFilesLog()
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if indexer.isServiceFile(f) {
			continue
		}
		filePath := strings.Join([]string{indexer.dataDir, f.Name()}, string(os.PathSeparator))
		if log.IsFileIndexed(filePath) {
			continue
		}
		paths = append(paths, filePath)
	}
	return paths, nil
}

func (indexer *simpleIndexer) isServiceFile(f os.FileInfo) bool {
	for _, serviceF := range dataServiceFileNames {
		if f.Name() == serviceF {
			return true
		}
	}
	return false
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

func (indexer *simpleIndexer) updateIndexes(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	var data *hourData
	var isFinished bool
	var err error
	indexesUpdated := make([]*hourIndexID, 0)
	var lastQ *parser.Query
	for {
		data, isFinished, lastQ, err = indexer.scanHour(scanner, lastQ)
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

func (indexer *simpleIndexer) scanHour(scanner *bufio.Scanner, firstHourQuery *parser.Query) (*hourData, bool, *parser.Query, error) {
	if !scanner.Scan() {
		return nil, true, nil, nil
	}
	line := scanner.Text()
	q, err := indexer.parser.Parse(line)
	if err != nil {
		return nil, false, nil, err
	}
	thisHour := time.Date(q.Date.Year(), q.Date.Month(), q.Date.Day(), q.Date.Hour(), 0, 0, 0, q.Date.Location())
	nextHour := time.Date(q.Date.Year(), q.Date.Month(), q.Date.Day(), q.Date.Hour()+1, 0, 0, 0, q.Date.Location())
	currMinute := q.Date.Minute()
	currSecond := q.Date.Second()
	data := &hourData{
		Hour: &thisHour,
	}
	data.Partition[currMinute][currSecond] = make(map[string]int)
	data.Partition[currMinute][currSecond][q.Query] = 1
	if firstHourQuery != nil {
		if _, ok := data.Partition[currMinute][currSecond][firstHourQuery.Query]; !ok {
			data.Partition[currMinute][currSecond][firstHourQuery.Query] = 0
		}
		data.Partition[currMinute][currSecond][firstHourQuery.Query]++
	}
	for scanner.Scan() {
		line := scanner.Text()
		q, err := indexer.parser.Parse(line)
		if err != nil {
			return nil, false, nil, err
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
		return data, false, q, nil
	}
	return data, true, nil, nil
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

func (indexer *simpleIndexer) createHourIndexPath(id *hourIndexID) (string, error) {
	if id == nil {
		return "", fmt.Errorf("passed hour index is nil")
	}
	pathParts := []string{indexer.indexesDir, strconv.Itoa(id.Year), strconv.Itoa(id.Month), strconv.Itoa(id.Day)}
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
				if err := indexer.writeIndexes(newDayIdx, year, month, day); err != nil {
					return err
				}
				monthIdx.Add(newDayIdx)
			}
			if err := indexer.writeIndexes(monthIdx, year, month); err != nil {
				return err
			}
			yearIdx.Add(monthIdx)
		}
		if err := indexer.writeIndexes(yearIdx, year); err != nil {
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
		index.AddWithoutRecalculation(hourData.Index)
	}
	index.CountQueries()
	return index, nil
}

func (indexer *simpleIndexer) loadHourIndex(hour int, parts []int) (*hourData, error) {
	hourIndexFileParts := []string{indexer.getIndexFileDir(parts), strconv.Itoa(hour)}
	hourIndexFilePath := strings.Join(hourIndexFileParts, string(os.PathSeparator))
	fileExists, err := doesFileExist(hourIndexFilePath)
	if err != nil {
		return nil, err
	}
	if !fileExists {
		return nil, nil
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
	//TODO: move to a subroutine
	totalIndexFileParts := []string{indexer.getIndexFileDir(parts), totalIndexName}
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
	pathsParts := []string{indexer.indexesDir}
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

func (indexer *simpleIndexer) writeIndexes(idx *index, pathParts ...int) error {
	if err := indexer.writeTotalIndexToFile(idx, pathParts...); err != nil {
		return err
	}
	if err := indexer.writeFastFetchDataToFile(idx, pathParts...); err != nil {
		return err
	}
	return nil
}

func (indexer *simpleIndexer) writeTotalIndexToFile(idx *index, pathParts ...int) error {
	// TODO: move to function
	fileDir := indexer.getIndexFileDir(pathParts)
	filePath := strings.Join([]string{fileDir, totalIndexName}, string(os.PathSeparator))

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

func (indexer *simpleIndexer) writeFastFetchDataToFile(idx *index, pathParts ...int) error {
	// TODO: move to function
	fileDir := indexer.getIndexFileDir(pathParts)
	filePath := strings.Join([]string{fileDir, fastFetchDataName}, string(os.PathSeparator))

	indexer.fileMux.Lock()
	defer indexer.fileMux.Unlock()

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	if err := idx.WriteFastFetchData(f); err != nil {
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

func doesFileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
