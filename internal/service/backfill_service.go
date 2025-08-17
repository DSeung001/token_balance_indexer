package service

import (
	"context"
	"fmt"
	"gn-indexer/internal/client"
	"gn-indexer/internal/producer"
	"gn-indexer/internal/types"
	"log"
)

// BackfillService handles backfilling of missing blockchain data
type BackfillService struct {
	syncer     *producer.Syncer
	wsEndpoint string
}

// NewBackfillService creates a new backfill service
func NewBackfillService(syncer *producer.Syncer, wsEndpoint string) *BackfillService {
	return &BackfillService{
		syncer:     syncer,
		wsEndpoint: wsEndpoint,
	}
}

// BackfillToLatest performs backfill to catch up with the latest block
func (bs *BackfillService) BackfillToLatest(ctx context.Context) error {
	// Get last saved block height from DB
	lastHeight, err := bs.syncer.GetLastSyncedHeight(ctx)
	if err != nil {
		return fmt.Errorf("get last synced height: %w", err)
	}

	log.Printf("BackfillService: starting backfill from height %d", lastHeight)

	// Create a separate websocket client for backfill
	backfillSubClient := client.NewSubscriptionClient(bs.wsEndpoint)
	defer backfillSubClient.Close()

	// Get current block height from websocket (one-time subscription)
	var currentHeight int
	err = backfillSubClient.SubscribeOnce(ctx, producer.SBlocks, nil, func(data types.BlocksData) error {
		currentHeight = data.GetBlocks.Height
		return nil
	})

	if err != nil {
		return fmt.Errorf("get current block height: %w", err)
	}

	if currentHeight < 1 {
		return fmt.Errorf("invalid current block height: %d", currentHeight)
	}

	// Check if backfill is needed
	if lastHeight >= currentHeight {
		log.Printf("BackfillService: no backfill needed (last=%d, current=%d)", lastHeight, currentHeight)
		return nil
	}

	// Sync missing blocks in smaller chunks to avoid GraphQL limits
	lastHeightPlus1 := lastHeight + 1
	log.Printf("BackfillService: backfilling from height %d to %d", lastHeightPlus1, currentHeight)

	// Use smaller chunks (e.g., 1000 blocks at a time)
	const chunkSize = 1000

	successCount := 0
	totalChunks := 0

	for from := lastHeightPlus1; from <= currentHeight; from += chunkSize {
		to := from + chunkSize - 1
		if to > currentHeight {
			to = currentHeight
		}

		totalChunks++
		log.Printf("BackfillService: syncing chunk %d/%d (height %d to %d)",
			totalChunks, (currentHeight-lastHeight)/chunkSize+1, from, to)

		if err := bs.syncer.SyncRange(ctx, from, to); err != nil {
			log.Printf("BackfillService: chunk %d-%d failed: %v", from, to, err)
			continue
		}

		successCount++
	}

	// Summary log
	log.Printf("BackfillService: completed - %d/%d chunks successful, synced blocks %d to %d",
		successCount, totalChunks, lastHeightPlus1, currentHeight)

	if successCount > 0 {
		return nil
	}

	return fmt.Errorf("all chunks failed during backfill")
}
