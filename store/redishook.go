package store

import (
	"context"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
	"github.com/go-redis/redis/v7"
)

type RedisHook struct{}

func (*RedisHook) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, sqlBeginTimeKey, time.Now()), nil
}

func (*RedisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	begin := ctx.Value(sqlBeginTimeKey).(time.Time)
	l := logger.NewLoggerWithContext(ctx)
	l.Infof("REDIS | %10s | %v", util.FormatDuration(time.Since(begin)), cmd)
	return nil
}

func (*RedisHook) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, sqlBeginTimeKey, time.Now()), nil
}

func (*RedisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	begin := ctx.Value(sqlBeginTimeKey).(time.Time)
	l := logger.NewLoggerWithContext(ctx)
	l.Infof("REDIS | %10s | %v", util.FormatDuration(time.Since(begin)), cmds)
	return nil
}
