// Package circuitbreaker provides a circuit breaker middleware for Fiber.
package circuitbreaker

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/config/middleware/circuitbreakerconf"
	"github.com/stretchr/testify/assert"
)

func TestFiberMiddleware_Success(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 1,
		Interval:    1,
		Timeout:     1,
	}
	mgr := NewManager(cfg)

	r := fiber.New()
	r.Use(FiberMiddleware(mgr))
	r.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req, _ := http.NewRequest("GET", "/ok", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestFiberMiddleware_ConsecutiveFailures(t *testing.T) {
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

	r := fiber.New()
	r.Use(FiberMiddleware(mgr))
	r.Get("/fail/:id", func(c *fiber.Ctx) error {
		return c.SendString("should not run")
	})

	req, _ := http.NewRequest("GET", "/fail/1", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mgr.Break("/fail/1", 1)
	req, _ = http.NewRequest("GET", "/fail/1", nil)
	resp, err = r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestFiberMiddleware_FailureRate(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 1000,
		Interval:    1,
		Timeout:     1,
		Policy: &circuitbreakerconf.PolicyConfig{
			Type:                 "failure_rate",
			MinRequests:          2,
			FailureRateThreshold: 0.5,
		},
	}
	mgr := NewManager(cfg)

	r := fiber.New()
	r.Use(FiberMiddleware(mgr))
	r.Get("/fail", func(c *fiber.Ctx) error {
		return c.SendString("should not run")
	})

	mgr.Break("/fail", 2)
	req, _ := http.NewRequest("GET", "/fail", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
