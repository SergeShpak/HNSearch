package regexp

import (
	"fmt"
	"regexp"
)

type RegexpParser struct {
	dateTimeSeparatorRegexp *regexp.Regexp
	dateTimeRegexp          *regexp.Regexp
	numbersRegexp           *regexp.Regexp
}

func NewRegexpParser() (*RegexpParser, error) {
	parser := &RegexpParser{}
	var err error
	parser.dateTimeSeparatorRegexp, err = regexp.Compile(dateTimeSeparatorRegexpStr)
	if err != nil {
		return nil, err
	}
	parser.dateTimeRegexp, err = regexp.Compile(dateTimeRegexpStr)
	if err != nil {
		return nil, err
	}
	parser.numbersRegexp, err = regexp.Compile(numbersRegexpStr)
	if err != nil {
		return nil, err
	}
	return parser, nil
}

func (p *RegexpParser) ExtractNumbers(str string) []string {
	numbers := p.numbersRegexp.FindAllString(str, -1)
	if len(numbers) == 1 && numbers[0] == "" {
		return []string{}
	}
	return numbers
}

func (p *RegexpParser) DateSplitDateTime(s string) ([]string, error) {
	if !reIsFullMatch(p.dateTimeRegexp, s) {
		return nil, fmt.Errorf("string \"%s\" is not a valid dateTime string", s)
	}
	dateTimeParts := p.dateTimeSeparatorRegexp.Split(s, -1)
	if len(dateTimeParts) < 1 {
		return nil, fmt.Errorf("dateTime string \"%s\" cannot be splitted", s)
	}
	if len(dateTimeParts) > 2 {
		return nil, fmt.Errorf("dateTime string \"%s\" consists of %d parts, while 2 parts were expected", s, len(dateTimeParts))
	}
	if len(dateTimeParts) == 1 {
		dateTimeParts = append(dateTimeParts, "")
	}
	return dateTimeParts, nil
}

func reIsFullMatch(re *regexp.Regexp, str string) bool {
	foundIndices := re.FindStringIndex(str)
	if foundIndices == nil || len(foundIndices) == 0 {
		return false
	}
	if foundIndices[0] == 0 && foundIndices[1] == len(str) {
		return true
	}
	return false
}
