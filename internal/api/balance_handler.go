package api

import (
	"net/http"
	"strconv"

	"gn-indexer/internal/repository"
	"gn-indexer/internal/types"

	"github.com/gin-gonic/gin"
)

// BalanceHandler handles balance-related API requests
type BalanceHandler struct {
	balanceRepo  repository.BalanceRepository
	tokenRepo    repository.TokenRepository
	transferRepo repository.TransferRepository
}

// NewBalanceHandler creates a new balance handler
func NewBalanceHandler(
	balanceRepo repository.BalanceRepository,
	tokenRepo repository.TokenRepository,
	transferRepo repository.TransferRepository,
) *BalanceHandler {
	return &BalanceHandler{
		balanceRepo:  balanceRepo,
		tokenRepo:    tokenRepo,
		transferRepo: transferRepo,
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

	responseBalances := make([]types.TokenBalance, 0, len(balances))
	for _, balance := range balances {
		responseBalances = append(responseBalances, types.TokenBalance{
			TokenPath: balance.TokenPath,
			Amount:    balance.Amount,
		})
	}

	response := types.BalanceResponse{
		Balances: responseBalances,
	}

	c.JSON(http.StatusOK, response)
}

// GetBalancesByTokenAndAddress handles GET /tokens/{tokenPath}/balances?address={address}
func (h *BalanceHandler) GetBalancesByTokenAndAddress(c *gin.Context) {
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
		balances, err := h.balanceRepo.GetBalancesByTokenAndAddress(c.Request.Context(), tokenPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get balances: " + err.Error(),
			})
			return
		}

		accountBalances := make([]types.AccountBalance, 0, len(balances))
		for _, balance := range balances {
			accountBalances = append(accountBalances, types.AccountBalance{
				Address:   balance.Address,
				TokenPath: balance.TokenPath,
				Amount:    balance.Amount,
			})
		}

		response := types.AccountBalanceResponse{
			AccountBalances: accountBalances,
		}

		c.JSON(http.StatusOK, response)
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

	accountBalances := []types.AccountBalance{
		{
			Address:   balance.Address,
			TokenPath: balance.TokenPath,
			Amount:    balance.Amount,
		},
	}

	response := types.AccountBalanceResponse{
		AccountBalances: accountBalances,
	}

	c.JSON(http.StatusOK, response)
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

	// Transfer history
	transfers, err := h.transferRepo.GetByAddress(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get transfer history: " + err.Error(),
		})
		return
	}

	responseTransfers := make([]types.TransferRecord, 0, len(transfers))
	for _, transfer := range transfers {
		responseTransfers = append(responseTransfers, types.TransferRecord{
			FromAddress: transfer.FromAddress,
			ToAddress:   transfer.ToAddress,
			TokenPath:   transfer.TokenPath,
			Amount:      transfer.Amount,
		})
	}

	response := types.TransferHistoryResponse{
		Transfers: responseTransfers,
	}

	c.JSON(http.StatusOK, response)
}
