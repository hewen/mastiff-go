// Package store redis hook for logging Redis commands
package store

import (
	"context"
	"net"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/redis/go-redis/v9"
)

// RedisHook implements redis.Hook interface for logging Redis commands.
type RedisHook struct{}

// DialHook is a no-op hook for dialing, it simply calls the next hook in the chain.
func (RedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

// ProcessHook logs Redis commands.
func (RedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		begin := time.Now()

		err := next(ctx, cmd)

		logger.NewLoggerWithContext(ctx).Infof(
			"REDIS | %10s | %v",
			util.FormatDuration(time.Since(begin)),
			cmd,
		)
		return err
	}
}

// ProcessPipelineHook logs Redis pipeline commands.
func (RedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		begin := time.Now()

		err := next(ctx, cmds)

		logger.NewLoggerWithContext(ctx).Infof(
			"REDIS | %10s | %v",
			util.FormatDuration(time.Since(begin)),
			cmds,
		)
		return err
	}
}
