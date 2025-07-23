package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/middlewareconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware/logging"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"github.com/stretchr/testify/assert"
)

// Helper function to test basic handler functionality.
func testBasicHandlerFunctionality(t *testing.T, handler HTTPHandler) {
	// Add a test route
	handler.Get("/test", func(ctx unicontext.UniversalContext) error {
		return ctx.JSON(http.StatusOK, map[string]string{"message": "test"})
	})

	t.Run("successful request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := handler.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "test")
	})

	t.Run("with timeout", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := handler.Test(req, 1000) // 1 second timeout
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("404 request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		resp, err := handler.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// Helper function to test router group functionality.
func testRouterGroupFunctionality(t *testing.T, handler HTTPHandler) {
	// Create a group
	apiGroup := handler.Group("/api")
	assert.NotNil(t, apiGroup)

	// Add route to group
	apiGroup.Get("/test", func(ctx unicontext.UniversalContext) error {
		return ctx.JSON(http.StatusOK, map[string]string{"message": "group test"})
	})

	// Test the grouped route
	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := handler.Test(req)
	assert.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "group test")
}

// Helper function to test router Use functionality.
func testRouterUseFunctionality(t *testing.T, handler HTTPHandler) {
	middlewareCalled := false
	middleware := func(ctx unicontext.UniversalContext) error {
		middlewareCalled = true
		return ctx.Next()
	}

	// Test handler-level Use method
	router := handler.Use(middleware)
	assert.NotNil(t, router)

	// Test router-level Use method by creating a group and calling Use on it
	group := handler.Group("/api")
	groupRouter := group.Use(middleware)
	assert.NotNil(t, groupRouter)

	// Add route to the group
	group.Get("/test", func(ctx unicontext.UniversalContext) error {
		return ctx.JSON(http.StatusOK, map[string]string{"message": "test"})
	})

	// Test that middleware is called
	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := handler.Test(req)
	assert.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.True(t, middlewareCalled)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Helper function to test HTTP methods.
func testHTTPMethods(t *testing.T, handler HTTPHandler) {
	// Test all HTTP methods
	methods := map[string]func(string, ...HTTPHandlerFunc) Router{
		"GET":     handler.Get,
		"POST":    handler.Post,
		"PUT":     handler.Put,
		"DELETE":  handler.Delete,
		"PATCH":   handler.Patch,
		"HEAD":    handler.Head,
		"OPTIONS": handler.Options,
	}

	for method, routerFunc := range methods {
		t.Run(method, func(t *testing.T) {
			path := fmt.Sprintf("/%s-test", strings.ToLower(method))
			router := routerFunc(path, func(ctx unicontext.UniversalContext) error {
				return ctx.JSON(http.StatusOK, map[string]string{"method": method})
			})
			assert.NotNil(t, router)

			req := httptest.NewRequest(method, path, nil)
			resp, err := handler.Test(req)
			assert.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			if method != "HEAD" { // HEAD requests don't return body
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.Contains(t, string(body), method)
			}
		})
	}
}

func TestGinHandler(t *testing.T) {
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
		FrameworkType: serverconf.FrameworkGin,
	}

	s, err := NewHandler(conf)
	assert.Nil(t, err)
	s.Use(logging.HttpxMiddleware())

	s.Get("/test", func(c unicontext.UniversalContext) error {
		ctx := contextkeys.ContextFrom(c)
		l := logger.NewLoggerWithContext(ctx)
		l.Infof("test")

		return c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	go func() {
		defer func() {
			_ = s.Stop()
		}()
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
}

func TestGinHandler_Stop(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := &serverconf.HTTPConfig{
		Addr:          fmt.Sprintf("localhost:%d", port),
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkGin,
	}

	handler, err := NewGinHandler(conf)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ginHandler := handler.(*GinHandler)

	// Test stopping without starting (should not error)
	err = ginHandler.Stop()
	assert.NoError(t, err)
}

func TestGinHandler_Name(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkGin,
	}

	handler, err := NewGinHandler(conf)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	name := handler.Name()
	assert.Contains(t, name, "gin")
	assert.Contains(t, name, "localhost:8080")
	assert.Contains(t, name, "server")
}

func TestGinHandler_Test(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkGin,
	}

	handler, err := NewGinHandler(conf)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	testBasicHandlerFunctionality(t, handler)
}

func TestGinRouterGroup_Group(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkGin,
	}

	handler, err := NewGinHandler(conf)
	assert.NoError(t, err)

	testRouterGroupFunctionality(t, handler)
}

func TestGinRouter_Use(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkGin,
	}

	handler, err := NewGinHandler(conf)
	assert.NoError(t, err)

	testRouterUseFunctionality(t, handler)
}

