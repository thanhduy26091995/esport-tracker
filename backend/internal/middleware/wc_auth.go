package middleware

import (
	"net/http"
	"strings"

	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	WcUserIDKey  = "wc_user_id"
	WcUserNameKey = "wc_user_name"
	WcIsAdminKey = "wc_is_admin"
)

func WcJWTMiddleware(authSvc *service.WcAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		claims, err := authSvc.VerifyToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(WcUserIDKey, claims.WcUserID)
		c.Set(WcUserNameKey, claims.Name)
		c.Set(WcIsAdminKey, claims.IsAdmin)
		c.Next()
	}
}
