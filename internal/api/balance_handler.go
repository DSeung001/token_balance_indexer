package api

import (
	"github.com/gin-gonic/gin"
	"gn-indexer/internal/repository"
	"gn-indexer/internal/types"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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
		// Get all balances for all addresses
		balances, err := h.balanceRepo.GetAllBalances(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get all balances: " + err.Error(),
			})
			return
		}

		responseBalances := make([]types.TokenBalance, 0, len(balances))
		for _, balance := range balances {
			amount := int64(0)
			if balance.Amount != nil {
				amount = balance.Amount.Int64()
			}
			responseBalances = append(responseBalances, types.TokenBalance{
				TokenPath: balance.TokenPath,
				Amount:    amount,
			})
		}

		response := types.BalanceResponse{
			Balances: responseBalances,
		}

		c.JSON(http.StatusOK, response)
		return
	}

	// Get balances for specific address
	balances, err := h.balanceRepo.GetBalancesByAddress(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get balances: " + err.Error(),
		})
		return
	}

	responseBalances := make([]types.TokenBalance, 0, len(balances))
	for _, balance := range balances {
		amount := int64(0)
		if balance.Amount != nil {
			amount = balance.Amount.Int64()
		}
		responseBalances = append(responseBalances, types.TokenBalance{
			TokenPath: balance.TokenPath,
			Amount:    amount,
		})
	}

	response := types.BalanceResponse{
		Balances: responseBalances,
	}

	c.JSON(http.StatusOK, response)
}

// precompile once at package level (optional but nice)
var reTokenBalances = regexp.MustCompile(`^/tokens/(.+)/balances$`)

// GetBalancesByTokenAndAddress handles GET /tokens/{tokenPath}/balances?address={address}
func (h *BalanceHandler) GetBalancesByTokenAndAddress(c *gin.Context) {
	raw := c.Param("tokenPath")

	// Example raw path: /tokens/gno.land/r/demo/wugnot/balances
	if raw == "" {
		path := c.Request.URL.Path
		// Normalize backslashes if any (defensive)
		path = strings.ReplaceAll(path, `\`, `/`)

		if m := reTokenBalances.FindStringSubmatch(path); len(m) == 2 {
			raw = m[1]
		}
	}

	// URL-decode in case client sent %2F etc.
	tokenPath, err := url.PathUnescape(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tokenPath: " + err.Error()})
		return
	}
	tokenPath = strings.Trim(tokenPath, "/")
	if tokenPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tokenPath parameter is required"})
		return
	}

	// (Optional) sanity check for allowed characters
	if !regexp.MustCompile(`^[A-Za-z0-9._\-/]+$`).MatchString(tokenPath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tokenPath format"})
		return
	}

	// Read optional address query
	address := c.Query("address")

	// Case A: no address → all balances for the token
	if address == "" {
		balances, err := h.balanceRepo.GetBalancesByTokenAndAddress(c.Request.Context(), tokenPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get balances: " + err.Error()})
			return
		}

		accountBalances := make([]types.AccountBalance, 0, len(balances))
		for _, balance := range balances {
			amount := int64(0)
			if balance.Amount != nil {
				amount = balance.Amount.Int64()
			}
			accountBalances = append(accountBalances, types.AccountBalance{
				Address:   balance.Address,
				TokenPath: balance.TokenPath, // or tokenPath to normalize
				Amount:    amount,
			})
		}

		response := types.AccountBalanceResponse{
			AccountBalances: accountBalances,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Case B: address provided → single balance for (tokenPath, address)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get balance: " + err.Error()})
		return
	}

	amount := int64(0)
	if balance.Amount != nil {
		amount = balance.Amount.Int64()
	}

	response := types.AccountBalanceResponse{
		AccountBalances: []types.AccountBalance{{
			Address:   balance.Address,
			TokenPath: balance.TokenPath, // or tokenPath to normalize
			Amount:    amount,
		}},
	}
	c.JSON(http.StatusOK, response)
}

// GetTransferHistory handles GET /tokens/transfer-history?address={address}
func (h *BalanceHandler) GetTransferHistory(c *gin.Context) {
	address := c.Query("address")

	if address == "" {
		// Get all transfer history
		transfers, err := h.transferRepo.GetAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get all transfer history: " + err.Error(),
			})
			return
		}

		responseTransfers := make([]types.TransferRecord, 0, len(transfers))
		for _, transfer := range transfers {
			amount := int64(0)
			if transfer.Amount != nil {
				amount = transfer.Amount.Int64()
			}
			responseTransfers = append(responseTransfers, types.TransferRecord{
				FromAddress: transfer.FromAddress,
				ToAddress:   transfer.ToAddress,
				TokenPath:   transfer.TokenPath,
				Amount:      amount,
			})
		}

		response := types.TransferHistoryResponse{
			Transfers: responseTransfers,
		}

		c.JSON(http.StatusOK, response)
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

	// Transfer history for specific address
	transfers, err := h.transferRepo.GetByAddress(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get transfer history: " + err.Error(),
		})
		return
	}

	responseTransfers := make([]types.TransferRecord, 0, len(transfers))
	for _, transfer := range transfers {
		amount := int64(0)
		if transfer.Amount != nil {
			amount = transfer.Amount.Int64()
		}
		responseTransfers = append(responseTransfers, types.TransferRecord{
			FromAddress: transfer.FromAddress,
			ToAddress:   transfer.ToAddress,
			TokenPath:   transfer.TokenPath,
			Amount:      amount,
		})
	}

	response := types.TransferHistoryResponse{
		Transfers: responseTransfers,
	}

	c.JSON(http.StatusOK, response)
}
