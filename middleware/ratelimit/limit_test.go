// Package ratelimit provides a rate limiter middleware.
package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/middleware/ratelimitconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
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

func TestLimiterManager_GetKeyFromGin(t *testing.T) {
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

	cfgs := []*ratelimitconf.RouteLimitConfig{
		{EnableRoute: true},
		{EnableIP: true},
		{EnableUserID: true},
	}

	for i, cfg := range cfgs {
		name := ""
		switch i {
		case 0:
			name = "EnableRoute"
		case 1:
			name = "EnableIP"
		case 2:
			name = "EnableUserID"
		}
		t.Run(name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request, _ = http.NewRequest("GET", "/path", nil)
			c.Request.URL.Path = "/urlpath"
			c.Request.RequestURI = "/urlpath"
			c.Set("somekey", "someval")
			c.Request.RemoteAddr = "127.0.0.1:12345"
			c.Request = c.Request.WithContext(contextkeys.SetUserID(c.Request.Context(), "uid123"))

			key := mgr.getKeyFromGin(c, cfg)

			assert.NotEmpty(t, key)
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
	assert.GreaterOrEqual(t, duration, time.Second)
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
