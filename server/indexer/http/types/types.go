package types

type QueriesByDateRequest struct {
	FromDate int64
	ToDate   int64
}

type DistinctQueriesCountResponse struct {
	Count int
}
