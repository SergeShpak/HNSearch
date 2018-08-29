package query_handler

import (
	"fmt"
	"time"

	"github.com/SergeyShpak/HNSearch/config"
)

type QueryHandler interface {
	CountQueries(from *time.Time, to *time.Time) int
	GetTopQueries(from *time.Time, to *time.Time, size int) int
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
