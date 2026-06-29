package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func runSiteAccessMW(repo fakeAccessRepo, path, token string) *httptest.ResponseRecorder {
	return runSiteAccessMWWithAuth(repo, path, token, "")
}

func runSiteAccessMWWithAuth(repo fakeAccessRepo, path, siteToken, authHeader string) *httptest.ResponseRecorder {
	eng := gin.New()
	eng.Use(siteAccessMiddlewareWithGetter(repo.Get))
	eng.GET("/*p", func(c *gin.Context) { c.Status(http.StatusOK) })
	eng.POST("/*p", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	if siteToken != "" {
		req.Header.Set("X-Site-Token", siteToken)
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	eng.ServeHTTP(w, req)
	return w
}

// fakeAccessRepo simulates Get() without a real DB.
type fakeAccessRepo struct {
	enabled    bool
	answerHash string
}

func (f fakeAccessRepo) Get() (enabled bool, answerHash string) {
	return f.enabled, f.answerHash
}

// ─── Exempt paths ─────────────────────────────────────────────────────────────

func TestSiteAccessMiddleware_ExemptPaths_PassWithoutToken(t *testing.T) {
	repo := fakeAccessRepo{enabled: true, answerHash: "somehash"}
	exemptPaths := []string{
		"/health",
		"/uploads/avatar.png",
		"/ws",
		"/ws/chat",
		"/api/v1/site-access/question",
		"/api/v1/site-access/validate",
		"/api/v1/wc/admin/site-access",
		"/api/v1/wc/admin/matches",
		"/api/v1/wc/auth/login",
	}
	for _, path := range exemptPaths {
		w := runSiteAccessMW(repo, path, "")
		assert.Equal(t, http.StatusOK, w.Code, "path %s should be exempt from gate", path)
	}
}

// ─── Gate disabled ────────────────────────────────────────────────────────────

func TestSiteAccessMiddleware_Disabled_PassesWithoutToken(t *testing.T) {
	repo := fakeAccessRepo{enabled: false, answerHash: "somehash"}
	w := runSiteAccessMW(repo, "/api/v1/users", "")
	assert.Equal(t, http.StatusOK, w.Code, "disabled gate must not block requests")
}

func TestSiteAccessMiddleware_Disabled_PassesWithWrongToken(t *testing.T) {
	repo := fakeAccessRepo{enabled: false, answerHash: "somehash"}
	w := runSiteAccessMW(repo, "/api/v1/wc/matches", "wrongtoken")
	assert.Equal(t, http.StatusOK, w.Code, "disabled gate ignores token value")
}

// ─── Gate enabled ─────────────────────────────────────────────────────────────

func TestSiteAccessMiddleware_Enabled_CorrectToken_Passes(t *testing.T) {
	const hash = "abc123"
	repo := fakeAccessRepo{enabled: true, answerHash: hash}
	w := runSiteAccessMW(repo, "/api/v1/wc/matches", hash)
	assert.Equal(t, http.StatusOK, w.Code, "correct token must pass through")
}

func TestSiteAccessMiddleware_Enabled_NoToken_Returns403(t *testing.T) {
	repo := fakeAccessRepo{enabled: true, answerHash: "abc123"}
	w := runSiteAccessMW(repo, "/api/v1/wc/matches", "")
	assert.Equal(t, http.StatusForbidden, w.Code, "missing token must be blocked")
}

func TestSiteAccessMiddleware_Enabled_WrongToken_Returns403(t *testing.T) {
	repo := fakeAccessRepo{enabled: true, answerHash: "abc123"}
	w := runSiteAccessMW(repo, "/api/v1/wc/matches", "wrongtoken")
	assert.Equal(t, http.StatusForbidden, w.Code, "wrong token must be blocked")
}

func TestSiteAccessMiddleware_Enabled_BearerToken_Bypasses(t *testing.T) {
	repo := fakeAccessRepo{enabled: true, answerHash: "abc123"}
	w := runSiteAccessMWWithAuth(repo, "/api/v1/wc/matches", "", "Bearer somejwttoken")
	assert.Equal(t, http.StatusOK, w.Code, "authenticated users bypass gate; JWT validity checked by WcJWTMiddleware")
}

func TestSiteAccessMiddleware_Enabled_BearerToken_NoSiteToken_Bypasses(t *testing.T) {
	repo := fakeAccessRepo{enabled: true, answerHash: "abc123"}
	w := runSiteAccessMWWithAuth(repo, "/api/v1/wc/leaderboard", "", "Bearer admintoken")
	assert.Equal(t, http.StatusOK, w.Code, "Bearer token bypasses even without X-Site-Token")
}

func TestSiteAccessMiddleware_Enabled_WcAdminRoute_IsExempt(t *testing.T) {
	repo := fakeAccessRepo{enabled: true, answerHash: "abc123"}
	w := runSiteAccessMW(repo, "/api/v1/wc/admin/site-access", "")
	assert.Equal(t, http.StatusOK, w.Code, "admin routes are exempt — protected by JWT + admin role")
}

func TestSiteAccessMiddleware_Enabled_WcAuthLogin_IsExempt(t *testing.T) {
	repo := fakeAccessRepo{enabled: true, answerHash: "abc123"}
	w := runSiteAccessMW(repo, "/api/v1/wc/auth/login", "")
	assert.Equal(t, http.StatusOK, w.Code, "login endpoint must be reachable before passing the gate")
}
