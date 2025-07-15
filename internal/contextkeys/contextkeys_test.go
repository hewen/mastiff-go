package contextkeys

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestCtxKeyString(t *testing.T) {
	key := ctxKey("custom")
	assert.Equal(t, "context key: custom", key.String())
}

func TestSetAndGetAuthInfo(t *testing.T) {
	ctx := context.Background()
	info := &Info{
		UserID: "user123",
		Claims: jwt.MapClaims{"role": "admin"},
	}

	ctx = SetAuthInfo(ctx, info)
	val, ok := GetAuthInfo(ctx)

	assert.True(t, ok)
	assert.Equal(t, "user123", val.UserID)
	assert.Equal(t, "admin", val.Claims["role"])
}

func TestSetAndGetUserID(t *testing.T) {
	ctx := context.Background()
	ctx = SetUserID(ctx, "u456")

	userID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "u456", userID)
}

func TestSetAndGetTraceID(t *testing.T) {
	ctx := context.Background()
	ctx = SetTraceID(ctx, "trace-abc")

	traceID, ok := GetTraceID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "trace-abc", traceID)
}

func TestSetAndGetSQLBeginTime(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	ctx = SetSQLBeginTime(ctx, now)

	got, ok := GetSQLBeginTime(ctx)
	assert.True(t, ok)
	assert.WithinDuration(t, now, got, time.Millisecond)
}

func TestSetAndGetRedisBeginTime(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	ctx = SetRedisBeginTime(ctx, now)

	got, ok := GetRedisBeginTime(ctx)
	assert.True(t, ok)
	assert.WithinDuration(t, now, got, time.Millisecond)
}

func TestGenericSetAndGetValue(t *testing.T) {
	ctx := context.Background()

	type customType struct {
		Foo string
		Bar int
	}

	val := customType{"baz", 42}
	ctx = SetValue(ctx, ctxKey("custom"), val)

	got, ok := GetValue[customType](ctx, ctxKey("custom"))
	assert.True(t, ok)
	assert.Equal(t, "baz", got.Foo)
	assert.Equal(t, 42, got.Bar)
}

func TestContextFrom_GinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.WithValue(context.Background(), LoggerTraceIDKey, "12345")

	// mock gin.Context
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	c := &gin.Context{
		Request: req,
	}

	got := ContextFrom(c)
	traceID, ok := GetTraceID(got)

	assert.True(t, ok)
	assert.Equal(t, "12345", traceID)
}

func TestContextFrom_FiberCtx(t *testing.T) {
	app := fiber.New()
	testCtx := context.WithValue(context.Background(), LoggerTraceIDKey, "abcde")

	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals(ContextKey, testCtx)

	got := ContextFrom(c)
	traceID, ok := GetTraceID(got)

	assert.True(t, ok)
	assert.Equal(t, "abcde", traceID)
}

func TestContextFrom_Context(t *testing.T) {
	ctx := context.WithValue(context.Background(), LoggerTraceIDKey, "xyz")

	got := ContextFrom(ctx)
	traceID, ok := GetTraceID(got)

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

	ctx := context.WithValue(context.Background(), LoggerTraceIDKey, "gin-trace")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	c := &gin.Context{Request: req}

	InjectContext(ctx, c)

	got := c.Request.Context()
	traceID, ok := GetTraceID(got)
	assert.True(t, ok)
	assert.Equal(t, "gin-trace", traceID)
}

func TestInjectContext_Fiber(t *testing.T) {
	app := fiber.New()
	ctx := context.WithValue(context.Background(), LoggerTraceIDKey, "fiber-trace")
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	InjectContext(ctx, c)

	val := c.Locals(ContextKey)
	assert.NotNil(t, val)

	retrieved, ok := val.(context.Context)
	assert.True(t, ok)

	traceID, found := GetTraceID(retrieved)
	assert.True(t, found)
	assert.Equal(t, "fiber-trace", traceID)
}
