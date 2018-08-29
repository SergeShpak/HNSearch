package query_handler

import (
	"fmt"
	"time"

	"github.com/SergeyShpak/HNSearch/config"
)

type QueryHandler interface {
	CountDistinctQueries(from *time.Time, to *time.Time) *DistinctQueriesCount
	GetTopQueries(from *time.Time, to *time.Time, size int) []*QueryCount
}

type DistinctQueriesCount struct {
	Count int
}

type QueryCount struct {
	Query string
	Count int
}

func GetQueryHandler(qh *config.QueryHandler) (QueryHandler, error) {
	if qh == nil {
		return nil, fmt.Errorf("passed QueryHandler configuration is nil")
	}
	var err error
	var handler QueryHandler
	switch qh.Type {
	case "QueryDump":
		handler, err = initSimpleQueryHandler(qh)
	default:
		handler, err = initSimpleQueryHandler(qh)
	}
	return handler, err
}

func initSimpleQueryHandler(qh *config.QueryHandler) (QueryHandler, error) {
	handler, err := newSimleQueryHanlder(qh.File)
	return handler, err
}
