// Package store provides configuration for database connections.
package store

type (
	// MysqlConf holds the configuration for MySQL connection.
	MysqlConf struct {
		// DataSourceName represents the data source name for the MySQL database.
		DataSourceName string
		// RegisterHookDriver represents whether to register the driver hook.
		RegisterHookDriver bool
	}

	// RedisConf holds the configuration for Redis connection.
	RedisConf struct {
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
