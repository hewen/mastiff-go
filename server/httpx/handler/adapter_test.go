package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func testSingleGinHandler(t *testing.T) {
	called := false
	handler := func(ctx unicontext.UniversalContext) error {
		called = true
		return ctx.Text(200, "test")
	}

	ginHandlers := AsGinHandler(handler)
	assert.Len(t, ginHandlers, 1)

	// Test the converted handler
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	ginHandlers[0](c)
	assert.True(t, called)
}

func testMultipleGinHandlers(t *testing.T) {
	var callOrder []int

	handler1 := func(_ unicontext.UniversalContext) error {
		callOrder = append(callOrder, 1)
		return nil
	}

	handler2 := func(_ unicontext.UniversalContext) error {
		callOrder = append(callOrder, 2)
		return nil
	}

	ginHandlers := AsGinHandler(handler1, handler2)
	assert.Len(t, ginHandlers, 2)

	// Test the converted handlers
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	ginHandlers[0](c)
	ginHandlers[1](c)

	assert.Equal(t, []int{1, 2}, callOrder)
}

// nolint
func TestAsFiberHandler(t *testing.T) {
	t.Run("single handler", func(t *testing.T) {
		called := false
		handler := func(ctx unicontext.UniversalContext) error {
			called = true
			return ctx.Text(200, "test")
		}

		fiberHandlers := AsFiberHandler(handler)
		assert.Len(t, fiberHandlers, 1)

		// Test the converted handler
		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)

		err := fiberHandlers[0](ctx)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("multiple handlers", func(t *testing.T) {
		var callOrder []int

		handler1 := func(_ unicontext.UniversalContext) error {
			callOrder = append(callOrder, 1)
			return nil
		}

		handler2 := func(_ unicontext.UniversalContext) error {
			callOrder = append(callOrder, 2)
			return nil
		}

		fiberHandlers := AsFiberHandler(handler1, handler2)
		assert.Len(t, fiberHandlers, 2)

		// Test the converted handlers
		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)

		err := fiberHandlers[0](ctx)
		assert.NoError(t, err)

		err = fiberHandlers[1](ctx)
		assert.NoError(t, err)

		assert.Equal(t, []int{1, 2}, callOrder)
	})

	t.Run("handler with error", func(t *testing.T) {
		expectedErr := assert.AnError
		handler := func(ctx unicontext.UniversalContext) error {
			return expectedErr
		}

		fiberHandlers := AsFiberHandler(handler)
		assert.Len(t, fiberHandlers, 1)

		// Test the converted handler
		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)

		err := fiberHandlers[0](ctx)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("empty handlers", func(t *testing.T) {
		fiberHandlers := AsFiberHandler()
		assert.Len(t, fiberHandlers, 0)
	})
}

func TestAsGinHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("single handler", testSingleGinHandler)
	t.Run("multiple handlers", testMultipleGinHandlers)

	t.Run("handler with error", func(t *testing.T) {
		expectedErr := assert.AnError
		handler := func(_ unicontext.UniversalContext) error {
			return expectedErr
		}

		ginHandlers := AsGinHandler(handler)
		assert.Len(t, ginHandlers, 1)

		// Test the converted handler (Gin handlers don't return errors, so error is ignored)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// Should not panic even with error
		assert.NotPanics(t, func() {
			ginHandlers[0](c)
		})
	})

	t.Run("empty handlers", func(t *testing.T) {
		ginHandlers := AsGinHandler()
		assert.Len(t, ginHandlers, 0)
	})
}

func TestFromHTTPHandlerFunc(t *testing.T) {
	t.Run("successful handler", func(t *testing.T) {
		called := false
		httpHandler := func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test response"))
		}

		universalHandler := FromHTTPHandlerFunc(httpHandler)
		assert.NotNil(t, universalHandler)

		// Create a mock context
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		ctx := &unicontext.GinContext{Ctx: c}

		err := universalHandler(ctx)
		assert.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, "test response", w.Body.String())
	})

	t.Run("handler with panic", func(t *testing.T) {
		httpHandler := func(_ http.ResponseWriter, _ *http.Request) {
			panic("test panic")
		}

		universalHandler := FromHTTPHandlerFunc(httpHandler)
		assert.NotNil(t, universalHandler)

		// Create a mock context
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		ctx := &unicontext.GinContext{Ctx: c}

		assert.Panics(t, func() {
			_ = universalHandler(ctx)
		})
	})
}

func testSuccessfulHTTPHandler(t *testing.T) {
	called := false
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("handler response"))
	})

	universalHandler := FromHTTPHandler(httpHandler)
	assert.NotNil(t, universalHandler)

	// Create a mock context
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", nil)

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	ctx := &unicontext.GinContext{Ctx: c}

	err := universalHandler(ctx)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "handler response", w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
}

// Helper function to test HTTP handler with custom headers.
func testHTTPHandlerWithHeaders(t *testing.T) {
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"message":"success"}`))
	})

	universalHandler := FromHTTPHandler(httpHandler)
	assert.NotNil(t, universalHandler)

	// Create a mock context
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/test", nil)

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	ctx := &unicontext.GinContext{Ctx: c}

	err := universalHandler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, `{"message":"success"}`, w.Body.String())
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Equal(t, "test-value", w.Header().Get("X-Custom-Header"))
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestFromHTTPHandler(t *testing.T) {
	t.Run("successful handler", testSuccessfulHTTPHandler)
	t.Run("handler with custom headers", testHTTPHandlerWithHeaders)

	t.Run("nil handler", func(t *testing.T) {
		universalHandler := FromHTTPHandler(nil)
		assert.NotNil(t, universalHandler)

		// Create a mock context
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		ctx := &unicontext.GinContext{Ctx: c}

		assert.Panics(t, func() {
			_ = universalHandler(ctx)
		})
	})
}
