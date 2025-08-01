// Package ratelimit provides a rate limiter middleware for Httpx.
package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hewen/mastiff-go/config/middlewareconf/ratelimitconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/server/httpx"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"github.com/stretchr/testify/assert"
)

func TestHttpxMiddleware(t *testing.T) {
	r, err := httpx.NewHTTPServer(&serverconf.HTTPConfig{
		FrameworkType: serverconf.FrameworkFiber,
	})
	assert.Nil(t, err)

	cfg := &ratelimitconf.Config{
		Default: &ratelimitconf.RouteLimitConfig{
			Rate:        1,
			Burst:       1,
			Mode:        ratelimitconf.ModeAllow,
			EnableRoute: true,
			EnableIP:    true,
		},
	}
	mgr := NewLimiterManager(cfg)
	defer mgr.Stop()

	r.Use(HttpxMiddleware(mgr))
	r.Get("/test", func(c unicontext.UniversalContext) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func TestHttpxMiddleware_NoConfig(t *testing.T) {
	r, err := httpx.NewHTTPServer(&serverconf.HTTPConfig{
		FrameworkType: serverconf.FrameworkFiber,
	})
	assert.Nil(t, err)

	cfg := &ratelimitconf.Config{}
	mgr := NewLimiterManager(cfg)
	defer mgr.Stop()

	r.Use(HttpxMiddleware(mgr))
	r.Get("/test", func(c unicontext.UniversalContext) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
