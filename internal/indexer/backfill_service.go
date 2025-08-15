package indexer

import (
	"context"
	"fmt"
	"log"
)

type BackfillService struct {
	syncer *Syncer
}

func (bs *BackfillService) BackfillToLatest(ctx context.Context) error {
	// last saved block height
	lastHeight, err := bs.syncer.GetLastSyncedHeight(ctx)
	if err != nil {
		return fmt.Errorf("get last synced height: %w", err)
	}

	// look up current block height
	currentHeight, err := bs.getCurrentBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("get current block height: %w", err)
	}

	// Run only if you need a backfill
	if lastHeight >= currentHeight {
		log.Printf("no backfill needed: last=%d, current=%d", lastHeight, currentHeight)
		return nil
	}

	// SyncRange method apply
	lastHeightPlus1 := lastHeight + 1
	log.Printf("backfilling from height %d to %d", lastHeightPlus1, currentHeight)
	return bs.syncer.SyncRange(ctx, lastHeightPlus1, currentHeight)
}

func (bs *BackfillService) getCurrentBlockHeight(ctx context.Context) (int, error) {
	var currentHeight int

	err := bs.syncer.subClient.SubcribeOnce(ctx, SBlocks, nil, func(data BlocksData) error {
		if len(data.GetBlocks) > 0 {
			currentHeight = data.GetBlocks[0].Height
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("get current block height from subscription: %w", err)
	}

	if currentHeight < 1 {
		return 0, fmt.Errorf("get current block height from subscription: %d < 1", currentHeight)
	}

	return currentHeight, nil
}
