package api

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MatchHandler struct {
	matchService *service.MatchService
	bonusService *service.ScoreBonusService
}

func NewMatchHandler(matchService *service.MatchService, bonusService *service.ScoreBonusService) *MatchHandler {
	return &MatchHandler{matchService: matchService, bonusService: bonusService}
}

type matchFeedItem struct {
	Type string `json:"type"`
	*model.Match
}

type bonusFeedItem struct {
	Type string `json:"type"`
	*model.ScoreBonus
}

type feedEntry struct {
	date time.Time
	item any
}

func (h *MatchHandler) buildFeed(matches []*model.Match, bonuses []*model.ScoreBonus) []any {
	entries := make([]feedEntry, 0, len(matches)+len(bonuses))
	for _, m := range matches {
		entries = append(entries, feedEntry{date: m.MatchDate, item: matchFeedItem{Type: "match", Match: m}})
	}
	for _, b := range bonuses {
		entries = append(entries, feedEntry{date: b.BonusDate, item: bonusFeedItem{Type: "bonus", ScoreBonus: b}})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].date.After(entries[j].date)
	})
	result := make([]any, len(entries))
	for i, e := range entries {
		result[i] = e.item
	}
	return result
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
		case "match_type must be '1v1', '2v2', or '1v2'":
			statusCode = http.StatusBadRequest
			code = "INVALID_MATCH_TYPE"
		case "winner_team must be 0, 1, or 2":
			statusCode = http.StatusBadRequest
			code = "INVALID_WINNER_TEAM"
		case "duplicate player found in match":
			statusCode = http.StatusBadRequest
			code = "DUPLICATE_PLAYER"
		default:
			if len(err.Error()) >= 3 && err.Error()[:3] == "for" { // Team size error: "for 1v2, team1 needs..."
				statusCode = http.StatusBadRequest
				code = "INVALID_TEAM_SIZE"
			} else if len(err.Error()) >= 9 && err.Error()[len(err.Error())-9:] == "not found" {
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

// GetAll returns all matches and score bonuses merged by date
func (h *MatchHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0")) // 0 = no limit; pagination happens after merge
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var playerID *uuid.UUID
	if playerIDStr := c.Query("player_id"); playerIDStr != "" {
		parsed, err := uuid.Parse(playerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_UUID", "message": "invalid player_id"})
			return
		}
		playerID = &parsed
	}

	matches, err := h.matchService.GetAllMatches(0, 0, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	bonuses, err := h.bonusService.GetAll(0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	feed := h.buildFeed(matches, bonuses)

	// Apply pagination on merged result
	total := len(feed)
	start := offset
	if start > total {
		start = total
	}
	end := total
	if limit > 0 && start+limit < total {
		end = start + limit
	}

	c.JSON(http.StatusOK, feed[start:end])
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

// GetRecent returns recent matches and bonuses merged by date
func (h *MatchHandler) GetRecent(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	matches, err := h.matchService.GetRecentMatches(0) // fetch all recent; trim after merge
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	bonuses, err := h.bonusService.GetAll(0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	feed := h.buildFeed(matches, bonuses)
	if limit > 0 && len(feed) > limit {
		feed = feed[:limit]
	}

	c.JSON(http.StatusOK, feed)
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
