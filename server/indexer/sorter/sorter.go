package sorter

import (
	"io"
)

type Sorter interface {
	SortSet(r io.Reader) error
}
