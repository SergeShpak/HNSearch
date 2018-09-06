package engine

import (
	"os"
	"testing"

	"github.com/SergeyShpak/HNSearch/server/indexer/config"
)

func TestSimpleIndexer(t *testing.T) {
	c := config.GetDefaultConfig()
	indexer, err := NewIndexer(c.Indexer)
	if err != nil {
		t.Fatalf("error when initializing an indexer: %v", err)
	}
	sortedPath := "../sorter/sorted/sorted_1438387200_1438667699"
	f, err := os.Open(sortedPath)
	if err != nil {
		t.Fatalf("could not open the sorted file: %v", err)
	}
	defer f.Close()
	if err := indexer.UpdateIndices(f); err != nil {
		t.Fatalf("error occured during indexes adding: %v", err)
	}
}
