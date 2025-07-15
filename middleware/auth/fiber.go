// Package auth provides authentication and authorization middleware for fiber.
package auth

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/config/middleware/authconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
)

// FiberMiddleware is a fiber middleware for authentication and authorization.
func FiberMiddleware(conf *authconf.Config) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if isWhiteListed(c.Path(), conf.WhiteList) {
			return c.Next()
		}

		token := extractTokenFromHeader(c.Get(conf.HeaderKey), conf.TokenPrefixes)
		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "forbidden"})
		}

		info, err := validateJWTToken(token, conf.JWTSecret)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		ctx := contextkeys.ContextFrom(c)
		ctx = contextkeys.SetAuthInfo(ctx, info)
		ctx = contextkeys.SetUserID(ctx, info.UserID)
		contextkeys.InjectContext(ctx, c)
		return c.Next()
	}
}
