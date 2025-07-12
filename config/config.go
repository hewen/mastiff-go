// Package config provides configuration loading and management for the application.
package config

import (
	"errors"
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/hewen/mastiff-go/config/loggerconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/config/storeconf"
	"github.com/spf13/viper"
)

// Config represents the global application configuration structure. It is thread-safe.
type Config struct {
	Logger *loggerconf.Config
	HTTP   *serverconf.HTTPConfig
	Grpc   *serverconf.GrpcConfig
	Queue  *serverconf.QueueConfig
	Mysql  *storeconf.MysqlConfig
	Redis  *storeconf.RedisConfig
	Custom any
}

var (
	cfg   *Config      // Global config instance
	mutex sync.RWMutex // Protects cfg
)

// LoadConfig loads the application configuration from file, unmarshals it into Config,
// binds custom config, and watches for live updates. The onChange callback will be called
// whenever the config file changes: - If changeErr is nil, newConfig contains the updated
// configuration. - If changeErr is non-nil, newConfig is nil and the error describes the
// problem.
//
// Note:
// - `custom` must be a pointer to a struct matching the "custom" section schema in the config file.
// - This function is safe for concurrent use.
func LoadConfig(configPath string, custom any, onChange func(newConfig *Config, changeErr error)) (*Config, error) {
	v, err := newViper(configPath)
	if err != nil {
		return nil, err
	}

	var c Config
	c.Custom = custom

	if err := UnmarshalAll(v, &c, custom); err != nil {
		return nil, fmt.Errorf("initial config unmarshal error: %w", err)
	}

	setConfig(&c)

	v.WatchConfig()
	v.OnConfigChange(func(_ fsnotify.Event) {
		var newC Config
		newC.Custom = custom

		if err := UnmarshalAll(v, &newC, custom); err != nil {
			if onChange != nil {
				onChange(nil, fmt.Errorf("config reload error: %w", err))
			}
			return
		}

		setConfig(&newC)

		if onChange != nil {
			onChange(&newC, nil)
		}
	})

	return &c, nil
}

// MustLoad loads config and panics on failure. It is thread-safe.
func MustLoad(configPath string, custom any) *Config {
	c, err := LoadConfig(configPath, custom, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return c
}

// GetConfig safely returns the current configuration. It is thread-safe.
func GetConfig() *Config {
	mutex.RLock()
	defer mutex.RUnlock()
	return cfg
}

// SetConfig allows manually overriding the global configuration.
func SetConfig(c *Config) {
	setConfig(c)
}

func setConfig(c *Config) {
	mutex.Lock()
	defer mutex.Unlock()
	cfg = c
}

func newViper(path string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", path, err)
	}
	return v, nil
}

// UnmarshalAll unmarshals the viper config into the Config struct and custom config.
func UnmarshalAll(v *viper.Viper, c *Config, custom any) error {
	if err := v.Unmarshal(c); err != nil {
		return fmt.Errorf("error unmarshaling base config: %w", err)
	}
	if custom != nil {
		// custom must be a pointer to struct matching 'custom' section schema
		sub := v.Sub("custom")
		if sub == nil {
			return errors.New("missing 'custom' section in config")
		}
		if err := sub.Unmarshal(custom); err != nil {
			return fmt.Errorf("error unmarshaling custom config: %w", err)
		}
	}
	return nil
}
