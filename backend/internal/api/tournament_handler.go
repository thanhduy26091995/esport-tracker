package api

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TournamentHandler struct {
	tournamentService *service.TournamentService
}

func NewTournamentHandler(tournamentService *service.TournamentService) *TournamentHandler {
	return &TournamentHandler{tournamentService: tournamentService}
}

func (h *TournamentHandler) GetAll(c *gin.Context) {
	tournaments, err := h.tournamentService.GetAllTournaments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tournaments)
}

func (h *TournamentHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid tournament ID"})
		return
	}
	t, err := h.tournamentService.GetTournament(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Tournament not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TournamentHandler) Create(c *gin.Context) {
	var req service.CreateTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}
	t, err := h.tournamentService.CreateTournament(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "CREATE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *TournamentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid tournament ID"})
		return
	}
	if err := h.tournamentService.DeleteTournament(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "DELETE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tournament deleted"})
}

func (h *TournamentHandler) Complete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid tournament ID"})
		return
	}
	t, err := h.tournamentService.CompleteTournament(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "COMPLETE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TournamentHandler) RecordResult(c *gin.Context) {
	tournamentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid tournament ID"})
		return
	}
	matchID, err := uuid.Parse(c.Param("matchId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid match ID"})
		return
	}

	var req service.RecordMatchResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	tm, err := h.tournamentService.RecordMatchResult(tournamentID, matchID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "RECORD_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tm)
}
