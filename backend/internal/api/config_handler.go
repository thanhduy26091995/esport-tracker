package api

import (
	"log"
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService *service.ConfigService
	tierService   *service.TierService
}

func NewConfigHandler(configService *service.ConfigService, tierService *service.TierService) *ConfigHandler {
	return &ConfigHandler{configService: configService, tierService: tierService}
}

// GetAll returns all configuration entries
func (h *ConfigHandler) GetAll(c *gin.Context) {
	configs, err := h.configService.GetAllConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// GetByKey returns a specific config value
func (h *ConfigHandler) GetByKey(c *gin.Context) {
	key := c.Param("key")

	config, err := h.configService.GetConfigByKey(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Config not found",
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateAll bulk-updates multiple config values in one request
func (h *ConfigHandler) UpdateAll(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": err.Error(),
		})
		return
	}

	configs, err := h.configService.UpdateAllConfig(req)
	if err != nil {
		statusCode := http.StatusBadRequest
		code := "VALIDATION_ERROR"
		if err.Error() == "invalid config key" {
			statusCode = http.StatusNotFound
			code = "NOT_FOUND"
		}
		c.JSON(statusCode, gin.H{"code": code, "message": err.Error()})
		return
	}

	if _, ok := req["min_matches_for_tier"]; ok {
		if err := h.tierService.RecalculateAllTiers(); err != nil {
			log.Printf("config: failed to recalculate tiers after min_matches_for_tier change: %v", err)
		}
	}

	c.JSON(http.StatusOK, configs)
}

// Update updates a config value
func (h *ConfigHandler) Update(c *gin.Context) {
	key := c.Param("key")

	var req struct {
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": err.Error(),
		})
		return
	}

	err := h.configService.UpdateConfig(key, req.Value)
	if err != nil {
		statusCode := http.StatusBadRequest
		code := "VALIDATION_ERROR"

		if err.Error() == "invalid config key" {
			statusCode = http.StatusNotFound
			code = "NOT_FOUND"
		}

		c.JSON(statusCode, gin.H{
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	if key == "min_matches_for_tier" {
		if err := h.tierService.RecalculateAllTiers(); err != nil {
			log.Printf("config: failed to recalculate tiers after min_matches_for_tier change: %v", err)
		}
	}

	// Return updated config
	config, err := h.configService.GetConfigByKey(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, config)
}
