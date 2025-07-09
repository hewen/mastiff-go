// Package auth provides authentication and authorization middleware for gin.
package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinAuthMiddleware is a gin middleware for authentication and authorization.
func GinAuthMiddleware(conf Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isWhiteListed(c.Request.URL.Path, conf.WhiteList) {
			c.Next()
			return
		}

		token := extractTokenFromHeader(c.GetHeader(conf.HeaderKey), conf.TokenPrefixes)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		info, err := validateJWTToken(token, conf.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		SetAuthInfoToGin(c, info)
		c.Next()
	}
}
