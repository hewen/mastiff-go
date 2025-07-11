// Package ratelimit provides a rate limiter middleware.
package ratelimit

import "time"

// LimitMode represents the mode of the rate limiter.
type LimitMode string

const (
	// ModeAllow is the mode that allows requests.
	ModeAllow LimitMode = "allow"
	// ModeWait is the mode that waits for requests.
	ModeWait LimitMode = "wait"

	// cleanerInterval is the interval at which the cleaner runs.
	cleanerInterval = 5 * time.Minute
	// limiterTTL is the time after which a limiter is removed from the cache.
	limiterTTL = 10 * time.Minute
)

// RouteLimitConfig represents the configuration for rate limiting per route.
type RouteLimitConfig struct {
	// Mode represents the mode of the rate limiter.
	Mode LimitMode
	// EnableRoute enables rate limiting per route.
	EnableRoute bool
	// EnableIP enables rate limiting per IP.
	EnableIP bool
	// EnableUserID enables rate limiting per user ID.
	EnableUserID bool
	// Burst is the maximum number of events that can be sent in a single burst.
	Burst int
	// Rate is the maximum number of events that can be sent per second.
	Rate float64
}

// Config represents the configuration for rate limiting.
type Config struct {
	// Default represents the default configuration for rate limiting.
	Default *RouteLimitConfig
	// PerRoute represents the configuration for rate limiting per route.
	PerRoute map[string]*RouteLimitConfig
}