func TestGinRouter_HTTPMethods(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkGin,
	}

	handler, err := NewGinHandler(conf)
	assert.NoError(t, err)

	testHTTPMethods(t, handler)
}

func TestGinRouter_Handle(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkGin,
	}

	handler, err := NewGinHandler(conf)
	assert.NoError(t, err)

	// Test Handle method with multiple HTTP methods to ensure coverage
	testCases := []struct {
		method string
		path   string
		status int
	}{
		{"CONNECT", "/gin-connect", http.StatusOK},
		{"TRACE", "/gin-trace", http.StatusAccepted},
	}

	for _, tc := range testCases {
		router := handler.Handle(tc.method, tc.path, func(ctx unicontext.UniversalContext) error {
			return ctx.JSON(tc.status, map[string]string{"gin_method": tc.method})
		})
		assert.NotNil(t, router)

		req := httptest.NewRequest(tc.method, tc.path, nil)
		resp, err2 := handler.Test(req)
		assert.NoError(t, err2)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, tc.status, resp.StatusCode)
		body, err3 := io.ReadAll(resp.Body)
		assert.NoError(t, err3)
		assert.Contains(t, string(body), tc.method)
	}
}

func TestGinRouter_Any(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkGin,
	}

	handler, err := NewGinHandler(conf)
	assert.NoError(t, err)

	// Test Any method - should respond to multiple HTTP methods
	router := handler.Any("/any-endpoint", func(ctx unicontext.UniversalContext) error {
		method := ctx.Method()
		return ctx.JSON(http.StatusOK, map[string]string{"received_method": method})
	})
	assert.NotNil(t, router)

	// Test with different methods
	testMethods := []string{"GET", "POST", "PUT", "DELETE"}
	for _, method := range testMethods {
		req := httptest.NewRequest(method, "/any-endpoint", nil)
		resp, err := handler.Test(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), method)
	}
}

func TestGinRouter_Match(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "localhost:8080",
		Mode:          "debug",
		FrameworkType: serverconf.FrameworkGin,
	}

	handler, err := NewGinHandler(conf)
	assert.NoError(t, err)

	// Test Match method with specific methods
	methods := []string{"GET", "POST", "PUT"}
	router := handler.Match(methods, "/match-endpoint", func(ctx unicontext.UniversalContext) error {
		return ctx.JSON(http.StatusOK, map[string]string{"matched": "true"})
	})
	assert.NotNil(t, router)

	// Test with allowed methods
	for _, method := range methods {
		req := httptest.NewRequest(method, "/match-endpoint", nil)
		resp, err2 := handler.Test(req)
		assert.NoError(t, err2)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err3 := io.ReadAll(resp.Body)
		assert.NoError(t, err3)
		assert.Contains(t, string(body), "matched")
	}

	// Test with non-allowed method (Gin returns 404 for unmatched routes)
	req := httptest.NewRequest("DELETE", "/match-endpoint", nil)
	resp, err := handler.Test(req)
	assert.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestNewGinHandler_EdgeCases(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		handler, err := NewGinHandler(nil)
		assert.Nil(t, handler)
		assert.Equal(t, ErrEmptyHTTPConf, err)
	})

	t.Run("with timeouts", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          "localhost:8080",
			Mode:          "release",
			FrameworkType: serverconf.FrameworkGin,
			ReadTimeout:   30,
			WriteTimeout:  30,
			IdleTimeout:   60,
		}

		handler, err := NewGinHandler(conf)
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		ginHandler := handler.(*GinHandler)
		assert.Equal(t, "localhost:8080", ginHandler.server.Addr)
		assert.Equal(t, 30*time.Second, ginHandler.server.ReadTimeout)
		assert.Equal(t, 30*time.Second, ginHandler.server.WriteTimeout)
		assert.Equal(t, 60*time.Second, ginHandler.server.IdleTimeout)
	})

	t.Run("release mode", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          "localhost:8080",
			Mode:          "release",
			FrameworkType: serverconf.FrameworkGin,
		}

		handler, err := NewGinHandler(conf)
		assert.NoError(t, err)
		assert.NotNil(t, handler)
		assert.Equal(t, gin.ReleaseMode, gin.Mode())
	})
}
