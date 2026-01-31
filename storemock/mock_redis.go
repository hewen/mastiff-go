// Package storemock redis
package storemock

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/hewen/mastiff-go/config/storeconf"
	"github.com/hewen/mastiff-go/store"
)

var miniredisRun = miniredis.Run

// InitMockRedis initializes a mock Redis connection using miniredis.
func InitMockRedis() *redis.Client {
	s, err := miniredisRun()
	if err != nil {
		panic(err)
	}

	conn, _ := store.InitRedis(storeconf.RedisConfig{
		Addr:               s.Addr(),
		RegisterHookDriver: true,
	})
	return conn
}
