package service

import (
	"context"
	"fmt"
	"gn-indexer/internal/domain"
	"gn-indexer/internal/repository"
	"log"
)

// BalanceService handles token balance calculations
type BalanceService struct {
	balanceRepo repository.BalanceRepository
	tokenRepo   repository.TokenRepository
}

// NewBalanceService creates a new balance service
func NewBalanceService(
	balanceRepo repository.BalanceRepository,
	tokenRepo repository.TokenRepository,
) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		tokenRepo:   tokenRepo,
	}
}

// ProcessEvent processes a parsed event and updates balances
func (bs *BalanceService) ProcessEvent(ctx context.Context, event *domain.ParsedEvent) error {
	log.Printf("BalanceService: processing event %s for token %s", event.Type, event.TokenPath)

	// Ensure token exists
	if err := bs.tokenRepo.RegisterIfNotExists(ctx, event.TokenPath); err != nil {
		return fmt.Errorf("register token: %w", err)
	}

	switch event.Type {
	case "Mint":
		return bs.processMintEvent(ctx, event)
	case "Burn":
		return bs.processBurnEvent(ctx, event)
	case "Transfer":
		return bs.processTransferEvent(ctx, event)
	default:
		log.Printf("BalanceService: unknown event type %s, skipping", event.Type)
		return nil
	}
}

// processMintEvent handles token mint events
func (bs *BalanceService) processMintEvent(ctx context.Context, event *domain.ParsedEvent) error {
	log.Printf("BalanceService: processing mint event for %s to %s", event.TokenPath, event.ToAddress)

	// Mint: increase balance for 'to' address
	if err := bs.updateBalance(ctx, event.TokenPath, event.ToAddress, event.Amount, true); err != nil {
		return fmt.Errorf("update balance for mint: %w", err)
	}

	log.Printf("BalanceService: mint event processed successfully")
	return nil
}

// processBurnEvent handles token burn events
func (bs *BalanceService) processBurnEvent(ctx context.Context, event *domain.ParsedEvent) error {
	log.Printf("BalanceService: processing burn event for %s from %s", event.TokenPath, event.FromAddress)

	// Burn: decrease balance for 'from' address
	if err := bs.updateBalance(ctx, event.TokenPath, event.FromAddress, event.Amount, false); err != nil {
		return fmt.Errorf("update balance for burn: %w", err)
	}

	log.Printf("BalanceService: burn event processed successfully")
	return nil
}

// processTransferEvent handles token transfer events
func (bs *BalanceService) processTransferEvent(ctx context.Context, event *domain.ParsedEvent) error {
	log.Printf("BalanceService: processing transfer event for %s from %s to %s",
		event.TokenPath, event.FromAddress, event.ToAddress)

	// Transfer: decrease balance for 'from' address and increase for 'to' address
	if err := bs.updateBalance(ctx, event.TokenPath, event.FromAddress, event.Amount, false); err != nil {
		return fmt.Errorf("update balance for transfer from: %w", err)
	}

	if err := bs.updateBalance(ctx, event.TokenPath, event.ToAddress, event.Amount, true); err != nil {
		return fmt.Errorf("update balance for transfer to: %w", err)
	}

	log.Printf("BalanceService: transfer event processed successfully")
	return nil
}

// updateBalance updates the balance for a specific token and address
func (bs *BalanceService) updateBalance(ctx context.Context, tokenPath, address string, amount int64, isIncrease bool) error {
	// Get current balance
	currentBalance, err := bs.balanceRepo.GetBalance(ctx, tokenPath, address)
	if err != nil {
		// If balance doesn't exist, create with 0
		if err == repository.ErrBalanceNotFound {
			currentBalance = &domain.Balance{
				TokenPath: tokenPath,
				Address:   address,
				Amount:    0,
			}
		} else {
			return fmt.Errorf("get current balance: %w", err)
		}
	}

	// Calculate new balance
	var newAmount int64
	if isIncrease {
		newAmount = currentBalance.Amount + amount
	} else {
		newAmount = currentBalance.Amount - amount
		// Ensure balance doesn't go negative
		if newAmount < 0 {
			log.Printf("BalanceService: warning - balance would go negative for %s %s, setting to 0", tokenPath, address)
			newAmount = 0
		}
	}

	// Update or create balance
	balance := &domain.Balance{
		TokenPath: tokenPath,
		Address:   address,
		Amount:    newAmount,
	}

	// Try to update first, if it fails (not found), create new
	if err := bs.balanceRepo.Update(ctx, balance); err != nil {
		if err == repository.ErrBalanceNotFound {
			// Create new balance
			if err := bs.balanceRepo.Create(ctx, balance); err != nil {
				return fmt.Errorf("create balance: %w", err)
			}
			log.Printf("BalanceService: created new balance for %s %s: %d", tokenPath, address, newAmount)
		} else {
			return fmt.Errorf("update balance: %w", err)
		}
	} else {
		log.Printf("BalanceService: updated balance for %s %s: %d -> %d", tokenPath, address, currentBalance.Amount, newAmount)
	}

	return nil
}
