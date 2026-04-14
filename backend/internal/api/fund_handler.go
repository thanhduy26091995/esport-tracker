package api

import (
	"net/http"
	"strconv"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

type FundHandler struct {
	fundService *service.FundService
}

func NewFundHandler(fundService *service.FundService) *FundHandler {
	return &FundHandler{fundService: fundService}
}

// GetBalance returns the current fund balance
func (h *FundHandler) GetBalance(c *gin.Context) {
	balance, err := h.fundService.GetBalance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
	})
}

// GetStats returns fund statistics
func (h *FundHandler) GetStats(c *gin.Context) {
	stats, err := h.fundService.GetFundStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTransactions returns all fund transactions
func (h *FundHandler) GetTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transactions, err := h.fundService.GetAllTransactions(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// CreateDeposit creates a deposit transaction
func (h *FundHandler) CreateDeposit(c *gin.Context) {
	var req service.CreateDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": err.Error(),
		})
		return
	}

	transaction, err := h.fundService.CreateDeposit(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// CreateWithdrawal creates a withdrawal transaction
func (h *FundHandler) CreateWithdrawal(c *gin.Context) {
	var req service.CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": err.Error(),
		})
		return
	}

	transaction, err := h.fundService.CreateWithdrawal(&req)
	if err != nil {
		statusCode := http.StatusBadRequest
		code := "VALIDATION_ERROR"

		if err.Error() == "insufficient fund balance" {
			statusCode = http.StatusUnprocessableEntity
			code = "INSUFFICIENT_BALANCE"
		}

		c.JSON(statusCode, gin.H{
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}
