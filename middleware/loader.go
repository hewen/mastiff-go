// Package middleware provides middleware for HTTP, gRPC, Gin, Fiber servers.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/config/middleware/authconf"
	"github.com/hewen/mastiff-go/config/middleware/circuitbreakerconf"
	"github.com/hewen/mastiff-go/config/middleware/ratelimitconf"
	"github.com/hewen/mastiff-go/middleware/auth"
	"github.com/hewen/mastiff-go/middleware/circuitbreaker"
	"github.com/hewen/mastiff-go/middleware/logging"
	"github.com/hewen/mastiff-go/middleware/metrics"
	"github.com/hewen/mastiff-go/middleware/ratelimit"
	"github.com/hewen/mastiff-go/middleware/recovery"
	"github.com/hewen/mastiff-go/middleware/timeout"
	"google.golang.org/grpc"
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

// LoadGRPCMiddlewares loads gRPC middlewares based on the provided configuration.
func LoadGRPCMiddlewares(conf Config) []grpc.UnaryServerInterceptor {
	conf.SetDefaults()

	var result []grpc.UnaryServerInterceptor

	if isEnabled(conf.EnableLogging) {
		result = append(result, logging.UnaryServerInterceptor())
	}
	if isEnabled(conf.EnableRecovery) {
		result = append(result, recovery.UnaryServerInterceptor())
	}
	if conf.TimeoutSeconds != nil && *conf.TimeoutSeconds > 0 {
		result = append(result, timeout.UnaryServerInterceptor(time.Duration(*conf.TimeoutSeconds)*time.Second))
	}
	if conf.Auth != nil {
		result = append(result, auth.UnaryServerInterceptor(*conf.Auth))
	}
	if conf.CircuitBreaker != nil {
		mgr := circuitbreaker.NewManager(conf.CircuitBreaker)
		result = append(result, circuitbreaker.UnaryServerInterceptor(mgr))
	}
	if conf.RateLimit != nil {
		mgr := ratelimit.NewLimiterManager(conf.RateLimit)
		result = append(result, ratelimit.UnaryServerInterceptor(mgr))
	}
	if isEnabled(conf.EnableMetrics) {
		result = append(result, metrics.UnaryServerInterceptor())
	}

	return result
}

// LoadGinMiddlewares loads Gin middlewares based on the provided configuration.
func LoadGinMiddlewares(conf Config) []gin.HandlerFunc {
	conf.SetDefaults()

	var result []gin.HandlerFunc

	if isEnabled(conf.EnableLogging) {
		result = append(result, logging.GinMiddleware())
	}
	if isEnabled(conf.EnableRecovery) {
		result = append(result, recovery.GinMiddleware())
	}
	if conf.Auth != nil {
		result = append(result, auth.GinMiddleware(conf.Auth))
	}
	if conf.CircuitBreaker != nil {
		mgr := circuitbreaker.NewManager(conf.CircuitBreaker)
		result = append(result, circuitbreaker.GinMiddleware(mgr))
	}
	if conf.RateLimit != nil {
		mgr := ratelimit.NewLimiterManager(conf.RateLimit)
		result = append(result, ratelimit.GinMiddleware(mgr))
	}
	if isEnabled(conf.EnableMetrics) {
		result = append(result, metrics.GinMiddleware())
	}

	return result
}

// LoadFiberMiddlewares loads Fiber middlewares based on the provided configuration.
func LoadFiberMiddlewares(conf Config) []func(*fiber.Ctx) error {
	conf.SetDefaults()

	var result []func(*fiber.Ctx) error

	if isEnabled(conf.EnableLogging) {
		result = append(result, logging.FiberMiddleware())
	}
	if isEnabled(conf.EnableRecovery) {
		result = append(result, recovery.FiberMiddleware())
	}
	if conf.Auth != nil {
		result = append(result, auth.FiberMiddleware(conf.Auth))
	}
	if conf.CircuitBreaker != nil {
		mgr := circuitbreaker.NewManager(conf.CircuitBreaker)
		result = append(result, circuitbreaker.FiberMiddleware(mgr))
	}
	if conf.RateLimit != nil {
		mgr := ratelimit.NewLimiterManager(conf.RateLimit)
		result = append(result, ratelimit.FiberMiddleware(mgr))
	}
	if isEnabled(conf.EnableMetrics) {
		result = append(result, metrics.FiberMiddleware())
	}

	return result
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

// isEnabled returns true if the flag is nil or true.
func isEnabled(flag *bool) bool {
	return flag == nil || *flag
}
