package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/stretchr/testify/assert"
)

func TestGinAuthMiddleware(t *testing.T) {
	conf := Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}

	t.Run("white list should pass", func(t *testing.T) {
		testWhiteList(t, conf)
	})

	t.Run("white list with prefix", func(t *testing.T) {
		testWhiteList(t, conf)
	})

	t.Run("missing token", func(t *testing.T) {
		testMissingToken(t, conf)
	})

	t.Run("invalid token", func(t *testing.T) {
		testInvalidToken(t, conf)
	})

	t.Run("valid token ok", func(t *testing.T) {
		testValidToken(t, conf)
	})
}

func testWhiteList(t *testing.T, conf Config) {
	r := gin.New()
	r.Use(GinAuthMiddleware(conf))
	r.GET("/public", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func testMissingToken(t *testing.T, conf Config) {
	r := gin.New()
	r.Use(GinAuthMiddleware(conf))
	r.GET("/secure", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing token")
}
func testInvalidToken(t *testing.T, conf Config) {
	r := gin.New()
	r.Use(GinAuthMiddleware(conf))
	r.GET("/secure", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}

func testValidToken(t *testing.T, conf Config) {
	r := gin.New()
	r.Use(GinAuthMiddleware(conf))
	r.GET("/secure", func(c *gin.Context) {
		ai, exists := contextkeys.GetAuthInfo(c.Request.Context())
		if !exists || ai.UserID != "123" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth info missing"})
			return
		}
		c.String(http.StatusOK, "hello "+ai.UserID)
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	tk, _ := GenerateJWTToken(map[string]any{"user_id": "123"}, conf.JWTSecret, time.Minute)
	req.Header.Set("Authorization", "Bearer "+tk)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello 123", w.Body.String())
}
