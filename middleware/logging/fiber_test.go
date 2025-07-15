package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestFiberMiddleware(t *testing.T) {
	router := fiber.New()
	router.Use(FiberMiddleware())

	router.Get("/log", func(ctx *fiber.Ctx) error {
		ctx.Locals("req", "req")
		ctx.Locals("resp", "resp")
		return nil
	})

	req := httptest.NewRequest("GET", "/log", nil)
	resp, err := router.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
