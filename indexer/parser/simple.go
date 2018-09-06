package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/SergeyShpak/HNSearch/indexer/config"
)

type simpleParser struct{}

func newSimpleParser(c *config.SimpleParser) *simpleParser {
	return &simpleParser{}
}

// Format: {yyyy-mm-dd \space hh:mm:ss \t query}
func (parser *simpleParser) Parse(s string) (*Query, error) {
	dateAndQuery, err := parser.separateDateAndQuery(s)
	if err != nil {
		return nil, err
	}
	date, err := parser.strToTime(dateAndQuery[0])
	if err != nil {
		return nil, err
	}
	e := &Query{
		Date:  date,
		Query: dateAndQuery[1],
	}
	return e, nil
}

func (parser *simpleParser) separateDateAndQuery(dateQueryStr string) ([]string, error) {
	dateAndQuery, err := splitAndCheck(dateQueryStr, "\t", 2)
	if err != nil {
		return nil, err
	}
	return dateAndQuery, nil
}

func (parser *simpleParser) strToTime(dateStr string) (*time.Time, error) {
	dateAndTime, err := parser.separateDateAndTime(dateStr)
	if err != nil {
		return nil, err
	}
	dateRFCStr := dateAndTime[0] + "T" + dateAndTime[1] + "Z"
	date, err := time.Parse(time.RFC3339, dateRFCStr)
	if err != nil {
		return nil, err
	}
	return &date, nil
}

func (parser *simpleParser) separateDateAndTime(dateTimeStr string) ([]string, error) {
	dateAndHour, err := splitAndCheck(dateTimeStr, " ", 2)
	if err != nil {
		return nil, err
	}
	return dateAndHour, nil
}

func splitAndCheck(s string, sep string, expectedPartsCount int) ([]string, error) {
	sParts := strings.Split(s, sep)
	if len(sParts) != expectedPartsCount {
		return nil, fmt.Errorf("string \"%s\" contains %d parts separated with \"%s\", expected %d parts", s, len(sParts), sep, expectedPartsCount)
	}
	return sParts, nil
}
