package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
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

	_, err := RedisConn.Set("test", 1, time.Minute).Result()
	assert.Nil(t, err)
	_, err = RedisConn.Get("test").Result()
	assert.Nil(t, err)

	ctx := context.TODO()
	_, err = RedisConn.Pipelined(func(pipe redis.Pipeliner) error {
		_, err = pipe.Get("test").Result()

		_, err = pipe.ExecContext(ctx)
		if err != nil {
			return err
		}

		return nil
	})
	assert.Nil(t, err)
}
