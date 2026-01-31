// Package store redis database
package store

import (
	"crypto/tls"

	"github.com/go-redis/redis/v7"
	"github.com/hewen/mastiff-go/config/storeconf"
)

// InitRedis initializes a Redis connection.
func InitRedis(conf storeconf.RedisConfig) (*redis.Client, error) {
	opt := &redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	}

	if conf.TLSConfig != nil && conf.TLSConfig.Enabled {
		var minVersion uint16 = tls.VersionTLS12
		if conf.TLSConfig.VersionTLS != 0 {
			minVersion = conf.TLSConfig.VersionTLS
		}

		opt.TLSConfig = &tls.Config{
			MinVersion:         minVersion,
			ServerName:         conf.TLSConfig.ServerName,
			InsecureSkipVerify: conf.TLSConfig.InsecureSkipVerify, // #nosec
		}
	}

	redisConn := redis.NewClient(opt)

	if conf.RegisterHookDriver {
		hook := &RedisHook{}
		redisConn.AddHook(hook)
	}

	_, err := redisConn.Ping().Result()
	return redisConn, err
}
