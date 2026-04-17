package api

import (
	"fmt"
	"net/http"

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

// GetAll handles GET /users
func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userService.GetAll()
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
