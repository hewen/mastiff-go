// Package storeconf provides configuration for database connections.
package storeconf

type (
	// MysqlConfig holds the configuration for MySQL connection.
	MysqlConfig struct {
		// DataSourceName represents the data source name for the MySQL database.
		DataSourceName string
	}

	// TLSConfig holds the configuration for TLS connection.
	TLSConfig struct {
		// ServerName represents the server name for the TLS connection.
		ServerName string
		// VersionTLS represents the minimum TLS version to use.
		VersionTLS uint16
		// Enabled represents whether to enable TLS for the connection.
		Enabled bool
		// InsecureSkipVerify
		InsecureSkipVerify bool
	}

	// RedisConfig holds the configuration for Redis connection.
	RedisConfig struct {
		// TLSConfig tls config for redis
		TLSConfig *TLSConfig
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
