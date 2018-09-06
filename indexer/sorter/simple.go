package sorter

/*
TODO:
	1. Move random file generation to a subroutine (use github.com/rs/xid)
	2. Add configuration checking
	3. Pass output directory as a parameter
*/

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/SergeyShpak/HNSearch/indexer/config"
	"github.com/SergeyShpak/HNSearch/indexer/parser"
)

type simpleSorter struct {
	config *config.SimpleSorter
	parser parser.Parser
}

func newSimpleSorter(c *config.SimpleSorter) (*simpleSorter, error) {
	sorter := &simpleSorter{
		config: c,
	}
	var err error
	sorter.parser, err = parser.NewParser(c.Parser)
	if err != nil {
		return nil, fmt.Errorf("error while initializing sorter's parser: %v", err)
	}
	return sorter, nil
}

func (sorter *simpleSorter) SortSet(r io.Reader) (string, error) {
	tmpDir, err := sorter.splitLargeSet(r)
	if err != nil {
		return "", err
	}
	resultFile, err := sorter.mergeDataSets(tmpDir)
	if err != nil {
		//os.RemoveAll(tmpDir)
		return "", err
	}
	fmt.Println("Here!")
	resultName, err := sorter.generateResultName(resultFile)
	if err != nil {
		//os.RemoveAll(tmpDir)
		return "", err
	}
	sortedFile, err := moveFile(resultFile, sorter.config.OutDir, resultName)
	if err != nil {
		//os.RemoveAll(tmpDir)
		return "", err
	}
	//os.RemoveAll(tmpDir)
	return sortedFile, nil
}

func (sorter *simpleSorter) splitLargeSet(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	maxBufLen := sorter.config.Buffer
	lines := make([]string, 0)
	var currLen uint64
	chunkNumber := 0
	dir := fmt.Sprintf("tmp_chunks_%v", time.Now().Unix())
	if err := os.MkdirAll(dir, 0777); err != nil {
		return dir, err
	}
	fileTemplate := fmt.Sprintf("./%s%cchunk_%%d", dir, os.PathSeparator)
	writeToFileFn := func() error {
		sort.Slice(lines, func(i int, j int) bool {
			return lines[i] < lines[j]
		})
		fileName := fmt.Sprintf(fileTemplate, chunkNumber)
		fmt.Println("file: ", fileName)
		if err := dumpToFile(fileName, lines); err != nil {
			fmt.Println("ERROR!: ", err)
			if err := os.RemoveAll(dir); err != nil {
				return err
			}
			return err
		}
		chunkNumber++
		lines = make([]string, 0)
		return nil
	}
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		currLen += uint64(len(line))
		lines = append(lines, line)
		if currLen > maxBufLen {
			if err := writeToFileFn(); err != nil {
				if err := os.RemoveAll(dir); err != nil {
					return "", err
				}
				return "", err
			}
			currLen = 0
		}
	}
	if err := writeToFileFn(); err != nil {
		if err := os.RemoveAll(dir); err != nil {
			return "", err
		}
		return "", err
	}
	return dir, nil
}

func (sorter *simpleSorter) mergeDataSets(dir string) (string, error) {
	files, err := getFiles(dir)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("No files found in the directory \"%s\"", dir)
	}
	if len(files) == 1 {
		path := fmt.Sprintf("%s%c%s", dir, os.PathSeparator, files[0].Name())
		return path, nil
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

func (sorter *simpleSorter) mergeDataSetsAux(first string, second string, tail []string, tmpDir string) (string, error) {
	tmpFile, err := sorter.mergeFiles(first, second, tmpDir)
	if err != nil {
		return "", err
	}
	if len(tail) == 0 {
		return tmpFile, nil
	}
	if len(tail) == 1 {
		tmpFile, err = sorter.mergeFiles(tmpFile, tail[0], tmpDir)
		if err != nil {
			return "", err
		}
		return tmpFile, err
	}
	return sorter.mergeDataSetsAux(tmpFile, tail[0], tail[1:], tmpDir)
}

func (sorter *simpleSorter) mergeFiles(first string, second string, tmpDir string) (string, error) {
	fileMaxBuf := sorter.config.Buffer / 4
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
	tmpFile, err := sorter.createTmpFile(tmpDir)
	if err != nil {
		return "", err
	}
	for {
		if len(fLines) == 0 {
			if isFirstFinished {
				break
			}
			fLines, isFirstFinished = readFromScanner(fScanner, fileMaxBuf)
		}
		if len(sLines) == 0 {
			if isSecondFinished {
				break
			}
			sLines, isSecondFinished = readFromScanner(sScanner, fileMaxBuf)
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
			fLines, isFirstFinished = readFromScanner(fScanner, fileMaxBuf)
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
		sLines, isSecondFinished = readFromScanner(sScanner, fileMaxBuf)
	}
	return tmpFile, nil
}

func (sorter *simpleSorter) createTmpFile(tmpDir string) (string, error) {
	timeNow := time.Now().Unix()
	fileNameTempl := fmt.Sprintf("%s%cmerged_%%d", tmpDir, os.PathSeparator)
	errCtr := 0
	for {
		if errCtr >= sorter.config.TmpCreationRetries {
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

func (sorter *simpleSorter) generateResultName(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	firstLine := scanner.Text()
	q, err := sorter.parser.Parse(firstLine)
	if err != nil {
		fmt.Println("First line: ", firstLine)
		return "", err
	}
	firstDate := q.Date.Unix()
	var line string
	for scanner.Scan() {
		line = scanner.Text()
	}
	q, err = sorter.parser.Parse(line)
	if err != nil {
		fmt.Println("Last line: ", line)
		return "", err
	}
	secondDate := q.Date.Unix()
	resultName := fmt.Sprintf("sorted_%d_%d", firstDate, secondDate)
	return resultName, nil
}

func readFromScanner(scanner *bufio.Scanner, maxBuf uint64) ([]string, bool) {
	lines := make([]string, 0)
	var linesLen uint64
	for scanner.Scan() {
		l := scanner.Text()

		lines = append(lines, l)
		linesLen += uint64(len(l))
		if linesLen >= maxBuf {
			return lines, false
		}
	}
	return lines, true
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
	writer.Flush()
	if err := f.Close(); err != nil {
		return err
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

func moveFile(filePath string, newDir string, newFileName string) (string, error) {
	if err := os.MkdirAll(newDir, 0777); err != nil {
		return "", err
	}
	newFilePath := fmt.Sprintf("%s%c%s", newDir, os.PathSeparator, newFileName)
	if err := os.Rename(filePath, newFilePath); err != nil {
		return "", err
	}
	return newFilePath, nil
}
