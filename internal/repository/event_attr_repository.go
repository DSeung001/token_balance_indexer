package repository

import (
	"context"
	"fmt"
	"gn-indexer/internal/domain"
	"gorm.io/gorm"
)

// EventAttrRepository handles event attribute data persistence
type EventAttrRepository interface {
	Create(ctx context.Context, attr *domain.TxEventAttr) error
	GetByEventID(ctx context.Context, eventID int64) ([]domain.TxEventAttr, error)
}

type postgresEventAttrRepository struct {
	db *gorm.DB
}

// NewEventAttrRepository creates a new PostgreSQL event attribute repository
func NewEventAttrRepository(db *gorm.DB) EventAttrRepository {
	return &postgresEventAttrRepository{db: db}
}

// Create saves an event attribute to the tx_event_attrs table
func (r *postgresEventAttrRepository) Create(ctx context.Context, attr *domain.TxEventAttr) error {
	// Check if attribute already exists
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.TxEventAttr{}).
		Where("event_id = ? AND attr_index = ?", attr.EventID, attr.AttrIndex).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to check attribute existence: %w", err)
	}

	if count > 0 {
		return nil
	}

	// new event attr save
	return r.db.WithContext(ctx).Table("indexer.tx_event_attrs").Create(attr).Error
}

// GetByEventID retrieves all attributes for an event
func (r *postgresEventAttrRepository) GetByEventID(ctx context.Context, eventID int64) ([]domain.TxEventAttr, error) {
	var attrs []domain.TxEventAttr
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("attr_index").
		Find(&attrs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get attributes by event ID: %w", err)
	}

	return attrs, nil
}
