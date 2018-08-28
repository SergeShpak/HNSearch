package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var AcceptableDateSeparators = []string{"-", "/", "."}
var AcceptableTimeSeparators = []string{":"}

var dateSeparatorsRegexpStr = fmt.Sprintf("(%s)", getSeparatorsString(AcceptableDateSeparators))
var dateRegexpStr = fmt.Sprintf("(19|20)[0-9]{2}(%[1]s((1[0-2])|(0?[1-9]))?(%[1]s((3[0-1])|([1-2][0-9])|(0?[1-9]))?)?)?", dateSeparatorsRegexpStr)
var timeSeparatorsRegexpStr = fmt.Sprintf("(%s)", getSeparatorsString(AcceptableTimeSeparators))
var timeRegexpStr = fmt.Sprintf("((1[0-9])|(2[0-3])|(0?[0-9]))((%[1]s(([1-5][0-9])|(0?[0-9]))?)(%[1]s(([1-5][0-9])|(0?[0-9]))?)?)?", timeSeparatorsRegexpStr)
var dateTimeSeparatorRegexpStr = " "
var dateTimeSeparatorRegexp = regexp.MustCompile(dateTimeSeparatorRegexpStr)
var dateTimeRegexpStr = fmt.Sprintf("(%s)((%s)(%s))?", dateRegexpStr, dateTimeSeparatorRegexpStr, timeRegexpStr)
var DateTimeRegexp = regexp.MustCompile(dateTimeRegexpStr)

var numbersRegexp = regexp.MustCompile("[0-9]*")

func getSeparatorsString(separators []string) string {
	wrappedSeparators := make([]string, 0)
	for _, sep := range separators {
		escapedSep := escapeSeparator(sep)
		wrappedSep := "(" + escapedSep + ")"
		wrappedSeparators = append(wrappedSeparators, wrappedSep)
	}
	separatorsStr := strings.Join(wrappedSeparators, "|")
	return separatorsStr
}

func escapeSeparator(sep string) string {
	re := regexp.MustCompile("[\\[\\\\^\\$\\.\\|\\?\\*\\+\\(\\)\\{\\}]")
	sepB := []byte(sep)
	escapedSepB := re.ReplaceAllFunc(sepB, func(src []byte) []byte {
		return []byte("\\" + string(src))
	})
	escapedSep := string(escapedSepB)
	return escapedSep
}

func ReIsFullMatch(re *regexp.Regexp, str string) bool {
	foundIndices := re.FindStringIndex(str)
	if foundIndices == nil || len(foundIndices) == 0 {
		return false
	}
	if foundIndices[0] == 0 && foundIndices[1] == len(str) {
		return true
	}
	return false
}

func ReExtractNumbers(str string) []string {
	numbers := numbersRegexp.FindAllString(str, -1)
	return numbers
}

func ReSplitDateTime(dateTimeStr string) (*string, *string, error) {
	if !ReIsFullMatch(DateTimeRegexp, dateTimeStr) {
		return nil, nil, fmt.Errorf("string \"%s\" is not a valid dateTime string", dateTimeStr)
	}
	dateTimeParts := dateTimeSeparatorRegexp.Split(dateTimeStr, -1)
	if len(dateTimeParts) < 1 {
		return nil, nil, fmt.Errorf("dateTime string \"%s\" cannot be splitted", dateTimeStr)
	}
	if len(dateTimeParts) > 2 {
		return nil, nil, fmt.Errorf("dateTime string \"%s\" consists of %d parts, while 2 parts were expected", dateTimeStr, len(dateTimeParts))
	}
	if len(dateTimeParts) == 1 {
		return &dateTimeParts[0], nil, nil
	}
	return &dateTimeParts[0], &dateTimeParts[1], nil
}
