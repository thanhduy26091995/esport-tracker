package api

import (
	"net/http"
	"strconv"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SettlementHandler struct {
	settlementService *service.SettlementService
}

func NewSettlementHandler(settlementService *service.SettlementService) *SettlementHandler {
	return &SettlementHandler{settlementService: settlementService}
}

// GetAll returns all settlements
func (h *SettlementHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	settlements, err := h.settlementService.GetAllSettlements(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, settlements)
}

// GetByID returns a settlement by ID
func (h *SettlementHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_UUID",
			"message": "Invalid settlement ID format",
		})
		return
	}

	settlement, err := h.settlementService.GetSettlementByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Settlement not found",
		})
		return
	}

	c.JSON(http.StatusOK, settlement)
}

// GetByDebtorID returns settlements for a specific debtor
func (h *SettlementHandler) GetByDebtorID(c *gin.Context) {
	debtorIDStr := c.Param("id")
	debtorID, err := uuid.Parse(debtorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_UUID",
			"message": "Invalid debtor ID format",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	settlements, err := h.settlementService.GetSettlementsByDebtorID(debtorID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, settlements)
}

// TriggerSettlement manually triggers settlement for a user
func (h *SettlementHandler) TriggerSettlement(c *gin.Context) {
	var req struct {
		DebtorID  uuid.UUID   `json:"debtor_id" binding:"required"`
		WinnerIDs []uuid.UUID `json:"winner_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": err.Error(),
		})
		return
	}

	settlement, err := h.settlementService.TriggerSettlement(req.DebtorID, req.WinnerIDs)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := "INTERNAL_ERROR"

		switch err.Error() {
		case "user does not have debt":
			statusCode = http.StatusBadRequest
			code = "NO_DEBT"
		case "no winners found for settlement":
			statusCode = http.StatusBadRequest
			code = "NO_WINNERS"
		case "debtor cannot be a winner":
			statusCode = http.StatusBadRequest
			code = "INVALID_WINNER"
		case "invalid winner ID":
			statusCode = http.StatusBadRequest
			code = "INVALID_WINNER"
		case "winners must have positive scores":
			statusCode = http.StatusBadRequest
			code = "INVALID_WINNER"
		}

		c.JSON(statusCode, gin.H{
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, settlement)
}

// GetStats returns settlement statistics
func (h *SettlementHandler) GetStats(c *gin.Context) {
	stats, err := h.settlementService.GetSettlementStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
