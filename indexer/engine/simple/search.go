package simple

import (
	"time"
)

type indexesPath struct {
	Path    []int
	Indexes []int
}

func newIndexesPath() *indexesPath {
	path := &indexesPath{
		Indexes: make([]int, 0),
	}
	return path
}

func (p *indexesPath) SetPath(parts ...int) {
	p.Path = make([]int, 0, len(parts))
	for _, part := range parts {
		p.Path = append(p.Path, part)
	}
}

func (p *indexesPath) AddIndex(i int) {
	p.Indexes = append(p.Indexes, i)
}

type minutesSecondsIndexesPaths indexesPath

type indexesToExamine struct {
	Years   []indexesPath
	Months  []indexesPath
	Days    []indexesPath
	Hours   []indexesPath
	Minutes []minutesSecondsIndexesPaths
	Seconds []minutesSecondsIndexesPaths
}

type addDateUnit struct {
	Years  int
	Months int
	Days   int
}

type datesDiff struct {
	From *time.Time
	To   *time.Time
	Diff [6]int
}

func newDatesDiff(from *time.Time, to *time.Time) *datesDiff {
	diffDate := &datesDiff{
		From: from,
		To:   to,
	}
	diffDate.Diff[0] = to.Year() - from.Year()
	diffDate.Diff[1] = int(to.Month()) - int(from.Month())
	diffDate.Diff[2] = to.Day() - from.Day()
	diffDate.Diff[3] = to.Hour() - from.Hour()
	diffDate.Diff[4] = to.Minute() - from.Minute()
	diffDate.Diff[5] = to.Second() - from.Second()
	return diffDate
}

