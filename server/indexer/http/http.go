package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/SergeyShpak/HNSearch/server/config"
	"github.com/SergeyShpak/HNSearch/server/indexer/http/types"
	serverTypes "github.com/SergeyShpak/HNSearch/server/types"
)

type HTTPIndexer struct {
	addr string
}

func NewHTTPIndexer(c *config.Config) (*HTTPIndexer, error) {
	if c == nil || c.Indexer == nil || c.Indexer.HTTP == nil {
		return nil, fmt.Errorf("no valid HTTPIndexer configuration was found")
	}
	addr := c.Indexer.HTTP.Addr
	if addr[len(addr)-1] == '/' {
		addr = addr[:len(addr)-1]
	}
	indexer := &HTTPIndexer{
		addr: addr,
	}
	return indexer, nil
}

func (indexer *HTTPIndexer) CountDistinctQueries(from *time.Time, to *time.Time) (int, error) {
	reqBody := &types.DistinctQueriesCountRequest{
		FromDate: from.Unix(),
		ToDate:   to.Unix(),
	}
	resp := &types.DistinctQueriesCountResponse{}
	if err := indexer.postJSON(reqBody, "/1/queries/count", resp, nil); err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (indexer *HTTPIndexer) GetTopQueries(from *time.Time, to *time.Time, size int) (*serverTypes.TopQueriesResponse, error) {
	reqBody := &types.TopQueriesRequest{
		FromDate: from.Unix(),
		ToDate:   to.Unix(),
	}
	resp := &types.TopQueriesResponse{}
	headers := make(map[string]string)
	headers["x-top-size"] = strconv.Itoa(size)
	if err := indexer.postJSON(reqBody, "/1/queries/top", resp, headers); err != nil {
		return nil, err
	}
	serverResp := &serverTypes.TopQueriesResponse{
		Queries: make([]*serverTypes.QueryCount, len(resp.Queries)),
	}
	for i := 0; i < len(resp.Queries); i++ {
		queryCount := serverTypes.QueryCount(*resp.Queries[i])
		serverResp.Queries[i] = &queryCount
	}
	return serverResp, nil
}

func (indexer *HTTPIndexer) postJSON(v interface{}, url string, resp interface{}, headers map[string]string) error {
	reqBodyB, err := json.Marshal(v)
	if err != nil {
		return err
	}
	url = indexer.addr + url
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyB))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for header, val := range headers {
		req.Header.Set(header, val)
	}

	client := &http.Client{}
	httpResp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBodyB, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response code: %v, response body: %v", httpResp.StatusCode, string(respBodyB))
	}
	if err := indexer.unmarshalResponse(httpResp, respBodyB, resp); err != nil {
		return err
	}
	return nil
}

func (indexer *HTTPIndexer) unmarshalResponse(httpResponse *http.Response, respBodyB []byte, resp interface{}) error {
	contentType := httpResponse.Header.Get("Content-Type")
	switch contentType {
	case "application/json":
		if err := json.Unmarshal(respBodyB, resp); err != nil {
			return err
		}
	default:
		if len(contentType) == 0 {
			fmt.Println(string(respBodyB))
			return fmt.Errorf("Content-Type header was not found in the response")
		}
		return fmt.Errorf("could not unmarshal %v Content-Type response", contentType)
	}
	return nil
}
