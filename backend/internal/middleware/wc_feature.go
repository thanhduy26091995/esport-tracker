package middleware

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/gin-gonic/gin"
)

func WcFeatureMiddleware(wcRepo *repository.WcRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg, err := wcRepo.GetConfig()
		if err != nil || !cfg.IsEnabled {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "WC2026 feature is currently disabled",
			})
			return
		}
		c.Next()
	}
}
