package types

import "gn-indexer/internal/domain"

// BlocksData represents block subscription response data (single block)
type BlocksData struct {
	GetBlocks domain.Block `json:"getBlocks"`
}

// BlocksDataArr represents block query response data (multiple blocks)
type BlocksDataArr struct {
	GetBlocks []domain.Block `json:"getBlocks"`
}

// TxsData represents transaction query response data
type TxsData struct {
	GetTransactions []domain.Transaction `json:"getTransactions"`
}

// BalanceResponse represents the response for /tokens/balances endpoint
type BalanceResponse struct {
	Balances []TokenBalance `json:"balances"`
}

// TokenBalance represents a single token balance
type TokenBalance struct {
	TokenPath string `json:"tokenPath"`
	Amount    int64  `json:"amount"`
}

// AccountBalanceResponse represents the response for /tokens/{tokenPath}/balances endpoint
type AccountBalanceResponse struct {
	AccountBalances []AccountBalance `json:"accountBalances"`
}

// AccountBalance represents a single account balance for a specific token
type AccountBalance struct {
	Address   string `json:"address"`
	TokenPath string `json:"tokenPath"`
	Amount    int64  `json:"amount"`
}

// TransferHistoryResponse represents the response for /tokens/transfer-history endpoint
type TransferHistoryResponse struct {
	Transfers []TransferRecord `json:"transfers"`
}

// TransferRecord represents a single transfer record
type TransferRecord struct {
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	TokenPath   string `json:"tokenPath"`
	Amount      int64  `json:"amount"`
}
