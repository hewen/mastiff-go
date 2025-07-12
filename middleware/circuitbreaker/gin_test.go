// Package circuitbreaker provides a circuit breaker middleware for Gin.
package circuitbreaker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/middleware/circuitbreakerconf"
	"github.com/stretchr/testify/assert"
)

func TestGinMiddleware_Success(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 1,
		Interval:    1,
		Timeout:     1,
	}
	mgr := NewManager(cfg)

	r := gin.New()
	r.Use(GinMiddleware(mgr))
	r.GET("/ok", func(c *gin.Context) {
		c.String(200, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

func TestGinMiddleware_ConsecutiveFailures(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 1,
		Interval:    1,
		Timeout:     1,
		Policy: &circuitbreakerconf.PolicyConfig{
			Type:                "consecutive_failures",
			ConsecutiveFailures: 1,
		},
	}
	mgr := NewManager(cfg)

	r := gin.New()
	r.Use(GinMiddleware(mgr))
	r.GET("/fail", func(c *gin.Context) {
		c.String(200, "should not run")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fail", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	mgr.Break("/fail", 1)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/fail", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestGinMiddleware_FailureRate(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 1,
		Interval:    1,
		Timeout:     1,
		Policy: &circuitbreakerconf.PolicyConfig{
			Type:                 "failure_rate",
			MinRequests:          1,
			FailureRateThreshold: 0.5,
		},
	}
	mgr := NewManager(cfg)

	r := gin.New()
	r.Use(GinMiddleware(mgr))
	r.GET("/fail", func(c *gin.Context) {
		c.String(500, "should not run")
	})

	mgr.Break("/fail", 2)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fail", nil)
	r.ServeHTTP(w, req)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
