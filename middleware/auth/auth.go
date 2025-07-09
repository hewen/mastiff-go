// Package auth provides authentication and authorization middleware.
package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"google.golang.org/grpc/metadata"
)

// Info represents the authentication information.
type Info struct {
	UserID string
	Claims jwt.MapClaims
}

// SetAuthInfoToContext sets auth info to context.
func SetAuthInfoToContext(ctx context.Context, info *Info) context.Context {
	return context.WithValue(ctx, contextkeys.AuthInfoKey, info)
}

// GetAuthInfoFromContext gets auth info from context.
func GetAuthInfoFromContext(ctx context.Context) (*Info, bool) {
	info, ok := ctx.Value(contextkeys.AuthInfoKey).(*Info)
	return info, ok
}

// SetAuthInfoToGin sets auth info to gin context.
func SetAuthInfoToGin(c *gin.Context, info *Info) {
	c.Set(string(contextkeys.AuthInfoKey), info)
}

// GetAuthInfoFromGin gets auth info from gin context.
func GetAuthInfoFromGin(c *gin.Context) (*Info, bool) {
	val, ok := c.Get(string(contextkeys.AuthInfoKey))
	if !ok {
		return nil, false
	}
	info, ok := val.(*Info)
	return info, ok
}

// ExtractTokenFromHeader extracts token from header.
func extractTokenFromHeader(value string, prefixes []string) string {
	for _, p := range prefixes {
		if value, ok := strings.CutPrefix(value, p); ok {
			return strings.TrimSpace(value)
		}
	}
	return value
}

// ExtractTokenFromGrpcMetadata extracts token from grpc metadata.
func extractTokenFromGrpcMetadata(md metadata.MD, headerKey string, prefixes []string) string {
	authHeaders := md.Get(headerKey)
	for _, h := range authHeaders {
		return extractTokenFromHeader(h, prefixes)
	}
	return ""
}

// IsWhiteListed checks if the path is in the whitelist.
func isWhiteListed(path string, whitelist []string) bool {
	for _, w := range whitelist {
		if (strings.HasSuffix(w, "/") && strings.HasPrefix(path, w)) ||
			path == w {
			return true
		}
	}
	return false
}

// ValidateJWTToken validates a JWT token string with the given secret.
func validateJWTToken(tokenStr, secret string) (*Info, error) {
	t, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(_ *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !t.Valid {
		return nil, errors.New("invalid jwt")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("missing user_id")
	}

	return &Info{
		UserID: userID,
		Claims: claims,
	}, nil
}

// GenerateJWTToken creates a JWT token string with custom claims and a secret.
// It ensures standard claims like "exp" and "iat" are injected.
func GenerateJWTToken(claims map[string]any, secret string, expiration time.Duration) (string, error) {
	// Set standard claims
	now := time.Now()
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(expiration).Unix()

	// Use the claims to create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	return token.SignedString([]byte(secret))
}
