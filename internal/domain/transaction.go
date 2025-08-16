package domain

// Todo Separated into models and domains

// TransactionMessage represents a transaction message
type TransactionMessage struct {
	Route string      `json:"route"`
	Value interface{} `json:"value"`
}

// GasFee represents gas fee information
type GasFee struct {
	Amount interface{} `json:"amount"`
	Denom  string      `json:"denom"`
}

// GnoEvent represents a blockchain event
type GnoEvent struct {
	Type    string `json:"type"`
	Func    string `json:"func"`
	PkgPath string `json:"pkg_path"`
	Attrs   []Attr `json:"attrs"`
}

// Attr represents an event attribute
type Attr struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Transaction represents a blockchain transaction
type Transaction struct {
	Index       int                   `json:"index" gorm:"column:tx_index"`
	Hash        string                `json:"hash" gorm:"primaryKey;column:hash"`
	Success     bool                  `json:"success" gorm:"column:success"`
	BlockHeight int                   `json:"block_height" gorm:"column:block_height"`
	GasWanted   int64                 `json:"gas_wanted" gorm:"column:gas_wanted"`
	GasUsed     int64                 `json:"gas_used" gorm:"column:gas_used"`
	Memo        string                `json:"memo" gorm:"column:memo"`
	ContentRaw  string                `json:"content_raw" gorm:"column:content_raw"`
	GasFee      *GasFee               `json:"gas_fee" gorm:"-"`
	Messages    *[]TransactionMessage `json:"messages" gorm:"-"`
	Response    *TransactionResponse  `json:"response" gorm:"-"`
}

// TransactionResponse represents transaction response data
type TransactionResponse struct {
	Events []GnoEvent `json:"events"`
}

// TableName returns the table name for Transaction
func (Transaction) TableName() string {
	return "indexer.transactions"
}
