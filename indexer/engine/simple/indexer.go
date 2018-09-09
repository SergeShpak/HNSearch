package simple

import (
	"fmt"
	"os"
	"sync"

	"github.com/SergeyShpak/HNSearch/indexer/config"
	"github.com/SergeyShpak/HNSearch/indexer/parser"
	"github.com/SergeyShpak/HNSearch/indexer/sorter"
)

type simpleIndexer struct {
	fileMux    *sync.Mutex
	dataDir    string
	indexesDir string
	parser     parser.Parser
	sorter     sorter.Sorter
}

func NewSimpleIndexer(c *config.Config) (*simpleIndexer, error) {
	if c == nil {
		return nil, fmt.Errorf("nil passed as simple indexer configuration")
	}
	indexer := &simpleIndexer{
		fileMux: &sync.Mutex{},
	}
	var err error
	// TODO: pass the whole configuration
	indexer.parser, err = parser.NewParser(c.Parser)
	if err != nil {
		return nil, err
	}
	indexer.sorter, err = sorter.NewSorter(c)
	if err != nil {
		return nil, err
	}
	if err := initServiceDirs(c, indexer); err != nil {
		return nil, err
	}
	return indexer, nil
}

func initServiceDirs(c *config.Config, indexer *simpleIndexer) error {
	if err := os.MkdirAll(c.Indexer.DataDir, 0755); err != nil {
		return fmt.Errorf("error occurred during data directory creation: %v", err)
	}
	indexer.dataDir = c.Indexer.DataDir
	if err := os.MkdirAll(c.Indexer.IndexesDir, 0755); err != nil {
		return fmt.Errorf("error occurred during indexes directory creation: %v", err)
	}
	indexer.indexesDir = c.Indexer.IndexesDir
	return nil
}
