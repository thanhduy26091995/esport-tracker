package middleware

import (
	"net/http"

	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/gin-gonic/gin"
)

func WcFeatureMiddleware(wcRepo *repository.WcRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentType, _ := c.Get(TournamentTypeKey)
		tt, _ := tournamentType.(string)
		cfg, err := wcRepo.GetConfig(tt)
		if err != nil || !cfg.IsEnabled {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Feature is currently disabled",
			})
			return
		}
		c.Next()
	}
}
