package indexer

import (
	"context"
	"fmt"
	"gn-indexer/internal/api"
	"log"
	"time"

	"gn-indexer/internal/domain"
)

type RealtimeSyncService struct {
	syncer    *Syncer
	subClient *SubscriptionClient
}

func (rs *RealtimeSyncService) Start(ctx context.Context) error {
	// Check if context is already cancelled
	if ctx.Err() != nil {
		return ctx.Err()
	}

	backfillSvc := &BackfillService{syncer: rs.syncer}
	if err := backfillSvc.BackfillToLatest(ctx); err != nil {
		return fmt.Errorf("BackfillToLatest: %w", err)
	}

	return rs.startSubscription(ctx)
}

func (rs *RealtimeSyncService) startSubscription(ctx context.Context) error {
	// Check if context is already cancelled
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return rs.subClient.Subscribe(ctx, api.SBlocks, nil, rs.handleSubscriptionData)
}

func (rs *RealtimeSyncService) handleSubscriptionData(data BlocksData) error {
	for _, block := range data.GetBlocks {
		// Blocking with retry logic
		if err := rs.processBlockWithRetry(block); err != nil {
			log.Printf("failed to process block %d after retries: %v", block.Height, err)
			// Return error if still failed (subscription stopped)
			// Todo : Advanced, queue and then analyze again
			return fmt.Errorf("block %d processing failed: %w", block.Height, err)
		}
	}
	return nil
}

// processBlockWithRetry Processing the block as it retries
func (rs *RealtimeSyncService) processBlockWithRetry(block domain.Block) error {
	const maxRetries = 3
	const retryDelay = time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := rs.processBlock(block); err != nil {
			if attempt == maxRetries {
				return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
			}

			log.Printf("attempt %d failed for block %d: %v, retrying in %v...",
				attempt, block.Height, err, retryDelay)

			time.Sleep(retryDelay)
			continue
		}

		// Return immediately if successful
		log.Printf("block %d processed successfully on attempt %d", block.Height, attempt)
		return nil
	}

	return fmt.Errorf("unexpected error in retry loop")
}

func (rs *RealtimeSyncService) processBlock(block domain.Block) error {
	// block processing
	return rs.syncer.handleRealtimeBlock(context.Background(), block)
}
