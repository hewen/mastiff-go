package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	testCase := []struct {
		out string
		in  time.Duration
	}{
		{in: time.Hour, out: "1h0m0s"},
		{in: 2 * time.Minute, out: "2m0s"},
		{in: time.Minute, out: "1m0s"},
		{in: time.Second, out: "1s"},
		{in: time.Millisecond, out: "1ms"},
	}

	for i := range testCase {
		act := FormatDuration(testCase[i].in)
		assert.Equal(t, testCase[i].out, act)
	}
}
