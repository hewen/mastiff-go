// Package handler provides a unified socket abstraction over gnet.
package handler

// contextKey is a custom type to avoid key collisions in context.
type contextKey string

const (
	// ContextKeyDeviceID is the key for device ID in context.
	ContextKeyDeviceID contextKey = "device_id"
)

// SetContextValue sets a value in the connection context.
func SetContextValue(c Conn, key string, value any) {
	ctx, ok := c.Context().(map[string]any)
	if !ok {
		ctx = make(map[string]any)
	}
	ctx[key] = value
	c.SetContext(ctx)
}

// GetContextValue gets a value from the connection context.
func GetContextValue[T any](c Conn, key string) (T, bool) {
	var zero T
	ctx, ok := c.Context().(map[string]any)
	if !ok {
		return zero, false
	}
	val, ok := ctx[key]
	if !ok {
		return zero, false
	}
	casted, ok := val.(T)
	if !ok {
		return zero, false
	}
	return casted, true
}
