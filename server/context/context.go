package context

type Key int

const (
	IndexerID Key = iota
	RequestParserID
	FromDateID
	ToDateID
	SizeParamID
)
