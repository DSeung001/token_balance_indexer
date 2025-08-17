package service

import (
	"context"
	"fmt"
	"gn-indexer/internal/client"
	"gn-indexer/internal/producer"
	"gn-indexer/internal/types"
	"log"
)

type DataIntegrityService struct {
	syncer    *producer.Syncer
	subClient *client.SubscriptionClient
}

func NewDataIntegrityService(syncer *producer.Syncer, subClient *client.SubscriptionClient) *DataIntegrityService {
	return &DataIntegrityService{
		syncer:    syncer,
		subClient: subClient,
	}
}

// SyncRange performs synchronization for a specific height range
func (dis *DataIntegrityService) SyncRange(ctx context.Context, fromHeight, toHeight int64) error {
	log.Printf("DataIntegrityService: syncing range %d to %d", fromHeight, toHeight)

	if err := dis.syncer.SyncRange(ctx, int(fromHeight), int(toHeight)); err != nil {
		return fmt.Errorf("sync range failed: %w", err)
	}

	log.Printf("DataIntegrityService: range %d to %d completed", fromHeight, toHeight)
	return nil
}

// CheckAndFixDataIntegrity performs data integrity check and fixes missing blocks/transactions
// This will re-sync all blocks and transactions from height 1 to the current latest height from network
func (dis *DataIntegrityService) CheckAndFixDataIntegrity(ctx context.Context) error {
	log.Printf("DataIntegrityService: starting data integrity check and fix from height 1")

	// Get the latest block height from network using SubscribeOnce (same as backfill service)
	var latestNetworkHeight int
	err := dis.subClient.SubscribeOnce(ctx, producer.SBlocks, nil, func(data types.BlocksData) error {
		latestNetworkHeight = data.GetBlocks.Height
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get latest network height: %w", err)
	}

	if latestNetworkHeight < 1 {
		return fmt.Errorf("invalid network height: %d", latestNetworkHeight)
	}

	log.Printf("DataIntegrityService: latest network height: %d, will sync entire range from height 1 to %d", latestNetworkHeight, latestNetworkHeight)
	return dis.syncRangeWithChunks(ctx, 1, int64(latestNetworkHeight))
}

// syncRangeWithChunks processes blocks in chunks to avoid GraphQL limits
func (dis *DataIntegrityService) syncRangeWithChunks(ctx context.Context, fromHeight, toHeight int64) error {
	if fromHeight > toHeight {
		log.Printf("DataIntegrityService: no range to sync (fromHeight %d > toHeight %d)", fromHeight, toHeight)
		return nil
	}

	const chunkSize = int64(1000)

	// Calculate total chunks
	var totalChunks int64
	if fromHeight == toHeight {
		totalChunks = 1
	} else {
		totalChunks = (toHeight - fromHeight) / chunkSize
		if (toHeight-fromHeight)%chunkSize != 0 {
			totalChunks++
		}
	}

	log.Printf("DataIntegrityService: processing %d chunks from height %d to %d", totalChunks, fromHeight, toHeight)

	successCount := int64(0)
	currentChunk := int64(0)

	for from := fromHeight; from <= toHeight; from += chunkSize {
		currentChunk++
		to := from + chunkSize - 1
		if to > toHeight {
			to = toHeight
		}

		log.Printf("DataIntegrityService: syncing chunk %d/%d (height %d~%d)",
			currentChunk, totalChunks, from, to)

		if err := dis.SyncRange(ctx, from, to); err != nil {
			log.Printf("DataIntegrityService: chunk %d-%d failed: %v", from, to, err)
			continue
		}

		successCount++
		log.Printf("DataIntegrityService: chunk %d/%d completed successfully", currentChunk, totalChunks)
	}

	log.Printf("DataIntegrityService: completed - %d/%d chunks successful, synced blocks %d~%d",
		successCount, totalChunks, fromHeight, toHeight)

	if successCount > 0 {
		return nil
	}

	return fmt.Errorf("all chunks failed during data integrity check")
}

// SyncSpecificRange syncs a specific height range (for manual control)
func (dis *DataIntegrityService) SyncSpecificRange(ctx context.Context, fromHeight, toHeight int64) error {
	log.Printf("DataIntegrityService: syncing specific range %d to %d", fromHeight, toHeight)

	if fromHeight < 1 {
		return fmt.Errorf("fromHeight must be >= 1, got %d", fromHeight)
	}

	if toHeight < fromHeight {
		return fmt.Errorf("toHeight must be >= fromHeight, got %d < %d", toHeight, fromHeight)
	}

	return dis.syncRangeWithChunks(ctx, fromHeight, toHeight)
}
