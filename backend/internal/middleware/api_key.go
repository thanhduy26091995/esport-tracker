package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// InternalKeyMiddleware guards non-WC routes with a shared secret header.
// If INTERNAL_API_KEY is unset, the middleware is a no-op (dev convenience).
func InternalKeyMiddleware() gin.HandlerFunc {
	key := os.Getenv("INTERNAL_API_KEY")
	return func(c *gin.Context) {
		if key != "" && c.GetHeader("X-Internal-Key") != key {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing API key"})
			return
		}
		c.Next()
	}
}
