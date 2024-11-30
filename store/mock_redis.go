package store

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
)

func InitMockRedis() *redis.Client {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	conn, _ := InitRedis(RedisConf{
		Addr:               s.Addr(),
		RegisterHookDriver: true,
	})
	return conn
}
