# GN Indexer 실행 가이드

## 환경 설정
`.env` 파일을 프로젝트 루트에 생성하고 다음 내용을 추가하세요:

```env

```

## 실행 순서

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
