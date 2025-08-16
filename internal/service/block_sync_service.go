package service

import (
	"context"
	"fmt"
	"gn-indexer/internal/indexer"
	"log"
	"sync"
)

// BlockSyncService coordinates backfill and real-time synchronization
type BlockSyncService struct {
	backfillSvc *BackfillService
	realtimeSvc *RealtimeSyncService
}

// NewBlockSyncService creates a new block sync service(BackfillService, RealtimeSyncService)
func NewBlockSyncService(syncer *indexer.Syncer, subClient *indexer.SubscriptionClient, wsEndpoint string) *BlockSyncService {
	return &BlockSyncService{
		backfillSvc: NewBackfillService(syncer, wsEndpoint),
		realtimeSvc: NewRealtimeSyncService(syncer, subClient),
	}
}

// StartParallelSync begins the complete synchronization process
func (bm *BlockSyncService) StartParallelSync(ctx context.Context) error {
	log.Printf("BlockSyncService: starting parallel sync process")

	// Create a new context for backfill that can be cancelled independently
	backfillCtx, cancelBackfill := context.WithCancel(ctx)
	defer cancelBackfill()

	var wg sync.WaitGroup
	var backfillErr, realtimeErr error

	// Step 1: Start real-time synchronization immediately (parallel with backfill)
	log.Printf("BlockSyncService: phase 1 - starting real-time sync (parallel)")
	wg.Add(1)
	go func() {
		defer wg.Done()
		realtimeErr = bm.realtimeSvc.Start(ctx)
		if realtimeErr != nil {
			log.Printf("BlockSyncService: real-time sync failed: %v", realtimeErr)
		}
	}()

	// Step 2: Run backfill in parallel (with separate context)
	log.Printf("BlockSyncService: phase 2 - starting backfill (parallel)")
	wg.Add(1)
	go func() {
		defer wg.Done()
		backfillErr = bm.backfillSvc.BackfillToLatest(backfillCtx)
		if backfillErr != nil {
			log.Printf("BlockSyncService: backfill failed: %v", backfillErr)
		} else {
			log.Printf("BlockSyncService: backfill completed successfully")
		}
	}()

	// Wait for both processes to complete or fail
	wg.Wait()

	// Check results
	if realtimeErr != nil {
		return fmt.Errorf("real-time sync failed: %w", realtimeErr)
	}

	if backfillErr != nil {
		log.Printf("BlockSyncService: backfill failed but real-time sync continues")
	}

	log.Printf("BlockSyncService: parallel sync process completed")
	return nil
}
