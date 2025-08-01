package circuitbreaker

import (
	"testing"

	"github.com/hewen/mastiff-go/config/middlewareconf/circuitbreakerconf"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
)

func TestConsecutiveFailuresPolicy_ShouldTrip(t *testing.T) {
	p := &ConsecutiveFailuresPolicy{ConsecutiveFailures: 3}

	tests := []struct {
		failures uint32
		expected bool
	}{
		{2, false},
		{3, true},
		{4, true},
	}

	for _, tt := range tests {
		counts := gobreaker.Counts{ConsecutiveFailures: tt.failures}
		assert.Equal(t, tt.expected, p.ShouldTrip(counts))
	}
}

func TestFailureRatePolicy_ShouldTrip(t *testing.T) {
	p := &FailureRatePolicy{MinRequests: 5, FailureRateThreshold: 0.5}

	tests := []struct {
		requests     uint32
		failures     uint32
		expectedTrip bool
	}{
		{4, 2, false}, // Not enough requests
		{10, 4, false},
		{10, 5, true},
		{10, 6, true},
	}

	for _, tt := range tests {
		counts := gobreaker.Counts{Requests: tt.requests, TotalFailures: tt.failures}
		assert.Equal(t, tt.expectedTrip, p.ShouldTrip(counts))
	}
}

func TestNewPolicyFromConfig(t *testing.T) {
	// nil config should apply defaults
	p := NewPolicyFromConfig(nil)
	assert.IsType(t, &FailureRatePolicy{}, p)

	// test consecutive_failures
	cfg1 := &circuitbreakerconf.PolicyConfig{
		Type:                "consecutive_failures",
		ConsecutiveFailures: 10,
	}
	p1 := NewPolicyFromConfig(cfg1)
	cfp, ok := p1.(*ConsecutiveFailuresPolicy)
	assert.True(t, ok)
	assert.Equal(t, uint32(10), cfp.ConsecutiveFailures)

	// test failure_rate
	cfg2 := &circuitbreakerconf.PolicyConfig{
		Type:                 "failure_rate",
		MinRequests:          20,
		FailureRateThreshold: 0.6,
	}
	p2 := NewPolicyFromConfig(cfg2)
	frp, ok := p2.(*FailureRatePolicy)
	assert.True(t, ok)
	assert.Equal(t, uint32(20), frp.MinRequests)
	assert.Equal(t, 0.6, frp.FailureRateThreshold)

	// test fallback to default on unknown type
	cfg3 := &circuitbreakerconf.PolicyConfig{
		Type:                 "unknown_type",
		MinRequests:          15,
		FailureRateThreshold: 0.75,
	}
	p3 := NewPolicyFromConfig(cfg3)
	frp2, ok := p3.(*FailureRatePolicy)
	assert.True(t, ok)
	assert.EqualValues(t, 15, frp2.MinRequests)
	assert.EqualValues(t, 0.75, frp2.FailureRateThreshold)
}
