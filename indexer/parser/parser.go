package parser

import (
	"fmt"

	"github.com/SergeyShpak/HNSearch/indexer/config"
)

type Parser interface {
	Parse(s string) (*Query, error)
}

func NewParser(c *config.Config) (Parser, error) {
	if c == nil {
		return nil, fmt.Errorf("passed parser configuration is nil")
	}
	if c.Parser.Simple != nil {
		return newSimpleParser(c.Parser.Simple), nil
	}
	return nil, fmt.Errorf("no parser found in configuration")
}
