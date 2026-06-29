package middleware

import (
	"net/http"
	"strings"

	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/gin-gonic/gin"
)

var siteAccessExemptPrefixes = []string{
	"/health",
	"/uploads/",
	"/ws",
	"/api/v1/site-access/",
	"/api/v1/wc/admin/",  // admin routes are protected by JWT + admin role
	"/api/v1/wc/auth/",   // login endpoint must be reachable before gate
}

// SiteAccessMiddleware gates all non-exempt routes behind X-Site-Token when enabled.
func SiteAccessMiddleware(repo *repository.SiteAccessRepository) gin.HandlerFunc {
	return siteAccessMiddlewareWithGetter(func() (bool, string) {
		cfg, err := repo.Get()
		if err != nil {
			return false, ""
		}
		return cfg.Enabled, cfg.AnswerHash
	})
}

// siteAccessMiddlewareWithGetter is the testable core — accepts a getter closure
// so tests can inject a fake without a real DB.
func siteAccessMiddlewareWithGetter(get func() (bool, string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, prefix := range siteAccessExemptPrefixes {
			if path == prefix || strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		enabled, answerHash := get()
		if !enabled {
			c.Next()
			return
		}

		// Authenticated users (any valid-looking Bearer token) bypass the gate.
		// JWT validity is still enforced by WcJWTMiddleware on protected routes.
		if strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
			c.Next()
			return
		}

		if c.GetHeader("X-Site-Token") != answerHash {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
