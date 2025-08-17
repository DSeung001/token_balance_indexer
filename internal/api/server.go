package api

import (
	"gn-indexer/internal/repository"
	"log"
	"net/http"
	"regexp"
	"strings"

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
	// Health
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "gn-indexer-api"})
	})

	// Concrete routes under /tokens
	s.router.GET("/tokens/balances", s.balanceHandler.GetBalancesByAddress)
	s.router.GET("/tokens/transfer-history", s.balanceHandler.GetTransferHistory)

	// /tokens/:tokenPath/balances handling
	s.router.NoRoute(s.tokenRouteFallback())
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

// tokenRouteFallback tries to recover tokenPath from a raw path like
func (s *Server) tokenRouteFallback() gin.HandlerFunc {
	// Precompile regex once
	var re = regexp.MustCompile(`^/tokens/(.+)/balances$`)

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Normalize backslashes to forward slashes (in case client sent '\')
		path = strings.ReplaceAll(path, `\`, `/`)

		// Try to match /tokens/{tokenPath}/balances
		m := re.FindStringSubmatch(path)
		if len(m) == 2 {
			tokenPath := strings.Trim(m[1], "/")
			if tokenPath != "" {
				// Inject tokenPath so the original handler can read c.Param("tokenPath")
				c.Params = append(c.Params, gin.Param{Key: "tokenPath", Value: tokenPath})

				// Call the existing handler
				s.balanceHandler.GetBalancesByTokenAndAddress(c)
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "tokenPath parameter is required"})
			return
		}

		// Not our pattern → keep 404
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "route not found",
			"path":   c.Request.URL.Path,
			"detail": "check path or query params",
		})
	}
}
