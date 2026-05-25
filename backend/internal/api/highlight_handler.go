package api

import (
	"net/http"
	"time"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

type HighlightHandler struct {
	highlightService *service.HighlightService
}

func NewHighlightHandler(highlightService *service.HighlightService) *HighlightHandler {
	return &HighlightHandler{highlightService: highlightService}
}

// GetHighlights computes and returns all player highlights grouped by section.
func (h *HighlightHandler) GetHighlights(c *gin.Context) {
	resp, err := h.highlightService.GenerateHighlights()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	resp.GeneratedAt = time.Now()
	c.JSON(http.StatusOK, resp)
}
