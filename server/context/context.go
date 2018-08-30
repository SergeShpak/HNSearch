package context

type Key int

const (
	QueryHandlerID Key = iota
	RequestParserID
	DateParamID
	TimeParamID
	SizeParamID
)
