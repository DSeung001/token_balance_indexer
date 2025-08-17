package repository

import (
	"context"
	"errors"
	"gn-indexer/internal/domain"

	"gorm.io/gorm"
)

// ErrBalanceNotFound is returned when a balance is not found
var ErrBalanceNotFound = errors.New("balance not found")

// BalanceRepository defines the interface for balance operations
type BalanceRepository interface {
	Create(ctx context.Context, balance *domain.Balance) error
	Update(ctx context.Context, balance *domain.Balance) error
	GetBalance(ctx context.Context, tokenPath, address string) (*domain.Balance, error)
	GetBalancesByAddress(ctx context.Context, address string) ([]*domain.Balance, error)
	GetBalancesByToken(ctx context.Context, tokenPath string) ([]*domain.Balance, error)
	GetAllBalances(ctx context.Context) ([]*domain.Balance, error)
}

// balanceRepository implements BalanceRepository
type balanceRepository struct {
	db *gorm.DB
}

// NewBalanceRepository creates a new balance repository
func NewBalanceRepository(db *gorm.DB) BalanceRepository {
	return &balanceRepository{db: db}
}

// Create creates a new balance record
func (r *balanceRepository) Create(ctx context.Context, balance *domain.Balance) error {
	result := r.db.WithContext(ctx).Create(balance)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Update updates an existing balance record
func (r *balanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	result := r.db.WithContext(ctx).Save(balance)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetBalance gets a balance for a specific token and address
func (r *balanceRepository) GetBalance(ctx context.Context, tokenPath, address string) (*domain.Balance, error) {
	var balance domain.Balance
	result := r.db.WithContext(ctx).
		Where("token_path = ? AND address = ?", tokenPath, address).
		First(&balance)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrBalanceNotFound
		}
		return nil, result.Error
	}

	return &balance, nil
}

// GetBalancesByAddress gets all balances for a specific address
func (r *balanceRepository) GetBalancesByAddress(ctx context.Context, address string) ([]*domain.Balance, error) {
	var balances []*domain.Balance
	result := r.db.WithContext(ctx).
		Where("address = ?", address).
		Find(&balances)

	if result.Error != nil {
		return nil, result.Error
	}

	return balances, nil
}

// GetBalancesByToken gets all balances for a specific token
func (r *balanceRepository) GetBalancesByToken(ctx context.Context, tokenPath string) ([]*domain.Balance, error) {
	var balances []*domain.Balance
	result := r.db.WithContext(ctx).
		Where("token_path = ?", tokenPath).
		Find(&balances)

	if result.Error != nil {
		return nil, result.Error
	}

	return balances, nil
}

// GetAllBalances gets all balances
func (r *balanceRepository) GetAllBalances(ctx context.Context) ([]*domain.Balance, error) {
	var balances []*domain.Balance
	result := r.db.WithContext(ctx).Find(&balances)

	if result.Error != nil {
		return nil, result.Error
	}

	return balances, nil
}
