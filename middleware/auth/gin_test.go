package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/middleware/authconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/stretchr/testify/assert"
)

func TestGinWhiteList(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}

	r := gin.New()
	r.Use(GinMiddleware(conf))
	r.GET("/public", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestGinMissingToken(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}

	r := gin.New()
	r.Use(GinMiddleware(conf))
	r.GET("/secure", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing token")
}

func TestGinInvalidToken(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}

	r := gin.New()
	r.Use(GinMiddleware(conf))
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

func TestGinValidToken(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}

	r := gin.New()
	r.Use(GinMiddleware(conf))
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
