# 📂 MSA 프로젝트 폴더 구조 선정 이유

## 1. 개요
이번 과제는 **MSA(Microservices Architecture)** 기반의 블록체인 인덱싱 시스템 구축입니다.
MSA 구조에서는 서비스별 역할을 분리하고, 공통 코드는 재사용 가능하도록 구성하는 것이 중요합니다.

이번 프로젝트에서는 **Monorepo** 방식을 채택하여, 하나의 저장소에서 모든 서비스를 관리하고,  
`cmd/`와 `internal/` 디렉토리를 활용한 Go 언어의 표준 프로젝트 레이아웃을 적용했습니다.

---

## 2. 폴더 구조

```
gn-indexer/
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
```

---

## 3. 구조 선정 이유

### 3.1 `cmd/` - 서비스 실행 진입점
- 각 서비스의 `main.go` 파일을 별도로 관리
- `go build ./cmd/api`처럼 개별 서비스만 빌드 가능
- MSA에서 서비스별 배포/확장 용이

### 3.2 `internal/` - 공통 모듈
- 서비스 간 공유하는 로직(DB 연결, Config, Queue, 도메인 로직 등)을 `internal/`에 모음
- Go 언어 특성상 `internal/`은 외부 패키지에서 접근 불가 → **캡슐화** 보장

### 3.3 Monorepo 채택 이유
- Monorepo: (Monolithic Repository) 방식은 여러 개의 서비스나 프로젝트를 하나의 저장소(repo)에 모아 관리하는 방식
- Polyrepo, Hybrid Repo 대신 이 방법을 쓴 이유는 공통 코드 재사용과 개발 환겨 세팅과 과제 제출의 간편화
- Docker Compose로 전체 서비스 및 인프라(PostgreSQL, SQS) 한 번에 실행 가능합니다.

### 3.4 MSA 아키텍처 대응
- **Syncer**: 블록/트랜잭션 수집 → 이벤트 추출 → SQS 전송
- **Processor**: SQS 이벤트 소비 → 잔액 계산 → DB 저장
- **API**: 잔액 및 이력 조회 REST API 제공
- 서비스 간 통신은 SQS 기반 → 결합도 낮음, 확장성 높음

---

## 4. 기대 효과
- 서비스별 독립 실행 및 배포 가능
- 공통 코드 관리 효율성 ↑
- 변경 영향 최소화 (하나의 서비스 수정이 다른 서비스에 최소 영향)
- 과제 요구사항(MSA, Queue, Docker Compose)에 적합
