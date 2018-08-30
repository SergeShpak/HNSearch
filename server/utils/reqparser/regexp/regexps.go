package regexp

import (
	"fmt"
	"regexp"
	"strings"
)

var acceptableDateSeparators = []string{"-", "/", "."}
var acceptableTimeSeparators = []string{":"}

var dateSeparatorsRegexpStr = fmt.Sprintf("(%s)", getSeparatorsString(acceptableDateSeparators))
var dateRegexpStr = fmt.Sprintf("(19|20)[0-9]{2}(%[1]s((1[0-2])|(0?[1-9]))?(%[1]s((3[0-1])|([1-2][0-9])|(0?[1-9]))?)?)?", dateSeparatorsRegexpStr)
var timeSeparatorsRegexpStr = fmt.Sprintf("(%s)", getSeparatorsString(acceptableTimeSeparators))
var timeRegexpStr = fmt.Sprintf("((1[0-9])|(2[0-3])|(0?[0-9]))((%[1]s(([1-5][0-9])|(0?[0-9]))?)(%[1]s(([1-5][0-9])|(0?[0-9]))?)?)?", timeSeparatorsRegexpStr)
var dateTimeSeparatorRegexpStr = " "
var dateTimeRegexpStr = fmt.Sprintf("(%s)((%s)(%s))?", dateRegexpStr, dateTimeSeparatorRegexpStr, timeRegexpStr)
var numbersRegexpStr = "[0-9]*"

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
