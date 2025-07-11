package circuitbreaker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Config{
		MaxRequests: 5,
		Interval:    10,
		Timeout:     20,
	}

	assert.Equal(t, uint32(5), cfg.MaxRequests)
	assert.Equal(t, int64(10), cfg.Interval)
	assert.Equal(t, int64(20), cfg.Timeout)
}
