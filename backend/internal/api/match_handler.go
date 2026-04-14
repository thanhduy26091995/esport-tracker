package api

import (
	"net/http"
	"strconv"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MatchHandler struct {
	matchService *service.MatchService
}

func NewMatchHandler(matchService *service.MatchService) *MatchHandler {
	return &MatchHandler{matchService: matchService}
}

// Create creates a new match
func (h *MatchHandler) Create(c *gin.Context) {
	var req service.CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": err.Error(),
		})
		return
	}

	match, err := h.matchService.CreateMatch(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := "INTERNAL_ERROR"

		// Handle specific errors
		switch err.Error() {
		case "match_type must be '1v1' or '2v2'":
			statusCode = http.StatusBadRequest
			code = "INVALID_MATCH_TYPE"
		case "winner_team must be 1 or 2":
			statusCode = http.StatusBadRequest
			code = "INVALID_WINNER_TEAM"
		case "duplicate player found in match":
			statusCode = http.StatusBadRequest
			code = "DUPLICATE_PLAYER"
		default:
			if err.Error()[:4] == "each" { // Team size error
				statusCode = http.StatusBadRequest
				code = "INVALID_TEAM_SIZE"
			} else if err.Error()[len(err.Error())-9:] == "not found" { // User not found
				statusCode = http.StatusNotFound
				code = "USER_NOT_FOUND"
			}
		}

		c.JSON(statusCode, gin.H{
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, match)
}

// GetAll returns all matches with pagination
func (h *MatchHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	matches, err := h.matchService.GetAllMatches(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, matches)
}

// GetByID returns a match by ID
func (h *MatchHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_UUID",
			"message": "Invalid match ID format",
		})
		return
	}

	match, err := h.matchService.GetMatchByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Match not found",
		})
		return
	}

	c.JSON(http.StatusOK, match)
}

// GetRecent returns recent matches
func (h *MatchHandler) GetRecent(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	matches, err := h.matchService.GetRecentMatches(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, matches)
}

// GetByUserID returns matches for a specific user
func (h *MatchHandler) GetByUserID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_UUID",
			"message": "Invalid user ID format",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	matches, err := h.matchService.GetMatchesByUserID(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, matches)
}

// Delete deletes a match and reverts scores
func (h *MatchHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_UUID",
			"message": "Invalid match ID format",
		})
		return
	}

	err = h.matchService.DeleteMatch(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := "INTERNAL_ERROR"

		if err.Error() == "cannot delete a locked match" {
			statusCode = http.StatusForbidden
			code = "MATCH_LOCKED"
		}

		c.JSON(statusCode, gin.H{
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Match deleted successfully",
	})
}

// GetStats returns match statistics
func (h *MatchHandler) GetStats(c *gin.Context) {
	stats, err := h.matchService.GetMatchStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
