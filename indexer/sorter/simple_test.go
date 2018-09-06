package sorter

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/SergeyShpak/HNSearch/indexer/config"
)

func TestSorter(t *testing.T) {
	c := &config.SimpleSorter{
		Buffer: 50 * 1024 * 1024,
		OutDir: "sorted",
		Parser: &config.Parser{
			Simple: &config.SimpleParser{},
		},
		TmpCreationRetries: 10,
	}
	sorter, err := newSimpleSorter(c)
	if err != nil {
		t.Fatalf("error during sorter initialization: %v", err)
	}
	fd, err := os.Open("../../../hn_logs.tsv")
	if err != nil {
		t.Fatalf("error during file open: %v", err)
	}
	sortedFile, err := sorter.SortSet(fd)
	if err != nil {
		t.Fatalf("error during set sorting: %v", err)
	}
	sortedCount, err := getQueryCount(sortedFile)
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
