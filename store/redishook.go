// Package store redis hook for logging Redis commands
package store

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
)

// RedisHook implements redis.Hook interface for logging Redis commands.
type RedisHook struct{}

// BeforeProcess is called before Redis command is processed.
func (*RedisHook) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	// Record the time when Redis command is about to be processed.
	return contextkeys.SetRedisBeginTime(ctx, time.Now()), nil
}

// AfterProcess is called after Redis command is processed.
func (*RedisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	begin, _ := contextkeys.GetRedisBeginTime(ctx)
	l := logger.NewLoggerWithContext(ctx)
	l.Infof("REDIS | %10s | %v", util.FormatDuration(time.Since(begin)), cmd)
	return nil
}

// BeforeProcessPipeline is called before a Redis pipeline is processed.
func (*RedisHook) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return contextkeys.SetRedisBeginTime(ctx, time.Now()), nil
}

// AfterProcessPipeline is called after a Redis pipeline is processed.
func (*RedisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	begin, _ := contextkeys.GetRedisBeginTime(ctx)
	l := logger.NewLoggerWithContext(ctx)
	l.Infof("REDIS | %10s | %v", util.FormatDuration(time.Since(begin)), cmds)
	return nil
}
