package simple

import (
	"fmt"
	"testing"
	"time"

	"github.com/SergeyShpak/HNSearch/indexer/config"
)

/*
func TestGetIndexesPath(t *testing.T) {
	testCases := []struct {
		from time.Time
		to   time.Time
	}{
		{
			from: time.Date(2015, 4, 6, 7, 31, 0, 0, time.UTC),
			to:   time.Date(2015, 8, 8, 10, 33, 0, 0, time.UTC),
		},
	}
	for _, tc := range testCases {
		datesDiff := newDatesDiff(&tc.from, &tc.to)
		paths, minSecPaths := datesDiff.getIndexesPaths()
		for _, msp := range minSecPaths {
			fmt.Printf("min sec: %v\n", msp)
		}
		for _, p := range paths {
			fmt.Printf("paths: %v\n", p)
		}
		//fmt.Printf("test %d: path: %v, minSecPaths: %v", i, paths, minSecPaths)
	}
}
*/
func TestCountDistinctQueries(t *testing.T) {
	testCases := []struct {
		from time.Time
		to   time.Time
	}{
		{
			from: time.Date(2015, 1, 0, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2016, 1, 0, 0, 0, 0, 0, time.UTC),
		},
	}
	config := config.GetDefaultConfig()
	config.Indexer.DataDir = "/home/bp/go/src/github.com/SergeyShpak/HNSearch/indexer/service/data"
	config.Indexer.IndexesDir = "/home/bp/go/src/github.com/SergeyShpak/HNSearch/indexer/service/indexes"
	indexer, err := newSimpleIndexer(config)
	if err != nil {
		t.Fatalf("could not initialize the simple indexer: %v", err)
	}
	for _, tc := range testCases {
		queriesCount, err := indexer.CountDistinctQueries(&tc.from, &tc.to)
		if err != nil {
			t.Fatalf("an error occurred: %v", err)
		}
		fmt.Println("queries count: ", queriesCount)
		//fmt.Printf("test %d: path: %v, minSecPaths: %v", i, paths, minSecPaths)
	}
}
