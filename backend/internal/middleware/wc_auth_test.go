package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func openMiddlewareTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping middleware tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Skipf("cannot connect to test DB (%v) — skipping", err)
	}
	require.NoError(t, db.AutoMigrate(&model.WcUser{}))
	return db
}

func seedMWUser(t *testing.T, db *gorm.DB, googleID *string, isAdmin bool) *model.WcUser {
	t.Helper()
	u := &model.WcUser{Name: "MWUser_" + uuid.NewString()[:8], GoogleID: googleID, IsAdmin: isAdmin}
	require.NoError(t, db.Create(u).Error)
	t.Cleanup(func() { db.Delete(u) })
	return u
}

// runGoogleLinkedMiddleware builds a gin engine that injects context keys (simulating
// WcJWTMiddleware) then runs WcGoogleLinkedMiddleware followed by a 200 handler.
func runGoogleLinkedMiddleware(db *gorm.DB, userID uuid.UUID, isAdmin bool) *httptest.ResponseRecorder {
	eng := gin.New()
	eng.Use(func(c *gin.Context) {
		c.Set(WcUserIDKey, userID)
		c.Set(WcIsAdminKey, isAdmin)
		c.Next()
	})
	eng.Use(WcGoogleLinkedMiddleware(db))
	eng.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	eng.ServeHTTP(w, req)
	return w
}

// ─── WcGoogleLinkedMiddleware ─────────────────────────────────────────────────

func TestWcGoogleLinkedMiddleware_LinkedUser_Passes(t *testing.T) {
	db := openMiddlewareTestDB(t)

	gid := "gsub_mw_" + uuid.NewString()[:8]
	user := seedMWUser(t, db, &gid, false)

	w := runGoogleLinkedMiddleware(db, user.ID, false)
	assert.Equal(t, http.StatusOK, w.Code, "linked user must pass through")
}

func TestWcGoogleLinkedMiddleware_UnlinkedUser_Returns403(t *testing.T) {
	db := openMiddlewareTestDB(t)

	user := seedMWUser(t, db, nil, false)

	w := runGoogleLinkedMiddleware(db, user.ID, false)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "google_not_linked")
}

func TestWcGoogleLinkedMiddleware_AdminUser_BypassesCheck(t *testing.T) {
	db := openMiddlewareTestDB(t)

	// Admin has no google_id — must still pass
	user := seedMWUser(t, db, nil, true)

	w := runGoogleLinkedMiddleware(db, user.ID, true)
	assert.Equal(t, http.StatusOK, w.Code, "admin must bypass google-link check")
}

func TestWcGoogleLinkedMiddleware_UnknownUserID_Returns403(t *testing.T) {
	db := openMiddlewareTestDB(t)

	// Random UUID not in DB — Scan returns zero-value with nil GoogleID
	w := runGoogleLinkedMiddleware(db, uuid.New(), false)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
