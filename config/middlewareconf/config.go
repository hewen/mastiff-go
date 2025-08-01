// Package middlewareconf provides configuration for middleware.
package middlewareconf

import (
	"github.com/hewen/mastiff-go/config/middlewareconf/authconf"
	"github.com/hewen/mastiff-go/config/middlewareconf/circuitbreakerconf"
	"github.com/hewen/mastiff-go/config/middlewareconf/ratelimitconf"
)

// Config is the configuration for middleware.
type Config struct {
	// Auth middleware configuration
	Auth *authconf.Config
	// Rate limit middleware configuration
	RateLimit *ratelimitconf.Config
	// Circuit breaker middleware configuration
	CircuitBreaker *circuitbreakerconf.Config
	// Timeout seconds for requests
	TimeoutSeconds *int
	// Enable metrics middleware
	EnableMetrics *bool
	// Enable recovery middleware, default enabled
	EnableRecovery *bool
	// Enable logging middleware, default enabled
	EnableLogging *bool
}

// SetDefaults sets default values for the configuration.
func (c *Config) SetDefaults() {
	if c.EnableLogging == nil {
		b := true
		c.EnableLogging = &b
	}
	if c.EnableRecovery == nil {
		b := true
		c.EnableRecovery = &b
	}
	if c.EnableMetrics == nil {
		b := false
		c.EnableMetrics = &b
	}
	if c.TimeoutSeconds == nil {
		d := 30
		c.TimeoutSeconds = &d
	}
}
