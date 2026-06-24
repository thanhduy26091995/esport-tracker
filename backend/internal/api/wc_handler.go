package api

import (
	"net/http"
	"strconv"
	"strings"

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

// GetPublicConfig handles GET /api/v1/wc/config — public, returns is_enabled + bet limits
func (h *WcHandler) GetPublicConfig(c *gin.Context) {
	cfg, err := h.svc.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"is_enabled": cfg.IsEnabled,
		"min_points": cfg.MinPoints,
		"max_points": cfg.MaxPoints,
	})
}

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
		IsEnabled *bool `json:"is_enabled"`
		MinPoints *int  `json:"min_points"`
		MaxPoints *int  `json:"max_points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)

	if req.IsEnabled != nil {
		if err := h.svc.SetConfig(*req.IsEnabled, adminID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
			return
		}
	}

	if req.MinPoints != nil || req.MaxPoints != nil {
		cfg, err := h.svc.GetConfig()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
			return
		}
		min, max := cfg.MinPoints, cfg.MaxPoints
		if req.MinPoints != nil {
			min = *req.MinPoints
		}
		if req.MaxPoints != nil {
			max = *req.MaxPoints
		}
		if err := h.svc.SetBetLimits(min, max, adminID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	cfg, _ := h.svc.GetConfig()
	c.JSON(http.StatusOK, cfg)
}

// ListMatches handles GET /api/v1/wc/matches
func (h *WcHandler) ListMatches(c *gin.Context) {
	f := repository.MatchFilter{
		Status:   c.Query("status"),
		Stage:    c.Query("stage"),
		Group:    c.Query("group"),
		Date:     c.Query("date"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
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

// GetMatchPredictions handles GET /api/v1/wc/matches/:id/predictions
func (h *WcHandler) GetMatchPredictions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	predictions, err := h.svc.ListPredictionsForMatchPublic(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch predictions"})
		return
	}
	c.JSON(http.StatusOK, predictions)
}

// GetScoreMultipliers handles GET /api/v1/wc/matches/:id/score-multipliers
func (h *WcHandler) GetScoreMultipliers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	multipliers, err := h.svc.ListScoreMultipliers(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch score multipliers"})
		return
	}
	c.JSON(http.StatusOK, multipliers)
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

// SubmitPrediction handles POST /api/v1/wc/predictions
func (h *WcHandler) SubmitPrediction(c *gin.Context) {
	var req struct {
		MatchID            string  `json:"match_id" binding:"required"`
		PredictionType     string  `json:"prediction_type" binding:"required"`
		PredictionChoice   *string `json:"prediction_choice"`
		PredictedHomeScore *int    `json:"predicted_home_score"`
		PredictedAwayScore *int    `json:"predicted_away_score"`
		Points             int     `json:"points" binding:"required,min=1"`
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
	prediction, err := h.svc.SubmitPrediction(wcUserID, service.SubmitPredictionRequest{
		MatchID:            matchID,
		PredictionType:     req.PredictionType,
		PredictionChoice:   req.PredictionChoice,
		PredictedHomeScore: req.PredictedHomeScore,
		PredictedAwayScore: req.PredictedAwayScore,
		Points:             req.Points,
	})
	if err != nil {
		if strings.Contains(err.Error(), "blocked") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, prediction)
}

// ListPredictions handles GET /api/v1/wc/predictions
func (h *WcHandler) ListPredictions(c *gin.Context) {
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	predictions, err := h.svc.ListPredictions(wcUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch predictions"})
		return
	}
	c.JSON(http.StatusOK, predictions)
}

// DeletePrediction handles DELETE /api/v1/wc/predictions/:id
func (h *WcHandler) DeletePrediction(c *gin.Context) {
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	predictionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prediction ID"})
		return
	}
	if err := h.svc.DeletePrediction(wcUserID, predictionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// UpdatePrediction handles PUT /api/v1/wc/predictions/:id
func (h *WcHandler) UpdatePrediction(c *gin.Context) {
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	predictionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prediction ID"})
		return
	}
	var body struct {
		Points int `json:"points"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdatePredictionPoints(wcUserID, predictionID, body.Points); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
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

// PlaceBet handles POST /api/v1/wc/bets
func (h *WcHandler) PlaceBet(c *gin.Context) {
	var req struct {
		MatchID            string  `json:"match_id" binding:"required"`
		BetType            string  `json:"bet_type" binding:"required"`
		BetChoice          *string `json:"bet_choice"`
		Stake              int     `json:"stake" binding:"required,min=1"`
		PredictedHomeScore *int    `json:"predicted_home_score"`
		PredictedAwayScore *int    `json:"predicted_away_score"`
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
		Stake:              req.Stake,
		PredictedHomeScore: req.PredictedHomeScore,
		PredictedAwayScore: req.PredictedAwayScore,
	})
	if err != nil {
		if strings.Contains(err.Error(), "blocked") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, bet)
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
	c.Status(http.StatusNoContent)
}

