package engine

import (
	"fmt"
	"time"

	"github.com/SergeyShpak/HNSearch/indexer/config"
	"github.com/SergeyShpak/HNSearch/indexer/engine/simple"
)

type Indexer interface {
	IndexData() error
	CountDistinctQueries(from *time.Time, to *time.Time) (int, error)
}

func NewIndexer(c *config.Config) (Indexer, error) {
	if c == nil {
		return nil, fmt.Errorf("passed indexer configuration is nil")
	}
	if c.Indexer.Simple != nil {
		// TODO: move simple indexer to an internal module
		indexer, err := simple.NewSimpleIndexer(c)
		if err != nil {
			return nil, err
		}
		return indexer, nil
	}
	return nil, fmt.Errorf("no indexer found in configuration")
}
