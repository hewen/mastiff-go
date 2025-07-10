// Package auth provides authentication and authorization middleware.
package auth

// Config defines auth-related configuration.
type Config struct {
	JWTSecret     string   // Secret for JWT
	WhiteList     []string // Path whitelist (exact match or prefix)
	HeaderKey     string   // e.g., "Authorization"
	TokenPrefixes []string // e.g., "Bearer", "Token"
}
