리포지터리와 서비스 부분을 더 자세히 설명해드리겠습니다. 먼저 현재 구현된 부분을 확인하고 단계별로 설명해드리겠습니다.

Ran tool
Read file: internal/repository/transaction_repository.go
이제 리포지터리와 서비스 부분을 단계별로 자세히 설명해드리겠습니다.

## 🏗️ **Phase 3: 리포지터리 구현 (Repository Layer)**

### **리포지터리가 뭔가요?**
리포지터리는 **데이터베이스와 직접 소통하는 계층**입니다. 데이터를 저장하고, 조회하고, 수정하는 모든 작업을 담당합니다.

---

## �� **3단계: 이벤트 관련 리포지터리 구현**

### **3-1. 이벤트 리포지터리** (`internal/repository/event_repository.go`)

```go
package repository

import (
    "context"
    "gn-indexer/internal/domain"
    "gorm.io/gorm"
)

// EventRepository handles event data persistence
type EventRepository interface {
    Create(ctx context.Context, event *domain.TxEvent) error
    GetByTxHash(ctx context.Context, txHash string) ([]domain.TxEvent, error)
    GetByTxHashAndIndex(ctx context.Context, txHash string, eventIndex int) (*domain.TxEvent, error)
}

type postgresEventRepository struct {
    db *gorm.DB
}

// NewEventRepository creates a new PostgreSQL event repository
func NewEventRepository(db *gorm.DB) EventRepository {
    return &postgresEventRepository{db: db}
}

// Create saves an event to the tx_events table
func (r *postgresEventRepository) Create(ctx context.Context, event *domain.TxEvent) error {
    // Check if event already exists (멱등성 보장)
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.TxEvent{}).
        Where("tx_hash = ? AND event_index = ?", event.TxHash, event.EventIndex).
        Count(&count).Error
    
    if err != nil {
        return fmt.Errorf("failed to check event existence: %w", err)
    }
    
    if count > 0 {
        return nil // 이미 존재함
    }
    
    // 새 이벤트 저장
    return r.db.WithContext(ctx).Table("indexer.tx_events").Create(event).Error
}

// GetByTxHash retrieves all events for a transaction
func (r *postgresEventRepository) GetByTxHash(ctx context.Context, txHash string) ([]domain.TxEvent, error) {
    var events []domain.TxEvent
    err := r.db.WithContext(ctx).
        Where("tx_hash = ?", txHash).
        Order("event_index").
        Find(&events).Error
    
    if err != nil {
        return nil, fmt.Errorf("failed to get events by tx hash: %w", err)
    }
    
    return events, nil
}

// GetByTxHashAndIndex retrieves a specific event by transaction hash and event index
func (r *postgresEventRepository) GetByTxHashAndIndex(ctx context.Context, txHash string, eventIndex int) (*domain.TxEvent, error) {
    var event domain.TxEvent
    err := r.db.WithContext(ctx).
        Where("tx_hash = ? AND event_index = ?", txHash, eventIndex).
        First(&event).Error
    
    if err != nil {
        return nil, fmt.Errorf("failed to get event: %w", err)
    }
    
    return &event, nil
}
```

### **3-2. 이벤트 속성 리포지터리** (`internal/repository/event_attr_repository.go`)

```go
package repository

import (
    "context"
    "gn-indexer/internal/domain"
    "gorm.io/gorm"
)

// EventAttrRepository handles event attribute data persistence
type EventAttrRepository interface {
    Create(ctx context.Context, attr *domain.TxEventAttr) error
    GetByEventID(ctx context.Context, eventID int64) ([]domain.TxEventAttr, error)
}

type postgresEventAttrRepository struct {
    db *gorm.DB
}

// NewEventAttrRepository creates a new PostgreSQL event attribute repository
func NewEventAttrRepository(db *gorm.DB) EventAttrRepository {
    return &postgresEventAttrRepository{db: db}
}

// Create saves an event attribute to the tx_event_attrs table
func (r *postgresEventAttrRepository) Create(ctx context.Context, attr *domain.TxEventAttr) error {
    // Check if attribute already exists
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.TxEventAttr{}).
        Where("event_id = ? AND attr_index = ?", attr.EventID, attr.AttrIndex).
        Count(&count).Error
    
    if err != nil {
        return fmt.Errorf("failed to check attribute existence: %w", err)
    }
    
    if count > 0 {
        return nil // 이미 존재함
    }
    
    // 새 속성 저장
    return r.db.WithContext(ctx).Table("indexer.tx_event_attrs").Create(attr).Error
}

// GetByEventID retrieves all attributes for an event
func (r *postgresEventAttrRepository) GetByEventID(ctx context.Context, eventID int64) ([]domain.TxEventAttr, error) {
    var attrs []domain.TxEventAttr
    err := r.db.WithContext(ctx).
        Where("event_id = ?", eventID).
        Order("attr_index").
        Find(&attrs).Error
    
    if err != nil {
        return nil, fmt.Errorf("failed to get attributes by event ID: %w", err)
    }
    
    return attrs, nil
}
```

