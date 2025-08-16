package main

import (
	"context"
	"flag"
	"gn-indexer/internal/config"
	"gn-indexer/internal/indexer"
	"gn-indexer/internal/repository"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gn-indexer/internal/service"

	"github.com/joho/godotenv"
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
		integrity  = flag.Bool("integrity", false, "check and fix data integrity from height 1")
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
	cliBlocks := indexer.NewGraphQLClient[indexer.BlocksDataArr](gqlEndpoint)
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
		// real-time synchronization using orchestrator
		log.Println("starting realtime sync mode...")

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Create and start block sync manager
		syncService := service.NewBlockSyncService(syncer, subClient, wsEndpoint)

		// Start parallel sync in a goroutine
		go func() {
			if err := syncService.StartParallelSync(ctx); err != nil {
				log.Printf("parallel sync failed: %v", err)
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
	} else if *integrity {
		// Data integrity check and fix (from height 1)
		log.Println("starting data integrity check and fix from height 1...")

		dataIntegritySvc := service.NewDataIntegrityService(syncer)

		if err := dataIntegritySvc.CheckAndFixDataIntegrity(ctx); err != nil {
			log.Fatalf("data integrity check and fix failed: %v", err)
		}

		log.Println("data integrity check and fix completed successfully")
		return
	} else if *fromHeight > 0 || *toHeight > 0 {
		// Specific range synchronization
		if *fromHeight == 0 {
			*fromHeight = 1 // Set default value to 1
		}
		if *toHeight == 0 {
			*toHeight = 1000 // default
		}

		// Add validation check
		if *fromHeight > *toHeight {
			log.Fatalf("invalid range: from height (%d) cannot be greater than to height (%d)", *fromHeight, *toHeight)
		}

		log.Printf("starting one-time sync from height %d to %d", *fromHeight, *toHeight)

		if err := syncer.SyncRange(ctx, *fromHeight, *toHeight); err != nil {
			log.Fatalf("failed to sync range: %v", err)
		}

		log.Println("sync completed successfully")
	} else {
		// Usage guide
		log.Println("Usage:")
		log.Println("  --integrity: Check and fix data integrity from height 1")
		log.Println("  --realtime: Start realtime sync mode")
		log.Println("  --from <height> --to <height>: Sync specific range (from defaults to 1)")
		log.Println("  No flags: Show this help message")
	}
}
