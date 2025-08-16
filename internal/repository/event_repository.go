package repository

import (
	"context"
	"fmt"
	"gn-indexer/internal/domain"
	"gorm.io/gorm"
)

// EventRepository handles event data persistence
type EventRepository interface {
	Create(ctx context.Context, event *domain.TxEvent) error
	GetByTxHash(ctx context.Context, txHash string) ([]domain.TxEvent, error)
	GetByTxHashAndIndex(ctx context.Context, txHash string, eventIndex int) (*domain.TxEvent, error)
}

type postgresEventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates a new PostgreSQL event repository
func NewEventRepository(db *gorm.DB) EventRepository {
	return &postgresEventRepository{db: db}
}

// Create saves an event to the tx_events table
func (r *postgresEventRepository) Create(ctx context.Context, event *domain.TxEvent) error {
	// Check if event already exists
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.TxEvent{}).
		Where("tx_hash = ? AND event_index = ?", event.TxHash, event.EventIndex).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to check event existence: %w", err)
	}

	if count > 0 {
		return nil
	}

	// new event save
	return r.db.WithContext(ctx).Table("indexer.tx_events").Create(event).Error
}

// GetByTxHash retrieves all events for a transaction
func (r *postgresEventRepository) GetByTxHash(ctx context.Context, txHash string) ([]domain.TxEvent, error) {
	var events []domain.TxEvent
	err := r.db.WithContext(ctx).
		Where("tx_hash = ?", txHash).
		Order("event_index").
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get events by tx hash: %w", err)
	}

	return events, nil
}

// GetByTxHashAndIndex retrieves a specific event by transaction hash and event index
func (r *postgresEventRepository) GetByTxHashAndIndex(ctx context.Context, txHash string, eventIndex int) (*domain.TxEvent, error) {
	var event domain.TxEvent
	err := r.db.WithContext(ctx).
		Where("tx_hash = ? AND event_index = ?", txHash, eventIndex).
		First(&event).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &event, nil
}
