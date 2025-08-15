package main

import (
	"context"
	"flag"
	"github.com/joho/godotenv"
	"log"

	"gn-indexer/internal/db"
	"gn-indexer/internal/indexer"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, continuing...")
	}
}

func main() {
	// flag: command line standardization
	var (
		fromHeight = flag.Int("from", 0, "from block height")
		toHeight   = flag.Int("to", 0, "to block height")
		continuous = flag.Bool("continuous", false, "continuous mode")
	)
	flag.Parse()

	// db connect
	gormDb := db.MustConnect()

	ctx := context.Background()

	// cliBlocks: Block client
	cliBlocks := indexer.NewClient[indexer.BlocksData]("https://indexer.onbloc.xyz/graphql/query")

	// cliTxs: Transaction client
	cliTxs := indexer.NewClient[indexer.TxsData]("https://indexer.onbloc.xyz/graphql/query")

	// new syncer
	syncer := indexer.NewSyncer(cliBlocks, cliTxs, gormDb)

	if *continuous {
		// Todo continous mode
		log.Println("continuous mode not implemented yet")
	} else {
		// one time sync
		if *toHeight == 0 {
			*toHeight = 1000 // default
		}

		if err := syncer.SyncRange(ctx, *fromHeight, *toHeight); err != nil {
			log.Fatalf("failed to sync range: %v", err)
		}
		log.Println("sync completed successfully")
	}
}