### **3-3. 전송 리포지터리** (`internal/repository/transfer_repository.go`)

```go
package repository

import (
    "context"
    "gn-indexer/internal/domain"
    "gorm.io/gorm"
)

// TransferRepository handles transfer data persistence
type TransferRepository interface {
    Create(ctx context.Context, transfer *domain.Transfer) error
    GetByTxHash(ctx context.Context, txHash string) ([]domain.Transfer, error)
    GetByAddress(ctx context.Context, address string) ([]domain.Transfer, error)
}

type postgresTransferRepository struct {
    db *gorm.DB
}

// NewTransferRepository creates a new PostgreSQL transfer repository
func NewTransferRepository(db *gorm.DB) TransferRepository {
    return &postgresTransferRepository{db: db}
}

// Create saves a transfer record to the transfers table
func (r *postgresTransferRepository) Create(ctx context.Context, transfer *domain.Transfer) error {
    // Check if transfer already exists
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.Transfer{}).
        Where("tx_hash = ? AND event_index = ?", transfer.TxHash, transfer.EventIndex).
        Count(&count).Error
    
    if err != nil {
        return fmt.Errorf("failed to check transfer existence: %w", err)
    }
    
    if count > 0 {
        return nil // 이미 존재함
    }
    
    // 새 전송 기록 저장
    return r.db.WithContext(ctx).Table("indexer.transfers").Create(transfer).Error
}

// GetByTxHash retrieves all transfers for a transaction
func (r *postgresTransferRepository) GetByTxHash(ctx context.Context, txHash string) ([]domain.Transfer, error) {
    var transfers []domain.Transfer
    err := r.db.WithContext(ctx).
        Where("tx_hash = ?", txHash).
        Order("event_index").
        Find(&transfers).Error
    
    if err != nil {
        return nil, fmt.Errorf("failed to get transfers by tx hash: %w", err)
    }
    
    return transfers, nil
}

// GetByAddress retrieves all transfers involving a specific address
func (r *postgresTransferRepository) GetByAddress(ctx context.Context, address string) ([]domain.Transfer, error) {
    var transfers []domain.Transfer
    err := r.db.WithContext(ctx).
        Where("from_address = ? OR to_address = ?", address, address).
        Order("block_height DESC, event_index DESC").
        Find(&transfers).Error
    
    if err != nil {
        return nil, fmt.Errorf("failed to get transfers by address: %w", err)
    }
    
    return transfers, nil
}
```

---

## ��️ **Phase 4: 서비스 구현 (Service Layer)**

### **서비스가 뭔가요?**
서비스는 **비즈니스 로직을 담당하는 계층**입니다. 리포지터리들을 조합해서 실제 업무를 처리합니다.

---

## 📋 **4단계: 이벤트 저장 서비스 구현**

### **4-1. 이벤트 저장 서비스** (`internal/service/event_storage_service.go`)

