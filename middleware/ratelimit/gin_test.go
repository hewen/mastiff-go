// Package ratelimit provides a rate limiter middleware for Gin.
package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGinRateLimitHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &Config{
		Default: &RouteLimitConfig{
			Rate:  1,
			Burst: 1,
			Mode:  ModeAllow,
			Strategy: Strategy{
				EnableRoute: true,
				EnableIP:    true,
			},
		},
	}
	mgr := NewLimiterManager(cfg)
	defer mgr.Stop()

	r.Use(GinRateLimitHandler(mgr))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
