package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/SergeyShpak/HNSearch/config"
	"github.com/SergeyShpak/HNSearch/server/model/query_handler"
)

var s *http.Server

func TestMain(m *testing.M) {
	var err error
	s, err = initServer()
	if err != nil {
		log.Printf("error occurred during tests setup: %v", err)
		os.Exit(-1)
	}
	os.Exit(m.Run())
}

func TestDateDistinctHandler(t *testing.T) {
	testCases := []struct {
		date     string
		expected *query_handler.DistinctQueriesCount
	}{
		{
			date:     "2015",
			expected: &query_handler.DistinctQueriesCount{Count: 573697},
		},
		{
			date:     "2015-08",
			expected: &query_handler.DistinctQueriesCount{Count: 573697},
		},
		{
			date:     "2015-08-03",
			expected: &query_handler.DistinctQueriesCount{Count: 198117},
		},
		{
			date:     "2015-08-01 00:04",
			expected: &query_handler.DistinctQueriesCount{Count: 617},
		},
	}
	for i, tc := range testCases {
		i := i
		tc := tc
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			t.Parallel()
			reqURL := fmt.Sprintf("/1/queries/count/%s", tc.date)
			req, err := http.NewRequest("GET", reqURL, nil)
			if err != nil {
				t.Fatalf("test #%d: could not create a test request: %v", i, err)
			}
			rec := httptest.NewRecorder()
			s.Handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("test #%d: expected HTTP status code: %d, actual code: %d", i, http.StatusOK, rec.Code)
			}
			var actual *query_handler.DistinctQueriesCount
			if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
				t.Errorf("test #%d: cannot unmarshal the response %v into *query_handler.DistinctQueriesCount", i, rec.Body.String())
			}
			if actual.Count != tc.expected.Count {
				t.Errorf("test #%d: expected count: %v, actual count: %v", i, tc.expected.Count, actual.Count)
			}
		})
	}
}

func TestDatePopularHandler(t *testing.T) {
	testCases := []struct {
		date     string
		size     string
		expected []*query_handler.QueryCount
	}{
		{
			date: "2015",
			size: "3",
			expected: []*query_handler.QueryCount{
				&query_handler.QueryCount{
					Query: "http%3A%2F%2Fwww.getsidekick.com%2Fblog%2Fbody-language-advice",
					Count: 6675,
				},
				&query_handler.QueryCount{
					Query: "http%3A%2F%2Fwebboard.yenta4.com%2Ftopic%2F568045",
					Count: 4652,
				},
				&query_handler.QueryCount{
					Query: "http%3A%2F%2Fwebboard.yenta4.com%2Ftopic%2F379035%3Fsort%3D1",
					Count: 3100,
				},
			},
		},
		{
			date: "2015-08-02",
			size: "5",
			expected: []*query_handler.QueryCount{
				&query_handler.QueryCount{
					Query: "http%3A%2F%2Fwww.getsidekick.com%2Fblog%2Fbody-language-advice",
					Count: 2283,
				},
				&query_handler.QueryCount{
					Query: "http%3A%2F%2Fwebboard.yenta4.com%2Ftopic%2F568045",
					Count: 1943,
				},
				&query_handler.QueryCount{
					Query: "http%3A%2F%2Fwebboard.yenta4.com%2Ftopic%2F379035%3Fsort%3D1",
					Count: 1358,
				},
				&query_handler.QueryCount{
					Query: "http%3A%2F%2Fjamonkey.com%2F50-organizing-ideas-for-every-room-in-your-house%2F",
					Count: 890,
				},
				&query_handler.QueryCount{
					Query: "http%3A%2F%2Fsharingis.cool%2F1000-musicians-played-foo-fighters-learn-to-fly-and-it-was-epic",
					Count: 701,
				},
			},
		},
	}
	for i, tc := range testCases {
		i := i
		tc := tc
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			t.Parallel()
			reqURL := fmt.Sprintf("/1/queries/count/%s?size=%s", tc.date, tc.size)
			req, err := http.NewRequest("GET", reqURL, nil)
			if err != nil {
				t.Fatalf("test #%d: could not create a test request: %v", i, err)
			}
			rec := httptest.NewRecorder()
			s.Handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("test #%d: expected HTTP status code: %d, actual code: %d", i, http.StatusOK, rec.Code)
			}
			var actual []*query_handler.QueryCount
			if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
				t.Errorf("test #%d: cannot unmarshal the response %v into *[]query_handler.QueryCount", i, rec.Body.String())
			}
			if len(actual) != len(tc.expected) {
				t.Errorf("test #%d: expected result lenght: %v, actual length: %v, actual result: %v", i, len(tc.expected), len(actual), actual)
			}
			for j, a := range actual {
				if tc.expected[j].Count != a.Count {
					t.Errorf("test #%d, element: %d: expected count: %v, actual count: %v, actual result: %v", i, j, tc.expected[i].Count, a.Count, actual)
				}
				if tc.expected[j].Query != a.Query {
					t.Errorf("test #%d, element: %d: expected query: %v, actual query: %v, actual result: %v", i, j, tc.expected[i].Query, a.Query, actual)
				}
			}
		})
	}
}

func initServer() (*http.Server, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	c := config.GetDefaultConfig()
	c.QueryHandler.File = "../hn_logs.tsv"
	s, err := InitServer(c)
	if err != nil {
		return nil, fmt.Errorf("could not initialize the server: %v", err)
	}
	return s, nil
}
