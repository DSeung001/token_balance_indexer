package indexer

import "gn-indexer/internal/domain"

// BlocksData represents block query response data
type BlocksData struct {
	GetBlocks []domain.Block `json:"getBlocks"`
}

// TxsData represents transaction query response data
type TxsData struct {
	GetTransactions []domain.Transaction `json:"getTransactions"`
}
