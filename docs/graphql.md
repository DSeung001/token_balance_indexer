# GraphQL 핵심 개념 정리 (필수 요소, 실전 체크리스트 포함)

> 목적: GraphQL을 처음 도입할 때 반드시 이해해야 하는 구성요소, HTTP 사용 규약, 에러/페이지네이션/보안 패턴, 그리고 Go에서의 견고한 호출 예시를 한 파일에 집약.

---

## 1) 스키마와 SDL (Schema Definition Language)

- **스키마**: 서버가 노출하는 타입과 연산의 계약(Contract).
- **SDL**: 스키마를 서술하는 언어. 파일 확장자 예: `schema.graphql`.
- **핵심 타입 종류**
  - Scalar: `Int`, `Float`, `String`, `Boolean`, `ID` (+ 커스텀 스칼라 가능)
  - Object: 필드의 집합. `type User { id: ID! name: String! }`
  - Enum: 제한된 리터럴 집합.
  - Interface: 공통 필드 계약, 구현 타입에서 실체화.
  - Union: 전혀 다른 객체 타입들 중 하나.
  - Input: 인자(입력) 객체 전용 타입(출력에 사용 불가).
  - Non-Null: `!` 필수값. List와의 조합 유의: `[T]` vs `[T!]` vs `[T]!`.
- **루트 타입**: `type Query`, `type Mutation`, `type Subscription`

```graphql
schema { query: Query mutation: Mutation }

type Query {
  me: User!
  user(id: ID!): User
}

type Mutation {
  updateName(id: ID!, name: String!): User!
}

type User {
  id: ID!
  name: String!
  role: Role!
}

enum Role { ADMIN USER }
```
**근거**: GraphQL 공통 스키마 모델과 SDL 표준 표기.


---

## 2) Selection Set, 필드, 인자, 별칭, 변수

- **Selection Set**: 반환 필드를 **명시적으로** 고른다(오버페치 방지).
- **인자(Arguments)**: `user(id: $id)` 형태로 필드에 전달.
- **변수(Variables)**: `query ($id: ID!)` 정의부와 `$id` 사용부 분리.
- **별칭(Alias)**: 동일 필드 두 번 가져올 때 이름 충돌 방지.

```graphql
query ($id: ID!) {
  a: user(id: $id) { id name }
  b: user(id: $id) { id role }
}
```
**근거**: Selection Set/변수/별칭은 클라이언트 오버페치/중복 요청 방지의 핵심.


---

## 3) 프래그먼트(Fragments)

- **Named Fragment**: 재사용 가능한 필드 묶음.
- **Inline Fragment**: 유니온/인터페이스의 **구체 타입**에 조건부 적용.

```graphql
fragment UserCore on User { id name }
query { me { ...UserCore } }

query {
  node(id: "42") {
    ... on User { id name }
    ... on Admin { id permissions }
  }
}
```
**근거**: 타입 다형성 처리 및 중복 제거.


---

## 4) 디렉티브(Directives)

- 빌트인: `@include(if:)`, `@skip(if:)`, `@deprecated(reason:)`, `@specifiedBy(url:)`
- 커스텀 디렉티브: 서버가 정의해 의미를 부여.

```graphql
query($debug: Boolean!) {
  me {
    id
    debugFields @include(if: $debug) { traceId }
  }
}
```
**근거**: 실행 시점에 결과 셰이프를 제어.


---

## 5) 에러 모델

- 응답 구조: `{"data": {...}, "errors": [...]}`
- **부분 성공** 가능: 일부 필드는 `null` + `errors`에 상세 포함.
- 프로덕션 방침: 클라이언트가 `errors`를 항상 검사하도록 강제.

```json
{
  "data": { "user": null },
  "errors": [{ "message": "not found", "path": ["user"] }]
}
```
**근거**: GraphQL은 HTTP 200에서도 비즈니스 에러를 `errors[]`로 반환.


---

## 6) 페이지네이션 패턴

