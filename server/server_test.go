package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/SergeyShpak/HNSearch/server/config"
	"github.com/SergeyShpak/HNSearch/server/initialization"
	"github.com/SergeyShpak/HNSearch/server/types"
)

var s *http.Server

func TestMain(m *testing.M) {
	var err error
	s, err = initTestServer()
	if err != nil {
		log.Printf("error occurred during tests setup: %v", err)
		os.Exit(-1)
	}
	os.Exit(m.Run())
}

func TestDateDistinctHandler(t *testing.T) {
	testCases := []struct {
		date     string
		expected *types.DistinctQueriesCountResponse
	}{
		{
			date:     "2015",
			expected: &types.DistinctQueriesCountResponse{Count: 573697},
		},
		{
			date:     "2015-08",
			expected: &types.DistinctQueriesCountResponse{Count: 573697},
		},
		{
			date:     "2015-08-03",
			expected: &types.DistinctQueriesCountResponse{Count: 198117},
		},
		{
			date:     "2015-08-01 00:04",
			expected: &types.DistinctQueriesCountResponse{Count: 617},
		},
	}
	for i, tc := range testCases {
		i := i
		tc := tc
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			t.Parallel()
			reqURL := fmt.Sprintf("http://127.0.0.1:8080/1/queries/count/%s", tc.date)
			resp, err := http.Get(reqURL)
			if err != nil {
				t.Fatalf("test #%d: could not send a request: %v", i, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("test #%d: expected HTTP status code: %d, actual code: %d", i, http.StatusOK, resp.StatusCode)
			}
			body, err := ioutil.ReadAll(resp.Body)
			var actual *types.DistinctQueriesCountResponse
			if err := json.Unmarshal(body, &actual); err != nil {
				t.Errorf("test #%d: cannot unmarshal the response %v into *types.DistinctQueriesCountResponse", i, string(body))
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
		expected []*types.QueryCount
	}{
		{
			date: "2015",
			size: "3",
			expected: []*types.QueryCount{
				&types.QueryCount{
					Query: "http%3A%2F%2Fwww.getsidekick.com%2Fblog%2Fbody-language-advice",
					Count: 6675,
				},
				&types.QueryCount{
					Query: "http%3A%2F%2Fwebboard.yenta4.com%2Ftopic%2F568045",
					Count: 4652,
				},
				&types.QueryCount{
					Query: "http%3A%2F%2Fwebboard.yenta4.com%2Ftopic%2F379035%3Fsort%3D1",
					Count: 3100,
				},
			},
		},
		{
			date: "2015-08-02",
			size: "5",
			expected: []*types.QueryCount{
				&types.QueryCount{
					Query: "http%3A%2F%2Fwww.getsidekick.com%2Fblog%2Fbody-language-advice",
					Count: 2283,
				},
				&types.QueryCount{
					Query: "http%3A%2F%2Fwebboard.yenta4.com%2Ftopic%2F568045",
					Count: 1943,
				},
				&types.QueryCount{
					Query: "http%3A%2F%2Fwebboard.yenta4.com%2Ftopic%2F379035%3Fsort%3D1",
					Count: 1358,
				},
				&types.QueryCount{
					Query: "http%3A%2F%2Fjamonkey.com%2F50-organizing-ideas-for-every-room-in-your-house%2F",
					Count: 890,
				},
				&types.QueryCount{
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
			reqURL := fmt.Sprintf("http://127.0.0.1:8080/1/queries/popular/%s?size=%s", tc.date, tc.size)
			resp, err := http.Get(reqURL)
			if err != nil {
				t.Fatalf("test #%d: could not send a request: %v", i, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("test #%d: expected HTTP status code: %d, actual code: %d", i, http.StatusOK, resp.StatusCode)
			}
			body, err := ioutil.ReadAll(resp.Body)
			var actual *types.TopQueriesResponse
			if err := json.Unmarshal(body, &actual); err != nil {
				t.Fatalf("test #%d: cannot unmarshal the response %v into *types.TopQueriesResponse", i, string(body))
			}
			if len(actual.Queries) != len(tc.expected) {
				t.Fatalf("test #%d: expected result lenght: %v, actual length: %v, actual result: %v", i, len(tc.expected), len(actual.Queries), actual)
			}
			for j, a := range actual.Queries {
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

func initTestServer() (*http.Server, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	c := config.GetDefaultConfig()
	s, err := initialization.InitServer(c)
	if err != nil {
		return nil, fmt.Errorf("could not initialize the server: %v", err)
	}
	return s, nil
}
