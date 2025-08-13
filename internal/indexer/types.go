package indexer

type Block struct {
	Hash     string `json:"hash"`
	Height   int    `json:"height"`
	Time     string `json:"time"`
	NumTxs   int    `json:"num_txs"`
	TotalTxs int    `json:"total_txs"`
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
	Index       int    `json:"index"`
	Hash        string `json:"hash"`
	Success     bool   `json:"success"`
	BlockHeight int    `json:"block_height"`
	GasFee      *struct {
		Amount string `json:"amount"`
		Denom  string `json:"denom"`
	} `json:"gas_fee"`
	Response *struct {
		Events []GnoEvent `json:"events"`
	} `json:"response"`
}

// 블록 쿼리 응답용
type BlocksData struct {
	GetBlocks []Block `json:"getBlocks"`
}

// 트랜잭션 쿼리 응답용  
type TxsData struct {
	GetTransactions []Tx `json:"getTransactions"`
}
