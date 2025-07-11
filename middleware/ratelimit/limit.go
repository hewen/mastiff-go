// Package ratelimit provides a rate limiter middleware for Gin and gRPC.
package ratelimit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/peer"
)

// routeLimiter represents a limiter for a route.
type routeLimiter struct {
	limiter  *rate.Limiter
	lastUsed time.Time
	mode     LimitMode
}

// LimiterManager manages the rate limiters.
type LimiterManager struct {
	config   *Config
	limiters map[string]*routeLimiter
	stopCh   chan struct{}
	mu       sync.RWMutex
}

// NewLimiterManager creates a new LimiterManager.
func NewLimiterManager(cfg *Config) *LimiterManager {
	mgr := &LimiterManager{
		limiters: make(map[string]*routeLimiter),
		config:   cfg,
		stopCh:   make(chan struct{}),
	}
	go mgr.cleaner()
	return mgr
}

// Stop stops the LimiterManager.
func (mgr *LimiterManager) Stop() {
	close(mgr.stopCh)
}

// cleaner removes old limiters from the cache.
func (mgr *LimiterManager) cleaner() {
	ticker := time.NewTicker(cleanerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mgr.cleanerOnce()
		case <-mgr.stopCh:
			return
		}
	}
}

// cleanerOnce removes old limiters from the cache.
func (mgr *LimiterManager) cleanerOnce() {
	mgr.mu.Lock()
	now := time.Now()
	for k, l := range mgr.limiters {
		if now.Sub(l.lastUsed) > limiterTTL {
			delete(mgr.limiters, k)
		}
	}
	mgr.mu.Unlock()
}

// getKeyFromGin returns the key for the limiter from the gin context.
func (mgr *LimiterManager) getKeyFromGin(ctx *gin.Context, cfg *RouteLimitConfig) string {
	parts := []string{}
	if cfg.EnableRoute {
		route := ctx.FullPath()
		if route == "" {
			route = ctx.Request.URL.Path
		}
		parts = append(parts, route)
	}
	if cfg.EnableIP {
		parts = append(parts, ctx.ClientIP())
	}
	if cfg.EnableUserID {
		if uid, ok := contextkeys.GetUserID(ctx.Request.Context()); ok {
			parts = append(parts, fmt.Sprint(uid))
		}
	}
	return strings.Join(parts, "|")
}

// getKeyFromContext returns the key for the limiter from the context.
func (mgr *LimiterManager) getKeyFromContext(ctx context.Context, route string, cfg *RouteLimitConfig) string {
	parts := []string{}
	if cfg.EnableRoute {
		parts = append(parts, route)
	}
	if cfg.EnableIP {
		if pr, _ := peer.FromContext(ctx); pr != nil {
			parts = append(parts, pr.Addr.String())
		}
	}
	if cfg.EnableUserID {
		if uid, ok := contextkeys.GetUserID(ctx); ok {
			parts = append(parts, uid)
		}
	}
	return strings.Join(parts, "|")
}

// getOrCreateLimiter returns the limiter for the key. If it doesn't exist, it creates it.
func (mgr *LimiterManager) getOrCreateLimiter(key string, cfg *RouteLimitConfig) *routeLimiter {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	if l, ok := mgr.limiters[key]; ok {
		l.lastUsed = time.Now()
		return l
	}
	lim := &routeLimiter{
		limiter:  rate.NewLimiter(rate.Limit(cfg.Rate), cfg.Burst),
		mode:     cfg.Mode,
		lastUsed: time.Now(),
	}
	mgr.limiters[key] = lim
	return lim
}

// AllowOrWait allows the request if the limiter allows it. If the limiter doesn't allow it, it waits for the limiter to allow it.
func (l *routeLimiter) AllowOrWait(ctx context.Context) error {
	switch l.mode {
	case ModeAllow:
		if l.limiter.Allow() {
			return nil
		}
		return context.DeadlineExceeded
	case ModeWait:
		return l.limiter.Wait(ctx)
	default:
		return context.DeadlineExceeded
	}
}
