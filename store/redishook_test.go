package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisHook(t *testing.T) {
	s, _ := miniredis.Run()
	RedisConn := redis.NewClient(&redis.Options{
		Addr:     s.Addr(),
		Password: "",
		DB:       0,
	})

	hook := &RedisHook{}
	RedisConn.AddHook(hook)

	ctx := context.TODO()
	_, err := RedisConn.Set(ctx, "test", 1, time.Minute).Result()
	assert.Nil(t, err)
	_, err = RedisConn.Get(ctx, "test").Result()
	assert.Nil(t, err)

	_, err = RedisConn.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		_, err = pipe.Get(ctx, "test").Result()

		_, err = pipe.Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	})
	assert.Nil(t, err)
}
