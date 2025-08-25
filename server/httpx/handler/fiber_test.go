package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/middlewareconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware/logging"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"github.com/stretchr/testify/assert"
)

// nolint
func TestFiberHandler(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	enableMetrics := true
	conf := &serverconf.HTTPConfig{
		Addr:         fmt.Sprintf("localhost:%d", port),
		PprofEnabled: true,
		Mode:         "debug",
		Middlewares: middlewareconf.Config{
			EnableMetrics: &enableMetrics,
		},
		FrameworkType: serverconf.FrameworkFiber,
	}

	s, err := NewHandler(conf)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	_ = s.Name()
	s.Use(logging.HttpxMiddleware())

	s.Get("/test", func(c unicontext.UniversalContext) error {
		ctx := unicontext.ContextFrom(c)
		l := logger.NewLoggerWithContext(ctx)
		l.Infof("test")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "ok",
		})
	})

	go func() {
		_ = s.Start()
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

	err = s.Stop()
	assert.Nil(t, err)
}

func TestGetFiberConfig(t *testing.T) {
	t.Run("release mode", func(t *testing.T) {
		conf := getFiberConfig("release")
		assert.Equal(t, true, conf.Prefork)
		assert.Equal(t, true, conf.CaseSensitive)
		assert.Equal(t, true, conf.StrictRouting)
		assert.Equal(t, false, conf.EnablePrintRoutes)
		assert.Equal(t, true, conf.DisableStartupMessage)
	})

	t.Run("debug mode", func(t *testing.T) {
		conf := getFiberConfig("debug")
		assert.Equal(t, false, conf.Prefork)
		assert.Equal(t, true, conf.CaseSensitive)
		assert.Equal(t, false, conf.StrictRouting)
		assert.Equal(t, false, conf.EnablePrintRoutes)
		assert.Equal(t, true, conf.DisableStartupMessage)
	})

	t.Run("other mode", func(t *testing.T) {
		conf := getFiberConfig("test")
		assert.Equal(t, false, conf.Prefork)
		assert.Equal(t, true, conf.CaseSensitive)
		assert.Equal(t, false, conf.StrictRouting)
		assert.Equal(t, false, conf.EnablePrintRoutes)
		assert.Equal(t, true, conf.DisableStartupMessage)
	})
}

func TestFiberHandler_Test(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkFiber,
	}

	handler, err := NewHandler(conf)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	testBasicHandlerFunctionality(t, handler)
}

func TestFiberRouterGroup_Group(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkFiber,
	}

	handler, err := NewHandler(conf)
	assert.NoError(t, err)

	testRouterGroupFunctionality(t, handler)
}

func TestFiberRouter_Use(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkFiber,
	}

	handler, err := NewHandler(conf)
	assert.NoError(t, err)

	testRouterUseFunctionality(t, handler)
}

func TestFiberRouter_Handle(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkFiber,
	}

	handler, err := NewHandler(conf)
	assert.NoError(t, err)

	// Test Handle method with Fiber-specific approach using different methods and paths
	router1 := handler.Handle("POST", "/fiber-post-handle", func(ctx unicontext.UniversalContext) error {
		return ctx.JSON(http.StatusCreated, map[string]string{"fiber": "post", "type": "handle"})
	})
	assert.NotNil(t, router1)

	router2 := handler.Handle("PATCH", "/fiber-patch-handle", func(ctx unicontext.UniversalContext) error {
		return ctx.JSON(http.StatusOK, map[string]string{"fiber": "patch", "type": "handle"})
	})
	assert.NotNil(t, router2)

	// Test POST
	req1 := httptest.NewRequest("POST", "/fiber-post-handle", nil)
	resp1, err1 := handler.Test(req1)
	assert.NoError(t, err1)
	defer func() { _ = resp1.Body.Close() }()
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	// Test PATCH
	req2 := httptest.NewRequest("PATCH", "/fiber-patch-handle", nil)
	resp2, err2 := handler.Test(req2)
	assert.NoError(t, err2)
	defer func() { _ = resp2.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

func TestFiberRouter_HTTPMethods(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkFiber,
	}

	handler, err := NewHandler(conf)
	assert.NoError(t, err)

	testHTTPMethods(t, handler)
}

func TestFiberRouter_Any(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkFiber,
	}

	handler, err := NewHandler(conf)
	assert.NoError(t, err)

	router := handler.Any("/any", func(ctx unicontext.UniversalContext) error {
		return ctx.JSON(http.StatusOK, map[string]string{"method": "any"})
	})
	assert.NotNil(t, router)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/any", nil)
			resp, err := handler.Test(req)
			assert.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), "any")
		})
	}
}

func TestFiberRouter_Match(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkFiber,
	}

	handler, err := NewHandler(conf)
	assert.NoError(t, err)

	// Test Match method with specific methods
	methods := []string{"GET", "POST", "PUT"}
	router := handler.Match(methods, "/match-test", func(ctx unicontext.UniversalContext) error {
		return ctx.JSON(http.StatusOK, map[string]string{"method": "match"})
	})
	assert.NotNil(t, router)

	// Test with allowed methods
	for _, method := range methods {
		t.Run(fmt.Sprintf("allowed_%s", method), func(t *testing.T) {
			req := httptest.NewRequest(method, "/match-test", nil)
			resp, err := handler.Test(req)
			assert.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), "match")
		})
	}

	// Test with disallowed method
	t.Run("disallowed_DELETE", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/match-test", nil)
		resp, err := handler.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Fiber returns 405 Method Not Allowed for disallowed methods on existing routes
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

func TestFiberHandler_Stop(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := &serverconf.HTTPConfig{
		Addr:          fmt.Sprintf("localhost:%d", port),
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkFiber,
	}

	handler, err := NewFiberHandler(conf)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	fiberHandler := handler.(*FiberHandler)

	// Test stopping without starting (should not error)
	err = fiberHandler.Stop()
	assert.NoError(t, err)
}

func TestFiberHandler_Name(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkFiber,
	}

	handler, err := NewFiberHandler(conf)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	name := handler.Name()
	assert.Contains(t, name, "fiber")
	assert.Contains(t, name, "localhost:8080")
	assert.Contains(t, name, "server")
}

func TestNewFiberHandler_EdgeCases(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		handler, err := NewFiberHandler(nil)
		assert.Nil(t, handler)
		assert.Equal(t, ErrEmptyHTTPConf, err)
	})

	t.Run("with timeouts", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          "localhost:8080",
			Mode:          "release",
			FrameworkType: serverconf.FrameworkFiber,
			ReadTimeout:   30,
			WriteTimeout:  30,
			IdleTimeout:   60,
		}

		handler, err := NewFiberHandler(conf)
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		fiberHandler := handler.(*FiberHandler)
		config := fiberHandler.app.Config()
		assert.Equal(t, 30*time.Second, config.ReadTimeout)
		assert.Equal(t, 30*time.Second, config.WriteTimeout)
		assert.Equal(t, 60*time.Second, config.IdleTimeout)
	})

	t.Run("release mode", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          "localhost:8080",
			Mode:          "release",
			FrameworkType: serverconf.FrameworkFiber,
		}

		handler, err := NewFiberHandler(conf)
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		fiberHandler := handler.(*FiberHandler)
		config := fiberHandler.app.Config()
		assert.True(t, config.Prefork) // Should be true in release mode
	})
}
