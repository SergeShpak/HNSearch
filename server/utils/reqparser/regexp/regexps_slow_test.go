//+build regexps_slow_test

package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
)

func TestDateRegexp(t *testing.T) {
	dateRegexp, err := regexp.Compile(dateRegexpStr)
	if err != nil {
		t.Fatalf("failed while compiling the date regexp: %s", dateRegexpStr)
	}
	t.Parallel()
	for year := 1900; year < 2100; year++ {
		yearStrs := getYearStrs(year)
		fmt.Println("single year strs: ", yearStrs)
		for _, yearStr := range yearStrs {
			yearStr := yearStr
			t.Run(fmt.Sprintf("testing a single year %s", yearStr), func(t *testing.T) {
				fmt.Println("testing", yearStr)
				if !ReIsFullMatch(dateRegexp, yearStr) {
					t.Fatalf("\"%s\" did not match the date regexp", yearStr)
				}
			})
		}
		for month := 1; month <= 12; month++ {
			monthStrs := getMonthStrs(month, yearStrs)
			for _, monthStr := range monthStrs {
				monthStr := monthStr
				t.Run(fmt.Sprintf("testing a single month %s", monthStr), func(t *testing.T) {
					fmt.Println("testing", monthStr)
					if !ReIsFullMatch(dateRegexp, monthStr) {
						t.Fatalf("\"%s\" did not match the date regexp", monthStr)
					}
				})
			}
			for day := 1; day <= 31; day++ {
				dayStrs := getDayStrs(day, monthStrs)
				for _, dayStr := range dayStrs {
					dayStr := dayStr
					t.Run(fmt.Sprintf("testing a single day %s", dayStr), func(t *testing.T) {
						fmt.Println("testing", dayStr)
						if !ReIsFullMatch(dateRegexp, dayStr) {
							t.Fatalf("\"%s\" did not match the date regexp", dayStr)
						}
					})
				}
			}
		}
	}
}

func getYearStrs(year int) []string {
	strs := make([]string, 0)
	yearStr := strconv.Itoa(year)
	strs = append(strs, yearStr)
	strs = addStringsWithTrailingSeparators(strs, AcceptableDateSeparators)
	return strs
}

func getMonthStrs(month int, yearStrs []string) []string {
	return getPartString(month, yearStrs, AcceptableDateSeparators, true)
}

func addStringsWithTrailingSeparators(strs []string, separators []string) []string {
	initStrsCount := len(strs)
	for _, s := range separators {
		for i := 0; i < initStrsCount; i++ {
			strs = append(strs, strs[i]+s)
		}
	}
	return strs
}

func getDayStrs(day int, monthStrs []string) []string {
	return getPartString(day, monthStrs, AcceptableDateSeparators, false)
}

func TestTimeRegexp(t *testing.T) {
	timeRegexp, err := regexp.Compile(timeRegexpStr)
	if err != nil {
		t.Fatalf("failed while compiling the time regexp: %s", timeRegexpStr)
	}
	t.Parallel()
	for hour := 0; hour <= 23; hour++ {
		hourStrs := getHourStrs(hour)
		for _, hourStr := range hourStrs {
			hourStr := hourStr
			t.Run(fmt.Sprintf("testing the hour %s", hourStr), func(t *testing.T) {
				fmt.Println("testing", hourStr)
				if !ReIsFullMatch(timeRegexp, hourStr) {
					t.Fatalf("\"%s\" did not match the time regexp", hourStr)
				}
			})
		}
		for minute := 0; minute <= 59; minute++ {
			minuteStrs := getMinuteStrs(minute, hourStrs)
			for _, minuteStr := range minuteStrs {
				minuteStr := minuteStr
				t.Run(fmt.Sprintf("testing the minute %s", minuteStr), func(t *testing.T) {
					fmt.Println("testing", minuteStr)
					if !ReIsFullMatch(timeRegexp, minuteStr) {
						t.Fatalf("\"%s\" did not match the time regexp", minuteStr)
					}
				})
			}
			for second := 0; second <= 59; second++ {
				secondStrs := getSecondStrs(second, minuteStrs)
				for _, secondStr := range secondStrs {
					secondStr := secondStr
					t.Run(fmt.Sprintf("testing the second %s", secondStr), func(t *testing.T) {
						fmt.Println("testing", secondStr)
						if !ReIsFullMatch(timeRegexp, secondStr) {
							t.Fatalf("\"%s\" did not match the time regexp", secondStr)
						}
					})
				}
			}
		}
	}
}

func getHourStrs(hour int) []string {
	strs := make([]string, 0)
	hourStr := strconv.Itoa(hour)
	strs = append(strs, hourStr)
	if len(hourStr) < 2 {
		strs = append(strs, "0"+hourStr)
	}
	strs = addStringsWithTrailingSeparators(strs, AcceptableTimeSeparators)
	return strs
}

func getMinuteStrs(minute int, hourStrs []string) []string {
	return getPartString(minute, hourStrs, AcceptableTimeSeparators, true)
}

func getSecondStrs(second int, minutesStrs []string) []string {
	return getPartString(second, minutesStrs, AcceptableTimeSeparators, false)
}

func getPartString(unit int, previousParts []string, separators []string, addTrailingSeparator bool) []string {
	strs := make([]string, 0)
	unitStr := strconv.Itoa(unit)
	for _, prev := range previousParts {
		if _, err := strconv.Atoi(string(prev[len(prev)-1])); err == nil {
			continue
		}
		strs = append(strs, prev+unitStr)
		if len(unitStr) < 2 {
			strs = append(strs, prev+"0"+unitStr)
		}
	}
	if addTrailingSeparator {
		strs = addStringsWithTrailingSeparators(strs, separators)
	}
	return strs
}
