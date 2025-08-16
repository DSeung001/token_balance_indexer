package service

import (
	"context"
	"fmt"
	"gn-indexer/internal/indexer"
	"log"
	"sync"
)

// SyncOrchestrator coordinates backfill and real-time synchronization
type SyncOrchestrator struct {
	backfillSvc *BackfillService
	realtimeSvc *RealtimeSyncService
}

// NewSyncOrchestrator creates a new sync orchestrator
func NewSyncOrchestrator(syncer *indexer.Syncer, subClient *indexer.SubscriptionClient, wsEndpoint string) *SyncOrchestrator {
	return &SyncOrchestrator{
		backfillSvc: NewBackfillService(syncer, wsEndpoint),
		realtimeSvc: NewRealtimeSyncService(syncer, subClient),
	}
}

// StartOrchestratedSync begins the complete synchronization process
func (so *SyncOrchestrator) StartOrchestratedSync(ctx context.Context) error {
	log.Printf("SyncOrchestrator: starting orchestrated sync process")

	// Create a new context for backfill that can be cancelled independently
	backfillCtx, cancelBackfill := context.WithCancel(ctx)
	defer cancelBackfill()

	var wg sync.WaitGroup
	var backfillErr, realtimeErr error

	// Step 1: Start real-time synchronization immediately (parallel with backfill)
	log.Printf("SyncOrchestrator: phase 1 - starting real-time sync (parallel)")
	wg.Add(1)
	go func() {
		defer wg.Done()
		realtimeErr = so.realtimeSvc.Start(ctx)
		if realtimeErr != nil {
			log.Printf("SyncOrchestrator: real-time sync failed: %v", realtimeErr)
		}
	}()

	// Step 2: Run backfill in parallel (with separate context)
	log.Printf("SyncOrchestrator: phase 2 - starting backfill (parallel)")
	wg.Add(1)
	go func() {
		defer wg.Done()
		backfillErr = so.backfillSvc.BackfillToLatest(backfillCtx)
		if backfillErr != nil {
			log.Printf("SyncOrchestrator: backfill failed: %v", backfillErr)
		} else {
			log.Printf("SyncOrchestrator: backfill completed successfully")
		}
	}()

	// Wait for both processes to complete or fail
	wg.Wait()

	// Check results
	if realtimeErr != nil {
		return fmt.Errorf("real-time sync failed: %w", realtimeErr)
	}

	if backfillErr != nil {
		log.Printf("SyncOrchestrator: backfill failed but real-time sync continues")
	}

	log.Printf("SyncOrchestrator: orchestrated sync process completed")
	return nil
}
