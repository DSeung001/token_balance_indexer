package domain

type EventType string

// Todo Separated into models and domains

const (
	EventTypeMint     EventType = "MINT"
	EventTypeBurn     EventType = "BURN"
	EventTypeTransfer EventType = "TRANSFER"
)

type ParsedEvent struct {
	Type        string
	Func        EventType
	TokenPath   string
	FromAddress string
	ToAddress   string
	Amount      int64
	TxHash      string
	BlockHeight int64
	EventIndex  int
}

// TxEvent represents a database event record
type TxEvent struct {
	ID         int64  `json:"id" gorm:"primaryKey;column:id"`
	TxHash     string `json:"tx_hash" gorm:"column:tx_hash"`
	EventIndex int    `json:"event_index" gorm:"column:event_index"`
	Type       string `json:"type" gorm:"column:type"`
	Func       string `json:"func" gorm:"column:func"`
	PkgPath    string `json:"pkg_path" gorm:"column:pkg_path"`
}

// TxEventAttr represents a database event attribute record
type TxEventAttr struct {
	ID        int64  `json:"id" gorm:"primaryKey;column:id"`
	EventID   int64  `json:"event_id" gorm:"column:event_id"`
	AttrIndex int    `json:"attr_index" gorm:"column:attr_index"`
	Key       string `json:"key" gorm:"column:key"`
	Value     string `json:"value" gorm:"column:value"`
}
