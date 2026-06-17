package api

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WcSyncHandler struct {
	syncSvc   *service.StatsApiSyncService
	poissonSvc *service.PoissonService
}

func NewWcSyncHandler(syncSvc *service.StatsApiSyncService, poissonSvc *service.PoissonService) *WcSyncHandler {
	return &WcSyncHandler{syncSvc: syncSvc, poissonSvc: poissonSvc}
}

// SetupMapping handles POST /admin/setup-statsapi-mapping
func (h *WcSyncHandler) SetupMapping(c *gin.Context) {
	var req struct {
		PreviewOnly bool `json:"preview_only"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.PreviewOnly = true // default to preview
	}
	adminID := adminIDFromCtx(c)
	result, err := h.syncSvc.SetupMapping(req.PreviewOnly, adminID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if req.PreviewOnly {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, gin.H{"ok": true, "mapped": len(result.Matched)})
	}
}

// ImportHandicap handles POST /admin/matches/:id/import-handicap
func (h *WcSyncHandler) ImportHandicap(c *gin.Context) {
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	var req struct {
		PreviewOnly bool `json:"preview_only"`
	}
	_ = c.ShouldBindJSON(&req)

	if req.PreviewOnly {
		preview, err := h.syncSvc.PreviewHandicap(matchID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, preview)
		return
	}
	adminID := adminIDFromCtx(c)
	if err := h.syncSvc.ImportHandicapForMatch(matchID, adminID); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ImportOU handles POST /admin/matches/:id/import-ou
func (h *WcSyncHandler) ImportOU(c *gin.Context) {
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	var req struct {
		PreviewOnly bool `json:"preview_only"`
	}
	_ = c.ShouldBindJSON(&req)

	if req.PreviewOnly {
		preview, err := h.syncSvc.PreviewOU(matchID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, preview)
		return
	}
	adminID := adminIDFromCtx(c)
	if err := h.syncSvc.ImportOUForMatch(matchID, adminID); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GeneratePoisson handles POST /admin/matches/:id/generate-poisson
func (h *WcSyncHandler) GeneratePoisson(c *gin.Context) {
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}
	var req struct {
		HomeLambda  float64 `json:"home_lambda" binding:"required,gt=0"`
		AwayLambda  float64 `json:"away_lambda" binding:"required,gt=0"`
		HouseMargin float64 `json:"house_margin"`
		MinProb     float64 `json:"min_prob"`
		PreviewOnly bool    `json:"preview_only"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.HouseMargin == 0 {
		req.HouseMargin = 0.10
	}
	if req.MinProb == 0 {
		req.MinProb = 0.01
	}

	input := service.PoissonInput{
		MatchID:     matchID,
		HomeLambda:  req.HomeLambda,
		AwayLambda:  req.AwayLambda,
		HouseMargin: req.HouseMargin,
		MinProb:     req.MinProb,
	}
	lines, oddsRows := h.poissonSvc.GenerateScoreOdds(input)
	if len(lines) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "no scorelines generated — check lambda values"})
		return
	}

	if req.PreviewOnly {
		c.JSON(http.StatusOK, gin.H{
			"match_id":     matchID,
			"score_odds":   lines,
			"count":        len(lines),
			"house_margin": req.HouseMargin,
		})
		return
	}

	// Save to wc_score_odds
	dbOdds := make([]model.WcScoreOdds, len(oddsRows))
	copy(dbOdds, oddsRows)
	// Use the existing AddScoreOdds upsert via the repo (through a pointer we pass in)
	// We'll bulk-upsert by calling the handler's repo indirectly via a service call not available here.
	// Instead, expose BulkUpsert via a thin wrapper in WcSyncHandler.
	if err := h.syncSvc.UpsertScoreOdds(dbOdds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save score odds"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "count": len(oddsRows)})
}

// GetSyncLogs handles GET /admin/sync-logs
func (h *WcSyncHandler) GetSyncLogs(c *gin.Context) {
	logs, err := h.syncSvc.GetSyncLogs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sync logs"})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func adminIDFromCtx(c *gin.Context) uuid.UUID {
	if v, ok := c.Get("wc_user_id"); ok {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}
