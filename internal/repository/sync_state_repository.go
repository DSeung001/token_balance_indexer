package repository

import (
	"context"
	"gn-indexer/internal/models"
	"gorm.io/gorm"
)

type SyncStateRepository interface {
	GetState(ctx context.Context, component string) (*models.SyncState, error)
	UpdateState(ctx context.Context, state *models.SyncState) error
	UpdateLastBlock(ctx context.Context, component string, height int64, txHash string) error
	GetLastSyncedHeight(ctx context.Context, component string) (int64, error)
	InitializeComponent(ctx context.Context, component string) error
}

type syncStateRepository struct {
	db *gorm.DB
}