- **Offset/Limit**: 단순, 대용량 테이블에서 비효율 발생 가능.
  - 인자 예: `limit`, `offset` 또는 `take`, `skip`
- **Cursor 기반(Relay 스타일)**: 안정적 페이지 이동.
  - 필드: `edges { cursor node { ... } } pageInfo { hasNextPage endCursor }`
- **정렬(order_by)** 필수. 커서/오프셋 모두 결정적 정렬 필요.

```graphql
query($first: Int!, $after: String) {
  users(first: $first, after: $after) {
    edges { cursor node { id name } }
    pageInfo { hasNextPage endCursor }
  }
}
```
**근거**: 커서 페이지네이션은 추가/삭제에도 안정적.


---

## 7) GraphQL over HTTP 핵심 규칙

- **엔드포인트 구분**
  - `/graphql` = Playground/GraphiQL 같은 **HTML UI**를 노출하는 경로인 경우가 많음
  - `/graphql/query` = **기계용 JSON API** 경로로 분리되는 배포도 흔함
- **요청 헤더**
  - `Content-Type: application/json`  → 요청 바디 형식 선언
  - `Accept: application/json` 또는 `application/graphql-response+json` → 응답 형식 요구
- **상태 코드**
  - 유효한 GraphQL 실행 결과면 대체로 200. 애플리케이션 에러는 `errors[]`.
  - 인증/권한/레이트리밋 등은 401/403/429 등 **HTTP 레이어**로도 표현 가능.
- **GET vs POST**
  - Query는 캐시 혜택을 위해 GET 허용하는 서버도 있음(쿼리 문자열 인코딩)
  - Mutation은 POST가 일반적.

```bash
# POST 예시
curl -X POST https://api.example.com/graphql \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"query":"query($id:ID!){user(id:$id){id name}}","variables":{"id":"1"}}'
```
**근거**: 콘텐츠 협상 실패 시 HTML이 반환될 수 있으므로 Accept를 명시. 엔드포인트 구분 필수.


---

## 8) 보안/성능

- **쿼리 복잡도/깊이 제한**: Worst-case 방지.
- **Persisted Query**: 해시 기반 고정 쿼리로 대형 쿼리 방지·캐시 유리.
- **N+1 문제 해결**: DataLoader(배칭/캐싱).
- **레이트 리밋/권한**: HTTP 레이어 + 리졸버 레벨 모두 고려.
- **Nullability 설계**: 비즈니스 모델의 안정성과 디폴트 동작을 좌우.


---

## 9) 스키마 진화

- 신규 필드 **추가**는 안전.
- 제거는 위험 → `@deprecated`로 사전 안내 후 마이그레이션 기간 부여.
- Non-Null 전환은 브레이킹 체인지.


---

## 10) 클라이언트/서버 도구

- 서버: Apollo Server, Yoga, Mercurius, Hasura, PostGraphile, Graphene, graphql-go, **gqlgen(Go)** 등
- 클라이언트: Apollo Client, Relay, URQL, graphql-request
- 검사/문서화: GraphiQL, GraphQL Playground, Voyager, 스키마 인트로스펙션


---

## 11) 응답 예시(질의/트랜잭션)

```graphql
# 블록 조회
query($gt:Int!, $lt:Int!){
  getBlocks(where:{height:{gt:$gt, lt:$lt}}){
    hash height time num_txs total_txs
  }
}
```

```graphql
# 트랜잭션 조회
query($gt:Int!, $lt:Int!){
  getTransactions(where:{block_height:{gt:$gt, lt:$lt}}){
    index hash success block_height
    gas_fee { amount denom }
    response {
      events {
        ... on GnoEvent {
          type func pkg_path
          attrs { key value }
        }
      }
    }
  }
}
```
**근거**: 선택 집합에 명시한 필드만 JSON에 포함. 인라인 프래그먼트로 구체 타입 선택.


---

## 12) Go에서의 견고한 호출 예시

