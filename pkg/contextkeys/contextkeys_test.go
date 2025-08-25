package contextkeys

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestCtxKeyString(t *testing.T) {
	key := ctxKey("custom")
	assert.Equal(t, "custom", key.String())
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
