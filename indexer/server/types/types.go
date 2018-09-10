package types

type dateIntervalQuery struct {
	FromDate int64
	ToDate   int64
}

type DistinctQueriesCountRequest dateIntervalQuery

type TopQueriesRequest dateIntervalQuery

type DistinctQueriesCountResponse struct {
	Count int
}

type Query struct {
	Query string
	Count int
}

type TopQueriesResponse struct {
	Queries []*Query
}
