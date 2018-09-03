package sorter

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"
)

type SimpleSorter struct{}

func NewSimpleSorter() *SimpleSorter {
	sorter := &SimpleSorter{}
	return sorter
}

func (sorter *SimpleSorter) SortSet(r io.Reader) error {
	tmpDir, err := sorter.splitLargeSet(r)
	if err != nil {
		return err
	}
	resultFile, err := sorter.mergeDataSets(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return err
	}
	if err := os.Rename(resultFile, "sorted.tsv"); err != nil {
		os.RemoveAll(tmpDir)
		return err
	}
	os.RemoveAll(tmpDir)
	return nil
}

func (sorter *SimpleSorter) splitLargeSet(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	maxBufLen := 50 * 1024 * 1024 // 50 MB
	lines := make([]string, 0)
	currLen := 0
	chunkNumber := 0
	dir := fmt.Sprintf("tmp_chunks_%v", time.Now().Unix())
	if err := os.MkdirAll(dir, 0777); err != nil {
		return dir, err
	}
	fileTemplate := fmt.Sprintf("./%s%cchunk_%%d", dir, os.PathSeparator)
	cleanFn := func() error {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		return nil
	}
	writeToFileFn := func() error {
		sort.Slice(lines, func(i int, j int) bool {
			dateI := strings.Split(lines[i], "\t")[0]
			dateJ := strings.Split(lines[j], "\t")[0]
			return dateI < dateJ
		})
		fileName := fmt.Sprintf(fileTemplate, chunkNumber)
		fmt.Println("file: ", fileName)
		if err := dumpToFile(fileName, lines); err != nil {
			fmt.Println("ERROR!: ", err)
			if err := cleanFn(); err != nil {
				return err
			}
			return err
		}
		chunkNumber++
		lines = lines[:0]
		return nil
	}
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		currLen += len(line)
		lines = append(lines, line)
		if currLen > maxBufLen {
			if err := writeToFileFn(); err != nil {
				return dir, err
			}
			currLen = 0
		}
	}
	if err := writeToFileFn(); err != nil {
		return dir, err
	}
	return dir, nil
}

func (sorter *SimpleSorter) mergeDataSets(dir string) (string, error) {
	files, err := getFiles(dir)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("No files found in the directory \"%s\"", dir)
	}
	if len(files) == 1 {
		// TODO: change
		return "", nil
	}
	fileNames := make([]string, 0, len(files))
	for _, f := range files {
		fileNames = append(fileNames, fmt.Sprintf("%s%c%s", dir, os.PathSeparator, f.Name()))
	}
	resultFile, err := sorter.mergeDataSetsAux(fileNames[0], fileNames[1], fileNames[2:], dir)
	if err != nil {
		return "", err
	}
	return resultFile, nil
}

func (sorter *SimpleSorter) mergeDataSetsAux(first string, second string, tail []string, tmpDir string) (string, error) {
	tmpFile, err := mergeFiles(first, second, tmpDir)
	if err != nil {
		return "", err
	}
	if len(tail) == 0 {
		return tmpFile, nil
	}
	if len(tail) == 1 {
		tmpFile, err = mergeFiles(tmpFile, tail[0], tmpDir)
		if err != nil {
			return "", err
		}
		return tmpFile, err
	}
	return sorter.mergeDataSetsAux(tmpFile, tail[0], tail[1:], tmpDir)
}

func dumpToFile(fileName string, data []string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(f)
	for _, s := range data {
		writer.WriteString(s)
	}
	return nil
}

func getFiles(dirName string) ([]os.FileInfo, error) {
	dirEntries, err := ioutil.ReadDir(dirName)
	if err != nil {
		return nil, err
	}
	for i, e := range dirEntries {
		if e.IsDir() {
			dirEntries = append(dirEntries[:i], dirEntries[i+1:]...)
		}
	}
	return dirEntries, nil
}

func mergeFiles(first string, second string, tmpDir string) (string, error) {
	maxBuf := 12 * 1024 * 1024
	firstF, err := os.Open(first)
	if err != nil {
		return "", err
	}
	secondF, err := os.Open(second)
	if err != nil {
		return "", err
	}
	fScanner := bufio.NewScanner(firstF)
	var fLines []string
	var isFirstFinished bool
	sScanner := bufio.NewScanner(secondF)
	var sLines []string
	var isSecondFinished bool
	tmpFile, err := createTmpFile(tmpDir)
	if err != nil {
		return "", err
	}
	for {
		if len(fLines) == 0 {
			if isFirstFinished {
				break
			}
			fLines, isFirstFinished = readFromScanner(fScanner, maxBuf)
		}
		if len(sLines) == 0 {
			if isSecondFinished {
				break
			}
			sLines, isSecondFinished = readFromScanner(sScanner, maxBuf)
		}
		var merged []string
		merged, fLines, sLines = mergeLines(fLines, sLines)
		if err := appendToFile(merged, tmpFile); err != nil {
			return "", err
		}
	}
	if len(fLines) != 0 {
		for {
			if err := appendToFile(fLines, tmpFile); err != nil {
				return "", err
			}
			if isFirstFinished {
				break
			}
			fLines, isFirstFinished = readFromScanner(fScanner, maxBuf)
		}
		return tmpFile, nil
	}
	for {
		if err := appendToFile(sLines, tmpFile); err != nil {
			return "", err
		}
		if isSecondFinished {
			break
		}
		sLines, isSecondFinished = readFromScanner(sScanner, maxBuf)
	}
	return tmpFile, nil
}

func readFromScanner(scanner *bufio.Scanner, max int) ([]string, bool) {
	lines := make([]string, 0)
	linesLen := 0
	for scanner.Scan() {
		l := scanner.Text()

		lines = append(lines, l)
		linesLen += len(l)
		if linesLen >= max {
			return lines, false
		}
	}
	return lines, true
}

func mergeLines(first []string, second []string) ([]string, []string, []string) {
	var fIdx int
	var sIdx int
	merged := make([]string, 0)
	for {
		if fIdx >= len(first) {
			return merged, nil, second[sIdx:]
		}
		if sIdx >= len(second) {
			return merged, first[fIdx:], nil
		}
		if first[fIdx] < second[sIdx] {
			merged = append(merged, first[fIdx])
			fIdx++
			continue
		}
		merged = append(merged, second[sIdx])
		sIdx++
	}
}

func createTmpFile(tmpDir string) (string, error) {
	timeNow := time.Now().Unix()
	fileNameTempl := fmt.Sprintf("%s%cmerged_%%d", tmpDir, os.PathSeparator)
	errCtr := 0
	for {
		// TODO: make configurable??
		if errCtr == 10 {
			break
		}
		fileName := fmt.Sprintf(fileNameTempl, timeNow)
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			f, err := os.Create(fileName)
			if err != nil {
				return "", err
			}
			if err := f.Close(); err != nil {
				return "", err
			}
			return fileName, nil
		}
		timeNow++
		errCtr++
	}
	return "", fmt.Errorf("could not create a new temporary file")
}

func appendToFile(buf []string, file string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	for _, s := range buf {
		if _, err := f.WriteString(s + "\n"); err != nil {
			if err := f.Close(); err != nil {
				return err
			}
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}
