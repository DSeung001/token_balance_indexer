# GN-Indexer 블록체인 이벤트 기반 Token Balance Indexer 구현 계획

## 📋 현재 상황 분석

### ✅ 이미 구현된 기능
- **데이터베이스 스키마**: 블록, 트랜잭션, 이벤트, 토큰, 전송, 잔액 테이블 완성
- **도메인 모델**: Block, Transaction, Event, Token 구조체 정의
- **이벤트 파싱**: 토큰 발행(Mint), 소멸(Burn), 전송(Transfer) 이벤트 파싱 로직
- **기본 서비스**: 블록 동기화, 백필, 실시간 동기화, 데이터 무결성 검사
- **Repository 계층**: PostgreSQL 데이터 접근 로직
- **MSA 폴더 구조**: Producer/Consumer 패턴으로 개선된 아키텍처

### ❌ 누락된 핵심 기능 (과제 요구사항)
1. **Queue 시스템**: LocalStack SQS 기반 메시지 큐 구현
2. **Event Processor**: 큐에서 이벤트를 소비하여 잔액 계산
3. **Balance API**: REST API 엔드포인트 구현 (3개 엔드포인트)
4. **이벤트 큐 연동**: Block Syncer에서 파싱된 이벤트를 SQS에 전송
5. **Docker Compose 환경**: LocalStack SQS 포함 전체 환경 구성

## 🎯 구현 우선순위 (오늘 하루 내 완성 목표)

### 🔥 **Phase 1: Queue 시스템 구현 (2-3시간)**
- LocalStack SQS 기반 이벤트 큐 구현
- Block Syncer와 Event Processor 간 SQS 통신 인터페이스 정의
- Docker Compose에 LocalStack SQS 환경 추가

### 🔥 **Phase 2: Event Processor 구현 (2-3시간)**
- 큐에서 이벤트 소비하여 잔액 계산 로직
- 잔액 테이블 업데이트 및 토큰 정보 관리

### 🔥 **Phase 3: Balance API 구현 (2-3시간)**
- REST API 서버 구현 (Gin 프레임워크 사용)
- 3개 엔드포인트 완성:
  - `GET /tokens/balances?address={address}`
  - `GET /tokens/{tokenPath}/balances?address={address}`
  - `GET /tokens/transfer-history?address={address}`

### 🔥 **Phase 4: 통합 및 테스트 (1-2시간)**
- 전체 시스템 연동 테스트
- 간단한 데모 시나리오 실행

## 📁 구현할 파일 구조

```
internal/
├── queue/
│   ├── event_queue.go          # 이벤트 큐 인터페이스 정의
│   └── sqs_queue.go            # LocalStack SQS 기반 큐 구현
├── service/
│   ├── balance_service.go      # 잔액 계산 서비스
│   └── event_processor.go      # 이벤트 처리 서비스
└── api/
    ├── handlers/
    │   ├── balance_handler.go  # 잔액 조회 핸들러
    │   └── token_handler.go    # 토큰 관련 핸들러
    └── server.go               # HTTP 서버 구현 (Gin)

cmd/
├── event-processor/
│   └── main.go                 # 이벤트 프로세서 메인
└── balance-api/
    └── main.go                 # API 서버 메인

docker-compose.yml              # LocalStack SQS 포함 환경
```

## 🔧 세부 구현 계획

### **Phase 1: Queue 시스템**
1. `internal/queue/event_queue.go`: 큐 인터페이스 정의
2. `internal/queue/sqs_queue.go`: LocalStack SQS 기반 큐 구현
3. `docker-compose.yml`: LocalStack SQS 환경 추가
4. Block Syncer에서 이벤트 파싱 후 SQS에 전송 로직 추가

### **Phase 2: Event Processor**
1. `internal/service/balance_service.go`: 잔액 계산 로직
2. `internal/service/event_processor.go`: 큐 소비 및 이벤트 처리
3. `cmd/event-processor/main.go`: 독립 실행 가능한 이벤트 프로세서

### **Phase 3: Balance API**
1. `internal/api/handlers/`: REST API 핸들러들 (3개 엔드포인트)
2. `internal/api/server.go`: HTTP 서버 구현 (Gin 프레임워크)
3. `cmd/balance-api/main.go`: API 서버 실행 파일
4. 잔액 조회 및 전송 내역 조회 로직 구현

### **Phase 4: 통합**
1. Block Syncer → SQS → Event Processor → Database 흐름 테스트
2. API 엔드포인트 동작 확인 (3개 엔드포인트)
3. Docker Compose 환경에서 전체 시스템 시연 준비
4. 단위 테스트 코드 작성

## ⚡ 빠른 구현을 위한 전략

### **1. 최소 기능 우선**
- 완벽한 에러 처리보다는 기본 동작에 집중
- 로깅은 간단하게, 핵심 로직에 집중

### **2. LocalStack 환경 최적화**
- LocalStack SQS로 실제 SQS와 동일한 환경 구성
- Docker Compose로 PostgreSQL + LocalStack SQS 실행

### **3. 테스트 데이터 활용**
- 기존 블록 동기화로 테스트 데이터 생성
- 실제 블록체인 연동보다는 로컬 테스트 우선

## 🚀 실행 순서

1. **LocalStack SQS 환경 구성** → Docker Compose 설정
2. **Queue 시스템 구현** → Block Syncer와 SQS 연동
3. **Event Processor 구현** → SQS 소비 및 잔액 계산
4. **Balance API 구현** → REST 엔드포인트 완성 (3개)
5. **통합 테스트** → 전체 플로우 검증

## 📊 예상 소요 시간

- **LocalStack SQS 환경**: 1시간
- **Queue 시스템**: 2-3시간
- **Event Processor**: 2-3시간  
- **Balance API**: 2-3시간
- **통합 및 테스트**: 1-2시간
- **총 예상 시간**: 8-12시간

## 🎯 최종 목표

오늘 하루 내에 블록체인 이벤트 기반 Token Balance Indexer 과제 요구사항을 모두 충족하는 **기능적인 MSA 시스템**을 완성하여 과제 제출이 가능한 상태로 만드는 것.

**완성 시점에서 확인할 수 있는 것:**
- Block Syncer 실행 → SQS에 이벤트 전송
- Event Processor 실행 → SQS에서 이벤트 소비 및 잔액 계산
- Balance API 실행 → 3개 REST 엔드포인트 응답
- Docker Compose 환경에서 전체 시스템 연동 동작 확인
- 단위 테스트 코드 포함
