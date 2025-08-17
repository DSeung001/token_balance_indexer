package client

import (
	"context"
	"encoding/json"
	"fmt"
	"gn-indexer/internal/types"
	"log"
	"sync"
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
type gqlWSMessage struct {
	Type    string                 `json:"type"`
	ID      string                 `json:"id,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// Subscription represents a single GraphQL subscription
type Subscription struct {
	ID      string
	Query   string
	Vars    map[string]interface{}
	Handler func(types.BlocksData) error
	Active  bool
}

// SubscriptionClient handles GraphQL subscriptions via WebSocket
type SubscriptionClient struct {
	Endpoint      string
	conn          *websocket.Conn
	nextID        int64
	mu            sync.Mutex
	subscriptions map[string]*Subscription
}

// NewSubscriptionClient creates a new subscription client
func NewSubscriptionClient(endpoint string) *SubscriptionClient {
	return &SubscriptionClient{
		Endpoint:      endpoint,
		subscriptions: make(map[string]*Subscription),
	}
}

// generateID creates a unique ID for each subscription
func (sc *SubscriptionClient) generateID() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.nextID++
	return fmt.Sprintf("%d", sc.nextID)
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
func (sc *SubscriptionClient) Subscribe(ctx context.Context, query string, vars map[string]interface{}, handler func(types.BlocksData) error) error {
	log.Printf("Subscribe: starting subscription with query: %s", query)

	if err := sc.Connect(ctx); err != nil {
		return fmt.Errorf("connect websocket: %w", err)
	}

	log.Printf("Subscribe: websocket connected successfully")

	// generate unique ID for this subscription
	subID := sc.generateID()

	// create and store subscription
	subscription := &Subscription{
		ID:      subID,
		Query:   query,
		Vars:    vars,
		Handler: handler,
		Active:  true,
	}

	sc.mu.Lock()
	sc.subscriptions[subID] = subscription
	sc.mu.Unlock()

	// send start message with subscription
	startMsg := gqlWSMessage{
		Type: GQL_START,
		ID:   subID,
		Payload: map[string]interface{}{
			"query":     query,
			"variables": vars,
		},
	}

	log.Printf("Subscribe: sending start message with ID %s: %+v", subID, startMsg)

	if err := sc.conn.WriteJSON(startMsg); err != nil {
		// remove subscription on error
		sc.mu.Lock()
		delete(sc.subscriptions, subID)
		sc.mu.Unlock()
		return fmt.Errorf("write start message: %w", err)
	}

	log.Printf("Subscribe: start message sent successfully")

	// handle subscription responses in goroutine
	go func() {
		defer func() {
			log.Printf("Subscribe: goroutine ending for subscription %s, closing connection", subID)
			// remove subscription
			sc.mu.Lock()
			delete(sc.subscriptions, subID)
			sc.mu.Unlock()
			sc.conn.Close()
		}()

		log.Printf("Subscribe: starting message handling loop for subscription %s", subID)

		for {
			select {
			case <-ctx.Done():
				log.Printf("Subscribe: context cancelled for subscription %s, stopping subscription", subID)
				// send stop message before closing
				stopMsg := gqlWSMessage{
					Type: GQL_STOP,
					ID:   subID,
				}
				sc.conn.WriteJSON(stopMsg)
				return
			default:
				log.Printf("Subscribe: waiting for message for subscription %s...", subID)

				var response gqlWSMessage
				if err := sc.conn.ReadJSON(&response); err != nil {
					log.Printf("Subscribe: read subscription response error for subscription %s: %v", subID, err)

					// Attempt to reconnect if connection is lost
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("Subscribe: connection closed for subscription %s, attempting to reconnect...", subID)

						// Wait briefly before attempting to reconnect
						time.Sleep(2 * time.Second)

						if err := sc.reconnect(ctx); err != nil {
							log.Printf("Subscribe: reconnection failed for subscription %s: %v", subID, err)
							return
						}

						// Restart subscription after reconnection
						if err := sc.resubscribe(subscription); err != nil {
							log.Printf("Subscribe: resubscription failed for subscription %s: %v", subID, err)
							return
						}

						log.Printf("Subscribe: reconnected and resubscribed successfully for subscription %s", subID)
						continue
					}

					return
				}

				log.Printf("Subscribe: received message for subscription %s, type: %s, payload: %+v", subID, response.Type, response.Payload)

				// check if this message is for our subscription
				if response.ID != subID {
					log.Printf("Subscribe: message ID %s doesn't match subscription ID %s, skipping", response.ID, subID)
					continue
				}

				switch response.Type {
				case GQL_DATA:
					log.Printf("Subscribe: handling DATA message for subscription %s", subID)
					// handle data message
					if payload, ok := response.Payload["data"]; ok {
						log.Printf("Subscribe: data payload found for subscription %s: %+v", subID, payload)

						if data, ok := payload.(map[string]interface{}); ok {
							// convert to BlocksData
							jsonData, _ := json.Marshal(data)
							log.Printf("Subscribe: marshaled data for subscription %s: %s", subID, string(jsonData))

							var blocksData types.BlocksData
							if err := json.Unmarshal(jsonData, &blocksData); err == nil {
								log.Printf("Subscribe: unmarshaled successfully for subscription %s, block height: %d", subID, blocksData.GetBlocks.Height)

								if err := handler(blocksData); err != nil {
									log.Printf("Subscribe: handler error for subscription %s: %v", subID, err)
								} else {
									log.Printf("Subscribe: handler executed successfully for subscription %s", subID)
								}
							} else {
								log.Printf("Subscribe: unmarshal error for subscription %s: %v", subID, err)
							}
						} else {
							log.Printf("Subscribe: payload data is not map[string]interface{} for subscription %s: %T", subID, payload)
						}
					} else {
						log.Printf("Subscribe: no data payload found in response for subscription %s", subID)
					}
				case GQL_ERROR:
					log.Printf("Subscribe: handling ERROR message for subscription %s", subID)
					if payload, ok := response.Payload["errors"]; ok {
						log.Printf("Subscribe: subscription error for subscription %s: %+v", subID, payload)
						// Todo: re connecting logic
					}
				case GQL_COMPLETE:
					log.Printf("Subscribe: handling COMPLETE message for subscription %s", subID)
					log.Printf("Subscribe: subscription %s completed", subID)
					return
				case GQL_CONNECTION_ERROR:
					log.Printf("Subscribe: handling CONNECTION_ERROR message for subscription %s", subID)
					// connection error handling
					if payload, ok := response.Payload["message"]; ok {
						log.Printf("Subscribe: connection error for subscription %s: %v", subID, payload)
					} else {
						log.Printf("Subscribe: connection error for subscription %s: unknown reason", subID)
					}
					stopMsg := gqlWSMessage{
						Type: GQL_STOP,
						ID:   subID,
					}
					_ = sc.conn.WriteJSON(stopMsg)
					return
				case GQL_CONNECTION_ACK:
					log.Printf("Subscribe: received connection_ack for subscription %s (this should not happen after start)", subID)
				default:
					log.Printf("Subscribe: unknown message type %s for subscription %s", response.Type, subID)
				}
			}
		}
	}()

	log.Printf("Subscribe: subscription %s started successfully", subID)
	return nil
}

// SubscribeOnce starts a one-time subscription, receives one data message, then stops
func (sc *SubscriptionClient) SubscribeOnce(ctx context.Context, query string, vars map[string]interface{}, handler func(types.BlocksData) error) error {
	log.Printf("SubscribeOnce: starting one-time subscription with query: %s", query)

	if err := sc.Connect(ctx); err != nil {
		return fmt.Errorf("connect websocket: %w", err)
	}

	log.Printf("SubscribeOnce: websocket connected successfully")

	// generate unique ID for this subscription
	subID := sc.generateID()

	// send start message with subscription
	startMsg := gqlWSMessage{
		Type: GQL_START,
		ID:   subID,
		Payload: map[string]interface{}{
			"query":     query,
			"variables": vars,
		},
	}

	log.Printf("SubscribeOnce: sending start message with ID %s: %+v", subID, startMsg)

	if err := sc.conn.WriteJSON(startMsg); err != nil {
		return fmt.Errorf("write start message: %w", err)
	}

	log.Printf("SubscribeOnce: start message sent successfully, waiting for data...")

	// wait for data message
	for {
		select {
		case <-ctx.Done():
			log.Printf("SubscribeOnce: context cancelled for subscription %s", subID)
			sc.conn.Close()
			return ctx.Err()
		default:
			var response gqlWSMessage
			if err := sc.conn.ReadJSON(&response); err != nil {
				log.Printf("SubscribeOnce: read subscription response error for subscription %s: %v", subID, err)
				sc.conn.Close()
				return fmt.Errorf("read subscription response: %w", err)
			}

			log.Printf("SubscribeOnce: received message for subscription %s, type: %s, payload: %+v", subID, response.Type, response.Payload)

			// check if this message is for our subscription
			if response.ID != subID {
				log.Printf("SubscribeOnce: message ID %s doesn't match subscription ID %s, skipping", response.ID, subID)
				continue
			}

			switch response.Type {
			case GQL_DATA:
				log.Printf("SubscribeOnce: handling DATA message for subscription %s", subID)
				// handle data message
				if payload, ok := response.Payload["data"]; ok {
					log.Printf("SubscribeOnce: data payload found for subscription %s: %+v", subID, payload)

					if data, ok := payload.(map[string]interface{}); ok {
						// convert to BlocksData
						jsonData, _ := json.Marshal(data)
						log.Printf("SubscribeOnce: marshaled data for subscription %s: %s", subID, string(jsonData))

						var blocksData types.BlocksData
						if err := json.Unmarshal(jsonData, &blocksData); err == nil {
							log.Printf("SubscribeOnce: unmarshaled successfully for subscription %s, block height: %d", subID, blocksData.GetBlocks.Height)

							if err := handler(blocksData); err != nil {
								log.Printf("SubscribeOnce: handler error for subscription %s: %v", subID, err)
								sc.conn.Close()
								return fmt.Errorf("handler error: %w", err)
							}

							log.Printf("SubscribeOnce: handler executed successfully for subscription %s", subID)
						} else {
							log.Printf("SubscribeOnce: unmarshal error for subscription %s: %v", subID, err)
							sc.conn.Close()
							return fmt.Errorf("unmarshal error: %w", err)
						}
					} else {
						log.Printf("SubscribeOnce: payload data is not map[string]interface{} for subscription %s: %T", subID, payload)
						sc.conn.Close()
						return fmt.Errorf("invalid payload data type: %T", payload)
					}
				} else {
					log.Printf("SubscribeOnce: no data payload found in response for subscription %s", subID)
					sc.conn.Close()
					return fmt.Errorf("no data payload found")
				}

				// send stop message and close connection
				stopMsg := gqlWSMessage{
					Type: GQL_STOP,
					ID:   subID,
				}
				sc.conn.WriteJSON(stopMsg)
				sc.conn.Close()
				log.Printf("SubscribeOnce: one-time subscription %s completed successfully", subID)
				return nil

			case GQL_ERROR:
				log.Printf("SubscribeOnce: handling ERROR message for subscription %s", subID)
				if payload, ok := response.Payload["errors"]; ok {
					log.Printf("SubscribeOnce: subscription error for subscription %s: %+v", subID, payload)
				}
				sc.conn.Close()
				return fmt.Errorf("subscription error for subscription %s: %+v", subID, response.Payload)

			case GQL_CONNECTION_ERROR:
				log.Printf("SubscribeOnce: handling CONNECTION_ERROR message for subscription %s", subID)
				if payload, ok := response.Payload["message"]; ok {
					log.Printf("SubscribeOnce: connection error for subscription %s: %v", subID, payload)
				} else {
					log.Printf("SubscribeOnce: connection error for subscription %s: unknown reason", subID)
				}
				sc.conn.Close()
				return fmt.Errorf("connection error for subscription %s: %+v", subID, response.Payload)

			case GQL_COMPLETE:
				log.Printf("SubscribeOnce: handling COMPLETE message for subscription %s", subID)
				sc.conn.Close()
				return fmt.Errorf("subscription %s completed without data", subID)

			case GQL_CONNECTION_ACK:
				log.Printf("SubscribeOnce: received connection_ack for subscription %s (this should not happen after start)", subID)

			default:
				log.Printf("SubscribeOnce: unknown message type %s for subscription %s", response.Type, subID)
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
func (sc *SubscriptionClient) resubscribe(subscription *Subscription) error {
	log.Printf("resubscribe: sending start message again for subscription %s...", subscription.ID)

	startMsg := gqlWSMessage{
		Type: GQL_START,
		ID:   subscription.ID,
		Payload: map[string]interface{}{
			"query":     subscription.Query,
			"variables": subscription.Vars,
		},
	}

	return sc.conn.WriteJSON(startMsg)
}

// StopSubscription stops a specific subscription by ID
func (sc *SubscriptionClient) StopSubscription(subscriptionID string) error {
	sc.mu.Lock()
	subscription, exists := sc.subscriptions[subscriptionID]
	sc.mu.Unlock()

	if !exists {
		return fmt.Errorf("subscription %s not found", subscriptionID)
	}

	if !subscription.Active {
		return fmt.Errorf("subscription %s is already inactive", subscriptionID)
	}

	// send stop message
	stopMsg := gqlWSMessage{
		Type: GQL_STOP,
		ID:   subscriptionID,
	}

	if err := sc.conn.WriteJSON(stopMsg); err != nil {
		return fmt.Errorf("failed to send stop message for subscription %s: %w", subscriptionID, err)
	}

	// mark subscription as inactive
	subscription.Active = false

	log.Printf("StopSubscription: subscription %s stopped successfully", subscriptionID)
	return nil
}

// GetActiveSubscriptions returns a list of active subscription IDs
func (sc *SubscriptionClient) GetActiveSubscriptions() []string {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	var activeIDs []string
	for id, sub := range sc.subscriptions {
		if sub.Active {
			activeIDs = append(activeIDs, id)
		}
	}
	return activeIDs
}

// Close closes the WebSocket connection and stops all subscriptions
func (sc *SubscriptionClient) Close() error {
	if sc.conn != nil {
		// stop all active subscriptions
		sc.mu.Lock()
		for id, sub := range sc.subscriptions {
			if sub.Active {
				stopMsg := gqlWSMessage{
					Type: GQL_STOP,
					ID:   id,
				}
				sc.conn.WriteJSON(stopMsg)
				sub.Active = false
			}
		}
		sc.mu.Unlock()

		// send connection_terminate message
		termMsg := gqlWSMessage{
			Type: GQL_CONNECTION_TERMINATE,
		}
		sc.conn.WriteJSON(termMsg)
		return sc.conn.Close()
	}
	return nil
}
