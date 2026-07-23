package api

import (
	"net/http"
	"strings"

	"github.com/duyb/esport-score-tracker/internal/middleware"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WcChampionHandler struct {
	svc *service.WcChampionService
}

func NewWcChampionHandler(svc *service.WcChampionService) *WcChampionHandler {
	return &WcChampionHandler{svc: svc}
}

// GetConfig handles GET /api/v1/{wc|ac}/champion/config
func (h *WcChampionHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.GetPublicConfig(tournamentType(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load champion config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// GetTeams handles GET /api/v1/{wc|ac}/champion/teams
func (h *WcChampionHandler) GetTeams(c *gin.Context) {
	teams, err := h.svc.ListTeams(tournamentType(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list teams"})
		return
	}
	c.JSON(http.StatusOK, teams)
}

// GetAllPredictions handles GET /api/v1/{wc|ac}/champion/predictions
func (h *WcChampionHandler) GetAllPredictions(c *gin.Context) {
	preds, err := h.svc.GetAllPredictions(tournamentType(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list predictions"})
		return
	}
	c.JSON(http.StatusOK, preds)
}

// GetMyPrediction handles GET /api/v1/{wc|ac}/champion/my-prediction
// Returns all predictions for the current user (array, may be empty).
func (h *WcChampionHandler) GetMyPrediction(c *gin.Context) {
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	preds, err := h.svc.GetMyPredictions(tournamentType(c), wcUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load predictions"})
		return
	}
	c.JSON(http.StatusOK, preds)
}

// PlacePredict handles POST /api/v1/{wc|ac}/champion/predict
func (h *WcChampionHandler) PlacePredict(c *gin.Context) {
	var req struct {
		TeamID uuid.UUID `json:"team_id" binding:"required"`
		Points int       `json:"points" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	pred, err := h.svc.PlaceOrUpdatePrediction(tournamentType(c), wcUserID, req.TeamID, req.Points)
	if err != nil {
		if strings.Contains(err.Error(), "blocked") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pred)
}

// DeletePredict handles DELETE /api/v1/{wc|ac}/champion/predict/:id
func (h *WcChampionHandler) DeletePredict(c *gin.Context) {
	predID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prediction id"})
		return
	}
	wcUserID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	if err := h.svc.DeletePredictionByID(tournamentType(c), wcUserID, predID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// AdminUpdateConfig handles PUT /api/v1/{wc|ac}/admin/champion/config
func (h *WcChampionHandler) AdminUpdateConfig(c *gin.Context) {
	var req struct {
		IsOpen bool `json:"is_open"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateConfig(tournamentType(c), req.IsOpen); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_open": req.IsOpen})
}

// AdminCreateTeam handles POST /api/v1/{wc|ac}/admin/champion/teams
func (h *WcChampionHandler) AdminCreateTeam(c *gin.Context) {
	var req struct {
		Name      string  `json:"name" binding:"required"`
		Code      string  `json:"code" binding:"required"`
		FlagEmoji string  `json:"flag_emoji"`
		Odds      float64 `json:"odds" binding:"required,gt=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := h.svc.CreateTeam(tournamentType(c), req.Name, req.Code, req.FlagEmoji, req.Odds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, team)
}

// AdminUpdateTeamOdds handles PUT /api/v1/{wc|ac}/admin/champion/teams/:id
func (h *WcChampionHandler) AdminUpdateTeamOdds(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}
	var req struct {
		Odds float64 `json:"odds" binding:"required,gt=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateTeamOdds(teamID, req.Odds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// AdminSettle handles POST /api/v1/{wc|ac}/admin/champion/settle
func (h *WcChampionHandler) AdminSettle(c *gin.Context) {
	var req struct {
		WinnerTeamID uuid.UUID `json:"winner_team_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	result, err := h.svc.SettleChampion(tournamentType(c), adminID, req.WinnerTeamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
