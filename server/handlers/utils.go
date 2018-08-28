package handlers

import (
	"fmt"
	"net/http"
)

func getParameter(param string, r *http.Request) ([]string, error) {
	results, ok := r.URL.Query()[param]
	if !ok || len(results) == 0 {
		return nil, fmt.Errorf("parameter \"%s\" was not found", param)
	}
	return results, nil
}
