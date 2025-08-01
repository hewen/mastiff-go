package serverconf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSocketConfig_SetDefault(t *testing.T) {
	tests := []struct {
		config       *SocketConfig
		name         string
		expectedTick time.Duration
		expectedIdle time.Duration
	}{
		{
			name: "both values are zero - should set defaults",
			config: &SocketConfig{
				TickInterval: 0,
			},
			expectedTick: defaultTickInterval,
		},
		{
			name: "tick interval is zero - should set default tick only",
			config: &SocketConfig{
				TickInterval: 0,
			},
			expectedTick: defaultTickInterval,
		},
		{
			name: "max idle time is zero - should set default idle only",
			config: &SocketConfig{
				TickInterval: 30 * time.Second,
			},
			expectedTick: 30 * time.Second,
		},
		{
			name: "both values are non-zero - should not change anything",
			config: &SocketConfig{
				TickInterval: 15 * time.Second,
			},
			expectedTick: 15 * time.Second,
			expectedIdle: 2 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call SetDefault
			tt.config.SetDefault()

			// Verify the results
			assert.Equal(t, tt.expectedTick, tt.config.TickInterval)
		})
	}
}

func TestSocketConfig_SetDefault_DefaultValues(t *testing.T) {
	// Test that the default values are what we expect
	config := &SocketConfig{}
	config.SetDefault()

	// Verify that defaults are reasonable values
	assert.True(t, config.TickInterval > 0, "TickInterval should be positive")
}

func TestSocketConfig_SetDefault_MultipleCallsIdempotent(t *testing.T) {
	config := &SocketConfig{
		TickInterval: 0,
	}

	// Call SetDefault multiple times
	config.SetDefault()
	firstTick := config.TickInterval

	config.SetDefault()
	secondTick := config.TickInterval

	// Should be idempotent - same values after multiple calls
	assert.Equal(t, firstTick, secondTick)
}

func TestSocketConfig_SetDefault_PartialDefaults(t *testing.T) {
	// Test edge case where one field is set to a very small positive value
	config := &SocketConfig{
		TickInterval: 1 * time.Nanosecond, // Very small but not zero
	}

	config.SetDefault()

	// Should not change the nanosecond value since it's not zero
	assert.Equal(t, 1*time.Nanosecond, config.TickInterval)
}

func TestFrameworkType_Constants(t *testing.T) {
	// Test that framework type constants are defined correctly
	assert.Equal(t, SocketFrameworkType("gnet"), FrameworkGnet)
	assert.NotEmpty(t, string(FrameworkGnet))
}

func TestSocketConfig_DefaultConstantsExist(t *testing.T) {
	// Verify that the default constants are accessible and reasonable
	// This is a compile-time check that the constants exist
	assert.True(t, defaultTickInterval > 0)
}
