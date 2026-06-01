package api

import (
	"net/http"
	"strconv"

	"github.com/duyb/esport-score-tracker/internal/middleware"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WcHandler struct {
	svc     *service.WcService
	authSvc *service.WcAuthService
}

func NewWcHandler(svc *service.WcService, authSvc *service.WcAuthService) *WcHandler {
	return &WcHandler{svc: svc, authSvc: authSvc}
}

// --- Config ---

// GetConfig handles GET /api/v1/wc/admin/config
func (h *WcHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// UpdateConfig handles PUT /api/v1/wc/admin/config
func (h *WcHandler) UpdateConfig(c *gin.Context) {
	var req struct {
		IsEnabled bool `json:"is_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	if err := h.svc.SetConfig(req.IsEnabled, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_enabled": req.IsEnabled})
}

// --- Matches ---

// ListMatches handles GET /api/v1/wc/matches
func (h *WcHandler) ListMatches(c *gin.Context) {
	f := repository.MatchFilter{
		Status: c.Query("status"),
		Stage:  c.Query("stage"),
		Group:  c.Query("group"),
		Date:   c.Query("date"),
	}
	matches, err := h.svc.ListMatches(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matches"})
		return
	}
	c.JSON(http.StatusOK, matches)
}

// GetMatch handles GET /api/v1/wc/matches/:id
func (h *WcHandler) GetMatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	m, err := h.svc.GetMatchWithOdds(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}
	c.JSON(http.StatusOK, m)
}

// GetMatchBets handles GET /api/v1/wc/matches/:id/bets
func (h *WcHandler) GetMatchBets(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	bets, err := h.svc.ListBetsForMatchPublic(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bets"})
		return
	}
	c.JSON(http.StatusOK, bets)
}

// GetScoreOdds handles GET /api/v1/wc/matches/:id/score-odds
func (h *WcHandler) GetScoreOdds(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	odds, err := h.svc.ListScoreOdds(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch score odds"})
		return
	}
	c.JSON(http.StatusOK, odds)
}

// GetLeaderboard handles GET /api/v1/wc/leaderboard
func (h *WcHandler) GetLeaderboard(c *gin.Context) {
	entries, err := h.svc.GetLeaderboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// --- Authenticated ---

// GetWallet handles GET /api/v1/wc/wallet
func (h *WcHandler) GetWallet(c *gin.Context) {
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	wallet, err := h.svc.GetWallet(wcUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch wallet"})
		return
	}
	c.JSON(http.StatusOK, wallet)
}

// PlaceBet handles POST /api/v1/wc/bets
func (h *WcHandler) PlaceBet(c *gin.Context) {
	var req struct {
		MatchID            string  `json:"match_id" binding:"required"`
		BetType            string  `json:"bet_type" binding:"required"`
		BetChoice          *string `json:"bet_choice"`
		PredictedHomeScore *int    `json:"predicted_home_score"`
		PredictedAwayScore *int    `json:"predicted_away_score"`
		Stake              int     `json:"stake" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	matchID, err := uuid.Parse(req.MatchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match_id"})
		return
	}
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)

	bet, err := h.svc.PlaceBet(wcUserID, service.PlaceBetRequest{
		MatchID:            matchID,
		BetType:            req.BetType,
		BetChoice:          req.BetChoice,
		PredictedHomeScore: req.PredictedHomeScore,
		PredictedAwayScore: req.PredictedAwayScore,
		Stake:              req.Stake,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, bet)
}

// ListBets handles GET /api/v1/wc/bets
func (h *WcHandler) ListBets(c *gin.Context) {
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	bets, err := h.svc.ListBets(wcUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bets"})
		return
	}
	c.JSON(http.StatusOK, bets)
}

// DeleteBet handles DELETE /api/v1/wc/bets/:id
func (h *WcHandler) DeleteBet(c *gin.Context) {
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	betID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bet ID"})
		return
	}
	if err := h.svc.DeleteBet(wcUserID, betID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// UpdateBet handles PUT /api/v1/wc/bets/:id
func (h *WcHandler) UpdateBet(c *gin.Context) {
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	betID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bet ID"})
		return
	}
	var body struct {
		Stake int `json:"stake"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateBetStake(wcUserID, betID, body.Stake); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// --- Admin ---

// SyncMatches handles POST /api/v1/wc/admin/sync
func (h *WcHandler) SyncMatches(c *gin.Context) {
	count, err := h.svc.SyncMatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"synced": count})
}

// UpdateMatch handles PUT /api/v1/wc/admin/matches/:id
func (h *WcHandler) UpdateMatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	var fields map[string]interface{}
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateMatch(id, fields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update match"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// LockMatch handles POST /api/v1/wc/admin/matches/:id/lock
func (h *WcHandler) LockMatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	if err := h.svc.LockMatch(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lock match"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// AddScoreOdds handles POST /api/v1/wc/admin/matches/:id/score-odds
func (h *WcHandler) AddScoreOdds(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	var req struct {
		HomeScore int     `json:"home_score" binding:"min=0"`
		AwayScore int     `json:"away_score" binding:"min=0"`
		Odds      float64 `json:"odds" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	so, err := h.svc.AddScoreOdds(id, req.HomeScore, req.AwayScore, req.Odds)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, so)
}

// UpdateScoreOdds handles PUT /api/v1/wc/admin/score-odds/:id
func (h *WcHandler) UpdateScoreOdds(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score odds ID"})
		return
	}
	var req struct {
		Odds float64 `json:"odds" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateScoreOdds(id, req.Odds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update score odds"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DeleteScoreOdds handles DELETE /api/v1/wc/admin/score-odds/:id
func (h *WcHandler) DeleteScoreOdds(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score odds ID"})
		return
	}
	if err := h.svc.DeleteScoreOdds(id); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// SettleMatch handles POST /api/v1/wc/admin/matches/:id/settle
func (h *WcHandler) SettleMatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	processed, totalPayout, err := h.svc.SettleMatch(id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bets_processed": processed, "total_paid_out": totalPayout})
}

// AdminTopUp handles PUT /api/v1/wc/admin/wallets/:wc_user_id
func (h *WcHandler) AdminTopUp(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("wc_user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	var req struct {
		Delta int    `json:"delta" binding:"required"`
		Note  string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	if err := h.svc.AdminTopUp(adminID, targetID, req.Delta, req.Note); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetWalletLogs handles GET /api/v1/wc/admin/wallets/:wc_user_id/logs
func (h *WcHandler) GetWalletLogs(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("wc_user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	logs, err := h.svc.GetWalletLogs(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch wallet logs"})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// SetUserRole handles PUT /api/v1/wc/admin/users/:wc_user_id/role
func (h *WcHandler) SetUserRole(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("wc_user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	var req struct {
		IsAdmin bool `json:"is_admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetAdminRole(targetID, req.IsAdmin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ListUsers handles GET /api/v1/wc/admin/users
func (h *WcHandler) ListUsers(c *gin.Context) {
	users, err := h.svc.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// ListAllWallets handles GET /api/v1/wc/admin/wallets
func (h *WcHandler) ListAllWallets(c *gin.Context) {
	wallets, err := h.svc.GetAllWallets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch wallets"})
		return
	}
	c.JSON(http.StatusOK, wallets)
}

// PreviewSettlement handles GET /api/v1/wc/admin/settlements/preview
func (h *WcHandler) PreviewSettlement(c *gin.Context) {
	rateStr := c.Query("point_rate")
	pointRate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil || pointRate <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "point_rate query param required (positive number)"})
		return
	}
	rows, err := h.svc.PreviewSettlement(pointRate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to preview settlement"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// CreateSettlement handles POST /api/v1/wc/admin/settlements
func (h *WcHandler) CreateSettlement(c *gin.Context) {
	var req struct {
		Name      string  `json:"name" binding:"required"`
		PointRate float64 `json:"point_rate" binding:"required,gt=0"`
		Note      string  `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	s, err := h.svc.CreateSettlement(adminID, req.Name, req.PointRate, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

// ListSettlements handles GET /api/v1/wc/admin/settlements
func (h *WcHandler) ListSettlements(c *gin.Context) {
	list, err := h.svc.ListSettlements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch settlements"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetSettlement handles GET /api/v1/wc/admin/settlements/:id
func (h *WcHandler) GetSettlement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settlement ID"})
		return
	}
	s, err := h.svc.GetSettlement(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "settlement not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// MarkSettlementDone handles PUT /api/v1/wc/admin/settlements/:id/details/:wc_user_id
func (h *WcHandler) MarkSettlementDone(c *gin.Context) {
	settlementID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settlement ID"})
		return
	}
	wcUserID, err := uuid.Parse(c.Param("wc_user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	var req struct {
		Status   string `json:"status" binding:"required"`
		DoneNote string `json:"done_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.MarkSettlementDone(settlementID, wcUserID, req.DoneNote); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settlement detail"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
