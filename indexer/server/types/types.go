package types

type DistinctQueriesCountRequest DateIntervalQuery

type TopQueriesRequest struct {
	DateInterval *DateIntervalQuery
	Size         int
}

type DateIntervalQuery struct {
	FromDate int64
	ToDate   int64
}

type DistinctQueriesCountResponse struct {
	Count int
}

// TODO: separate web and internal types
type QueryCount struct {
	Query string
	Count int
}
