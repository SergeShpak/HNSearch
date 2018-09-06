package engine

import (
	"fmt"
	"io"

	"github.com/SergeyShpak/HNSearch/server/indexer/config"
)

type Indexer interface {
	UpdateIndices(r io.Reader) error
}

func NewIndexer(c *config.Indexer) (Indexer, error) {
	if c == nil {
		return nil, fmt.Errorf("passed indexer configuration is nil")
	}
	if c.Simple != nil {
		indexer, err := newSimpleIndexer(c.Simple)
		if err != nil {
			return nil, err
		}
		return indexer, nil
	}
	return nil, fmt.Errorf("no indexer found in configuration")
}
