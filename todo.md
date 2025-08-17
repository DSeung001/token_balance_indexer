1. 임의 데이터 넣으면 api 잘 동작하는 지 (sql문을 실행해서 직접 시스템 테스트 할 수 있게끔)
2. go run ./cmd/block-syncer -integrity, go run ./cmd/block-syncer -realtime을 했을 때 consumer 실행안하면 누락되는 지 체크
   A: 누락됨
