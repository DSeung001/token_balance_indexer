package api

import (
	"gn-indexer/internal/repository"
	"log"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP API server
type Server struct {
	router         *gin.Engine
	balanceHandler *BalanceHandler
}

// NewServer creates a new API server
func NewServer(
	balanceRepo repository.BalanceRepository,
	tokenRepo repository.TokenRepository,
	transferRepo repository.TransferRepository,
) *Server {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Create handlers
	balanceHandler := NewBalanceHandler(balanceRepo, tokenRepo, transferRepo)

	server := &Server{
		router:         router,
		balanceHandler: balanceHandler,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "gn-indexer-api"})
	})

	// API routes
	s.router.GET("/tokens/balances", s.balanceHandler.GetBalancesByAddress)
	s.router.GET("/tokens/:tokenPath/balances", s.balanceHandler.GetBalancesByTokenAndAddress)
	s.router.GET("/tokens/transfer-history", s.balanceHandler.GetTransferHistory)
}

// Run starts the HTTP server
func (s *Server) Run(addr string) error {
	log.Printf("Starting API server on %s", addr)
	return s.router.Run(addr)
}

// GetRouter returns the underlying Gin router (useful for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
