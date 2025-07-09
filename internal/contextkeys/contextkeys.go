// Package contextkeys provides context keys.
package contextkeys

// ContextKeys is the context keys.
type ctxKey string

// String returns the string representation of the context key.
func (k ctxKey) String() string {
	return "context key: " + string(k)
}

var (
	// LoggerTraceIDKey is the key for logger trace id.
	LoggerTraceIDKey = ctxKey("logid")
	// AuthInfoKey is the key for auth info.
	AuthInfoKey = ctxKey("auth_info")
	// SQLBeginTimeKey is the key for sql begin time.
	SQLBeginTimeKey = ctxKey("sql_begin_time")
	// RedisBeginTimeKey is the key for redis begin time.
	RedisBeginTimeKey = ctxKey("redis_begin_time")
)
