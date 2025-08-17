package repository

import (
	"context"
	"fmt"
	"gn-indexer/internal/domain"
	"gorm.io/gorm"
)

// BlockRepository handles block data persistence
type BlockRepository interface {
	SaveBlock(ctx context.Context, block domain.Block) error
	GetLastSyncedHeight(ctx context.Context) (int, error)
	GetBlockByHash(ctx context.Context, hash string) (*domain.Block, error)
	GetBlockByHeight(ctx context.Context, height int) (*domain.Block, error)
}

type postgresBlockRepository struct {
	db *gorm.DB
}

// NewBlockRepository creates a new PostgreSQL block repository
func NewBlockRepository(db *gorm.DB) BlockRepository {
	return &postgresBlockRepository{db: db}
}

// SaveBlock saves a block to database with duplication check
func (r *postgresBlockRepository) SaveBlock(ctx context.Context, block domain.Block) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Block{}).Where("hash = ?", block.Hash).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check block existence: %w", err)
	}
	if count > 0 {
		return nil // already exists
	}

	return r.db.WithContext(ctx).Create(&block).Error
}

// GetLastSyncedHeight returns the height of the last synchronized block
func (r *postgresBlockRepository) GetLastSyncedHeight(ctx context.Context) (int, error) {
	var lastHeight int
	err := r.db.WithContext(ctx).Model(&domain.Block{}).Select("COALESCE(MAX(height), 1)").Scan(&lastHeight).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get last block height: %w", err)
	}
	return lastHeight, nil
}

// GetBlockByHash retrieves a block by its hash
func (r *postgresBlockRepository) GetBlockByHash(ctx context.Context, hash string) (*domain.Block, error) {
	var block domain.Block
	err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&block).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %w", err)
	}
	return &block, nil
}

// GetBlockByHeight retrieves a block by its height
func (r *postgresBlockRepository) GetBlockByHeight(ctx context.Context, height int) (*domain.Block, error) {
	var block domain.Block
	err := r.db.WithContext(ctx).Where("height = ?", height).First(&block).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get block by height: %w", err)
	}
	return &block, nil
}
