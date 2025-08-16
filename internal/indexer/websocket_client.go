package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
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
	log.Printf("Subscribe: starting subscription with query: %s", query)

	if err := sc.Connect(ctx); err != nil {
		return fmt.Errorf("connect websocket: %w", err)
	}

	log.Printf("Subscribe: websocket connected successfully")

	// send start message with subscription
	startMsg := gqlWSMessage{
		Type: GQL_START,
		ID:   "1",
		Payload: map[string]interface{}{
			"query":     query,
			"variables": vars,
		},
	}

	log.Printf("Subscribe: sending start message: %+v", startMsg)

	if err := sc.conn.WriteJSON(startMsg); err != nil {
		return fmt.Errorf("write start message: %w", err)
	}

	log.Printf("Subscribe: start message sent successfully")

	// handle subscription responses in goroutine
	go func() {
		defer func() {
			log.Printf("Subscribe: goroutine ending, closing connection")
			sc.conn.Close()
		}()

		log.Printf("Subscribe: starting message handling loop")

		for {
			select {
			case <-ctx.Done():
				log.Printf("Subscribe: context cancelled, stopping subscription")
				// send stop message before closing
				stopMsg := gqlWSMessage{
					Type: GQL_STOP,
					ID:   "1",
				}
				sc.conn.WriteJSON(stopMsg)
				return
			default:
				log.Printf("Subscribe: waiting for message...")

				var response gqlWSMessage
				if err := sc.conn.ReadJSON(&response); err != nil {
					log.Printf("Subscribe: read subscription response error: %v", err)

					// 연결이 끊어진 경우 재연결 시도
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("Subscribe: connection closed, attempting to reconnect...")

						// 잠시 대기 후 재연결 시도
						time.Sleep(2 * time.Second)

						if err := sc.reconnect(ctx); err != nil {
							log.Printf("Subscribe: reconnection failed: %v", err)
							return
						}

						// 재연결 후 구독 재시작
						if err := sc.resubscribe(ctx, query, vars); err != nil {
							log.Printf("Subscribe: resubscription failed: %v", err)
							return
						}

						log.Printf("Subscribe: reconnected and resubscribed successfully")
						continue
					}

					return
				}

				log.Printf("Subscribe: received message type: %s, payload: %+v", response.Type, response.Payload)

				switch response.Type {
				case GQL_DATA:
					log.Printf("Subscribe: handling DATA message")
					// handle data message
					if payload, ok := response.Payload["data"]; ok {
						log.Printf("Subscribe: data payload found: %+v", payload)

						if data, ok := payload.(map[string]interface{}); ok {
							// convert to BlocksData
							jsonData, _ := json.Marshal(data)
							log.Printf("Subscribe: marshaled data: %s", string(jsonData))

							var blocksData BlocksData
							if err := json.Unmarshal(jsonData, &blocksData); err == nil {
								log.Printf("Subscribe: unmarshaled successfully, block height: %d", blocksData.GetBlocks.Height)

								if err := handler(blocksData); err != nil {
									log.Printf("Subscribe: handler error: %v", err)
								} else {
									log.Printf("Subscribe: handler executed successfully")
								}
							} else {
								log.Printf("Subscribe: unmarshal error: %v", err)
							}
						} else {
							log.Printf("Subscribe: payload data is not map[string]interface{}: %T", payload)
						}
					} else {
						log.Printf("Subscribe: no data payload found in response")
					}
				case GQL_ERROR:
					log.Printf("Subscribe: handling ERROR message")
					if payload, ok := response.Payload["errors"]; ok {
						log.Printf("Subscribe: subscription error: %+v", payload)
						// Todo: re connecting logic
					}
				case GQL_CONNECTION_ERROR:
					log.Printf("Subscribe: handling CONNECTION_ERROR message")
					// connection error handling
					if payload, ok := response.Payload["message"]; ok {
						log.Printf("Subscribe: connection error: %v", payload)
					} else {
						log.Printf("Subscribe: connection error: unknown reason")
					}
					stopMsg := gqlWSMessage{
						Type: GQL_STOP,
						ID:   "1",
					}
					_ = sc.conn.WriteJSON(stopMsg)
					return
				case GQL_COMPLETE:
					log.Printf("Subscribe: handling COMPLETE message")
					log.Printf("Subscribe: subscription completed")
					return
				case GQL_CONNECTION_ACK:
					log.Printf("Subscribe: received connection_ack (this should not happen after start)")
				default:
					log.Printf("Subscribe: unknown message type: %s", response.Type)
				}
			}
		}
	}()

	log.Printf("Subscribe: subscription started successfully")
	return nil
}