```go
package service

import (
    "context"
    "fmt"
    "gn-indexer/internal/domain"
    "gn-indexer/internal/parsing"
    "gn-indexer/internal/repository"
    "log"
)

// EventStorageService handles storing parsed events to database
type EventStorageService struct {
    eventRepo      repository.EventRepository
    eventAttrRepo  repository.EventAttrRepository
    transferRepo   repository.TransferRepository
    tokenRepo      repository.TokenRepository
    eventParser    *parsing.EventParser
}

// NewEventStorageService creates a new event storage service
func NewEventStorageService(
    eventRepo repository.EventRepository,
    eventAttrRepo repository.EventAttrRepository,
    transferRepo repository.TransferRepository,
    tokenRepo repository.TokenRepository,
) *EventStorageService {
    return &EventStorageService{
        eventRepo:     eventRepo,
        eventAttrRepo: eventAttrRepo,
        transferRepo:  transferRepo,
        tokenRepo:     tokenRepo,
        eventParser:   parsing.NewEventParser(),
    }
}

// ProcessTransaction processes a transaction and stores its events
func (ess *EventStorageService) ProcessTransaction(ctx context.Context, tx *domain.Transaction) error {
    log.Printf("Processing transaction %s for events", tx.Hash)
    
    // 1단계: 트랜잭션에서 이벤트 파싱
    parsedEvents, err := ess.eventParser.ParseEventsFromTransaction(tx)
    if err != nil {
        return fmt.Errorf("parse events: %w", err)
    }
    
    if len(parsedEvents) == 0 {
        log.Printf("No token events found in transaction %s", tx.Hash)
        return nil
    }
    
    log.Printf("Found %d token events in transaction %s", len(parsedEvents), tx.Hash)
    
    // 2단계: 각 이벤트를 하나씩 처리
    for _, parsedEvent := range parsedEvents {
        if err := ess.processSingleEvent(ctx, &parsedEvent, tx); err != nil {
            return fmt.Errorf("process event: %w", err)
        }
    }
    
    return nil
}

// processSingleEvent processes a single parsed event
func (ess *EventStorageService) processSingleEvent(ctx context.Context, event *domain.ParsedEvent, tx *domain.Transaction) error {
    log.Printf("Processing event %s for transaction %s", event.Type, event.TxHash)
    
    // 1단계: tx_events 테이블에 이벤트 저장
    txEvent := &domain.TxEvent{
        TxHash:     event.TxHash,
        EventIndex: event.EventIndex,
        Type:       string(event.Type),
        Func:       event.Type.String(), // MINT, BURN, TRANSFER
        PkgPath:    event.TokenPath,
    }
    
    if err := ess.eventRepo.Create(ctx, txEvent); err != nil {
        return fmt.Errorf("create tx_event: %w", err)
    }
    
    log.Printf("Saved event to tx_events table with ID: %d", txEvent.ID)
    
    // 2단계: tx_event_attrs 테이블에 속성 저장
    attrs := []domain.TxEventAttr{
        {EventID: txEvent.ID, AttrIndex: 0, Key: "from", Value: event.FromAddress},
        {EventID: txEvent.ID, AttrIndex: 1, Key: "to", Value: event.ToAddress},
        {EventID: txEvent.ID, AttrIndex: 2, Key: "value", Value: fmt.Sprintf("%d", event.Amount)},
    }
    
    for _, attr := range attrs {
        if err := ess.eventAttrRepo.Create(ctx, &attr); err != nil {
            return fmt.Errorf("create tx_event_attr: %w", err)
        }
    }
    
    log.Printf("Saved %d attributes to tx_event_attrs table", len(attrs))
    
    // 3단계: transfers 테이블에 전송 기록 저장
    transfer := &domain.Transfer{
        TxHash:      event.TxHash,
        EventIndex:  event.EventIndex,
        TokenPath:   event.TokenPath,
        FromAddress: event.FromAddress,
        ToAddress:   event.ToAddress,
        Amount:      event.Amount,
        BlockHeight: event.BlockHeight,
    }
    
    if err := ess.transferRepo.Create(ctx, transfer); err != nil {
        return fmt.Errorf("create transfer: %w", err)
    }
    
    log.Printf("Saved transfer record to transfers table")
    
    // 4단계: tokens 테이블에 새 토큰 등록 (아직 없는 경우)
    if err := ess.tokenRepo.RegisterIfNotExists(ctx, event.TokenPath); err != nil {
        return fmt.Errorf("register token: %w", err)
    }
    
    log.Printf("Successfully processed event %s for transaction %s", event.Type, event.TxHash)
    return nil
}
```

---

## 🔄 **전체 흐름 이해하기**

### **데이터 흐름**
```
1. 트랜잭션 데이터 (JSON) 
   ↓
2. 이벤트 파서가 파싱
   ↓
3. 이벤트 저장 서비스가 처리
   ↓
4. 각 리포지터리가 DB에 저장
```

### **각 계층의 역할**
- **도메인 모델**: 데이터 구조 정의
- **파서**: JSON에서 이벤트 추출
- **서비스**: 비즈니스 로직 (파싱 → 저장)
- **리포지터리**: DB 작업 (INSERT, SELECT 등)

---

## �� **다음 단계**

1. **리포지터리 구현** (위의 3개 파일)
2. **이벤트 저장 서비스 구현** (위의 1개 파일)
3. **main.go에서 서비스 연결**
4. **테스트 실행**

이제 이해가 되셨나요? 어떤 부분이 더 궁금하신지 알려주세요!