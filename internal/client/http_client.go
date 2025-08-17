package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

// GraphQLClient handles GraphQL HTTP requests
type GraphQLClient[T any] struct {
	Endpoint string
	httpc    *http.Client
}

// gqlReq GraphQL request structure
type gqlReq struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// gqlResp GraphQL response structure
type gqlResp[T any] struct {
	Data   T             `json:"data"`
	Errors []interface{} `json:"errors"`
}

// NewGraphQLClient creates a new GraphQL client
func NewGraphQLClient[T any](endpoint string) *GraphQLClient[T] {
	return &GraphQLClient[T]{
		Endpoint: endpoint,
		httpc:    &http.Client{Timeout: 20 * time.Second},
	}
}

// Do executes a GraphQL query
func (c *GraphQLClient[T]) Do(ctx context.Context, query string, vars map[string]interface{}, out *T) error {
	if out == nil {
		return errors.New("out is nil")
	}

	body, err := json.Marshal(gqlReq{Query: query, Variables: vars})
	if err != nil {
		return fmt.Errorf("marshal gql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.httpc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// check HTTP status code
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("http %d from %s: %s", res.StatusCode, c.Endpoint, sample(raw, 600))
	}

	// check content type
	ct := res.Header.Get("Content-Type")
	mt, _, _ := mime.ParseMediaType(ct)
	if mt != "" && mt != "application/json" && mt != "application/graphql-response+json" {
		return fmt.Errorf("unexpected content-type %q from %s: %s", ct, c.Endpoint, sample(raw, 600))
	}

	// decode JSON response
	var r gqlResp[T]
	if err := json.Unmarshal(raw, &r); err != nil {
		return fmt.Errorf("decode json: %w; body: %s", err, sample(raw, 600))
	}
	if len(r.Errors) > 0 {
		return fmt.Errorf("graphql errors: %+v", r.Errors)
	}
	*out = r.Data
	return nil
}

// sample helper function for sampling response data
func sample(b []byte, n int) string {
	s := strings.TrimSpace(string(b))
	if len(s) > n {
		return s[:n] + "...(truncated)"
	}
	return s
}
