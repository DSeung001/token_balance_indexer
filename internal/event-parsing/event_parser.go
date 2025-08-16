package event_parsing

import (
	"fmt"
	"gn-indexer/internal/domain"
	"strconv"
)

// EventParser parses blockchain events from transaction data
type EventParser struct{}

// NewEventParser creates a new event parser
func NewEventParser() *EventParser {
	return &EventParser{}
}

// ParseEventsFromTransaction parses all events from a transaction
func (ep *EventParser) ParseEventsFromTransaction(tx *domain.Transaction) ([]domain.ParsedEvent, error) {
	var events []domain.ParsedEvent

	// Parse events from response_json
	if tx.Response != nil {
		for i, event := range tx.Response.Events {
			if ep.IsTokenEvent(&event) {
				parsedEvent, err := ep.ParseTokenEvent(&event, tx, i)
				if err != nil {
					return nil, fmt.Errorf("parse event %d: %w", i, err)
				}
				events = append(events, *parsedEvent)
			}
		}
	}

	return events, nil
}

// IsTokenEvent checks if an event is a token-related event
func (ep *EventParser) IsTokenEvent(event *domain.GnoEvent) bool {
	// Must be Transfer type, it's all transfer type
	if event.Type != "Transfer" {
		return false
	}

	// Must have one of the token functions
	switch event.Func {
	case "Mint", "Burn", "Transfer":
		return true
	default:
		return false
	}
}

// ParseTokenEvent parses a single token event
func (ep *EventParser) ParseTokenEvent(event *domain.GnoEvent, tx *domain.Transaction, eventIndex int) (*domain.ParsedEvent, error) {
	// Extract attributes
	var fromAddr, toAddr string
	var amount int64

	for _, attr := range event.Attrs {
		switch attr.Key {
		case "from":
			fromAddr = attr.Value
		case "to":
			toAddr = attr.Value
		case "value":
			if val, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
				amount = val
			} else {
				return nil, fmt.Errorf("invalid amount value: %s", attr.Value)
			}
		}
	}

	// Determine event type based on function and addresses
	eventType := ep.determineEventType(event.Func, fromAddr, toAddr)

	return &domain.ParsedEvent{
		Type:        event.Type,
		Func:        eventType,
		TokenPath:   event.PkgPath,
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      amount,
		TxHash:      tx.Hash,
		BlockHeight: int64(tx.BlockHeight),
		EventIndex:  eventIndex,
	}, nil
}

// determineEventType determines the event type based on function and addresses
func (ep *EventParser) determineEventType(funcName, fromAddr, toAddr string) domain.EventType {
	switch funcName {
	case "Mint":
		// Mint: from="", to=address
		if fromAddr == "" && toAddr != "" {
			return domain.EventTypeMint
		}
	case "Burn":
		// Burn: from=address, to=""
		if fromAddr != "" && toAddr == "" {
			return domain.EventTypeBurn
		}
	case "Transfer":
		// Transfer: from=address, to=address
		if fromAddr != "" && toAddr != "" {
			return domain.EventTypeTransfer
		}
	}

	// Default to Transfer if pattern doesn't match
	return domain.EventTypeTransfer
}
