// Package middleware provides middleware for HTTP, gRPC, and Gin servers.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
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
	Auth           *authconf.Config           // Auth middleware configuration
	RateLimit      *ratelimitconf.Config      // Rate limit middleware configuration
	CircuitBreaker *circuitbreakerconf.Config // Circuit breaker middleware configuration
	TimeoutSeconds *int                       // Timeout seconds for requests
	EnableMetrics  *bool                      // Enable metrics middleware
	EnableRecovery *bool                      // Enable recovery middleware, default enabled
	EnableLogging  *bool                      // Enable logging middleware, default enabled
}

const (
	// defaultTimeout is the default timeout for requests.
	defaultTimeout = 30
)

// LoadGRPCMiddlewares loads gRPC middlewares based on the provided configuration.
func LoadGRPCMiddlewares(conf Config) []grpc.UnaryServerInterceptor {
	var result []grpc.UnaryServerInterceptor
	if conf.EnableLogging == nil || *conf.EnableLogging {
		result = append(result, logging.UnaryServerInterceptor())
	}
	if conf.EnableRecovery == nil || *conf.EnableRecovery {
		result = append(result, recovery.UnaryServerInterceptor())
	}
	if conf.TimeoutSeconds == nil || *conf.TimeoutSeconds > 0 {
		var timeoutSeconds = defaultTimeout
		if conf.TimeoutSeconds != nil {
			timeoutSeconds = *conf.TimeoutSeconds
		}
		result = append(result, timeout.UnaryServerInterceptor(time.Duration(timeoutSeconds)*time.Second))
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
	if conf.EnableMetrics != nil {
		result = append(result, metrics.UnaryServerInterceptor())
	}
	return result
}

// LoadGinMiddlewares loads Gin middlewares based on the provided configuration.
func LoadGinMiddlewares(conf Config) []gin.HandlerFunc {
	var result []gin.HandlerFunc

	if conf.EnableRecovery == nil || *conf.EnableRecovery {
		result = append(result, recovery.GinMiddleware())
	}
	if conf.EnableLogging == nil || *conf.EnableLogging {
		result = append(result, logging.GinMiddleware())
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
	if conf.EnableMetrics != nil {
		result = append(result, metrics.GinMiddleware())
	}

	return result
}
