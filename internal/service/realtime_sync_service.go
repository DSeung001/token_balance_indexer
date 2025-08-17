package service

import (
	"context"
	"fmt"
	"gn-indexer/internal/api"
	"gn-indexer/internal/client"
	"gn-indexer/internal/domain"
	"gn-indexer/internal/producer"
	"gn-indexer/internal/types"
	"log"
	"time"
)

// RealtimeSyncService handles real-time blockchain data synchronization
type RealtimeSyncService struct {
	syncer    *producer.Syncer
	subClient *client.SubscriptionClient
}

// NewRealtimeSyncService creates a new realtime sync service
func NewRealtimeSyncService(syncer *producer.Syncer, subClient *client.SubscriptionClient) *RealtimeSyncService {
	return &RealtimeSyncService{
		syncer:    syncer,
		subClient: subClient,
	}
}

// Start begins the real-time synchronization process
func (rs *RealtimeSyncService) Start(ctx context.Context) error {
	log.Printf("RealtimeSyncService: starting real-time sync")

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Start real-time subscription for new blocks
	return rs.startSubscription(ctx)
}

// startSubscription starts the websocket subscription
func (rs *RealtimeSyncService) startSubscription(ctx context.Context) error {
	log.Printf("RealtimeSyncService: starting subscription")

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Start the subscription
	err := rs.subClient.Subscribe(ctx, api.SBlocks, nil, rs.handleSubscriptionData)
	if err != nil {
		return fmt.Errorf("failed to start subscription: %w", err)
	}

	log.Printf("RealtimeSyncService: subscription started successfully")
	return nil
}

// handleSubscriptionData processes incoming real-time block data
func (rs *RealtimeSyncService) handleSubscriptionData(data types.BlocksData) error {
	block := data.GetBlocks
	log.Printf("RealtimeSyncService: processing block %d", block.Height)

	// Process real-time blocks with retry logic
	if err := rs.processBlockWithRetry(block); err != nil {
		log.Printf("RealtimeSyncService: block %d failed after retries: %v", block.Height, err)
		return fmt.Errorf("block %d processing failed: %w", block.Height, err)
	}

	log.Printf("RealtimeSyncService: block %d processed successfully", block.Height)
	return nil
}

// processBlockWithRetry processes the block with retry logic
func (rs *RealtimeSyncService) processBlockWithRetry(block domain.Block) error {
	const maxRetries = 3
	const retryDelay = time.Millisecond * 500

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := rs.processBlock(block); err != nil {
			if attempt == maxRetries {
				return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
			}

			log.Printf("RealtimeSyncService: attempt %d failed for block %d: %v, retrying...",
				attempt, block.Height, err)
			time.Sleep(retryDelay)
			continue
		}

		return nil
	}

	return fmt.Errorf("unexpected error in retry loop")
}

// processBlock processes a single block
func (rs *RealtimeSyncService) processBlock(block domain.Block) error {
	return rs.syncer.HandleRealtimeBlock(context.Background(), block)
}
