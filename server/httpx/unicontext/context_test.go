package unicontext

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/pkg/contextkeys"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestContextFrom_GinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.WithValue(context.Background(), contextkeys.LoggerTraceIDKey, "12345")

	// mock gin.Context
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	c := &gin.Context{
		Request: req,
	}

	got := ContextFrom(c)
	traceID, ok := contextkeys.GetTraceID(got)

	assert.True(t, ok)
	assert.Equal(t, "12345", traceID)
}

func TestContextFrom_FiberCtx(t *testing.T) {
	app := fiber.New()
	testCtx := context.WithValue(context.Background(), contextkeys.LoggerTraceIDKey, "abcde")

	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals(contextkeys.ContextKey, testCtx)

	got := ContextFrom(c)
	traceID, ok := contextkeys.GetTraceID(got)

	assert.True(t, ok)
	assert.Equal(t, "abcde", traceID)
}

func TestContextFrom_Unicontext(t *testing.T) {
	testCtx := context.WithValue(context.Background(), contextkeys.LoggerTraceIDKey, "abcde")

	app := fiber.New()
	ctx := &FiberContext{
		Ctx: app.AcquireCtx(&fasthttp.RequestCtx{}),
	}
	ctx.Set(contextkeys.ContextKey, testCtx)

	got := ContextFrom(ctx)
	traceID, ok := contextkeys.GetTraceID(got)

	assert.True(t, ok)
	assert.Equal(t, "abcde", traceID)
}

func TestContextFrom_Context(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.LoggerTraceIDKey, "xyz")

	got := ContextFrom(ctx)
	traceID, ok := contextkeys.GetTraceID(got)

	assert.True(t, ok)
	assert.Equal(t, "xyz", traceID)
}

func TestContextFrom_UnknownType(t *testing.T) {
	got := ContextFrom("unknown")
	assert.NotNil(t, got)
	assert.Equal(t, context.Background(), got)
}

func TestInjectContext_Gin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := context.WithValue(context.Background(), contextkeys.LoggerTraceIDKey, "gin-trace")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	c := &gin.Context{Request: req}

	InjectContext(ctx, c)

	got := c.Request.Context()
	traceID, ok := contextkeys.GetTraceID(got)
	assert.True(t, ok)
	assert.Equal(t, "gin-trace", traceID)
}

func TestInjectContext_Fiber(t *testing.T) {
	app := fiber.New()
	ctx := context.WithValue(context.Background(), contextkeys.LoggerTraceIDKey, "fiber-trace")
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	InjectContext(ctx, c)

	val := c.Locals(contextkeys.ContextKey)
	assert.NotNil(t, val)

	retrieved, ok := val.(context.Context)
	assert.True(t, ok)

	traceID, found := contextkeys.GetTraceID(retrieved)
	assert.True(t, found)
	assert.Equal(t, "fiber-trace", traceID)
}

func TestInjectContext_Unicontext(t *testing.T) {
	app := fiber.New()
	ctx := context.WithValue(context.Background(), contextkeys.LoggerTraceIDKey, "fiber-trace")

	c := &FiberContext{
		Ctx: app.AcquireCtx(&fasthttp.RequestCtx{}),
	}

	InjectContext(ctx, c)

	val, ok := c.Get(contextkeys.ContextKey)
	assert.NotNil(t, val)
	assert.True(t, ok)

	retrieved, ok := val.(context.Context)
	assert.True(t, ok)

	traceID, found := contextkeys.GetTraceID(retrieved)
	assert.True(t, found)
	assert.Equal(t, "fiber-trace", traceID)
}
