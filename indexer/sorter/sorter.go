package sorter

import (
	"fmt"
	"io"

	"github.com/SergeyShpak/HNSearch/indexer/config"
)

type Sorter interface {
	SortSet(r io.Reader) (string, error)
}

func NewSorter(c *config.Config) (Sorter, error) {
	if c == nil {
		return nil, fmt.Errorf("passed sorter configuration is nil")
	}
	if c.Sorter.Simple != nil {
		sorter, err := newSimpleSorter(c)
		if err != nil {
			return nil, err
		}
		return sorter, nil
	}
	return nil, fmt.Errorf("no sorter found in configuration")
}
