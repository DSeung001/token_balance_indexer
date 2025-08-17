package producer

import (
	"context"
	"fmt"
	"gn-indexer/internal/client"
	"gn-indexer/internal/types"
	"log"

	"gn-indexer/internal/domain"
	"gn-indexer/internal/repository"
)

// EventProcessor defines the interface for processing transactions
type EventProcessor interface {
	ProcessTransaction(ctx context.Context, tx *domain.Transaction) error
}

type Syncer struct {
	blockClient *client.GraphQLClient[types.BlocksDataArr]
	txClient    *client.GraphQLClient[types.TxsData]
	subClient   *client.SubscriptionClient

	// Use repositories instead of direct DB access
	blockRepo       repository.BlockRepository
	transactionRepo repository.TransactionRepository

	// Use interface instead of concrete type to avoid circular import
	eventProcessor EventProcessor
}

func NewSyncer(
	client *client.GraphQLClient[types.BlocksDataArr],
	txClient *client.GraphQLClient[types.TxsData],
	subClient *client.SubscriptionClient,
	blockRepo repository.BlockRepository,
	transactionRepo repository.TransactionRepository,
	eventProcessor EventProcessor,
) *Syncer {
	syncer := &Syncer{
		blockClient:     client,
		txClient:        txClient,
		subClient:       subClient,
		blockRepo:       blockRepo,
		transactionRepo: transactionRepo,
		eventProcessor:  eventProcessor,
	}

	return syncer
}

// SyncBlocks synchronizes blocks within a height range
func (s *Syncer) SyncBlocks(ctx context.Context, fromHeight, toHeight int) error {
	var bd types.BlocksDataArr
	if err := s.blockClient.Do(ctx, QBlocks, map[string]interface{}{
		"gt": fromHeight - 1, // fromHeight-1보다 큰 값 = fromHeight부터
		"lt": toHeight + 1,   // toHeight+1보다 작은 값 = toHeight까지
	}, &bd); err != nil {
		return fmt.Errorf("sync blocks: %w", err)
	}

	for _, block := range bd.GetBlocks {
		if err := s.blockRepo.SaveBlock(ctx, block); err != nil {
			log.Printf("failed to save block: %v", err)
			continue
		}
	}
	// 실제 저장된 블록의 높이 범위 계산
	if len(bd.GetBlocks) > 0 {
		minHeight := bd.GetBlocks[0].Height
		maxHeight := bd.GetBlocks[len(bd.GetBlocks)-1].Height
		log.Printf("synced %d blocks from height %d to %d", len(bd.GetBlocks), minHeight, maxHeight)
	} else {
		log.Printf("synced 0 blocks (no blocks in range %d to %d)", fromHeight+1, toHeight)
	}
	return nil
}

// SyncTxs synchronizes transactions within a height range
func (s *Syncer) SyncTxs(ctx context.Context, fromHeight, toHeight int) error {
	var td types.TxsData
	if err := s.txClient.Do(ctx, QTxs, map[string]interface{}{
		"gt":   fromHeight - 1, // fromHeight-1보다 큰 값 = fromHeight부터
		"lt":   toHeight + 1,   // toHeight+1보다 작은 값 = toHeight까지
		"imax": 1000,
	}, &td); err != nil {
		return fmt.Errorf("sync transactions: %w", err)
	}

	for _, tx := range td.GetTransactions {
		if err := s.transactionRepo.SaveTransaction(ctx, tx); err != nil {
			log.Printf("failed to save transaction: %v", err)
			continue
		}

		// Process transaction for events and send to SQS queue
		if s.eventProcessor != nil {
			// tx is already domain.Transaction type, convert to pointer
			if err := s.eventProcessor.ProcessTransaction(ctx, &tx); err != nil {
				log.Printf("failed to process transaction events: %v", err)
				// Don't fail the sync if event processing fails
			}
		}
	}
	// 실제 저장된 트랜잭션의 높이 범위 계산
	if len(td.GetTransactions) > 0 {
		minHeight := td.GetTransactions[0].BlockHeight
		maxHeight := td.GetTransactions[len(td.GetTransactions)-1].BlockHeight
		log.Printf("synced %d transactions from height %d to %d", len(td.GetTransactions), minHeight, maxHeight)
	} else {
		log.Printf("synced 0 transactions (no transactions in range %d to %d)", fromHeight+1, toHeight)
	}
	return nil
}

// GetLastSyncedHeight returns the height of the last synchronized block
func (s *Syncer) GetLastSyncedHeight(ctx context.Context) (int, error) {
	return s.blockRepo.GetLastSyncedHeight(ctx)
}

// SyncRange synchronizes both blocks and transactions within a height range
func (s *Syncer) SyncRange(ctx context.Context, fromHeight, toHeight int) error {
	// Call SyncBlocks with fromHeight to toHeight (inclusive)
	if err := s.SyncBlocks(ctx, fromHeight, toHeight); err != nil {
		return fmt.Errorf("failed to sync blocks: %w", err)
	}
	// Call SyncTxs with fromHeight to toHeight (inclusive)
	if err := s.SyncTxs(ctx, fromHeight, toHeight); err != nil {
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
