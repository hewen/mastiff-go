// Package auth provides authentication and authorization middleware for fiber.
package auth

import (
	"net/http"

	"github.com/hewen/mastiff-go/config/middlewareconf/authconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// HttpxMiddleware is a fiber middleware for authentication and authorization.
func HttpxMiddleware(conf *authconf.Config) func(unicontext.UniversalContext) error {
	return func(c unicontext.UniversalContext) error {
		if isWhiteListed(c.FullPath(), conf.WhiteList) {
			return c.Next()
		}

		token := extractTokenFromHeader(c.Header(conf.HeaderKey), conf.TokenPrefixes)
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "forbidden"})
		}

		info, err := validateJWTToken(token, conf.JWTSecret)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		ctx := unicontext.ContextFrom(c)
		ctx = contextkeys.SetAuthInfo(ctx, info)
		ctx = contextkeys.SetUserID(ctx, info.UserID)
		unicontext.InjectContext(ctx, c)
		return c.Next()
	}
}
