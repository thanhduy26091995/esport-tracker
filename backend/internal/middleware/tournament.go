package middleware

import "github.com/gin-gonic/gin"

const TournamentTypeKey = "tournament_type"

func TournamentMiddleware(tournamentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(TournamentTypeKey, tournamentType)
		c.Next()
	}
}
