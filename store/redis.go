// Package store redis database
package store

import (
	"github.com/go-redis/redis/v7"
)

// InitRedis initializes a Redis connection.
func InitRedis(conf RedisConf) (*redis.Client, error) {
	redisConn := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})

	if conf.RegisterHookDriver {
		hook := &RedisHook{}
		redisConn.AddHook(hook)
	}

	_, err := redisConn.Ping().Result()
	return redisConn, err
}
