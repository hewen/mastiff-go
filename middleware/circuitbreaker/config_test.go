package circuitbreaker

import (
	"testing"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Config{
		Name:        "test",
		MaxRequests: 5,
		Interval:    10,
		Timeout:     20,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 3
		},
	}

	assert.Equal(t, "test", cfg.Name)
	assert.Equal(t, uint32(5), cfg.MaxRequests)
	assert.Equal(t, int64(10), cfg.Interval)
	assert.Equal(t, int64(20), cfg.Timeout)

	result := cfg.ReadyToTrip(gobreaker.Counts{ConsecutiveFailures: 4})
	assert.True(t, result)
}
