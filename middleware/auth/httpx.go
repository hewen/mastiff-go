// Package auth provides authentication and authorization middleware for fiber.
package auth

import (
	"net/http"

	"github.com/hewen/mastiff-go/config/middlewareconf/authconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/contextkeys"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// HttpxMiddleware is a fiber middleware for authentication and authorization.
func HttpxMiddleware(conf *authconf.Config) func(unicontext.UniversalContext) error {
	return func(c unicontext.UniversalContext) error {
		token := extractTokenFromHeader(c.Header(conf.HeaderKey), conf.TokenPrefixes)
		if token != "" {
			authInfo, err := validateJWTToken(token, conf.JWTSecret)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}

			ctx := unicontext.ContextFrom(c)
			logger.NewLoggerWithContext(ctx).Infof("auth info: %v", authInfo.Claims)
			ctx = contextkeys.SetAuthInfo(ctx, authInfo)
			ctx = contextkeys.SetUserID(ctx, authInfo.UserID)
			unicontext.InjectContext(ctx, c)
			return c.Next()
		}

		if isWhiteListed(c.FullPath(), conf.WhiteList) {
			return c.Next()
		} else {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "forbidden"})
		}
	}
}
