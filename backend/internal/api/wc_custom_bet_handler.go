package api

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/middleware"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WcCustomBetHandler struct {
	svc *service.WcCustomBetService
}

func NewWcCustomBetHandler(svc *service.WcCustomBetService) *WcCustomBetHandler {
	return &WcCustomBetHandler{svc: svc}
}

// AdminCreateCustomBet handles POST /admin/matches/:id/custom-bets
func (h *WcCustomBetHandler) AdminCreateCustomBet(c *gin.Context) {
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}
	var req struct {
		Title   string                          `json:"title" binding:"required"`
		Line    *float64                        `json:"line"`
		Options []service.CreateCustomBetOption `json:"options" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	result, err := h.svc.CreateCustomBet(matchID, adminID, req.Title, req.Line, req.Options)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// AdminUpdateCustomBet handles PUT /admin/custom-bets/:id
func (h *WcCustomBetHandler) AdminUpdateCustomBet(c *gin.Context) {
	betID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bet id"})
		return
	}
	var req struct {
		Title  *string  `json:"title"`
		Line   *float64 `json:"line"`
		Status *string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.UpdateCustomBet(betID, req.Title, req.Line, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// AdminSettleCustomBet handles POST /admin/custom-bets/:id/settle
func (h *WcCustomBetHandler) AdminSettleCustomBet(c *gin.Context) {
	betID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bet id"})
		return
	}
	var req struct {
		WinningOptionID uuid.UUID `json:"winning_option_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	if err := h.svc.Settle(betID, req.WinningOptionID, adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "settled"})
}

// AdminVoidCustomBet handles PUT /admin/custom-bets/:id/void
func (h *WcCustomBetHandler) AdminVoidCustomBet(c *gin.Context) {
	betID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bet id"})
		return
	}
	if err := h.svc.VoidBet(betID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "voided"})
}

// AdminListCustomBets handles GET /admin/matches/:id/custom-bets
func (h *WcCustomBetHandler) AdminListCustomBets(c *gin.Context) {
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}
	bets, err := h.svc.ListForMatchAdmin(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bets)
}

// ListCustomBets handles GET /matches/:id/custom-bets (player)
func (h *WcCustomBetHandler) ListCustomBets(c *gin.Context) {
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}
	userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	bets, err := h.svc.ListForMatchPlayer(matchID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, bet := range bets {
		bet.Entries = nil
	}
	c.JSON(http.StatusOK, bets)
}

// PlaceEntry handles POST /custom-bets/:id/entry (player)
func (h *WcCustomBetHandler) PlaceEntry(c *gin.Context) {
	betID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bet id"})
		return
	}
	var req struct {
		OptionID uuid.UUID `json:"option_id" binding:"required"`
		Stake    int       `json:"stake" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	if err := h.svc.PlaceEntry(betID, userID, req.OptionID, req.Stake); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "entry placed"})
}

// GetMyCustomBetEntries handles GET /wc/custom-bet-entries (player)
func (h *WcCustomBetHandler) GetMyCustomBetEntries(c *gin.Context) {
	userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	entries, err := h.svc.GetMyEntries(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// CancelEntry handles DELETE /custom-bet-entries/:id (player)
func (h *WcCustomBetHandler) CancelEntry(c *gin.Context) {
	entryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry id"})
		return
	}
	userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	if err := h.svc.CancelEntry(entryID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "entry cancelled"})
}
