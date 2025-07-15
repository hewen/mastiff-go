// Package ratelimit provides a rate limiter middleware for Fiber.
package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/config/middleware/ratelimitconf"
	"github.com/stretchr/testify/assert"
)

func TestFiberMiddleware(t *testing.T) {
	r := fiber.New()
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

	r.Use(FiberMiddleware(mgr))
	r.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
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

func TestFiberMiddleware_NoConfig(t *testing.T) {
	r := fiber.New()

	cfg := &ratelimitconf.Config{}
	mgr := NewLimiterManager(cfg)
	defer mgr.Stop()

	r.Use(FiberMiddleware(mgr))
	r.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
