package models

import "time"

type SyncState struct {
	Component    string    `json:"component" gorm:"primaryKey;column:component"`
	LastBlockH   int64     `json:"last_block_h" gorm:"column:last_block_h"`
	LastTxHash   string    `json:"last_tx_hash" gorm:"column:last_tx_hash"`
	LastSyncTime time.Time `json:"last_sync_time" gorm:"column:last_sync_time"`
	Status       string    `json:"status" gorm:"column:status"`
	ErrorCount   int       `json:"error_count" gorm:"column:error_count"`
	LastError    string    `json:"last_error" gorm:"column:last_error"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (SyncState) TableName() string {
	return "indexer.app_state"
}
