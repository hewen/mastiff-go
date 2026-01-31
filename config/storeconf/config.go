// Package storeconf provides configuration for database connections.
package storeconf

import "crypto/tls"

type (
	// MysqlConfig holds the configuration for MySQL connection.
	MysqlConfig struct {
		// DataSourceName represents the data source name for the MySQL database.
		DataSourceName string
	}

	// RedisConfig holds the configuration for Redis connection.
	RedisConfig struct {
		// TLSConfig tls config for redis
		TLSConfig *tls.Config
		// Addr represents the address of the Redis server.
		Addr string
		// Password represents the password for the Redis server.
		Password string
		// DB represents the database index for the Redis server.
		DB int
		// RegisterHookDriver represents whether to register the driver hook.
		RegisterHookDriver bool
	}
)
