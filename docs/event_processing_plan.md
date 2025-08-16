# 🎯 이벤트 처리 작업 계획 (Mint, Burn, Transfer)

## 📋 현재 상황 분석

### ✅ 이미 구현된 부분
- **DB 스키마**: `tx_events`, `tx_event_attrs`, `transfers`, `balances` 테이블 완성
- **도메인 모델**: `Transaction`, `Block`, `Token`, `Balance`, `Transfer` 구조체 정의
- **기본 서비스**: `BackfillService`로 블록 동기화 로직 구현
- **데이터베이스**: PostgreSQL 연결 및 마이그레이션 완료

### ❌ 구현이 필요한 부분
- **이벤트 파싱 로직**: Mint, Burn, Transfer 이벤트 식별 및 파싱
- **이벤트 큐 전송**: 파싱된 이벤트를 큐에 전송
- **이벤트 처리 서비스**: 큐에서 이벤트를 받아 잔액 계산 및 저장

---

## 🏗️ 구현해야 할 구조

### 1. **이벤트 파싱 서비스** (`internal/parsing/event_parser.go`)

#### 이벤트 타입 정의
```go
type EventType string

const (
    EventTypeMint     EventType = "Mint"
    EventTypeBurn     EventType = "Burn"
    EventTypeTransfer EventType = "Transfer"
)
```

#### 파싱된 이벤트 구조체
```go
type ParsedEvent struct {
    Type        EventType
    TokenPath   string
    FromAddress string
    ToAddress   string
    Amount      int64
    TxHash      string
    BlockHeight int64
    EventIndex  int
}
```

#### 이벤트 파서 인터페이스
```go
type EventParser interface {
    ParseEvents(tx *domain.Transaction) ([]ParsedEvent, error)
    IsTokenEvent(event *domain.GnoEvent) bool
    ParseTokenEvent(event *domain.GnoEvent, tx *domain.Transaction) (*ParsedEvent, error)
}
```

### 2. **이벤트 큐 서비스** (`internal/queue/event_queue.go`)

#### 큐 메시지 구조체
```go
type EventMessage struct {
    EventType   string `json:"event_type"`
    TokenPath   string `json:"token_path"`
    FromAddress string `json:"from_address"`
    ToAddress   string `json:"to_address"`
    Amount      int64  `json:"amount"`
    TxHash      string `json:"tx_hash"`
    BlockHeight int64  `json:"block_height"`
    EventIndex  int    `json:"event_index"`
}
```

#### 큐 인터페이스
```go
type EventQueue interface {
    SendEvent(ctx context.Context, event *EventMessage) error
    ReceiveEvents(ctx context.Context) (<-chan *EventMessage, error)
}
```

### 3. **이벤트 처리 서비스** (`internal/service/event_processor.go`)

#### 이벤트 처리 서비스 구조체
```go
type EventProcessorService struct {
    eventQueue    EventQueue
    balanceRepo   repository.BalanceRepository
    transferRepo  repository.TransferRepository
    tokenRepo     repository.TokenRepository
}
```

#### 잔액 계산 로직
```go
func (eps *EventProcessorService) ProcessEvent(ctx context.Context, event *EventMessage) error {
    // 1. 토큰 등록/업데이트
    // 2. Transfer 테이블에 기록
    // 3. Balance 테이블 업데이트 (UPSERT)
}
```

---

## 📅 구현 순서 및 단계

### **Phase 1: 이벤트 파싱 로직 구현**
1. **이벤트 파서 생성** (`internal/parsing/event_parser.go`)
   - Mint, Burn, Transfer 이벤트 식별 로직
   - 이벤트 속성 파싱 (`from`, `to`, `value`)
   - 유효성 검증 (주소 형식, 금액 등)

2. **도메인 모델 확장**
   - `ParsedEvent` 구조체 추가
   - 이벤트 타입 상수 정의

### **Phase 2: 큐 시스템 구현**
1. **로컬 큐 구현** (`internal/queue/local_queue.go`)
   - 메모리 기반 큐 (개발용)
   - SQS 호환 인터페이스

2. **이벤트 전송 로직**
   - `BackfillService`에 이벤트 파싱 및 큐 전송 추가
   - 실시간 동기화 시에도 이벤트 전송

### **Phase 3: 이벤트 처리 서비스**
1. **이벤트 프로세서 구현**
   - 큐에서 이벤트 수신
   - 잔액 계산 로직
   - DB 저장 (UPSERT)

2. **잔액 계산 알고리즘**
   ```go
   // Mint: to_address에 amount 추가
   // Burn: from_address에서 amount 차감
   // Transfer: from_address에서 차감, to_address에 추가
   ```

### **Phase 4: 통합 및 테스트**
1. **서비스 연결**
   - `main.go`에서 이벤트 프로세서 실행
   - 백필과 실시간 동기화 연동

2. **테스트 코드 작성**
   - 이벤트 파싱 테스트
   - 잔액 계산 테스트
   - 전체 플로우 테스트

---

## 🔧 핵심 구현 포인트

