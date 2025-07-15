package metrics

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestFiberMiddleware(t *testing.T) {
	router := fiber.New()
	router.Use(FiberMiddleware())
	router.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	req, _ := http.NewRequest("GET", "/ping", nil)
	resp, err := router.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
