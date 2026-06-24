package api

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

type WcChatHandler struct {
	svc *service.WcChatService
}

func NewWcChatHandler(svc *service.WcChatService) *WcChatHandler {
	return &WcChatHandler{svc: svc}
}

// ListMessages handles GET /wc/chat/messages — returns last 100 messages, no auth required.
func (h *WcChatHandler) ListMessages(c *gin.Context) {
	msgs, err := h.svc.ListHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load chat history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": msgs})
}
