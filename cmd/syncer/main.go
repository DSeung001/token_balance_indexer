package main

import (
	"context"
	"flag"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gn-indexer/internal/db"
	"gn-indexer/internal/indexer"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, continuing...")
	}
}

func main() {
	const gqlEndpoint = "https://indexer.onbloc.xyz/graphql/query"
	const wsEndpoint = "wss://indexer.onbloc.xyz/graphql/query"

	// flag: command line standardization
	var (
		fromHeight = flag.Int("from", 0, "from block height")
		toHeight   = flag.Int("to", 0, "to block height")
		realtime   = flag.Bool("realtime", false, "start realtime sync")
	)
	flag.Parse()

	// db connect
	gormDb := db.MustConnect()

	ctx := context.Background()

	// http client
	cliBlocks := indexer.NewGraphQLClient[indexer.BlocksData](gqlEndpoint)
	cliTxs := indexer.NewGraphQLClient[indexer.TxsData](gqlEndpoint)

	// websocket client
	subClient := indexer.NewSubscriptionClient(wsEndpoint)

	// sync
	syncer := indexer.NewSyncer(cliBlocks, cliTxs, subClient, gormDb)

	if *realtime {
		// real-time synchronization
		log.Println("starting realtime sync mode...")

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// signal handling for gracefull shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigChan
			log.Printf("received signal %v, shutting down gracefully...", sig)
			cancel()
		}()

		if err := syncer.StartRealtimeSync(ctx); err != nil {
			log.Fatalf("realtime sync failed: %v", err)
		}

		log.Println("realtime sync stopped")
	} else {
		// test/dev
		if *toHeight == 0 {
			*toHeight = 1000 // default
		}

		log.Printf("starting one-time sync from height %d to %d", *fromHeight, *toHeight)

		if err := syncer.SyncRange(ctx, *fromHeight, *toHeight); err != nil {
			log.Fatalf("failed to sync range: %v", err)
		}

		log.Println("sync completed successfully")
	}
}
