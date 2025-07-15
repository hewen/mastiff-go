package httpx

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware"
	"github.com/hewen/mastiff-go/middleware/logging"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestFiberHandlerBuilder(t *testing.T) {
	initRoute := func(app *fiber.App) {
		app.Get("/test", func(c *fiber.Ctx) error {
			ctx := contextkeys.ContextFrom(c)
			l := logger.NewLoggerWithContext(ctx)
			l.Infof("test")
			return c.JSON(map[string]string{
				"message": "ok",
			})
		})

	}
	assert.NotNil(t, initRoute)

	port, err := util.GetFreePort()
	assert.Nil(t, err)

	enableMetrics := true
	conf := &serverconf.HTTPConfig{
		Addr:         fmt.Sprintf("localhost:%d", port),
		PprofEnabled: true,
		Mode:         "debug",
		Middlewares: middleware.Config{
			EnableMetrics: &enableMetrics,
		},
	}

	builder := &FiberHandlerBuilder{
		Conf:             conf,
		InitRoute:        initRoute,
		ExtraMiddlewares: []func(*fiber.Ctx) error{logging.FiberMiddleware()},
	}

	s, err := NewHTTPServer(builder)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	_ = s.Name()

	go func() {
		s.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
	assert.Nil(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), `"message":"ok"`)

	s.Stop()
}
