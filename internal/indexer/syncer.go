package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type Syncer struct {
	blockClient *GraphQLClient[BlocksData]
	txClient    *GraphQLClient[TxsData]
	subClient   *SubscriptionClient
	db          *gorm.DB

	backfillSvc *BackfillService
	realtimeSvc *RealtimeSyncService
}

func NewSyncer(client *GraphQLClient[BlocksData], txClient *GraphQLClient[TxsData], subClient *SubscriptionClient, db *gorm.DB) *Syncer {
	syncer := &Syncer{
		blockClient: client,
		txClient:    txClient,
		subClient:   subClient,
		db:          db,
	}

	// service init
	syncer.backfillSvc = &BackfillService{syncer: syncer}
	syncer.realtimeSvc = &RealtimeSyncService{syncer: syncer, subClient: subClient}

	return syncer
}

// SyncBlocks block range synchronization
func (s *Syncer) SyncBlocks(ctx context.Context, fromHeight, toHeight int) error {
	// get block data
	var bd BlocksData
	if err := s.blockClient.Do(ctx, QBlocks, map[string]interface{}{
		"gt": fromHeight,
		"lt": toHeight,
	}, &bd); err != nil {
		return fmt.Errorf("sync blocks: %w", err)
	}

	// save each block to db
	for _, block := range bd.GetBlocks {
		if err := s.saveBlock(ctx, block); err != nil {
			log.Printf("failed to save block: %v", err)
			continue
		}
	}
	log.Printf("synced %d blocks from height %d to %d", len(bd.GetBlocks), fromHeight, toHeight)
	return nil
}

// SyncTxs Transaction range synchronization
func (s *Syncer) SyncTxs(ctx context.Context, fromHeight, toHeight int) error {
	// get transaction data
	var td TxsData
	if err := s.txClient.Do(ctx, QTxs, map[string]interface{}{
		"gt":   fromHeight,
		"lt":   toHeight,
		"imax": 1000,
	}, &td); err != nil {
		return fmt.Errorf("sync txs: %w", err)
	}

	// save each transaction to db
	for _, tx := range td.GetTransactions {
		if err := s.saveTxs(ctx, tx); err != nil {
			log.Printf("failed to save transaction: %v", err)
			continue
		}
	}
	log.Printf("synced %d transactions from height %d to %d", len(td.GetTransactions), fromHeight, toHeight)
	return nil
}

// saveBlock block save to db (duplication check)
func (s *Syncer) saveBlock(ctx context.Context, block Block) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&Block{}).Where("hash = ?", block.Hash).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check block existence: %w", err)
	}
	if count > 0 {
		return nil // already exists
	}

	// save block
	if err := s.db.WithContext(ctx).Create(&block).Error; err != nil {
		return fmt.Errorf("failed to insert block: %w", err)
	}

	return nil
}

// saveTxs transaction save to db (duplication check)
func (s *Syncer) saveTxs(ctx context.Context, tx Tx) error {
	// duplication check
	var count int64
	if err := s.db.WithContext(ctx).Model(&Tx{}).Where("hash = ?", tx.Hash).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check transaction existence: %w", err)
	}
	if count > 0 {
		return nil // already exists
	}

	// gas_fee to JSON
	var gasFeeJSON []byte
	if tx.GasFee != nil {
		gasFeeJSON, _ = json.Marshal(tx.GasFee)
	}

	// response to JSON
	var responseJSON []byte
	if tx.Response != nil {
		responseJSON, _ = json.Marshal(tx.Response)
	}

	// message to JSON
	var messagesJSON []byte
	if tx.Messages != nil {
		messagesJSON, _ = json.Marshal(tx.Messages)
	}

	// save transaction
	txRecord := map[string]interface{}{
		"hash":          tx.Hash,
		"block_height":  tx.BlockHeight,
		"tx_index":      tx.Index,
		"success":       tx.Success,
		"gas_wanted":    tx.GasWanted,
		"gas_used":      tx.GasUsed,
		"memo":          tx.Memo,
		"content_raw":   tx.ContentRaw,
		"gas_fee":       gasFeeJSON,
		"messages_json": messagesJSON,
		"response_json": responseJSON,
	}

	if err := s.db.WithContext(ctx).Table("indexer.transactions").Create(txRecord).Error; err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	return nil
}

func (s *Syncer) GetLastSyncedHeight(ctx context.Context) (int, error) {
	var lastHeight int
	err := s.db.WithContext(ctx).Model(&Block{}).Select("COALESCE(MAX(height), 0)").Scan(&lastHeight).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get last block height: %w", err)
	}
	return lastHeight, nil
}

func (s *Syncer) SyncRange(ctx context.Context, fromHeight, toHeight int) error {
	// block sync
	if err := s.SyncBlocks(ctx, fromHeight, toHeight); err != nil {
		return fmt.Errorf("failed to sync blocks: %w", err)
	}

	// transaction sync
	if err := s.SyncTxs(ctx, fromHeight, toHeight); err != nil {
		return fmt.Errorf("failed to sync txs: %w", err)
	}
	return nil
}

func (s *Syncer) StartRealtimeSync(ctx context.Context) error {
	return s.realtimeSvc.Start(ctx)
}

func (s *Syncer) BackfillToLatest(ctx context.Context) error {
	return s.backfillSvc.BackfillToLatest(ctx)
}

// handleRealtimeBlock : Real-time block processing only method
func (s *Syncer) handleRealtimeBlock(ctx context.Context, block Block) error {
	// save block
	if err := s.saveBlock(ctx, block); err != nil {
		return fmt.Errorf("save realtime block: %w", err)
	}

	// transaction sync
	if block.NumTxs > 0 {
		if err := s.SyncTxs(ctx, block.NumTxs, block.Height); err != nil {
			return fmt.Errorf("sync txs: %w", err)
		}
	}

	log.Printf("realtime sync: block %d saved", block.Height)
	return nil
}
