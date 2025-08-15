package indexer

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

// SubscriptionClient handles GraphQL subscriptions via WebSocket
type SubscriptionClient struct {
	Endpoint string
	conn     *websocket.Conn
}

// NewSubscriptionClient creates a new subscription client
func NewSubscriptionClient(endpoint string) *SubscriptionClient {
	return &SubscriptionClient{
		Endpoint: endpoint,
	}
}

// Connect establishes WebSocket connection
func (sc *SubscriptionClient) Connect(ctx context.Context) error {
	if sc.conn != nil {
		return nil
	}

	// open websocket connection
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, sc.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	sc.conn = conn
	return nil
}

// Subscribe starts a persistent subscription
func (sc *SubscriptionClient) Subscribe(ctx context.Context, query string, vars map[string]interface{}, handler func(BlocksData) error) error {
	if err := sc.Connect(ctx); err != nil {
		return fmt.Errorf("connect websocket: %w", err)
	}

	// send subscription request
	subReq := gqlReq{
		Query:     query,
		Variables: vars,
	}

	if err := sc.conn.WriteJSON(subReq); err != nil {
		return fmt.Errorf("write subscription request: %w", err)
	}

	// handle subscription responses in goroutine
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

				// call handler function
				if err := handler(response.Data); err != nil {
					log.Printf("handle subscription response: %v", err)
				}
			}
		}
	}()

	return nil
}

// SubscribeOnce executes a subscription for a single response
func (sc *SubscriptionClient) SubscribeOnce(ctx context.Context, query string, vars map[string]interface{}, handler func(BlocksData) error) error {
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

	// wait for single response
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

// Close closes the WebSocket connection
func (sc *SubscriptionClient) Close() error {
	if sc.conn != nil {
		return sc.conn.Close()
	}
	return nil
}
