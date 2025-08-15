package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

// GraphQL over WebSocket Protocol
// reference: https://github.com/enisdenjo/graphql-ws/blob/master/PROTOCOL.md
// reference: https://github.com/GraphQL/graphql-over-http/blob/main/rfcs/GraphQLOverWebSocket.md

const (
	GQL_CONNECTION_INIT      = "connection_init"      // client → server: initialize connection
	GQL_START                = "start"                // client → server: start subscription
	GQL_STOP                 = "stop"                 // client → server: stop subscription
	GQL_CONNECTION_TERMINATE = "connection_terminate" // client → server: terminate connection

	GQL_DATA             = "data"             // server → client: send data
	GQL_ERROR            = "error"            // server → client: error occurred
	GQL_COMPLETE         = "complete"         // server → client: subscription complete
	GQL_CONNECTION_ACK   = "connection_ack"   // server → client: connection acknowledged
	GQL_CONNECTION_ERROR = "connection_error" // server → client: connection error
)

// GraphQL WebSocket message structure
// Todo: id uniqueness
type gqlWSMessage struct {
	Type    string                 `json:"type"`
	ID      string                 `json:"id,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

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

// Connect establishes WebSocket connection and initializes GraphQL protocol
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

	// send connection_init message
	initMsg := gqlWSMessage{
		Type: GQL_CONNECTION_INIT,
		Payload: map[string]interface{}{
			"type": "connection_init",
		},
	}

	if err := sc.conn.WriteJSON(initMsg); err != nil {
		sc.conn.Close()
		return fmt.Errorf("send connection_init: %w", err)
	}

	// wait for connection_ack
	var ackMsg gqlWSMessage
	if err := sc.conn.ReadJSON(&ackMsg); err != nil {
		sc.conn.Close()
		return fmt.Errorf("read connection_ack: %w", err)
	}

	if ackMsg.Type != GQL_CONNECTION_ACK {
		sc.conn.Close()
		return fmt.Errorf("expected connection_ack, got: %s", ackMsg.Type)
	}

	return nil
}

// Subscribe starts a persistent subscription
func (sc *SubscriptionClient) Subscribe(ctx context.Context, query string, vars map[string]interface{}, handler func(BlocksData) error) error {
	if err := sc.Connect(ctx); err != nil {
		return fmt.Errorf("connect websocket: %w", err)
	}

	// send start message with subscription
	startMsg := gqlWSMessage{
		Type: GQL_START,
		ID:   "1",
		Payload: map[string]interface{}{
			"query":     query,
			"variables": vars,
		},
	}

	if err := sc.conn.WriteJSON(startMsg); err != nil {
		return fmt.Errorf("write start message: %w", err)
	}

	// handle subscription responses in goroutine
	go func() {
		defer sc.conn.Close()

		for {
			select {
			case <-ctx.Done():
				// send stop message before closing
				stopMsg := gqlWSMessage{
					Type: GQL_STOP,
					ID:   "1",
				}
				sc.conn.WriteJSON(stopMsg)
				return
			default:
				var response gqlWSMessage
				if err := sc.conn.ReadJSON(&response); err != nil {
					log.Printf("read subscription response: %v", err)
					return
				}

				switch response.Type {
				case GQL_DATA:
					// handle data message
					if payload, ok := response.Payload["data"]; ok {
						if data, ok := payload.(map[string]interface{}); ok {
							// convert to BlocksData
							jsonData, _ := json.Marshal(data)
							var blocksData BlocksData
							if err := json.Unmarshal(jsonData, &blocksData); err == nil {
								if err := handler(blocksData); err != nil {
									log.Printf("handle subscription response: %v", err)
								}
							}
						}
					}
				case GQL_ERROR:
					if payload, ok := response.Payload["errors"]; ok {
						log.Printf("subscription error: %+v", payload)
						// Todo: re connecting logic
					}
				case GQL_CONNECTION_ERROR:
					// 연결 에러 처리
					if payload, ok := response.Payload["message"]; ok {
						log.Printf("connection error: %v", payload)
					} else {
						log.Printf("connection error: unknown reason")
					}
					stopMsg := gqlWSMessage{
						Type: GQL_STOP,
						ID:   "1",
					}
					_ = sc.conn.WriteJSON(stopMsg)
					return
				case GQL_COMPLETE:
					log.Printf("subscription completed")
					return
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

	// send start message
	startMsg := gqlWSMessage{
		Type: GQL_START,
		ID:   "1",
		Payload: map[string]interface{}{
			"query":     query,
			"variables": vars,
		},
	}

	if err := sc.conn.WriteJSON(startMsg); err != nil {
		return fmt.Errorf("write start message: %w", err)
	}

	// wait for data message
	for {
		var response gqlWSMessage
		if err := sc.conn.ReadJSON(&response); err != nil {
			return fmt.Errorf("read subscription response: %w", err)
		}

		switch response.Type {
		case GQL_DATA:
			// handle data message
			if payload, ok := response.Payload["data"]; ok {
				if data, ok := payload.(map[string]interface{}); ok {
					// convert to BlocksData
					jsonData, _ := json.Marshal(data)
					var blocksData BlocksData
					if err := json.Unmarshal(jsonData, &blocksData); err == nil {
						if err := handler(blocksData); err != nil {
							log.Printf("handle subscription response: %v", err)
						}
					}
				}
			}
			// send stop message
			stopMsg := gqlWSMessage{
				Type: GQL_STOP,
				ID:   "1",
			}
			sc.conn.WriteJSON(stopMsg)
			break
		case GQL_CONNECTION_ERROR:
			if payload, ok := response.Payload["message"]; ok {
				log.Printf("connection error: %v", payload)
			} else {
				log.Printf("connection error: unknown reason")
			}
			return fmt.Errorf("connection error: %v", response.Payload)
		case GQL_ERROR:
			return fmt.Errorf("subscription error: %+v", response.Payload)
		case GQL_COMPLETE:
			break
		}
	}

	sc.conn.Close()
	sc.conn = nil
	return nil
}

// Close closes the WebSocket connection
func (sc *SubscriptionClient) Close() error {
	if sc.conn != nil {
		// send connection_terminate message
		termMsg := gqlWSMessage{
			Type: GQL_CONNECTION_TERMINATE,
		}
		sc.conn.WriteJSON(termMsg)
		return sc.conn.Close()
	}
	return nil
}
