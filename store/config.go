// Package store config
package store

type (
	// MysqlConf holds the configuration for MySQL connection.
	MysqlConf struct {
		DataSourceName     string
		RegisterHookDriver bool
	}

	// RedisConf holds the configuration for Redis connection.
	RedisConf struct {
		Addr               string
		Password           string
		DB                 int
		RegisterHookDriver bool
	}
)
