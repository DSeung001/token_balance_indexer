package main

import (
	"flag"
	"gn-indexer/internal/api"
	"gn-indexer/internal/config"
	"gn-indexer/internal/repository"
	"log"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, continuing...")
	}
}

func main() {
	// Parse command line flags
	var (
		port = flag.String("port", "8080", "server port")
		host = flag.String("host", "127.0.0.1", "server host")
	)
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, continuing...")
	}

	// Database connection
	dbConfig := config.NewDatabaseConfig()
	gormDb, err := dbConfig.Connect()
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	// Create repositories
	balanceRepo := repository.NewBalanceRepository(gormDb)
	tokenRepo := repository.NewTokenRepository(gormDb)
	transferRepo := repository.NewTransferRepository(gormDb)

	// Create and start API server
	server := api.NewServer(balanceRepo, tokenRepo, transferRepo)

	addr := *host + ":" + *port
	log.Printf("Starting GN Indexer Balance API on %s", addr)
	log.Printf("Available endpoints:")
	log.Printf("  GET /health")
	log.Printf("  GET /tokens/balances?address={address}")
	log.Printf("  GET /tokens/{tokenPath}/balances?address={address}")
	log.Printf("  GET /tokens/transfer-history?address={address}")

	if err := server.Run(addr); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
