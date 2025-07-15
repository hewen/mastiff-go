package middleware

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hewen/mastiff-go/config/middleware/authconf"
	"github.com/hewen/mastiff-go/config/middleware/circuitbreakerconf"
	"github.com/hewen/mastiff-go/config/middleware/ratelimitconf"
)

func TestLoadGRPCMiddlewares(t *testing.T) {
	t.Run("All features enabled", func(t *testing.T) {
		timeoutSec := 5
		enable := true

		conf := Config{
			Auth: &authconf.Config{
				JWTSecret:     "secret",
				WhiteList:     []string{"/health"},
				HeaderKey:     "Authorization",
				TokenPrefixes: []string{"Bearer"},
			},
			CircuitBreaker: &circuitbreakerconf.Config{
				MaxRequests: 5,
				Interval:    60,
				Timeout:     10,
			},
			RateLimit: &ratelimitconf.Config{
				Default: &ratelimitconf.RouteLimitConfig{
					Rate:  5,
					Burst: 10,
				},
			},
			EnableMetrics:  &enable,
			EnableRecovery: &enable,
			TimeoutSeconds: &timeoutSec,
		}

		mws := LoadGRPCMiddlewares(conf)
		assert.NotEmpty(t, mws)
		assert.GreaterOrEqual(t, len(mws), 5)
		for _, mw := range mws {
			assert.NotNil(t, mw)
		}
	})

	t.Run("Minimal config", func(t *testing.T) {
		conf := Config{}
		mws := LoadGRPCMiddlewares(conf)
		assert.NotEmpty(t, mws)
	})
}

func TestLoadGinMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("All features enabled", func(t *testing.T) {
		enable := true

		conf := Config{
			Auth: &authconf.Config{
				JWTSecret:     "secret",
				WhiteList:     []string{"/health"},
				HeaderKey:     "Authorization",
				TokenPrefixes: []string{"Bearer"},
			},
			CircuitBreaker: &circuitbreakerconf.Config{
				MaxRequests: 5,
				Interval:    60,
				Timeout:     10,
			},
			RateLimit: &ratelimitconf.Config{
				Default: &ratelimitconf.RouteLimitConfig{
					Rate:  5,
					Burst: 10,
				},
			},
			EnableMetrics:  &enable,
			EnableRecovery: &enable,
		}

		mws := LoadGinMiddlewares(conf)
		assert.NotEmpty(t, mws)
		assert.GreaterOrEqual(t, len(mws), 5)
		for _, mw := range mws {
			assert.NotNil(t, mw)
		}
	})

	t.Run("Minimal config", func(t *testing.T) {
		conf := Config{}
		mws := LoadGinMiddlewares(conf)
		assert.NotEmpty(t, mws)
	})
}

func TestLoadFiberMiddlewares(t *testing.T) {
	t.Run("All features enabled", func(t *testing.T) {
		enable := true

		conf := Config{
			Auth: &authconf.Config{
				JWTSecret:     "secret",
				WhiteList:     []string{"/health"},
				HeaderKey:     "Authorization",
				TokenPrefixes: []string{"Bearer"},
			},
			CircuitBreaker: &circuitbreakerconf.Config{
				MaxRequests: 5,
				Interval:    60,
				Timeout:     10,
			},
			RateLimit: &ratelimitconf.Config{
				Default: &ratelimitconf.RouteLimitConfig{
					Rate:  5,
					Burst: 10,
				},
			},
			EnableMetrics:  &enable,
			EnableRecovery: &enable,
		}

		mws := LoadFiberMiddlewares(conf)
		assert.NotEmpty(t, mws)
		assert.GreaterOrEqual(t, len(mws), 5)
		for _, mw := range mws {
			assert.NotNil(t, mw)
		}
	})

	t.Run("Minimal config", func(t *testing.T) {
		conf := Config{}
		mws := LoadFiberMiddlewares(conf)
		assert.NotEmpty(t, mws)
	})
}
