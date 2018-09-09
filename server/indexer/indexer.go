package indexer

import (
	"fmt"
	"time"

	"github.com/SergeyShpak/HNSearch/server/config"
	"github.com/SergeyShpak/HNSearch/server/indexer/http"
)

type Indexer interface {
	CountDistinctQueries(from *time.Time, to *time.Time) (int, error)
}

func NewIndexer(c *config.Config) (Indexer, error) {
	if c == nil {
		return nil, fmt.Errorf("passed config in nil")
	}
	if c.Indexer == nil {
		return nil, fmt.Errorf("passed indexer config is nil")
	}
	if c.Indexer.HTTP != nil {
		indexer, err := http.NewHTTPIndexer(c)
		if err != nil {
			return nil, err
		}
		return indexer, nil
	}
	return nil, fmt.Errorf("no valid indexer configuration found")
}
