package store

import (
	"context"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
)

type Hooks struct{}

var sqlBeginTimeKey struct{}

func (h *Hooks) Before(ctx context.Context, _ string, _ ...interface{}) (context.Context, error) {
	return context.WithValue(ctx, sqlBeginTimeKey, time.Now()), nil
}

func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	begin := ctx.Value(sqlBeginTimeKey).(time.Time)
	l := logger.NewLoggerWithContext(ctx)
	l.Infof("SQL | %10s | %s %v", util.FormatDuration(time.Since(begin)), query, args)
	return ctx, nil
}
