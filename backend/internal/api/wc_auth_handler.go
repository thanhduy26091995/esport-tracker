package api

import (
	"errors"
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/middleware"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WcAuthHandler struct {
	authSvc *service.WcAuthService
}

func NewWcAuthHandler(authSvc *service.WcAuthService) *WcAuthHandler {
	return &WcAuthHandler{authSvc: authSvc}
}

type wcLoginRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type wcGoogleRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// Login handles POST /api/v1/wc/auth/login
func (h *WcAuthHandler) Login(c *gin.Context) {
	var req wcLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.Login(req.Name, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GoogleLoginOrCreate handles POST /api/v1/wc/auth/google
func (h *WcAuthHandler) GoogleLoginOrCreate(c *gin.Context) {
	var req wcGoogleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.GoogleLoginOrCreate(c.Request.Context(), req.IDToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidGoogleToken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired Google token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process Google login"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GoogleLink handles POST /api/v1/wc/auth/google/link (requires JWT, no google-link check)
func (h *WcAuthHandler) GoogleLink(c *gin.Context) {
	var req wcGoogleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	avatarURL, err := h.authSvc.LinkGoogleToAccount(c.Request.Context(), userID, req.IDToken)
	_ = avatarURL
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidGoogleToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired Google token"})
		case errors.Is(err, service.ErrGoogleAlreadyLinked):
			c.JSON(http.StatusConflict, gin.H{"error": "This Google account is already linked to another player"})
		case errors.Is(err, service.ErrAlreadyLinked):
			c.JSON(http.StatusConflict, gin.H{"error": "Your account is already linked to a Google account"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link Google account"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"google_linked": true, "avatar_url": avatarURL})
}
