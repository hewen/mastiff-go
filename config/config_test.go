package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CustomAppConfig is an example struct to test the "custom" field binding.
type CustomAppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

func createTempConfigFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)
	return tmpFile
}

func verifyConfig(t *testing.T, cfg *Config, customCfg *CustomAppConfig) {
	assert.Equal(t, "mastiff-app", customCfg.Name)
	assert.Equal(t, 9000, customCfg.Port)
	assert.Equal(t, ":8080", cfg.HTTP.Addr)
	assert.True(t, cfg.HTTP.PprofEnabled)
}

func TestLoadConfig_WithCustomAndOnChange(t *testing.T) {
	const configYAML = `
logger:
  level: "debug"

http:
  addr: ":8080"
  mode: "debug"
  timeoutRead: 5
  timeoutWrite: 10
  pprofEnabled: true

grpc:
  addr: ":50051"
  timeout: 10
  reflection: true

queue:
  queueName: "test-queue"
  poolSize: 5
  emptySleepInterval: 100ms

mysql:
  dsn: "root:password@tcp(localhost:3306)/testdb"

redis:
  addr: "localhost:6379"
  db: 0

custom:
  name: "mastiff-app"
  port: 9000
`

	tmpFile := createTempConfigFile(t, configYAML)

	var customCfg CustomAppConfig
	onChangeCalled := false
	onChange := func(newCfg *Config, err error) {
		require.NoError(t, err)
		require.NotNil(t, newCfg)

		verifyConfig(t, newCfg, &customCfg)

		onChangeCalled = true
	}

	cfg, err := LoadConfig(tmpFile, &customCfg, onChange)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	verifyConfig(t, cfg, &customCfg)

	// 模拟调用 onChange
	onChange(cfg, nil)
	assert.True(t, onChangeCalled)
}

func TestMustLoad_PanicOnError(t *testing.T) {
	assert.Panics(t, func() {
		MustLoad(".nonexistent", nil)
	})
}

type CustomConfig struct {
	FieldA string `mapstructure:"field_a"`
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(file, []byte(content), 0600)
	assert.NoError(t, err)
	return file
}

func TestSetAndGetConfig(t *testing.T) {
	cfg := &Config{}
	SetConfig(cfg)

	got := GetConfig()
	assert.Equal(t, cfg, got)
}

func TestLoadConfig_WatchConfigChange(t *testing.T) {
	// Initial config with custom field
	content := `
logger:
  # put logger config here if needed
http:
  addr: ":8080"
custom:
  field_a: "initial"
`
	configPath := writeTempConfig(t, content)

	var mu sync.Mutex
	var changedConfig *Config
	var changedErr error
	callback := func(c *Config, err error) {
		mu.Lock()
		defer mu.Unlock()
		changedConfig = c
		changedErr = err
	}

	custom := &CustomConfig{}

	cfg, err := LoadConfig(configPath, custom, callback)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "initial", custom.FieldA)

	// Modify config file to trigger OnConfigChange
	newContent := `
logger:
http:
  addr: ":9090"
custom:
  field_a: "updated"
`
	err = os.WriteFile(configPath, []byte(newContent), 0600)
	assert.NoError(t, err)

	// Wait for the watcher to detect change (allow some buffer time)
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.NotNil(t, changedConfig)
	assert.Nil(t, changedErr)
	assert.Equal(t, "updated", custom.FieldA)
}

func TestMustLoad_PanicAndSuccess(t *testing.T) {
	// Test panic on invalid file
	assert.Panics(t, func() {
		MustLoad(".nonexistent", nil)
	})

	// Test successful load
	content := `
logger:
http:
  addr: ":8080"
`
	configPath := writeTempConfig(t, content)
	cfg := MustLoad(configPath, nil)
	assert.NotNil(t, cfg)
	assert.Equal(t, ":8080", cfg.HTTP.Addr)
}

func TestUnmarshalAll_Errors(t *testing.T) {
	v := NewViperForTest(t, `
logger:
http:
  addr: ":8080"
custom:
  field_a: "val"
`)

	c := &Config{}
	custom := &CustomConfig{}

	// Happy path
	err := UnmarshalAll(v, c, custom)
	assert.NoError(t, err)
	assert.Equal(t, "val", custom.FieldA)

	// custom section missing
	vNoCustom := NewViperForTest(t, `
logger:
http:
  addr: ":8080"
`)
	err = UnmarshalAll(vNoCustom, c, custom)
	assert.ErrorContains(t, err, "missing 'custom' section")

	// unmarshaling error: pass custom of incompatible type (e.g. int)
	err = UnmarshalAll(v, c, 123)
	assert.Error(t, err)
}

func NewViperForTest(t *testing.T, content string) *viper.Viper {
	t.Helper()
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(strings.NewReader(content))
	assert.Nil(t, err)
	return v
}
