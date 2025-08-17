package producer

import (
	"context"
	"fmt"
	"gn-indexer/internal/api"
	"gn-indexer/internal/client"
	"gn-indexer/internal/types"
	"log"

	"gn-indexer/internal/domain"
	"gn-indexer/internal/repository"
)

type Syncer struct {
	blockClient *client.GraphQLClient[types.BlocksDataArr]
	txClient    *client.GraphQLClient[types.TxsData]
	subClient   *client.SubscriptionClient

	// Use repositories instead of direct DB access
	blockRepo       repository.BlockRepository
	transactionRepo repository.TransactionRepository
}

// ... existing types ...

func NewSyncer(
	client *client.GraphQLClient[types.BlocksDataArr],
	txClient *client.GraphQLClient[types.TxsData],
	subClient *client.SubscriptionClient,
	blockRepo repository.BlockRepository,
	transactionRepo repository.TransactionRepository,
) *Syncer {
	syncer := &Syncer{
		blockClient:     client,
		txClient:        txClient,
		subClient:       subClient,
		blockRepo:       blockRepo,
		transactionRepo: transactionRepo,
	}

	return syncer
}

// SyncBlocks synchronizes blocks within a height range
func (s *Syncer) SyncBlocks(ctx context.Context, fromHeight, toHeight int) error {
	var bd types.BlocksDataArr
	if err := s.blockClient.Do(ctx, api.QBlocks, map[string]interface{}{
		"gt": fromHeight,
		"lt": toHeight,
	}, &bd); err != nil {
		return fmt.Errorf("sync blocks: %w", err)
	}

	for _, block := range bd.GetBlocks {
		if err := s.blockRepo.SaveBlock(ctx, block); err != nil {
			log.Printf("failed to save block: %v", err)
			continue
		}
	}
	log.Printf("synced %d blocks from height %d to %d", len(bd.GetBlocks), fromHeight, toHeight)
	return nil
}

// SyncTxs synchronizes transactions within a height range
func (s *Syncer) SyncTxs(ctx context.Context, fromHeight, toHeight int) error {
	var td types.TxsData
	if err := s.txClient.Do(ctx, api.QTxs, map[string]interface{}{
		"gt":   fromHeight,
		"lt":   toHeight,
		"imax": 1000,
	}, &td); err != nil {
		return fmt.Errorf("sync transactions: %w", err)
	}

	for _, tx := range td.GetTransactions {
		if err := s.transactionRepo.SaveTransaction(ctx, tx); err != nil {
			log.Printf("failed to save transaction: %v", err)
			continue
		}
	}
	log.Printf("synced %d transactions from height %d to %d", len(td.GetTransactions), fromHeight, toHeight)
	return nil
}

// GetLastSyncedHeight returns the height of the last synchronized block
func (s *Syncer) GetLastSyncedHeight(ctx context.Context) (int, error) {
	return s.blockRepo.GetLastSyncedHeight(ctx)
}

// SyncRange synchronizes both blocks and transactions within a height range
func (s *Syncer) SyncRange(ctx context.Context, fromHeight, toHeight int) error {
	// Call SyncBlocks with fromHeight-1 to include fromHeight
	if err := s.SyncBlocks(ctx, fromHeight-1, toHeight); err != nil {
		return fmt.Errorf("failed to sync blocks: %w", err)
	}
	// Call SyncTxs with fromHeight-1 to include fromHeight
	if err := s.SyncTxs(ctx, fromHeight-1, toHeight); err != nil {
		return fmt.Errorf("failed to sync transactions: %w", err)
	}
	return nil
}

// StartRealtimeSync starts real-time synchronization
func (s *Syncer) StartRealtimeSync(ctx context.Context) error {
	return nil // No longer managed by Syncer
}

// HandleRealtimeBlock processes real-time block data
func (s *Syncer) HandleRealtimeBlock(ctx context.Context, block domain.Block) error {
	// save block
	if err := s.blockRepo.SaveBlock(ctx, block); err != nil {
		return fmt.Errorf("save realtime block: %w", err)
	}

	// transaction sync
	if block.NumTxs > 0 {
		if err := s.SyncTxs(ctx, block.Height, block.Height); err != nil {
			return fmt.Errorf("sync transactions: %w", err)
		}
	}

	log.Printf("realtime sync: block %d saved", block.Height)
	return nil
}

// GetSubscriptionClient returns the subscription client
func (s *Syncer) GetSubscriptionClient() *client.SubscriptionClient {
	return s.subClient
}
