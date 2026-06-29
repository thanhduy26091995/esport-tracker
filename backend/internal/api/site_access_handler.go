package api

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

type SiteAccessHandler struct {
	svc *service.SiteAccessService
}

func NewSiteAccessHandler(svc *service.SiteAccessService) *SiteAccessHandler {
	return &SiteAccessHandler{svc: svc}
}

// GetQuestion handles GET /api/v1/site-access/question — public, no auth required.
func (h *SiteAccessHandler) GetQuestion(c *gin.Context) {
	question, enabled, err := h.svc.GetQuestion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"question": question, "enabled": enabled})
}

// Validate handles POST /api/v1/site-access/validate — public, no auth required.
func (h *SiteAccessHandler) Validate(c *gin.Context) {
	var req struct {
		Answer string `json:"answer" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "answer is required"})
		return
	}
	token, err := h.svc.Validate(req.Answer)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect answer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetAdminConfig handles GET /api/v1/wc/admin/site-access.
func (h *SiteAccessHandler) GetAdminConfig(c *gin.Context) {
	question, enabled, err := h.svc.GetQuestion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"question": question, "enabled": enabled})
}

// UpdateAdminConfig handles PUT /api/v1/wc/admin/site-access.
func (h *SiteAccessHandler) UpdateAdminConfig(c *gin.Context) {
	var req struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
		Enabled  bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.svc.Update(req.Question, req.Answer, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
