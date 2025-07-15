package recovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestFiberMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(FiberMiddleware())

	app.Get("/panic", func(_ *fiber.Ctx) error {
		panic("simulated panic for test")
	})

	req := httptest.NewRequest("GET", "/panic", nil)

	resp, err := app.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
