# [BE 과제] 블록체인 이벤트 기반 Token Balance Indexer 개발 

## 📋 과제 설명

블록체인 네트워크에서 발생하는 **블록 및 트랜잭션 이벤트를 기반으로**, 토큰별 계정 잔액 정보를 **정합성 있게 계산 및 저장**하고, 이를 API로 조회할 수 있는 **MSA 기반 인덱싱 시스템**을 구축합니다.

이 과제를 통해 다음 세 가지 역량을 확인하고자 합니다:

1. **새로운 도메인 학습 능력**: 블록체인 동작 원리, Tranaction 및 Event 구조에 대한 이해
2. **데이터 비동기 처리 역량**: 실시간 이벤트 수신 및 누락 방지 설계
3. **Queue 기반 MSA 아키텍처 설계 능력**: 생산자-소비자 분리와 서비스 간 메시지 전달

## **🧭** 사전 가이드

1. Gno.land 블록체인 데이터
    1. 블록
        - 블록은 일정 시간 동안 발생한 트랜잭션을 모아 저장하는 데이터 구조입니다. 각 블록은 이전 블록의 해시를 포함하여 체인 형태로 연결됩니다.
        - 하나의 블록에 여러개의 트랜잭션이 포함될 수 있습니다.
        - 블록 생성과정
            1. **트랜잭션 수집**: 네트워크에서 발생한 트랜잭션을 수집합니다.
            2. **블록 구성**: 수집된 트랜잭션을 블록에 포함시킵니다.
            3. **해시 계산**: 블록의 데이터를 해시 함수에 입력하여 고유한 해시 값을 생성합니다.
            4. **블록 연결**: 생성된 해시를 다음 블록에 포함시켜 체인을 형성합니다.
        1. 블럭 구성 요소
            - **블록 해시**: 블록의 고유 식별자
            - **이전 블록 해시**: 이전 블록의 해시 값
            - **타임스탬프**: 블록 생성 시간
            - **트랜잭션 목록**: 블록에 포함된 트랜잭션들
            - …
    2. 트랜잭션
        - 트랜잭션은 블록체인에서 상태를 변경하는 작업 단위입니다. 예를 들어, 토큰 전송, 스마트 컨트랙트 호출 등이 트랜잭션에 해당합니다.
        - 하나의 트랜잭션에 여러개의 트랜잭션 메세지와 트랜잭션 이벤트가 포함될 수 있습니다.
        - **트랜잭션의 구성 요소**
            - **송신자 주소**: 트랜잭션을 생성한 계정의 주소
            - **수수료**: 트랜잭션 처리에 대한 보상
            - **서명**: 트랜잭션의 진위 확인을 위한 디지털 서명
            - …
    3. 트랜잭션 메세지
        - 트랜잭션 메시지는 블록체인에서 실행되는 명령어로, 블록체인 상태를 변경하는 데 사용됩니다.
        - **주요 메시지 유형**
            - **BankMsgSend**: 토큰 전송
            - **MsgAddPackage**: 새로운 패키지(컨트렉트) 추가
            - **MsgCall**: 기존 패키지(컨트렉트)의 함수 호출
            - **MsgRun**: 스크립트 실행
    4. 트랜잭션 이벤트
        - 이벤트는 트랜잭션 실행 중 발생한 특정 동작을 기록하는 로그입니다. 이벤트는 블록체인 상태 변경을 추적하고, 오프체인 시스템과의 연동에 활용됩니다.
        - **이벤트의 구성 요소**
            - **이벤트 타입**: 이벤트 타입
            - **함수 이름**: 이벤트를 발생시킨 함수
            - **패키지 경로**: 이벤트를 발생시킨 패키지의 경로
            - **속성**: 이벤트에 대한 추가 정보, key-value로 구성 (예: 전송 주소, 금액 등)
        - 이벤트 예시 데이터

            ```json
            {
              "type": "Transfer",
              "func": "Mint",
              "pkg_path": "gno.land/r/gnoswap/v1/test_token/bar",
              "attrs": [
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
              ]
            }
            ```

    5. 토큰 전송 이벤트
        - 토큰 Mint 이벤트 (토큰 발행)
            1. func: `Mint`
            2. attrs: `from`, `to`, `value` 의 값이 존재
                1. `from` : 공백 문자열
                2. `to` : bech32 형태의 주소
                3. `value` : 숫자형태
                4. 형태가 다른 경우 토큰 전송 이벤트로 판단하지 않음

            ```json
            {
              "type": "Transfer",
              "func": "Mint",
              "pkg_path": "gno.land/r/gnoswap/v1/test_token/bar",
              "attrs": [
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
              ]
            }
            ```

        - 토큰 Burn 이벤트 (토큰 소멸)
            1. func: `Burn`
            2. attrs: `from`, `to`, `value` 의 값이 존재
                1. `from` : bech32 형태의 주소
                2. `to` : 공백 문자열
                3. `value` : 숫자형태
                4. 형태가 다른 경우 토큰 전송 이벤트로 판단하지 않음

            ```json
            {
              "type": "Transfer",
              "func": "Burn",
              "pkg_path": "gno.land/r/gnoswap/v1/test_token/bar",
              "attrs": [
                {
                  "key": "from",
                  "value": "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"
                },
                {
                  "key": "to",
                  "value": ""
                },
                {
                  "key": "value",
                  "value": "100000000000000"
                }
              ]
            }
            ```

        - 토큰 Transfer 이벤트 (토큰 전송)
            1. func: `Transfer`
            2. attrs: `from`, `to`, `value` 의 값이 존재
                1. `from` : bech32 형태의 주소
                2. `to` : bech32 형태의 주소
                3. `value` : 숫자형태
                4. 형태가 다른 경우 토큰 전송 이벤트로 판단하지 않음

            ```json
            {
              "type": "Transfer",
              "func": "Transfer",
              "pkg_path": "gno.land/r/gnoswap/v1/test_token/bar",
              "attrs": [
                {
                  "key": "from",
                  "value": "g16a7etgm9z2r653ucl36rj0l2yqcxgrz2jyegzx"
                },
                {
                  "key": "to",
                  "value": "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"
                },
                {
                  "key": "value",
                  "value": "100000000000000"
                }
              ]
            }
            ```

