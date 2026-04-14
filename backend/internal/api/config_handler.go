package api

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService *service.ConfigService
}

func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
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
