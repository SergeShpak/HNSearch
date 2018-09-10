package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/SergeyShpak/HNSearch/server/config"
	"github.com/SergeyShpak/HNSearch/server/indexer/http/types"
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
	fmt.Println("From: %v, To: %v", reqBody.FromDate, reqBody.ToDate)
	resp := &types.DistinctQueriesCountResponse{}
	if err := indexer.postJSON(reqBody, "/1/queries/count", resp); err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (indexer *HTTPIndexer) postJSON(v interface{}, url string, resp interface{}) error {
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
