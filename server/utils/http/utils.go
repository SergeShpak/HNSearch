package http

import (
	"net/http"
)

func GetParameter(param string, r *http.Request) []string {
	results := r.URL.Query()[param]
	return results
}