### **1. 이벤트 식별 로직**
```go
func (ep *EventParser) IsTokenEvent(event *domain.GnoEvent) bool {
    // Transfer 타입이면서
    if event.Type != "Transfer" {
        return false
    }
    
    // Mint, Burn, Transfer 함수 중 하나인지 확인
    switch event.Func {
    case "Mint", "Burn", "Transfer":
        return true
    default:
        return false
    }
}
```

### **2. 이벤트 속성 파싱**
```go
func (ep *EventParser) ParseTokenEvent(event *domain.GnoEvent, tx *domain.Transaction) (*ParsedEvent, error) {
    // from, to, value 속성 추출
    var fromAddr, toAddr string
    var amount int64
    
    for _, attr := range event.Attrs {
        switch attr.Key {
        case "from":
            fromAddr = attr.Value
        case "to":
            toAddr = attr.Value
        case "value":
            if val, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
                amount = val
            }
        }
    }
    
    // Mint: from="", to=주소
    // Burn: from=주소, to=""
    // Transfer: from=주소, to=주소
    
    return &ParsedEvent{
        Type:        ep.determineEventType(event.Func, fromAddr, toAddr),
        TokenPath:   event.PkgPath,
        FromAddress: fromAddr,
        ToAddress:   toAddr,
        Amount:      amount,
        TxHash:      tx.Hash,
        BlockHeight: int64(tx.BlockHeight),
        EventIndex:  event.Index, // 이 필드 추가 필요
    }, nil
}
```

### **3. 잔액 계산 로직**
```go
func (eps *EventProcessorService) updateBalance(ctx context.Context, event *EventMessage) error {
    // Mint: to_address 잔액 증가
    if event.EventType == "Mint" {
        return eps.balanceRepo.AddBalance(ctx, event.ToAddress, event.TokenPath, event.Amount)
    }
    
    // Burn: from_address 잔액 감소
    if event.EventType == "Burn" {
        return eps.balanceRepo.SubtractBalance(ctx, event.FromAddress, event.TokenPath, event.Amount)
    }
    
    // Transfer: from_address 감소, to_address 증가
    if event.EventType == "Transfer" {
        if err := eps.balanceRepo.SubtractBalance(ctx, event.FromAddress, event.TokenPath, event.Amount); err != nil {
            return err
        }
        return eps.balanceRepo.AddBalance(ctx, event.ToAddress, event.TokenPath, event.Amount)
    }
    
    return nil
}
```

---

## 📊 이벤트 처리 플로우

### **전체 시스템 플로우**
```
1. Block Syncer (Producer)
   ↓ 블록/트랜잭션 수신
2. 이벤트 파싱
   ↓ Mint/Burn/Transfer 이벤트 식별
3. 이벤트 큐 전송
   ↓ 큐에 메시지 저장
4. Event Processor (Consumer)
   ↓ 큐에서 이벤트 수신
5. 잔액 계산 및 DB 저장
   ↓ Balance, Transfer 테이블 업데이트
6. Balance API
   ↓ 잔액 조회 응답
```

### **이벤트별 처리 로직**
| 이벤트 타입 | From Address | To Address | 잔액 변경 |
|------------|--------------|------------|-----------|
| **Mint**   | "" (빈 문자열) | 실제 주소 | To 주소에 +amount |
| **Burn**   | 실제 주소 | "" (빈 문자열) | From 주소에 -amount |
| **Transfer** | 실제 주소 | 실제 주소 | From 주소에 -amount, To 주소에 +amount |

---

## 🧪 테스트 전략

### **단위 테스트**
1. **이벤트 파싱 테스트**
   - 각 이벤트 타입별 파싱 정확성
   - 잘못된 이벤트 데이터 처리

2. **잔액 계산 테스트**
   - Mint/Burn/Transfer 시나리오별 계산
   - 음수 잔액 방지

3. **큐 시스템 테스트**
   - 메시지 전송/수신 정확성
   - 동시성 처리

### **통합 테스트**
1. **전체 플로우 테스트**
   - 블록 동기화 → 이벤트 파싱 → 큐 전송 → 처리 → DB 저장

2. **데이터 무결성 테스트**
   - 중복 이벤트 처리
   - 순서 보장
   - 장애 복구

---

## 🚀 다음 단계

### **즉시 시작 가능한 작업**
1. **이벤트 파서 구현** (`internal/parsing/event_parser.go`)
2. **기본 테스트 코드 작성**
3. **도메인 모델 확장**

### **단계별 목표**
- **1주차**: 이벤트 파싱 로직 완성
- **2주차**: 큐 시스템 및 이벤트 처리 서비스 구현
- **3주차**: 통합 테스트 및 최적화

---

## 📚 참고 자료

- [과제 요구사항](./task.md)
- [데이터베이스 스키마](./db_table.md)
- [블록 동기화 전략](./block_sync_strategy.md)
- [GraphQL API 가이드](./graphql.md)

---

## 💡 구현 시 주의사항

1. **멱등성 보장**: 동일한 이벤트가 여러 번 처리되어도 안전해야 함
2. **트랜잭션 처리**: 잔액 업데이트 시 원자성 보장
3. **에러 처리**: 파싱 실패, DB 오류 등에 대한 적절한 처리
4. **성능 최적화**: 대량 이벤트 처리 시 배치 처리 고려
5. **로깅**: 디버깅을 위한 상세한 로그 기록

