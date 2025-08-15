0. 과제에 충족하게 폴더 구조 규칙을 잘 지키는 지
   - gn-indexer/
     ├─ cmd/
     │  ├─ api/           # Balance API 서비스 (REST API Server)
     │  ├─ processor/     # Event Processor 서비스 (SQS Consumer)
     │  └─ syncer/        # Block Synchronizer 서비스 (Producer)
     ├─ internal/
     │  ├─ api/           # API 라우팅, 핸들러
     │  ├─ config/        # 환경변수 로드, 설정값 관리
     │  ├─ db/            # DB 연결 및 마이그레이션
     │  ├─ domain/        # 핵심 도메인 로직, 엔티티
     │  ├─ indexer/       # GraphQL 호출 및 블록/트랜잭션 파싱 로직
     │  ├─ parsing/       # 이벤트 파싱 유틸리티
     │  └─ queue/         # SQS 송신/수신 로직
     ├─ schema/           # DB 스키마(SQL)
     ├─ .localstack/      # 로컬 SQS 환경 데이터
     ├─ docker-compose.yml
     ├─ Makefile
     ├─ go.mod / go.sum
     └─ README.md
1. internal/indexer/realtime_sync_service.go의 역할
2. syncer에 너무 많은 기능이 부담, msa 위반
3. internal/indexer/realtime_sync_service.go, 와 internal/indexer/syncer.go의 파일구조가 이상함
4. internal/indexer/subscription_client.go에서 ID 분리해서 구분처리