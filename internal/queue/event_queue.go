package queue

import (
	"context"
	"gn-indexer/internal/domain"
)

// EventQueue defines the interface for event queue operations
type EventQueue interface {
	// SendEvent sends a parsed event to the queue
	SendEvent(ctx context.Context, event *domain.ParsedEvent) error

	// ReceiveEvents receives multiple events from the queue
	ReceiveEvents(ctx context.Context) ([]*domain.ParsedEvent, error)

	// Close closes the queue connection
	Close() error
}

// QueueConfig holds configuration for queue operations
type QueueConfig struct {
	QueueName          string
	EndpointURL        string
	Region             string
	AccessKeyID        string
	SecretAccessKey    string
	MaxReceiveMessages int
	VisibilityTimeout  int
}
