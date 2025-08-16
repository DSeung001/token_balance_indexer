package domain

import "time"

// Block represents a blockchain block
type Block struct {
	Hash          string    `json:"hash" gorm:"primaryKey;column:hash"`
	Height        int       `json:"height" gorm:"column:height;uniqueIndex"`
	LastBlockHash string    `json:"last_block_hash" gorm:"column:last_block_hash"`
	Time          time.Time `json:"time" gorm:"column:time"`
	NumTxs        int       `json:"num_txs" gorm:"column:num_txs"`
	TotalTxs      int       `json:"total_txs" gorm:"column:total_txs"`
}

// TableName returns the table name for Block
func (Block) TableName() string {
	return "indexer.blocks"
}
