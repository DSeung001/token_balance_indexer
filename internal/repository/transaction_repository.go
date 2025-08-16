package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"gn-indexer/internal/domain"
	"gorm.io/gorm"
)

// TransactionRepository handles transaction data persistence
type TransactionRepository interface {
	SaveTransaction(ctx context.Context, tx domain.Transaction) error
	GetTransactionByHash(ctx context.Context, hash string) (*domain.Transaction, error)
	GetTransactionsByBlockHeight(ctx context.Context, blockHeight int) ([]domain.Transaction, error)
}

type postgresTransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new PostgreSQL transaction repository
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &postgresTransactionRepository{db: db}
}

// SaveTransaction saves a transaction to database with duplication check
func (r *postgresTransactionRepository) SaveTransaction(ctx context.Context, tx domain.Transaction) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Transaction{}).Where("hash = ?", tx.Hash).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check transaction existence: %w", err)
	}
	if count > 0 {
		return nil // already exists
	}

	// gas_fee to JSON
	var gasFeeJSON []byte
	if tx.GasFee != nil {
		gasFeeJSON, _ = json.Marshal(tx.GasFee)
	}

	// response to JSON
	var responseJSON []byte
	if tx.Response != nil {
		responseJSON, _ = json.Marshal(tx.Response)
	}

	// message to JSON
	var messagesJSON []byte
	if tx.Messages != nil {
		messagesJSON, _ = json.Marshal(tx.Messages)
	}

	// save transaction
	txRecord := map[string]interface{}{
		"hash":          tx.Hash,
		"block_height":  tx.BlockHeight,
		"tx_index":      tx.Index,
		"success":       tx.Success,
		"gas_wanted":    tx.GasWanted,
		"gas_used":      tx.GasUsed,
		"memo":          tx.Memo,
		"content_raw":   tx.ContentRaw,
		"gas_fee":       gasFeeJSON,
		"messages_json": messagesJSON,
		"response_json": responseJSON,
	}

	return r.db.WithContext(ctx).Table("indexer.transactions").Create(txRecord).Error
}

// GetTransactionByHash retrieves a transaction by its hash
func (r *postgresTransactionRepository) GetTransactionByHash(ctx context.Context, hash string) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&tx).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by hash: %w", err)
	}
	return &tx, nil
}

// GetTransactionsByBlockHeight retrieves all transactions for a specific block
func (r *postgresTransactionRepository) GetTransactionsByBlockHeight(ctx context.Context, blockHeight int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.db.WithContext(ctx).Where("block_height = ?", blockHeight).Order("tx_index").Find(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by block height: %w", err)
	}
	return transactions, nil
}