// SubscribeOnce starts a one-time subscription, receives one data message, then stops
func (sc *SubscriptionClient) SubscribeOnce(ctx context.Context, query string, vars map[string]interface{}, handler func(BlocksData) error) error {
	log.Printf("SubscribeOnce: starting one-time subscription with query: %s", query)

	if err := sc.Connect(ctx); err != nil {
		return fmt.Errorf("connect websocket: %w", err)
	}

	log.Printf("SubscribeOnce: websocket connected successfully")

	// send start message with subscription
	startMsg := gqlWSMessage{
		Type: GQL_START,
		ID:   "1",
		Payload: map[string]interface{}{
			"query":     query,
			"variables": vars,
		},
	}

	log.Printf("SubscribeOnce: sending start message: %+v", startMsg)

	if err := sc.conn.WriteJSON(startMsg); err != nil {
		return fmt.Errorf("write start message: %w", err)
	}

	log.Printf("SubscribeOnce: start message sent successfully, waiting for data...")

	// wait for data message
	for {
		select {
		case <-ctx.Done():
			log.Printf("SubscribeOnce: context cancelled")
			sc.conn.Close()
			return ctx.Err()
		default:
			var response gqlWSMessage
			if err := sc.conn.ReadJSON(&response); err != nil {
				log.Printf("SubscribeOnce: read subscription response error: %v", err)
				sc.conn.Close()
				return fmt.Errorf("read subscription response: %w", err)
			}

			log.Printf("SubscribeOnce: received message type: %s, payload: %+v", response.Type, response.Payload)

			switch response.Type {
			case GQL_DATA:
				log.Printf("SubscribeOnce: handling DATA message")
				// handle data message
				if payload, ok := response.Payload["data"]; ok {
					log.Printf("SubscribeOnce: data payload found: %+v", payload)

					if data, ok := payload.(map[string]interface{}); ok {
						// convert to BlocksData
						jsonData, _ := json.Marshal(data)
						log.Printf("SubscribeOnce: marshaled data: %s", string(jsonData))

						var blocksData BlocksData
						if err := json.Unmarshal(jsonData, &blocksData); err == nil {
							log.Printf("SubscribeOnce: unmarshaled successfully, block height: %d", blocksData.GetBlocks.Height)

							if err := handler(blocksData); err != nil {
								log.Printf("SubscribeOnce: handler error: %v", err)
								sc.conn.Close()
								return fmt.Errorf("handler error: %w", err)
							}

							log.Printf("SubscribeOnce: handler executed successfully")
						} else {
							log.Printf("SubscribeOnce: unmarshal error: %v", err)
							sc.conn.Close()
							return fmt.Errorf("unmarshal error: %w", err)
						}
					} else {
						log.Printf("SubscribeOnce: payload data is not map[string]interface{}: %T", payload)
						sc.conn.Close()
						return fmt.Errorf("invalid payload data type: %T", payload)
					}
				} else {
					log.Printf("SubscribeOnce: no data payload found in response")
					sc.conn.Close()
					return fmt.Errorf("no data payload found")
				}

				// send stop message and close connection
				stopMsg := gqlWSMessage{
					Type: GQL_STOP,
					ID:   "1",
				}
				sc.conn.WriteJSON(stopMsg)
				sc.conn.Close()
				log.Printf("SubscribeOnce: one-time subscription completed successfully")
				return nil

			case GQL_ERROR:
				log.Printf("SubscribeOnce: handling ERROR message")
				if payload, ok := response.Payload["errors"]; ok {
					log.Printf("SubscribeOnce: subscription error: %+v", payload)
				}
				sc.conn.Close()
				return fmt.Errorf("subscription error: %+v", response.Payload)

			case GQL_CONNECTION_ERROR:
				log.Printf("SubscribeOnce: handling CONNECTION_ERROR message")
				if payload, ok := response.Payload["message"]; ok {
					log.Printf("SubscribeOnce: connection error: %v", payload)
				} else {
					log.Printf("SubscribeOnce: connection error: unknown reason")
				}
				sc.conn.Close()
				return fmt.Errorf("connection error: %+v", response.Payload)

			case GQL_COMPLETE:
				log.Printf("SubscribeOnce: handling COMPLETE message")
				sc.conn.Close()
				return fmt.Errorf("subscription completed without data")

			case GQL_CONNECTION_ACK:
				log.Printf("SubscribeOnce: received connection_ack (this should not happen after start)")

			default:
				log.Printf("SubscribeOnce: unknown message type: %s", response.Type)
			}
		}
	}
}

// reconnect attempts to reconnect to the websocket
func (sc *SubscriptionClient) reconnect(ctx context.Context) error {
	log.Printf("reconnect: attempting to reconnect...")

	if sc.conn != nil {
		sc.conn.Close()
		sc.conn = nil
	}

	return sc.Connect(ctx)
}

// resubscribe sends the subscription start message again
func (sc *SubscriptionClient) resubscribe(ctx context.Context, query string, vars map[string]interface{}) error {
	log.Printf("resubscribe: sending start message again...")

	startMsg := gqlWSMessage{
		Type: GQL_START,
		ID:   "1",
		Payload: map[string]interface{}{
			"query":     query,
			"variables": vars,
		},
	}

	return sc.conn.WriteJSON(startMsg)
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
