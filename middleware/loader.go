// Package middleware provides middleware for HTTP, gRPC, Gin, Fiber servers.
package middleware

import (
	"time"

	"github.com/hewen/mastiff-go/config/middlewareconf"
	"github.com/hewen/mastiff-go/middleware/auth"
	"github.com/hewen/mastiff-go/middleware/circuitbreaker"
	"github.com/hewen/mastiff-go/middleware/logging"
	"github.com/hewen/mastiff-go/middleware/metrics"
	"github.com/hewen/mastiff-go/middleware/ratelimit"
	"github.com/hewen/mastiff-go/middleware/recovery"
	"github.com/hewen/mastiff-go/middleware/timeout"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"google.golang.org/grpc"
)

// LoadGRPCMiddlewares loads gRPC middlewares based on the provided configuration.
func LoadGRPCMiddlewares(conf middlewareconf.Config) []grpc.UnaryServerInterceptor {

	conf.SetDefaults()

	var result []grpc.UnaryServerInterceptor

	if IsEnabled(conf.EnableLogging) {
		result = append(result, logging.UnaryServerInterceptor())
	}
	if IsEnabled(conf.EnableRecovery) {
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
	if IsEnabled(conf.EnableMetrics) {
		result = append(result, metrics.UnaryServerInterceptor())
	}

	return result
}

// LoadHttpxMiddlewares loads Fiber middlewares based on the provided configuration.
func LoadHttpxMiddlewares(conf middlewareconf.Config) []func(unicontext.UniversalContext) error {
	conf.SetDefaults()

	var result []func(unicontext.UniversalContext) error

	if IsEnabled(conf.EnableLogging) {
		result = append(result, logging.HttpxMiddleware())
	}
	if IsEnabled(conf.EnableRecovery) {
		result = append(result, recovery.HttpxMiddleware())
	}
	if conf.Auth != nil {
		result = append(result, auth.HttpxMiddleware(conf.Auth))
	}
	if conf.CircuitBreaker != nil {
		mgr := circuitbreaker.NewManager(conf.CircuitBreaker)
		result = append(result, circuitbreaker.HttpxMiddleware(mgr))
	}
	if conf.RateLimit != nil {
		mgr := ratelimit.NewLimiterManager(conf.RateLimit)
		result = append(result, ratelimit.HttpxMiddleware(mgr))
	}
	if IsEnabled(conf.EnableMetrics) {
		result = append(result, metrics.HttpxMiddleware())
	}

	return result
}

// IsEnabled returns true if the flag is nil or true.
func IsEnabled(flag *bool) bool {
	return flag == nil || *flag
}
