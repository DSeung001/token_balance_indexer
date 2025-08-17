# GN Indexer 

## 실행 방법

## 설계 의도

## 프로세스 흐름도

## 폴더 구조
```
gn-indexer/
├── cmd/                    # 메인 애플리케이션 진입점
│   ├── block-syncer/       # 블록 동기화 서비스 (Producer)
│   ├── balance-api/        # 잔액 조회 API 서비스
│   └── event-processor/    # 이벤트 처리 서비스 (Consumer)
├── internal/               # 내부 패키지들
│   ├── client/             # 외부 API 통신 (HTTP, WebSocket)
│   ├── producer/           # Producer 로직 (블록 동기화)
│   ├── consumer/           # Consumer 로직 (이벤트 처리)
│   ├── types/              # 공통 타입 정의
│   ├── service/            # 비즈니스 로직 서비스
│   ├── domain/             # 도메인 모델
│   ├── repository/         # 데이터 접근 계층
│   ├── queue/              # 메시지 큐 처리
│   ├── api/                # API 관련
│   └── config/             # 설정 관리
├── db/                     # 데이터베이스 관련 파일
├── docs/                   # 프로젝트 문서
├── docker-compose.yml      # Docker 환경 설정
├── go.mod                  # Go 모듈 정의
├── go.sum                  # Go 모듈 체크섬
└── README.md
```

## 환경 설정
로컬 개발 용이므로 .env는 저장소에 있는 그대로 가져다가 쓰시면 됩니다.

```env
# Postgres
POSTGRES_USER=app
POSTGRES_PASSWORD=app
POSTGRES_DB=indexer
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
DATABASE_URL=postgres://app:app@127.0.0.1:5432/indexer?sslmode=disable

# LocalStack (SQS)
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_DEFAULT_REGION=ap-northeast-2
AWS_REGION=ap-northeast-2
LOCALSTACK_EDGE_PORT=4566
LOCALSTACK_URL=http://localhost:4566
LOCALSTACK_INTERNAL_URL=http://localstack:4566
SQS_QUEUE_NAME=gn-token-events

# Compose
COMPOSE_PROJECT_NAME=gnindexer
NETWORK_NAME=app-net

# GraphQL URL
GRAPHQL_ENDPOINT=https://dev-indexer.api.gnoswap.io/graphql/query
GRAPHQL_WS_ENDPOINT=wss://dev-indexer.api.gnoswap.io/graphql/query
```

## 고도화

## 실행 명령어

### 1. 도커 인프라
```bash
# 처음 실행 또는 설정 변경 시
docker-compose up -d --build

# 기존 실행
docker-compose up -d

# 재시작
docker-compose restart

# 중지
docker-compose down

# 재빌드
docker-compose down -v
docker-compose up -d --build
```

### 2. 블록 동기화 + 이벤트 처리
```bash
# 실시간 동기화 (Producer + Consumer)
go run ./cmd/block-syncer -realtime

# 특정 범위 동기화
go run ./cmd/block-syncer -from 1 -to 1000

# 데이터 무결성 검사
go run ./cmd/block-syncer -integrity
```

### 3. 개별 서비스
```bash
# Balance API
go run ./cmd/balance-api

# Event Processor  
go run ./cmd/event-processor
```

## SQS 테스트

### 기본 테스트
```bash
# 큐 목록 확인
aws --endpoint-url http://127.0.0.1:4566 sqs list-queues

# 테스트 메시지 전송
aws --endpoint-url http://127.0.0.1:4566 sqs send-message --queue-url "http://sqs.ap-northeast-2.localhost.localstack.cloud:4566/000000000000/gn-token-events" --message-body "{\"test\": \"message\"}"

# 메시지 수신
aws --endpoint-url http://127.0.0.1:4566 sqs receive-message --queue-url "http://sqs.ap-northeast-2.localhost.localstack.cloud:4566/000000000000/gn-token-events"
```

### PowerShell (Windows)
```powershell
# 큐 목록 확인
aws --endpoint-url http://127.0.0.1:4566 sqs list-queues

# 테스트 메시지 전송
aws --endpoint-url http://127.0.0.1:4566 sqs send-message --queue-url "http://sqs.ap-northeast-2.localhost.localstack.cloud:4566/000000000000/gn-token-events" --message-body "{\"test\": \"message\"}"

# 메시지 수신
aws --endpoint-url http://127.0.0.1:4566 sqs receive-message --queue-url "http://sqs.ap-northeast-2.localhost.localstack.cloud:4566/000000000000/gn-token-events"
```
