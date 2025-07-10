// Package circuitbreaker provides a circuit breaker middleware for Gin.
package circuitbreaker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
)

func TestGinCircuitBreakerHandler_Success(t *testing.T) {
	cfg := &Config{
		MaxRequests: 1,
		Interval:    1,
		Timeout:     1,
		ReadyToTrip: func(_ gobreaker.Counts) bool {
			return false
		},
	}
	mgr := NewManager(cfg)

	r := gin.New()
	r.Use(GinCircuitBreakerHandler(mgr))
	r.GET("/ok", func(c *gin.Context) {
		c.String(200, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

func TestGinCircuitBreakerHandler_Failure(t *testing.T) {
	cfg := &Config{
		MaxRequests: 1,
		Interval:    1,
		Timeout:     1,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 0
		},
	}
	mgr := NewManager(cfg)

	r := gin.New()
	r.Use(GinCircuitBreakerHandler(mgr))
	r.GET("/fail", func(c *gin.Context) {
		c.String(200, "should not run")
	})

	mgr.Break("/fail", 1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fail", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
