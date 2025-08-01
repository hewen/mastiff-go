package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hewen/mastiff-go/config/middlewareconf"
	"github.com/hewen/mastiff-go/config/middlewareconf/authconf"
	"github.com/hewen/mastiff-go/config/middlewareconf/circuitbreakerconf"
	"github.com/hewen/mastiff-go/config/middlewareconf/ratelimitconf"
)

func TestLoadGRPCMiddlewares(t *testing.T) {
	t.Run("All features enabled", func(t *testing.T) {
		timeoutSec := 5
		enable := true

		conf := middlewareconf.Config{
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
		conf := middlewareconf.Config{}
		mws := LoadGRPCMiddlewares(conf)
		assert.NotEmpty(t, mws)
	})
}

func TestLoadHttpxMiddlewares(t *testing.T) {
	t.Run("All features enabled", func(t *testing.T) {
		enable := true

		conf := middlewareconf.Config{
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

		mws := LoadHttpxMiddlewares(conf)
		assert.NotEmpty(t, mws)
		assert.GreaterOrEqual(t, len(mws), 5)
		for _, mw := range mws {
			assert.NotNil(t, mw)
		}
	})

	t.Run("Minimal config", func(t *testing.T) {
		conf := middlewareconf.Config{}
		mws := LoadHttpxMiddlewares(conf)
		assert.NotEmpty(t, mws)
	})
}
