package queries

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	ctxIDs "github.com/SergeyShpak/HNSearch/indexer/server/context"
	"github.com/SergeyShpak/HNSearch/indexer/server/types"
)

func ParseQueriesByDateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := getQueriesByDateRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		ctx := context.WithValue(r.Context(), ctxIDs.QueryByDateRequestID, req)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getQueriesByDateRequest(r *http.Request) (*types.DistinctQueriesCountRequest, error) {
	req := &types.DistinctQueriesCountRequest{}
	b, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(b, req); err != nil {
		return nil, err
	}
	return req, nil
}
