package api

import (
	"errors"
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/middleware"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WcProfileHandler struct {
	profileSvc *service.WcProfileService
}

func NewWcProfileHandler(profileSvc *service.WcProfileService) *WcProfileHandler {
	return &WcProfileHandler{profileSvc: profileSvc}
}

type wcUpdateProfileRequest struct {
	Name      *string `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}

// GetProfile handles GET /api/v1/wc/profile
func (h *WcProfileHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	user, err := h.profileSvc.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"name":         user.Name,
		"avatar_url":   user.AvatarURL,
		"is_admin":     user.IsAdmin,
		"google_linked": user.GoogleLinked(),
		"created_at":   user.CreatedAt,
	})
}

// UpdateProfile handles PUT /api/v1/wc/profile
func (h *WcProfileHandler) UpdateProfile(c *gin.Context) {
	var req wcUpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
	user, err := h.profileSvc.UpdateProfile(userID, req.Name, req.AvatarURL)
	if err != nil {
		if errors.Is(err, service.ErrNameTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Name is already taken by another player"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"name":         user.Name,
		"avatar_url":   user.AvatarURL,
		"is_admin":     user.IsAdmin,
		"google_linked": user.GoogleLinked(),
		"created_at":   user.CreatedAt,
	})
}
