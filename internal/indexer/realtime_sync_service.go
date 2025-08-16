package indexer

import (
	"context"
	"fmt"

	"gn-indexer/internal/domain"
)

type RealtimeSyncService struct {
	syncer    *Syncer
	subClient *SubscriptionClient
}

func (rs *RealtimeSyncService) Start(ctx context.Context) error {
	backfillSvc := &BackfillService{syncer: rs.syncer}
	if err := backfillSvc.BackfillToLatest(ctx); err != nil {
		return fmt.Errorf("BackfillToLatest: %w", err)
	}

	return rs.startSubscription(ctx)
}

func (rs *RealtimeSyncService) startSubscription(ctx context.Context) error {
	return rs.subClient.Subscribe(ctx, SBlocks, nil, rs.handleSubscriptionData)
}

func (rs *RealtimeSyncService) handleSubscriptionData(data BlocksData) error {
	for _, block := range data.GetBlocks {
		if err := rs.processBlock(block); err != nil {
			return fmt.Errorf("processBlock: %w", err)
		}
	}
	return nil
}

func (rs *RealtimeSyncService) processBlock(block domain.Block) error {
	// block processing
	return rs.syncer.handleRealtimeBlock(context.Background(), block)
}
