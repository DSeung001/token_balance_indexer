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

	// cliBlocks: Block client
	cliBlocks := indexer.NewClient[indexer.BlocksData]("https://indexer.onbloc.xyz/graphql/query")
	var bd indexer.BlocksData
	if err := cliBlocks.Do(ctx, indexer.QBlocks, map[string]interface{}{"gt": 0, "lt": 1000}, &bd); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("blocks fetched: %d\n", len(bd.GetBlocks))

	// cliTxs: Transaction client
	cliTxs := indexer.NewClient[indexer.TxsData]("https://indexer.onbloc.xyz/graphql/query")
	var td indexer.TxsData
	if err := cliTxs.Do(ctx, indexer.QTxs, map[string]interface{}{"gt": 0, "lt": 1000, "imax": 1000}, &td); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("txs fetched: %d\n", len(td.GetTransactions))
}
