package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to test pprof endpoints for any handler.
func testPprofEndpoints(t *testing.T, handler HTTPHandler) {
	// Test pprof endpoints
	pprofEndpoints := []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/profile?seconds=1", // Use 1 second for faster testing
		"/debug/pprof/symbol",
		"/debug/pprof/trace?seconds=1", // Use 1 second for faster testing
	}

	for _, endpoint := range pprofEndpoints {
		t.Run(fmt.Sprintf("endpoint_%s", endpoint), func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			// Add timeout for profile and trace endpoints
			if strings.Contains(endpoint, "profile") || strings.Contains(endpoint, "trace") {
				resp, err := handler.Test(req, 2000) // 2 second timeout
				require.NoError(t, err)
				defer func() {
					_ = resp.Body.Close()
				}()
				// All pprof endpoints should be accessible (200 or other valid status)
				assert.True(t, resp.StatusCode < 500, "endpoint %s should not return server error", endpoint)
			} else {
				resp, err := handler.Test(req)
				require.NoError(t, err)
				defer func() {
					_ = resp.Body.Close()
				}()
				// All pprof endpoints should be accessible (200 or other valid status)
				assert.True(t, resp.StatusCode < 500, "endpoint %s should not return server error", endpoint)
			}
		})
	}
}

// Helper function to test metrics and pprof together.
func testMetricsAndPprof(t *testing.T, handler HTTPHandler) {
	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := handler.Test(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test pprof endpoint
	req = httptest.NewRequest("GET", "/debug/pprof/", nil)
	resp, err = handler.Test(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.True(t, resp.StatusCode < 500)
}

// nolint
func TestNewHandler(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		handler, err := NewHandler(nil)
		assert.Nil(t, handler)
		assert.Equal(t, ErrEmptyHTTPConf, err)
	})

	t.Run("gin handler", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8080",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkGin,
		}

		handler, err := NewHandler(conf)
		assert.NoError(t, err)
		assert.NotNil(t, handler)
		assert.Contains(t, handler.Name(), "gin")
	})

	t.Run("fiber handler", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8081",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkFiber,
		}

		handler, err := NewHandler(conf)
		assert.NoError(t, err)
		assert.NotNil(t, handler)
		assert.Contains(t, handler.Name(), "fiber")
	})

	t.Run("unknown framework type", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8082",
			Mode:          "debug",
			FrameworkType: "unknown",
		}

		assert.Panics(t, func() {
			NewHandler(conf)
		})
	})

	t.Run("with server options", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8083",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkGin,
		}

		optionCalled := false
		testOption := func(h HTTPHandler) {
			optionCalled = true
			assert.NotNil(t, h)
		}

		handler, err := NewHandler(conf, testOption)
		assert.NoError(t, err)
		assert.NotNil(t, handler)
		assert.True(t, optionCalled)
	})

	t.Run("with multiple server options", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8084",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkFiber,
		}

		var callOrder []int
		option1 := func(h HTTPHandler) {
			callOrder = append(callOrder, 1)
		}
		option2 := func(h HTTPHandler) {
			callOrder = append(callOrder, 2)
		}

		handler, err := NewHandler(conf, option1, option2)
		assert.NoError(t, err)
		assert.NotNil(t, handler)
		assert.Equal(t, []int{1, 2}, callOrder)
	})
}

func TestWithMetrics(t *testing.T) {
	t.Run("gin handler with metrics", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8085",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkGin,
		}

		handler, err := NewHandler(conf, WithMetrics())
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		// Test that metrics endpoint is available
		req := httptest.NewRequest("GET", "/metrics", nil)
		resp, err := handler.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		// Metrics endpoint should return 200 and contain prometheus metrics
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		// Should contain some prometheus metrics
		bodyStr := string(body)
		assert.Contains(t, bodyStr, "# HELP")
	})

	t.Run("fiber handler with metrics", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8086",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkFiber,
		}

		handler, err := NewHandler(conf, WithMetrics())
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		// Test that metrics endpoint is available
		req := httptest.NewRequest("GET", "/metrics", nil)
		resp, err := handler.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		// Metrics endpoint should return 200 and contain prometheus metrics
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		// Should contain some prometheus metrics
		bodyStr := string(body)
		assert.Contains(t, bodyStr, "# HELP")
	})
}

func TestWithPprof(t *testing.T) {
	t.Run("gin handler with pprof", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8087",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkGin,
		}

		handler, err := NewHandler(conf, WithPprof())
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		testPprofEndpoints(t, handler)
	})

	t.Run("fiber handler with pprof", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8088",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkFiber,
		}

		handler, err := NewHandler(conf, WithPprof())
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		testPprofEndpoints(t, handler)
	})
}

func TestWithMetricsAndPprof(t *testing.T) {
	t.Run("gin handler with both metrics and pprof", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8089",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkGin,
		}

		handler, err := NewHandler(conf, WithMetrics(), WithPprof())
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		testMetricsAndPprof(t, handler)
	})

	t.Run("fiber handler with both metrics and pprof", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8090",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkFiber,
		}

		handler, err := NewHandler(conf, WithMetrics(), WithPprof())
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		testMetricsAndPprof(t, handler)
	})
}

func TestServerOptionIntegration(t *testing.T) {
	t.Run("custom routes with options", func(t *testing.T) {
		conf := &serverconf.HTTPConfig{
			Addr:          ":8091",
			Mode:          "debug",
			FrameworkType: serverconf.FrameworkGin,
		}

		customOption := func(h HTTPHandler) {
			h.Get("/custom", func(ctx unicontext.UniversalContext) error {
				return ctx.JSON(200, map[string]string{"message": "custom"})
			})
		}

		handler, err := NewHandler(conf, WithMetrics(), customOption, WithPprof())
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		// Test custom endpoint
		req := httptest.NewRequest("GET", "/custom", nil)
		resp, err := handler.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "custom")

		// Test that metrics still works
		req = httptest.NewRequest("GET", "/metrics", nil)
		resp, err = handler.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
