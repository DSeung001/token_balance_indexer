package main

import (
	"context"
	"flag"
	"gn-indexer/internal/config"
	"gn-indexer/internal/queue"
	"gn-indexer/internal/repository"
	"gn-indexer/internal/service"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, continuing...")
	}
}

func main() {
	// flag: command line standardization
	var (
		batchSize = flag.Int("batch", 10, "batch size for processing events")
		manual    = flag.Bool("manual", false, "manual batch processing mode")
	)
	flag.Parse()

	// database connection
	dbConfig := config.NewDatabaseConfig()
	gormDb, err := dbConfig.Connect()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Load queue configuration
	queueConfig := &queue.QueueConfig{
		QueueName:          getEnv("SQS_QUEUE_NAME", "token-events"),
		EndpointURL:        getEnv("LOCALSTACK_EDGE_PORT", "http://localhost:4566"),
		Region:             getEnv("AWS_DEFAULT_REGION", "ap-northeast-2"),
		AccessKeyID:        getEnv("AWS_ACCESS_KEY_ID", "test"),
		SecretAccessKey:    getEnv("AWS_SECRET_ACCESS_KEY", "test"),
		MaxReceiveMessages: *batchSize,
		VisibilityTimeout:  30,
	}

	// create repositories directly
	balanceRepo := repository.NewBalanceRepository(gormDb)
	tokenRepo := repository.NewTokenRepository(gormDb)

	// create queue
	eventQueue, err := queue.NewSQSQueue(queueConfig)
	if err != nil {
		log.Fatalf("failed to create SQS queue: %v", err)
	}
	defer eventQueue.Close()

	// create services
	balanceService := service.NewBalanceService(balanceRepo, tokenRepo)
	eventProcessor := service.NewEventProcessorService(eventQueue, balanceService)

	if *manual {
		// Manual batch processing mode - process one batch and exit
		log.Printf("starting manual batch processing mode with batch size: %d", *batchSize)

		// Process exactly one batch
		processedCount, err := eventProcessor.ProcessSingleBatch(ctx, *batchSize)
		if err != nil {
			log.Fatalf("manual batch processing failed: %v", err)
		}

		log.Printf("manual batch processing completed successfully. Processed %d events", processedCount)
		return
	} else {
		// Continuous event processing mode (default) - keep processing until interrupted
		log.Printf("starting continuous event processing mode with batch size: %d", *batchSize)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Start event processing in a goroutine
		go func() {
			if err := eventProcessor.Start(ctx); err != nil {
				log.Printf("event processing failed: %v", err)
			}
		}()

		// Wait for signal
		sig := <-sigChan
		log.Printf("received signal %v, shutting down gracefully...", sig)

		// Cancel context to stop all operations
		cancel()

		// Close queue connection
		if err := eventQueue.Close(); err != nil {
			log.Printf("error closing queue: %v", err)
		}

		log.Println("shutdown completed")
		log.Println("")
		log.Println("Usage:")
		log.Println("  --manual: Manual batch processing mode (process one batch and exit)")
		log.Println("  --batch <size>: Set batch size (default: 10)")
		log.Println("  No flags: Continuous event processing mode (default behavior)")
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
