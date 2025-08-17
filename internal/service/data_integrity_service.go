package service

import (
	"context"
	"fmt"
	"gn-indexer/internal/producer"
	"log"
)

type DataIntegrityService struct {
	syncer *producer.Syncer
}

func NewDataIntegrityService(syncer *producer.Syncer) *DataIntegrityService {
	return &DataIntegrityService{
		syncer: syncer,
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
// This will re-sync all blocks and transactions from height 1 to the current highest height in DB
func (dis *DataIntegrityService) CheckAndFixDataIntegrity(ctx context.Context) error {
	log.Printf("DataIntegrityService: starting data integrity check and fix from height 1")

	dbHighestHeight, err := dis.syncer.GetLastSyncedHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get highest height from DB: %w", err)
	}

	if dbHighestHeight == 0 {
		log.Printf("DataIntegrityService: no blocks in DB, starting initial sync from height 1")
		// Perform initial sync if database is empty
		log.Printf("DataIntegrityService: performing initial sync from height 1 to 1000")
		if err := dis.syncRangeWithChunks(ctx, 1, 1000); err != nil {
			return fmt.Errorf("initial sync failed: %w", err)
		}
		log.Printf("DataIntegrityService: initial sync completed successfully")
		return nil
	} else {
		log.Printf("DataIntegrityService: DB has blocks up to height %d, will check and fix integrity from height 1", dbHighestHeight)
	}

	return dis.syncRangeWithChunks(ctx, 1, int64(dbHighestHeight))
}

// syncRangeWithChunks processes blocks in chunks to avoid GraphQL limits
func (dis *DataIntegrityService) syncRangeWithChunks(ctx context.Context, fromHeight, toHeight int64) error {
	if fromHeight > toHeight {
		log.Printf("DataIntegrityService: no range to sync (fromHeight %d > toHeight %d)", fromHeight, toHeight)
		return nil
	}

	const chunkSize = int64(1000)

	totalChunks := (toHeight - fromHeight) / chunkSize
	if (toHeight-fromHeight)%chunkSize != 0 {
		totalChunks++
	}

	log.Printf("DataIntegrityService: processing %d chunks from height %d to %d", totalChunks, fromHeight, toHeight)

	successCount := 0
	currentChunk := int64(0)

	for from := fromHeight; from <= toHeight; from += chunkSize {
		currentChunk++
		to := from + chunkSize - 1
		if to > toHeight {
			to = toHeight
		}

		log.Printf("DataIntegrityService: syncing chunk %d/%d (height %d to %d)",
			currentChunk, totalChunks, from, to)

		if err := dis.SyncRange(ctx, from, to); err != nil {
			log.Printf("DataIntegrityService: chunk %d-%d failed: %v", from, to, err)
			continue
		}

		successCount++
		log.Printf("DataIntegrityService: chunk %d/%d completed successfully", currentChunk, totalChunks)
	}

	log.Printf("DataIntegrityService: completed - %d/%d chunks successful, synced blocks %d to %d",
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
