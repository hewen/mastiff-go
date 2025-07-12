// Package store mock redis
package store

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/hewen/mastiff-go/config/storeconf"
)

var miniredisRun = miniredis.Run

// InitMockRedis initializes a mock Redis connection using miniredis.
func InitMockRedis() *redis.Client {
	s, err := miniredisRun()
	if err != nil {
		panic(err)
	}

	conn, _ := InitRedis(storeconf.RedisConfig{
		Addr:               s.Addr(),
		RegisterHookDriver: true,
	})
	return conn
}
