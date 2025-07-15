package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/config/middleware/authconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/stretchr/testify/assert"
)

func TestFiberWhiteList(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}

	r := fiber.New()
	r.Use(FiberMiddleware(conf))
	r.Get("/public", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/public", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestFiberMissingToken(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}
	r := fiber.New()
	r.Use(FiberMiddleware(conf))
	r.Get("/public", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
func TestFiberInvalidToken(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}
	r := fiber.New()
	r.Use(FiberMiddleware(conf))
	r.Get("/secure", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "invalid-token")
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestFiberValidToken(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}
	r := fiber.New()
	r.Use(FiberMiddleware(conf))
	r.Get("/secure", func(c *fiber.Ctx) error {
		ctx := contextkeys.ContextFrom(c)
		ai, exists := contextkeys.GetAuthInfo(ctx)
		if !exists || ai.UserID != "123" {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "auth info missing"})
		}
		return c.SendString("hello " + ai.UserID)
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	tk, _ := GenerateJWTToken(map[string]any{"user_id": "123"}, conf.JWTSecret, time.Minute)
	req.Header.Set("Authorization", "Bearer "+tk)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
