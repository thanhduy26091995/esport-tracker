package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

type WcChatHandler struct {
	svc *service.WcChatService
}

func NewWcChatHandler(svc *service.WcChatService) *WcChatHandler {
	return &WcChatHandler{svc: svc}
}

// ListMessages handles GET /wc/chat/messages?limit=20&before=<RFC3339>
// Returns messages ordered oldest → newest. No auth required.
func (h *WcChatHandler) ListMessages(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	var before *time.Time
	if b := c.Query("before"); b != "" {
		if t, err := time.Parse(time.RFC3339, b); err == nil {
			before = &t
		}
	}

	msgs, err := h.svc.ListHistory(limit, before)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load chat history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": msgs, "has_more": len(msgs) == limit})
}
