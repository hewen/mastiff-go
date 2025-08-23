// Package ratelimit provides a rate limiter middleware.
package ratelimit

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/middlewareconf/ratelimitconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/pkg/contextkeys"
	"github.com/hewen/mastiff-go/server/httpx"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"github.com/stretchr/testify/assert"
)

func TestLimiterManager_CleanerOnce(t *testing.T) {
	cfg := &ratelimitconf.Config{
		Default: &ratelimitconf.RouteLimitConfig{
			Rate:         1,
			Burst:        1,
			Mode:         ratelimitconf.ModeAllow,
			EnableRoute:  true,
			EnableIP:     true,
			EnableUserID: true,
		},
	}
	mgr := NewLimiterManager(cfg)
	defer mgr.Stop()

	ctx := context.Background()
	ctx = contextkeys.SetValue(ctx, "ip", "1.2.3.4")
	ctx = contextkeys.SetUserID(ctx, "u1")
	key := mgr.getKeyFromContext(ctx, "/cleaneronce", cfg.Default)
	mgr.getOrCreateLimiter(key, cfg.Default)

	mgr.mu.Lock()
	mgr.limiters[key].lastUsed = time.Now().Add(-11 * time.Minute)
	mgr.mu.Unlock()

	mgr.cleanerOnce()

	mgr.mu.RLock()
	_, exist := mgr.limiters[key]
	mgr.mu.RUnlock()
	assert.False(t, exist)
}

func TestLimiterManager_GetKeyFromHttpx(t *testing.T) {
	cfgs := []*ratelimitconf.RouteLimitConfig{
		{EnableRoute: true},
		{EnableIP: true},
		{EnableUserID: true},
	}

	for k, v := range cfgs {
		name := ""
		switch k {
		case 0:
			name = "EnableRoute"
		case 1:
			name = "EnableIP"
		case 2:
			name = "EnableUserID"
		}
		t.Run(name, func(t *testing.T) {
			cfg := &ratelimitconf.Config{
				Default: &ratelimitconf.RouteLimitConfig{
					Rate:         1,
					Burst:        1,
					Mode:         ratelimitconf.ModeAllow,
					EnableRoute:  v.EnableRoute,
					EnableIP:     v.EnableIP,
					EnableUserID: v.EnableUserID,
				},
			}

			mgr := NewLimiterManager(cfg)
			defer mgr.Stop()

			r, err := httpx.NewHTTPServer(&serverconf.HTTPConfig{
				FrameworkType: serverconf.FrameworkFiber,
			})
			assert.Nil(t, err)

			r.Use(HttpxMiddleware(mgr))
			r.Get("/path", func(c unicontext.UniversalContext) error {
				return c.String(http.StatusOK, "path")
			})

			req, _ := http.NewRequest("GET", "/path", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			req = req.WithContext(contextkeys.SetUserID(req.Context(), "uid123"))

			resp, err := r.Test(req)
			defer func() {
				_ = resp.Body.Close()
			}()
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestLimiterManager_Allow(t *testing.T) {
	cfg := &ratelimitconf.Config{
		Default: &ratelimitconf.RouteLimitConfig{
			Rate:         2,
			Burst:        1,
			Mode:         ratelimitconf.ModeAllow,
			EnableRoute:  true,
			EnableIP:     true,
			EnableUserID: true,
		},
	}
	mgr := NewLimiterManager(cfg)
	defer mgr.Stop()

	ctx := context.Background()
	ctx = contextkeys.SetValue(ctx, "ip", "1.2.3.4")
	ctx = contextkeys.SetUserID(ctx, "u1")

	key := mgr.getKeyFromContext(ctx, "/test/route", cfg.Default)
	limiter := mgr.getOrCreateLimiter(key, cfg.Default)

	err := limiter.AllowOrWait(ctx)
	assert.Nil(t, err)

	err = limiter.AllowOrWait(ctx)
	for i := 0; i < 10; i++ {
		err = limiter.AllowOrWait(ctx)
		if err != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Error(t, err)
}

func TestLimiterManager_Wait(t *testing.T) {
	cfg := &ratelimitconf.Config{
		Default: &ratelimitconf.RouteLimitConfig{
			Rate:         1,
			Burst:        1,
			Mode:         ratelimitconf.ModeWait,
			EnableRoute:  true,
			EnableIP:     true,
			EnableUserID: true,
		},
	}
	mgr := NewLimiterManager(cfg)
	defer mgr.Stop()

	ctx := context.Background()
	ctx = contextkeys.SetValue(ctx, "ip", "1.2.3.4")
	ctx = contextkeys.SetUserID(ctx, "u1")

	key := mgr.getKeyFromContext(ctx, "/test/wait", cfg.Default)
	limiter := mgr.getOrCreateLimiter(key, cfg.Default)

	err := limiter.AllowOrWait(ctx)
	assert.Nil(t, err)

	t1 := time.Now()
	err = limiter.AllowOrWait(ctx)
	duration := time.Since(t1)
	assert.Nil(t, err)
	assert.InDelta(t, 1.0, duration.Seconds(), 0.05) // ±5%
}

func TestLimiterManager_Cleanup(t *testing.T) {
	cfg := &ratelimitconf.Config{
		Default: &ratelimitconf.RouteLimitConfig{
			Rate:         10,
			Burst:        2,
			Mode:         ratelimitconf.ModeAllow,
			EnableRoute:  true,
			EnableIP:     true,
			EnableUserID: true,
		},
	}
	mgr := NewLimiterManager(cfg)

	ctx := context.Background()
	ctx = contextkeys.SetValue(ctx, "ip", "1.2.3.4")
	ctx = contextkeys.SetUserID(ctx, "u1")
	key := mgr.getKeyFromContext(ctx, "/cleanup", cfg.Default)
	_ = mgr.getOrCreateLimiter(key, cfg.Default)

	mgr.mu.Lock()
	mgr.limiters[key].lastUsed = time.Now().Add(-11 * time.Minute)
	mgr.mu.Unlock()

	go mgr.cleaner()
	mgr.cleanerOnce()

	mgr.mu.RLock()
	limit, ok := mgr.limiters[key]
	assert.Equal(t, false, ok)
	assert.Nil(t, limit)
	mgr.mu.RUnlock()

	mgr.Stop()
}
