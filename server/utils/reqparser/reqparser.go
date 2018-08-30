package reqparser

import (
	"github.com/SergeyShpak/HNSearch/server/utils/reqparser/regexp"
)

type Parser interface {
	ExtractNumbers(s string) []string
	DateSplitDateTime(s string) ([]string, error)
}

func GetRequestsParser() (Parser, error) {
	parser, err := regexp.NewRegexpParser()
	if err != nil {
		return nil, err
	}
	return parser, nil
}
