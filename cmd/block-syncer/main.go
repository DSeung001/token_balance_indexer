package main

import (
	"context"
	"flag"
	"github.com/joho/godotenv"
	"gn-indexer/internal/config"
	"gn-indexer/internal/indexer"
	"gn-indexer/internal/repository"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// database connection
	connConfig := config.NewDatabaseConfig()
	gormDb, err := connConfig.Connect()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// http client
	cliBlocks := indexer.NewGraphQLClient[indexer.BlocksData](gqlEndpoint)
	cliTxs := indexer.NewGraphQLClient[indexer.TxsData](gqlEndpoint)

	// websocket client
	subClient := indexer.NewSubscriptionClient(wsEndpoint)

	// create repositories directly
	blockRepo := repository.NewBlockRepository(gormDb)
	transactionRepo := repository.NewTransactionRepository(gormDb)

	// sync with repositories
	syncer := indexer.NewSyncer(
		cliBlocks,
		cliTxs,
		subClient,
		blockRepo,
		transactionRepo,
	)

	if *realtime {
		// real-time synchronization
		log.Println("starting realtime sync mode...")

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Start realtime sync in a goroutine
		go func() {
			if err := syncer.StartRealtimeSync(ctx); err != nil {
				log.Printf("realtime sync failed: %v", err)
			}
		}()

		// Wait for signal
		sig := <-sigChan
		log.Printf("received signal %v, shutting down gracefully...", sig)

		// Cancel context to stop all operations
		cancel()

		// Close WebSocket connection
		if err := subClient.Close(); err != nil {
			log.Printf("error closing websocket: %v", err)
		}

		log.Println("shutdown completed")
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
