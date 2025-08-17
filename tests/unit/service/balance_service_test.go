package service_test

import (
	"context"
	"gn-indexer/internal/domain"
	"gn-indexer/internal/repository"
	"gn-indexer/internal/service"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
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

type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) Create(ctx context.Context, token *domain.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) GetByPath(ctx context.Context, path string) (*domain.Token, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Token), args.Error(1)
}

func (m *MockTokenRepository) RegisterIfNotExists(ctx context.Context, tokenPath string) error {
	args := m.Called(ctx, tokenPath)
	return args.Error(0)
}

func (m *MockTokenRepository) GetAll(ctx context.Context) ([]domain.Token, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Token), args.Error(1)
}

func TestBalanceService_ProcessMintEvent(t *testing.T) {
	// Setup
	mockBalanceRepo := new(MockBalanceRepository)
	mockTokenRepo := new(MockTokenRepository)
	balanceService := service.NewBalanceService(mockBalanceRepo, mockTokenRepo)
	ctx := context.Background()

	event := &domain.ParsedEvent{
		Type:        "MINT",
		TokenPath:   "test-token",
		ToAddress:   "test-address",
		Amount:      100,
		FromAddress: "",
	}

	// Mock expectations - MINT event flow
	mockBalanceRepo.On("GetBalance", ctx, "test-token", "test-address").Return(nil, repository.ErrBalanceNotFound)
	mockBalanceRepo.On("Update", ctx, mock.AnythingOfType("*domain.Balance")).Return(repository.ErrBalanceNotFound)
	mockBalanceRepo.On("Create", ctx, mock.AnythingOfType("*domain.Balance")).Return(nil)

	// Execute
	err := balanceService.ProcessEvent(ctx, event)

	// Assert
	assert.NoError(t, err)
	mockBalanceRepo.AssertExpectations(t)
}

func TestBalanceService_ProcessBurnEvent(t *testing.T) {
	// Setup
	mockBalanceRepo := new(MockBalanceRepository)
	mockTokenRepo := new(MockTokenRepository)
	balanceService := service.NewBalanceService(mockBalanceRepo, mockTokenRepo)
	ctx := context.Background()

	event := &domain.ParsedEvent{
		Type:        "BURN",
		TokenPath:   "test-token",
		FromAddress: "test-address",
		Amount:      30,
		ToAddress:   "",
	}

	existingBalance := &domain.Balance{
		TokenPath:  "test-token",
		Address:    "test-address",
		Amount:     domain.NewU64(100),
		LastTxHash: "test-tx",
		LastBlockH: 1000,
		UpdatedAt:  time.Now(),
	}

	// Mock expectations - BURN event flow
	mockBalanceRepo.On("GetBalance", ctx, "test-token", "test-address").Return(existingBalance, nil)
	mockBalanceRepo.On("Update", ctx, mock.AnythingOfType("*domain.Balance")).Return(nil)

	// Execute
	err := balanceService.ProcessEvent(ctx, event)

	// Assert
	assert.NoError(t, err)
	mockBalanceRepo.AssertExpectations(t)
}

func TestBalanceService_ProcessTransferEvent(t *testing.T) {
	// Setup
	mockBalanceRepo := new(MockBalanceRepository)
	mockTokenRepo := new(MockTokenRepository)
	balanceService := service.NewBalanceService(mockBalanceRepo, mockTokenRepo)
	ctx := context.Background()

	event := &domain.ParsedEvent{
		Type:        "TRANSFER",
		TokenPath:   "test-token",
		FromAddress: "from-address",
		ToAddress:   "to-address",
		Amount:      50,
	}

	fromBalance := &domain.Balance{
		TokenPath:  "test-token",
		Address:    "from-address",
		Amount:     domain.NewU64(200),
		LastTxHash: "test-tx",
		LastBlockH: 1000,
		UpdatedAt:  time.Now(),
	}

	// Mock expectations - TRANSFER event flow
	// From address (decrease balance)
	mockBalanceRepo.On("GetBalance", ctx, "test-token", "from-address").Return(fromBalance, nil)
	mockBalanceRepo.On("Update", ctx, mock.AnythingOfType("*domain.Balance")).Return(nil)

	// To address (increase balance) - Update succeeds
	mockBalanceRepo.On("GetBalance", ctx, "test-token", "to-address").Return(nil, repository.ErrBalanceNotFound)
	mockBalanceRepo.On("Update", ctx, mock.AnythingOfType("*domain.Balance")).Return(nil)

	// Execute
	err := balanceService.ProcessEvent(ctx, event)

	// Assert
	assert.NoError(t, err)
	mockBalanceRepo.AssertExpectations(t)
}
