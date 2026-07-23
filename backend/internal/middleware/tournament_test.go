package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTournamentMiddleware_SetsWorldCup(t *testing.T) {
	eng := gin.New()
	eng.Use(TournamentMiddleware("world_cup"))
	var got string
	eng.GET("/", func(c *gin.Context) {
		v, _ := c.Get(TournamentTypeKey)
		got, _ = v.(string)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	eng.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "world_cup", got)
}

func TestTournamentMiddleware_SetsAseanCup(t *testing.T) {
	eng := gin.New()
	eng.Use(TournamentMiddleware("asean_cup"))
	var got string
	eng.GET("/", func(c *gin.Context) {
		v, _ := c.Get(TournamentTypeKey)
		got, _ = v.(string)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	eng.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "asean_cup", got)
}

func TestTournamentMiddleware_CallsNext(t *testing.T) {
	called := false
	eng := gin.New()
	eng.Use(TournamentMiddleware("world_cup"))
	eng.GET("/", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	eng.ServeHTTP(w, req)

	assert.True(t, called, "downstream handler must be called after TournamentMiddleware")
}

func TestTournamentMiddleware_KeyConstant(t *testing.T) {
	assert.Equal(t, "tournament_type", TournamentTypeKey)
}