2. tx-indexer를 통한 데이터 수신방법
    1. GraphQL Dashboard
        1. Dashboard Url (1번 인덱서에 트랜잭션 데이터가 없다면 2번에서 확인):
            1. 인덱서1:  https://dev-indexer.api.gnoswap.io/graphql
            2. 인덱서2: https://indexer.onbloc.xyz/graphql
        2. 좌측 상단 문서아이콘을 통해 Docs 확인 가능

           ![image.png](attachment:7bec3538-a0e0-4c5e-b85c-2802f10bc645:image.png)

    2. GraphQL을 통한 블록 데이터 조회
        - GraphQL 쿼리

            ```graphql
            {
              getBlocks(
                where: {
                  height: {
                    gt: 0,  		# Greater than 1989.
                    lt: 1000   	# Less than 2001.
                  }
                }
              ) {
                # Fields to retrieve for each block.
                hash         # The unique hash identifier of the block.
                height       # The block's height in the blockchain.
                time         # Timestamp when the block was created.
                total_txs    # Total number of transactions up to this block.
                num_txs      # Number of transactions in the block.
                total_txs		 # Number of total transactions.
              }
            }
            ```

    3. GraphQL을 통한 트랜잭션 데이터 조회
        - GraphQL 쿼리

            ```graphql
            {
              getTransactions(
                where: {
                  block_height: {
                    gt: 0,     # block height > 0
                    lt: 1000   # block height < 1000
                  }, 
                  index: {
                    lt: 1000   # transaction index < 1000
                  }
                }
              ) {
                index
                hash
                success
                block_height
                gas_wanted
                gas_used
                memo
                gas_fee {
                  amount
                  denom
                }
                messages {
                  route
                  typeUrl
                  value {
                    ... on BankMsgSend {
                      from_address
                      to_address
                      amount
                    }
                    ... on MsgAddPackage {
                      creator
                      deposit
                      package {
                        name
                        path
                        files {
                          name
                          body
                        }
                      }
                    }
                    ... on MsgCall {
                      pkg_path
                      func
                      send
                      caller
                      args
                    }
                    ... on MsgRun {
                      caller
                      send
                      package {
                        name
                        path
                        files {
                          name
                          body
                        }
                      }
                    }
                  }
                }
                response {
                  log
                  info
                  error
                  data
                  events {
                    ... on GnoEvent {
                      type
                      func
                      pkg_path
                      attrs {
                        key
                        value
                      }
                    }
                  }
                }
              }
            }
            ```

    4. GraphQL을 통한 블록 데이터 구독
        - GraphQL 쿼리

            ```graphql
            subscription{
              getBlocks(
                where: {}
              ) {
                # Fields to retrieve for each block.
                hash         # The unique hash identifier of the block.
                height       # The block's height in the blockchain.
                time         # Timestamp when the block was created.
                total_txs    # Total number of transactions up to this block.
                num_txs      # Number of transactions in the block.
                total_txs		 # Number of total transactions.
              }
            }
            ```



