package sorter

import (
	"fmt"
	"io"

	"github.com/SergeyShpak/HNSearch/indexer/config"
)

type Sorter interface {
	SortSet(r io.Reader) (string, error)
}

func NewSorter(c *config.Sorter) (Sorter, error) {
	if c == nil {
		return nil, fmt.Errorf("passed sorter configuration is nil")
	}
	if c.Simple != nil {
		sorter, err := newSimpleSorter(c.Simple)
		if err != nil {
			return nil, err
		}
		return sorter, nil
	}
	return nil, fmt.Errorf("no sorter found in configuration")
}
