# Balance API 테스트 가이드
트랜잭션에서 가져온 이벤트에서 Type이 Transfer, func가 Mint, Burn, Transfer인 것을 찾을 수 없어서 API 테스트를 위해 추가하였습니다.<br/>
아래와 같이 func은 ""이고 type이 "Transfer"로 데이터를 받아왔지만 가이드 문서의 토큰 전송 이벤트와 동일하지 않아 파싱 부분은 가이드대로 개발하고 보류해두었습니다.

데이터 때문에 [balance_api_insert_mock_data.sql](mock-data/balance_api_insert_mock_data.sql)가 안될 경우 [balance_api_cleanup_mock_data.sql](mock-data/balance_api_cleanup_mock_data.sql)을 먼저 실행 시킨 뒤 실행
<br/>
※ 주의: 테스트를 위한 sql 문이므로 사용의 주의 필요 ※

```json
{
  "events": [
    {
      "func": "",
      "type": "Transfer",
      "attrs": [
        {
          "key": "token",
          "value": "gno.land/r/gnoswap/v1/test_token/baz.BAZ"
        },
        {
          "key": "from",
          "value": ""
        },
        {
          "key": "to",
          "value": "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"
        },
        {
          "key": "value",
          "value": "100000000000000"
        }
      ],
      "pkg_path": "gno.land/p/demo/grc/grc20"
    },
    {
      "func": "",
      "type": "register",
      "attrs": [
        {
          "key": "pkgpath",
          "value": "gno.land/r/gnoswap/v1/test_token/baz"
        },
        {
          "key": "slug",
          "value": ""
        }
      ],
      "pkg_path": "gno.land/r/demo/grc20reg"
    },
    {
      "func": "",
      "type": "StorageDeposit",
      "attrs": [
        {
          "key": "Deposit",
          "value": "220300ugnot"
        },
        {
          "key": "Storage",
          "value": "2203 bytes"
        }
      ],
      "pkg_path": "gno.land/r/demo/grc20reg"
    },
    {
      "func": "",
      "type": "StorageDeposit",
      "attrs": [
        {
          "key": "Deposit",
          "value": "1548500ugnot"
        },
        {
          "key": "Storage",
          "value": "15485 bytes"
        }
      ],
      "pkg_path": "gno.land/r/gnoswap/v1/test_token/baz"
    }
  ]
}
```

## 빠른 시작

### 1. 환경 준비
```bash
# 기존 환경 정리
docker-compose down -v

# 새로 빌드 및 실행
docker-compose up -d --build
```

### 2. 데이터 준비
```bash
# PostgreSQL에 연결 후 더미 데이터 삽입
psql -h localhost -U username -d database_name -f test/mock-data/balance_api_insert_mock_data.sql
```

### 3. API 서버 실행
```bash
go run cmd/balance-api/main.go
```

## API 테스트 요청

### 1. 특정 주소의 모든 토큰 잔액 조회

#### 테스트 계정 1 (g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5)
```bash
curl "http://localhost:8080/tokens/balances?address=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
```
**예상 결과**: WUGNOT(500000), GNS(1000000), BAR(300000) - 총 3개 토큰

#### 테스트 계정 2 (g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae)
```bash
curl "http://localhost:8080/tokens/balances?address=g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae"
```
**예상 결과**: WUGNOT(500000), GNS(2000000), BAR(700000) - 총 3개 토큰

#### 테스트 계정 3 (g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d)
```bash
curl "http://localhost:8080/tokens/balances?address=g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"
```
**예상 결과**: WUGNOT(1000000), GNS(1500000), FOO(2500000) - 총 3개 토큰

### 2. 특정 토큰의 모든 주소 잔액 조회

#### WUGNOT 토큰 (gno.land/r/demo/wugnot)
```bash
curl "http://localhost:8080/tokens/gno.land/r/demo/wugnot/balances"
```
**예상 결과**: 3개 계정의 WUGNOT 잔액

#### GNS 토큰 (gno.land/r/gnoswap/v1/gns)
```bash
curl "http://localhost:8080/tokens/gno.land/r/gnoswap/v1/gns/balances"
```
**예상 결과**: 3개 계정의 GNS 잔액

#### BAR 토큰 (gno.land/r/gnoswap/v1/test_token/bar)
```bash
curl "http://localhost:8080/tokens/gno.land/r/gnoswap/v1/test_token/bar/balances"
```
**예상 결과**: 2개 계정의 BAR 잔액

#### FOO 토큰 (gno.land/r/gnoswap/v1/test_token/foo)
```bash
curl "http://localhost:8080/tokens/gno.land/r/gnoswap/v1/test_token/foo/balances"
```
**예상 결과**: 1개 계정의 FOO 잔액

### 3. 특정 주소의 전송 내역 조회

#### 테스트 계정 1의 전송 내역
```bash
curl "http://localhost:8080/tokens/transfer-history?address=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
```
**예상 결과**: 5개 전송 내역 (보내기: BAR 150000, WUGNOT 100000 / 받기: BAR 300000, GNS 1000000, WUGNOT 500000)

#### 테스트 계정 2의 전송 내역
```bash
curl "http://localhost:8080/tokens/transfer-history?address=g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae"
```
**예상 결과**: 5개 전송 내역 (보내기: GNS 500000 / 받기: WUGNOT 100000, BAR 700000, GNS 2000000, WUGNOT 500000)

#### 테스트 계정 3의 전송 내역
```bash
curl "http://localhost:8080/tokens/transfer-history?address=g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"
```
**예상 결과**: 5개 전송 내역 (받기만: BAR 150000, GNS 500000, FOO 2500000, GNS 1500000, WUGNOT 1000000)

### 4. 특정 토큰의 특정 주소 잔액 조회

#### 테스트 계정 1의 WUGNOT 잔액
```bash
curl "http://localhost:8080/tokens/gno.land/r/demo/wugnot/balances?address=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
```
**예상 결과**: WUGNOT 500000

#### 테스트 계정 2의 GNS 잔액
```bash
curl "http://localhost:8080/tokens/gno.land/r/gnoswap/v1/gns/balances?address=g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae"
```
**예상 결과**: GNS 2000000

#### 테스트 계정 3의 FOO 잔액
```bash
curl "http://localhost:8080/tokens/gno.land/r/gnoswap/v1/test_token/foo/balances?address=g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"
```
**예상 결과**: FOO 2500000

## 테스트 데이터

### 테스트 계정 주소
- `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5` - 테스트 계정 1
- `g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae` - 테스트 계정 2  
- `g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d` - 테스트 계정 3

### 테스트 토큰
- `gno.land/r/demo/wugnot` (WUGNOT)
- `gno.land/r/gnoswap/v1/gns` (GNS)
- `gno.land/r/gnoswap/v1/test_token/bar` (BAR)
- `gno.land/r/gnoswap/v1/test_token/foo` (FOO)

### 더미 데이터 요약
- **블록**: 7개 (높이 1000-1006)
- **트랜잭션**: 12개 (tx_hash_001 ~ tx_hash_012)
- **토큰**: 4개
- **잔액**: 9개 계정-토큰 조합
- **전송 내역**: 12개 (Mint 9개 + Transfer 3개)

## 테스트 완료 후 정리
```bash
# 더미 데이터 정리
psql -h localhost -U username -d database_name -f test/mock-data/balance_api_cleanup_mock_data.sql
```
