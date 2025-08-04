// Package contextkeys provides strongly typed context keys and helper functions.
package contextkeys

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// ContextKey is the key for context in fiber.
	ContextKey = "context"
)

// ctxKey is a custom type to avoid key collisions in context.
type ctxKey string

func (k ctxKey) String() string {
	return "context key: " + string(k)
}

// Predefined context keys for common use cases.
var (
	// LoggerTraceIDKey is the key for logger trace id.
	LoggerTraceIDKey = ctxKey("logid")
	// AuthInfoKey is the key for auth info.
	AuthInfoKey = ctxKey("auth_info")
	// UserIDKey is the key for user id.
	UserIDKey = ctxKey("user_id")
	// SQLBeginTimeKey is the key for sql begin time.
	SQLBeginTimeKey = ctxKey("sql_begin_time")
	// RedisBeginTimeKey is the key for redis begin time.
	RedisBeginTimeKey = ctxKey("redis_begin_time")
)

// Info represents authentication information.
type Info struct {
	Claims jwt.MapClaims
	UserID string
}

// SetValue sets a typed value into the context using a custom key.
func SetValue[T any](ctx context.Context, key ctxKey, val T) context.Context {
	return context.WithValue(ctx, key, val)
}

// GetValue retrieves a typed value from the context by key.
func GetValue[T any](ctx context.Context, key ctxKey) (T, bool) {
	val := ctx.Value(key)
	v, ok := val.(T)
	return v, ok
}

// SetAuthInfo sets auth info into the context.
func SetAuthInfo(ctx context.Context, info *Info) context.Context {
	return context.WithValue(ctx, AuthInfoKey, info)
}

// GetAuthInfo retrieves auth info from the context.
func GetAuthInfo(ctx context.Context) (*Info, bool) {
	info, ok := ctx.Value(AuthInfoKey).(*Info)
	return info, ok
}

// SetUserID sets user id into the context.
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID retrieves user id from the context.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// SetTraceID sets trace id into the context.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, LoggerTraceIDKey, traceID)
}

// GetTraceID retrieves trace id from the context.
func GetTraceID(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(LoggerTraceIDKey).(string)
	return traceID, ok
}

// SetSQLBeginTime sets sql begin time into the context.
func SetSQLBeginTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, SQLBeginTimeKey, t)
}

// GetSQLBeginTime retrieves sql begin time from the context.
func GetSQLBeginTime(ctx context.Context) (time.Time, bool) {
	t, ok := ctx.Value(SQLBeginTimeKey).(time.Time)
	return t, ok
}

// SetRedisBeginTime sets redis begin time into the context.
func SetRedisBeginTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, RedisBeginTimeKey, t)
}

// GetRedisBeginTime retrieves redis begin time from the context.
func GetRedisBeginTime(ctx context.Context) (time.Time, bool) {
	t, ok := ctx.Value(RedisBeginTimeKey).(time.Time)
	return t, ok
}
