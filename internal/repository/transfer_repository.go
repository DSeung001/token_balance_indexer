package repository

import (
	"context"
	"fmt"
	"gn-indexer/internal/domain"

	"gorm.io/gorm"
)

// TransferRepository handles transfer data persistence
type TransferRepository interface {
	Create(ctx context.Context, transfer *domain.Transfer) error
	GetByTxHash(ctx context.Context, txHash string) ([]domain.Transfer, error)
	GetByAddress(ctx context.Context, address string) ([]domain.Transfer, error)
	GetByTokenPath(ctx context.Context, tokenPath string) ([]domain.Transfer, error)
	GetAll(ctx context.Context) ([]domain.Transfer, error)
}

type postgresTransferRepository struct {
	db *gorm.DB
}

// NewTransferRepository creates a new PostgreSQL transfer repository
func NewTransferRepository(db *gorm.DB) TransferRepository {
	return &postgresTransferRepository{db: db}
}

// Create saves a transfer record to the transfers table
func (r *postgresTransferRepository) Create(ctx context.Context, transfer *domain.Transfer) error {
	// Check if transfer already exists (멱등성 보장)
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Transfer{}).
		Where("tx_hash = ? AND event_index = ?", transfer.TxHash, transfer.EventIndex).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to check transfer existence: %w", err)
	}

	if count > 0 {
		return nil
	}

	// new transfer save
	return r.db.WithContext(ctx).Create(transfer).Error
}

// GetByTxHash retrieves all transfers for a transaction
func (r *postgresTransferRepository) GetByTxHash(ctx context.Context, txHash string) ([]domain.Transfer, error) {
	var transfers []domain.Transfer
	err := r.db.WithContext(ctx).
		Where("tx_hash = ?", txHash).
		Order("event_index"). // 이벤트 순서대로 정렬
		Find(&transfers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get transfers by tx hash: %w", err)
	}

	return transfers, nil
}

// GetByAddress retrieves all transfers involving a specific address
func (r *postgresTransferRepository) GetByAddress(ctx context.Context, address string) ([]domain.Transfer, error) {
	var transfers []domain.Transfer
	err := r.db.WithContext(ctx).
		Where("from_address = ? OR to_address = ?", address, address).
		Order("block_height DESC, event_index DESC").
		Find(&transfers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get transfers by address: %w", err)
	}

	return transfers, nil
}

// GetByTokenPath retrieves all transfers for a specific token
func (r *postgresTransferRepository) GetByTokenPath(ctx context.Context, tokenPath string) ([]domain.Transfer, error) {
	var transfers []domain.Transfer
	err := r.db.WithContext(ctx).
		Where("token_path = ?", tokenPath).
		Order("block_height DESC, event_index DESC").
		Find(&transfers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get transfers by token path: %w", err)
	}

	return transfers, nil
}

// GetAll retrieves all transfers
func (r *postgresTransferRepository) GetAll(ctx context.Context) ([]domain.Transfer, error) {
	var transfers []domain.Transfer
	err := r.db.WithContext(ctx).
		Order("block_height DESC, event_index DESC").
		Find(&transfers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get all transfers: %w", err)
	}

	return transfers, nil
}
