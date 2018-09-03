package sorter

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestSorter(t *testing.T) {
	sorter := NewSimpleSorter()
	fd, err := os.Open("../../../hn_logs.tsv")
	if err != nil {
		t.Fatalf("error during file open: %v", err)
	}
	if err := sorter.SortSet(fd); err != nil {
		t.Fatalf("error during set sorting: %v", err)
	}
	sortedCount, err := getQueryCount("sorted.tsv")
	if err != nil {
		t.Fatalf("error when counting sorted queries")
	}
	unsortedCount, err := getQueryCount("../../../hn_logs.tsv")
	if err != nil {
		t.Fatalf("error when counting unsorted values")
	}
	for k, v := range sortedCount {
		unsortedValue, ok := unsortedCount[k]
		if !ok {
			t.Fatalf("key %s not found in unsorted", k)
		}
		if v != unsortedValue {
			t.Fatalf("values for key %s is different in unsorted", k)
		}
	}
	for k, v := range unsortedCount {
		sortedValue, ok := sortedCount[k]
		if !ok {
			t.Fatalf("key %s not found in sorted", k)
		}
		if v != sortedValue {
			t.Fatalf("values for key %s is different in sorted", k)
		}
	}
}

func getQueryCount(file string) (map[string]int, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	sortedDict := make(map[string]int)
	for scanner.Scan() {
		line := scanner.Text()
		query := strings.Split(line, "\t")[1]
		if _, ok := sortedDict[query]; !ok {
			sortedDict[query] = 0
		}
		sortedDict[query]++
	}
	return sortedDict, nil
}
