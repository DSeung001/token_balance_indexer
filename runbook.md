# Dev Ops Runbook — Gno Token Balance Indexer

이 문서는 **Docker Compose + Postgres + LocalStack(SQS) + golang-migrate**를 사용하는 로컬 개발환경을 *명령어 위주*로 정리한 실행 가이드입니다.  
아래 예시는 `.env` 파일을 사용하는 전제입니다.

---

## 0) 준비물
- Docker / Docker Compose
- AWS CLI (LocalStack 조작용)
- psql (선택, DB 셸 접속용)

---

## 1) 환경변수(.env) 예시
```env
# Postgres
POSTGRES_USER=app
POSTGRES_PASSWORD=app
POSTGRES_DB=indexer
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
DATABASE_URL=postgres://app:app@postgres:5432/indexer?sslmode=disable

# LocalStack (SQS)
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_DEFAULT_REGION=ap-northeast-2
LOCALSTACK_EDGE_PORT=4566

# Compose 공용
COMPOSE_PROJECT_NAME=gnindexer
NETWORK_NAME=app-net
```
> 주의: `DATABASE_URL`의 **호스트명은 컨테이너 기준**으로 `postgres` 사용.

---

## 2) 컨테이너 올리기 / 내리기
```bash
# 백그라운드 기동
docker compose --env-file .env up -d

# 종료 + 볼륨 삭제
docker compose --env-file .env down -v
```

### 상태/로그 확인
```bash
docker compose ps
docker compose logs -f --tail=200
```

---

## 3) 데이터베이스
### 3.1 Postgres 헬스체크(Compose에서 자동)
`postgres` 서비스는 `pg_isready` 기반 헬스체크가 설정되어 있습니다.

### 3.2 psql 접속
```bash
psql "host=localhost port=5432 user=app password=app dbname=indexer sslmode=disable"
```
> `.env`를 로컬 쉘에 export 했거나, 값을 숫자/문자열로 치환해서 실행.

### 3.3 스키마 로딩(수동 SQL)
```bash
cat schema/schema.sql | psql "host=localhost port=5432 user=app password=app dbname=indexer sslmode=disable"
```

---

## 4) 마이그레이션 (golang-migrate)
> 컨테이너 `migrate` 서비스가 있다면 Compose 기동 시 자동으로 `up`합니다.  
> 수동으로 실행하고 싶다면 아래 중 택1.

### 4.1 도커로 1회성 실행
```bash
docker run --rm --network host -v "$PWD/db/migrations:/migrations" migrate/migrate:4   -path=/migrations   -database "postgres://app:app@localhost:5432/indexer?sslmode=disable"   up
```

### 4.2 버전 내리기 / 특정 버전으로 강제
```bash
# 한 버전 내리기
docker run --rm --network host -v "$PWD/db/migrations:/migrations" migrate/migrate:4   -path=/migrations -database "postgres://app:app@localhost:5432/indexer?sslmode=disable"   down 1

# 깨졌을 때 강제로 특정 버전으로 맞추기 (예: v2)
docker run --rm --network host -v "$PWD/db/migrations:/migrations" migrate/migrate:4   -path=/migrations -database "postgres://app:app@localhost:5432/indexer?sslmode=disable"   force 2
```

### 4.3 로컬 마이그레이션
```bash
migrate -path ./migrations -database "postgres://app:app@localhost:5432/indexer?sslmode=disable" up
```

---

## 5) LocalStack(SQS)
### 5.1 헬스체크
`localstack` 서비스는 `http://localhost:4566/_localstack/health`로 체크합니다.

### 5.2 큐 만들기 / 리스트 / 퍼지
```bash
# 편의 변수
SQS_URL="http://localhost:4566"
QUEUE_NAME="${SQS_QUEUE_NAME:-token-events}"
ACCOUNT_ID="000000000000"

# 큐 생성
aws --endpoint-url="${SQS_URL}" sqs create-queue --queue-name "${QUEUE_NAME}"
echo "Created queue at: ${SQS_URL}/${ACCOUNT_ID}/${QUEUE_NAME}"

# 큐 목록
aws --endpoint-url="${SQS_URL}" sqs list-queues

# 큐 비우기
aws --endpoint-url="${SQS_URL}" sqs purge-queue --queue-url "${SQS_URL}/${ACCOUNT_ID}/${QUEUE_NAME}"
```

> 애플리케이션에서 사용할 **Queue URL** 예시:  
> `http://localstack:4566/000000000000/${SQS_QUEUE_NAME}` (컨테이너 내부 기준)

---

## 6) 애플리케이션 (Gin + GORM)
- 실행 전, DB 마이그레이션이 완료되어 있어야 합니다.
- 환경변수 예시:
```bash
export DATABASE_URL="postgres://app:app@localhost:5432/indexer?sslmode=disable"
export SQS_ENDPOINT="http://localhost:4566"
export SQS_QUEUE_URL="http://localhost:4566/000000000000/${SQS_QUEUE_NAME}"
# go run ./cmd/api  # 또는 docker compose 서비스로 기동
```

---

## 7)
- 재배포
```bash
# 1. 모든 컨테이너 중지 및 삭제
docker stop $(docker ps -aq)
docker rm $(docker ps -aq)

# 2. 모든 이미지 삭제
docker rmi -f $(docker images -q)

# 3. 모든 볼륨 삭제
docker volume rm $(docker volume ls -q)

# 4. (선택) 모든 네트워크 삭제
docker network rm $(docker network ls -q)

# 5. 재배포 (compose 기준)
docker compose --env-file .env up -d --build
```

---

## 8) 빠른 시작(요약)
```bash
# 1) 컨테이너 기동
docker compose --env-file .env up -d

# 2) 마이그레이션(필요 시 수동)
docker run --rm --network host -v "$PWD/db/migrations:/migrations" migrate/migrate:4   -path=/migrations -database "postgres://app:app@localhost:5432/indexer?sslmode=disable" up
migrate -path ./migrations -database "postgres://app[docker-compose.yml](docker-compose.yml):app@localhost:5432/indexer?sslmode=disable" up

# 3) SQS 큐 생성
SQS_URL="http://localhost:4566"; QUEUE_NAME="${SQS_QUEUE_NAME:-token-events}"
aws --endpoint-url="${SQS_URL}" sqs create-queue --queue-name "${QUEUE_NAME}"

# 4) 앱 실행 (예시)
export DATABASE_URL="postgres://app:app@localhost:5432/indexer?sslmode=disable"
export SQS_ENDPOINT="http://localhost:4566"
export SQS_QUEUE_URL="http://localhost:4566/000000000000/${QUEUE_NAME}"
# go run ./cmd/api
```