```go
// 수신자 제네릭(타입 고정) 또는 비제네릭 + 함수 제네릭 중 택1.
// 여기서는 수신자 제네릭 버전으로 표기.
type Client[T any] struct {
  Endpoint string
  httpc    *http.Client
}

func NewClient[T any](endpoint string) *Client[T] {
  return &Client[T]{
    Endpoint: endpoint,
    httpc:    &http.Client{Timeout: 20 * time.Second},
  }
}

type gqlReq struct {
  Query     string                 `json:"query"`
  Variables map[string]interface{} `json:"variables,omitempty"`
}

type gqlResp[T any] struct {
  Data   T             `json:"data"`
  Errors []interface{} `json:"errors"`
}

func (c *Client[T]) Do(ctx context.Context, query string, vars map[string]interface{}, out *T) error {
  if out == nil { return errors.New("out must not be nil") }

  body, err := json.Marshal(gqlReq{Query: query, Variables: vars})
  if err != nil { return fmt.Errorf("marshal: %w", err) }

  req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body))
  if err != nil { return fmt.Errorf("new request: %w", err) }
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Accept", "application/json")

  res, err := c.httpc.Do(req)
  if err != nil { return fmt.Errorf("do: %w", err) }
  defer res.Body.Close()

  raw, err := io.ReadAll(res.Body)
  if err != nil { return fmt.Errorf("read: %w", err) }

  if res.StatusCode < 200 || res.StatusCode >= 300 {
    return fmt.Errorf("http %d: %s", res.StatusCode, sample(raw, 600))
  }

  mt, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
  if mt != "" && mt != "application/json" && mt != "application/graphql-response+json" {
    return fmt.Errorf("unexpected content-type %q: %s", res.Header.Get("Content-Type"), sample(raw, 600))
  }

  var r gqlResp[T]
  if err := json.Unmarshal(raw, &r); err != nil {
    return fmt.Errorf("decode: %w; body: %s", err, sample(raw, 600))
  }
  if len(r.Errors) > 0 {
    return fmt.Errorf("graphql errors: %+v", r.Errors)
  }
  *out = r.Data
  return nil
}

func sample(b []byte, n int) string {
  s := strings.TrimSpace(string(b))
  if len(s) > n { return s[:n] + "...(truncated)" }
  return s
}
```
**근거**
- `Accept` 미설정 시 HTML 반환 가능 → 명시로 방지.
- Content-Type/Status/본문 샘플 검사로 리다이렉트/프록시/플레이그라운드 응답 즉시 식별.
- GraphQL `errors` 존재 시 실패 처리.


---

## 13) 실전 체크리스트

- [ ] 엔드포인트가 **플레이그라운드(HTML)** 인지 **API(JSON)** 인지 구분했다.
- [ ] `Content-Type`/`Accept`를 올바르게 설정했다.
- [ ] 상태 코드, `Content-Type` 방어 로직이 있다.
- [ ] GraphQL `errors[]`를 항상 검사한다.
- [ ] 페이지네이션 인자 이름(`limit/offset` vs `take/skip` vs `first/after`)을 **스키마 기준**으로 맞췄다.
- [ ] 정렬(order_by)을 명시한다.
- [ ] N+1 방지용 배처/캐시 전략(DataLoader 등)을 갖췄다.
- [ ] `@deprecated`와 Nullability 정책으로 스키마 진화를 관리한다.
- [ ] 최신 윈도우로 탐침 후 범위를 확정한다(초기 블록 구간 공백 대비).
- [ ] 서버/클라이언트 로깅으로 원문 본문 샘플 확인 가능하게 했다.


---

## 14) 용어 정리 초간단

- **SDL**: 스키마 정의 언어
- **Selection Set**: 반환 필드 목록
- **Fragment**: 재사용 가능한 필드 묶음(또는 타입별 분기)
- **Directive**: 실행 시점 조건/메타
- **Non-Null(!)**: 필수값
- **Introspection**: 스키마 쿼리(운영 환경에서 제한 가능)
- **Relay Connection**: 커서 페이지네이션 규약
- **Persisted Query**: 해시 기반 고정 쿼리
- **DataLoader**: 배칭/캐시로 N+1 해결
