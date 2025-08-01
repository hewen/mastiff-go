package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	t.Run("HTTPTimeoutDefault", func(t *testing.T) {
		assert.Equal(t, int(10), HTTPTimeoutDefault)
	})
}

func TestErrors(t *testing.T) {
	t.Run("ErrEmptyHTTPConf", func(t *testing.T) {
		assert.NotNil(t, ErrEmptyHTTPConf)
		assert.Equal(t, "http config is empty", ErrEmptyHTTPConf.Error())
	})
}

func TestToDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected time.Duration
	}{
		{
			name:     "zero value uses default",
			input:    0,
			expected: HTTPTimeoutDefault * time.Second,
		},
		{
			name:     "positive value",
			input:    5,
			expected: 5 * time.Second,
		},
		{
			name:     "large value",
			input:    3600,
			expected: 3600 * time.Second,
		},
		{
			name:     "one second",
			input:    1,
			expected: 1 * time.Second,
		},
		{
			name:     "default timeout value",
			input:    HTTPTimeoutDefault,
			expected: HTTPTimeoutDefault * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDuration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToDurationEdgeCases(t *testing.T) {
	t.Run("negative value", func(t *testing.T) {
		// Negative values should still be converted to duration
		result := toDuration(-5)
		expected := -5 * time.Second
		assert.Equal(t, expected, result)
	})

	t.Run("very large value", func(t *testing.T) {
		// Test with a very large value
		result := toDuration(9223372036) // Close to max int64 when multiplied by time.Second
		expected := 9223372036 * time.Second
		assert.Equal(t, expected, result)
	})

	t.Run("boundary values", func(t *testing.T) {
		// Test boundary values around zero
		testCases := []int64{-1, 0, 1}
		for _, tc := range testCases {
			result := toDuration(tc)
			if tc == 0 {
				assert.Equal(t, HTTPTimeoutDefault*time.Second, result)
			} else {
				assert.Equal(t, time.Duration(tc)*time.Second, result)
			}
		}
	})
}

func TestToDurationConsistency(t *testing.T) {
	t.Run("multiple calls with same input", func(t *testing.T) {
		// Ensure function is deterministic
		input := int64(30)
		result1 := toDuration(input)
		result2 := toDuration(input)
		assert.Equal(t, result1, result2)
	})

	t.Run("zero value consistency", func(t *testing.T) {
		// Ensure zero always returns default
		result1 := toDuration(0)
		result2 := toDuration(0)
		assert.Equal(t, result1, result2)
		assert.Equal(t, HTTPTimeoutDefault*time.Second, result1)
		assert.Equal(t, HTTPTimeoutDefault*time.Second, result2)
	})
}

func TestToDurationMathematicalProperties(t *testing.T) {
	t.Run("conversion accuracy", func(t *testing.T) {
		// Test that conversion is mathematically correct
		input := int64(42)
		result := toDuration(input)

		// Convert back to seconds
		seconds := int64(result / time.Second)
		assert.Equal(t, input, seconds)
	})

	t.Run("time unit relationships", func(t *testing.T) {
		// Test relationships between different time units
		oneMinute := toDuration(60)
		sixtyUnits := toDuration(1) * 60
		assert.Equal(t, oneMinute, sixtyUnits)

		oneHour := toDuration(3600)
		sixtyMinutes := toDuration(60) * 60
		assert.Equal(t, oneHour, sixtyMinutes)
	})
}

func TestToDurationWithTimeOperations(t *testing.T) {
	t.Run("duration arithmetic", func(t *testing.T) {
		d1 := toDuration(10)
		d2 := toDuration(20)
		sum := d1 + d2
		expected := toDuration(30)
		assert.Equal(t, expected, sum)
	})

	t.Run("duration comparison", func(t *testing.T) {
		short := toDuration(5)
		long := toDuration(10)
		assert.True(t, short < long)
		assert.True(t, long > short)
		assert.False(t, short > long)
	})

	t.Run("duration with time.Now", func(t *testing.T) {
		// Test that duration can be used with time operations
		duration := toDuration(1)
		now := time.Now()
		future := now.Add(duration)
		assert.True(t, future.After(now))
		assert.Equal(t, duration, future.Sub(now))
	})
}

func TestToDurationDocumentationExamples(t *testing.T) {
	t.Run("common timeout values", func(t *testing.T) {
		// Test common timeout values used in HTTP servers
		commonTimeouts := map[int64]time.Duration{
			5:   5 * time.Second,   // Short timeout
			30:  30 * time.Second,  // Medium timeout
			60:  60 * time.Second,  // Long timeout
			300: 300 * time.Second, // Very long timeout
		}

		for input, expected := range commonTimeouts {
			result := toDuration(input)
			assert.Equal(t, expected, result, "Failed for input %d", input)
		}
	})

	t.Run("default behavior demonstration", func(t *testing.T) {
		// Demonstrate the default behavior when input is 0
		defaultDuration := toDuration(0)
		explicitDefault := toDuration(HTTPTimeoutDefault)

		assert.Equal(t, defaultDuration, explicitDefault)
		assert.Equal(t, 10*time.Second, defaultDuration)
	})
}
