package unicontext

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/internal/contextkeys"
)

// ContextFrom extracts context.Context from gin.Context, fiber.Ctx or context.Context itself.
// If none matched, returns context.Background.
func ContextFrom(v any) context.Context {
	// NOTE: Order matters in type switch — match *gin.Context and *fiber.Ctx
	// before context.Context to avoid premature capture.
	switch c := v.(type) {
	case UniversalContext:
		if val, ok := c.Get(contextkeys.ContextKey); ok && val != nil {
			if ctx, ok := val.(context.Context); ok {
				return ctx
			}
		}
	case *gin.Context:
		if req := c.Request; req != nil {
			return req.Context()
		}
	case *fiber.Ctx:
		if val := c.Locals(contextkeys.ContextKey); val != nil {
			if ctx, ok := val.(context.Context); ok {
				return ctx
			}
		}
	case context.Context:
		return c
	}

	return context.Background()
}

// InjectContext sets the updated context.Context back into the carrier (gin/fiber).
func InjectContext(ctx context.Context, carrier any) {
	switch c := carrier.(type) {
	case UniversalContext:
		c.Set(contextkeys.ContextKey, ctx)
	case *gin.Context:
		c.Request = c.Request.WithContext(ctx)
	case *fiber.Ctx:
		c.Locals(contextkeys.ContextKey, ctx)
	}
}
