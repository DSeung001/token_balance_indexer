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
