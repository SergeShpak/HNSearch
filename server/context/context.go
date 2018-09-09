package context

type Key int

const (
	IndexerID Key = iota
	RequestParserID
	DateParamID
	TimeParamID
	SizeParamID
)
