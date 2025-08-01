// Package authconf provides authentication and authorization middleware.
package authconf

// Config defines auth-related configuration.
type Config struct {
	// Secret for JWT
	JWTSecret string
	// Path whitelist (exact match or prefix)
	WhiteList []string
	// HeaderKey  e.g., "Authorization"
	HeaderKey string
	// TokenPrefixes e.g., "Bearer", "Token"
	TokenPrefixes []string
}
