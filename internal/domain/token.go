package domain

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"time"
)

// U64 represents a u64 domain type from PostgreSQL
type U64 struct {
	*big.Int
}

// NewU64 creates a new U64 from int64
func NewU64(value int64) *U64 {
	return &U64{big.NewInt(value)}
}

// NewU64FromString creates a new U64 from string
func NewU64FromString(value string) (*U64, error) {
	bi, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, fmt.Errorf("invalid u64 string: %s", value)
	}
	return &U64{bi}, nil
}

// Value implements driver.Valuer for database serialization
func (u U64) Value() (driver.Value, error) {
	if u.Int == nil {
		return nil, nil
	}
	return u.String(), nil
}

// Scan implements sql.Scanner for database deserialization
func (u *U64) Scan(value interface{}) error {
	if value == nil {
		u.Int = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		bi, ok := new(big.Int).SetString(v, 10)
		if !ok {
			return fmt.Errorf("invalid u64 string: %s", v)
		}
		u.Int = bi
	case []byte:
		bi, ok := new(big.Int).SetString(string(v), 10)
		if !ok {
			return fmt.Errorf("invalid u64 bytes: %s", string(v))
		}
		u.Int = bi
	default:
		return fmt.Errorf("cannot scan %T into U64", value)
	}
	return nil
}

// Int64 converts U64 to int64 (may lose precision for large values)
func (u U64) Int64() int64 {
	if u.Int == nil {
		return 0
	}
	return u.Int.Int64()
}

// String returns string representation
func (u U64) String() string {
	if u.Int == nil {
		return "0"
	}
	return u.Int.String()
}

// Token requests token information
type Token struct {
	Path      string    `json:"token_path" gorm:"primaryKey;column:token_path"`
	Symbol    string    `json:"symbol"`
	Decimals  int       `json:"decimals"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName returns the table name for Token
func (Token) TableName() string {
	return "indexer.tokens"
}

// Balance represents account balance for a specific token
type Balance struct {
	Address    string    `json:"address" gorm:"primaryKey;column:address"`
	TokenPath  string    `json:"token_path" gorm:"primaryKey;column:token_path"`
	Amount     *U64      `json:"amount" gorm:"column:amount;type:numeric"`
	LastTxHash string    `json:"last_tx_hash" gorm:"column:last_tx_hash"`
	LastBlockH int64     `json:"last_block_h" gorm:"column:last_block_h"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name for Balance
func (Balance) TableName() string {
	return "indexer.balances"
}

// Transfer represents a token transfer event
type Transfer struct {
	ID          int64     `json:"id" gorm:"primaryKey;column:id"`
	TxHash      string    `json:"tx_hash" gorm:"column:tx_hash"`
	EventIndex  int       `json:"event_index" gorm:"column:event_index"`
	TokenPath   string    `json:"token_path" gorm:"column:token_path"`
	FromAddress string    `json:"from_address" gorm:"column:from_address"`
	ToAddress   string    `json:"to_address" gorm:"column:to_address"`
	Amount      *U64      `json:"amount" gorm:"column:amount;type:numeric"`
	BlockHeight int64     `json:"block_height" gorm:"column:block_height"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName returns the table name for Transfer
func (Transfer) TableName() string {
	return "indexer.transfers"
}
