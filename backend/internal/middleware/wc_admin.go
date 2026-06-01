package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WcAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, _ := c.Get(WcIsAdminKey)
		if isAdmin != true {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
