package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	testCase := []struct {
		in  time.Duration
		out string
	}{
		{time.Hour, "1h0m0s"},
		{2 * time.Minute, "2m0s"},
		{time.Minute, "1m0s"},
		{time.Second, "1s"},
		{time.Millisecond, "1ms"},
	}

	for i := range testCase {
		act := FormatDuration(testCase[i].in)
		assert.Equal(t, testCase[i].out, act)
	}
}