// GetMatchBets handles GET /api/v1/wc/matches/:id/bets
func (h *WcHandler) GetMatchBets(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	bets, err := h.svc.ListBetsForMatch(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bets"})
		return
	}
	c.JSON(http.StatusOK, bets)
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
	c.JSON(http.StatusOK, gin.H{"bets_processed": processed, "total_payout": totalPayout})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete score odds"})
		return
	}
	c.Status(http.StatusNoContent)
}

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

// OpenMatch handles POST /api/v1/wc/admin/matches/:id/open
func (h *WcHandler) OpenMatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	if err := h.svc.OpenMatch(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open match"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// CloseMatch handles POST /api/v1/wc/admin/matches/:id/close
func (h *WcHandler) CloseMatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	if err := h.svc.CloseMatch(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close match"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// AddScoreMultiplier handles POST /api/v1/wc/admin/matches/:id/score-multipliers
func (h *WcHandler) AddScoreMultiplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	var req struct {
		HomeScore  int     `json:"home_score" binding:"min=0"`
		AwayScore  int     `json:"away_score" binding:"min=0"`
		Multiplier float64 `json:"multiplier" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sm, err := h.svc.AddScoreMultiplier(id, req.HomeScore, req.AwayScore, req.Multiplier)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sm)
}

// UpdateScoreMultiplier handles PUT /api/v1/wc/admin/score-multipliers/:id
func (h *WcHandler) UpdateScoreMultiplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score multiplier ID"})
		return
	}
	var req struct {
		Multiplier float64 `json:"multiplier" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateScoreMultiplier(id, req.Multiplier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update score multiplier"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DeleteScoreMultiplier handles DELETE /api/v1/wc/admin/score-multipliers/:id
func (h *WcHandler) DeleteScoreMultiplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score multiplier ID"})
		return
	}
	if err := h.svc.DeleteScoreMultiplier(id); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// FinalizeAll handles POST /api/v1/wc/admin/matches/finalize-all
func (h *WcHandler) FinalizeAll(c *gin.Context) {
	result, err := h.svc.FinalizeAllMatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// RefinalizeAll handles POST /api/v1/wc/admin/matches/refinalize-all
// Re-calculates points_earned for all scored matches, correcting any rounded/integer values.
func (h *WcHandler) RefinalizeAll(c *gin.Context) {
	result, err := h.svc.RefinalizeAllMatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// FinalizeMatch handles POST /api/v1/wc/admin/matches/:id/finalize
func (h *WcHandler) FinalizeMatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	res, err := h.svc.FinalizeMatch(id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// PreviewFinalizeMatch handles GET /api/v1/wc/admin/matches/:id/finalize-preview
func (h *WcHandler) PreviewFinalizeMatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	result, err := h.svc.PreviewFinalizeMatch(id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// PreviewFinalizeAll handles GET /api/v1/wc/admin/matches/finalize-all-preview
func (h *WcHandler) PreviewFinalizeAll(c *gin.Context) {
	result, err := h.svc.PreviewFinalizeAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build preview"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// PreviewRefinalizeAll handles GET /api/v1/wc/admin/matches/refinalize-all-preview
func (h *WcHandler) PreviewRefinalizeAll(c *gin.Context) {
	result, err := h.svc.PreviewRefinalizeAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build preview"})
		return
	}
	c.JSON(http.StatusOK, result)
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

// ListUsersForMention handles GET /api/v1/wc/users — returns minimal user info for @mention autocomplete.
// Requires JWT auth; excludes blocked users.
func (h *WcHandler) ListUsersForMention(c *gin.Context) {
	users, err := h.svc.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	type userItem struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		AvatarURL *string `json:"avatar_url"`
	}
	items := make([]userItem, 0, len(users))
	for _, u := range users {
		if u.IsBlocked {
			continue
		}
		items = append(items, userItem{ID: u.ID.String(), Name: u.Name, AvatarURL: u.AvatarURL})
	}
	c.JSON(http.StatusOK, gin.H{"users": items})
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

// GetHousePnL handles GET /api/v1/wc/admin/house-pnl
func (h *WcHandler) GetHousePnL(c *gin.Context) {
	pnl, err := h.svc.GetHousePnL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute P&L"})
		return
	}
	c.JSON(http.StatusOK, pnl)
}

// BlockUser handles PUT /api/v1/wc/admin/users/:wc_user_id/block
func (h *WcHandler) BlockUser(c *gin.Context) {
	adminID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	targetID, err := uuid.Parse(c.Param("wc_user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	voidedCount, err := h.svc.BlockUser(adminID, targetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "voided_bets": voidedCount})
}

// UnblockUser handles PUT /api/v1/wc/admin/users/:wc_user_id/unblock
func (h *WcHandler) UnblockUser(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("wc_user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	if err := h.svc.UnblockUser(targetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unblock user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
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

// GetGroupStandings handles GET /api/v1/wc/standings
func (h *WcHandler) GetGroupStandings(c *gin.Context) {
	resp, err := h.svc.GetGroupStandings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch standings"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
