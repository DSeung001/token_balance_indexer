package api

import (
	"net/http"
	"strconv"

	"gn-indexer/internal/repository"

	"github.com/gin-gonic/gin"
)

// BalanceHandler handles balance-related API requests
type BalanceHandler struct {
	balanceRepo repository.BalanceRepository
	tokenRepo   repository.TokenRepository
}

// NewBalanceHandler creates a new balance handler
func NewBalanceHandler(
	balanceRepo repository.BalanceRepository,
	tokenRepo repository.TokenRepository,
) *BalanceHandler {
	return &BalanceHandler{
		balanceRepo: balanceRepo,
		tokenRepo:   tokenRepo,
	}
}

// GetBalancesByAddress handles GET /tokens/balances?address={address}
func (h *BalanceHandler) GetBalancesByAddress(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "address parameter is required",
		})
		return
	}

	balances, err := h.balanceRepo.GetBalancesByAddress(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get balances: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":  address,
		"balances": balances,
		"count":    len(balances),
	})
}

// GetBalancesByToken handles GET /tokens/{tokenPath}/balances?address={address}
func (h *BalanceHandler) GetBalancesByToken(c *gin.Context) {
	tokenPath := c.Param("tokenPath")
	address := c.Query("address")

	if tokenPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tokenPath parameter is required",
		})
		return
	}

	if address == "" {
		// Get all balances for the token
		balances, err := h.balanceRepo.GetBalancesByToken(c.Request.Context(), tokenPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get balances: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tokenPath": tokenPath,
			"balances":  balances,
			"count":     len(balances),
		})
		return
	}

	// Get specific balance for token and address
	balance, err := h.balanceRepo.GetBalance(c.Request.Context(), tokenPath, address)
	if err != nil {
		if err == repository.ErrBalanceNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":     "balance not found",
				"tokenPath": tokenPath,
				"address":   address,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get balance: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokenPath": tokenPath,
		"address":   address,
		"balance":   balance,
	})
}

// GetTransferHistory handles GET /tokens/transfer-history?address={address}
func (h *BalanceHandler) GetTransferHistory(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "address parameter is required",
		})
		return
	}

	// Get page and limit parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// TODO: Implement transfer history repository
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"address":   address,
		"page":      page,
		"limit":     limit,
		"message":   "Transfer history endpoint - implementation pending",
		"transfers": []gin.H{},
		"total":     0,
	})
}
