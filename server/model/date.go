package model

import (
	"fmt"
	"time"
)

func GetDistinctQueries(qd *QueryDump, from *time.Time, to *time.Time) int {
	fmt.Println("Query dump: ", qd)
	return 0
}
