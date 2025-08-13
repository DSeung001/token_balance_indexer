// cmd/syncer/main.go
package main

import (
	"context"
	"fmt"
	"log"

	"gn-indexer/internal/indexer"
)

func main() {
	ctx := context.Background()
	cli := indexer.NewClient("https://indexer.onbloc.xyz/graphql")

	// 1) 블록 샘플
	var bd indexer.BlocksData
	if err := cli.Do(ctx, indexer.QBlocks, map[string]interface{}{"gt": 0, "lt": 1000}, &bd); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("blocks fetched: %d\n", len(bd.GetBlocks))

	// 2) 트랜잭션 샘플
	var td indexer.TxsData
	if err := cli.Do(ctx, indexer.QTxs, map[string]interface{}{"gt": 0, "lt": 1000, "imax": 1000}, &td); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("txs fetched: %d\n", len(td.GetTransactions))
}
