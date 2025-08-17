package repository_test

import (
	"context"
	"gn-indexer/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository for testing
type MockBalanceRepository struct {
	mock.Mock
}

func (m *MockBalanceRepository) Create(ctx context.Context, balance *domain.Balance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

func (m *MockBalanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

func (m *MockBalanceRepository) GetBalance(ctx context.Context, tokenPath, address string) (*domain.Balance, error) {
	args := m.Called(ctx, tokenPath, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Balance), args.Error(1)
}

func (m *MockBalanceRepository) GetBalancesByAddress(ctx context.Context, address string) ([]*domain.Balance, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Balance), args.Error(1)
}

func (m *MockBalanceRepository) GetBalancesByTokenAndAddress(ctx context.Context, tokenPath string) ([]*domain.Balance, error) {
	args := m.Called(ctx, tokenPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Balance), args.Error(1)
}

func (m *MockBalanceRepository) GetAllBalances(ctx context.Context) ([]*domain.Balance, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Balance), args.Error(1)
}

func TestBalanceRepository_Create(t *testing.T) {
	// Setup
	mockRepo := new(MockBalanceRepository)
	ctx := context.Background()

	balance := &domain.Balance{
		TokenPath:  "test-token",
		Address:    "test-address",
		Amount:     domain.NewU64(100),
		LastTxHash: "test-tx",
		LastBlockH: 1000,
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("Create", ctx, balance).Return(nil)

	// Execute
	err := mockRepo.Create(ctx, balance)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBalanceRepository_GetBalance(t *testing.T) {
	// Setup
	mockRepo := new(MockBalanceRepository)
	ctx := context.Background()

	expectedBalance := &domain.Balance{
		TokenPath:  "test-token",
		Address:    "test-address",
		Amount:     domain.NewU64(100),
		LastTxHash: "test-tx",
		LastBlockH: 1000,
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("GetBalance", ctx, "test-token", "test-address").
		Return(expectedBalance, nil)

	// Execute
	retrieved, err := mockRepo.GetBalance(ctx, "test-token", "test-address")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, retrieved)
	mockRepo.AssertExpectations(t)
}

func TestBalanceRepository_Update(t *testing.T) {
	// Setup
	mockRepo := new(MockBalanceRepository)
	ctx := context.Background()

	balance := &domain.Balance{
		TokenPath:  "test-token",
		Address:    "test-address",
		Amount:     domain.NewU64(200),
		LastTxHash: "updated-tx",
		LastBlockH: 1001,
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("Update", ctx, balance).Return(nil)

	// Execute
	err := mockRepo.Update(ctx, balance)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBalanceRepository_GetBalancesByAddress(t *testing.T) {
	// Setup
	mockRepo := new(MockBalanceRepository)
	ctx := context.Background()

	expectedBalances := []*domain.Balance{
		{
			TokenPath:  "token-1",
			Address:    "test-address",
			Amount:     domain.NewU64(100),
			LastTxHash: "tx-1",
			LastBlockH: 1000,
			UpdatedAt:  time.Now(),
		},
		{
			TokenPath:  "token-2",
			Address:    "test-address",
			Amount:     domain.NewU64(200),
			LastTxHash: "tx-2",
			LastBlockH: 1001,
			UpdatedAt:  time.Now(),
		},
	}

	// Mock expectations
	mockRepo.On("GetBalancesByAddress", ctx, "test-address").
		Return(expectedBalances, nil)

	// Execute
	balances, err := mockRepo.GetBalancesByAddress(ctx, "test-address")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, balances, 2)
	assert.Equal(t, expectedBalances, balances)
	mockRepo.AssertExpectations(t)
}
