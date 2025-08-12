# 📦 인프라 환경 구성 (infra.md)

## 1. 선정 이유
이번 과제에서는 **블록체인 네트워크 이벤트 기반 MSA 인덱싱 시스템**을 구현해야 합니다.  
이 과정에서 안정적이고 확장 가능한 **오프체인 인프라** 구성이 필수적입니다.

- **PostgreSQL**  
  - 이벤트와 잔액, 전송 이력 등을 **영구 저장**하기 위해 사용
  - 관계형 구조와 트랜잭션 지원으로 데이터 정합성 보장
- **LocalStack SQS (Queue)**  
  - **서비스 간 비동기 메시지 전달**을 위해 선택
  - AWS SQS를 로컬에서 시뮬레이션 가능
  - Producer(블록 동기화) ↔ Consumer(이벤트 처리) 구조를 안전하게 분리
- **Docker Compose**  
  - Postgres와 LocalStack 환경을 한 번에 관리
  - 컨테이너 기반이라 로컬 개발 환경 오염 없이 재현 가능

---

## 2. 환경 구성
### 사용 기술
- **Docker Desktop** (Windows 환경)
- **Docker Compose V2**
- **PostgreSQL 14**
- **LocalStack (SQS)**
- **AWS CLI** (LocalStack SQS 테스트용)

### 디렉토리 구조
```
gn-indexer/
 ├─ docker-compose.yml
 ├─ .env
 ├─ schema/        # DB 초기 스키마 파일
 ├─ .localstack/   # LocalStack 데이터 저장 (옵션)
 └─ infra.md
```

---

## 3. 관련 명령어

### 3.1 Docker Compose로 인프라 실행
```bash
docker compose --env-file .env up -d
```

### 3.2 컨테이너 상태 확인
```bash
docker compose ps
```

---

### 3.3 LocalStack SQS 설정
#### AWS CLI 환경 설정
```bash
aws configure
# AWS Access Key ID: test
# AWS Secret Access Key: test
# Default region name: ap-northeast-2
# Default output format: json
```

#### SQS 큐 생성
```bash
aws --endpoint-url http://localhost:4566 sqs create-queue --queue-name token-events
```

#### 큐 목록 확인
```bash
aws --endpoint-url http://localhost:4566 sqs list-queues
```

---

### 3.4 PostgreSQL 접속
#### 로컬 `psql` 사용 시
```bash
psql "host=localhost port=5432 user=app password=app1234 dbname=gnodb sslmode=disable"
```

#### 컨테이너 내부에서 접속
```bash
docker exec -it gnindexer-postgres psql -U app -d gnodb
```

---

### 3.5 PostgreSQL 테스트 쿼리
```sql
\dt;
CREATE TABLE test_table (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
INSERT INTO test_table (name) VALUES ('first row');
SELECT * FROM test_table;
\q
```

---

## 4. 현재 상태
- **Postgres** 연결 확인 완료 ✅
- **LocalStack SQS** 큐 생성 및 조회 완료 ✅
- 과제에서 요구하는 **오프체인 인프라 환경 준비 완료**
