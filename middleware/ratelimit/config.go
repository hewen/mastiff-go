// Package ratelimit provides a middleware that limits the number of concurrent in-flight requests.
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

// Strategy represents the strategy for rate limiting.
type Strategy struct {
	EnableRoute  bool `json:"enable_route"`
	EnableIP     bool `json:"enable_ip"`
	EnableUserID bool `json:"enable_user_id"`
}

// RouteLimitConfig represents the configuration for rate limiting per route.
type RouteLimitConfig struct {
	Rate     float64   `json:"rate"`
	Burst    int       `json:"burst"`
	Mode     LimitMode `json:"mode"`
	Strategy Strategy  `json:"strategy"`
}

// Config represents the configuration for rate limiting.
type Config struct {
	Default  *RouteLimitConfig            `json:"default"`
	PerRoute map[string]*RouteLimitConfig `json:"per_route"`
}
