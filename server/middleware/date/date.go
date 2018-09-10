package date

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	ctxIDs "github.com/SergeyShpak/HNSearch/server/context"
	"github.com/SergeyShpak/HNSearch/server/types"
	http_utils "github.com/SergeyShpak/HNSearch/server/utils/http"
	"github.com/SergeyShpak/HNSearch/server/utils/reqparser"
)

func ParseDateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestParser, ok := r.Context().Value(ctxIDs.RequestParserID).(reqparser.Parser)
		if !ok || requestParser == nil {
			msg := "no request parser found"
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(msg))
			return
		}
		var err error
		r, err = storeDateInCtx(r, requestParser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		r, err = storeSizeInCtx(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func storeDateInCtx(r *http.Request, reqParser reqparser.Parser) (*http.Request, error) {
	dateStr, err := extractDate(r)
	if err != nil {
		return nil, err
	}
	dateTimeParts, err := reqParser.DateSplitDateTime(dateStr)
	if err != nil {
		return nil, err
	}
	if len(dateTimeParts) < 1 || len(dateTimeParts) > 2 {
		err := fmt.Errorf("dateTimeParts should cotain either one or two elements, it actually contains %d: %v", len(dateTimeParts), dateTimeParts)
		return nil, err
	}
	dateParts := reqParser.ExtractNumbers(dateTimeParts[0])
	date, err := types.NewDate(dateParts)
	if err != nil {
		return nil, err
	}
	timeParts := reqParser.ExtractNumbers(dateTimeParts[1])
	time, err := types.NewTime(timeParts)
	if err != nil {
		return nil, err
	}
	ctx := context.WithValue(r.Context(), ctxIDs.DateParamID, date)
	ctx = context.WithValue(ctx, ctxIDs.TimeParamID, time)
	rWithCtx := r.WithContext(ctx)
	return rWithCtx, nil
}

func extractDate(r *http.Request) (string, error) {
	dateStr, ok := mux.Vars(r)["date"]
	if !ok {
		return "", fmt.Errorf("No date was passed")
	}
	return dateStr, nil
}

func storeSizeInCtx(r *http.Request) (*http.Request, error) {
	sizeStr, err := extractSize(r)
	if err != nil {
		return nil, err
	}
	size, err := parseSize(sizeStr)
	if err != nil {
		return nil, err
	}
	ctx := context.WithValue(r.Context(), ctxIDs.SizeParamID, size)
	rWithCtx := r.WithContext(ctx)
	return rWithCtx, nil
}

func extractSize(r *http.Request) (string, error) {
	sizeParams := http_utils.GetParameter("size", r)
	if len(sizeParams) == 0 {
		return "0", nil
	}
	if len(sizeParams) > 1 {
		return "-1", fmt.Errorf("multiple size parameters passed")
	}
	return sizeParams[0], nil
}

func parseSize(sizeStr string) (int, error) {
	s, err := strconv.Atoi(sizeStr)
	if err != nil {
		return s, err
	}
	if s < 0 {
		return s, fmt.Errorf("size %d is invalid", s)
	}
	return s, nil
}
