package domain

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
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	TokenPath   string `json:"token_path"`
	Amount      int64  `json:"amount"`
	BlockHeight int64  `json:"block_height"`
}
