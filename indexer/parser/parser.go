package parser

import (
	"fmt"

	"github.com/SergeyShpak/HNSearch/indexer/config"
)

type Parser interface {
	Parse(s string) (*Query, error)
}

func NewParser(c *config.Parser) (Parser, error) {
	if c == nil {
		return nil, fmt.Errorf("passed parser configuration is nil")
	}
	if c.Simple != nil {
		return newSimpleParser(c.Simple), nil
	}
	return nil, fmt.Errorf("no parser found in configuration")
}
