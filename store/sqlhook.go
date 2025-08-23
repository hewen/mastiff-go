// Package store sql hooks
package store

import (
	"context"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/contextkeys"
	"github.com/hewen/mastiff-go/pkg/util"
)

// SQLHooks is a hook collection.
type SQLHooks struct{}

// Before records SQL execution time.
func (h *SQLHooks) Before(ctx context.Context, _ string, _ ...any) (context.Context, error) {
	return contextkeys.SetSQLBeginTime(ctx, time.Now()), nil
}

// After records SQL execution time.
func (h *SQLHooks) After(ctx context.Context, query string, args ...any) (context.Context, error) {
	begin, _ := contextkeys.GetSQLBeginTime(ctx)
	l := logger.NewLoggerWithContext(ctx)
	l.Infof("SQL | %10s | %s %v", util.FormatDuration(time.Since(begin)), query, args)
	return ctx, nil
}
