# Token Balance Indexer - 테이블 설계 설명

이 문서는 `0001_init.up.sql`의 테이블 구조를 바탕으로, **[BE 과제] 블록체인 이벤트 기반 Token Balance Indexer 개발** 요구사항에 맞추어 각 테이블의 역할과 사용 시나리오를 정리한 것입니다.

---

## 1. blocks
블록체인의 블록 헤더 정보를 저장하는 테이블입니다.
- **주요 컬럼**: `hash`(PK), `height`(UNIQUE), `time`, `total_txs`, `num_txs`
- **용도**:
  - Block Synchronizer(Producer)가 인덱서에서 블록을 수신하여 저장
  - 블록 height로 순차 동기화 및 누락 블록 백필 처리
  - `transactions` 테이블과 1:N 관계

---

## 2. transactions
블록에 포함된 트랜잭션의 메타데이터 및 원본 JSON을 저장합니다.
- **주요 컬럼**: `hash`(PK), `block_height`(FK), `index_in_block`(UNIQUE within block), `success`, `messages_json`, `response_json`
- **용도**:
  - 트랜잭션 원본 데이터를 보존하여 재파싱 가능
  - 이벤트 파싱의 근거 데이터
  - 블록 삭제 시 함께 삭제(CASCADE)

---

## 3. tx_events
트랜잭션 내 발생한 개별 이벤트를 저장합니다.
- **주요 컬럼**: `id`(PK), `tx_hash`(FK), `event_index`(UNIQUE with tx_hash), `type`, `func`, `pkg_path`
- **용도**:
  - 이벤트 단위 멱등성 보장
  - Transfer/Mint/Burn 이벤트 필터링 근거
  - 큐로 전송할 이벤트 데이터의 기초

---

## 4. tx_event_attrs
이벤트의 속성(key-value 쌍)을 순서(index)까지 포함해 저장합니다.
- **주요 컬럼**: `event_id`(FK), `attr_index`(UNIQUE with event_id), `key`, `value`
- **용도**:
  - `from`, `to`, `value` 값 검증 및 재구성
  - 이벤트 삭제 시 속성도 함께 삭제(CASCADE)

---

## 5. tokens
토큰 메타데이터로 나중에 확장성을 위한 테이블로 데이터 무결성 유지와 lazy registration, 품질 표시 향상등 다양한 측면에서 활용 가능 
- **주요 컬럼**: `token_path`(PK), `symbol`, `decimals`
- **용도**:
  - 이벤트에서 처음 보는 토큰을 등록
  - balances, transfers 참조
  - API 응답 시 심볼/소수자릿수 정보 제공

---

## 6. transfers
토큰 전송 내역을 저장합니다.
- **주요 컬럼**: `id`(PK), `tx_hash`(FK), `event_index`(UNIQUE with tx_hash), `token_path`(FK), `from_address`, `to_address`, `amount`, `block_height`
- **용도**:
  - 전송 이력 API의 소스
  - Mint/Burn/Transfer 이벤트 처리 결과 저장
  - 멱등성 보장

---

## 7. balances
주소별 토큰 잔액 스냅샷을 저장합니다.
- **주요 컬럼**: `(address, token_path)`(PK), `amount`, `last_tx_hash`, `last_block_h`
- **용도**:
  - Event Processor가 잔액 계산 후 UPSERT
  - `last_*` 컬럼으로 처리 지점 추적 및 리플레이 최적화
  - Balance API의 데이터 소스

---

## 8. app_state
각 컴포넌트의 동기화 상태를 저장하는 걸로 예기치 못한 이유로 종료되도 영속성, 일관성을 지키기 위한 체크포인트
- **주요 컬럼**: `component`(PK), `last_block_h`, `last_tx_hash`
- **용도**:
  - block_sync, event_consumer 등 서비스별 진행 포인터 관리
  - 재시작 시 이어받기 가능

---

## 과제 요구사항과 매핑

- **Block Synchronizer (Producer)** → `blocks`, `transactions`, `tx_events`, `tx_event_attrs`, `tokens`
- **Event Processor (Consumer)** → `transfers`, `balances`, `tokens`
- **Balance API** → `balances` + `tokens` 조인, `transfers` 조회
- **멱등성 보장** → `UNIQUE` 제약 조건 다수, FK와 CASCADE/RESTRICT 전략
