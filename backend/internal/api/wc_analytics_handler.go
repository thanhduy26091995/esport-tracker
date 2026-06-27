package api

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/middleware"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WcAnalyticsHandler struct {
	svc *service.WcAnalyticsService
}

func NewWcAnalyticsHandler(svc *service.WcAnalyticsService) *WcAnalyticsHandler {
	return &WcAnalyticsHandler{svc: svc}
}

// GetMyAnalytics handles GET /api/v1/wc/analytics/my
func (h *WcAnalyticsHandler) GetMyAnalytics(c *gin.Context) {
	userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)

	param := c.DefaultQuery("period", "30d")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	period := service.PeriodFromParam(param, dateFrom, dateTo)

	resp, err := h.svc.BuildMyResponse(userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load analytics"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetCommunityAnalytics handles GET /api/v1/wc/analytics/community
func (h *WcAnalyticsHandler) GetCommunityAnalytics(c *gin.Context) {
	resp, err := h.svc.BuildCommunityResponse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load community analytics"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetCompareAnalytics handles GET /api/v1/wc/analytics/compare
func (h *WcAnalyticsHandler) GetCompareAnalytics(c *gin.Context) {
	userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)

	resp, err := h.svc.BuildCompareResponse(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load compare analytics"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetWorldCup2026Analytics handles GET /api/v1/wc/analytics/world-cup-2026
func (h *WcAnalyticsHandler) GetWorldCup2026Analytics(c *gin.Context) {
	resp, err := h.svc.GetWorldCup2026Analytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load tournament analytics"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
