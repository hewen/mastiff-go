// Package handler provides a context interface for HTTP handlers.
package handler

import "context"

// Context is the interface for HTTP handlers.
type Context interface {
	JSON(code int, obj any) error
	BindJSON(obj any) error
	Set(key string, val any)
	RequestContext() context.Context
}
