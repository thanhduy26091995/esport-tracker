package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name         string  `json:"name" binding:"required"`
	Tier         string  `json:"tier"`
	HandicapRate float64 `json:"handicap_rate"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name         string  `json:"name"`
	Tier         string  `json:"tier"`
	HandicapRate float64 `json:"handicap_rate"`
}

// GetAll handles GET /users. With ?include_inactive=true it also returns soft-deleted
// players (for the head-to-head picker); by default only active players are returned.
func (h *UserHandler) GetAll(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "true"

	var users []*model.UserWithStats
	var err error
	if includeInactive {
		users, err = h.userService.GetAllIncludingInactive()
	} else {
		users, err = h.userService.GetAll()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch users",
			},
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetHeadToHead handles GET /users/head-to-head?player1=&player2=
// Returns the opponents-only head-to-head record between two players (all match types).
func (h *UserHandler) GetHeadToHead(c *gin.Context) {
	p1, err := uuid.Parse(c.Query("player1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid or missing player1 ID"},
		})
		return
	}
	p2, err := uuid.Parse(c.Query("player2"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid or missing player2 ID"},
		})
		return
	}

	result, err := h.userService.GetHeadToHead(p1, p2)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSamePlayer):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
			})
		case errors.Is(err, service.ErrPlayerNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": err.Error()},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to compute head-to-head"},
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID handles GET /users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Create handles POST /users
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	user, err := h.userService.CreateUser(req.Name, req.Tier, req.HandicapRate)
	if err != nil {
		// Check if it's a duplicate name error
		if err.Error() == "user with name '"+req.Name+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "CONFLICT",
					"message": err.Error(),
				},
			})
			return
		}

		// Check if it's a validation error
		if err.Error() == "name cannot be empty" ||
			err.Error() == "name must be at least 2 characters" ||
			err.Error() == "name cannot exceed 100 characters" ||
			err.Error() == "tier must be one of: pro, normal, noop" ||
			err.Error() == "handicap_rate must be between 0 and 5" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to create user",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Update handles PUT /users/:id
func (h *UserHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	user, err := h.userService.UpdateUser(id, req.Name, req.Tier, req.HandicapRate)
	if err != nil {
		// Check if user not found
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "NOT_FOUND",
					"message": err.Error(),
				},
			})
			return
		}

		// Check if it's a duplicate name error
		if err.Error() == "user with name '"+req.Name+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "CONFLICT",
					"message": err.Error(),
				},
			})
			return
		}

		// Check if it's a validation error
		if err.Error() == "name cannot be empty" ||
			err.Error() == "name must be at least 2 characters" ||
			err.Error() == "name cannot exceed 100 characters" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to update user",
			},
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete handles DELETE /users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	if err := h.userService.DeleteUser(id); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "NOT_FOUND",
					"message": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to delete user",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// GetPaymentRanking handles GET /users/payment-ranking
func (h *UserHandler) GetPaymentRanking(c *gin.Context) {
	users, err := h.userService.GetPaymentRanking()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch payment ranking",
			},
		})
		return
	}
	c.JSON(http.StatusOK, users)
}

// UploadAvatar handles PUT /users/:id/avatar
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid user ID"}})
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Missing avatar file"}})
		return
	}
	defer file.Close()

	avatarURL, err := h.userService.UploadAvatar(id, file, header)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		} else if strings.HasPrefix(err.Error(), "file") || strings.HasPrefix(err.Error(), "unsupported") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": gin.H{"code": "UPLOAD_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"avatar_url": avatarURL})
}

// SetAvatarURL handles PUT /users/:id/avatar/url
func (h *UserHandler) SetAvatarURL(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid user ID"}})
		return
	}
	var req struct {
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.AvatarURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "avatar_url is required"}})
		return
	}
	if err := h.userService.SetAvatarURL(id, req.AvatarURL); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": gin.H{"code": "UPDATE_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"avatar_url": req.AvatarURL})
}

// DeleteAvatar handles DELETE /users/:id/avatar
func (h *UserHandler) DeleteAvatar(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid user ID"}})
		return
	}
	if err := h.userService.DeleteAvatar(id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": gin.H{"code": "DELETE_ERROR", "message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateClub handles PUT /users/:id/club
func (h *UserHandler) UpdateClub(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid user ID"}})
		return
	}
	var req struct {
		FavoriteClub string `json:"favorite_club"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid request body"}})
		return
	}
	if err := h.userService.UpdateClub(id, req.FavoriteClub); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		} else if strings.HasPrefix(err.Error(), "unknown") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": gin.H{"code": "UPDATE_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"favorite_club": req.FavoriteClub})
}

// GetLeaderboard handles GET /users/leaderboard
func (h *UserHandler) GetLeaderboard(c *gin.Context) {
	limit := 0 // Default: no limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if _, err := fmt.Sscanf(limitParam, "%d", &limit); err != nil {
			limit = 0
		}
	}

	users, err := h.userService.GetLeaderboard(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch leaderboard",
			},
		})
		return
	}

	c.JSON(http.StatusOK, users)
}
