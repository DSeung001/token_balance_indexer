# 블록 동기화 전략 및 작업 계획

## 개요
마지막 저장된 블록 높이부터 현재 블록까지 안전하게 동기화하는 시스템을 구축합니다. 이는 블록체인 인덱서의 핵심 기능으로, 중단 없이 연속적인 블록 동기화를 보장해야 합니다.

## 현재 상황 분석

### 기존 구현된 기능
- `GetLastSyncedHeight()`: 데이터베이스에서 마지막 동기화된 블록 높이 조회
- `SyncRange()`: 지정된 범위의 블록과 트랜잭션 동기화
- 기본적인 중복 체크 및 에러 처리

### 부족한 부분
- 연속 동기화 모드 미구현
- 현재 블록 높이 조회 기능 부재
- 동기화 상태 추적 및 복구 메커니즘 부족
- 에러 발생 시 재시도 로직 부재

## 안전한 마지막 블록 높이 추적 방법

### 1. 데이터베이스 기반 추적 (과제 단계 권장)
```sql
-- app_state 테이블 활용
SELECT last_block_h FROM indexer.app_state WHERE component = 'block_sync';

-- blocks 테이블에서 최대 높이 조회 (백업)
SELECT COALESCE(MAX(height), 0) FROM indexer.blocks;
```

**장점:**
- 영속성 보장 (서비스 재시작 시에도 상태 유지)
- 트랜잭션 내에서 원자적 업데이트 가능
- 여러 컴포넌트 간 상태 공유
- 구현 단순, 일관성 보장

**구현 방법:**
- `app_state` 테이블에 `block_sync` 컴포넌트의 `last_block_h` 업데이트
- 동기화 완료 시마다 체크포인트 저장
- DB 연결 재시도 및 헬스체크 로직 추가

### 2. 파일 시스템 기반 추적 (참고용)
```go
type SyncState struct {
    LastSyncedHeight int64 `json:"last_synced_height"`
    LastSyncTime     time.Time `json:"last_sync_time"`
    Checksum         string `json:"checksum"`
}
```

**장점:**
- 데이터베이스와 독립적
- 간단한 구현

**단점:**
- 파일 손상 시 상태 손실 위험
- 동시성 제어 어려움
- 과제 단계에서는 불필요한 복잡성

### 3. 단순 DB 기반 접근법 (과제 단계 권장)
- 주 상태: 데이터베이스 (`app_state`)만 사용
- 장점: 구현 단순, 일관성 보장, 유지보수 용이
- 단점: DB 장애 시 상태 접근 불가
- 향후 확장: 필요시 백업 시스템 추가

## 작업 계획

### Phase 1: 기본 동기화 상태 관리 (1-2일)

#### 1.1 SyncState 구조체 정의
```go
type SyncState struct {
    Component     string    `json:"component"`
    LastBlockH    int64     `json:"last_block_h"`
    LastTxHash    string    `json:"last_tx_hash"`
    LastSyncTime  time.Time `json:"last_sync_time"`
    Status        string    `json:"status"` // "syncing", "completed", "error"
    ErrorCount    int       `json:"error_count"`
    LastError     string    `json:"last_error"`
}
```

#### 1.2 SyncStateRepository 구현
```go
type SyncStateRepository interface {
    GetState(ctx context.Context, component string) (*SyncState, error)
    UpdateState(ctx context.Context, state *SyncState) error
    UpdateLastBlock(ctx context.Context, component string, height int64, txHash string) error
    GetLastSyncedHeight(ctx context.Context, component string) (int64, error)
}
```

#### 1.3 현재 블록 높이 조회 기능
```go
func (c *Client[T]) GetCurrentBlockHeight(ctx context.Context) (int64, error)
```

### Phase 2: 연속 동기화 모드 구현 (2-3일)

#### 2.1 ContinuousSyncer 구조체
```go
type ContinuousSyncer struct {
    syncer        *Syncer
    stateRepo     SyncStateRepository
    client        *Client[BlocksData]
    interval      time.Duration
    maxRetries    int
    retryDelay    time.Duration
}
```

#### 2.2 동기화 루프 구현
```go
func (cs *ContinuousSyncer) Start(ctx context.Context) error {
    ticker := time.NewTicker(cs.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := cs.syncNextBatch(ctx); err != nil {
                log.Printf("sync error: %v", err)
                // 재시도 로직
            }
        }
    }
}
```

#### 2.3 배치 동기화 로직
```go
func (cs *ContinuousSyncer) syncNextBatch(ctx context.Context) error {
    // 1. 현재 상태 조회
    state, err := cs.stateRepo.GetState(ctx, "block_sync")
    if err != nil {
        return err
    }
    
    // 2. 현재 블록 높이 조회
    currentHeight, err := cs.client.GetCurrentBlockHeight(ctx)
    if err != nil {
        return err
    }
    
    // 3. 동기화 범위 계산
    fromHeight := state.LastBlockH + 1
    toHeight := min(fromHeight + batchSize - 1, currentHeight)
    
    if fromHeight > toHeight {
        return nil // 동기화할 블록 없음
    }
    
    // 4. 동기화 실행
    if err := cs.syncer.SyncRange(ctx, int(fromHeight), int(toHeight)); err != nil {
        return err
    }
    
    // 5. 상태 업데이트
    return cs.stateRepo.UpdateLastBlock(ctx, "block_sync", toHeight, "")
}
```

