package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs method, path, status, and duration for every request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Milliseconds()
		status := c.Writer.Status()

		level := "INFO"
		if status >= 500 {
			level = "ERROR"
		} else if status >= 400 {
			level = "WARN"
		}

		log.Printf("[%s] %s %s → %d (%dms) ip=%s",
			level,
			c.Request.Method,
			c.Request.URL.Path,
			status,
			duration,
			c.ClientIP(),
		)

		// Log slow requests (>500ms) separately for easy grepping
		if duration > 500 {
			log.Printf("[SLOW] %s %s took %dms",
				c.Request.Method, c.Request.URL.Path, duration)
		}

		// Log error details if present
		if len(c.Errors) > 0 {
			log.Printf("[ERROR_DETAIL] %s %s errors: %s",
				c.Request.Method, c.Request.URL.Path, fmt.Sprintf("%v", c.Errors))
		}
	}
}
