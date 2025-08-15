package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"
)

type Client[T any] struct {
	Endpoint string
	httpc    *http.Client
}

type gqlReq struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type gqlResp[T any] struct {
	Data   T             `json:"data"`
	Errors []interface{} `json:"errors"`
}

type SubscriptionClient struct {
	Endpoint string
	conn     *websocket.Conn
}

func NewClient[T any](endpoint string) *Client[T] {
	return &Client[T]{
		Endpoint: endpoint,
		httpc:    &http.Client{Timeout: 20 * time.Second},
	}
}

func NewSubscriptionClient(endpoint string) *SubscriptionClient {
	return &SubscriptionClient{
		Endpoint: endpoint,
	}
}

func (c *Client[T]) Do(ctx context.Context, query string, vars map[string]interface{}, out *T) error {
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

	// http code check
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("http %d from %s: %s", res.StatusCode, c.Endpoint, sample(raw, 600))
	}

	// content type check
	ct := res.Header.Get("Content-Type")
	mt, _, _ := mime.ParseMediaType(ct)
	if mt != "" && mt != "application/json" && mt != "application/graphql-response+json" {
		return fmt.Errorf("unexpected content-type %q from %s: %s", ct, c.Endpoint, sample(raw, 600))
	}

	// json decoding
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

func (sc *SubscriptionClient) Connect(ctx context.Context) error {
	if sc.conn != nil {
		return nil
	}

	// open websocket
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, sc.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	sc.conn = conn
	return nil
}

func (sc *SubscriptionClient) Subscribe(ctx context.Context, query string, vars map[string]interface{}, handler func(BlocksData) error) error {
	if err := sc.Connect(ctx); err != nil {
		return fmt.Errorf("connect websocket: %w", err)
	}

	// subscription request
	subReq := gqlReq{
		Query:     query,
		Variables: vars,
	}

	if err := sc.conn.WriteJSON(subReq); err != nil {
		return fmt.Errorf("write subscription request: %w", err)
	}

	// subscription response handling
	go func() {
		defer sc.conn.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				var response gqlResp[BlocksData]
				if err := sc.conn.ReadJSON(&response); err != nil {
					log.Printf("read subscription response: %v", err)
					return
				}

				if len(response.Errors) > 0 {
					log.Printf("subscription errors: %+v", response.Errors)
					continue
				}

				// handler
				if err := handler(response.Data); err != nil {
					log.Printf("handle subscription response: %v", err)
				}
			}
		}
	}()

	return nil
}

func (sc *SubscriptionClient) SubcribeOnce(ctx context.Context, query string, vars map[string]interface{}, handler func(BlocksData) error) error {
	if err := sc.Connect(ctx); err != nil {
		return fmt.Errorf("connect websocket: %w", err)
	}

	subReq := gqlReq{
		Query:     query,
		Variables: vars,
	}

	if err := sc.conn.WriteJSON(subReq); err != nil {
		return fmt.Errorf("write subscription request: %w", err)
	}

	// once response
	var response gqlResp[BlocksData]
	if err := sc.conn.ReadJSON(&response); err != nil {
		return fmt.Errorf("read subscription response: %w", err)
	}

	if len(response.Errors) > 0 {
		return fmt.Errorf("subscription errors: %+v", response.Errors)
	}

	if err := handler(response.Data); err != nil {
		log.Printf("handle subscription response: %v", err)
	}

	sc.conn.Close()
	sc.conn = nil

	return nil
}

func (sc *SubscriptionClient) Close() error {
	if sc.conn != nil {
		return sc.conn.Close()
	}
	return nil
}

func sample(b []byte, n int) string {
	s := strings.TrimSpace(string(b))
	if len(s) > n {
		return s[:n] + "...(truncated)"
	}
	return s
}