### Phase 3: 안전성 및 복구 메커니즘 (2-3일)

#### 3.1 체크포인트 시스템
```go
type Checkpoint struct {
    BlockHeight int64     `json:"block_height"`
    TxHash      string    `json:"tx_hash"`
    Timestamp   time.Time `json:"timestamp"`
    Checksum    string    `json:"checksum"`
}
```

#### 3.2 검증 로직
```go
func (cs *ContinuousSyncer) validateSync(ctx context.Context, fromHeight, toHeight int64) error {
    // 1. 블록 개수 검증
    expectedCount := toHeight - fromHeight + 1
    actualCount, err := cs.countBlocksInRange(ctx, fromHeight, toHeight)
    if err != nil {
        return err
    }
    
    if actualCount != expectedCount {
        return fmt.Errorf("block count mismatch: expected %d, got %d", expectedCount, actualCount)
    }
    
    // 2. 체인 연결성 검증
    return cs.validateChainContinuity(ctx, fromHeight, toHeight)
}
```

#### 3.3 재시도 및 백오프 전략
```go
type RetryStrategy struct {
    MaxRetries    int
    BaseDelay     time.Duration
    MaxDelay      time.Duration
    BackoffFactor float64
}

func (rs *RetryStrategy) Execute(operation func() error) error {
    var lastErr error
    delay := rs.BaseDelay
    
    for attempt := 0; attempt <= rs.MaxRetries; attempt++ {
        if err := operation(); err == nil {
            return nil
        } else {
            lastErr = err
            if attempt < rs.MaxRetries {
                time.Sleep(delay)
                delay = time.Duration(float64(delay) * rs.BackoffFactor)
                if delay > rs.MaxDelay {
                    delay = rs.MaxDelay
                }
            }
        }
    }
    
    return fmt.Errorf("operation failed after %d attempts: %w", rs.MaxRetries, lastErr)
}
```

### Phase 4: 모니터링 및 로깅 (1-2일)

#### 4.1 동기화 메트릭
```go
type SyncMetrics struct {
    TotalBlocksSynced    int64
    TotalTxsSynced       int64
    SyncDuration         time.Duration
    LastSyncTime         time.Time
    ErrorRate            float64
    AverageBatchSize     float64
}
```

#### 4.2 구조화된 로깅
```go
type SyncLog struct {
    Timestamp     time.Time `json:"timestamp"`
    Level         string    `json:"level"`
    Component     string    `json:"component"`
    BlockHeight   int64     `json:"block_height"`
    Message       string    `json:"message"`
    Error         string    `json:"error,omitempty"`
    Duration      string    `json:"duration,omitempty"`
}
```

## 구현 우선순위

### High Priority (1주차)
1. `SyncStateRepository` 구현
2. 현재 블록 높이 조회 기능
3. 기본 연속 동기화 루프
4. DB 연결 재시도 및 헬스체크

### Medium Priority (2주차)
1. 체크포인트 시스템
2. 검증 로직
3. 재시도 전략
4. 에러 처리 및 로깅

### Low Priority (3주차)
1. 기본 모니터링
2. 성능 최적화
3. 설정 관리
4. 향후 확장 고려사항

## 테스트 전략

### 단위 테스트
- `SyncStateRepository` 메서드별 테스트
- 동기화 로직 단위 테스트
- 검증 로직 테스트

### 통합 테스트
- 전체 동기화 플로우 테스트
- 에러 상황 시뮬레이션
- 성능 테스트

### E2E 테스트
- 실제 블록체인과의 연동 테스트
- 장시간 실행 안정성 테스트

## 위험 요소 및 대응 방안

### 1. 네트워크 불안정
- **위험**: 블록체인 노드 연결 실패
- **대응**: 다중 노드 지원, 자동 재연결

### 2. 데이터베이스 성능 저하
- **위험**: 대량 블록 처리 시 성능 저하
- **대응**: 배치 크기 조정, 인덱스 최적화

### 3. 메모리 누수
- **위험**: 장시간 실행 시 메모리 증가
- **대응**: 정기적인 메모리 정리, 가비지 컬렉션 모니터링

### 4. 동시성 문제
- **위험**: 여러 프로세스 동시 동기화
- **대응**: 락 메커니즘, 프로세스 ID 검증

## 성공 지표

1. **안정성**: 99.9% 이상의 동기화 성공률
2. **성능**: 평균 배치 처리 시간 < 5초
3. **복구성**: 에러 발생 후 1분 이내 자동 복구
4. **모니터링**: 실시간 동기화 상태 가시성

## 다음 단계

1. `SyncStateRepository` 인터페이스 정의
2. 현재 블록 높이 조회 GraphQL 쿼리 구현
3. 기본 연속 동기화 루프 구현
4. DB 연결 재시도 로직 구현
5. 단위 테스트 작성 및 실행

## 과제 단계 접근법

### 핵심 원칙: **단순하고 견고하게**
- 복잡한 백업 시스템보다는 DB 기반 단일 소스로 구현
- 에러 처리와 재시도 로직에 집중
- 실제 운영 환경에서 발생할 수 있는 문제들에 대한 기본적인 대응
- 코드 가독성과 유지보수성 우선

### 향후 확장 고려사항
- 필요시 파일 기반 백업 시스템 추가
- 다중 DB 지원 (Redis, MongoDB 등)
- 분산 환경에서의 동기화 상태 공유
