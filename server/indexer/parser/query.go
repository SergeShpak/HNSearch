package parser

import "time"

type Query struct {
	Date  *time.Time
	Query string
}

type QueryBatch interface {
	From() *time.Time
	To() *time.Time
	Queries() []Query
}

type SimpleQueryBatch struct {
	From    *time.Time
	To      *time.Time
	Queries []Query
}
