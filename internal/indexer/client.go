package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	Endpoint string
	httpc    *http.Client
}

func NewClient(endpoint string) *Client {
	return &Client{
		Endpoint: endpoint,
		httpc:    &http.Client{Timeout: 20 * time.Second},
	}
}

type gqlReq struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Go 1.18 제네릭 기능, 구조체 선언시 타입을 받아서 data의 타입을 사용
type gqlResp[T any] struct {
	Data   T             `json:"data"`
	Errors []interface{} `json:"errors"`
}

func (c *Client) Do[T any](ctx context.Context, query string, vars map[string]interface{}, out *T) error {
	body, _ := json.Marshal(gqlReq{Query: query, Variables: vars})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var r gqlResp[T]
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return err
	}
	if len(r.Errors) > 0 {
		return fmt.Errorf("graphql errors: %+v", r.Errors)
	}
	*out = r.Data
	return nil
}
