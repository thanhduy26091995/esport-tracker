package api

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScoreBonusHandler struct {
	bonusService *service.ScoreBonusService
}

func NewScoreBonusHandler(bonusService *service.ScoreBonusService) *ScoreBonusHandler {
	return &ScoreBonusHandler{bonusService: bonusService}
}

func (h *ScoreBonusHandler) Create(c *gin.Context) {
	var req service.CreateScoreBonusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	bonus, err := h.bonusService.CreateBonus(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		switch err.Error() {
		case "points must be positive":
			statusCode, code = http.StatusBadRequest, "INVALID_POINTS"
		default:
			if len(err.Error()) >= 9 && err.Error()[len(err.Error())-9:] == "not found" {
				statusCode, code = http.StatusNotFound, "USER_NOT_FOUND"
			}
		}
		c.JSON(statusCode, gin.H{"code": code, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bonus)
}

func (h *ScoreBonusHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_UUID", "message": "Invalid bonus ID format"})
		return
	}

	if err := h.bonusService.DeleteBonus(id); err != nil {
		statusCode := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		if err.Error() == "score bonus not found" {
			statusCode, code = http.StatusNotFound, "NOT_FOUND"
		}
		c.JSON(statusCode, gin.H{"code": code, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Score bonus deleted successfully"})
}
