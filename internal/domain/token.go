package domain

import "time"

// Todo Separated into models and domains

// Token requests token information
type Token struct {
	Path     string `json:"token_path" gorm:"primaryKey;column:token_path"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

// Balance represents account balance for a specific token
type Balance struct {
	Address    string `json:"address" gorm:"primaryKey;column:address"`
	TokenPath  string `json:"token_path" gorm:"primaryKey;column:token_path"`
	Amount     int64  `json:"amount"`
	LastTxHash string `json:"last_tx_hash"`
	LastBlockH int64  `json:"last_block_h"`
}

// Transfer represents a token transfer event
type Transfer struct {
	ID          int64     `json:"id" gorm:"primaryKey;column:id"`
	TxHash      string    `json:"tx_hash" gorm:"column:tx_hash"`
	EventIndex  int       `json:"event_index" gorm:"column:event_index"`
	TokenPath   string    `json:"token_path" gorm:"column:token_path"`
	FromAddress string    `json:"from_address" gorm:"column:from_address"`
	ToAddress   string    `json:"to_address" gorm:"column:to_address"`
	Amount      int64     `json:"amount" gorm:"column:amount"`
	BlockHeight int64     `json:"block_height" gorm:"column:block_height"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName returns the table name for Transfer
func (Transfer) TableName() string {
	return "indexer.transfers"
}