## 📄 참고 내용

- Gno.land Docs: https://docs.gno.land/
- Tx Indexer 저장소: https://github.com/gnolang/tx-indexer

## **🛠** 기능 요구사항

### **1. Block Synchronizer (Producer)**

- 블록 및 트랜잭션을 수신하여 PostgreSQL에 저장
    - 테이블 스키마는 자유롭게 설계
- 서버 시작 시 누락된 블록 범위를 스캔하여 백필 처리
- 트랜잭션 내 토큰 발행, 소멸, 전송 **이벤트**를 파싱하여 Queue(SQS)에 전송

### **2. Event Processor (Consumer)**

- Queue에서 수신한 이벤트 기반으로 잔액 계산 수행
    - 토큰 발행, 소멸, 전송 이벤트에 따라 잔액 계산
- 계산된 잔액 정보를 DBMS에 저장
- 대량 처리 또는 병렬 소비가 가능한 구조 고려

### **3. Balance API (REST API Server)**

- **[GET] /tokens/balances?address={address}**
    - 토큰들의 토큰 잔액 조회 API
    - **[파라미터]**
        - (Optional) `address` : 계정 주소
            - 값이 존재하는 경우: `address` 가 보유한 토큰의 잔액만 반환
            - 값이 없거나 비어있는 경우: 전체 계정의 토큰 잔액 반환
    - **[예상 응답]**

        ```json
        {
          "balances": [
            {
              "tokenPath": "gno.land/r/demo/wugnot",
              "amount": 1000000
            },
            {
              "tokenPath": "gno.land/r/gnoswap/v1/gns",
              "amount": 1000000
            }
          ]
        }
        ```

- **[GET] /tokens/{tokenPath}/balances?address={address}**
    - 특정 토큰의 토큰 잔액 조회 API
    - **[파라미터]**
        - (Optional) `address` : 계정 주소
            - 값이 존재하는 경우: `address` 가 보유한 특정 토큰의 잔액만 반환
            - 값이 없거나 비어있는 경우: 전체 계정의 특정 토큰 잔액 반환
    - **[예상 응답]**

        ```json
        {
          "accountBalances": [
            {
              "address": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
              "tokenPath": "gno.land/r/demo/wugnot",
              "amount": 500000
            },
            {
              "address": "g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae",
              "tokenPath": "gno.land/r/demo/wugnot",
              "amount": 500000
            }
          ]
        }
        ```

- **[GET] /tokens/transfer-history?address={address}**
    - 토큰 전송내역 조회 API
    - **[파라미터]**
        - (Optional) `adddress` : 계정 주소
            - 값이 존재하는 경우: `address` 가 fromAddress 또는 toAddress에 포함된 transfer 기록 반환
            - 값이 없거나 비어있는 경우: 전체 계정의 토큰내역 조회
    - **[예상 응답]**

        ```json
        {
          "transfers": [
            {
              "fromAddress": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
              "toAddress": "g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae",
              "tokenPath": "gno.land/r/demo/wugnot",
              "amount": 500000
            }
          ]
        }
        ```


### **4. 시스템 구성**

- MSA 서비스
    - Block Synchronizer (producer)
    - Event Processor (consumer)
    - Balance API (REST API Server)
- 서비스 간 통신은 Queue 기반
- Docker Compose 환경 구성

## 💻 기술 요구사항

- 개발언어: Golang
- DBMS: PostgreSQL, (Optional) Redis
- Queue: Local Stack (SQS)
- Library: Gorm(DB), Gin(Web Framework)

## 📝 평가 기준

1. 아키텍쳐: MSA 구조, 확장성/복원성 설계
2. 코드품질: 모듈 구조, 서비스 및 도메인 별 역할 분리
3. 기능 완성도: 데이터 정합성, 누락 이벤트 처리

## 📄 제출 문서

- 제출: GitHub 저장소 제출 후 링크 공유
- 필수 문서:
    - README.md: 실행 방법, 주요 설계 의도, flow-diagram 혹은 처리 흐름 등
    - schema.sql: DB 테이블 생성 DDL
    - 단위 테스트 코드 포함