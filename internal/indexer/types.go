package indexer

type Block struct {
	Hash     string `json:"hash" gorm:"primaryKey;column:hash"`
	Height   int    `json:"height" gorm:"column:height;uniqueIndex"`
	Time     string `json:"time" gorm:"column:time"`
	NumTxs   int    `json:"num_txs" gorm:"column:num_txs"`
	TotalTxs int    `json:"total_txs" gorm:"column:total_txs"`
}

func (Block) TableName() string {
	return "indexer.blocks"
}

type Attr struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GnoEvent struct {
	Type    string `json:"type"`
	Func    string `json:"func"`
	PkgPath string `json:"pkg_path"`
	Attrs   []Attr `json:"attrs"`
}

type Tx struct {
	Index       int    `json:"index" gorm:"column:index_in_block"`
	Hash        string `json:"hash" gorm:"primaryKey;column:hash"`
	Success     bool   `json:"success" gorm:"column:success"`
	BlockHeight int    `json:"block_height" gorm:"column:block_height"`
	GasFee      *struct {
		Amount string `json:"amount"`
		Denom  string `json:"denom"`
	} `json:"gas_fee" gorm:"-"`
	Response *struct {
		Events []GnoEvent `json:"events"`
	} `json:"response" gorm:"-"`
}

func (Tx) TableName() string {
	return "indexer.transactions"
}

// BlocksData block query responses
type BlocksData struct {
	GetBlocks []Block `json:"getBlocks"`
}

// TxsData transaction query responses
type TxsData struct {
	GetTransactions []Tx `json:"getTransactions"`
}
