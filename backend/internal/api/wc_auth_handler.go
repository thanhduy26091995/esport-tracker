package api

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

type WcAuthHandler struct {
	authSvc *service.WcAuthService
}

func NewWcAuthHandler(authSvc *service.WcAuthService) *WcAuthHandler {
	return &WcAuthHandler{authSvc: authSvc}
}

type wcRegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=4"`
}

type wcLoginRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type wcResetPasswordRequest struct {
	Name string `json:"name" binding:"required"`
}

// Register handles POST /api/v1/wc/auth/register
func (h *WcAuthHandler) Register(c *gin.Context) {
	var req wcRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authSvc.Register(req.Name, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":    token,
		"user_id":  user.ID,
		"name":     user.Name,
		"is_admin": user.IsAdmin,
	})
}

// Login handles POST /api/v1/wc/auth/login
func (h *WcAuthHandler) Login(c *gin.Context) {
	var req wcLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authSvc.Login(req.Name, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"user_id":  user.ID,
		"name":     user.Name,
		"is_admin": user.IsAdmin,
	})
}

// ResetPassword handles POST /api/v1/wc/auth/reset-password
func (h *WcAuthHandler) ResetPassword(c *gin.Context) {
	var req wcResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authSvc.ResetPassword(req.Name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset to {name}_@123"})
}
