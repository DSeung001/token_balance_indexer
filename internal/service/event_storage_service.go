package service

import (
	"context"
	"fmt"
	event_parsing "gn-indexer/internal/consumer"
	"gn-indexer/internal/domain"
	"gn-indexer/internal/queue"
	"gn-indexer/internal/repository"
	"log"
	"time"
)

// EventStorageService handles storing parsed events to database
type EventStorageService struct {
	eventRepo     repository.EventRepository
	eventAttrRepo repository.EventAttrRepository
	transferRepo  repository.TransferRepository
	tokenRepo     repository.TokenRepository
	eventParser   *event_parsing.EventParser
	eventQueue    queue.EventQueue
}

// NewEventStorageService creates a new event storage service
func NewEventStorageService(
	eventRepo repository.EventRepository,
	eventAttrRepo repository.EventAttrRepository,
	transferRepo repository.TransferRepository,
	tokenRepo repository.TokenRepository,
	eventQueue queue.EventQueue,
) *EventStorageService {
	return &EventStorageService{
		eventRepo:     eventRepo,
		eventAttrRepo: eventAttrRepo,
		transferRepo:  transferRepo,
		tokenRepo:     tokenRepo,
		eventParser:   event_parsing.NewEventParser(),
		eventQueue:    eventQueue,
	}
}

// ProcessTransaction processes a transaction and stores its events
func (ess *EventStorageService) ProcessTransaction(ctx context.Context, tx *domain.Transaction) error {
	log.Printf("Processing transaction %s for events", tx.Hash)

	// Parse events from transaction
	parsedEvents, err := ess.eventParser.ParseEventsFromTransaction(tx)
	if err != nil {
		return fmt.Errorf("parse events: %w", err)
	}

	if len(parsedEvents) == 0 {
		log.Printf("No token events found in transaction %s", tx.Hash)
		return nil
	}

	log.Printf("Found %d token events in transaction %s", len(parsedEvents), tx.Hash)

	// Process each event
	for _, parsedEvent := range parsedEvents {
		if err := ess.processSingleEvent(ctx, &parsedEvent, tx); err != nil {
			return fmt.Errorf("process event: %w", err)
		}

		// Send event to queue for balance calculation
		if err := ess.eventQueue.SendEvent(ctx, &parsedEvent); err != nil {
			log.Printf("Failed to send event to queue: %v", err)
			// Don't fail the transaction processing if queue fails
		}
	}

	return nil
}

// processSingleEvent processes a single parsed event
func (ess *EventStorageService) processSingleEvent(ctx context.Context, event *domain.ParsedEvent, tx *domain.Transaction) error {
	log.Printf("Processing event %s for transaction %s", event.Type, event.TxHash)

	// 1. Store event in tx_events table
	txEvent := &domain.TxEvent{
		TxHash:     event.TxHash,
		EventIndex: event.EventIndex,
		Type:       event.Type,
		Func:       string(event.Func),
		PkgPath:    event.TokenPath,
	}

	if err := ess.eventRepo.Create(ctx, txEvent); err != nil {
		return fmt.Errorf("create tx_event: %w", err)
	}

	log.Printf("Saved event to tx_events table with ID: %d", txEvent.ID)

	// 2. Store event attributes in tx_event_attrs table
	attrs := []domain.TxEventAttr{
		{EventID: txEvent.ID, AttrIndex: 0, Key: "from", Value: event.FromAddress},
		{EventID: txEvent.ID, AttrIndex: 1, Key: "to", Value: event.ToAddress},
		{EventID: txEvent.ID, AttrIndex: 2, Key: "value", Value: fmt.Sprintf("%d", event.Amount)},
	}

	for _, attr := range attrs {
		if err := ess.eventAttrRepo.Create(ctx, &attr); err != nil {
			return fmt.Errorf("create tx_event_attr: %w", err)
		}
	}

	log.Printf("Saved %d attributes to tx_event_attrs table", len(attrs))

	// 3. Store transfer record
	transfer := &domain.Transfer{
		TxHash:      event.TxHash,
		EventIndex:  event.EventIndex,
		TokenPath:   event.TokenPath,
		FromAddress: event.FromAddress,
		ToAddress:   event.ToAddress,
		Amount:      domain.NewU64(event.Amount),
		BlockHeight: event.BlockHeight,
		CreatedAt:   time.Now(),
	}

	if err := ess.transferRepo.Create(ctx, transfer); err != nil {
		return fmt.Errorf("create transfer: %w", err)
	}

	log.Printf("Saved transfer record to transfers table")

	// 4. Register token if new
	if err := ess.tokenRepo.RegisterIfNotExists(ctx, event.TokenPath); err != nil {
		return fmt.Errorf("register token: %w", err)
	}

	log.Printf("Successfully processed event %s for transaction %s", event.Type, event.TxHash)
	return nil
}