//TODO: refactor
func (diff *datesDiff) getIndexesPaths() ([]*indexesPath, []*minutesSecondsIndexesPaths) {
	paths := make([]*indexesPath, 0)
	minSecPaths := make([]*minutesSecondsIndexesPaths, 0)
	ceiled := *diff.From

	ceilTimeStepFn := func(ceil time.Time, path []int, ceiledUnit int, maxCeiledUnit int, toUnit int, interval time.Duration) *indexesPath {
		idxPath := newIndexesPath()
		idxPath.SetPath(path...)
		if diff.To.Before(ceil) || diff.To.Equal(ceil) {
			for i := ceiledUnit; i < toUnit; i++ {
				idxPath.AddIndex(i)
			}
			ceiled = ceiled.Add(interval * time.Duration((toUnit - ceiledUnit)))
			if len(idxPath.Indexes) == 0 {
				return nil
			}
			return idxPath
		}
		for i := ceiledUnit; i <= maxCeiledUnit; i++ {
			idxPath.AddIndex(i)
		}
		ceiled = ceiled.Add(interval * time.Duration((maxCeiledUnit - ceiledUnit + 1)))
		if len(idxPath.Indexes) == 0 {
			return nil
		}
		return idxPath
	}

	ceilDateStepFn := func(ceil time.Time, path []int, ceiledUnit int, maxCeiledUnit int, toUnit int, intervalUnit *addDateUnit) *indexesPath {
		idxPath := newIndexesPath()
		idxPath.SetPath(path...)
		if diff.To.Before(ceil) || diff.To.Equal(ceil) {
			for i := ceiledUnit; i < toUnit; i++ {
				idxPath.AddIndex(i)
				ceiled = ceiled.AddDate(intervalUnit.Years, intervalUnit.Months, intervalUnit.Days)
			}
			if len(idxPath.Indexes) == 0 {
				return nil
			}
			return idxPath
		}
		for i := ceiledUnit; i <= maxCeiledUnit; i++ {
			idxPath.AddIndex(i)
			ceiled = ceiled.AddDate(intervalUnit.Years, intervalUnit.Months, intervalUnit.Days)
		}
		if len(idxPath.Indexes) == 0 {
			return nil
		}
		return idxPath
	}

	// ceiling seconds
	if ceiled.Second() != 0 {
		ceil := time.Date(
			ceiled.Year(),
			ceiled.Month(),
			ceiled.Day(),
			ceiled.Hour(),
			ceiled.Minute()+1,
			0,
			0,
			diff.From.Location(),
		)
		indexes := ceilTimeStepFn(
			ceil,
			[]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day(), ceiled.Hour(), ceiled.Minute()},
			ceiled.Second(),
			59,
			diff.To.Second(),
			time.Second,
		)
		if indexes != nil {
			minSecIndexes := minutesSecondsIndexesPaths(*indexes)
			minSecPaths = append(minSecPaths, &minSecIndexes)
		}
	}
	// ceiling minutes
	if ceiled.Minute() != 0 {
		ceil := time.Date(
			ceiled.Year(),
			ceiled.Month(),
			ceiled.Day(),
			ceiled.Hour()+1,
			0,
			0,
			0,
			diff.From.Location(),
		)
		indexes := ceilTimeStepFn(
			ceil,
			[]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day(), ceiled.Hour()},
			ceiled.Minute(),
			59,
			diff.To.Minute(),
			time.Minute,
		)
		if indexes != nil {
			minSecIndexes := minutesSecondsIndexesPaths(*indexes)
			minSecPaths = append(minSecPaths, &minSecIndexes)
		}

	}
	// ceiling hours
	if ceiled.Hour() != 0 {
		ceil := time.Date(
			ceiled.Year(),
			ceiled.Month(),
			ceiled.Day()+1,
			0,
			0,
			0,
			0,
			ceiled.Location(),
		)
		indexes := ceilTimeStepFn(
			ceil,
			[]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day()},
			ceiled.Hour(),
			23,
			diff.To.Hour(),
			time.Hour,
		)
		if indexes != nil {
			paths = append(paths, indexes)
		}
	}
	if ceiled.Day() != 0 {
		ceil := time.Date(ceiled.Year(), ceiled.Month()+1, 0, 0, 0, 0, 0, ceiled.Location())
		interval := &addDateUnit{
			Days: 1,
		}
		indexes := ceilDateStepFn(ceil, []int{ceiled.Year(), int(ceiled.Month())}, ceiled.Day(), 31, diff.To.Day(), interval)
		if indexes != nil {
			paths = append(paths, indexes)
		}
	}
	if int(ceiled.Month()) != 0 {
		ceil := time.Date(ceiled.Year()+1, ceiled.Month(), 0, 0, 0, 0, 0, ceiled.Location())
		interval := &addDateUnit{
			Months: 1,
		}
		indexes := ceilDateStepFn(ceil, []int{ceiled.Year()}, int(ceiled.Month()), 11, int(diff.To.Month()), interval)
		if indexes != nil {
			paths = append(paths, indexes)
		}
	}
	getPathsFromCeiled := func(parts []int, ceiledUnit int, toUnit int) (*indexesPath, int) {
		p := newIndexesPath()
		p.SetPath(parts...)
		ctr := 0
		for i := ceiledUnit; i < toUnit; i++ {
			p.AddIndex(i)
			ctr++
		}
		if ctr == 0 {
			return nil, 0
		}
		return p, ctr
	}
	p, yearsAdded := getPathsFromCeiled(nil, ceiled.Year(), diff.To.Year())
	if p != nil {
		paths = append(paths, p)
		ceiled = ceiled.AddDate(yearsAdded, 0, 0)
	}
	p, monthsAdded := getPathsFromCeiled([]int{ceiled.Year()}, int(ceiled.Month()), int(diff.To.Month()))
	if p != nil {
		paths = append(paths, p)
		ceiled = ceiled.AddDate(0, monthsAdded, 0)
	}
	p, daysAdded := getPathsFromCeiled([]int{ceiled.Year(), int(ceiled.Month())}, ceiled.Day(), diff.To.Day())
	if p != nil {
		paths = append(paths, p)
		ceiled = ceiled.AddDate(0, 0, daysAdded)
	}
	p, hoursAdded := getPathsFromCeiled([]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day()}, ceiled.Hour(), diff.To.Hour())
	if p != nil {
		paths = append(paths, p)
		ceiled = ceiled.Add(time.Hour * time.Duration(hoursAdded))
	}
	p, minutesAdded := getPathsFromCeiled([]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day(), ceiled.Hour()}, ceiled.Minute(), diff.To.Minute())
	if p != nil {
		minP := minutesSecondsIndexesPaths(*p)
		minSecPaths = append(minSecPaths, &minP)
		ceiled = ceiled.Add(time.Minute * time.Duration(minutesAdded))
	}
	p, secondsAdded := getPathsFromCeiled([]int{ceiled.Year(), int(ceiled.Month()), ceiled.Day(), ceiled.Hour(), ceiled.Minute()}, ceiled.Second(), diff.To.Second())
	if p != nil {
		secP := minutesSecondsIndexesPaths(*p)
		minSecPaths = append(minSecPaths, &secP)
		ceiled = ceiled.Add(time.Second * time.Duration(secondsAdded))
	}
	return paths, minSecPaths
}

func (indexer *simpleIndexer) CountDistinctQueries(from *time.Time, to *time.Time) int {
	return 0
}

func (indexer *simpleIndexer) getIndexesToExamine(from *time.Time, to *time.Time) *indexesToExamine {
	if to.Before(*from) {
		from, to = to, from
	}
	toToInclude := time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), to.Minute(), to.Second()-1, 0, to.Location())
	_ = newDatesDiff(from, &toToInclude)
	for i := 0; i < 6; i++ {
	}

	return nil
}
