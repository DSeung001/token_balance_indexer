package service

import (
	"context"
	"fmt"
	"gn-indexer/internal/queue"
	"log"
)

// EventProcessorService handles consuming events from queue and processing them
type EventProcessorService struct {
	eventQueue     queue.EventQueue
	balanceService *BalanceService
}

// NewEventProcessorService creates a new event processor service
func NewEventProcessorService(
	eventQueue queue.EventQueue,
	balanceService *BalanceService,
) *EventProcessorService {
	return &EventProcessorService{
		eventQueue:     eventQueue,
		balanceService: balanceService,
	}
}

// Start begins continuous processing events from the queue
func (eps *EventProcessorService) Start(ctx context.Context) error {
	log.Printf("EventProcessorService: starting continuous event processing")

	for {
		select {
		case <-ctx.Done():
			log.Printf("EventProcessorService: context cancelled, stopping")
			return ctx.Err()
		default:
			// Use SQS Long Polling to receive events
			if err := eps.processBatch(ctx); err != nil {
				log.Printf("EventProcessorService: error processing events: %v", err)
				// Continue processing even if batch fails
			}
		}
	}
}

// ProcessSingleBatch processes exactly one batch of events and returns the count
func (eps *EventProcessorService) ProcessSingleBatch(ctx context.Context, batchSize int) (int, error) {
	log.Printf("EventProcessorService: processing single batch with size: %d", batchSize)

	// Receive events from queue (SQS Long Polling handles the waiting)
	events, err := eps.eventQueue.ReceiveEvents(ctx)
	if err != nil {
		return 0, fmt.Errorf("receive events from queue: %w", err)
	}

	// If no events available, return 0 (this is normal)
	if len(events) == 0 {
		log.Printf("EventProcessorService: no events available in queue")
		return 0, nil
	}

	log.Printf("EventProcessorService: processing %d events", len(events))

	// Process all events
	processedCount := 0
	for _, event := range events {
		log.Printf("EventProcessorService: processing event %s for token %s", event.Type, event.TokenPath)

		// Process the event
		if err := eps.balanceService.ProcessEvent(ctx, event); err != nil {
			log.Printf("EventProcessorService: error processing event %s: %v", event.Type, err)
			// Continue processing other events even if one fails
			continue
		}

		processedCount++
		log.Printf("EventProcessorService: successfully processed event %s", event.Type)
	}

	log.Printf("EventProcessorService: batch processing completed, processed %d/%d events", processedCount, len(events))
	return processedCount, nil
}

// processBatch processes multiple events from the queue (internal method for continuous processing)
func (eps *EventProcessorService) processBatch(ctx context.Context) error {
	// Receive events from queue (SQS Long Polling handles the waiting)
	events, err := eps.eventQueue.ReceiveEvents(ctx)
	if err != nil {
		return fmt.Errorf("receive events from queue: %w", err)
	}

	// If no events available, return (SQS will wait up to 20 seconds for new messages)
	if len(events) == 0 {
		return nil
	}

	log.Printf("EventProcessorService: processing %d events", len(events))

	// Process all events
	processedCount := 0
	for _, event := range events {
		log.Printf("EventProcessorService: processing event %s for token %s", event.Type, event.TokenPath)

		// Process the event
		if err := eps.balanceService.ProcessEvent(ctx, event); err != nil {
			log.Printf("EventProcessorService: error processing event %s: %v", event.Type, err)
			// Continue processing other events even if one fails
			continue
		}

		processedCount++
		log.Printf("EventProcessorService: successfully processed event %s", event.Type)
	}

	log.Printf("EventProcessorService: processing completed, processed %d/%d events", processedCount, len(events))
	return nil
}
