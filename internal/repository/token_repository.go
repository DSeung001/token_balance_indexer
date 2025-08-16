package repository

import (
	"context"
	"gn-indexer/internal/domain"
	"gorm.io/gorm"
)

// TokenRepository handles token data persistence
type TokenRepository interface {
	Create(ctx context.Context, token *domain.Token) error
	GetByPath(ctx context.Context, tokenPath string) (*domain.Token, error)
	RegisterIfNotExists(ctx context.Context, tokenPath string) error
	GetAll(ctx context.Context) ([]domain.Token, error)
}

type postgresTokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new PostgreSQL token repository
func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &postgresTokenRepository{db: db}
}

// Create saves a token to the tokens table
func (r *postgresTokenRepository) Create(ctx context.Context, token *domain.Token) error {
	return r.db.WithContext(ctx).Table("indexer.tokens").Create(token).Error
}

// GetByPath retrieves a token by its path
func (r *postgresTokenRepository) GetByPath(ctx context.Context, tokenPath string) (*domain.Token, error) {
	var token domain.Token
	err := r.db.WithContext(ctx).
		Where("token_path = ?", tokenPath).
		First(&token).Error

	if err != nil {
		return nil, err
	}

	return &token, nil
}

// RegisterIfNotExists registers a token if it doesn't exist
func (r *postgresTokenRepository) RegisterIfNotExists(ctx context.Context, tokenPath string) error {
	// Check if token already exists
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Token{}).
		Where("token_path = ?", tokenPath).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	// Create new token with default values
	token := &domain.Token{
		Path:     tokenPath,
		Symbol:   "", // Todo: updatable
		Decimals: 0,  // Todo: updatable
	}

	return r.Create(ctx, token)
}

// GetAll retrieves all tokens
func (r *postgresTokenRepository) GetAll(ctx context.Context) ([]domain.Token, error) {
	var tokens []domain.Token
	err := r.db.WithContext(ctx).Find(&tokens).Error
	return tokens, err
}
